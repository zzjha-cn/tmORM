package tmorm

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type (
	// 错误类型分类
	ErrorType  string
	ErrorLevel string

	// 增强的结果结构体
	MResult struct {
		Val       any            `json:"value,omitempty"`
		Err       error          `json:"-"`
		Code      string         `json:"code,omitempty"`
		Message   string         `json:"message,omitempty"`
		ErrorType ErrorType      `json:"error_type,omitempty"`
		Level     ErrorLevel     `json:"level,omitempty"`
		Details   map[string]any `json:"details,omitempty"`
		Timestamp time.Time      `json:"timestamp,omitempty"`
		Retryable bool           `json:"retryable,omitempty"`
	}

	// 错误回调函数类型
	ErrorCallback func(result *MResult, ctx *MiddleCtx)
)

// 错误类型常量
const (
	ErrorTypeConnection ErrorType = "CONNECTION"
	ErrorTypeTimeout    ErrorType = "TIMEOUT"
	ErrorTypeValidation ErrorType = "VALIDATION"
	ErrorTypeQuery      ErrorType = "QUERY"
	ErrorTypeUpdate     ErrorType = "UPDATE"
	ErrorTypeInsert     ErrorType = "INSERT"
	ErrorTypeDelete     ErrorType = "DELETE"
	ErrorTypeAggregate  ErrorType = "AGGREGATE"
	ErrorTypeAuth       ErrorType = "AUTH"
	ErrorTypeIndex      ErrorType = "INDEX"
	ErrorTypeUnknown    ErrorType = "UNKNOWN"
)

// 错误级别常量
const (
	ErrorLevelInfo    ErrorLevel = "INFO"
	ErrorLevelWarning ErrorLevel = "WARNING"
	ErrorLevelError   ErrorLevel = "ERROR"
	ErrorLevelFatal   ErrorLevel = "FATAL"
)

func NewMResult() *MResult {
	return &MResult{
		Timestamp: time.Now(),
		Details:   make(map[string]any),
	}
}

// SetSuccess 设置成功结果
func (r *MResult) SetSuccess(value any) *MResult {
	r.Val = value
	r.Err = nil
	r.Level = ErrorLevelInfo
	r.Timestamp = time.Now()
	return r
}

// SetError 设置错误结果
func (r *MResult) SetError(err error, errorType ErrorType, code string, message string) *MResult {
	r.Err = err
	r.ErrorType = errorType
	r.Code = code
	r.Message = message
	r.Level = ErrorLevelError
	r.Timestamp = time.Now()
	r.Retryable = isRetryableError(err)

	// 根据错误类型设置详细信息
	if r.Details == nil {
		r.Details = make(map[string]any)
	}
	r.Details["original_error"] = err.Error()

	return r
}

// SetFatalError 设置致命错误
func (r *MResult) SetFatalError(err error, errorType ErrorType, code string, message string) *MResult {
	r.SetError(err, errorType, code, message)
	r.Level = ErrorLevelFatal
	r.Retryable = false
	return r
}

// SetWarning 设置警告
func (r *MResult) SetWarning(message string, details map[string]any) *MResult {
	r.Level = ErrorLevelWarning
	r.Message = message
	r.Timestamp = time.Now()
	if details != nil {
		if r.Details == nil {
			r.Details = make(map[string]any)
		}
		for k, v := range details {
			r.Details[k] = v
		}
	}
	return r
}

// AddDetail 添加详细信息
func (r *MResult) AddDetail(key string, value any) *MResult {
	if r.Details == nil {
		r.Details = make(map[string]any)
	}
	r.Details[key] = value
	return r
}

// IsSuccess 检查是否成功
func (r *MResult) IsSuccess() bool {
	return r.Err == nil
}

// IsRetryable 检查是否可重试
func (r *MResult) IsRetryable() bool {
	return r.Retryable && r.Err != nil
}

// GetErrorInfo 获取错误信息摘要
func (r *MResult) GetErrorInfo() map[string]any {
	if r.Err == nil {
		return nil
	}

	info := map[string]any{
		"code":       r.Code,
		"message":    r.Message,
		"error_type": r.ErrorType,
		"level":      r.Level,
		"retryable":  r.Retryable,
		"timestamp":  r.Timestamp,
	}

	if r.Details != nil && len(r.Details) > 0 {
		info["details"] = r.Details
	}

	return info
}

// ClassifyError 根据错误自动分类
func ClassifyError(err error) (ErrorType, string, string) {
	if err == nil {
		return ErrorTypeUnknown, "SUCCESS", "Operation completed successfully"
	}

	errorMsg := err.Error()

	// MongoDB 网络错误
	if mongo.IsNetworkError(err) {
		return ErrorTypeConnection, "NETWORK_ERROR", "Network connection error"
	}

	// MongoDB 超时错误
	if mongo.IsTimeout(err) {
		return ErrorTypeTimeout, "TIMEOUT_ERROR", "Operation timeout"
	}

	// 认证错误
	if strings.Contains(errorMsg, "authentication") || strings.Contains(errorMsg, "unauthorized") {
		return ErrorTypeAuth, "AUTH_ERROR", "Authentication failed"
	}

	// 查询错误
	if strings.Contains(errorMsg, "query") || strings.Contains(errorMsg, "find") {
		return ErrorTypeQuery, "QUERY_ERROR", "Query execution failed"
	}

	// 更新错误
	if strings.Contains(errorMsg, "update") || strings.Contains(errorMsg, "modify") {
		return ErrorTypeUpdate, "UPDATE_ERROR", "Update operation failed"
	}

	// 插入错误
	if strings.Contains(errorMsg, "insert") || strings.Contains(errorMsg, "duplicate key") {
		return ErrorTypeInsert, "INSERT_ERROR", "Insert operation failed"
	}

	// 删除错误
	if strings.Contains(errorMsg, "delete") || strings.Contains(errorMsg, "remove") {
		return ErrorTypeDelete, "DELETE_ERROR", "Delete operation failed"
	}

	// 聚合错误
	if strings.Contains(errorMsg, "aggregate") || strings.Contains(errorMsg, "pipeline") {
		return ErrorTypeAggregate, "AGGREGATE_ERROR", "Aggregation operation failed"
	}

	// 索引错误
	if strings.Contains(errorMsg, "index") {
		return ErrorTypeIndex, "INDEX_ERROR", "Index operation failed"
	}

	// 验证错误
	if strings.Contains(errorMsg, "validation") || strings.Contains(errorMsg, "invalid") {
		return ErrorTypeValidation, "VALIDATION_ERROR", "Data validation failed"
	}

	return ErrorTypeUnknown, "UNKNOWN_ERROR", "Unknown error occurred"
}

// 示例错误回调函数
// func ExampleErrorCallback(result *MResult, ctx *MiddleCtx) {
// 	log.Printf("[ERROR_CALLBACK] Operation: %s, Error: %s, Code: %s, Type: %s",
// 		ctx.Typ, result.Err.Error(), result.Code, result.ErrorType)
// 	// 可以在这里发送告警、记录到监控系统等
// 	if result.Level == ErrorLevelFatal {
// 		log.Printf("[ERROR_CALLBACK] FATAL ERROR - sending alert!")
// 		// sendAlert(result)
// 	}
// }
