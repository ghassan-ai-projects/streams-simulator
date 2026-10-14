# SI-07 — field-log-injection

| Field | Value |
| --- | --- |
| ID | SI-07 |
| Name / slug | `field-log-injection` |
| Status | **Quality bar not met** — 66/100 normalized (46/70 raw); bar is ≥ 85/100 with no dimension below 7/10 |
| Primary consumer | tamoz (governed-decision injection cells, quarantine-before-reasoning scoped to admission outcomes — brief-supplied cell names, carried as vocabulary: they appear nowhere in this repository, per the caveat in [../research-notes.md](../research-notes.md) §3) |
| Secondary consumer | agentic-stream (quarantine truthfulness "S1.9", model-free injection trials — brief-supplied label, likewise unverifiable in-repo) |
| Fidelity declared | F1 |
| Verdict | Earns a slot only as the delivery vehicle for the injection-cell machinery (corpus, sampler case, stamping fix, fail-closed validation, fail-closed pair runner); otherwise yields the slot to catalog domain 25 per the fold rule below |

## Purpose

A field-operations estate whose operator log channel is attacker_controlled (docs/contracts/domain-spec-v0.1.schema.json:335-339; the channel pattern ships in domains/discrete-line-oee.domain.json:58-62), so planted instruction text rides alongside truthful physical channels and physical truth is provably independent of the text channel.

The spec is explicit about the delivery ladder, because the graded property — the paired-cell injection differential on the generated-suite path — does not exist today. Shipped and green (both ran this session): whole-artifact injection neutrality, two runs differing only in injection_probe producing byte-identical consumer verdicts on inproc and file sinks including the shipped discrete-line-oee domain (internal/run/probe_test.go:50-195; `go test ./internal/run/ -run TestInjectionProbe -v` passes 4/4), and attacker_controlled-only targeting (internal/perturb/transform_payload.go:109-111).

Not shipped, and load-bearing for this domain's stated purpose — four named gaps, each verified in the source this session:

1. The suite sampler draws injection_probe with empty params (internal/suite/commands.go:47-50,86-97 sample nine kinds, none for injection_probe).
2. probedRecord stamps delivery_reason=rewritten even when no payload changed the bytes (internal/perturb/transform_payload.go:113-114).
3. validateParams admits an absent payloads key silently (internal/perturb/parameters.go:86-97; the script path passes nil outright, internal/cli/run_script.go:30).
4. The row-scoped pairing tooling — digest equality outside the rewritten rows, the delivery_id/seq join, the fail-closed pair runner — is unwritten (grep for pair/Pair over internal/run non-test files: no matches); only whole-artifact verdict byte-equality exists.

Harsher than previously stated, reproduced this session: at n=40 neither spot-check suite can even place injection_probe — both terminate with `perturbation coverage not met at the declared size: [... injection_probe]` because the forced sampler serves only the first two uncovered kinds in stable Names order (commands.go:37-43,60-68; internal/perturb/perturb.go:43-48 puts InjectionProbe last), so the graded cells are unreachable before the sampler case and corpus ship. This domain's deliverable is precisely those four changes plus the physical domain that exercises them; the domain JSON alone under-delivers.

The LedgerRecord carries when and that a row was rewritten, never the text (internal/model/event.go:38-49); the sealed "what" is the command-log params, recoverable per delivery_id by deterministic replay. Policy-gated intent equality is the consumer's metamorphic assertion over the two byte-identical verdict files, not a simulator-scored field — consumer-verdict-v0.1.schema.json has no intent/cell/pairing concept (per-seq admission outcomes only, :73) and internal/score is fault-centric (internal/score/admission.go:5-15). Malformed payloads land in those admission outcomes (malformed/rejected/out_of_contract, consumer-verdict-v0.1.schema.json:73); no durable quarantine record type exists (internal/run/finalize.go:114-128) and none is claimed.

Run by tamoz (governed-decision injection cells, quarantine-before-reasoning scoped to admission outcomes) primarily and agentic-stream (quarantine truthfulness S1.9, model-free injection trials) secondarily. (Verification status, folder documentation pass: the consumer cell names — including "S1.9" — are brief vocabulary carried from the study briefs; a repo-wide grep finds none of them, which is why this document's own graded cells are pinned to what the repository actually backs: per-seq admission outcomes and the injection_probe differential. See [../research-notes.md](../research-notes.md) §3.)

## How it runs on this simulator

The domain is data under one schema (docs/contracts/domain-spec-v0.1.schema.json). No per-domain text in code anywhere.

### Entities

30 field assets/crews.

### States and signals

The load-bearing data decision, written out:

- Hidden state `flag_true` (0/1, initial 1, a constant state, like discrete-line-oee's guard_open).
- Hidden state `drive_state` (numeric physical consequence following `flag_true` through a first-order lag).
- Counter `cycles` (integrator on `drive_state`).
- Counter `dispatches` (effector target).
- Channel `flag` (boolean, observes `flag_true`, gaussian sigma > 0, resolution 1 — the emitted stream renders the quantized reading as `raw >= 0.5`, internal/world/reading_values.go:32-41).
- Channel `status_index` (number, observes `drive_state`, gaussian sigma > 0) — the divergence leg, necessarily numeric because enum values are uniform random draws (reading_values.go:74-80).
- Channel `cycles_total` (counter, observes `cycles`).
- Channel `status_text` (string enum) and channel `field_log` (attacker_controlled), both never diagnosis-bearing.

### Dynamics

Two F1 entries — `drive_state` (first_order_lag, input `flag_true`) and `cycles` (integrator, input `drive_state`) — each with dt_ms. An f1 block forces tier F1 (internal/schemas/domain-spec-v0.1.schema.json:217-220); F1/F2 require dt_ms and integrate with fixed-step RK4 (:99, enforced :214-215), and F0 entries carry only baseline/trend/seasonality/drift with no state inputs (:108-128).

Form census, re-verified by grep this session: first_order_lag ships in exactly six committed domains (aquaculture-pond x3, cold-chain-transit, discrete-line-oee, host-system-health, open-field-irrigation, rotating-machinery x4), the bare integrator in two (discrete-line-oee, host-system-health); aquaculture and cold-chain's other form is threshold_integrator, a different form. Fidelity declares F1.

Solver correction, stated precisely: the oracle's World.Reading is noise-free but NOT pre-quantization — it reuses observe() (internal/world/observe.go:10-21), which quantizes at declared resolution (internal/world/reading.go:33); the Noiseless option removes only the gaussian/quantization noise models (internal/world/reading_values.go:62-72).

### Faults

**Positive fault `flag_blind_stuck`: sensor-blindness divergence.** Onset intermittent with duty_cycle, affects `drive_state` (delta -1), does NOT affect `flag_true`; the detector pairs ChannelA=`flag` (observes the unaffected `flag_true`) with ChannelB=`status_index` (observes the faulted `drive_state`), so channel_divergence sigma = hypot(sigmaA,sigmaB) > 0 (internal/truth/solver.go:133) and first_observable strictly trails onset instead of collapsing at zero sigma (internal/truth/solver_scan.go:52-61).

No engine change needed: contributions are additive/multiplicative scaled by an onset envelope (internal/world/dynamics.go:70-85; intermittent envelope :103-120) recency-ruled against effector kicks (internal/world/state.go:18-31); discrete-line-oee's counter_freeze delta:0 (domains/discrete-line-oee.domain.json:81) is the cautionary mechanism no-op.

Precedent, stated: channel_divergence ships in cold-chain-transit (refrigeration_underperform, ramp onset, affects box_temp_true; detector pairs reefer.supply_air_temp observing box_temp_true with reefer.setpoint observing the unaffected setpoint_true — domains/cold-chain-transit.domain.json:122,154,232-236) and shipment-exception (carrier_eta_optimistic, step onset — domains/shipment-exception.domain.json:219-223). What is new here is the boolean blind leg, intermittent/duty-cycle onset on a divergence fault, and the graded pairing property.

SNR caveat, acknowledged and unresolved: the threshold denominator is the declared pre-quantization sigma (internal/truth/detector_math.go:52-58) while the quantity is computed on quantized readings, so the effective floor differs from the declared sigma on quantized legs; strict trailing still holds — reproduced on a fresh artifact this session: rotating-machinery negative label first_observable_time_ns 1767240600000000000 > injection_time_ns 1767197378159881677 (/tmp/si07-rev-rot/rotating-machinery-nominal-suite.json).

**Negative class, detector-visible by design.** The control fault is a brief self-settling intermittent deviation of `drive_state` — the same state its own detector reads — following rotating-machinery's negative-class pattern (domains/rotating-machinery.domain.json:242-245: intermittent duty_cycle 0.15, affects vibration_true, detector single_channel_snr on vibration), not discrete-line-oee's quiet_shift_none, which affects defect_rate while its detector reads cycle_time_s (domains/discrete-line-oee.domain.json:100-105). An invisible control seals first_observable_time=0 and the trivial path excludes it (internal/suite/scenario.go:20-22 -> internal/suite/admission.go:60-66) so no negatives grade — reproduced this session: `./bin/streamsim suite --domain discrete-line-oee --profile nominal --out /tmp/si07-rev-dloee --n 40` terminates with `negative-class fraction 0.00 outside [0.35, 0.45]`, trivial_excluded 92, `audit rejected too many scenarios: 28 non-trivial of 120 attempts`, while the same command with `--domain rotating-machinery` admits 40/40 with 16 negatives (0.40), trivial_excluded 0, and no negative shortfall. The visible control grades and counts toward the 35-45% gate via countComposition (internal/suite/admission.go:23-36, called from admitScenario internal/suite/generation.go:73).

Audit asymmetry, kept honest: for a negative scenario sc.Fault IS the panel's control fault (internal/suite/scenario_draw.go:42-49 == internal/audit/evidence.go:9-16; control arm internal/audit/audit.go:70), so the panel compares the world against itself and negative admission hinges solely on first_observable > 0; positives are audit-gated below the 0.9 cutoff (internal/audit/audit.go:18). Both pre-commit gates are manual process until G-06's generate-audit-regenerate loop ships (docs/design/GAP_ANALYSIS.md:268, open, scheduled S5).

### Perturbations

Injection episodes ride on top of either class; the perturbation layer changes nothing physical (internal/perturb/perturb.go:1-7), so injection orthogonality to the label is the oracle, not the negative-class mechanism.

**Payload path — the graded cells are four budgeted named changes, none written, in this order:**

1. A committed, versioned, domain-agnostic neutral probe corpus as data (no payloads key exists anywhere in domains/ or adapters/ — grep verified this session; docs/fixtures/ does not exist as a directory in the worktree).
2. A suite sampler case for injection_probe (commands.go:47-50,86-97 sample nine kinds today, none for it), drawing payload families from the corpus — acceptance: every suite-drawn injection_probe command carries non-empty corpus payloads (the command log carries params verbatim, internal/suite/command_log.go:21-26).
3. A stamping fix: probedRecord stamps delivery_reason=rewritten unconditionally on targeted rows (transform_payload.go:113-114) while applyProbePayload rewrites only when payloads is a non-empty string array (:118-125); the fix stamps rewritten only on actual byte change, preserving the deterministic payload draw, with enlargedRecord's changed-only stamping (:87-91) as the in-file precedent — acceptance: a probe payload equal to the current value leaves the reason unchanged.
4. Fail-closed validation: validateParams iterates present keys only and passes nil outright (parameters.go:86-97), so an injection_probe without payloads is admitted and silently no-ops; the change makes absent payloads an explicit Apply error — no committed artifact in runs/ carries an injection_probe command (grep verified), so no replay breaks. Acceptance for the whole path: an empty-params probe is an error, never a silent stamp.

**Suite-coverage trap, verified this session:** the coverage gate requires 5 of each of the 19 kinds (generation.go:79-81) but the forced sampler serves only the first two uncovered kinds in Names order (commands.go:37-43,60-68; perturb.go:43-48), so 2n forced placements saturate only the first ~2n/5 kinds — at n=40 both spot-check runs terminate listing injection_probe (and the two kinds ahead of it) as uncovered; larger suites reach it, with empty params until (2) ships.

Reconciliation of the ~2n/5 arithmetic with the two recorded runs (folder documentation pass, from code re-read this pass: `internal/suite/commands.go:37-43,60-68`, `internal/suite/generation.go:67-74`, `internal/suite/admission.go:60-66`): the arithmetic's n is **admitted** scenarios, not attempts — `admitScenario` increments perturbation coverage only for admitted scenarios, and trivially-excluded ones consume attempts without adding coverage. rotating-machinery admitted 40/40 → ~80 placements ≈ 16 kinds covered, 3 uncovered: exactly the recorded tail `[time_encoding precision_edge injection_probe]`. discrete-line-oee admitted only 28 of 120 attempts → ~56 placements ≈ 11 kinds, so ~8 uncovered predicted against the 9 recorded (the recorded list is likewise the exact tail of `Names` order); the one-kind residual is within the arithmetic's rounding and is not further reconcilable from the recorded artifacts, whose full attempt logs were not retained.

### Ground truth and its analytic oracle

Two-sided:

- **Physical trajectory:** solved by the oracle's dual-world noiseless integration over the same verified RK4 integrator the run uses (internal/truth/solver.go:18-21), held to the committed analytic cross-check — not closed forms. Disclosed weakness, stated plainly (folder documentation pass): the committed analytic cross-check verifies the **integrator** (the `internal/world` oracle tests), not this domain's label computation — truth and run share one integrator, so this spec names no second, independent implementation of its own onsets. The independent content it does carry is the strictly-trailing `first_observable` property (reproduced on a fresh artifact, §G2 below) and the injection half's sealed-params/replay join; a closed form for this domain's F1 chain is constructible from committed oracle pieces (`oracle_test.go:57-75` integrator; the first-order-lag closed-form sweep) but is named work this spec does not commit. This gap is part of why determinismOracle scored 6/10.
- **Injection truth:** the sealed command-log params (payloads + window) joined to the ledger's rewritten rows on delivery_id via deterministic replay (replay applies logged params, internal/run/replay_commands.go:29). The ledger answers which rows and when, never the text — a ledger payload_digest column is the rejected alternative (a G4/G6-sensitive contract change needing its own design change).

**Pairing — budgeted new code, none written** (grep for pair/Pair over internal/run non-test files: no matches): the pair key is the suite scenario identity (seed+idx, internal/suite/scenario_draw.go:36-40); the injected cell is the same command log plus exactly one injection_probe command; the pair runner computes per-row digests and fails closed — a missing cell, a rewritten row with no joinable sealed param, or any byte differing outside exactly the rewritten delivery_ids fails the pair, never passes by default. Committed today is whole-artifact equality only (probe_test.go:87-106 trace differ plus verdict bytes, green this session).

Planned regression tests:

- A G1 metamorphic extension of the discrete-line-oee-hardcoded differential (probe_test.go:131-144) to this domain via the generated-suite path.
- A tampered-control test — one flipped byte outside the rewritten rows must fail the pair.
- The (3) stamping regression.

The join over-marks until (3) lands, since rewritten is stamped unconditionally.

### Effectors

`dispatch_technician` advisory declares a real state delta (dispatches counter increment), because effect.state_deltas requires minItems 1 (internal/schemas/domain-spec-v0.1.schema.json:518-525) — a pure no-op is not schema-expressible.

## Feasibility against G1–G9

Gate definitions from docs/QUALITY_BAR.md:57-72. The spec text addresses G1, G2, G4, G6, G7 and G9 by name or content; G3 partially; G5 and G8 are not addressed by the feasibility text and are reported as such.

- **G1 Determinism.** The strongest constraint named as hardest: adding the sampler case without disturbing the RNG draws of the nine already-sampled kinds or the uncovered-forcing order the coverage gate depends on. Sealed command-log params keep the delivery-ledger contract, canonical digests and replay untouched; the injection truth is joined to the ledger on delivery_id via deterministic replay (internal/run/replay_commands.go:29), and proving that join reproduces exactly the ledger's rewritten rows under replay is named among the hardest parts once the stamping fix makes rewritten mean changed-bytes.
- **G2 Analytic cross-check.** The physical oracle is dual-world noiseless integration over the same verified RK4 integrator the run uses (internal/truth/solver.go:18-21), held to the committed analytic cross-check — not closed forms. The boundary/negative-recovery behavior is specified: first_observable strictly trails onset for the divergence fault (solver_scan.go:52-61), verified strictly trailing on a fresh rotating-machinery artifact this session. The sigma denominator is the declared pre-quantization sigma (detector_math.go:52-58) while the quantity is computed on quantized readings — a solver-wide property disclosed, evidenced on a fresh artifact this session, and correctly scoped out as a separate truth-package change.
- **G3 Reference consumer.** Policy-gated intent equality is the consumer's metamorphic assertion over the two byte-identical verdict files; whole-artifact verdict byte-equality on inproc and file sinks passes today (probe_test.go, 4/4 this session), but the graded row-scoped pair runner does not exist yet.
- **G4 Delivery ledger.** The ledger records when and that a row was rewritten, never the text (internal/model/event.go:38-49); the design deliberately keeps the ledger contract untouched, and the payload_digest column alternative is rejected as G4/G6-sensitive pending its own design change.
- **G5 Quiescence.** Not addressed by the feasibility text; nothing in this domain's sketch exercises the quiescence barrier beyond what every domain inherits.
- **G6 Sealed oracle.** The sealed "what" is the command-log params (payloads + window), recoverable per delivery_id by deterministic replay; the sealed-params join proof under replay is the named hardest part. Prefix-equivalence as such is not named in the final text.
- **G7 Injection probe (the headline gate).** Committed and verified this session: whole-artifact paired-cell byte equality on inproc and file sinks including the shipped discrete-line-oee domain (`go test ./internal/run/ -run TestInjectionProbe -v` passes 4/4), attacker_controlled-only targeting (transform_payload.go:109-116). The graded cells exist only after the four named changes ship (corpus, sampler case, stamping fix, fail-closed validation), none written; the domain JSON alone under-delivers the stated purpose on the generated-suite path. The suite cannot place injection_probe at n=40 at all — the forced sampler serves the first two uncovered kinds in Names order and InjectionProbe is last (commands.go:37-43,60-68; perturb.go:43-48), so the coverage gate (generation.go:79-81) is unreachable for this kind until suites grow or the sampler case ships.
- **G8 silent_no_effect.** Not addressed by the feasibility text by name.
- **G9 Trivial baseline.** Stays honest: the audit cannot separate injected text from noise — a legitimate mechanism finding — and the G-06 reportable failure path exists if the physics is trivial (docs/design/GAP_ANALYSIS.md:268, open, scheduled S5). The two suite-floor risks are reproduced this session, not argued: (1) graded negatives require a detector-visible control — discrete-line-oee's invisible control yields negative fraction 0.00, trivial_excluded 92 and 28 admitted of 120 attempts, while rotating-machinery's visible control admits 40/40 with 16 negatives, trivial_excluded 0 and first_observable strictly trailing injection time on the negative class (exact commands and artifacts named in the sketch; n=40 spot checks, not release evidence); (2) the positive class must pass the hindsight audit below the 0.9 cutoff (internal/audit/audit.go:18) — the discrete-line run rejected its positives too. Negative scenarios cannot be audit-gated at all because their fault equals the panel's control arm (scenario_draw.go:42-49 == evidence.go:9-16; audit.go:70), so admission hinges solely on first_observable > 0; both gates are manual process until G-06 ships.

Hardest parts, stated plainly: sizing the drive_state deviation so the positive fault is detector-visible yet hindsight-ambiguous while the transient control self-settles; proving the sealed-params join reproduces exactly the ledger's rewritten rows under replay once the stamping fix makes rewritten mean changed-bytes; and adding the sampler case without disturbing the RNG draws of the nine already-sampled kinds or the uncovered-forcing order the coverage gate depends on. Sealed command-log params keep the delivery-ledger contract, canonical digests and replay untouched; the cost, stated plainly, is that the ledger alone answers "which rows, when", never "what text".

## Quality bar result

Scores are 0-10 per dimension. Bar: total at least 85/100 across the seven dimensions, with no single dimension below 7/10.

| Dimension | Score (0-10) |
| --- | --- |
| consumerValue | 6 |
| dataExpressibility | 7 |
| determinismOracle | 6 |
| ledgerQuiescence | 6 |
| novelty | 5 |
| sealedTruth | 7 |
| specificity | 9 |
| **Raw total** | **46/70** |
| **Normalized total** | **66/100** |
| **Met the bar?** | **Not met** — 66 < 85, and four dimensions sit below 7/10 (consumerValue 6, determinismOracle 6, ledgerQuiescence 6, novelty 5) |

## Five-lens review

### 1. Determinism and analytic oracle (score 7)

**Found:** The stated ground-truth anchor was factually wrong — "the sealed delivery ledger knows exactly which text entered the stream and when": LedgerRecord carries delivery_id/seq/times/delivered/reason and no payload content (internal/model/event.go:40-52), and it is director-side always. The pairing protocol was underspecified — "byte-equality between clean and injected cells" requires two runs differing only in injection activation, but which cell is negative-class for the physical detector and how the pair is keyed in the suite (scenario identity, internal/suite/scenario.go:53-65) was unstated, so scoring could not fail closed when one cell of a pair is missing.

**Changed:** The injection ground truth was re-anchored on the delivered stream plus ledger reasons, with the sealed "what" being the command-log params recoverable per delivery_id by deterministic replay; the fail-closed rule is now explicit (a missing cell, a rewritten row with no joinable sealed param, or any byte differing outside exactly the rewritten delivery_ids fails the pair); a G1 metamorphic extension of the discrete-line-oee differential to this domain via the generated-suite path is a named planned regression; the negative-recovery oracle (injection window closing with no physical fault, first_observable_time=0 handling via the suite's trivial path) is covered by the detector-visible negative-class design.

### 2. Data-not-code expressibility (score 7)

**Found:** The planted-text path is empty in generated suites — the suite never sets the payloads param (internal/suite/commands.go:86-97), and applyProbePayload without payloads leaves the text unchanged (transform_payload.go:118-125), so on the benchmark path the injection stamps Reason=DeliveryRewritten but plants no instruction text; an unpriced, domain-agnostic runtime gap the original spec called "all committed". Arithmetic error: value_shape=boolean is 1 today, so this domain raises it to 2, not 3. Otherwise the cleanest gates story of the pool: attacker_controlled (schema :335-339) and attacker_controlled-only targeting are committed, and the refusal to let diagnosis depend on enum/text is exactly right (enum values are random draws, reading_values.go:74-80).

**Changed:** The payload path is pre-committed as four domain-agnostic budgeted changes — a committed neutral probe corpus as data, a suite sampler case drawing from it, the stamping fix, and fail-closed absent-payloads validation — each with an acceptance criterion and an ordering (the stamping fix must land before the join); the boolean arithmetic is corrected to 1->2, and discrete-line-oee's shift_log (human_driven, attacker_controlled) is cited as the shipped channel pattern.

### 3. Consumer testing value (score 6)

**Found:** As a domain, its graded content likely collapsed: with an all-F0 stuck-boolean fault scored by channel_divergence between two discrete channels, noise-free discrete channels give hindsight-fitted fixed_threshold ~1.0 balanced accuracy against any control (audit cutoff 0.9, audit.go:18), so virtually every fault scenario trivials out and the domain cannot meet the suite floor — a mechanism fixture wearing a domain slot. The real consumer value (the injection differential) does not need a new domain to start; discrete-line-oee already declares an attacker_controlled shift_log, so the quarantine cells could be prototyped on shipped data. The value_shape arithmetic did not reconcile with the portfolio.

**Changed:** Option (b) was taken — the physical fault was given real ambiguity: intermittent stuck-at with duty-cycle onset and a noisy derived-status numeric coupling (the F1 lag + integrator chain), with the negative control made detector-visible following rotating-machinery's pattern; the byte-equality differential is now specified precisely (same command log plus exactly one injection_probe command; per-row digest equality outside exactly the rewritten delivery_ids; fail-closed) and the quarantine evidence the consumer must produce is enumerated to what exists — per-seq admission outcomes (malformed/rejected/out_of_contract), not a claimed durable quarantine record.

### 4. Delivery ledger and quiescence (score 6)

**Found:** The central oracle claim was false as written — the ledger has no payload/value field (internal/model/event.go:38-49) and knows that a row was rewritten, never what text; injection ground truth was re-derivable only from the sealed perturbation seed/params or from the delivered trace, the latter being consumer-visible evidence that cannot double as the sealed oracle without a provenance argument. The tamoz "durable quarantine" cells need a record type that does not exist — the run artifact publishes trace/ledger/world_state_history/verdict only (internal/run/finalize.go:114-128). The G7 fit itself is the strongest in the pool and the committed machinery is verified.

**Changed:** The injection-truth record was chosen: ground truth is sealed perturbation params joined on delivery ids under deterministic replay (the second offered option), with the ledger payload_digest extension rejected as a G4/G6-sensitive contract change needing its own design change; the neutrality differential is extended beyond the discrete-line-oee hard-coding (probe_test.go:131-144) as a planned regression; the quarantine cells are scoped to what exists (admission outcomes) with no durable quarantine record implied.

### 5. Novelty and portfolio balance (score 4)

**Found:** Both headline novelty claims were contradicted by the repo — "first domain where text is the attack surface" is false against discrete-line-oee (declares the shift log attacker_controlled and calls itself the injection probe's home domain, domains/discrete-line-oee.domain.json:5,58-62) and against catalog 25 auth-and-edr, "the native home of the injection probe" (docs/design/DOMAIN_CATALOG.md:631-634); the boolean axis-fill arithmetic was wrong (coverage shows boolean=1, so 1->2); the residual was largely 25 auth-and-edr minus its actual novelty (the adversarial-adaptive fault model, DOMAIN_CATALOG.md:624-630); it overlaps SI-06 and duplicates discrete-line-oee's shipped injection role — the weakest marginal world in a ten-domain set, occupying the slot where the portfolio most needs the unhosted adversarial frontier.

**Changed:** The novelty claim was re-scoped to the genuinely conditional increment — the paired-cell byte-equality differential as a graded, schema-declared, suite-generated property with injection orthogonality enforced cell-against-cell exists for no domain today because the probe kind is unreachable at n=40 and payload-less when reached, so the first domain to exercise the injection probe end-to-end does so by shipping the budgeted machinery — with the fold recommendation restated as a decision rule: if the portfolio ships the injection-cell machinery under SI-06 instead, this domain reduces to a second boolean domain and yields the slot to catalog 25; it is explicitly not the adversarial-adaptive fault model, which outranks it for the slot. Narrower unconditional increments are named (boolean blind leg, intermittent/duty-cycle onset on a divergence fault, second boolean-shaped domain), both citations corrected, and both arithmetic claims fixed against a fresh `./bin/streamsim catalog coverage` run (value_shape=boolean(1), text(1)).

## Improvement history

The material for this ask provides the original consolidated spec and the final spec but not the per-round transcripts of the 10 improvement rounds; the history below is therefore the cumulative delta between the two endpoints, organized by theme. Each item names what the original said and what the final spec says instead.

1. **Ground-truth anchor corrected.** Original: "the sealed delivery ledger knows exactly which text entered the stream and when" and "the injection ground truth is the ledger's record of injected text (what and when)". Final: the ledger carries when and that a row was rewritten, never the text (internal/model/event.go:38-49); the sealed "what" is the command-log params recoverable per delivery_id by deterministic replay, and the ledger answers which rows and when only.
2. **"All committed" corrected to a four-change budget.** Original feasibility: "injection_probe, attacker_controlled, and confirmation requirements are all committed". Final: attacker_controlled-only targeting and whole-artifact neutrality are committed, but the graded cells need the corpus, the sampler case, the stamping fix, and fail-closed validation — none written — each with an acceptance criterion.
3. **Suite-coverage trap discovered.** Absent from the original. Final: the forced sampler serves the first two uncovered kinds in Names order with InjectionProbe last, so at n=40 neither spot-check suite can place injection_probe at all; the coverage gate is unreachable for this kind until suites grow or the sampler case ships (reproduced this session on both spot-check domains).
4. **Dynamics corrected from all-F0 to F1, and the negative class made detector-visible.** Original: "Dynamics: all F0 ... the physical trajectory is F0 closed-form", with a stuck boolean sensor fault and the injection variant as the negative class. Final: two F1 entries (first_order_lag, integrator) with dt_ms; the negative-class control is a brief self-settling intermittent deviation of drive_state following rotating-machinery's pattern, because an invisible control seals first_observable_time=0 and yields negative fraction 0.00 (reproduced: discrete-line-oee at n=40 gives 0.00 negatives, trivial_excluded 92, 28 of 120 admitted; rotating-machinery admits 40/40 with 16 negatives). "Cheapest of the ten to build (no physics beyond F0)" is replaced by the honest cost statement.
5. **Solver precision corrected.** Original: closed-form F0 ground truth. Final: dual-world noiseless RK4 integration over the same verified integrator, held to the committed analytic cross-check; the oracle's readings are noise-free but not pre-quantization (it reuses observe()); the pre-quantization-sigma denominator on quantized legs is disclosed as a solver-wide caveat, evidenced, and scoped out as a separate truth-package change.
6. **Pairing made a budgeted deliverable with fail-closed semantics.** Original: a one-line "byte-equality between clean and injected cells". Final: a named pair key (suite scenario identity, seed+idx), a per-row digest comparison outside exactly the rewritten delivery_ids, the delivery_id/seq join against sealed params, fail-closed on any mismatch, the ordering constraint that the stamping fix must land first, and three named regression tests (metamorphic extension, tampered-control, stamping).
7. **Quarantine claims scoped to what exists.** Original: "malformed payloads must land in durable quarantine rather than silent drop". Final: malformed payloads land in per-seq admission outcomes (malformed/rejected/out_of_contract, consumer-verdict-v0.1.schema.json:73); no durable quarantine record type exists (internal/run/finalize.go:114-128) and none is claimed; policy-gated intent equality is the consumer's metamorphic assertion, not a simulator-scored field.
8. **Novelty claims corrected and a fold rule added.** Original: "First domain where text is the attack surface rather than evidence ... Raises value_shape=text to 2 and value_shape=boolean to 3". Final: the occupied ground is stated first (discrete-line-oee's attacker_controlled shift_log; catalog 25's native-home role and its Q7 pairing with 09), the increment is made conditional on shipping the machinery, boolean is corrected to 1->2 against a fresh catalog coverage run, and the fold/decision rule is explicit: yield the slot to catalog 25 if SI-06 ships the injection-cell machinery.
9. **Audit honesty made explicit.** Absent from the original. Final: for negatives, sc.Fault IS the panel's control arm, so negative admission hinges solely on first_observable > 0; positives are gated below the 0.9 hindsight cutoff; both gates are manual until G-06's generate-audit-regenerate loop ships (docs/design/GAP_ANALYSIS.md:268, open, scheduled S5); the discrete-line spot check rejected its positives as well as grading no negatives.
10. **Effector and schema constraints pinned.** Original: "dispatch_technician advisory keeps the loop scoreable" (unconstrained). Final: the advisory declares a real state delta (dispatches counter increment) because effect.state_deltas requires minItems 1 (schema :518-525) — a pure no-op is not schema-expressible; likewise the status leg is numeric rather than enum because enum values are uniform random draws (reading_values.go:74-80), and the boolean renders as raw >= 0.5 (:32-41).

## Open questions

1. **Slot decision vs SI-06 and catalog 25.** If the portfolio ships the injection-cell machinery under SI-06 (drift+poison+injection), this domain reduces to a second boolean domain and should yield the slot to catalog 25. The decision rule is written; the portfolio-level call is not made here.
2. **SNR floor on quantized legs.** The threshold denominator is the declared pre-quantization sigma while the quantity is computed on quantized readings (detector_math.go:52-58 vs observe.go/reading.go:33). Acknowledged and unresolved by design — strict trailing still holds (evidenced on a fresh artifact) — and scoped out as a separate truth-package change.
3. **Suite floor sizing.** The positive fault must be detector-visible yet hindsight-ambiguous below the 0.9 cutoff while the transient control self-settles; no sizing has been proven against the audit yet — the discrete-line spot check rejected both its positives and graded no negatives, so the n=40 spot checks are risk evidence, not release evidence.
4. **Sampler ordering.** Adding the injection_probe sampler case must not disturb the RNG draws of the nine already-sampled kinds or the uncovered-forcing order the coverage gate depends on; how to guarantee that (append-only case, separate stream) is undecided.
5. **Audit automation dependency.** Both pre-commit gates (positive hindsight gating, negative first_observable admission) are manual process until G-06's generate-audit-regenerate loop ships (docs/design/GAP_ANALYSIS.md:268, open, scheduled S5).
6. **Portfolio axis counts beyond 2.** `./bin/streamsim catalog coverage` reports value_shape=boolean(1) and text(1), so this domain takes text 1->2 and boolean 1->2; counts beyond 2 depend on other candidates and could not be verified from this repository.
7. **Ledger payload digest.** Rejected here as a G4/G6-sensitive contract change needing its own design change; if the join-under-replay proof proves brittle once rewritten means changed-bytes, that design change is the fallback.

## Verification performed for this document

Checks executed during this writing pass (n=40 suite runs are spot checks, not release evidence):

- `go test ./internal/run/ -run TestInjectionProbe -v` — PASS, 4/4 subtests (inproc-synthetic, file-synthetic, inproc-shipped, file-shipped).
- `./bin/streamsim suite --domain discrete-line-oee --profile nominal --out /tmp/si07-doc-dloee --n 40` — terminal_state: `perturbation coverage not met at the declared size: [unit_mismatch oversize malformed nan_inf storm producer_flap time_encoding precision_edge injection_probe]; negative-class fraction 0.00 outside [0.35, 0.45]; correlated_cascade scenarios: 0 < 3; sensor_pathology scenarios: 0 < 10; audit rejected too many scenarios: 28 non-trivial of 120 attempts`, trivial_excluded 92.
- `./bin/streamsim suite --domain rotating-machinery --profile nominal --out /tmp/si07-doc-rot --n 40` — 40/40 scenarios admitted, 16 negatives (0.40), trivial_excluded 0, all negative labels first_observable > 0 (checked in the artifact JSON); terminal_state lists injection_probe among the uncovered kinds: `perturbation coverage not met at the declared size: [time_encoding precision_edge injection_probe]` (plus profile-specific correlated_cascade/sensor_pathology shortfalls unrelated to negatives).
- `./bin/streamsim catalog coverage` — value_shape `boolean: 1`, `text: 1`.
- Grep for pair/Pair over internal/run non-test Go files — 0 files.
- Grep for `payloads` over domains/ and adapters/ — 0 matches; docs/fixtures/ does not exist as a directory (so the spec's "grep over domains/, adapters/, docs/fixtures finds no payloads key anywhere" holds, with the caveat that docs/fixtures is absent rather than merely clean).
- Line-level source checks: internal/model/event.go:38-49 (LedgerRecord, no payload field), internal/perturb/transform_payload.go:105-125 (attacker_controlled gate, unconditional rewritten stamp, payloads no-op), internal/perturb/parameters.go:86-97 (absent keys admitted, nil passed), internal/cli/run_script.go:30 (nil params), internal/suite/commands.go:35-97 (forced sampler, no injection_probe params), internal/suite/generation.go:79-81 (coverage gate), internal/perturb/perturb.go:43-48 (InjectionProbe last), internal/run/probe_test.go, internal/run/replay_commands.go:29, internal/run/finalize.go:114-128, internal/world/reading_values.go:32-41,62-80, internal/world/observe.go:10-21, internal/world/reading.go:33, internal/world/dynamics.go:70-120, internal/world/state.go:18-31, internal/truth/solver.go:18-21,131-135, internal/truth/solver_scan.go:52-61, internal/truth/detector_math.go:52-58, internal/suite/scenario.go:20-22,53-65, internal/suite/scenario_draw.go:36-49, internal/suite/admission.go:23-36,60-66, internal/suite/command_log.go:21-26, internal/audit/audit.go:16-20,68-72, internal/audit/evidence.go:9-16, internal/score/admission.go:5-15, internal/schemas/domain-spec-v0.1.schema.json:97-99,108-128,212-220,518-525, docs/contracts/domain-spec-v0.1.schema.json:335-339, docs/contracts/consumer-verdict-v0.1.schema.json:68-78 (per-seq outcomes; no intent/cell/pair concept), docs/design/GAP_ANALYSIS.md:268, docs/design/DOMAIN_CATALOG.md:624-634,679, docs/QUALITY_BAR.md:57-72 (gate table), domains/discrete-line-oee.domain.json:5,58-62,81,100-105, domains/cold-chain-transit.domain.json:122,154,232-236, domains/shipment-exception.domain.json:219-223, domains/rotating-machinery.domain.json:242-245, and the first_order_lag/integrator domain census.
