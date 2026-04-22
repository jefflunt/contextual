# Architecture overview

## Table of contents
- [Repository shape](#repository-shape)
- [Primary execution flow](#primary-execution-flow)
- [Core packages](#core-packages)
- [Key conventions](#key-conventions)

## Repository shape

```
cmd/contextual/main.go                     — CLI entrypoint; flag parsing; config + logger construction; writes context.md; invokes planner
internal/spider/spider.go                  — argument classification + BFS traversal; orchestrates fetchers
internal/fetcher/jira.go                   — Jira fetch + ADF→Markdown conversion
internal/fetcher/confluence.go             — Confluence fetch + XHTML→Markdown conversion
internal/fetcher/web.go                    — generic HTTP fetch
internal/config/config.go                  — config loading from ~/.contextual/config.yml
internal/types/types.go                    — shared domain types (Item, ItemType)
internal/logger/logger.go                  — Silent/Progress/Verbose logger; always writes to ~/.contextual/log.log
internal/planner/planner.go               — ResolveOutputDir, ConfirmOverwrite, ItemSlug
internal/planner/copilot.go               — RunPlanner: temp file + <promptFile> substitution + sh -c execution
internal/prompt/prompt.go                  — BuildPlanPrompt; embeds plan spec via go:embed
internal/prompt/writing-plan-files.md      — embedded plan file spec
script/build                               — build binary to bin/contextual
script/test                                — run all tests
script/install                             — install binary to $GOPATH/bin
script/build-test-install                  — run all three in sequence
contextual.config.example.yml             — annotated config example for users
```

## Primary execution flow

### Fetch mode (`contextual [--verbose|-v] [--progress|-p] <item> [<item> ...]`)

```
CLI args
  |
  v
config.Load()              (internal/config)  <- ~/.contextual/config.yml
  |
  v
logger.New(mode)           (internal/logger)  <- --verbose / --progress / silent
  |
  v
spider.New(cfg, lg)        (internal/spider)  <- credentials from cfg.Atlassian.*
  |
  v
Spider.Run(args)
  |
  +--> ParseItem(arg) classifies inputs
  |
  +--> Priority-driven BFS traversal
         |
         +--> fetcher.FetchJira(host, email, token, id)       -> Item + links (parent/subtasks/confluence/web)
         +--> fetcher.FetchConfluence(host, email, token, id) -> Item + links (children/jira/web)
         +--> fetcher.FetchWeb(url)                           -> Item only (no further spidering)
  |
  v
ConfirmOverwrite(contextPath)   <- interactive prompt if context.md exists
  |
  v
Write context.md to cwd         <- AI-readable preamble + item blocks (truncated at MaxContextLength)
  |
  v
print contextPath to stdout
```

### Plan mode (`contextual plan [--verbose|-v] [--progress|-p] <item> [<item> ...]`)

Same as fetch mode through writing context.md, then:

```
  |
  v
planner.ResolveOutputDir(items[0])
  ├── agent_docs/plans/<slug>/   if agent_docs/plans/ exists
  ├── prompt to create it        if only agent_docs/ exists
  └── ./<slug>/                  fallback
  |
  v
Write context.md to outputDir
  |
  v
prompt.BuildPlanPrompt(contextPath, primaryItem)
  |
  v
planner.RunPlanner(cfg.Planner, promptText, outputDir, lg)
  <- writes prompt to temp file
  <- replaces <promptFile> in cfg.Planner with temp file path
  <- executes full command string via sh -c
  |
  v
print plan.md path to stdout
```

## Core packages

### `cmd/contextual`
- Owns process concerns: flag parsing, config loading, logger construction, exit codes, stdout printing.
- Flags: `--verbose`/`-v`, `--progress`/`-p`
- Subcommand: `plan` (detected before flag parsing via `args[0] == "plan"`)

### `internal/spider`
- Owns classification and traversal orchestration.
- `New(cfg *config.Config, log *logger.Logger) *Spider`
- `ParseItem(arg string) (*types.Item, error)`
- `Run(args []string) ([]types.Item, error)`
- Reads credentials from `cfg.Atlassian.{Host, APIUser, APIToken, MaxSpiderJumps}` — **no env vars**.

### `internal/fetcher`
- Owns I/O for each source.
- Emits **Markdown** content (not plaintext) from both ADF (Jira) and XHTML storage (Confluence).
- Returns structured results (Item + extracted references for enqueueing).

### `internal/config`
- `Config.Atlassian.{Host, APIUser, APIToken, MaxSpiderJumps}` — Atlassian credentials and spidering depth control.
- `Config.Planner` — shell command template; must contain `<promptFile>` placeholder.
- CLI tolerates config load failure (warn + continue with empty config).

### `internal/logger`
- Three modes: `ModeSilent`, `ModeProgress`, `ModeVerbose`.
- Always writes structured entries to `~/.contextual/log.log` (ISO 8601 UTC).
- Methods: `Info`, `Warn`, `Error`.

### `internal/planner`
- `ResolveOutputDir(item)` — three-tier output directory logic.
- `ConfirmOverwrite(path)` — interactive prompt if file already exists.
- `ItemSlug(item)` — slug derived from item ID or URL.
- `RunPlanner(plannerCmd, promptText, outputDir, lg)` — executes configured planner via shell.

### `internal/prompt`
- `BuildPlanPrompt(contextPath, primaryItem)` — constructs the plan prompt string.
- Plan file spec embedded via `//go:embed writing-plan-files.md`.

### `internal/types`
- `Item`, `ItemType` — shared domain types used across all packages.

## Key conventions

- **stdout is data** (output file path only), stderr is logs.
- On per-item failures: log `[ERROR]` and continue traversal — do not abort the run.
- All credentials come from `~/.contextual/config.yml`, not environment variables.
- Scripts in `script/` are the canonical build/test/install interface (no Makefile).
- `RunPlanner` executes via `sh -c` so the full command string (quoting, backticks, etc.) is shell-interpreted.
