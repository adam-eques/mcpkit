package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// Standard JSON-RPC 2.0 error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// Error is a JSON-RPC error object. It implements the error interface so it can
// flow through ordinary Go error handling.
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("jsonrpc: code %d: %s", e.Code, e.Message)
}

// NewError builds an error with the given code and message.
func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Errorf builds an error with a formatted message.
func Errorf(code int, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// WithData attaches structured data to the error. A marshalling failure leaves
// the data field unset rather than losing the error.
func (e *Error) WithData(v any) *Error {
	if raw, err := json.Marshal(v); err == nil {
		e.Data = raw
	}
	return e
}

// Convenience constructors for the common cases.

// ParseError reports malformed JSON.
func ParseError(msg string) *Error { return NewError(CodeParseError, orDefault(msg, "parse error")) }

// InvalidRequest reports a structurally invalid request.
func InvalidRequest(msg string) *Error {
	return NewError(CodeInvalidRequest, orDefault(msg, "invalid request"))
}

// MethodNotFound reports an unknown method.
func MethodNotFound(method string) *Error {
	return Errorf(CodeMethodNotFound, "method not found: %s", method)
}

// InvalidParams reports malformed parameters.
func InvalidParams(msg string) *Error {
	return NewError(CodeInvalidParams, orDefault(msg, "invalid params"))
}

// InternalError reports an unexpected server failure.
func InternalError(msg string) *Error {
	return NewError(CodeInternalError, orDefault(msg, "internal error"))
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
