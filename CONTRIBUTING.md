# Contributing

Thanks for your interest in mcpkit.

## Development

```bash
make test    # run the suite
make race    # run with the race detector
make vet     # static checks
make bench   # benchmarks
```

The module has **no third-party dependencies** and that is a deliberate
constraint — please keep it standard-library-only.

## Branching

Work happens on `dev`. Each change is a focused commit; features are merged into
`main` at milestone boundaries. Please keep commits small and their messages in
the imperative mood ("add", "fix", "refactor").

## Adding a tool

See `docs/tools.md`. In short: implement the `tools.Handler` interface in a new
subpackage, wire it into `internal/app/build.go` behind a config flag, and add a
table-driven test.

## Checklist before opening a PR

- [ ] `make vet test` passes
- [ ] New behaviour is covered by tests
- [ ] Exported identifiers are documented
- [ ] No new dependencies
