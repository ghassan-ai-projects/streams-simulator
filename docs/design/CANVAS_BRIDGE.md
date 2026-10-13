# Canvas bridge (`streamsim canvas`) — protocol, threat model, module

- **Phase**: P8b. **Decision record**: `docs/simulation-canvas/canvas-technology.md` (Go bridge, HTTP JSON + Server-Sent
  Events). **Gates**: `governance/CANVAS_QUALITY_BAR.md` §3 (security row), `governance/CANVAS_UI_ACCEPTANCE.md`.
- The bridge is the canvas's only way to the simulator. It holds **no simulation, validation, composition, truth or
  scoring logic**: every action is a director MCP tool or a `streamsim` CLI command, run by the binary itself.

## 1. Command

```
streamsim canvas --workspace <dir> --ui-dir <built canvas> [--addr 127.0.0.1:0] [--adapters-dir adapters]
                 [--components-dir <dir>]...
```

It binds `127.0.0.1` only (a non-loopback `--addr` is refused), prints the address and a **pairing code** on stderr, and
serves until interrupted. The workspace is a folder:

| Path | Holds | Canvas access |
|---|---|---|
| `domains/*.domain.json` | Domains | read, write |
| `compositions/*.composition.json` | Compositions | read, write |
| `scenarios/*.scenario.json` | Scenario files | read, write |
| `layouts/*.layout.json` | Layout sidecars | read, write |
| `runs/<world_id>/…` | Run folders written by the director | read only, truth-gated (§4) |

Any other path, an absolute path, `..`, a symlink leaving the workspace or a file over 8 MiB is refused. A folder carrying
the `STREAMSIM_CUSTOMER_WORKSPACE` marker is served like any other, and must live outside every repository.

## 2. Protocol

All bodies are JSON. Errors are `{"error": {"code", "message", "issues"?}}` with the HTTP status of the class (400 invalid,
401 unpaired, 403 refused, 404 absent, 409 conflict, 413 too large, 502 binary failure).

| Method and path | Body / query | Returns |
|---|---|---|
| `GET /`, `GET /assets/*` | — | The built UI from `--ui-dir`, with a strict CSP |
| `GET /api/session` | — | `{paired}` |
| `POST /api/session` | `{code}` | `{paired: true}` and an `HttpOnly`, `SameSite=Strict` session cookie |
| `GET /api/workspace` | — | `{name, domains[], compositions[], scenarios[], layouts[], runs[]}`; a domain or composition entry is `{path, id, title}` (read from the file's top-level `id` and `title`, nothing else), a run `{world_id, ended, mode}` |
| `GET /api/file?path=` | — | `{path, content, etag}` |
| `PUT /api/file` | `{path, content, etag}` | `{path, etag}`; `etag` is the version the edit started from (`""` for a new file); a different file on disk is `409 conflict` with the current `etag` |
| `GET /api/actions` | — | The action registry: `{name, kind: "mcp" | "cli", tool \| command, truth}` per action |
| `POST /api/actions/{name}` | `{args}` | `{result}`: the tool's structured result or the command's JSON document |
| `GET /api/runs/{world_id}/events` | `Last-Event-ID` | Server-Sent Events: each observer journal record as event `record` (id = the record's `seq`), then `closed` |
| `GET /api/events` | — | Server-Sent Events: `workspace` when a workspace file changes (polled every second) |

The **action registry** is the parity list (GOAL item 7): every action names its director MCP tool or CLI command. CLI
actions build an argument vector from typed, workspace-relative arguments; no shell is involved. MCP actions go to one
`streamsim mcp --role director` child per bridge, started with the workspace's `domains/` and `runs/`. When a domain file
changes and no world is open, the child is restarted before the next catalog or world action so its catalog is current.

## 3. Threat model (one test per threat)

| # | Threat | Control |
|---|---|---|
| T1 | Another host on the network reaches the bridge | Binds loopback only; a non-loopback address is refused at start |
| T2 | A web page in the same browser drives the API (CSRF, DNS rebinding) | `Host` must be the bound loopback address; `Origin`, when sent, must be the bridge's origin; `Sec-Fetch-Site`, when sent, must be `same-origin` or `none`; the session cookie is `SameSite=Strict` |
| T3 | Another local process or user drives the API | The pairing code (60 bits, printed only on the bridge's stderr) is exchanged once for a session cookie; five wrong codes lock pairing until restart; codes are compared in constant time |
| T4 | The secret leaks through URLs or logs | The code and the session id never appear in a URL, a response body or a log line; the cookie is `HttpOnly` |
| T5 | Path traversal or symlink escape | Every file access goes through an `os.Root` at the workspace; only the paths of §1 are accepted |
| T6 | A write clobbers someone else's edit | Writes are atomic (temporary file and rename inside the root) and conditional on the `etag` the edit started from |
| T7 | Oversized requests exhaust memory | Bodies over 8 MiB (files) or 1 MiB (actions) are refused before they are read |
| T8 | Sealed truth reaches the page before sealing (Evaluate) | §4 |
| T9 | Command injection through action arguments | Arguments are typed and become separate `argv` entries; paths are workspace-relative and checked; the child runs the bridge's own executable with a scrubbed environment |
| T10 | Script injection in the served UI | `Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer` |

## 4. Truth boundary

- The live stream carries the **observer** journal only. The director stream (`sealed/journal.jsonl`), `run.json`, the
  ledger and the state history of a run are readable only once the run has ended (its `run.json` exists), or from the start
  for an Explore run (`run.opened` with `mode: explore`).
- The registry has no action for `sim.fault.list` and passes `unblind` to no tool: in Evaluate mode the canvas knows its own
  disturbances only from the commands it sent, and truth comes with the report after the run ends.

## 5. Module

`internal/canvasbridge`, a core with edges:

| Layer | Holds |
|---|---|
| facade | `Serve(ctx, Config)` |
| `internal/domain` | Pairing and session rules, request admission (T2), workspace path classes (§1), the truth gate (§4), the action registry and its argument-to-argv rules, entity tags |
| `internal/app` | Use cases: pair, list the workspace, read and write a file, run an action, follow a run, watch the workspace |
| `internal/transport` | `net/http` server, routes, Server-Sent Events, static files |
| `internal/files` | The workspace `os.Root`: reads, conditional atomic writes, journal tailing, change polling |
| `internal/process` | The director MCP child (MCP Go SDK client over stdio) and CLI commands of the bridge's own executable |
| `internal/entropy` | Pairing codes and session ids from `crypto/rand` |

Imports: `journal` (record reader), `kernel` is not needed, `model` for shared names; the MCP Go SDK client in
`internal/process` only. The CLI wires `streamsim canvas` to `canvasbridge.Serve`.
