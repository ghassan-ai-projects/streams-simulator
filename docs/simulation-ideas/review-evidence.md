# Simulation ideas — review evidence

- **Date**: 2026-10-05
- **Status**: review-stage record for the [simulation-ideas portfolio](README.md)
- **Companion**: [research-notes.md](research-notes.md) (raw pool, research angles, briefs)

This file records the adversarial half of the pipeline: the five-lens verdicts per idea (§2),
the lenses' portfolio-level notes (§3), and the final independent scores the improvement loop
was scored against (§4).

## 1. The five lenses

| Lens | Question it asks |
|---|---|
| Determinism and analytic oracle | Is the sealed label actually computable by the shipped truth machinery, with a real G2 cross-check — or is oracle readiness claimed from machinery in the wrong layer? |
| Data-not-code expressibility | Is every load-bearing mechanism data under `domain-spec-v0.1.schema.json` — or does the idea quietly need a runtime concept that does not exist? |
| Consumer testing value | Does the domain grade a consumer behaviour a team can act on, with the graded cells actually reachable — or is value asserted on parsed-only fields and inverted semantics? |
| Delivery ledger and quiescence | Does the run reach terminal ledger state and provable quiescence at this domain's shapes — and is transport-vs-reasoning attribution real? |
| Novelty and portfolio balance | Is the delta real against shipped domains, the 25-domain catalog, and the rest of the portfolio — or is this a decoration under the catalog's own rule? |

**Instrument note.** The per-lens scores grade the draft the lens reviewed; the seven
dimension scores in each idea's *Quality bar result* grade the final spec after the
improvement rounds. They are different instruments and do not always move together (most
visibly SI-04's novelty: lens 4/10 → final 7/10). The portfolio numbers are the final
independent scores, in §4.

## 2. Five-lens verdicts per idea

Condensed from the full *Five-lens review* and *Improvement history* sections of each idea
document, which carry the line citations.

### SI-01 · telemetry-estate — final 70/100, not met (1 round)

| Lens | Score | Found | Changed |
|---|---|---|---|
| Determinism / oracle | 7 | Headline oracle misanchored: the shipped solver labels by SNR scan over four reading-based detector forms; no absence/heartbeat form exists, so "closed-form tick counting" is new oracle work; no sealed-record field was named for transport-vs-death. | Label declared on value physics (`single_channel_snr` on the frozen counter); tick counting spec'd as the G2 independent oracle, asserted equal to the scan label before unblind with named boundary tests; per-scenario transport-vs-death recorded from ledger reasons, fail-closed. |
| Data-not-code | 5 | Load-bearing mechanism not expressible: world faults are state deltas and cannot silence emission; the claimed first-missed-heartbeat detector form does not exist; k8s `network_partition` already ships the freeze pattern. | Fault re-anchored on the counter freeze (the `host_down` pattern); absence arises from ambient availability + `store_and_forward` buffering; a silence/heartbeat detector form deliberately **not** proposed (form-without-equation violates the detector contract). |
| Consumer value | 6 | Silence-only fault auto-trivials (`FirstObservableNS == 0`, excluded); G8 mechanism-only; "three wire-identical worlds" asserted, not evidenced against the audit's hindsight `channel_silence` detector. | One truth story picked (stale-on-time vs late-and-fresh vs lull, re-derived honestly); declared `silent_no_effect` added; ambiguity claim gated on a scratch suite run showing balanced accuracy < 0.9; step-freeze demoted to mechanism fixture. |
| Ledger / quiescence | 7 | Ledger discriminator verified real (dead world emits no rows; `store_and_forward` cannot stall delivery), but `run.End` never flushes perturbation buffers (`finalize.go:30-78`) — a flap ending at/past the last advance silently loses ledger rows; no run-level test exercises flap/backfill at End. | Run-level conservation test with `producer_flap` active at End made a ship precondition (terminal delayed rows reusing `DeliveryDelayed`, or fail the run); `first_observable` forbidden from arrival-time anchoring. |
| Novelty / balance | 7 | Novelty measured only against the 8 shipped domains; `absence=heartbeat` carried by ~8 planned catalog domains; neighbors 03/18 never differentiated; kill/absorb claim and F-002/S1.8 citations unverifiable (0 grep hits). | Declares `absence=ambiguous`; nearest-neighbor delta stated (three-way disambiguation inside one producer's channel set, graded under a scored dispatch deadline); explicit SI-01/SI-05 boundary; kill claim withdrawn; grounding re-anchored on the committed agentic-stream adapter and CRITICAL_REVIEW; mttr-scaled `deadline_s` added. |

### SI-02 · feeder-switching — final 80/100, met as supplied, discrepancy documented (1 round)

| Lens | Score | Found | Changed |
|---|---|---|---|
| Determinism / oracle | 8 | "Apply-events as sealed ground truth" overstated — sealed truth is the closed label record; the compound loading/underfrequency interlock predicate is not expressible (single state, one operator). | Exactly-once oracle re-homed to a run-artifact EffectorCall audit; interlock collapsed to `loading_pct >= 95` with underfrequency dropped; post-expiry duplicate added as the scored boundary cell; metamorphic one-`recordCall`-per-id test specified. |
| Data-not-code | 8 | `duplicate_burst` misattributed — the perturbation layer never sees commands (emissions only); novelty claim partly false (discrete-line-oee ships interlock + `silent_no_effect` + confirmation channels). | Duplicate-command stimulus re-routed to its two real sources (consumer retries under ambiguous acks; driver-replayed command logs) with the missing generator named as a generic suite prerequisite; novelty restated as "no shipped domain makes ack/effect decoupling the primary scored question". |
| Consumer value | 8 | Duplicate mechanism misattributed; the stated hardest part already solved by the EffectorCall record; redundant-sensor wording risked blurring `confirmed_no_effect` (truth-telling) with `silent_no_effect` (counterfactual shadow). | Duplicate cell rewritten as retry behavior resolved deterministically by the idempotency window; confirmed/silent split stated explicitly per mode; kept as the portfolio's reference design for G8; interlock-refusal metric ("refusal produces no retry") made explicit. |
| Ledger / quiescence | 6 | "G4 ledger proves apply-once" impossible — `LedgerRecord` carries no command/effector fields; perturbations process emissions only. | EffectorCall log made the exactly-once authority; delivery ledger demoted to background G4 green; regression test asserting one `recordCall` per distinct `command_id` inside and outside the window specified. |
| Novelty / balance | 7 | "discrete-line-oee has honest acks" false (ships `silent_no_effect` p=0.05); catalog neighbors 14/19 undifferentiated; safety fill asserted in four words; kill claim unverifiable. | Exact residual delta stated (ack_lost ambiguity + in-window duplicates + post-expiry cell + reconcile-to-`unknown`); neighbors named and differentiated; safety scored via `deadline_s` + counterfactual lost to a false-success loop; portfolio pre-commitment recorded (SI-02 owns command/ack discipline); kill claim removed. |

### SI-03 · district-metering — final 69/100, not met (1 round)

| Lens | Score | Found | Changed |
|---|---|---|---|
| Determinism / oracle | 8 | Batch-shaped arrival needs new runtime — `batch_size` is parsed-only and `batch` cadence routes identically to periodic; the solver's conservation detector is between-worlds, so the leak must also be visible in a consumer-formable residual. | Channel wiring stated (district-owned aggregate; residual formable from the wire); byte-identical replay and group-order determinism evidence required for the batch capability; boundary oracle (onset inside a withheld batch labelled from world state) and the G2 negative-recovery case added. |
| Data-not-code | 5 | Shipping today would emit a periodic stream under a `cadence=batch` axis declaration and lie to `catalog coverage`; the headline cross-entity balance ("master − Σ submeters") is not computable by any declared detector form — the solver reads one entity. | Detector re-modelled **single-entity** (one district entity carries the balance channels; aggregate authored data asserted by the cross-check); the batch-emission capability staged as a design change before domain authoring, with the explicit periodic fallback (re-scores below the bar). |
| Consumer value | 6 | `time_reference=batch_relative` decorative (parsed-only, zero runtime consumers); without genuine batch arrival the domain degenerates toward open-field-irrigation's existing conservation-under-gross-lateness coverage. | `batch_relative` claim withdrawn (vector declares `wall`); capability staged first with a regression test proving batch-shaped arrival in the ledger; delivery-invariant `first_observable` stated as the design spine; kill claims withdrawn as unverifiable. |
| Ledger / quiescence | 7 | No ledger semantics defined for a batch; emission counts feed the conservation invariant end to end; the only >10k conservation evidence is periodic-shaped; quiescence itself safe (`observe_only`, bounded windows). | Batch emission unit specified — every record gets its own primary ledger row and unique `DeliveryID`; batches are an emission-count shape, not an aggregation; batch-shaped >10k conservation under drop+`duplicate_burst` required as admission evidence; partial groups flush at horizon end. |
| Novelty / balance | 6 | Fault family is catalog 15 water-distribution's, originally unacknowledged; two of three "zero-coverage" fills shipped-relative only; `observe_only` the only genuine two-sided gap; third conservation-leak in the portfolio. | Repositioned honestly as "catalog 15's conservation fault under batched late delivery" (with the `paint-booth-emissions` cut precedent); `batch_relative` withdrawn; `:656` citation fixed to `:659`; third-leak risk flagged for the run owner. |

### SI-04 · ami-headend-backpressure — final 73/100, not met (2 rounds)

| Lens | Score | Found | Changed |
|---|---|---|---|
| Determinism / oracle | 9 | Counter rollover unmodelled — the mitigation must be a machine-checked data rule, not prose; the release burst must flush exactly once at the advance boundary (an off-by-one scores as a consumer miss). | Closed-form boundary oracle across the flush edge; rollover made a suite-admission invariant computing the register bound from declared `max_rate_kw`/`horizon_s`; attribution invariant named (plateau-interior miss ledger-attributable, tamper label-attributable, never the reverse). |
| Data-not-code | 9 | Producer-side withhold is wire-identical to a stalled pull but must not claim pull-side semantics; detector phrasing wrong (levels, not slope). | Purpose states the producer-side model and disclaims pull-side; detector redeclared as `single_channel_snr` on the register level; width/horizon bounds declared in data with the no-rollover admission check. |
| Consumer value | 7 | One fault family carried the whole judgment value; trivial risk asserted, not shown; the headline load shape is transport, not world, and must not carry fault-family weight. | Three fault families declared on one register substrate (linear, intermittent duty-cycle, bypass); quiescence-after-release made a check with the drained ledger state named exactly; register bounds moved into suite admission. |
| Ledger / quiescence | 6 | `run.End` performs no flush and no run-level test exercises backfill — plateau rows could vanish from the ledger, corrupting the exact separation the domain scores; the >10k claim was asserted, not evidenced. | "Release boundary strictly interior to the run horizon" adopted as a scenario-admission rule; a >10k `gross_backfill`+`duplicate_burst` conservation test committed as the domain's G4 evidence; deterministic burst-shape quiescence test on the fake `QuiescenceClock` written into the G5 plan. |
| Novelty / balance | 4 | Headline novelty contradicted by shipped cold-chain-transit (ships the plateau, "840 buffered readings at one arrival instant"); fifth ledger-as-oracle candidate, third reconnect narrative; kill claim unverifiable and mis-cited. | Novelty re-anchored on the genuine deltas (producer-vs-transport withholding is ledger-attributable; watermark freeze-then-jump as the scored quantity; settlement-integral scoring against sealed per-period truth); binding fold-in contingency; F-002 and the kill claim removed as unverifiable. |

### SI-05 · tank-dip-replay-aliveness — final 77/100, not met (2 rounds)

| Lens | Score | Found | Changed |
|---|---|---|---|
| Determinism / oracle | 6 | Core mechanism not expressible as claimed: `id_reuse` takes only a rate and duplicates the in-flight record with a shifted payload — it cannot re-deliver yesterday's window; `gross_backfill` is a delay, not a replay of old content; the "zero new machinery" framing collapsed. | Replay regime re-spec'd as committed kinds issued through the director command log (`duplicate_burst` ordered **before** `gross_backfill` spanning `[disconnect, reconnect)`); G2 boundary case added (onset inside the masked window; negative-class windows scoring zero labels). |
| Data-not-code | 8 | Same `id_reuse` finding; the heartbeat count arithmetic depended on an unshipped sibling. | Replay window sizing moved into profile data; ledger `Reason` values named as the wire discriminators; `id_reuse` demoted to what it actually does (same-id/different-payload collisions as a second-order toxin); heartbeat arithmetic corrected to standalone +1. |
| Consumer value | 7 | No `silent_no_effect` anywhere (no false-success contribution); "replayed rows delivered as duplicated" imprecise (flush rows read `delayed`; only live duplicate-burst rows read `duplicated`). | `silent_no_effect` added on `raise_loss_report` with the conservation residual as the free confirmation-channel substrate; approval gating replaced with declarable `risk_class=high` + `consequence=regulatory`, approval latency measured from consumer verdicts; verdict table pins exact ledger reasons per row class. |
| Ledger / quiescence | 7 | Replay semantics the oracle needs verified (`flush.go:95-105`); same End-flush exposure as SI-01/SI-04; no detector may read delivered text. | `gross_backfill`+`duplicate_burst` pinned unconditionally with flush conformance cited; End-flush guard test named ship-blocking; benign routine-replay negatives sealed as separate `is_negative_class` scenarios; composition-order conformance test required. |
| Novelty / balance | 8 | Four-surface overlap with SI-01 with no stated boundary; the masked-window fault was the portfolio's third conservation leak; nearest shipped neighbor (cold-chain, the only reconnect dump) uncited; kill claim unverifiable. | SI-01/SI-05 boundary declared in purpose; leak replaced with a biased level sensor (breaking the three-leak pattern and doubling as the mandatory `sensor_pathology` profile); cold-chain positioning stated (there the dump is context to keep, here it is the mask); kill claim withdrawn. |

### SI-06 · ev-plaza-firmware-split — final 74/100, not met (2 rounds)

| Lens | Score | Found | Changed |
|---|---|---|---|
| Determinism / oracle | 7 | Poison content could flip the audit verdict between replay cells unless the audit's treatment of mangled records is deterministic; the divergence channel must be a state observed by a channel (setpoint echo assumed, not specified). | Audit poison policy named (mangled/malformed rows excluded by ledger reason and type, never value inspection) plus a metamorphic test: identical scenario with and without the poison window yields identical labels and `first_observable_time`; setpoint echo channel specified as data via `Channel.Observes`. |
| Data-not-code | 6 | Perturbations are window-global with no entity scoping — the half-fleet cohort premise inexpressible without a schema extension; enum values are random draws, so "new enum state names" is only realizable as `out_of_enum`'s random `undeclared_###` tokens. | Restructured to a whole-plaza temporal split (one global migration window, fault inside it); cohort enum vocabularies dropped; the per-entity multiplicative fault route rejected (one-fault-per-scenario label model would make the migration the labeled fault). |
| Consumer value | 7 | `peer_residual` pools ALL siblings with no cohort filter — "same-dialect peers" not expressible; the poison-window/fault-onset overlap was asserted without a scratch proof. | Detector swapped to `channel_divergence` (commanded vs delivered current) on one charger, supported today; `peer_residual` reserved for SI-10; benign control per poison window defined; one scratch scenario executed under the poison window, admitting non-trivial (best hindsight detector 0.668, below the 0.9 exclusion bar — the poison does not make the scenario trivially solvable; an earlier draft's "the audit still separates derate from controls" phrasing was backwards, since a ≥0.9 score would have meant trivial exclusion). |
| Ledger / quiescence | 7 | Poison kinds committed and terminal-conserving; G7 overstated — the neutrality differential covers the synthetic fixture and discrete-line-oee only, so a new domain gets no neutrality evidence for free. | Domain-specific probe/poison-neutrality differential added to acceptance evidence (delivered bytes differ, verdict byte-identical); "poison overlaps onset in delivery, never in truth" made a sealed scenario rule; >10k conservation with malformed rows added as G4 evidence. |
| Novelty / balance | 8 | Shares its poison/quarantine surface with SI-07; discrete-line-oee already keeps out-of-enum values as declared evidence; the kill claim was unverifiable. | SI-07 boundary explicit (this domain owns drift-driven quarantine under a diagnosis deadline); repositioned against §09 and §18 catalog neighbors; the unit-mismatch counterfactual stated in ground truth (peer comparison across the coerced/uncoerced boundary is the trap); kill claim withdrawn. |

### SI-07 · field-log-injection — final 66/100, not met (10 rounds, cap)

| Lens | Score | Found | Changed |
|---|---|---|---|
| Determinism / oracle | 7 | Ground-truth anchor factually wrong ("the sealed ledger knows which text entered") — `LedgerRecord` carries no payload; the pairing protocol underspecified and not fail-closed. | Injection truth re-anchored on sealed command-log params joined per `delivery_id` under deterministic replay; fail-closed rule explicit (missing cell, unjoinable rewritten row, or any byte differing outside exactly the rewritten ids fails the pair); metamorphic extension and negative-recovery coverage named. |
| Data-not-code | 7 | The planted-text path is empty in generated suites — the sampler draws no `payloads` param and the probe stamps `rewritten` unconditionally while planting nothing; boolean axis arithmetic wrong (1→2, not 1→3). | Payload path pre-committed as four domain-agnostic budgeted changes (corpus, sampler case, stamping fix, fail-closed validation) with acceptance criteria and ordering; arithmetic corrected; discrete-line-oee's `shift_log` cited as the shipped channel pattern. |
| Consumer value | 6 | As a domain, the graded content likely collapses — discrete channels give hindsight `fixed_threshold` ~1.0, so nearly every fault scenario trivials out: "a mechanism fixture wearing a domain slot"; the injection differential needs no new domain to start (discrete-line-oee already hosts an attacker_controlled channel). | Fault given real ambiguity (intermittent duty-cycle divergence on an F1 lag+integrator chain); negative control made detector-visible following rotating-machinery (reproduced: rotating-machinery admits 40/40 with 16 negatives; discrete-line-oee 0.00 negatives, 92 trivial-excluded); byte-equality differential specified precisely; quarantine evidence enumerated to what exists (per-seq admission outcomes). |
| Ledger / quiescence | 6 | The ledger has no payload field and never will without a contract change; the tamoz "durable quarantine" record type does not exist (run artifact publishes trace/ledger/history/verdict only); the G7 fit itself is the pool's strongest. | Sealed-params join chosen; the ledger `payload_digest` extension rejected as a G4/G6-sensitive contract change needing its own design change; neutrality differential extended as a planned regression; quarantine scoped to admission outcomes, no durable quarantine record claimed. |
| Novelty / balance | 4 | Both headline novelty claims false — discrete-line-oee is the shipped attacker_controlled home and catalog 25 auth-and-edr "the native home of the injection probe"; the residual is largely catalog 25 minus its actual novelty (the adversarial-adaptive fault model); weakest marginal world in the set. | Novelty re-scoped to the conditional increment — the paired-cell byte-equality differential as a graded, suite-generated property, reachable only by shipping the budgeted machinery; fold rule explicit (yield the slot to catalog 25 if SI-06 ships the injection-cell machinery); citations and arithmetic corrected against a fresh `catalog coverage` run. |

### SI-08 · amr-fleet — final 63/100, not met (10 rounds, cap)

| Lens | Score | Found | Changed |
|---|---|---|---|
| Determinism / oracle | 6 | "Churn is sealed exactly (born/retired ids with times)" had no home — the sealed record is closed with no churn manifest, and the suite samples only initial ids; a fault on a robot retiring before onset degrades to zero deviation and trivials out. | Where churn truth lives made a precondition of any scoreability claim; the final spec took the other fork — retiree closure declared unscored, retirement demoted to background state, the lifecycle surface moved to a separate sealed-truth effort, removing every unmet in-scope acceptance criterion. |
| Data-not-code | 7 | Scoreability asserted, not shown; dock-event-correlated bursts not expressible (storm is a time-windowed multiplier, not event-triggered). | Churn truth located before claims; churn-adjacent behaviour scored consumer-side via admission dispositions and counters; dock bursts shaped with profile timing; dock correlation not claimed. |
| Consumer value | 6 | The reference consumer does not handle enum channels as evidence (non-numeric values refresh the presence marker and return); the whole value was scenario-density tuning with no committed evidence; battery slope at high trivial risk; per-partition watermarks not domain-expressible (no partition concept). | G3 restated honestly (an out-of-process consumer with budget policy is needed; the neutral verdict vocabulary exists because those outcomes have no simulator-side names); tempo falsification defined numerically — producing the measured k8s stand-in (true-episode 0.255/entity-hour, an order of magnitude below the binding band) and the pre-registered tuning contract; `cohort_drift` added as a second form; watermarks and torrent claims dropped. |
| Ledger / quiescence | 7 | Churn is entity-isolated and deterministic and retirement leaves no rows — cleanly distinct from transport miss; but storm-multiplicity conservation at burst rate has no evidence anywhere (only the periodic >10k recipe), and the 30 s bounded await fails the run rather than degrading. | Storm-shaped >10k conservation named as the domain's G4 evidence to add; quiescence evidence specified as a deterministic burst-advance test on the fake `QuiescenceClock`; budget-exhaustion/admission outcomes declared consumer-side — the domain supplies only tempo. |
| Novelty / balance | 6 | Property-vector collision with shipped k8s-cluster triggers the catalog's own decoration rule; all three axis raises shipped-relative (each value already held twice); the kill claim was unverifiable. | Retirement withdrawn entirely to unscored; the delta restated as ONE measurable candidate (budget-binding admission tempo), self-judged by the decoration rule until the tuned run exists; per-partition watermarks and torrent claims dropped (0 grep hits); kill retraction verified. |

### SI-09 · reservoir-release — final 73/100, met as supplied, discrepancy documented (10 rounds, cap)

| Lens | Score | Found | Changed |
|---|---|---|---|
| Determinism / oracle | 7 | The counterfactual had nowhere sealed to live (two ≤200-char strings); "sealed truth is a deterministic release schedule" conflated layers — the per-run record has no schedule field; "supersession scored against physical consequence" implies scoring policy that does not exist. | Counterfactual named an explicit third solver world (alternate command log) — a design change, not implied machinery; schedule demoted to v2 prose sealed via `domain_digest`; G6 monotonic-unblind test required; the 72 h solver horizon flagged as a determinism-tuple item if ever made configurable. |
| Data-not-code | 6 | The headline sealed truth had no record type; the world has no autonomous schedule executor (dynamics are F0 sinusoids, faults, effector deltas — a piecewise calendar is none of these); the counterfactual needed a solver mode that does not exist (`Solve` runs exactly clean-vs-faulted). | Truth reframed around the gate-failure family and dead-time consequence (existing records); slot-per-scenario casting explicitly rejected; machine-readable `release_calendar` named as the v2 prerequisite, flagged as the largest new-concept load in the portfolio. |
| Consumer value | 7 | The G8 cell was specified backwards — confirmation channels report the counterfactual shadow under `silent_no_effect`, so the original "release counter confirms the halt" would have shown the release that did not happen; supersession reduces to existing loop metrics unless a design change is taken; with period >> duration the annual term is near-linear on the wire. | `release_volume` declared the confirmation channel (the consistent ack lie) with detection on real-state channels plus the false-success metric; supersession restated in terms the scorer already computes (timing vs `effect_eta_ns`/deadline from the EffectorCall log) or named as an explicit design item; the 96 h profile made the annual axis load-bearing (phase-at-onset as the scored signature). |
| Ledger / quiescence | 7 | Delivery risk low and verified; same command-path misattribution as SI-02 — duplicate wakeups and supersession are consumer re-invocation phenomena; perturbations cannot touch commands. | Duplicate-wakeup semantics attributed to the idempotency window and the EffectorCalls authority log (one entry per distinct `command_id`; in-window replay returns the cached result); a skew+delay composition test documents that observed-order is enforced pre-perturbation only. |
| Novelty / balance | 7 | Axis fills shipped-relative (catalog 05 carries annual/slow_physical; catalog 01 owns the actuation-dead-time stress without SI-09 positioning against it); the seepage leak was a third conservation leak the story does not need; kill claim unverifiable. | Catalog 01 named as the dead-time neighbor with the delta stated (01 scores an agent fighting its own lag; SI-09 sets up which command wins under late/skewed/duplicated evidence — consumer experiment design, since no supersession verdict exists); seepage dropped, gate jams made load-bearing; `time_reference=calendar` declared and honestly priced as metadata-only; kill claim removed. |

### SI-10 · tick-feed-integrity — final 77/100, not met (1 round)

| Lens | Score | Found | Changed |
|---|---|---|---|
| Determinism / oracle | 8 | Detector forms undeclared per fault; `crossed_venue_divergence` via peer_residual assumes venue grouping, which is not a solver concept (`peerReadings` walks all entities); "bound volume in profile weights" was a soft guarantee leaving G4/G5 evidence tuning-dependent. | Detector form declared per fault (stall/stale_quote/fat_finger/phantom_depth → `single_channel_snr`; crossed → `channel_divergence` on `(a_mid, b_mid)`); venues reframed as channel families so the solver's actual math applies; data-pinned volume budget (~30–35k events/scenario, ~1.1k per 30 s advance, ~3–4k storm peak) with a deterministic soak. |
| Data-not-code | 7 | Venues-as-entities made `crossed_venue_divergence` non-computable (`channel_divergence` reads two channels of the same entity); "exact update counts via single_channel_snr" conflated ledger facts with detector output; sequence counters as time-integrals advance with the clock, not per emission. | One entity per symbol with per-venue channel families; per-event count truth attributed exclusively to the ledger seq stream; counters declared time-integral ACTIVITY integrals with that caveat (mirroring the counter-reset sidestep). |
| Consumer value | 8 | The G8 wire story was inverted: as the halt's confirmation channel, `a_trade_count` would report the halted counterfactual — the wire would show the halt as effective; volume economics unaddressed (G-07/G-08, the most expensive domain by construction); dedup cells overlap SI-05 and torrent overlaps SI-08. | Lying/telling channel split spec'd — `a_trade_count` reports the frozen counterfactual while `b_trade_count` and both mids report truth: the asymmetry is the falsifiable false-success verdict, strictly stronger than "counters keep moving"; per-scenario cost published (~300 runs ≈ 11M ledger rows), mechanism-only until measured; portfolio partitions vs SI-02/SI-05/SI-08 written into purpose with the residual pairing/join gap stated. |
| Ledger / quiescence | 7 | Torrent duplicate+reorder flows through a synchronous per-row durable ledger append inside the 30 s bounded await; the only >10k conservation evidence is periodic-heartbeat — torrent-shaped conservation evidence, which this domain most needs, does not exist; reorder windows inherit the End-flush caveat. | Torrent+storm conservation beyond 10k rows made ship-blocking G4 gate evidence; per-advance burst size bounded in profile data as a scenario-admission constraint; reorder/delay windows required closed at least one advance before End; `halt_symbol` silent-no-effect evidence named as a test with the corrected wire story. |
| Novelty / balance | 7 | "First non-industrial world" false (host-system-health and k8s-cluster are digital estates); `cadence=human_driven` "fills from zero" shipped-relative and, in the original sketch, decoration — the pattern the catalog rejects; the toll-gantry kill claim had zero grep hits. | Claims restated as defensible: first financial-market world, first venue-consensus (crossed-market) fault family, first domain whose halt interlock reconciles against a human-paced quote; human_driven made load-bearing (the manual quote anchors the halt band); coverage restated from an actual `catalog coverage` run (torrent/interlocked/regulatory 1→2); kill withdrawn entirely with the pairing/join family named as still-unhosted. |

## 3. Lens portfolio notes

The five lenses' portfolio-level conclusions, as recorded by the lens reviewer (formatting
light; content preserved).

### 3.1 Determinism and analytic oracle

The portfolio's common failure mode is claiming G2/G6 readiness from machinery that lives in
the wrong layer: the shipped oracle is a four-form SNR scan over a clean-vs-faulted world
written into one closed label record (schemas/domain-spec-v0.1.schema.json:408,
schemas/ground-truth-v0.1.schema.json additionalProperties:false,
internal/truth/solver_scan.go:40-61), while SI-01, SI-08 and SI-09 ground their value in oracle
semantics that have no detector form and no record field (missed-heartbeat tick counting, churn
manifests, counterfactual trajectories). Requirement imposed before admission: (a) state the
exact detector form and the sealed-record fields the label consumes; (b) name the new
independent closed-form oracle test it adds to `internal/world`, including boundary and
negative-recovery cases (the G2 bar of docs/QUALITY_BAR.md:65); (c) name any new seeded
perturbation or emission machinery as a bounded domain-agnostic capability with its own G1
conformance test — killing any spec whose core mechanism is claimed to "already be committed"
when it is not, as SI-05's replay and SI-07's ledger-as-truth currently were.

### 3.2 Data-not-code expressibility

The portfolio's citations are unusually accurate (schema lines, runtime files, and the
axis-coverage counts all checked out against the built binary and the code), but four of ten
candidates (SI-01, SI-03, SI-08, SI-09) park their headline mechanism in places the runtime
cannot reach — a detector form outside the declared four, a cross-entity balance, a churn
registry, a schedule truth — and two more (SI-06, SI-07) depend on perturbation capabilities
that don't exist (entity-scoped windows, a payload data path). Admission rule imposed: every
candidate must name the ONE existing detector form, computed on the ONE labeled entity, that
produces `first_observable_time` (SI-04 and SI-05 already pass this test), and anything needing
a perturbation-layer or truth-record capability must be decided as an explicit domain-agnostic
design change first — otherwise domains-are-data degrades into domains-that-quietly-amend-the-
runtime.

### 3.3 Consumer testing value

Three of ten candidates rested a headline mechanism on a field the runtime parses and ignores
or constrains backwards — absence-as-fault (the truth solver is emission-blind: detector forms
read state readings, internal/truth/detector_math.go, so a silence fault auto-trivials at
suite/scenario.go:20-21), batch_size/time_reference (parsed-only, model/domain.go:182/:31), and
silent_no_effect confirmation channels (which report the counterfactual shadow, reading.go:23-27,
not the truth — SI-09 and SI-10 both specified their flagship G8 cell inverted). Rule imposed:
admit no candidate whose verdict cannot name the exact committed record it is read from
(ground-truth timestamp, ledger row, or effector-call tuple) and show one scratch
`streamsim suite --domain --profile` audit verdict below the 0.9 bar before it becomes a domain.
Scope note (as recorded): the agentic-stream/tamoz brief ticket IDs (F-002, S1.8, S1.9, F-010,
kill-matrix, watch-expiry) appear nowhere in this repository (grep over docs/ and .agents/
returned nothing; Tamoz appears only as a product name in GAP_ANALYSIS.md:46 and
CRITICAL_REVIEW.md:145), so brief alignment was judged as given by the study; the
killed/absorbed candidate names are likewise not in the repo, so those claims were checked only
for internal consistency and axis arithmetic against the coverage run
(`go build -o bin/streamsim ./cmd/streamsim` then `./bin/streamsim catalog coverage`, which
confirmed every coverage number cited: heartbeat=1, interlocked=1, safety=1, regulatory=1,
batch/human_driven absent, torrent=1, churning=1, independent=1, heavy_tail=1, gross=2,
annual=1, slow_physical=1, text=1, boolean=1).

### 3.4 Delivery ledger and quiescence

Checks run by this lens: `go test ./internal/run -run TestConservationAtScaleBeyondTenThousand
-count=1` (G4, PASS), `go test ./internal/run -run Quiescence -count=1 -v` (G5, 6/6 PASS),
`go test ./internal/run -run TestInjectionProbeNeutral -count=1 -v` (G7, 2/2 PASS); `make
ci-check` and the full `go test ./...` were not run, and nothing here should be read as a
whole-tree release claim. The portfolio treats the ledger/quiescence/probe machinery as a
finished, gate-proven substrate, but four candidates sit on gate-level holes that need a
runtime/contract slice first — `run.End` never flushes perturbation buffers
(finalize.go:30-78), which endangers SI-01, SI-04 and SI-05's withhold/release patterns; the
ledger records no payload (model/event.go:38-49), so SI-07's injection ground truth rests on a
capability that does not exist; and the only >10k conservation evidence is a
periodic-heartbeat recipe, while SI-03/04/08/10 need batch- and burst-shaped conservation proof.
Recommended as the portfolio's first item: a "gate-carrying slice" — End-time buffer
terminality, an injection-provenance record, and shaped conservation tests — with every new
domain declaring `attacker_controlled` shipping its own neutrality differential instead of
inheriting the hard-coded discrete-line-oee case (probe_test.go:131-144).

### 3.5 Novelty and portfolio balance

The set over-concentrates on a single trick — late/duplicated/replayed delivery resolved by the
ledger's transport-vs-reasoning split (SI-01, 03, 04, 05, 07, 10, with command/ack discipline
again in SI-02/09/10) — while leaving the catalog's actually-uncovered frontier untouched. The
coverage run (after `make build`) shows shipped zero-coverage values that NO candidate fills:
rate=sparse, cardinality=singleton/large, value_shape=distribution, lateness=none, fidelity=F2,
and time_reference=batch_relative/cycle_relative/calendar; and DOMAIN_CATALOG.md's only
adversarial-adaptive fault model (25 auth-and-edr, lines 624-634) goes unhosted while SI-07
half-copies its injection half. The catalog's own decoration test ("if two domains have the
same property vector, one of them is decoration", lines 9-14) argues to fold SI-04 into the
SI-01/SI-05 absence family and spend that slot on one frontier domain — an adversarial-fault or
F2/distribution world — instead of a sixth delivery-path variant. Two process debts apply to
all ten: the "kills/absorbs X" names return 0 grep hits in this repo (kill rationale
unverifiable), and the agentic-stream/tamoz cell citations (F-002, S1.8, S1.9, F-010) appear in
no repo doc — those claims were taken on trust, and each candidate's novelty section should be
re-checked against the 25-domain catalog, not just `catalog coverage`'s 8-domain view, before
any is admitted.

## 4. Final independent scores

The seven-dimension quality-bar scores recorded in each idea document's *Quality bar result*
section after the improvement loop (each dimension 0–10; raw sum /70, normalised to /100 as
recorded in the documents). The bar: total ≥ 85/100 AND no dimension below 7/10.

**Where these numbers come from** (provenance added during the folder's documentation pass,
answering "who scored this"): the seven-dimension scores and the two "met" verdicts were
produced by the automated workflow that ran this study — its quality-bar evaluation step and
its post-loop "independent final check" scoring pass. That evaluator is not a tool in this
repository and not a human review board, and its runs are not reproducible from the tree; the
bar itself (≥85/100 total, no dimension below 7/10) is the study's own method definition, not
a repository artifact. What **is** independently checkable here is the code-level evidence each
idea document records — line citations, quoted commands, suite-run outputs — which is what the
five lenses attacked. Treat the dimension scores as one evaluator's structured judgment,
calibrated by the two recorded verdict-vs-bar discrepancies, and the feasibility sections as
the load-bearing content.

| Idea | consumerValue | dataExpressibility | determinismOracle | ledgerQuiescence | novelty | sealedTruth | specificity | Raw | Total | Recorded verdict |
|---|---|---|---|---|---|---|---|---|---|---|
| SI-01 telemetry-estate | 7 | 8 | 6 | 6 | 6 | 7 | 9 | 49/70 | 70/100 | not met |
| SI-02 feeder-switching | 8 | 8 | 8 | 8 | 7 | 8 | 9 | 56/70 | 80/100 | met as supplied¹ |
| SI-03 district-metering | 7 | 6 | 7 | 6 | 6 | 8 | 8 | 48/70 | 69/100 | not met |
| SI-04 ami-headend-backpressure | 7 | 7 | 7 | 7 | 7 | 7 | 9 | 51/70 | 73/100 | not met |
| SI-05 tank-dip-replay-aliveness | 7 | 8 | 8 | 7 | 7 | 8 | 9 | 54/70 | 77/100 | not met |
| SI-06 ev-plaza-firmware-split | 7 | 8 | 7 | 7 | 7 | 7 | 9 | 52/70 | 74/100 | not met |
| SI-07 field-log-injection | 6 | 7 | 6 | 6 | 5 | 7 | 9 | 46/70 | 66/100 | not met |
| SI-08 amr-fleet | 6 | 7 | 6 | 6 | 4 | 6 | 9 | 44/70 | 63/100 | not met |
| SI-09 reservoir-release | 7 | 8 | 7 | 7 | 6 | 7 | 9 | 51/70 | 73/100 | met as supplied¹ |
| SI-10 tick-feed-integrity | 8 | 9 | 8 | 7 | 7 | 7 | 8 | 54/70 | 77/100 | not met |

¹ Verdict as supplied by the workflow's quality-bar evaluation and recorded — not smoothed
over — in the idea documents: **SI-02** sums to 56/70 = 80/100, below the stated 85/100 total
while holding the per-dimension floor (minimum 7, novelty); **SI-09** sums to 51/70 = 73/100
and fails both clauses (novelty 6/10). Taken literally against the stated bar, **no idea meets
it**; the eight "not met" verdicts fail on the total (all of them), on the floor (SI-01, SI-03,
SI-07, SI-08), or both. The recorded totals are the documents' normalised values (raw/70
rounded; e.g. 48/70 → 69).

## 5. Provenance of this record

All passes on 2026-10-05, branch `new-sim-ideas`, worktree clean at `867026a`.

**Index pass** (wrote the three frames and the `docs/README.md` §7 row):

- Read all ten idea documents in full and the archive index; confirmed `domains/` holds
  exactly the eight shipped domain files and no candidate artifact exists.
- Transcribed the five-lens verdicts and the seven-dimension scores from the idea documents;
  cross-checked every dimension table and total in §4 against the portfolio record supplied to
  that pass — all ten agree.
- §3 preserves the lens reviewer's portfolio notes as recorded, including the tests that lens
  ran in its own session (§3.4); those tests were **not** re-run.

**Correction pass** (same day; fixed the defects a first-read review of the folder found,
re-reading code and re-running checks where documents contradicted each other — each
correction is also marked inline where it lands):

- **Reconciled SI-02's status**: header corrected from "idea accepted" to evaluated/not
  admitted with the recorded verdict kept and the discrepancy stated; README §1 now defines
  *met the bar* vs *admitted*.
- **Scoped the per-idea "zero grep hits" claims** to their writing passes and re-ran the
  killed-name greps with `-E` (`grep -rEn` over the nine names, `--include=*.go/*.md/*.json`):
  zero hits outside `docs/simulation-ideas/` — the withdrawals' substance holds; annotated
  SI-01, SI-02, SI-03, SI-06, SI-10.
- **Corrected two recorded no-op greps** (missing `-E`, so `|` was literal): SI-10's
  toll-gantry/custody-transfer check and SI-09's horizon-constant check; both re-run with `-E`
  this pass — conclusions unchanged (zero hits; no horizon constant).
- **Resolved the batch-cadence contradiction from code**: `internal/domain/
  validation_records.go:32-35` rejects channel `cadence=batch`; the `batch→periodicEmission`
  case at `internal/world/cadence.go:22-24` is unreachable dead code behind it. SI-03 and
  research-notes now state both halves.
- **Reconciled SI-02's suite-run evidence** with the structural shortfall diagnosis:
  re-read `internal/suite/admission.go:50-58` (unconditional cascade ≥3 / pathology ≥10) and
  `internal/cli/suite.go:22` (one profile per invocation) this pass; SI-02 now flags the
  omitted terminal-state lines and the throwaway domain's unrecorded disposition (no
  `feeder-switching` file exists under `domains/`, checked this pass).
- **Named SI-03's rubric scorer** from code re-read: `internal/score/judgment.go:12-39`
  (offline detection grading vs the sealed record; `Detection` at
  `internal/model/verdict.go:27-34`) — and quoted the `catalog coverage` output SI-03's
  header had promised but never shown (re-run this pass).
- **Read review finding R2** at `docs/reviews/ARCHITECTURE_REVIEW.md:127-146`: a G2 coverage
  gap (link-delay models lack analytic oracles) with an oracle-or-scope-decision
  recommendation, not a usage ban; SI-01 now states its position and SI-05 its refusal.
- **Explained the schema paths**: `docs/contracts/` holds the committed originals;
  `internal/schemas/` embeds byte-identical copies, enforced by `TestEmbeddedMatchCommitted`
  (`internal/schemas/schemas_test.go:16-37`); verified byte-identical this pass (`diff -q`
  exit 0). Detector-form enum is at line 408 (grep, this pass) — the higher of the discrepant
  citations.
- **Reconciled SI-07's coverage arithmetic** against the sampler code (commands.go:37-43,
  generation.go:67-74, admission.go:60-66, all re-read this pass): coverage counts admitted
  scenarios only, which matches the rotating-machinery tail exactly and the discrete-line-oee
  tail to one kind; disclosed SI-07's G2 gap (no second, independent implementation of its own
  onsets).
- **Fixed the SI-06 "separates" wording** (0.668 < 0.9 is the non-trivial pass condition, not
  evidence of separation), **defined SI-08's 1200 s factor**, **corrected research-notes'**
  "seven ideas" to the angle's seven candidates, and **annotated the pre-index sibling
  claims** (SI-04 open Q2, SI-05 header/census/open Q3) as pass-relative.
- Documentation-only throughout: nothing outside `docs/` was modified, and no test suites,
  builds, or soak runs were executed.
