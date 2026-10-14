# P9 latest-main integration review — 2026-10-14

Owner authorized merging latest main after the test-speed work. Merge `856bcb9` imports PR #33 (`02fe45a`)
and source commit `762c7fa`. Its explicit `Owner-approved-gate-change:` line covers moving long simulations,
component sweeps and race detection out of required pipelines. No new gate edit was made during this integration.

Root reviewed every changed file: AGENTS/context testing, Makefile, CI, scripts/test-ci and the two contract tests.
A Luna engineering reviewer independently reviewed the same diff read-only. The three CI matrix jobs together
cover ci-check targets; the separate Canvas job remains. Coverage floors are unchanged. Required test-ci now runs
once with short mode, simdet, shuffle and coverage; test-long and test-race are on demand.

Finding MI-1: phase-gate/CI no longer runs the untagged production wall-clock Go tests. This is a narrower evidence
signal than the earlier two-tier gate. Standalone `make quick`, `make test`, `make test-short` and
`make test-coverage` still run untagged tests; root verified this against Makefile, correcting the reviewer's
initial broader statement that no target did. Post-merge quick passed. Disposition: documented coverage boundary;
no unapproved gate change. Run relevant untagged/race checks when changing clock or concurrency behavior.

Finding MI-2: the KPI-tracking equality test is permanently shortened from six hours to one hour/three steps;
`test-long` does not restore the old six-hour sequence. It still exercises a fault, three effector calls and a
second fault. Disposition: incoming main change reviewed; no evidence that this invalidates the specific alarm
activation scan/copy/report tests or real-binary Chrome navigation. Do not claim the old horizon was rerun.

Long contract/acceptance skips include catalog preset sweep, ten-hour factory cascade, campus speed/baseline
checks and longer binary runs. Report/verify performance and dynamics work must run the relevant on-demand tests
and actual campus measurements. Prior full alarm evidence used the old two-tier gate; future phase-gate passes
have the new, explicitly approved scope. Incoming changes introduce no simulator production-code change.
