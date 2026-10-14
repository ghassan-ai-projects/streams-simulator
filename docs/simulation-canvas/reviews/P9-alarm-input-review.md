# P9 alarm playback input review

Date: 2026-10-14. Reviewer: round 7 desktop-input review. Scope: R9-24 alarm playback on the frozen real binary and UI, using Chrome. This is a bounded functional review, not full P9 sign-off or a separate accessibility audit.

Frozen manifest: `/private/tmp/streamsim-p9-alarm-review/manifest.json` (`sourceBase` `53aefb8`; binary SHA-256 `96e3450742b948dadb85931a54d6e254f2950000309d3b2da731d042c3c6ec0a`; UI SHA-256 `3f7b7e5758d1665598a0314792dde6d4059b0015790c9dd166afbd7268cfefc1`). Tested the manifest's `w-alarm` one-heater workspace and `dist-final` UI.

**Scoped score: 5/5.** No new U7 functional findings.

## Evidence

- Opened `w-alarm` in the paired Chrome tab. The real report showed `Verified`, `Evaluate · ended`, simulator `0.3.2`, and 180 observations.
- From 00:30:00, clicking the alarm event chip sought to 00:02:20 (slider value 140 seconds), selected HEAT-1, switched to `Inspect · HEAT-1`, and showed the observed temperature as 85.0 °C at 00:02:20. The event announcement said it moved to 00:02:20 and selected HEAT-1.
- Keyboard navigation reached the alarm chip; Enter activated it with the same result. The focused element remained the activated chip. `Jump to alarm: HEAT-1 temp alarm activated` worked from the run end by mouse and by focusing it with Shift+Tab and pressing Enter; focus remained on the jump button. From 00:30:00, `Previous event` also navigated to 00:02:20.
- The HEAT-1 outline button responded to Space selection, and Inspect retained the selected entity and its 85.0 °C observation. With the Playback time slider focused, ArrowRight moved from 00:02:20 to 00:02:21 (140 to 141 seconds).
- Checked dark and light themes at 1280×900, 1440×900, and 1920×1080. The one-heater card, selection outline, alarm chip, playback controls, and Inspect value remained visible at each width; labels and values were readable. The temporary viewport override and theme preference were reset afterward.

## Disposition checks

| Prior item | Round 7 check |
|---|---|
| E7-2, alarm activation batch completeness | The frozen build exposed the alarm chip and Jump to alarm; both navigated to the known activation and showed the selected entity in Inspect. Passed for this workspace. |
| E7-3, exact nanoseconds versus displayed seconds | The known 140-second activation navigated exactly; this review does not claim nanosecond UI precision or test sub-second activation times. |
| Alarm-navigation Chrome regression, Inspect not opened after event reveal | Fixed in this build: event reveal selected HEAT-1, opened Inspect, and showed 85.0 °C at the activation time. |
| E7-1, incomplete verification caution in Inspect | Not exercised: the frozen workspace contains only the verified `w-alarm` run. The prior code review records the incomplete-stamp regression test as the disposition evidence. |
| Prior round 6 desktop findings | No campus re-review or broad visual-optimization pass was performed, per this batch's scope. |

No quality-bar gap was observed for the R9-24 input flow in this fixture. The review used the real frozen binary and UI; it made no code or gate changes. The only repository modification is this review record.
