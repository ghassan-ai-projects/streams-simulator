# SI-04 · `ami-headend-backpressure`

| | |
|---|---|
| **ID / slug** | SI-04 / `ami-headend-backpressure` |
| **Status** | Candidate — **quality bar NOT met** (73/100 against the 85/100 bar) |
| **Serves** | `agentic-stream` (watermark stall, `trace_end` until-horizon trailer) and `tamoz` (misfire/overlap under delayed delivery against a deterministic world clock) |
| **Reviewed** | 2026-10-05 · five-lens review applied · two improvement rounds consolidated |

## Purpose

An AMI electric-meter fleet whose head-end poll stalls for hours, then releases: interval register reads accumulate unpolled and arrive compressed into one exact arrival-time plateau — every buffered record shares one `ObservedTime` (`internal/perturb/perturb_test.go:112-117`) — while event time stays true. Wire-identical to a stalled pull, but modeled producer-side (withhold/flush), so the domain claims no pull-side backpressure semantics. It owns exactly one question: watermark freeze-then-jump scored against settlement integrals — a consumer that bills on arrival time mis-assigns all withheld energy to the single release period, and the watermark's freeze span equals the withhold window before the plateau jump. The ledger separates a plateau-interior transport miss from the tamper's reasoning miss. Distinct from shipped cold-chain-transit, whose `store_and_forward` plateau is transport-side and spread per record, not producer-side and one-instant. Run by agentic-stream (`internal/run/delivery.go:53-77`) and tamoz.

## How it runs on this simulator

The domain is data. Everything below is declared in domain JSON against the one `domain-spec-v0.1` schema; no per-domain code branches.

### Entities

- 200–500 meters (fleet cardinality duplicates shipped coverage and is not claimed as new).

### Signals

- Hourly interval register counters (`value_type=counter`) plus power scalars, periodic cadence.
- **Rollover parameterization, declared in domain data and machine-checked.** The runtime renders counters as `int64(round(raw))` with no width modeling (`internal/world/reading_values.go:36-37`), so the domain declares avg demand 10 kW over a 7-day horizon — closed-form max register ≈ 1.7e3 kWh, far under 2^62. A suite-admission invariant test computes that closed-form bound from the declared `max_rate_kw`/`horizon_s` and rejects any scenario crossing it (JSON Schema cannot sum; the bounds are data, the check is code-in-test).

### Dynamics

- All F0: integrator consumption with closed form `x(t) = x0 + g·u·t` (`internal/world/oracle_test.go:57-58`).
- Three fault families on one register substrate, all schema-expressible — multiplicative deltas (`internal/schemas/domain-spec-v0.1.schema.json:388-394`, applied in `internal/world/dynamics.go:78-82`) and `duty_cycle` onset (`internal/schemas/domain-spec-v0.1.schema.json:379`):
  - `underregister_linear` — step onset, mult −δ: register grows at (1−δ)·rate.
  - `underregister_intermittent` — intermittent onset, duty_cycle d, mult −δ. The envelope is a deterministic square wave (period P = 3600/rate_per_hour, on for d·P, default d = 0.5 — `internal/world/dynamics.go:109-120`, period arithmetic `:135-141`; no substream draw).
  - `register_bypass` — step onset, mult −1: register freezes flat while consumption continues.
- Detector declared as `single_channel_snr` on the register **level**, not slope — the form compares clean-vs-faulted reading levels against the noise floor (`internal/truth/detector_math.go:10-14`).
- Per-family `first_observable`, all closed-form on the cumulative shortfall S(t) = |faulted − clean| register level:
  - linear: S(t) = δ·rate·(t − t0); bypass: S(t) = rate·(t − t0);
  - intermittent: S(t) = δ·rate·D(t), where D(t) = d·P·floor(e/P) + min(e mod P, d·P) in exact floor arithmetic over the deterministic envelope;
  - `first_observable` = smallest t with S(t) ≥ SNR·σ; a truth-solver test asserts the closed form agrees with a scan over `envelopeAt` within one duty period P.

### Perturbations

- `producer_flap` / `gross_backfill` withhold-then-release realizes the plateau: `withholdBackfill` buffers during the window (`internal/perturb/transform_timing.go:33-46`); flush stamps the whole buffer with one `ObservedTime` (`internal/perturb/flush.go:83-105`); release happens at the advance boundary (`internal/run/advance.go:74-79`).
- Cold-chain's `store_and_forward` spreads per record (`internal/world/link_delay.go:17-18`), so the exact one-instant plateau stays native to this domain's withhold/flush.
- `duplicate_burst` litters the release burst.

### Ground truth and its analytic oracle

- **Scope, explicit:** a `GroundTruthRecord` labels one entity per scenario (`BuildRecord` takes a single entityID, `internal/truth/truth.go:16-27`; `sc.EntityID` at `internal/suite/scenario.go:32`). The optional settlement block seals the **labeled meter only**: 168 hourly integrals E_p = R(t_{p+1}) − R(t_p) for that meter, ~10 kWh each. The rest of the fleet delivers normally as peer contrast and is neither sealed nor settlement-scored.
- **Mis-settlement oracle:** the consumer's `file_regulatory_report` args carry per-period figures for the labeled meter. Billed on arrival time, the period k containing the release instant T_r absorbs the withheld sum (~140 kWh for a 14 h withhold), withheld periods settle ≈ 0, and all other periods are exact within 3·σ√2 ≈ 2.1 kWh at σ = 0.5 kWh. Period windows are half-open: T_r lands in exactly one period, so a plateau straddling a boundary corrupts all-or-nothing per period.
- **No-file semantics:** no filed report on a positive scenario is a settlement **miss** (counted as a mechanism miss, mirroring the contract's missing-episode rule and `internal/score/loop.go:14`'s `Resolved=false` path for missing expected-effector calls); a filed report on the negative class is a false positive. Never unscored.

### Effectors

- `file_regulatory_report` (regulatory consequence): carries the per-period settlement figures; declares the standard failure-mode distribution including `confirmed_no_effect`/`silent_no_effect` (`internal/schemas/domain-spec-v0.1.schema.json:481-492`), scored separately per C-05. A no-effect invocation means the figures never arrive — scored as a settlement miss, with the ack-lies mode attributed by the confirmed/silent distinction.
- `request_field_visit` (advisory).

## Feasibility against G1–G9

**Honest capability accounting.** Two machinery deltas are **design intent, not committed**: (1) an optional settlement block in the ground-truth contract, and (2) a generic reported-vs-sealed per-period comparison in `internal/score` — verified absent: `internal/score/judgment.go:12-39` and `internal/score/offline.go:12-23` contain only verdict-vs-label detection grading plus ledger/loop/consumer instruments (re-verified this session; a grep for "settlement" under `internal/` and `docs/contracts/` returns no hits). The domain is inadmissible until both land with tests.

**Contract discipline.** `ground-truth-v0.1` gains ONE additive optional settlement property — in-kind with the six optional fields v0.1 already carries (`deadline_ns`, `counterfactual`, `expected_effector`, `pre_degraded`, `perturbations`, `world_state_reference`, `internal/model/truth.go:18-25`) — no v0.2 bump. `additionalProperties:false` (schema line 7) means the in-repo validator updates in the same change, and there is no external consumer of the sealed contract.

**Digest implications.** Sealed-truth bytes change only for domains declaring settlement; the run determinism tuple `(sim_version, domain_digest, adapter_digest, seed, command_log, sink)` is untouched because the block is a pure function of `(seed, domain data)`, and replay reproduces identical bytes. The scorer stays domain-agnostic: it compares figures the sealed record itself carries against the tolerance the record declares.

Everything else is committed: withhold/flush, advance-boundary flush, multiplicative deltas, `duty_cycle`, deterministic intermittent envelope, ledger rows with `Delivered` plus reason (`internal/model/event.go:38-49`).

| Gate | SI-04 coverage |
|---|---|
| **G1 Determinism** | Digest analysis above: the settlement block is pure in `(seed, domain data)`; the tuple is untouched; replay reproduces identical bytes. |
| **G2 Analytic cross-check** | Boundary oracle as closed-form sum vs integrator across the flush edge (partial periods at the release boundaries); the per-family `first_observable` agreement test (closed form vs `envelopeAt` scan within one duty period); a dedicated exactly-once flush invariant at the advance boundary — an off-by-one there scores as a consumer miss. |
| **G3 Reference consumer** | Not addressed in the feasibility note. |
| **G4 Delivery ledger** | Release boundary strictly interior to the run horizon as a scenario-admission rule, because `End` performs no flush (`internal/run/finalize.go:30-43`) and no run-level test exercises backfill (grep over `internal/run/*_test.go` for `gross_backfill`/`producer_flap`: zero hits — re-run this session). Plus a >10k conservation run with `gross_backfill`+`duplicate_burst` at the release boundary (today's only >10k test is drop+duplicate, `internal/run/ledger_test.go:119-174`), and the attribution invariant: a plateau-interior miss is ledger-attributable, the tamper is label-attributable, never the reverse. |
| **G5 Quiescence** | Evidence writable today with the injectable fake `QuiescenceClock` (`internal/run/quiescence_test.go:20-31`): quiesced-through past the trace end reported only after draining the plateau; drained state named exactly — every buffered DeliveryID terminal-delivered, reason `delayed`, one shared `ObservedTime`. Caveat carried in the G5 evidence plan: the await backstop is wall-clock (`DefaultQuiescenceTimeout`, `internal/run/quiescence.go:31`; `realQuiescenceClock`, `internal/run/run.go:53-66`), so its timeout failure path is not simulated-time-deterministic. |
| **G6 Sealed oracle** | The settlement contract extension discipline above: one additive optional property, sealed bytes reproducible from `(seed, domain data)`, no external consumer. |
| **G7 Injection probe** | Not addressed in the feasibility note. |
| **G8 `silent_no_effect`** | The report effector declares the standard failure-mode distribution including `confirmed_no_effect`/`silent_no_effect` (schema `:481-492`), scored separately per C-05; a no-effect invocation is a settlement miss with the ack-lies mode attributed by the confirmed/silent distinction. |
| **G9 Trivial baseline** | A constant-register meter. Suite floor cleared by three fault families plus negatives (min 10 per label, 35–45% negative class — `docs/design/GROUND_TRUTH_AND_SCORING.md:291-292`, enforced at `internal/suite/admission.go:40-43`). |

**Hardest parts:** the settlement contract extension, the score comparison, and the flush-edge off-by-one.

## Quality bar result

| Dimension | Score |
|---|---|
| consumerValue | 7/10 |
| dataExpressibility | 7/10 |
| determinismOracle | 7/10 |
| ledgerQuiescence | 7/10 |
| novelty | 7/10 |
| sealedTruth | 7/10 |
| specificity | 9/10 |
| **Total** | **73/100** (raw sum 51/70 across the seven 0–10 dimensions, normalised to the 100-point scale) |

**Bar: NOT met.** The bar is a total of at least 85/100 across the seven dimensions with no single dimension below 7/10. The per-dimension floor passes at its exact minimum (7/10 on six dimensions, 9/10 on specificity), but 73/100 is 12 points short of the 85/100 total.

## Five-lens review

### 1. Determinism and analytic oracle — 9/10

Found:

- Counter rollover is unmodelled and the original mitigation (parameterize horizon/widths) was prose of the same kind shipped limitations carry; the sealed settlement integral must never cross a rollover boundary or the closed form diverges from the register — a data-validation rule, not documentation.
- All cited machinery verified live: `withholdBackfill` (`transform_timing.go:33-46`), flush at the advance boundary (`advance.go:74-79`), monotonic quiescence watermark (`internal/run/quiescence.go:14-22`), multiplicative tamper deltas; the G2 integral oracle covered by the integrator closed form (`oracle_test.go:57-58`).
- Residual risk: a release burst must flush exactly once at the advance boundary — an off-by-one is scored as a consumer miss and needed a dedicated invariant test.

Changed in response:

- Added the boundary oracle for partial metering periods at the release edges (closed-form sum vs integrator across the flush boundary) and the attribution invariant — plateau-interior miss ledger-attributable via `internal/model/event.go:38-49` rows, tamper label-attributable, never the reverse.
- Rollover made machine-checked: the suite-admission invariant computes the closed-form register bound from declared `max_rate_kw`/`horizon_s` and rejects crossing scenarios.

### 2. Data-not-code expressibility — 9/10

Found:

- The AMI story is a head-end not polling; producer-side withhold is wire-identical, but the domain text must not claim pull-side backpressure semantics the perturbation layer does not have.
- Detector phrasing was wrong: the declared form compares clean-vs-faulted reading **levels** (`detector_math.go:10-14`), not slope; for a linear tamper the first threshold crossing is equivalent, but the domain must declare level.
- Everything else verified committed: producer_flap withhold/flush, `store_and_forward` in channel data, counter rendering, multiplicative deltas, ledger `Delivered`+reason.

Changed in response:

- Purpose now states the producer-side model and explicitly disclaims pull-side semantics.
- Detector redeclared as `single_channel_snr` on the register level; the linear crossing time does the rest.
- Register width and run horizon declared in domain data with the no-rollover admission check (shared with lens 1's fix).

### 3. Consumer testing value — 7/10

Found:

- One fault family rode the whole domain's judgment value; a hindsight-fitted slope threshold was at trivial risk unless frozen-interval bounding genuinely degrades it — asserted, not shown.
- The "same limitation host-system-health ships" comparison was imprecise: shipped host-system-health declares no `counter_rollover` fault; the rollover mention is catalog-plan only.
- The headline load shape is transport, not world; it must not carry fault-family weight.

Changed in response:

- Two more fault shapes declared on the same register substrate — `underregister_intermittent` (duty_cycle, schema-supported) and `register_bypass` — giving the graded suite a non-trivial fault taxonomy.
- Quiescence-after-release made a check, not prose: the drained ledger state named exactly (every buffered DeliveryID terminal-delivered, reason `delayed`, one shared `ObservedTime`) and cited as the G5 evidence.
- Register-width/horizon bounds moved into suite admission so no generated scenario can cross rollover.

### 4. Delivery ledger and quiescence — 6/10

Found:

- The one real G4 hole: withheld records become terminal only when an advance boundary crosses `UntilNS`; `run.End` performs no flush (`finalize.go:30-43`) and no run-level test exercises backfill (grep: zero hits). A plateau whose release window ends at the scenario horizon — the design's own shape — risked rows vanishing from the ledger, corrupting the exact transport-miss vs reasoning-miss separation the domain scores.
- The G5 headline is achievable and deterministic (monotonic watermark `quiescence.go:14-22`, injectable clock, trailer horizon computed from delivered arrivals `delivery.go:53-77`), but no deterministic test existed at burst shape.
- "Conservation beyond 10k records is stressed by the burst" was asserted, not evidenced — the only >10k test is the periodic-heartbeat recipe with drop+duplicate (`ledger_test.go:119-174`).

Changed in response:

- "Release boundary strictly interior to the run horizon" adopted as a scenario-admission rule.
- A >10k conservation test with `gross_backfill`+`duplicate_burst` at the release boundary committed as the domain's G4 evidence.
- A deterministic burst-shape quiescence test using the fake `QuiescenceClock` written into the G5 evidence plan (writable today).

### 5. Novelty and portfolio balance — 4/10

Found:

- The headline novelty claim was contradicted by the shipped domain it named: cold-chain-transit ships the plateau — `store_and_forward` with `max_s=50400` (`domains/cold-chain-transit.domain.json:123-131`), "dumps 840 buffered readings at one arrival instant" (`docs/design/DOMAIN_CATALOG.md:186-192`).
- Residual novelty was thinner than framed; each piece real, none a new world property.
- Portfolio redundancy: fifth ledger-as-transport-vs-reasoning-oracle candidate, third reconnect-release narrative; the sibling set is not verifiable in-repo (no SI catalog exists under `docs/`).
- Fleet cardinality and counter substrates duplicate shipped coverage; the grain-storage-backfill kill claim was unverifiable and mis-cited.

Changed in response:

- Novelty re-anchored and led with the genuine deltas: producer-side vs transport-side withholding ("who withheld the data" is ledger-attributable — exact one-instant plateau vs per-record exponential spread), watermark freeze-then-jump as the scored quantity, and settlement-integral scoring against sealed per-period truth, which no shipped domain does.
- Both "first plateau / no shipped backpressure" claims dropped; the distinct-from-cold-chain statement made precise.
- The fold-in contingency made binding: if a sibling owns watermark freeze-then-jump, SI-04 reduces to the settlement-integral delta alone and its withhold/plateau machinery folds into the sibling's scenario.
- Removed as unverifiable in this repo: "F-002 stalled-progress reproduction" (zero hits repo-wide at this document's writing pass — the brief IDs now also appear in [../research-notes.md](../research-notes.md) §3, created later the same day; a re-run during the folder's documentation pass confirms zero hits outside `docs/simulation-ideas/`) and the grain-storage-backfill kill claim; fleet cardinality and counter substrates disclaimed. The lateness=gross 2→3 claim retained (verified: `grep -l '"gross"' domains/*.domain.json` → `cold-chain-transit`, `open-field-irrigation`; `DOMAIN_CATALOG.md:61`).

## Improvement history

Method note: the pipeline states two improvement rounds ran between the original consolidated spec and the final spec; the material supplied here comprises those two specs plus the five-lens findings, with no per-round logs. The round split below is therefore reconstructed: round 1 = the corrections the five lenses demanded; round 2 = the verification and consolidation hardening visible only in the final spec.

### Round 1 — five-lens corrections (structural)

- **Novelty rebuilt.** Original: "First world where the simulator deliberately withholds load and releases it: no shipped domain exercises backpressure arrival shape or post-release quiescence." False as stated — cold-chain ships the plateau. Final: deltas led with producer-vs-transport withholding, watermark freeze-then-jump, and settlement scoring; distinct-from-cold-chain statement and the binding fold-in contingency added.
- **Fault content tripled.** Original: one linear under-registration tamper delta. Final: three families (linear, intermittent duty-cycle, bypass) with per-family closed-form `first_observable` on the cumulative shortfall S(t).
- **Detector corrected:** `single_channel_snr` on register slope → on the register level.
- **Plateau mechanism simplified.** Original realized the plateau with "`store_and_forward` link_delay in domain data plus producer_flap withhold/flush." Final drops `store_and_forward` from the domain's own design entirely — withhold/flush natively produce the exact one-instant plateau, and `store_and_forward` survives only as the cold-chain contrast.
- **Rollover moved from prose caveat** ("declared no-rollover caveat") **to a machine-checked suite-admission invariant.**
- **Ledger hole closed:** release-strictly-interior admission rule; the drained-ledger-state quiescence check named; the >10k withhold/release conservation test committed as G4 evidence.

### Round 2 — verification and consolidation hardening

- **"One schema, zero new machinery" corrected.** The original feasibility claimed zero new machinery; the final declares the two uncommitted pieces — the settlement ground-truth block and the score per-period comparison — as design intent, verified absent (`judgment.go:12-39`, `offline.go:12-23`), with the domain inadmissible until both land with tests.
- **Contract and digest discipline written out:** one additive optional property in-kind with the six existing optional fields (`internal/model/truth.go:18-25`), no v0.2 bump, `additionalProperties:false` (schema line 7) so the in-repo validator updates in the same change, no external consumer; the determinism tuple untouched because the block is pure in `(seed, domain data)`.
- **Ground-truth scope tightened and quantified:** labeled meter only (`truth.go:16-27`, `scenario.go:32`); 168 hourly integrals E_p; mis-settlement numbers (~140 kWh per 14 h withhold; 3·σ√2 ≈ 2.1 kWh tolerance at σ = 0.5 kWh); half-open period windows; no-file semantics pinned to never-unscored (mechanism miss on positive, false positive on negative).
- **G8 wired:** no-effect invocations scored as settlement misses with confirmed/silent attribution per C-05 (schema `:481-492`).
- **G5 caveat recorded:** the await backstop is wall-clock, so its timeout failure path is not simulated-time-deterministic.
- **Purpose narrowed** to exactly one owned question (watermark freeze-then-jump scored against settlement integrals), per the novelty lens's "own exactly one question" requirement; cross-repo unverifiables (F-002, grain-storage kill claim) deleted.

The review-time lens scores and the final quality-bar dimensions are different instruments, but the direction is visible: novelty moved from the novelty lens's 4/10 to a final 7/10 after re-anchoring — not enough to lift the total over the bar.

## Open questions

1. **Admission blockers.** Both design-intent deltas must land with tests before SI-04 is admissible: the optional settlement block in the ground-truth-v0.1 contract (and `internal/model`), and the generic reported-vs-sealed per-period comparison in `internal/score`. In what round do they land, and does the block mirror the existing six-field optionality pattern exactly?
2. **Fold-in arbitration.** The sibling set (the earlier ledger-as-oracle candidates and the SI-01/SI-05 reconnect narratives) was not verifiable in-repo when this document was written — no SI catalog existed under `docs/` at its writing pass. (Superseded in part, folder documentation pass: the portfolio index now lives at [docs/simulation-ideas/](../README.md), created later the same day; the sibling documents are `../ideas/SI-01-…` through `SI-10-…`.) If a sibling owns watermark freeze-then-jump, SI-04 reduces to the settlement-integral delta and folds in. Which pipeline artifact arbitrates sibling ownership?
3. **G5 backstop.** The await deadline is wall-clock (`DefaultQuiescenceTimeout`, `internal/run/run.go:48-51`; `realQuiescenceClock`), so its timeout failure path is not simulated-time-deterministic. Is that acceptable for the G5 evidence plan, or does the await deadline need injectable-time treatment end to end?
4. **End-time flush as defense in depth.** The ledger lens offered two closes for the End-flush hole; the admission rule was chosen. Should `run.End` also gain flush-as-terminal rows so that no admissible scenario can drop ledger rows, or is the admission rule the whole fix?
5. **Hindsight-fit demonstration.** That frozen-interval bounding genuinely degrades a hindsight-fitted register-level threshold is asserted, not shown. Which demonstration scenario proves it before the suite freezes?

## Citation verification (this session, 2026-10-05, branch `new-sim-ideas`)

All line citations above were re-read at their cited locations in the worktree; ranges were corrected where this session's read differed from the research material (`advance.go:74-79`, `finalize.go:30-43`, `dynamics.go:78-82` / `:109-120` / `:135-141`, `quiescence_test.go:20-31`, `delivery.go:53-77`, `run.go:48-66`). Commands run:

- `grep -rn 'gross_backfill\|producer_flap' internal/run/*_test.go` → no output (confirms no run-level backfill/flap test exists).
- `grep -l '"gross"' domains/*.domain.json` → `cold-chain-transit.domain.json`, `open-field-irrigation.domain.json` (lateness=gross ships in 2 domains; SI-04 raises it to 3; also `docs/design/DOMAIN_CATALOG.md:61`).
- `grep -rni 'settlement' internal/ docs/contracts/` → no output (settlement machinery absent, as the feasibility note claims).
- `grep -rn 'F-002' --include='*.go' --include='*.md' .` → no output (claim removed as unverifiable, as the novelty section states).
- G1–G9 gate definitions read at `docs/QUALITY_BAR.md:64-72`.

No Go tests were run for this ask; it is a documentation task and names no test command.
