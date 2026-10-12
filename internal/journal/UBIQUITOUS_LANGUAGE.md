# Ubiquitous language — journal

The *journal* is the record of a run that grows while the run happens. It is two
append-only JSON Lines streams, so a person or agent can watch a run, and anyone
can investigate one that crashed, without the consumer side ever seeing sealed
truth (docs/design/RUN_JOURNAL.md).

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Journal | The pair of streams of one run | `Journal` | `journal.jsonl` and `sealed/journal.jsonl` |
| Observer stream | What a consumer could know: identity, delivered observations, clock advances, effector acknowledgements, consumer reports, the close | `StreamObserver` | `journal.jsonl`, `"stream": "observer"` |
| Director stream | What makes the run true: commands, delivery rows, world state, effector outcomes, unblinding, the close | `StreamDirector` | `sealed/journal.jsonl`, `"stream": "director"` |
| Record | One line: the envelope (`stream`, `seq`, `type`, `at_ns`) and the fields of its type | `Record` | one line of `run-journal-v0.1` |
| Truth-bearing record | A record that would tell a consumer something the world has not shown it; only the director stream holds them | – | – |
| Replayable record | A record replay can rebuild from the command log; it takes the next `seq` and enters the stream digest | `IsSession` is false | – |
| Session record | An out-of-band report or live control that is not in the command log (`consumer.reported`, `pacing`); it is in the file but outside the stream digest | `IsSession` | `consumer.reported` |
| Stream digest | SHA-256 of the replayable lines of a stream | `Stream.Digest`, `Digest` | `journal_digest`, `sealed_journal_digest` in `run.json` |
| Torn line | An unterminated last line left by a crash; readers ignore it | `Read` | – |
| Explore mode | A run unblinded from creation; it is never scored | `ModeExplore` | `run.opened.mode`, `mode` in `run.json` |
| Incomplete run | A journal that stops without `run.closed` (a crash) or closes with `incomplete: true` | – | `run.closed` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| trace (for the journal) | journal | the trace is the consumer's delivered bytes (`trace.jsonl`); the journal also records commands, state and acknowledgements |
| sealed journal (for the observer stream) | director stream | only the director stream is sealed; the observer stream is readable by the consumer |
