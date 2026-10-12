# P6c review record — trips, lockouts, anti-recycle, boundaries, preconditions

- **Reviewed**: the P6c working-tree diff against `5358417` (2026-10-12)
- **Reviewer**: one independent Sonnet subagent, with probes on a scratch copy and a behaviour-pin run.
- **Outcome**: 0 High, 3 Medium, 5 Low. Resolved below, accepted, or carried.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | Med | `ViolatedBoundaries` at a time after the clock integrated ahead and committed, losing a later latch | Fixed: it reads on the look-ahead copy; `TestReadingBoundariesAheadOfTheClockChangesNothingEarlier` |
| 2 | Med | The look-ahead copy was built for every future-dated call | Fixed: only for effectors with an interlock or preconditions |
| 3 | Low-Med | Oracle and audit worlds apply setup calls before faults, so a setup call dated after an onset is checked without the fault | Carried (R6c-01, P12) |
| 4 | Low | Precondition refusals are not cached for idempotency | Accepted and documented (spec section 17, as for interlock refusals) |
| 5 | Low | `tripped_value` equal to the initial value makes a trip indistinguishable | Fixed: refused at `/…/f1/tripped_value`; overlays on trip states are the domain's choice |
| 6 | Low | The dependency graph ignores `when`, `watch`, `reset_on`, `selector` | Accepted and documented: read at step start, so a cycle is a one-step delay |
| 7 | Low | The `requires` schema description named the director-only reason | Fixed: reworded in both copies |
| 8 | Low | Edge and reset detection react to any change; a safe-stop effector with `requires` surfaces as `ErrPlantUnavailable` | Documented (discrete watched states); safe-stop carried (R6c-02, P7a) |
