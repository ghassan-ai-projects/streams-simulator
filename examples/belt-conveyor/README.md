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
- `scenarios/<story>.do-nothing.scenario.json` and `<story>.reference.scenario.json`: seed row 1 as `scenario-v0.1`
  runs (S0 and S4 have no policy, so no reference file). Run one with
  `streamsim run --scenario examples/belt-conveyor/scenarios/s2-a.reference.scenario.json --domains-dir examples`.
- `sessions/<story>.session-plan.json`: seed row 1 as a live session plan (`session-plan-v0.1`), S2 and S3 paced by the
  wall clock. The `host_socket` `/tmp/streamsim-host.sock` is a placeholder and there are no `policies`: set both to
  the host's own. Each plan carries the story's `horizon_s`, so the session ends itself there (`stopped: false`).

The scenario and session files are generated from `stories.json` by `test/contract/belt_stories_files_test.go`, which
fails when a committed file differs; after editing the table, regenerate them with
`UPDATE_GOLDEN=1 go test ./test/contract -run TestTheCommittedStoryFilesAreWhatTheStoryTableGenerates`.
`test/contract/belt_stories_check_test.go` runs every seed row of every story, with and without its policy, and checks
the story's claims (skipped under `-short`).
