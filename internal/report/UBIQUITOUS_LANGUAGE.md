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
| Truth | What was done to the run (faults, perturbations, environment, effector calls with real outcomes); present only for a sealed or Explore run | `Truth` | `truth` |
| Verification | Whether the folder's files agree with the digests the run recorded; not a replay | `Verification`, `Check` | `verification.status`: `files_match`, `files_differ`, `not_checked` |
| Consumer | Who submitted the verdict; its `config_digest` appears only when the verdict recorded one | `Consumer` | `consumer` |
| Scoring | The offline scorecard, or why there is none | `Scoring` | `scoring.status`: `scored`, `not_scored`, `failed` |
| Format | The rendering of a report | `FormatJSON`, `FormatMarkdown` | `--format` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| verified (a report's) | `files_match` | A report does not replay; only `streamsim verify` proves a run |
