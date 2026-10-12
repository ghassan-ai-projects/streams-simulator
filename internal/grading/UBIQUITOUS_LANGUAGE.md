# Ubiquitous language — grading

*Grading* is scoring a finished run offline, from the files it published, with
exactly the scorer the director uses online. The module composes `run` (replay)
and `score` (policies), which do not know each other.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Grade | Score a finished run from its published files and the sealed label | `Grade` | `score` command |
| Run artifact | The record of the run's inputs, command log and digests | `model.RunArtifact` | `run.json` |
| Verdict | What the consumer reported, checked against its contract | `model.Verdict` | `verdict.json` |
| Label | The sealed ground truth the run is graded against | `model.GroundTruthRecord` | `label.json` |
| Ledger | The delivery ledger, one JSON record per line | `DecodeLedger` | `ledger.jsonl` |
| State history | The director-only hidden-state snapshots, one per emission | `DecodeHistory` | `world_state_history.jsonl` |
| Replayed evidence | The effector calls, perturbations and counts recovered by replaying the artifact, because the command log records invocations but not whether each took effect | `ArtifactEvidence` | – |
| Replay mismatch | The refusal to grade a run whose replay does not reproduce its recorded trace | `ReplayMismatchError` | CLI `replay_mismatch` |
| Unblinded run | A run stamped unblinded (an explore run, or one whose truth was revealed while it was open); it is never scored | `ErrRunUnblinded`, `RequireBlind` | `unblinded`, `mode: "explore"` in `run.json`; CLI `score_failed` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `cli` score workflow | the `grading` module | loading, replaying and scoring a run are not command-line wiring |
