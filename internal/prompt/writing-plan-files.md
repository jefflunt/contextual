# Plan Files: How the Work Gets Done

A practical guide for writing plan files to structure and execute well-defined
work with an AI agent.

---

## What a Plan File Is

A plan file is a living markdown document created alongside an AI agent before
implementation begins. It captures the goal, the relevant files, the decisions
made, the open questions and their answers, and a concrete phase-by-phase
implementation plan. It travels with the work from the first agent session
through to the pull request.

---

## The Structure of a Plan File

A plan file has a small set of well-defined sections. Each one serves a specific
purpose.

### `# Goal`

A concise statement of what this work is meant to achieve. Write it in plain
terms: what is being added, changed, or fixed, and what the expected outcome is.
Keep it to a few bullet points.

### `# Reference Material`

A structured list of everything the agent should have access to before it starts
work. This section does two things: it focuses the agent's attention on the
right inputs, and it gives reviewers and future readers a fast map of where the
requirements came from.

- **Jira tickets** — link or paste the tickets that define this work, including
  any sub-tasks or linked issues that affect scope.
- **Confluence docs** — any architecture overviews, API contracts, data model
  docs, or design documents relevant to this implementation.
- **Other materials** — discussion notes, Slack decisions, design files,
  runbooks, or any other input that informed the plan.
- **Relevant files** — every file in the codebase that the agent is expected to
  read, create, or modify.

### `# Need to Know`

Constraints, product decisions, and non-obvious rules that must hold for the
implementation to be correct. Think of this as the set of things you would tell
a new engineer before they started the task — the things that aren't in the code
but that the code must respect.

### `# Questions from AI Agent`

During the planning phase, the agent will surface things that need a human
answer before it can proceed. Record these questions here, along with your
answers. Don't skip this section even if the questions seem minor — the answers
represent explicit decisions that would otherwise be implicit and undocumented.

If certain questions cannot be answered yet, record them as unanswered. They
should be resolved before implementation of the relevant phase begins.

### `# Investigation Notes`

A record of what the agent discovered while exploring the codebase. This is the
output of the exploratory phase applied to a specific ticket: which files were
examined, what patterns were found, how the existing system works, and what that
means for this implementation. It connects the plan to the reality of the
codebase.

### `# Development Plan`

The step-by-step implementation plan. This must be specific enough that the
agent can execute it without ambiguity. Each item should map to a concrete
change: a file to create, a config entry to add, a method to implement. Vague
items like "update the frontend" are not useful here —
"add a new route in `config/routes.rb`: `resources :widgets, only: :index`" is.

This section is where you capture the _what_ and the _where_ of the
implementation, grounded in the investigation notes above it.

Work is broken into **phases**. A phase should represent a coherent unit of work
that leaves the codebase in a valid state when complete — even if not yet
feature-complete. Good phase boundaries tend to fall at natural seams:
infrastructure before behaviour, backend before frontend, implementation before
tests.

Within each phase, list a checklist of TODOs. Mark items complete as you go.

### `# Test Plan`

A description of the automated test coverage for this work, broken down by
behaviour being tested. For each item, include the spec file and the specific
test cases expected. This serves as a checklist for the agent and as
documentation for reviewers of what the tests are actually verifying.

#### `## Test Execution Results`

Once the tests have been written and run, record the exact command used and the
result. Paste the actual command so anyone picking up the work can reproduce it
exactly.

### `# Implementation Status`

A running summary of where the work stands. Update this as phases are completed.
At the top, state which phase is done and what comes next. In a sub-section,
list every file that was created or modified with a brief note on what changed.

This section makes handoffs cheap. If you need to continue the work in a new
session, or hand it to a colleague, the state of the implementation is
immediately clear without re-reading the entire file.

---

## Breaking Work into Phases

The Development Plan section describes _what_ to build. Phases describe _in
what order_ to build it. Breaking work into phases keeps implementation sessions
focused, makes it easier to checkpoint progress, and reduces the cost of
stopping and restarting.

A phase should represent a coherent unit of work that leaves the codebase in a
valid state when complete — even if it's not yet feature-complete. Good phase
boundaries tend to fall at natural seams: infrastructure before behaviour,
backend before frontend, implementation before tests.

Within each phase, the agent should work through a checklist of TODOs. Mark
items as complete as you go — this keeps the plan file accurate and gives you a
clear picture of progress when you review the work.

---

## Quick Reference

| Section                  | Purpose                                                            |
|--------------------------|--------------------------------------------------------------------|
| `# Goal`                 | What this work achieves, in a few bullet points                    |
| `# Reference Material`   | Jira tickets, Confluence docs, other materials, and relevant files |
| `# Need to Know`         | Constraints and decisions that must hold throughout implementation |
| `# Questions from AI Agent` | Open questions surfaced during planning, with answers recorded  |
| `# Investigation Notes`  | What the agent learned about the existing system                   |
| `# Development Plan`     | Concrete, file-level steps to implement the work                   |
| `# Test Plan`            | Specific test cases, spec files, and the exact command to run them |
| `# Implementation Status`| Current phase and a list of all files modified or created          |
