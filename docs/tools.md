# Tools

Every tool implements a two-method interface:

```go
type Handler interface {
    Definition() mcp.Tool
    Call(ctx context.Context, args json.RawMessage) (*mcp.CallToolResult, error)
}
```

`Definition` advertises a JSON Schema built with the `tools` schema helpers;
`Call` parses the arguments and returns content. A returned Go `error` is a
protocol failure, while an expected runtime failure returns a result with
`isError: true`.

## Built-in tools

| Tool | Package | Notes |
| --- | --- | --- |
| `calculate` | `tools/calc` | Recursive-descent expression evaluator |
| `http_fetch` | `tools/fetch` | Guarded HTTP client; blocks private addresses |
| `fs_read`, `fs_write`, `fs_list` | `tools/fs` | Sandboxed to a root directory |
| `shell_exec` | `tools/shell` | Allowlisted executables, no shell interpolation |
| `kv_set`, `kv_get`, `kv_list`, `kv_delete` | `tools/kv` | Persisted key/value store |
| `web_search` | `tools/websearch` | DuckDuckGo Instant Answer API |
| `rag_index`, `rag_search` | `tools/rag` | Feature-hashing embeddings + cosine search |
| `time_now`, `time_convert` | `tools/timeutil` | Time-zone aware |
| `hash`, `uuid`, `base64` | `tools/textutil` | Text and encoding utilities |
| `json_query` | `tools/jsonq` | Dotted-path extraction from JSON |

## Writing a new tool

1. Create a package under `tools/` with a type implementing `Handler`.
2. Build the input schema with `tools.Object`, `tools.Str`, `tools.Int`, etc.
3. Register it in `internal/app/build.go` behind a config flag.
4. Add a table-driven test; tools are pure functions of their arguments and are
   simple to cover.

The `rag` package is a good reference: it shows an embedding function, a
concurrency-safe store and two cooperating tools that share it.
