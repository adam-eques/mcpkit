package server

import (
	"context"

	"github.com/adam-eques/mcpkit/internal/log"
	"github.com/adam-eques/mcpkit/internal/metrics"
	"github.com/adam-eques/mcpkit/mcp"
)

// ResourceProvider supplies resources for the resources/* methods. A server
// without a provider does not advertise the resources capability.
type ResourceProvider interface {
	ListResources(ctx context.Context) ([]mcp.Resource, error)
	ReadResource(ctx context.Context, uri string) ([]mcp.ResourceContents, error)
}

// PromptProvider supplies prompts for the prompts/* methods.
type PromptProvider interface {
	ListPrompts(ctx context.Context) ([]mcp.Prompt, error)
	GetPrompt(ctx context.Context, name string, args map[string]string) (*mcp.GetPromptResult, error)
}

// Option customises a Server.
type Option func(*Server)

// WithInstructions sets the free-text instructions returned during initialize.
func WithInstructions(s string) Option {
	return func(srv *Server) { srv.instructions = s }
}

// WithLogger sets the structured logger. It must not write to stdout when the
// server runs over the stdio transport.
func WithLogger(l *log.Logger) Option {
	return func(srv *Server) {
		if l != nil {
			srv.log = l
		}
	}
}

// WithMetrics attaches a metrics registry.
func WithMetrics(m *metrics.Registry) Option {
	return func(srv *Server) {
		if m != nil {
			srv.metrics = m
		}
	}
}

// WithResources enables the resources capability backed by p.
func WithResources(p ResourceProvider) Option {
	return func(srv *Server) { srv.resources = p }
}

// WithPrompts enables the prompts capability backed by p.
func WithPrompts(p PromptProvider) Option {
	return func(srv *Server) { srv.prompts = p }
}

// WithConcurrency caps the number of requests handled simultaneously. A value of
// zero or less means unlimited.
func WithConcurrency(n int) Option {
	return func(srv *Server) {
		if n > 0 {
			srv.sem = make(chan struct{}, n)
		}
	}
}
