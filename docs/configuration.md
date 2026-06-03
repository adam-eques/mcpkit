# Configuration

mcpkit reads a default configuration, overlays an optional JSON file
(`-config path`), then applies `MCPKIT_*` environment overrides.

## Precedence

```
defaults  <  config file  <  environment variables
```

## Safe defaults

Out of the box the calculator, RAG, time, text, JSON-query and key/value tools
are enabled. The tools that touch the network, filesystem or shell are **off**
until you enable them explicitly:

- `fetch` — reaches the network; blocks private addresses unless `allowPrivate`.
- `fs` — reads and writes under a configured `root`, read-only by default.
- `shell` — runs only executables you list in `allowlist`.
- `webSearch` — calls an external API.

## Environment variables

| Variable | Effect |
| --- | --- |
| `MCPKIT_LOG_LEVEL` | `debug`, `info`, `warn`, `error` |
| `MCPKIT_LOG_FORMAT` | `text` or `json` |
| `MCPKIT_CONCURRENCY` | Max simultaneous requests |
| `MCPKIT_GATEWAY_ADDR` | Gateway listen address |
| `MCPKIT_FS_ROOT` | Enables the filesystem tools rooted here |
| `MCPKIT_FETCH` | `true`/`false` to toggle the fetch tool |
| `MCPKIT_WEBSEARCH` | `true`/`false` to toggle web search |

See `config.example.json` for a complete annotated file.
