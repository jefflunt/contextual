# agent_docs/ (AI agent knowledge base)

This folder is a **living knowledge base written by and for AI coding agents**. Its purpose is to let a future agent ramp on this repo quickly enough to implement full-ticket changes with minimal exploration.

These docs are **agent-first**: dense, concrete, and optimized for correctness + navigation, not narrative.

## Maintenance model

- Agents should update these docs when they find drift (new files, changed flows, new conventions).
- These docs do not need to be perfectly human-friendly; they should be **actionable**.

## Progressive disclosure

Read as little as you need:

1. Start with `01_orientation/` for repo-wide mental model
2. Use `02_patterns/` when implementing changes
3. Use `03_deep_dives/` only when changing internals
4. `plans/` is for short-lived alignment notes, not long-term reference

## plans/

`plans/` contains short-lived planning docs created during engineer+agent collaboration (e.g., "approach options", "decision log", "migration plan"). Treat them as disposable.

## Flat file listing

### Orientation (`01_orientation/`)
- `01_orientation/01_architecture_overview.md` — repo layout + primary control/data flows for both fetch and plan modes; where to change what.
- `01_orientation/02_domain_model.md` — core domain concepts: Item, ItemType, Spider traversal outputs.
- `01_orientation/03_runtime_and_configuration.md` — config file structure (no env vars); required settings for Atlassian and plan mode.
- `01_orientation/04_testing_and_quality.md` — where tests live and how to run them (`script/test`, `script/build-test-install`).

### Patterns (`02_patterns/`)
- `02_patterns/01_adding_a_new_fetcher.md` — how to add a new fetcher module and wire it into traversal.
- `02_patterns/02_adding_a_new_item_type.md` — how to extend `types.ItemType` and ensure CLI output remains consistent.
- `02_patterns/03_logging_and_error_handling.md` — logger package usage, modes, file logging, skip-vs-fail decisions.
- `02_patterns/04_writing_tests.md` — how to write tests similar to existing `internal/config` and `internal/spider` tests.
- `02_patterns/05_parsing_and_classification_rules.md` — how CLI args are classified; safe ways to extend parsing.
- `02_patterns/06_output_contract_and_streams.md` — stdout contract (one path); context.md format; preamble; overwrite protection.

### Deep dives (`03_deep_dives/`)
- `03_deep_dives/01_spider_queue_and_traversal.md` — BFS queue, visited key strategy, and expansion rules by item type.
- `03_deep_dives/02_atlassian_fetchers.md` — Jira/Confluence fetch behavior; credentials from config (not env); Markdown output; returned link extraction.
- `03_deep_dives/03_web_fetcher_and_non_spidering.md` — web fetch behavior and why we intentionally do not spider web links.

### Plans (`plans/`)
- `plans/.gitkeep` — keeps the folder in git; plans are added ad-hoc.
