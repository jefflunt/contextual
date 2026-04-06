# Domain model

## Table of contents
- [Item](#item)
- [ItemType](#itemtype)
- [Traversal mental model](#traversal-mental-model)
- [Uniqueness / visited key](#uniqueness--visited-key)

## Item

Defined in `internal/types/types.go`.

An `Item` is the normalized unit of output printed by the CLI.

Fields (conceptually):
- `Type` — what kind of thing this is (Jira, Confluence, Web)
- `ID` — optional identifier (Jira key, Confluence numeric ID); empty for web URLs
- `URL` — canonical URL (or best-effort) for the item
- `Title` — human label
- `Content` — body text emitted to stdout

## ItemType

Item types used by:
- `cmd/contextual/main.go` (pretty name mapping)
- `internal/spider` (switch for traversal behavior)

Current types inferred from usage:
- Jira issue
- Confluence page
- Web page

## Traversal mental model

This CLI does **seeded context expansion**.

You give it “seed” items; it fetches them; and it discovers new items to fetch based on extracted references:

- From Jira:
  - parent issue (if any)
  - subtasks
  - referenced Confluence pages
  - referenced web URLs
- From Confluence:
  - child pages
  - referenced Jira issues
  - referenced web URLs
- From Web:
  - **no expansion** (explicitly does not spider further)

## Uniqueness / visited key

Spider de-duplicates with a key:

- `<type>:<id>` when `ID` is present
- `<type>:<url>` when `ID` is empty

Implication:
- Web items are unique by URL.
- Jira/Confluence items are unique by ID (URL is secondary).