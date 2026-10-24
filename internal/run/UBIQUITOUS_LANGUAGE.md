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
| Explore mode | A run unblinded from creation (`run --explore`, a world created with `explore`); it is never scored and `score` refuses it | `Config.Explore` | `mode: "explore"`, `unblinded` |
| Run journal | The two JSON Lines streams a run writes while it happens (observer `journal.jsonl`, director `sealed/journal.jsonl`); replay rebuilds both and verify compares their digests | `Run.journal`, `Config.JournalDir`, `journal.Journal` | `journal_digest`, `sealed_journal_digest` |
| Reproducible | Whether replay can reproduce the run byte-for-byte (stepped time, quiesced consumer) | `Reproducible` | `reproducible` |
| Hidden-state history | Director-only per-emission state snapshots | `History`, `model.StateSnapshot` | `world_state_history.jsonl` |
| Inputs | The domain and adapter a run names, resolved by id from a domains directory and an adapters directory | `LoadInputs` | `--domains-dir`, `--adapters-dir` |
| Scenario export | The scenario that reproduces a recorded run, with the effector calls the world refused dropped, checked as `run --scenario` would check it, and its file content; loading the artifact and its inputs is part of the entry | `ExportScenarioFile`, `ScenarioExport`, `ScenarioFromArtifact`, `PruneRefusedCalls` | `scenario from-run` |
| Scenario hooks | The two callbacks a host gives the scenario executor: `Before` sees each command and may stop the loop cleanly or fail it, `After` sees each executed command with its `StepResult` (the effector acknowledgement of an invoke, including a recorded refusal such as `interlock_refused` or `command_id_reused`) | `ScenarioHooks`, `StepResult`, `ScenarioOutcome` | — |
| Scenario run | Resolve a scenario's inputs, check it before any sink opens (a violation is a `ScenarioCheckError`), execute it and end the run | `RunScenario`, `RunSummary`, `ScenarioCheckError` | `run`, `run --scenario` |
| Replay file | Reproduce the run artifact at a path against the domain and adapter it recorded | `ReplayFile`, `Directories` | `replay`, `verify` |
| Domain check | The check of a domain against the kind files its kinds implement, given to a scenario run by its surface (`model.DomainCheck`); with none given, a domain that implements a kind file is refused, so a missing check is never a skipped one. It returns what it checked (`implements_checked`). A recorded run (`ReplayFile`, `ExportScenarioFile`) and `suite` run no domain check: they reproduce or describe a run and connect to no host | `CheckDomain`, `ScenarioRunOptions.CheckDomain`, `rules.RefuseImplements`, `RunSummary.ImplementsChecked` | `implements_checked` |
| Start bound | The latest a run's world may start, the world's own bound | `MaxStartNS` | `--start-time` |
| Wire trace | The run journal's observer stream written without a world: only `wire` session records and a closing record, opened by the plant wire relay on its folder (`OpenWireTrace`, `WireTrace`); a hardware relay has no truth, so its director stream holds only the closing record | `WireTrace` | `journal.jsonl` |
| Field trace | What a gateway sent, in the order a relay received it: one canonical `{ordinal, record}` line per record and nothing else, an input of the run beside the command log; its digest (the SHA-256 of the file) is the optional `field_trace_digest` of the artifact and is checked with the other evidence files. A run without one has neither the member nor the file | `FieldTrace`, `NewFieldTrace`, `ReadFieldTrace`, `Run.SetFieldTrace`, `ErrFieldTrace*` | `field-trace.jsonl`, `field_trace_digest` |
| Interrupt log | What a plant wire node was told to refuse or stop and what it decided: operator controls, host commands, node clock readings and the node's decisions about them, one canonical `{ordinal, record}` line each (`interrupt-log-v0.1`); evidence of the run, its digest is the optional `interrupt_log_digest` of the artifact and is checked with the other evidence files. A run without a control socket has neither the member nor the file | `InterruptLog`, `NewInterruptLog`, `ReadInterruptLog`, `Run.SetInterruptLog`, `ErrInterruptLog*` | `interrupts.jsonl`, `interrupt_log_digest` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `RecordHistory`, `RenderRecord`, `Unblinded`, `TraceDigest`, `AddEntity`, `ConfigureEnvTarget` (exported) | removed from the facade | no caller outside the module |
| `Run.Perturb`, `Run.Engine`, `Run.Sink` (exported fields) | app-layer fields | no outside reader |
| `ledgerWriter`/`ledgerFile` fields | `durable.Ledger` | the file is the durable edge's |
