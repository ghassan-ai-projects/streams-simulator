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

## Round 9 — Generation, audit capture and consumer processing

Suite generation reads as initialize, populate, report shortfalls and return.
A private generation context owns the existing state shared by those steps;
candidate construction names recipe drawing, label sealing and delivered-stream
audit, eliminating the former twelve-argument internal call. Reference processing
names trace consumption, silence detection and verdict reporting. Audit capture
owns immutable delivered series and sample-grid projection.

Pre-extraction regressions pin cumulative re-feeding, malformed records,
quiescence-before-report failure, dropped audit records and a complete fixed-seed
suite JSON digest (`9247a8f2…`). They pass afterward. RNG draw order, rejected-attempt
identity, adaptive sampling, setup/fault order, audit horizon and terminal strings
are unchanged. Focused race checks pass; reference-consumer coverage is 66.8% and
audit coverage 92.9%. The suite snapshot also passes under race detection; full
suite coverage is part of final validation. Architecture checks, zero-issue lint
and whitespace review pass. No intended behavior changes.

## Round 10 — Final architecture review and enforcement

Added a function-review guard covering production declarations and anonymous
functions. Three cohesive encoding/schema/numerical operations retain documented
bounds; new long functions or growth fail. Strengthened root/tool import checks.
Corrected agent/public package maps and replaced unrelated server/service/store
advice with actual simulator ownership. Published the 226-file inventory,
review decisions, explicit behavior corrections and remaining release work.
Grouped imports in 88 changed files; AST printing confirms every non-import
declaration is unchanged by that formatting pass.

Final validation: `make ci-check` passes, including full simdet tests, bounded
fuzzing and vulnerability scan. Full shuffled race tests with coverage pass;
suite generation takes 475.9 s and total coverage is 69.4%. Every modified runtime
package has nonzero coverage. Final architecture additions pass race/simdet checks;
baseline-wide lint reports zero issues. Documentation links, docs smoke and
whitespace checks pass. Optional pre-commit is absent. Four baseline deadcode
findings remain reported by the successful tool invocation and are disclosed in
REVIEW.md. No remote release, manifest or physical-HIL evidence is claimed.

The refactoring bar is met. Existing advisory review files and the initial
README edit remain untouched and excluded from these round commits.

## Round 11 — Enola architecture gate

Applied the Tamoz Enola workflow to this repository: generated a fresh single-repo
snapshot, pinned before the setup edits, regenerated and checked comparable
receipts/delta. Added repository-owned configuration, ignored generated state,
explicit local baseline/check targets and AGENTS instructions. The check enforces
new cycles/layers at confidence 0.8; it does not merely print a report.

Enola finds no package dependency cycles or upward inferred-layer imports. The
existing business import guard remains necessary because generic Go layers group
all internal packages together. Reviewed heuristic candidates and extraction
limits in ENOLA.md. Disposable fixtures prove rejection of new cycles/upward
layers, missing baselines and incomparable extraction inputs. Architecture/domain
tests, docs smoke/links and whitespace checks pass. No runtime source changes;
the previously completed full runtime gate is not repeated for this setup round.


## Round 12 — Strict 15-line source bar

Recorded the user-selected 15-line maximum and the instruction to leave tests
unchanged. Added a standalone, build-excluded AST source checker and the local
`make function-length` target. It checks named functions, methods and literals,
counts all body lines and fails without exceptions. The initial audit reports
288 overlength bodies (281 named, seven anonymous). The new bar remains open.

The checker itself meets the limit. Disposable source fixtures prove boundary
counting, methods/literals, test exclusion and parse-error failure. Existing
architecture checks and documentation smoke checks pass. No runtime behavior or
test files changed in this round. Subsequent rounds migrate production packages
and run their existing tests before committing.


## Round 13 — Canonical serialization steps

Refactored canonical traversal into collection/scalar dispatch, number conversion,
UTF-16 object ordering, member writing and string escaping. Number rendering now
names exact values, scientific decomposition and decimal/exponent notation. Every
production function in `internal/canonical` meets the 15-line maximum.

Existing canonical race tests, package vet/lint and architecture checks pass. Test
files are unchanged. Canonical bytes, supported input types, integer formatting,
negative-zero handling, escaping and error wrapping are preserved. Existing
RFC 8785 deviations remain separate correctness findings.

Enola receipts are comparable, there are no new findings or package spillover,
and three heuristic findings clear. The object traversal algorithm is unchanged;
cleared performance heuristics are not evidence of a complexity improvement. The
strict repository audit still reports 283 overlength bodies outside this package.


## Round 14 — Schema compiler and validator steps

All production functions in `internal/jsonschema` now meet the 15-line maximum.
Compilation names definition registration, reference resolution, keyword filling
and schema-valued assignment. Validation names type/value admission, primitive
constraints, ordered object checks, array constraints and composition branches.
No new runtime package or dependency was added. Test files remain unchanged.

Existing shuffled-independent race checks pass for jsonschema, adapter, domain,
model, device and architecture. The device socket test first hit sandbox binding
restrictions, then passed with socket access. Package lint/vet and whitespace
checks pass. Disposable old/new modules prove equal canonical output/error strings
for 20,023 cases and equal validation diagnostics for six shipped schema fixtures
across mixed valid/invalid values. Reference depth, keyword order, wrapping and
additional-property diagnostic order are preserved.

Enola receipts are comparable and changes remain within jsonschema. Two complexity
findings clear. A new low-confidence quadratic warning on validateAdditional was
source-reviewed: it still gathers and sorts extra names once, then validates each
extra once; extraction preserves the previous traversal. It is a heuristic, not a
new algorithmic regression. The enforced cycle/layer gate passes.

The strict repository audit remains open with 263 overlength production bodies.


## Round 15 — High-impact adapter and run workflows

Prioritized duplicated loader/framing code and long lifecycle entry points. Adapter
file/byte loading now shares decoding and contract validation while retaining each
public error layout. Streaming and batch rendering share ordered pre/postamble
steps. Run creation names world/state construction, pipeline opening and emitter
attachment; completion names locked admission, trace closing and ordered evidence
publication. Every function in the four modified source files meets 15 lines.

Existing shuffled race tests pass for adapter, run and architecture; run takes
43.9 seconds. Package lint/vet and whitespace checks pass. Test files and exported
signatures remain unchanged. Terminal publication failures, constructor cleanup
and short hash-suffix admission remain separate correctness findings; this round
preserves behavior. The strict audit now reports 250 overlength production bodies.

Per user steering, Enola is checked once at the end of the batch, immediately
before the round commit, using the retained pre-change baseline and the enforced
cycle/layer policy. Subsequent work prioritizes straightforward workflow and
duplication improvements before lower-impact helper splitting.


## Round 16 — CLI command and catalog dispatch

Replaced the long command switch with a declarative handler registry and named
help/error/exit-status steps. Shared catalog/adapter verb parsing and split their
loading, selection and reporting. The help text is a constant; its exact bytes
remain unchanged. Every function in both modified source files meets 15 lines.

Existing shuffled CLI/architecture race tests, lint/vet, documentation smoke and
whitespace checks pass. Test files remain unchanged. Directory loading order,
verb defaults, duplicate adapter handling, exit codes and error wrapping are
preserved. Validation is batched once before the round commit, including the
retained Enola baseline check. The strict audit remains open at 244 bodies.


## Round 17 — Declarative MCP input contracts

Replaced the 185-line schema switch with a registry of fresh input builders,
grouped by director/operator responsibility. Shared world/run/token contracts
avoid duplication. Every function in the four schema source files meets 15 lines.

A disposable old/new source comparison proves all 28 tool schema JSON documents
and the unknown-tool panic match exactly. Existing shuffled MCP/architecture race
tests, lint/vet and whitespace checks pass; tests are unchanged. The retained
Enola baseline is checked once before committing. The strict audit remains open
at 243 production bodies.


## Round 18 — World initialization and actuation workflows

World creation now names identity, state/runtime construction, initial entities
and churn. Effector invocation names admission, idempotent replay, terminal
interlock refusal, outcome application and authority-log caching. Each function
in the three modified source files meets 15 lines. Mode selection and latency
draws keep their order; physical/partial/shadow effects retain their semantics.

Existing shuffled world, scoring and architecture race tests pass, including
effect-order and replay/argument-admission regressions. Package lint/vet and
whitespace checks pass; test files remain unchanged. The retained Enola baseline
is verified once before committing. The strict audit remains open at 240 bodies.

## Round 19 — Domain loading and admission

Domain parsing names document validation, decoding, digesting and compilation.
Symbol registration shares duplicate detection while retaining declaration order
and channel defaults. Per-record validation names reference, fidelity, detector,
interlock and delay checks without changing error priority. Dynamics traversal
and catalog coverage now expose their steps. Every domain-package production
function meets 15 lines; all changed files remain below 300 lines.

Existing shuffled domain, world, adapter and architecture race tests pass;
domain lint/vet and whitespace checks pass. Test files remain unchanged. The
retained Enola baseline is verified once before committing. The strict repository
audit remains open at 226 production bodies. High-impact workflows and repeated
validation take priority; full-repository verification remains a final gate.

## Round 20 — Shared scoring workflows and publication handoff

Online/offline scorecards share construction. Admission, action/dropped-event
matching, evidence grounding, ledger accounting, judgment and recovery scoring
now name their concrete steps. History lookup shares latest-sample selection.
All score-package production functions meet 15 lines. Matching remains greedy;
existing recovery scope and zero-time semantics are preserved for separate,
regression-proven corrections. Test files remain unchanged.

Focused scoring/architecture race tests and score lint/vet pass. The full CI gate and shuffled race/coverage suite pass (69.4% total coverage);
the final retained-baseline Enola check is clean and comparable.
The strict audit remains open at 211 bodies. HANDOFF.md records the package
inventory, source-confirmed improvement candidates and remaining CI integration.


## Round 21 — Publication evidence

Record the final full-gate results and package-level remaining work in HANDOFF.md.
The strict 15-line migration is incomplete at 211 production bodies, despite the
passing existing CI gate. Disposable public-API probes reproduce adapter
hash-suffix bounds and authority-log argument mutation. Other scan findings are
labeled source-confirmed risks. No production or test source changes in this
round; whitespace checks suffice for the evidence update. Preserve the concurrent
publication commit and its included architecture-review documents. Publish the
report through the existing PR rather than creating a duplicate.

## Round 22 — Adapter projection and conformance

Complete the adapter package's 15-line migration. Rendering names context,
guards, encoding and individual fields; transform evaluation separates concrete
operations. Validation and conformance name templates, identity, schemas,
fixtures and golden comparison. Ordered JSON output, null handling, expression
errors and byte-divergence positions retain their semantics. All adapter
production functions meet 15 lines; test files remain unchanged.

Existing adapter and run replay/conformance race tests, architecture guards,
adapter lint/vet and whitespace checks pass. One retained-baseline Enola check
precedes the commit. The strict audit remains open at 193 production bodies.

## Round 23 — Foundations and sealed observability

Complete model, randutil, sink and truth production functions to 15 lines.
Decoding separates trailing-document checks and coefficient defaults. Weighted
sampling retains insertion sorting and total accumulation order. Sink helpers
retain locking, close/error ordering and HTTP body ownership. The truth solver
names oracle-world preparation, threshold scanning and detector arithmetic;
label construction names identity, observability and scenario context.

Existing shuffled foundation/sink/truth race tests and architecture guards pass,
including sealed-label and analytic regressions. Package lint/vet and whitespace
checks pass; tests are unchanged. The retained Enola baseline is checked once
before committing. The strict audit remains open at 181 production bodies.

## Round 24 — Seeded delivery perturbations

Complete perturbation functions and callbacks to 15 lines. Activation,
parameter admission, pending drops, reorder displacement and buffered recovery
are named steps. Transform routing follows multiplicity, delivery, timing and
payload responsibilities. Record helpers share duplicate IDs and numeric
parameter conversion while preserving seeded draws, delivery order and flags.
Reorder displacement avoids allocating an intermediate slice per pair.

Existing shuffled perturbation/run race and architecture tests, package lint/vet
and whitespace checks pass; test files are unchanged. One retained-baseline
Enola check precedes the commit. The strict audit remains open at 166 bodies.

## Round 25 — Run commands, delivery and replay

Complete run-package functions to 15 lines. Command logging shares sequenced
record construction and records refusals exactly once. Delivery steps retain
native versus perturbed timestamps and evidence-hook placement. Quiescence
retains watermark locking, wakeups, cancellation and timeout semantics. Artifact
assembly names identity, inputs, counts and state; replay names admission,
configuration, commands, completion and divergence. Existing cleanup/publication
and serialization limitations remain documented, without behavioral correction.

Existing shuffled run, scoring and architecture race tests pass, including replay
identity and quiescence/failure boundaries. Package lint/vet and whitespace checks
pass; tests are unchanged. The retained Enola baseline is checked once before
committing. The strict audit remains open at 148 production bodies.

## Round 26 — Device world bindings and safe stops

Complete deviceworld production functions to 15 lines. Binding admission names
source compilation, entity requirements and catalog checks. Plant commands name
world advancement, argument resolution, invocation and effect completion. Safe
stops retain explicit admission, interlock errors and applied-effect requirements.

Existing shuffled deviceworld/device race tests and architecture guards pass.
Package lint/vet and whitespace checks pass; tests are unchanged. Enola is clean,
comparable and adds no findings against the retained baseline. The strict audit
remains open at 140 production bodies.

## Round 27 — Reference-consumer observation and reporting

Complete reference-consumer functions to 15 lines. Separate delivery admission,
series tracking, anomaly detection, actuation, silence detection and verdict
publication. MCP decoding names refusal and structured-content admission;
nameplate projection preserves ordered append and nil-slice representations.

Existing shuffled reference-consumer race and architecture tests, package lint/vet
and whitespace checks pass. Test files are unchanged. Enola is clean and
comparable with no added findings. The strict audit remains open at 132 bodies.

## Round 28 — Trivial-baseline audit evidence and fitting

Complete audit production functions to 15 lines. World preparation and delivered
evidence capture lead into sampling and detector grading. Threshold, difference,
z-score, median and silence detectors share balanced-accuracy pooling. Preserve
quantile grids, numerical accumulation, strict comparisons and tie priority.

Existing shuffled audit race tests and architecture guards, lint/vet and whitespace
checks pass. Tests remain unchanged. Enola is clean and comparable with no added
findings. The strict audit remains open at 120 production bodies.

## Round 29 — Scenario generation, admission and replay commands

Complete suite production functions to 15 lines. Name generation defaults,
attempts, sampling, composition accounting, label grading and command logging.
Preserve random draw order, weighted accumulation order, failed-attempt identities,
trivial exclusions and replay-command sequence. Existing profile data remains
the source of setup; no domain-specific branch is introduced.

Existing focused generation regression, audit and architecture race tests pass;
complete expensive suite tests are reserved for the final repository gate.
Package lint/vet and whitespace checks pass; test files are unchanged. Enola is
clean and comparable with no added findings. The strict audit has 108 bodies left.

## Round 30 — Run, replay, scoring and suite command workflows

Complete command orchestration, run option parsing, script execution and evidence
loading functions to 15 lines. Shared simulator-input loading retains catalog,
adapter, domain and adapter-selection error priority. Preserve flags, defaults,
CLI output shapes, ignored artifact-load behavior and empty-ledger representation.

Existing shuffled CLI race and architecture tests, package lint/vet and whitespace
checks pass. Tests remain unchanged. Enola is clean and comparable with no added
findings. The strict audit remains open at 98 production bodies.

## Round 31 — Device, manifest and MCP command wiring

Complete all CLI production functions to 15 lines. Name device options,
capability loading, world selection and safe-stop binding requirements; separate
listener shutdown from startup. Release manifests name options, file digests,
identity, signing and publication. MCP startup and reference-consumer connection
retain endpoint lifecycle and close/error order. Flags and output shapes persist.

Existing shuffled CLI, deviceworld, reference-consumer and architecture race tests
pass; lint/vet and whitespace checks pass. Tests are unchanged. Enola is clean
and comparable with no added findings. The strict audit has 89 bodies left.
