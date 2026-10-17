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
| `kinds/*.kind.json` | Kind files: the interface of a device (plant wire `NBIRTH.kinds` entry) | read, write |
| `components/equipment/*.component.json` | Workspace equipment components | read, write |
| `components/instruments/*.instrument.json` | Workspace instrument components | read, write |
| `runs/<world_id>/…` | Run folders written by the director | read only, truth-gated (§4) |

`components/` is a customer component directory in the library's own layout (`equipment/` and `instruments/` below it,
as `--components-dir` expects), so the bridge hands that folder to every command that loads the library. `kinds/` is a
kind directory (`--kinds-dir`). Neither folder is created at start: a workspace without `kinds/` or `components/` passes
no flag for it, and a domain that `implements` a kind file then fails closed, because its kind file is not loaded.

Any other path, an absolute path, `..`, a symlink leaving the workspace or a file over 8 MiB is refused. A folder carrying
the `STREAMSIM_CUSTOMER_WORKSPACE` marker is served like any other, and must live outside every repository.

## 2. Protocol

All bodies are JSON. Errors are `{"error": {"code", "message", "issues"?, "result"?}}` (`result` is the command's own result document when it
failed after checking, e.g. `run.verify` with `matches: false`) with the HTTP status of the class (400 invalid,
401 unpaired, 403 refused, 404 absent, 409 conflict, 413 too large, 422 a director tool or streamsim command refused the
request with its own `code`, `message` and `issues`, 500 the bridge itself failed). Any other `GET /api/...` path is
`404 not_found` (never the UI) and passes the same admission as the API. A conflict from a write or from `domain.compose`
and `scenario.export` also carries the file's current `etag`. A message never names the absolute workspace folder: the bridge
removes it from every error and result it relays.

| Method and path | Body / query | Returns |
|---|---|---|
| `GET /`, `GET /assets/*` | — | The built UI from `--ui-dir`, with a strict CSP |
| `GET /api/session` | — | `{paired}` |
| `POST /api/session` | `{code}` | `{paired: true}` and an `HttpOnly`, `SameSite=Strict` session cookie that lasts 12 hours (`Max-Age`), so a browser restart does not need a new code; the code itself stays single use and the session lives in the bridge's memory only; the cookie is named `streamsim_canvas_<port>` after the bridge's bound port, because browsers share cookies across ports of one host and two bridges on `127.0.0.1` would otherwise unpair each other |
| `GET /api/workspace` | — | `{name, domains[], compositions[], scenarios[], layouts[], kinds[], components[], runs[]}`; a domain or composition entry is `{path, id, title}` (read from the file's top-level `id` and `title`, nothing else), a kind entry `{path, kind, title?}` (the file's top-level `kind` text, and `title` when present), a component entry `{path, id, title, class}` (`id` and `title` from the file, `class` is `"equipment"` or `"instrument"` from its folder); a file that does not parse is listed with its `path` (and a component's `class`) only; `kinds` and `components` are `[]` when the folders are absent; a run `{world_id, ended, complete, incomplete_reason, explore, seed, plant, duration_ns, faults, recorded_at}` (`explore` true for an Explore run, false otherwise; `complete` true only for a run that ended without being abandoned or aborted, with `incomplete_reason` the run's recorded reason and `""` when complete; `seed` the run's seed from `run.json`, `null` until the run has a `run.json`) |
| `GET /api/file?path=` | — | `{path, content, etag}` |
| `PUT /api/file` | `{path, content, etag}` | `{path, etag}`; `etag` is the version the edit started from (`""` for a new file); a different file on disk is `409 conflict` with the current `etag`; content over 8 MiB is `413` (the write limit is the read limit) |
| `GET /api/actions` | — | The action registry: `{name, kind: "mcp" | "cli", tool \| command, truth}` per action |
| `POST /api/actions/{name}` | `{args}` | `{result}`: the tool's structured result or the command's JSON document |
| `GET /api/runs/{world_id}/events` | `Last-Event-ID` | Server-Sent Events: each observer journal record as event `record` (id = the record's line index), then `closed`. `world_id` is one run id (letters, digits, `.`, `_`, `-`; an encoded `/` is `400 bad_path`) and always names the observer journal; a run folder that does not exist is `404`. Resuming with the id of the closing record sends `closed` alone |
| `GET /api/events` | — | Server-Sent Events: `workspace` when a file of `domains/`, `compositions/`, `scenarios/`, `layouts/`, `kinds/`, `components/equipment/`, `components/instruments/` or `runs/` changes (polled every second) |

The **action registry** is the parity list (GOAL item 7): every action names its director MCP tool or CLI command. CLI
actions build an argument vector from typed, workspace-relative arguments; no shell is involved. The device actions are
`kind.validate` (`streamsim kind validate`, one `kinds/*.kind.json` file) and `component.validate` (`streamsim component
validate`, the whole library: built-in, `--components-dir` and the workspace's `components/`). The bridge, never the canvas,
supplies the workspace folders: every library action (`component.list`, `component.show`, `component.validate`,
`domain.compose`) gets `--components-dir <workspace>/components` when that folder exists, and `domain.validate` and
`domain.compose` get `--kinds-dir <workspace>/kinds` when it exists, so the "behaviour implements interface" check runs
against the workspace's own kind files. A kind file that does not parse stops `domain.validate` too, as it stops the CLI.
The director child's `world.create` does not yet check `implements`; that check stays in `streamsim run`. MCP actions go to one
`streamsim mcp --role director` child per bridge, started with the workspace's `domains/` and `runs/`. When a domain file
changes and no world is open, the child is restarted before the next catalog or world action so its catalog is current. When the child is gone (a transport
error, not a refusal from a tool) the bridge drops it, forgets its worlds and retries the call once on a new child; the
child's stderr tail (it names, for instance, an invalid domain file) is part of the failure message. `domain.compose` and
`scenario.export` take an optional `etag` for their `out` file: an existing `out` whose version differs (or that has no
`etag` at all) is `409 conflict` with the current `etag`, checked before the command runs.

Concurrency: director calls are serialized (one child, one call at a time); pairing and session checks, file reads and
writes, listings and streams use their own short locks and never wait behind a running `clock.advance`.

## 3. Threat model (one test per threat)

| # | Threat | Control |
|---|---|---|
| T1 | Another host on the network reaches the bridge | Binds loopback only; a non-loopback address is refused at start |
| T2 | A web page in the same browser drives the API (CSRF, DNS rebinding) | `Host` must be the bound loopback address; `Origin`, when sent, must be the bridge's origin; `Sec-Fetch-Site`, when sent, must be `same-origin` or `none`; the session cookie is `SameSite=Strict` |
| T3 | Another local process or user drives the API | The pairing code (60 bits, printed only on the bridge's stderr) is exchanged once for a session cookie; five wrong codes lock pairing until restart; codes are compared in constant time |
| T4 | The secret leaks through URLs or logs | The code and the session id never appear in a URL, a response body or a log line; the cookie is `HttpOnly` |
| T5 | Path traversal or symlink escape | Every file access goes through an `os.Root` at the workspace; only the paths of §1 are accepted |
| T6 | A write clobbers someone else's edit | Writes are atomic (temporary file and rename inside the root) and conditional on the `etag` the edit started from |
| T7 | Oversized requests exhaust memory | A body is read through a limit of 8 MiB plus JSON overhead (files) or 1 MiB (actions) and refused when it exceeds it; run files and writes over 8 MiB are `413` (a long run's `trace.jsonl` or `sealed/journal.jsonl` therefore cannot be read whole; the live stream has no such limit) |
| T8 | Sealed truth reaches the page before sealing (Evaluate) | §4 |
| T9 | Command injection through action arguments | Arguments are typed and become separate `argv` entries; paths are workspace-relative and checked; the child runs the bridge's own executable with a scrubbed environment |
| T10 | Script injection in the served UI | `Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer` |

## 4. Truth boundary

- The live stream carries the **observer** journal only, and its run id is one validated segment: no path can be built
  from it, so it cannot name `sealed/journal.jsonl`. The director stream (`sealed/journal.jsonl`), `run.json`, the ledger
  and the state history of a run are readable only once the run has **ended**, or from the start for an Explore run
  (`run.opened` with `mode: explore`). A run has ended when the last record of its observer journal is `run.closed` and the
  bridge no longer counts it as running (`run.end` or `world.destroy` succeeded); a file that exists (`run.json`) proves
  nothing, because a folder can hold files of an earlier run.
- The director never reuses a run folder: a world takes the next id whose `<out>/<world_id>` folder does not exist, so a
  restarted director (ids restart at `w-1`) cannot inherit, append to or unlock the files of an earlier run.
- Kind files and components are definitions, readable and writable at any time: a kind file carries no faults, onsets or
  truth (PHYSICAL_DEVICES §1.3), and the device actions return validation results only, never run state.
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
| `internal/process` | The director MCP child (MCP Go SDK client over stdio, its stderr tail kept, a lost connection reported as `ErrGone`) and CLI commands of the bridge's own executable |
| `internal/entropy` | Pairing codes and session ids from `crypto/rand` |

Imports: `journal` (record reader), `kernel` is not needed, `model` for shared names; the MCP Go SDK client in
`internal/process` only. The CLI wires `streamsim canvas` to `canvasbridge.Serve`.
