# Model Context Protocol (MCP) reference

> Status: Implemented reference. Authority: `internal/mcp/internal/protocol/` (`schemas.go`, tool registration) and `internal/director/` (`internal/domain/operator.go`, `internal/app/director.go`). Verified by: MCP strictness and operator tests. Last verified: 2026-08-17.

Streams Simulator exposes one MCP server with two role surfaces:

- **Director** — creates and drives worlds, injects faults and perturbations, owns truth, verifies artifacts, and scores runs.
- **Operator** — the consumer-facing capability surface for one world: read its nameplate, list/invoke effectors, and submit a verdict.

The director runs over stdio. A director may also serve a streamable HTTP operator endpoint. The operator endpoint resolves a world from the capability token, so one endpoint can serve multiple worlds.

## Start the server

```bash
go run ./cmd/streamsim mcp --role director --operator-addr 127.0.0.1:0 --out runs
```

The built-in component library is embedded in the binary and always served; `--components-dir <dir>` (repeatable) and `STREAMSIM_COMPONENTS_DIRS` add customer libraries after it.

The `sim.world.create` response includes the endpoint when configured and always includes the world capability token. Never commit or log the token.

## Director tools

| Tool | Purpose | Required identity |
| --- | --- | --- |
| `sim.catalog.list` | List domains and property vectors. | Optional `group`. |
| `sim.catalog.describe` | Describe one domain. | `domain`. |
| `sim.catalog.coverage` | Return axis coverage. | None. |
| `sim.adapter.list` | List installed adapters. | None. |
| `sim.component.list` | List the component library: the built-in one plus any `--components-dir` customer directories. | None. |
| `sim.component.describe` | Describe a library entry. | `ref` (`namespace/name` or `namespace/name@major`). |
| `sim.domain.compose` | Compose a composition into a domain; with `install` it joins the catalog so `sim.world.create` can use it (an id already installed with another digest is refused). | `composition` (object), `install` (boolean). |
| `sim.world.create` | Create a seeded world and operator capability. | `domain`; `sink_target` for file/http-push sinks. |
| `sim.world.describe` | Report world configuration, clock, and emitted count. | `world_id`. |
| `sim.world.destroy` | End and remove a world. A world destroyed before `sim.run.end` keeps its run artifact as an incomplete run, with the reason in `run.json`. | `world_id`. |
| `sim.world.effectors` | List the world's declared effectors with their argument schemas, as the director (no capability token). | `world_id`. |
| `sim.world.invoke` | Invoke an effector on a world as its director: the same operator logic and acknowledgement as `sim.effector.invoke`; an omitted `command_id` is numbered `director-N`. | `world_id`, `entity_id`, `effector`; optional `args`, `command_id`. |
| `sim.clock.advance` | Advance by `by_ns` or to `to_ns`; optionally await consumer quiescence. | `world_id` and exactly one clock target. |
| `sim.clock.state` | Report clock, next event, and pending effects. | `world_id`. |
| `sim.fault.inject` | Inject a world fault into an entity. | `world_id`, `entity_id`, `fault`. |
| `sim.fault.clear` | Clear a fault. | `world_id`, `fault_id`. |
| `sim.fault.list` | List active faults. | `world_id`. |
| `sim.perturb.apply` | Apply a delivery perturbation. | `world_id`, `perturbation`. |
| `sim.perturb.clear` | Clear a perturbation. | `world_id`, `perturb_id`. |
| `sim.env.inject` | Inject a configured environment fault. | `world_id`, `target`, `fault`. |
| `sim.run.begin` | Open a run; truth is sealed. | `world_id`. |
| `sim.run.end` | Close a run and write its artifact. | `world_id`. |
| `sim.truth.seal` | Install and seal ground truth before run begin. | `run_id`, `ground_truth`. |
| `sim.run.verify` | Replay and verify an artifact. | `run_artifact_path`. |
| `sim.truth.reveal` | Reveal truth; `unblind` permanently stamps the run. | `run_id`; optional `unblind`. |
| `sim.truth.seal_status` | Report sealed/unblinded state. | `run_id`. |
| `sim.score` | Score a submitted verdict against truth and ledger. | `run_id`. |
| `sim.scenario.audit` | Run the trivial-baseline audit for one injection. | `domain`, `entity_id`, `fault`; optional times. |
| `sim.entity.retire` | Retire an entity. | `world_id`, `entity_id`, `reason`. |
| `sim.job.start` | Run a plan as a job: create a world and execute the planned run on the director's own goroutine, paced against the wall clock, and return at once. The plan is checked as a scenario before anything starts; one job runs at a time. | `domain`, `seed`, `duration_s`; optional `explore`, `rate`, `faults`, `calls`. |
| `sim.job.status` | Read a job: state (`running`, `paused`, `stopping`, `completed`, `stopped`, `failed`), simulated time reached, requested end and the outcome of each effector call so far. Never waits for an advance in flight. | `world_id`. |
| `sim.job.stop` | Ask a job to end early. The run closes with the commands executed so far and a `stopped_by_operator` termination; repeating the call is harmless. | `world_id`. |
| `sim.job.list` | List every job since the director started, finished ones included. | None. |
| `sim.job.pause` | Freeze a running job at its next command boundary. The call returns once the job's simulated clock stands still (state `paused`); pausing a paused job is harmless. A pause never reaches the world: the run records no pause. | `world_id`. |
| `sim.job.resume` | Let a paused job go on at its rate from the moment of the call; pacing does not make up the paused wall time. Resuming a running job is harmless. | `world_id`. |
| `sim.job.step` | Release a paused job through exactly one clock chunk (any planned actions still due at the current instant, then the next advance, without pacing) and pause it again; returns when the job is paused again. Refused with `job_not_paused` when the job is not paused. | `world_id`. |
| `sim.job.rate` | Set the pacing rate of a running or paused job (simulated seconds per wall second, `0` as fast as possible), counted from the simulated instant the job has reached. | `world_id`, `rate`. |
| `sim.job.inject_fault` | Inject a fault into a running or paused job now, at the run's current simulated clock, with no parameters. It is logged in the run's command log in execution order and replays. An unknown fault or entity is refused and the job is untouched. | `world_id`, `entity_id`, `fault`. |
| `sim.job.invoke_effector` | Invoke an effector in a running or paused job now, at the run's current simulated clock. The outcome (accepted, or refused with its reason) is appended to the job's `calls` with `immediate: true` and returned as `call`; the command is logged with id `job-now-<n>`. An effector the domain does not declare is refused and the job is untouched. | `world_id`, `entity_id`, `effector`; optional `args`. |

The authoritative input schemas are defined and enforced beside the handlers in [`internal/mcp/internal/protocol/schemas.go`](../../internal/mcp/internal/protocol/schemas.go). Unknown properties are rejected at the MCP schema boundary; domain-dependent checks such as effector names and argument schemas happen in the handler.

## Operator tools

| Tool | Purpose | Required arguments |
| --- | --- | --- |
| `sim.nameplate.read` | Read static world identity, entities, channels, and effectors. | `token`. |
| `sim.effector.list` | Read declared effectors and argument schemas. | `token`. |
| `sim.effector.invoke` | Invoke a declared effector. | `token`, `effector`, `entity_id`, `command_id`; optional `args`. |
| `sim.consumer.report` | Report quiescence and submit a verdict; never returns a score. | `token`, `run_id`; optional `quiesced_through_ns`, `verdict`. |

The operator view contains no truth store, fault registry, perturbation log, or hidden state field. Its narrow shape is a structural leak-prevention measure, not only a convention.

## Important argument fields and defaults

| Tool | Fields and defaults |
| --- | --- |
| `sim.world.create` | `domain` is required. `seed` defaults to `1` and must be an integer from 0 to 9007199254740991 (2^53-1; larger values would not survive JSON numbers exactly), `adapter` to `native-jsonl`, `sink` to `inproc`, and `time_mode` to `stepped`. `sink_target` is required for `file` and `http-push`. Optional `entities`, `scenario_profile`, `start_time`, and `label` are recorded in the world configuration. `explore: true` stamps the world's run unblinded from creation: it is never scored. The world writes its run journal (`journal.jsonl`, `sealed/journal.jsonl`) in its run directory beside the ledger. |
| `sim.clock.advance` | `world_id` plus exactly one of `by_ns` or `to_ns` is required. `by_ns` is non-negative. `await_consumer` defaults false; when true, the call waits for the consumer’s quiescence report or returns `consumer_not_quiesced`. |
| `sim.effector.invoke` | `token`, `effector`, `entity_id`, and unique `command_id` are required. `args` defaults to an empty object and is checked against the domain-declared argument schema. The call acts at the current world clock: an operator cannot date a call, because a dated call would probe hidden state at another time. |
| `sim.consumer.report` | `token` and `run_id` are required. `quiesced_through_ns` is optional; `verdict` is optional when the consumer is reporting only quiescence. The response acknowledges acceptance and never includes a score. |
| `sim.truth.seal` | `run_id` and a complete `ground_truth` object conforming to `ground-truth-v0.1` are required. It must happen before `sim.run.begin`. |
| `sim.job.start` | `domain`, `seed` and `duration_s` (a positive number of seconds) are required. `rate` is simulated seconds per wall second and defaults to `0`, as fast as possible; a positive rate runs 30 chunked advances, `0` runs 20, and each paced advance waits until the wall clock has reached its simulated time. `faults` are `{entity_id, fault, at_s}` and `calls` are `{entity_id, effector, args, at_s}`, with `at_s` from `0` to `duration_s`. At each boundary the run advances to the time, then injects the faults, then invokes the calls, in plan order; actions at `0` run first. The world is an ordinary world (`w-N`, run folder `<out>/w-N`). A refused effector call (interlock, rejection) is recorded as an outcome with its reason and the job goes on. `job_active` refuses a second job while one is running, paused or stopping. The run records the plan (`requested_end_ns`, `rate`) and why it ended (`completed` or `stopped_by_operator`). |
| `sim.job.rate` | `world_id` and `rate` (a finite number, `0` or more) are required. The job's `rate` in its status changes at once; the pacer re-anchors at the simulated instant the job has reached. |
| `sim.job.inject_fault` | `world_id`, `entity_id` and `fault` are required; the fault takes no parameters and starts at the run's clock. The result is the job status plus `fault_id`. |
| `sim.job.invoke_effector` | `world_id`, `entity_id` and `effector` are required; `args` defaults to an empty object and is checked against the domain-declared argument schema. The result is the job status plus `call`, the recorded outcome. A call the world refuses as invalid (unknown entity, arguments outside the schema) is an `invalid_args` refusal; it is still recorded in the run's command log as an invalid invocation, as every surface does. |
| `sim.truth.reveal` | `run_id` is required. `unblind` defaults false; `unblind: true` permanently stamps the run and is allowed only after the scoring decision. |

## Minimal director/operator sequence

The identifiers come from earlier responses: `world_id`, `run_id` and `token` are returned by world creation (and `run_id` again by run begin). The ground-truth record is created by the director, never by the consumer.

1. Call `sim.world.create` with a domain and, for a closed loop, a configured operator endpoint:

   ```json
   {"domain":"rotating-machinery","seed":7,"adapter":"native-jsonl","sink":"file","sink_target":"runs/w-1/trace.jsonl","time_mode":"stepped"}
   ```

   The response includes `world_id`, `run_id`, `world_digest`, `entity_ids`, `clock`, `token`, `implements_checked`, and `operator_endpoint` when HTTP operator serving is enabled. `implements_checked` lists `{kind, kind_digest}` for each kind file the domain was checked against; it is empty when no kind of the domain implements one. A domain that implements a kind file is refused when the server was started without `--kinds-dir`.
2. Call `sim.truth.seal` with the director-generated `ground_truth`, then call `sim.run.begin` with `world_id`. `run.begin` returns `run_id` and confirms `truth_sealed`.
3. Call `sim.clock.advance` with either `by_ns` or `to_ns`. Set `await_consumer: true` at a boundary where the consumer must have processed the delivered evidence.
4. Over the operator endpoint, call `sim.nameplate.read` and `sim.effector.list`, then invoke declared actions with `sim.effector.invoke`. Preserve the same `command_id` across retries. Submit the verdict and quiescence through `sim.consumer.report`.
5. Call `sim.run.end`; the response includes `run_artifact_path`, `trace_digest`, and `reproducible`.
6. Call `sim.score` while truth is still sealed. Only after scoring call `sim.truth.reveal` with `unblind: true` when post-hoc inspection requires the label.

`sim.env.inject` records a configured environment-fault request in the run evidence; it is not a replacement for a world fault or a delivery perturbation. `sim.run.verify` replays a completed artifact and returns digest/version comparison fields.

## Resources

- `sim://catalog` — installed catalog as JSON.
- `sim://domains/{id}/spec` — one domain specification as JSON.

Resources are part of the director surface. They are not a substitute for the operator nameplate and must not be exposed through the consumer endpoint.

## Error taxonomy

The stable error codes are:

`domain_invalid`, `adapter_invalid`, `world_not_found`, `clock_backwards`, `consumer_not_quiesced`, `profile_rate_exceeded`, `truth_sealed`, `run_unblinded`, `unknown_effector`, `invalid_args`, `missing_command_id`, `capability_denied`, `effector_refused`, `interlock_refused`, `not_implemented`, `job_active`, `job_not_running` (a job control or immediate command reached a job that is not running or paused) and `job_not_paused` (`sim.job.step` on a job that is not paused).

Operator-facing errors intentionally remain coarse so error text cannot become a hidden truth channel. A retry must preserve the `command_id`; a new command ID is a new action attempt.

This is the declared taxonomy. Some codes are reserved for validation paths and may not be reachable from every tool or domain today; clients should handle unknown future codes without treating their text as a data channel.

## Role boundary

MCP carries request/response control, actuation, and audit. Event evidence travels through the configured sink. This is a deliberate design decision documented in [`docs/research/TRANSPORT_ANALYSIS.md`](../../docs/research/TRANSPORT_ANALYSIS.md) and tested by the role-separation and operator endpoint tests.

## Next reads

- [Consumer integration](../guides/consumer-integration.md)
- [Security model](../architecture/security-model.md)
- [Truth and scoring](../architecture/truth-and-scoring.md)

## Run-wide limits

A world starts between 0 and 4611686018427387903 ns. `sim.clock.advance` refuses to move the
clock more than 30 days past the world's start, and `sim.perturb.apply` refuses a window whose
`until_ns` is set and before its `from_ns`. See `docs/design/SCENARIO_FILE.md` §3.
