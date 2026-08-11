# Level 2 remaining-work plan

Status: working plan — updated per slice
Date: 2026-08-11 (rev 5: Slice E landed)
Target: **Level 2 — benchmark-ready** per [QUALITY_BAR.md](QUALITY_BAR.md)
Baseline: `5a269a0` (Level 1 green; all tests, vet, ci-check pass)

This is the execution plan for the remaining Level 2 blockers recorded in
[IMPLEMENTATION_REBASELINE.md](IMPLEMENTATION_REBASELINE.md), informed by four independent
current-state surveys of the code and the external interaction review
[MCP_INTERACTION_REVIEW.md](MCP_INTERACTION_REVIEW.md) — two of whose findings (tool-list
drift, open top-level schemas) were already fixed by `18af187` and are tracked here only
as residuals.

The highest bar for this project is defined in QUALITY_BAR.md: Level 2 with all nine
non-negotiable gates green on production paths and a release manifest. Work stops only
when that bar is reached.

Revision notes (rev 2): incorporated the adversarial review — added the online/offline
scoring identity gate (P0, missing and already diverging), corrected the conservation
identity, fixed the crash-recovery flush ordering, decided the operator topology up front,
added command serialization to Slice E, specified oracle semantics per dynamic form, and
corrected several factual scopes.

## Slice ordering and dependencies

```
A (MCP strictness) ──► B (quiescence) ──► E (operator topology + mutex) ──► C (durable ledger)
                                             │
D (oracles, independent) ────────────────────┴──► F (G6–G9 + scoring identity) ──► G (release)
```

Each slice: plan → implement → review → fix → validate → commit, per the QUALITY_BAR
per-change loop. Commits name their gate.

---

## Slice A — MCP contract strictness (rebaseline blocker 5)

Status: **landed** (this plan rev). Commits: Slice A commit (this round).

Issue: top-level tool schemas are already typed and closed for all 28 tools
(`internal/mcp/schemas.go`, `additionalProperties: false`; unknown/missing args rejected
by the SDK, proven by `TestMCPRejectsUnknownAndMissingArguments`). Residuals:

1. **Six nested free-form objects** (`object()` helper, schemas.go:17-19):
   `sim.fault.inject.params`, `sim.perturb.apply.params`, `sim.env.inject.params`,
   `sim.effector.invoke.args`, `sim.consumer.report.verdict`, `sim.truth.seal.ground_truth`.
   The two contract-defined ones — `verdict` (consumer-verdict-v0.1) and `ground_truth`
   (ground-truth contract) — get typed nested schemas. `params`/`args` stay declared
   objects by design (domain/effector-dependent; `sim.effector.list` exposes per-effector
   argument schemas), but handlers must reject unknown keys.
2. **Handler-side fail-closed params**: `perturb.validateParams` (perturb.go:524-540)
   silently ignores unknown keys; `fault.inject.params` reads only `severity` and ignores
   the rest; `env.inject.params` validates nothing (run.go). All three reject unknown
   keys. Effector `args` validation already exists (`world.validateArgs`,
   effectors.go:160-173) — verify with a test, no code change expected.
3. **Time base**: the MCP default already matches the CLI default (mcp_test.go:193). The
   one real defect: `start_time: 0` is treated as "unset" (run.go:102), so epoch-0 cannot
   be requested — make presence-aware (honor explicitly supplied 0).
4. Stale comment at server.go:31-33 ("validated inside the handlers") — update.

Tests: schema advertisement asserts nested `verdict`/`ground_truth` shapes; handler tests
prove unknown param keys rejected for fault/perturb/env; effector args unknown-key
rejection test; `start_time=0` honored; presence of typed nested schemas in the tool
list.

## Slice B — G5 quiescence barrier, injectable and deterministic

Status: **landed** (this plan rev). Deviation from plan: no per-wait generation
token was added. The monotonic watermark (`quiescedThroughNS`, monotone under
`ReportQuiesced`) already prevents a stale report from satisfying a later wait
(a report only moves the watermark forward, and each wait checks `>= toNS`),
and a strict token would break the legitimate fast path where the consumer
reported before the harness issued the advance. The property is now proven by
deterministic tests instead.

Issue (survey): `awaitQuiescence` (run.go:532-551) uses a real 30 s wall timer, has no
cancellation path, no generation token; its only test sleeps (run_test.go:160,167);
`internal/wall` is an unused seam. On timeout the world has advanced but no command is
appended — the failed advance is not replayable, and `Advance` discards the emitted count
on error. `RunArtifact.Incomplete` exists (model/run.go:47) but replay never reads it
(replay.go:88-135).

Plan:
- Thread a per-request `context.Context` through `Advance` → `awaitQuiescence` (the MCP
  handler owns a request ctx; `ReportQuiesced` keeps its own lock).
- Injectable clock: small clock interface with real + fake (manually firable)
  implementations, injected via functional option; the 30 s bound becomes the injected
  deadline. No real time in tests.
- Per-wait generation token so a report cannot satisfy a wait it did not belong to; the
  monotonic watermark stays the primary guard.
- On timeout: return the emitted count alongside the error, append the advance command
  with its outcome to the command log, and mark the run incomplete. Replay reads
  `Incomplete` and reports incomplete for such runs (never silent success).
- Tests (no sleeps): timeout expiry, request-cancel, wake-on-report, stale-generation
  rejection, replay of a timed-out advance reproduces incomplete.

## Slice E — G3 out-of-process operator endpoint and golden closed loop

Status: **landed** (this plan rev). Deviations: the operator endpoint is streamable
HTTP (the SDK has no TCP/socket transport); `Run` gained a command-serialization mutex
with the quiescence wait releasing it (otherwise the consumer's mid-wait effector call
deadlocks); a scoring bug found by the golden loop was fixed — `resolveTime`'s pre-onset
baseline now samples strictly before the onset (emission-boundary alignments collapsed
the deviation to zero). Adversarial review of the slice: all P1s fixed (lint gate,
absence re-fire, post-End verdict guard), P2s fixed (dedup semantics, error reason,
timeout-path race, e2e assertions), nits fixed (`--world` removed, `--effector` requires
`--mcp`). The untracked external review `docs/MCP_INTERACTION_REVIEW.md` was committed
as review evidence.

Issue (survey): `NewOperatorServer` is test-only; `streamsim mcp --role operator` errors
(cli.go:417-418); refconsumer is offline-only with no quiescence reporting and no `--mcp`;
no out-of-process harness exists. Also: `Run` has no command-execution mutex, and the
world is single-goroutine with no locks — a second endpoint serving operator calls from
another goroutine races `Advance`/`Invoke`/`End`/`Report` (only `quiesceMu` exists).

Topology decision (made, not deferred): the SDK v1.7.0 has no TCP/Unix-socket transport.
Use **streamable HTTP on a localhost listener** for the operator role:
- One operator server per director process; a **token-resolving registry** maps the
  capability token (a tool argument, not a request attribute) to the owning world's
  `OperatorView`. `NewOperatorServer` takes the registry; handlers resolve per call.
- `sim.world.create` returns the operator endpoint URL alongside the token; the consumer
  connects with an SDK client (`StreamableClientTransport`).
- `Run` gains a **command mutex** serializing all mutations (Advance, InjectFault,
  operator Invoke/Report, End, SealTruth/Unblind). Serialized order = command-log order,
  preserving determinism. `ReportQuiesced`'s existing lock composes with it.
- CLI serves both servers with coordinated shutdown (director stdio + operator HTTP),
  and the usage text stops advertising `mcp --role operator` (currently errors, cli.go:27).
- Document the topology in MCP_SURFACE.md §2.

Refconsumer: add quiescence reporting (`quiesced_through_ns` in its verdict sink) and an
MCP operator-client mode (`streamsim refconsumer --mcp <url> --token ... --trace ...`)
per TECHNICAL_DESIGN §9.5 — read the trace, detect, invoke effectors over MCP, report
quiescence, submit the verdict.

Tests: black-box closed-loop golden run — director server up, world created, an external
SDK client connects to the operator endpoint, reads the nameplate, invokes an effector,
reports quiescence, submits a verdict, scores; prove director tools are unreachable from
the operator endpoint (tools/list and direct call attempts); repeat per shipped domain
(data-driven). Refconsumer-driven loop: `refconsumer --mcp` completes the same loop
out-of-process.

Out of scope: streamable HTTP for the director (stdio stays); TLS.

## Slice C — G4 durable append-only ledger

Issue (survey): ledger is an in-memory slice in `internal/run`, written once at `End()`;
no mid-run durability, no crash recovery; no conservation test beyond ~10 rows. The file
sink writes through `bufio.Writer` and only `Close()` persists bytes (sink.go:52-108) —
so a ledger that claims `Delivered=true` must not outlive unpersisted trace bytes.

Plan:
- Append-only ledger persistence: write each finalized delivery record to
  `ledger.jsonl` as it is produced, flushing at command boundaries (end of each
  serialized Advance/command execution, guaranteed by Slice E's command mutex).
- **Sink/ledger consistency**: at the same command boundary, `Flush()` (not fsync) the
  file sink writer so trace bytes and ledger rows land together; `End()` fsyncs both.
  Crash recovery is therefore at **command granularity**: rows of fully-executed commands
  are the recoverable prefix; a crash mid-command loses that command's rows, and the
  recovered run reports incomplete.
- Crash-recovery evidence: a test abandons a run mid-run (no `End`), reopens the ledger
  and trace files, and proves (a) recovered ledger rows equal the rows of fully-executed
  commands, (b) the recovered trace is consistent with them, (c) the run reports
  incomplete.
- Conservation at scale: a deterministic long stepped run emitting >10,000 delivery
  instances asserting the precise identity:
  `emitted == primary_rows` where primary rows are the unique-seq delivered rows
  (`delivered && reason != duplicated`), and `duplicated` rows are surplus; per-seq
  exactly one primary row; plus the `score.instrument` invariants (unique non-zero
  DeliveryIDs, seq coverage of [0, Emitted)).
  Note: `duplicate_burst`/`storm`/`id_reuse` produce two rows per emitted event, so a
  naive sum over reason categories is wrong; the identity above is the correct one.

Out of scope: WAL for world state; extracting the ledger into a new package (keep in
`run`, extract only the writer).

## Slice D — G2 independent oracles and fail-closed noise

Issue (survey): only `first_order_lag` has an analytic oracle. Fail-open bug: noise model
`none` with `sigma > 0` silently simulates gaussian (observe.go:176-179). `pink` is
rejected at load with no test. Perturbation params are fail-open on unknown keys. 12 of
19 perturbations have no conformance test (drop/duplicate_burst/producer_flap/reorder/
out_of_enum/out_of_range/injection_probe are tested; `delay_tail` has only determinism
coverage). Dead code: unreachable `rc_network` case in `f1Derivative` (dynamics.go:295).

Plan — oracle semantics per form (each oracle is an independent implementation, the S1
"implement twice" pattern, not a copy of the production code):
- `rc_network`: double-exponential closed form (composition of two first-order lags),
  RK4-vs-analytic tolerance pattern from `TestAnalyticCrossCheck` (world_test.go:97, tol
  0.005) swept over both time constants.
- `integrator`: `x0 + gain·u·t` (and threshold variant: piecewise linear in t).
- `dead_time`: **discrete** oracle — the production form is a ring buffer of
  `cap = ceil(DeadTimeS/dt)` feeding an optional lag (dynamics.go:334-346); the oracle
  re-implements the exact discrete recurrence independently in the test.
- F0 trend/seasonality: closed-form sinusoid sum vs emitted values.
- Fault envelopes `ramp`/`exponential`/`intermittent`: closed-form envelope shapes at
  sampled instants.
- Noise oracles: gaussian (seeded moments within tolerance), quantization (exact grid
  behavior: every emitted value is a multiple of sigma).
- **Fail-closed fixes**: reject `noise.model == none` with `sigma > 0` at load (and any
  unknown model); add the `pink` rejection test; add the `batch` cadence rejection
  conformance row.
- Availability: exponential sojourn means against seeded expectations;
  silence-during-down window test.
- Perturbation: fail-closed param validation (unknown keys rejected — shared with Slice
  A's handler work); conformance tests for the 12 untested perturbations
  (id_reuse, gross_backfill, clock_skew, non_monotonic, unit_mismatch, oversize,
  malformed, nan_inf, storm, time_encoding, precision_edge, delay_tail).
- Remove the dead `rc_network` branch in `f1Derivative`.

Out of scope: implementing pink/batch (keep fail-closed, per design).

## Slice F — G6/G7/G8/G9 hardening + scoring identity

Issue (survey + review):
- **F-0 (P0) Online/offline scoring identity** — QUALITY_BAR Level 2 requires identical
  results from one versioned bundle. The two paths already diverge: online requires
  exact labels (score.go:373) while offline accepts empty labels (offline.go:42); online
  checks seq coverage against `Emitted` (score.go:131-136) while offline has no coverage
  check (offline.go:58-70). Fix the divergence and add a test running the same
  verdict/label/ledger through both paths asserting identical scorecards, plus a
  versioned scoring-bundle constant.
- G6: the single prefix test drives only `nameplate.read`+`effector.list` via direct
  `OperatorView` calls; extend the harness through the real server tool layer (SDK
  client over the operator endpoint) including `invoke`, `report`, and error-path
  responses; run per shipped domain.
- G7: probe every delivery path (file sink, inproc, http-push where configured) and the
  shipped domain with attacker-controlled channels; assert byte-identical verdicts.
- G8: strengthen `silent_no_effect` to assert the full tuple (effector/entity/time/args/
  outcome) leaves scoped truth unchanged; confirm action scoring checks the tuple.
- G9: route the audit through the run pipeline — build the world, run it through
  perturb + adapter + ledger, and feed the hindsight detectors the **delivered ledger
  records in monotonic order** (detector series sampled at record timestamps; fixtures
  re-baselined, thresholds re-verified). Move audit entity selection into domain/profile
  data, removing the `site-a/pond-%d` literal (director.go:463). Suite byte-identity
  test (two Generate calls byte-equal) and rejected-candidate counter-accounting test.

Out of scope: redesigning the hindsight-detector panel.

## Slice G — release evidence

Issue (survey): `ci-check` skips deadcode/govulncheck when absent (exit 0); no
fuzz/soak/perf/bench targets; no manifest anywhere; no cross-process tests.

Plan:
- Makefile: `ci-check` **fails closed** when `deadcode`/`govulncheck` are unavailable
  (Level 2 requires the checks be available and fail closed); install the tools locally;
  add `fuzz`, `soak`, `perf` targets — go native fuzzing (bounded), a deterministic soak
  run asserting conservation + determinism + bounded memory, benchmarks for the emission
  path.
- Release manifest generator (`streamsim manifest` or `make manifest`): records
  simulator/domain/adapter/suite/toolchain/consumer digests plus author and independent
  review identity, optionally signed with stdlib ed25519 when a key file is supplied
  (no new dependencies).
- Cross-process determinism evidence: spawn `streamsim run` twice as subprocesses and
  assert byte-identical artifacts/traces (current replay tests are in-process only).

Out of scope: cosign/sigstore (external tooling).

---

## Acceptance definition (done)

- All nine gates G1–G9 green on production paths per QUALITY_BAR.md evidence columns.
- Online and offline scoring produce identical scorecards from the same versioned bundle.
- `make ci-check` green with deadcode/govulncheck present, and failing closed when absent.
- Release manifest produced for the current commit.
- IMPLEMENTATION_REBASELINE.md updated to Level 2 with commit evidence; QUALITY_BAR.md
  current-bar-status updated; MCP_SURFACE.md documents the operator topology and the
  file-sink flush semantics.
