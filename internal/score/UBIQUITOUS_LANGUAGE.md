# Ubiquitous language — score

Scoring grades one run from its evidence and the sealed label. The scorer
reads only what the simulator generated or the consumer submitted through
`sim.consumer.report`; it never reads a consumer's database or logs.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Scorecard | The per-run scoring output, JSON-stable | `Scorecard` | `scorecard.json`, bundle `scorecard-bundle-v0.2` |
| Evidence | Everything the scorer reads about one run, packed by the host | `Evidence` | – |
| Bundle | The version of the scoring policies shared by the online and offline paths | `scoringBundleVersion` | `bundle` |
| Instrument metrics | The simulator grading itself: ledger completeness, perturbation fidelity, effector idempotency, delivery counts | `InstrumentMetrics` | `instrument` |
| Consumer metrics | What the consumer did with delivery faults: duplicates, identity conflict, lateness, dropped events, clock skew, evidence grounding, action fidelity, interlock handling | `ConsumerMetrics` | `consumer` |
| Loop metrics | What only a closed loop can measure: resolution, deadline, unnecessary action, false success | `LoopMetrics` | `loop` |
| Judgment metrics | Detection quality against truth: detected, label correct, latency, false positive | `JudgmentMetrics` | `judgment` |
| Online / offline | Scoring a live run's evidence versus scoring artifacts alone; offline cannot compute loop metrics from history | `Score`, `Offline` | – |
| Hidden-state history | The director-only per-emission state snapshots the loop metrics read | `model.StateSnapshot` | `world_state_history.jsonl` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `Score(*run.Run, …)` | `Score(Evidence, …)` | the scorer no longer imports `run`; the host packs the evidence |
| `Scorecard.Marshal` | removed | no caller; callers marshal the record |
