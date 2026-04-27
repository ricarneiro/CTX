package helper

import "encoding/json"

// Request is a newline-delimited JSON-RPC request sent to the helper process.
type Request struct {
	ID     int             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// Response is a newline-delimited JSON-RPC response from the helper process.
type Response struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *RpcError       `json:"error,omitempty"`
}

// RpcError is a structured error from the helper process.
type RpcError struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *RpcError) Error() string { return e.Code + ": " + e.Message }
