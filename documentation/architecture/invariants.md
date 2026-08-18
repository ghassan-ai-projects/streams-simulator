# Nine non-negotiables

> Status: Implemented test gates with conditional release status. Authority: named tests and the design archive. Verified by: per-gate evidence below; full working-tree validation remains conditional. Last verified: 2026-08-17.

These guarantees distinguish a test instrument from a trace generator. The design may cut other features; it may not quietly cut these.

| Gate | Guarantee | Current evidence | Validation and scope |
| --- | --- | --- | --- |
| 1 | Determinism: a run is replayable from its declared inputs and command log. | [`internal/run/`](../../internal/run/), [`internal/world/`](../../internal/world/), CLI subprocess tests | `go test ./internal/run ./internal/cli`; deterministic local path |
| 2 | Analytic cross-check: the production integrator agrees with an independent oracle. | [`internal/world/oracle_test.go`](../../internal/world/oracle_test.go) | `go test ./internal/world -run Oracle`; oracle cross-check |
| 3 | Reference consumer: the product can complete its own consumer workflow. | [`internal/refconsumer/`](../../internal/refconsumer/), [`internal/mcp/operator_e2e_test.go`](../../internal/mcp/operator_e2e_test.go) | `go test ./internal/refconsumer ./internal/mcp`; loopback tests need socket permission |
| 4 | Delivery ledger: transport misses and reasoning misses remain distinguishable. | [`internal/run/ledger_test.go`](../../internal/run/ledger_test.go) | `go test ./internal/run -run Ledger`; ledger classification |
| 5 | Quiescence barrier: advancing time can wait for consumer work or fail incomplete. | [`internal/run/quiescence_test.go`](../../internal/run/quiescence_test.go) | `go test ./internal/run -run Quiescence`; consumer wait boundary |
| 6 | Sealed oracle: truth is role-scoped and cannot leak through the consumer surface. | [`internal/truth/truth_test.go`](../../internal/truth/truth_test.go), MCP tests | `go test ./internal/truth ./internal/mcp`; operator surface excludes truth |
| 7 | Injection probe: producer text cannot change the consumer conclusion. | [`internal/run/probe_test.go`](../../internal/run/probe_test.go) | `go test ./internal/run -run Probe`; built-in/reference-consumer path, not arbitrary external consumers |
| 8 | `silent_no_effect`: acknowledgement without world change is scored as such. | [`internal/score/score_test.go`](../../internal/score/score_test.go) | `go test ./internal/score -run Silent`; action outcome uses world evidence |
| 9 | Trivial-baseline audit: label-leaking scenarios do not enter the graded suite. | [`internal/audit/audit_test.go`](../../internal/audit/audit_test.go), [`internal/suite/`](../../internal/suite/) | `go test ./internal/audit ./internal/suite`; suite admission control |

## Evidence status

The repository contains named production-path tests for all nine gates. That is not the same as a current release claim: the working tree currently fails a shipped-domain count test because an untracked seventh domain is present. See [limitations](../limitations.md) and [release evidence](../operations/release.md).

## How to use this page

When changing a gate, start from the corresponding design section, identify the proving test, add an adversarial or metamorphic case where the risk warrants it, and update the status evidence. Do not replace an invariant with a weaker “similar” test without a recorded decision.

## Next reads

- [Quality bar](../QUALITY_BAR.md)
- [Benchmark methodology](../benchmark/methodology.md)
- [Contributor quality guide](../governance/quality.md)
