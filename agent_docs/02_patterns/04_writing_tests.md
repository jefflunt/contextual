# Pattern: Writing tests

## When to use this pattern

Use when adding parsing rules, new item types, or config handling.

## Canonical test targets

### Parsing/classification
- Test `(*Spider).ParseItem(arg)` in `internal/spider/spider_test.go`.
- Add cases for:
  - Jira key
  - Confluence ID
  - Jira URL (host-bound)
  - Confluence URL (host-bound)
  - generic web URL
  - invalid input

### Config
- Test config parsing in `internal/config/config_test.go`.

## Avoid

- Avoid live network calls in unit tests.
- If adding HTTP behavior, consider:
  - refactoring fetchers to accept an `http.Client` (if/when needed)
  - or using httptest servers in tests (if repo evolves that way)