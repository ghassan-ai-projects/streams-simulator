# SI-02 · feeder-switching

- **ID:** SI-02
- **Slug:** `feeder-switching`
- **Status:** evaluated — **not admitted.** The workflow's quality-bar evaluation recorded this idea "met"/"accepted" at 80/100, but 80/100 is below the stated bar (≥85/100 total, no dimension below 7/10); the discrepancy is documented in [Quality bar result](#quality-bar-result) and reconciled folder-wide in [../README.md](../README.md) §3. No domain artifact exists, so "accepted" never meant admission to a shipped catalog — the two words are kept strictly apart (see [../README.md](../README.md) §1).
- **Serves:** `tamoz` primarily (ack/effect reconciliation; the cell names "kill-matrix exactly-once oracle" and "three-valued reconcile journal" are brief-supplied vocabulary — they appear nowhere in this repository and are carried as such, per the caveat in [../research-notes.md](../research-notes.md) §3); `agentic-stream` (policy-gated intents, reconsideration when a correction invalidates a succeeded command) secondarily.
- **Prerequisites declared:** (1) a generic suite/driver capability that emits replayable command logs repeating a `command_id`; (2) an exactly-once run-artifact audit over existing `EffectorCall` records. Neither is domain data; both are named, not hidden.

## Purpose

A radial distribution feeder whose switch/breaker commands make ack/effect decoupling and
command idempotency the primary scored question. Exposes four cells the machinery already
implements:

1. **Applied-but-unacked** — mode `ack_lost` applies the effect and returns `Accepted=false`
   (internal/world/effectors.go:149-157).
2. **Acked-before-applied** — the ack returns synchronously while the physical effect is still
   inside `dead_time_s`; kicks land at `atNS+dead_time` with the lag tau
   (internal/world/effects.go:19-30).
3. **Duplicate `command_id` inside the idempotency window** — the world replays the cached
   result and applies no second effect (internal/world/effectors.go:85-93).
4. **The scored boundary cell** — the same duplicate after window expiry, which applies a
   second effect (`replayInvocation` deletes the expired entry,
   internal/world/effectors.go:90).

A correct consumer reconciles to `outcome_believed=unknown` under an ambiguous ack and never
treats an ack as verification
(internal/schemas/consumer-verdict-v0.1.schema.json:93-97).

Duplicate commands are neither domain data nor a perturbation: the perturbation layer only
transforms world-to-consumer emissions (internal/run/emission.go:18) while commands reach
`InvokeEffector` directly (internal/world/effectors.go:44-56), so
internal/perturb/transform_duplicate.go:7-15 never sees them. They arise from consumer retries
that sampled `ack_lost`/slow acks legitimately provoke, or from a driver replaying a command
log that repeats a `command_id`. The replayable log format already carries `command_id`
(internal/suite/command_log.go:14-19) but no generator emits a repeat today
(internal/suite/commands.go:26-35), so this spec names that small generic suite/driver
capability as a declared prerequisite rather than presenting it as existing perturbation
vocabulary.

## How it runs on this simulator

Everything below is domain DATA against the one schema
(internal/schemas/domain-spec-v0.1.schema.json). No domain code branches.

### Entities

6–12 radial feeders off one bus, each with hidden states:

- `loading_pct` — initial 62, clamped 0–130.
- `breaker_open` — boolean.
- a CT-A corruption state (bias source for the sensor-pathology faults).

### Signals

- Redundant current-transformer channels `feeder_current_a` / `feeder_current_b`, both
  observing `loading_pct` at gain 4 (62 % → 248 A), gaussian sigma 3 A each, 10 s periodic.
- `breaker_position` — boolean event channel on `breaker_open`.
- The redundant pair is declared in `confirmation_channels`, so the G8 `silent_no_effect`
  shadow covers the current channels, not just breaker position. Confirmation channels read
  the counterfactual shadow value: `shadowValue` at internal/world/state.go:55-62, selected
  for confirmation channels at internal/world/reading.go:26.

### Dynamics

- Effects carry `dead_time_s` 10–20 and `time_constant_s` 30–45 (schema :543-544); kicks land
  at `atNS + dead_time_s` and relax with tau (internal/world/effects.go:19-30).
- `idempotency_window_s` 900 (schema :547, runtime default 3600 at
  internal/world/effectors.go:116-122). A re-invoked id inside the window returns the cached
  result **before** any failure-mode sample or `recordCall`
  (internal/world/effectors.go:49-51); after expiry the entry is deleted and a second effect
  applies (internal/world/effectors.go:90).
- The relay's autonomous trip is recorded on the refusal call itself
  (`refuseInterlock` → `recordCall` mode `reject`,
  internal/world/effectors.go:96-105).

### Perturbations

- Transport-side only and background: `drop` / `duplicate_burst` / `delay_tail` from the
  existing catalog (internal/perturb/perturb.go:42-46) corrupt the emission stream for G4
  green. No perturbation touches commands, and ack-failure reasons stay world-side
  enum-like strings, never routed through delivery perturbations.
- Command-side stimulus has exactly two sources:
  - (a) consumer retries provoked by sampled `ack_lost`/slow acks — the world's idempotency
    window resolves apply-once with no new machinery;
  - (b) driver-replayed command logs repeating a `command_id` inside and deliberately after
    the window — a generic suite capability to add (no domain code), which also yields the
    scored expiry cell.

### Effectors

`open_switch` and `close_switch`, `actuation=interlocked`.

- **Interlock:** one state and one operator — `close_switch` refused while
  `loading_pct >= 95` — because the runtime evaluates a single state with one
  `lt/lte/gt/gte` operator (internal/world/effector_policy.go:63-76; schema :494-512). A
  compound loading/underfrequency predicate is not expressible; underfrequency is dropped,
  not hand-waved. Refusal is terminal (`ErrInterlockRefused`,
  internal/world/effectors.go:33-35) and the consumer metric is "refusal produces no retry"
  (docs/design/GROUND_TRUTH_AND_SCORING.md:197).
- **Ack failure modes**, declared per effector over the schema enum (schema :485):
  - `open_switch`: ok 0.50 / slow 0.15 / ack_lost 0.10 / partial 0.05 /
    confirmed_no_effect 0.10 / silent_no_effect 0.10
  - `close_switch`: ok 0.70 / slow 0.10 / ack_lost 0.10 / silent_no_effect 0.10
- **The two no-effect modes are never blurred when naming the sensor pair:** under
  `confirmed_no_effect` the CTs keep reporting truth while the ack lies
  (internal/world/effectors.go:144-146); under `silent_no_effect` they report the
  counterfactual shadow, so the loop watches a recovery that never happened
  (internal/world/effectors.go:159-169).

### Ground truth and its analytic oracle

Fault side, computed not asserted — the sketch was built as a throwaway domain and run
through the shipped tooling (`go build ./cmd/streamsim`, then `./bin/streamsim suite -domain
feeder-switching -profile nominal -n 100 -seed 7` and `-profile sensor_pathology -seed 11`;
both runs admitted 100/100 scenarios with `trivial_excluded 0`, negative fraction 0.41,
~26 s each):

Provenance and a declared evidence gap, added during the folder's documentation pass: the
throwaway domain JSON was never committed and its disposition (kept or deleted) was not
recorded — no `feeder-switching` file exists under `domains/` (checked this pass) — so these
numbers are development-run evidence, not reproducible from the tree; the same status SI-06
gives its removed scratch harness, stated here because SI-02's own text never stated it. The
run summaries above are also **incomplete**: `internal/suite/admission.go:50-58` demands
`correlated_cascade scenarios ≥ 3` and `sensor_pathology scenarios ≥ 10` unconditionally while
`internal/cli/suite.go:22` generates exactly one profile per invocation (both re-read this
pass; the diagnosis is SI-08's, verified in code), so both single-profile runs must have
printed `correlated_cascade scenarios: 0 < 3` and `sensor_pathology scenarios: 0 < 10` in
`terminal_state`. Their omission from this document is an evidence-report gap, flagged rather
than resolved; the quoted admission counts, audit figures and timings stand as recorded.

- **`feeder_overload`** ramps `loading_pct` +45 pct (+180 A on the gain-4 channels) with
  detector `peer_residual`; measured effective_sigma 3.286 A = 3·√(1+1/5) for six siblings
  (internal/truth/detector_math.go:60-63), `first_observable` +120 s (one 120 s scan tick) at
  declared SNR 3, unavoidable at SNR 10.
- **`ct_a_drift` / `ct_a_stuck`** bias CT-A through the `observation_bias` hook (schema
  :256-266; coef 45 A full-scale, `noise_scale_at_full` 0.1), detector `channel_divergence`;
  measured effective_sigma 4.243 A = √(3²+3²)
  (internal/truth/detector_math.go:16-22). The ramp form crosses first-observable one scan
  tick after onset; the intermittent stuck form lands anywhere from +0 s to +27657 s
  depending on duty phase — the solver scans noiseless clean-vs-faulted worlds against
  declared SNR × sigma (internal/truth/solver_scan.go:36-54).
- **`downstream_fault_trip`** steps +55 pct (+220 A) and is kept deliberately as the audit
  boundary case (see [Feasibility](#feasibility-against-g1g9), G9).

**Safety is a scored question, not a fill.** `feeder_overload` and `downstream_fault_trip`
declare `deadline_s` 1200/900 with counterfactual `{if_no_action: irreversible conductor
thermal damage; if_action_by_deadline: section opened inside the thermal window}` (schema
:437-448). The twist: under `silent_no_effect` the loop reports success while the deadline
runs out. The false-success metric comes from verdict actions × the EffectorCall log, which
is the exactly-once authority (`EffectorCall`: Mode/Accepted/InterlockRefused/EffectApplied/
ResultDigest, internal/world/world.go:115-129) — not the delivery ledger (`LedgerRecord` has
no command/effector fields, internal/model/event.go:38-49).

**Sealed truth stays fault-shaped** — label plus three onsets in the closed
`ground-truth-v0.1` schema (:7-12) with `world_state_reference` (:96-99). Per-command
reconcile verdicts aggregate through the existing cross-checks: `matchesActions` over
EffectorCalls (internal/score/actions.go:8-17), `respectsInterlocks`
(internal/score/actions.go:19-33), and `outcome_believed=succeeded` against
`EffectApplied=false`, landing in the loop-metric class of
docs/design/GROUND_TRUTH_AND_SCORING.md §3 (:159; loop metrics table :221-231, false-success
rate :231). No new truth record type; the exactly-once invariant is a run-artifact audit,
not sealed truth.

**Exactly-once audit, explicit:** for each `command_id` exactly one `EffectApplied=true`
EffectorCall per window; a re-invoked id inside the window returns the cached result before
any mode sample or `recordCall` (internal/world/effectors.go:49-51). Proven by a metamorphic
regression test that replays a command log containing duplicates (extends
`TestEffectorIdempotencyAndInterlock`, internal/world/effector_test.go:11) and asserting one
`recordCall` per distinct `command_id` inside and outside the declared window.

**Profiles:** mandatory `nominal` and `sensor_pathology` (fault_weights
`{ct_a_stuck 0.35, ct_a_drift 0.35, morning_spike_none 0.30}`); `correlated_cascade`
declared `not_applicable` — all faults are per-feeder injections (schema :585; note the
declaration makes the catalog honest about intent but does **not** suppress the suite's
unconditional shortfall demand — see the evidence-gap note under Ground truth above);
`negative_class_fraction` 0.40, enforced at [0.35, 0.45] by schema (:621-628) and admission
(internal/suite/admission.go:41-42). The negative control `morning_spike_none`
(intermittent +8 pct, duty 0.25) is also the audit panel's control series
(internal/audit/evidence.go:30-36).

## Feasibility against G1–G9

Gate IDs follow docs/QUALITY_BAR.md:62-72. One schema, one compile, zero new record types:
every mechanism above is verified present and data-expressible in
internal/schemas/domain-spec-v0.1.schema.json — the seven ack modes with the C-05
scored-separately note (:485-487), single-state interlock with `autonomous_action`
(:494-512), mandatory `confirmation_channels` for no-effect modes (:513-517, allOf
:549-570), `dead_time_s`/`time_constant_s` (:543-544), `idempotency_window_s` (:547),
`observation_bias` (:256-266), onset shapes (:376-379), deadline/counterfactual (:437-448).
The interlocked pattern already ships in
domains/discrete-line-oee.domain.json:122-127 and is reused, not reinvented.

| Gate | Standing for this domain |
| --- | --- |
| G1 Determinism | Untouched by the sketch: the domain adds data only; the driver-replayed duplicate log is a deterministic input to the same replay function. |
| G2 Analytic cross-check | Carried by the solver, live-verified: effective-sigma formulas (internal/truth/detector_math.go:16-22, :60-63) and the clean-vs-faulted SNR scan (internal/truth/solver_scan.go:36-54) produced the first-observable numbers quoted above. |
| G3 Reference consumer | Not specifically addressed by the spec; nothing in the sketch changes the operator surface — invokes travel the standard `sim.effector.invoke` path. |
| G4 Delivery ledger | Background green only. The ledger is **not** this domain's oracle: `LedgerRecord` has no command/effector fields (internal/model/event.go:38-49). Apply-once evidence lives in the EffectorCall log (internal/world/effector_log.go:8-17) plus the idempotency cache (internal/world/effectors.go:116-122). |
| G5 Quiescence | Not specifically addressed; effects are ordinary scheduled kicks, so the world's existing pending-kick accounting (internal/world/effector_log.go:59-69) applies unchanged. |
| G6 Sealed oracle | Untouched — only the fault label is sealed (closed `ground-truth-v0.1` schema, `additionalProperties:false`); exactly-once lives in the run artifact's EffectorCall log, deliberately outside sealed truth. |
| G7 Injection probe | Unthreatened: no attacker-controlled channels are declared; ack-failure reasons are world-side enum-like strings, never routed through delivery perturbations. |
| G8 `silent_no_effect` | The headline: declaring the CT pair in `confirmation_channels` scopes the shadow-state proof (internal/world/state.go:55-62 via internal/world/reading.go:26) to the current channels, and the confirmed/silent distinction is scored against the actual machinery (internal/world/effectors.go:144-169). |
| G9 Trivial baseline | Verified live rather than asserted: a direct panel run through `audit.Panel` on a tight 24 h window with a clean +6 h onset marked the step/ramp load faults TRIVIAL — `fixed_threshold` 0.982-0.983 — while the intermittent CT fault survived at 0.874 < 0.9 (`BalancedAccuracyCutoff`, internal/audit/audit.go:18); the same step fault under suite geometry (72 h window, onset +18 h) scored 0.677, non-trivial, which is why the suite runs excluded none. Two conclusions are carried into the spec: audit survival is geometry-dependent (window, onset draw, negative control), so the domain must pin duration/onset/negative-class and re-run the audit on any change, with trivial step faults expected and kept as mechanism fixtures; and the detection-audit panel does not measure the command/ack dimension at all (its five detectors run on channel streams, internal/audit/audit.go:21-23), so this domain's primary question is orthogonal to the trivial-baseline audit and is scored by the loop metrics instead. |

Two prerequisites are honestly named as generic changes, not domain data: (1) a suite/driver
capability that emits replayable command logs repeating a `command_id` (format ready at
internal/suite/command_log.go:14-19; generator absent at internal/suite/commands.go:26-35);
(2) the exactly-once audit plus the post-expiry duplicate cell as run-artifact checks over
existing EffectorCall records and the idempotency cache
(internal/world/effector_log.go:8-17, internal/world/effectors.go:116-122).

## Quality bar result

| Dimension | Score |
| --- | --- |
| consumerValue | 8/10 |
| dataExpressibility | 8/10 |
| determinismOracle | 8/10 |
| ledgerQuiescence | 8/10 |
| novelty | 7/10 |
| sealedTruth | 8/10 |
| specificity | 9/10 |
| **Total** | **56/70 = 80/100** |

**Recorded verdict: "met" as supplied** by the workflow's quality-bar evaluation — the bar being:
total at least 85/100 across the seven dimensions, with no single dimension below 7/10.
Recorded discrepancy, not smoothed over: the seven scores sum to 56/70 = 80/100, which is
**below** the stated 85/100 total threshold, while the per-dimension floor **is** satisfied
(minimum 7/10, novelty). Taken literally against the stated bar text, the total clause alone
would fail; the "met" verdict and the stated threshold do not agree, and both are reported
here so the reader can judge.

## Five-lens review

### 1. Determinism and analytic oracle (8/10)

**Found:** "apply-events as sealed ground truth without new truth record types" overstated
the current shape — sealed truth is the closed label record; the EffectorCall log is
director-side evidence consumed by action scoring, not part of sealed truth. The compound
"loading/underfrequency" interlock predicate is not expressible against the single-state
`lt/lte/gt/gte` evaluator (internal/world/effector_policy.go:63-76). Citations otherwise
verified.

**Changed:** the spec now states the exactly-once oracle lives in the run artifact +
EffectorCall log, not sealed truth; the interlock collapses to one state and one operator
(`loading_pct >= 95`) with underfrequency explicitly dropped; the exactly-once invariant is
added as an explicit audit with a metamorphic replay test, and the redundant CTs are declared
as `confirmation_channels` so the G8 shadow covers them. The post-expiry duplicate cell is
added as a scored boundary case.

### 2. Data-not-code expressibility (8/10)

**Found:** mechanism attribution error on the headline — `duplicate_burst` duplicates event
records on the world→adapter path (internal/perturb/transform_duplicate.go:7-15); commands
are never perturbed, and the generated command log contains no effector invokes or duplicate
ids (internal/suite/commands.go:26-35). Novelty claim partly false: discrete-line-oee
already ships interlock + `silent_no_effect` + confirmation channels
(domains/discrete-line-oee.domain.json:122-127).

**Changed:** the duplicate-command stimulus is re-routed to its two real sources (consumer
retry under ambiguous acks; driver-replayed logs) with the missing generator named as a
generic suite prerequisite; the novelty text is corrected to "no shipped domain makes
ack/effect decoupling the primary scored question" with `reset_line` cited as prior
machinery, not an honest-ack counterexample.

### 3. Consumer testing value (8/10)

**Found:** the duplicate-command mechanism was misattributed to the perturbation layer,
which cannot see the command path; the stated "hardest part" was already solved by the
EffectorCall record; redundant-sensor wording risked blurring `confirmed_no_effect` (truth-
telling confirmation) with `silent_no_effect` (counterfactual shadow).

**Changed:** the duplicate cell is rewritten as consumer retry behavior resolved
deterministically by the world's idempotency window; the confirmed/silent distinction is
stated explicitly per mode and this domain is kept as the portfolio's reference design for
G8; the interlock refusal cell is made explicit with the "refusal produces no retry" metric
(docs/design/GROUND_TRUTH_AND_SCORING.md:197).

### 4. Delivery ledger and quiescence (6/10)

**Found:** two misattributions — the perturbation layer processes native emissions only
(internal/run/emission.go:18), and "G4 ledger proves apply-once" is impossible because
`LedgerRecord` carries no command/effector fields (internal/model/event.go:38-49). The
machinery the domain needs (terminal refusal, shadow state, confirmation channels) was
verified live, so the verdict was recoverable.

**Changed:** the exactly-once oracle is restated against EffectorCalls — a re-invoked id
inside the window returns the cached result and appends no new call
(internal/world/effectors.go:49-51) — with the delivery ledger demoted to background G4
green; a regression test proving one `recordCall` per distinct `command_id` inside and
outside the declared window is specified.

### 5. Novelty and portfolio balance (7/10)

**Found:** the load-bearing claim "discrete-line-oee has honest acks" is false — it ships
`silent_no_effect` at p=0.05 on its interlocked `reset_line`; nearest catalog neighbors 14
bess-thermal and 19 rail-trackside (docs/design/DOMAIN_CATALOG.md:361-383, :476-498) were
never differentiated; the safety fill was asserted in four words with no scored question;
the pump-station-relay-bank kill claim was unverifiable.

**Changed:** positioning corrected with the exact residual delta (ack_lost ambiguity +
idempotency-window duplicates + post-expiry cell + reconcile-to-`unknown` as scored
behavior), citing internal/world/effectors.go:85-93,116-122 as the existing machinery the
domain stress-tests; neighbors named and differentiated (theirs is refusal semantics;
this is ack/effect reconciliation under at-least-once delivery, which catalog Q8 at
docs/design/DOMAIN_CATALOG.md:680 does not cover); safety specified as an asymmetric
irreversible cost behind `deadline_s` lost to a false-success loop; portfolio
pre-commitment recorded (SI-02 owns command/ack discipline; merge if SI-09/SI-10 center the
same question). The unverifiable kill claim was removed — re-verified this session: grep
over the repository returned zero references to pump-station-relay-bank in the tree of this
document's writing pass. (Scope correction, folder documentation pass: the pool name now also
appears in [../research-notes.md](../research-notes.md) §1, created later the same day; a
re-run of `grep -rE "pump-station-relay-bank" --include="*.go" --include="*.md"
--include="*.json" .` during that pass confirms zero hits outside `docs/simulation-ideas/`.)

## Improvement history

One improvement round separated the original consolidated spec from the final spec. What it
changed:

1. **Duplicate-command mechanism re-attributed.** Original: "`duplicate_burst` delivers one
   command twice inside the idempotency window" as perturbation vocabulary. Final: the
   perturbation layer only transforms emissions (internal/run/emission.go:18,
   internal/perturb/transform_duplicate.go:7-15); commands reach `InvokeEffector` directly
   (internal/world/effectors.go:44-56); duplicates arise from consumer retries or a
   driver-replayed command log, with the missing generator declared as a prerequisite
   instead of claimed as existing machinery.
2. **Apply-once oracle re-homed.** Original: "G4 ledger proves apply-once". Final: the
   EffectorCall log is the exactly-once authority (internal/world/world.go:115-129,
   internal/world/effector_log.go:8-17); the delivery ledger is background G4 green only
   (internal/model/event.go:38-49).
3. **Sealed-truth overclaim removed.** Original: "sealed world state is exact — apply-once
   is a closed-form function… as sealed ground truth… without new truth record types".
   Final: sealed truth stays fault-shaped (label + three onsets,
   internal/schemas/ground-truth-v0.1.schema.json:7-12); per-command reconcile verdicts
   aggregate through existing cross-checks into the loop-metric class; the exactly-once
   invariant is an explicit run-artifact audit with a named metamorphic test.
4. **Interlock predicate made data-expressible.** Original: a compound
   "loading/underfrequency predicate". Final: single-state `loading_pct >= 95`
   (internal/world/effector_policy.go:63-76), underfrequency dropped rather than
   hand-waved.
5. **Novelty repositioned.** Original: "discrete-line-oee has interlocks but honest acks"
   and "absorbs killed pump-station-relay-bank". Final: both claims retracted —
   discrete-line-oee ships `silent_no_effect` + interlock + confirmation channels
   (domains/discrete-line-oee.domain.json:122-127) and the kill claim has zero grep hits
   (re-verified this session). The delta is stated exactly, catalog neighbors 14/19 and Q8
   are named and differentiated, and portfolio ownership is pre-committed.
6. **Safety made a scored question.** Original: "safety consequence axis" in four words.
   Final: `deadline_s` 1200/900 with counterfactuals (schema :437-448) and the
   false-success-under-`silent_no_effect` twist scored via verdict actions × EffectorCall
   log.
7. **Sketch hardened with live evidence.** The final spec adds concrete entities/states,
   channel gains and sigmas, per-effector ack-mode distributions, and solver numbers
   produced by actually building the sketch as a throwaway domain and running the shipped
   binary — including the discovery that audit survival is geometry-dependent and that the
   audit panel does not measure the command/ack dimension at all. Two line citations were
   also adjusted to the verified locations (shadow read: internal/world/reading.go:26 and
   internal/world/state.go:55-62; audit cutoff: internal/audit/audit.go:18), and the
   feasibility section is now structured against all nine gates.

Net effect on scores: ledgerQuiescence rose 6 → 8 and specificity reached 9/10 after the
round; novelty stayed 7/10 (the machinery is prior art; the scored question is new), giving
the final 80/100.

## Open questions

1. **Duplicate-log generator.** The suite/driver capability that emits replayable command
   logs repeating a `command_id` is named but unimplemented
   (internal/suite/commands.go:26-35 emits no effector invokes). It needs a small generic
   design slice before the domain's primary cells can be scored end-to-end in the suite.
2. **Exactly-once metamorphic test.** Extending
   `TestEffectorIdempotencyAndInterlock` (internal/world/effector_test.go:11) to replay a
   duplicate-bearing command log and assert one `recordCall` per distinct id inside and
   outside the window — where does it live: `internal/world` test or `test/architecture`?
3. **Audit geometry pinning.** Duration, onset draw and negative class must be pinned and
   the audit re-run on any change; the step faults will be trivial by design. Does the
   portfolio accept a domain whose headline dimension (command/ack) is orthogonal to the
   trivial-baseline audit panel?
4. **Post-expiry duplicate semantics.** After window expiry the world applies a second
   effect by design (internal/world/effectors.go:90). Confirm scoring treats that as the
   legitimate boundary cell (consumer must re-decide) rather than an instrument bug.
5. **SI-09 / SI-10 ownership.** If either centers ack/effect reconciliation, merge rather
   than duplicate; SI-02 keeps command/ack discipline as its primary question.
6. **Underfrequency interlock.** Dropped as inexpressible. Is a derived-state extension to
   the interlock schema ever worth it, or is single-state refusal a permanent, documented
   scope decision?
