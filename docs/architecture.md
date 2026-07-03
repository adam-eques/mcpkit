# Architecture

mcpkit is a layered, dependency-free implementation of the Model Context
Protocol. Each layer depends only on the ones beneath it, which keeps the pieces
independently testable.

```
            ┌───────────────────────────────────────────────┐
 cmd/       │  mcpkit (stdio)          mcpkit-gateway (HTTP) │
            └───────────────┬───────────────────┬───────────┘
                            │                   │
 internal/app  ── Build(config) ──► wires tools into a server
                            │
            ┌───────────────▼───────────────────────────────┐
 server/    │  Server: dispatch, session, cancellation       │
            └───────┬───────────────────────┬───────────────┘
                    │                       │
 tools/    Registry + Handlers      transport/  Stdio · Pipe · SSE
                    │                       │
 mcp/       protocol types          jsonrpc/  JSON-RPC 2.0 core
```

## Layers

- **`jsonrpc`** — JSON-RPC 2.0 messages, identifiers and error objects. Knows
  nothing about MCP.
- **`mcp`** — the protocol vocabulary: initialize handshake, capabilities,
  content blocks, and the tools/resources/prompts request and result types.
- **`transport`** — moves opaque JSON-RPC frames. The stdio transport is
  newline-delimited; `Pipe` connects two in-memory endpoints for tests; the SSE
  helper frames server-to-client events for the gateway.
- **`tools`** — the `Handler` interface, a JSON-Schema builder and a
  concurrency-safe registry. Individual tools live in subpackages and depend only
  on `mcp` and `tools`.
- **`server`** — parses frames, enforces the initialize-first rule, dispatches
  each method, and manages per-request cancellation and a concurrency limit.
- **`internal/app`** — the single wiring point that turns a `config.Config` into
  a populated `server.Server`.
- **`cmd`** — two thin binaries: a stdio server and an HTTP gateway.

## Concurrency model

`Server.Serve` reads frames sequentially but dispatches each request on its own
goroutine, bounded by an optional semaphore (`WithConcurrency`). Writes are
serialised by the transport. Every request runs under a `context.Context` derived
from the connection; a `notifications/cancelled` from the peer cancels the
matching in-flight context via a registry keyed by request id.

## Design choices

- **No third-party dependencies.** Everything is built on the Go standard
  library, so the module builds and tests anywhere with no supply chain to audit.
- **Protocol errors vs tool errors.** A malformed request yields a JSON-RPC
  error; a tool that fails at runtime returns a result with `isError: true`, which
  is what a model should see and reason about.
- **Stdout is sacred.** In stdio mode the protocol owns stdout, so all logging is
  routed to stderr through the `internal/log` package.
