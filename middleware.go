package tmorm

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/zzjha-cn/tm_orm/impl"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	MiddlewareFunc func(mctx *MiddleCtx, next func(m *MiddleCtx))

	IOperation interface {
		GetBsonD() bson.D
	}

	// 中间流程上下文
	MiddleCtx struct {
		UserCtx        context.Context
		Typ            MethodTyp
		Operation      IOperation
		UpdateOp       impl.IUpdateOperation
		AggOp          impl.IAggregationOperation
		Result         *MResult
		DBName         string
		CollectionName string
	}
)

func NewMiddleContext(ctx context.Context, typ MethodTyp, dbName, collectionName string) *MiddleCtx {
	return &MiddleCtx{
		UserCtx:        ctx,
		Typ:            typ,
		Result:         NewMResult(),
		DBName:         dbName,
		CollectionName: collectionName,
	}
}

func Executor(mctx *MiddleCtx, list []MiddlewareFunc) {
	var et func(i int, ctx *MiddleCtx)
	et = func(i int, ctx *MiddleCtx) {
		if i >= len(list) {
			// 如果执行后有错误，进行自动分类
			if ctx.Result.Err != nil && ctx.Result.Code == "" {
				errorType, code, message := ClassifyError(ctx.Result.Err)
				ctx.Result.SetError(ctx.Result.Err, errorType, code, message)
			}
			return
		}
		md := list[i]
		md(mctx, func(m *MiddleCtx) {
			et(i+1, m)
		})
	}
	et(0, mctx)
}

//	================================================================================================

// LoggingMiddleware 日志中间件
func LoggingMiddleware(logger *log.Logger) MiddlewareFunc {
	if logger == nil {
		logger = log.Default()
	}

	return func(mctx *MiddleCtx, next func(m *MiddleCtx)) {
		start := time.Now()
		var d bson.D
		if mctx.Operation != nil {
			d = mctx.Operation.GetBsonD()
		}
		logger.Printf("[MongoDB] Starting %s operation with filter: %+v", mctx.Typ, d)

		next(mctx)

		duration := time.Since(start)
		mctx.Result.AddDetail("duration", duration.String())
		mctx.Result.AddDetail("operation", string(mctx.Typ))

		if mctx.Result.Err != nil {
			errorInfo := mctx.Result.GetErrorInfo()
			logger.Printf("[%s] Operation: %v, Duration: %v, Code: %s, Error: %v",
				mctx.Result.Level, mctx.Typ, duration, mctx.Result.Code, mctx.Result.Err)
			if errorInfo != nil {
				logger.Printf("[ERROR_DETAILS] %+v", errorInfo)
			}
		} else {
			logger.Printf("[INFO] Operation: %v, Duration: %v, Success", mctx.Typ, duration)
		}
	}
}

func Recovery() MiddlewareFunc {
	return func(mctx *MiddleCtx, next func(m *MiddleCtx)) {
		defer func() {
			if r := recover(); r != nil {
				panicErr := fmt.Errorf("panic recovered: %v", r)
				mctx.Result.SetFatalError(panicErr, ErrorTypeUnknown, "PANIC_ERROR", "System panic occurred")
				mctx.Result.AddDetail("panic_value", r)
				mctx.Result.AddDetail("stack_trace", string(debug.Stack()))
				log.Printf("[PANIC] Operation: %v, Panic: %v", mctx.Typ, r)
			}
		}()
		next(mctx)
	}
}

// TimingMiddleware 性能监控中间件
func TimingMiddleware(threshold time.Duration) MiddlewareFunc {
	return func(mctx *MiddleCtx, next func(m *MiddleCtx)) {
		start := time.Now()
		next(mctx)

		duration := time.Since(start)
		mctx.Result.AddDetail("execution_time", duration.String())
		mctx.Result.AddDetail("threshold", threshold.String())

		if duration > threshold {
			var d bson.D
			if mctx.Operation != nil {
				d = mctx.Operation.GetBsonD()
			}
			if mctx.Result.IsSuccess() {
				mctx.Result.SetWarning(fmt.Sprintf("Slow operation detected: %v", duration),
					map[string]any{
						"duration":  duration.String(),
						"threshold": threshold.String(),
						"operation": string(mctx.Typ),
					})
			}
			log.Printf("[MongoDB SLOW QUERY] %s operation took %v (threshold: %v), filter: %+v",
				mctx.Typ, duration, threshold, d)
		}
	}
}

// RetryMiddleware 重试中间件
func RetryMiddleware(maxRetries int, retryDelay time.Duration) MiddlewareFunc {
	return func(mctx *MiddleCtx, next func(m *MiddleCtx)) {
		var lastErr error
		retryCount := 0

		for i := 0; i <= maxRetries; i++ {
			next(mctx)
			err := mctx.Result.Err
			if err == nil {
				if retryCount > 0 {
					mctx.Result.AddDetail("retry_count", retryCount)
					mctx.Result.AddDetail("retry_success", true)
				}
				return
			}

			lastErr = err

			// 检查是否为可重试的错误
			if !mctx.Result.IsRetryable() {
				mctx.Result.AddDetail("retry_stopped_reason", "error_not_retryable")
				break
			}

			if i < maxRetries {
				retryCount++
				log.Printf("[RETRY] %s operation failed (attempt %d/%d): %v, retrying in %v",
					mctx.Typ, i+1, maxRetries+1, err, retryDelay)
				time.Sleep(retryDelay)
				// 重置结果状态，但保留重试信息
				oldDetails := mctx.Result.Details
				mctx.Result = NewMResult()
				if oldDetails != nil {
					for k, v := range oldDetails {
						if strings.HasPrefix(k, "retry_") {
							mctx.Result.Details[k] = v
						}
					}
				}
			}
		}

		if mctx.Result.Err != nil {
			mctx.Result.AddDetail("retry_count", retryCount)
			mctx.Result.AddDetail("retry_success", false)
			mctx.Result.AddDetail("max_retries", maxRetries)
			mctx.Result.Err = fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
		}
	}
}

// isRetryableError 判断错误是否可重试
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// MongoDB 网络错误通常可以重试
	if mongo.IsNetworkError(err) {
		return true
	}

	// 超时错误可以重试
	if mongo.IsTimeout(err) {
		return true
	}

	// 其他特定错误码
	errorMsg := err.Error()
	retryableMessages := []string{
		"connection refused",
		"no reachable servers",
		"server selection timeout",
	}

	for _, msg := range retryableMessages {
		if strings.Contains(errorMsg, msg) {
			return true
		}
	}

	return false
}

// ErrorHandlingMiddleware 错误处理中间件，支持错误回调
func ErrorHandlingMiddleware(callback ErrorCallback) MiddlewareFunc {
	return func(mctx *MiddleCtx, next func(m *MiddleCtx)) {
		next(mctx)

		// 如果有错误且设置了回调函数
		if mctx.Result.Err != nil && callback != nil {
			// 确保错误已分类
			if mctx.Result.Code == "" {
				errorType, code, message := ClassifyError(mctx.Result.Err)
				mctx.Result.SetError(mctx.Result.Err, errorType, code, message)
			}

			// 调用错误回调
			callback(mctx.Result, mctx)
		}
	}
}

// CircuitBreakerMiddleware 熔断器中间件
func CircuitBreakerMiddleware(failureThreshold int, resetTimeout time.Duration) MiddlewareFunc {
	// 简化的熔断器实现
	var (
		failureCount int
		lastFailTime time.Time
		state        string = "CLOSED" // CLOSED, OPEN, HALF_OPEN
	)

	return func(mctx *MiddleCtx, next func(m *MiddleCtx)) {
		// 检查熔断器状态
		if state == "OPEN" {
			if time.Since(lastFailTime) > resetTimeout {
				state = "HALF_OPEN"
			} else {
				mctx.Result.SetError(
					fmt.Errorf("circuit breaker is open"),
					ErrorTypeConnection,
					"CIRCUIT_BREAKER_OPEN",
					"Circuit breaker is open, request rejected",
				)
				mctx.Result.AddDetail("circuit_breaker_state", state)
				mctx.Result.AddDetail("failure_count", failureCount)
				return
			}
		}

		next(mctx)

		// 更新熔断器状态
		if mctx.Result.Err != nil {
			failureCount++
			lastFailTime = time.Now()
			if failureCount >= failureThreshold {
				state = "OPEN"
			}
		} else {
			if state == "HALF_OPEN" {
				state = "CLOSED"
				failureCount = 0
			}
		}

		mctx.Result.AddDetail("circuit_breaker_state", state)
		mctx.Result.AddDetail("failure_count", failureCount)
	}
}
