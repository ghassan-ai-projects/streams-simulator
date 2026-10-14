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
| Run lifecycle | Seal truth, begin the run, end it, reveal truth, score, verify, audit, in that order | `SealTruth`, `BeginRun`, `EndRun`, `RevealTruth`, `Score`, `VerifyRun`, `AuditScenario` | `sim.truth.*`, `sim.run.*`, `sim.score`, `sim.scenario.audit` |
| Operator | The role the consumer uses; sees the nameplate, the effectors and an acknowledgement, never truth, faults or hidden state | `OperatorNameplate`, `OperatorEffectors`, `OperatorInvoke`, `OperatorReport` | `sim.nameplate.read`, `sim.effector.*`, `sim.consumer.report` |
| Operator view | What one capability token resolves to: nameplate, effector table, invocation surface, verdict sink; it has no field for truth | `OperatorView` (domain) | – |
| Nameplate | The static, time-invariant description of a world | `Nameplate`, `EffectorInfo` | `nameplate` |
| Capability token | Opaque unguessable bearer value that names exactly one world to the operator | `capability.NewToken` | `token`, `t-…` |
| Operator endpoint | The URL the director advertises for operators to connect to | `SetOperatorEndpoint` | `operator_endpoint` |
| Explore world | A world created with `explore: true`: its run is unblinded from creation and `sim.score` refuses it | `WorldRequest.Explore` | `explore` in `sim.world.create` |
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
