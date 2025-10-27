// Package jsonrpc implements the subset of JSON-RPC 2.0 used by the Model
// Context Protocol: single (non-batched) requests, responses and notifications
// encoded as UTF-8 JSON objects.
package jsonrpc

import "encoding/json"

// Version is the only protocol version permitted on the wire.
const Version = "2.0"

// Request is a JSON-RPC request or notification. When ID is nil the message is a
// notification and the peer must not send a response.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *ID             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// IsNotification reports whether the request omits an identifier.
func (r *Request) IsNotification() bool { return r == nil || r.ID == nil }

// Response is a JSON-RPC response. Exactly one of Result or Error is set. The
// identifier is always encoded, defaulting to null when the request could not be
// parsed.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *ID             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// NewRequest builds a request that expects a response.
func NewRequest(id ID, method string, params json.RawMessage) *Request {
	return &Request{JSONRPC: Version, ID: &id, Method: method, Params: params}
}

// NewNotification builds a fire-and-forget notification.
func NewNotification(method string, params json.RawMessage) *Request {
	return &Request{JSONRPC: Version, Method: method, Params: params}
}

// NewResponse builds a successful response carrying result.
func NewResponse(id *ID, result json.RawMessage) *Response {
	return &Response{JSONRPC: Version, ID: id, Result: result}
}

// NewErrorResponse builds a failing response carrying err.
func NewErrorResponse(id *ID, err *Error) *Response {
	return &Response{JSONRPC: Version, ID: id, Error: err}
}

// EncodeResult marshals v into a response result. It never fails for values that
// the standard library can encode; encoding errors are surfaced to the caller.
func EncodeResult(id *ID, v any) (*Response, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return NewResponse(id, raw), nil
}
