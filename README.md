# contextual

CLI tool for fetching Jira issues, Confluence pages, and generic web pages and assembling them into **AI-ready context blocks** (`context.md`).

Features **priority-driven spidering** and **context size truncation** to ensure only the most relevant, nearest content fits into your LLM's context window.

## Usage

```bash
contextual [--verbose|-v] [--progress|-p] <item> [<item> ...]
contextual plan [--verbose|-v] [--progress|-p] <item> [<item> ...]
contextual version
contextual help
```

## Build

```bash
./script/build
```

## Configuration

Required: `~/.contextual/config.yml`. See `contextual.config.example.yml` for the structure.

- `max_context_length`: Maximum size (in bytes) of the generated context.md.
- `spider.max_hops`: Maximum number of jumps to make when spidering connected items.
- `atlassian`: Credentials and host for Jira/Confluence.
- `planner`: Command string for `plan` mode (with `<promptFile>` placeholder).

## Examples

Fetch items and generate context (truncated to `max_context_length`):

```bash
contextual CTX-1234
```

Generate a plan file for an AI agent:

```bash
contextual plan CTX-1234
```

## Agent docs

See [`agent_docs/`](agent_docs/README.md) for agent-first documentation (architecture, patterns, deep dives).