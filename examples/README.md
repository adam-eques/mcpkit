# Examples

## Go client

A minimal client that launches the server and drives a full session:

```bash
go run ./examples/client
```

It prints each request (`->`) and response (`<-`) so you can see the protocol
exchange end to end.

## Piping a recorded session

`session.jsonl` contains one JSON-RPC message per line. Feed it straight into the
stdio server:

```bash
go run ./cmd/mcpkit < examples/session.jsonl
```

Because the server answers requests concurrently, responses may arrive in a
different order than the requests were sent — match them by `id`, as any MCP
client does.

## Development config

`config.dev.json` enables the filesystem and fetch tools against a scratch
directory for local experimentation:

```bash
go run ./cmd/mcpkit -config examples/config.dev.json
```
