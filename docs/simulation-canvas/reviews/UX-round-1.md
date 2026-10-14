# UX goal, round 1

- **Goal:** [GOAL-UI-UX.md](../GOAL-UI-UX.md).
- **Bar:** [UI-EXPECTATIONS.md](../UI-EXPECTATIONS.md).
- **Frozen build:** `b9349a4`. It was captured as 49 screenshots at 1280, 1440 and 1920 px in light and dark, against the real binary.
- **Reviewers:** three Sonnet agents, all read-only:
  - design director;
  - desktop input and accessibility;
  - process and test engineer.

## Scores

| Reviewer | Score | High findings |
|---|---|---|
| Design director | 3 | UXD-1 editor bar loses the plant name; UXD-2 nothing opens fitted; UXD-3 Evaluate truth unexplained |
| Desktop input and accessibility | 3 | UXA-3 a hidden details panel hides the run panel and its Cancel |
| Process and test engineer | 3 | UXE-1 "Cancel and discard … keeps nothing" is false: destroying a world writes the partial run |

## Status corrections accepted

Each line gives the expectation, the status it moves to, and the gap behind it.

- **W-1** moves to ◐: rows have no instance count.
- **W-3** moves to ◐: rows have no fault count or verdict.
- **W-6** moves to ◐: there is no scenario line.
- **W-11** moves to ◐: the run chips carry little meaning.
- **P-2** moves to ◐: kinds are clipped on short nodes.
- **E-11** moves to ◐: disabled controls look live in dark.
- **V-1** moves to ◐: the verdict chip does not explain itself, and the details popover has no Escape.
- **V-2** moves to ◐: Space conflicts.
- **RUN-3** moves to ◐: the hidden panel hides Cancel, and the cancel wording is false.

## Dispositions

Fixes are committed one per finding. The commit is noted where it already exists.

| ID | Disposition |
|---|---|
| UXD-3 | Fixed `2785563`: the report says when the truth was revealed; "True KPIs"; the Evaluate wording names the reveal at the run's end. |
| UXA-3 / UXE-5 | Fixed `3f1e787`: anything that opens the run panel shows the details panel. |
| UXA-1, UXA-2, UXA-4, UXA-11 | Fixed `eb146bc`: one rule for which keys a focused control owns. |
| UXD-13 (part) | Fixed `d41c032`: library search, description and a suggested name. |
| UXD-14 (part) | Fixed `ba90274`: links in words. |
| UXE-1 | Next: Stop ends the run early and keeps it as a shorter run, which is what the binary does, then opens its review. |
| UXE-2 | Next: Play before the journal has loaded must not wedge playback. |
| UXE-4 | Next: starting a run must not jump to an old, unreviewed run. |
| UXD-2 | Next: fit on open, on panel toggle and on resize, on all three graph screens, with no clipped frame titles. |
| UXD-1, UXD-8, UXD-22, UXA-7, G-13 | Next: one shell, with the breadcrumb in one place, never truncated, and labelled theme, panel and keyboard controls. |
| UXA-5, UXA-6, UXA-8, UXA-9 | Next: quiet live regions, focus after dismiss, a real popover, and announced play state. |
| UXD-5 | Next: scrubber marks on the track, centred, with a legend. |
| UXD-10, UXD-11, UXD-12 | Next: run panel finish. |
| UXE-3, UXE-7 | Binary first: which keys an empty plant keeps, and id validation, belong to the binary (`component scaffold` or similar). Interim: none. |
| UXD-4 / V-8, UXD-15 | Binary first: plausibility flags and issue wording. |
| UXE-8 | Bridge clean-up: build `runState` on `SummarizeRun`. |
| Remaining Mediums and Lows | Worked through after the above, in the order of the expectations document. |
