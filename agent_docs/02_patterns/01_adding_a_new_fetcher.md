# Pattern: Adding a new fetcher

## When to use this pattern

Use when adding a new external content source (e.g., GitHub issue, Slack thread export, Google Doc export) that should produce a `types.Item`.

## Canonical approach

1. Add a new fetcher file in `internal/fetcher/`:
   - e.g. `internal/fetcher/github.go`
2. The fetcher should:
   - accept minimal inputs (host/creds/id/url)
   - return an `Item`
   - optionally return extracted references for expansion (if applicable)

3. Wire it into traversal:
   - update `internal/spider/spider.go` switch in `Run(...)` to call the new fetcher when appropriate
   - decide whether discovered references should be enqueued

## Do / Don’t

### Do
- Keep fetchers focused on I/O + parsing into domain outputs.
- Keep traversal policy (what gets enqueued) in `internal/spider`.

### Don’t
- Don’t print to stdout/stderr inside fetchers (leave printing/logging to CLI/spider).
- Don’t introduce uncontrolled spidering (web fetcher is intentionally non-spidering).

## Files to reference

- `internal/fetcher/jira.go`
- `internal/fetcher/confluence.go`
- `internal/fetcher/web.go`
- `internal/spider/spider.go`