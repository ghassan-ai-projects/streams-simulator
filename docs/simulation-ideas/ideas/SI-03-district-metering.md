# SI-03 · District metering — conservation leak under batched, late, reordered, duplicated delivery

- **Id / slug**: SI-03 / `district-metering`
- **Status**: documented; **quality bar not met** (69/100 against an 85/100 bar; three dimensions below 7/10)
- **Date**: 2026-10-05 · **Branch**: `new-sim-ideas` (working tree clean at `867026a`)
- **Evidence base**: every repository citation below was re-read at the given path during this
  documentation pass, and the checks named in the text were run here: `go run ./cmd/streamsim
  catalog coverage` (the two axis blocks this document leans on are quoted directly below — an
  earlier revision of this header promised the output "quoted under Purpose and Improvement
  history", where it never appeared; fixed in the folder's documentation pass by quoting it
  here), and greps over `internal/` and `cmd/`
  (non-test) for `BatchSize` (exactly one hit: `internal/model/domain.go:182`),
  `observe_only` (zero hits), and the withdrawn kill-target names `ev-charging-plaza` /
  `grain-storage-backfill` (zero hits in `*.go`/`*.md`/`*.json` at this document's writing
  pass; the names now also appear in [../research-notes.md](../research-notes.md) §1, created
  later the same day — a re-run during the folder's documentation pass confirms zero hits
  outside `docs/simulation-ideas/`). Gate definitions G1–G9 are
  read from [docs/QUALITY_BAR.md:64-72](../../QUALITY_BAR.md).

  `go run ./cmd/streamsim catalog coverage`, re-run during the folder's documentation pass
  (2026-10-05) — the two blocks this document's zero-coverage claims rest on:

  ```json
  "cadence":   { "event_driven": 7, "periodic": 7, "report_by_exception": 1 },
  "actuation": { "advisory": 4, "fast_physical": 2, "interlocked": 1, "slow_physical": 1 }
  ```

  No `batch` cadence value and no `observe_only` actuation value appears — both are absent
  from `by_axis` entirely, which is the claim the Purpose and Improvement-history sections
  make.

## Purpose

From the final spec (`spec.purpose`, quoted verbatim):

> Catalog 15 water-distribution's district-metered-area conservation leak
> ([docs/design/DOMAIN_CATALOG.md:385-406](../../design/DOMAIN_CATALOG.md)), re-angled at the
> delivery axes 15 does not carry: batched, heavily-late, reordered, duplicated arrival. One
> district entity carries the balance — master inlet register, outlet register, and the
> district's own aggregate metered-consumption channel — while 24 submeter entities emit
> consumption registers the consumer must sum; a background leak grows master−Σsubmeters
> linearly. The detector is single-entity and expressible today; the delivery story is the new
> content. Exposes the window-policy matrix — on-time fold, corrected inside allowance,
> beyond-lateness fold, duplicate batch replay — as four scored cells against a leak whose
> magnitude is known exactly; consumers fail by scoring on arrival time or silently
> mis-folding late batches. Run by agentic-stream teams tuning watermark and
> window-attribution policies.

The fault family is catalog 15's `background_leak_growth`
([docs/design/DOMAIN_CATALOG.md:385-406](../../design/DOMAIN_CATALOG.md)); the new content is
the delivery shape (batch cadence, heavy-tail lateness, `observe_only` actuation) and the
window-policy matrix scored against delivery-invariant sealed times.

## How it runs on this simulator

The domain is **data** — one JSON file through `internal/schemas/domain-spec-v0.1.schema.json`,
no per-domain code (AGENTS.md: "domains are data"). The sketch below is the domain as
authored, organised by spec section. Every mechanism claim is pinned to code that was read.

### Entities

- **1 district entity** carrying the balance channels (master inlet register, outlet register,
  district aggregate metered consumption).
- **24 submeter entities**, each emitting a consumption register.

The single-entity detector is a load-bearing modelling decision, not a convenience: the
solver's conservation balance sums detector channels of **one** entity
([internal/truth/solver.go:117-123](../../../internal/truth/solver.go) —
`sum += w.Reading(entityID, c, t)`), and `channel_divergence` likewise reads two channels of
the same entity ([internal/truth/detector_math.go:16-22](../../../internal/truth/detector_math.go)).
"master − Σ submeters" across 25 entities is **not** computable by any declared detector form;
re-modelling 25 meters as 25 channels of one entity would break per-entity fault injection
([internal/world/faults.go:29-38](../../../internal/world/faults.go)) and ledger entity
semantics. Hence: the district entity carries the balance, the submeters stay entities for
fault/ledger structure, and the district's `metered_consumption` channel is **authored data**
equal to the sum of the 24 submeter states, asserted by the analytic cross-check (see below).
There is no cross-entity state coupling in the model — `State`/`Dynamics` are per-entity.

### Signals (channels)

| Channel | Entity | Kind | Cadence |
|---|---|---|---|
| `master_register` | district | counter | `batch`, with `batch_size` |
| `outlet_register` | district | counter | `batch`, with `batch_size` |
| `metered_consumption` | district (owned aggregate) | counter | `batch`, with `batch_size` |
| consumption register × 24 | each submeter | counter | `batch`, with `batch_size` |

Channels declare `cadence=batch` with `batch_size` and heavy-tail link delay. Both are
schema-legal (`batch` in the cadence enum,
[internal/schemas/domain-spec-v0.1.schema.json:28](../../../internal/schemas/domain-spec-v0.1.schema.json);
`batch_size` at [:281](../../../internal/schemas/domain-spec-v0.1.schema.json)). Schema-legal
is not shippable today, stated plainly: the loader **rejects** any channel declaring
`cadence=batch` — "channel %q uses batch cadence, which is not implemented"
([internal/domain/validation_records.go:32-35](../../../internal/domain/validation_records.go),
re-read during the folder's documentation pass — resolution of the research stage's
"loader-rejected" vs this document's earlier "routes to periodicEmission": both fragments are
true; the `batch` case at
[internal/world/cadence.go:22-24](../../../internal/world/cadence.go) is unreachable dead
code behind the rejection). The consumer-visible
residual master−Σ(submeter registers) is formable from the wire, so the solver's
`first_observable` — computed on the same noiseless channels — is against a signal the
consumer can actually form.

### Dynamics

- **F0**: diurnal/weekly demand flows (the clean world).
- **F1**: one leak integrator on the district balance — the fault family is catalog 15's
  `background_leak_growth`
  ([docs/design/DOMAIN_CATALOG.md:385-406](../../design/DOMAIN_CATALOG.md)).

### Perturbations

`delay_tail` + `gross_backfill` + `duplicate_burst` + `reorder` + `drop`, over scenario
`FromNS`/`UntilNS` windows bounded inside the duration. All kinds exist in the perturbation
catalog ([internal/perturb/perturb.go:21-39](../../../internal/perturb/perturb.go)); the window
bounding is `drawPerturbations` in
[internal/suite/scenario.go:67-77](../../../internal/suite/scenario.go) (`from` starts in the
first half of the duration, `until = from + (0.25..0.75)·duration/2`, so windows never escape
the horizon).

### Ground truth and its analytic oracle

- **Detector**: `conservation_residual` on the district entity alone — inputs
  `[master_register]`, outputs `[outlet_register, metered_consumption]`.
- **What cancels**: `conservationDeviation` computes `(fi−fo)−(ci−co)` between the clean and
  faulted worlds ([internal/truth/detector_math.go:31-37](../../../internal/truth/detector_math.go)),
  so all clean-world terms (seasonal, demand) cancel. **The only closed form is the leak
  deviation `leak·(t−t_onset)`** against the effective pooled sigma `√Σσ²` over the three
  channels ([internal/truth/detector_math.go:65-71](../../../internal/truth/detector_math.go),
  `balanceSigma`). There is **no** closed form for the exact integral of seeded flows in
  general. The solver itself derives no per-form closed forms — it reuses the verified
  integrator and reports `Method=SolveNumeric`
  ([internal/truth/solver.go:18-21](../../../internal/truth/solver.go) doc comment;
  `Method: model.SolveNumeric` at [:70](../../../internal/truth/solver.go)).
- **The independent check is an acceptance obligation of this idea, not an existing
  capability**: new test code in the `TestAnalyticCrossCheck` mold
  ([internal/world/world_test.go:93-139](../../../internal/world/world_test.go)) asserting the
  leak integrator against its closed form, with the integrator-oracle precedent
  (`TestIntegratorAnalyticOracle`, dx/dt = g·u vs x(t) = x₀+g·u·t,
  [internal/world/oracle_test.go:58](../../../internal/world/oracle_test.go)).
- **Design spine — delivery-invariant sealed times**: `first_observable` is computed on
  noiseless truth-state balances and is therefore invariant to delivery. This is what lets the
  ledger separate a **withheld batch (transport miss)** from a **mishandled one (reasoning
  miss)**.
- **Boundary oracle**: a leak onset landing inside a withheld batch is labelled from world
  state, never from the batch boundary.
- **G2 negative-recovery case**: leak onset and `first_observable` both inside the
  late-arrival window; closed-form integral and RK4 asserted equal across the
  fold/duplicate/replay cells.
- **Window-policy rubric** (consumer-agnostic through the neutral verdict vocabulary,
  [internal/model/verdict.go:37-45](../../../internal/model/verdict.go) — the seven outcomes
  `accepted`/`duplicate`/`conflict`/`late`/`out_of_contract`/`malformed`/`rejected`;
  "billing" appears only in domain prose, never in the binary).

  **The scorer for cells A–D, named** (added during the folder's documentation pass; an
  earlier revision left "Detection.DetectedAt" without saying what reads it): detection
  latency and label correctness are graded by the offline judgment instrument —
  [internal/score/judgment.go:12-39](../../../internal/score/judgment.go) reads the consumer
  verdict's `detections` array (each `Detection` carries `EntityID`, `DetectedAt`, `Label`,
  [internal/model/verdict.go:27-34](../../../internal/model/verdict.go)), keeps the best
  detection on the labeled entity, computes
  `DetectionLatencyNS = DetectedAt − FirstObservableTimeNS`, flags any detection earlier than
  `first_observable` as suspicious, and grades `LabelCorrect` against the sealed label. This
  path grades the verdict artifact and needs no effector, which is how an `observe_only`
  domain scores a detection question; a reference-consumer run exercising it remains
  unspecified (see G3 below).

| Cell | Policy | Scored as |
|---|---|---|
| A | On-time fold | `Detection.DetectedAt` vs sealed `first_observable` |
| B | Corrected inside allowance | delayed rows declared late ([internal/score/admission.go:55-62](../../../internal/score/admission.go)), detection within tolerance of `first_observable` — deliberately a cell an arrival-anchored consumer can still pass |
| C | Beyond-lateness fold | sealed label asserts the threshold crossing occurred inside the withheld window; scored as detection latency vs the sealed times |
| D | Duplicate batch replay | every duplicated ledger row declared duplicate ([internal/score/admission.go:43-53](../../../internal/score/admission.go)) and detection latency unchanged vs cell A (replay must not move the estimate) |

### Effectors

**None.** `actuation=observe_only`, realised as an empty effector set — schema-legal: the
top-level `required` list omits `effectors`
([internal/schemas/domain-spec-v0.1.schema.json:8](../../../internal/schemas/domain-spec-v0.1.schema.json);
`observe_only` in the actuation enum at
[:34](../../../internal/schemas/domain-spec-v0.1.schema.json)). This isolates event-time
semantics from any effector confound.

## Feasibility against G1–G9

Gate definitions: [docs/QUALITY_BAR.md:64-72](../../QUALITY_BAR.md). What follows is
`spec.feasibility`, organised per gate; gates the spec does not speak to are marked as such
rather than filled in.

**The load-bearing decision, taken before domain authoring** (per AGENTS.md's
design-change-first rule): **batch emission does not exist.**
[internal/world/cadence.go:22-24](../../../internal/world/cadence.go) routes mode `batch` to
`periodicEmission`, and `batch_size` is parsed-only
([internal/model/domain.go:182](../../../internal/model/domain.go); the grep over `internal/`
and `cmd/` non-test returns exactly that one line — run and confirmed during this pass).
Shipping today, a **channel** declaring `batch` fails validation
([internal/domain/validation_records.go:32-35](../../../internal/domain/validation_records.go),
re-read during the folder's documentation pass); the lie to `catalog coverage` comes from the
other half — an **axes-level** `cadence=batch` declaration riding periodic channels, where the
catalog counts `batch` while the emitted behaviour is periodic. Both halves are why the
capability is a staged design change, not a data edit.

**The staged slice** (one bounded, domain-agnostic capability in `internal/world`): a batch
channel emits on its periodic schedule, buffers `batch_size` scheduled records, releases the
group at the Nth tick in scheduled order; per-record event time and seq are preserved; every
record keeps its own primary ledger row and unique `DeliveryID` (precedent
[internal/perturb/transform_duplicate.go:44-50](../../../internal/perturb/transform_duplicate.go);
per-seq `containsEveryEmission`
[internal/score/ledger_metrics.go:35-42](../../../internal/score/ledger_metrics.go)) — batches
are an **emission-count shape, not an aggregation**; buffering is keyed on the existing
per-entity/channel substream discipline with **no new RNG draws**; partial groups flush at
horizon end so quiescence holds.

**Required evidence before the domain admits**: a batch-shaped ledger boundary test in the
world package, byte-identical replay, and a >10k batch-shaped conservation run under
drop+`duplicate_burst` — the existing recipe
([internal/run/ledger_test.go:119-174](../../../internal/run/ledger_test.go),
`TestConservationAtScaleBeyondTenThousand`) is the template but is periodic-shaped, so batch
shape is unproven.

**Fallback**: if design review rejects the capability, the domain ships periodic and drops
`cadence=batch` from its axis claims (on the periodic fallback the idea re-scores below the
bar against open-field-irrigation's coverage).

**Hardest parts**: the aggregate-as-authored-data cross-check; batch-shaped conservation at
scale; the oracle is the leak deviation only — no closed form for the exact integral of seeded
flows (only the clean-world terms cancel in `conservationDeviation`).

Per gate:

- **G1 Determinism** — addressed. No new RNG draws (buffering keyed on the existing
  per-entity/channel substream discipline); byte-identical replay is required admission
  evidence. Batch grouping changes emission order and ledger row patterns, so it needs its own
  metamorphic evidence (group boundaries, group-internal emission order), not just a schema
  entry.
- **G2 Analytic cross-check** — addressed, as an obligation. Independent closed-form check for
  the leak integrator (new test code, `TestAnalyticCrossCheck` mold), the negative-recovery
  case across cells, the boundary oracle for onsets inside withheld batches, and the
  aggregate-as-authored-data assertion (district `metered_consumption` = Σ 24 submeter states).
- **G3 Reference consumer** — not addressed in `spec.feasibility`. The window-policy rubric is
  stated consumer-agnostically (neutral verdict vocabulary,
  [internal/model/verdict.go:37-45](../../../internal/model/verdict.go)), but no
  reference-consumer run is specified.
- **G4 Delivery ledger** — addressed. One-emission-one-primary-row semantics defined for batch
  members; unique non-zero `DeliveryID` per record; >10k batch-shaped conservation under
  drop+`duplicate_burst` required (existing >10k evidence is periodic-shaped).
- **G5 Quiescence** — addressed. `observe_only` removes the effector confound; perturbation
  windows are bounded inside the scenario duration
  ([internal/suite/scenario.go:67-77](../../../internal/suite/scenario.go)); partial batch
  groups flush at horizon end.
- **G6 Sealed oracle** — partially addressed. The design leans on sealed `first_observable`
  (delivery-invariant, computed on noiseless truth-state balances) and on sealed labels for
  in-window crossings (from world state). The spec does not name a prefix-equivalence
  obligation for the batch capability specifically.
- **G7 Injection probe** — not addressed in `spec.feasibility`.
- **G8 `silent_no_effect`** — not discussed explicitly; the surface is empty by construction
  (zero effectors), so there is no declared command to prove no-effect on.
- **G9 Trivial baseline** — carried from the original consolidated spec ("trivial baseline is
  a balanced ledger"); the final spec does not re-argue it.

## Quality bar result

| Dimension | Score (0–10) |
|---|---|
| consumerValue | 7 |
| dataExpressibility | 6 |
| determinismOracle | 7 |
| ledgerQuiescence | 6 |
| novelty | 6 |
| sealedTruth | 8 |
| specificity | 8 |
| **Total** | **48/70 raw = 69/100** (normalised, as recorded in the quality-bar data) |

**Bar**: total at least 85/100 across the seven dimensions, with no single dimension below
7/10. **Result: not met** — 69 < 85, and three dimensions sit below 7 (`dataExpressibility`
6, `ledgerQuiescence` 6, `novelty` 6). Note the per-lens review scores below (8/5/6/7/6) are
the five lenses' own scores, not these seven dimension scores.

## Five-lens review

### 1. Determinism and analytic oracle — 8/10

**Found**: batch-shaped arrival needs new runtime — `batch_size` is parsed-only with zero
consumers and `batch` cadence is routed identically to periodic, so batch grouping is new
emission-scheduling behaviour needing its own G1 metamorphic evidence (>10k conservation
check, group boundaries), not just a schema entry. The solver's conservation detector is a
between-worlds difference, not the wire residual, so the leak must also be visible in the
consumer-visible master−Σsubmeters residual or `first_observable` is computed against a
signal the consumer cannot form — the channel wiring must be stated. Citations verified
(perturb kinds, scenario windows, schema enums, zero shipped coverage for `cadence=batch` and
`actuation=observe_only`).

**Changed in response**: the final spec states the channel wiring explicitly (district-owned
`metered_consumption` aggregate; consumer-visible residual formable from the wire), requires
byte-identical replay and group-order determinism evidence for the batch capability, adds the
boundary oracle (onset inside a withheld batch labelled from world state), and adds the G2
negative-recovery case (onset and `first_observable` both inside the late-arrival window,
closed form vs RK4 across the fold/duplicate/replay cells).

### 2. Data-not-code expressibility — 5/10

**Found**: the admitted runtime gap is real and load-bearing — shipping today would emit a
periodic stream under a `cadence: ["batch"]` axis declaration and lie to `catalog coverage`
(the definition of secretly needing a new runtime concept, even though the capability is
domain-agnostic). The headline detector was **not computable by the declared forms**:
`conservation_residual` balances channels of one entity
([internal/truth/solver.go:117-123](../../../internal/truth/solver.go)) and the original
"balance = master − Σ submeters" spanned 25 entities; re-modelling 25 meters as 25 channels of
one entity would break per-entity fault injection and ledger semantics. Cross-checks that did
hold: schema legality of `batch_size`/`batch`, perturbation windows and kinds, and the G2
integral-oracle claim given the solver's clean-vs-faulted double run.

**Changed in response**: the detector was re-modelled **single-entity** — one district entity
carries `master_register`, `outlet_register`, and the district-owned aggregate
`metered_consumption` so `conservation_residual` is the declared form; the 24 submeters remain
entities for fault/ledger structure only, with the aggregate authored as data and asserted by
the analytic cross-check. The batch-emission capability decision is staged **before** domain
authoring as a design change, with the explicit fallback (ship periodic, drop the batch-axis
novelty claim).

### 3. Consumer testing value — 6/10

**Found**: the axis-carrying machinery was missing **twice**, and the original spec named only
one gap — `time_reference=batch_relative` was decorative (`TimeRef` parsed at
[internal/model/domain.go:31](../../../internal/model/domain.go) with zero runtime consumers),
so novelty was claimed on unimplemented surfaces. Without genuine batch-shaped arrival the
domain degenerates to conservation residuals under delay/reorder — open-field-irrigation
already runs conservation-style residuals on gross-lateness periodic streams — and the graded
content collapses toward existing coverage. The leak oracle itself is fine and
emission-independent, which is exactly why the arrival shape is load-bearing. Kill claims were
only partially checkable (targets not in this repo); the P1-staging citation was right in
substance but pointed at the wrong line.

**Changed in response**: the `time_reference=batch_relative` claim was **withdrawn** (vector
declares `wall`; event-time-vs-release-time anchoring is carried by the batch cell's ledger
instead). The batch capability is staged first as its own slice with a regression test proving
batch-shaped arrival in the delivery ledger; the fallback re-score against open-field coverage
(falls below the bar) is stated. The delivery-invariance of `first_observable` is now stated
as the design's spine — the property that makes the transport-miss/reasoning-miss verdict
falsifiable. Kill claims on `ev-charging-plaza` and `grain-storage-backfill` were withdrawn as
unverifiable here (grep confirms zero hits in the repo).

### 4. Delivery ledger and quiescence — 7/10

**Found**: the parsed-only `batch_size` gap was confirmed. Batch emission changes emission
counts, which feed the conservation invariant end to end (`World.EmittedCount()`,
[internal/run/ledger_test.go:150-174](../../../internal/run/ledger_test.go); per-seq
`containsEveryEmission`,
[internal/score/ledger_metrics.go:35-42](../../../internal/score/ledger_metrics.go)), and the
original spec defined no ledger semantics for a batch. The only >10k conservation evidence is
periodic-shaped (1ms heartbeat recipe,
[internal/run/ledger_test.go:119-153](../../../internal/run/ledger_test.go)). Quiescence
itself is safe: `observe_only` and bounded perturbation windows guarantee flush.

**Changed in response**: the batch emission unit is now specified — every record in a batch
group gets its own primary ledger row and unique `DeliveryID`
([internal/perturb/transform_duplicate.go:44-50](../../../internal/perturb/transform_duplicate.go)
precedent); batches are an emission-count shape, not an aggregation. Admission evidence now
explicitly requires the batch-shaped >10k conservation test (drop + `duplicate_burst` over
batch groups) and a boundary test in the world package, with the capability kept
domain-agnostic (a `batch_size` consumer in `internal/world`, no per-domain branch). Partial
groups flush at horizon end.

### 5. Novelty and portfolio balance — 6/10

**Found**: the fault family is catalog 15 water-distribution's, originally unacknowledged —
15's stresses sentence is nearly verbatim this domain's balance
([docs/design/DOMAIN_CATALOG.md:385-406](../../design/DOMAIN_CATALOG.md)). Two of the three
"zero-coverage" fills were shipped-relative only (`cadence=batch` already declared by catalog
02, [:105](../../design/DOMAIN_CATALOG.md); `time_reference=batch_relative` by catalog 11,
[:291](../../design/DOMAIN_CATALOG.md)); only `actuation=observe_only` is a genuine two-sided
gap (it appears in no catalog vector and in zero Go references — confirmed by the coverage run
and grep in this pass). `batch_relative` was declared but never mechanised — axis decoration,
which [DOMAIN_CATALOG.md:41-43](../../design/DOMAIN_CATALOG.md) calls a domain out for. The
`:656` citation for grain-storage P1 staging was wrong (line 656 is the table preamble; the
P1 row listing catalog 05 grain-storage is
[:659](../../design/DOMAIN_CATALOG.md)). And the set leans on a third conservation-leak fault
(SI-03, SI-05, SI-09).

**Changed in response**: repositioned honestly — "catalog 15's conservation fault under
batched late delivery", with the catalog's own redundancy-cut precedent
(`paint-booth-emissions` was cut for overlapping 15,
[:696-697](../../design/DOMAIN_CATALOG.md)); SI-03 is not a 26th domain but 15's slot filled
with the delivery axes 15 lacks. The `batch_relative` claim was withdrawn. The `:656`
citation was fixed to `:659`. The external kill claims were withdrawn as unverifiable.
Detector differentiation is stated (three-term district-aggregate summation balance with
pooled sigma vs open-field-irrigation's two-channel same-entity in/out balance on periodic
streams — [domains/open-field-irrigation.domain.json:192](../../../domains/open-field-irrigation.domain.json)
and cadence axis at [:11](../../../domains/open-field-irrigation.domain.json)), and the
third-slow-leak portfolio risk is flagged for the run owner.

## Improvement history

One improvement round, from the original consolidated spec to the final spec. The changes,
grouped by what drove them:

1. **Detector re-modelled from cross-entity to single-entity** (lens 2, the blocking
   finding). Original: "balance = master − Σ submeters" across 1 master meter + 24 submeters,
   with a `conservation_residual` detector — a form the solver cannot compute
   ([internal/truth/solver.go:117-123](../../../internal/truth/solver.go) reads one entity).
   Final: 1 district entity carries `master_register`, `outlet_register`, and the
   district-owned aggregate `metered_consumption`; the submeters stay entities whose registers
   the consumer sums on the wire; the aggregate is authored data asserted by the analytic
   cross-check.
2. **Batch emission staged as an explicit pre-domain design change with a fallback**
   (lenses 1–4). Original feasibility said "cadence batch is routed
   (internal/world/cadence.go:23-24)" — literally true but misleading, since `batch` routes to
   `periodicEmission`; the gap (`batch_size` parsed-only) was named but with no plan. Final:
   the routing is stated as a blocker, the capability is specified (buffer `batch_size`
   scheduled records, release at the Nth tick, per-record event time/seq/ledger-row/DeliveryID,
   no new RNG draws, partial-group flush), admission evidence is named (batch-shaped ledger
   boundary test, byte-identical replay, >10k batch-shaped conservation under
   drop+`duplicate_burst`), and the fallback (ship periodic, drop the axis claim, re-score
   below the bar) is explicit.
3. **Oracle claim corrected** (lens 1). Original: "oracle is the exact integral of seeded
   flows ± leak·(t−t_onset), integrated by the solver's RK4 core and independently in closed
   form (G2)" — an overclaim. Final: only the leak deviation `leak·(t−t_onset)` has a closed
   form (clean-world terms cancel in `conservationDeviation`,
   [internal/truth/detector_math.go:31-37](../../../internal/truth/detector_math.go)); the
   solver derives no per-form closed forms (`Method=SolveNumeric`); the independent check is
   new test code and an acceptance obligation. The G2 negative-recovery case and the boundary
   oracle for onsets inside withheld batches were added.
4. **`time_reference=batch_relative` claim withdrawn** (lenses 3 and 5). `TimeRef` is
   parsed-only with zero runtime consumers, so filling the axis changes nothing observable.
   The vector declares `wall`; anchoring is carried by the batch cell's ledger.
5. **Novelty repositioned honestly** (lens 5). Original claimed three zero-coverage axis fills
   and killed two external ideas. Final: acknowledges catalog 15 as the fault family (with the
   `paint-booth-emissions` cut precedent, [:696-697](../../design/DOMAIN_CATALOG.md)); marks
   `cadence=batch` real only if the capability lands (catalog 02 already declares it,
   [:105](../../design/DOMAIN_CATALOG.md)); identifies `observe_only` as the only two-sided
   gap; withdraws the `ev-charging-plaza`/`grain-storage-backfill` kills as unverifiable in
   this repo; fixes the `:656` citation to `:659`; flags the third-conservation-leak portfolio
   risk.
6. **Delivery-invariant `first_observable` stated as the design spine** (lens 3), and the
   window-policy matrix made concrete as four scored cells (A–D) mapped to the neutral verdict
   vocabulary ([internal/model/verdict.go:37-45](../../../internal/model/verdict.go)) and
   admission labels ([internal/score/admission.go:43-62](../../../internal/score/admission.go)),
   replacing the original's consumer-flavoured "billing on arrival time" framing ("billing"
   appears only in domain prose, never in the binary).
7. **Smaller corrections**: `drop` added to the perturbation set; scenario-window citation
   corrected to [internal/suite/scenario.go:67-77](../../../internal/suite/scenario.go);
   effectors stated as an explicit empty set (schema-legal omission,
   [internal/schemas/domain-spec-v0.1.schema.json:8](../../../internal/schemas/domain-spec-v0.1.schema.json)).

The scores that resulted — `dataExpressibility` 6 and `ledgerQuiescence` 6 remain capped by
the unlanded batch capability, `novelty` 6 by the repositioned (and partially conditional)
axis fills — are why the idea still sits at 69/100, below the bar, after the round.

## Open questions

1. **Does design review accept the batch-emission capability?** The whole delivery story and
   two of the axis fills depend on it; the stated fallback (ship periodic, drop
   `cadence=batch`) re-scores the idea below the bar against open-field-irrigation's existing
   conservation-residual coverage. If the fallback is taken, does the idea proceed at all?
2. **Where does the aggregate-as-authored-data cross-check live** — `internal/world` or
   `internal/truth` — and does the authored `metered_consumption` = Σ 24 submeter states
   assertion need its own G2-style test identity beyond the leak-integrator check?
3. **Sealed labels inside withheld windows (cell C)**: the boundary oracle says the crossing
   label comes from world state. Does producing that label need solver support beyond the
   current onset scan (`searchOnsets`), and how is the sealed label validated against the
   withheld batch boundary?
4. **Batch-scale parameters for the >10k evidence**: what `batch_size`/period combination
   reaches >10k records in the suite harness, and does byte-identical replay cover the
   partial-group flush at horizon end?
5. **Portfolio balance**: SI-03, SI-05 and SI-09 are all conservation-leak specs (per the lens
   review; those specs are not in this repo, so not verifiable here). Should one convert to a
   different fault family, and who owns that call — flagged for the run owner.
6. **Batch-relative anchoring, later**: the `time_reference=batch_relative` stamp format was
   withdrawn. Is the ledger-carried event-time-vs-release-time anchoring sufficient for the
   consumer rubric, or does a batch-origin stamp need a schema extension in a later round?
