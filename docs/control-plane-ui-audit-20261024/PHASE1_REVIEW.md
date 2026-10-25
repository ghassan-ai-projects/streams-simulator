# Phase 1 review: quick trust fixes

Branch `control-plane-p1-trust-fixes`, from the audit commit `10bd5a4e` on `3c6d78c8`. Gate: `make test-ci` green (coverage floors satisfied); `npm run check` green on the F01 change (1957 unit tests, 125 e2e).

## Closed

| Finding | Commit | Evidence |
|---|---|---|
| F03 failed start recorded as complete | `5ec22453` | Node and mixed-relay tests: red with `abandonFailed` neutered, green with it; operator-stop control stays complete. Chrome on the fresh binary: review shows "Did not finish" with the plantwire reason, no Verified badge, replay "Not checked" |
| F01 Evaluate shows the injected fault | `86caefa5` | Model, DOM canary and real-binary e2e tests, red then green. Chrome: no "fan failure" or "Fault ·" in page text, node, inspector or panel during an Evaluate run; the sealed review lists the fault; Explore still shows it |
| Scrub bug (`/privatesockets/...`) | `58ed3367` | Audit string reproduced as a failing table case, then fixed. Chrome: launch failure path reads `sockets/absent2.sock` |
| F20 contradictory Explore wording | `38686771` | Launch panel now reads "The truth is visible while the run is going; this run is not for scoring." Verified in Chrome |

## Gate changes

One line in `scripts/coverage-floors.txt`: `internal/canvasbridge/internal/files` 90 to 92 (ratchet; a floor may rise). No cap, allow-list or hook change. `git diff origin/main -- test/architecture scripts` shows only that line. (`git diff main` also shows the `uds` 95 to 94 change; that is commit `1eb00d81`, already on `origin/main`, and not part of this work.)

## Decisions taken without a blocking answer

- F01: Evaluate is observer-only; the plan's fault count ("1 fault") is still shown in the run panel, identities are not. Revisit if the owner wants the count hidden or an explicit director view.
- `Aliases` in `canvasbridge/internal/files` walks symlinks with `os.Readlink` because `filepath.EvalSymlinks` is not in the declared I/O inventory and adding it needs an `Owner-approved-gate-change:` line. If approved, swap the body for `EvalSymlinks` and delete about 25 lines.
- The relay rule "a mixed session that closed with zero host sessions is incomplete" is new behavior; it follows from the F03 requirement but was a judgment call.

## Found during the Chrome review (backlog)

| Item | Where it goes |
|---|---|
| The Home run row and the report's "the run is incomplete: ..." line show the unscrubbed absolute workspace path; only action errors are scrubbed | Round 2.0b (listing and report text go through the same scrub) |
| A failed-to-start run still shows "True results" KPI tables with no "initial state only" label | Phase 2.4 (Review wording) |
| Failure panel offers only "Back to the session plan", no link to the retained run | F16, with Phase 5 |
| Explore fault chip overlaps and clips the node title and icon during a run | F17, Phase 3 plant focus |
| Check-plan message for an over-long socket path says "Give the host socket" although a path was entered; the real reason is hidden in developer text | F14/F16, Phase 5 |
| The bridge resolves the adapters directory relative to its working directory | Small cli/bridge fix, not scheduled |
| Stale `bin/streamsim` used by the audit | Caveat in PLAN.md |

## Not verified

- Dark mode, other window sizes (Chrome frame was 1384 x 868), hover titles on nodes (the tool blocked the script that would list them; the DOM canary test covers attributes).
- F02, F04, F05, F10 to F13 are not in this phase.
- No race or long test run.
