# Belt conveyor stories

The stories of the multi-drive belt conveyor (`docs/belt-conveyor-drives/PLAN.md` section 2.3) as data, for the
plants `belt-conveyor-drives-a` (drives with margin) and `belt-conveyor-drives-b` (without), and for S6 the
master-follower variants of plant A, `belt-conveyor-drives-s6-stop` and `-s6-promote`.

- `stories.json`: one entry per story variant (S0-A, S1-A, S1-B, S2-A, S2-B, S3-A, S4-A, S4-B, S5-A, S5b-A, S6-A-stop,
  S6-A-promote). Each has its plant, horizon, disturbances, ten seed rows and a hand-written reference policy; an S6
  story names its `regime` (`a`) beside its plant, and `unavoidable: true` marks S6-A-stop, which no reaction saves: its
  reference policy (feed 0, no restart) only does no harm and ties doing nothing.
  - A disturbance strikes `instance` with the qualified `fault` at the row field named by `at` (or at `at_s` seconds);
    `severity` is a number or a row field; `duration_s` clears it that much later. A row field that is a list (S4's
    pulses) repeats the disturbance once per element.
  - A seed row holds the world `seed` and the drawn values, in seconds after the world start: `onset_s`, `severity`,
    `stop_s` (the downstream stop), `trip_s` (S2, S6), `frozen_s` (S5), `dropout_s` (S5b), `pulse_s` and `pulse_severity`
    (S4). They were drawn once from a fixed generator and are committed as numbers. S5b's rows are S5-A's, with
    `frozen_s` as `dropout_s`: D-3's vibration channel drops out (`conveyor_drive__vibration__dropout`, no reading
    delivered from the injection on) where S5-A freezes it, and nothing else differs. The dropout is a label of its own
    on D-3 beside `lube_loss`, observable at its injection and expecting no action. S1-B's rows also hold
    `planned_stop_s` (below), computed from the row, not drawn.
  - A policy call invokes `effector` on `instance` with `args` at `after_s` seconds after the row field named by
    `anchor`. The policies know the drawn onset: they only prove that a saving reaction exists.
  - S1-B's policy anchors on `planned_stop_s`, the row's do-nothing D-3 trip less the story's `planned_stop_lead_s`
    (1.5 h), rounded down to 5 min, where a perfect prognosis would put it: full feed until then, feed 0 for 180 s to
    empty the belt, stop D-3, then the safe limp feed. The long tier checks the lead on every row.
  - `safe_limp_feed_pct`: the feed each plant runs at on two drives, a few points below the highest feed at which two
    drives can still restart the belt they leave loaded and stay at or under their rating.
- The drive faults name their expected actions in both compositions (`expected_actions`), so `streamsim score` credits
  them on the instances a label resolves: `lube_loss` expects `set_feed` on the feeder and `stop_drive` on the faulted
  drive, `gear_oil_loss` the same plus `stop_belt` on the belt, `vfd_fault` `set_feed` on the feeder.
- The drives share the belt's demand by weight: each drive's `weight` is 1 unless the instance fault
  `conveyor_drive__share_skew` moves it by its severity (a step, from t = 0 in a seed row that skews a drive); a drive then
  carries `weight·n/Σweight` of the equal share.
- `kinds/`: the plants' interface to a plant-wire host, which every `domain compose`, `run`, `plantwire node` and
  `plantwire connect` of either plant needs as `--kinds-dir examples/belt-conveyor/kinds` (each kind `implements` its file).
  - `bulk_feeder.kind.json`: the feeder as a host sees it, `feed` (kg/s) and `feed_setpoint` (%, writable) and the command
    `set_feed` with the argument `feed_setpoint`, so a host confirms a feed command by reading the setpoint back.
  - `feeder-controller.board.json`: the board mapping from that kind file to the composed feeder's raw names
    (`bulk_feeder.feed`, `bulk_feeder.feed_setpoint`, `bulk_feeder__set_feed` with `percent` from `feed_setpoint`). No
    conveyor command carries an id argument: the call (or the plant wire device) addresses the instance.
  - `conveyor_drive.kind.json`: a drive as a host sees it, `current` (A), `winding` (degC), `vibration` (mm/s),
    `run_status` (boolean, what the drive does) and `run_command` (writable, what it was last told: 1 or 0, read back from
    the controller's command register by the instrument `streamsim/command-readback`); the commands `start_drive`
    (`run_command` 1), `stop_drive` (`run_command` 0; the plant still refuses a stop its permit forbids) and `reset_trip`
    (no argument). `drive-controller.board.json` maps it to the composed raw names (`conveyor_drive.current`, ...,
    `conveyor_drive__start_drive`).
  - `belt_conveyor.kind.json`: the belt, `tail_speed` (rpm) and `run_command` (writable), with `start_belt`
    (`run_command` 1) and `stop_belt` (`run_command` 0); `belt-controller.board.json` maps it to the raw names.
  - A host confirms every feed, start and stop by reading its argument back on the metric of the same name; plants A and
    B announce the same three kind digests.
  - The S6 plants split the drives into the kinds `drive_master` (D-1) and `drive_follower` (D-2, D-3), both
    `streamsim/conveyor-drive@2` (the conveyor drive plus a speed-reference link: D-1 -> D-2 -> D-3). A kind file names
    the kind it describes, so each role has its own, `drive_master.kind.json` and `drive_follower.kind.json` (the drive's
    metrics and commands under the role's name), mapped by `drive-master-controller.board.json` and
    `drive-follower-controller.board.json` to the raw names `drive_master.*` and `drive_follower.*`. With
    `master_loss: stop_all` the followers stop one reference sample apart when D-1 is lost and the loaded belt is stranded;
    with `promote` they keep their torque and D-2 takes the speed loop 2 s later.
- `scenarios/<story>.do-nothing.scenario.json` and `<story>.reference.scenario.json`: seed row 1 as `scenario-v0.1`
  runs (S0 and S4 have no policy, so no reference file). Run one with
  `streamsim run --scenario examples/belt-conveyor/scenarios/s2-a.reference.scenario.json --domains-dir examples --kinds-dir examples/belt-conveyor/kinds`.
- `sessions/<story>.session-plan.json`: seed row 1 as a live session plan (`session-plan-v0.1`), S2 and S3 paced by the
  wall clock. The `host_socket` `/tmp/streamsim-host.sock` is a placeholder and there are no `policies`: set both to
  the host's own. Each plan carries the story's `horizon_s`, so the session ends itself there (`stopped: false`).

To grade a story run on its consequences, run its scenario into a folder and score that folder:

```bash
streamsim run --scenario examples/belt-conveyor/scenarios/s2-a.reference.scenario.json --domains-dir examples --kinds-dir examples/belt-conveyor/kinds --sink file --out /tmp/s2-a
streamsim outcome --run /tmp/s2-a --domains-dir examples
```

`outcome` replays the run as it ran, with every effector call removed (`do_nothing`) and under each of the plant's
baseline policies (PLAN.md section 2.2 as decided in O27): `b1` (a drive's vibration above 4.5 mm/s for 60 s: feed 0, and
60 s later stop that drive), `b2` (a drive's current above 1.1 times its current at full feed on three drives, 3.385 A on
plant A and 4.688 A on plant B, for 60 s: feed 67 %), `b3` (from the first second, the plant's safe limp feed: 54 % on A,
18 % on B) and `b4` (feed to 67 % when a drive is unavailable), prices each with the plant's `outcome_weights` (equivalent production
hours) and says whether the run beats, ties or loses against each. `streamsim report /tmp/s2-a --outcome` carries the same result in
its `outcome_score` section (a report without `--outcome` does not compute it). A live session that ended at its `horizon_s` is graded the same way from its run folder.

The scenario and session files are generated from `stories.json` by `test/contract/belt_stories_files_test.go`, which
fails when a committed file differs; after editing the table, regenerate them with
`UPDATE_GOLDEN=1 go test ./test/contract -run TestTheCommittedStoryFilesAreWhatTheStoryTableGenerates`.
`test/contract/belt_stories_check_test.go` runs every seed row of every story, with and without its policy, and checks
the story's claims (skipped under `-short`).
