# Simulation ideas — research notes

- **Date**: 2026-10-05
- **Status**: research-stage record for the [simulation-ideas portfolio](README.md); preserved as produced
- **Companion**: [review-evidence.md](review-evidence.md) (five-lens verdicts and final scores)

> Provenance note: this file records the research stage of the pipeline as it was produced.
> The five-lens review and the improvement loop later corrected several of its feasibility
> claims — most visibly the research angles' "zero new Go code" claims, which the lenses
> shrank to a much smaller expressible residue per idea. Two further corrections were made
> during the folder's documentation pass and are marked inline: (a) the batch-cadence lines
> (§2.1, §2.3) recorded the loader rejection (`internal/domain/validation_records.go:32-35`)
> without the dead routing case (`internal/world/cadence.go:22-24`) that SI-03's pass
> recorded — both halves are reconciled in place, from code re-read; (b) one grep quoted in
> §1.1 (via SI-10) was recorded without `-E` and was a literal-pattern no-op as written — the
> corrected command and its re-run are recorded in the dispositions table. The corrected
> record lives in each idea document and in [review-evidence.md](review-evidence.md); this
> file is the input the correction pass worked from. See [README.md](README.md) §5 for the
> write order (research → per-idea → index/correction passes, all 2026-10-05).

## 1. Raw candidate pool

Twenty-five candidates were produced across the four research angles (§2). Self-scores are the
research stage's own 0–10 fit estimate, recorded before any adversarial review. "Admitted as"
links the candidate to its portfolio idea; all other dispositions are in §1.1.

| # | Candidate | Angle | Self-score | Pitch (condensed) | Disposition |
|---|---|---|---|---|---|
| 1 | fleet-telemetry-gateway | A | 9 | 30–60 moving assets behind a `store_and_forward` uplink; separate powered-off vs gateway-buffering vs deadband-lull from heartbeat absence | consolidated into SI-01 |
| 2 | ev-charging-plaza | A | 8 | 40 charge points, human-paced plug-ins; session boundaries as sealed time-structure; sessionization under disorder | cut |
| 3 | custody-transfer-pairing | A | 8 | GPS trip closures × dock scans paired in an event-time interval; the multi-stream join is the only evidence surface | cut |
| 4 | pump-station-relay-bank | A | 8 | 12 pump panels; applied-but-unacked vs acked-before-applied vs duplicate-id command reconciliation under interlocks | consolidated into SI-02 |
| 5 | grain-storage-backfill | A | 7 | catalog 05 built as designed: weeks of sparse hourly data with buffered bursts whose event times precede any watermark | cut |
| 6 | toll-gantry-transactions | A | 7 | gantry passes at torrent rate with a churning vehicle population; at-least-once dedup with id-reuse as the adversarial cell | cut |
| 7 | lab-analyzer-recalibration | A | 6 | reagent-lot changes shift the observation scale mid-run; state/schema migration as a streaming problem | cut |
| 8 | telemetry-estate | B | 9 | sparse remote-telemetry estate; absence itself as evidence — stale-on-time counter freeze, late-and-fresh buffering, benign lull | admitted as SI-01 |
| 9 | feeder-switching | B | 9 | radial distribution feeder; interlock-gated switch commands exercising the full ack/effect decoupling and idempotency family | admitted as SI-02 |
| 10 | amr-fleet | B | 8 | churning autonomous-mobile-robot warehouse fleet; bounded reasoning budgets and admission tempo under load | admitted as SI-08 |
| 11 | district-metering | B | 8 | batched water-district metering; a four-cell window-policy matrix against a conservation leak of exactly known magnitude | admitted as SI-03 |
| 12 | reservoir-release | B | 7 | 2–4 peer reservoirs on a compressed annual hydrograph; schedule drift, misfire/overlap, 600 s gate dead time | admitted as SI-09 |
| 13 | qc-release | B | 7 | regulated QC lab; human_driven cadence + batch_relative time; batch release must not fire while completeness is provisional | cut |
| 14 | field-log-injection | B | 8 | attacker_controlled log channel; planted instructions must never move policy-gated intents | admitted as SI-07 |
| 15 | distribution-grid-bad-data | C | 8 | substation + feeders under mixed PMU/SCADA cadence; conservation-oracle bad-data detection under 5–20% missing data | cut |
| 16 | tick-feed-integrity | C | 9 | consolidated market feed; torrent dedup and sequence-gap integrity, halt interlock reconciled against a human quote | admitted as SI-10 |
| 17 | arterial-detector-network | C | 7 | ~1,200 loop detectors; the standard detector-fault taxonomy and threshold-integrator queue oracles | cut |
| 18 | air-quality-network | C | 7 | 1,000+ node regulatory air network; sensor drift vs a reference station that can itself be the liar | cut |
| 19 | terminal-fleet-churn | C | 8 | payments terminal fleet with k8s-style churn; ambiguous silence and reboot ack/effect decoupling | cut |
| 20 | infusion-fleet-safety | C | 8 | hospital smart-pump fleet; silent_no_effect and exactly-once at physical stakes; one-compartment PK oracle | cut |
| 21 | metocean-buoy-partition | D | 8 | buoy fleet through a dilution→reconvergence partition; Jepsen's reconvergence question with drift inside the gap | cut |
| 22 | ami-headend-backpressure | D | 9 | AMI head-end stalls hours then releases one-instant plateaus; watermark freeze-then-jump vs sealed settlement integrals | admitted as SI-04 |
| 23 | toll-gantry-pairing | D | 8 | two gantries read churning tags; exactly-once entry/exit pairing vs cloned-tag alarm | cut |
| 24 | ev-plaza-firmware-split | D | 9 | plaza-wide firmware migration window (unit labels, enum tokens, poison records) hiding one genuine derate fault | admitted as SI-06 |
| 25 | mesonet-floodfront-skew | D | 7 | weather mesonet watches a flood front; reconstruct true cross-station ordering under windowed clock skew | cut |

### 1.1 Dispositions — what was cut and why

Nine candidates were admitted directly; SI-05 `tank-dip-replay-aliveness` has **no raw-pool
entry** in the recorded material — it entered the portfolio during consolidation as the
replay-mask idea, its only recorded consolidation link being a metocean-buoy kill claim later
withdrawn (below). The sixteen cuts fall into three recorded groups:

**Consolidated into an admitted idea (2).**

- **fleet-telemetry-gateway → SI-01.** Same absence/gateway ambiguity story. SI-01's
  improvement history records that its original "absorbs killed fleet-telemetry-gateway" claim
  was withdrawn as unverifiable (zero repository references); the consolidation itself is what
  SI-01 documents.
- **pump-station-relay-bank → SI-02.** Same interlocked-command world. SI-02's improvement
  history likewise retracts the kill claim (grep over the repository: zero references), while
  keeping the ack/effect-decoupling question as its own subject.

**Recorded kill claims, all withdrawn as unverifiable (6).** Each idea document withdraws the
claim after grep returned zero repository hits for the killed name; the dispositions below are
therefore recorded as "cut in consolidation", with the claimed killer named for the record:

| Cut candidate | Kill claimed by | Status of the claim |
|---|---|---|
| ev-charging-plaza | SI-03 | withdrawn as unverifiable (SI-03 improvement history); its sessionization premise is hosted by no idea; its sibling candidate (ev-plaza-firmware-split) became SI-06 |
| grain-storage-backfill | SI-03 and SI-04 | withdrawn as unverifiable in both (SI-03 §5, SI-04 §5) |
| lab-analyzer-recalibration | SI-06 | withdrawn as unverifiable (SI-06 §5); the pool's self-score of 6 — the pool's lowest — is the only externally consistent part of the "scored lowest at 6" claim |
| terminal-fleet-churn | SI-08 | retracted, 0 grep hits (SI-08 §6) |
| metocean-buoy-partition | SI-05 | withdrawn as unverifiable (SI-05 §5) |
| toll-gantry-pairing | SI-10 | withdrawn as unverifiable. (Command correction, documentation pass: the grep SI-10 §3 originally quoted lacked `-E`, so `toll-gantry&#124;custody-transfer` matched only the literal string containing a pipe — a guaranteed no-op on any input. Corrected command re-run: `grep -rEn "toll-gantry&#124;custody-transfer" --include="*.go" --include="*.md" --include="*.json" .` → zero hits outside `docs/simulation-ideas/`, so the withdrawal's substance holds on the corrected evidence.) |

**Cut with no recorded rationale (8).** The recorded material gives no explicit cut reason for
custody-transfer-pairing, toll-gantry-transactions, qc-release, distribution-grid-bad-data,
arterial-detector-network, air-quality-network, infusion-fleet-safety, or
mesonet-floodfront-skew; the pool entries above are preserved as the record. Two documented
consequences are worth stating plainly, because the review stage recorded them as the
portfolio's residual gaps:

- the **pairing/join family** (custody-transfer-pairing, toll-gantry-pairing) stayed unhosted —
  SI-10's purpose states the residual gap explicitly (`peer_residual` is sibling-deviation, not
  pairing);
- the **coverage-fill worlds** (air-quality-network: sparse/large/none/annual/heartbeat/
  regulatory in one file; distribution-grid-bad-data: rate=high, report_by_exception;
  arterial-detector-network: cardinality=large) were cut, and the novelty lens lists
  rate=sparse, cardinality=singleton/large, value_shape=distribution, lateness=none, fidelity
  F2, and the calendar time references among the shipped zero-coverage values **no admitted
  candidate fills**.

## 2. Research angle findings

Four angles ran in parallel. Each combined web research with repo grounding; the repo
grounding was re-derived independently by all four (the catalog coverage counts below are
theirs and agree). Sources are linked; limitations are stated per angle, not buried.

### 2.1 Angle A — Event-time semantics pain

**Scope.** Out-of-order and late data, watermark behaviour, exactly-once vs at-least-once,
replay/backfill, state and schema migration, sessionization, multi-stream joins — mapped onto
streams-simulator's data-defined domain surface and the testing gaps of agentic-stream and
tamoz.

**Findings (external sources).**

1. The completeness problem is unsolvable in general — systems must embrace refinement, not
   certainty. Streaming 101 (Akidau,
   [oreilly.com/radar/the-world-beyond-batch-streaming-101](https://www.oreilly.com/radar/the-world-beyond-batch-streaming-101/)):
   event-time vs processing-time skew is "non-zero and variable" (phones leaving airplane
   mode); event-time windowing is "the gold standard" but carries the completeness problem:
   "how can you determine when you've observed all the data for a given event time X? For many
   real-world data sources, you simply can't." Sessions are "dynamic, unaligned windows …
   event bursts ended by an inactivity gap".
2. Watermarks trade latency against correctness and are always wrong somewhere. Streaming 102
   ([oreilly.com/radar/the-world-beyond-batch-streaming-102](https://www.oreilly.com/radar/the-world-beyond-batch-streaming-102/)):
   a watermark of X asserts all data with event times < X has been observed; watermarks can be
   "too slow" (latency) or "too fast" (incorrect results from stragglers); "You simply cannot
   get both low latency and correctness out of a system that relies solely on notions of
   completeness." Allowed lateness bounds window lifetime; records beyond the horizon "are
   simply dropped"; accumulation modes include "accumulating & retracting" — "I previously told
   you the result was X, but I was wrong".
3. Flink semantics a consumer must honor
   ([nightlies.apache.org/flink/flink-docs-stable/docs/concepts/time](https://nightlies.apache.org/flink/flink-docs-stable/docs/concepts/time/),
   …`/generating_watermarks/`, …`/operators/windows/`, …`/operators/joining/`): Watermark(t)
   declares event time has reached t; two-input operators take the minimum of both inputs;
   `withIdleness` exists because a silent split stalls the global minimum — and a resumed idle
   source "can pull the combined watermark backward until it catches up"; watermark alignment
   exists because a fast input forces unbounded state in windowed joins; session windows merge
   on gaps and "late firings … may bridge the gap between two pre-existing, unmerged windows";
   default allowed lateness is 0 (drop), side outputs capture the discards; interval joins are
   inner-only and event-time-only.
4. Deterministic event-time testing is a solved recipe everywhere except in the system under
   test. Apache Beam TestStream ([beam.apache.org/blog/test-stream](https://beam.apache.org/blog/test-stream/)):
   scripts element adds, watermark advances and processing-time advances, each event run to
   quiescence, so tests are deterministic while emulating nondeterministic arrival; it
   enumerates exactly the cells that matter — on-time, "unobservably late", late-within-allowed-
   lateness (refinement), "droppably late" (dropped), and speculative-trigger panes.
   streams-simulator's quiescence barrier (G5) and perturbation layer are the same idea from
   the world side; the gap is that no shipped domain makes event-time structure itself the
   evidence.
5. Exactly-once is end-to-end only with application cooperation. Confluent
   ([confluent.io/blog/exactly-once-semantics-are-possible-heres-how-apache-kafka-does-it](https://www.confluent.io/blog/exactly-once-semantics-are-possible-heres-how-apache-kafka-does-it/)):
   idempotent producer sequence numbers survive leader failover; transactions fence zombie
   producers; but guarantees stop at Kafka's edge — "if the app makes an RPC call to update
   some remote stores … the resulting side effects would not be guaranteed exactly once".
   Exactly the boundary where tamoz's three-valued reconciliation and agentic-stream's
   event_inbox/idempotency live.
6. State/schema migration is brittle and key structure is frozen. Flink schema-evolution docs
   ([nightlies.apache.org/…/fault-tolerance/serialization/schema_evolution](https://nightlies.apache.org/flink/flink-docs-stable/docs/dev/datastream/fault-tolerance/serialization/schema_evolution/)):
   "Schema evolution of keys is not supported"; only POJO/Avro evolve; Kryo "cannot be used for
   schema evolution".
7. Sessionization pain (search-derived, second-hand — the primary page 404'd): gap duration is
   "often cited as the hardest part of sessionization"; session windows are "particularly
   sensitive to late data since late events can extend closed sessions or merge previously
   separate sessions". One production-tuning anecdote surfaced via search (unverified):
   maxOutOfOrderness=5s dropping >10% of mobile events as late.
8. Honest limitation: WebSearch was rate-limited (HTTP 429) in early rounds and several
   targeted URLs 404'd (Conduktor session-windows slug, s.apache.org short links, Flink master
   docs; Kafka dsl-api returned empty content). Public postmortems specifically on
   watermark/late-data incidents did not surface — search itself noted such writeups are rare.
   Item 7 and the >10%-late quote rest on search-result snippets, not full-page reads.

**Repo grounding (commands run).**

- `go run ./cmd/streamsim catalog coverage` — thin list: absence=heartbeat(1), absence=normal(1),
  actuation=interlocked(1), actuation=slow_physical(1), cadence=report_by_exception(1),
  cardinality=churning(1), consequence=regulatory(1), consequence=safety(1),
  correlation=independent(1), lateness=heavy_tail(1), rate=torrent(1), seasonality=annual(1),
  value_shape=boolean(1), value_shape=text(1); zero-coverage values (sparse/high rate,
  singleton/large cardinality, distribution shape, batch/human_driven cadence, lateness=none,
  batch_relative/cycle_relative/calendar time, observe_only, F2/F3) are absent from `by_axis`
  entirely.
- `internal/world/link_delay.go:9-21` — store_and_forward delay = MeanS + Exp(MeanS): a
  buffered-burst backfill is a per-channel data parameter today.
- `internal/perturb/perturb.go:20-48` — all 19 perturbation names, including duplicate_burst,
  id_reuse, reorder, delay_tail, gross_backfill, clock_skew, non_monotonic, producer_flap,
  injection_probe — every event-time corruption the research surfaced is already a
  data-declared delivery fault.
- `internal/world/cadence.go:10-31,43-89` — report_by_exception deadband, event_driven trigger,
  periodic+jitter, human_driven exponential inter-arrival draws (seeded, deterministic):
  session-like arrival structure needs no new code.
- `docs/design/DOMAIN_CATALOG.md:159-176` (grain-storage 05), `:185-192` (cold-chain 06, the
  store-and-forward birth-burst story), `:339-359` (solar 13), `:385-406` (water-distribution
  15, "the fault is only visible in a conservation law across entities").

Constraints honored per idea: all truth via the four existing detector forms
(`internal/truth/solver.go:65-141`), arrival structure via cadence/link-delay/perturbation
data, no batch cadence — loader-rejected (`validation_records.go:32-35`, re-read during the
documentation pass: `channelDelivery` errors on `Cadence.Mode == "batch"`; the `batch` case
at `internal/world/cadence.go:22-24` is unreachable dead code behind that rejection — this
note recorded the rejection half, SI-03's pass the routing half; reconciled in SI-03 §2), no
F2/F3, no enum-content-dependent truth (enum values are random draws,
`reading_values.go:74-80`).

### 2.2 Angle B — Testing agentic and durable consumers

**Scope.** Situation lifecycle triggers, bounded reasoning budgets, policy-gated intents,
idempotent command effects, checkpoint/resume under stream faults, schedule drift — hosted as
new data-only domains.

**Findings (six searches, Oct 2026, external sources).**

1. Event-time/late-data tuning remains the canonical stream-testing pain: watermarks trade
   accuracy vs latency under clock skew and delayed arrivals (Flink "Timely Stream Processing"
   docs, nightlies.apache.org; Conduktor watermark glossary; RisingWave watermark post) — the
   consensus is that late-data branch coverage requires deliberately shaped disorder that ad-hoc
   fixtures do not produce.
2. Durable execution: frameworks give at-least-once; exactly-once effects depend on application
   idempotency keys and dedup, and determinism is a straitjacket where replay versioning breaks
   workflows; subtle failure modes appear only under crash/retry harnesses like Temporal's
   time-skipping test server and Restate's testkit (matthewswong.com durable-execution post;
   prachub.com exactly-once analysis; ai.engineer Temporal activity guidance).
3. Agentic testing: agents fail subtly with no concrete error and need the
   trajectory/outcome/cost eval triple, hard caps on steps/spend, and graceful escalation on
   budget exhaustion (newline.co unit-testing AI agents; zenml.io production monitoring;
   padiso.co agent anatomy; calibreos.com agentic eval guide).
4. Schedule drift: DST/misfire handling is a recurring bug farm (Hangfire DST thread,
   discuss.hangfire.io; croniter issue #61; Quartz misfire-policy docs), with
   deterministic/simulated clocks the standard remedy — exactly the controlled-time suite
   tamoz's P13 §9 wishlist still owes.
5. Checkpoint/recovery: standard chaos checks return-to-steady-state, not behaviour DURING
   recovery (ACM "Verifying Transient Behavior Specifications in Chaos Engineering");
   post-recovery correctness needs output validation against an oracle (Conduktor
   chaos-for-streaming guide; Flink fault-tolerance docs).
6. Situation lifecycle: alert correlation needs temporal dedup plus topology-aware inhibition;
   naive content dedup folds unrelated incidents into one "situation" (Site24x7/ilert
   correlation docs; Google Chronicle dedup docs).

**Repo grounding (commands run).**

- `go run ./cmd/streamsim catalog list` → 8 shipped domains (aquaculture-pond,
  cold-chain-transit, discrete-line-oee, host-system-health, k8s-cluster, open-field-irrigation,
  rotating-machinery, shipment-exception).
- `go run ./cmd/streamsim catalog coverage` → zero coverage at cadence=batch/human_driven,
  time_reference=batch_relative/cycle_relative/calendar, actuation=observe_only,
  value_shape=distribution, cardinality=singleton/large, rate=sparse/high, lateness=none,
  fidelity=F2/F3; thin (1 domain) at absence=heartbeat/normal, actuation=interlocked/
  slow_physical, cadence=report_by_exception, cardinality=churning, consequence=regulatory/
  safety, correlation=independent, lateness=heavy_tail, rate=torrent, seasonality=annual,
  value_shape=boolean/text.
- Schema greps of `docs/contracts/domain-spec-v0.1.schema.json`: batch/human_driven cadence
  (lines 28, 277, batch_size 281), entities.churn (line 67), attacker_controlled channels
  (line 335, used at `domains/discrete-line-oee.domain.json:59`), effector failure modes
  ok/slow/ack_lost/reject/partial/confirmed_no_effect/silent_no_effect with mandatory
  confirmation_channels (lines 485-516), interlock with autonomous_action (494-505),
  idempotency_window_s default 3600 (547), link model store_and_forward (329). Churn machinery
  is live and used by exactly one domain (`grep -l churn domains/` → k8s-cluster only; block
  `{births_per_hour:120, mean_lifetime_s:1800}`); the 19 perturbation names at
  `internal/perturb/perturb.go:21-39`; all four detector forms used by shipped domains;
  combination rules at `internal/truth/solver.go:126-127`;
  `adapters/agentic-stream.adapter.json:1-6` confirms consumer wire knowledge stays in adapter
  data only.

**Limitation.** NOT RUN: any test or command inside agentic-stream or tamoz — all claims about
those two repos come from the briefs supplied for this study (§3). The research stage's
conclusion that "all seven of this angle's candidates need zero new Go code" (the count is
this angle's seven pool candidates — §1, angle B — not the ten portfolio ideas) was the
round's most consequential claim;
the five-lens review later corrected it idea by idea (see
[review-evidence.md](review-evidence.md)).

### 2.3 Angle C — New physical/operational worlds

**Scope.** Worlds absent from the 25-domain catalog (power grids, financial ticks, traffic
networks, environmental monitoring, payments/branch device fleets, healthcare device fleets),
each judged strictly by whether it is expressible as domain DATA under
`domain-spec-v0.1.schema.json` with an analytic oracle — no new world code, no per-domain
branches, wall-clock banned.

**Repo grounding (all run this session).**

1. `go run ./cmd/streamsim catalog coverage` reproduces the measured gaps (same thin/zero lists
   as §2.1/§2.2; all 8 shipped domains declare `time_reference=wall`).
2. `go run ./cmd/streamsim catalog list` confirms the shipped set and that discrete-line-oee
   owns both torrent=1 and interlocked=1, k8s-cluster owns churning=1, shipment-exception owns
   heavy_tail=1 (`domains/shipment-exception.domain.json:9-18`).
3. Expressibility envelope verified in schema + code: F1 forms are exactly first_order_lag,
   rc_network, dead_time, saturation, hysteresis, integrator, threshold_integrator (schema line
   136; RK4 at `internal/world/integration_forms.go:80-88`); detector forms
   single_channel_snr/channel_divergence/peer_residual/conservation_residual (schema line 407);
   conservation_residual proven per-entity in shipped data —
   `domains/discrete-line-oee.domain.json:96` and `domains/open-field-irrigation.domain.json:192`
   — so a conservation law over ONE entity's channel set is data-expressible today (cross-entity
   district leaks, as catalog 15 wants, are not). human_driven cadence is implemented
   (`internal/world/cadence.go:27`) and only `batch` is loader-rejected
   (`internal/domain/validation_records.go:33-34`; the `batch→periodicEmission` case at
   `internal/world/cadence.go:22-24` is dead code behind that rejection — reconciled during
   the documentation pass, which is the resolution SI-03 §2 records). Effector risk_class low/medium/high,
   failure modes ok..silent_no_effect, interlock lt/lte/gt/gte all in schema; `expected_effector`
   is optional and loop scoring guards on it (`internal/score/loop.go:13`) and on `DeadlineNS`
   (loop.go:16). Entity count is free-form default/min/max (schema lines 57-64); churn is the
   shipped k8s pattern (`domains/k8s-cluster.domain.json:9,28`). Enum-value payload truth does
   NOT exist (random enum draws), so no candidate depends on text/enum content as evidence;
   counter reset/rollover is not modelled, so counter-based faults are pitched as
   integrator-freeze gaps plus perturbation-layer duplicates instead.

**Findings (external sources).**

- **Power grid**: NASPI's synchrophasor data-quality reports ([naspi.org](https://www.naspi.org))
  and IEEE bad-data work document 5–20% missing synchrophasor data in practice, with bad-data
  detection built on redundancy/estimation residuals — exactly `peer_residual` and
  `conservation_residual` under heavy perturbation; PMU (~20 ms) vs SCADA (2–4 s) cadence mixing
  on one entity is a documented stream challenge (sciopen.com review 2025; OSTI status-word
  paper 2022).
- **ICU/clinical alarms**: PhysioNet/CinC Challenge 2015, fetched directly
  ([physionet.org/content/challenge-2015/1.0.0](https://physionet.org/content/challenge-2015/1.0.0/)) —
  false-alarm rates up to 86%, only 2–9% of alarms clinically important, score
  TP+TN/(TP+TN+FP+5·FN): a 5× penalty for a missed alarm, the canonical asymmetric-cost
  scoring a simulator can replicate.
- **Air quality**: EPA long-term sensor performance evaluations and colocation-calibration
  literature (epa.gov; EGUsphere; MDPI 2020) document drift, short sensor lifespans,
  recalibration cycles, and network-wide calibration against a reference — a peer-residual
  problem where the reference itself can be the liar.
- **Market data**: NYSE Pillar client spec (nyse.com) and FINRA TDDS specs (finra.org) make
  sequence numbers and gap handling first-class feed-handler mechanics; MiFID II RTS 25 clock
  synchronization (100 µs granularity for HFT) makes timestamping a regulatory event; BIS
  (bis.org) documents surveillance over out-of-order tick streams.
- **Traffic**: FHWA traffic-detector QC procedures and PeMS-style 30-second detector diagnostics
  (constant-value/dropout/implausible-count fault flags, daily imputation) are the standard
  failure taxonomy — CAVEAT: these two reached the study only as general-knowledge fallback
  after repeated search rate-limiting, not as fetched pages; treat that taxonomy as well-known
  but unverified-here.
- **Watermark/late-data mechanics**: Apache Flink windows docs (nightlies.apache.org) and IBM
  event-processing docs (ibm.github.io) document idle-partition watermark stalls
  (earliest-of-partition-watermarks) and the too-tight/too-loose lateness-allowance tradeoff.

**Not done / not found.** No production postmortem combining late data + idle partitions + wrong
results was found in one source (search rate-limits forced broad queries); the first Flink
event-time URL tried 404'd. No test suite or build was executed; the only executed checks are
the two catalog commands, the schema enum extraction, and the greps cited above.

### 2.4 Angle D — Delivery-corruption frontiers

**Scope.** Partitions and reconvergence, duplication and reordering, dilution and throttling,
clock skew, schema drift, poison records, slow-consumer backpressure, and replay attacks on the
delivery path.

**Repo verification (run this session).**

1. `go run ./cmd/streamsim catalog coverage` — the same thin/zero lists (only
   `time_reference=wall(8)` confirms 8 shipped domains).
2. `go run ./cmd/streamsim catalog list` — the 8 shipped domains, with discrete-line-oee owning
   torrent=1 and interlocked=1, k8s-cluster churning=1, shipment-exception heavy_tail=1.
3. `internal/perturb/perturb.go:20-48` — exactly 19 perturbations, all run-level between world
   and adapter, with Active windows (FromNS/UntilNS) and per-ID substream RNG; Delivered records
   carry the same seq on duplicate (perturb.go:50-59) so an exactly-once oracle is scoreable.
4. `internal/perturb/transform_timing.go` — producer_flap buffers and flushes (67-76),
   gross_backfill withholds then flushes (33-46), clock_skew is a constant ± offset on
   ObservedTime (48-61), non_monotonic forces observed_time < event_time (100-112).
5. `internal/world/link_delay.go:9-21` — constant/exponential/lognormal/store_and_forward; the
   store_and_forward sampler is just a mean+Exp delay sample, so burst semantics for
   cold-chain-style flushes live in the perturbation layer, not the link sampler.
6. `docs/contracts/domain-spec-v0.1.schema.json` (via python): per-channel link_delay,
   availability, cadence, observation_bias/gain/offset, enum_values (line 242: out_of_enum
   emits values outside this set); axis enums confirmed (rate sparse/low/medium/high/torrent;
   cadence allows batch/human_driven as labels even though channel batch cadence is
   loader-rejected; time_reference wall/batch_relative/cycle_relative/calendar; fidelity F0–F3);
   ground_truth requires negative_class_fraction only at top level.
7. `internal/suite/scenario.go:58-69` and `internal/suite/command_log.go:21-23` — scenarios
   draw perturbations with params/windows and record them in the command log: delivery
   corruption is scenario DATA over the 19-name catalog, not domain code.

**Findings (external sources, by frontier).**

- **Out-of-order/lateness/completeness**: Akidau, Streaming 101/102 (URLs in §2.1) — event-time
  vs processing-time skew "often a highly variable function"; "you simply can't" know when all
  data for event time X has been seen; watermarks fail two ways (too fast → late data and wrong
  sums; too slow → delayed output); allowed-lateness GC drops records beyond a horizon.
- **Partitions and reconvergence**: Jepsen Kafka
  ([aphyr.com/posts/293-call-me-maybe-kafka](https://aphyr.com/posts/293-call-me-maybe-kafka)) —
  "987 acknowledged, 468 survivors, 520 acknowledged writes lost" under partition + leader
  election; recommendation "block and sound an alarm instead of silently dropping data"; a 2017
  comment shows mitigations (unclean.leader.election.enable=false, min.insync.replicas) reduce
  but do not eliminate it. The consumer-side frontier after a partition is reconvergence
  reconciliation: backfill bursts, re-acked duplicates, and deciding partitioned-but-alive vs
  dead.
- **Slow-consumer backpressure / dilution**: retention-window data loss is silent — messages
  deleted before consumers catch up when retention.ms < catch-up time (Kannika "Which scenarios
  lead to data loss in Kafka"; Confluent "Kafka Retention Explained"; operatex.dev lag-by-shape
  diagnosis); the community guidance is alerting on lag relative to the retention window
  (time-to-expire), because "nothing breaks; data just disappears". The AWS Kinesis throttling
  doc body did not render (fetch returned title only).
- **Poison records**: search-confirmed pattern (tanhdev, irisidea, daily.dev summaries) — a
  record that consistently fails on consume halts the whole partition in a redelivery loop;
  mitigations are retry limits, skip-and-log, and DLQ quarantine with manual replay. Spring
  Kafka error-handling-deserializer page 404'd; not retrieved.
- **Clock skew**: Google leap smear ([developers.google.com/time/smear](https://developers.google.com/time/smear)) —
  smeared and stepping clocks disagree by up to ~0.5 s on each side of a leap second, "breaking
  timestamp comparisons and ordering across systems". Cloudflare's leap-second postmortem body
  404'd at the guessed slug; not retrieved.
- **Schema drift**: registry-level compatibility ≠ end-to-end safety — "a platform can accept a
  produce request while a schema change still breaks a downstream application", and the most
  common failure is semantic breakage that passes compatibility checks (AutoMQ streaming data
  contracts; systemoverflow schema failure-mode survey). streams-simulator's analogue is
  unit_mismatch/out_of_enum/nan_inf/malformed/oversize over a bounded window.
- **Replay on the delivery path**: replay attacks capture telemetry and retransmit it later to
  forge device state; freshness requires timestamps/nonces/sequence ids (Alotaibi 2024, PMC;
  MQTT security guidance). In the bus world the same frontier is consumer rewind/resume replay
  (Last-Event-ID, auto.offset.reset) plus duplicated re-acks — exactly tamoz's
  "replayed/duplicated/reordered events to stress inbox dedup" need.

**Not run / not retrieved.** No repository tests were executed. Full bodies not retrieved due to
404/JS-render/rate-limit: Cloudflare leap-second postmortem, Datadog consumer-lag article,
Spring Kafka error-handling-deserializer, lucumr "Tale of Two Clocks", kafka.apache.org
documentation body; WebSearch was 429-throttled for most of the session.

**Idea selection logic (as recorded).** Every candidate is domain DATA on the shipped pattern
(shipment-exception proves the non-catalog route; k8s-cluster proves churn is live machinery),
exercises only existing F0/F1 dynamics, cadence, link_delay, churn and effector surfaces, pairs
a physical world story with the existing 19 run-level perturbations so the delivery frontier is
scenario data rather than code, names the sealed analytic truth (solver detector form), and
declares the axes it lifts off the measured thin/zero lists.

## 3. Project briefs, condensed

Three briefs framed what the ideas are for. The host brief is grounded in this repository; the
two consumer briefs were supplied by the study and could only be judged for internal
consistency here (see the caveat at the end of this section).

### 3.1 streams-simulator (host instrument)

A standalone, deterministic, closed-loop world simulator for testing stream processors: it
generates event streams from data-defined domains, corrupts delivery in specified ways, carries
sealed ground truth, accepts commands through declared effectors, and scores whatever consumed
it. Non-negotiables: determinism (a run is a pure function of `(sim_version, domain_digest,
adapter_digest, seed, command_log, sink)`), the analytic cross-check, the reference consumer,
the delivery ledger, the quiescence barrier, the sealed oracle, the injection probe, the
`silent_no_effect` test, the trivial-baseline audit. Domains, adapters, and effectors are data;
if any domain needs a code branch in the binary, the simulator is wrong and the domain found
the bug. The simulator has no knowledge of its consumers. Eight domains are shipped; the
coverage gaps the ideas answer to are the thin/zero axis values measured in §2.

### 3.2 agentic-stream (as supplied by the brief)

A durable stream consumer whose own audit and suite catalog name the cells it cannot yet test:
missing-heartbeat situations held uncertain from durable timers; the F-002 timer-stall class
during silent ingress; `trace_end` until-horizon as the post-silence time advance; the
window-mode matrix (on_close vs on_update vs early_and_close) with late corrections inside and
outside allowedLateness; corrected re-emission and reconsideration dedup; event_inbox
suppression and the `evt-[0-9]{6}` id cap as its known scale risk; virtual-partition
per-partition watermarks (its `internal/engine/engine_events.go:253-275` per the brief);
durable quarantine truthfulness (S1.9); restart/resume (S1.8); and admission-outcome
explainability under bounded reasoning budgets (invariant 10, the unenforced F-010
trigger-completeness gate).

### 3.3 tamoz (as supplied by the brief)

A governed decision agent whose release audit names what it cannot prove in constrained
environments: the kill-matrix exactly-once oracle generalized to device effects; the
three-valued reconcile journal (`:unknown` on applied-but-unacked ambiguity vs fenced retry);
MAX_RECONCILE_GRANTS discipline; watch-condition expiry and quiescence terminals; the P13 §9
schedule wishlist (delayed delivery, duplicate wakeups, long downtime, every misfire/overlap
combination, against a deterministic clock); byte-identical-payload dedup with
CheckpointConflictError; Last-Event-ID resume; approval-gated high-risk commands and approval
tempo; governed-decision byte-equality between clean and injected cells (WP-T4) with
digest-verified snapshot termination before any model call; and Q3 — what retires an entity and
what becomes of its timers.

**Verification caveat (recorded by the review stage; scope corrected during the folder's
documentation pass).** The brief ticket IDs above (F-002,
S1.8, S1.9, F-010, kill-matrix, watch-expiry) appeared **nowhere in the repository at the
review stage** (grep over
docs/ and .agents/ returned nothing then — this file, which quotes them, did not exist yet).
Today they live in exactly two kinds of place: this folder's own records (§3.2–3.3 here, the
idea documents' headers, [review-evidence.md](review-evidence.md) §3.3) and the two product
names in
`docs/design/GAP_ANALYSIS.md:46` and `docs/design/CRITICAL_REVIEW.md:145` — a re-run during
the documentation pass finds no other hits. Brief alignment
was judged as given by the study. The improvement loop re-anchored each idea's consumer claims
on what is verifiable in-repo — the committed agentic-stream adapter
(`docs/design/CONSUMERS.md:98-102`, `adapters/agentic-stream.adapter.json`), tamoz as
documented in `docs/design/CRITICAL_REVIEW.md:143-163`, and the neutral verdict vocabulary of
`internal/model/verdict.go` — and withdrew every unverifiable cell **where the idea document records the withdrawal** (SI-01, SI-04, SI-05, SI-10). Scope correction, documentation pass: SI-02's and SI-07's headers retained brief-supplied cell names ("kill-matrix exactly-once oracle", "three-valued reconcile journal", "quarantine truthfulness S1.9") — both headers now annotate those names as brief vocabulary rather than verified cells.

## 4. What was run and what was not

- **Run, repeatedly and consistently across all four angles**: `go run ./cmd/streamsim
  catalog coverage` and `catalog list` (after `go build`/`make build` where noted); schema
  enum extraction and greps over `internal/`, `domains/`, `docs/contracts/`; line-level reads
  of the runtime files cited above.
- **Never run**: any test or command inside agentic-stream or tamoz (claims come from the
  briefs, per §3's caveat); no `go test` at the research stage; the SOAK=1 soak; any suite run
  over candidate domain data (no candidate artifact existed).
- **Not retrieved** (rate-limits and 404s, per angle): listed in each angle's limitations;
  where a claim rests on a search snippet rather than a fetched page, it is marked second-hand
  in place.
