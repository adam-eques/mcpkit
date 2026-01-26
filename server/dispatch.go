package server

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/adam-eques/mcpkit/jsonrpc"
	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// HandleMessage processes a single inbound frame and returns the encoded
// response, or nil when the frame was a notification. It is safe for concurrent
// use and is the integration point for both the stdio loop and the HTTP gateway.
func (s *Server) HandleMessage(ctx context.Context, raw []byte) ([]byte, error) {
	var req jsonrpc.Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return s.encode(jsonrpc.NewErrorResponse(nil, jsonrpc.ParseError("")))
	}
	if req.JSONRPC != jsonrpc.Version {
		return s.encode(jsonrpc.NewErrorResponse(req.ID, jsonrpc.InvalidRequest("unsupported jsonrpc version")))
	}
	if req.IsNotification() {
		s.handleNotification(ctx, &req)
		return nil, nil
	}

	start := time.Now()
	result, rpcErr := s.route(ctx, &req)
	s.metrics.Observe(req.Method, time.Since(start), rpcErr != nil)
	s.metrics.Inc("requests_total")
	if rpcErr != nil {
		s.metrics.Inc("errors_total")
		return s.encode(jsonrpc.NewErrorResponse(req.ID, rpcErr))
	}
	return s.encode(jsonrpc.NewResponse(req.ID, result))
}

func (s *Server) encode(resp *jsonrpc.Response) ([]byte, error) {
	return json.Marshal(resp)
}

// route dispatches a request by method, returning either a raw result or a
// JSON-RPC error.
func (s *Server) route(ctx context.Context, req *jsonrpc.Request) (json.RawMessage, *jsonrpc.Error) {
	if req.Method != mcp.MethodInitialize && req.Method != mcp.MethodPing && !s.initialized.Load() {
		return nil, jsonrpc.InvalidRequest("session not initialized")
	}
	switch req.Method {
	case mcp.MethodInitialize:
		return s.handleInitialize(req.Params)
	case mcp.MethodPing:
		return marshalResult(mcp.PingResult{})
	case mcp.MethodToolsList:
		return marshalResult(mcp.ListToolsResult{Tools: s.tools.List()})
	case mcp.MethodToolsCall:
		return s.handleToolsCall(ctx, req.Params)
	case mcp.MethodResourcesList:
		return s.handleResourcesList(ctx)
	case mcp.MethodResourcesRead:
		return s.handleResourcesRead(ctx, req.Params)
	case mcp.MethodPromptsList:
		return s.handlePromptsList(ctx)
	case mcp.MethodPromptsGet:
		return s.handlePromptsGet(ctx, req.Params)
	case mcp.MethodLoggingSetLevel:
		return s.handleSetLevel(req.Params)
	default:
		return nil, jsonrpc.MethodNotFound(req.Method)
	}
}

func (s *Server) handleInitialize(params json.RawMessage) (json.RawMessage, *jsonrpc.Error) {
	var p mcp.InitializeParams
	if err := json.Unmarshal(nonNull(params), &p); err != nil {
		return nil, jsonrpc.InvalidParams(err.Error())
	}
	s.mu.Lock()
	s.clientIn = p.ClientInfo
	s.mu.Unlock()
	s.initialized.Store(true)
	s.log.Info("session initialized",
		"client", p.ClientInfo.Name,
		"clientVersion", p.ClientInfo.Version,
		"protocol", p.ProtocolVersion)
	return marshalResult(mcp.InitializeResult{
		ProtocolVersion: mcp.NegotiateVersion(p.ProtocolVersion),
		Capabilities:    s.capabilities(),
		ServerInfo:      s.info,
		Instructions:    s.instructions,
	})
}

func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (json.RawMessage, *jsonrpc.Error) {
	var p mcp.CallToolParams
	if err := json.Unmarshal(nonNull(params), &p); err != nil {
		return nil, jsonrpc.InvalidParams(err.Error())
	}
	if p.Name == "" {
		return nil, jsonrpc.InvalidParams("tool name is required")
	}
	result, err := s.tools.Call(ctx, p.Name, p.Arguments)
	if err != nil {
		if errors.Is(err, tools.ErrToolNotFound) {
			return nil, jsonrpc.InvalidParams(err.Error())
		}
		if errors.Is(err, context.Canceled) {
			return nil, jsonrpc.Errorf(jsonrpc.CodeInternalError, "request cancelled")
		}
		// An unexpected tool failure is surfaced to the model, not the protocol.
		s.log.Error("tool call failed", "tool", p.Name, "err", err)
		return marshalResult(mcp.ErrorResult(err.Error()))
	}
	return marshalResult(result)
}

func marshalResult(v any) (json.RawMessage, *jsonrpc.Error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, jsonrpc.InternalError(err.Error())
	}
	return raw, nil
}

// nonNull substitutes an empty object for absent params so downstream
// unmarshalling of a struct always succeeds.
func nonNull(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage("{}")
	}
	return raw
}
