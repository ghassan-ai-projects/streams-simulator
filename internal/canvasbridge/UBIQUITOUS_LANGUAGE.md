# canvasbridge — ubiquitous language

| Term | Meaning |
|---|---|
| Bridge | The local HTTP server behind `streamsim canvas`; the canvas's only way to the simulator |
| Workspace | The folder the bridge serves: `domains/`, `compositions/`, `scenarios/`, `layouts/`, `runs/` |
| Pairing code | A one-time code the bridge prints on stderr; exchanged once for a session cookie |
| Session | The paired browser, identified by an `HttpOnly`, `SameSite=Strict` cookie |
| Admission | The check every API request passes: the bridge's host, its origin, a same-origin fetch |
| Action | A canvas operation in the registry, mapped to one director MCP tool or one streamsim command |
| Truth gate | The rule that keeps a run's sealed files from the canvas until the run ends (or from the start of an Explore run) |
| Follow | Streaming a run's observer journal to the page as Server-Sent Events, by line index |
| Version (etag) | A digest of a file's content that a write must name to replace it |
