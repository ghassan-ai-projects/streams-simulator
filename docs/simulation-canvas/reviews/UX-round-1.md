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
| UXE-1 | Fixed `f2afe1c`: Stop ends the run early and keeps it, as the binary does. |
| UXE-2 | Fixed `5c61d95`: Play before the run has loaded no longer wedges playback. |
| UXE-4 | Fixed `0baaf4d`: a run panel never jumps to a run it did not watch. |
| UXD-2 | Fixed `d0f9d88`, `d8a8549`: graphs open fitted at the top and refit on resize and panel toggle; the editor opens on the whole plant and zooms to a selection. |
| UXD-1, UXD-8, UXD-9, UXD-22, UXA-7, G-13 | Fixed `4a3d258`: one top bar with the breadcrumb; labelled theme, panel and Keyboard buttons; grouped key list with a scrim. |
| UXA-5, UXA-6 | Fixed `1c03d29`. |
| UXA-8 | Fixed `54ff2fe`: About this run is a real popover. |
| UXA-9 | Fixed `b46c0d0`. The scrubber's value during playback is not throttled; not verified with a screen reader. |
| UXA-10, UXD-25 | Fixed `063a011`. |
| UXD-4 / V-8 | Fixed in the canvas `593f431`: readings outside the instrument range are marked. The simulator defect (noise not bounded by the range) is D-41 in `governance/DEFERRED.md`. |
| UXD-5, UXD-18 | Fixed `2fe6ad4`, `b6d4803`, `5941949`: marks on the track, run time everywhere, legend, thin-space counts, the opening explained. |
| UXD-6, P-11 | Fixed `21ac9b8`. |
| UXD-7 | Fixed `16c51da`: instance counts from the bridge. Run rows still carry no fault count. |
| UXD-10, UXD-11, UXD-12 | Fixed `9998053`. |
| UXD-13 | Fixed `d41c032`, `d8a8549`, `e4d5de0`. |
| UXD-14 | Fixed `ba90274`, `e9c8974`, `21ac9b8`. |
| UXD-15 | Fixed in the binary `04f1869`: patterns can say in words what they expect. |
| UXD-16, UXD-17 | Fixed `6b797ef`. |
| UXD-19 | Fixed `d8a8549`, `0df96ae`. |
| UXD-20 | Fixed `8a81f75`. Halls are still laid out by rank. |
| UXD-21 | Fixed `55b7692`: scroll cues. |
| UXD-23 | Fixed `5510f2b`, `21ac9b8`. |
| UXD-24 | Fixed `b6d4803`. |
| UXE-6 | Fixed `e2e9c1f`. |
| UXE-8 | Fixed `c25966c`. |
| UXE-9, UXE-10 | Fixed `8d5d8a6`. `.panels` is used through `composes`, so it stays. `recorded_at` remains the journal's file time. |
| UXE-3, UXE-7 | Open, binary first: needs a CLI verb for a new composition (empty or copied, id checked). The CLI file caps need an owner-approved gate change. |
