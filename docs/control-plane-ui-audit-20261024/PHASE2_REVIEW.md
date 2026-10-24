# Phase 2 review: the binary runs the experiment

Design: [PHASE2_DESIGN.md](PHASE2_DESIGN.md). Branch `control-plane-p1-trust-fixes`, rounds 2.0 to 2.6 (commits `2702b7be` to `5a4260f6`).

## Closed

| Finding | Round and commit | Evidence |
|---|---|---|
| F04 running run listed as Did not finish | 2.0 `2702b7be` | Go listing table test and canvas model table; e2e on Home while a Watch run is going. Chrome: Home row reads Running |
| F02 reload loses the run | 2.2 `ad11c1a3`, 2.3 `5f439e03`, 2.4 `ac09ebec` | The job lives in `director`; the canvas adopts it from `job.list` after a reload. Acceptance test: a second client sees the job and its progress mid-run. Chrome: after `location.reload()` the banner, progress and Stop now return and the run finishes normally |
| F05 stop loses intent | 2.1 `a364c0fb`, 2.4, 2.6 `5a4260f6` | Artifact records planned end and termination; report, Review header ("Stopped by operator at 00:09:00 of 00:30:00") and Home row ("9 of 30 min, stopped"). Chrome confirmed the Review header |
| F10 rerun drops effector calls | 2.5 `550a3556` | The decoder dropped `args`; a model table covers complete, stopped, fast and refused-call runs. Acceptance test: a job rebuilt from its report gives equal trace, world and ledger digests. Chrome: the repeat form shows seed, requested duration, pace, fault and the setpoint 7 call |
| F11 incomplete rerun silently resets | 2.5, 2.6 | An unfinished run opens fresh settings with an explicit "could not be recovered" note, for simulation and connected-session sources; earlier failure state is cleared. e2e for both |
| F12 (saved plan) | 2.5 partly | A run's plan is recoverable from its report; named, saved, exportable plans remain Phase 4 |
| F13 (clears, delivery perturbations) | not done | Moved to Phase 4 with the plan editor, as written in the design |

## The key bar

A plan run as a job and the same plan built and run through `RunScenario` give the same `expected_trace_digest`, `world_digest` and `command_log` (`TestAJobAndRunScenarioOfTheSamePlanProduceTheSameTrace`, three plans including an interlock refusal). A repeat from a run's report reproduces the trace, world and ledger digests. `run.verify` passes on job-produced and stopped runs. The journal digest differs between repeats because it contains the run id; that is expected and tested.

## Gate changes

- `test/architecture/modules_test.go`: `internal/director` gains the `runner` layer.
- `test/architecture/io_edges_test.go`: one line for `internal/director/internal/runner/runner.go` (`time.NewTimer`, `time.Now`, `time.Until`, a goroutine). New package registered per MODULARITY_STANDARD; no cap, allow-list or floor loosened. `make gate-ratchet` passes.
- `scripts/coverage-floors.txt`: `canvasbridge/internal/files` 90 to 92 and `canvasbridge/internal/domain` 97 to 99 (floors may rise).

## Design deviations

- Clears and perturbations in the plan: deferred to Phase 4 (as the design note says).
- `plan.json` stored with the run: not built. Repeat exact rebuilds the plan from the run's report instead. Exact for complete runs; for a stopped run, faults and calls planned after the stop are not recorded, and the form says so. Delivery perturbations, environment changes and explicit fault clears are listed as not copied.
- `Workspace` now takes a context (needed by the job reconciliation, which may not create its own deadline in the app layer).
- The listing field is `requested_duration_ns` (a duration), not the artifact's absolute `requested_end_ns`.
- Actions at exactly the plan's end time are executed after the final advance (the browser loop dropped them).
- Pacing waits before each advance; the browser loop advanced first and then waited, so simulated time now lags wall time slightly instead of leading it.

## Known limits

- After a reload in Explore the plan's fault and call details are gone from the plant (no fault marks, no progress ticks, no counts); progress, rate, Stop now and reported call outcomes remain. Evaluate never shows them. Fix by letting `job.status` return the plan for Explore jobs.
- After a reload the run panel is closed; Stop now is one click away (toolbar "Running").
- A director crash loses its jobs; the run is listed as not running and not ended (shown as Did not finish), not finalised.
- `job.start` runs a full dry-run check before starting, so a very long plan delays the first response; not measured on a 24 h plan.
- Arguments of effector calls with nested or non-scalar values are dropped when a plan is rebuilt for a repeat (none of the shipped effectors uses them).

## Process notes

- Binary: the audit and my first review used `bin/streamsim` (stale). `make build` writes `bin/streams-simulator`; every check from Phase 1 on used it.
- Two tests flake when the machine is busy with other work: `TestAPlantWireHostWhoseNodeStopsDuringABarrierWaitPrintsItsReport` (cli/app) and `TestAnAcceptedPeerHasNoReadDeadline` (plantwire/uds). Both pass alone and are untouched by this work. A racy assertion in the new job status test was fixed.
- Chrome review of rounds 2.0 to 2.5 found four defects (note placement, running pill, Home row, recovery note on the session path); all fixed in 2.6 and covered by tests.

## Not verified

- Round 2.6 in Chrome (verified by Playwright screenshots only); dark mode; other window sizes (Chrome frame was 1841 x 1155).
- Jobs longer than a few minutes of wall time, a 24 h plan, two browsers at once, and director restart mid-job beyond the unit test.
- The race detector on the whole repository (run only on the new director packages).
