# ADR Guidelines

## Purpose
Use ADRs to record architectural decisions that change system structure, interfaces, dependencies, operating assumptions, or long-lived constraints.

## When To Write An ADR
- A new component, service boundary, or integration pattern is introduced.
- A public contract changes or a compatibility guarantee is added or removed.
- A significant dependency is adopted, replaced, or deprecated.
- A non-obvious tradeoff is accepted for scale, reliability, security, or maintainability.
- A decision is important enough that future contributors will ask "why was this done?"

## When Not To Write An ADR
- The change is local implementation detail with no architectural consequence.
- The decision is already fully covered by an existing ADR.
- The proposal is still too vague to describe the problem, options, and consequences.

## Recommended Structure
- `Title`: short statement of the decision.
- `Status`: `Proposed`, `Accepted`, `Deprecated`, or `Superseded`.
- `Context`: what problem exists and what constraints matter.
- `Decision`: the chosen approach in plain language.
- `Alternatives`: the main options considered and why they were not chosen.
- `Tradeoffs`: what is gained and what is given up.
- `Consequences`: what must change, what becomes easier, and what becomes harder.

## Writing Rules
- Describe the problem before the solution.
- Prefer concrete language over abstract principles.
- Capture the real alternatives, not strawmen.
- State the tradeoff explicitly even when the choice seems obvious.
- Keep the ADR understandable without opening code.

## Quality Check
- Is the decision specific enough to guide future work?
- Would a new contributor understand why this option won?
- Does the ADR name the affected interfaces, boundaries, or dependencies?
- Are follow-on obligations and migration consequences visible?
