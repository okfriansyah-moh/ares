# Architecture Management

## Purpose
Create, review, and evolve repository or system architecture with clear decisions and explicit tradeoffs.

## Responsibilities
- Define system boundaries, major components, interfaces, and dependencies.
- Keep architecture aligned with current implementation and stated requirements.
- Record significant decisions as ADRs.
- Detect drift, coupling, and unnecessary complexity.

## Principles
- Prefer the simplest design that satisfies the requirement.
- Make decisions explicit, including alternatives and tradeoffs.
- Evolve from the current state before proposing replacement.
- Keep responsibilities and interfaces clear.

## Inputs
- Requirements, specs, RFCs, or problem statements.
- Existing architecture documents, ADRs, and relevant code context.
- Constraints such as scale, compliance, integrations, or team boundaries.

## Outputs
- `architecture.md` or an updated architecture section for the repository.
- ADRs in `docs/decisions/` for material decisions.
- Review findings grouped as confirmed, drift, gaps, and recommendations.

## Usage Guidance
- Read the current architecture before proposing change.
- Describe each component in one clear responsibility statement.
- Justify external dependencies and identify coupling risks.
- Write an ADR when a decision changes structure, contracts, or operating assumptions.

## References
- `references/adr-guidelines.md` for ADR structure, status handling, and decision quality checks.
- `references/tradeoff-analysis.md` for comparing options and documenting tradeoffs clearly.
- `references/architecture-review.md` for review criteria, drift detection, and reporting format.
