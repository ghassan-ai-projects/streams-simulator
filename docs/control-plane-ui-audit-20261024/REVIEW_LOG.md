# Browser review log and evidence

## Method and boundaries

Date: 24 October 2026, approximately 11:29–11:45 Europe/Berlin. Review used Chrome at its existing desktop viewport, with screenshots 1841 × 1099. No viewport override, mobile testing, other-size testing, or theme matrix was performed. The browser was in its existing system/light appearance.

The review followed an operator journey rather than inspecting source files. Repository instructions, README, canvas acceptance criteria, and operating/CLI documentation were read to launch and understand the supported product. No production source, tests, stylesheets, or application implementation was inspected. Build commands executed the existing project tooling; this was not a code review.

Findings derive from visible Chrome UI, accessible/rendered-page content, screenshots, and the resulting recording inventory. Files named `.txt` in evidence contain browser DOM snapshots, not source code. Screenshots sometimes capture a later progress tick than the paired snapshot; the active-run identity and relevant UI state remain the same.

This is an expert operator heuristic audit, not a measured usability study. No frame-rate benchmark, axe scan, full keyboard traversal, backend payload audit, adversarial access test, or physical safety test was performed. One post-reload browser-console check returned no captured warnings/errors; it is not evidence of a clean console across the entire review.

## Checkout and runtime

- Initial worktree was clean on `CC2640R2F-wireless-microcontroller`.
- Fetched `origin/main`, then created `codex/control-plane-ui-audit-20261024` at `3c6d78c8faa6ce86ac8119ef34d31536a9668c67`.
- Ran `make build` and `npm run build` in `canvas/`; both completed. Vite emitted a PostCSS `from` warning and a bundle-size warning. These are build warnings, not findings about observed UI performance.
- Existing server on port 8765 was left untouched. The audit used a separate loopback server on `http://127.0.0.1:8766` and its own paired browser session. The pairing code is intentionally omitted from the record.
- The isolated audit server was stopped after the review; its workspace and recordings were retained.
- Isolated temporary workspace: `/private/tmp/streamsim-ui-audit-20261024` (also reachable through `/tmp/streamsim-ui-audit-20261024`). It contains runtime artifacts and authored model data; they are not added under `docs/`.
- Copied shipped domain/example data into its `domains/`, factory/chiller compositions into `compositions/`, and the example thermal/BLE device definitions into `kinds/`. This setup used data files, not implementation inspection.
- Created an owner-only `sockets/` directory for the absent-host test. No real board or successful host integration was available.

Command used to start the isolated bridge:

```sh
bin/streamsim canvas --workspace /tmp/streamsim-ui-audit-20261024 --ui-dir canvas/dist --addr 127.0.0.1:8766
```

The build was necessary to audit current main rather than an unknown pre-existing bundle. No code or test change was made. Application and architecture gates were not run because the deliverable is documentation only; document links, evidence references, scope, and whitespace were checked instead.

Final documentation checks: four Markdown documents, 28 finding headings, 50 evidence files, no missing local Markdown targets, no trailing whitespace, and `git diff --check` passed. Git status contained only the new audit folder. No commit or push was made.

## Run inventory

| Run | Intent | Browser result |
|---|---|---|
| `w-1` | Factory; Explore; seed 42; 30 min; Watch; CT-1 fan failure at 5 min; AC-1 fails to load at 8 min; CH-1 set setpoint 9 at 12 min | Completed; UI Verified; 14,490 observations; effector applied, 251 ms reported; no outside-system score |
| `c-093f3e8c` | Factory; all simulated; host `/tmp/ui-audit-absent-host.sock`; no allowed policy digests | Failed security/path precondition: socket location not owner-only; attempt retained |
| `c-6dc5acbb` | Factory; all simulated; host in isolated owner-only sockets directory; no allowed policy digests | Failed `file_not_found`; later review Verified, zero time/observations, ordinary KPI/alarm tables |
| `w-2` | Authored twin; Evaluate; seed 1; 30 min; Watch; fan set cooling 40 at 1 min; temperature dropout at 5 min | Reload around minute 6; controls lost; review remained incomplete with 474 observations and 6 min; no final run artifact in recording inventory |
| `w-3` | Authored twin; Evaluate; seed 1; 10 min; as fast as possible; fan set cooling 40 at 1 min | Completed; UI Verified; 800 observations; command applied, 51 ms reported; temperature 30.0 → 26.0 Cel and fan 0 → 40% by review end |
| `w-4` | Authored twin; Explore; seed 1; requested 60 min; Watch; no disturbances/calls | Normal navigation retained banner; active workspace row said Did not finish; Stop now finalized at 28 min; UI Verified, 2,240 observations |
| `w-5` | Authored twin; Evaluate; seed 1; 30 min; Watch; temperature drift at minute 0 | Independent live fault-label disclosure reproduced; later completed, UI Verified; 2,400 observations; no outside-system score |

Verified above means the UI's file/replay check. It is not an independent assertion that the plant was healthy, a consumer passed, or hardware was controlled correctly.

## Complex factory exercise

Used the existing factory composition/domain: 23 instances in five areas and 59 links. It includes production equipment, cooling plant, air systems, meters, transformer, generator, weather and calendar dependencies. This is complex enough to expose cross-asset diagnosis and navigation problems without inventing a new plant merely for the review.

1. Open Factory utilities, expand Run plant.
2. Set seed 42, 30 minutes, Explore, Watch.
3. Add CT-1 fan failure at minute 5 and AC-1 fails to load at minute 8.
4. Add CH-1 set setpoint with `setpoint c = 9` at minute 12.
5. Start; close the panel; inspect CT-1 and CH-1 while running.
6. Reopen the run panel to check execution controls.
7. Open the completed review; select CT-1 and click the minute-5 event; inspect report and timeline.
8. Choose Set up another run and check retained inputs, particularly the missing effector call.

An independent area-only authoring exercise later added `Audit staging` to the copied factory composition and saved it. UI confirmed `Ready to run. Saved.` with composed channel/fault/effector counts. It did not change the source checkout or the earlier `w-1` inputs. Failed connected attempts preceded that edit.

## Authored twin exercise

From Home → Edit fan → Behaviour:

- Included fan, thermal_sensor and ble_sensor_node.
- Tried adding indicator; UI refused that addition with the unsupported-setting message. It was not included in the generated twin.
- Set twin name `audit-fan-thermal`, title `Audit fan thermal response`.
- Fan command: sets `level_percent`; temperature follows gradually with gain `-0.1` and time constant `30` seconds.
- Thermal sensor temperature starts at `30`; other readings retained the offered defaults.
- Created the twin and opened Check, then Open as a plant.
- Checked input validation by entering fan level `70` against the declared `0 to 60` range. UI blocked launch and focused the field. Corrected it to `40` before running.

The generated plant has three instances and one causal link. Its graph reports no port data, but the causal edge is rendered. No claim is made that the edge is missing. The fan's default instrument is a valve position transmitter; that is an observed semantic concern, not proof of a numeric error.

## Targeted reproductions

### Evaluate disclosure

Run `w-2` showed dropout on the selected asset while live. Independently, configure the authored twin in Evaluate for 30 minutes, with temperature drift at minute 0; start Watch. Keep the launch panel visible. The graph node says `Fault · temperature · drift`, while the same screen says `Evaluate`. See evidence 29–30. The later completed review is evidence 31.

This demonstrates visible pre-seal disclosure on the browser's plant view. It does not prove payload disclosure to an external system, a role bypass, or a numeric-truth leak. The current acceptance text explicitly requires no hidden fault labels on Evaluate pages; if product intent is to permit a privileged director view, that needs an explicit, visible boundary.

### Reload/interruption

Start the 30-minute Evaluate Watch run with fan level 40 at minute 1 and temperature dropout at minute 5. Close the run panel and inspect the thermal sensor. Reload around minute 6. The plant loses the running banner and offers ordinary Run/Review actions. Review reports absent `run.json` and six minutes of evidence. Revisit after the original run should have finished: the same incomplete state remains. Choose Set up another run: the form resets to Explore with no disturbances or calls.

Normal in-app navigation was tested separately in `w-4`: it preserved the running banner and returned to the active run. This difference narrows the reproduction to reload/recovery rather than all page changes. The actual process/request cancellation mechanism was not inspected.

### Failed connected attempt and status ambiguity

1. Factory → Run → Connected session.
2. Leave every instance simulated, with no allowed policy digests.
3. Enter an absent host path; Check the plan returns 23 simulated instances and correctly warns that normal commands are refused.
4. Start; record the error and retained attempt ID.
5. Home lists the attempt as a connected session. Open its review.
6. `c-6dc5acbb` is Verified at zero duration/observations and presents initialized KPI/alarm tables without leading with its launch failure.

The second path was typed as `/private/tmp/streamsim-ui-audit-20261024/sockets/absent.sock`; the rendered error shortened it to `/privatesockets/absent.sock`. The audit records the displayed error but does not diagnose that path transformation.

### Stop and lifecycle

Start a 60-minute Explore Watch run. Navigate Home while it is active: see the running banner and Did not finish row simultaneously. Return through the banner, Stop now, then inspect the finalized review and About this run. The 28-minute run is Verified without visible original 60-minute intent or operator-stop reason.

### Authoring, capture, and discovery

Inspected factory editor, CH-1's edit controls, compatible output-target filtering, area add/save, device command risk/bounds/safe stop, twin behaviour/check, board capture, new-plant form, new-device form, and keyboard help. Selecting CH-1 supply output narrowed possible link targets to compatible receivers; the attempted incompatible tower target was unavailable. This is a positive type-guidance observation, not a failed product interaction.

Capture from a board offered unsuitable simulation recordings; Read this run returned a clear no-announcement result and did not save anything. No hardware announcement was supplied or invented.

## Evidence index

Each basename below has `.txt` and `.jpg` unless a type is stated. Text is a rendered-page snapshot; images are desktop Chrome screenshots.

| Evidence | Subject |
|---|---|
| `01-factory-running` (jpg) | Factory live fault, static channel inspector, dense topology |
| `02-run-controls` | Active factory run; Stop now and progress |
| `03-factory-report` | Completed factory review, KPI/alarm tables, truth copy |
| `04-fault-inspect` | Minute-5 tower event and per-asset evidence |
| `05-repeat-drops-call` (txt) | Rerun prefill explicitly omits effector call |
| `06-connected-plan` | Technical setup, real-board mapping fields |
| `07-plan-check-absent-host` (txt) | Structural check and policy refusal warning |
| `08-session-failure` (txt) | Owner-only path refusal and retained attempt |
| `09-host-unavailable` | Absent-host launch failure |
| `10-workspace-after-failures` | Inventory and failed attempts without operational outcomes |
| `11-failed-session-verified` | Zero-time/zero-observation failed attempt marked Verified |
| `12-editor-asset` (txt) | Existing CH-1 editing controls |
| `13-editor-save` (txt) | Successful copied-factory area edit/save |
| `14-device-command` | Risk, bounds, safe-stop explanation |
| `15-twin-authoring` | Suggested instrument, multiple devices, unsupported indicator |
| `16-twin-check` | Created twin's check summary |
| `17-generated-twin-topology` | Generated plant with no port metadata |
| `18-command-validation` (txt) | Out-of-range command blocked before run |
| `19-evaluate-running` | First live Evaluate injected-fault disclosure |
| `20-reload-loses-running` | Plant after reload loses running controls |
| `21-evaluate-incomplete-after-reload` | Incomplete review after reload |
| `22-capture-no-announcement` (txt) | Clear refusal for unsuitable capture source |
| `23-reload-still-incomplete` (txt) | Later revisit still incomplete |
| `24-incomplete-rerun-resets-mode` (txt) | Retry defaults to Explore, no faults/calls |
| `25-thermal-response-end` | Completed fan intervention; 26.0 Cel |
| `26-running-labelled-did-not-finish` | Running banner and conflicting workspace outcome |
| `27-stopped-run` | Deliberately shortened run shown as ordinary Verified recording |
| `28-new-plant` (txt) | Copy/empty-template creation and restricted source selection |
| `29-evaluate-repro-setup` (txt) | Independent Evaluate drift experiment configuration |
| `30-evaluate-fault-disclosure` | Evaluate mode and injected-fault label in the same active view |
| `31-completed-evaluate-report` | Completed blind-mode recording with no outside-system score |

Input SHA-256 values recorded for reproduction:

| Input | SHA-256 |
|---|---|
| Shipped factory domain | `d147eccbfacdd520d94aea8b27a43f8415d1c402c2ab36822a24b16fbe629451` |
| Shipped factory composition | `0e1f8598bfccd3e94cc6b1e945417fa66f39f75efa51f1a519f46beabb3aba59` |
| Authored audit-fan-thermal domain | `8b0e9eaa878c66e46701c5d1feb4b0b66a21b5b8d4e08825a2bd6f26f70350ba` |

The runtime workspace is temporary and may be cleaned by the OS. Screenshots, page evidence and the reproduction instructions are retained in this folder. Simulator data, contracts and fixtures are not placed in the engineering-documentation tree.

## What remains unvalidated

- Successful live connected session, real-device discovery, command receipt/report/effect under load, human approvals, policy/lease interaction and competing controllers.
- Host/gateway disconnect and rebirth; emergency stop with stopped/skipped/unreachable targets; physical device behavior after the session ends.
- Very large campus rendering, long-history retention, high-rate telemetry, alarm floods and performance budgets.
- Full assistive-technology accessibility, full keyboard operation, dark mode, mobile, and alternate desktop sizes.
- External consumer blindness and payload-level sealed-truth enforcement. Only browser-visible disclosure was tested.
- Numerical model correctness, analytic cross-checks, replay determinism independently of the UI's verification result, and whole-repository gates.

For these areas, [the control-plane direction](CONTROL_PLANE_DIRECTION.md) gives concrete acceptance exercises rather than claiming current capability or safety.
