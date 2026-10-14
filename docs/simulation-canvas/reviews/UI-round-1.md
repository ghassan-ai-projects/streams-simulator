# UI goal, round 1: flows 2 and 3

- **Date:** 2026-10-14.
- **Goal:** [GOAL-UI-PRODUCTION.md](../GOAL-UI-PRODUCTION.md).
- **Frozen build:** `0a00224`, which is flow 2 on top of flow 3 (`b74cb5d`).
- **Reviewers:** three Sonnet subagents, all read-only:
  - design director;
  - desktop input and accessibility;
  - process and test engineer.
- **Evidence:** screenshots of the real binary at 1280, 1440 and 1920 px in light and dark.

## Scores

| Reviewer | Score | High |
|---|---|---|
| Design director | 4 | none |
| Desktop input and accessibility | 4 | none |
| Process and test engineer | 4 | none |

The round meets the bar: every reviewer scored at least 4, with no High finding. The Medium findings that touch correctness or keyboard use were fixed anyway; the dispositions are below.

## Dispositions

### Fixed

| ID | Finding | Fix |
|---|---|---|
| E-1 | A run left from an earlier visit opened when the panel next mounted. | The panel resets a finished or failed run when it mounts, and cancels its own run when it unmounts. |
| E-2 | After a conflict, recovery needed a page reload. | "Reload from disk" re-reads the file. `open` re-checks the file version when the draft is clean, and keeps a dirty draft. |
| E-4 | The domain etag was read just before compose, so the guard never protected anything. | The etag is captured when the plant opens and refreshed after each compose. A new plant whose domain already exists is refused. |
| E-7 | The plant store did not reload when its composition appeared in the listing later. | The store reloads when the paired composition path changes. |
| E-8 | Seeds above 2^53 were rounded. | The seed must be a safe integer. |
| E-9 / A-13 | `null` kept the deleted instance selected. | `undefined` now means "unchanged". |
| E-10 | "Not saved" was shown when the file was saved but compose failed. | The status reads "Saved, not composed: …" with the binary's message. |
| E-13 | An `/equipment/i/names/j` pointer flagged every name in the entry. | It flags only that name. |
| E-14 | `start` had no re-entry guard. | A run in progress makes `start` a no-op. |
| E-16 | The workspace spec's copy had the same title as the base. | The copy has its own title. |
| E-3 | Instances from an `id_template` could be renamed, which corrupted the wires. | They are read-only in the editor. |
| A-1, A-2 | Focus was lost when the panel swapped. | Panel actions move focus to the panel that replaces them: add and issue links go to the instance heading, delete and close go to Build, rename goes to Name. Selecting in the graph keeps focus in the graph. |
| A-3, A-4 | Focus was lost during a run. | On run, focus goes to Cancel; on failure, to "Change the plan"; when the form opens, to Seed; after removing a fault, to Add fault; on close, back to Run plant. Run plant is a real toggle and is locked while a run is in progress. |
| A-5 | Cancel gave no feedback. | "Canceling at the next step…" is shown and the button is disabled. |
| A-6 | The plan was lost after a failure or cancel. | The draft lives in the panel and survives both. |
| A-7 | Revert discarded work in one step. | Revert needs a second click on "Discard changes". Leaving the page with unsaved changes asks first. Delete stays one step, with Revert as the undo. |
| A-8 / D-6 | Pins could not be reached at fit zoom. | The editor opens at 100 %, the pin hit area is larger, and the hint names both the drag and the Link an output form. |
| A-9 | Focus was not managed in the New plant form. | The form focuses Id when it opens. Escape and Cancel return focus to New plant. |
| A-11 | Disabled buttons gave no reason. | Add and Link explain why they are disabled through `aria-describedby`. |
| A-12 / D-7 | Instance issues were not grouped or linked to the field. | They are grouped under "Issues on X", repeated messages are counted, and Name is described by them. |
| A-14 | The fault minute field's accessible name did not contain its visible label. | It is named "Fault N starts at minute", and a new fault row takes focus. |
| D-1 | The editor status was clipped at 1280 px. | The status keeps its width and the breadcrumb shrinks instead. |
| D-2 | "Create and edit" did not read as the primary action. | It is the primary button, and disabled buttons are visibly dimmed. |
| D-4 | The run mode was not visible as a choice. | Explore and Evaluate are a named pair of radios. |
| D-5 | Plan problems showed before the user tried to start. | They are shown when Start is pressed, and focus moves to them. |
| D-9 | The instance options did not name the instance kind. | Options read "CT-1 · cooling tower", and "Choose instance" and "Choose fault" replace "Instance…" and "Fault…". |
| D-10 | The minimum duration was given in seconds, and the seed had no explanation. | The duration is at least one minute, and the seed has a hint. |
| D-11 | Progress showed only the world id, and failures had no title. | Progress names the plant and its world, and failures have a title. |
| Tests | Coverage gaps in the end-to-end tests. | The launch end-to-end test checks the injected fault in the review. The editor end-to-end test checks that the saved domain is byte-identical to `streamsim domain compose` of the saved composition. |

### Postponed (one line each, per the goal)

| ID | Finding |
|---|---|
| E-5, E-6 | The id rule and the file naming (`domains/<id>.domain.json`, `compositions/<id>.composition.json`) should come from the binary. This needs a small binary capability: compose returns its output path and validates the id. |
| E-11 | Issues stay keyed to the last saved document until the next save. |
| E-12 | Opening another composition drops an unsaved draft. The leave-page guard covers reloads, not in-app navigation. |
| E-15 | Renaming onto an existing name is left to the binary's issue. |
| D-3 | The words "plant" and "composition" are kept as the owner uses them. |
| D-8 | Save stays enabled when clean; saving recomposes the domain. |
| D-12, D-13, D-14 | Graph margins, workspace width at 1920 px, and link-row polish. |
| A-10 | The create failure is not tied to the Id field. |
| A-15 | The key help does not yet mention Link an output. |

## Gates after the fixes

The canvas check passes:

- 165 unit and component tests.
- 22 Playwright tests against the real binary, which ran in about 26 s and passed three times in a row.
- The canvas pipeline takes about 60 s.

Checked in Chrome at 1280, 1440 and 1920 px in light and dark.
