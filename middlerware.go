package tmorm

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	MethodTyp string
)
type (
	MiddlewareFunc func(mctx *MiddleCtx, next func(m *MiddleCtx))

	IOperation interface {
		GetBsonD() bson.D
	}

	// 中间流程上下文
	MiddleCtx struct {
		UserCtx   context.Context
		Typ       MethodTyp
		Operation IOperation
		Result    *MResult
	}

	MResult struct {
		Val any
		Err error
	}
)

func NewMiddleContext(ctx context.Context, typ MethodTyp) *MiddleCtx {
	return &MiddleCtx{
		UserCtx: ctx,
		Typ:     typ,
		Result:  &MResult{},
	}
}

func Executor(mctx *MiddleCtx, list []MiddlewareFunc) {
	var et func(i int, ctx *MiddleCtx)
	et = func(i int, ctx *MiddleCtx) {
		if i >= len(list) {
			return
		}
		md := list[i]
		md(mctx, func(m *MiddleCtx) {
			et(i+1, m)
		})
	}
	et(0, mctx)
}

//	=======================================================

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
		if mctx.Result.Err != nil {
			logger.Printf("[MongoDB] %s operation failed in %v: %v", mctx.Typ, duration, mctx.Result.Err)
		} else {
			logger.Printf("[MongoDB] %s operation completed in %v", mctx.Typ, duration)
		}
	}
}

func Recovery() MiddlewareFunc {
	return func(mctx *MiddleCtx, next func(m *MiddleCtx)) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[MongoDB] panic , %s operation")
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

		if duration > threshold {
			var d bson.D
			if mctx.Operation != nil {
				d = mctx.Operation.GetBsonD()
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
		for i := 0; i <= maxRetries; i++ {
			next(mctx)
			err := mctx.Result.Err
			if err == nil {
				return
			}

			lastErr = err

			// 检查是否为可重试的错误
			if !isRetryableError(err) {
				break
			}

			if i < maxRetries {
				log.Printf("[RETRY] %s operation failed (attempt %d/%d): %v, retrying in %v",
					mctx.Typ, i+1, maxRetries+1, err, retryDelay)
				time.Sleep(retryDelay)
			}
		}

		mctx.Result.Err = fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
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
		if contains(errorMsg, msg) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
