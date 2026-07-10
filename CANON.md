# The BuzzPi Canon

When two documents disagree, this hierarchy determines which one wins.

## The Canonical Hierarchy

```
North Star
    │
Constitution
    │
Canon
    │
Books
    │
RFCs
    │
Specifications
    │
Reference
    │
Code
```

### North Star — The Aspiration

"A Raspberry Pi should feel like a Bluetooth speaker."

The North Star is the emotional and experiential goal of the entire platform. It is not technical. It is not measurable in the traditional sense. It exists to answer one question: *Does this decision bring us closer to that feeling?*

**Priority:** Highest. If any other document conflicts with the North Star, the North Star wins.

### Constitution — The Immutable Principles

Five articles that define what BuzzPi will never compromise on. Users First. Open by Default. Protocol Stability. No Vendor Lock-In. Clients Are Equal.

**Priority:** Second. The Constitution constrains the books, RFCs, specifications, and code. It can only be changed by amendment, which is rare by design.

### Canon — This Document

The connective tissue. Explains how all the pieces fit together. Defines the vocabulary, the hierarchy itself, and the rules for resolving conflicts between documents.

**Priority:** Third. The Canon interprets the Constitution and North Star but does not override them.

### Books — The Enduring Knowledge

Eight books that each answer a fundamental question about BuzzPi:

1. **Product** — Why does BuzzPi exist?
2. **Experience** — How should BuzzPi feel?
3. **Engineering** — How does BuzzPi work?
4. **Protocol** — How do devices communicate?
5. **Community** — How do we build this together?
6. **Patterns** — How does BuzzPi think?
7. **Reference** — Where are the facts?
8. **Lexicon** — What do we call things?

**Priority:** Fourth. Books describe intended behavior and architecture. An accepted RFC can update a book, but a book reflects current consensus.

### RFCs — The Decision-Making Process

Requests for Comments document major decisions: the motivation, alternatives considered, tradeoffs, and final design. Every accepted RFC is permanent — it remains in the repository as a historical record, even if superseded.

**Priority:** Fifth. A newer accepted RFC supersedes an older one. RFCs can update the Books but cannot override the Constitution or North Star.

### Specifications — The Precise Contracts

Formal specifications for the BuzzPi Protocol (BPP), plugin API, and other interfaces. Specifications are precise enough for a third party to implement without reading source code.

**Priority:** Sixth. Specifications must conform to the protocol layer defined in the Engineering book. If a specification contradicts an RFC, the RFC's rationale wins, but the specification's precision wins for implementation.

### Reference — The Facts

The BuzzPi dictionary. Canonical identifiers, error codes, event types, design tokens, JSON schemas, environment variables, CLI commands. Reference is the source of truth for *what exists*, not *why it exists*.

**Priority:** Seventh. Reference entries are machine-verified where possible. If reference contradicts a specification, reference wins for concrete values. If reference contradicts a book, the book wins for intent.

### Code — The Implementation

Code is the last artifact, not the first. It is an implementation of everything above it. If the code does not reflect the canon, the code is wrong — not the canon.

**Priority:** Lowest. Code must conform to specifications. Pull requests that violate the canon should be rejected with a reference to the relevant document.

---

## Conflict Resolution

When any two documents in the canon disagree:

1. Find the higher-priority document in the hierarchy.
2. The higher-priority document wins.
3. The lower-priority document should be corrected to match.
4. If the conflict reveals a flaw in the higher-priority document, follow the amendment process for that document type.

---

## The Canon Does Not Change Often

The North Star, Constitution, and Canon together form the stable core of BuzzPi. They should be revisited only when the project's fundamental understanding of itself changes. Books, RFCs, specifications, reference, and code evolve continuously. The hierarchy ensures that evolution never breaks the foundation.
