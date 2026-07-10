# Lexicon Design Constraints

- Every concept has exactly one canonical term. No synonyms in the platform vocabulary.
- Every noun has exactly one icon. The mapping is permanent.
- Terminology in the Lexicon overrides personal preference. A contributor may prefer "node" but the platform says "device."
- UI strings, API responses, documentation, and error messages use the canonical terms.
- New terms require an RFC. The Lexicon is not changed by casual agreement.
- The Lexicon applies across all clients: Android, Desktop, CLI, web, documentation, and SDK.
- Definitions are written for end users, not engineers. If a definition requires technical knowledge to understand, it needs rewriting.
