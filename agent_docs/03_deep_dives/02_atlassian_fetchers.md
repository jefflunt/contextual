# Deep dive: Atlassian fetchers (Jira + Confluence)

## Executive summary

Jira and Confluence fetchers live under `internal/fetcher/` and require both a configured host and API credentials. They return the fetched `Item` plus extracted references that drive spider expansion. Content is emitted as **Markdown**, not plaintext.

## Files

- `internal/fetcher/jira.go`
- `internal/fetcher/confluence.go`

## Dependency inputs

Provided by `internal/spider` (which reads them from `cfg.Atlassian.*`):
- `host` — from `cfg.Atlassian.Host`
- `email` — from `cfg.Atlassian.APIUser`
- `token` — from `cfg.Atlassian.APIToken`
- `id` — Jira key (e.g. `CTX-1234`) or Confluence numeric page ID

**No environment variables are involved.** All credentials come from `~/.contextual/config.yml`.

## Content rendering

- **Jira**: body is ADF (Atlassian Document Format, JSON). The fetcher walks the ADF tree and emits Markdown — headings, lists, bold/italic, code blocks, tables, links.
- **Confluence**: body is XHTML storage format. The fetcher walks the HTML tree and emits Markdown using the same conventions.

## Output responsibilities

Fetchers return:
- `types.Item` with `Type`, `ID`, `URL`, `Title`, `Content` (Markdown)
- Extracted references for spider expansion:
  - Jira: parent issue key, subtask keys, Confluence page IDs, web URLs
  - Confluence: child page IDs, Jira issue keys, web URLs

Traversal decisions (what to enqueue) remain in `internal/spider`.

## Gotchas

- If `cfg.Atlassian.Host` is empty, spider skips both fetchers entirely with `[ERROR] atlassian.host not configured`.
- Missing credentials are not validated at startup but will cause auth failures when the Atlassian API is called.
