# Replay and debugging

> Status: Implemented deterministic replay workflow. Authority: run artifact, replay library, and CLI. Verified by: quickstart replay and run tests. Last verified: 2026-11-03.

Treat a run artifact as the primary failure report. It contains the inputs and command history needed to reproduce the run without an MCP server or a live consumer.

## Verify first

```bash
go run ./cmd/streamsim verify path/to/run.json
```

`verify` exits 1 with a `replay_mismatch` document when the run does not reproduce; `replay` prints the same report and always exits 0. If verification fails, preserve the complete JSON output and read `reason` first: `recorded_spec` or `world_digest` (the recorded domain or adapter no longer matches its digest), `trace_divergence` (with `first_divergence`), `journal_divergence`, `evidence_divergence`, `file_differs` or `file_missing` (with `file`), `digest_missing`, `incomplete` (the run closed without quiescence or after a sink failure and is never verified) or `incomplete_stamp`. `version_match` and `detail` report a simulator version difference, which alone does not fail. Then check the command log and, in the ledger, the count and terminal delivery reasons. The full list is in the [CLI reference](../reference/cli.md#replay-and-verify).
## Inspect the evidence set

A run folder holds `run.json`, `trace.jsonl`, `ledger.jsonl`, `world_state_history.jsonl`, the observer `journal.jsonl` and the truth-bearing `sealed/journal.jsonl` (plus `field-trace.jsonl` and `interrupts.jsonl` for a connected session). Use [the artifact reference](../reference/artifacts.md) to identify which file is authoritative for each question. Never infer transport loss from a trace without consulting the ledger. `streamsim report <run-dir> [--replay]` summarises the folder (identity, digests, whether the files still agree, delivery counts, faults, effector calls) as JSON or Markdown, and `streamsim domain diff` compares the domains of two runs.

## Compare experiments

Change one input at a time: seed, domain digest, adapter digest, fault, perturbation, effector command, or sink. Keep the original artifact and record the changed input in the issue or review. A different digest is expected when the experiment changed. `streamsim scenario from-run run.json --out scenario.json` exports the scenario that reproduces a run, so you can rerun it with one input changed, and `streamsim report compare --a <ref-run> --b <run>` lists the differing inputs and the comparability limits without ranking the runs.

## MCP failures

For operator runs, record the world ID, endpoint, capability-token handling (never the token itself), command IDs, and the sequence of clock advances. A missing command ID, wrong role, incorrect token, unquiesced advance, or unknown effector is a control-plane problem and must not be scored as consumer reasoning.

## Next reads

- [Determinism](../architecture/determinism.md)
- [Artifact reference](../reference/artifacts.md)
- [Troubleshooting](../operations/troubleshooting.md)
