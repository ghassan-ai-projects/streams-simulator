# P9 alarm playback design review

Date: 2026-10-14. Reviewer: design director (Round 7). Score: **5/5 for this bounded batch**.

This is a functional review of the alarm-playback batch only. It is not a full P9 sign-off. Owner direction for this round was to finish functionality first and defer visual optimization.

## Frozen build and evidence

- Manifest: `/private/tmp/streamsim-p9-alarm-review/manifest.json`
- Frozen binary SHA-256: `96e3450742b948dadb85931a54d6e254f2950000309d3b2da731d042c3c6ec0a`
- Frozen UI SHA-256: `3f7b7e5758d1665598a0314792dde6d4059b0015790c9dd166afbd7268cfefc1`
- Run: `w-alarm`, simulator 0.3.2, one heater, 180 observations.
- Paired the frozen build in Chrome through my own bridge, PID 45579. Tested 1280×900, 1440×900, and 1920×1080 in light and dark themes. Stopped only PID 45579 after review.

On the completed run, the Events lane shows `00:02:20 HEAT-1 temp alarm activated`; the alarm-system report shows one activation. Clicking the chip seeks to 00:02:20, selects HEAT-1, opens Inspect, and shows the observation `85.0 °C`. Moving playback to the 00:30:00 run end and pressing Jump to alarm wraps to the activation at 00:02:20. The alarm mark remains a point event; the UI does not infer a cleared instant or active duration.

At all six checked desktop theme/width combinations, the alarm chip and jump control remained present and usable. The initial report-loading state showed `Jump to alarm: none` and `Next event: none` as disabled controls. The current frozen workspace contains only the alarm run, so I did not observe a completed zero-alarm run. The component test covers the no-alarm case and verifies clicking its disabled jump control does not seek.

## Code-review dispositions

| Finding | Disposition |
|---|---|
| E7-1 — KPI-only Inspect hid the unverified-truth caution | Fixed in `RunPanel.tsx` by passing the ready verification state to `ReportPanel`. The real-binary incomplete-run E2E now selects an entity and asserts the caution beside KPI truth. |
| E7-2 — activation instants were emitted before canvas navigation was complete | Completed in this batch. The frozen Chrome flow verifies the report instant, chip, jump, selection, Inspect tab, and reading. The world test independently scans alarm transitions and verifies returned-slice ownership. |
| E7-3 — JavaScript numbers cannot preserve arbitrary int64 nanosecond timestamps | Deferred contract requirement for the next playback-query batch. This frozen run's 140-second activation navigates correctly; this test does not establish nanosecond precision for arbitrary large absolute timestamps. |
| Alarm-navigation Chrome regression — source selection/time changed but Report stayed open | Fixed: `ReviewStore.reveal` opens Inspect, with an app-state regression and the frozen Chrome observation above. |

## D7 findings and disposition

No new D7 functional findings. Alarm-chip navigation, alarm jump wrap, observed source reading, and empty/loading control states behaved as expected in the frozen build. The one-node graph's placement and available whitespace are visual optimization topics deferred by owner direction; I am not making a visual-layout finding in this round.

This verdict covers only the alarm-playback batch. It does not approve the larger P9 functional blockers or claim full P9 completion.
