# BuzzPi Governance

This document describes how BuzzPi is governed, how decisions are made, and how community members can take on leadership roles.

---

## Project Structure

BuzzPi uses a BDFL (Benevolent Dictator for Life) model that transitions toward core maintainer governance as the project matures.

### Roles

**BDFL**
The BDFL has final decision authority on all project matters, including RFC acceptance, maintainer appointments, and dispute resolution. This authority is intended to be used rarely and only when consensus cannot be reached.

**Core Maintainers**
Core Maintainers are trusted community members who have demonstrated a deep understanding of BuzzPi's architecture and philosophy. They:
- Review and accept RFCs
- Approve major pull requests
- Set project direction and priorities
- Nominate new Core Maintainers

**Committers**
Committers have merge access and are responsible for reviewing pull requests within their area of expertise. They are expected to uphold BuzzPi's quality standards and design philosophy.

**Contributors**
Anyone who contributes via pull requests, issues, documentation, design, testing, or community support is a Contributor. Every contribution is valued regardless of its size.

---

## Decision-Making Process

### RFC-Driven Decisions

Major changes to BuzzPi require an RFC (Request for Comments). This includes:
- New features or subsystems
- Protocol changes
- Architecture changes
- Breaking API changes
- New plugins that ship with the agent

### RFC Lifecycle

1. **Proposal** — The author submits an RFC as a pull request in the `rfcs/` directory. The RFC must include: motivation, architecture, alternatives considered, tradeoffs, security implications, and migration path.

2. **Discussion** — Community feedback is gathered over a minimum of two weeks. The author may revise the RFC based on feedback.

3. **Comment Period** — Core Maintainers review the final RFC. The comment period lasts one week.

4. **Decision** — The RFC is either:
   - **Accepted** — approved for implementation
   - **Declined** — rejected with reasoning
   - **Withdrawn** — withdrawn by the author

5. **Implementation** — The accepted RFC is implemented, typically by the author or a volunteer.

### Minor Decisions

Minor changes such as bug fixes, documentation improvements, and small enhancements can be made through normal pull request review. Lazy consensus is assumed after 72 hours if no objections are raised.

---

## Adding Core Maintainers

1. A current Core Maintainer nominates a candidate
2. Existing Core Maintainers discuss and evaluate
3. The BDFL confirms the appointment
4. The new maintainer is announced to the community

Nominees are evaluated on their technical contributions, understanding of BuzzPi's philosophy, community interactions, and availability.

---

## Conflict Resolution

If a dispute cannot be resolved through discussion among Core Maintainers, the BDFL has final authority. This authority is expected to be exercised sparingly and only when the project's health is at stake.

---

## Code of Conduct Enforcement

Code of Conduct violations are handled by the Core Maintainers. Reports are confidential. Enforcement follows the guidelines in [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

---

## Role Expectations

All roles are volunteer positions. There is no expectation of sustained commitment, though regular contributors may be invited to take on formal roles. Anyone may step down from a role at any time.

---

## Amendments

This governance document may be amended through the RFC process. Amendments require approval by the BDFL.
