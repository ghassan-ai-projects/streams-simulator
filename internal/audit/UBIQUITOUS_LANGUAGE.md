# Ubiquitous language — audit

The audit asks whether a scenario is trivially solvable: if a one-line
detector tuned with hindsight on the answer can already separate the fault
from its control, the scenario is not evidence that reasoning helped.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Panel | The set of one-line baseline detectors run against one injection | `Panel` | – |
| Trivial baseline | A detector so simple it needs no reasoning: fixed threshold, z-score, first difference, moving-median residual, channel silence | `DetectorNames` (domain layer) | `scores` keys |
| Hindsight fit | Fitting the detector's parameters on the scenario's own labeled series, deliberately unfair to the scenario | – | – |
| Balanced accuracy | The score of a detector separating fault series from control series | `Verdict.Scores`, `BalancedAccuracyCutoff` (0.9) | `scores`, `best_score` |
| Verdict | The audit outcome: trivial or not, per-detector scores, best detector, channels, sample count | `Verdict` | `trivial`, `best`, `scores` |
| Trivial | Some detector reaches the cutoff; the scenario stays a regression fixture but never counts as evidence | `Verdict.Trivial` | `trivial_excluded` in suites |
| Audit perturbation | A delivery perturbation applied to the audited stream exactly as the scenario declares it | `Perturbation` | `perturbations[]` |
| Control | The same world without the fault, run alongside | – | – |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `audit.DetectorNames`, `BalancedAccuracyCutoff` (exported) | domain-layer only | no caller outside the module (the manifest keeps its own literal, DEFERRED P-06/duplicates) |
