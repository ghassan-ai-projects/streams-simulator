# Streams Simulator implementation-readiness review

Status: **Level 1 replayable prototype; not benchmark-ready**
Review date: 2026-08-11
Initial reviewed commit: `63b1f1c`
Correctness-hardening rebaseline: `943083c`
Scope: repository design, contracts, implementation, shipped domains and adapters, tests,
build automation, and representative end-to-end executions

## 1. Executive verdict

This document is the initial gap inventory. The post-fix status and remaining benchmark
blockers are maintained in [IMPLEMENTATION_REBASELINE.md](IMPLEMENTATION_REBASELINE.md).

The project has a strong design premise, unusually clear architectural intent, and a
substantial S0-S5 prototype. It is not yet a trustworthy test instrument.

The most important problem is not missing feature breadth. Several implemented paths can
produce plausible, deterministic-looking output while violating the world model, delivery
contract, replay contract, or scoring semantics. Those are instrument-integrity failures:
they can produce a confident but wrong evaluation of a consumer.

Feature work should pause until the correctness spine is repaired. In particular, do not
publish benchmark results or describe the nine non-negotiables as implemented. The current
code is suitable for continued development and design validation, not for judging another
system.

The recommended sequence is:

1. Re-baseline the documented status and convert the non-negotiables into release gates.
2. Fix entity isolation, time/model correctness, exact numeric representations, and churn.
3. Unify the real streaming path with adapter, ledger, perturbation, and replay contracts.
4. Rebuild the oracle, audit, and score calculations on immutable time-ordered evidence.
5. Complete the operator lifecycle and remove domain-specific suite logic.
6. Add adversarial, per-domain, performance, and end-to-end consumer evidence.

## 2. Review method and evidence

This was a code and design review, not a review based only on the existing test results.
The following evidence was used:

- all repository instructions and context under `.agents/context/`
- the design, research, contracts, examples, and decision records under `docs/`
- all production Go packages, tests, the CLI, Makefile, and CI workflows
- all six shipped domains and both shipped output adapters
- `make ci-check`, including lint, unit, race, and simulator-determinism runs
- representative CLI run, replay, verify, scoring, and MCP-role smoke tests
- inspection of emitted trace, ledger, history, and run artifacts

`make ci-check` passed in the review environment. `deadcode` and `govulncheck` were not
installed locally, however, and the Makefile skipped them while still printing that the CI
check passed. A separate coverage run could not start because the local Go installation was
missing its `covdata` tool. These limitations are recorded below; they are not counted as
product failures by themselves.

The review confirmed these concrete runtime results:

- A nominal 120-second run emitted 148 records and replayed to the same trace.
- The run artifact had no usable `world_digest`.
- An `agentic-stream` trace omitted its declared preamble and postamble.
- A matching `verify` result reported `first_divergence: 0` rather than no divergence.
- The default CLI file sink failed because its target was empty.
- `score --verdict` was advertised but rejected as an unknown flag.
- `mcp --role operator` reported that the role is served per-world, but there is no CLI or
  live-server path that actually serves it.

## 3. What is already good

The project is worth repairing. Its strongest properties are:

- The product boundary is clear: domains, adapters, and effectors are intended to be data.
- The package layout is readable and mostly follows the documented dependency direction.
- Contracts are versioned, embedded, and checked against their documentation copies.
- The core has a deliberately small dependency surface.
- The PRNG is abstracted and named substreams are used in many important places.
- Nominal runs are reproducible in the tested path.
- The first-order analytic cross-check is a real independent check, not just a snapshot.
- Adapter golden tests, race tests, schema checks, and deterministic test wrappers exist.
- Six varied domains load and validate, providing useful pressure on a generic design.
- The documentation explicitly recognizes that a faulty instrument is worse than none.

These strengths reduce the redesign needed. The architectural concepts can remain; the
runtime semantics and proof need to catch up to them.

## 4. Non-negotiable gate assessment

| Gate | Result | Evidence and reason |
| --- | --- | --- |
| Determinism | **Partial / fail** | A simple nominal trace repeats, but the world digest is empty, exact integers are converted through `float64`, suite selection iterates a map, some RNG streams omit entity identity, and quiescence uses wall time. |
| Analytic cross-check | **Partial** | `first_order` is cross-checked. `rc_network` ignores a declared coefficient, and other dynamic forms have no independent oracle. |
| Reference consumer | **Partial / fail** | An offline native-event observer exists. It is not a live operator client and does not demonstrate the declared closed loop or terminal-silence semantics. |
| Delivery ledger | **Partial / fail** | Basic emitted/drop/duplicate rows exist. Adapter omissions, buffered clear, pre/post records, several perturbations, and delivery-instance identity are incomplete or absent. History is silently capped. |
| Quiescence barrier | **Fail** | It polls unsynchronized state with wall-clock sleep and lacks a working live operator topology. The existing test does not establish the real concurrent protocol. |
| Sealed oracle | **Fail** | Normal run lifecycle does not seal truth, unblind state is not tracked correctly, and mutable pointers remain exposed. |
| Injection probe | **Partial / fail** | A synthetic reference-consumer test exists, but it proves only that that consumer ignores a string. Flush paths bypass evidence capture, and shipped domains/real consumers are not covered. |
| `silent_no_effect` | **Partial** | A direct test exists, but shared entity state and weak action scoring prevent it from proving end-to-end absence of effect. |
| Trivial-baseline audit | **Fail** | The audit advances the mutable world and then reads it backward in time, mixes entities, truncates the event horizon, and counts rejected candidates. Its result is not a reliable baseline audit. |

Release policy should treat every row as binary: a gate is complete only when an adversarial
test exercises the production path and fails after the protected invariant is deliberately
broken.

## 5. Findings

Severity means:

- **P0** — blocks trustworthy use of the simulator or its scores.
- **P1** — must be resolved before a production-quality release.
- **P2** — maintainability, usability, documentation, or hardening work.

### IR-001 — P0: dynamic effects are shared across entities

**Evidence.** `internal/world.World` stores kicks and shadow actions in maps keyed only by
state name. A kick has no entity identity. `stateAt` therefore applies a state kick to every
entity that has a state with the same name. Autonomous actions use the same pattern.

**Impact.** Actuating entity A can change entity B. Scores can credit an action for a
recovery that happened to a different asset. This violates the closed-loop premise and
entity-isolated PRNG rule.

**Required change.** Key all mutable state, effects, caches, emission state, and relevant RNG
substreams by stable entity ID plus state/channel identity. Make cross-entity effects
explicit data in a domain specification rather than an accidental shared key.

**Acceptance evidence.** A two-entity test applies opposite effects at the same timestamp
and proves isolated analytic trajectories, isolated observations, and stable results when a
third entity is added.

### IR-002 — P0: entity churn collides and then stops

**Evidence.** Initial entities use `{n}` values beginning at one while the runtime
`nextIndex` starts at zero. The first birth increments to one, collides with the first
entity, and returns on `addEntity` error without scheduling another birth.

**Impact.** Domains can appear to support churn while never producing a valid birth. A
headline Kubernetes-domain behavior is therefore not exercised.

**Required change.** Initialize the allocator from existing identities, make allocation
collision-proof, and schedule the next lifecycle event even when one transition is rejected.
Record lifecycle transitions in the artifact and ledger where required for replay.

**Acceptance evidence.** Long-running tests prove multiple births/deaths, uniqueness,
deterministic replay, and unaffected pre-existing entity streams.

### IR-003 — P0: the world digest is empty and errors are discarded

**Evidence.** `worldDigest` passes `[]string` to a canonical encoder that accepts canonical
arrays as `[]any`. The marshal error is ignored and an empty string is returned. A real run
artifact consequently omitted or null-valued `world_digest`.

The digest construction also converts the `uint64` seed through `float64` and does not
clearly cover all inputs in the documented determinism identity.

**Impact.** Artifacts cannot prove which world they represent. Cache keys and replay
comparisons can silently collide. The core reproducibility claim is ungrounded.

**Required change.** Define one typed, versioned digest input; fail the run if canonicalization
fails; never convert exact seeds/timestamps through `float64`; and test every field included
or intentionally excluded. Use strings or another lossless representation where the JSON
number model cannot hold the value.

**Acceptance evidence.** Golden digest vectors cover boundary `uint64` seeds, nanosecond
timestamps, reordered maps, changed domain/adapter bytes, and every run parameter. No digest
function may return a value while discarding an error.

### IR-004 — P0: the production adapter path violates adapter framing

**Evidence.** A run calls `RenderRecord` for individual events and never invokes the adapter
preamble/postamble lifecycle or supplies final metadata. A smoke trace using the shipped
`agentic-stream` adapter contained only event rows. Its declared `runtime_config` and
`trace_end` records were absent. The `json-array` encoding similarly has no real streaming
array lifecycle. Golden tests exercise `RenderRun`, not the production run path.

**Impact.** A consumer receives a different protocol than conformance tests certify.
Terminal metadata, completeness assertions, and array validity are lost.

**Required change.** Create one adapter session API with explicit begin, record, and end
states. Both batch conformance and live runs must use it. Define failure behavior for a
partially written stream and represent an incomplete end state.

**Acceptance evidence.** Every shipped adapter's actual file and HTTP output validates as
its declared encoding, byte-matches its golden lifecycle, and includes correct terminal
counts/digest. Add zero-record and mid-stream-failure cases.

### IR-005 — P0: runtime output errors can panic the process

**Evidence.** The run failure helper panics on adapter or sink failures. Malformed-record
writes discard sink errors.

**Impact.** A single bad sink can crash a long-lived MCP server and lose the exact failure
artifact needed to diagnose the consumer. A panic is also different from the designed
“abort and mark incomplete” behavior.

**Required change.** Propagate contextual errors through the run state machine, close the
adapter/sink best-effort, write an incomplete artifact when possible, and make the MCP/CLI
surface return a structured failure. Never ignore a sink error.

**Acceptance evidence.** Fault-injecting sinks fail on begin, middle, flush, and close;
tests prove no panic, a single terminal state, accurate ledger/artifact status, and no
subsequent writes.

### IR-006 — P0: the artifact is not self-contained and replay is not authenticated

**Evidence.** The artifact records domain/adapter references and digests rather than the
canonical source material needed to reconstruct them. Replay depends on the current
external catalog and does not first reject a supplied domain or adapter whose digest differs
from the artifact. The original trace is not retained, so `firstDivergentRecord` compares
counts rather than records. Exact command arguments and nanosecond values pass through
`float64` in command-log maps.

**Impact.** “One artifact reproduces any failure” is false after a catalog changes. Replay
can use different semantics while appearing successful, and it cannot identify the first
real divergence.

**Required change.** Either embed canonical domain/adapter inputs in the artifact or use a
content-addressed immutable bundle included with it. Verify all digests before executing.
Store enough ordered output hashes or records to locate a divergence. Replace untyped
numeric maps with versioned typed commands.

**Acceptance evidence.** A run replays in an empty directory, rejects one-byte spec changes,
preserves max seeds and nanosecond timestamps exactly, and reports the first differing record
or “none” for an exact match.

### IR-007 — P0: the declared operator MCP lifecycle does not exist end to end

**Evidence.** Operator tools can be constructed in tests, but `streamsim mcp --role
operator` only returns an error saying the role is per-world. The director stdio process has
no mechanism that exposes a per-world operator server to a consumer. `BeginRun` only toggles
a Boolean and ignores its label; it does not establish the complete lifecycle or seal truth.

**Impact.** The primary closed-loop protocol described in the design and README cannot be
used by an external consumer. Quiescence, effector invocation, and verdict collection are
therefore not validated in their intended topology.

**Required change.** Decide and document one implementable topology: a single server with
capability-scoped sessions, a separately bound per-world server, or another concrete design.
Specify endpoint discovery, authentication, lifecycle states, reconnect behavior, and
shutdown. Then test it with an out-of-process consumer.

**Acceptance evidence.** A black-box test starts the director, creates a world, obtains only
operator capabilities, invokes an effector, participates in quiescence, submits a verdict,
and proves truth/director tools are inaccessible.

### IR-008 — P0: truth sealing and unblinding are not reliable state transitions

**Evidence.** Normal begin/run paths do not seal truth automatically. `Store.SealStatus`
always reports unblinded false, the store does not own an explicit unblind transition, and
sealed records remain reachable through mutable pointers.

**Impact.** A harness can observe or modify oracle data at the wrong time, and audit logs
cannot establish that scoring was blind.

**Required change.** Introduce a monotonic lifecycle such as created → sealed → running →
quiescent → scored → unblinded. Store immutable snapshots or defensive copies. Put every
read behind capability and state checks and append an auditable transition record.

**Acceptance evidence.** State-machine tests reject all illegal transitions and concurrent
reads. Prefix-equivalence tests cover every evidence path, buffered delivery, and at least
one real external consumer.

### IR-009 — P0: scoring reports unevaluated qualities as successful

**Evidence.** `LedgerComplete` is set true without proving unique sequence coverage.
`IdentityConflict` and `ClockSkewRejection` are hard-coded true. Lateness can pass without an
explicit admission decision. Deadline adherence is not compared with a deadline.

**Impact.** Missing evidence becomes a passing score. This is the most dangerous possible
failure mode for a test instrument.

**Required change.** Every metric needs an explicit numerator, denominator, evidence source,
and state: pass, fail, or not evaluated. No default success is allowed. Validate evidence
completeness before computing derived metrics.

**Acceptance evidence.** Mutation tests remove or corrupt each required evidence item and
prove the score becomes fail/not-evaluated, never pass. Publish a versioned scorecard JSON
schema.

### IR-010 — P0: detection and action matching can award unrelated behavior

**Evidence.** One detection near one dropped record can satisfy all dropped-event
detections. Verdicts are keyed by sequence number even though original and duplicate
deliveries share it. Action fidelity compares essentially the command ID, not the complete
effector/entity/time/argument/outcome tuple. An empty detection label can count as correct.
The earliest-judgment logic compares an absolute timestamp with a latency duration, making
input order affect the winner.

**Impact.** A vague, duplicated, late, or wrong-entity response can receive full credit.

**Required change.** Define one-to-one matching with delivery-instance IDs, typed detection
and action identities, stable tie-breaking, and explicit label semantics. Keep event time,
observed time, report time, and duration in distinct types.

**Acceptance evidence.** Adversarial tables cover many-to-one, one-to-many, duplicate,
wrong-entity, wrong-effector, empty-label, reordered-verdict, and tie cases.

### IR-011 — P0: recovery and baseline scoring read a mutable world backward

**Evidence.** Scoring first advances/integrates a world and later asks that same mutable
world for earlier values. The dynamics and noise streams are forward stateful and cannot
rewind. Negative-deviation recovery also uses a formula whose direction is incorrect.

**Impact.** Baseline, response, and recovery values may describe a different trajectory
than the delivered run. Scores can change depending on call order.

**Required change.** Score from immutable, monotonic ground-truth samples captured during
the run, or from a pure random-access analytic evaluator with counter-based noise. Do not
reuse an advanced mutable world for historical queries. Define recovery for both deviation
directions mathematically.

**Acceptance evidence.** Score results are identical under repeated and reordered queries;
analytic positive and negative recovery fixtures have known expected values.

### IR-012 — P0: the trivial-baseline audit is not evaluating the generated stream

**Evidence.** The audit advances the world to a horizon and then samples it from the start,
which is a backward query. Its emission log is keyed only by channel and mixes entities. It
stops logging after a fixed horizon while later sample loops may continue. It does not
evaluate the actual perturbed/adapted delivery seen by the consumer.

Rejected candidates also update suite coverage counters before admission.

**Impact.** A candidate can pass the non-triviality gate for invalid reasons, and suite
coverage reports can include scenarios that do not exist.

**Required change.** Run baseline consumers over the exact immutable delivered evidence in
monotonic time, separately per entity/channel. Define horizon and terminal-silence behavior.
Only admitted scenarios may affect suite accounting.

**Acceptance evidence.** Known trivial and non-trivial fixtures pass/fail independently of
query order; rejected candidates leave all counters unchanged; audit results reproduce from
the artifact alone.

### IR-013 — P0: the suite generator contains domain-specific code

**Evidence.** `internal/suite` branches on `aquaculture-pond` and `start_aerator`.
`Director.AuditScenario` constructs `site-a/pond-%d` identities. These are domain and
effector names embedded in the binary.

**Impact.** The genericity invariant is already broken. A seventh domain may require another
code branch, while existing genericity tests remain green.

**Required change.** Put admissible setup actions and audit identity selection in the
versioned domain/profile schema. The binary should interpret generic selectors and
effector declarations only. Add a repository scan/lint for shipped domain and effector
literals outside data and tests.

**Acceptance evidence.** Rename every shipped domain/effector in data and prove the binary
needs no source change. Add a novel fixture domain that exercises setup and audit.

### IR-014 — P0: suite generation is not deterministically or accurately accounted

**Evidence.** Weighted fault selection iterates a Go map without sorting. Rejected attempts
reuse an admitted-count-derived scenario index/seed. Pre-degraded, negative, profile, and
perturbation counters can be incremented before the scenario passes admission. Several
percentages are hard-coded, and per-domain minimums/defaults are ignored.

**Impact.** The same inputs can yield a different suite; coverage summaries can claim
requirements that admitted scenarios do not satisfy; rejected attempts can collide in ID
and seed.

**Required change.** Sort all weighted choices, separate attempt identity from admitted
index, make candidate construction transactional, and derive constraints from one declared
profile. Validate terminal counts from admitted scenarios only.

**Acceptance evidence.** Repeated generation across processes is byte-identical; forced
rejection cannot change final counters; scenario IDs/seeds are unique; every declared
constraint has a failing boundary test.

### IR-015 — P0: declared world models are only partially implemented

**Evidence.** `rc_network` ignores `time_constant_2_s`. `batch_size` is not used. Declared
pink and quantization noise collapse into Gaussian handling. Boolean, counter, and string
channels are not emitted with faithful types. Fault onset magnitude is unused. Availability
uses fixed durations rather than the declared stochastic renewal behavior.

**Impact.** A valid specification can load while the runtime silently simulates different
physics or data types.

**Required change.** Make validation reject every unsupported declared option until it is
implemented. Then add an independent oracle per dynamic/noise/value form. Prefer a smaller
truthful schema over aspirational options.

**Acceptance evidence.** A contract-to-runtime matrix has one conformance test per enum and
field. Removing any field's implementation must fail a behavioral test, not only decoding.

### IR-016 — P0: dynamics admit dangerous and ambiguous specifications

**Evidence.** First-order dependency cycles are not rejected and can recurse indefinitely.
A stochastic fault walk starts at absolute epoch step zero, so a first query at a modern
Unix timestamp can execute tens of millions of iterations. Zero-valued gains/coefficients
are treated as missing defaults, making valid zero behavior impossible. One-sided zero
bounds are not representable reliably with scalar fields. Fault severity accepts invalid
ranges/types in some paths.

**Impact.** Valid-looking data can cause a denial of service, stack exhaustion, or silently
different equations.

**Required change.** Validate the dependency graph, bound all iterative work, anchor random
walks to world start, represent optional numbers with presence-aware types, and validate
severity/parameter ranges before mutating a run.

**Acceptance evidence.** Cycle, far-future timestamp, zero coefficient, one-sided bound,
wrong JSON numeric type, negative severity, and extreme-size tests all fail fast or produce
the specified bounded result.

### IR-017 — P0: perturbation outcomes are not faithfully represented

**Evidence.** Multiple perturbations leave the ledger reason as `ok`, including cases where
time or ordering changed. Reorder, clock-skew, non-monotonic-time, timestamp-encoding,
precision, and injection outcomes are not consistently distinguishable. Producer flap
ignores its period and buffered records flush on `Advance`, not their declared window.
Clearing perturbations can discard buffers without ledger evidence. Some configured cases
are no-ops.

**Impact.** The scorer cannot tell what the consumer actually received or why. Delivery
fidelity, rejection, and probe results become unverifiable.

**Required change.** Define a typed perturbation outcome per delivery instance: transforms,
hold/release times, drop reason, source instance, and final wire identity. Give every
buffering perturbation an explicit scheduling/close policy and ledger all discarded items.
Reject unknown or invalid parameters.

**Acceptance evidence.** A conformance table covers every perturbation alone and in declared
compositions, including boundary timestamps, clear, close, and replay.

### IR-018 — P0: the delivery record is neither complete nor instance-addressable

**Evidence.** Adapter-guard omissions can produce no ledger row. Flushed deliveries bypass
the evidence recorder used by the probe harness. History is silently capped at 10,000
records despite the “every emission” requirement. `WrittenAtNS` stores an ordinal-like value
rather than a nanosecond timestamp. Original and duplicate deliveries lack distinct stable
IDs.

**Impact.** Transport misses and reasoning misses cannot always be separated. Large runs
lose evidence silently, and duplicate-specific verdicts cannot be reconciled.

**Required change.** Establish a monotonic delivery-instance ID and a state transition for
every native emission through transform, hold, omit/drop, render, write, and error. Stream
the ledger to durable storage rather than truncating it in memory.

**Acceptance evidence.** Conservation tests reconcile every native emission and derived
delivery instance for runs larger than 10,000, with adapter omissions, duplicates, buffers,
and failures.

### IR-019 — P0: quiescence is a polling approximation with concurrency risk

**Evidence.** The barrier reads mutable pending state while another goroutine may report it,
using repeated `time.Now`/`time.Sleep` for up to 30 seconds. The synchronization contract is
not explicit. Deterministic wrappers do not remove this wall clock. Existing tests can
report before the wait and therefore do not prove the blocking interaction.

**Impact.** Race behavior and scheduler timing can change whether a run is scored, while a
simulated-time instrument waits on real time. A missing consumer costs 30 seconds per run.

**Required change.** Use a mutex-protected generation/ack state machine with a condition or
channel and injectable timeout clock. Specify reconnect, duplicate ack, late ack, cancel,
and consumer-death behavior.

**Acceptance evidence.** Race-enabled concurrent tests cover ack-before-wait, ack-during-
wait, duplicate, stale generation, timeout, cancel, and shutdown without real sleeps.

### IR-020 — P0: the reference consumer does not prove the advertised closed loop

**Evidence.** The reference consumer processes offline native JSONL and has no live operator
endpoint. It does not use the declared end time to detect terminal silence, fully report
admission outcomes, generate all required effector arguments, or participate in quiescence.
Its “records seen” value counts series rather than records in some paths.

**Impact.** The project lacks the minimum external witness that the protocol, evidence,
actions, and scoring work together.

**Required change.** Build a deliberately simple out-of-process reference consumer against
the real operator protocol. It must handle the shipped adapter, event-time/admission policy,
absence through end-of-run, generic effector schemas, idempotency, and quiescence.

**Acceptance evidence.** It completes a closed-loop golden run for every shipped domain,
while deliberate protocol/model mutations make the expected tests fail.

### IR-021 — P1: world identity, capability tokens, and lifecycle keys are weak

**Evidence.** Creating the same domain/seed can derive the same registry key and overwrite a
world. The externally exposed world ID and the internal world ID can differ. Capability
tokens are predictable strings derived from sequence/seed.

**Impact.** Concurrent/repeated runs can alias, responses can name inconsistent worlds, and
tokens are not suitable once an operator endpoint is reachable outside a trusted process.

**Required change.** Separate immutable run identity, world identity, scenario identity, and
human label. Generate collision-resistant opaque capability tokens from a cryptographic RNG
and store only hashes. Make registry collision an error.

**Acceptance evidence.** Repeated seeds create distinct requested runs without overwrite;
all surfaces return one canonical identity; tokens are unguessable and revocable.

### IR-022 — P1: MCP requests are weakly typed and discard request context

**Evidence.** Tools do not expose meaningful input schemas, arguments silently default, and
the shared registration helper invokes handlers with `context.Background()` rather than the
request context.

**Impact.** Clients cannot discover the protocol, malformed requests may change behavior
silently, and cancellation/deadlines do not propagate to I/O or long runs.

**Required change.** Publish strict schemas generated from typed request structs, reject
unknown/missing values, and pass the request context through every handler.

**Acceptance evidence.** Protocol tests validate discovery schemas, malformed inputs,
cancellation, deadlines, and stable structured error codes.

### IR-023 — P1: sink configuration is incomplete and failure semantics differ from design

**Evidence.** Director-created worlds do not provide a sink target, so file and HTTP-push
sinks are unusable through MCP. Environment injection targets are never configured. The
HTTP sink has no documented retry/retention behavior and fails immediately. The CLI defaults
to a file sink with an empty target.

**Impact.** Only the in-process path is practically usable in the control plane, and the
same transient failure behaves differently from the design.

**Required change.** Make sink configuration a validated typed object with explicit target,
timeouts, retry policy, redirect/network policy, and artifact behavior. Pick a CLI default
that works without hidden arguments.

**Acceptance evidence.** Black-box tests exercise every sink, retries and permanent errors,
and confirm byte/ledger equivalence where transport semantics permit it.

### IR-024 — P1: canonical JSON and schema numeric semantics conflict

**Evidence.** The canonical encoder preserves arbitrary integer `json.Number` text while
the claimed RFC 8785/JCS data model is based on I-JSON/IEEE-754 semantics. Other schema paths
convert integer values through `float64`/`int64`, while artifacts allow full `uint64` seeds.

**Impact.** Two components can disagree on validation or digest bytes for large values.

**Required change.** Choose one explicit numeric contract. Either conform strictly to JCS
and encode out-of-range exact integers as strings, or define and version a documented
extended canonical format without calling it RFC 8785.

**Acceptance evidence.** Cross-language vectors cover ±2^53 boundaries, max `uint64`,
exponents, negative zero, decimals, and rejected non-I-JSON values.

### IR-025 — P1: decoders and recursive adapter values need resource limits

**Evidence.** Some trailing-input checks use decoder state in a way that may not prove EOF.
Adapter values are recursively evaluated without a clear instance-depth/size budget. URI,
hostname, and external conformance-file handling are minimally constrained.

**Impact.** Untrusted or accidentally pathological data can consume excessive CPU/stack or
smuggle trailing content. Future remote control surfaces could widen the exposure.

**Required change.** Decode exactly one value then require EOF; impose bytes, nesting,
collection, template, and output limits; validate external locations at the trust boundary.

**Acceptance evidence.** Fuzz and boundary tests cover trailing JSON, deep nesting, huge
arrays/strings, recursive templates, redirect targets, and local-file scope.

### IR-026 — P1: RNG isolation is incomplete

**Evidence.** Some link-delay and fault RNG substream keys omit entity or fault-instance
identity. Availability currently does not consume its declared renewal RNG. Adding a
neighboring entity/fault can therefore perturb another series or leave a stochastic model
deterministic for the wrong reason.

**Impact.** The isolation guarantee—add one entity without changing existing entities—is
not consistently true.

**Required change.** Document a canonical substream-key hierarchy and require all stochastic
sites to include the full stable scope. Add a static helper so callers cannot assemble keys
ad hoc.

**Acceptance evidence.** Metamorphic tests add/reorder unrelated entities, channels, faults,
and perturbations and prove all pre-existing scoped outputs stay byte-identical.

### IR-027 — P1: command and result representations are not durable contracts

**Evidence.** Effector arguments are loose maps; the CLI advertises arguments but does not
faithfully parse/pass them. A value named `ResultDigest` contains serialized JSON rather
than a cryptographic digest. Command IDs include wall-clock-derived values in some paths.

**Impact.** Commands can lose type/precision, artifacts vary for non-simulation reasons, and
a field name promises integrity it does not provide.

**Required change.** Version typed command/result envelopes, generate deterministic IDs from
run-local sequence or accept caller idempotency IDs, and compute an actual digest over a
canonical result.

**Acceptance evidence.** Round-trip tests cover every JSON type, large integers, duplicate
idempotency keys, CLI parity, and digest tampering.

### IR-028 — P1: offline scoring cannot reconstruct the evidence it claims to score

**Evidence.** The scoring CLI attempts to unmarshal a JSONL ledger as one JSON array and
does not load a complete call/action/truth bundle. Its help advertises a `--verdict` flag
that is not registered.

**Impact.** The offline audit path is unusable or incomplete, and CLI documentation cannot
be trusted as an executable contract.

**Required change.** Define a versioned scoring bundle containing truth, delivery instances,
verdicts, commands/results, configuration, and digests. Reuse the online scoring engine and
make CLI help generated from actual flags.

**Acceptance evidence.** An online run exported to disk scores byte-identically offline;
missing/tampered components are rejected with specific errors.

### IR-029 — P1: per-domain behavioral truth is missing

**Evidence.** Shipped domains are tested mainly for loading, profiles, and selected negative
cases. There are no independent golden invariants for unit semantics, state equations,
fault effects, churn, availability, and effector response per domain. For example, OEE
throughput units and integration deserve an explicit oracle rather than visual plausibility.

**Impact.** A domain can emit credible-looking but dimensionally wrong data and all generic
tests still pass.

**Required change.** Give every shipped domain a short deterministic validation scenario
with dimensional analysis and hand/independently computed checkpoints. Record units for
state derivatives and integrations explicitly.

**Acceptance evidence.** Each domain has nominal, fault, recovery, lifecycle, and boundary
goldens reviewed by someone other than the scenario author.

### IR-030 — P1: required owner and benchmark-governance decisions remain unresolved

**Evidence.** The implementation plan required decisions on initial domain count, real
customer/consumer selection, and scenario-author separation before later stages. Six domains
exist, but the customer target and independent-author process are not recorded. Suite
constants are not fully driven by a frozen, digested profile.

**Impact.** The project can optimize against internally invented expectations and publish a
benchmark whose author/evaluator independence is unclear.

**Required change.** Record the owner decisions explicitly. Name the validation consumer or
state that no benchmark claim is allowed yet. Define author/reviewer separation and freeze
all benchmark-affecting inputs in a signed/digested suite manifest.

**Acceptance evidence.** A release artifact identifies the exact suite manifest, authors,
reviewers, consumer contract, simulator version, and all content digests.

### IR-031 — P1: tests demonstrate happy paths more strongly than invariants

**Evidence.** There are no broad per-domain behavioral tests, load/soak tests, property tests,
or fuzz targets for critical parsers and state machines. The deterministic-source lint
misses relevant cross-package maps. Several non-negotiables are represented by tests that
exercise a synthetic or direct path rather than the production path.

**Impact.** Green automation overstates assurance. Major defects above coexist with passing
lint, race, and determinism checks.

**Required change.** Organize tests around claims and failure injection, not packages alone.
For every non-negotiable, add a mutation test or a deliberately faulty implementation
fixture showing that the gate detects the defect.

**Acceptance evidence.** The release checklist links each claim to a black-box or
independent-oracle test and includes fault-injection, fuzz, soak, and scale jobs.

### IR-032 — P1: performance and capacity claims are unmeasured

**Evidence.** The design names run sizes and timing behavior, but the repository contains no
benchmarks or enforced budgets for events/sec, memory per event, ledger size, replay time,
large entity catalogs, or adversarial specs.

**Impact.** Silent in-memory caps and quadratic/epoch-length work can reach production before
being visible in CI.

**Required change.** Define representative small/medium/maximum profiles and budgets. Measure
generation, adaptation, perturbation, durable ledger, replay, scoring, and MCP overhead.

**Acceptance evidence.** Reproducible benchmarks and one bounded soak test publish resource
curves and fail on material regression.

### IR-033 — P2: canonical project status contradicts the repository

**Evidence.** `AGENTS.md`, `.agents/context/project.md`, `.agents/context/architecture.md`,
and this documentation index describe a design-only repository with no `cmd/` or `internal/`
tree. The README says S0-S5 and all nine non-negotiables are implemented. The code and this
review support neither description. Design documents still contain superseded choices such
as SQLite in places, while decision records only partially reconcile the change.

**Impact.** New contributors follow false instructions and reviewers cannot tell which
document is authoritative.

**Required change.** After correctness scope is agreed, update all status documents in one
change. Distinguish normative contract, accepted decision, current implementation, and
future plan. Mark superseded text rather than leaving silent contradictions.

**Acceptance evidence.** A repository documentation check confirms versions/status/commands
agree, and every design deviation has a decision record.

### IR-034 — P2: Go and toolchain policy is inconsistent

**Evidence.** Repository surfaces refer to both Go 1.25.12 and Go 1.26. Local `ci-check`
skips absent `deadcode` and `govulncheck` yet reports success; CI installs them. Coverage is
not an enforced release signal.

**Impact.** “CI-equivalent” has different meanings locally and remotely, and contributors
can debug toolchain-specific failures that the project itself created.

**Required change.** Pin one Go/tool version source, install tools reproducibly, and make
required checks fail when unavailable. Provide a separate explicitly named best-effort
target if useful.

**Acceptance evidence.** A clean machine executes the documented command with the same
versions and checks as CI; coverage is collected for all changed production packages.

### IR-035 — P2: CLI behavior and help are not contract-tested

**Evidence.** The default run configuration fails, adapter listing can inherit map iteration
order, score flags disagree with help, operator role is advertised but unavailable, and
matching verification uses a misleading divergence sentinel.

**Impact.** The first user experience is failure and automation must special-case unstable
or ambiguous output.

**Required change.** Treat CLI JSON and exit codes as versioned interfaces. Choose working
defaults, sort output, use nullable/omitted divergence, and remove or implement advertised
commands.

**Acceptance evidence.** Golden black-box tests cover help, defaults, errors, JSON output,
exit codes, and every documented example.

### IR-036 — P2: security and trust-boundary documentation is incomplete

**Evidence.** The security policy delegates the private contact to an adopting project but
does not identify one. Domain/adapter inputs can drive templates, files, HTTP destinations,
and potentially large computations. The intended trust level of director callers, spec
authors, sink targets, and consumers is not consolidated in a threat model.

**Impact.** Safe local-development assumptions can accidentally become remote-service
assumptions when MCP/HTTP deployment is completed.

**Required change.** Write a concise threat model: actors, capabilities, trusted inputs,
network/file boundaries, resource limits, prompt-injection boundary, token lifecycle, and
vulnerability contact. Review HTTP redirect/DNS behavior and local conformance-file access.

**Acceptance evidence.** Security tests cover capability separation, path/network policies,
resource exhaustion, secret-free artifacts/logs, and injection evidence isolation.

## 6. Systemic root causes: a 5 Whys summary

### Why can the simulator produce a confident wrong score?

1. Because several metrics treat absent or ambiguous evidence as success.
2. Because evidence is not represented as a complete, immutable set of delivery instances
   and world-history samples.
3. Because each subsystem evolved its own convenient representation—maps of `any`, sequence
   numbers, mutable world queries, and local buffers.
4. Because tests mostly validate those subsystem implementations in isolation instead of
   validating end-to-end conservation and adversarial invariants.
5. Because stage completion was inferred from implemented components and green happy paths,
   rather than from falsifiable release gates for the nine non-negotiables.

### Why did genericity and determinism regress despite explicit design rules?

1. Because suite/audit features needed domain setup and identity selection.
2. Because those needs were resolved with local string branches and map iteration.
3. Because the schemas did not yet express setup policies and stable selection order.
4. Because the genericity/determinism checks inspect only a subset of source patterns and
   do not use metamorphic black-box tests.
5. Because design rules were documented as architecture principles, not enforced as
   executable constraints at the extension boundary.

### Why is documentation simultaneously “design-only” and “S0-S5 complete”?

1. Implementation advanced faster than canonical contributor context was updated.
2. The README marked milestones by feature presence.
3. Runtime deviations and incomplete proof were not recorded as stage blockers.
4. There is no single generated or reviewed implementation-status ledger.
5. Therefore different documents optimized for different moments and none remained a
   reliable current-state source.

## 7. Recommended implementation plan

### Phase 0 — Re-baseline and protect users

- Mark the product experimental and explicitly prohibit benchmark claims.
- Reconcile `AGENTS.md`, `.agents/context/`, README, docs status, Go version, and decisions.
- Turn every non-negotiable into a named required release gate.
- Add the missing scorecard and scoring-bundle contracts.
- Make required CI tools fail closed when unavailable.

**Exit criterion:** a contributor can determine the real state, supported behavior, and
release gates from one consistent documentation path.

### Phase 1 — Repair the deterministic world core

- Scope effects, state, caches, emissions, and RNG streams by entity.
- Fix churn allocation/scheduling and availability renewal.
- Reject unsupported model/noise/value features or implement them with independent oracles.
- Validate dependency cycles and bound all iterative stochastic work.
- Replace lossy untyped numbers with exact typed representations.
- Build a typed world digest and fail on canonicalization errors.

**Exit criterion:** multi-entity metamorphic tests, analytic domain checkpoints, extreme
numeric vectors, and long-horizon churn all pass and deliberately injected isolation/model
faults are caught.

### Phase 2 — Make delivery, adapters, and replay one audited pipeline

- Introduce one begin/record/end adapter session used by batch and streaming paths.
- Model every emission and delivery instance through a durable state-transition ledger.
- Give perturbations typed outcomes and correct scheduling/flush/clear behavior.
- Embed or content-address immutable replay inputs and verify them before execution.
- Store record hashes/evidence sufficient for a real first-divergence report.
- Replace panics and ignored errors with an incomplete-run state.

**Exit criterion:** conservation holds for every perturbation/adapter/sink composition,
actual outputs pass adapter goldens, and an artifact replays independently after catalogs
are removed.

### Phase 3 — Rebuild oracle, audit, and scoring semantics

- Capture immutable, monotonic truth and delivery evidence during the run.
- Rewrite baseline and recovery calculations without backward mutable-world queries.
- Define every metric's evidence, matching, denominator, and not-evaluated behavior.
- Persist all verdict/action/truth inputs in a versioned scoring bundle.
- Make online and offline scoring use the same pure engine.
- Add adversarial/mutation tests for each score.

**Exit criterion:** scores are order-independent, fail closed on missing evidence, match
hand-computed fixtures, and reproduce offline byte-for-byte.

### Phase 4 — Complete the closed-loop control plane

- Choose and document a deployable director/operator topology.
- Implement typed schemas, request-context propagation, canonical identities, and secure
  capability tokens.
- Implement the monotonic world/truth lifecycle and synchronized quiescence state machine.
- Make every sink usable through the chosen control plane.
- Build the external reference consumer against this surface.

**Exit criterion:** a black-box process test completes a generic action/recovery/scoring run
without importing internal packages and proves capability isolation.

### Phase 5 — Restore generic suite construction and benchmark governance

- Move setup/audit selectors into generic versioned data.
- Make all weighted selection ordered and all candidate admission transactional.
- Honor one frozen suite profile rather than hard-coded distributions.
- Give attempts and admitted scenarios distinct, collision-free identities.
- Record customer target, author separation, review, and manifest digests.

**Exit criterion:** a novel domain joins without Go changes; suite generation is
byte-identical across processes; only admitted scenarios satisfy reported coverage.

### Phase 6 — Establish release evidence

- Add per-domain independent behavioral goldens.
- Add perturbation composition, fuzz, property, scale, soak, and failure-injection jobs.
- Measure and enforce resource budgets.
- Complete the security threat model and boundary tests.
- Run at least one real target consumer without changing the simulator binary.

**Exit criterion:** all nine gates are green on production paths, the reference and target
consumer runs reproduce from sealed artifacts, and an independent reviewer signs off on the
benchmark manifest.

## 8. Suggested first implementation slices

These slices are deliberately narrow and independently reviewable:

1. **Exact identity slice:** typed digest input, exact seed/timestamp encoding, no ignored
   canonical errors, replay digest verification, boundary vectors.
2. **Entity isolation slice:** entity-keyed kicks/shadows/RNG plus two-entity metamorphic
   tests.
3. **Adapter lifecycle slice:** begin/record/end API, real `agentic-stream` and `json-array`
   conformance, sink-failure state.
4. **Ledger conservation slice:** delivery-instance IDs, no cap, adapter omissions and
   buffer transitions, conservation assertions.
5. **Immutable oracle slice:** captured truth samples and pure baseline/recovery fixtures.
6. **Score fail-closed slice:** not-evaluated states, one-to-one matching, scorecard schema,
   adversarial tables.
7. **Generic suite slice:** remove every domain/effector literal, deterministic selection,
   transactional accounting.
8. **Live loop slice:** one operator topology, secure capability, reference consumer, and
   quiescence integration test.

Do not combine all phases in one refactor. Each slice should leave behind a stronger
executable invariant and a migration note for artifacts/contracts it changes.

## 9. Release checklist derived from this review

A benchmark-capable release must answer “yes” to all of these:

- Can one artifact reproduce a run with no catalog or repository present?
- Does replay reject every changed input before simulation?
- Can adding an unrelated entity leave all existing entity bytes unchanged?
- Does every native emission reconcile to explicit delivery-instance terminal states?
- Does the real output path honor adapter begin/end framing and encoding?
- Can every supported domain field be tied to a behavioral conformance test?
- Can no score pass when its required evidence is absent or ambiguous?
- Are baseline/recovery scores computed from immutable time-ordered evidence?
- Is truth inaccessible and immutable from seal until audited unblind?
- Does quiescence work concurrently without polling simulation state or real sleeps in tests?
- Can an external reference consumer close the loop through only operator capabilities?
- Can a new domain be added without a Go source branch or literal?
- Is suite generation byte-identical and based only on admitted scenarios?
- Are all documentation, toolchain, CLI, and CI claims executable and current?
- Have performance, security, and independent scenario-review gates passed?

Until all answers are yes, the honest project status is: **implementation prototype under
correctness hardening; not a benchmark authority**.
