# Spaghetti Code Review

## Purpose
Review a codebase for tangled, hard-to-change code and classify functions, files, and folders by maintainability risk with concrete recommendations and solutions.

## Inputs
- Repository structure, source files, tests, generated artifacts, and build configuration.
- Existing architecture, contribution, style, and testing guidance.
- Recent diffs or target scope when the review is limited to a feature, PR, package, or module.
- Available static analysis, test coverage, dependency graph, complexity, duplication, and lint output.

## Classification
- `Healthy`: clear responsibility, low coupling, readable flow, focused tests, and stable boundaries.
- `Watchlist`: mostly understandable, but showing early signs of growth, duplication, long functions, weak names, or scattered knowledge.
- `Tangled`: multiple responsibilities, hidden side effects, unclear dependencies, brittle conditionals, or tests that make change difficult.
- `Spaghetti`: control flow, state, dependencies, naming, and boundaries are so mixed that local changes are risky without refactoring.
- `Critical Hotspot`: high-change or high-impact spaghetti code that also affects security, data integrity, user-visible behavior, releases, or core architecture.

## Review Method
- Start with a repository map: identify languages, entry points, major folders, generated code, test boundaries, and ownership conventions.
- Review from broad to narrow: classify folders first, then files, then the functions or methods that explain the risk.
- Use evidence, not taste: cite concrete signals such as branching depth, function length, global state, cyclic dependencies, duplicate logic, mixed abstraction levels, temporal coupling, poor error handling, missing tests, or architecture drift.
- Separate symptoms from causes: describe what is tangled, why it became risky, and what change pressure will make it worse.
- Distinguish code smells from intentional design: do not penalize adapters, generated files, simple scripts, or framework glue without evidence of change risk.
- Prefer incremental repair plans over rewrites unless the code is isolated, low-risk, and cheaper to replace than untangle.

## Findings Format
- Lead with the highest-risk hotspots.
- Include file, folder, function, or symbol references where possible.
- For each finding, include:
  - `Classification`: one of the defined classification levels.
  - `Scope`: folder, file, class, function, or code path affected.
  - `Evidence`: specific maintainability signals observed in the code.
  - `Risk`: how this makes future changes, debugging, testing, or releases harder.
  - `Recommendation`: the preferred direction of change.
  - `Solution`: concrete next steps, ordered so they can be implemented safely.
- End with a short remediation order that groups fixes into immediate, near-term, and later work.

## Recommendation Guidance
- For long functions, extract cohesive helper functions only after naming the separate responsibilities.
- For large files, split by stable domain concepts, not by arbitrary line count.
- For tangled folders, introduce clear package boundaries before moving code.
- For duplication, remove it only after confirming the duplicated behavior is truly the same.
- For hidden state or side effects, make dependencies explicit and push mutation to narrow edges.
- For complex conditionals, prefer table-driven logic, small strategy objects, or explicit state transitions when they match the domain.
- For weak tests, add characterization tests before refactoring risky behavior.
- For architecture drift, propose the smallest boundary correction that restores the intended dependency direction.

## Output Expectations
- Classify only the code that was actually inspected.
- State when a classification is inferred from partial evidence.
- Keep recommendations proportional to the risk and likely change frequency.
- Tie every proposed solution to an observed maintainability problem.
- Avoid style-only findings unless they materially affect comprehension or change safety.

## Anti-Patterns
- Calling code spaghetti because it is unfamiliar.
- Recommending a rewrite without comparing incremental refactoring cost and risk.
- Reporting generic clean-code advice without file or function evidence.
- Treating generated code, vendored code, or framework boilerplate as refactoring targets.
- Mixing unrelated product, architecture, security, and style opinions into maintainability findings.
- Ignoring tests and release risk when proposing refactors.
