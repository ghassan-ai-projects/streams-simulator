# SI-06 · ev-plaza-firmware-split

| | |
|---|---|
| ID / slug | SI-06 · `ev-plaza-firmware-split` |
| Status | Candidate evaluated against the quality bar — **bar not met** (74/100; see Quality bar result) |
| Documented | 2026-10-05 |
| Improvement rounds | 2 (consolidated first draft → final spec) |
| Consumer target | agentic-stream — `docs/design/CONSUMERS.md` §5.2 |
| Neighbors | shipped: `discrete-line-oee` (DOMAIN_CATALOG §09, `docs/design/DOMAIN_CATALOG.md:240`); catalog: `ev-fleet-telematics` (§18, `docs/design/DOMAIN_CATALOG.md:458`); siblings: SI-07 (quarantine boundary), SI-10 (`peer_residual` reservation) |

## Purpose

A DC fast-charging plaza whose entire fleet crosses a firmware migration mid-run: for one
bounded window every record's unit label is stripped to `"err"` (values stay Wh-scale), the
charger-state channel emits out-of-vocabulary tokens, and poison records (nan/inf, oversize,
malformed) ride the same window — while one charger has a genuine contactor-thermal-derate
fault (a flapping weld/bounce derate under thermal cycling) that must be diagnosed through
the drift.

The split is temporal (pre/post-migration dialects), not cohort-scoped: perturbations are
window-global (`internal/perturb/perturb.go:141-154`) with fail-closed params and no entity
key (`internal/perturb/parameters.go:63-84`; `internal/mcp/world_commands.go:105-111`), so
the original half-fleet cohort premise was inexpressible without a schema extension.

The Wh↔kWh step is stated honestly: `unit_mismatch` never rescales
(`internal/perturb/transform_payload.go:41-50`), so the 1000x error exists only when a
consumer coerces the unitless window to its assumed dialect — a self-inflicted corruption
the sealed, unit-exact oracle scores as the reasoning miss it is.

The domain exposes the three-way failure: quarantine loudly and miss the diagnosis
deadline; coerce units and corrupt every peer comparison; stall and miss everything. It runs
against agentic-stream, whose documented closed-loop obligations (classify `ack_lost` vs
`reject` vs `interlock_refused`, idempotent dispatch, quiescence — `docs/design/CONSUMERS.md`
§5.2, verified at `docs/design/CONSUMERS.md:104-109`) are exactly what a
quarantine-vs-coerce decision collides with.

## How it runs on this simulator

One JSON domain file plus a suite profile plus acceptance tests; zero new machinery; no code
branches (domains are data). Everything below is schema-declarable.

### Entities

40 DC fast chargers; one fleet; no cohorts.

### Signals (channels)

- `commanded_current_a` — observes `current_setpoint`, `observation_gain` 1.0, offset 0. The
  setpoint echo, expressible as data via `Channel.Observes`
  (`internal/model/domain.go:148-151`).
- `delivered_current_a` — observes `delivered_current`; σ = 2 A.
- `contactor_temp_c` — observes `contactor_temp`; σ = 0.5 °C. Sigmas are chosen large enough
  that hindsight thresholds cannot cheat on noise, per the schema's own warning that "an
  unrealistically small sigma silently makes every scenario easy"
  (`docs/contracts/domain-spec-v0.1.schema.json:301`).
- `energy_wh` — `value_type` counter, unit `"Wh"`.
- `charger_state` — `value_type` string, `enum_values` `[idle, precharge, charging,
  discharging, fault]`, event-driven on trigger state `current_setpoint`. Values are random
  draws from `enum_values` at runtime (`internal/world/reading_values.go:74-80`), so
  diagnosis stays numeric-only by construction.
- No `attacker_controlled` channel is declared.

### Dynamics (states)

- `current_setpoint` — F0, baseline 350 A, two declared periodic components: seasonality
  `[{period_s: 3600, amplitude: 40}, {period_s: 14400, amplitude: 25}]`. The schema's only
  seasonality concept is periodic `period_s`/`amplitude`/`phase_rad`
  (`docs/contracts/domain-spec-v0.1.schema.json:114-127`); there is no "session seasonality"
  concept, only this named approximation.
- `delivered_current` — F1 `first_order_lag` on `current_setpoint`, gain 1.0,
  `time_constant_s` 4.
- `contactor_temp` — F1 `rc_network` on `delivered_current`, tau1 300 s, tau2 60 s, gain 0.1.
- `energy_wh` — F1 `integrator` on `delivered_current`, gain 1.

### Faults (world-side; magnitudes fixed by the admission proof)

- `contactor_thermal_derate` (positive class) — onset intermittent, `rate_per_hour` 2,
  `duty_cycle` 0.5 (the committed envelope honors `duty_cycle`,
  `internal/world/dynamics.go:109-120`), affects `delivered_current` delta −120 A. Detector:
  `channel_divergence`(`channel_a`: `commanded_current_a`, `channel_b`:
  `delivered_current_a`) — supported today (`internal/truth/detector_math.go:16-22`; the
  schema requires `channel_a`/`channel_b` for this form,
  `docs/contracts/domain-spec-v0.1.schema.json:423-424`). `peer_residual` is deliberately
  **not** the oracle: it pools all siblings with no cohort filter
  (`internal/truth/detector_math.go:39-50`, `internal/truth/solver.go:108-115`) and is
  reserved for SI-10's cohort-agnostic residual.
- `demand_response_event` (negative class) — onset intermittent, `rate_per_hour` 2,
  `duty_cycle` 0.5, affects `current_setpoint` delta −120 A. This is the **flap-matched
  confound**: fault and benign control share shape, rate, duty and depth, so every
  per-channel distribution coincides (phase is the only difference and none of the five
  audit detectors uses timing), and only commanded-vs-delivered divergence separates them.
  Its own detector is `single_channel_snr` on `delivered_current_a` — and it fails to
  separate, which is the point.

### Perturbations (delivery-side, window-global)

All windows are `[FromNS, UntilNS)` and apply to every event in the window
(`internal/perturb/perturb.go:141-154`); parameters fail closed, and the three poison kinds
accept no parameters at all (`internal/perturb/parameters.go:63-84`).

- **Migration window**: `unit_mismatch` (no parameters; every record's unit label becomes
  `"err"`, value untouched — `internal/perturb/transform_payload.go:41-50`) plus
  `out_of_enum` (rate 0.5; `undeclared_<0..999>` tokens on `charger_state` —
  `internal/perturb/transform_payload.go:10-23`; *not* named v2 states — that claim was
  dropped as inexpressible).
- **Poison window** (same `FromNS`/`UntilNS`): `nan_inf` (no parameters; value becomes
  `"NaN"`/`"Infinity"` strings — `internal/perturb/transform_payload.go:67-77`), `oversize`
  (bytes 4096 — `internal/perturb/transform_payload.go:83-103`), `malformed` (every record
  flagged Malformed — `internal/perturb/transform_payload.go:56-65`; each written as one
  broken JSONL line without poisoning the rest of the file —
  `internal/run/emission.go:61-67`, `internal/run/delivery.go:124-133`).

### Ground truth and its analytic oracle

- Sealed from the noiseless world diff: the solver reads a Noiseless world
  (`internal/world/observe.go:19-20`), and the perturbation layer changes nothing physical
  (`internal/perturb/perturb.go:3-4`). `first_observable_time` comes from the divergence
  crossing the combined noise floor and is unaffected by delivery corruption — even when a
  derate on-window opens inside the poison window.
- Oracles, double-implemented: the first-order lag closed form
  `x(t) = g·u·(1−e^(−t/τ))` (`internal/world/oracle_test.go:57-75`), the `rc_network`
  cascade closed form (`internal/world/oracle_test.go:24-56`), and `delta·envelope(t)` with
  the intermittent envelope closed form (`internal/world/oracle_test.go:221-266`).

### Effector (data)

- `limit_current`, advisory. Args `{charger_id, limit_a 0..500}`.
- Ack 100 ms with the full failure-mode family — ok 0.90, slow 0.03, `ack_lost` 0.02,
  `reject` 0.03, `confirmed_no_effect` 0.01, `silent_no_effect` 0.01 (the rotating-machinery
  `stop_motor` template, `domains/rotating-machinery.domain.json:307-315`; mode constants
  `internal/world/effectors.go:11-18`). `silent_no_effect` is what G8 scores on —
  `internal/score/loop_metrics.go:35` keys on call mode `"silent_no_effect"`, and
  confirmation reads the shadow value through the declared confirmation channel
  (`internal/world/state.go:55-63`).
- `confirmation_channels` `[commanded_current_a]`; effect `state_deltas`
  `[{state: current_setpoint, assign_from_arg: limit_a}]`; `time_constant_s` 20 (schema
  fields verified: `docs/contracts/domain-spec-v0.1.schema.json:513-543`).
- Fault `deadline_s` 1800, counterfactual `if_no_action =
  session_throttled_thermal_escalation`, `if_action_by_deadline = safe_cap_until_service`.

## Feasibility against G1–G9

Gate names follow the non-negotiables table in `docs/QUALITY_BAR.md:57-72`.

### One schema, zero new machinery

The domain is one JSON file plus a suite profile plus acceptance tests. Every corruption is
a committed perturbation kind plus a scenario window; every dynamic and detector form is a
committed schema concept. The curated five-kind migration window and the flap-matched fault
pair are the only design choices, and both are data.

### Admission magnitudes are evidence, not assertion (G9)

*Reported from the idea-development run; not re-runnable in this documentation pass — the
scratch directory was removed by design (see Open questions).*

A scratch harness (candidate specs generated in `/tmp/si06-scratch`, loaded with the
committed `domain.Load`, audited with the committed panel; run via a temporary untracked
`cmd/si06scratch` directory because `internal/` packages cannot be imported outside the
module, removed afterward with `git status --short` clean) executed `go run ./cmd/si06scratch
<spec>` against three schema-valid variants:

| Variant | Verdict |
|---|---|
| Original magnitudes (control −60 A, step fault) | TRIVIAL — best `fixed_threshold` 0.957 |
| Control deepened to −120 A, sustained | still TRIVIAL at 0.902 — exactly the boundary the critic predicted |
| Flap-matched design (fault and control both intermittent, 2/h, duty 0.5, both −120 A) | admits in both cells — clean 0.830 non-trivial; five-kind poison window over onset 0.668 non-trivial; all five detectors below the 0.9 bar (`fixed_threshold` 0.830, `channel_silence` 0.726, `moving_median_residual` 0.662, `zscore` 0.608, `first_difference` 0.584); `first_observable_time` = injection time (lag 0) |

This reproduces the determinism lens's finding by execution and fixes it in the design:
hindsight sweeps are per-channel with the whole control arm scored label-0
(`internal/audit/statistics.go:8-67`, `internal/audit/detector.go:81-88`), so per-channel
class overlap is a design requirement, and the flap-match delivers it while divergence stays
the only separator. Caveat stated in the spec: that is one seed, one onset, one geometry;
the final domain JSON must reproduce the non-trivial verdict across the suite's seed sweep,
and scenarios that do go trivial are excluded by the audit automatically — the domain gate
is suite fill (`min_scenarios` with coverage), not per-scenario luck.

### Suite mechanics (G1)

The generator draws windows per perturbation name with ≥5 coverage each across all 19 kinds
(`internal/suite/commands.go:60-68`; catalog `internal/perturb/perturb.go:43-45`), window
start uniform in the first half and length 0.25–0.75 of duration/2
(`internal/suite/scenario.go:67-77`), onset in the first half
(`internal/suite/scenario_draw.go:21-34`) — so poison-over-onset overlap, the actually
graded confound, arises by draw and is mass-producible with zero suite changes.

### The curated five-kind migration window

Honestly scoped as curated-artifact-only: profiles carry
`fault_weights`/`entity_count`/`duration`/`setup` and no perturbation composition
(`internal/model/effector.go:62-71`), so it exists as a run-level scenario pinned by its
command-log artifact (`internal/suite/command_log.go:21-24`) and replayed from it
(`ApplyPerturb` accepts arbitrary windows, `internal/mcp/world_commands.go:105-111`). No
suite machinery change is claimed or proposed.

### Audit poison policy (G9 determinism)

Stated as the named deterministic rule the code already implements: the audit's series
builder excludes undelivered records, Malformed-flagged records, and non-numeric payloads —
by ledger flag and type, never by value inspection (`internal/audit/scenario.go:49-61`) — so
mangled-but-numeric `unit_mismatch` rows stay in, `nan_inf`/`oversize`/`malformed` rows
drop, and because the same perturbations are applied to both arms
(`internal/audit/audit.go:69-80`) the verdict cannot flip between replay cells; the measured
cells agree (same-shape scores, poison lowering all detectors uniformly).

### Why the migration is not a fault (G6)

The multiplicative fault-effect flag exists
(`docs/contracts/domain-spec-v0.1.schema.json:392`) but is deliberately unused: expressing
the migration as a world fault would make it the labeled fault in every scenario (the
ground-truth label is one fault id; one fault per scenario —
`internal/audit/audit.go:70-74`), destroying the drift-as-context design.

### Required acceptance evidence — deliverables, not evidence in hand

1. **(G7)** This domain's own probe/poison-neutrality differential following
   `internal/run/probe_test.go:118-195` — delivered bytes differ, consumer verdict
   byte-identical — with the benign control run defined per poison window as the same
   windows with no poison. That test covers only the synthetic fixture and
   discrete-line-oee today (verified by running `go test ./internal/run -run
   TestInjectionProbeNeutralAcrossDeliveryPaths -count=1` → `ok
   github.com/ghassan-ai-projects/streams-simulator/internal/run 0.702s` in this pass), so
   neutrality is not free for a new domain.
2. **(G9/G6)** A metamorphic test — the identical scenario with and without the poison
   window yields the same fault label and the same `first_observable_time_ns` (the
   audit-level analog is measured above: both cells non-trivial on the same scenario).
3. **(G9)** The admission proof re-run on the final domain JSON across the seed sweep,
   extending the scratch result (0.830 clean / 0.668 poison, single seed) to a filled suite.
4. **(G4)** A >10k-event conservation run with terminal mangled rows present as the domain's
   G4 evidence.

### Gate-by-gate summary

| Gate | Spec claim | Evidence status |
|---|---|---|
| G1 Determinism | Data-only domain, fixed windows, sealed seed; rides the platform's byte-identical replay | implicit; no new machinery claimed |
| G2 Analytic cross-check | Lag, `rc_network` and envelope closed forms double-implemented | in hand — `internal/world/oracle_test.go:24-56`, `:57-75`, `:221-266` (read in this pass) |
| G3 Reference consumer | No domain-specific claim; agentic-stream runs the documented operator surface (CONSUMERS.md §5.2) | platform gate, unchanged |
| G4 Delivery ledger | >10k conservation with terminal mangled rows | required deliverable (4) |
| G5 Quiescence | No domain-specific claim | platform gate, unchanged |
| G6 Sealed oracle | Truth from the Noiseless diff; delivery corruption cannot touch it | in hand — `internal/world/observe.go:19-20`, `internal/perturb/perturb.go:3-4` |
| G7 Injection probe | Domain's own probe/poison differential | required deliverable (1); existing test re-run in this pass covers synthetic + discrete-line-oee only |
| G8 `silent_no_effect` | Full ack family declared; scenarios exercise the G8 machinery | mechanisms in hand — `internal/score/loop_metrics.go:35`, `internal/world/state.go:55-63`; domain-level tests are deliverables |
| G9 Trivial baseline | Admission proof by execution + metamorphic + seed sweep | scratch evidence is single-seed (not re-runnable here); deliverables (2) and (3) |

## Quality bar result

| Dimension | Score |
|---|---|
| consumerValue | 7/10 |
| dataExpressibility | 8/10 |
| determinismOracle | 7/10 |
| ledgerQuiescence | 7/10 |
| novelty | 7/10 |
| sealedTruth | 7/10 |
| specificity | 9/10 |
| **Total** | **52/70 = 74/100** |

**Bar:** total at least 85/100 across the seven dimensions, with no single dimension below
7/10.

**Result: not met.** The total (74/100) is below the required 85/100. The
minimum-dimension criterion is satisfied (lowest score is 7/10; no dimension scores below
7), but the bar requires both conditions. The headroom sits exactly where the four required
acceptance deliverables above are: executing them converts the spec's remaining
uncertainties (seed-sweep admission, domain-specific G7 neutrality, metamorphic label
stability) into evidence.

## Five-lens review

The lenses reviewed the original consolidated spec; the per-lens scores grade that draft.
The seven dimension scores above grade the final spec after both improvement rounds.

### 1. Determinism and analytic oracle — 7/10

**Found:**
- The G2/G9 boundary risk is real and only half-mitigated: the trivial-baseline audit fits
  detectors with hindsight on the delivered (perturbed) stream, and the poison window
  deliberately puts nan/inf/malformed/oversize content in numeric channels. "Diagnosis runs
  on numeric channels only" does not protect the audit panel itself from poison; the spec
  needed to state how the audit's numeric fits treat mangled records (skip vs impute)
  deterministically, or the trivial verdict can flip between replay cells.
- The derate oracle's "commanded current" divergence channel must be a state observed by a
  channel (`channel_divergence` requires `channel_a`/`channel_b`,
  `docs/contracts/domain-spec-v0.1.schema.json:423-424`) — a setpoint echo channel the
  original sketch assumed rather than specified.
- G6 is clean: labels come from the noiseless world diff, so delivery-side poison cannot
  touch sealed truth; probe-neutrality evidence for G7 is committed
  (`docs/QUALITY_BAR.md:126`).

**Changed in response:**
- The audit's poison policy is now a named deterministic rule — mangled/malformed rows
  excluded by ledger reason, never by value inspection (`internal/audit/scenario.go:49-61`)
  — plus a metamorphic test requirement: identical scenario with and without the poison
  window yields identical fault labels and identical `first_observable_time`.
- The setpoint-echo channel is specified (`Channel.Observes`,
  `internal/model/domain.go:148-151`) and the divergence oracle's boundary is asserted:
  a derate on-window opening inside the poison window still yields
  `first_observable_time` from the world diff, unaffected by delivery corruption.

### 2. Data-not-code expressibility — 6/10

**Found:**
- Perturbations are window-global with no entity scoping: `Layer.Process` applies every
  active perturbation to every event in the window (`internal/perturb/perturb.go:141-154`).
  "unit_mismatch on the migrated cohort" and cohort-specific new enum names cannot be
  scoped to half the chargers.
- Enum channel values are random draws from `enum_values`, not state-derived
  (`internal/world/reading_values.go:74-80`), so "new enum state names appear" is only
  realizable as `out_of_enum`'s random `undeclared_###` values in a global window.
- Realizing the 1000x unit step as a per-entity multiplicative FAULT (the only per-entity
  mechanism) collides with the one-fault-per-scenario label model; verified good: all cited
  corruption kinds exist, probe neutrality supports G7, and numeric-only diagnosis is
  correct given the random-enum runtime.

**Changed in response:**
- Restructured to one global migration window (whole plaza migrates) with the
  thermal-derate fault inside it — the quarantine-vs-coerce-vs-stall consumer matrix
  survives intact. No entity-scoped-perturbation design change was pre-committed.
- Cohort enum vocabularies dropped; the drift carries as the temporal dialect split plus
  `out_of_enum`'s `undeclared_<0..999>` tokens — all data.

### 3. Consumer testing value — 7/10

**Found:**
- "peer_residual against same-dialect chargers" is not expressible: peer deviation pools
  ALL sibling entities with no cohort filter (`internal/truth/detector_math.go:39-50`; the
  Detector struct has only form/channel/channel_a/channel_b/inputs/outputs — no sibling-set
  selector). With two dialect cohorts, a mixed peer pool during unit_mismatch flags the
  whole migrated cohort.
- The poison-window/fault-onset overlap was asserted without a scratch proof: the audit
  runs clean and faulted worlds each WITH the scenario's perturbations
  (`internal/audit/audit.go:69-80`) — fairness is protected, but whether balanced accuracy
  stays below 0.9 was exactly the open question.
- G7 composition is strong, but the probe-neutrality metric needs the benign-control run
  defined per poison window, which the original sketch did not specify.

**Changed in response:**
- The fault's detector swapped to `channel_divergence` between two numeric channels on the
  same charger (`commanded_current_a` vs `delivered_current_a`), supported today
  (`internal/truth/detector_math.go:16-22`); `peer_residual` explicitly reserved for
  SI-10's cohort-agnostic cross-venue residual; the sibling-set selector design change
  explicitly refused.
- The poison-window control run is defined (same windows, no poison) so injection/poison
  neutrality is a differential with an exact expected result, and one scratch scenario was
  executed with the poison window active, admitting **non-trivial**: the hindsight panel's
  best detector scored 0.668, below the 0.9 exclusion bar (`internal/audit/audit.go:18`) —
  i.e. the poison window does **not** make the scenario trivially solvable, which is the G9
  requirement (development-run evidence; wording correction, folder documentation pass: an
  earlier draft said the scratch run "proves the audit still separates derate from controls",
  which is backwards — a hindsight score ≥ 0.9 would have meant the audit *does* separate them
  and the scenario is trivial-excluded; 0.668 < 0.9 is the pass condition, and separation for
  a real consumer remains the commanded-vs-delivered divergence's job).

### 4. Delivery ledger and quiescence — 7/10

**Found:**
- The poison kinds are committed and terminal-conserving: `nan_inf` writes NaN/Infinity
  tokens, `oversize` enlarges deterministically, `malformed` writes one broken line and
  records a terminal mangled row (`internal/run/delivery.go:124-133`,
  `internal/run/emission.go:61-67`) — none touch sealed world-side truth, which (not the
  probe) is the actual reason poison cannot alter oracle evidence.
- The G7 claim was overstated: the neutrality differential covers `injection_probe` on
  `attacker_controlled` channels only and is tested against the synthetic fixture plus
  discrete-line-oee (`internal/run/probe_test.go:118-195`) — a new domain gets no
  neutrality evidence for free.
- "Diagnosis runs on numeric channels only" was stated as intent; it should be an admission
  rule, or enum/oversize content can leak into suite-detector inputs.

**Changed in response:**
- This domain's own probe/poison-neutrality differential added to its acceptance evidence
  (delivered bytes differ, consumer verdict byte-identical), following the
  `probe_test.go:118-195` pattern.
- "Poison window never overlaps fault onset in truth, only in delivery" made a sealed
  scenario rule: `first_observable_time` comes from the world diff, so delivery corruption
  cannot move it.
- A >10k conservation run with malformed rows present (terminal mangled rows, file not
  poisoned) added as the domain's G4 evidence.

### 5. Novelty and portfolio balance — 8/10

**Found:**
- Shares its poison/malformed/quarantine surface with SI-07; the two candidates must be
  sharply partitioned (migration-window poison vs injected instruction text).
- Nearest shipped neighbor discrete-line-oee already keeps out-of-enum values as declared
  evidence (`docs/design/DOMAIN_CATALOG.md:240`, §09: "the out-of-enum values the stream is
  specified to keep as evidence") — the `out_of_enum` half is an intensification, not a
  first; the coordinated migration and unit step are the genuinely new parts.
- The kill claim (lab-analyzer-recalibration, "scored lowest at 6") was unverifiable — the
  name and score existed in no repo file at this document's writing pass. (Scope note, folder
  documentation pass: the pool record in [../research-notes.md](../research-notes.md) §1 now
  preserves both the name and the self-score 6 — created later the same day; the substance of
  the withdrawal is unchanged, and the claim's verification is now the scoped re-run recorded
  under *Evidence recorded* below.)
- Suite admission determinism (overlapping-window audit behavior) was undemonstrated.

**Changed in response:**
- The SI-07 boundary is explicit: this domain owns drift-driven quarantine under a diagnosis
  deadline (deadline_s plus counterfactual scoring over delivery-ledger mangled rows); only
  the repo-verifiable half of the earlier partition claim is retained.
- Repositioned against §09's out-of-enum-as-evidence (the delta is coordinated migration +
  poison as a competing hypothesis, not operator fatigue) and named ev-fleet-telematics
  (§18, `docs/design/DOMAIN_CATALOG.md:458` — `charger_negotiation_failure`,
  `limit_charge_rate`; a mobile fleet with peer-relative labels: no plaza-side migration
  window, no divergence label) as the closest catalog neighbor.
- The unit-mismatch counterfactual is stated up front in ground truth: peer comparison
  across the coerced/uncoerced boundary is the trap, and the oracle stays unit-exact.
- The kill claim is withdrawn (grep in this pass confirms `lab-analyzer-recalibration`
  appears nowhere in `*.go`/`*.json`/`*.md`).

## Improvement history

Two rounds separate the original consolidated spec from the final spec. The original
proposed a half-fleet migration with a Wh↔kWh 1000x semantic step "passing any registry
compatibility check", new enum state names, a `peer_residual` same-dialect oracle, a joint
run by agentic-stream *and* tamoz, and a kill of lab-analyzer-recalibration. What survived
both rounds: the plaza, 40 chargers, the contactor-thermal-derate diagnosis under drift, the
advisory `limit_current` effector with deadline and counterfactual, the
quarantine-vs-coerce-vs-stall consumer matrix, and "one schema, zero new machinery".

### Round 1 — align the design with what the code can express

Driven by the expressibility, consumer-value and determinism lenses:

1. **Half-fleet cohort → whole-plaza temporal split.** Perturbations are window-global with
   no entity key (`internal/perturb/perturb.go:141-154`,
   `internal/perturb/parameters.go:63-84`; verified in this pass).
2. **Oracle swapped.** `peer_residual` pools all siblings with no cohort filter
   (`internal/truth/detector_math.go:39-50`, `internal/truth/solver.go:108-115`) and was
   replaced by `channel_divergence(commanded_current_a, delivered_current_a)`
   (`internal/truth/detector_math.go:16-22`); `peer_residual` reserved for SI-10.
3. **Setpoint echo channel specified as data** via `Channel.Observes`
   (`internal/model/domain.go:148-151`) rather than assumed.
4. **"New enum state names" dropped.** Channels draw enum values at random
   (`internal/world/reading_values.go:74-80`); the drift carries as `out_of_enum`'s
   `undeclared_<0..999>` tokens (`internal/perturb/transform_payload.go:10-23`).
5. **Multiplicative per-entity fault route rejected** (the flag exists at
   `docs/contracts/domain-spec-v0.1.schema.json:392` but is deliberately unused): the
   migration would be the labeled fault in every scenario
   (`internal/audit/audit.go:70-74`), destroying drift-as-context.
6. **tamoz co-runner withdrawn;** consumer grounded in agentic-stream's documented §5.2
   obligations. (This pass verified no consumer named tamoz exists — but see the spec-text
   correction under Open questions.)
7. **Audit poison policy named and grounded** (`internal/audit/scenario.go:49-61`;
   `internal/audit/audit.go:69-80`).
8. **G7 claim corrected** from "the existing probe neutrality already guarantees" to a
   required domain-specific deliverable: `internal/run/probe_test.go:118-195` covers only
   the synthetic fixture and discrete-line-oee.

### Round 2 — convert assertions into execution and scope honestly

1. **Admission proof by execution.** Three schema-valid variants through the scratch
   harness; TRIVIAL → TRIVIAL → admits; the flap-matched fault/control pair became the
   design (numbers under Feasibility; development-run evidence, scratch removed).
2. **The 1000x Wh↔kWh step restated.** `unit_mismatch` never rescales
   (`internal/perturb/transform_payload.go:41-50`); the step exists only when a consumer
   coerces; the oracle stays unit-exact and the miss is scored where it physically arises —
   in the consumer.
3. **The curated five-kind migration window scoped honestly** as curated-artifact-only:
   profiles carry no perturbation composition (`internal/model/effector.go:62-71`); pinned
   via the command-log artifact (`internal/suite/command_log.go:21-24`), replayed via
   `ApplyPerturb` (`internal/mcp/world_commands.go:105-111`); no suite machinery change
   claimed.
4. **Four required acceptance deliverables enumerated** as deliverables, not evidence in
   hand (G7 differential, metamorphic label/latency stability, seed-sweep admission proof,
   >10k conservation with mangled rows).
5. **Full ack failure-mode family declared** (including `silent_no_effect` 0.01 and
   `confirmed_no_effect` 0.01) so the domain's own scenarios exercise the G8 false-success
   machinery (`internal/score/loop_metrics.go:35`,
   `internal/world/state.go:55-63`) — the earlier ack_lost-only sketch never would have.
6. **Neighbors named and checked:** discrete-line-oee (§09; out-of-enum-as-evidence) and
   ev-fleet-telematics (§18; mobile fleet, no migration window, no divergence label); no
   shipped or catalog domain combines a migration window with an in-window world fault.
7. **Kill claim withdrawn:** lab-analyzer-recalibration appeared in no repo file at this
   document's writing pass (grep in that pass: no hits; the name is now preserved in
   [../research-notes.md](../research-notes.md) §1, created later the same day — see the
   scoped re-run under *Evidence recorded* below).

## Open questions

1. **Seed-sweep robustness.** The admission verdict (0.830 clean / 0.668 poison) is one
   seed, one onset, one geometry, and the scratch harness was removed by design — it cannot
   be re-run from this document. Does the final domain JSON reproduce non-trivial admission
   across the suite's seed sweep, or do magnitudes need retuning?
2. **Shape of the G7 differential.** Should the domain's probe/poison-neutrality test extend
   the table in `internal/run/probe_test.go:118-195` (currently synthetic fixture +
   discrete-line-oee only) or live as a new domain-scoped test? The benign control per
   poison window (same windows, no poison) needs a committed definition either way.
3. **Distribution channel for the curated window.** The five-kind migration window is
   curated-artifact-only because profiles cannot compose perturbations
   (`internal/model/effector.go:62-71`). Is command-log pinning and replay
   (`internal/suite/command_log.go:21-24`, `internal/mcp/world_commands.go:105-111`)
   enough, or does the portfolio want a declared curated-scenario mechanism (a design
   change SI-06 deliberately does not request)?
4. **SI-10 dependency.** `peer_residual` has no sibling-set selector
   (`internal/truth/detector_math.go:39-50`). If SI-10's cohort-agnostic residual needs one,
   that is the detector-struct design change SI-06 explicitly refused — it should be scoped
   against SI-10, not absorbed here.
5. **Spec-text correction.** The final spec's novelty section says "tamoz … appear nowhere
   (grep over *.go/*.json/*.md returns nothing)". Re-run in this pass: `grep -rni tamoz`
   over those extensions returns 10 hits, all as a workflow/tooling name (e.g.
   `AGENTS.md:139`, `docs/design/CRITICAL_REVIEW.md:145-161`), never as a consumer or
   candidate. The consumer-withdrawal stands on substance, but the sentence should be
   reworded to "appears nowhere as a consumer" before this document is quoted.
6. **Unit-exact counterfactual representation.** The spec asserts the oracle stays
   unit-exact and the 1000x coercion miss is scored in the consumer. It does not specify
   whether `ground-truth-v0.1` needs an explicit unit-exact counterfactual field or whether
   the existing label plus the sealed noiseless diff suffices at score time.

## Evidence recorded in this documentation pass

- Ran `go test ./internal/run -run TestInjectionProbeNeutralAcrossDeliveryPaths -count=1`
  → `ok  github.com/ghassan-ai-projects/streams-simulator/internal/run  0.702s`.
- Ran `grep -rni tamoz` over `*.go`/`*.json`/`*.md` (excluding `.git/`) → 10 hits, all
  workflow/tooling references (`AGENTS.md:139`, `docs/design/CRITICAL_REVIEW.md:145,153,161`,
  `docs/design/GAP_ANALYSIS.md:46`, `docs/DOCUMENTATION_ORGANIZATION_PLAN.md:3,117`,
  `docs/refactoring/clean-code-20261002/ENOLA.md:11,23`,
  `docs/refactoring/clean-code-20261002/ROUNDS.md:185`); none is a consumer.
- Ran `grep -rni lab-analyzer-recalibration` over the same extensions → no hits **at this
  document's writing pass**. (Re-run during the folder's documentation pass: the only hits are
  inside `docs/simulation-ideas/` — `../research-notes.md` §1, which records the pool entry,
  and this document — so the claim is pass-relative, not false; the target is in no shipped or
  catalog artifact.)
- Read and confirmed every code citation used above: `internal/perturb/perturb.go:1-7,43-45,125-164`;
  `internal/perturb/parameters.go:59-95`; `internal/perturb/transform_payload.go` (whole);
  `internal/mcp/world_commands.go:95-125`; `internal/world/reading_values.go:65-81`;
  `internal/world/dynamics.go:100-130`; `internal/world/observe.go:1-22`;
  `internal/world/state.go:40-70`; `internal/world/effectors.go:10-18`;
  `internal/truth/detector_math.go` (whole); `internal/truth/solver.go:95-124`;
  `internal/model/domain.go:135-165`; `internal/model/effector.go:55-85`;
  `internal/score/loop_metrics.go:20-46`; `internal/audit/scenario.go:40-70`;
  `internal/audit/audit.go:60-81`; `internal/audit/statistics.go:1-70`;
  `internal/audit/detector.go:72-89`; `internal/suite/commands.go:50-79`;
  `internal/suite/scenario.go:60-90`; `internal/suite/scenario_draw.go:15-45`;
  `internal/suite/command_log.go:10-27`; `internal/run/emission.go:55-75`;
  `internal/run/delivery.go:115-134`; `internal/run/probe_test.go:115-196`;
  `internal/world/oracle_test.go:22-76,218-267`; `docs/QUALITY_BAR.md:57-72`;
  `docs/contracts/domain-spec-v0.1.schema.json:108-133,385-430,505-548`;
  `docs/design/CONSUMERS.md:91-122`; `docs/design/DOMAIN_CATALOG.md:240-256,455-476`;
  `domains/rotating-machinery.domain.json:290-340`.
- **Not run here:** the scratch admission harness (removed during idea development; its
  numbers are reported, not reproduced), and the seed-sweep admission proof (no final domain
  JSON exists yet — it is deliverable (3)).
