# S rounds review A (correctness and architecture)

Range `5951812d..0065a1e5` on `fan-sim` (S3, S6, S7, S2, S4a, S1a, S5a, S4b1, S4b2, S10, S4c). Read-only review, 2026-10-20.
Run here: `make quick` (pass), `make behaviour-pin-check` (pass), `make gate-ratchet` (pass), `go test -count=1 ./internal/plantwire/...
./test/contract/ ./test/architecture/` (pass; `test/contract` takes 34 s), coverage of the touched packages against
`scripts/coverage-floors.txt` (all at or above floor; none 2 points above, so no ratchet failure).

Verdict: no High finding. Three Medium findings, all about something the instrument does or accepts silently. Everything else is Low.

## Findings

### M1. Medium: S5 ships scripted actuator faults while D-116 says they are silently cancelled

- Evidence: `governance/DEFERRED.md` D-116 (added in this range) says a `fan_stall` or `led_stuck_off` is cancelled by the next
  command on that state while the truth keeps the fault open, and names its trigger as "before any scripted actuator fault in a
  live session". S5a is exactly that path. `internal/plantwire/internal/domain/disturbances.go:68-79` (`faultIssues`) accepts any fault
  the kind declares, and `documentation/reference/cli.md:385-392` lists the plan rules without a caveat. In a closed-loop session the
  host commands the fan, so a scripted `fan_stall` measures the wrong thing and the sealed truth disagrees with what happened.
  `fan_stall` (domain line 315) and `led_stuck_off` (line 226) are in the shipped thermal domain.
- Fix: until D-116 is fixed, either refuse in the plan check a fault whose `affects` target a state an effector assigns
  (new keyword beside `unknown_fault`, with a test), or state the limit in the `disturbances` paragraph of `cli.md` and in the S5
  line of PLAN.md. Prefer the refusal: an instrument must not accept a scenario it cannot execute faithfully.

### M2. Medium: a board mapping cannot be validated on its own and its `board_digest` is reachable by nobody

- Evidence: S-FINDINGS 2.7 puts "`kind validate` accepting `.board.json`" in S4a and PLAN S4's exit is "`kind validate` prints the
  new digests". `streamsim kind validate examples/devices/thermal-board/arduino-thermal.board.json` fails with
  `kind file fails its schema: /commands: missing required property "commands"`. `board_digest` is computed
  (`internal/plantwire/internal/domain/board_files.go:109-116`, SPEC section 15 defines it) but `BoardSummary` (`board_files.go:61-66`)
  is discarded by its only production caller (`internal/plantwire/internal/app/kind_files.go:42`, `boards, _, err :=`) and no output document
  carries it (not `implements_checked`, not the CLI). The gateway and Agentic Stream cannot get the digest the SPEC defines
  except by reimplementing RFC 8785, which G1 explicitly says not to do. The deviation is not recorded in PLAN.
- Fix: add `.board.json` handling to `kind validate` (one facade call in `internal/plantwire`, output carries `board_type`,
  `board_digest`, `kinds`), or record in PLAN that S4a dropped it and add a round; then `BoardSummary` has a caller.

### M3. Medium: a reading that does not translate is dropped silently, and the mapping is not required to translate every raw word

- Evidence: `internal/plantwire/internal/domain/board_translate.go:118-122` (`semanticValue`): a raw word absent from a metric `table`
  returns `known == false`, `app/devices.go` then does not carry the reading, nothing is journalled or counted. The projection
  check (`board_fit.go:45-53`, `tableReadingFits`) only demands that the semantic words used cover the kind's `enum_values`; it does not
  demand that every `raw_enums` word of the raw channel has a table entry. The scripts fixture `LED_KIND` (`scripts/plant-wire/fixtures.py`)
  is itself such a mapping (raw `slow_blink` has no metric row), and the shipped example only avoids the gap by hand. Same effect for
  two metrics sharing one raw channel: `SemanticReading` (`board_translate.go:108-116`) returns the alphabetically first, the other
  metric is never reported and no rule refuses it.
- Fix: add a rule (`board_table_incomplete`: every word of the raw enum has a row; `board_raw_shared`: one metric per raw reading)
  in `board_rules.go`, with fixtures under `contracts/plant-wire/fixtures/rules/`, and a SPEC section 15 sentence; or, if a partial table
  is wanted, make the drop explicit (a counted `skipped` note in the observer journal).

### L1. Low: mapped kinds ignore instance-set overrides

`app/devices.go:78-87` (`kindOf`) returns the face projected from `d.spec.Kind(effective.Name)`, the base kind. An instance set whose
`override` adds a channel or effector or narrows a range (`internal/domain/internal/domain/overrides.go:90-100`) is honoured by the
world and by `planFaults` (uses `InstanceKind`) but not by the wire: `birth()` passes the face as both `base` and `effective`. Fix: refuse
`board_type` on a kind that has an overriding instance set in `CheckInterfaces`, with a test.

### L2. Low: catch-up after a late start is not marked historical; the wall clock can step

`internal/plantwire/internal/domain/pace.go:19` (`WaitNS` returns 0 for an instant in the past) and `app/session.go:129`: every slice between
the paced start and "now" is released at once, with event times in the past and `historical` false. S-FINDINGS 1.3 and cli.md ("catch-up")
document the burst; the open point is the host's reading of it (Agentic Stream drops historical data, Q5, but cannot know these are late).
`time.Now().UnixNano()` (`uds/await.go:110`) is not monotonic, so an NTP step stalls or bursts the world. Fix: record in DEFERRED as the
S-FINDINGS risks say; consider `historical: true` for data more than one slice old.

### L3. Low: the `io_edges` reason for `uds/await.go` does not mention the pacer

`test/architecture/io_edges_test.go:67` still says "wall-time cap of the plant wire barrier and the wait it measured" while the file now also holds
`WallPacer` (`await.go:105-132`). Reusing the inventoried clock access is what PLAN required (no new gate entry), and the other gate edits are
tightening or neutral (below). Fix: widen the sentence in the same commit that owns the file; no capability change.

### L4. Low: `internal/plantwire/internal/files` is at 70.3 % against the 70 % default floor

New `files/run_folder.go` (`presentNames` error branch) is untested. A later small change fails the floor. Fix: one test with an unreadable
run folder.

### L5. Low: the canvas decides "has a gateway" from the environment string

`canvas/src/adapters/bridge/internal/connect.ts:31` computes `mixed` as `environment === "mixed" || environment === "hardware"`. The meaning
"a gateway is attached" belongs to the binary (the check document); the canvas should read a field, not a pair of words.

### L6. Low: real-time tests with a wide upper bound

`internal/plantwire/internal/app/connect_paced_test.go:65` allows `minimum+5s`; on a loaded CI box a barrier round trip plus scheduling can exceed it. The
lower bound is the real assertion; consider dropping or widening the upper one.

### L7. Low: small plan/doc drifts

- S-FINDINGS 3.2 listed `severity_invalid` and `disturbance_after_limit` as plan rules; the schema covers both (negative severity, `at_s`/`duration_s` maximum),
  which is fine, but PLAN does not record the deviation. `at_s` and `duration_s` can each be 30 days, so a clear can lie 60 days out (`ScheduleDisturbances`).
- `connect.go` loads the interfaces for the plan and `relay_sim.go` (`openSim`) loads them again, then overwrites `Node.Interfaces`: one load would do.
- `contracts/plant-wire/SPEC.md` section 14 lists 1.4.0 before 1.1.0 (the 1.1.0 row was already last).

## Checked and found sound (per round)

| Round | Result |
|---|---|
| S3 | `ambient_warm_spell` magnitude 1 and `zone_heat_load` are proven against the closed form (`test/contract/thermal_faults_test.go`); the negative class flag is tested. |
| S6 | `freshRunFolder` runs before the plan is read and writes nothing; logic in the plantwire domain (`run_folder.go`), the listing in the files edge; `sealed/` and `interrupts.rejected.*` covered. Not applied to `CheckConnect`, correct. |
| S7 | `endsHostCleanly` = `Interrupted` or `kernel.PeerGone`; any other error still fails. `kernel` facade only delegates. |
| S2 | `AnnouncedEnvironment` (pure, domain) is used by the plan check and by the merged birth; cli-output and run-journal schemas widened consistently; all-real test replays. |
| S4a | `DIGEST`, `VERSION`, `ContractVersion`, fixtures, NBIRTH fixtures, `internal/plantwire/contract/schemas` copy and doc counters all say 1.4.0 (grep: no stale 1.3.0 outside history). BOARD is closed at every fixed-member level and the lease keeps `arg` and `ms` required (`test/contract/plant_wire_board_test.go`); `board_digest` is forbidden in the file. |
| S1a | Pure schedule in `pw/domain/pace.go`, edge in `uds/await.go` reusing an inventoried clock access, port `Pacer` in the app layer, typed-nil trap avoided (`connect_run.go` guards `p.pacer != nil`). Start time is the recorded run input, replay unchanged; `replaysWhole` is asserted for node and mixed. The barrier stays a cap (three-miss test). Cancel during the wait ends as a stop. |
| S5a | Schedule is pure (`disturbances.go`), injected through `run.InjectFault/ClearFault` so the command log and sealed journal carry it and replay needs no plan. Test asserts the fault name is absent from the observer journal, the trace, every record the host received, and the check document. |
| S4b1/S4b2 | Domains stay raw; the command log records the raw command and raw arguments (`connect_mapped_test.go`), so replay and the analytic oracle are untouched. Translation is deterministic (sorted keys). A raw command name from the host is `unknown_command`, an out-of-range or unknown word is `out_of_range`, a safe stop translates through the same path. No Go production file names a thermal command (grep). |
| S10 | Reference host beats on a wall ticker owned by the `uds` edge; `Peer.Beat` is pure state; relay watchdog is suspended while `!r.host.shaken` and re-armed at the first host `STATE`, with tests for both absences and for forwarded heartbeats not refreshing it. |
| S4c | Canvas sends `pace: "wall"`, the bridge accepts `kinds/<type>.board.json` only in `kinds/`, docs and briefs use the semantic names (grep of `documentation/` and the briefs: raw names only where the board's firmware is meant). |

## Gate-file changes (`git diff 5951812d..0065a1e5 -- test/architecture scripts`)

| File | Change | Verdict |
|---|---|---|
| `scripts/coverage-floors.txt` | `plantwire/internal/uds` 93 to 95 | Floor rose: allowed. |
| `test/architecture/io_edges_test.go` | `uds/listen.go`: `call:time.Now` removed | Allow-list shrank: allowed. |
| `test/architecture/io_edges_test.go` | `uds/ticks.go`: reason text widened (reference host heartbeats) | Text only; the capability (`time.NewTicker`) is unchanged. Allowed. |
| `scripts/plant-wire/fixtures.py`, `schemas.py` | BOARD schema, fixtures, rules fixtures, two fixture fixes | Contract generators, consistent with the committed files; not a gate loosening. |

No cap raised, no allow-list grown, no coverage floor lowered, no test hook added. See L3 for the one stale reason text.
