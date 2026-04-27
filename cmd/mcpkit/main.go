// Command mcpkit runs the MCP server over the stdio transport. It reads
// JSON-RPC frames from stdin and writes responses to stdout; all logging goes to
// stderr so it never corrupts the protocol stream.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/adam-eques/mcpkit/internal/app"
	"github.com/adam-eques/mcpkit/internal/config"
	"github.com/adam-eques/mcpkit/internal/log"
	"github.com/adam-eques/mcpkit/internal/version"
	"github.com/adam-eques/mcpkit/transport"
)

func main() {
	configPath := flag.String("config", "", "path to a JSON config file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	version.FromBuildInfo()
	if *showVersion {
		fmt.Println("mcpkit", version.String())
		return
	}

	if err := run(*configPath); err != nil {
		fmt.Fprintln(os.Stderr, "mcpkit:", err)
		os.Exit(1)
	}
}

func run(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
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
	logger.Info("starting mcpkit over stdio",
		"version", version.String(),
		"tools", len(toolNames),
		"toolNames", toolNames)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	t := transport.NewStdio(os.Stdin, os.Stdout)
	return srv.Serve(ctx, t)
}
