# Maintainers Guide

**This guide describes the responsibilities, expectations, and workflows for BuzzPi maintainers.**

---

## Core Values

1. **Lead by example** — Uphold the highest standards of code quality, community interaction, and documentation
2. **Be responsive** — PRs and issues should not languish. If you cannot review in 48 hours, say so
3. **Say no kindly** — Not every feature belongs in BuzzPi. Explain reasoning, offer alternatives
4. **Delegate** — You are not the only maintainer. Trust other maintainers, distribute the load
5. **Know when to escalate** — Disagreements among maintainers go to the BDFL

---

## PR Review Process

### Triage

When a PR is opened:
1. **Label it** — `feature`, `bug`, `docs`, `RFC`, `enhancement`, `question`
2. **Assign a reviewer** — Assign yourself or another relevant maintainer within 24 hours
3. **Check basics** — Does the PR have a description? Does it reference an issue or RFC? Does CI pass?

### Review

1. **Read the code** — Not just the diff. Understand the intent
2. **Run the tests** — If CI is not sufficient for confidence, pull the branch and run locally
3. **Leave comments** — Be specific, be kind. "This approach has a race condition because..." not "This is wrong"
4. **Request changes** or **Approve** — Partial approval: approve with blocking comments

### Merge

1. **Squash and merge** — Default for feature branches
2. **Rebase and merge** — For PRs with clean, atomic commit history (rare)
3. **Delete branch** — After merge

**Never force push to `main`.** Never merge your own PR unless it's trivial (typo fix, CI config) and another maintainer has explicitly delegated.

### Merge Requirements

- At least one maintainer approval
- All CI checks passing
- No unresolved review threads
- Changelog entry (for user-facing changes)
- RFC linked (for major changes)

---

## Issue Management

### Labels

| Label | Meaning |
|-------|---------|
| `bug` | Something is broken |
| `feature` | New capability request |
| `enhancement` | Improvement to existing feature |
| `docs` | Documentation issue |
| `RFC` | Needs an RFC before implementation |
| `good first issue` | Good entry point for new contributors |
| `help wanted` | Maintainers want community help |
| `blocked` | Waiting on something else |
| `duplicate` | Already reported |
| `wontfix` | Valid but will not address |
| `priority/critical` | Needs immediate attention |
| `priority/high` | Needs attention soon |

### Triage

- Every issue gets a label within 48 hours
- Every bug report gets a severity assessment within 24 hours
- Issues with `priority/critical` are assigned immediately

### Closing

- Close resolved issues with a comment referencing the PR
- Close duplicates with a link to the original
- Close `wontfix` with an explanation

---

## RFC Review

Maintainers are responsible for shepherding RFCs through the lifecycle:

1. **Initial review** — Does the RFC template have all required sections?
2. **Discussion period** — Ensure community feedback is heard. Respond to questions within 72 hours
3. **Final decision** — Vote: accept, decline, or request revisions. Simple majority among active maintainers
4. **Implementation tracking** — Tag accepted RFCs with the target release milestone

---

## Release Responsibilities

One maintainer is designated **Release Shepherd** for each release cycle:

1. Cut the release branch
2. Track remaining PRs targeting the release
3. Run the release process (see [release-process.md](release-process.md))
4. Write release notes
5. Announce the release

---

## Onboarding New Maintainers

When a new maintainer is appointed:

1. Add them to the `@buzzpi/maintainers` GitHub team
2. Grant them write access to the repository
3. Grant them access to:
   - Project board
   - CI/CD configuration
   - Social media accounts (optional)
   - Package publishing credentials
4. Schedule a 30-minute onboarding call to review this guide

---

## Communication Expectations

| Channel | Purpose | Response Time |
|---------|---------|---------------|
| GitHub Issues | Bug reports, feature requests | 48 hours |
| GitHub PRs | Code review | 48 hours |
| GitHub Discussions | Community questions, ideas | 72 hours |
| Internal chat (Discord) | Maintainer coordination | 24 hours |
| Emergency security | Critical vulnerability reports | < 4 hours |

**Communication is asynchronous by default.** If something is urgent, use the emergency channel. Maintainers are volunteers — respect their time and boundaries.

---

## Burnout Prevention

- No maintainer is expected to respond to every notification
- If you need a break, say so. Tag another maintainer to cover your area
- If you feel overwhelmed, raise it with the BDFL
- Maintainers may step down at any time with no hard feelings

---

## Conflict Resolution

1. Discuss the disagreement in a private maintainer channel
2. If unresolved, bring in a neutral third-party maintainer
3. If still unresolved, the BDFL makes the final call
4. All discussions remain confidential

---

## Security Incident Response

1. **Receive report** — Via security@buzzpi.dev (private)
2. **Triage** — Assess severity within 4 hours
3. **Fix** — Develop fix in a private fork
4. **Release** — Ship as a hotfix with release notes after the fix is live
5. **Disclose** — Public disclosure 30 days after the fix is available

Do not discuss security vulnerabilities in public issues.
