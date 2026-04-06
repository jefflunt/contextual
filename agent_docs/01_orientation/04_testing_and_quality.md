# Testing and quality

## Table of contents
- [Where tests are](#where-tests-are)
- [How to run tests](#how-to-run-tests)
- [Testing style](#testing-style)
- [Linting](#linting)

## Where tests are

- `internal/config/config_test.go`
- `internal/spider/spider_test.go`

## How to run tests

```bash
script/test
```

Equivalent:

```bash
go test ./...
```

To build, test, and install in one step:

```bash
script/build-test-install
```

## Testing style

Follow existing patterns:
- Prefer table-driven tests where it improves coverage.
- For classification behavior, test `Spider.ParseItem(...)` directly.
- For traversal behavior, isolate side effects; avoid live network calls in unit tests.
- Config tests write a temp `config.yml` to a temp dir and call `config.Load()` directly.

## Linting

```bash
go vet ./...
```

There is no dedicated `script/lint`; run `go vet` directly.
