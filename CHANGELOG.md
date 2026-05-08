# Changelog

All notable changes to mcpkit are documented here.

## 0.0.1 — Scaffolding

- Initialize the Go module and repository tooling.
- Add the MIT license, Makefile and editor configuration.

## 0.0.2 — JSON-RPC 2.0

- Implement JSON-RPC 2.0 requests, notifications, responses and errors.
- Preserve string and integer identifiers on the wire.

## 0.0.3 — MCP types

- Add the initialize handshake, capabilities and content blocks.
- Add tools, resources, prompts and notification parameter types.

## 0.0.4 — Transports

- Add a context-aware, newline-delimited stdio transport.
- Add an in-memory pipe and a Server-Sent Events writer.

## 0.0.5 — Observability

- Add a structured logger that never writes to stdout.
- Add a dependency-free metrics registry and build stamping.

## 0.0.6 — Tool framework

- Add the Handler interface, a schema builder and a thread-safe registry.

## 0.1.0-rc1 — Server core

- Dispatch the full method set with an initialize-first guard.
- Run each request on its own goroutine with per-request cancellation.

## 0.1.0-rc2 — Core tools

- Add the calculator, hash/uuid/base64, json_query and time tools.

## 0.1.0-rc3 — RAG and state

- Add an in-process RAG toolset with a pure-Go embedder.
- Add a persistent, concurrency-safe key/value store.

## 0.1.0-rc4 — Network and system tools

- Add http_fetch with an SSRF guard and web_search.
- Add a sandboxed filesystem toolset and an allowlisted shell tool.

## 0.1.0-rc5 — Configuration and CLI

- Add layered configuration and the tool wiring layer.
- Add the stdio server and HTTP gateway binaries.
