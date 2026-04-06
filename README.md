# contextual

CLI tool for fetching Jira issues, Confluence pages, and generic web pages and printing them as **AI-ready context blocks** (one block per item).

## Build

Requires Go 1.21+.

```bash
make build
```

Binary is built to:

```bash
./bin/contextual
```

## Run

```bash
./bin/contextual <item> [<item> ...]
```

If no args are provided, it prints:

```text
Usage: contextual <item> [<item> ...]
```

### Supported `<item>` inputs

`contextual` classifies each CLI argument into an item type (see `internal/spider/spider.go`):

1. **Jira issue key** (uppercase project key + hyphen + digits)
   - Example: `IPM-1234`
2. **Confluence page ID** (8+ digits)
   - Example: `12345678`
3. **Atlassian Jira URL** (must include your configured host and `/browse/<KEY>`)
   - Example: `https://your-company.atlassian.net/browse/IPM-1234`
4. **Atlassian Confluence URL** (must include your configured host and `/wiki/` and contain a numeric page ID)
   - Example: `https://your-company.atlassian.net/wiki/spaces/ABC/pages/12345678/...`
5. **Generic web URL**
   - Example: `https://example.com/some/page`

If an arg is not recognized, it is skipped and logged to stderr:
- `unrecognised argument: "<arg>"`

## Configuration

### Atlassian host (required to fetch Jira/Confluence)

Jira/Confluence fetching requires `atlassian_host` to be configured (otherwise the tool will log an error and skip those items).

Config is loaded via `internal/config.Load()` (see `internal/config/config.go`). If config can’t be loaded, the CLI continues with an empty config and logs a warning.

You should set:

- `atlassian_host`: e.g. `your-company.atlassian.net`

### Atlassian credentials (recommended)

Set these environment variables:

- `CONTEXTUAL_ATLASSIAN_API_EMAIL`
- `CONTEXTUAL_ATLASSIAN_API_TOKEN`

If they aren’t set, the CLI continues, but logs warnings and your Atlassian fetches will likely fail.

## Output format (stdout contract)

For each fetched item, `contextual` prints a block like:

- a separator line of 80 `=` characters
- `TYPE: ...`
- optionally `ID: ...` (only when non-empty)
- `URL: ...`
- `TITLE: ...`
- separator line again
- then the item content body

This is implemented in `cmd/contextual/main.go` (`printItem()`).

Notes:
- Logs and errors go to **stderr**.
- Content blocks go to **stdout** (so you can redirect/pipeline stdout cleanly).

## Examples

Fetch a Jira issue and its related context (parents/subtasks + referenced Confluence/web links):

```bash
./bin/contextual IPM-1234
```

Fetch a Confluence page and its expanded context (child pages + any Jira keys/web links found in it):

```bash
./bin/contextual 12345678
```

Fetch arbitrary URLs (no further link-spidering occurs for generic web pages):

```bash
./bin/contextual https://example.com/blog/post
```

## Agent docs

See [`agent_docs/`](agent_docs/README.md) for agent-first documentation (architecture, patterns, deep dives).