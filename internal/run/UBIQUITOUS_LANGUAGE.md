# Ubiquitous language — run

A run is the deterministic execution of one world through the delivery
pipeline, recorded so that replay reproduces it byte-for-byte.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Run | One deterministic execution: world, perturbation layer, adapter engine, sink and ledger behind one command log | `Run` | `run_id`, `run.json` |
| Determinism tuple | `(sim_version, domain_digest, adapter_digest, seed, command_log, sink)`: everything a run is a pure function of | `Config`, `Digest` | `world_config`, digests |
| Command | A recorded mutation of the run: advance, fault, perturb, effector, entity, env | `model.Command` | `command_log[]` |
| Command log | The ordered commands that, replayed, reproduce the run | `commandLog` | `command_log` |
| Scenario | A repeatable run as a file: domain, adapter, seed, world settings and the ordered director commands; flags and files both execute one | `model.Scenario`, `ScenarioConfig`, `ExecuteScenario` | `scenario-v0.1` |
| Scenario command | One director operation with the arguments the command log records, without `seq` and `at_ns`; a clock advance carries only its target | `model.ScenarioCommand`, `rules.IsScenarioOp` | `commands[]` |
| Scenario check | Every way a scenario breaks its domain or the run's rules, found before anything is delivered: a static check, then a dry run | `CheckScenario`, `rules.CheckScenario`, `rules.ScenarioFacts` | `scenario_invalid` issues |
| Dry run | A scenario executed through the run's executor on its world with emission off and nothing opened, to find the first command the run would refuse as it unfolds | `newShadow`, `dryRunScenario` | `/commands/<i>` issue |
| Refused call | An effector call the world refused at admission (gone or unknown entity, undeclared effector, refused arguments): it changed nothing, so an export drops it | `PruneRefusedCalls`, `world.ErrInvalidInvocation` | `refused` in the export result |
| Run span | How far a run's clock may move past its start: 30 days of simulated time, on every surface | `rules.MaxRunSpanNS`, `rules.ExceedsRunSpan`, `ErrRunSpanExceeded` | `--duration` bound |
| Command executor | The one executor behind replay and scenario runs; its refusal policy keeps every recorded effector refusal on replay and only the world's run-time refusals (interlock, reused command id) for authored commands | `executeCommand`, `refusalPolicy` | — |
| Unknown profile | A run created with a profile its domain does not declare; refused on every surface, replay included | `app.ErrUnknownProfile` | — |
| Delivery ledger | One row per emitted event with its terminal delivery reason, kept in memory and optionally durably | `ledger`, `durable.Ledger` | `ledger.jsonl`, `delivery_reason` |
| Trace | The delivered byte stream the sink produced; its digest is the replay oracle | `Trace`, `traceDigest` | `trace.jsonl`, `expected_trace_digest` |
| Run artifact | The self-contained record of a finished run: inputs, command log, counts, digests | `model.RunArtifact` | `run.json` |
| Replay | Re-executing an artifact's command log against a fresh world and comparing the trace digest | `ReplayArtifact`, `ReplayResult` | `verify`, `replay` |
| Quiescence | The consumer reporting it has processed everything through a time; `await_consumer` waits for it | `ReportQuiesced`, `Advance(awaitConsumer)` | `consumer_not_quiesced` |
| Verdict | The consumer's conclusion for the run, submitted once | `SubmitVerdict`, `Verdict` | `verdict.json` |
| Unblind | Stamping the run permanently so it is excluded from every scorecard | `Unblind`, `UnblindedStamp` | `unblinded`, `unblinded_at` |
| Reproducible | Whether replay can reproduce the run byte-for-byte (stepped time, quiesced consumer) | `Reproducible` | `reproducible` |
| Hidden-state history | Director-only per-emission state snapshots | `History`, `model.StateSnapshot` | `world_state_history.jsonl` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `RecordHistory`, `RenderRecord`, `Unblinded`, `TraceDigest`, `AddEntity`, `ConfigureEnvTarget` (exported) | removed from the facade | no caller outside the module |
| `Run.Perturb`, `Run.Engine`, `Run.Sink` (exported fields) | app-layer fields | no outside reader |
| `ledgerWriter`/`ledgerFile` fields | `durable.Ledger` | the file is the durable edge's |
