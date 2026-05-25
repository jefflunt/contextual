# Design: Completely Remove Planner from Contextual

## User Story
- **Headline**: Remove the planner subcommand and config from contextual
- **Problem Statement**: The `contextual` tool currently supports a `plan` mode that runs an external planner command with a generated prompt. We now have a separate, dedicated tool to handle planning, and having this logic in `contextual` adds unnecessary complexity, maintenance overhead, and bloat.
- **Objective**: Completely remove the `plan` subcommand, the `planner` command execution logic, and the planning configuration settings, while keeping the interactive terminal prompt helpers.
- **Expected Outcome**:
  - Running `contextual plan ...` will result in a standard flag parsing / usage error (no longer a recognized subcommand).
  - The `planner` package is completely deleted.
  - The `planner` field is removed from the YAML `Config` struct.
  - Interactive terminal prompts (`PromptYesNo` and `ConfirmOverwrite`) are preserved and moved to the `internal/prompt` package.
  - All old planning prompt builder logic (`BuildPlanPrompt` and `writing-plan-files.md`) in `internal/prompt` is deleted.
  - Tests pass, and documentation is updated to reflect the removal of the planner.

## Architecture Overview
This refactoring removes a significant portion of the system related to plan-generation, including the subprocess command runner. No database changes or schema changes are involved.

The core changes are:
1. **CLI Commands (`cmd/contextual/main.go`)**:
   - Remove `plan` subcommand keyword handling.
   - Remove `planMode` boolean, conditional directory resolution (remove `planner.ResolveOutputDir`), and subprocess runner invocation (`planner.RunPlanner`).
   - Standard output path is always the current working directory (`os.Getwd()`), and the output file is always written to `./context.md` (unless specified otherwise via some other means, but currently it's hardcoded to `./context.md` in non-plan mode).
   - Update help menu output to remove `contextual plan`.
2. **Interactive Helpers (`internal/prompt`)**:
   - Delete `BuildPlanPrompt` and `writing-plan-files.md`.
   - Move `ConfirmOverwrite` and `PromptYesNo` from `internal/planner/planner.go` into `internal/prompt/prompt.go` as package `prompt`.
   - Update imports in `main.go` from `github.com/jefflunt/contextual/internal/planner` to `github.com/jefflunt/contextual/internal/prompt`.
3. **Configuration (`internal/config/config.go`)**:
   - Remove `Planner` field from `Config` struct.
   - Update `contextual.config.example.yml` to remove the `planner` keys and comments.
4. **Clean up Files**:
   - Delete entire `internal/planner/` folder.
   - Delete `internal/prompt/writing-plan-files.md`.

## Implementation Backlog

### Pending
- *None*

### Current
- *None*

### Completed
- **Task 1: Repurpose `internal/prompt` for CLI prompt utilities**
  - Delete `internal/prompt/writing-plan-files.md`.
  - Rewrite `internal/prompt/prompt.go` to house `PromptYesNo` and `ConfirmOverwrite` from `internal/planner/planner.go`.
  - Write unit tests for `internal/prompt` if applicable (e.g. mocking/redirecting stdin/stderr).
- **Task 2: Update `cmd/contextual/main.go` to remove plan mode**
  - Remove `plan` subcommand detection, `planMode` flags/branches.
  - Remove all references to `planner.ResolveOutputDir` and `planner.RunPlanner`.
  - Update imports from `internal/planner` to `internal/prompt`.
  - Update interactive prompt calls in `main.go` to use `prompt.PromptYesNo` and `prompt.ConfirmOverwrite`.
  - Update usage/help output to remove references to `plan` subcommand.
- **Task 3: Clean up configuration**
  - Remove `Planner` field from the `Config` struct in `internal/config/config.go`.
  - Remove all planner-related configurations and descriptions from `contextual.config.example.yml`.
- **Task 4: Delete the planner package**
  - Completely delete `internal/planner/` directory.
- **Task 5: Update project documentation**
  - Update architectural documentation in `agent_docs/01_orientation/01_architecture_overview.md`, `03_runtime_and_configuration.md`, `agent_docs/02_patterns/03_logging_and_error_handling.md`, etc. to remove references to the `planner` package, plan subcommand, and config.
- **Verification: Run all tests and verify build**
  - Compile the binary via `./script/build`.
  - Run all tests via `./script/test` (including prompt unit tests).
  - Install binary via `./script/install`.

## Checklist & TDD Requirements
- All tests in the test suite must pass (`go test ./...`).
- We must provide a test for the repurposed `prompt` package to verify `PromptYesNo` and `ConfirmOverwrite`.
- Verify compilation works successfully by building the binary.
