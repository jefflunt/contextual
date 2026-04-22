# Plan: Contextual Prioritization and Truncation

This plan outlines the implementation of prioritized spidering and context-aware truncation in the `contextual` tool.

## Phase 1: Configuration Updates
*   **Action**: Add `MaxContextLength` (`int`) to the top-level `Config` struct.
*   **Action**: Update `config.Load` to require `max_context_length`. If missing, return an error.
*   **Action**: Update `contextual.config.example.yml` to include `max_context_length: 10240`.

## Phase 2: Priority-Driven Spider Refactoring
*   **Action**: Replace the simple `queue` slice in `internal/spider/spider.go` with a Priority Queue implementation using `container/heap`.
*   **Action**: Define `Direction` enum (Source, Down, Up, Sibling).
*   **Action**: Implement a `Score` function to sort the queue based on:
    1.  Nearness (`jumps` - ascending).
    2.  Direction (Prioritize `Down` > `Up` > `Sibling`).
    *   *Note*: This priority is applied recursively, ensuring children/grandchildren are prioritized over parents/grandparents, and deeper dependency chains are prioritized over shallower ones or siblings at the same depth.
    3.  Discovery Order (tie-breaker).
*   **Result**: The spider will continue fetching until it hits `max_jumps` or exhausts connected items, ensuring the full graph is available in-memory, ordered by relevance.

## Phase 3: Context Serialization and Truncation
*   **Action**: Update `cmd/contextual/main.go` to iterate through the prioritized result slice returned by the spider.
    *   *Crucial*: The serialization order MUST match the fetch order defined by the Priority Queue to maintain relevance.
*   **Action**: Implement byte-counting logic when writing to `context.md`.
*   **Action**: Terminate writing once the cumulative byte count (content + metadata) exceeds `MaxContextLength`.

## Phase 4: Verification
*   **Action**: Create test cases that verify both the fetch order (via logs) and the final `context.md` content size.

---

## Example Scenario: Jira-1 (Source)
- **Source**: Jira-1
- **Children**: Jira-2, Jira-3 (Nearness: 1, Direction: Down)
- **Parent**: Jira-Epic (Nearness: 1, Direction: Up)

### 1. Spidering Order (Fetch Priority)
The spider fetches content based on its priority score (Nearness, then Direction).

```text
[Queue]
Step 1: Fetch(Jira-1) (Nearness: 0)
Step 2: Fetch(Jira-2) (Nearness: 1, Down)
Step 3: Fetch(Jira-3) (Nearness: 1, Down)
Step 4: Fetch(Jira-Epic) (Nearness: 1, Up)
```
*The spider ensures that children are processed before parents when nearness ties, adhering to the dependency-first requirement.*

### 2. Context Serialization (Write Priority)
Content is written to `context.md` in the exact order items were fetched.

```text
[context.md]
+-----------------+
|   Jira-1 (Header + Content)
+-----------------+
|   Jira-2 (Header + Content)
+-----------------+
|   Jira-3 (Header + Content)
+-----------------+
|   Jira-Epic (Header + Content) ... (stop if > 10kb)
+-----------------+
```
*The serialization process uses the same prioritized order as the spider. If `Jira-Epic` causes the total size to exceed 10KB, it (and subsequent items) will be omitted.*
