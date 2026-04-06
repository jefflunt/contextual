# Pattern: Logging and error handling

## When to use this pattern

Use when adding new functionality that can fail per-item (network errors, parse errors, missing config).

## Logger package

All logging goes through `internal/logger.Logger` — **do not use `log.Printf` or `fmt.Fprintf(os.Stderr, ...)` directly** in spider or fetcher code.

The logger is constructed in `cmd/contextual/main.go` and passed down to `spider.New(cfg, lg)` and `planner.RunPlanner(...)`.

### Modes (set via CLI flags)

| Mode | Flag | stderr behavior |
|------|------|----------------|
| `ModeSilent` | (default) | nothing printed |
| `ModeProgress` | `--progress` / `-p` | prints `loading context ` then `.` per success, `X` per error |
| `ModeVerbose` | `--verbose` / `-v` | prints each `[INFO]`/`[WARN]`/`[ERROR]` line |

### File logging

Regardless of mode, **all log entries are always written** to `~/.contextual/log.log` with ISO 8601 UTC timestamps:

```
2026-04-04T22:07:39Z [INFO] Fetching jira: IPM-1234
2026-04-04T22:07:40Z [ERROR] Planner exited with code 1: copilot -p ...
```

### Methods

- `lg.Info(format, args...)` — fetch progress, invocation details
- `lg.Warn(format, args...)` — configuration issues, non-fatal setup problems
- `lg.Error(format, args...)` — fetch failures, planner failures

## Strategy: skip vs fail

Prefer "skip item and continue" for:
- unrecognized CLI args
- missing config needed for a specific fetch type
- fetch errors for a single item

Fail the whole run (exit non-zero) only for:
- missing or misconfigured `planner` key when in plan mode
- planner process exiting non-zero
- truly fatal setup errors (logger init failure, etc.)

## Where to implement

- Parse errors: in `internal/spider.Run` seed loop — call `s.logError(...)`, `continue`.
- Fetch errors: in each `case` branch of `Run(...)` — call `s.logError(...)`, `continue`.
- Planner errors: in `internal/planner/copilot.go` — call `lg.Error(...)`, return error to `main.go`.
- Config/setup warnings: in `cmd/contextual/main.go` — call `lg.Warn(...)`, continue.
