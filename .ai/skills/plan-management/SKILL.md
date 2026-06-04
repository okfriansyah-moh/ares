# Plan Management

## Purpose
Create, review, and maintain implementation plans that are easy to execute and verify.

## Responsibilities
- Break work into small sequenced tasks.
- Make dependencies, validation steps, and risks explicit.
- Keep plans aligned with current requirements and architecture.
- Update plans without introducing ambiguity or hidden work.

## Principles
- Tasks should be small enough for a focused session.
- Every task needs a concrete validation step.
- Sequence work so prerequisites exist before they are used.
- Prefer working increments over big-bang delivery.

## Inputs
- Requirements, specs, tickets, or change requests.
- Existing plans, architecture guidance, and known constraints.
- Current execution status when reviewing or extending a plan.

## Outputs
- `docs/PLAN.md` or `docs/PLAN-<feature>.md`.
- Review findings grouped as complete, missing, sequencing issues, oversized tasks, and recommendations.
- Added tasks, risks, or milestone updates when the plan changes.

## Usage Guidance
- Read the full source material before decomposing work.
- Put scaffolding and shared contracts before feature tasks.
- Avoid tasks that span multiple layers unless that coupling is the point.
- When adding tasks to an existing plan, preserve numbering and insert them where dependencies make sense.

## References
- `references/task-decomposition.md` for task sizing, ordering, and validation rules.
- `references/roadmap-construction.md` for building milestones and sequencing delivery increments.
- `references/risk-analysis.md` for identifying planning risks and documenting mitigation.
