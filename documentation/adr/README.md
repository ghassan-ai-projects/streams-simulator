# Architecture decision summary

The full decision record is [`docs/DECISIONS.md`](../../docs/DECISIONS.md). This page publishes the decisions that constrain the public architecture and records their current status without duplicating every historical note. It is a compact summary, not a numbered set of standalone ADR files.

| Decision | Current position | Status |
| --- | --- | --- |
| Control versus evidence transport | MCP carries control, actuation, and audit; sinks carry event evidence. | Implemented |
| Native event plus adapters | The simulator emits native `sim-event-v0.1`; consumer formats are projections. | Implemented |
| Data-defined domains/effectors | Domain behavior is JSON data loaded through one path; no per-domain production branch. | Implemented |
| File-based run artifacts | Run replay and persistence use JSON/JSONL artifacts rather than a database. | Implemented; supersedes earlier storage design text |
| Director/operator separation | The director owns truth; the operator receives only nameplate, effectors, invocation, and report. | Implemented |
| Quiescence before scoring | Consumer work must settle or the run is marked incomplete. | Implemented |

## How to read decisions

For each decision, read the current summary in this documentation set, then the full record in `docs/DECISIONS.md`, then the implementation and tests named by the relevant page. If the archive still contains an earlier alternative, treat it as historical or superseded.

## Decision hygiene

Create a new decision record when a change constrains architecture, a public contract, trust boundary, or replay semantics. Do not create an ADR for ordinary refactoring. Every decision should name context, decision, invariants, rejected alternatives, consequences, and verification evidence.

## Next reads

- [Design index](../design/README.md)
- [Architecture overview](../architecture/overview.md)
- [Contributor quality guide](../governance/quality.md)
