# Rule promotion: the streams-simulator plan

- **Date**: 2026-10-27
- **Repository**: streams-simulator, branch `rule-promotion` (cut from main `74a0c77c`)
- **Session order**: this is **session 1 of 4**. The repositories are worked one at a time, never in parallel:
  1. streams-simulator (this plan)
  2. agentic-stream: `/Users/ghassan/my-projects/agentic-stream/documentation/design/rule-promotion.md`
  3. tamoz: `/Users/ghassan/my-projects/tamoz/docs/RULE_PROMOTION_PLAN.md`
  4. agentic-stream again: the first-slice run (round AS-9 of its plan)
- **Background, not required to start**: [README.md](README.md) (the 18 decisions), [SCENARIOS.md](SCENARIOS.md)
  (the campaign and scenarios), [PLAN.md](PLAN.md) (the cross-repository overview and findings F1-F9)

This file is self-contained: a session that reads only this file and the repository's `AGENTS.md` can do the work.

## 1. What rule promotion is, and the simulator's part

Agentic Stream turns plant readings into situations and calls Tamoz, an LLM agent, to decide what to do. Rule
promotion lets Tamoz learn from cases it solved and that were verified, and draft a **deterministic rule**; a
human approves it; from then on Agentic Stream handles that kind of case **without an LLM call**. The rules live
in a per-site playbook file outside every repository.

The simulator's job is to **prove whether this is worth it and safe**, before any real site. It supplies a
simulated pump station with sealed ground truth, runs it as an ordered **campaign** (learning weeks, holdout
weeks, disturbances), and grades every run. The same campaign is run against four consumers ("arms"): static
rules, an LLM on every case, learning with a careful approver, learning with an approve-all approver. The truth
decides who did better.

The simulator **never sees, reads or grades with a rule**, and knows nothing of Agentic Stream or Tamoz
(`test/architecture/consumer_knowledge_test.go` forbids their names in production code). Everything here is
generic: a domain is data, a campaign is a list of runs, phase names are plain strings.

## 2. Decisions that bind this repository

| Decision | What it means here |
|---|---|
| The success bar is a simulated campaign on the rotating-machinery world (decision 7) | Build the world and the campaign; the real board is not the bar. |
| Harness split (decision 8) | The simulator owns a **generic phased scenario suite and its grading**. The campaign driver, scripted approvers and campaign score belong to Agentic Stream. |
| F5: one fault per run | The simulator grades one ground-truth record per run (`internal/score/internal/domain/score.go:82`). The first campaign is a **sequence of one-fault runs**; the consumer keeps its state across runs, the world starts fresh each run. |
| F6: never edit the reference domain | `rotating-machinery` is pinned (`internal/domain/runtime_equivalence_test.go:22`, `release-manifest.json`, `test/pin/BEHAVIOUR_PIN.txt`). All changes go into **new files** under `examples/pump-station/`. |
| First slice only | Build what scenarios **S1** (repeat offenders), **S2** (cavitation look-alike) and **S7** (new probes) need. Everything else is listed in §6 and not built now. |
| Good enough, not perfect | The owner prefers the simplest version that works. State limits instead of building more. |

## 3. Starting state (read on 2026-10-27)

- **The reference world**: `domains/rotating-machinery.domain.json`. One kind `pump` (8 instances
  `site-{site}/pump-{n}`), states including `pump_load`, `vibration_true`, `axial_vibration_true`,
  `lubrication`, `maintenance_due`, `vibration_bias`; channels `vibration`, `axial_vibration`, `temperature`,
  `motor_current`, `rpm`, `mode`, `heartbeat`; seven faults (`bearing_wear`, `misalignment`, `overload`,
  `lubrication_due`, `sensor_drift`, `sensor_failure`, `transient_none` the negative class); three effectors
  (`schedule_maintenance` with args `pump_id`, `work_order`; `reduce_load`; `stop_motor`).
- **No correct action ends a fault today.** Effectors only add state deltas; no effector has a `responder`, so a
  fault's contribution persists. `schedule_maintenance` changes only `maintenance_due` and `lubrication`, so after
  maintenance on a worn bearing the vibration stays high, and the consumer can never verify the right action.
  `reduce_load` (the wrong action for wear) lowers vibration at once because vibration follows load. Left as is,
  the world would make the masking trap the only action that "works".
- **The responder machinery already exists** (`internal/world/internal/domain/follow_up.go`, `model.Responder`
  in `internal/model/vocabulary.go:52`): an effector may declare a responder with `match_arg`, `delay_s`,
  `on_match`, `on_miss` and `clear_fault`. When the call's `match_arg` value equals the id of a fault active on
  that entity, `on_match` applies and, with `clear_fault`, the fault is cleared; otherwise `on_miss` applies. No
  shipped domain uses it yet.
- **D-116 guard**: a plant wire session refuses a planned fault that acts on a state a command of the instance's
  kind sets (`fault_on_commanded_state`, `internal/plantwire/internal/domain/disturbances.go:75-90`), because the
  next command would cancel it (`governance/DEFERRED.md` D-116). In this world that refuses `overload`
  (`pump_load`, set by `reduce_load`) and `lubrication_due` (`lubrication`, `maintenance_due`, set by
  `schedule_maintenance`). `bearing_wear`, `misalignment`, `transient_none`, `sensor_drift` and `sensor_failure`
  are scriptable.
- **D-121**: `schedule_maintenance` never clears `vibration_bias`, so the sensor faults cannot be repaired
  (`governance/DEFERRED.md`).
- **Plant wire** runs any domain (`plantwire.Serve` takes any run; `TestEveryShippedDomainSpeaksValidWire` in
  `internal/plantwire/internal/app/review_test.go:26`). `plantwire connect --plan <session-plan> --domains-dir
  <dir>` runs a session; the session plan (`contracts/session-plan-v0.1.schema.json`) has `domain`, `seed`,
  `host_socket`, `disturbances` (planned faults with `at_s`, `duration_s`) and `pace`.
- **Kind digest**: sha256 of the canonical kind document: metrics (name, data type, resolution, writable, unit,
  range, enum values) and commands (name, argument validation, risk class, safe stop)
  (`internal/plantwire/internal/domain/digest.go:65`, `templates.go:10-60`). A channel's `observation_gain` and
  `observation_offset` are **not** in it.
- **Suite and scoring**: `suite.Publish` writes one-fault scenarios with a declared negative-class share
  (`internal/suite`); `streamsim score --run --label` scores one run against one label; `streamsim report <run-dir>`
  summarises a run folder. `simultaneous_entities` in profiles is declared but never read.
- **Each pump does not have its own normal**: every pump starts from the same initial values; only noise differs.
  Not needed for the first slice (S4 is later).

## 4. First-slice deviation (recorded per the owner's rule to improve on a plan and say so)

SCENARIOS.md S1 names `lubrication_due`, `bearing_wear` and `overload` as the repeating families. Two of them act
on commanded states and are refused over plant wire until D-116 is fixed. **The first slice uses
`bearing_wear`, `misalignment` and `sensor_drift` instead**, all repaired by `schedule_maintenance` with a
different work order each, with `transient_none` as the 40% negative class. This keeps the slice free of the
D-116 change, makes the diagnosis decide the work order (so a wrong diagnosis visibly fails), and keeps the
masking trap testable: `reduce_load` on a worn bearing still lowers vibration while the wear goes on.
`overload` and `lubrication_due` return after the slice with the D-116 fix (§6).

## 5. Rounds

Each round: write the failing test first where a defect is fixed; keep functions within 15 lines and files within
300; run `make quick` while working; one commit per round (or per improvement inside it), on `rule-promotion`.
Run `make phase-gate` once before the session ends.

### SIM-1 The pump-station world (data, size S)

Goal: a copy of the reference world in which the right maintenance repairs the fault it names.

1. Create `examples/pump-station/pump-station.domain.json`: a copy of `domains/rotating-machinery.domain.json`
   with its own id `pump-station` and title.
2. Give `schedule_maintenance` a responder: `match_arg: "work_order"`, `clear_fault: true`, a `delay_s` that
   models the job (start with 3600 s), `on_match` restoring lubrication as today, `on_miss` restoring lubrication
   only (the visit happened, the fault was not the one fixed). A work order equal to an active fault's id
   (`bearing_wear`, `misalignment`, `sensor_drift`, `sensor_failure`, `lubrication_due`) clears that fault on that
   pump.
3. This also fixes **D-121** for this world: `sensor_drift` and `sensor_failure` are cleared by a work order
   naming them, so the bias falls. Update the D-121 entry: fixed in `pump-station` through the responder; the
   pinned reference keeps the defect.
4. Tests (world or domain layer, table-driven), each proving one fact:
   - `bearing_wear` then `schedule_maintenance{work_order: bearing_wear}`: the fault is cleared and the
     `vibration` reading returns to its pre-fault level within the declared time.
   - the same fault with `work_order: misalignment`: the fault stays, vibration stays high.
   - `sensor_drift` then the matching work order: the bias clears (the D-121 regression test).
   - `bearing_wear` then `reduce_load`: vibration falls while the fault stays active (the masking trap exists).
   - `transient_none` with no action: the reading settles on its own.
5. `streamsim domain validate examples/pump-station/pump-station.domain.json` passes.

Done when: the five tests pass and the world validates.

### SIM-2 The probe variant (data, size S)

Goal: the same pumps born again with another `kind_digest` (S7).

1. Create `examples/pump-station/pump-station-probe-b.domain.json`: identical to `pump-station` except the
   `vibration` channel: another `observation_gain`, and another `range` or `resolution`, so the kind digest
   changes (gain alone does not change it). Same kind name `pump`, same instance template, so the same device ids
   are born with a new digest.
2. Test: the two worlds' `pump` kind digests differ, and the device ids are identical.

### SIM-3 The campaign (contract and code, size M)

Goal: an ordered, generic list of one-fault runs with phase labels, from which a consumer can run a campaign
without seeing any truth.

1. Contract `contracts/campaign-v0.1.schema.json` (with its embedded copy in `internal/schemas`, kept equal by the
   existing test): `schema_version`, `domain`, `seed`, and an ordered `runs[]`, each with `index`, `phase` (a
   free string such as `learn`, `holdout`, `disturb`), `domain` (to switch to the probe variant), `seed`,
   `duration_s`, and the session-plan `disturbances` for that run. **No truth in it**: no fault family names
   beyond the disturbance the session must inject (already the session plan's content, never sent to the host),
   no expected effector, no deadline. Truth stays in each run's sealed record.
2. Generator in `suite` (facade `suite.PublishCampaign`, rules in its domain layer, I/O in its edge): input is a
   small campaign spec (phases with run counts, fault families and weights, the negative-class share, the domain
   for each phase, a seed); output is `campaign.json` plus one `session-plan-v0.1` file per run. Per-run seeds
   come from the campaign seed through the existing splitmix helpers, so the campaign is a pure function of its
   spec.
3. CLI: `streamsim suite campaign --spec <file> --out <dir>` (one facade call, the output document, the failure
   mapping; no logic in `internal/cli`).
4. The first-slice campaign spec lives at `examples/pump-station/first-slice.campaign-spec.json`: learning 20 runs
   per family (`bearing_wear`, `misalignment`, `sensor_drift`) plus cavitation at 40%, holdout 10 per family plus
   cavitation, then a disturb phase of 10 runs on `pump-station-probe-b`. Each run lasts 48 simulated hours (the
   suite's default), with the fault starting in the first 12 h, so detection, the 1 h maintenance job and the
   consumer's 24 h settle window all fit inside the run.
5. Tests: the same spec gives the same files byte for byte; the negative share is within tolerance; a run's
   disturbance is valid for its domain (`plantwire connect --check` rules); no truth field appears in
   `campaign.json` (a test that reads it).

### SIM-4 Plant wire checks for the pump station (tests, size S)

Goal: prove the closed loop the consumer will use.

1. A test that runs a short `pump-station` session against the in-repository reference host: the node's birth
   carries the `pump` kind and its digest; a `DCMD` for `schedule_maintenance` with a `work_order` reaches the
   world and its responder fires; a `DCMD` for `reduce_load` changes `pump_load`.
2. A test that a second session on `pump-station-probe-b` is born with the other digest for the same device ids.
3. Capture the `pump` kind file of each world with `streamsim kind capture` from a session journal and save them
   as `examples/pump-station/kinds/pump.kind.json` and `examples/pump-station/kinds/pump-probe-b.kind.json`;
   Agentic Stream commissions its pumps with them. A test checks each file's digest equals the world's birth.
4. A short note in `documentation/` only if a public command changed (the new `suite campaign` command goes into
   `documentation/reference/cli.md`).

## 6. After the first slice (not now)

| Scenario | Simulator work |
|---|---|
| S1 full families, S5 | **D-116 fix**: a fault declares in the domain data whether it persists over later commands; the guard accepts faults that declare it; replay stays byte for byte. Brings back `overload` and `lubrication_due`. |
| S4 | Per-pump baselines: each pump its own normal (an instance-level offset or parameter), a schema change. |
| S6 | Header coupling as data (a `header` kind, links, a trip and boundary) and a **"defer to a human" truth**: `expected_response: defer` or forbidden effectors in the domain schema, ground truth and scoring. |
| S6, recurrence within a run | Several faults per run with per-fault labels and an entity and time filter in scoring; make `simultaneous_entities` real. |
| S8 | An operating-point fault (`pump_load` +0.2 on every pump, negative class). |
| S9 | A `maintenance_log` free-text channel with `attacker_controlled: true` (copy the `shift_log` pattern of `discrete-line-oee`). |
| S10 | A probe fault on `motor_current`; entity-scoped perturbations to drop one pump's data for hours. |
| S11 | Maintenance lowering a pump's normal (a `bearing_age`-like state feeding vibration). |

## 7. Rules for this session

- `AGENTS.md` is binding: logic lives in its module, the CLI only wires, one use case has one implementation,
  **never loosen a gate** (`make gate-ratchet` enforces it), no new package without a stated design, no consumer
  knowledge in the binary, domains are data.
- Commit each improvement separately on `rule-promotion`; nothing merges to main until the first-slice result
  is written (session 4).
- Better ideas are welcome: deviate from this plan when it is clearly better, and record the deviation and the
  reason in §8.
- Subagents: the main session writes; investigation and review go to Sonnet subagents; command runs, lint and
  long scripts go to Haiku subagents in the background. Lint findings are fixed without new `//nolint`.
- Comments: none beyond each package's one-line doc comment.
- Keep the inner loop fast (`make quick`); heavy checks (race, long tests) on demand only.

## 8. Status and handoff (update at the end of the session)

Deviations, each better than the plan or forced by the code:

- **SIM-1**: the responder has no `on_match` / `on_miss`. The effector's own effect already restores lubrication and clears the backlog on every call, so a second delta would double it. A work order naming no active fault therefore still restores lubrication and leaves the fault.
- **SIM-3, spec location**: `examples/pump-station/campaign/first-slice.campaign-spec.json`, not the folder root. The domain catalog loads every `.json` in a domains directory, so the spec cannot sit beside the domains.
- **SIM-3, spec content**: the spec also carries `host_socket` and optional `policies`, because a session plan requires the first and refuses every command without the second. The committed `host_socket` is a placeholder; the campaign driver sets both and generates again (the output is a pure function of the spec).
- **SIM-3, counts**: phase counts are exact (largest remainders over the fault weights, negatives by `negative_share`), then shuffled from the seed, instead of random draws. The learn phase is 100 runs: 20 per family, 40 `transient_none`. `negative_duration_s` (3600) bounds the negative-class fault; positive faults last to the end of the run.
- **SIM-3, campaign.json**: each run also names its `session_plan` file. Campaign runs are not put through the trivial-baseline audit that `suite` runs; each run is graded by `score` against its own sealed record.
- **SIM-3, gates**: additions only. `cli_wiring_test` registers `suite_campaign.go`, `facade_rules_test` registers `CampaignSpecError`, `BEHAVIOUR_PIN.txt` takes the new usage line. No cap raised, no allow-list grown, no floor lowered.

| Round | State | Commit | Notes |
|---|---|---|---|
| SIM-1 | done | cfb7f42c | `examples/pump-station/pump-station.domain.json`; five world tests in `internal/world/pump_station_test.go`; D-121 entry updated. |
| SIM-2 | done | 2b3dabc6 | `pump-station-probe-b.domain.json`: vibration range 0-60, resolution 0.05, gain 1.1. |
| SIM-3 | done | 85dfb3e4 | `streamsim suite campaign`, `campaign-v0.1`, `suite.PublishCampaign`. |
| SIM-4 | done | 529ec472 | Session tests in `internal/plantwire/internal/app/connect_pump_station_test.go`; kind files in `examples/pump-station/kinds/`. |

Hand-off to session 2 (agentic-stream):

- **Worlds and files**: `examples/pump-station/` (README there). Domains directory for every command: `examples/pump-station`.
- **Pump kind digests**: `pump-station` `sha256:17f5e82f53bc294d167f8543598ba2d49f5c60373985b78aa9ebe17288e79f19` (`kinds/pump.kind.json`); `pump-station-probe-b` `sha256:ad23e9ee18970de5f2f8d749e1e506108dd30535cd7779e80ab61766b4786c6a` (`kinds/pump-probe-b.kind.json`). Both files define the kind `pump`, so commission one at a time.
- **Devices**: `site-a:pump-1` to `site-a:pump-8` on the wire (instances `site-a/pump-N`). Commands: `schedule_maintenance {pump_id, work_order}`, `reduce_load {pump_id, target_load}`, `stop_motor {pump_id}`. `work_order` is a fault id: `bearing_wear`, `misalignment`, `sensor_drift`, `sensor_failure`. A DCMD needs `fallback_on_lease_loss` and a `policy_digest` the plan lists.
- **Generate the first-slice campaign** (160 runs, 100 learn, 50 holdout, 10 disturb on the probe variant). Set `host_socket` and `policies` in a copy of the spec first:

```text
streamsim suite campaign --spec examples/pump-station/campaign/first-slice.campaign-spec.json --domains-dir examples/pump-station --out <campaign-dir>
```

- **Run one session**: `streamsim plantwire connect --plan <campaign-dir>/run-0001.session-plan.json --domains-dir examples/pump-station --out <run-dir>`. The consumer's state persists across runs; the world starts fresh each run, so the host reconnects per run.
- **Timing**: a fault starts within the first 12 h of a 48 h run; `schedule_maintenance` takes 1 h to clear a fault, and vibration returns to its level within about 3 h more; the cavitation burst lasts 1 h and needs no action.
- **Grading**: `streamsim score --run <run-dir>/run.json --label <label>` per run, with that run's sealed record.
- **Known limits**: one fault per run (F5); `overload` and `lubrication_due` wait for the D-116 fix (§6); every pump shares one normal; the world has no header coupling.
