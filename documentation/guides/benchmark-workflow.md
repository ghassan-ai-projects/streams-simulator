# Evaluation workflow

> Status: Partially automated workflow. Authority: suite, run, score, and MCP implementations. Verified by: CLI smoke tests and current limitations. Last verified: 2026-08-17.

This is the operational path for generating and scoring a consumer evaluation. It assumes the exact domain, profile, adapter, seed, and consumer version have been chosen before the run begins.

## 1. Generate an audited suite

```bash
go run ./cmd/streamsim suite --domain rotating-machinery --profile nominal --n 100 --seed 7 --out suites
```

The command reports the suite path, admitted scenarios, trivial exclusions, and terminal audit state. The suite records the domain, profile, seed, scenarios, and labels; it does not select an adapter or execute the scenarios. Do not treat requested scenario count as admitted count.

## 2. Run the consumer

This step is currently a harness responsibility, not a `streamsim suite` subcommand. Choose and record the adapter, run every admitted scenario, and keep the run artifact, trace, ledger, command log, and consumer output together. For an MCP consumer, use [consumer integration](consumer-integration.md). Copy the scenario label into the `label.json` input expected by offline scoring, or retain it in the director-side truth store.

## 3. Score offline

```bash
go run ./cmd/streamsim score --run /path/to/run.json --label /path/to/label.json
```

The current CLI reads `verdict.json` and `ledger.jsonl` beside the run artifact and uses `label.json` beside it unless `--label` is supplied. It prints the derived scorecard to standard output; a batch harness must persist that output. Offline CLI scoring does not receive an effector-call log, so action-loop metrics are incomplete; use the MCP director score path when action evidence is part of the benchmark.

## 4. Verify and preserve evidence

```bash
go run ./cmd/streamsim verify /path/to/run.json
```

Store the verification output and the manifest with the score. A score without replay identity and input digests is not a reproducible result.

## Next reads

- [Benchmark methodology](../benchmark/methodology.md)
- [Release evidence](../benchmark/evidence.md)
- [Troubleshooting](../operations/troubleshooting.md)
