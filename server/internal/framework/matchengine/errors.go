package matchengine

import "fmt"

// EngineError 是引擎层返回的结构化错误。
type EngineError struct {
	Code    string `json:"code"`    // 稳定的错误码，供调用方分类处理。
	Message string `json:"message"` // 面向日志或调用方的错误说明。
}

func (e *EngineError) Error() string {
	return e.Code + ": " + e.Message
}

func newError(code, format string, args ...interface{}) *EngineError {
	return &EngineError{Code: code, Message: fmt.Sprintf(format, args...)}
}
