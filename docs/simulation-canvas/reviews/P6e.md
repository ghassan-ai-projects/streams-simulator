# P6e review record — event clock

- **Reviewed**: the P6e working-tree diff against `83f2b20` (2026-10-12)
- **Reviewer**: one independent Sonnet subagent, with probes on a copy of the tree and a HEAD archive.
- **Outcome**: 2 High, 1 Medium, 7 Low. Resolved below, accepted, or carried.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | High | Every event-driven channel of the entity was checked at pulse instants, so a noisy channel unrelated to the pulse emitted at the exact injection time and duration (truth leak) | Fixed: only channels that observe, or are triggered by, a state the pulse affects are checked; `TestAnEventDrivenChannelRecordsASagAndOtherStreamsDrawAsBefore` asserts an unrelated noisy channel never records at the instants |
| 2 | High | `ClearFault` could extend a pulse that had already ended, rewriting history | Fixed: a clear only shortens; `TestClearingAPulseOnlyShortensItAndRecordsTheNewEnd` |
| 3 | Med | The hold-until-next-sample rule tripped on a hold that ended inside the last step | Fixed: the condition must still hold just before the next sample; `TestAHoldThatEndsBeforeItsDurationDoesNotTrip` (9.8 s against 10 s) |
| 4 | Low | An early clear left the stale end check and none at the clear | Fixed: a clear of a pulse is checked at its time (same test) |
| 5 | Low | A pulse dated in the past scheduled checks before the clock | Fixed: checks are scheduled no earlier than the clock |
| 6 | Low | Event checks advance the channel's availability | Documented in the draw-order table; limited to channels that see the pulse |
| 7 | Low | The draw-order table's conditions were loose; the stochastic envelope stream is shared per declaration | Table tightened; the shared stream carried (R6e-02) |
| 8 | Low | `duration_s` accepted on non-pulse shapes | Fixed: refused at `/…/onset/duration_s` |
| 9 | Low | `pulseInstants` scans every injection | Accepted: injections per run are few |
| 10 | Low | Test gaps | Covered by the tests above |
| — | FYI | A direct future read outside the look-ahead commits integration (pre-existing) | Carried (R6e-03) |
