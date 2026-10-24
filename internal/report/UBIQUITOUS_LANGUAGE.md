# Ubiquitous language — report

A *report* summarises one run folder as a single deterministic document, in JSON
(`run-report-v0.1`) or Markdown. It is a pure function of the folder's files. The
module composes `run` (artifact loading), `journal` (stream reading and digests) and
`grading` (scoring).

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Report | The document describing a run folder | `Report` | `run-report-v0.1`, `streamsim report` |
| Run folder | The directory `run --out` publishes: `run.json`, `trace.jsonl`, `journal.jsonl`, `sealed/journal.jsonl`, `ledger.jsonl`, optionally `verdict.json` and `label.json` | `Folder` | – |
| Document | One file of the folder: its bytes and whether it exists | `Document` | – |
| Ended | The run ended and wrote `run.json`, so its truth may be read | `Run.Ended` | `run.ended` |
| Complete | The run ended and its artifact is not marked incomplete | `Run.Complete` | `run.complete`, `run.incomplete_reason` |
| Name and purpose | The operator's authored plain text for the run, read from the artifact's `plan`; shown beside the run's identity, never a measurement or a verdict | `Run.Name`, `Run.Purpose` | `run.name`, `run.purpose` |
| Truth | What was done to the run (faults, perturbations, environment, effector calls with real outcomes); present only for a sealed or Explore run | `Truth` | `truth` |
| Verification | Whether the folder's files agree with the digests the run recorded; not a replay | `Verification`, `Check` | `verification.status`: `files_match`, `files_differ`, `not_checked` |
| Consumer | Who submitted the verdict; its `config_digest` appears only when the verdict recorded one | `Consumer` | `consumer` |
| Scoring | The offline scorecard, or why there is none | `Scoring` | `scoring.status`: `scored`, `not_scored`, `failed` |
| Format | The rendering of a report or comparison | `FormatJSON`, `FormatMarkdown` | `--format` |
| Replay | Whether the recording was replayed and reproduced the run; `not_checked` unless `--replay` asked for it | `Replay`, `ReplayReproduces`, `ReplayDiverges`, `ReplayIncomplete` | `replay.status` |
| Simulated clock text | A span of simulated time written in a human sentence as `hh:mm:ss` rounded to the second (hours not wrapped), the form the canvas shows; machine fields keep nanoseconds and table cells that pair an instant with its offset keep `(+Ns)` | `clockText` | `open_questions[].message`, `comparability_limits[].message` |
| Open question | Something a reader cannot conclude from the run, asked by a rule of the binary in a fixed order | `Question`, `questionsOf` | `open_questions[]` (`code`, `message`) |
| Handover | The Markdown report read by someone who did not attend the run: purpose, what ran, outcome and integrity as separate rows, evidence coverage, results, plan, open questions | `Markdown` | `--format markdown` |
| Comparison | Two ended runs set side by side, run A the reference: inputs, plan difference, comparability limits, whole-run results; it never ranks runs | `Comparison`, `Compare` | `run-comparison-v0.1`, `streamsim report compare` |
| Plan difference | The faults, perturbations, environment faults and effector calls of two runs as identical, different (naming the fields), only in A or only in B | `Diff`, `Changed` | `plan` |
| Comparability limit | A reason the two runs are not like for like, as a stable code, an optional side and a sentence | `Limit` | `comparability_limits[]` |
| Presence | Whether a result row exists in both runs or in one | `PresenceBoth`, `PresenceOnlyA`, `PresenceOnlyB` | `presence` |
| Scope | What a result covers: `whole_run`, each run over its own simulated length; never an event window | `Scope` | `scope`, `windowed_results` |
| Not ended | A run to compare has no `run.json` or its truth is unreadable, so the comparison is refused | `ErrNotEnded` | `run_not_ended` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| verified (a report's) | `files_match` | A report does not replay; only `streamsim verify` proves a run |
