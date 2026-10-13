# Ubiquitous language — lint

The lint answers one question about a domain that already validates: does a
nominal day of it behave plausibly? It simulates the default population with
no faults and no effector calls, reads every state and numeric channel
noise-free on a fixed grid and judges the series with a few rules.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Nominal day | The default population from the world start for the requested hours, noiseless, with no fault and no effector call | `Check` | `hours`, `seed` |
| Sample grid | Every 60 s from the world start, plus the end of the run | `grid` | `samples` |
| Series | The sampled values of one state (`StateValue`) or one numeric channel (noise-free `Reading`) of one entity | `series` | – |
| Finding | One judgement about one series, or about one conservation balance of one entity | `Finding` | `findings[]` |
| Severity | `error` fails the command, `warning` is reported only | `Finding.Severity` | `severity` |
| Report | The whole outcome: domain identity, sampling, findings, counts | `Report` | `domain_lint` output |
| Idle state | A state that never moves over the run; it is not pinned, it is unused in a nominal day | `pinnedAtBound` | – |
| Cyclic state | An integrated state driven, through its inputs, by a seasonal or scheduled source; it cycles daily and is not expected to settle | `cyclic` | – |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| validate (for this check) | lint | `domain validate` checks the document against the schema and cross references; lint judges simulated behaviour |
