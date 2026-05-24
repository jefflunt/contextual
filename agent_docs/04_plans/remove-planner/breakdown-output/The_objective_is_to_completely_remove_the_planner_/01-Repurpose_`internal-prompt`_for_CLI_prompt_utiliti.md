# Repurpose `internal/prompt` for CLI prompt utilities by moving `ConfirmOverwrite` and `PromptYesNo` from the planner, and deleting old plan prompt logic.

This task focuses on refactoring the `internal/prompt` package to serve as a general-purpose terminal utility package rather than holding planner-specific system prompts. The work involves two primary operations: 

First, extract the CLI confirmation utilities, specifically `ConfirmOverwrite` and `PromptYesNo`, from the `internal/planner` package (where they currently reside) and move them into `internal/prompt/prompt.go`. Ensure that any references to these utilities in other packages (like `cmd/contextual/main.go`) are updated to point to the `prompt` package.

Second, delete any old, planner-specific prompt templates or text-generation logic currently residing in `internal/prompt`. This ensures the `prompt` package is strictly dedicated to user terminal interactions, neatly decoupling it from the planner package which is scheduled for complete removal.
