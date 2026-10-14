# SI-01 · `telemetry-estate`

- **ID / slug**: SI-01 / `telemetry-estate`
- **Date**: 2026-10-05
- **Status**: evaluated — quality bar **not met** (70/100; bar is ≥85/100 with no dimension below 7/10)
- **Consumer grounding**: agentic-stream (primary), tamoz (secondary) — verified in-repo consumers only

## 1. Purpose

A sparse remote-telemetry estate (wells, pumps, weather sites behind store-and-forward
gateways) where the wire's story of a gap is the evidence. Three worlds open the same way
— data stops or stalls — and resolve differently:

1. **asset powered off** (declared fault): the telemetry counter freezes mid-count and
   keeps arriving on schedule — *stale-on-time*;
2. **gateway buffering** (transport): fresh values arrive late — *late-and-fresh*;
3. **benign availability lull**: brief silence that self-heals.

A consumer must separate the three while holding the situation uncertain through the gap
and closing it only on rejoin, and dispatch repair inside a scored, mttr-scaled deadline
without wasting dispatches on lulls.

Grounded on verified in-repo consumers: **agentic-stream** (`trace-record-v0.1` framing,
`arrival_time` fed from `observed_time` — docs/design/CONSUMERS.md:98-102,
adapters/vendor/agentic-stream/trace-record-v0.1.schema.json committed) primarily;
**tamoz** (docs/design/CRITICAL_REVIEW.md:143-163) secondarily. The earlier
F-002/S1.8/fleet-telemetry-gateway grounding is withdrawn: zero references in the tree of this
document's writing pass (grep over md/json/go returned 0 hits for each; re-run for this
document, same result). Scope note, folder documentation pass: the pool name
`fleet-telemetry-gateway` now also appears in [../research-notes.md](../research-notes.md) §1,
created later the same day as this document; a re-run of
`grep -rE "fleet-telemetry-gateway" --include="*.go" --include="*.md" --include="*.json" .`
during that pass confirms zero hits outside `docs/simulation-ideas/` (see
[../README.md](../README.md) §5, write order).

## 2. How it runs on this simulator

Everything below is domain DATA against the shipped schema and runtime
(`internal/schemas/domain-spec-v0.1.schema.json`); no per-domain code branches. Where a
piece is *not* zero-code, §2.6 says so plainly.

### 2.1 Entities

40–80 sites (wells, pumps, weather stations). Some sites are **dual-reporting**: two
channels observe one upstream quantity, so the paired pair feeds `channel_divergence`
(schema :408, :423; `internal/truth/detector_math.go:16-22`).

### 2.2 Signals

- Per-site **periodic heartbeat channel carrying a monotonic counter** (the telemetry
  upload counter). The counter — not channel silence — is the labeled evidence surface.
- 2–3 **slow scalar process channels** (pressure/level) on `first_order_lag` dynamics
  (schema :136).

### 2.3 Dynamics — the declared fault (labeled world)

`site_down` is a **state-delta fault that zeroes the counter's driving state**, freezing
it mid-count while the emission schedule continues — *stale-on-time*, the shipped
`host_down` pattern (domains/host-system-health.domain.json:233-241: `host_powered`
delta, detector `single_channel_snr` on `uptime_s`). This re-anchor is forced by the
engine: a declared fault cannot silence emission. Faults only apply state deltas
(`internal/world/faults.go:27-38`, registration couples `fault.Affects` to states at
:90-92); emissions gate only on availability renewal and entity liveness
(`internal/world/observe.go:24-33`, `:82-94`), none of which is fault-scoped. So the
fault rides value physics.

The **graded lane declares the freeze intermittent** (onset shape `intermittent` +
`duty_cycle`, schema :376-379; precedent `memory_thrash`,
domains/k8s-cluster.domain.json:123-131) so delivered-freeze evidence must survive
delivery noise. A **step-freeze variant stays a mechanism fixture**: a clean freeze is
expected to trivial-out under the audit's hindsight bar (§3, G9).

Detector: `single_channel_snr` on the frozen counter, computed by the shipped
clean-vs-faulted scan (`internal/truth/solver_scan.go:40-61`). Expected effector
`dispatch_repair_tech` with `deadline_s` scaled to lull mttr and a declared
counterfactual (schema :437).

### 2.4 Perturbations — the transport world and the benign lull

- **Transport world** (delivery layer, not a fault): `link_delay` `store_and_forward`
  (`internal/world/link_delay.go:17-18`, `mean + Exp(mean)`, schema :329 — a pure
  observed-time rewrite) plus `delay_tail` and `producer_flap` shaping the resume burst;
  `duplicate_burst` + `id_reuse` litter the reconnect dump
  (`internal/perturb/perturb.go:21-40`). Fresh `event_time`, lagging `observed_time`.
- **Benign lull** (negative class): ambient availability windows on the heartbeat
  channel — channel-level availability `uptime < 1` with short `mttr_s`
  (schema :313-321; runtime `internal/world/availability.go:88-109`, seeded per
  entity/channel substream at availability.go:14) — present in **every** world including
  control; plus one declared `is_negative_class` benign fault in the `transient_none`
  pattern (domains/host-system-health.domain.json:243-249): a routine operational pause
  that flattens process channels and recovers.

**Relation to review finding R2, stated so it binds** (added during the folder's
documentation pass; sibling ideas treat this list as load-bearing): `store_and_forward` is on
the architecture review's G2 gap list — the link-delay models `lognormal`/`store_and_forward`
"have no analytic oracle"
([docs/reviews/ARCHITECTURE_REVIEW.md:127-146](../../reviews/ARCHITECTURE_REVIEW.md), finding
R2, whose recommendation is "add an oracle or record an explicit scope decision"). R2 is a
truth-coverage gap, not a usage ban, and this domain stays inside it because **nothing in its
oracle consumes the delay draw**: the sealed label is SNR on the counter's value physics,
`first_observable` is defined against the seeded emission schedule (§2.5), and the ledger's
`delayed` rows are transport evidence, never oracle input — the delay's missing closed form
cannot corrupt the label. SI-05 makes the opposite, stated choice (its oracle needs exact
flush content, so it uses `gross_backfill` instead, SI-05 §2); the portfolio carries both
positions explicitly rather than by omission.

### 2.5 Ground truth and its analytic oracle

- **Sealed label**: from the shipped scan solver — `single_channel_snr` on the frozen
  counter (`internal/truth/solver_scan.go:40-61`), sealed via `truth.BuildRecord` in
  suite generation (`internal/suite/scenario.go:27-38`). Truth is director-only
  (`internal/truth/solver.go:1-5`).
- **G2 independent oracle (new truth-side test code)**: a tick-counting closed form —
  *first seeded period whose counter value deviates; reconnect backlog = exact count of
  buffered periods* — a second integration of the seeded periodic schedule, asserted
  equal to the scan label **before unblind**, with explicit boundary tests for reconnect
  mid-period, resume burst overlapping the solver's 60 s default scan sample
  (`internal/truth/solver.go:33-38`), and benign-lull negatives.
- **`first_observable_time` is defined against the seeded schedule** (world-side truth),
  never against arrivals: `delay_tail` is uncapped Exp+Exp and arrivals can exceed world
  end, handled deterministically by `traceEndTime` (`internal/run/delivery.go:68-77`).
- **The delivery ledger separates transport from death**: buffered rows carry
  `delayed` / `dropped_by_perturbation` reasons (`internal/model/event.go:24-33`); an
  availability-dead world emits nothing at all, so leaves no rows
  (`internal/world/observe.go:82-94`); scoring consumes the ledger fail-closed, so a
  buffered-gateway miss is a transport miss, never a reasoning miss
  (`internal/perturb/perturb.go:1-7`). Ledger rows are **director-side evidence, not
  consumer input and not the label** (`internal/model/event.go:36-37`).

### 2.6 Effectors

`dispatch_repair_tech` carries:

- a declared **`silent_no_effect` probability** — feeding the false-success metric
  (`internal/score/loop_metrics.go:35-44`; schema :485-486);
- **`confirmed_no_effect`** for the two-signal mechanism cell (schema :486), with
  `confirmation_channels` declared as the schema requires when any no-effect mode is
  present (schema :513-516, :557).

The **benign twin of the trap is NOT `confirmed_no_effect` semantics**: dispatching on a
merely-buffering site is a *correct no-op*, scored as `UnnecessaryAction`
(`internal/score/loop_metrics.go:19`), which vetoes `DeadlineAdhered`
(`internal/score/loop.go:16-18`), with the site's healthy confirmation channels carrying
the truth — no stretched semantics.

## 3. Feasibility against G1–G9

Gate definitions: docs/QUALITY_BAR.md:64-72. One schema expresses every data surface;
key lines re-verified this session: absence enum incl. `ambiguous` (:30, channel-level
:313-314), availability `mttr_s` (:321), `store_and_forward` (:329), exactly four
reading-based detector forms (:408-409), `deadline_s` (:437), scored no-effect modes
(:485-486), `confirmation_channels` required (:516, :557), `first_order_lag` (:136),
`intermittent`/`duty_cycle` (:376-379). Live runtime verified: availability renewal
(`internal/world/availability.go:88-109`), link delay
(`internal/world/link_delay.go:9-21`), perturbation catalog
(`internal/perturb/perturb.go:21-48`), effect application, deadline and
unnecessary-action scoring (`internal/score/loop.go:16-18`).

- **G1 determinism** — seeded draws throughout: availability RNG per entity/channel
  substream (availability.go:14), perturbation RNG per activation
  (`internal/perturb/perturb.go:173`).
- **G2 analytic cross-check** — the label rides value physics via the shipped scan
  solver; the tick-counting closed form is the G2 independent oracle, asserted equal to
  the scan label before unblind, with the boundary cases listed in §2.5. Not zero-code:
  no absence/heartbeat detector form exists — the shipped solver computes SNR crossings
  on clean-vs-faulted *readings* only (`internal/truth/solver.go:65-76`,
  `:94-106`; `internal/truth/detector_math.go:10-37`), and adding an absence form would
  be a schema design change (form-without-equation violates the detector contract,
  schema :409), deliberately not proposed here.
- **G3 reference consumer** — demonstrable with zero new code: the shipped reference
  consumer already detects channel silence (`internal/refconsumer/series.go:95-105`,
  `absence_factor × median period`) and invokes the declared detection effector through
  the operator surface (`internal/refconsumer/detection.go:34-58`).
- **G4 delivery ledger** — the transport-vs-death discriminator (§2.5), with immutable
  terminal reasons (`internal/model/event.go:24-33`).
- **G5 quiescence** — via the post-reconnect flush plus a **ship precondition**: a
  run-level conservation test with `producer_flap` active at End. `run.End` never
  flushes windowing buffers (`internal/run/finalize.go:30-78` contains no
  `perturb.Flush`; `internal/perturb/flush.go:83-93` releases only when an advance
  boundary crosses `UntilNS`), so a flap whose window ends at/past the last advance
  silently drops buffered emissions from the ledger and breaks the tick-count
  conservation this domain's oracle leans on. Fix as terminal delayed rows reusing
  `recoveredRecord`'s `DeliveryDelayed` marking (`internal/perturb/flush.go:95-105`) or
  fail the run. No run-level test exercises `producer_flap` or `gross_backfill` at End
  (grep over `internal/run/*_test.go`: 0 hits; re-run for this document, same result).
- **G6 sealed oracle** — inherited platform structure: truth is director-only
  (`internal/truth/solver.go:1-5`), labels sealed before suite admission
  (`internal/suite/scenario.go:27-38`), tick-vs-scan reconciliation happens before
  unblind.
- **G7 injection probe** — no producer text: the perturbation layer changes nothing
  physical and consumer knowledge never enters it (`internal/perturb/perturb.go:1-7`,
  `:86-88`); the event envelope never carries perturbation metadata, fault identity or
  labels (`internal/model/event.go:3-7`).
- **G8 `silent_no_effect`** — declared no-effect probability feeding the false-success
  metric (§2.6), plus `UnnecessaryAction` scoring of the benign twin.
- **G9 trivial baseline** — **admission gate before any wire-indistinguishability
  claim**: a scratch `streamsim suite --domain --profile` run in which the audit's
  `channel_silence` hindsight detector over both worlds
  (`internal/audit/detector.go:5-15`; panel build `internal/audit/audit.go:69-80`;
  cutoff 0.9 at `internal/audit/audit.go:18`) scores < 0.9 balanced accuracy on
  fault-vs-lull, and the intermittent-freeze scenarios enter non-trivial (trivial
  exclusion at `internal/suite/scenario.go:20-22`). Proxy evidence (run this session):
  `./bin/streamsim suite --domain host-system-health --profile sensor_pathology -n 40`
  kept 40 of 87 candidates and trivial-excluded 47, hindsight detectors separating most
  drawn positives — so a clean step-freeze `site_down` should be expected to trivial-out
  and is planned as the mechanism fixture, not the graded flagship. **The domain's own
  gate run has not happened and gates the claim.**

**Hardest part, restated honestly**: availability/deadband parameterization so lulls and
fault gaps are genuinely hard — decided by the admission gate, not asserted. The three
worlds are not value-identical (stale-on-time, late-fresh, briefly-silent differ); the
hardness is that each opens ambiguously and resolves only over time or at rejoin.

## 4. Quality bar result

| Dimension | Score (0–10) |
| --- | --- |
| consumerValue | 7 |
| dataExpressibility | 8 |
| determinismOracle | 6 |
| ledgerQuiescence | 6 |
| novelty | 6 |
| sealedTruth | 7 |
| specificity | 9 |
| **Total** | **49/70 → 70/100** |

**Bar: not met.** The bar is a total of at least 85/100 across the seven dimensions with
no single dimension below 7/10. This idea totals 70/100, and three dimensions —
determinismOracle, ledgerQuiescence, novelty — sit at 6/10, below the per-dimension
floor.

## 5. Five-lens review

### 5.1 Determinism and analytic oracle (score 7)

**Found**: the headline oracle was misanchored — the spec claimed "closed-form tick
counting … trivially double-implemented", but the shipped solver labels by scanning
clean-vs-faulted noiseless worlds for SNR-threshold crossings on one of four
reading-based detector forms (`internal/truth/solver.go:65-76`,
`internal/truth/solver_scan.go:40-61`; forms enum schema :408). No heartbeat/absence
form exists, `World.Reading` never observes absence (`internal/world/observe.go:10-21`),
and a `site_down` fault was labelable only as a state-freeze deviation — which times the
freeze, not the first missed heartbeat. The tick-counting oracle and its reconciliation
with the scan label at the rejoin boundary is new, unspecified oracle work. The spec also
never stated which sealed-record field carries "the gap was transport, not death" — the
ledger discriminates (G4) yet is director-side evidence, not the label
(`internal/model/event.go`).

**Changed**: the label is declared on value physics (`single_channel_snr` on the frozen
counter); the tick-counting closed form is spec'd as the G2 independent oracle with
explicit boundary tests (reconnect mid-period; resume burst overlapping the 60 s scan
sample, solver.go:33-38; benign-lull negatives) asserted equal to the scan label before
unblind; per-scenario transport-vs-death recording is stated (ledger reasons
`delayed`/`dropped_by_perturbation` vs no emission row at all) with fail-closed scoring.

### 5.2 Data-not-code expressibility (score 5)

**Found**: the load-bearing mechanism was not expressible — world faults only apply
state deltas (`internal/world/faults.go:27-38`, `:81-94`); nothing couples a fault to
emission, so a labeled "asset powered off" fault cannot silence the heartbeat channel;
the claimed first-missed-heartbeat oracle needs a detector form that does not exist
(`internal/truth/solver.go:94-106`; schema :408). The expressible residue was already
shipped: k8s-cluster's `network_partition` (domains/k8s-cluster.domain.json:133-141)
carries the value-freeze + heartbeat-continues pattern, undercutting the "three
wire-identical worlds" novelty if the fault re-anchors on values.

**Changed**: the labeled fault is re-anchored on value physics — the counter freeze (the
`host_down` pattern) — and heartbeat absence arises from ambient availability plus
`store_and_forward` buffering, so all three worlds are wire-real without new machinery;
the reduced novelty is accepted (`absence=ambiguous`, heartbeat claim withdrawn). A
silence/heartbeat detector form is deliberately **not** proposed (form-without-equation
violates the detector contract, schema :409) and the spec says so.

### 5.3 Consumer testing value (score 6)

**Found**: the headline fault was not expressible as sealed silence — a silence-only
`site_down` has zero deviation, `FirstObservableNS == 0`, and is auto-classified trivial
and excluded from the graded suite (`internal/suite/scenario.go:20-22`). The "closed-form
tick counting" oracle is not what the truth machinery computes. G8 was mechanism-only:
`confirmed_no_effect` is the two-signal mechanism cell (schema :486), while the
false-success metric is graded only for `silent_no_effect`
(`internal/score/loop_metrics.go:35-44`) — none was declared. The "three wire-identical
worlds" hinge on deadband tuning and was asserted, not evidenced: the audit's
`channel_silence` hindsight detector fits over both fault and control worlds
(`internal/audit/detector.go:5-29`, `internal/audit/audit.go:69-80`) and would separate
shorter/less-frequent lulls at ≥0.9, trivialling everything. Positive counterweight:
the shipped reference consumer already detects channel silence
(`internal/refconsumer/series.go:95-105`) and invokes the detection effector via the
documented surface (`internal/refconsumer/detection.go:34-58`), so the closed loop (G3)
is demonstrable with zero new code.

**Changed**: one truth story picked explicitly — `site_down` as a state-delta freeze on
the counter, and the three-worlds premise re-derived honestly (stale-on-time vs
late-and-fresh vs lull) rather than "wire-identical"; a declared `silent_no_effect` mode
added on `dispatch_repair_tech` so the domain feeds the false-success metric, not just
the mechanism cell; the negative class must be proven empirically — a scratch
`streamsim suite` run showing `channel_silence` balanced accuracy < 0.9 on
fault-vs-benign-lull — before the ambiguity claim is made; the clean step-freeze variant
is demoted to a mechanism fixture.

### 5.4 Delivery ledger and quiescence (score 7)

**Found**: the ledger-as-discriminator claim is verified and real — availability-down
emits nothing (`internal/world/observe.go:82-94`) so a dead world leaves no rows, while
`drop` leaves terminal `dropped_by_perturbation` rows; `store_and_forward` is a pure
observed-time rewrite (`internal/world/link_delay.go:17-19`) so buffering cannot stall
delivery or quiescence. But `producer_flap` resume bursts use a held buffer that reaches
terminal ledger state only when an advance boundary crosses `UntilNS`
(`internal/perturb/flush.go:83-93`), and `run.End` never flushes buffers
(`internal/run/finalize.go:30-78`) — a flap ending at/past the last advance silently
loses emissions from the ledger, breaking the conservation the tick-counting oracle
leans on; no run-level test exercises `producer_flap`/`gross_backfill` at End.
`delay_tail` is uncapped Exp+Exp; arrivals can exceed world end, handled by
`traceEndTime` (`internal/run/delivery.go:68-77`).

**Changed**: a run-level conservation test with `producer_flap` active at End is a ship
precondition (flush remaining buffers as terminal rows in End — reusing the committed
`DeliveryDelayed` marking at `internal/perturb/flush.go:95-105` as conformance evidence
for reconnect-backlog rows — or fail the run); the heartbeat oracle is stated against
the seeded periodic schedule with ledger reasons used only as the transport
discriminator; arrival-time-based `first_observable` is forbidden in the spec.

### 5.5 Novelty and portfolio balance (score 7)

**Found**: novelty was measured only against the 8 shipped domains. `absence=heartbeat`
is 1 in the shipped set but carried by ~8 planned catalog domains
(docs/design/DOMAIN_CATALOG.md vectors/prose at :87, :205-226, :291, :367, :419, :440,
:508; grep re-run for this document), so "fills absence=heartbeat" is not novel. The
candidate never argued against its nearest neighbors: 03 `livestock-herd`
(DOMAIN_CATALOG.md:121-137 — ambiguous absence resolved by a second producer, the
parlour weight, stopping at advisory flagging) and 18 `ev-fleet` (:455-472 —
connectivity-vs-death separated by learned geography, stopping at service scheduling).
The domain's point is three-way undecidability from the wire — the `ambiguous` absence
semantics (DOMAIN_CATALOG.md:31) — yet it declared `heartbeat`. SI-01 and SI-05 share
heartbeat, a reconnect window, `store_and_forward`, and the ledger-as-discriminator
trick (five of ten ideas reuse it). The kill/absorb claim ("absorbs killed
fleet-telemetry-gateway") and the cited consumer cells (F-002, S1.8) are unverifiable —
zero repo references (grep re-run for this document: 0 hits each).

**Changed**: declares `absence=ambiguous` — the sharper shipped gap (coverage:
`ambiguous=2`; `./bin/streamsim catalog coverage`, re-run this session) — because
three-way undecidability from one channel is the point; the `absence=heartbeat` claim is
withdrawn and `lateness=heavy_tail` is raised instead (currently 1, same run). The
nearest-neighbor delta is stated: this domain disambiguates **inside one producer's
channel set** — counter staleness plus event-vs-observed time divergence on paired
channels — and grades the dispatch decision under the ambiguity (deadline adherence,
`UnnecessaryAction`, `silent_no_effect` false-success), which neither neighbor carries.
The SI-01/SI-05 boundary is explicit — SI-01 grades live three-way separation of an
ongoing gap under a scored dispatch deadline; post-hoc replay masking stays SI-05's lane
— so the absence-axis credit is not double-counted; if both ship, the axis earns its
third entry honestly, otherwise fold one. The kill/absorb claim is withdrawn; consumer
grounding rests on the committed agentic-stream adapter
(docs/design/CONSUMERS.md:98-102) and tamoz
(docs/design/CRITICAL_REVIEW.md:143-163). `dispatch_repair_tech` gained a scored,
mttr-scaled `deadline_s` with a declared counterfactual (schema :437) so the advisory
loop has a measurable false-positive cost.

## 6. Improvement history

One improvement round separated the original consolidated spec from the final spec.
Substance of what changed:

1. **Fault mechanism re-anchored on value physics.** Original: a `site_down`
   *availability-freeze* fault plus `absence=heartbeat` — silence itself as the labeled
   evidence. Final: a state-delta fault zeroing the counter's driving state (stale-on-time,
   `host_down` pattern), with an intermittent-freeze graded lane (`duty_cycle`) and the
   step-freeze variant demoted to a mechanism fixture. Cause: the engine's faults are
   state deltas and cannot silence emission (`internal/world/faults.go:27-38`;
   `internal/world/observe.go:24-33`).
2. **Oracle re-anchored and demoted to honest scope.** Original: "oracle is closed-form
   tick counting … trivially double-implemented", `first_observable_time` = first missed
   heartbeat. Final: the sealed label comes from the shipped scan solver
   (`single_channel_snr` on the frozen counter); tick counting is new truth-side test
   code serving as the G2 independent oracle, asserted equal to the scan label before
   unblind, with named boundary tests; `first_observable_time` is defined against the
   seeded schedule, never arrivals.
3. **Absence semantics and novelty corrected.** Original: `absence=heartbeat` ("thin at
   1") plus "absorbs killed fleet-telemetry-gateway". Final: `absence=ambiguous` (the
   three-way undecidability is the point), heartbeat claim withdrawn (~8 planned catalog
   domains carry it), kill/absorb claim withdrawn (zero repo references), nearest
   neighbors 03/18 argued, explicit SI-05 boundary, `lateness=heavy_tail` raised.
4. **Consumer grounding re-verified.** Original: F-002 timer stall, restart/resume
   S1.8, tamoz watch-expiry cells. Final: only verifiable grounding kept — the committed
   agentic-stream adapter (`arrival_time` from `observed_time`,
   docs/design/CONSUMERS.md:98-102) and tamoz as documented; the unverifiable cells are
   explicitly not claimed.
5. **G8 coverage widened.** Original: advisory `dispatch_repair_tech` with a
   `confirmed_no_effect` cell as "the benign twin of the silent_no_effect trap". Final:
   declared `silent_no_effect` probability feeding the false-success metric, plus
   `confirmed_no_effect` as the two-signal mechanism cell; the benign twin re-scoped as
   a correct no-op scored `UnnecessaryAction` vetoing `DeadlineAdhered` — no stretched
   semantics — with an mttr-scaled `deadline_s` and declared counterfactual.
6. **Quiescence made a ship precondition.** Original: "G5 exercised by post-reconnect
   flush". Final: a run-level conservation test with `producer_flap` active at End
   (terminal delayed rows reusing `DeliveryDelayed`, or fail the run), because `run.End`
   never flushes windowing buffers.
7. **Hardness claim gated instead of asserted.** Original: "three wire-identical
   worlds … data tuning, not code". Final: the worlds "open the same way and resolve
   differently"; any wire-indistinguishability claim is gated on a scratch suite run with
   the audit's `channel_silence` hindsight detector (< 0.9 balanced accuracy on
   fault-vs-lull, intermittent-freeze scenarios non-trivial); the three not-zero-code
   items (tick oracle, End-flush fix, admission gate) are stated plainly.

## 7. Open questions

1. **Admission gate outcome.** Does the domain's own suite run hold `channel_silence`
   below 0.9 balanced accuracy on fault-vs-lull and admit the intermittent-freeze
   scenarios as non-trivial? Untested; this gates the ambiguity claim. (Proxy:
   host-system-health `sensor_pathology` at n=40 kept 40/87 — run this session.)
2. **Parameterization hardness.** Can availability/deadband parameters make lulls and
   fault gaps genuinely hard on the wire? Decided by the gate run, not asserted.
3. **End-flush fix shape.** Terminal delayed rows reusing the `DeliveryDelayed` marking
   vs failing the run at End — decide, land with the run-level conservation test, and
   cover `gross_backfill` too.
4. **Tick-vs-scan agreement at the boundaries.** Reconnect mid-period and resume bursts
   overlapping the solver's 60 s default sample: the pre-unblind assertion is the test,
   but no seeded draw has exercised it yet.
5. **Intermittent-freeze survival.** Does delivered-freeze evidence survive
   `duplicate_burst`/`id_reuse` litter and `delay_tail` resume bursts at suite scale,
   or does the graded lane collapse toward the step fixture?
6. **Portfolio fold decision.** If SI-05 also ships, the `absence=ambiguous` axis earns
   its third entry; otherwise fold one of the pair.
7. **Long-run detector form.** Does this domain eventually need an
   emission-aware/absence detector form (a schema design change under the detector
   contract, schema :409), or does value-physics labeling suffice?
