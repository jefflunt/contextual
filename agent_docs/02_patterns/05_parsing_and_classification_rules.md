# Pattern: Parsing and classification rules

## When to use this pattern

Use when extending what `contextual` accepts as CLI input.

## Current rules (see `internal/spider/spider.go`)

Order matters:

1. Jira key regex: `^[A-Z]+-\d+$`
2. Confluence ID regex: `^\d{8,}$`
3. URL cases:
   - if URL contains `<host>/browse/` -> Jira URL -> extract key from `/browse/`
   - if URL contains `<host>/wiki/`   -> Confluence URL -> extract numeric ID from path
   - else -> generic web URL

## Guidance for changes

- Keep rules deterministic and mutually exclusive where possible.
- Prefer extracting a stable `ID` for visited-key stability.
- Add tests for any new parse branches.

## Gotchas

- Jira URL matching is host-dependent; if `atlassian_host` is empty, Jira/Confluence URLs become generic web URLs.
- Confluence ID extraction uses a numeric regex searching the URL path; ensure it doesn’t create false positives for unrelated URLs.