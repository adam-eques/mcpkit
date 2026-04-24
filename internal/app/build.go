// Package app wires a configuration into a fully populated MCP server. It is the
// single place that knows about every tool, keeping the command binaries thin.
package app

import (
	"fmt"
	"time"

	"github.com/adam-eques/mcpkit/internal/config"
	"github.com/adam-eques/mcpkit/internal/log"
	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/server"
	"github.com/adam-eques/mcpkit/tools"
	"github.com/adam-eques/mcpkit/tools/calc"
	"github.com/adam-eques/mcpkit/tools/fetch"
	"github.com/adam-eques/mcpkit/tools/fs"
	"github.com/adam-eques/mcpkit/tools/jsonq"
	"github.com/adam-eques/mcpkit/tools/kv"
	"github.com/adam-eques/mcpkit/tools/rag"
	"github.com/adam-eques/mcpkit/tools/shell"
	"github.com/adam-eques/mcpkit/tools/textutil"
	"github.com/adam-eques/mcpkit/tools/timeutil"
	"github.com/adam-eques/mcpkit/tools/websearch"
)

// Build constructs a registry and server from cfg. It returns the tool names
// that were enabled so the caller can report them.
func Build(cfg config.Config, logger *log.Logger) (*server.Server, []string, error) {
	reg := tools.NewRegistry()

	if cfg.Tools.Calc {
		reg.MustRegister(calc.New())
	}
	if cfg.Tools.RAG {
		reg.MustRegister(rag.Handlers(rag.NewStore())...)
	}
	if cfg.Tools.Time {
		reg.MustRegister(timeutil.Handlers(nil)...)
	}
	if cfg.Tools.Text {
		reg.MustRegister(textutil.Handlers()...)
	}
	if cfg.Tools.JSONQuery {
		reg.MustRegister(jsonq.New())
	}
	if cfg.Tools.KV.Enabled {
		store := kv.NewStore()
		if cfg.Tools.KV.Path != "" {
			opened, err := kv.Open(cfg.Tools.KV.Path)
			if err != nil {
				return nil, nil, fmt.Errorf("open kv store: %w", err)
			}
			store = opened
		}
		reg.MustRegister(kv.Handlers(store)...)
	}
	if cfg.Tools.Fetch.Enabled {
		reg.MustRegister(fetch.New(
			fetch.WithAllowPrivate(cfg.Tools.Fetch.AllowPrivate),
			fetch.WithMaxBytes(cfg.Tools.Fetch.MaxBytes),
		))
	}
	if cfg.Tools.WebSearch.Enabled {
		opts := []websearch.Option{}
		if cfg.Tools.WebSearch.Endpoint != "" {
			opts = append(opts, websearch.WithEndpoint(cfg.Tools.WebSearch.Endpoint))
		}
		reg.MustRegister(websearch.New(opts...))
	}
	if cfg.Tools.FS.Enabled {
		if cfg.Tools.FS.Root == "" {
			return nil, nil, fmt.Errorf("filesystem tool enabled but no root configured")
		}
		sb, err := fs.New(cfg.Tools.FS.Root, cfg.Tools.FS.ReadOnly)
		if err != nil {
			return nil, nil, fmt.Errorf("filesystem sandbox: %w", err)
		}
		reg.MustRegister(fs.Handlers(sb)...)
	}
	if cfg.Tools.Shell.Enabled {
		timeout := time.Duration(cfg.Tools.Shell.TimeoutSeconds) * time.Second
		reg.MustRegister(shell.New(cfg.Tools.Shell.Allowlist, timeout))
	}

	names := make([]string, 0, reg.Len())
	for _, tl := range reg.List() {
		names = append(names, tl.Name)
	}

	srv := server.New(
		mcp.Implementation{Name: cfg.Server.Name, Version: cfg.Server.Version},
		reg,
		server.WithInstructions(cfg.Server.Instructions),
		server.WithLogger(logger),
		server.WithConcurrency(cfg.Server.Concurrency),
	)
	return srv, names, nil
}
