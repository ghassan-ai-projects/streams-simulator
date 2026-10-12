# P2b evidence — operator truth leak

- **Date**: 2026-10-12
- **Change**: `sim.effector.invoke` returns `model.EffectorAcknowledgement` (six fields) instead of
  the world's internal result; acknowledging modes are byte-identical; a lost acknowledgement
  reads as a timeout; the reference consumer reads the same acknowledgement.

| Gate | Result |
|---|---|
| `make phase-gate` | Passed |
| Lint (Haiku 5.5 pass, then `lint-ci`) | 0 issues |
| Differential test | `TestEveryAcknowledgingModeReadsLikeSuccess`: ok, slow, partial, confirmed_no_effect and silent_no_effect responses are byte-identical; fails on the previous code |
| Allow-list | `TestTheOperatorResponseCarriesOnlyItsContractFields` over all seven modes |
| Findings | S-01 closed; R2b-01 carried |

Review: [reviews/P2b.md](../reviews/P2b.md).
