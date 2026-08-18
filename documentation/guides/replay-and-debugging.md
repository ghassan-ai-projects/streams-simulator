# Replay and debugging

> Status: Implemented deterministic replay workflow. Authority: run artifact, replay library, and CLI. Verified by: quickstart replay and run tests. Last verified: 2026-08-17.

Treat a run artifact as the primary failure report. It contains the inputs and command history needed to reproduce the run without an MCP server or a live consumer.

## Verify first

```bash
go run ./cmd/streamsim verify path/to/run.json
```

If verification fails, preserve the complete JSON output. Check, in order:

1. simulator version mismatch;
2. domain or adapter digest mismatch;
3. command-log divergence;
4. trace digest mismatch and first divergence index;
5. incomplete/quiescence status;
6. ledger count and terminal delivery reasons.

## Inspect the evidence set

The normal file-sink output includes the run artifact and rendered trace; the run may also contain ledger and state-history files. Use [the artifact reference](../reference/artifacts.md) to identify which file is authoritative for each question. Never infer transport loss from a trace without consulting the ledger.

## Compare experiments

Change one input at a time: seed, domain digest, adapter digest, fault, perturbation, effector command, or sink. Keep the original artifact and record the changed input in the issue or review. A different digest is expected when the experiment changed.

## MCP failures

For operator runs, record the world ID, endpoint, capability-token handling (never the token itself), command IDs, and the sequence of clock advances. A missing command ID, wrong role, incorrect token, unquiesced advance, or unknown effector is a control-plane problem and must not be scored as consumer reasoning.

## Next reads

- [Determinism](../architecture/determinism.md)
- [Artifact reference](../reference/artifacts.md)
- [Troubleshooting](../operations/troubleshooting.md)
