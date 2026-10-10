# Plan and status

Branch `harden-weak-spots`, from `main` after PR #18. Tiers run in order;
within a tier, items are independent.

| Tier | Scope | Items |
| --- | --- | --- |
| 1 | Oracle, evidence and safety (severity H) | D-01, D-02, D-03, D-04, D-05, D-06 |
| 2 | Run lifecycle and concurrency | D-07, D-08, D-09, D-10, D-15, D-16, D-17 |
| 3 | Truth, world, consumer and surface correctness | D-11, D-12, D-13, D-19, D-20, D-21, D-22, D-23, D-24, D-25, D-27, D-41, D-42, D-50, D-51 |
| 4 | Loaders, adapters, schemas, sinks | D-26, D-28, D-30, D-31, D-32, D-33, D-34, D-38, D-40 |
| 5 | Structure and test weak spots | D-46 (mcp protocol edge), D-47, D-48, T-01 (error assertions), run facade test hooks, determinism-gate gaps D-43 |

## Needs a decision (not applied unprompted)

| ID | Why | Recommendation |
| --- | --- | --- |
| D-14 | Making reads pure changes simulated numbers | Pin the current numbers; make reads pure behind a new `sim_version` |
| D-18 | Two delivery paths diverge; intent unclear | Choose the ledger-first path, unify, record as a version change |
| D-35 | Zero equals unset for omitempty floats | Keep; document in the domain contract |
| D-36, D-39 | Changing digests invalidates stored artifacts | Keep the legacy digest, add the RFC 8785 digest beside it |
| P-01 | `wall` seam has no importers | Delete the package; amend DECISIONS D-13 |

## Status

| Item | State | Commit | Regression test |
| --- | --- | --- | --- |
| D-01 | done | `see git log` | `TestRevealRefusesALabelForARunTheDirectorDoesNotKnow` |
