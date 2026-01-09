// Package server implements a Model Context Protocol server: it parses inbound
// JSON-RPC frames, dispatches the initialize handshake and the tools, resources
// and prompts methods, and streams responses back over a transport.
package server

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/adam-eques/mcpkit/internal/log"
	"github.com/adam-eques/mcpkit/internal/metrics"
	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// Server handles MCP requests for a single peer connection. A Server may serve
// many sequential connections but tracks one session's state at a time.
type Server struct {
	info         mcp.Implementation
	instructions string
	tools        *tools.Registry
	resources    ResourceProvider
	prompts      PromptProvider
	log          *log.Logger
	metrics      *metrics.Registry
	sem          chan struct{}

	initialized atomic.Bool
	logLevel    atomic.Value // mcp.LoggingLevel

	mu       sync.Mutex
	clientIn mcp.Implementation
	inflight map[string]context.CancelFunc
}

// New builds a Server that advertises info and dispatches tool calls to reg.
func New(info mcp.Implementation, reg *tools.Registry, opts ...Option) *Server {
	if reg == nil {
		reg = tools.NewRegistry()
	}
	s := &Server{
		info:     info,
		tools:    reg,
		log:      log.Discard(),
		metrics:  metrics.New(),
		inflight: make(map[string]context.CancelFunc),
	}
	s.logLevel.Store(mcp.LogInfo)
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Tools returns the registry backing this server.
func (s *Server) Tools() *tools.Registry { return s.tools }

// Metrics returns the server's metrics registry.
func (s *Server) Metrics() *metrics.Registry { return s.metrics }

// capabilities reports which primitives the server exposes given its providers.
func (s *Server) capabilities() mcp.ServerCapabilities {
	caps := mcp.ServerCapabilities{Logging: &struct{}{}}
	if s.tools != nil {
		caps.Tools = &mcp.ToolsCapability{}
	}
	if s.resources != nil {
		caps.Resources = &mcp.ResourcesCapability{}
	}
	if s.prompts != nil {
		caps.Prompts = &mcp.PromptsCapability{}
	}
	return caps
}

func (s *Server) registerInflight(id string, cancel context.CancelFunc) {
	s.mu.Lock()
	s.inflight[id] = cancel
	s.mu.Unlock()
}

func (s *Server) clearInflight(id string) {
	s.mu.Lock()
	if cancel, ok := s.inflight[id]; ok {
		delete(s.inflight, id)
		cancel()
	}
	s.mu.Unlock()
}

// cancelInflight cancels a request the peer asked to abort.
func (s *Server) cancelInflight(id string) {
	s.mu.Lock()
	cancel, ok := s.inflight[id]
	if ok {
		delete(s.inflight, id)
	}
	s.mu.Unlock()
	if ok {
		cancel()
	}
}

func (s *Server) acquire() {
	if s.sem != nil {
		s.sem <- struct{}{}
	}
}

func (s *Server) release() {
	if s.sem != nil {
		<-s.sem
	}
}
