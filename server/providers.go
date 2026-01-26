package server

import (
	"context"
	"encoding/json"

	"github.com/adam-eques/mcpkit/jsonrpc"
	"github.com/adam-eques/mcpkit/mcp"
)

func (s *Server) handleResourcesList(ctx context.Context) (json.RawMessage, *jsonrpc.Error) {
	if s.resources == nil {
		return nil, jsonrpc.MethodNotFound(mcp.MethodResourcesList)
	}
	res, err := s.resources.ListResources(ctx)
	if err != nil {
		return nil, jsonrpc.InternalError(err.Error())
	}
	return marshalResult(mcp.ListResourcesResult{Resources: res})
}

func (s *Server) handleResourcesRead(ctx context.Context, params json.RawMessage) (json.RawMessage, *jsonrpc.Error) {
	if s.resources == nil {
		return nil, jsonrpc.MethodNotFound(mcp.MethodResourcesRead)
	}
	var p mcp.ReadResourceParams
	if err := json.Unmarshal(nonNull(params), &p); err != nil {
		return nil, jsonrpc.InvalidParams(err.Error())
	}
	if p.URI == "" {
		return nil, jsonrpc.InvalidParams("resource uri is required")
	}
	contents, err := s.resources.ReadResource(ctx, p.URI)
	if err != nil {
		return nil, jsonrpc.InternalError(err.Error())
	}
	return marshalResult(mcp.ReadResourceResult{Contents: contents})
}

func (s *Server) handlePromptsList(ctx context.Context) (json.RawMessage, *jsonrpc.Error) {
	if s.prompts == nil {
		return nil, jsonrpc.MethodNotFound(mcp.MethodPromptsList)
	}
	ps, err := s.prompts.ListPrompts(ctx)
	if err != nil {
		return nil, jsonrpc.InternalError(err.Error())
	}
	return marshalResult(mcp.ListPromptsResult{Prompts: ps})
}

func (s *Server) handlePromptsGet(ctx context.Context, params json.RawMessage) (json.RawMessage, *jsonrpc.Error) {
	if s.prompts == nil {
		return nil, jsonrpc.MethodNotFound(mcp.MethodPromptsGet)
	}
	var p mcp.GetPromptParams
	if err := json.Unmarshal(nonNull(params), &p); err != nil {
		return nil, jsonrpc.InvalidParams(err.Error())
	}
	if p.Name == "" {
		return nil, jsonrpc.InvalidParams("prompt name is required")
	}
	result, err := s.prompts.GetPrompt(ctx, p.Name, p.Arguments)
	if err != nil {
		return nil, jsonrpc.InternalError(err.Error())
	}
	return marshalResult(result)
}

func (s *Server) handleSetLevel(params json.RawMessage) (json.RawMessage, *jsonrpc.Error) {
	var p mcp.SetLevelParams
	if err := json.Unmarshal(nonNull(params), &p); err != nil {
		return nil, jsonrpc.InvalidParams(err.Error())
	}
	if p.Level != "" {
		s.logLevel.Store(p.Level)
		s.log.Info("log level changed", "level", string(p.Level))
	}
	return marshalResult(struct{}{})
}
