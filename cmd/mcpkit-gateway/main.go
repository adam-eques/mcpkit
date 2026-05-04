// Command mcpkit-gateway exposes the MCP server over HTTP instead of stdio,
// which is convenient for local testing, container health checks and clients
// that speak plain JSON-RPC over HTTP.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adam-eques/mcpkit/internal/app"
	"github.com/adam-eques/mcpkit/internal/config"
	"github.com/adam-eques/mcpkit/internal/gateway"
	"github.com/adam-eques/mcpkit/internal/log"
	"github.com/adam-eques/mcpkit/internal/version"
)

func main() {
	configPath := flag.String("config", "", "path to a JSON config file")
	addr := flag.String("addr", "", "listen address (overrides config)")
	flag.Parse()

	if err := run(*configPath, *addr); err != nil {
		fmt.Fprintln(os.Stderr, "mcpkit-gateway:", err)
		os.Exit(1)
	}
}

func run(configPath, addr string) error {
	version.FromBuildInfo()
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if addr != "" {
		cfg.Gateway.Address = addr
	}

	logger := log.New(log.Options{
		Level:  log.ParseLevel(cfg.Log.Level),
		Format: log.Format(cfg.Log.Format),
		Writer: os.Stderr,
	})

	srv, toolNames, err := app.Build(cfg, logger)
	if err != nil {
		return err
	}

	httpSrv := &http.Server{
		Addr:              cfg.Gateway.Address,
		Handler:           gateway.Handler(srv, logger),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}()

	logger.Info("gateway listening",
		"address", cfg.Gateway.Address,
		"version", version.String(),
		"tools", len(toolNames))

	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
