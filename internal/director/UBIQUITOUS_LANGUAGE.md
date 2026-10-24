# Ubiquitous language — director

The director module runs worlds on behalf of a surface. It is the application
logic behind the MCP tools: the registry of worlds and runs, the world
commands, the run lifecycle and the operator view. It knows no protocol:
callers pass typed values, and a refusal carries a stable code a surface
renders in its own terms.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Director | The role that creates worlds, advances the clock, injects faults and perturbations, holds sealed truth and scores | `Director`, `New`, `NewFromDirs` | `sim.*` director tools |
| World | One created simulation under the director's control, named by a world id | `WorldRecord` (app), `CreateWorld` | `world_id`, `w-…` |
| World request | The typed input of world creation; empty strings select the documented defaults, a nil start time selects the default start | `WorldRequest` | `sim.world.create` arguments |
| World command | A director action on a live world: advance, fault, perturb, environment fault, retire | `Advance`, `AdvanceBy`, `InjectFault`, `ApplyPerturb`, `EnvInject`, `RetireEntity`, `WorldEffectors`, `InvokeWorldEffector` | `sim.clock.*`, `sim.world.effectors`, `sim.world.invoke`, `sim.fault.*`, `sim.perturb.*`, `sim.env.inject`, `sim.entity.retire` |
| Job | A run plan executed by the director on its own goroutine, paced against the wall clock and stoppable: it owns a world, builds one scenario from the plan, runs it through the run's scenario executor and ends the run when the plan completes, a stop is requested or a command fails; one runs at a time, and a finished job stays listed until the director exits | `StartJob`, `JobStatus`, `StopJob`, `ListJobs`, `JobRequest`, `JobStatus` | `sim.job.*` |
| Job plan | The scenario a job request becomes: chunked `clock.advance` commands with the faults and effector calls at their boundaries (advance, then faults, then calls; time-0 actions first), command ids `job-<n>` | `PlanScenario` | `scenario-v0.1` |
| Pacer | What a job asks to hold an advance until the wall clock has reached its simulated time, and to re-anchor when it resumes or its rate changes; the wall implementation lives in the `runner` edge, tests supply their own | `Pacer`, `runner.Pacer` | `rate` |
| Job controls | What the operator does to a live job without touching the world: `pause` parks it at its next command boundary (the call returns once the clock stands still), `resume` lets it go on at its rate with the pacer re-anchored at the reached instant (no catch-up for the paused time), `step` releases one clock chunk of a paused job without pacing, `rate` re-anchors the pacer at the reached instant with a new rate; none is recorded in the run | `PauseJob`, `ResumeJob`, `StepJob`, `SetJobRate`, `JobPaused` | `sim.job.pause`, `sim.job.resume`, `sim.job.step`, `sim.job.rate` |
| Immediate command | A fault or effector call applied to a running or paused job at the run's current simulated clock, through the run's own command methods, so it is logged in execution order with its time and replays; a refused effector call is a `calls[]` outcome flagged `immediate`, its command ids are `job-now-<n>` | `JobInjectFault`, `JobInvokeEffector`, `JobImmediate`, `InjectFaultNow`, `InvokeAcknowledgedNow` | `sim.job.inject_fault`, `sim.job.invoke_effector` |
| Run lifecycle | Seal truth, begin the run, end it, reveal truth, score, verify, audit, in that order | `SealTruth`, `BeginRun`, `EndRun`, `RevealTruth`, `Score`, `VerifyRun`, `AuditScenario` | `sim.truth.*`, `sim.run.*`, `sim.score`, `sim.scenario.audit` |
| Operator | The role the consumer uses; sees the nameplate, the effectors and an acknowledgement, never truth, faults or hidden state | `OperatorNameplate`, `OperatorEffectors`, `OperatorInvoke`, `OperatorReport` | `sim.nameplate.read`, `sim.effector.*`, `sim.consumer.report` |
| Operator view | What one capability token resolves to: nameplate, effector table, invocation surface, verdict sink; it has no field for truth | `OperatorView` (domain) | – |
| Nameplate | The static, time-invariant description of a world | `Nameplate`, `EffectorInfo` | `nameplate` |
| Capability token | Opaque unguessable bearer value that names exactly one world to the operator | `capability.NewToken` | `token`, `t-…` |
| Operator endpoint | The URL the director advertises for operators to connect to | `SetOperatorEndpoint` | `operator_endpoint` |
| Explore world | A world created with `explore: true`: its run is unblinded from creation and `sim.score` refuses it | `WorldRequest.Explore` | `explore` in `sim.world.create` |
| Kind check | The function (`model.DomainCheck`) a director is given to check a domain against the kind files its kinds `implements`, run when a world is created through `run.CheckDomain`, the one implementation `run` shares; a director given none refuses every domain that implements a kind file. The world's result lists what was checked | `check` (`NewFromDirs`), `run.CheckDomain` | `domain_invalid` at `/kinds/<i>/implements`; `implements_checked` |
| Run directory | Where a world's run publishes its evidence: the ledger and the two journal streams while it runs, the artifact at the end | `Director.OutDir/<world_id>` | `ledger.jsonl`, `journal.jsonl`, `sealed/journal.jsonl` |
| Harness run handle | The run behind a world, for composition tests only (the MCP surface cannot name its type) | `Director.RunOf` | – |
| Refusal | A typed failure carrying a stable code; operator-facing codes are deliberately coarse | `Code`, `ToolErrorf` | `code` in the tool error text |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `mcp.Director`, `mcp/internal/app` | `director.Director` | the director is application logic, not protocol; `mcp` keeps tool definitions, argument decoding and transport |
| `SingleView`, `OperatorResolver`, `NewOperatorServer(view)` | `OperatorNameplate`, `OperatorInvoke` and the token | one operator endpoint serves every world; the token, not a handle, names the world |
| `Str`, `Num`, `IntArg` (in the app layer) | `stringArg`, `numberArg`, `intArg` in `mcp/internal/protocol` | decoding tool-call argument maps is protocol work |
| `ServeOptions` in the app layer | `LoadDirector(ctx, domainsDir, adaptersDir, outDir)` and `mcp.ServeOptions` | the option set that includes the operator address is a serving concern |
