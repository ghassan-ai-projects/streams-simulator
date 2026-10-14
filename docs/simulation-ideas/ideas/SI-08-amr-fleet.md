# SI-08 · `amr-fleet`

- **ID / slug**: SI-08 / `amr-fleet`
- **Date**: 2026-10-05
- **Status**: evaluated — quality bar **not met** (63/100; bar is ≥85/100 with no dimension below 7/10)
- **Serves**: `agentic-stream` (episode admission outcomes) primarily; `tamoz` (budget pressure under load) secondarily
- **Claimed delta**: exactly one measurable candidate — budget-binding admission tempo — carried by a pre-registered tuning contract whose pass/fail is a single named suite run on authored data. Until that run exists the idea has **no verified delta** over the shipped `k8s-cluster` and is judged a decoration-risk proposal under the catalog's own rule.
- **Evidence basis**: every citation below was re-verified against this worktree (branch `new-sim-ideas`, HEAD `867026a`) on 2026-10-05 during this documentation pass; the verification record is in §3. Line drift against the improvement-pass citations was found in three places and corrected here (`internal/suite/commands.go:90-91` not :88-89; `domain-spec-v0.1.schema.json` churn block :67-75; the "linear" enum resolved to the drift-model enum :309 / f1 form enum :137, not :136).

## 1. Purpose

An autonomous-mobile-robot warehouse fleet whose data exercises the situation-lifecycle boundary, and whose **scored surface — stated exactly — is the one this repository can seal today**: detection latency against the sealed onsets plus the consumer's per-event admission dispositions, counters and resource tallies (`internal/model/verdict.go:36-45` — the seven neutral admission outcomes; `internal/schemas/consumer-verdict-v0.1.schema.json:63-79` — the admission array and its seven-outcome enum; `:100-105` — free-form counters "reported verbatim alongside the scorecard and never interpreted").

The flagship consumer question — *does a consumer close a situation whose subject stopped existing, instead of alerting on the silence of a correctly retired robot* — has **no scorer here**, and this spec claims none and carries no unmet acceptance criteria:

1. the born/retired set is world-memory state only: `RetiredNS` is set at `internal/world/availability.go:83`, the field lives at `internal/world/world.go:95`, a repo-wide grep finds no other reference, and `End` publishes only trace, ledger, `world_state_history.jsonl` and verdict (`internal/run/finalize.go:114-128`);
2. the sealed record is `additionalProperties:false` with no churn manifest (`internal/schemas/ground-truth-v0.1.schema.json:7-12` — the required list has no lifecycle field);
3. suite entity sampling renders only the initial ids `1..n` (`internal/suite/entities.go:11-24`).

Building a lifecycle surface would extend a sealed-oracle boundary — a design change first per AGENTS.md — and belongs to a separate effort, not this domain's work.

Retirement here is **background state**. A fault on a robot that retires before `first_observable` fails closed: `Reading` returns 0 for dead entities (`internal/world/observe.go:10-14`), and both solver worlds churn identically on the shared seed and world id (`internal/truth/solver.go:156-167` builds oracle worlds with the same seed and `"w-solver"` id; `internal/world/clock.go:50-58` keys substreams by world id and seed), so `first_observable = 0` and the scenario is excluded as a trivial fixture (`internal/suite/scenario.go:20-22`). No claim below depends on scoring retiree closure. The consumer-facing failure mode — silence that is ambiguous between retirement and transport miss — is still what the data creates: churn births/retirements emit no lifecycle events to the stream, and the only `Birth=true` producer in the binary is producer-flap recovery (`internal/perturb/flush.go:99`; sole grep hit, re-verified this pass). But the consumer's handling of that ambiguity is explicitly **unscored**.

The scored headline is **admission tempo in operator units**, with the definitions pinned:

- The EEMUA number (< 6 episodes/entity/operator-hour, `docs/design/GROUND_TRUTH_AND_SCORING.md:254`) is the **TRUE-EPISODE rate** — `expected_episode=true` labels only, divided by (initial `entity_count` × duration-hours); negative-class controls never enter the numerator.
- **OFFERED-LOAD** (all labels ÷ the same denominator — GROUND_TRUTH_AND_SCORING.md's density companion) is reported alongside as the knob, never substituted.
- The ISA-18.2 flood check (≤ 10 episodes in any 10-minute fleet window, `GROUND_TRUTH_AND_SCORING.md:255`) counts positives-only labels in any exactly-600e9-ns window computed over `first_observable_time_ns` — the declared latency reference; `injection_time_ns` is explicitly not it (`ground-truth-v0.1.schema.json:31-38`) — with the injection basis reported alongside, never mixed.

**Measured stand-in** on the only shipped churn domain, reproduced fresh: two consecutive runs of `streamsim suite -domains-dir domains -domain k8s-cluster -profile nominal -n 100 -seed 7` (fresh build) produced byte-identical `suites/k8s-cluster-nominal-suite.json` (sha1 `0097ad8ae4a6f406104d26d46b237ee1891b493f`), `terminal_state` verbatim, 100 labels / 49 positives / 51 `node_drain` controls, offered-load 0.521 and TRUE-EPISODE 0.255 per entity-hour (8 entities × 24 h, `domains/k8s-cluster.domain.json:253-257`), ISA maxima at the pinned 600e9 ns: 4 positives injection-basis and 3 on the declared `first_observable` basis, positives-only (5 and 6 with controls under an inclusive-window scan), `trivial_excluded` 136 — every tempo figure an order of magnitude below the band where budget decisions bind.

That negative delta is this domain's **one implementation deliverable**: a manifest tuned INTO the binding band under the pre-registered contract in §2.7, verified by one named suite run on its own data.

Per-partition watermarks are **NOT claimed**: no partition concept exists in the contracts or model (grep re-run this pass: 0 hits in `docs/contracts/*.json` and `internal/model/*.go`).

## 2. How it runs on this simulator

Everything below is domain DATA against the shipped schema and runtime (`docs/contracts/domain-spec-v0.1.schema.json`); no per-domain code branches. Where a piece is *not* expressible as data, it says so plainly.

### 2.1 Entities

A churning fleet via `entities.churn` (`docs/contracts/domain-spec-v0.1.schema.json:67-75`). The machinery is live and re-verified:

- births draw from the entity-isolated `churn/births` substream (`internal/world/availability.go:66`);
- birth ids come from `reserveBirthID` (`availability.go:111-119`), a seeded function of the id template;
- lifetimes draw from `churn/lifetimes/<id>` (`availability.go:121-127`);
- retirement drops scheduled emissions by liveness (`availability.go:77-86`; `internal/world/observe.go:70-79` — `emissionChannel` returns nil for dead entities).

`k8s-cluster` is the only shipped domain declaring churn (grep `"churn"` `domains/*.json` → single hit, `domains/k8s-cluster.domain.json:28`: 120 births/hour, 1800 s mean lifetime), and its suite passes (§3, G1).

### 2.2 Signals

- `event_driven` task-state enum channels (per-robot task states);
- per-robot battery state-of-charge scalars feeding the linear integrator (§2.3);
- dock-event counters.

### 2.3 Dynamics and faults

The battery is a **first-order linear integrator**: dx/dt = g·u, closed form x(t) = x₀ + g·u·t, covered by `TestIntegratorAnalyticOracle` (`internal/world/oracle_test.go:57-58`; re-run green this pass: `go test ./internal/world/ -run TestIntegratorAnalyticOracle -count=1` → ok 0.309s). The form is **not new ground**: `"linear"` dynamics already ship in six domains — aquaculture-pond, cold-chain-transit, host-system-health, k8s-cluster, open-field-irrigation, rotating-machinery (grep `"linear"` `domains/*.json`, six files). Two enum homes, verified: the drift-model enum `"model": {"enum": ["linear", "random_walk"]}` at `domain-spec-v0.1.schema.json:309`, and the f1 first-order form enum (including `integrator`) at `:137`.

Two fault families, both expressible in **existing detector forms**:

1. **`parasitic_drain`** — battery slope deviation via `single_channel_snr` (`internal/truth/detector_math.go:10-14`).
2. **`cohort_drift`** — peer residual across ALL live siblings (`peerReadings` iterates `EntityIDs`, `detector_math.go:39-50`; `EntityIDs` is alive-only, `internal/world/clock.go:61-69`). No age/cohort selector exists and adding one would be a new detector form — declared out of scope. Onset magnitude is sized against the all-siblings pooled sigma (`peerSigma`, `detector_math.go:60-63`) in a mixed-age fleet.

The **dock-loop-thrash enum-transition shape is NOT buildable**: all four detector forms read numeric `Readings` only (`detector_math.go:10-37`). It is not claimed.

### 2.4 Perturbations

The storm is a **time-windowed multiplier** (`internal/perturb/transforms.go:16-24` — storm routes to `multiplyRecords`; suite multiplier range 2..5, `internal/suite/commands.go:90-91`). Dock-burst load is therefore a **windowed-multiplier artifact, not dock-event-correlated**, and no torrent claim is made off it.

Tempo knobs are `fault_weights` / `entity_count` / `duration_s` only (`internal/model/effector.go:62-71`, the `Profile` struct; `domain-spec-v0.1.schema.json` profile block).

### 2.5 Ground truth and its analytic oracle

- **Oracle honesty**: `first_observable` is a numeric scan reusing the verified integrator, not a closed form (`internal/truth/solver.go:18-21`, `:70` — `Method: model.SolveNumeric`). No closed-form-onset claim is made.
- **Latency reference**: `first_observable_time_ns` is the declared reference for detection latency; `injection_time_ns` is explicitly not (`ground-truth-v0.1.schema.json:31-38`).
- **Churn truth is NOT in the sealed record**: the record is `additionalProperties:false` with a required list that has no lifecycle field (`ground-truth-v0.1.schema.json:7-12`). Born/retired ids with times remain world-memory only. This domain makes no scoreability claim on retiree closure (§1).
- **Fail-closed on retire-before-onset**: a fault on a robot that retires before `first_observable` yields `first_observable = 0` — `Reading` returns 0 for dead entities (`observe.go:10-14`) and both solver worlds churn identically on shared seed and world id (`solver.go:156-167`; `clock.go:50-58`) — and the scenario is excluded as a trivial fixture (`internal/suite/scenario.go:20-22`).
- **The ambiguity is still created**: churn births/retirements emit no lifecycle events; the only `Birth=true` producer is producer-flap recovery (`internal/perturb/flush.go:99`, sole grep hit). The stream shows exactly the silence a consumer must interpret — but its interpretation is unscored here.

### 2.6 Effectors

- `recall_robot` — advisory.
- `send_to_dock` — `fast_physical`, for loop scoring.

### 2.7 Tempo tuning contract (pre-registered; currently UNMEASURED — no amr-fleet data exists)

**Binding band**: true-episode R ∈ [2.5, 5.0) per entity-hour (≥ 10× the measured 0.255 stand-in, still under the < 6 line) **AND** positives-only `first_observable`-basis max window ∈ [6, 10] — 60–100% of the ≤ 10 ceiling, where any further uniform tempo increase violates a published target.

**Identity behind the band**: under the suite recipe (one episode label per scenario; onsets uniform over the first half of the trace, `internal/suite/scenario_draw.go:26-29`; all scenarios sharing one start instant, `internal/suite/generation.go:37`), expected window occupancy ≈ 1200·N_pos/duration_s and R = N/(n·duration_s/3600). The 1200 s factor defined (folder documentation pass): it is 2× the 600 s ISA window, because onsets draw uniformly over the **first half** of the trace, so a fixed 600 s window's share of the onset support is 600/(duration_s/2) = 1200/duration_s. Identity check on the reproduced stand-in: 1200·49/86400 ≈ 0.68 expected vs max 3 observed (clumping over ~72 windows) — consistent.

**Candidate first search space**: `entity_count` 18–24, `duration_s` 7200–10800, suite N 80–120, `fault_weights` favoring panel-surviving forms. Measured survival on the stand-in (recomputed this pass from the published `admitted` + `trivial_excluded` arrays of the sha1-verified artifact; 236 attempts = 100 admitted + 136 excluded):

| Fault | Admitted/attempts | Survival |
| --- | --- | --- |
| node_drain (control) | 51/51 | 100% |
| config_mismatch | 26/43 | 60% |
| memory_thrash | 13/20 | 65% |
| pod_eviction | 5/51 | 10% |
| crash_loop | 2/35 | 6% |
| network_partition | 3/36 | 8% |

Predicted R ≈ 2.2–3.0 and mean window ≈ 4.4–8.9.

**Pass condition on the authored data**, from one run of `streamsim suite -domains-dir domains -domain amr-fleet -profile nominal -n 100 -seed 7`: measured R and max window inside the band, and a `terminal_state` reduced to AT MOST the two structurally-unreachable cross-profile shortfalls (`correlated_cascade scenarios: 0 < 3; sensor_pathology scenarios: 0 < 10` — unreachable in any single-profile run, see §3); the negative-fraction and pre-degraded shortfalls must be **GONE**.

**Composition levers, data-level only**:

- Declare `ground_truth.negative_class_fraction` **below 0.40**. The audit admits controls at 100% and positives at ~26% (49/185), which fully explains the admitted 0.51 share, and the adaptive sampler (`internal/suite/generation.go:52-65`, floor-clamped at 0.1 sampling) cannot correct a survival asymmetry. E.g. a declared ~0.29 with positive survival ~0.6 lands the admitted share at 0.40 (band checked on the ADMITTED share, `internal/suite/admission.go:38-47`).
- Size `pre_degraded_fraction` to clear 10% **ADMITTED** (9/100 on the stand-in, `internal/suite/admission.go:44-45`) — survival-weighted `fault_weights` is the lever.

Trivial-graded faults stay in as mechanism fixtures per the audit rule (panel: `fixed_threshold`, `zscore`, `first_difference`, `moving_median_residual`, `channel_silence`, `internal/audit/audit.go:20-23`), never as evidence.

## 3. Feasibility against G1–G9

Gates as defined in `docs/QUALITY_BAR.md:62-72`.

| Gate | Status | Evidence |
| --- | --- | --- |
| G1 Determinism | **Pass — reproduced this pass** | Churn draws are entity-isolated seeded substreams keyed by world id and seed (`availability.go:66,123`; `clock.go:50-58`). `k8s-cluster` is the only shipped churn domain (grep single hit, `domains/k8s-cluster.domain.json:28`) and its suite passes. This pass: `go build -o /tmp/streamsim-si08 ./cmd/streamsim` OK; two consecutive `streamsim suite -domains-dir domains -domain k8s-cluster -profile nominal -n 100 -seed 7` runs byte-identical, sha1 `0097ad8ae4a6f406104d26d46b237ee1891b493f` (matches the improvement pass), `terminal_state` verbatim `negative-class fraction 0.51 outside [0.35, 0.45]; pre-degraded fraction below 10%; correlated_cascade scenarios: 0 < 3; sensor_pathology scenarios: 0 < 10`, `trivial_excluded` 136 — all 136 positive faults, zero controls excluded (fault mix re-derived from the artifact); recomputed from the labels array (`Suite.Labels`, `internal/suite/suite.go:54`): 100 labels / 49 positives / 51 controls, offered-load 0.521, true-episode 0.255, ISA maxima 4 (injection) and 3 (`first_observable`) positives-only — matching every prior execution. |
| G2 Analytic cross-check | **Pass for the declared dynamics** | `go test ./internal/world/ -run TestIntegratorAnalyticOracle -count=1` → ok 0.309s (this pass). No closed-form-onset claim: the onset scan is numeric (`solver.go:18-21`, `:70`). |
| G3 Reference consumer | **Honest gap, named** | The shipped reference consumer cannot demonstrate admission tempo — it is a config-driven moving-window numeric detector with no domain knowledge (`internal/refconsumer/refconsumer.go:1-8`); non-numeric values only refresh the presence marker and return (`internal/refconsumer/process.go:44-50`). The loop needs an out-of-process consumer with a budget policy submitting the seven neutral dispositions (`verdict.go:36-45`) and free-form counters (`consumer-verdict-v0.1.schema.json:100-105`, reported verbatim, never interpreted). Debounce/coalesce/defer/expire and budget exhaustion have no simulator-side scoring vocabulary — which is why the vocabulary is neutral. |
| G4 Delivery ledger | **Partial** | `go test ./internal/run/ -run 'Conservation' -count=1` → ok 0.713s (this pass). The >10k evidence is drop+duplicate_burst on a heartbeat (`internal/run/ledger_test.go:115-128`, `TestConservationAtScaleBeyondTenThousand`). Storm-multiplicity conservation at burst rate has no evidence anywhere; the `SOAK=1` soak is gated (`internal/run/soak_test.go:22-25`, drop applied at `:40`) and was NOT run this pass. |
| G5 Quiescence | **Bounded machinery; domain evidence not yet written** | `DefaultQuiescenceTimeout` is a 30 s bound that fails the run (`internal/run/run.go:48-51`; `internal/run/advance.go:46-59`). The review asked for a deterministic burst-advance test at max burst rate on the injectable `QuiescenceClock` as this domain's G5 evidence — not yet written. |
| G6 Sealed oracle | **Boundary respected by not extending it** | The sealed record is `additionalProperties:false` with no churn manifest (`ground-truth-v0.1.schema.json:7-12`); born/retired state is world-memory only (§1, items 1–3). Building a lifecycle surface is a sealed-truth design change first (AGENTS.md) — a separate effort. A G6 prefix-equivalence test for born/retired sets was proposed by the review and is contingent on that design change. |
| G7 Injection probe | **No new surface; not run for this domain** | The domain is data over existing perturbation machinery; it introduces no new delivery path or producer-controlled-string surface. No domain-specific evidence claimed. |
| G8 `silent_no_effect` | **No new surface; not run for this domain** | Both effectors reuse existing advisory / `fast_physical` modes; no new effector mechanism. No domain-specific evidence claimed. |
| G9 Trivial baseline | **Pass as a consumer of the audit machinery** | The panel is the shipped hindsight-fitted five (`internal/audit/audit.go:20-23`). The reproduced run excluded 136 scenarios as trivial fixtures, all positive faults, zero controls — the audit's discrimination is visible in the artifact. Trivial-graded faults stay as mechanism fixtures, never evidence (§2.7). |

**Composition diagnosis, verified in code this pass.** The cascade/pathology shortfalls are structurally unreachable in ANY single-profile run: `countComposition` tests `sc.Profile` equality (`internal/suite/admission.go:30-35`) against a value pinned to the one generated profile (`internal/suite/scenario_draw.go:37`) by a CLI that generates one profile per invocation (`internal/cli/suite.go:13-27`), while `profileShortfalls` demands cascade ≥ 3 AND pathology ≥ 10 unconditionally (`admission.go:50-57`). The shipped tests tolerate non-empty terminal states provided the shortfall is reported (`internal/suite/suite_test.go:103-107`). Fixing that is a suite-generator change outside this spec; until it lands, a `terminal_state` reduced to exactly those two lines is acceptable for this domain's pass condition, while the negative-fraction and pre-degraded shortfalls are data-level and must be tuned out via the levers in §2.7.

**Not run this pass**: the `SOAK=1` million-record soak; storm-shaped >10k conservation; any suite run over amr-fleet data (none exists — the tuning contract is a pre-registered prediction, not a measurement).

## 4. Quality bar result

| Dimension | Score |
| --- | --- |
| consumerValue | 6/10 |
| dataExpressibility | 7/10 |
| determinismOracle | 6/10 |
| ledgerQuiescence | 6/10 |
| novelty | 4/10 |
| sealedTruth | 6/10 |
| specificity | 9/10 |
| **Total** | **44/70 → 63/100** |

**Verdict: not met.** The bar is a total of at least 85/100 across the seven dimensions, with no single dimension below 7/10. The idea fails both conditions: the scaled total 63/100 is well below 85, and five of seven dimensions sit under the per-dimension floor (novelty 4/10; consumerValue, determinismOracle, ledgerQuiescence and sealedTruth at 6/10). Only dataExpressibility (7) and specificity (9) clear the floor — the spec is precise about what it can and cannot claim, but what it can claim is mostly not new.

## 5. Five-lens review

| Lens | Score | What it found | What changed |
| --- | --- | --- | --- |
| Determinism and analytic oracle | 6 | The claim "churn is sealed exactly (born/retired ids with times)" had no home: the sealed label record is closed (`ground-truth-v0.1.schema.json`, `additionalProperties:false`) and carries no churn manifest; `world_state_history` is a director-only pointer. Either churn truth is derived deterministically from (seed, domain) at scoring time — say so and prove replay-equivalence — or a record shape changes, a design change first. The suite's scenario entity pool is the initial entity list (`internal/suite/entities.go:11-24`), so born entities never enter it and "a situation opened for a retired entity is scoreable" had no generative path. A fault on an entity retiring mid-scenario degrades to zero deviation (`observe.go:10-14`) and a trivial verdict (`scenario.go:20-22`) — fail-closed, but the tempo claim had to survive it. Machinery verified good: churn births/lifetimes via named substreams, deterministic DES tiebreak, battery linear drain covered by the integrator oracle. | Deciding where sealed churn truth lives was made a precondition of any scoreability claim, with a G6 prefix-equivalence test and deterministic churn-born id sampling named as the design change's own work. The final spec took the other fork: retiree closure declared unscored, retirement demoted to background state, and the lifecycle surface moved to a separate sealed-truth effort — removing every unmet acceptance criterion from this spec's scope. |
| Data-not-code expressibility | 7 | "Sealed exactly … so scoreable" was asserted, not shown: the sealed truth record is one entity_id + one fault label + three timestamps (`ground-truth-v0.1.schema.json:14-27`); there is no churn-registry record and the suite label is drawn from the faults array. Dock-event-correlated bursts are not expressible: storm is a time-windowed multiplier (`internal/perturb/transforms.go:16-24`; `internal/suite/commands.go:90-91`), not event-triggered. Machinery claims verified: churn fields (`domain-spec-v0.1.schema.json:67-75`; `availability.go:57-69,121-127`), k8s-cluster shipping `cardinality=churning`, battery-slope fault via `single_channel_snr` on a linear integrator, admission outcomes mapping onto the neutral verdict vocabulary. | Where churn truth lives is now stated before any scoreability claim (§2.5); churn-adjacent behavior is scored consumer-side via admission dispositions and counters; dock bursts are shaped with profile timing; the storm is explicitly declared time-windowed and dock-correlation is not claimed. |
| Consumer testing value | 6 | The G3 claim was wrong in detail: the reference consumer does not "handle enum channels" as evidence — non-numeric values refresh the presence marker and return (`internal/refconsumer/process.go:44-50`), so on enum series it produces at most silence detections, never state-transition detections. The whole value — admission tempo tripping real budget decisions — was scenario-density tuning with no committed evidence anywhere in the repo that any domain data was ever tuned to trip budget behavior. Per-partition watermarks are not domain-expressible (no partition concept in schema or runtime). The battery-slope fault is at high trivial risk: a linear drain deviation is exactly what `fixed_threshold`/`zscore` fit with hindsight (`internal/audit/audit.go:21-23`). | The G3 claim was replaced with the honest version: the reference consumer exercises presence/silence, and the admission-tempo loop needs an out-of-process consumer with budget policy submitting the seven neutral dispositions — the neutral vocabulary exists precisely because those outcomes have no simulator-side scoring names. Tempo falsification was defined numerically (episodes/entity/operator-hour + the ISA-18.2 flood check computed from one scratch suite run, entity-isolated churn draws) — which produced the measured stand-in and the pre-registered contract. A second fault shape not expressible as a slope was added (`cohort_drift` via `peer_residual` across live siblings); the dock-loop-thrash enum shape was declared not buildable and mechanism-fixture status accepted where trivial. |
| Delivery ledger and quiescence | 7 | Churn is entity-isolated and deterministic, and retirement drops scheduled emissions by liveness — a retired entity leaves no rows, cleanly distinct from a transport miss. But storm-multiplicity conservation at torrent rate has no evidence: the only >10k conservation evidence is the periodic-heartbeat recipe (`internal/run/ledger_test.go:115-128`), and this domain is the one most likely to produce burst shapes. If the consumer chokes on burst volume the only guard is the 30 s bounded await (`run.go:51`), which fails the run (`advance.go:46-59`) rather than degrading — so "admission tempo is data tuning" needs deterministic burst evidence, not a soak. | Storm-shaped >10k conservation was named as this domain's G4 evidence to add; quiescence evidence was specified as a deterministic burst-advance test on the fake `QuiescenceClock` at max burst rate; and the spec now states outright that budget-exhaustion/admission outcomes are consumer-side — the domain supplies only tempo. |
| Novelty and portfolio balance | 6 | Property-vector collision with the shipped k8s-cluster triggers the catalog's own decoration rule (`docs/design/DOMAIN_CATALOG.md:13`): churning cardinality, event_driven enum channels, F0/linear-F1 dynamics, and the churn machinery are all already shipped, and churning and torrent are each already held twice in the catalog (`:55-56`) — so all three axis "raises" were shipped-relative and the novelty section never checked the catalog. The genuinely new content was admitted to be scenario tuning, not domain data — a weak domain under the stresses rule (`:41-43`). The declared fault observability is the exact single-channel model `docs/design/GAP_ANALYSIS.md` (G-03) calls confidently wrong — defensible for a linear drain, but the sketch had to say why. "First non-pods non-stationary identity" was a thin distinction carrying a whole domain; the kill claim (`terminal-fleet-churn`) was unverifiable (0 grep hits). | The retirement/lifecycle question was made the lead — then withdrawn entirely to unscored when no scorer exists, taking the four unmet bring-up criteria with it. The k8s-cluster delta was restated as ONE measurable candidate (budget-binding admission tempo) with the decoration rule applied to this idea itself: until the tuned run exists, SI-08 is a decoration-risk proposal whose pass/fail is that single measurement, and any axis value declared differently must be TRUE of the authored streams. `cohort_drift` was paired with `single_channel_snr` as a second form; per-partition watermarks were dropped (0 grep hits); the torrent-rate claim was dropped; the kill-claim retraction was verified. |

## 6. Improvement history

The ten improvement rounds are not individually recorded in the material available to this pass; the table below summarises their cumulative effect — the delta between the original consolidated spec and the final spec. Every load-bearing mechanism claim and every verifiable number changed.

| Aspect | Original consolidated spec | Final spec |
| --- | --- | --- |
| Scored surface | "Situation-admission discipline — every admission outcome (debounce, coalesce, defer, expire) must be durably explainable" plus per-partition watermarks. | Scored surface stated exactly: detection latency + admission dispositions, counters, resource tallies (`verdict.go:36-45`; `consumer-verdict-v0.1.schema.json:63-79,100-105`). Per-partition watermarks dropped — 0 grep hits in `docs/contracts/*.json` and `internal/model/*.go`. |
| Retiree-closure flagship | "Churn is sealed exactly (born/retired ids with times), so a situation opened for a retired entity is scoreable" — load-bearing. | Withdrawn: no scorer, no record shape, no sampling path for born ids; retiree closure declared unscored; a lifecycle surface named a sealed-truth design change for a separate effort. The prior four-part bring-up (persisted lifecycle surface, lifecycle-equivalence gate, churn-born id sampling, composition on own data) no longer appears as acceptance criteria, so the spec carries no unmet in-scope criterion. |
| Ground-truth claims | Churn sealed exactly; "closed-form `first_observable` from the slope delta". | Churn truth located in world memory only (closed sealed record, `ground-truth-v0.1.schema.json:7-12`); `first_observable` is a numeric scan reusing the verified integrator (`solver.go:18-21`, `:70`) — no closed form claimed. |
| Rate claims | Dock events "burst at torrent rate"; storm applied on dock bursts. | Torrent claim dropped: storm is a time-windowed multiplier (`transforms.go:16-24`; suite 2..5, `commands.go:90-91`); dock-burst load is a windowed-multiplier artifact; no torrent claim is made. |
| Fault family | Parasitic-drain via `single_channel_snr` only. | Plus `cohort_drift` via `peer_residual` across all live siblings, onset sized against the pooled sigma (`detector_math.go:60-63`); no age/cohort selector — declared out of scope as a new detector form; dock-loop-thrash enum-transition declared NOT buildable (`detector_math.go:10-37`). |
| Novelty | "Raises cardinality=churning to 2, rate=torrent to 2, correlation=independent to 2. First non-pods non-stationary identity. Kills terminal-fleet-churn." | All axis raises dropped (each value already held twice, `DOMAIN_CATALOG.md:55-56`); the retirement differentiator withdrawn; the AMR axis vector shown to duplicate k8s-cluster's shipped vector; kill claim retracted ("terminal-fleet-churn": 0 grep hits, re-verified). Remaining delta: ONE measurable candidate, self-judged by the decoration rule (`:13`) until measured. |
| Tempo claims | "Entity counts and perturbation weights must be tuned in data until the episode rate trips real budget decisions" — a hope. | A pre-registered, checkable tuning contract: binding band (R ∈ [2.5, 5.0); positives-only `first_observable`-basis max window ∈ [6, 10]), the occupancy identity behind it, knob ranges, measured per-fault survival, and the exact pass/fail command on the authored manifest (§2.7). |
| Operator definitions | None. | EEMUA < 6 pinned as the TRUE-EPISODE rate (`expected_episode=true` only, over initial entities × hours); OFFERED-LOAD reported alongside, never substituted; ISA-18.2 ≤ 10/10-min counted positives-only over `first_observable_time_ns` (the declared latency reference, `ground-truth-v0.1.schema.json:31-38`), injection basis alongside, never mixed (`GROUND_TRUTH_AND_SCORING.md:254-255`). |
| Reference consumer | "The reference consumer already handles enum channels shipped domains emit." | Corrected: non-numeric values refresh the presence marker and return (`process.go:44-50`) — silence only, never state transitions. The admission-tempo loop needs an out-of-process consumer with budget policy; the neutral verdict vocabulary exists because those outcomes have no simulator-side scoring names. |
| Feasibility / acceptance | "Machinery verified live" plus the hardest-part admission about tuning. | Per-gate evidence (§3) with a composition diagnosis: the cascade/pathology shortfalls are structurally unreachable in single-profile runs (`admission.go:30-35,50-57`; `scenario_draw.go:37`; `cli/suite.go:13-27`), so the pass condition tolerates exactly those two terminal lines while the negative-fraction and pre-degraded shortfalls are data-level and must be tuned out (declared `negative_class_fraction` < 0.40; survival-weighted `fault_weights`; `pre_degraded_fraction` sized to clear 10% admitted). |
| Dynamics ground | Implicitly novel linear battery ODE. | Stated as the same linear ground six shipped domains already use (grep list in §2.3; drift-model enum `:309`, f1 form enum `:137`), with the analytic oracle re-run green. |

## 7. Open questions

1. **Does the authored manifest land in the band?** The pre-registered pass/fail run (`streamsim suite -domains-dir domains -domain amr-fleet -profile nominal -n 100 -seed 7`) does not exist — no amr-fleet data has been written. Until it does, this idea has no verified delta over `k8s-cluster`.
2. **Where should churn truth live, if ever?** World-history artifact + digest vs a sealed-record extension is a sealed-truth design change (AGENTS.md: design first). It would bring its own G6 prefix-equivalence test for born/retired sets and deterministic churn-born id sampling in the suite.
3. **Suite-generator fix for cross-profile composition.** The cascade/pathology shortfalls are structurally unreachable in single-profile runs (§3); fixing that is a `internal/suite` change outside this spec. Until then the pass condition tolerates exactly those two terminal lines.
4. **Storm-shaped >10k conservation and the SOAK=1 soak** remain unrun; this domain (or its tuned manifest) is the natural shape that would exercise them.
5. **The actual G3 loop** — an out-of-process consumer with a budget policy submitting the seven neutral dispositions and free-form counters — is a consumer-team contract, not domain work.
6. **Portfolio decision after measurement.** If the tempo measurement lands in-band, does the surviving axis vector justify a second churn domain alongside `k8s-cluster` under the catalog rule, or should the delta ship as a `k8s-cluster` profile instead? The stresses rule (`DOMAIN_CATALOG.md:41-43`) applies to whichever answer.
7. **ISA window convention.** The flood check is pinned to positives-only over `first_observable_time_ns` in any exactly-600e9-ns window, but the with-controls inclusive scan yields different maxima (5 and 6 on the stand-in). The exact window-boundary and inclusion conventions should be fixed once before cross-domain tempo numbers are published.
