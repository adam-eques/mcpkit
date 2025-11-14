package mcp

import "encoding/json"

// Prompt describes a reusable prompt template advertised to the client.
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument describes a single template argument.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// ListPromptsResult is returned by prompts/list.
type ListPromptsResult struct {
	Prompts    []Prompt `json:"prompts"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

// GetPromptParams are the parameters of prompts/get.
type GetPromptParams struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments,omitempty"`
}

// PromptMessage is one message in a rendered prompt.
type PromptMessage struct {
	Role    Role    `json:"role"`
	Content Content `json:"content"`
}

// UnmarshalJSON decodes the polymorphic content block of a prompt message.
func (m *PromptMessage) UnmarshalJSON(data []byte) error {
	var wire struct {
		Role    Role            `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	content, err := DecodeContent(wire.Content)
	if err != nil {
		return err
	}
	m.Role = wire.Role
	m.Content = content
	return nil
}

// GetPromptResult is returned by prompts/get.
type GetPromptResult struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}
