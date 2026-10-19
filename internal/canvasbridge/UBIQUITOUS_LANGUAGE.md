# canvasbridge — ubiquitous language

| Term | Meaning |
|---|---|
| Bridge | The local HTTP server behind `streamsim canvas`; the canvas's only way to the simulator |
| Workspace | The folder the bridge serves: `domains/`, `compositions/`, `scenarios/`, `layouts/`, `kinds/`, `twins/`, `components/` (`equipment/`, `instruments/`), `runs/` |
| Twin choices | A `twins/*.twin.json` file: the instrument of each metric and the response of each command from which the `kind.twin` action scaffolds a twin into `domains/` |
| Kind stems | The `kinds` argument of `kind.twin` and `kind.suggest`: the stems of `kinds/<stem>.kind.json` that are in the twin, resolved by the bridge to workspace paths, repeats counted once; the whole `kinds/` folder is never passed to these actions (`ShapeStems`, `StemPath`, `ArgvPart.Existing`) |
| Pairing code | A one-time code the bridge prints on stderr; exchanged once for a session cookie |
| Session | The paired browser, identified by an `HttpOnly`, `SameSite=Strict` cookie |
| Admission | The check every API request passes: the bridge's host, its origin, a same-origin fetch |
| Action | A canvas operation in the registry, mapped to one director MCP tool or one streamsim command |
| Action input | The one argument of an action whose text goes on the command's standard input and nowhere else; `kind.check` and `kind.capture` have `content`; capped at 1 MiB (`Action.Input`, `Action.Stdin`) |
| One of | The arguments of which an action takes exactly one: `kind.capture` takes the text of a birth (`content`, on standard input) or a run (`run`, resolved to the observer journal `runs/<run>/journal.jsonl`, never to the sealed one), not both and not neither (`Action.OneOf`, `Action.StdinFlag`) |
| Issue | `{pointer?, message, keyword?, member?}`: where, in words, why for a program, and which named member; relayed unchanged from the command, never parsed by the bridge |
| Effector call | The canvas's `effector.list` and `effector.invoke` actions: the director lists a world's effectors and invokes one on an entity; the result is the operator acknowledgement, the capability token stays in the director |
| Connected session | A `plantwire connect` child the bridge owns, one at a time: `session.start` first runs the binary's plan check, then claims a fresh run folder `runs/c-<8 hex>` (a taken id is retried), and starts the child with the plan on standard input, `--barrier-cap-ms 10000` and its own process group; `session.status` never waits and says `none`, `starting`, `running`, `stopping`, `ended` or `failed` (with the command's error), always with the session's domain and the binary's check document; `session.stop` interrupts the child and waits up to 30 s, longer than the barrier cap, before it kills the group; the workspace listing marks the runs this bridge started `connected` (and `running` while the child lives); the bridge stops the session when it closes; it passes `--control-socket <path>` (a path in an owner-only directory it makes for the session and removes at the end), reports `controls` in the status while that socket exists, and `session.control` sends one node-control request line to it and returns the node's reply unchanged (failures `no_session`, `control_unavailable`, `control_gone`, `control_timeout`, `control_unreadable`) |
| Truth gate | The rule that keeps a run's sealed files from the canvas until the run ends (or from the start of an Explore run) |
| Follow | Streaming a run's observer journal to the page as Server-Sent Events, by line index |
| Version (etag) | A digest of a file's content that a write must name to replace it |
