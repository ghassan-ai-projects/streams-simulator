# Belt conveyor stories

The M0 stories of the multi-drive belt conveyor (`docs/belt-conveyor-drives/PLAN.md` section 2.3) as data, for the
plants `belt-conveyor-drives-a` (drives with margin) and `belt-conveyor-drives-b` (without).

- `stories.json`: one entry per story variant (S0-A, S1-A, S1-B, S2-A, S2-B, S3-A, S4-A, S4-B, S5-A). Each has its
  plant, horizon, disturbances, ten seed rows and a hand-written reference policy.
  - A disturbance strikes `instance` with the qualified `fault` at the row field named by `at` (or at `at_s` seconds);
    `severity` is a number or a row field; `duration_s` clears it that much later. A row field that is a list (S4's
    pulses) repeats the disturbance once per element.
  - A seed row holds the world `seed` and the drawn values, in seconds after the world start: `onset_s`, `severity`,
    `stop_s` (the downstream stop), `trip_s` (S2), `frozen_s` (S5), `pulse_s` and `pulse_severity` (S4). They were drawn
    once from a fixed generator and are committed as numbers.
  - A policy call invokes `effector` on `instance` with `args` at `after_s` seconds after the row field named by
    `anchor`. The policies know the drawn onset: they only prove that a saving reaction exists.
  - `safe_limp_feed_pct`: the feed each plant runs at on two drives, a few points below the highest feed at which two
    drives can still restart the belt they leave loaded and stay at or under their rating.
- The drive faults name their expected actions in both compositions (`expected_actions`), so `streamsim score` credits
  them on the instances a label resolves: `lube_loss` expects `set_feed` on the feeder and `stop_drive` on the faulted
  drive, `gear_oil_loss` the same plus `stop_belt` on the belt, `vfd_fault` `set_feed` on the feeder.
- `kinds/`: the plants' interface to a plant-wire host, which every `domain compose`, `run`, `plantwire node` and
  `plantwire connect` of either plant needs as `--kinds-dir examples/belt-conveyor/kinds` (each kind `implements` its file).
  - `bulk_feeder.kind.json`: the feeder as a host sees it, `feed` (kg/s) and `feed_setpoint` (%, writable) and the command
    `set_feed` with the argument `feed_setpoint`, so a host confirms a feed command by reading the setpoint back.
  - `feeder-controller.board.json`: the board mapping from that kind file to the composed feeder's raw names
    (`bulk_feeder.feed`, `bulk_feeder.feed_setpoint`, `bulk_feeder__set_feed` with `feeder_id` fixed to `F-1` and `percent`
    from `feed_setpoint`).
  - `belt_conveyor.kind.json`, `conveyor_drive.kind.json`: captured from a session birth with `streamsim kind capture`;
    plants A and B announce the same three kind digests.
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

`outcome` replays the run as it ran, with every effector call removed (`do_nothing`) and under the plant's baseline policy
`b4` (feed to 67 % when a drive is unavailable), prices each with the plant's `outcome_weights` (equivalent production
hours) and says whether the run beats, ties or loses against each. `streamsim report /tmp/s2-a --outcome` carries the same result in
its `outcome_score` section (a report without `--outcome` does not compute it). A live session that ended at its `horizon_s` is graded the same way from its run folder.

The scenario and session files are generated from `stories.json` by `test/contract/belt_stories_files_test.go`, which
fails when a committed file differs; after editing the table, regenerate them with
`UPDATE_GOLDEN=1 go test ./test/contract -run TestTheCommittedStoryFilesAreWhatTheStoryTableGenerates`.
`test/contract/belt_stories_check_test.go` runs every seed row of every story, with and without its policy, and checks
the story's claims (skipped under `-short`).
