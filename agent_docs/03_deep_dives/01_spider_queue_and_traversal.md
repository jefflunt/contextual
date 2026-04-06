# Deep dive: Spider queue and traversal

## Executive summary

`internal/spider.Spider.Run` performs a BFS traversal over a queue of items, expanding the queue based on references discovered in fetched content. It de-duplicates via a `visited` key to avoid infinite loops and redundant fetches.

## Control flow diagram

```
seed args
  |
  v
ParseItem -> enqueue seed
  |
  v
while queue not empty:
  pop front
  switch item.Type:
    Jira:
      FetchJira
      enqueue parent/subtasks
      enqueue referenced Confluence IDs
      enqueue referenced Web URLs
    Confluence:
      FetchConfluence
      enqueue child pages
      enqueue referenced Jira keys
      enqueue referenced Web URLs
    Web:
      FetchWeb
      (no enqueue)
```

## De-duplication

Visited key:
- `<type>:<id>` when ID exists
- `<type>:<url>` otherwise

Implications:
- If a Jira key appears in many places, it is fetched once.
- Web URLs are fetched once each.

## Non-obvious traversal rule: fromConfluence

Queue entry carries `fromConfluence bool` (currently only tracked; not used to change behavior elsewhere in shown code).

This suggests future logic might treat Jira items discovered via Confluence differently (e.g., fetch fewer related issues). If you introduce such behavior, update docs + tests.

## Gotchas

- Missing `atlassian_host` prevents Jira/Confluence fetches entirely (they are skipped with logs).
- Web fetches are intentionally non-spidering to prevent unbounded crawling.