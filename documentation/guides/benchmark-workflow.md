# Evaluation workflow

> Status: Partially automated workflow. Authority: suite, run, score, and MCP implementations. Verified by: CLI smoke tests and current limitations. Last verified: 2026-11-03.

This is the operational path for generating and scoring a consumer evaluation. It assumes the exact domain, profile, adapter, seed, and consumer version have been chosen before the run begins.

## 1. Generate an audited suite

```bash
go run ./cmd/streamsim suite --domain rotating-machinery --profile nominal --n 100 --seed 7 --out suites
```

The command prints `{suite, scenarios, trivial_excluded, terminal_state}`: the suite path, admitted scenarios, trivial exclusions, and the terminal audit state. A small `--n` cannot meet the suite's coverage rules, and `terminal_state` then lists what is unmet; read it before using the suite. The suite records the domain, profile, seed, scenarios, and labels; it does not select an adapter or execute the scenarios. Do not treat requested scenario count as admitted count.

## 2. Run the consumer

This step is currently a harness responsibility, not a `streamsim suite` subcommand. Choose and record the adapter, run every admitted scenario, and keep the run artifact, trace, ledger, command log, and consumer output together. For an MCP consumer, use [consumer integration](consumer-integration.md). Copy the scenario label into the `label.json` input expected by offline scoring, or retain it in the director-side truth store.

## 3. Score offline

```bash
go run ./cmd/streamsim score --run /path/to/run.json --label /path/to/label.json
```

The CLI reads `verdict.json` beside the run artifact and uses `label.json` (or `labels.json`) beside it unless `--label` is supplied. It replays the artifact, so the effector calls, ledger and state history that decide action-loop metrics come from the replay, not from files a consumer could edit, and the scorecard agrees with the online one. It prints the scorecard to standard output; a batch harness must persist that output. For many one-fault runs, `suite campaign` generates the plans and labels; see the [CLI reference](../reference/cli.md#suite-campaign).

## 4. Verify and preserve evidence

```bash
go run ./cmd/streamsim verify /path/to/run.json
```

Store the verification output and the manifest with the score. A score without replay identity and input digests is not a reproducible result.

## Next reads

- [Benchmark methodology](../benchmark/methodology.md)
- [Release evidence](../benchmark/evidence.md)
- [Troubleshooting](../operations/troubleshooting.md)
