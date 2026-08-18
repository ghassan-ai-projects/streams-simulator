# Artifact reference

> Status: Implemented reference. Authority: `internal/model/run.go`, `internal/run/run.go`, and the run-artifact schema. Verified by: quickstart output inventory and replay tests. Last verified: 2026-08-17.

An output directory is the evidence bundle for one run. The exact contents depend on the sink and whether a verdict was submitted, but the standard file-based path writes these files:

| File | Meaning | Authority |
| --- | --- | --- |
| `run.json` | Versioned run artifact with inputs, digests, command log, counts, expected trace digest, and completion state. | Replay and identity. |
| `trace.jsonl` | Adapter-rendered consumer-visible evidence. | What the consumer could read. |
| `ledger.jsonl` | Per-delivery terminal outcomes. | Transport classification and conservation. |
| `world_state_history.jsonl` | Director-side world state snapshots at emissions. | Post-hoc world analysis; never consumer evidence. |
| `verdict.json` | Consumer report when submitted during the run. | Consumer claim to score. |
| `label.json` | Ground-truth label supplied beside a run when offline scoring needs it. | External scoring input; not written by `run.End`. |

The standard run output is the five files through `verdict.json` in the table above. `label.json` is an external input. `streamsim score` prints a derived scorecard to standard output; it does not write a `scorecard.json` file. A reporting harness may save that output beside the run, but such a file is not a simulator contract.

`run.json` also carries the experiment identity and status fields needed for replay: `world_config`, `platform`, `world_digest`, `applied_perturbations`, `unblinded`, `error`, and nested `counts`. Consult the [run-artifact schema](../../docs/contracts/run-artifact-v0.1.schema.json) for exact field types and required properties.

## Artifact lifecycle

`run.json` is written by `run.End`. `streamsim verify` loads it and replays the command log. A successful replay must match the expected trace digest and input digests. An incomplete run remains incomplete; do not normalize it into a successful artifact.

## Handling rules

- Keep the artifact, trace, ledger, verdict, label, and manifest together.
- Do not edit a run artifact in place when investigating; copy it and record the change.
- Treat trace fields as consumer-controlled data for injection testing.
- Treat world history, truth, and ledger files as director-side evidence.
- Protect output directories if domains contain sensitive test data.

## Next reads

- [Determinism and replay](../architecture/determinism.md)
- [Replay and debugging](../guides/replay-and-debugging.md)
- [Contracts](contracts.md)
