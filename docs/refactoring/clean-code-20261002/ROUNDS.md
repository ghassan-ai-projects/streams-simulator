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

## Round 5 — Validation and adapter steps

Domain cross-checks now read as channels, dynamics, cycle rejection, faults,
effectors and profiles. Private validation context owns source-aware errors.
Schema compilation names value/object/array/string/number/composition/conditional
keyword phases. Runtime validation preserves type short-circuiting, primitive
checks, object/array traversal and composition order. Adapter validation replaces
mutually recursive closures with named steps; expression evaluation delegates
object/concat/template/time/counter operations. Verification separates fixture
rendering, schema conformance and golden comparison.

Characterization tests were green before extraction and remain green afterward:
keyword error order, type short-circuit, nested projection identity/source context
and channel-before-fault error priority. Existing schema and golden-byte tests pass.
Race coverage: adapter 69.1%, domain 76.1%, jsonschema 68.8%. Final focused tests,
size/dependency guards, zero-issue lint and whitespace checks pass. No intended
runtime or contract changes in this round.

## Round 6 — World effects and perturbation ownership

The perturbation dispatcher delegates to named duplication, timing and payload
transforms in cohesive files. Each extracted operation retains its original body,
short-circuit checks and PRNG calls. World actuation separates admission/replay/
interlock checks from selected-mode execution. Fault severity parsing, state-kick
evaluation and native-event publication now sit below their callers.

Captured a fixed-seed overlapping perturbation stream before refactoring and pinned
its complete delivery digest (`d34d9a91…`). The digest is unchanged afterward.
Added an argument-validation-before-idempotency regression. Existing analytic
oracles, observed-time ordering, substream isolation, closed-loop and replay tests
pass. Numerical RK4 arithmetic remains together for independent oracle review.
Race coverage: perturb 85.2%, world 64.5%, run 68.2%. Architecture guards, zero-issue
lint and whitespace review pass. No intended behavior changes.

## Round 7 — Device admission and transport exchange

Capability loading separates strict decoding, digest preparation and route/target
construction. Admission names boot/freshness checks, safe-stop authorization and
capability parameter checks. Command handling separates scheduled execution faults,
freshness injections and response faults while retaining lock/lease/dedup order.
Connection framing delegates state queries, malformed exchange and command outcome
emission to small private steps; receipt precedes result as before.

Pre-extraction regressions pin boot-before-freshness-before-target rejection and a
mixed blank/malformed/query/command stream with exact frame order. Those pass after
extraction, alongside lease, safe-stop, ack-loss retry, duplicate, reboot and world
binding tests. Race tests pass with local socket access. Coverage: device 84.1%,
deviceworld 76.8%. Architecture guards, zero-issue lint and whitespace checks pass.
No intended behavior changes.

## Round 8 — Application workflows

CLI run reads as parse options, load config, create run, apply scripted faults /
perturbations / effects, advance and publish. Flags/defaults and wall-derived
command IDs are unchanged. Device-world setup, reference-consumer connection/output
and manifest signing sit below their workflows. MCP world creation separates
configuration, identity reservation and registry publication. Director server
registration reads as catalog, world/clock, injection, run, truth and resources;
registration order and the original handlers/schemas are retained.

Added regressions for explicit epoch zero, default CLI inputs, derived file targets,
script command order, ed25519 signing over the canonical-body digest and distinct
MCP world/run/token identities. Extraction tests use existing process replay and
role/capability suites; new characterization was added afterward where existing
integration tests already covered the path. Full CLI tests pass (26.5% coverage),
MCP race tests pass (65.7%). Final focused CLI tests, architecture guards, zero-issue
lint and whitespace checks pass.
