# Pattern: Adding a new ItemType

## When to use this pattern

Use when you want `contextual` to support a new class of content with distinct traversal and output naming.

## Steps

1. Extend `internal/types/types.go`
   - add a new `ItemType` constant
   - ensure `Item` supports whatever identification scheme you need (`ID` vs URL)

2. Update CLI output naming
   - `cmd/contextual/main.go`: update `itemTypeName(...)` switch for friendly display

3. Update parsing/classification
   - `internal/spider/spider.go`: update `ParseItem(...)` to recognize new input forms

4. Update traversal
   - `internal/spider/spider.go`: update `Run(...)` switch to fetch and (optionally) expand

5. Add tests
   - Add tests for parsing/classification behavior (`ParseItem`)
   - Add tests for any special-case traversal decisions if feasible without live calls

## Gotchas

- The `visited` key uses `<type>:<id_or_url>`. Decide:
  - what should be `ID` (stable identifier) vs URL
- Don’t break the stdout block contract (other tooling may rely on it).