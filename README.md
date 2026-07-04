# mcpkit

A production-minded **Model Context Protocol (MCP) server written in Go with zero
third-party dependencies** — the JSON-RPC 2.0 core, the MCP protocol layer, the
transports and a batteries-included toolset are all built on the standard
library. It compiles and tests anywhere, with no supply chain to audit.

[![CI](https://github.com/adam-eques/mcpkit/actions/workflows/ci.yml/badge.svg)](https://github.com/adam-eques/mcpkit/actions/workflows/ci.yml)
![Go](https://img.shields.io/badge/go-1.26-00ADD8)
![Dependencies](https://img.shields.io/badge/dependencies-none-brightgreen)
![License](https://img.shields.io/badge/license-MIT-blue)

## Why

There is an official MCP Go SDK, and for most production servers it's the right
choice — it tracks the spec and is battle-tested across many clients. mcpkit is a
deliberate alternative: it implements the protocol from the wire up — the
initialize handshake, capability negotiation, tools/resources/prompts,
cancellation and progress — on the standard library alone. That buys two things
the SDK route can't:

- **A readable reference** for how MCP actually works, with nothing hidden behind
  a dependency. The demux, the session state and the cancellation plumbing are
  all here to read.
- **A zero-dependency, auditable base** for security-sensitive tooling. This
  server ships an SSRF guard, a sandboxed filesystem and an allowlisted shell —
  and there is no third-party supply chain to vet, patch or keep in sync.

Reach for the official SDK when you want to ship and forget; reach for mcpkit when
you want to understand the protocol end to end or extend a base you fully control.

## Highlights

- **Full protocol** — MCP `2025-06-18` with negotiation down to `2024-11-05`.
- **Two transports** — newline-delimited **stdio** (the default) and an **HTTP
  gateway** with `/rpc`, `/healthz` and `/metrics`.
- **Concurrent dispatch** — each request runs on its own goroutine under a
  configurable limit, with per-request `context` cancellation driven by
  `notifications/cancelled`.
- **A real toolset** — including an in-process **RAG** tool (feature-hashing
  embeddings + cosine search), a guarded HTTP fetcher with an **SSRF guard**, a
  **sandboxed** filesystem, an allowlisted shell, a persistent key/value store,
  web search, a recursive-descent **calculator**, JSON path query and crypto
  utilities.
- **Observability** — structured logging (stderr, never stdout) and built-in
  metrics.
- **Tested** — table-driven unit tests, HTTP tests, an in-memory transport for
  integration tests, and benchmarks. No dependency downloads required to run them.

## Quick start

```bash
# Build both binaries into ./bin
make build

# Run the stdio server and drive it with a recorded session
go run ./cmd/mcpkit < examples/session.jsonl

# Or watch a full client/server exchange
go run ./examples/client
```

Minimal handshake:

```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","clientInfo":{"name":"demo","version":"1.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"calculate","arguments":{"expression":"2 ^ 10 + sqrt(81)"}}}
```

## HTTP gateway

```bash
go run ./cmd/mcpkit-gateway -addr :8080
curl -s localhost:8080/rpc -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}'
curl -s localhost:8080/metrics
```

## Tools

| Tool | What it does |
| --- | --- |
| `calculate` | Evaluate a math expression (custom parser) |
| `rag_index` / `rag_search` | Index passages and retrieve by semantic similarity |
| `http_fetch` | Fetch a URL with an SSRF guard and size limits |
| `fs_read` / `fs_write` / `fs_list` | Sandboxed filesystem access |
| `shell_exec` | Run an allowlisted executable, no shell interpolation |
| `kv_*` | Persistent key/value store |
| `web_search` | DuckDuckGo Instant Answer search |
| `time_now` / `time_convert` | Time-zone aware clock |
| `hash` / `uuid` / `base64` | Text and encoding utilities |
| `json_query` | Extract a value from JSON by dotted path |

Network, filesystem and shell tools are **disabled by default**; enable them in
config. See [`docs/configuration.md`](docs/configuration.md).

## Project layout

```
cmd/            stdio server and HTTP gateway binaries
mcp/            MCP protocol types
jsonrpc/        JSON-RPC 2.0 core
transport/      stdio, in-memory pipe, SSE
server/         dispatch, session, cancellation
tools/          the Handler interface, registry and every tool
internal/       config, logging, metrics, wiring
docs/           architecture, protocol, tools, configuration, security
examples/       a Go client and a recorded session
```

## Documentation

- [Architecture](docs/architecture.md)
- [Protocol support](docs/protocol.md)
- [Tools](docs/tools.md)
- [Configuration](docs/configuration.md)
- [Security model](docs/security.md)

## License

MIT — see [LICENSE](LICENSE).
