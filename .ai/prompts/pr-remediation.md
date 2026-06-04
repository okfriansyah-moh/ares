# PR Remediation

## Use

Review a pull request diff for this repo and produce concise, actionable remediation guidance.

## Inputs

- PR title, description, and changed files
- key diff snippets or summary of affected code
- relevant repo invariants from `docs/architecture.md` and `docs/PLAN.md`
- failing tests or review comments if available

## Instructions

- Be credit-efficient: terse output, short fragments, no filler or hedging.
- Follow caveman-style brevity: keep meaning, drop fluff, preserve technical accuracy.
- Validate against this repo’s standard: `.ai/` canonical source, generated provider artifact rules, and security/path invariants.
- Prefer exact remediation steps and minimal fix guidance.
- Reference file paths, sections, or rule checks when possible.
- Use one-line action items or a short numbered list.

## Check

- Output is concise and actionable.
- Recommendations cite repo invariants or files.
- No generic or polite wording.
- Fix advice is specific and easy to apply.
