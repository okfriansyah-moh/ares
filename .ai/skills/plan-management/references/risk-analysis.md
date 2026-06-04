# Risk Analysis

## Purpose
Identify where a plan can fail, stall, or expand unpredictably, then make mitigation visible inside the plan.

## Risk Categories
- Dependency risk: a task relies on undefined or unstable upstream work.
- Scope risk: a task or milestone is too broad to stay predictable.
- Knowledge risk: implementers lack required schema, domain, or integration context.
- Coordination risk: multiple tasks touch the same files or interfaces.
- External risk: vendors, APIs, approvals, or environments may block progress.

## How To Analyze Risk
1. Review each task for hidden prerequisites.
2. Check where the plan depends on external systems or decisions.
3. Identify bottlenecks that block many later tasks.
4. Mark areas where validation is weak or indirect.
5. Propose mitigation that reduces uncertainty early.

## Risk Record Format
- `Risk`: short description
- `Likelihood`: low, medium, or high
- `Impact`: low, medium, or high
- `Affected tasks`: where the risk shows up
- `Mitigation`: what reduces or contains the risk
- `Trigger`: what sign shows the risk is materializing

## Typical Mitigations
- Split a large task into smaller verified steps.
- Add a prerequisite task for schema, interface, or environment setup.
- Move a high-uncertainty task earlier to learn sooner.
- Document missing knowledge directly in the plan.
- Reduce parallel edits to the same files or contracts.

## Quality Check
- Does the mitigation change the plan, not just describe the problem?
- Are the highest-impact risks visible before execution starts?
- Would a teammate know what to watch for while carrying out the plan?
