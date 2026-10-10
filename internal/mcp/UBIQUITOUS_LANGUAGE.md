# Ubiquitous language — mcp

The MCP module is the simulator's protocol surface: one server, two roles.
The director role runs the world; the operator role is the only door the
system under test has, and it never sees truth.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Director | The role that creates worlds, advances the clock, injects faults and perturbations and reads sealed truth | `Director`, `NewDirectorServer` | `sim.*` director tools |
| Operator | The role the consumer uses: nameplate, effectors, verdict; no truth, no ledger | `NewOperatorServerResolver` | `sim.nameplate.read`, `sim.effector.*`, `sim.consumer.report` |
| Capability token | Opaque unguessable bearer value that names exactly one world to the operator role | `capability.NewToken` | `token`, `t-…` |
| Operator endpoint | The URL a director advertises for operators to connect to | `SetOperatorEndpoint` | `operator_endpoint` |
| Tool | One named, schema-validated MCP operation; unknown properties are refused | `director_tools.go`, `operator_tools.go` | tool name |
| Stable error code | The machine-readable refusal class carried in the tool error text | `errors.go` | `code` in error content |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `internal/mcp` flat package symbols | the facade (`Director`, `NewDirectorServer`, `NewOperatorServerResolver`, `SetOperatorEndpoint`) | handlers and schemas moved behind `internal/mcp/internal/app` |
| `capabilityToken` | `capability.NewToken` | entropy is an I/O edge, not a use-case concern |
