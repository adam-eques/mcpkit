# Protocol support

mcpkit implements the `2025-06-18` revision of the Model Context Protocol and
negotiates down to `2025-03-26` or `2024-11-05` when a client requests them.

## Lifecycle

1. The client sends `initialize` with its protocol version and capabilities.
2. The server replies with the negotiated version, its capabilities and its
   identity. Any method other than `initialize` or `ping` before this step is
   rejected with `-32600 invalid request`.
3. The client sends the `notifications/initialized` notification.
4. Normal request/response traffic proceeds.

## Methods

| Method | Description |
| --- | --- |
| `initialize` | Handshake and capability negotiation |
| `ping` | Liveness check, always available |
| `tools/list` | Enumerate registered tools |
| `tools/call` | Invoke a tool by name |
| `resources/list`, `resources/read` | Present when a resource provider is set |
| `prompts/list`, `prompts/get` | Present when a prompt provider is set |
| `logging/setLevel` | Adjust the server log verbosity |

## Notifications

| Notification | Effect |
| --- | --- |
| `notifications/initialized` | Marks the session ready |
| `notifications/cancelled` | Cancels the referenced in-flight request |

Unknown notifications are ignored, and unknown methods return `-32601 method not
found`, per JSON-RPC 2.0.

## Errors

Standard JSON-RPC codes are used throughout: `-32700` parse error, `-32600`
invalid request, `-32601` method not found, `-32602` invalid params and `-32603`
internal error.
