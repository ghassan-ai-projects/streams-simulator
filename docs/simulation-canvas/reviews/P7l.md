# P7l review record — factory utilities world, scale, KPI truth, last P7 findings

- **Reviewed**: `origin/main..0f84458` on `pre-ui-hardening` (2026-10-13), merged as PR #31 before the review completed
- **Reviewer**: one independent Sonnet subagent: 8 shipped domains × 48 h main-vs-head digest comparison, suites on main vs
  head, tracked-vs-untracked factory runs (13 h, five faults, dated effectors), four mutations, `-race` runs
- **Outcome**: 2 High, 2 Medium, 7 Low/Info. Fixed on `p7l-review-fixes` except as noted.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | High | The pulse recorder judged a recording before the clean oracle world had advanced to it: in a churning domain the clean world lacked the births, so the label fell at the onset (R7l-01) | Clean is advanced before faulted; `TestAPulseInAChurningDomainIsLabelledFromWorldsAtTheSameInstant` against an independent reference (fails with the old order) |
| 2 | High | Oracle worlds used the literal id `w-solver`, so their substreams differed from the run's: R7h-01 was only half fixed (R7l-02) | Oracle worlds take the derived identity a run uses; `TestTheOracleWorldIsBuiltAsARunBuildsItsWorld` (fails with the old id) |
| 3 | Med | A retirement less than one step before the close extended active alarm time to the clock (R7l-03) | The end of a retired entity is its retirement; its alarms are not standing; `TestARetiredEntitysAlarmEndsAtItsRetirementWhateverTheRunDoesAfter` |
| 4 | Med | Test gaps: sampling order at an instant, retirement, a longer invariance run | Retirement and chunking tests added; the invariance test runs six hours with hourly setpoint calls and a mid-run fault. Sampling before versus after the events of the same instant (mutation M1) is an equivalent mutant: every event at an instant acts by time (fault onsets, kicks) or schedules the new entity's own first sample, so no state read at that instant depends on the order; the rule stays for determinism of the record |
| 5 | Low | The sample at the end of an advance missed a fault injected at that instant (R7l-04) | Deferred to the next advance or the close; `TestAFaultInjectedRightAfterAnAdvanceIsInTheSampleAtThatInstant` (fails without the deferral) |
| 6 | Low | A fault clear before the clock rewrote sampled history (R7l-04) | Refused with `world.ErrFaultClearInPast` once the world has run, like an onset |
| 7 | Low | A retired entity's active alarm counted as standing | Excluded (finding 3); documented in §29 |
| 8 | Low | The pulse recorder's reads advance a `random_walk` drift (latent) | Open as R7l-05 (P12) |
| 9 | Low | Advancing oracle worlds processed every emission (k8s suite 34 s → 87 s) | Oracle worlds without a pulse schedule no emission (`world.Options.Unobserved`, R7l-06) |
| 10 | Low | CHANGELOG lacked the P7l changes | Added |
| 11 | Info | One `kpi.truth` record grows with every entity ever born | Noted; acceptable for a sealed end-of-run record |

Verified correct by the reviewer: the world performance change is bit-identical across all shipped domains; tracking KPIs
changes no emitted event (churn, faults, dated effectors); peak rate, flood and chattering rules; labels on main and head
identical for every shipped suite; `follow` sources; journal file digests and the report identity checks; no gate
loosened.
