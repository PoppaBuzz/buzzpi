# Protocol Design Constraints

These constraints define the boundaries of the BuzzPi Protocol (BPP). They ensure the protocol remains open, stable, and implementable by third parties.

- BPP is an open specification. No license or royalty is required to implement it.
- A developer can implement a working client or agent using only the protocol specification. No reference implementation source code is required.
- The protocol is transport-agnostic. Identity, Services, and Capabilities layers work over any transport.
- Each layer versions independently. Identity v2 can ship while Transport v1 remains stable.
- Backward compatibility is maintained within a defined version range.
- Breaking changes require an RFC, migration strategy, version negotiation, and a deprecation window of at least one minor version.
- Messages are human-readable during development (JSON). Binary frames are permitted for high-throughput data (video, file transfer).
- The protocol works entirely over LAN. No cloud service is required for local operations.
- Custom message types and vendor extensions are first-class citizens. The protocol does not require a central registry for extensions.
- The protocol specification is version-controlled in the same repository as the reference implementation, ensuring they never diverge.
