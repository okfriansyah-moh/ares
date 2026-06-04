# Architecture Review

## Purpose
Review the current architecture for clarity, correctness, drift, coupling, and missing decisions.

## Review Scope
- System boundaries and component responsibilities
- Interfaces and dependency direction
- External services and major libraries
- Alignment between documentation, ADRs, and implementation
- Places where complexity has grown without an explicit decision

## Review Questions
- Does each component still match its documented responsibility?
- Are interfaces stable, narrow, and intentional?
- Are dependencies explicit and flowing in the intended direction?
- Has any important decision been implemented without documentation?
- Has a shortcut become a permanent architecture path without review?

## Drift Signals
- Components doing work outside their stated boundary
- Direct access that bypasses intended interfaces
- New dependencies missing from architecture docs
- Shared modules accumulating unrelated responsibilities
- Performance or reliability decisions present in code but absent from ADRs

## Reporting Format
- `Confirmed`: architecture choices that remain correct and consistent
- `Drift`: places where implementation and documentation disagree
- `Gaps`: undocumented dependencies, boundaries, or decisions
- `Recommendations`: specific changes with rationale and tradeoff

## Review Quality Check
- Findings cite the component, interface, or decision they refer to.
- Recommendations solve a concrete issue, not a style preference.
- The report distinguishes urgent structural risk from minor cleanup.
