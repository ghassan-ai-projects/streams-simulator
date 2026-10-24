# Phase 2 design: the binary runs the experiment

Status: proposed, written before any Phase 2 code (AGENTS.md rule 4). Facts are from a read-only trace at `38686771`.

## Problem

A canvas-started run is driven by the browser (`launchStore.start`: `world.create`, then a loop of `clock.advance`, `fault.inject`, `effector.invoke` with a wall-clock wait, then `run.end`). The plan, the cursor, the stop request and the effector outcomes exist only in browser memory. Consequences: reload kills the run (F02); a stopped run records no intent (F05); "Set up another run" rebuilds the plan from a report that lacks call arguments (F10, F11).

The binary already owns the pieces: the scenario executor (`Run.ExecuteScenario`), the world registry and run lifecycle (`director`), and a precedent for a binary-owned long job (the connected session child process). What is missing is an asynchronous, paced, stoppable job in `director`, and the run record of intent.

## Decisions

1. **Owner.** `director` owns the job registry and lifecycle (it already owns worlds, `RunOf`, `finalizeRun`). `run` owns command execution; the job reuses it, so there is one command-execution implementation (rule 2).
2. **Request is a run plan, not a scenario.** `job.start` takes `{domain, seed, explore, duration_s, rate, faults[], calls[]}` (the fields the launch form collects). The director builds the scenario from the plan in one function; the canvas builds no scenario JSON. Clears and perturbations extend the plan DTO in Phase 4 (deviation from the earlier plan, which listed them in Phase 2: they need the Phase 4 lane editor to be usable).
3. **Execution reuses the scenario executor.** `Run.ExecuteScenario` gains an optional per-command hook `Before(ctx, command) error`. The job's hook paces `clock.advance` against a wall clock and returns `ErrStopRequested` when stop is requested. No second executor.
4. **Pacing lives behind `internal/wall`.** Domain layers read no clock; the job takes a `Pacer` interface, the edge supplies the wall implementation, tests supply a fake.
5. **Asynchronous.** `job.start` returns the world id at once. The bridge serialises director tool calls on `directorMu`, so a synchronous run would block every other call.
6. **The job is the world.** The job runs on the world created by `director.CreateWorld`, so the folder `runs/<world>` and the SSE journal follow the canvas already uses for live views keep working unchanged.
7. **Intent is recorded.** `RunArtifact` gains `plan` (requested end, rate) and `termination` (`completed` | `stopped_by_operator`, at_ns). The report exposes both; Review shows "Stopped by operator at 28:00 of 60:00".
8. **Bridge.** New MCP-backed actions `job.start`, `job.status`, `job.stop`, `job.list` in the registry (A7 parity applies automatically). `job.start` registers the world in `b.running` so the director is not restarted under it. The workspace listing reconciles `b.running` with `job.list`, so a job that finished while no browser was open does not stay Running.
9. **Running flag for every run (F04).** The bridge emits `running` for all runs; the canvas decodes it for all runs. This is independent of the job and lands first.
10. **Reload recovery.** On start the canvas asks `job.list`; an active job restores the Running state, progress, call outcomes and Stop.
11. **Interruption.** If the director process dies, its in-memory jobs die with it. The folder has an unclosed journal and the listing shows it as not running and not ended. The label becomes Interrupted (not "Did not finish"); no attempt is made to finalise it in this phase.
12. **Repeat exact.** The submitted plan is stored with the run in the sealed director stream (`plan.json`), because it names faults. "Set up another run" reads it through a sealed-aware action once the run has ended. For a run that never ended, the canvas says the plan could not be recovered and requires a deliberate mode choice (F11).

## Out of scope for Phase 2

Pause, step and rate change (Phase 3); fault clears, perturbations, scenario files (Phase 4); finalising interrupted runs; moving the CLI `--duration` scenario construction to the same builder (candidate follow-up, the builder should be written so the CLI can adopt it).

## Rounds

| Round | Content | Owner module |
|---|---|---|
| 2.0 | Running flag for all runs, Interrupted label (F04) | `canvasbridge`, canvas |
| 2.1 | Executor hook; artifact `plan` and `termination`; report fields | `run`, `model`, `report`, contracts |
| 2.2 | Plan to scenario builder, job registry, pacing, stop; MCP tools | `director`, `mcp` |
| 2.3 | Bridge actions, running reconciliation, acceptance test | `canvasbridge` |
| 2.4 | Canvas: job port, poller store, reload recovery, Review wording; delete the browser loop | canvas |
| 2.5 | Repeat exact from the stored plan; incomplete-run recovery message (F10, F11) | `canvasbridge`, canvas |

## Quality bars specific to this phase

- A plan run as a job and the same plan built as a scenario and run through `RunScenario` give the same trace digest (or, if run identity enters the digest, the same replay result and command log). This proves one implementation.
- Replay and `run.verify` of a job-produced run pass.
- Fake-pacer tests: start, status, stop at a boundary, stop before the first advance, natural completion, two jobs refused (one at a time is acceptable for the canvas), job error leaves an incomplete run with the reason.
- Reload e2e: start a watch run, `page.reload()`, banner, progress and Stop return, the run completes and Review is complete.
- Stop e2e: a stopped run shows requested and actual duration and "Stopped by operator".
- The browser issues no `clock.advance`, `fault.inject` or `effector.invoke` for a launched run (network assertion).
- The old loop, `planOfRun` calls reconstruction, `leftOut` and the browser-side badge rules are deleted.
