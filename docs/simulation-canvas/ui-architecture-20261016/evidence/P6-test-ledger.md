# Phase 6 evidence — test ledger

- **Scope:** [PLAN.md](../PLAN.md) §5 Phase 6 against [BAR.md](../BAR.md) §2 (phase 6 row).
- **Subject:** every Vitest file under `canvas/src` (145 files). The 70 Playwright tests run against the real binary and are the
  unchanged real-wire oracle; none was touched and none is listed as removable.
- **Method:** each file was read; each outcome is stated at its lowest owning layer; the unique failure is the production
  mutation only that file would notice. Every removal or merge was proven by mutating the production behaviour the test pins
  (the mutation was reverted at once; `git diff -- src` shows no production file changed) and by confirming another surviving
  test fails for the same mutation.
- **Layer vocabulary:** `model` (pure rules), `app` (stores and coordinators over ports), `adapter` (wire to port), `ui`
  (jsdom rendering and wiring), `support` (test doubles and harnesses), `e2e` (real binary, listed only for context).

## 1. Suite size and time

| Measure                               | Before                     | After                      |
| ------------------------------------- | -------------------------- | -------------------------- |
| Test files                            | 145                        | 145                        |
| Tests                                 | 725                        | 722                        |
| `npx vitest run` Duration, three runs | 15.90 s, 15.61 s, 15.65 s  | 14.12 s, 14.10 s, 14.14 s  |
| Median Duration                       | **15.65 s**                | **14.12 s**                |
| `npx vitest run --coverage`           | green, every threshold met | green, every threshold met |

Three tests fewer cannot by themselves explain 1.5 s; jsdom creation (145 environments, about 54 % of tracked time) dominates
and the difference is mostly machine state. The honest reading: the suite time is unchanged within noise. The audit's value is
the ledger and the guarantee that nothing distinct was traded for a smaller count.

## 2. Coverage per module folder (before equals after)

Measured with `npx vitest run --coverage --coverage.reporter=json-summary` before any edit and again after the last edit,
aggregated per folder. The two outputs are byte-identical, so no folder lost a line, function or branch.

| Folder                             | Lines  | Functions | Branches |
| ---------------------------------- | ------ | --------- | -------- |
| `adapters/**` (floor 99 / 99 / 97) | 100.00 | 100.00    | 98.06    |
| `adapters/bridge`                  | 100.00 | 100.00    | 98.01    |
| `adapters/browser`                 | 100.00 | 100.00    | 100.00   |
| `app/**` (floor 80 / 80 / 75)      | 99.22  | 97.44     | 92.32    |
| `app/editor`                       | 98.60  | 96.30     | 85.93    |
| `app/launch`                       | 100.00 | 100.00    | 96.15    |
| `app/plant`                        | 99.01  | 95.45     | 88.76    |
| `app/review`                       | 98.56  | 95.83     | 96.72    |
| `app/session`                      | 100.00 | 100.00    | 100.00   |
| `app/shared`                       | 100.00 | 100.00    | 100.00   |
| `app/shell`                        | 100.00 | 100.00    | 97.37    |
| `app/workspace`                    | 100.00 | 100.00    | 100.00   |
| `model/**` (floor 90 / 90 / 85)    | 98.93  | 98.45     | 90.03    |
| `model/editor`                     | 100.00 | 100.00    | 93.14    |
| `model/graph`                      | 98.92  | 99.01     | 85.01    |
| `model/launch`                     | 100.00 | 100.00    | 98.43    |
| `model/plant`                      | 99.34  | 100.00    | 93.64    |
| `model/review`                     | 97.00  | 93.88     | 91.83    |
| `model/shared`                     | 100.00 | 100.00    | 100.00   |
| `model/shell`                      | 100.00 | 100.00    | 97.30    |
| `model/workspace`                  | 100.00 | 100.00    | 91.84    |
| `ui/**` (no floor)                 | 75.65  | 75.57     | 65.43    |
| `ui/editor`                        | 87.03  | 84.08     | 79.58    |
| `ui/graph`                         | 62.61  | 61.97     | 43.53    |
| `ui/launch`                        | 97.83  | 96.09     | 92.41    |
| `ui/plant`                         | 44.91  | 47.67     | 52.46    |
| `ui/primitives`                    | 97.40  | 97.06     | 95.16    |
| `ui/review`                        | 58.89  | 59.44     | 52.48    |
| `ui/session`                       | 95.24  | 100.00    | 87.50    |
| `ui/shell`                         | 89.68  | 87.50     | 81.71    |
| `ui/workspace`                     | 97.06  | 96.34     | 88.51    |

Statement coverage is saturated: with every test run alone, 298 of 646 cases add no statement or branch that another test of
the same or a lower layer does not already reach. Coverage therefore cannot prove a duplicate; only the assertion and the
mutation can. That is why each removal below is argued from a mutation, not from a coverage number.

## 3. Changes made

### 3.1 Removed or merged tests (725 to 722)

| #   | Verdict          | Test                                                                                                                                                                                                                                      | Surviving test that keeps the failure                                                                                                                                                                                                                          | Proof                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| --- | ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | remove-duplicate | `ui/workspace/WorkspaceScreen.test.tsx` "shows the badge of a run already checked in this session"                                                                                                                                        | `ui/workspace/RunList.test.tsx` "gives a run with no recorded end its own chip, not the success tokens Verified uses" (renders the same `RunList` with a `badgeFor` that returns `verified` and asserts the `Verified` text and chip class)                    | Three mutations, each failed both tests and nothing else: `RunList.tsx` `badge={badgeFor(...)}` to `badge={null}`; `WorkspaceScreen.tsx` passing `() => null` to `Contents`; `RunList.tsx` `Status` rendering `Checked` instead of the `verified` label                                                                                                                                                                                             |
| 2   | merge            | `app/review/internal/reviewStore.test.ts` "finishes a run listed as open when its journal closes while the review watches" into `reviewCompletion.test.ts` "finishes the review and loads the recorded plant when the watched run closes" | the merged test (same scenario: listing open, report open, journal closes; it now asserts the empty reads and the single `run.report` before close, then two `run.report`, one `run.verify`, the recorded plant, the ended report and the ready verdict after) | Three mutations (drop the `playback.phase` guard in `refreshIfLive`; drop `domain` from the deferred set in `refreshIfLive`; disable the `domain` branch of `resumeFinalLoads`) were applied to the original files and to the merged files: the failing sets are the same, with the removed test replaced by the merged test. Dropping the report reload in `refreshIfLive` fails the merged test. The union of both tests' assertions is preserved |
| 3   | merge            | `ui/shell/App.test.tsx` "asks an unpaired browser for the pairing code" into "asks an unpaired browser for the pairing code, then goes to the workspace once it pairs"                                                                    | the merged test (the heading assertion moved to its start, every other assertion unchanged)                                                                                                                                                                    | The removed test was a strict prefix of the other (same render, same `FakeBridge`); no assertion was dropped                                                                                                                                                                                                                                                                                                                                        |

### 3.2 Fixture simplified

- `src/test/plantBuilder.ts` (a simulation of the binary's `component.new`: merge a frame over the base, copy the id and
  title, strip the body for an empty start) is deleted, together with its fallback in `FakeBridge.canned`. Seven
  `WorkspaceScreen` tests relied on it (found by breaking the fallback: 7 failures, all in that file). They now set a canned
  `component.new` payload (`generated(...)` helper) and assert what the UI owns: the arguments it sends (`id`, `title`,
  `from` or `example`, `empty`), the path it writes to, the hash it navigates to and the alert it shows. The behaviour the
  fake used to imitate is pinned by the real binary: `e2e/editor.spec.ts` (copy from a plant, copy from an example, full
  copy runnable) and `e2e/empty-workspace.spec.ts` (blank plant from a name and an id), plus
  `app/plant/plantMaker.test.ts`, which already used canned payloads.
- Remaining policy in `FakeBridge`, kept on purpose: it is served as canned wire responses through the real
  `bridgeAdapter`, so it exercises the real decoders and failure mappers (breaking `pair`'s mapper in `client.ts` fails the
  adapter test, `sessionStore` and `PairingScreen` together). The two pieces of state it keeps are the pairing flag flipped
  by a successful `pair` and the etag progression of `PUT /api/file`; the editor store tests assert on both (the version a
  save writes over). The real behaviour is covered by `e2e/pairing.spec.ts` and `e2e/editor.spec.ts`; they stay in the
  double because a canned queue cannot express "the next write returns the etag the last write returned" without
  reimplementing it in every test.

### 3.3 Titles

No title was `works`, `test 1` or a bare number, and no `describe` name is duplicated across the suite (checked with a
scripted scan of all 722 titles). Twenty-one `describe` names were made specific so a failure line identifies the module:
`launch store`, `review store`, `session store`, `panel store`, `navigation store`, `theme store`, `app shell`,
`plant inspector`, `run progress words`, `ready report`, `plant title`, `library payload`/`presets payload`/
`examples payload`/`component payload` (in `adapters/bridge/internal/components.test.ts`), `simulated clock`,
`outline rail`, `shared words`, `bridge adapter wiring`, `area counts`, `unit grouping`; `AlarmJump.test.tsx` gained its
missing `describe("alarm jump")`. No assertion changed.

## 4. Duplicates examined and kept

Each pair was read and then mutated. A pair is kept when mutating the named production line fails only one of them, or
fails the lower-layer test plus tests of a higher layer (which do not qualify as owners).

| Suspected duplicate                                                                                                             | Why it is not one                                                                                                                                                                                                    | Mutation and result                                                                                                                                                                                                                |
| ------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ui/review/RunStatus.test.tsx` "warns when playback holds fewer observations" vs `model/review/banners.test.ts`                 | the model pins the warning text; the UI pins that `Banners` feeds `state.playback` into `bannerWarnings`                                                                                                             | `RunStatus.tsx` passes a constant playback record: only the UI test fails                                                                                                                                                          |
| `ui/editor/AddEquipment.test.tsx` vs `model/editor/addWords.test.ts`                                                            | the model pins `libraryWords`/`libraryRows`; the UI pins the count handed to them                                                                                                                                    | `found.length` to `found.length + 5` in `AddEquipment.tsx`: both `AddEquipment` tests and `EditorPanel` "adds equipment" fail, no model test                                                                                       |
| `ui/editor/InstanceEditor.test.tsx` vs `model/editor/linkWords.test.ts`                                                         | the model pins words and part order; the UI pins how the parts are drawn                                                                                                                                             | arrow glyph changed in `InstanceEditor.tsx`: only the UI test fails                                                                                                                                                                |
| `adapters/bridge/bridgeClient.test.ts` vs `http.test.ts`, `failures.test.ts`, `client.test.ts`                                  | it is the one end-to-end bridge case: transport, envelope, mapper and decoder through each gateway; the others pin one stage each                                                                                    | `pair` wired to `workspaceFailure`: it fails, and so do `sessionStore` (3) and `PairingScreen` (2) through `FakeBridge`, which are higher layers, so it stays as the lowest owner                                                  |
| `adapters/bridge/failures.test.ts` editor block vs plant block                                                                  | `editorFailure` is an alias of `plantFailure`; the editor block is the only adapter-level test of the alias                                                                                                          | `editorFailure` returning `plantFailure("read")`: the editor block fails, plus 6 higher-layer tests                                                                                                                                |
| `app/shell/proposeRerun.test.ts` vs `app/review/rerunProposal.test.ts` vs `ui/review/RunHeader.test.tsx`                        | proposal, coordinator and button wiring                                                                                                                                                                              | `RunHeader.tsx` button no longer calls `proposeRerun`: only the `RunHeader` test fails                                                                                                                                             |
| `reviewStore.test.ts` "reviews a run the listing marks did-not-finish..." vs "answers a run the listing still shows as open..." | different listing input: an abandoned run (`ended`) and a run still open; both are unfinished                                                                                                                        | `knownUnfinished` also requiring `ended`: the open-listing test fails (with two `reviewCompletion` tests); also requiring `!ended`: the did-not-finish test fails (with two others). Each test catches a mutation the other cannot |
| `app/review/reviewStore.test.ts` "re-asks nothing doomed..."                                                                    | the only test that reaches the `refreshed` flag of an abandoned listing                                                                                                                                              | `refreshed: known?.ended === true` to `false`: only this test fails                                                                                                                                                                |
| `ui/editor/EditorStatus.test.tsx` "says which list holds the issues..." vs `EditorPanel.test.tsx` save-with-issues              | both assert the save sentence, only the first drives `Show issues` and the focus move                                                                                                                                | removing `request.ask("build")` in `EditorBar.tsx`: only the `EditorStatus` test fails                                                                                                                                             |
| `model/graph/wirePins.test.ts` vs `wireOffice.test.ts` vs `wireView.test.ts`                                                    | three oracles over the router (pin entry side and nothing behind a node; corridor bound; anchor arithmetic)                                                                                                          | not mutation-tested; kept as distinct oracles by reading the assertions                                                                                                                                                            |
| Class-token tests (`Outline` row style, `RunList` chip class, `MetaParts` span class)                                           | they pin CSS-module class names because jsdom computes no style; the weakest tests in the suite, and no other layer guards the regression (no e2e computed-style check for the heading uppercase or the chip tokens) | not mutation-tested; a candidate for an e2e computed-style check later, not removed                                                                                                                                                |

## 5. The ledger

Columns: **Outcome** is what the file pins; **Layer** is the lowest owning layer; **Unique failure** is what only this file
would notice; **Verdict** is `keep` unless stated (the three changed files are marked).

### 5.1 `adapters/bridge` (9 files, 143 tests)

| File                       | Outcome                                                                                                                                                                                             | Layer   | Unique failure                                                            | Verdict                               |
| -------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ------------------------------------------------------------------------- | ------------------------------------- |
| `bridgeClient.test.ts`     | The real adapter over a canned transport: pair body and credentials, failure mapping through each gateway, file read, verify verdict recovery, `advance` narrowing, journal follow, workspace watch | adapter | A gateway wired to the wrong request, decoder or mapper                   | keep (the one end-to-end bridge case) |
| `browserTransport.test.ts` | `fetch`/`EventSource` are reached through `window` with the same input                                                                                                                              | adapter | Transport that alters the request or opens another URL                    | keep                                  |
| `client.test.ts`           | Wire names and bodies of launch, plant, editor and review gateways; closed refusals of a malformed answer                                                                                           | adapter | Wrong action name, argument key or unreadable-vs-ok decision              | keep                                  |
| `components.test.ts`       | Library, presets, examples and component payload decoding and the category rule                                                                                                                     | adapter | A decoder that accepts a malformed library or mis-categorises a class     | keep                                  |
| `decoders.test.ts`         | Simple payloads, acknowledgement, workspace, run domain, confirmations, composition: accept and refuse                                                                                              | adapter | A decoder accepting a malformed twin                                      | keep                                  |
| `failures.test.ts`         | Protocol code to feature failure for session, workspace, plant, editor, launch, review                                                                                                              | adapter | A code mapped to the wrong feature failure or a protocol code leaking out | keep                                  |
| `http.test.ts`             | Error envelope read; transport failure kept apart from an unreadable body                                                                                                                           | adapter | Envelope fields lost or a transport error reported as a body error        | keep                                  |
| `journalDecoder.test.ts`   | Journal record kinds, scalar reading values, refusal of a bad line                                                                                                                                  | adapter | A malformed journal line accepted                                         | keep                                  |
| `reportDecoder.test.ts`    | Report sections, alarm instants, scoring, replay verdict                                                                                                                                            | adapter | Report field dropped or alarm instant invented                            | keep                                  |

### 5.2 `adapters/browser` (3 files, 13 tests)

| File               | Outcome                                                       | Layer   | Unique failure                              | Verdict |
| ------------------ | ------------------------------------------------------------- | ------- | ------------------------------------------- | ------- |
| `browser.test.ts`  | Preferences survive a throwing storage; media query forwarded | adapter | Storage exception escaping                  | keep    |
| `clock.test.ts`    | Wall clock, scheduler and abortable delay (listener detached) | adapter | A leaked timer or listener                  | keep    |
| `location.test.ts` | Hash read, write (push/replace), idempotence, notification    | adapter | A hash write that adds history when current | keep    |

### 5.3 `app/editor`, `app/launch`, `app/plant` (9 files, 60 tests)

| File                              | Outcome                                                                                                              | Layer | Unique failure                                                         | Verdict |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------------- | ----- | ---------------------------------------------------------------------- | ------- |
| `app/editor/editorStore.test.ts`  | Open, edit, save over the last etag, compose over the domain read, issues on instances, undo/redo, late-visit guards | app   | A save over a stale etag; a late answer applied to another composition | keep    |
| `app/editor/saveStatus.test.ts`   | Status text and tone; `checked`, `canSave`, `isDirty`, `isSaving` selectors                                          | app   | A wrong Save availability or status sentence                           | keep    |
| `app/launch/launchStore.test.ts`  | Create, inject, advance, end; refusals; stop early; pacing; stop wakes a waiting run                                 | app   | A step order change or a leaked world                                  | keep    |
| `app/launch/proposalSlot.test.ts` | A proposal is offered to its plant only, latest wins                                                                 | app   | A proposal offered to another plant                                    | keep    |
| `app/launch/runCursor.test.ts`    | Speed sampled over five seconds                                                                                      | app   | A wrong rate or an unbounded sample window                             | keep    |
| `app/launch/runWatch.test.ts`     | Hold a watched run two seconds, ignore other plants, release                                                         | app   | A hold on a run nobody watched                                         | keep    |
| `app/plant/baseChoice.test.ts`    | What a new plant starts from; missing-field order                                                                    | app   | A wrong first missing field                                            | keep    |
| `app/plant/plantMaker.test.ts`    | Write what the binary generated, compose a copy not an empty frame, refusals, examples                               | app   | The maker composing an empty frame or writing the wrong file           | keep    |
| `app/plant/plantStore.test.ts`    | Load, selection, reload on change, overlapping and late loads                                                        | app   | A late load applied; a lost selection                                  | keep    |

### 5.4 `app/review`, `app/session`, `app/shared`, `app/shell`, `app/workspace` (18 files, 77 tests)

| File                                   | Outcome                                                                             | Layer | Unique failure                                          | Verdict                                            |
| -------------------------------------- | ----------------------------------------------------------------------------------- | ----- | ------------------------------------------------------- | -------------------------------------------------- |
| `app/review/readyReport.test.ts`       | `readyReport` is null until loaded; start time                                      | app   | A report exposed before it is ready                     | keep                                               |
| `app/review/rerunProposal.test.ts`     | The rerun proposal from a ready report                                              | app   | A proposal built without the report                     | keep                                               |
| `app/review/reviewBadge.test.ts`       | Badge recorded once report and verdict arrived; kept through playback               | app   | A badge recorded early or lost on reopen                | keep                                               |
| `app/review/reviewCompletion.test.ts`  | A run listed before completion: close, publication, stale listing, ended incomplete | app   | Plant not loaded or verdict asked for an unfinished run | keep (absorbs the open-listing close test, 3.1 #2) |
| `app/review/reviewLifecycle.test.ts`   | Superseded runs and retained callbacks publish nothing                              | app   | A closed session publishing                             | keep                                               |
| `app/review/reviewOrdering.test.ts`    | The newest report wins over a late earlier one                                      | app   | An earlier request overwriting a newer report           | keep                                               |
| `app/review/reviewPlant.test.ts`       | The plant a review draws and its fallbacks                                          | app   | Wrong plant drawn or title                              | keep                                               |
| `app/review/reviewStore.test.ts`       | Report, verify, journal stream, playback, failure and listing-aware loads           | app   | A doomed verdict asked for; playback past the end       | keep (one test merged away, 3.1 #2)                |
| `app/session/sessionStore.test.ts`     | Pairing phases, code normalisation, blocked refusals                                | app   | Typing allowed after a lock                             | keep                                               |
| `app/shared/observable.test.ts`        | Notify on change only                                                               | app   | A listener woken for an equal state                     | keep                                               |
| `app/shared/operationIdentity.test.ts` | Newer operation invalidates older                                                   | app   | A resurrected old operation                             | keep                                               |
| `app/shell/navigationStore.test.ts`    | Address read/write, unsaved guard, back-button hold                                 | app   | A guard disposer removing its replacement               | keep                                               |
| `app/shell/panelStore.test.ts`         | Panels default open, toggle and persist                                             | app   | A panel choice not remembered                           | keep                                               |
| `app/shell/proposeRerun.test.ts`       | Propose and open the launch panel; nothing before the report                        | app   | Navigation without a proposal                           | keep                                               |
| `app/shell/sessionGate.test.ts`        | Workspace opens only when paired, re-checks on unpaired                             | app   | Workspace opened for an unpaired browser                | keep                                               |
| `app/shell/showBuild.test.ts`          | Build button swaps the inspector and toggles                                        | app   | Build panel stuck over a selection                      | keep                                               |
| `app/shell/themeStore.test.ts`         | Stored preference and system theme follow                                           | app   | Theme not persisted                                     | keep                                               |
| `app/workspace/workspaceStore.test.ts` | Listing, watch once, ordered overlapping reloads, close                             | app   | A reload published out of request order                 | keep                                               |

### 5.5 `model/*` (54 files, 253 tests)

| File                                    | Outcome                                                                   | Layer | Unique failure                                                     | Verdict |
| --------------------------------------- | ------------------------------------------------------------------------- | ----- | ------------------------------------------------------------------ | ------- |
| `model/editor/addWords.test.ts`         | First missing field of the add form; library row and count words          | model | A wrong problem order or count sentence                            | keep    |
| `model/editor/areaEdits.test.ts`        | Add, rename and delete areas with their rules                             | model | An area deleted while it holds instances                           | keep    |
| `model/editor/composition.test.ts`      | Composition document edits keep every other field                         | model | A field lost by an edit                                            | keep    |
| `model/editor/compositionKinds.test.ts` | Instances of one component share or split an entry                        | model | A reused kind name                                                 | keep    |
| `model/editor/instanceNameRule.test.ts` | The instance name rule text                                               | model | A rule drifting from the binary's                                  | keep    |
| `model/editor/issueTargets.test.ts`     | Binary issue pointers become instance targets and words                   | model | An issue assigned to the wrong instance                            | keep    |
| `model/editor/linkWords.test.ts`        | Link words and parts, delete question, name check                         | model | An arrow pointing the wrong way                                    | keep    |
| `model/graph/areas.test.ts`             | Area counts in plant order                                                | model | Areas sorted by count                                              | keep    |
| `model/graph/frameTitle.test.ts`        | Frame title cut, lines and bounds                                         | model | A cut word or a title past its cap                                 | keep    |
| `model/graph/graphKeys.test.ts`         | Arrow keys move to the nearest node                                       | model | A move to a node behind                                            | keep    |
| `model/graph/halls.test.ts`             | Hall prefix, single-prefix plant, members by kind                         | model | A single prefix read as a hall                                     | keep    |
| `model/graph/headerFit.test.ts`         | Header text cut and shrink to the pixel floor                             | model | Id cut before it shrank                                            | keep    |
| `model/graph/layout.test.ts`            | Ranking, columns, area packing, header reserve                            | model | Unstable ranking or overlapping rows                               | keep    |
| `model/graph/outline.test.ts`           | Outline groups, issue marks, empty sub-area order                         | model | An empty area between populated ones                               | keep    |
| `model/graph/planMarks.test.ts`         | Planned fault and call marks over time                                    | model | A mark shown outside its window                                    | keep    |
| `model/graph/plantView.test.ts`         | View nodes, pins, edges, halls, note lines                                | model | A wrong pin or edge                                                | keep    |
| `model/graph/readout.test.ts`           | Readout vocabulary, range flag, attention, hall roll-up, times            | model | A reading outside range unflagged                                  | keep    |
| `model/graph/viewport.test.ts`          | Focus, tiers, opening floor, fit, clamp                                   | model | A selection left off screen                                        | keep    |
| `model/graph/wireGrid.test.ts`          | Frame margin lines                                                        | model | Hug lanes mismarked                                                | keep    |
| `model/graph/wireOffice.test.ts`        | Office-tower corridor bound at every tier                                 | model | A wire leaving its corridor (distinct oracle)                      | keep    |
| `model/graph/wirePath.test.ts`          | Rounded path string and label point                                       | model | A wrong corner radius                                              | keep    |
| `model/graph/wirePins.test.ts`          | Every wire ends on its pin; none behind a node                            | model | A wire entering the wrong side (distinct oracle)                   | keep    |
| `model/graph/wireRoute.test.ts`         | Router clearance, determinism, no shared corridor                         | model | A wire cutting a node or title                                     | keep    |
| `model/graph/wireSearch.test.ts`        | Lane occupancy and hugged-lane cost                                       | model | Two wires on one lane                                              | keep    |
| `model/graph/wireView.test.ts`          | Pin row anchors, spacing, tiers, title box                                | model | Anchors off the pin row                                            | keep    |
| `model/launch/draft.test.ts`            | Typed text to plan, plan to draft                                         | model | A unit or NaN conversion error                                     | keep    |
| `model/launch/launch.test.ts`           | Steps, plan problems, effector calls, pace                                | model | A step order change                                                | keep    |
| `model/launch/progress.test.ts`         | Progress words                                                            | model | Speed shown when unknown                                           | keep    |
| `model/launch/words.test.ts`            | Mode, problem, cue, summary words                                         | model | A cue naming the wrong field                                       | keep    |
| `model/plant/domainFile.test.ts`        | Domain file and id read                                                   | model | A non-text id spelled out                                          | keep    |
| `model/plant/effectors.test.ts`         | Effector specs, call arguments, argument problems                         | model | A typed argument wrongly built                                     | keep    |
| `model/plant/plant.test.ts`             | Plant read from a domain and composition                                  | model | A lost link port                                                   | keep    |
| `model/plant/ports.test.ts`             | Port families, pin marks, matching ports, no-target reason                | model | Hue as the only carrier                                            | keep    |
| `model/plant/title.test.ts`             | Title falls back to id                                                    | model | An empty title shown                                               | keep    |
| `model/review/banners.test.ts`          | Banner warnings                                                           | model | A warning text drift                                               | keep    |
| `model/review/live.test.ts`             | A node's and a hall's live values                                         | model | A stale value shown                                                | keep    |
| `model/review/noteLines.test.ts`        | Note line reservation                                                     | model | A node too short for its note                                      | keep    |
| `model/review/playback.test.ts`         | Channels, last reading, settle                                            | model | A reading after the time returned                                  | keep    |
| `model/review/report.test.ts`           | Badge, replay line, span, wall clock, KPI filter, identity, rerun, faults | model | The badge derived from its own checks                              | keep    |
| `model/review/reportViews.test.ts`      | File line, KPI figures, steady rows                                       | model | Decimals on large values                                           | keep    |
| `model/review/reviewAnchor.test.ts`     | The node a review opens on                                                | model | The wrong first node                                               | keep    |
| `model/review/reviewRows.test.ts`       | Readout rows per port, hidden channels                                    | model | A port row missing                                                 | keep    |
| `model/review/timeLayout.test.ts`       | Event chips and ticks                                                     | model | Overlapping chips                                                  | keep    |
| `model/review/timeline.test.ts`         | Event marks, lanes, next alarm                                            | model | A wrap past the last alarm                                         | keep    |
| `model/review/words.test.ts`            | Alarm figure words, mode words                                            | model | A wrong explanation                                                | keep    |
| `model/shared/areaTree.test.ts`         | Area order, path and depth; orphan and cycle kept                         | model | A lost orphan area (already beside `areaTree.ts`; imports only it) | keep    |
| `model/shared/collect.test.ts`          | Group and tally                                                           | model | Key order change                                                   | keep    |
| `model/shared/duration.test.ts`         | Clock and duration words                                                  | model | Negative time shown                                                | keep    |
| `model/shared/json.test.ts`             | Narrowing helpers                                                         | model | A non-record accepted                                              | keep    |
| `model/shared/units.test.ts`            | Thousands grouping                                                        | model | A wrong separator                                                  | keep    |
| `model/shared/words.test.ts`            | Failure line, humanise, kind label                                        | model | A wrong humanised id                                               | keep    |
| `model/shell/route.test.ts`             | Route round trip and fallback                                             | model | A screen unreachable                                               | keep    |
| `model/shell/theme.test.ts`             | Theme cycle and labels                                                    | model | Cycle order                                                        | keep    |
| `model/workspace/workspace.test.ts`     | Plants, runs, filters, when words, suggested id                           | model | A duplicate id suggested                                           | keep    |

`model/plant/ports.test.ts` builds plants with `compositionPlant` from `model/editor` through `src/test/modelFixtures.ts`;
that is input construction, not a test of the editor code, and `model/shared/areaTree.test.ts` was already beside its code.
Nothing was moved.

### 5.6 `ui/*` (49 files, 167 tests)

| File                                       | Outcome                                                                        | Layer | Unique failure                                                     | Verdict                                                |
| ------------------------------------------ | ------------------------------------------------------------------------------ | ----- | ------------------------------------------------------------------ | ------------------------------------------------------ |
| `ui/editor/AddEquipment.test.tsx`          | Library count and scroll cue wiring                                            | ui    | Count handed to `libraryWords` wrong                               | keep                                                   |
| `ui/editor/AreaEditor.test.tsx`            | Area rename everywhere; delete only an empty area                              | ui    | An occupied area deletable                                         | keep                                                   |
| `ui/editor/ConnectForm.test.tsx`           | Link targets named by id and kind; missing-target reason                       | ui    | A wrong target label                                               | keep                                                   |
| `ui/editor/EditorPanel.test.tsx`           | Add, rename, move, link, unlink, delete, save, issues, undo                    | ui    | Any panel control wired to the wrong edit                          | keep                                                   |
| `ui/editor/EditorStatus.test.tsx`          | Show issues, unchecked plant Save, profile weight remedy                       | ui    | The remedy removing another weight                                 | keep                                                   |
| `ui/editor/EmptyEditor.test.tsx`           | First step guidance for an empty plant (axe)                                   | ui    | Guidance kept after equipment exists                               | keep                                                   |
| `ui/editor/InstanceEditor.test.tsx`        | Link list wording and ends                                                     | ui    | Link read from the wrong end                                       | keep                                                   |
| `ui/editor/LeaveDialog.test.tsx`           | Stay, Escape, Discard, Save and leave, no discard while saving                 | ui    | Leaving with unsaved changes                                       | keep                                                   |
| `ui/graph/GraphControls.test.tsx`          | Zoom tooltips, legend in review only, close rules                              | ui    | A legend left open                                                 | keep                                                   |
| `ui/graph/Outline.test.tsx`                | Selection, filter, indentation, empty areas, scroll cue                        | ui    | An area heading uppercased (class token)                           | keep                                                   |
| `ui/graph/PlantGraph.test.tsx`             | Description by editor or viewer mode                                           | ui    | Delete wording promised where nothing deletes                      | keep                                                   |
| `ui/graph/flow.test.ts`                    | Wire edges, markers and frame titles for the renderer                          | ui    | Edges drawn over nodes                                             | keep                                                   |
| `ui/graph/labelFit.test.ts`                | Label hide, short form, refit at zoom                                          | ui    | A label overflowing its box                                        | keep                                                   |
| `ui/launch/LaunchCalls.test.tsx`           | Effector call rows and the review gate                                         | ui    | A call row not checked                                             | keep                                                   |
| `ui/launch/LaunchFailure.test.tsx`         | A refused fault named in one sentence                                          | ui    | Binary message lost                                                | keep                                                   |
| `ui/launch/LaunchKeys.test.tsx`            | Enter in a select does not start the run                                       | ui    | A run started by Enter                                             | keep                                                   |
| `ui/launch/LaunchPanel.test.tsx`           | Plan, validation, run, stop, picking, proposals                                | ui    | Run started with problems                                          | keep                                                   |
| `ui/launch/LaunchPrefill.test.tsx`         | A prefilled fault the plant lost is cleared                                    | ui    | A stale fault run                                                  | keep                                                   |
| `ui/launch/LaunchRunning.test.tsx`         | A run outlives its panel; another plant's run                                  | ui    | A run cancelled by closing                                         | keep                                                   |
| `ui/plant/Inspector.test.tsx`              | Instance kind, links, faults, title (axe)                                      | ui    | A wrong link side                                                  | keep                                                   |
| `ui/plant/InspectorLinks.test.tsx`         | Hall and receiver ordering, select                                             | ui    | Links ordered as text                                              | keep                                                   |
| `ui/plant/PlantFailure.test.tsx`           | Missing plant and any other failure                                            | ui    | A failure without a way back                                       | keep                                                   |
| `ui/plant/PlantOverview.test.tsx`          | Campus and area counts, order agreement with the rail, recent runs             | ui    | The rail and list disagreeing                                      | keep                                                   |
| `ui/primitives/Breadcrumb.test.tsx`        | Crumbs in place or in the top bar slot, long names                             | ui    | A truncated name without a tooltip                                 | keep                                                   |
| `ui/primitives/MetaParts.test.tsx`         | Separator glued to the part before it, empty parts dropped                     | ui    | A stranded dot (class equality is an implementation detail, noted) | keep                                                   |
| `ui/primitives/OutsideMark.test.tsx`       | Out-of-range mark present only when named                                      | ui    | A mark with no text                                                | keep                                                   |
| `ui/primitives/holdEnter.test.tsx`         | Enter on a select held, others submit                                          | ui    | A select submitting a form                                         | keep                                                   |
| `ui/primitives/keys.test.ts`               | Text field, Space ownership, plain key                                         | ui    | A shortcut firing in a field                                       | keep                                                   |
| `ui/primitives/useCutBottom.test.tsx`      | Cut cue on scroll, disclosure and text change                                  | ui    | A cue stuck on                                                     | keep                                                   |
| `ui/primitives/useInterval.test.tsx`       | Latest callback each period; stop                                              | ui    | An interval leaking                                                | keep                                                   |
| `ui/primitives/useResizeObserver.test.tsx` | Observe, latest callback, disconnect                                           | ui    | An observer leaking                                                | keep                                                   |
| `ui/primitives/useScreen.test.tsx`         | Screen heading takes focus except over a dialog                                | ui    | Focus stolen from a dialog                                         | keep                                                   |
| `ui/review/AlarmJump.test.tsx`             | Next alarm and wrap, disabled when none                                        | ui    | No wrap at the end                                                 | keep (describe added)                                  |
| `ui/review/MissingRun.test.tsx`            | Missing run alert with a way back                                              | ui    | A dead end                                                         | keep                                                   |
| `ui/review/PlayControls.test.tsx`          | Play and pause announced on change only                                        | ui    | A silent state change                                              | keep                                                   |
| `ui/review/ReportPanel.test.tsx`           | Every report section, integrity block, KPI filter, verify caution              | ui    | A section missing (axe)                                            | keep                                                   |
| `ui/review/RunHeader.test.tsx`             | Set up another run from this run                                               | ui    | The button bypassing the proposal                                  | keep                                                   |
| `ui/review/RunInfo.test.tsx`               | Disclosure open, Escape, blur                                                  | ui    | Focus lost on Escape                                               | keep                                                   |
| `ui/review/RunStatus.test.tsx`             | Badge, did-not-finish, file changed, verify failed, playback banners           | ui    | A banner missing its alert role                                    | keep                                                   |
| `ui/review/Scrubber.test.tsx`              | Clock, slider, events, speed, next event                                       | ui    | A seek to the wrong time                                           | keep                                                   |
| `ui/session/PairingScreen.test.tsx`        | Code grouping, wrong code alert, lock, paired renders nothing (axe)            | ui    | A field editable after a lock                                      | keep                                                   |
| `ui/shell/App.test.tsx`                    | Pairing to workspace, unreachable bridge, theme cycle, axe                     | ui    | The shell wired to the wrong store                                 | keep (merge, 3.1 #3)                                   |
| `ui/shell/KeyHelp.test.tsx`                | Key help per screen, focus, scrim                                              | ui    | A key listed for another screen                                    | keep                                                   |
| `ui/shell/PanelToggles.test.tsx`           | Panel keys, focus off a hidden control, right toggle                           | ui    | Focus lost in a hidden panel                                       | keep                                                   |
| `ui/shell/RunIndicator.test.tsx`           | Progress, link back, review, dismiss                                           | ui    | A failure that cannot be dismissed                                 | keep                                                   |
| `ui/shell/TopBar.test.tsx`                 | Skip link; theme button text                                                   | ui    | A skip link that does not focus                                    | keep                                                   |
| `ui/workspace/NewPlant.test.tsx`           | Derived id until typed; id rule shown on leave                                 | ui    | An id re-derived over a typed one                                  | keep                                                   |
| `ui/workspace/RunList.test.tsx`            | Did-not-finish, chip tokens, singular "Show 1 more run"                        | ui    | An unfinished run shown as checked                                 | keep (carries the badge assertion of 3.1 #1)           |
| `ui/workspace/WorkspaceScreen.test.tsx`    | Plants and runs, filter and paging, create from base or example, blank, errors | ui    | A create flow sending the wrong arguments                          | keep (one test removed, 3.1 #1; fixture replaced, 3.2) |

### 5.7 `src/test` support (3 files, 9 tests)

| File                             | Outcome                                                                                         | Layer       | Unique failure                               | Verdict |
| -------------------------------- | ----------------------------------------------------------------------------------------------- | ----------- | -------------------------------------------- | ------- |
| `test/fakeClock.test.ts`         | The fake clock honours time order, cancel, abort and pending reports as the real scheduler does | support     | A double that drifts from `adapters/browser` | keep    |
| `test/strictMode.test.tsx`       | The strict-mode harness replays effects and renders per mount                                   | support     | A harness that stopped catching leaks        | keep    |
| `test/strictMode.route.test.tsx` | A route survives setup-cleanup-setup with one watcher                                           | app (route) | A session surviving one mount                | keep    |

## 6. Totals by folder after the audit

| Folder           | Files   | Tests   |
| ---------------- | ------- | ------- |
| `adapters`       | 12      | 156     |
| `app`            | 27      | 137     |
| `model`          | 54      | 253     |
| `ui`             | 49      | 167     |
| `test` (support) | 3       | 9       |
| **Total**        | **145** | **722** |
