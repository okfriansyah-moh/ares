# Tradeoff Analysis

## Purpose
Compare architectural options in a way that makes the decision boundary, the winning choice, and the cost of that choice explicit.

## What To Compare
- Simplicity versus flexibility
- Consistency versus local optimization
- Delivery speed versus long-term maintainability
- Operational complexity versus runtime efficiency
- Vendor convenience versus portability

## Method
1. Define the decision to be made in one sentence.
2. List the real options, including keeping the current state.
3. Identify the evaluation criteria that matter for this repository.
4. Compare each option against those criteria.
5. State the chosen option and the reason it wins.
6. Record what is intentionally given up.

## Evaluation Criteria
- Fit to current requirements
- Compatibility with existing architecture
- Cognitive load for contributors
- Impact on interfaces and dependencies
- Operational and maintenance cost
- Migration difficulty
- Reversibility if the decision proves wrong

## Warning Signs
- The options are not actually distinct.
- The criteria were chosen to force a predetermined answer.
- Portability, maintenance, or migration cost is ignored.
- The analysis describes benefits but never names the downside.

## Output Pattern
- `Decision`: what is being chosen.
- `Options`: the serious alternatives.
- `Criteria`: how they are being evaluated.
- `Outcome`: which option wins and why.
- `Tradeoff`: what the repository accepts in exchange.
