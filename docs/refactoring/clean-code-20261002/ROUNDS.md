# Refactoring rounds

## Round 1 — Bar and architecture guard

Established the 300-total-line rule, named ownership and stepdown criteria in
AGENTS.md and BAR.md. Inventoried all 103 Go files; 30 exceed the cap. The current
business package graph needs internal organization, not additional Go modules.
Added a source-based import guard that requires deliberate review for new edges.

Baseline short tests pass outside socket-dependent device/MCP/sink tests. Those
fail with `bind: operation not permitted` in the restricted environment; rerun with
local socket access before claiming the full gate.

## Round 2 — Responsibility files and enforced size limit

Partitioned all 30 oversized production/test files inside their existing packages.
Files now name loading, validation, expressions, encoding, capabilities, admission,
clock, delivery, commands, artifacts, scoring policies and test scenarios.
No package ownership or import edges changed. The source-walking size test includes
all Go files and proves handling of blank/comment lines and a missing final newline.

Compared all 1,099 baseline declarations against the partitioned declarations:
identical text after whitespace normalization (two one-line functions were realigned
by gofmt). No function bodies, signatures or data literals changed. Existing runtime
regressions and oracles remain intact. This relocation round uses that preservation
proof and the new architecture tests; behavioral rounds add focused runtime tests.

Validation: `go test -short -cover ./...` passes with local socket access; every
modified production package except suite has nonzero short-test coverage (suite's
runtime tests are deliberately non-short; full-suite coverage is required at final
validation). All 199 Go files are at most 300 lines; largest is 298. Dependency and
size guards pass. Lint and whitespace review pass.

## Round 3 — One owner for shared scoring policies

Online and offline paths now share judgment, admission, dropped-event detection,
evidence grounding, action/interlock fidelity, instrument and basic loop policies.
Consumer orchestration reads as named metric operations instead of 170 lines of
bookkeeping. Resolution and deadline checks remain online where world history is
available. Empty available ledgers retain vacuous success; unavailable offline
ledgers retain conservative zero metrics.

Tests first reproduced order-sensitive offline judgment, missing offline identity
conflict gating and wrong-entity action credit. Those are intentional corrections.
Online malformed evidence citations now fail grounding, matching offline behavior.
The scoring bundle advances to `scorecard-bundle-v0.2` to identify corrected metric
semantics; JSON fields and function signatures are unchanged.

Validation: scoring race tests and architecture guards pass; scoring coverage is
83.7%. Lint has zero issues. Self-review checked entity/effector/command matching,
nil versus empty evidence, earliest detection and retained history-only metrics.

## Round 4 — Run orchestration and resource ownership

Run construction now names defaults, ledger, adapter, sink and trace steps.
Emission separates native-time validation, hidden-state capture and delivery.
A single delivery-row helper replaces fourteen identical constructors without
changing values. Advance separates the locked mutation/durable boundary from the
unlocked consumer wait. Replay separates inputs, identity validation and config.
End names trace completion, durable-ledger close and evidence publication.

A failing regression proved that End retained the durable ledger descriptor, with
and without output publication. End now flushes, syncs and closes that owned file.
This is an intentional resource-leak correction. Added terminal delivery identity/
reason tests. Preserved native versus buffered recorder ordering, first-error
semantics, locks, command logging before quiescence and byte-identical replay.
Publication failure remains terminal as before; retry/atomic publication needs a
separate lifecycle design rather than an implicit refactor behavior change.

Validation: run/score race tests and architecture guards pass; run coverage 65.0%.
New close/delivery regressions pass after final edits. Lint has zero issues and
whitespace checks pass. All files remain under 300 lines.
