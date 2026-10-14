# SI-05 · tank-dip-replay-aliveness

- **Status:** proposed candidate; quality bar **not met** this round (77/100 total — bar requires ≥85/100)
- **Primary consumer:** tamoz — at-least-once inbox dedup (live-stream duplicates graded directly; dump-pair collapse via late-admission completeness) and replayed-window handling, plus approval pressure its consumer must expose
- **Secondary consumer:** agentic-stream — event_inbox suppression, stale-state refusal after boot
- **Declared boundary with SI-01:** that domain owns live ambiguity (which live sensor is speaking); this one owns the replay mask (a window of the past substituted for the present under the same ids). The shared absence=heartbeat, reconnect-dump and ledger-as-oracle surfaces are deliberate pairing, per the catalog's decoration rule ([docs/design/DOMAIN_CATALOG.md:9](../../design/DOMAIN_CATALOG.md)). SI-01 itself was not verifiable in this repository at this document's writing pass (zero grep hits in tracked content **at that time** — the portfolio index, including [SI-01](SI-01-telemetry-estate.md), was created later the same day; see [../README.md](../README.md) §5, write order). Axis update from that sibling's own improvement round: SI-01 withdrew `absence=heartbeat` in favour of `absence=ambiguous` (SI-01 §6, item 3), so this document's "shared absence=heartbeat" phrasing reflects the pre-withdrawal state; the still-shared surfaces are the reconnect dump and the ledger-as-oracle.

## 1. Purpose

An unattended fuel-tank farm — 8–16 storage tanks on one shared fill manifold behind a single gateway — whose gateway, on reconnect, dumps the disconnect window's dip records at one arrival instant: original ids and event times, `observed_time` rewritten to the flush instant, alongside same-id duplicate rows.

The true fault, a slowly drifting dip probe, ramps inside that masked window, so consumers fail by:

1. treating recent arrival as freshness (dating the drift after the reconnect),
2. discarding the dump as beyond lateness allowance without flagging the onset it hides, or
3. failing to collapse same-id rows inside the dump — that last cell graded only indirectly: dump rows read `delayed` and `duplicateAdmissions` keys on reason `duplicated` alone ([internal/score/admission.go:47](../../../internal/score/admission.go)), so dump-dedup failure surfaces through late-admission completeness (admission.go:55-62), while the directly graded duplicate cell is the live stream's `duplicated` rows.

tamoz runs it for at-least-once inbox dedup and replayed-window handling — the two cells the instrument actually backs — plus approval pressure its consumer must expose: a `risk_class=high` regulatory filing under `consequence=regulatory`, with approval latency measured from consumer verdicts, not from domain data. The acked-but-not-applied-after-reconnect cell has **no instrument-side mechanism** (the ledger records only what the observer got, [internal/model/event.go:25-49](../../../internal/model/event.go); run has no consumer-ack channel), so this domain does not claim it. agentic-stream (event_inbox suppression, stale-state refusal after boot) is secondary.

## 2. How it runs on this simulator

Everything below is domain, perturbation and profile **data** plus committed perturbation kinds issued through the existing director command surface — one schema, zero new machinery, no per-domain code branches.

### Entities

12 tanks (declared range 8–16) on one shared fill manifold behind one gateway. Every channel — `dip_level`, `inlet_flow`/`outlet_flow` counters, `product_temperature`, per-tank probe heartbeat — rides that gateway, so during the masked window nothing is live, pre-unblind inference is impossible by construction, and scoring necessarily collapses to the post-unblind flush plus the resumed live stream.

### Signals

- **absence=heartbeat per tank**: the probe electronics keep polling while the sensing element lies, so absence gives no aliveness signal. (Property-axis coverage: only [domains/rotating-machinery.domain.json:13](../../../domains/rotating-machinery.domain.json) ships `absence: "heartbeat"` today; this domain raises it to 2.)
- Periodic scalars (dip level, product temperature).
- Monotonic counters (inlet/outlet flow).

### Dynamics and faults

`level_true` is an F1 integrator of inflow − outflow with closed form `x(t) = x0 + g·u·t` (oracle-tested at [internal/world/oracle_test.go:57-75](../../../internal/world/oracle_test.go)); inlet/outlet flows are F0 diurnal. Faults act on hidden states with closed-form envelopes ([internal/world/oracle_test.go:221-249](../../../internal/world/oracle_test.go)):

- **dip_probe_drift** — ramp on a hidden `dip_bias` state corrupting `dip_level` through observation bias; the shipped rotating-machinery `sensor_drift` pattern ([domains/rotating-machinery.domain.json:220-228](../../../domains/rotating-machinery.domain.json)). `expected_effector: raise_loss_report`; counterfactual: if_no_action "unreported inventory loss compounds" / if_action_by_deadline "inspection and recalibration".
- **probe_stuck** — step: pins the dip high and stable.
- **delivery_rezero** — `is_negative_class=true` ([internal/model/domain.go:217](../../../internal/model/domain.go)): a legitimate manifold delivery that momentarily shifts flows and mimics drift onset.

**No leak.** Physical inventory conserves exactly; the loss is a pure instrument lie, so a threshold rule cannot get the labels right.

### Perturbations — the replay regime

Committed kinds only ([internal/perturb/perturb.go:21-40](../../../internal/perturb/perturb.go)) issued as **command-log data** via `sim.perturb.apply` ([internal/mcp/director_tools.go:38](../../../internal/mcp/director_tools.go)): `id_reuse` takes only a rate ([internal/perturb/parameters.go:66](../../../internal/perturb/parameters.go)) and cannot target a window, so the regime is the director's command log — `duplicate_burst` ordered **before** `gross_backfill` spanning `[disconnect, reconnect)`, and `id_reuse` on the live post-reconnect stream as a second-order toxin.

Exact row classes, as the scorer joins them:

| Row class | Ledger reason | Grading path |
| --- | --- | --- |
| Dump rows (original seq and event time; `observed_time` = flush instant) | `delayed` — the flush overwrites any duplicated reason ([internal/perturb/flush.go:95-105](../../../internal/perturb/flush.go)) | Late-admission completeness ([internal/score/admission.go:55-62](../../../internal/score/admission.go)) |
| Same-id pairs inside the dump (two `delayed` rows of one seq) | `delayed` | Late-admission completeness (admission.go:55-62) — indirect, not duplicate admissions |
| Live pass-through duplicates | `duplicated` ([internal/perturb/transform_duplicate.go:7-15,44-50](../../../internal/perturb/transform_duplicate.go)) | `duplicateAdmissions` ([internal/score/admission.go:43-53](../../../internal/score/admission.go)) — direct |
| `id_reuse` collisions | same-seq, shifted-payload (`+100` / `"(reused)"`) `duplicated` rows ([internal/perturb/transform_duplicate.go:52-61](../../../internal/perturb/transform_duplicate.go)) | Duplicate admissions (admission.go:43-53) |

Timing is pinned: disconnect opens before drift onset; onset strictly inside the masked window; flush delivered at the advance boundary ([internal/run/advance.go:74-79](../../../internal/run/advance.go)); horizon leaves detection and recovery room after reconnect.

`store_and_forward` link delay is deliberately **not** used: it is a per-event heavy tail ([internal/world/link_delay.go:17-18](../../../internal/world/link_delay.go)) and its sampler is on the architecture review's G2 gap list — link-delay `lognormal`/`store_and_forward` "have no analytic oracle" ([docs/reviews/ARCHITECTURE_REVIEW.md:127-146](../../reviews/ARCHITECTURE_REVIEW.md), finding R2; R2 is a truth-coverage gap with a recommended oracle-or-scope-decision resolution, not a usage ban — this domain refuses the sampler because its oracle needs exact flush content, while SI-01 uses it as transport scenery with its own stated reason, [SI-01 §2.4](SI-01-telemetry-estate.md)); `gross_backfill` is a ledgered perturbation with exact scoring joins (admission.go:43-62).

### Ground truth and its analytic oracle

- **Detector:** `conservation_residual` (inputs `inlet_flow`, `outlet_flow`; outputs `dip_level`) with closed-form residual `−coef·d(env)/dt` during the bias ramp, computed independently clean-vs-faulted ([internal/truth/detector_math.go:31-37](../../../internal/truth/detector_math.go), [internal/truth/solver.go:117-123](../../../internal/truth/solver.go)) — the analytic cross-check as an oracle, not a consistency check.
- **Persistent phase:** the signal is peer divergence against healthy sibling tanks.
- **Masked window:** the ramp is recoverable only by computing the residual over the dump in event-time order — the freshness test itself.
- **The delivery ledger is the direct oracle.** Any verdict scores exactly right (onset dated inside the masked window from event times, dump deduped by seq, mask flagged), exactly wrong (dump treated as current state), or blocked-by-transport (flush short of quiescence — a transport miss, not a reasoning miss).
- **Benign routine-replay windows** are sealed separate scenarios with `is_negative_class` ([internal/model/domain.go:217](../../../internal/model/domain.go)) distinguishable in the label, never producing a nonzero residual label.

### Effectors

- **raise_loss_report** (`risk_class=high`) — the effect is a contrastive regulatory chain, not a bare counter: `loss_report_filed` +1 immediately, `regulatory_inspection` +1 after `dead_time_s`, `dip_bias` −1.0 on `time_constant_s` (inspection forces recalibration) — kick-scheduled with dead time and tau in the real world or the shadow world ([internal/world/effects.go:17-31](../../../internal/world/effects.go); shadow branch :56-64), the `schedule_maintenance` shape ([domains/rotating-machinery.domain.json:264-269](../../../domains/rotating-machinery.domain.json)). Its ack declares `silent_no_effect` and `confirmed_no_effect` with `dip_level` as confirmation channel (schema-required, [internal/schemas/domain-spec-v0.1.schema.json:549-570](../../../internal/schemas/domain-spec-v0.1.schema.json)), and the channel is **contrastive**: under the real effect the bias decays and `dip_level` converges to the conserved true level; under `silent_no_effect` the shadow reports the counterfactual convergence while the true `dip_bias` keeps ramping, so only the absent recovery betrays the ack (false-success graded at [internal/score/loop_metrics.go:35-44](../../../internal/score/loop_metrics.go)); under `confirmed_no_effect` `dip_level` tells the truth.
- **schedule_probe_recalibration** (`risk_class=medium`; direct remediation, restores `dip_bias`) — the non-regulatory path.

### Profiles

- **nominal**.
- **correlated_cascade** — shared-manifold overpressure stepping every tank's `inlet_flow` at once (`simultaneous_entities` — one declared physical driver, the only honest cascade a single-gateway farm supports; rotating-machinery supply-grid precedent, [domains/rotating-machinery.domain.json:341-344](../../../domains/rotating-machinery.domain.json)).
- **sensor_pathology** — `fault_weights` weighted to `dip_probe_drift` and `probe_stuck`, with `delivery_rezero` as control.

## 3. Feasibility against G1–G9

One schema, zero new machinery. All three oracle segments are already committed: the integrator closed form (oracle_test.go:57-75), fault-envelope closed forms (oracle_test.go:221-249), and the independently implemented conservation deviation (detector_math.go:31-37). The replay side has shipped conformance coverage — `TestGrossBackfillConformance` ([internal/perturb/conformance_test.go:53-78](../../../internal/perturb/conformance_test.go): withhold inside the gap, resume after, identity preserved at the flush) — and the contrastive effector chain rides committed kick scheduling with shadow-world support (effects.go:17-31, :56-64).

| Gate | Claim | Evidence |
| --- | --- | --- |
| G1 determinism | Covered: the replay regime is command-log data, which replay determinism already covers | §2 perturbation regime |
| G2 analytic cross-check | **Exact.** Named boundary tests: onset inside the masked window (world closed-form residual and solver conservation deviation agree to tolerance); every `is_negative_class` window scores a zero label; recovery — after `raise_loss_report` or `schedule_probe_recalibration`, `dip_bias` decays on the effect's tau and the residual returns to zero along the closed-form envelope, asserted against the solver on the recovery segment | oracle_test.go:57-75, 221-249; detector_math.go:31-37 |
| G3 reference consumer | Not addressed by the spec; the named consumers (tamoz, agentic-stream) run against the shipped operator surface and no new consumer machinery is introduced | §1 |
| G4 delivery ledger | The ledger is the oracle itself; the verdict table names the grading path per row class exactly as the scorer joins it (`delayed` → late-admission completeness, admission.go:55-62; `duplicated` → duplicate admissions, admission.go:43-53) | §2 row-class table |
| G5 quiescence | Addressed via timing only: flush delivered at the advance boundary; the blocked-by-transport cell defines flush-short-of-quiescence as a transport miss | advance.go:74-79 |
| G6 sealed oracle | Sealed: perturbations change nothing physical ([internal/perturb/perturb.go:1-7](../../../internal/perturb/perturb.go)); no oracle reads delivered values | perturb.go:1-7 |
| G7 injection probe | No surface grows: the domain declares no free-text channel, and `id_reuse`'s payload shift is a delivery-layer mutation outside the `attacker_controlled` gate (the spec makes this point inside its G6 sentence) | transform_duplicate.go:52-61 |
| G8 `silent_no_effect` | The `raise_loss_report` ack declares `silent_no_effect`/`confirmed_no_effect` with a schema-required confirmation channel; false-success grading produces the instrument's strongest metric | schema :549-570; loop_metrics.go:35-44 |
| G9 trivial baseline | A balanced tank whose dump arrives and changes nothing | §2 |

**Ship-blocking prerequisites, not polish** — until their tests exist the domain's central rows are fragile:

1. **Composition-order conformance test.** The same-id dump pairs exist only if `duplicate_burst` is applied before `gross_backfill`, an ordering the command log carries but no schema declaration enforces (`Process` iterates application order, [internal/perturb/perturb.go:141-154](../../../internal/perturb/perturb.go); both rows buffer at [internal/perturb/transform_timing.go:40-42](../../../internal/perturb/transform_timing.go); flush reason-overwrite at [internal/perturb/flush.go:103](../../../internal/perturb/flush.go)). The composition-order conformance test is the enforcement point, matching the repo's own enforcement-over-convention rule (ARCHITECTURE_REVIEW §2, theme 1), and must pin flush reasons with reorder windows and producer-flap buffers in both orders.
2. **End-flush guard test** (shared with SI-01/SI-04). No perturbation flush exists at run.End ([internal/run/finalize.go:66-88](../../../internal/run/finalize.go) — `finishTrace` closes sink and ledger and runs only the adapter postamble), so the test must assert every buffered row is ledgered (delivered or `dropped_by_perturbation`, perturb.go:119-124) and the scenario must pin reconnect before the horizon.

The suite requires every perturbation name covered at declared size ([internal/suite/generation.go:79-80](../../../internal/suite/generation.go)), so all profiles must stay audit-clean under arbitrary co-sampled perturbations, not only the pinned replay.

**Checks.** `go test -count=1 ./internal/perturb/ ./internal/truth/` re-run this ask: both `ok` (0.287s, 0.442s). Not run: the full repository gate (spec-only change, no code touched) and any suite fixtures — the domain does not exist yet, so feasibility claims are code-level, not run-level.

## 4. Quality bar result

| Dimension | Score |
| --- | --- |
| consumerValue | 7/10 |
| dataExpressibility | 8/10 |
| determinismOracle | 8/10 |
| ledgerQuiescence | 7/10 |
| novelty | 7/10 |
| sealedTruth | 8/10 |
| specificity | 9/10 |
| **Total** | **77/100** (54/70 scaled) |

**Verdict: not met.** The bar is a total of at least 85/100 across the seven dimensions with no single dimension below 7/10. Every dimension clears the per-dimension floor (minimum 7/10), but the scaled total 77/100 is below 85/100, so the idea misses the bar on the total alone.

## 5. Five-lens review

| Lens | Score | What it found | What changed |
| --- | --- | --- | --- |
| Determinism and analytic oracle | 6 | The core mechanism was not expressible with committed machinery as claimed: `id_reuse` takes only a rate (parameters.go:66) and duplicates the in-flight record with a shifted payload (transform_duplicate.go:35-61) — it cannot re-deliver yesterday's window; `gross_backfill` is a delay, not a replay of old content. The "one schema, zero new machinery" framing collapsed, and the spec's fallback claim was wrong. | The replay capability is now spec'd explicitly as committed kinds issued through the director command log (`duplicate_burst` before `gross_backfill` spanning the disconnect window); the G2 boundary case (leak/drift onset inside the masked window, negative-class windows scoring zero labels) was added. |
| Data-not-code expressibility | 8 | Same `id_reuse` finding; the heartbeat count arithmetic depended on SI-01 shipping. Strong otherwise: single-tank conservation detector is same-entity (solver.go:117-123), the fault is a constant state delta (schema :382-394), ledger reasons discriminate `duplicated` vs `delayed` (event.go:47). | Replay window sizing moved into profile data; ledger `Reason` values named as the wire discriminators in the domain description; `id_reuse` demoted to what it actually does — same-id/different-payload collisions as a second-order toxin; heartbeat arithmetic corrected to standalone +1. |
| Consumer testing value | 7 | The named perturbation trio was one part wrong, two parts right; no `silent_no_effect` anywhere (approval gating has no simulator-side machinery — only `risk_class` exists, schema :465), so the domain contributed nothing to the false-success metric; "replayed rows delivered as duplicated" was imprecise (flush arrives `delayed`, only duplicate-burst rows read `duplicated`). | Replay re-spec'd as `gross_backfill`-with-original-event-times + `duplicate_burst`; `silent_no_effect` added on `raise_loss_report` with the conservation residual as the free confirmation-channel substrate; "approval-gated" replaced by `risk_class=high` + `consequence=regulatory` with approval latency measured from consumer verdicts; the verdict table now names exact ledger reasons per row class. |
| Delivery ledger and quiescence | 7 | The replay semantics the oracle needs already exist and are verified (flush.go:95-105); `id_reuse` has no window targeting; same End-flush exposure as SI-01/SI-04; the shifted string value is a producer-side mutation outside the `attacker_controlled` gate but the spec had to state no detector may read delivered text. | `gross_backfill` (+`duplicate_burst`) pinned unconditionally with flush.go:95-105 cited as conformance evidence; End-time flush guard test added as a prerequisite; benign routine-reconnect negative class defined as a separate sealed window so `is_negative_class` is distinguishable in the ledger. |
| Novelty and portfolio balance | 8 | Four-surface overlap with SI-01 with no stated boundary; the masked-window fault was the portfolio's third conservation leak; the nearest shipped neighbor (cold-chain, the catalog's only reconnect dump) was uncited; the kill claim (metocean-buoy-partition) had zero grep hits. | SI-01/SI-05 boundary declared in purpose; leak replaced with a biased level sensor (the sensor_pathology profile) breaking the three-leak pattern while keeping the replay-mask question; `id_reuse` window targeting verified unavailable before relying on the mechanics; positioning vs cold-chain stated (there the dump is context to keep, here it is the mask); unverifiable kill claim withdrawn. |

## 6. Improvement history

The original consolidated spec and the final spec differ on the mechanism, the fault, the effector, and every verifiable number:

| Aspect | Original consolidated spec | Final spec |
| --- | --- | --- |
| Replay mechanism | `gross_backfill + id_reuse + duplicate_burst`, with `store_and_forward` link delay creating the gap; "verify id_reuse params support window targeting ... if not, gross_backfill + duplicate_burst already express it" | `id_reuse` verified rate-only (parameters.go:66); regime issued as command-log data via `sim.perturb.apply` with `duplicate_burst` ordered **before** `gross_backfill` spanning `[disconnect, reconnect)`; `id_reuse` demoted to live-stream second-order toxin; `store_and_forward` dropped (R2 no-analytic-oracle list) |
| True fault | Slow leak, constant siphon δ | Slowly drifting dip probe (ramp on hidden `dip_bias` via observation bias); no leak — inventory conserves exactly; `probe_stuck` + `delivery_rezero` controls added |
| Effector | `raise_loss_report`, advisory/approval-gated under `consequence=regulatory` | `risk_class=high` contrastive regulatory chain (`loss_report_filed` → delayed `regulatory_inspection` → `dip_bias` −1.0 on tau) with `silent_no_effect`/`confirmed_no_effect` ack and `dip_level` confirmation channel; `schedule_probe_recalibration` added |
| tamoz cells | Three, including "acked-but-not-applied ambiguity after reconnect" | Two the instrument backs (dedup, replayed-window) plus approval pressure; acked-but-not-applied explicitly **not** claimed (ledger records only what the observer got, event.go:25-49; no consumer-ack channel) |
| Dedup grading | "replayed-window rows delivered as duplicated" | Dump rows read `delayed` (flush reason-overwrite, flush.go:103): dump-pair collapse graded indirectly via late-admission completeness (admission.go:55-62); only live duplicates grade directly as `duplicated` (admission.go:43-53) |
| Novelty numbers | "raises absence=heartbeat to 3 and consequence=regulatory to 2; kills metocean-buoy-partition" | absence=heartbeat → 2 (only rotating-machinery ships it, domains/rotating-machinery.domain.json:13); consequence=regulatory → 2 (only cold-chain-transit); metocean kill claim withdrawn (zero grep hits) |
| Neighbors | Not cited | Cold-chain positioned (DOMAIN_CATALOG.md:180-188): there the dump is context the consumer must keep, here it is the mask the fault hides in; property-vector deltas declared per the decoration rule; SI-01 boundary declared in purpose |

**Round 1 — five-lens review (major revision).** The replay mechanism was re-based on committed machinery and the spec's incorrect fallback claim was removed; the fault swapped from a conservation leak to a lying sensor (breaking the portfolio's three-leak pattern and doubling as the mandatory `sensor_pathology` profile, which also makes the false-success test physical — the regulatory filing closes a real world loop); `silent_no_effect` was added so the domain produces false-success evidence; approval gating was replaced with declarable `risk_class`/`consequence` data; the negative class was sealed as separate `is_negative_class` scenarios; the End-flush guard and composition-order conformance tests were named ship-blocking prerequisites; the SI-01 boundary and cold-chain positioning were written into purpose.

**Round 2 — grep-verified correction pass.** Every count was re-derived with stated commands: domains/ contains exactly 3 `replay` lines (all descriptions); docs/ + documentation/ + domains/ together 174 lines across 51 files and repo-wide over .go/.md/.json 369 lines across 98 files (case-insensitive; re-run this ask, matching) — every one replay-determinism or run-artifact material, none a tank-farm or replay domain; the prior revision's error of generalizing the domains/-scoped count of 3 to "repo-wide" was corrected. (Scope note, folder documentation pass: the census describes the tree of this document's writing pass, which did not yet contain the simulation-ideas folder — created later the same day, that folder is itself replay-heavy, so the count is pass-relative, not a claim about the current tree; see [../README.md](../README.md) §5.) The property counts were fixed to what the repository supports (heartbeat 2 not 3 — a count of 3 required an unshipped sibling; regulatory 2), the unverifiable metocean kill claim was withdrawn, and the grading path per row class was pinned to the scorer's actual joins, which narrowed the tamoz claim to the cells the instrument backs.

## 7. Open questions

1. **Co-sampling vs the pinned regime.** The suite requires every perturbation name covered at declared size (generation.go:79-80) and all profiles audit-clean under arbitrary co-sampling, while the replay regime is command-log data with a required order. Does the composition-order conformance test pin flush reasons under co-sampled reorder windows and producer-flap buffers in **both** orders, and can the suite ever generate the reversed order in a way that breaks the dump-pair invariant in an admitted scenario?
2. **Discriminative power of indirect grading.** Dump-dedup failure surfaces through late-admission completeness (admission.go:55-62), not through a dedicated duplicate counter. Can a completeness failure be attributed specifically to uncollapsed same-id dump rows rather than to any other delayed-row mishandling, or does the verdict table need a distinguishing signal?
3. **Shared End-flush test with an unverifiable sibling.** The End-flush guard is shared with SI-01/SI-04, but SI-01 had zero grep hits in this repository at this document's writing pass. (Partly answered since: the sibling specs live at `../ideas/` — the portfolio index was created later the same day; SI-01's zero-grep claim was true of its own pass only, see [../README.md](../README.md) §5.) Remaining: does the shared test land with SI-05's fixtures or before them?
4. **Peer signal under correlated drift.** `sensor_pathology` weights faults toward `dip_probe_drift` across tanks; if several siblings drift in one scenario, the persistent-phase peer-divergence signal weakens and detection falls entirely to the conservation residual over the dump. Is that degradation acceptable in the profile's declared verdicts, and does `correlated_cascade`'s simultaneous inlet step mask or mimic drift onset inside the same window?
5. **Recovery-segment oracle coverage.** The named G2 boundary tests include recovery along the effect's tau (`dip_bias` −1.0 on `time_constant_s`). The committed oracle tests cover fault onset envelopes (oracle_test.go:221-249); whether an existing test covers the effect-kick decay envelope on the recovery segment, or that assertion is new work in the recovery boundary test, is unverified.
6. **Horizon fit.** The scenario must fit disconnect → masked onset → reconnect flush → detection → regulatory remediation (with `dead_time_s` and tau) inside one horizon while leaving recovery room after reconnect — a run-level property that cannot be confirmed until the domain fixture exists.
