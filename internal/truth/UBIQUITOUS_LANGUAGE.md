# Ubiquitous language — truth

Ground truth is generated, not asserted: the module reconstructs when a fault
becomes observable by running the world twice, and keeps the label sealed
from the operator role until the run closes.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Ground truth record | The sealed label of one scenario: what was injected and when it can be seen | `model.GroundTruthRecord`, built by `BuildRecord` | `label.json`, ground-truth-v0.1 |
| Injection | The facts a label is built from: scenario, seed, entity, fault, onset, entities, setup | `Injection` | – |
| Setup call | A pre-fault effector invocation applied to both oracle worlds so the scenario context exists | `model.SetupCall` | `setup[]` |
| Solver | The oracle: runs a clean and a faulted noiseless world and scans the detector quantity | `Solver` | – |
| Injection time | When the fault was applied | `InjectionTimeNS` | `injection_time_ns` |
| First observable | The first time the detector quantity crosses the declared first-observable SNR | `FirstObservableTimeNS` | `first_observable_time_ns` |
| Unavoidable | The first time it crosses the unavoidable SNR: no consumer could have missed it | `UnavoidableTimeNS` | `unavoidable_time_ns` |
| Solution method | How the onsets were found (numeric scan) | `SolutionMethod` | `solution_method` |
| Seal | Recording a label for a run so only the director can read it | `Store.Seal` | – |
| Reveal | Reading a sealed label; refused on an open run unless unblinded | `Store.Reveal` | – |
| Unblind | Revealing on an open run, which stamps the run and excludes it from every scorecard | `unblind` argument | `unblinded` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `OpenChecker` (settable hook) | `NewStore(runIsOpen)` | a required constructor argument; no unwired store can reveal |
| `Result`, `Solve` (exported) | unexported in the domain layer | only `BuildRecord` used them |
| `truth.SetupCall` | `model.SetupCall` | shared by `audit`, `suite` and `truth`; a plain record |
