# Model Context Protocol (MCP) reference

> Status: Implemented reference. Authority: `internal/mcp/server.go`, `schemas.go`, and `operator.go`. Verified by: MCP strictness and operator tests. Last verified: 2026-08-17.

Streams Simulator exposes one MCP server with two role surfaces:

- **Director** — creates and drives worlds, injects faults and perturbations, owns truth, verifies artifacts, and scores runs.
- **Operator** — the consumer-facing capability surface for one world: read its nameplate, list/invoke effectors, and submit a verdict.

The director runs over stdio. A director may also serve a streamable HTTP operator endpoint. The operator endpoint resolves a world from the capability token, so one endpoint can serve multiple worlds.

## Start the server

```bash
go run ./cmd/streamsim mcp --role director --operator-addr 127.0.0.1:0 --out runs
```

The `sim.world.create` response includes the endpoint when configured and always includes the world capability token. Never commit or log the token.

## Director tools

| Tool | Purpose | Required identity |
| --- | --- | --- |
| `sim.catalog.list` | List domains and property vectors. | Optional `group`. |
| `sim.catalog.describe` | Describe one domain. | `domain`. |
| `sim.catalog.coverage` | Return axis coverage. | None. |
| `sim.adapter.list` | List installed adapters. | None. |
| `sim.world.create` | Create a seeded world and operator capability. | `domain`; `sink_target` for file/http-push sinks. |
| `sim.world.describe` | Report world configuration, clock, and emitted count. | `world_id`. |
| `sim.world.destroy` | End and remove a world. | `world_id`. |
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

The authoritative input schemas are defined and enforced beside the handlers in [`internal/mcp/schemas.go`](../../internal/mcp/schemas.go). Unknown properties are rejected at the MCP schema boundary; domain-dependent checks such as effector names and argument schemas happen in the handler.

## Operator tools

| Tool | Purpose | Required arguments |
| --- | --- | --- |
| `sim.nameplate.read` | Read static world identity, entities, channels, and effectors. | `token`. |
| `sim.effector.list` | Read declared effectors and argument schemas. | `token`. |
| `sim.effector.invoke` | Invoke a declared effector. | `token`, `effector`, `entity_id`, `command_id`; optional `args`, `at_ns`. |
| `sim.consumer.report` | Report quiescence and submit a verdict; never returns a score. | `token`, `run_id`; optional `quiesced_through_ns`, `verdict`. |

The operator view contains no truth store, fault registry, perturbation log, or hidden state field. Its narrow shape is a structural leak-prevention measure, not only a convention.

## Important argument fields and defaults

| Tool | Fields and defaults |
| --- | --- |
| `sim.world.create` | `domain` is required. `seed` defaults to `1`, `adapter` to `native-jsonl`, `sink` to `inproc`, and `time_mode` to `stepped`. `sink_target` is required for `file` and `http-push`. Optional `entities`, `scenario_profile`, `start_time`, and `label` are recorded in the world configuration. |
| `sim.clock.advance` | `world_id` plus exactly one of `by_ns` or `to_ns` is required. `by_ns` is non-negative. `await_consumer` defaults false; when true, the call waits for the consumer’s quiescence report or returns `consumer_not_quiesced`. |
| `sim.effector.invoke` | `token`, `effector`, `entity_id`, and unique `command_id` are required. `args` defaults to an empty object and is checked against the domain-declared argument schema. `at_ns` defaults to the current world clock. |
| `sim.consumer.report` | `token` and `run_id` are required. `quiesced_through_ns` is optional; `verdict` is optional when the consumer is reporting only quiescence. The response acknowledges acceptance and never includes a score. |
| `sim.truth.seal` | `run_id` and a complete `ground_truth` object conforming to `ground-truth-v0.1` are required. It must happen before `sim.run.begin`. |
| `sim.truth.reveal` | `run_id` is required. `unblind` defaults false; `unblind: true` permanently stamps the run and is allowed only after the scoring decision. |

## Minimal director/operator sequence

The identifiers come from earlier responses: `world_id`, `token`, and `run_id` are returned by world creation or run begin. The ground-truth record is created by the director, never by the consumer.

1. Call `sim.world.create` with a domain and, for a closed loop, a configured operator endpoint:

   ```json
   {"domain":"rotating-machinery","seed":7,"adapter":"native-jsonl","sink":"file","sink_target":"runs/w-1/trace.jsonl","time_mode":"stepped"}
   ```

   The response includes `world_id`, `world_digest`, `entity_ids`, `clock`, `token`, and `operator_endpoint` when HTTP operator serving is enabled.
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

`domain_invalid`, `adapter_invalid`, `world_not_found`, `clock_backwards`, `consumer_not_quiesced`, `profile_rate_exceeded`, `truth_sealed`, `run_unblinded`, `unknown_effector`, `invalid_args`, `missing_command_id`, `capability_denied`, `effector_refused`, `interlock_refused`, and `not_implemented`.

Operator-facing errors intentionally remain coarse so error text cannot become a hidden truth channel. A retry must preserve the `command_id`; a new command ID is a new action attempt.

This is the declared taxonomy. Some codes are reserved for validation paths and may not be reachable from every tool or domain today; clients should handle unknown future codes without treating their text as a data channel.

## Role boundary

MCP carries request/response control, actuation, and audit. Event evidence travels through the configured sink. This is a deliberate design decision documented in [`docs/research/TRANSPORT_ANALYSIS.md`](../../docs/research/TRANSPORT_ANALYSIS.md) and tested by the role-separation and operator endpoint tests.

## Next reads

- [Consumer integration](../guides/consumer-integration.md)
- [Security model](../architecture/security-model.md)
- [Truth and scoring](../architecture/truth-and-scoring.md)
