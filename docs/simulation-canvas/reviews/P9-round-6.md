# P9 round 6 — frozen build, Chrome review

Date: 2026-10-14. Branch: `codex/canvas-p9-playback`, created from latest `origin/main` at
`6721cdf8bc7930e13af40ef4faa75244d69127d6`. PR #32 has merged. No new PR or push is authorized.

## Build and workspace

Frozen binary SHA-256: `30ebf5ee370bda45211717d2ee6b73a9d836d4f0ee323c42ebfe3358fab28f9b`.
Frozen JavaScript SHA-256: `308220a57fad11a0f5cbf6c221666392e8c171e8500a6a1a8b0114df925d3f53`.
The binary and `canvas/dist` were copied before recording; all runs use that binary.
The synthetic scratch workspace follows the handoff recipe: factory Evaluate and Explore, seed 11, eight hours;
edited factory trace, seed 5, two hours; campus with 26 halls, seed 3, one hour and three injected faults.
The factory description was edited after recording. An additional 120-second run has `incomplete: true` deliberately
flipped in `run.json`, contradicting its complete journals. No customer data or configuration is involved.
Local manifest: `/private/tmp/streamsim-p9-r6-wu1LiP/manifest.json`; scratch paths are ephemeral, not release evidence.

Each reviewer used its own frozen-binary bridge and stopped only its own PID: design 25052, desktop input 25195,
engineering 25330. All browser work used installed Chrome. The owner replaced the separate accessibility lens with
desktop keyboard/input usability. Existing automated accessibility gates remain unchanged.

## Independent findings

| ID | Lens | Severity | Evidence and disposition |
|---|---|---|---|
| E6-1 | Engineering | High | Incomplete stamp: badge Incomplete, replay explicitly rejects `incomplete_stamp`, but KPI truth lacks the caution because the component checks only the Not verified badge. Fixed after review: any rejected verify result warns above truth in the full report and hall KPI view. Two component regressions fail before the change and pass after; real-binary Chrome test asserts the caution and absence of Reproduces. |
| E6-2 / R9-14 | Engineering | High | Campus: 725,253 observations, report loading at 55 s and verify at 77 s; a theme interaction timed out while the browser was busy. Open; report/verify retention and main-thread playback need bounded work. |
| E6-3 / R9-22 | Engineering | High | Edited factory playback shows impossible negative air flow: M-1 −2.8 and M-3 −3.5 m³/h. Open; fix measurement semantics in the simulator, not by hiding values in the canvas. |
| E6-4 / R9-24 | Engineering | Medium | Evaluate report records one alarm activation, but the lane contains only fault chips and no alarm jump. Open; simulator supplies activation truth first. |
| E6-5 / R9-13 | Engineering | Medium | Explore Inspect shows observed values only. Open; simulator supplies bounded, truth-gated true state before canvas comparison. |
| U6-1 | Desktop input | High | Reviewer reports clipped right-side hall nodes at 1280×900 and 1440×900 after Fit. Design reviewer instead observes full factory fit at 1280. Open, needs reproduction on the same hall/width and a geometric Chrome regression before attributing a root cause. |
| U6-2 | Desktop input | Medium | Campus heading to graph requires 65 Tab stops through the outline and controls. Open; a direct graph-focus route is functional work. |
| D6-01 | Design | Medium | At 59% compact zoom, MB-COOL and MB-PROC kind labels truncate. Deferred visual detail after functionality per owner direction. |
| D5-12 | Design | Medium | Hall tiles still show member count without a summary figure. Deferred visual detail after functionality. |
| E6-boundary | Engineering | Candidate Medium | Report/replay responses cross the adapter as unknown and are permissively parsed in the app/model. Needs contract-boundary review; do not claim schema validation at the adapter until proved. |

Scores on the frozen build: design **4/5**, desktop input **3/5**, engineering **2/5**. Round 6 is recorded, **not approved**.
No reviewer re-reviewed the truth-caution fix yet. P9 remains In progress.

## Previous findings checked

- E5-1: the incomplete stamp never says Reproduces. All three reviewers and the parent confirmed this in Chrome.
- E5-2: new real-binary Chrome regression deletes `ledger_digest` and proves both file and replay checks fail.
- E5-4: source/Go evidence says open Evaluate truth remains sealed; this frozen workspace contains ended runs,
  so the round does not claim a live Evaluate canary test.
- Selection follow, event-chip navigation, hall entry/back, readable wires and edited-trace feedback passed.

## Validation after the first fix

`make quick` passed on the merged base. `npm run check` with the real binary passed after the truth-caution fix:
126 unit/component tests and 17 installed-Chrome Playwright tests, including incomplete stamp and missing digest.
No architecture gate, script, facade, production Go or dependency changed in this first fix.
The full phase gate has not yet been rerun for this round; no push has occurred.

## Owner direction and next work

Complete functionality before visual optimization: P9 simulator blockers and real-binary Chrome acceptance first,
then P10 live control and P11 builder; visual details after working workflows. Misleading evidence and controls
that block desktop operation remain immediate functional fixes. Findings and dispositions are always persisted here.
Sonnet delegation could not authenticate and changed nothing; the owner subsequently requested direct implementation.
