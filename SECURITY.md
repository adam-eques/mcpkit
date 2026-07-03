# Security Policy

## Reporting a vulnerability

Please report security issues privately via Telegram
[@dracoeques](https://t.me/dracoeques) rather than opening a public issue. You
will get an acknowledgement within a few days and a fix or mitigation plan.

## Scope

mcpkit runs tools on behalf of a language model, so the threat model centres on
untrusted tool inputs. The network, filesystem and shell tools are disabled by
default and each has documented guardrails — see `docs/security.md`.

## Supported versions

The `main` branch receives security fixes. Tagged releases are patched on a
best-effort basis.
