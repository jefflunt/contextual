# Deep dive: Spider queue and traversal

## Executive summary

`internal/spider.Spider.Run` performs a **priority-driven traversal** over a queue of items, expanding the queue based on references discovered in fetched content.

Prioritization is managed by a `container/heap` Priority Queue, which orders items based on:
1.  **Nearness (`jumps`)**: Ascending (1, 2, 3...).
2.  **Direction**: Prioritizing Children (`Down`) > Parents (`Up`) > Siblings (`Sibling`).
3.  **Discovery Order**: Tie-breaker.

This ensures that the most relevant, dependent content is fetched first.

## Control flow diagram

```
seed args
  |
  v
ParseItem -> enqueue seed (PriorityQueue)
  |
  v
while PriorityQueue not empty:
  pop item (highest priority)
  switch item.Type:
    Jira:
      FetchJira
      enqueue parent (DirectionUp)
      enqueue subtasks (DirectionDown)
      enqueue referenced Confluence/Web (DirectionDown)
    ...
```

## De-duplication

Visited key:
- `<type>:<id>` when ID exists
- `<type>:<url>` otherwise

## Context Truncation

After spidering, `cmd/contextual/main.go` iterates through the prioritized results.

Content is written to `context.md` in the exact priority order determined by the spider. Truncation is enforced by monitoring a byte counter (`bytesWritten`) against `MaxContextLength` during the writing phase. If adding the next item exceeds the limit, that item (and all subsequent, less relevant items) is omitted.

## Gotchas

- Missing `max_context_length` causes startup error.
- Missing `atlassian_host` prevents Jira/Confluence fetches entirely.
- Web fetches are intentionally non-spidering to prevent unbounded crawling.