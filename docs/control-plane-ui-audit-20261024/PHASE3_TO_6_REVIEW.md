# Phases 3 to 6 review

Branch `control-plane-p3-live-diagnosis` (PR 53). Records what each phase closed, the gate changes, the deviations from `PLAN.md`, the known limits and what was not verified. Phase 5 (connected operation) is deliberately not started: it needs a real plant-wire host and the fan experiment's real run.

## Closed

| Finding | Where it landed | Evidence |
|---|---|---|
| F06 inspector is a catalogue while running | Phase 3A: a third review mode ("watch") streams the job's observer journal; the inspector shows value, unit symbol, observed-at, age in simulated time, quality when not good and a gap-aware mini-trend; memory bounded for long runs | Component tests, canary tests, real-binary e2e; Chrome review |
| F07 no multi-signal trend view | Phase 3B: per-plant watch list, a Trends panel with a shared time axis and cursor, one scale and unit per channel, gaps as gaps, the operator's calls always and faults only where truth is allowed | Model, component (axe, keyboard) and e2e tests; render counts in `PHASE3_PERF.md` |
| F08 watch run cannot be operated | Phase 3b: pause, step, speed and immediate effector or fault commands on a running job (Go, then canvas); Evaluate offers no fault injection | Pause, step and rate never change the trace digest (tested against `RunScenario`); immediate commands replay |
| F18 missing data unexplained | "Awaiting first reading", "Never observed in this run", no cause named | Tests grep for causal words |
| F28 raw unit codes | Unit table widened to every unit string in the shipped data, guarded by an e2e that walks the shipped folders | `e2e/shipped-units.spec.ts` |
| F12, F13 plan richness | Phase 4a/4b: name and purpose, faults with an end, delivery perturbations from the binary's catalog, fault categories, one time-ordered plan list | Job = `RunScenario` digest equality incl. clears and perturbations; repeat from the report is exact |
| F19, F20 report answers the question | Phase 4c/4d: report compare (plan diff, comparability limits, whole-run deltas with scope, no ranking), handover report with open questions, This run block, one visibility sentence | Golden and acceptance tests; Chrome review |
| F21 Evaluate without a scorer | "Unscored investigation" stated in Before you start and in the This run block | |
| F22, F23 twin defaults and messages | Phase 6 R1: unsupported settings named with their owner, partial twins listed, defaults tagged Suggested, four separate claims (interface, plausibility, calibration not performed, hardware validation not run) | Component and e2e tests |
| F24, F26 provenance and read-only plants | Phase 6 R2/R3: `domain lineage` (binary) and canvas views: Included in twin, why a plant is read-only and the right route, a lineage drawer; a hand-authored domain is never auto-converted | Acceptance tests with the real binary |
| F25 parameter editing | Phase 6 R4: preset and parameter editing with declared units and bounds, the binary's refusal at the field, digest and a plausibility row after save | |
| OP08 model assumptions and diff | `domain assumptions` and `domain diff` (binary) with canvas views, including the warning when the model changed but its version did not | OP08 thermal-gain acceptance test |

## Truth-boundary defects found and fixed on the way

- Round 4a put planned faults and perturbations into `job.status`, `job.list` and `job.start` results; for an Evaluate job these reached the browser's network responses while it ran. One redaction function now blanks them for an active Evaluate job, and `job.inject_fault` is refused for Evaluate jobs (`evaluate_blind`). A real-binary test scans raw response bodies for planted names.
- The Chrome review found the reload path showed stale fault marks, a missing action-story panel, a contradictory integrity sentence and a running Home row reading "Unknown plant"; all fixed (see the commits "browser review").

## Gate changes (`git diff origin/main -- test/architecture scripts test/pin`)

None loosens a gate. `test/architecture`: `modules_test.go` and `io_edges_test.go` register the new `internal/director/internal/runner` edge; `cli_wiring_test.go` registers three new CLI files (`domain_assumptions.go`, `domain_lineage.go`, `domain_diff.go`) with their sizes as caps. `scripts/test-hooks.txt`: one entry removed. `scripts/coverage-floors.txt`: three floors raised. `test/pin/BEHAVIOUR_PIN.txt`: one help-text line. `canvas/architecture.json` allow-lists, floors and caps are untouched; the Phase 6 canvas round reused existing modules, so the `responsibility` strings of `model/workspace`, `app/workspace`, `ui/plant`, `ui/review` and `ui/editor` are now out of date and need a register edit that was not made.

## Deviations from the plan

- No new canvas modules for provenance (allow-list edits were not made); the code lives in existing modules.
- Plans: pause, step and rate are not recorded in the run (only immediate commands are, as they change the world).
- Repeat exact rebuilds the plan from the run's report rather than storing `plan.json`.
- A suggestion is shown as "Suggested by the binary" or "Differs from the binary's suggestion", never "Your choice" (the choices file records no choice).

## Known limits

- Staleness by expected cadence is not classified (it needs a binary capability and an Evaluate-safe way to expose cadence); the canvas says only "Awaiting first reading" or "Never observed".
- A reopened twin cannot say it is partial (the choices file does not record exclusions); Home shows "Twin created".
- Hardware validation always reads "not run": the devices screens have no twin-report port.
- `cleared_at_s` exists only for clears the director executes; responder and pulse clears are not reported while the job runs.
- Canvas check time: about 5 minutes locally (end-to-end about 4); the cap in `testing.md` is 2 minutes. See `CANVAS_TEST_TIME.md` for the measurements and the options that need an owner decision (a separate sharded end-to-end CI job, a smoke set on pull requests).
- `make vulncheck` fails on this machine's Go 1.27.2 (fixed in 1.27.3); three plantwire timing tests are flaky under load (the host-barrier one hangs about 1 run in 3 on a clean `origin/main`); a task to fix them was proposed.

## Not verified

- A Chrome review of the final tip after the last two canvas rounds (the Chrome review covered the tip before the review fixes and Phase 6 canvas); dark mode; other window sizes.
- A real board or host (Phase 5).
- The race detector on the whole repository; long runs (hours).
