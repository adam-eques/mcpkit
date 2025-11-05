package mcp

import (
	"encoding/json"
	"fmt"
)

// Role identifies the speaker of a prompt message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Content is a single block within a tool result or prompt message. Concrete
// types embed a discriminating "type" field when marshalled.
type Content interface {
	contentType() string
}

// Annotations carry optional hints about how a content block should be used.
type Annotations struct {
	Audience []Role  `json:"audience,omitempty"`
	Priority float64 `json:"priority,omitempty"`
}

// TextContent is a UTF-8 text block.
type TextContent struct {
	Text        string       `json:"text"`
	Annotations *Annotations `json:"annotations,omitempty"`
}

func (TextContent) contentType() string { return "text" }

// MarshalJSON injects the type discriminator.
func (c TextContent) MarshalJSON() ([]byte, error) {
	type alias TextContent
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "text", alias: alias(c)})
}

// ImageContent is a base64-encoded image block.
type ImageContent struct {
	Data        string       `json:"data"`
	MimeType    string       `json:"mimeType"`
	Annotations *Annotations `json:"annotations,omitempty"`
}

func (ImageContent) contentType() string { return "image" }

// MarshalJSON injects the type discriminator.
func (c ImageContent) MarshalJSON() ([]byte, error) {
	type alias ImageContent
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "image", alias: alias(c)})
}

// EmbeddedResource inlines resource contents in a result.
type EmbeddedResource struct {
	Resource ResourceContents `json:"resource"`
}

func (EmbeddedResource) contentType() string { return "resource" }

// MarshalJSON injects the type discriminator.
func (c EmbeddedResource) MarshalJSON() ([]byte, error) {
	type alias EmbeddedResource
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "resource", alias: alias(c)})
}

// Text is a convenience constructor for a text content block.
func Text(s string) TextContent { return TextContent{Text: s} }

// Textf formats and returns a text content block.
func Textf(format string, args ...any) TextContent {
	return TextContent{Text: fmt.Sprintf(format, args...)}
}

// DecodeContentList parses a JSON array of content blocks.
func DecodeContentList(raw json.RawMessage) ([]Content, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	out := make([]Content, 0, len(items))
	for _, item := range items {
		c, err := DecodeContent(item)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// DecodeContent parses a single content block, dispatching on its type field.
func DecodeContent(raw json.RawMessage) (Content, error) {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, err
	}
	switch head.Type {
	case "text":
		var c TextContent
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, err
		}
		return c, nil
	case "image":
		var c ImageContent
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, err
		}
		return c, nil
	case "resource":
		var c EmbeddedResource
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, err
		}
		return c, nil
	default:
		return nil, fmt.Errorf("mcp: unknown content type %q", head.Type)
	}
}
