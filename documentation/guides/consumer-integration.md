# Consumer integration

> Status: Implemented file and MCP integration paths. Authority: CLI/refconsumer and MCP operator code. Verified by: reference-consumer and operator tests. Last verified: 2026-08-17.

The reference consumer demonstrates the supported boundary: evidence arrives through a sink, while control and actuation happen through the operator MCP surface.

## File-based consumer path

For a local trace, the reference consumer reads native JSONL and writes a verdict:

```bash
go run ./cmd/streamsim refconsumer --trace /path/to/trace.jsonl --out /path/to/verdict.json
```

The detector parameters are explicit flags (`--threshold` and `--window`). This path is useful for smoke tests and offline scoring; it does not close the action loop unless the MCP options are supplied.

## Closed-loop operator path

Start a director with an operator endpoint:

```bash
go run ./cmd/streamsim mcp --role director --operator-addr 127.0.0.1:0 --out runs
```

The director serves control over stdio and the operator role over streamable HTTP. `sim.world.create` returns the operator endpoint and a capability token. The consumer uses that token to read the nameplate, list effectors, invoke a declared effector with a unique `command_id`, and submit its verdict.

The operator role is not a separate standalone process. A request for `--role operator` is rejected; the director owns world state and resolves the token to the corresponding operator view.

## Consumer obligations

- Consume only the delivered adapter output for its evidence.
- Treat event time and observed time according to the native/adapter contract.
- Keep command IDs stable across retries.
- Wait for the quiescence boundary when the workflow requires it.
- Report detections and actions through the declared verdict contract.
- Never rely on hidden truth, director-only resources, or acknowledgement alone to claim an effect.

## Integration tests

The out-of-process operator endpoint and closed-loop behavior are exercised in [`internal/mcp/operator_e2e_test.go`](../../internal/mcp/operator_e2e_test.go). The reference consumer implementation is under [`internal/refconsumer/`](../../internal/refconsumer/).

## Next reads

- [MCP reference](../reference/mcp.md)
- [Truth and scoring](../architecture/truth-and-scoring.md)
- [Artifacts](../reference/artifacts.md)
