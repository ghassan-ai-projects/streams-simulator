# Walkthrough and evidence

## Scope

Desktop Chrome, 24 October 2026. Continued the existing audit branch and isolated workspace at `/private/tmp/streamsim-ui-audit-20261024`, using the same binary/UI build from commit `3c6d78c8faa6ce86ac8119ef34d31536a9668c67`. Restarted the isolated loopback bridge on port 8766; the pre-existing server on 8765 was left alone.

The owner refined the task during the review: put more effort into the run page, stop concurrency checks, and prioritize improvements rather than bugs. Subsequent work followed a single-tab run journey. Preliminary multi-tab probes were discontinued and their scratch captures excluded from this package. Their behavior is not used to justify the run-page recommendations.

No implementation source was inspected. No code changes, mobile review, viewport resizing, physical operation, or successful host-backed approvals were performed. The report is a design-improvement review grounded in browser interaction. It makes no numerical correctness, hardware-readiness, or full-gate claim.

## The main exercise

The factory utilities model contains 23 instances and 59 links across the site's modeled systems. It was reused from round 1. The area-only edit in the prior round remained; all three reference runs below used the same current factory data.

### Planned response: w-6

Explore, seed 314, 30 simulated minutes, Watch pacing:

| Time | Planned event | Why it helps review the UI |
|---|---|---|
| 00:00 | Nominal operation | Establish initial context |
| 03:00 | CT-1 basin-temperature dropout | Follow measurement availability and sample age |
| 05:00 | CT-1 fan failure | Follow the process disturbance and affected relationships |
| 08:00 | CH-1 stop | Inspect an intervention and its subsequent observations |
| 12:00 | CH-1 start | Follow restart and phase changes |
| 15:00 | CH-1 set setpoint, `setpoint c = 9` | Reconstruct numeric command intent and response |
| 30:00 | End | Explain outcome and coverage |

The sequence is a simulator exercise for interface design. It is not advice to operate a real cooling plant this way.

Walked through planning, graph-assisted selection, execution progress, asset inspection, completion, event selection, keyboard time stepping, end-of-run evidence, and the report. The UI marked the recording Verified and showed 14,320 observations. Reported simulator command outcomes:

- Stop: applied, 338 ms.
- Start: applied, 334 ms.
- Set setpoint: applied, 285 ms.

These are UI-reported outcomes and latency values. They do not independently prove a physical effect or define the exact meaning of a latency field.

### Disturbance reference: w-7

Used Set up another run from w-6. Kept seed 314, duration 30 minutes, Explore, and the same two disturbances. Deliberately retained no effector calls. Selected as-fast-as-possible execution. The UI marked the recording Verified, with 14,320 observations.

### Nominal reference: w-8

Prepared from w-7, removed both disturbances, retained seed 314 and duration 30 minutes, no commands, Explore, as-fast-as-possible. The UI marked the recording Verified, with 14,482 observations.

## What the references demonstrate for the design

| Reference | Whole-run true mean CH-1 COP shown in UI | Intended role |
|---|---:|---|
| w-8 | 5.09 | Nominal context |
| w-7 | 5.02 | Same disturbances, no planned response |
| w-6 | 4.40 | Disturbances and staged response |

The purpose of these numbers is to illustrate the need for coherent comparison. They are one-seed, whole-run model summaries. They do not rank the response or establish broad effectiveness. The stop interval, subsequent conditions, production evidence, coverage and an explicit objective all belong in that assessment. Aligning event-relative windows is a proposed workflow; the current UI was inspected through individual run reviews.

## Concrete operating observations

- The plan contains meaningful targets and commands, but reading the combined chronology requires merging fault and call sections. A dedicated chronological plan would make this sequence immediately understandable.
- Selecting an asset on the plant can fill a run-plan target. The Add controls then name that asset. This is a useful interaction to extend into watch lists.
- During execution the panel shows run identity, counts, progress and Stop now. A phase band, Now/Next view and watched evidence would give that space more operating value.
- Selecting the stop event moves the cursor to 08:00 and focuses CH-1. Its observed power there is 20.9 kW; stepping to 09:00 shows -0.70 kW with instrument-range information. Preserve the distinction between noisy observed evidence and true model values. The proposed action story should connect the intervention to subsequent observations and its actual assessment criteria.
- Selecting the 15:00 setpoint event focuses CH-1 and shows its applied outcome. The proposal adds numeric argument and rationale context beside that evidence.
- At 30:00, CT-1 basin temperature retains 23.5°C with observation time 03:00. The existing table preserves that timestamp. The proposal makes its 27-minute age at the cursor explicit and visible on the trend.
- Whole-run truth aggregates appear beside cursor-time observations. Explicit interval and source labels would help a reader combine them correctly.
- The report supplies event outcomes and detailed KPI tables. A question-led overview, phase summaries and aligned references would reduce the work needed for a conclusion or handover.
- Round-1 w-1 still opened as a verified historical record after the copied factory's area edit. This supports preserving historical provenance while giving comparisons an explicit input-difference view; it is not a general test of every model-revision case.

## Evidence index

Only the run-page-relevant captures are retained. Numbering has intentional gaps from the earlier discarded exploratory probes. A shared basename has `.txt` and `.jpg` unless noted.

| Capture | Content |
|---|---|
| [01-history-after-plant-edit.txt](evidence/01-history-after-plant-edit.txt) | Older factory review after the prior area edit |
| [08-schedule-bounds.txt](evidence/08-schedule-bounds.txt) | Existing focused guidance for an event outside the run interval |
| [09-run-plan.jpg](evidence/09-run-plan.jpg), [text](evidence/09-run-plan.txt) | Tall plan with two faults and three commands |
| [10-run-live.jpg](evidence/10-run-live.jpg), [text](evidence/10-run-live.txt) | Active execution panel and plant |
| [11-run-completion.jpg](evidence/11-run-completion.jpg), [text](evidence/11-run-completion.txt) | Handoff from run execution into review |
| [12-review-overview.jpg](evidence/12-review-overview.jpg), [text](evidence/12-review-overview.txt) | Event timeline and full report for w-6 |
| [13-sensor-loss-review.jpg](evidence/13-sensor-loss-review.jpg), [text](evidence/13-sensor-loss-review.txt) | Held reading, observation timestamp and selected cursor |
| [14-intervention-selected.jpg](evidence/14-intervention-selected.jpg), [text](evidence/14-intervention-selected.txt) | Event selection focuses CH-1 and its evidence |
| [15-after-stop-observation.txt](evidence/15-after-stop-observation.txt) | Step to the subsequent observation period |
| [16-setpoint-event.txt](evidence/16-setpoint-event.txt) | Selected numeric-intervention event |
| [17-fault-only-reference.txt](evidence/17-fault-only-reference.txt) | Full report for matched disturbance reference w-7 |
| [18-baseline-plan.txt](evidence/18-baseline-plan.txt) | Nominal reference plan after removing disturbances |
| [19-nominal-reference.txt](evidence/19-nominal-reference.txt) | Full report for nominal reference w-8 |
| [20-baseline-chiller-end.jpg](evidence/20-baseline-chiller-end.jpg), [text](evidence/20-baseline-chiller-end.txt) | Nominal chiller observations and KPI context |
| [21-graph-assisted-plan.txt](evidence/21-graph-assisted-plan.txt) | Selecting CT-1 fills its target and names relevant Add actions |

Screenshots preserve the existing desktop viewport; no other screen size was tested. Text captures are rendered-page snapshots, not implementation files. Progress can advance between a text capture and its paired screenshot.

## Proposed workflow acceptance

Using this exact exercise, ask a person to:

1. Name the investigation, arrange its events chronologically, and select watched signals.
2. Confirm environment, mode, complete inputs and intended output before launch.
3. Follow the current phase, upcoming intervention, and evidence quality without losing context.
4. Select the stop action and explain its request, outcome and subsequent evidence.
5. Identify the last valid tower observation and its age at the selected cursor.
6. Compare the same phase/window across nominal, disturbance and response runs, with input differences visible.
7. Record a conclusion and open questions, then export a handover linked to exact evidence.

These are acceptance tasks for the proposed improvements, not completed tests of features that have not been built. Live approval and connected-equipment interaction will require a representative host/device setup for their own validation.

## Delivery checks

The package contains four second-round Markdown documents, 20 recommendations and 22 retained evidence files. Local Markdown targets and whitespace were checked across all eight documents in the combined audit, with no failures. `git diff --check` passed. Git status contains only the audit documentation folder; no production or test changes were made. Full application gates were not run for this documentation-only review.

The isolated audit server was stopped after the walkthrough. Runtime recordings remain in the temporary workspace; the round-1 w-1 recording's SHA-256 was unchanged by this round. No commit or push was performed.
