# Security model

A tool server executes actions on behalf of a language model, so mcpkit treats
every tool input as untrusted.

## Network (`http_fetch`, `web_search`)

- Only `http` and `https` URLs are accepted.
- The fetch dialer inspects the **resolved** IP and refuses loopback, private,
  link-local, unique-local and unspecified addresses. Because the check runs at
  dial time it also defeats DNS-rebinding, which a pre-flight hostname lookup
  would miss.
- `allowPrivate` disables the guard and is intended only for trusted, isolated
  environments.
- Response bodies are size-limited.

## Filesystem (`fs_*`)

- Every path is joined to the configured root and rejected if the cleaned result
  escapes it, so `../` traversal and absolute paths are blocked.
- The sandbox is read-only unless writes are explicitly enabled.

## Shell (`shell_exec`)

- The executable and arguments are passed separately; nothing is interpreted by a
  shell, so there is no command-injection surface.
- The executable must appear on an explicit allowlist; an empty list disables the
  tool entirely.
- Every invocation runs under a timeout and its output is truncated.

## Reporting

Please report suspected vulnerabilities privately via the contact in
`SECURITY.md` rather than opening a public issue.
