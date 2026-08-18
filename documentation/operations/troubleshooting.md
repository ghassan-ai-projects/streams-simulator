# Troubleshooting

> Status: Current evidence-first troubleshooting guide. Authority: CLI errors, MCP error taxonomy, artifacts, and known limitations. Verified by: failure-mode review. Last verified: 2026-08-17.

Diagnose from evidence in this order: command output, run artifact, replay result, ledger, then world history and consumer verdict.

| Symptom | Likely boundary | First check |
| --- | --- | --- |
| `unknown domain` or `domain_invalid` | Input/catalog | `catalog list`, domain path, schema validation, digest. |
| `unknown adapter` or `adapter_invalid` | Adapter/input | `adapter list`, adapter schema, golden verification. |
| Replay digest mismatch | Determinism/input | Version, domain/adapter digest, command log, time mode, sink. |
| `consumer_not_quiesced` | Consumer/barrier | Consumer report, `await_consumer`, duplicate/stale report, endpoint health. |
| `capability_denied` | Operator authority | Use the token returned for that world; never substitute a world ID. |
| `missing_command_id` | Effector idempotency | Supply a stable non-empty `command_id`; preserve it on retry. |
| `unknown_effector` | Domain contract | Read the operator nameplate/effector list; do not invent names. |
| Trace is missing records | Delivery | Inspect `ledger.jsonl` before blaming the consumer. |
| Full test suite finds seven domains | Working-tree inventory | The untracked cold-chain domain conflicts with the six-domain assertion; see [limitations](../limitations.md). |
| `make ci-check` stops early | Tooling or code | Read the first failing target; missing release tools are intentional failures. |

## Escalation packet

When opening an issue, include the exact commit, command, sanitized artifact path, replay output, first divergence if any, ledger classification, and the smallest reproduction. Do not attach capability tokens or private receiver URLs.

## Next reads

- [Replay and debugging](../guides/replay-and-debugging.md)
- [Security model](../architecture/security-model.md)
- [Security policy](../../SECURITY.md)
