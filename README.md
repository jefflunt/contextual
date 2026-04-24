# contextual

CLI tool for fetching Jira issues, Confluence pages, and generic web pages and assembling them into **AI-ready context blocks** (`context.md`).

When working with LLMs, managing context windows is hard. `contextual` features **priority-driven spidering** and **context size truncation** to ensure only the most relevant, nearest content fits into your LLM's context window.

## Quick Start / Installation

The fastest way to install `contextual` is via our installation script. This works seamlessly on macOS, Linux, and in automated AI agent environments.

```bash
curl -sSfL https://raw.githubusercontent.com/jefflunt/contextual/main/install.sh | bash
```

## Basic Usage

Fetch items and generate a `context.md` file (automatically truncated to your configured `max_context_length`):

```bash
contextual CTX-1234
contextual https://example.com/docs/api
```

You can pass multiple items at once. Add `--verbose` (`-v`) or `--progress` (`-p`) to see the spidering graph in action.

```bash
contextual -v -p SIPS-123 SIPS-124
```

## Usage for AI Agents

`contextual` is designed to be highly scriptable by AI coding agents. Agents can use the `plan` feature to generate an execution plan file based on context:

```bash
contextual plan CTX-1234
```

For more deep-dive documentation aimed specifically at AI agents (architecture, patterns), see [`agent_docs/README.md`](agent_docs/README.md).

## Configuration

`contextual` requires a configuration file located at `~/.contextual/config.yml`.

See `contextual.config.example.yml` in this repository for a complete structure. Key settings include:

- `max_context_length`: Maximum size (in bytes) of the generated `context.md`. Ensures you don't blow out your LLM context window.
- `spider.max_hops`: Maximum number of jumps to make when spidering connected items (e.g., from an Epic to a Task to a Subtask).
- `atlassian`: Credentials and host for fetching Jira issues and Confluence pages.
- `planner`: Command string for `plan` mode (with `<promptFile>` placeholder).

## Building from Source

If you want to contribute or build locally, we recommend running the full build, test, and install suite to ensure everything works smoothly on your machine:

```bash
./script/build-test-install
```
