# Operations guide

> Status: Implemented local operations guidance. Authority: run/artifact lifecycle and sink implementations. Verified by: artifact and sink tests. Last verified: 2026-08-17.

Streams Simulator is a local instrument. Operations are about preserving evidence, controlling egress, and stopping safely when a run is incomplete or ambiguous.

## Run directory layout

Use a dedicated output directory per world or scenario. A normal file-sink run produces:

```text
runs/<world-or-run>/
├── run.json
├── trace.jsonl
├── ledger.jsonl
├── world_state_history.jsonl
└── verdict.json       optional
```

Keep the label/ground truth and release manifest beside the run when doing an evaluation. Use restrictive permissions when data is not public.

## Safe lifecycle

1. Inspect the catalog and adapter.
2. Create the world with an explicit seed, adapter, sink, and output directory.
3. Seal truth before the consumer phase.
4. Advance time through the declared clock surface.
5. Await consumer quiescence where required.
6. End the run and persist the artifact.
7. Verify replay before scoring.
8. Reveal/unblind and score only at the declared transition.

If a consumer does not quiesce, the run is incomplete. Preserve it as evidence and diagnose the cause; do not retry until the artifact looks successful.

## Egress

`http-push` is an explicit network boundary. Use `inproc` or `file` for deterministic local work and use a controlled receiver for integration tests. Do not put credentials in adapter or domain files.

## Cleanup

Run artifacts are ordinary files. Archive the complete evidence bundle before removing a run. Never delete a run that is referenced by a published score or incident without retaining a recoverable copy.

## Next reads

- [Troubleshooting](troubleshooting.md)
- [Release procedure](release.md)
- [Artifact reference](../reference/artifacts.md)
