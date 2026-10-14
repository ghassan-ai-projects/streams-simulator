# SI-09 · reservoir-release

- **ID:** SI-09
- **Slug:** `reservoir-release`
- **Status:** proposal — no domain artifact exists yet (`ls domains/ | grep -ic reservoir` returns 0, exit 1, re-run while writing this document; the eight shipped domains are untouched). Quality-bar score 73/100 — see [Quality bar result](#quality-bar-result).
- **Serves:** `agentic-stream` (reconsideration deadlines vs dead time, late correction of accepted commands, duplicate wakeups vs the idempotency window — all measurable from the EffectorCalls log and existing loop metrics) primarily — verifiable in-repo: `adapters/agentic-stream.adapter.json` ships, and the device contract is vendored from Agentic Stream per `documentation/reference/contracts.md:21`. `tamoz` secondarily (deterministic-clock misfire/overlap/duplicate-wakeup matrix; named in-repo at `docs/design/GAP_ANALYSIS.md:46` and `docs/design/CRITICAL_REVIEW.md:145-161`), whose matrix probes the documented supersession/harness-artifact hole (`docs/design/GAP_ANALYSIS.md:61-62`) against that substrate.
- **Prerequisites declared:** none in v1 code. The v1 domain is one schema, every field an existing knob. Two named future items, neither assumed present: (1) a machine-readable top-level `release_calendar` object (a v2 schema extension) before the calendar story becomes policy input; (2) a `SuiteConfig` solver-horizon field, only if a future slow-onset fault family needs crossings beyond 72 h post-onset — and it must stay inside the determinism tuple if added.

## Purpose

A v1 domain that ships seasonal gate physics and command discipline under corrupted evidence, plus an explicitly deferred v2 calendar story. v1: 2–4 parallel peer reservoirs on a compressed annual hydrograph (period_s 345600). A declared 96 h nominal profile puts the full 360-degree cycle on the wire with onsets sampling 0–180 degrees — structurally scannable today because the observability scan is onset-anchored, `internal/truth/solver_scan.go:40-51` — while a 72 h profile spans 270 degrees with onsets over 0–135 degrees. Open-field declares the same annual axis but its nominal scenarios sample 172800/31536000 = 0.548% of a cycle, a near-linear drift: exactly the nominal-on-the-wire failure mode. Actuation is slow (`slow_physical`) with a real 600 s dead time.

Sealed truth is per-scenario `ground-truth-v0.1` records over a gate-failure family. The record shape — one fault-or-negative label, one `expected_effector`, one `deadline_ns`, two ≤200-char counterfactual strings (`internal/schemas/ground-truth-v0.1.schema.json:19-72`; `internal/truth/record.go:29-49`) — cannot hold a schedule, and slot-per-scenario casting (one scenario per calendar slot) was rejected as the downgrade the critic named. The scored signature is **phase-at-onset**: the suite draws onsets in `[startNS, startNS + durationNS/2]` (`internal/suite/scenario_draw.go:21-34`, the `durationNS/2` at :28-29), so the same label lands at different points of the seasonal cycle — across the 72 h profile's sampled arc the hydrograph runs 12.0 → 18.0 → 16.24 m3/s (computed: 12 + 6·sin at 0/90/135 degrees of the 96 h cycle; 45 degrees gives 16.24) — and `first_observable`, `unavoidable` and storage headroom at onset are functions of that phase through the closed-form oracle. Measurable calendar structure with zero new machinery.

The discrete release calendar is a v2 candidate, not v1 product. Its only carrier today is prose in the 2000-char domain description (`internal/schemas/domain-spec-v0.1.schema.json:17`) — lossy as a configuration source, non-load-bearing by construction, and no scored question depends on it. Serving it as machine-readable policy input is gated on a named `release_calendar` schema extension.

What the instrument measures: whether a consumer re-evaluating on late/skewed telemetry stacks a correction inside the dead time (overlap is emergent kick composition — independent add-or-assign kicks at `atNS+dead_time`, `internal/world/effects.go:19-31` — a late command never invalidates an accepted one); whether duplicate wakeups collide with the idempotency window; whether it mistakes a jammed gate for weather; and whether detection latency tracks seasonal phase. Verdicts are exactly what the loop metric computes: first successful expected-effector call vs `deadline_ns`, action appropriateness, false success on `silent_no_effect`, unnecessary action on the negative class (`internal/score/loop.go:9-20`, `internal/score/loop_metrics.go:8-44`). There is no which-command-won supersession verdict, so the misfire/overlap matrix is the consumer's experiment design executed against the simulator's outputs — physics, delivery corruption, the EffectorCalls authority log, the loop metric — not new simulator evidence.

## How it runs on this simulator

Everything below is domain DATA against the one schema (`internal/schemas/domain-spec-v0.1.schema.json`). No domain code branches. Every section is closed (`additionalProperties: false`, schema :7), so anything not expressible below is expressible nowhere.

### Entities

2–4 parallel peer reservoirs `reservoir-{n}` (count default 3), explicitly **not** cascaded: F1 inputs read same-entity states only (`internal/world/integration.go:101-105`), so cross-entity physics coupling is not schema-expressible. Siblings share one headwater hydrograph — which is exactly what makes `peer_residual` detectors valid.

### State variables and dynamics

| State | Form | Inputs / parameters | Units |
| --- | --- | --- | --- |
| `inflow_true` | F0 | baseline 12.0 + seasonality amplitude 6.0, period_s 345600, phase_rad 0 (schema :114-127; integration `internal/world/integration.go:120-129`) | m3/s |
| `storage_true` | F1 `integrator`, dt_ms 60000 (schema :107, required for F1 per :215) | inputs `inflow_true` +1 / `release_true` −1; clamp 0..3,000,000; initial 2,700,000 | m3 |
| `gate_command_true` | assigned by effector | `assign_from_arg position` | percent |
| `gate_jam_offset` | relative, initial 0 | fault target | rel |
| `gate_position_true` | F1 `first_order_lag` tau 120 | inputs `gate_command_true` +1 / `gate_jam_offset` +55; clamp 0..100 | percent |
| `release_true` | F1 `first_order_lag` tau 240 | input `gate_position_true` coef 0.12 | m3/s |
| `release_volume_true` | F1 `integrator` | input `release_true` +1 — the counter | m3 |
| `level_sensor_bias` | relative, initial 0 | fault target | rel |

**Seasonal phase as the scored variable.** Onset draws land in `[startNS, startNS + durationNS/2]` (`internal/suite/scenario_draw.go:21-34`). A 72 h profile therefore samples onset phases over the first 135 degrees of the 96 h cycle (first three-eighths) with each trace covering 270 degrees, and the hydrograph across the sampled arc runs 12.0 → 16.24 → 18.0 → 16.24 m3/s (12 + 6·sin at 0/45/90/135 degrees, computed). Headroom at onset, net drift rate, and therefore `first_observable`/`unavoidable` vary with phase through the closed form. A declared 96 h nominal profile samples onsets over 0–180 degrees and covers the full 360-degree cycle on the wire; full 360-degree *onset* sampling needs duration_s of about 192 h (2× period, so the half-duration draw window spans the cycle) — data-only, no code change, because scanning is onset-anchored (`internal/truth/solver_scan.go:40-51`, horizon line :41). This is what makes `seasonality=annual` exercised rather than decorated: open-field's annual term is near-linear on the wire at 0.548% cycle coverage per scenario.

### Signals

| Channel | Observes | Cadence / noise | Notes |
| --- | --- | --- | --- |
| `inflow_rate` | `inflow_true` | periodic 60 s, gaussian sigma 0.2 | m3/s |
| `storage_level` | `storage_true` | sigma 1000 | m3; `observation_bias` via `level_sensor_bias` |
| `gate_position` | `gate_position_true` | sigma 0.5 | percent |
| `release_rate` | `release_true` | sigma 0.1 | m3/s; real-state detection channel |
| `release_volume` | `release_volume_true` | counter | the declared confirmation channel |
| `status` | — | enum | string channel |

### Effectors

| Effector | Effect | Timing | Ack modes |
| --- | --- | --- | --- |
| `set_gate_release` | assigns `gate_command_true` from arg `position` | `dead_time_s` 600, `time_constant_s` 300 (schema :543-544) | `silent_no_effect` 0.01 + `confirmed_no_effect` 0.01, so `confirmation_channels: [release_volume]` is required by schema :513-517 (modes enum :557); `idempotency_window_s` 900 (:547) |
| `service_gate` | delta −1 on `gate_jam_offset` | dead time 1800 s | clears the jam by the same additive cancellation open-field's `maintenance_reset` uses |
| `calibrate_gauge` | delta −1 on `level_sensor_bias` | — | removes the injected gauge bias |

**The G8 split, stated correctly:** under `silent_no_effect` the `release_volume` counter reports the shadow release that *did not happen* — the ack lie internally consistent per `internal/world/reading.go:25-27` (confirmation channels read `shadowValue`) and schema :516 — while detection lands on the real-state `release_rate`/`gate_position` channels plus the false-success metric (`internal/score/loop_metrics.go:31-44`).

### Fault family

A gate/weather family, deliberately **no conservation leak**, with onset fields named exactly as the schema declares them (`internal/schemas/domain-spec-v0.1.schema.json:370-380`: `shape` enum including `intermittent` at :376, `rate_per_hour` :377, `magnitude` :378, `duty_cycle` range 0–1 at :379). Envelope semantics per `internal/world/dynamics.go:95-107` (verified): the ramp envelope is `min(rate·elapsed_h, 1)`, so full deviation lands at 1/rate hours independent of magnitude, and deviation grows at magnitude·rate per hour.

| Fault | Onset | Target | Detector | Expected effector | Deadline | Class / notes |
| --- | --- | --- | --- | --- | --- | --- |
| `gate_jam_open` | step, magnitude +1 | `gate_jam_offset` | `single_channel_snr` `gate_position` | `service_gate` | 5400 s | positive; counterfactuals "spill at capacity" / "release back on schedule" |
| `gate_seized_closed` | step, magnitude −1 | `gate_jam_offset` | `single_channel_snr` `release_rate` | `service_gate` | 5400 s | positive |
| `level_gauge_drift` | ramp, rate_per_hour 0.4, magnitude +1 | `level_sensor_bias` | `peer_residual` `storage_level` | `calibrate_gauge` | — | positive; full deviation at 2.5 h (1/rate) |
| `headwater_drop` | ramp, rate_per_hour 0.4, magnitude −6 | `inflow_true` | `single_channel_snr` `inflow_rate` | `set_gate_release` | — | the cascade root, weighted in `correlated_cascade` with `simultaneous_entities` 3; growth 2.4 m3/s per hour, so the 3-sigma crossing against sigma 0.2 lands at ≈0.25 h and 10-sigma at ≈0.83 h (computed) |
| `headwater_swell_none` | intermittent, rate_per_hour 2.0, magnitude +6, duty_cycle 0.25 | `inflow_true` | `peer_residual` `inflow_rate` | none | — | negative class — the unnecessary-action trap; the exact field name and pattern open-field ships at `domains/open-field-irrigation.domain.json:211` |

The load-bearing faults are gate jams a schedule-vs-feedback comparison catches, so the domain owns command discipline end-to-end.

### Perturbations

Emission path only — commands are never perturbed (`internal/run/emission.go:18`):

- `delay_tail` + `clock_skew` — a single windowed offset with sign (`internal/perturb/transform_timing.go:48-61`) — plus `duplicate_burst` over director-declared windows (`internal/perturb/perturb.go:21-27`).
- The strict observed-order guard runs **pre**-perturbation (`internal/run/emission.go:12-23`, guard at :89-101), so delay/skew compose safely and the guard is not a delivery constraint.
- **Duplicate wakeups are consumer-side re-invocations**, read against the EffectorCalls authority log: one entry per distinct `command_id`; an in-window replay returns the cached result and applies no second effect (`internal/world/effectors.go:49-51, 84-93`; command timing vs `effect_eta_ns` at :156/:168; log shape `internal/world/world.go:113-129`).

### Ground truth and its analytic oracle

Per-scenario `ground-truth-v0.1` records: label, `expected_effector`, `deadline_ns` (= `first_observable_time_ns` + `deadline_s`, `internal/truth/record.go:36-38`), and two ≤200-char counterfactual strings (`internal/schemas/ground-truth-v0.1.schema.json:63-70`; `internal/model/domain.go:259-263`).

Oracle: piecewise closed-form storage under stepwise commands and sinusoidal inflow, cross-checked against the solver's independent RK4. The integrator closed form and the dead-time discrete oracle exist (`internal/world/oracle_test.go:57-75` verified; the dead-time oracle at :108-183 per the consolidated spec's round evidence). The G2 test must include (a) the boundary case of a correction landing inside the 600 s dead time — the second kick is scheduled independently, `internal/world/effects.go:19-31` — and (b) a phase sweep asserting solved timestamps track the closed form across the sampled onset arc.

**Disclosure boundary.** The calendar prose is sealed by `domain_digest` and reaches the consumer only through whoever configures the suite — the simulator owns no consumer configuration (`docs/design/GAP_ANALYSIS.md:263`). The nameplate carries no domain description (`internal/mcp/operator.go:56-68`), but `EffectorInfo` carries per-effector description text to the operator (`internal/mcp/operator.go:48-54`): the domain author must keep the calendar prose in the domain description and never migrate it into effector or channel nameplate text, or it leaks.

### The release calendar is v2-only context

`axes.time_reference=calendar` is a legal enum (`internal/schemas/domain-spec-v0.1.schema.json:31`) but metadata-only in the binary — catalog code stringifies and counts it, no behavior branch (`internal/domain/catalog.go:75-84`, `internal/domain/catalog_coverage.go:14-35`). The slot table (four slots per cycle at offsets 0/0.25/0.5/0.75 of period_s, target gate positions 15/35/60/30%) lives as prose in the domain description: context for construction and consumer experiments, lossy as a configuration source, never sealed truth, no schedule-adherence metric, and no scored question depends on it. It has **no schema home** — every section is closed (schema :7) and the only free-form container, `profiles[].setup[].args`, carries effector args that all fire at scenario `startNS` (`internal/suite/scenario.go:107-119`) — so a machine-readable top-level `release_calendar` object is the named v2 prerequisite this spec does not assume.

## Feasibility against G1–G9

Gate IDs follow `docs/QUALITY_BAR.md:62-72`. One schema, every field an existing knob: F1 `integrator`/`first_order_lag` forms, inputs with coefficients, clamp (schema :130-168), F0 seasonality (:114-127), slow_physical effect `dead_time_s`/`time_constant_s` (:543-544), `confirmation_channels` required when no-effect modes are declared (:513-517, modes :557), `idempotency_window_s` (:547), profiles `duration_s` (:588), onset shapes with `rate_per_hour`/`magnitude`/`duty_cycle` (:370-380), and all three axis values legal enums (:31-34). Sealed truth fits `ground-truth-v0.1` unchanged — labels, `expected_effector`, `deadline_ns` from `first_observable`+`deadline_s`, counterfactual strings are exactly what the record builder emits (`internal/truth/record.go:29-49`).

| Gate | Standing for this domain |
| --- | --- |
| G1 Determinism | Data-only today. One forward-looking constraint: a `SuiteConfig` horizon field, if ever added for slow-onset families, must be a function of domain-committed inputs, since the run digest covers (sim_version, domain digest, seed, world config) only (`internal/run/quiescence.go:47-50`) and a per-run override would sit outside the determinism tuple — proven by a test. |
| G2 Analytic cross-check | Piecewise closed-form storage vs the solver's independent RK4, including the in-dead-time correction boundary (second kick scheduled independently, `internal/world/effects.go:19-31`) and a phase sweep across the onset arc. Integrator closed form verified at `internal/world/oracle_test.go:57-75`. |
| G3 Reference consumer | Not specifically addressed by the spec; nothing in the sketch changes the operator surface — invokes travel the standard path, and the shipped `adapters/agentic-stream.adapter.json` is the primary consumer substrate. |
| G4 Delivery ledger | Background green. Perturbations are emission-path only (`internal/run/emission.go:18`); the strict observed-order guard runs pre-perturbation (`internal/run/emission.go:12-23`, guard :89-101), so delay/skew compose safely and the guard is not a delivery constraint. Duplicate wakeups are consumer re-invocations resolved by the idempotency window (`internal/world/effectors.go:49-51, 84-93`) and evidenced by EffectorCalls (`internal/world/world.go:113-129`) — the ledger is not the command oracle. |
| G5 Quiescence | Not specifically addressed beyond bounds: solver worlds advance to at most durationNS + 72 h (≈120 h of world horizon for the declared 96 h profile); effects are ordinary scheduled kicks under the world's existing pending-kick accounting. |
| G6 Sealed oracle | The sealed description and labels stay behind the director boundary — the nameplate carries no domain description (`internal/mcp/operator.go:56-68`). Required test: truth records unblind monotonically with the label, no partial reads. The `EffectorInfo` description field (`internal/mcp/operator.go:48-54`) is a named constraint on the domain author (calendar prose never migrates into nameplate text), not a leak today. |
| G7 Injection probe | Not specifically addressed; no attacker-controlled text channels are declared. The fault family injects through the standard onset/affects data path. |
| G8 `silent_no_effect` | `release_volume` as the confirmation channel — the consistent shadow lie (`internal/world/reading.go:25-27`, schema :516) — with detection on real-state channels plus the false-success metric (`internal/score/loop_metrics.go:31-44`). |
| G9 Trivial baseline | The one remaining empirical gate. The **structural scannability** case is verified: `searchOnsets` anchors the horizon at the onset — `horizon := max(onsetNS+maxHorizonNS, startNS+maxHorizonNS)`, scanning from `max(onsetNS, startNS)` (`internal/truth/solver_scan.go:40-51`, line :41) — so every onset gets its own 72 h post-onset budget; the solver-horizon hardcode is therefore **not binding** for this domain. The **trivial-baseline outcome** is not verified: passage of the audit panel under the generate-audit-regenerate loop is a suite run no one has executed, and this spec claims nothing about its verdict. A grep for `maxHorizon`/`HorizonNS`/`72*3600` across `internal/audit/*.go` found no horizon constant of its own. The hardcode at the NewSolver call site (`internal/suite/generation_setup.go:26`) with no `SuiteConfig` horizon field (`internal/suite/suite.go:61-68`) remains a robustness item for future slow-onset fault families. |

Crossing times, per the envelope code `min(rate·elapsed_h, 1)` (`internal/world/dynamics.go:95-107` — full deviation at 1/rate hours independent of magnitude, growth magnitude·rate per hour): step faults immediately; the declared ramps (both rate_per_hour 0.4) reach full deviation at 2.5 h; `headwater_drop` grows at 2.4 m3/s per hour, so `first_observable` (3-sigma = 0.6 m3/s against sigma 0.2) lands at ≈0.25 h and `unavoidable` (10-sigma) at ≈0.83 h — computed. The round-6 critic's 15–30 h full-deviation figures follow a magnitude/rate formula the envelope code does not implement; even under that slower reading, the critic's own `first_observable` (1.5–3 h) and `unavoidable` (5–10 h) sit far inside the 72 h budget, so the conclusion is unaffected under either reading.

Cost, bounded: per scenario the onset scan is at most 72 h / 60 s = 4320 SNR evaluations (onset-anchored, `internal/truth/solver_scan.go:40-51`), and the solver worlds advance to at most durationNS + 72 h — about 120 h of world horizon for the declared 96 h profile, on the order of 10^4–10^5 RK4 steps per world at dt 60000 ms; the 192 h full-phase variant raises this about 1.4×. Empirical at authoring time, bounded by construction.

Two items are honestly named as **not** implied machinery: (1) the on-time-vs-late counterfactual has no runner — the solver builds exactly two worlds (`internal/truth/solver_scan.go:19-28`) and `SetupCall` feeds only those two (`internal/truth/solver.go:52-61`); alternate-command-log replay exists (`internal/run/replay.go:42-58`) but a third counterfactual command world is an explicit design change; (2) zero new scoring code — misfire/overlap/late-correction reduce to the loop metric's existing verdicts (`internal/score/loop.go:9-20`); which command wins inside dead time is visible only as kick superposition (`internal/world/effects.go:19-31, 56-64`) and is the consumer's measurement to design.

Validation status: schema-suite validation of the eventual domain JSON and the full repo gate (`make ci-check`) have **not** been run — the artifact does not exist (`ls domains/ | grep -ic reservoir` returns 0, exit 1, re-run this session) and the gate is not named by this ask. All claims above are schema/code reads or computed checks; see [Verification performed for this document](#verification-performed-for-this-document).

## Quality bar result

| Dimension | Score |
| --- | --- |
| consumerValue | 7/10 |
| dataExpressibility | 8/10 |
| determinismOracle | 7/10 |
| ledgerQuiescence | 7/10 |
| novelty | 6/10 |
| sealedTruth | 7/10 |
| specificity | 9/10 |
| **Total** | **51/70 = 73/100** |

**Recorded verdict: "met" as supplied** by the workflow's quality-bar evaluation — the bar being: total at least 85/100 across the seven dimensions, with no single dimension below 7/10. Recorded discrepancy, not smoothed over: the seven scores sum to 51/70 = 73/100, which is **below** the stated 85/100 total threshold, and the per-dimension floor is **also** violated (novelty 6/10 < 7/10) — unlike SI-02, whose floor held, this idea fails both clauses of the stated bar taken literally. The supplied "met" verdict and the stated threshold do not agree; both are reported here so the reader can judge.

## Five-lens review

### 1. Determinism and analytic oracle (7/10)

**Found:** the counterfactual was under-specified relative to the record shape — the sealed label's counterfactual is two descriptive strings ≤200 chars (`internal/schemas/ground-truth-v0.1.schema.json:63-70`; `internal/model/domain.go:259-263`), so "the same integrator run under sealed alternate inputs" produces a trajectory with nowhere sealed to live, and "misfire, overlap, supersession scored against physical consequence" implies new scoring policy over the command log plus alternate-world runs that the sketch treated as free (`SetupCall` machinery exists at `internal/truth/solver.go:52-61,169-176` but feeds only the clean/faulted pair). "Sealed truth is a deterministic release schedule" conflated layers: the per-run label record has no schedule field. Physics verified expressible: multi-input F1 (`internal/model/domain.go:101,111-115`), kick ramp analytic (`internal/world/dynamics.go:33-41`), dead-time and integrator oracles (`internal/world/oracle_test.go:113`, `:58`), windowed `clock_skew` (`internal/perturb/transform_timing.go:48-61`); the 72 h solver horizon (`solver.go:37`) flagged for revisit since solver config is part of the determinism tuple.

**Changed:** the counterfactual is named as an explicit third solver world (alternate command log) — a design change, not implied machinery — with the G2 test that its closed form and RK4 agree, including the in-dead-time boundary; the release schedule is now domain data sealed by `domain_digest` (v2 prose), with a G6 test that unblinding exposes truth monotonically with the label, no partial schedule reads before unblind.

### 2. Data-not-code expressibility (6/10)

**Found:** the headline sealed truth had no record type — a day-of-year release schedule cannot live in `ground_truth` (one fault label, three timestamps, a single `expected_effector` string, `internal/schemas/ground-truth-v0.1.schema.json:19-72`), and the world has no autonomous schedule executor (dynamics are F0 sinusoids, faults, and effector deltas; a piecewise calendar is none of these). The counterfactual needed a new solver mode: `Solve` runs clean-vs-faulted worlds with identical setup calls (`internal/truth/solver.go:65-106`); there is no alternate-command-history world. Verified expressible residue: `seasonality=annual` and `slow_physical` both ship in open-field, compressed period is legal F0 data, windowed `clock_skew` matches the killed rival's concession, seepage via same-entity `conservation_residual`, G8 via a release-counter confirmation channel.

**Changed:** sealed truth reframed around the gate-failure family and dead-time physical consequence (existing records), with the release schedule demoted to the consumer's policy under test — the slot-per-scenario casting explicitly rejected; the machine-readable `release_calendar` object proposed as a design change first if schedule adherence ever becomes the product, flagged as the largest new-concept load in the portfolio.

### 3. Consumer testing value (7/10)

**Found:** the G8 cell was specified backwards — a channel declared as confirmation reports the shadow (counterfactual) value under `silent_no_effect` (`internal/world/reading.go:23-27`; schema :516), so the original "gate commands get silent_no_effect confirmation via a downstream release counter" would have the counter show the release that did not happen; the tell must be a non-confirmation channel reading real state, or the loop metric (`internal/score/loop_metrics.go:31-44`). The command-timing counterfactual is not built by the solver (`prepareScan` constructs exactly two worlds, `internal/truth/solver_scan.go:19-28`); "supersession scored against physical consequence" reduces to deadline adherence + action appropriateness (`internal/score/loop.go:17`, `loop_metrics.go:31-34`) unless a design change is taken. Annual compression is feasible (seasonality is a data-declared sinusoid with period_s, `internal/model/domain.go:88`, integrated at `internal/world/integration.go:120-127`), but with period_s >> durationNS the seasonal term is near-linear over the scenario and the annual-axis claim becomes nominal on the wire.

**Changed:** the `silent_no_effect` cell re-specified — `release_volume` declared as the confirmation channel so the ack lie is internally consistent, detection moved to a downstream real-state channel plus the false-success metric (mirroring SI-02's confirmed/silent split); the supersession verdict restated in terms the scorer already computes (action timing vs `effect_eta_ns`/deadline from the EffectorCall log, `internal/world/world.go:113-129`) or named as an explicit design item, never implied; the duration-vs-period constraint realized in profile data — the declared 96 h nominal profile covers the full cycle on the wire and `seasonality=annual` is exercised rather than decorated.

### 4. Delivery ledger and quiescence (7/10)

**Found:** delivery-layer risk is low and verified — `clock_skew` is a fixed-offset rewrite with sign (`internal/perturb/transform_timing.go:48-61`), monotone per direction, composing safely with `duplicate_burst`; the strict observed-order guard runs before perturbation transforms (`internal/run/emission.go:12-23`), so delay/skew cannot fail the guard. Same command-path misattribution as SI-02, milder: "duplicate wakeups" and supersession are consumer re-invocation phenomena — perturbations cannot touch commands (`Perturb.Process` is emission-only, `internal/run/emission.go:18`) — so the misfire/overlap matrix must be driven through the effector failure-mode data and the idempotency window (`internal/world/effectors.go:85-93`). The counterfactual is world-side and adds no delivery evidence; it must not be described as ledger-supported.

**Changed:** duplicate-wakeup semantics attributed explicitly to the idempotency window and EffectorCalls log — one entry per distinct `command_id`, in-window replay returns the cached result (`internal/world/effectors.go:49-51, 84-93`); a skew+delay composition test documents that observed-time monotonicity is enforced pre-perturbation only (`internal/run/emission.go:89-101`), so reviewers do not mistake the guard for a delivery constraint; the compressed seasonal phase stays in domain data.

### 5. Novelty and portfolio balance (7/10)

**Found:** the axis fills are shipped-relative and the nearest catalog neighbor was unmentioned — `seasonality=annual` and `actuation=slow_physical` are also carried by catalog 05 grain-storage (`docs/design/DOMAIN_CATALOG.md:165-166`), and catalog 01 greenhouse-climate owns the actuation-dead-time stress (`:88-97`: "any agent that re-evaluates before the dead time expires will stack corrections and oscillate") without SI-09 positioning against it. The seepage leak was a third conservation-leak fault the supersession story does not need. Against the shipped set the claim holds (annual=1 and slow_physical=1, both open-field-irrigation), and schedule-as-sealed-truth plus misfire/overlap/supersession scoring is genuinely absent — that should lead, not axis arithmetic. Persistent per-entity clock skew is genuinely not schema-expressible (`link_delay` has no per-entity offset field, schema :324-334), but should be stated as a documented exclusion per the catalog's honest-absence standard (`DOMAIN_CATALOG.md:45-48`). The killed candidate's name appears in 0 repo files — kill claim unverifiable.

**Changed:** catalog 01 named as the dead-time neighbor with the delta stated (01 scores an agent fighting its own actuation lag; SI-09 sets up which command wins under late/skewed/duplicated evidence — answered by consumer experiment design since no supersession verdict exists); the seepage leak dropped, the gate-jam family made load-bearing so the domain owns command discipline end-to-end without adding a fourth leak; `time_reference=calendar` declared (0→1 shipped, all eight shipped domains declare wall — verified by reading the axes blocks), honestly priced as metadata-only (`internal/domain/catalog.go:75-84`) labelling v2 context; the compressed phase specified concretely (period_s 345600, profile durations, computed arc) so feasibility is auditable; the unverifiable kill claim removed — only the windowed-skew exclusion is asserted, from in-repo evidence.

## Improvement history

Ten improvement rounds separate the original consolidated spec from the final spec. The final spec narrates rounds 4–8 explicitly (plus "a prior round" grep); the remaining rounds are visible only as the original-vs-final delta, and the grouping below is the documenter's clustering of that delta, not recorded round numbers. What changed:

1. **Sealed truth rebuilt** (early rounds). Original: "sealed truth is a deterministic release schedule (a function of day-of-year)" with the counterfactual as "the same integrator run under sealed alternate inputs". Final: the record shape cannot hold a schedule (`ground-truth-v0.1.schema.json:19-72`), slot-per-scenario casting rejected, and truth is a per-scenario gate-failure family; the schedule becomes v2 prose in the domain description, sealed only via `domain_digest`.
2. **The supersession verdict claim retracted; consumers swapped** (early rounds). Original: "misfire, overlap, and supersession scored against physical consequence", tamoz primary. Final: `loop.go:9-20` and `loop_metrics.go:8-44` score resolution, false-success and unnecessary-action only — no which-command-won verdict exists; the misfire/overlap matrix is the consumer's experiment design against the substrate, and the primary consumer flips to agentic-stream (shipped adapter, vendored contract) with Tamoz secondary probing `GAP_ANALYSIS.md:61-62`.
3. **Fault family replaced** (early rounds). Original: seepage leak via `conservation_residual` — a third conservation leak the story does not need. Final: gate jams, gauge drift, and headwater drop/swell; deliberately no conservation leak; the load-bearing faults are gate jams a schedule-vs-feedback comparison catches.
4. **The calendar demoted and its leak path closed** (early rounds). Original: the schedule as the scored product. Final: v2-only, 2000-char prose (schema :17), non-load-bearing, no scored question depends on it; the `release_calendar` schema extension named as the prerequisite; the disclosure boundary stated — nameplate carries no domain description (`operator.go:56-68`) but `EffectorInfo` carries effector descriptions (`operator.go:48-54`), so the prose must never migrate into nameplate text.
5. **Seasonality made load-bearing** (through round 5). Original: "parameterize a compressed seasonal phase in domain data" with no numbers. Final: period_s 345600; the 96 h nominal profile covers the full 360-degree cycle on the wire with onsets over 0–180 degrees; the 192 h duration gives full onset sampling (2× period, data-only); open-field's annual term exposed as near-linearity on the wire (0.548% coverage, computed); phase-at-onset made the scored signature.
6. **Round 4→5: the sampling-arc and horizon overstatements corrected.** Round 4 overstated both the sampling arc and the horizon constraint; round 5 recomputed the arithmetic against `scenario_draw.go:21-34` (the `durationNS/2` draw window) and re-derived the scannability case from `solver_scan.go:40-51`, and the same round's grep for `maxHorizon`/`HorizonNS`/`72*3600` across `internal/audit/*.go` found no horizon constant of its own.
7. **Round 6: ramp-timing slip fixed; crossings recomputed under the real envelope.** The round-6 critic's 15–30 h full-deviation figures follow a magnitude/rate formula the envelope code does not implement; the final spec declares onset rates in the fault data and restates crossings per `dynamics.go:95-107` (`min(rate·elapsed_h, 1)`): ramps fully deviated at 2.5 h, `headwater_drop` first-observable ≈0.25 h, unavoidable ≈0.83 h (computed) — with the conclusion shown robust under either reading.
8. **Round 7: onset field names fixed to the schema's exact vocabulary.** `duty_cycle` (schema :379) and `intermittent` (:376) replace paraphrase, matching the shipped open-field negative control (`domains/open-field-irrigation.domain.json:211`).
9. **Round 8: the horizon line re-verified verbatim and the empirical gate isolated.** `solver_scan.go:41` re-read character-for-character (`horizon := max(onsetNS+maxHorizonNS, startNS+maxHorizonNS)`); the claim split into its verified half (structural scannability, onset-anchored) and its unverified half (trivial-baseline audit outcome — a suite run no one has executed), with the NewSolver hardcode (`internal/suite/generation_setup.go:26`) and the absent `SuiteConfig` horizon field (`internal/suite/suite.go:61-68`) recorded as a determinism-bounded robustness item, not a v1 blocker.
10. **The G8 cell reversed and delivery status priced honestly.** Original: "gate commands get silent_no_effect confirmation via a downstream release counter" — backwards, since confirmation channels report the shadow (`reading.go:25-27`). Final: `release_volume` is the confirmation channel (consistent ack lie), detection lives on real-state channels plus the false-success metric; and every round re-ran `ls domains/ | grep -ic reservoir` (0, exit 1) so the absence of the artifact, the un-run schema-suite validation and the un-run `make ci-check` stay stated rather than implied.

Net effect: the idea kept its substrate (dead time, idempotency window, delivery corruption, authority log, compressed annual phase) and surrendered every claim that required machinery that does not exist — schedule-as-truth, supersession verdicts, counterfactual command worlds, per-entity skew — replacing them with scored questions the current code already answers.

## Open questions

1. **The trivial-baseline verdict for the 96 h full-period profiles.** The structural scannability case is verified in-repo; whether the audit panel passes the generate-audit-regenerate loop on these profiles is a suite run no one has executed. It is the one remaining empirical gate on the annual-exercised claim.
2. **A `SuiteConfig` solver-horizon field.** Not needed for this domain (onset-anchored horizon). If ever added for genuinely slow-onset families, it must be a function of domain-committed inputs to stay inside the determinism tuple (`internal/run/quiescence.go:47-50`) — proven by a test. Who owns that design slice?
3. **The v2 `release_calendar` schema extension.** A machine-readable top-level object is the prerequisite for serving the calendar as policy input. Until it lands, the calendar is 2000-char prose with no schema home (schema :7, :17). Is the v2 extension worth a design change, or does the consumer-side policy framing make it permanent prose?
4. **The supersession/supersession-vs-consequence scoring.** A third solver world over alternate command logs is an explicit design change; artifact replay of logged command runs exists (`internal/run/replay.go:42-58`) but is not truth-side and adds no delivery evidence. Take the design change or leave the question to consumer experiment design permanently?
5. **Enforcing the `EffectorInfo` description boundary.** Keeping calendar prose out of effector/channel nameplate text (`internal/mcp/operator.go:48-54`) is an author constraint today. Should the schema-suite lint check it mechanically?
6. **Ownership overlap with SI-02.** SI-02 pre-committed "SI-02 owns command/ack discipline; merge if SI-09/SI-10 center the same question" (SI-02 doc, Five-lens §5 and Open questions §5). SI-09's final posture — the misfire/overlap matrix as consumer-side measurement over a `slow_physical` substrate — looks compatible, but the split should be confirmed before both domains are authored.
7. **Cost of the 192 h full-phase variant.** Bounded by construction (~1.4× the 96 h profile's ≈120 h world horizon), but the empirical per-scenario cost is an authoring-time measurement not yet taken.
8. **Duration-vs-period as declared data.** The consumer lens required that scenario duration sample a declared fraction of the seasonal period. The schema's `duration_s` (:588) is a bare positive number with no such constraint — enforce by domain convention, or extend the schema?

## Verification performed for this document

Ran during this document's preparation (2026-10-05):

- `ls domains/ | grep -ic reservoir` → `0`, exit 1 (eight shipped domain files listed, none reservoir).
- `grep -nE "maxHorizon|HorizonNS|72\*3600" internal/audit/*.go` → no matches, exit 1: the audit panel carries no horizon constant of its own. (Command correction, folder documentation pass: the originally recorded command lacked `-E`, so `|` was literal and the pattern could only match a string containing a pipe character — a guaranteed no-op on any input. The conclusion is re-evidenced by the corrected `-E` command, re-run during that pass: still no matches, exit 1.)
- `grep -n "seasonality|actuation|time_reference" domains/*.domain.json`: all eight shipped domains declare `"time_reference": "wall"`; only open-field-irrigation carries `"annual"` and `"slow_physical"`.
- `grep -n "dt_ms" internal/schemas/domain-spec-v0.1.schema.json` → :107 (required for F1/F2 per :215), confirming the sketch's integrator tick is authored data.
- Computed with python3: hydrograph arc 12 + 6·sin at 0/45/90/135 degrees = 12.00/16.24/18.00/16.24 m3/s; open-field wire coverage 172800/31536000 = 0.548%; `headwater_drop` 3-sigma/10-sigma crossings 0.25 h / 0.833 h; full-cycle onset-sampling duration 192 h; onset-scan ceiling 72 h/60 s = 4320 SNR evaluations.

Read and verified line-by-line this session (all citations above rest on these reads): `internal/truth/solver_scan.go:19-61` (two worlds at :19-28; onset-anchored horizon at :41); `internal/suite/scenario_draw.go:21-34, 94-100`; `internal/suite/scenario.go:107-119`; `internal/suite/suite.go:61-68`; `internal/suite/generation_setup.go:26`; `internal/score/loop.go:9-20`; `internal/score/loop_metrics.go:8-44`; `internal/world/effects.go:19-31, 56-64`; `internal/world/effectors.go:41-169`; `internal/world/reading.go:23-34`; `internal/world/dynamics.go:95-107`; `internal/world/integration.go:96-129`; `internal/world/world.go:113-129`; `internal/world/oracle_test.go:57-75`; `internal/run/emission.go:12-23, 89-101`; `internal/run/quiescence.go:48-50`; `internal/run/replay.go:42-58`; `internal/truth/solver.go:30-74`; `internal/truth/record.go:29-49`; `internal/mcp/operator.go:46-68`; `internal/domain/catalog.go:75-84`; `internal/domain/catalog_coverage.go:14-35`; `internal/perturb/transform_timing.go:48-61`; `internal/perturb/perturb.go:20-29`; `internal/model/domain.go:259-263`; `internal/schemas/domain-spec-v0.1.schema.json` (:7, :17, :31-34, :107, :114-127, :130-168, :324-334, :370-380, :513-517, :543-544, :547, :557, :588); `internal/schemas/ground-truth-v0.1.schema.json:14-78`; `domains/open-field-irrigation.domain.json:55, :211, :330`; `docs/QUALITY_BAR.md:62-72`; `docs/design/GAP_ANALYSIS.md:46, :61-62, :263`; `docs/design/CRITICAL_REVIEW.md:145-161`; `docs/design/DOMAIN_CATALOG.md:45-48, :65, :88-97, :165-166`; `documentation/reference/contracts.md:21`; `adapters/agentic-stream.adapter.json` (exists).

One citation drift was corrected against the verified file: the 2000-char description bound is at `domain-spec-v0.1.schema.json:17` (the consolidated spec cited :20).

Not run (and not claimed): schema-suite validation of the domain JSON (the artifact does not exist), `make ci-check`, `go test ./...`, and any suite/audit run. The `oracle_test.go:108-183` dead-time oracle range and `internal/model/domain.go:101,111-115` F1-inputs lines are cited from the consolidated spec's round evidence, not re-read line-by-line this session.
