# Task Decomposition

## Purpose
Break a requirement into tasks that are small, ordered, and independently verifiable.

## Decomposition Goals
- Each task has one clear outcome.
- Each task can be validated without relying on later tasks.
- Dependencies are visible in the order of work.
- The plan can be executed with minimal interpretation.

## Sizing Rules
A good task usually:
- touches one primary concern or layer
- creates or updates a small number of related files
- has one concrete validation step
- fits within a focused work session

A task is too large when:
- it mixes architecture, infrastructure, backend, and frontend work in one step
- it requires multiple future tasks before validation is possible
- it hides several decisions behind one broad label like "build feature"

## Sequencing Rules
1. Start with scaffolding, shared contracts, or foundational data structures.
2. Define interfaces before implementation details that depend on them.
3. Place infrastructure and integration prerequisites before feature work.
4. End with integration, migration, or final verification tasks.

## Task Template
- `Goal`: what changes by the end of the task.
- `Files`: what is created or updated.
- `Dependencies`: what must already exist.
- `Validation`: the command, test, or observable check.
- `Notes`: any narrow context the implementer must know.

## Quality Check
- Can someone execute the task without guessing missing prerequisites?
- Does the validation prove completion rather than partial progress?
- Would splitting the task reduce coordination risk or merge conflict risk?
