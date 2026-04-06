# Deep dive: Web fetcher and intentional non-spidering

## Executive summary

Generic web URLs are fetched and emitted as items, but `contextual` intentionally does not spider hyperlinks discovered in web content. This keeps traversal bounded and avoids turning the tool into a general-purpose crawler.

## Files

- `internal/fetcher/web.go`
- `internal/spider/spider.go` (`case types.ItemTypeWeb`)

## Behavior

- Spider enqueues web URLs discovered from Jira/Confluence.
- When a web item is processed:
  - `fetcher.FetchWeb(url)` is called
  - resulting `Item` is appended to results
  - **no new items are enqueued**

## Rationale / gotchas

- Web pages can contain many links; spidering would explode the queue.
- Many links are irrelevant (nav links, trackers, etc.)
- If you ever introduce controlled web spidering, it must be:
  - depth-limited
  - domain-allowlisted
  - aggressively de-duplicated
  - documented + tested