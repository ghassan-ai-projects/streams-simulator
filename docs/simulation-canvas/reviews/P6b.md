# P6b review record — `schedule`, `regime`, `peak`

- **Reviewed**: the P6b working-tree diff against `8d09ee3` (2026-10-12)
- **Reviewer**: one independent Sonnet subagent, with a scratch read-order experiment (three start offsets, thirty random read
  orders per time) and a utility-plant run through the binary.
- **Outcome**: 0 High, 2 Medium, 6 Low. Resolved below, accepted, or carried.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | Med | The peak oracle ran only at a clock-aligned start, so a regression to start-relative blocks would pass | Fixed: `TestDemandBlocksFollowTheClockWhateverTheStart` starts 450 s into a block and asserts the hard-coded 07:00–07:15 clock-block mean (a start-relative block gives 88 instead of 106.7) |
| 2 | Med | No read-order test for the new forms; a regime selected by another state was untested | Fixed: `TestScheduleRegimeAndPeakDoNotDependOnReadOrder` (schedule, peak, regime selected by another integrated state, off-grid start, rotating read order against an in-order world) |
| 3 | Low | Block and period lengths truncated to nanoseconds | Fixed: rounded |
| 4 | Low | A world start or birth inside a block gives a partial first block | Accepted and documented (spec section 16) |
| 5 | Low | A read further in the past than the longest step returns the current value (P6a history horizon) | Carried (R6b-02) |
| 6 | Low | `stepRegime` builds its lag per step; a top-level `gain` on a regime is ignored | Accepted: allocation is small next to RK4; the schema documents `gain` per regime |
| 7 | Low | `dynamicPeak` message for an unreachable zero `dt_ms` | No change: the schema requires `dt_ms` for F1 |
| 8 | Low | `transfer-plant.domain.json` diff is formatting only; `transfer_test.go` replaced by `vocabulary_runs_test.go` | Intended: `domain format` output and the dupl lint |
