# Ubiquitous language — mcp

The MCP module is the simulator's protocol surface: one server, two roles.
It frames the tools of the `director` module: tool definitions and closed
schemas, argument decoding from tool calls, the wire form of a refusal, and
the stdio and HTTP transport. It decides nothing about a world, a run, truth or
scoring (`TestMCPHoldsOnlyProtocol` fails when it regrows logic). The director
role runs the world; the operator role is the only door the system under test
has, and it never sees truth.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Director | The role that creates worlds, advances the clock, injects faults and perturbations and reads sealed truth; its logic is the `director` module | `NewDirectorServer` | `sim.*` director tools |
| Operator | The role the consumer uses: nameplate, effectors, verdict; no truth, no ledger | `NewOperatorServerResolver` | `sim.nameplate.read`, `sim.effector.*`, `sim.consumer.report` |
| Capability token | Opaque unguessable bearer value that names exactly one world to the operator role; minted by the `director` module | `capability.NewToken` (director) | `token`, `t-…` |
| Argument decoding | Reading a tool call's JSON argument map into the typed values the director takes | `stringArg`, `numberArg`, `intArg`, `worldRequest` | tool arguments |
| Serving | Running the director role over stdio until the session ends, with the operator role optionally bound over streamable HTTP; the `transport` edge holds the listener and the session | `ServeDirector`, `transport.Serve` | `mcp --role director --operator-addr` |
| Tool | One named, schema-validated MCP operation; unknown properties are refused | `director_tools.go`, `operator_tools.go` | tool name |
| Stable error code | The machine-readable refusal class carried in the tool error text; the taxonomy belongs to the director | `director.Code` | `code` in error content |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `internal/mcp` flat package symbols | the facade (`ServeDirector`, `NewDirectorServer`, `NewOperatorServerResolver`) | handlers and schemas moved behind `internal/mcp/internal/protocol` |
| `mcp.Director`, `mcp.NewDirector`, `mcp.SetOperatorEndpoint` | `director.Director`, `director.New`, `director.SetOperatorEndpoint` | the director's logic is its own module; `mcp` takes one |
| `capabilityToken` | `capability.NewToken` in the director module | entropy is an I/O edge of the director, not a protocol concern |
| `serve.Director` (in `cli`) | `mcp.ServeDirector` | serving MCP is the protocol module's work, not command-line wiring |
