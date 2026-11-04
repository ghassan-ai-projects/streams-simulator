# Canvas modularity audit and refactoring plan

- **Date:** 2026-11-04 · **Branch:** `canvas-modularity-audit`, from `main` at `c3906add`
- **Scope:** `canvas/` only. No Go, contract, `test/architecture` or root `scripts/` change.
- **Hard rule:** no behaviour change. DOM, accessible names, classes visible to e2e, CSS cascade, bridge calls, timing of effects and user-visible wording all stay as they are.
- **Status:** revised after two reviews (§6). R1, R4, R5 and R7 need no owner decision. R2 and R3 wait for the owner (§3).

## 1. Audit result

The canvas is in better shape than most codebases of its size, so this plan is deliberately narrower than a rewrite. Baseline on this branch: format, types, lint, layers, register, duplication, dead code, tests with coverage, build and emitted CSS all pass in about 86 s; 2,902 tests pass, 5 are skipped; coverage is 98.4 % statements, 93.8 % branches; e2e is 170 passed, 1 failed (`editor-parameters.spec.ts:13`, which passes alone).

**Already sound, not touched:** the four layers and their gates; facades and `contract.ts` (types may only leave a module through `contract.ts`, the facade rule requires it); the store foundation (`Observable`, `DraftStore`, `createOperationIdentity`, used by all 20 stores); per-consumer port files; a composition root with almost no logic; few and narrow casts; no `eslint-disable` in production code.

**What is actually weak:**

| #   | Finding                                                                            | Evidence                                                                                                                                                                           |
| --- | ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| F1  | Large modules are flat. Nothing in `internal/` has a subfolder, whatever the size. | Production TS files: `ui/review` 77, `model/review` 42, `ui/devices` 38, `ui/launch` 36, `adapters/bridge` 33                                                                      |
| F2  | `model/review` is several domains behind one facade.                               | 234-line `index.ts` plus 83-line `contract.ts`, about 290 names; clusters measured in §2                                                                                           |
| F3  | Three features sit inside modules that do not name them and have clean edges.      | Capture in `ui/workspace` has **zero** imports to or from the rest; compare in `ui/review` is imported only by `ReviewActions`; `model/review` twin and compare are one-way leaves |
| F4  | A few true duplicates, each under the 75-token clone gate.                         | §4 R1, each read and confirmed                                                                                                                                                     |
| F5  | Small pieces of logic in the UI.                                                   | §4 R5, each read and confirmed                                                                                                                                                     |
| F6  | Confusing file names.                                                              | `policyWords.test.ts` has no `policyWords.ts`; `twinStore.ts` in both `app/review` and `app/devices`, with different jobs                                                          |

**Audit claims checked and dropped:** `Row`, `Status`, `Legend`, `Chip`, `Section`, `Choices`, `Frame`, `Field`, `TwinForm` share a name only (markup differs). `refusalWords`, `originWords`, `callWords`, `faultWords`, `rangeWords`, `kindOf`, `commandOf`, `strings`, `scopeOf`, `isDirty`, `canSave`, `saveStatusOf`, `plausibilityRow` differ in meaning or wording; merging them would change text. `model/review/index.ts` re-exporting `recordedPolicyWords` from `model/launch` stays: removing it would add a `ui/review` → `model/launch` allow-list edge.

## 2. Measured cuts

Cross-cluster imports were counted over production files, excluding tests.

`model/review`: `report` ← `twin`, `compare`, `interrupts`; `interrupts` ← `playback`. Trends and live import each other (`trendReadout` → `observedRows` → `trend`), and both sit on `playback`, so these three stay one folder. `report` imports none of the others, so `twin` (3 edges) and `compare` (2 edges) are one-way leaves and can become modules, needing no new facade names.

`ui/review`: `compare` has only incoming edges (2, both from `ReviewActions`). `twin` cannot be a module: `TwinParts` and others use `ReportParts` while `RunPanel` uses `TwinPanel`, so a module would create a cycle.

`ui/workspace`: capture has no edge to the rest in either direction.

`app/devices`: `capture` ↔ `draft` through `api.ts`, `deviceFailures` and `deviceDesk` (the device list desk creates the capture session), and `twin` ↔ `check`. Neither a module nor folders are clean, so it stays flat, as does `model/devices`.

## 3. Rules for every round

**Proof of no behaviour change.** Tested on trial moves before writing this plan: moving TSX, CSS and tests into subfolders leaves the production bundle **byte-identical** once CSS-module hashes are normalised. The oracle is:

1. `rm -rf dist && vite build`, failing if the build fails. Removing `dist` matters: a first trial reported "identical" only because the build had failed and the old output was still there.
2. In `dist/assets/index-*.{js,css}` replace each CSS-module hash with its order of first appearance in the CSS, so `_name_hash_line` becomes `_name_s<n>_line`, and blank the asset hash. Numbering rather than deleting the hash keeps sheet identity: a class wired to the wrong sheet still shows up (`Provenance.module.css` exists twice).
3. Compare with the previous round's output.

Negative controls, already run: one changed string fails the JS comparison, one changed declaration fails the CSS comparison, and a clean rebuild is identical.

| Round type                                   | Required result                                                                                                                                                                                            |
| -------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Pure move (R4, R2, R3)                       | JS and CSS byte-identical. In a module split, sorted JS statements may match instead if the split reorders modules, but the CSS must stay byte-identical: fix the facade or import order, never the oracle |
| Duplicate removal, logic extraction (R1, R5) | CSS byte-identical. JS differs by design, and the diff is reviewed against the stated change                                                                                                               |

Further evidence:

- **Tests stay fixed.** In production rounds, `*.test.*` files may differ only in relative import specifiers: each changed test is compared with its previous version with those specifiers masked and whitespace collapsed. New tests are added only for code that moved to a new owner. The test count is compared every round.
- **Full e2e** against the real binary after R4, at the end, and after R3 if it runs. A failure is re-run alone before it counts.
- **Gates each round:** `npm run check:static`, `npm run lint`, `npm run test`. A round that changes modules also updates `architecture.json` and regenerates callers and `ARCHITECTURE.md` in the same commit, or `check:static` fails. `make architecture-baseline` before R1, `make architecture` and `make phase-gate` at the end.

**Gates.** No cap, floor, `testSupport` entry or `.dependency-cruiser.cjs` change. R1, R4, R5 and R7 add no allow-list entry. `git diff main -- canvas/scripts canvas/architecture-fixtures canvas/.dependency-cruiser.cjs canvas/eslint.config.js canvas/vitest.config.ts` must be empty at the end.

**Owner decision for R2 and R3.** AGENTS.md rule 3 lists "add an allowed import" as loosening, and exempts registering a new package with its prescribed imports. Carving out a module also adds entries to existing consumers, which the exemption does not clearly cover and the ratchet does not check. No code becomes newly reachable: each consumer already reaches the same code through the source module. If approved, the commits carry `Owner-approved-gate-change:` and these entries:

| Round | Existing module | Gains                               |
| ----- | --------------- | ----------------------------------- |
| R2    | `ui/shell`      | `ui/capture`                        |
| R3    | `app/ports`     | `model/compare`, `model/twinReport` |
| R3    | `app/review`    | `model/compare`, `model/twinReport` |
| R3    | `adapters`      | `model/compare`, `model/twinReport` |
| R3    | `ui/review`     | `ui/compare`, `model/twinReport`    |

The final list is confirmed from the code in each round. Invariant: a new module's allow-list is a subset of its source module's plus the carved-out model module, a consumer gains an edge only to a module carved out of one it already imports, and the source module drops every edge it stops using.

**Order.** R1, R4, R5, R7, then R2 and R3 if approved. Everything before R2 stands on its own if they are declined, and R2 and R3 then mostly move folders R4 already grouped. Sheets named in the register stay at their `internal/` root, because the ratchet accepts cap and allowance moves only between modules (`ratchetMoves.ts`): `ReviewScreen.module.css`, `LaunchPanel.module.css`, and the untouched `vendor.module.css` and `shared.module.css`.

**Workflow.** One writer at a time in this checkout. Each change gets its own commit, reviewed before committing. Move rounds run through a script that takes a file-to-folder table and rewrites relative imports, so the diff is mechanical. Imports keep their order within each file, because the order drives CSS emission; new imports go after existing ones. Test changes never share a commit with production changes.

**Stop rule.** If a cut needs a new cycle, a cast, more than five new facade names, or an oracle difference that cannot be explained, it is dropped and written down as a known limit rather than forced.

## 4. Rounds

### R1. True duplicates (one commit each)

| Change                                                                                                                                                                                                                                       | Evidence                                                                                                                                                                                                                    |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `finiteOrNull` in `adapters/bridge/internal/decodeKit.ts` replaces `finite`, `bound`, `instant`, `numberOrNull`, `finiteNumber`. The type guard in `twinReportDecoder.ts` stays, it has another contract                                     | `jobs.ts:27`, `jobPlan.ts:10`, `compareResults.ts:15`, `compareItems.ts:14`, `parameterDecoder.ts:4`, `twinDecoders.ts:5`, `assumptionsDecoder.ts:14`, `decoders.ts:240`; inline copies only if the expression is identical |
| `unescaped` (JSON-pointer unescape) moves to `model/shared/json.ts`                                                                                                                                                                          | `model/editor/internal/issueTargets.ts:78`, `model/devices/internal/twinIssues.ts:27`                                                                                                                                       |
| `holds(operation, path)` becomes a plain `app/shared` function bounded on `phase` and `path`; dropped if that needs a cast                                                                                                                   | Verbatim in `editorStore.ts:75`, `app/devices` `twinStore.ts:215`, `deviceStore.ts:127`                                                                                                                                     |
| `SaveStatus` and `status()` move to `app/shared`; `saveStatusOf` stays in each module, it differs                                                                                                                                            | `app/devices` and `app/editor` `saveStatus.ts`                                                                                                                                                                              |
| The save-conflict test becomes a selector of `app/editor` and `app/devices`                                                                                                                                                                  | Same line in `EditorBar.tsx:76`, `DeviceBar.tsx:38`, `TwinBar.tsx:30`                                                                                                                                                       |
| `StatusLine` inside `ui/devices` for the shared `tones` map and `role="status"` span, with an optional `id` (`TwinBar` has one, `DeviceBar` does not). `DeviceBar` keeps its `!asking` guard and readiness span; `EditorBar` is out of scope | `DeviceBar.tsx:30,118`, `TwinBar.tsx:22,107`                                                                                                                                                                                |
| `Pick<DevicesGateway, …>` on device-store constructors, as `app/review` already does (types only)                                                                                                                                            | `app/ports/devices.ts`, 11 methods, no store uses them all                                                                                                                                                                  |

### R4. Feature folders inside `internal/` (one commit per module, plain moves)

Applies to `ui/review`, `model/review`, `ui/launch`, `ui/devices` and `adapters/bridge`; the file table is in §7. Criterion: the folder graph has no cycle over TypeScript imports, measured with this table. Files stay at the `internal/` root when they are the module's entry or a sheet shared across folders.

Then a rename commit: `app/review` `twinStore.ts` → `twinReportStore.ts`, and `policyWords.test.ts` to match what it tests.

### R5. Pure logic out of the UI (one commit each, unit test in the owning layer)

| From                                     | To                                                                  |
| ---------------------------------------- | ------------------------------------------------------------------- |
| `ReviewCentre.tsx:52-63` run span        | `app/review` (it reads the live `review.index`)                     |
| `LiveRoutes.tsx` `listedRunning`, `live` | pure function in `app/shell` over states the component already read |
| `PlantRoute.tsx` `useFollowing` decision | pure function in `app/shell` or `app/launch`, same rule             |
| `PlantScreen.tsx:108-110` hall check     | pure function on the plant; navigation stays in the UI              |

`useStore` calls stay in the components: a new selector subscription would change render counts. Excluded: `Scrubber` geometry (view arithmetic, and it would grow the facade this plan shrinks), and because timing could shift, the `TrendsToggle` restore effect, `PendingApprovals` state from props, `useReviewOpen`.

### R7. Documentation

`canvas/README.md` gains the feature-folder convention; add `evidence.md` here with gate results and oracle output per round.

### R2. Module `ui/capture`, out of `ui/workspace` (needs the owner)

- **Kind/layer:** ui, components. **Responsibility:** the capture-from-a-board screen. **Caller:** `ui/shell`.
- **Files:** `Capture*.tsx`, `useCardFocus.ts`, `Capture.module.css`, `CaptureCard.module.css` and their tests.
- **May import:** `app/devices`, `app/shell`, `app/workspace`, `model/devices`, `model/workspace`, `ui/primitives`.
- `ui/workspace` keeps its allow-list minus anything it stops using, and its register text drops capture.
- `App.tsx:6` becomes two adjacent imports, capture first, matching the facade's current load order.

### R3. Modules out of `review` (needs the owner)

**`model/compare`.** Kind: model. Responsibility: run comparison values, plan lanes, tables and choices. Callers: `app/ports`, `app/review`, `adapters`, `ui/compare`, `test`. Files: `comparison`, `comparePlan`, `compareTables`, `compareChoices`. May import: `model/review`, `model/shared`.

**`model/twinReport`.** Kind: model. Responsibility: the twin-versus-board report values, chart and words, named to keep it apart from twin authoring in `model/devices`. Callers: `app/ports`, `app/review`, `adapters`, `ui/review`, `test`. Files: `twin`, `twinWords`, `twinSelect`, `twinChart`. May import: `model/review`, `model/shared`.

**`ui/compare`.** Kind: ui. Responsibility: compare two runs, and diff a run's model against the plant. Caller: `ui/review` (`ReviewActions`). Files: the `compare` folder of §7. `HandoverNote` stays in `ui/review`: it uses neither. May import: `app/review`, `app/workspace`, `model/compare`, `model/review`, `model/workspace`, `ui/primitives`.

Together these take about 67 of the 290 names off the `model/review` facade. The `ui/compare` facade loads `CompareDialog` first and `ReviewActions` imports it before `HandoverNote`, which keeps today's CSS order.

## 5. Not doing, and known limits

| Item                                                                                         | Why                                                                                                                                                    |
| -------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `Dialog` primitive                                                                           | `aria-modal` and focus handling differ between the four dialogs, so unifying changes accessibility or takes so many options it adds nothing            |
| `useDismiss` hook, `ConfirmAction`, `Button`                                                 | Escape handlers differ in `stopPropagation` and focus return; `ConfirmAction` has two uses; `Button` adds nothing over the shared classes              |
| Splitting `LaunchPanel`, `Devices`, `ReportParts` stylesheets; promoting repeated CSS blocks | Changes emitted order, and the cascade across module sheets is not pinned, so the oracle could not prove it safe. A later CSS round with its own proof |
| `settle` async helper, `PathState` type, `controlStore` generation counter                   | Little gain; the counter may be deliberate                                                                                                             |
| `app/capture`, `ui/twin` as modules, folders in `app/devices` and `model/devices`            | Cycles or tangled clusters (§2)                                                                                                                        |
| `src/test` subfolders or co-located support                                                  | `testSupport` cannot grow; 673 imports moved for little gain                                                                                           |
| Test clones (17 at 60 tokens), e2e setup repetition                                          | Below the production clone gate, and touching tests during a refactor weakens the safety net. A separate test-only change                              |
| Other modules under 30 production files                                                      | No clear clusters; stay flat                                                                                                                           |
| Folder boundaries inside `internal/` are not gate-enforced                                   | A convention only. Modules remain the enforced boundary                                                                                                |
| The ratchet does not guard allow-lists of modules newer than its baseline                    | Real gap, pre-existing. Out of scope; it is a gate change and the owner's call                                                                         |

## 6. Review dispositions

Two independent reviewers, both **approve with changes**, then both **approve** on the revision. Their findings are applied above unless marked otherwise.

| Lens         | Finding                                                                  | Disposition                                                        |
| ------------ | ------------------------------------------------------------------------ | ------------------------------------------------------------------ |
| Behaviour    | Same-module moves of register-listed sheets trip the ratchet             | Confirmed in `ratchetMoves.ts`; those sheets stay put              |
| Behaviour    | `check:static` fails unless the register is regenerated in module rounds | Applied                                                            |
| Behaviour    | Blanking hashes hides a wrong-sheet import                               | Applied: hashes numbered                                           |
| Behaviour    | R2, R3 CSS order needs explicit import order; no sorted fallback for CSS | Applied                                                            |
| Behaviour    | `holds` generic, `StatusLine` differences                                | Applied                                                            |
| Behaviour    | R5 must keep `useStore` in components; `review.index` is a live call     | Applied                                                            |
| Behaviour    | Define the test check; e2e after R4; test count each round               | Applied                                                            |
| Behaviour    | Run R4 before R2 and R3                                                  | Applied, for a different reason: R2 and R3 need the owner          |
| Architecture | Rule 3 is overstated: consumers gain allow-list entries                  | Applied: owner decision, entries listed                            |
| Architecture | `app/devices`, `model/devices` folders are arbitrary                     | Applied: stay flat                                                 |
| Architecture | `ui/devices` folders are only suspected to be clean                      | Rejected: measured acyclic, and both steps depend only on `common` |
| Architecture | `model/review` table incomplete, trends and live cycle                   | Applied                                                            |
| Architecture | `words.ts`, `twin.ts` renames are inexact                                | Applied: dropped                                                   |
| Architecture | `Scrubber` extraction has poor value                                     | Applied: dropped                                                   |
| Architecture | Name `ui/compare` accurately; drop `model/compare` hedge                 | Applied                                                            |
| Architecture | `main.tsx` factory, test-support round have poor value                   | Applied: dropped                                                   |

## 7. R4 file table

Production files per folder; tests follow their source. `.css` means `.module.css`. (root) stays at the `internal/` root.

**`ui/review`**

| Folder       | Files                                                                                                                                                                                                                                                                                                                                  |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `compare`    | Compare.css, CompareDialog, CompareDiff, CompareResults, ComparisonView, ModelDiff.css, ModelDiffDialog, ModelDiffPanel, ModelDiffView, ModelLink, useModal                                                                                                                                                                            |
| `evidence`   | CommandEvidence, PolicyRefusal, SessionEvidence, SessionEvidence.css                                                                                                                                                                                                                                                                   |
| `interrupts` | ActiveInterrupts, InterruptButtons, InterruptLine, InterruptOutcome, Interrupts, Interrupts.css, LiveControls, LiveControls.css, ModeBadge, PendingApprovals, useShownLive                                                                                                                                                             |
| `playback`   | AlarmJump, PlayControls, PlayControls.css, Scrubber, Scrubber.css, ScrubberLane, ScrubberLane.css, useEventKeys, usePlaybackKeys                                                                                                                                                                                                       |
| `report`     | ActionStory, ActionStory.css, KpiTable, OpenQuestions, ReportEvents, ReportIntegrity, ReportPanel, ReportPanel.css, ReportParts, ReportParts.css, ReportTable.css, ThisRun, ThisRun.css                                                                                                                                                |
| `screen`     | ConnectedPlant, GraphBar, GraphBar.css, HandoverNote, HandoverNote.css, MissingRun, MissingRun.css, Notices, Notices.css, ReviewActions, ReviewCentre, ReviewGraph, ReviewSide, RunHeader, RunHeader.css, RunInfo, RunInfo.css, RunPanel, RunPanel.css, RunStatus, RunStatus.css, useLive, useReviewOpen, useRunPlant, useShownOrigins |
| `trends`     | MiniTrend, Observed, Observed.css, ObservedReadings, ObservedTable, ObservedTable.css, TrendLane, TrendPlot.css, TrendPlotSvg, TrendRow, Trends.css, TrendsHeader, TrendsHelp, TrendsPanel, TrendsToggle, WatchButton, plotHandlers, useProbe, useTrends, useWatch                                                                     |
| `twin`       | Twin.css, TwinChartView, TwinForm, TwinInspect, TwinPanel, TwinParts, TwinResult, TwinVerdict, useTwinSeen                                                                                                                                                                                                                             |
| (root)       | ReviewScreen, ReviewScreen.css                                                                                                                                                                                                                                                                                                         |

**`model/review`**

| Folder       | Files                                                                                                                                                    |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `compare`    | compareChoices, comparePlan, compareTables, comparison                                                                                                   |
| `interrupts` | interruptCover, interruptMarks, interruptScope, interruptSelect, interruptWords, interrupts, sessionLive, sessionWords                                   |
| `playback`   | callLog, live, observedRows, playback, reviewRows, savedTrends, series, trend, trendOverlay, trendPlot, trendProbe, trendReadout, trendWindow, watchList |
| `report`     | actionStory, banners, noteLines, origins, report, reportViews, rerun, reviewAnchor, thisRun, timeLayout, timeline, words                                 |
| `twin`       | twin, twinChart, twinSelect, twinWords                                                                                                                   |

**`ui/launch`**

| Folder     | Files                                                                                                                                                                                                                                          |
| ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `controls` | ActNow, CallNow, ControlRefusal, FaultNow, RunControls, RunControls.css                                                                                                                                                                        |
| `plan`     | CallArgs, CallRows, DeliveryParams, DeliveryRows, FaultRows, LaunchForm, PaceChoice, PlanAgenda, PlanAgenda.css, PlanIdentity, PlanIdentity.css, PlanRows.css, Readiness, Readiness.css, RowParts, RunSettings, useLaunchDraft, usePlanCatalog |
| `progress` | LaunchBody, LaunchFailure, LaunchProgress, LaunchResult, ProposalNote, RunningMark                                                                                                                                                             |
| `session`  | PolicyField, RealRows, RunMode, Session, Session.css, SessionCheck, SessionFields, SessionForm, SessionLive, SessionPolicies, useRunMode                                                                                                       |
| (root)     | LaunchPanel, LaunchPanel.css                                                                                                                                                                                                                   |

**`ui/devices`**

| Folder      | Files                                                                                                                                                                                                                                |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `behaviour` | CheckPanel, CommandRow, DriveRows, HeldIssues, HeldNotice, InstrumentPicker, InterfaceNeeds, KindCard, NeedsFirst, Provenance.css, ReadingRow, Suggestions, TwinBar, TwinDetails, TwinForm, TwinScope, TwinSide, TwinStep, twinProps |
| `common`    | ConfirmButton, DeviceTitle, Field, FormTitle, LeaveDevice, Notes.css, stepFocus, useDeviceKeys, useStepNavigation                                                                                                                    |
| `interface` | ArgumentRows, CommandForm, DeviceBar, DeviceDetails, DeviceOutline, InterfaceStep, KindForm, MetricExtras, MetricForm, formProps                                                                                                     |
| (root)      | Behaviour.css, DeviceScreen, Devices.css, StepTabs                                                                                                                                                                                   |

**`adapters/bridge`**

| Folder       | Files                                                                                                                                                                                   |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `devices`    | captureDecoder, devices, kindIssues, lintAction, twinDecoders                                                                                                                           |
| `launch`     | connect, jobPlan, jobs, planCatalog, policyDefault                                                                                                                                      |
| `provenance` | assumptionsDecoder, lineageDecoder, provenanceGateway                                                                                                                                   |
| `review`     | compareDecoder, compareItems, compareResults, control, diffDecoder, interruptDecoder, journalDecoder, liveDecoder, reportDecoder, reportQuestions, twinReportDecoder, twinReportGateway |
| `transport`  | browserTransport, decodeKit, failures, http                                                                                                                                             |
| (root)       | client, decoders, follow, parameterDecoder                                                                                                                                              |
