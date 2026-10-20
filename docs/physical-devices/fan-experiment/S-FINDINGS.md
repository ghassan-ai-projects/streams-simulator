# Fan experiment: streamsim rounds S1, S4, S5 (investigation, 2026-10-20)

Read-only investigation on branch `fan-sim` (80b16d23). Nothing was edited, built or run except `bin/streams-simulator kind validate -h`.
Every claim below was read in code at the cited place; section 5 lists what was not verified. Paths are relative to the repo root;
`pw/app`, `pw/domain`, `pw/uds` mean `internal/plantwire/internal/{app,domain,uds}`.

## 1. S1: paced live sessions (closes D-71 for `plantwire connect`)

### 1.1 Where time lives today

| Fact | Evidence |
|---|---|
| A connect session builds its run in `TimeStepped` from `DefaultStartTimeNS` (2026-01-01) | `pw/app/connect_run.go:64-66` (`newRun`); mixed path `pw/app/relay_sim.go:67-68` (`newSimRun`) |
| The node's end time is `DefaultStartTimeNS + openEndedNS` (100 years) | `connect_run.go:14,147` |
| One slice per loop turn, `at += SliceNS`, no wall reference | `pw/app/session.go:122-126` (`Session.Run`), `slices.go:12-25` (`slice`: Advance, data, `NSYNC`, barrier) |
| The barrier waits at most the cap, then records a miss and advances; three consecutive misses end the world | `slices.go:36-46`, `:81-90`, const `:10`; wall wait in `pw/uds/await.go:35-48` and (mixed) `pw/uds/sim.go:47-52` |
| The wall time waited is journalled as `pacing` (observer session record, outside the stream digest) | `slices.go:38` `JournalPacing`; `internal/journal/internal/domain/record.go:51,69` |
| Start time is already a recorded run input: `run.json` `world_config.start_time_ns` | `internal/run/internal/app/artifact_metadata.go:65`; replay reads it back, `replay.go:134-135` |
| `TimeWall` makes a run non-reproducible; `TimeStepped` stays reproducible | `internal/run/internal/app/initialize.go:126` |
| Only declared edge files may read the clock or wait (inventory `ioEdges`, one entry per file, with a reason) | `test/architecture/io_edges_test.go:30-72`; `uds/await.go`, `uds/sim.go`, `uds/ticks.go` are listed. `internal/wall` is a seam but the plantwire edges call `time.Now` directly (lines 66-72); the suite runs with `-tags simdet` (`Makefile:187`), under which `wall.Now()` is the zero time |
| The seconds-since-start model has no UTC dependence (seasonality uses `elapsed`) | `internal/world/internal/domain/integration.go:127-135`; no hard-coded 2026 in `canvas/src` |
| `NBIRTH.node_metrics.clock_mode` already admits `wall` (schema enum `virtual`, `wall`) but the node always sends `virtual` | `contracts/plant-wire/schemas/NBIRTH.schema.json:264-269`; `pw/domain/births.go:16,32`; `NodeIdentity.ClockMode` is never set (`pw/app/host_records.go:95`) |

Consequence: **the recorded start is already an input**; S1 changes only the value passed to `run.Config.StartTimeNS` (plus
`StartTimeSet: true`) and when slices are released. Replay and verify need no change: `replayConfig` reuses the recorded start and
replays the command log (`clock.advance` entries); nothing in replay reads a wall clock. A paced run is therefore reproducible
exactly as a stepped one: *paced is still stepped in virtual time, only released by wall time*.

### 1.2 Design

**Option name and place: a session-plan member `"pace": "wall"`** (absent = today's unpaced behaviour). No `connect` flag.
Reasons: the plan is the one recorded session input (policies were added the same way, `CHANGELOG.md:26-31`, in place in
`session-plan-v0.1`); the canvas bridge launches `plantwire connect --stdin --barrier-cap-ms 10000` and passes only the plan
(`internal/canvasbridge/internal/domain/session.go:38`); PLAN R1 wants simulate and real plans to differ only in
`gateway_socket` and `real`. `plantwire node` and the gateway-less `relay --domain` keep their flags and stay unpaced
(D-71 stays open for them; say so in DEFERRED).

Pieces (each file below 300 lines, each function below 16):

1. `model.SessionPlan.Pace string` + schema enum `["wall"]` (`internal/model/session_plan.go:9-17`, `contracts/session-plan-v0.1.schema.json`, its copy `internal/schemas/`, kept equal by `internal/schemas/schemas_test.go:36`).
2. **Schedule (pure)**: `pw/domain/pace.go`: `WallAt(startWorldNS, startWallNS, worldNS) int64 = startWallNS + (worldNS - startWorldNS)` and the lag. Absolute mapping, not cumulative sleeps, so jitter never accumulates and `world time == UTC` holds for the whole session.
3. **Edge**: `pw/uds/pace.go`, `WallPacer{startWorld, startWall}`; `Start() int64` reads `time.Now()` (floor to the millisecond); `Release(ctx, worldNS) (lag time.Duration, err error)` waits on a `time.Timer` or `ctx.Done()`; no wait when already late (catch-up).
4. **Port in the app layer**: `pw/app/session.go` gets `type Pacer interface{ Release(ctx, int64) (time.Duration, error) }` and `Options.Pace Pacer` (nil = unpaced). `Session.Run` calls `s.release(ctx, at)` before `s.slice(ctx, at)` (`session.go:122-124`), so readings with event time up to `at` leave at wall time ~`at`, and `NSYNC.world_time_ns` equals UTC. The fake in tests implements the interface; no test-only symbol is added to production code (so `scripts/test-hooks.txt` is untouched).
5. **Wiring** (`connect_run.go`, `relay_sim.go`): when `plan.Pace == "wall"`, `Connect` creates the pacer first, takes `StartNS`, and passes it to `newRun` / `SimulatedOptions` (new `StartNS` field replacing the `DefaultStartTimeNS` literals at `connect_run.go:66,147` and `relay_sim.go:68`); `nodeOptions` gets `UntilNS: start+openEndedNS`; `Options.Pace` set for both the node (`connectNode`) and the world inside the relay (`SimulatedOptions.Node`). Unpaced path keeps `DefaultStartTimeNS`.
6. **Signal to the host**: `NodeIdentity.ClockMode = "wall"` when paced (`host_records.go:95`). No schema change. Whether Agentic Stream should key its UTC handling on it is the host module's call (UNVERIFIED, section 5).
7. Barrier stays a cap: `awaitSync`, `countBarrier` and the three-miss rule are untouched. Pacing happens before `Advance`, the cap starts after `NSYNC`.

Optional, small: add `lag_ms` to the `pacing` record (`internal/journal/internal/domain/observer.go:82-85`; it is a session record outside the digest, but `run-journal-v0.1` is a contract and gets an additive member).

### 1.3 Interactions

- **Interrupts (hold/approve/e-stop)**: controls are taken only at a barrier (`interrupt_side.go:23-37`, D-81). Paced, the worst-case latency becomes one slice (1 s by default; `--slice-s 0.5` halves it) instead of milliseconds. Gate decisions use the world clock (`domain.ClockWorld`), which is now UTC-aligned, so the canvas countdown of a waiting command (D-85: counted from the newest snapshot) moves once per second instead of in bursts. The e-stop that matters physically is the gateway's, not the simulated node's.
- **Live history**: the snapshot is written once per barrier (`R10 as-built`), capped at 2000 snapshots in the canvas store; unpaced that trims within seconds, paced it holds about 33 minutes. A benefit, not a risk.
- **Mixed relay**: the world run is created at relay open (`relay_sim.go:41-55`), but its session starts at the first host `STATE` (`relay_sim.go:108-112,151-164`). With the absolute mapping, any delay before the host appears becomes a burst of catch-up slices at host speed (not paced, still deterministic). With `connect --plan` the host is normally already listening, so the lag is small; if it is not, the first readings arrive with old event times. Record the lag in `pacing`; consider `historical: true` for catch-up data later (SPEC section 7 already defines it).
- **Run span**: `MaxRunSpanNS` is 30 days (`internal/run/internal/domain/span.go:5`); a paced session longer than that ends with `ErrRunSpanExceeded`. Acceptable; say it in the docs.

### 1.4 Tests

| Test | Layer |
|---|---|
| `WallAt` and lag arithmetic, table-driven (before, on, after the schedule; overflow-free for 2026 starts) | `pw/domain` |
| Fake pacer: `Release` is called once per slice, in order, before the data of that slice; a cancelled context during the wait ends the session with the interrupt (`Stopped`), not an error | `pw/app` (pattern: `session(t, host, n)` in `barrier_epoch_test.go`) |
| Paced plus a silent host still ends after three misses (cap is a cap, pacing is not) | `pw/app`, beside `barrier_epoch_test.go:20-27` |
| Real `WallPacer` with 20 to 50 ms slices: total time at least n*slice, each slice on the absolute schedule, no wait when started in the past | `pw/uds` (small real timers, within the 2-minute rule) |
| `Connect` with `pace: wall` and 20 ms slices against `serveWireHost`: `world_config.start_time_ns` lies between wall readings taken around the call and is not `DefaultStartTimeNS`; `replaysWhole(t, out, ...)` matches (`connect_test.go:147-155`); `NBIRTH` journalled with `clock_mode: wall`; the same for the mixed path with the stand-in gateway | `pw/app` |
| Plan schema: `pace: "fast"` refused with a pointer; `pace` absent unchanged | `internal/model/session_plan_test.go` |
| Interrupt replay under pacing (reuse `interrupt_replay_test.go`) | `pw/app` |
| `ioEdges` entry for `pw/uds/pace.go` (`call:time.Now`, `call:time.NewTimer`) | `test/architecture/io_edges_test.go` |

Gate note for the PR list (AGENTS rule 3 asks every gate change to be listed): this adds one `ioEdges` entry for a new edge file.
The gate-ratchet does not cover `ioEdges` (no match in `test/architecture/gate_ratchet*.go`) and the failing message of
`TestIOStaysInDeclaredEdges` prescribes exactly this edit, but the owner should see it named.

### 1.5 Risks

- Catch-up bursts after a stall (up to 3 x barrier cap = 90 s at the default, 30 s from the bridge: `--barrier-cap-ms 10000`). Deterministic, but the host sees old event times.
- World lag versus UTC at start in relay mode (above). A rebase would need the run to be created after the first host `STATE`, which the birth path does not allow (`simBirth` needs the session, which needs the run: `relay_sim.go:93-106`). Not proposed.
- Timer drift is excluded by the absolute schedule; clock steps (NTP) are not: a backward step stalls the world, a forward step bursts. Nothing is done.
- A paced test suite must never become slow: only the `pw/uds` pacer test and one connect test use real time.
- Size: M. S1a (pacer, port, plan member, wiring, tests) is one round; S1b (docs: `documentation/reference/cli.md`, `contracts.md`, guide, DEFERRED D-71 narrowed) is a small second one.

## 2. S4: semantic kind commands and the board mapping (D7)

### 2.1 What the code allows today

| Question | Answer, with evidence |
|---|---|
| What does `KIND` allow? | `contracts/plant-wire/schemas/KIND.schema.json`: metric `data_type` in number, integer, boolean, enum, string; `enum_values` (max 64, 64 chars each), `unit`, `eu_range`, `resolution`, `writable`; command `args_schema` is any JSON object, so string enums are already legal arguments (`"enum": [...]`). **No member declares a lease**; `commands[]` has no `additionalProperties: false`, but see next row. |
| Who rejects unknown members? | The twin scaffolder only: `internal/component/internal/domain/twin_kindfile.go:16-19` closes the member lists, `twin_requirements.go:29,74,132`; it also refuses enum and string metrics (`:97`, `twin_type_unsupported`) and needs numeric `sets` arguments (`twin_commands.go:59`). `kind validate --twin` runs these rules (`bin/streams-simulator kind validate -h`). The plain validator and the digest accept any member (the digest covers the canonical entry). |
| How is a kind file proven against a domain? | `pw/domain/implements.go:67-95`: the domain kind's own template `KindTemplates(kind)` (`pw/domain/templates.go:10-72`: channels become metrics, effectors become commands with `validationOnly(args_schema)`) must equal the file's canonical tree; any difference is `implements_drift` at the first differing path. The announced `NBIRTH.kinds` come from the same templates (`births.go:19-23`, `host_records.go:94-97`, `devices.go:69`). |
| How does the thermal domain do the fan today? | Effector `set_pwm_lease {duty_permille 0..600, lease_ms 5000..10000}` assigns state `fan_duty` (`assign_from_arg`), `lease: {arg: lease_ms, unit: ms, fallback: safe_off, safe_deltas}`; the zone follows through a first-order `cools` drive (gain -0.02 per permille, 180 s; `thermal-board.twin.json`). The LED has two raw channels (`brightness_permille`, `pattern` as integer code 0..2) and `set_led {brightness_permille, pattern}`. (`examples/devices/thermal-board/thermal-board.domain.json`, `fan.kind.json`, `indicator.kind.json`.) |
| Can the world hold a word-valued state? | No. Effects only assign numbers (`internal/world/internal/domain/effects.go:62-72`), enum channels are random picks of `enum_values` (`reading_values.go:75-80`), a lease is read from an argument (`follow_up.go:120-126`, `model.Lease{Arg,...}` in `internal/model/vocabulary.go:60`). |
| Does production code name the raw commands? | No Go production file names `set_pwm_lease`, `set_led`, `duty_permille` or `brightness_permille`; only data, tests, fixtures and docs do (grep). The consumer-knowledge gate stays satisfied. |
| How does a lease reach the wire? | `DCMD.fallback_on_lease_loss` is required; the node does nothing with `lease_ms` (grep: no production use in `internal/plantwire`); the world applies it. SPEC section 9: a command whose arguments carry `lease_ms` falls back when it lapses. |

### 2.2 Two ways to build it; recommendation

| | A. Translate at the node boundary (recommended) | B. Make the twin domain itself semantic |
|---|---|---|
| Domain | stays raw (`duty`, `set_pwm_lease`, integer pattern); hand-written file unchanged in shape | states in percent, a word state, a fixed lease |
| New code | projection of a raw `model.Kind` to its semantic view (pure, `pw/domain`), a value/command translator (pure), three call sites in `pw/app` | world: word args and word-valued channels; model/schema: fixed-duration lease; scaffolder: a `mapped` response |
| Mapping used at run time by the twin | **yes**: the twin exercises the same file the gateway reads, which is D7's reason ("the twin and the real board cannot drift") | no: only at composition |
| Blast radius | node, twin-report, implements check, examples | world core, domain schema (digest of every domain), behaviour pin, scaffolder |

Recommended: **A**. The projection is the key simplification: a function `Project(rawKind, mapping) *model.Kind` returns a
`model.Kind` whose channels, effectors, ranges and args schema are the semantic ones; births, aliases, the live projection,
`KindTemplates` and the implements check then work unchanged on the projected kind. Only values and commands need translation.

### 2.3 The board mapping file (shape)

One file per board type, `<board>.board.json`, found in the same `--kinds-dir` directories (suffix beside `.kind.json`,
`pw/domain/kind_files.go:11`); schema `contracts/plant-wire/schemas/BOARD.schema.json` next to `KIND`. Three closed forms
only (`value`, linear `from`+`scale`, `table`); no expression language (AGENTS "Forbidden changes").

```json
{
  "board": "arduino-thermal",
  "kinds": {
    "fan": {
      "commands": {
        "set_cooling": {"raw": "set_pwm_lease", "args": {"duty_permille": {"from": "level_percent", "scale": 10}, "lease_ms": {"value": 5000}}},
        "stop_fan": {"raw": "stop_fan"}
      },
      "metrics": {"level_percent": {"raw": "duty_permille", "scale": 0.1}}
    },
    "indicator": {
      "commands": {
        "show_state": {"raw": "set_led", "by": "state", "table": {
          "off": {"brightness_permille": 0, "pattern": "off"},
          "watch": {"brightness_permille": 300, "pattern": "slow_blink"},
          "alert": {"brightness_permille": 1000, "pattern": "solid"}}},
        "stop_led": {"raw": "stop_led"}
      },
      "metrics": {"state": {"raw": "pattern", "table": {"off": "off", "slow_blink": "watch", "solid": "alert"}}}
    }
  },
  "raw_enums": {"pattern": ["off", "slow_blink", "solid"]}
}
```

- `raw_enums` gives the twin the code of a firmware string (index in the list), because the twin's raw `pattern` is the integer 0..2 while the firmware says `slow_blink`. This is the one place both spellings meet (UNVERIFIED as the best form; see 5).
- **Lease**: kind side, a new optional command member in `KIND` (for example `lease_ms: {min, max}`, 5000 and 10000 for `set_cooling`): the host reads it from `NBIRTH.kinds` (needed for renewal, D6). Board side, the mapping's `lease_ms` value must lie inside it (`mapping_lease_outside_kind`). This sits uneasily with the plan's own test ("would it mean the same on another board?": a lease is board-specific); the plan resolves it as "declared by the kind, satisfied by the board", and the check enforces that.
- **Safe stops** stay raw names mapped one to one (`stop_fan`, `stop_led`); the gateway maps them to the firmware's `safe_stop` as the plan already says.
- **Validation** (new `pw/domain/board_*.go`, called by a `board`/`kind validate` path in `pw/app/kind_files.go`): every kind-file command and metric has an entry and only those; `table` keys equal the kind file's `enum_values` / argument `enum`; `from` names an argument of the kind command and `scale` maps its range into the raw range; raw names exist in the domain kind (and, for the gateway, in the firmware catalog, which is out of this repo); the lease rule above; duplicate `kinds` across board files are refused like duplicate kind files (`ErrKindFileDuplicate`).

### 2.4 Where the twin applies it

Domain stays raw; mapping at the wire boundary, in three places:

1. **Kind view** (`pw/app/devices.go:69` `kinds()` and `:80-103` `observe`/`carried`): `d.spec.InstanceKind(...)` becomes the projected kind for any kind with a mapping; `KindTemplates(projected)` equals the kind file, so `NBIRTH` announces the file's own `kind_digest` (this also makes the R9 binding check, `kind_mismatch`, compare like with like for real versus simulated).
2. **Commands** (`pw/app/execution.go:25-28` `invoke`): `set_cooling {level_percent: 45}` becomes `set_pwm_lease {duty_permille: 450, lease_ms: 5000}` before `run.InvokeEffector`; the idempotency key and the ledger keep the *semantic* content (the host's view). `out_of_range` for the semantic argument is checked against the kind file's schema, raw bounds by the world as today.
3. **Readings** (`devices.go:110` `wire.Sample`, and `currentValues` in `pw/domain/births.go:46-60` for `DBIRTH`): raw `duty_permille` 450 is sent as `level_percent` 45; raw `pattern` 1 as `state: "watch"`; the raw `brightness_permille` channel is not carried (it has no semantic metric). Aliases follow the projected kind's metric order.
4. **Twin report** (`pw/app/twin_world.go:58-75`): `Apply` translates the command, `Predict` translates the metric back, so the report compares the host-visible quantities with the board trace.
5. **Implements check** (`pw/domain/implements.go:67-95`): compare the file with `KindTemplates(Project(kind, mapping))`; a missing mapping entry gives `implements_drift`/`implements_missing` at the first path, as today.

Twin internals stay raw: the 45 % fan behaves as 450 permille did by construction (same effector, same drive), which is the S4 exit.

### 2.5 Contract and fixture changes (1.4.0)

SPEC section 11 says a schema change needs a reviewed `DIGEST` change, and the digest is over **every file** in `schemas/` in name
order (`pw/domain/digest.go:18-38`), so adding `BOARD.schema.json` changes it by itself. Procedure (`scripts/plant-wire/README.md`):

1. Edit `scripts/plant-wire/schemas.py` (add `BOARD`, add the command `lease_ms` member to `KIND`), run it into `contracts/plant-wire/schemas`, then `fixtures.py ... | cut -d' ' -f1 > contracts/plant-wire/DIGEST`; copy the schemas into `internal/plantwire/contract/schemas/` (embedded copy, equality tested in `internal/plantwire/contract/contract_test.go`).
2. Version: `const ContractVersion = "1.3.0"` (`pw/domain/births.go:9`); `fixtures.py` writes `"1.3.0"` into the NBIRTH fixtures (all of `contracts/plant-wire/fixtures/**/NBIRTH*.json`); `internal/plantwire/environment_test.go:25`; the SPEC header and section 14 history row; `documentation/reference/contracts.md:30`, `documentation/guides/physical-devices.md:46,245`, `documentation/architecture/physical-devices.md:53`.
3. New fixtures: valid and invalid `BOARD` files (and a kind fixture with a lease); the guard tests glob `*.schema.json` (`test/contract/plant_wire_helpers_test.go:36-39`) so the new schema is compiled automatically; check that no test assumes "every schema is a record type" (UNVERIFIED).
4. `capability_digest` changes with any command change (it covers `{kind, commands}`, SPEC section 5), so every recorded NBIRTH digest in tests moves.
5. Sync: this repository holds the master copy (`Makefile:268-274`, `.env.example`); Agentic Stream and the gateway vendor the folder byte-identically, per D5 all three together.
6. Semantic kind files (`examples/devices/thermal-board/{fan,indicator}.kind.json`): `fan` gets metric `level_percent` (integer, unit `%`, 0..60, writable) and `set_cooling {level_percent}` plus `stop_fan`; `indicator` gets enum metric `state` (`off`, `watch`, `alert`) and `show_state {state}` plus `stop_led`. `kind validate --twin` will refuse the enum metric (`twin_type_unsupported`) and the `lease_ms` member (`twin_member_unsupported`); see the scaffolder question in 5.

### 2.6 Everything that names the raw commands (to migrate or leave)

- **Data**: `examples/devices/thermal-board/` (`thermal-board.domain.json` 13 hits, `thermal-board.twin.json` 6, both kind files 4 each; the domain is being edited by another agent, keep it raw); `domains/cold-chain-transit.domain.json` (7 hits) is an unrelated domain with its own `set_led`/`duty_permille`: leave.
- **Go tests that run the thermal kinds or domain** (update expected commands/metrics): `test/contract/thermal_board_test.go`, `thermal_twin_test.go` (8 hits), `thermal_actuator_faults_test.go`; `internal/plantwire/thermal_board_test.go`; `pw/app/`: `twin_world_test.go`, `twin_fixture_test.go`, `twin_relay_test.go`, `twin_review_test.go`, `twin_refusal_test.go`, `relay_mixed_test.go`, `relay_chain_test.go`, `interrupt_session_test.go`, `interrupt_review_test.go`, `policies_test.go`; `pw/standin/{commands,play,bench,config}_test.go`; `pw/domain/twin_{classify,evaluate,tolerance,fake,node_stop}_test.go`; `internal/cli/internal/app/{plantwire_twin,kind_twin}_test.go`; `internal/component/twin_suggest_test.go`; `internal/model/twin_report_test.go`.
- **Left alone on purpose** (firmware side, device-wire-v1, which the gateway targets): `internal/device/**`, `internal/device/testdata/thermal_capability_catalog.json`, `internal/deviceworld/**`, `test/acceptance/testdata/device-serve-transcript.jsonl`. These are the raw vocabulary the mapping points at.
- **Canvas** (all tests and fixtures, no production source names them): `canvas/e2e/{policies,interrupts-layout,twin-report,devices}.spec.ts`, `canvas/e2e/support/interruptBench.ts` (these drive the real binary with `set_pwm_lease`, e.g. `policies.spec.ts:27,49`, so they change); unit fixtures `canvas/src/test/{liveFixture,benchLiveFixture,interruptFixture}.ts`, `canvas/src/adapters/bridge/internal/{interruptDecoder,twinReportDecoder}.test.ts`, `canvas/src/model/review/internal/{interruptMarks,interruptWords}.test.ts`, `canvas/src/ui/review/internal/Interrupts.test.tsx` (string fixtures; update for realism only).
- **Docs and briefs**: `docs/physical-devices/briefs/{gateway-plant-wire,bench-instrument,bench-checks-m2-m5}.md`, `docs/design/PHYSICAL_DEVICES.md`, `docs/simulation-canvas/{plant-wire-protocol,impact-agentic-stream}.md`, `documentation/reference/cli.md`.

### 2.7 Rounds

| Round | Content | Size |
|---|---|---|
| S4a | Contract 1.4.0 only: `BOARD` schema, `KIND` lease member, fixtures, SPEC, `DIGEST`, `ContractVersion`, doc counters, plus the pure board-file parse/validate (`board_*.go`, no node behaviour) and `kind validate` accepting `.board.json`. Example kind files untouched, so the repo stays green. **Unblocks Agentic Stream A0 and gateway G1 to vendor 1.4.0 early.** | M |
| S4b | Projection, translator, implements-through-mapping, node boundary (`kinds()`, `invoke`, `carried`, births), semantic example kind files and `arduino-thermal.board.json`, migration of the Go tests above, `kind validate`/`plantwire` flags for the mapping | L (two commits: projection plus tests on a tiny fixture domain first, then the thermal migration) |
| S4c | Twin report and canvas e2e migration, standin, docs/briefs, the scaffolder decision (below) | M |

## 3. S5: scripted disturbances in a live session

### 3.1 Findings

- Offline faults go through the scenario script (`internal/cli/internal/app/run_script.go:70`, `model.ScenarioCommand{Op: fault.inject}`) and the director tools into `Run.InjectFault(entity, fault, onsetNS, params)` / `Run.ClearFault(faultID, atNS)` (`internal/run/internal/app/commands.go:13-41`). Each call is appended to the command log (`recordWorldCommand`, ops `fault.inject`/`fault.clear`, `internal/model/run.go:79-80`) and the sealed truth journal (`journalFaultTruth`); replay re-executes them (`internal/run/internal/app/command_executor.go:33,50-57`).
- The world accepts an onset in the future and a clear in the future: `admitOnset`/`admitClear` refuse only times before the clock of a world that already ran (`internal/world/internal/domain/faults.go:165-185`); `schedule(kindObservationCapture, ..., onsetNS)` and `state.go:39` (`t < onsetNS`) honour it. A fault's effect is a function of world time, so **pre-injecting every disturbance at session start with absolute onset times is deterministic and fires at the right slice in a live session and on replay**; no slice-boundary hook is needed. Constraint: the same call needs a world that has not run past the onset (true at session start).
- `severity` is the only fault parameter (`faults.go:59-65`); `zone_heat_load` (step, +8 on `heat_load`) and `dht11_stuck`, `ambient_warm_spell`, `fan_stall`, `led_stuck_off` exist in the current thermal domain. A fault must be declared by the instance's kind (`faults.go:36`), else `unknown fault`. `duration_s` exists only for pulse faults (`FaultOnset.DurationS`, `internal/model/domain.go:200-206`); for a step fault the end is a `ClearFault` at `onset+duration`.
- The session plan has no scenario member (`contracts/session-plan-v0.1.schema.json`); `model.DecodeSessionPlan` validates against the embedded copy (`internal/model/session_plan.go:28-38`); rules that need the domain live in `pw/domain/connect_plan.go` (`CheckPlan(plan, []PlanInstance)`, `PlanInstance{ID, Kind, Implements, KindDigest}`), fed by `planInstances` (`pw/app/connect.go:105-116`). Bound (real) instances are known there (`planCheck.instances`, `.bound`).
- The canvas reads the plan only as the checked document in the `session` journal record (`sessionNote`, `connect_run.go:73-84`: domain, environment, instances with origin, implements_checked, policies) and builds the plan from its form with a constant `schema_version` (`canvas/src/model/launch/internal/session.ts:79`). It reads nothing about faults; a hand-written plan file driven from the CLI (S8) carries them.
- Sealed oracle: the observer journal and the wire must not learn the schedule. The `session` record is an observer record, so the schedule must not be echoed into it.

### 3.2 Design

Schema (in place in `session-plan-v0.1`, as `policies` was; PLAN says "v0.2", renaming costs the file, `schemas.go:77`, `schemas_test.go:36`, `ContractSessionPlan`, the canvas constant and seven doc mentions for no gain while pre-release; record the deviation in PLAN):

```json
"disturbances": {"type": "array", "maxItems": 64, "items": {"type": "object", "additionalProperties": false,
  "required": ["at_s", "instance", "fault"],
  "properties": {"at_s": {"type": "number", "minimum": 0, "maximum": 2592000},
                 "instance": {"pattern": "same as real[].instance"}, "fault": {"pattern": "an id"},
                 "duration_s": {"type": "number", "exclusiveMinimum": 0, "maximum": 2592000},
                 "severity": {"type": "number", "minimum": 0}}}}
```

`at_s` is seconds after world start (so +120 s is relative to the session start whether paced or not). Entries are applied in array order for equal times.

Validation (`pw/domain/connect_plan.go`, `PlanInstance` gains `Faults []string` from `kind.FaultDecl`; pointers `/disturbances/N/...`): `unknown_instance` (reused), `unknown_fault` (kind of that instance does not declare it), `disturbance_on_real_instance` (the instance is in `real`: its simulated twin is only a shadow), `duration_with_pulse_fault` (the domain fixes the length), `severity_invalid` (reuse `World.CheckFaultParams`, `internal/world/operations.go:121`; needs a world or a spec-level function: UNVERIFIED how cheaply the plan check can get one; `run/app/scenario.go:96` does it with a created world), `disturbance_after_limit` (`at_s` beyond `MaxRunSpanNS`).

Scheduling: `pw/domain/disturbances.go` (pure) turns the plan list and the world start into ordered `{entity, fault, onsetNS, clearNS?, params}`; `Options.Disturbances` carries it; `Session.Run` calls `s.scheduleDisturbances()` right after the handshake (`session.go:119-121`), before the first slice, via `s.run.InjectFault` and `s.run.ClearFault`. Both start paths use `Session.Run`, so node and mixed relay behave the same; the entity is the plant instance id directly (`devices.go` maps wire ids separately).

Recording: nothing new. The absolute `fault.inject`/`fault.clear` entries are in `run.json` `command_log` and in the sealed truth journal; replay does not need the plan. The check document does **not** echo the schedule (it would leak into the observer `session` record and would need `cli-output` and `run-journal` contract edits); the operator has the plan file.

Tests: schema table (valid, each bad pointer); `CheckPlan` per keyword with pointers (pattern: `connect_policies_test.go:188`); a node session against a scripted host with a fake clock-free slice loop: the zone reading is unchanged before the onset slice and rises after it, the fault is absent from every wire record and from `NBIRTH`/`DBIRTH` (SPEC section 5 "what a birth never says"); `run.json` holds `fault.inject` with `onset_ns = start + at_s`; `replaysWhole` matches; the same plan twice gives byte-identical wire traces; real-bound and unknown-fault plans are refused; combined with the fake pacer of S1. Size: M (S5a plan schema, validation and scheduling with tests; S5b docs and the S8 plan files).

## 4. Recommended order, sizes, risks

1. **S4a first (M)**: the contract is what Agentic Stream A0 and the gateway G1 vendor; changing it after they start is rework. Data and validators only, the example kinds stay raw, so no other streamsim test moves. Branch from `fan-sim`; the thermal domain file is untouched.
2. **S1a (M)**, then S1b docs (S): independent of S4; needed for P4 (UTC alignment). Land the plan member `pace` here.
3. **S5a (M)**: depends on S1 only for the combined test; plan schema is touched twice, so keep S1 and S5 sequential, not parallel.
4. **S4b (L), S4c (M)**: biggest blast radius; do after the other repositories have the contract. The simulate rung (P4) needs S1, S4b and S5.
5. S2, S3, S6 to S9 as in PLAN section 5 (not part of this investigation).

Per-round gate list to state in each PR: S1 adds one `ioEdges` entry (new edge file); S4 adds no gate change if the new files respect the 300-line and 15-line limits and the board code sits in `pw/domain` (pure) and `pw/app`; S5 none. No cap raised, no allow-list grown. The cross-repo contract copy is vendored, so Agentic Stream and the gateway need the sync in the same milestone (D5).

Main risks: (a) S4 touches every path that assumes announced kind equals domain kind; the projection confines it, but the number of touched tests is large (about 30 Go files and 9 canvas files). (b) Catch-up bursts and start lag in paced relay mode (1.3). (c) A mapping form that grows (scale, table, then arithmetic) violates "no expression language"; keep three forms and refuse the rest in the schema. (d) Disturbances are sealed truth: any later request to show them in the canvas needs a reveal design, not an echo.

## 5. UNVERIFIED and open questions

- Nothing was executed beyond reading and `kind validate -h`; no test was run. Line numbers are from the working tree at 80b16d23 plus the untracked runbook.
- How Agentic Stream should use `node_metrics.clock_mode: wall` (or whether it needs it): its host module decides.
- Whether any contract test assumes one record type per schema file (affects adding `BOARD.schema.json`): `plant_wire_helpers_test.go` only compiles by glob, `plant_wire_test.go` and the guard tests were not read in full.
- **Scaffolder fate** (needs the owner): `thermal-board.twin.json` is scaffolded from the kind files into `arduino-thermal-twin` and pinned by `test/contract/thermal_twin_test.go` (known-document test at `:165`). With enum metrics, a lease member and semantic commands, the scaffolder (`component/internal/domain/twin_*.go`) refuses the new kind files. Options: retire that scaffold test and keep the hand-written raw domain (cheapest, recommended for the experiment), or later generate the raw domain from the board mapping plus choices (the mapping carries the raw interface), which is a further round not planned here.
- Mapping selection: by "kind file name" across all loaded `.board.json` files (no domain or plan change). If two boards of different types are ever loaded for one kind, a plan or domain member is needed.
- `raw_enums` as the home of the firmware string to twin code table (pattern); alternatives are a `code` beside each table entry or turning the twin's raw `pattern` into a string (the world cannot assign strings today).
- Whether the lease belongs in the kind (plan D7 text) or only in the mapping (the plan's own value test); this file follows the plan and adds a consistency check.
- How a connected run's sealed truth is revealed afterwards for scoring or review (S5 recording relies on the command log and the sealed journal being written as in offline runs).
- Whether `World.CheckFaultParams` can be reached in the plan check without building a world.
- Behaviour of the host (Agentic Stream) on catch-up readings with old `event_time_ns`; not testable here.
