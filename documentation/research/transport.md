# Transport research summary

The simulator keeps MCP and event evidence on different paths.

## Adopted

- MCP is used for request/response control, world creation, clock advancement, fault/perturbation injection, actuation, truth lifecycle, and audit.
- Sinks carry the stream evidence that a consumer is judged on.
- File and in-process sinks provide the strongest local replay path.
- The delivery ledger records terminal outcomes so a missing record is classifiable after the consumer phase.

## Rejected

Using MCP server-to-client push as the benchmark stream would make delivery timing depend on a best-effort control channel, and a lost message would be indistinguishable from the absence signal being tested. Putting raw producer text directly into an LLM host would also weaken the injection boundary.

## Source

See the detailed rationale in [`docs/research/TRANSPORT_ANALYSIS.md`](../../docs/research/TRANSPORT_ANALYSIS.md). Current implementation is in [`internal/mcp/`](../../internal/mcp/), [`internal/sink/`](../../internal/sink/), and [`internal/run/`](../../internal/run/).

## Next reads

- [Pipeline](../architecture/pipeline.md)
- [Security model](../architecture/security-model.md)
