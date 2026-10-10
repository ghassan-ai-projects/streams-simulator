# Ubiquitous language — refconsumer

The reference consumer is a deliberately simple moving-window detector that
closes the loop against the simulator the way a real consumer would. It is
the benchmark's baseline: config-driven, with no domain knowledge.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Reference consumer | The shipped baseline detector and its runner | `Runner` | `streamsim-refconsumer` |
| Detector | A trailing-window z-score detector over each channel series | `Config` | `--threshold`, `--window` |
| Episode | A run of consecutive suspicious readings the detector declares | detection | `detections[]` |
| Silence | A channel quiet for several of its own periods | `AbsenceFactor` | absence detection |
| Nameplate | The static world description the consumer is given | `Nameplate` | `sim.nameplate.read` |
| Port | The narrow interface the consumer actuates and reports through | `EffectorInvoker`, `VerdictSink`, `QuiescenceReporter` | – |
| Operator client | The out-of-process client of the MCP operator endpoint, holding a capability token | `MCPOperator` | `--mcp`, `--token`, `--run` |
| Verdict | What the consumer reports: detections, admission outcomes, actions | `model.Verdict` | `verdict.json` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| consumer identity literals in three places | one set in the operator client and the verdict | single source (DEFERRED duplicates list) |
