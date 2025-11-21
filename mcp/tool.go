package mcp

import "encoding/json"

// Tool describes a callable tool advertised to the client.
type Tool struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	InputSchema json.RawMessage  `json:"inputSchema"`
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
}

// ToolAnnotations carry optional behavioural hints for a tool.
type ToolAnnotations struct {
	Title           string `json:"title,omitempty"`
	ReadOnlyHint    bool   `json:"readOnlyHint,omitempty"`
	DestructiveHint bool   `json:"destructiveHint,omitempty"`
	IdempotentHint  bool   `json:"idempotentHint,omitempty"`
	OpenWorldHint   bool   `json:"openWorldHint,omitempty"`
}

// ListToolsResult is returned by tools/list.
type ListToolsResult struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// CallToolParams are the parameters of tools/call.
type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// CallToolResult is returned by tools/call. IsError distinguishes a tool-level
// failure (reported to the model) from a protocol error.
type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// UnmarshalJSON decodes the polymorphic content array so the type is usable by
// MCP clients as well as servers.
func (r *CallToolResult) UnmarshalJSON(data []byte) error {
	var wire struct {
		Content json.RawMessage `json:"content"`
		IsError bool            `json:"isError"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	content, err := DecodeContentList(wire.Content)
	if err != nil {
		return err
	}
	r.Content = content
	r.IsError = wire.IsError
	return nil
}

// NewToolResult builds a successful result from content blocks.
func NewToolResult(content ...Content) *CallToolResult {
	return &CallToolResult{Content: content}
}

// TextResult builds a successful result containing a single text block.
func TextResult(s string) *CallToolResult {
	return &CallToolResult{Content: []Content{Text(s)}}
}

// ErrorResult builds a tool-level error result containing a text block.
func ErrorResult(s string) *CallToolResult {
	return &CallToolResult{Content: []Content{Text(s)}, IsError: true}
}
