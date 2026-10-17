# canvasbridge — ubiquitous language

| Term | Meaning |
|---|---|
| Bridge | The local HTTP server behind `streamsim canvas`; the canvas's only way to the simulator |
| Workspace | The folder the bridge serves: `domains/`, `compositions/`, `scenarios/`, `layouts/`, `kinds/`, `twins/`, `components/` (`equipment/`, `instruments/`), `runs/` |
| Twin choices | A `twins/*.twin.json` file: the instrument of each metric and the response of each command from which the `kind.twin` action scaffolds a twin into `domains/` |
| Pairing code | A one-time code the bridge prints on stderr; exchanged once for a session cookie |
| Session | The paired browser, identified by an `HttpOnly`, `SameSite=Strict` cookie |
| Admission | The check every API request passes: the bridge's host, its origin, a same-origin fetch |
| Action | A canvas operation in the registry, mapped to one director MCP tool or one streamsim command |
| Effector call | The canvas's `effector.list` and `effector.invoke` actions: the director lists a world's effectors and invokes one on an entity; the result is the operator acknowledgement, the capability token stays in the director |
| Truth gate | The rule that keeps a run's sealed files from the canvas until the run ends (or from the start of an Explore run) |
| Follow | Streaming a run's observer journal to the page as Server-Sent Events, by line index |
| Version (etag) | A digest of a file's content that a write must name to replace it |
