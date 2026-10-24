# Canvas test time: baseline and after

Measured on a 10-core Mac whose load average was 20 to 55 the whole time (a Go test run was using about 6 cores
alongside), so every absolute number is inflated and noisy. Compare columns, not seconds. Playwright used its default
5 workers locally; the CI runner has 2 cores and so 2 workers.

## Per step of `npm run check` (local)

| Step | Baseline | After | What changed |
|---|---|---|---|
| format:check | 7-10 s | 6 s | - |
| typecheck | 0-1 s | 1 s | - |
| lint | 70 s | 22-27 s | `eslint --concurrency auto` (same rules; the type-aware program was built once, serially) |
| deps / architecture / dup / knip | 1 / 4 / 1 / 1 s | same | - |
| vitest with coverage | 63 s (88 s on a busier minute), jsdom created 322 times, 181-203 s | 28-29 s | `pool: "vmThreads"` |
| build + css:emitted | 2 s | 1-2 s | - |
| e2e | 212 s (sum of test time 936 s) | 173-177 s (sum 811 s) | see below |
| Whole check | about 365 s | 239, 245, 246 s (three runs in a row, all green, 160/160 e2e, no flake) | |

CI (real run 40128196816, 2 workers): eslint 94 s, vitest 257 s (jsdom 287 s tracked), e2e 9.7 min, job 17.5 min.
Expect the lint and vitest changes to give about 1 and 2 minutes there. The e2e gain is smaller (see "What is left").

## Vitest

- 321 files, 2634 tests. 25 slowest files before (under load): PlantWatch 25 s, TrendsPlant 19.5 s, PlantScreen 17 s,
  ReviewActions 16 s, ReviewTimeline 15 s, DeviceForms 14 s, BehaviourSteps 13 s, BehaviourKeys 11 s, WorkspaceScreen 9 s,
  EditorPanel 9 s, LaunchPanel 8 s ... All of it was jsdom creation per file (42% of tracked time) plus React render and
  axe; no single test above 5 s. After `vmThreads` the slowest file is 9 s and the slowest test 1.3 s.
- `pool: "vmThreads"` keeps a fresh context per file. Green in 4 runs, 3 of them with `--sequence.shuffle`.
- `isolate: false` was faster (16 s) but `viewportFit.test.tsx` (its `vi.mock("@xyflow/react")`) leaks across files, so it
  was not taken.
- Coverage instrumentation costs about 13 s of the 28 s (15 s without). v8 stays; no threshold or path touched.
- `BehaviourKeys` timed out at the 5 s default once under heavy load before the change; it no longer comes near it.

## Playwright

Slowest tests before: interrupts 55.8 s, interrupts-layout 44.8 s, launch Evaluate-sealed 38.7 s, job-run
"ends by itself" 34.7 s, twin-report 34.6 s, launch planned-fault 26.9 s, policies 23.6 s, review hall/recorded 18 s each.
Dominant causes: real-time script delays of the plant wire host (40 to 50 s after the host was spawned), Watch-paced runs
waited to their natural end (30 s), then bridge boot plus a 395 MB workspace copy per test (about 1.5 s per test).

Changes (all keep their assertions):

| Test | Change | Why still equivalent |
|---|---|---|
| host-scripted tests (interrupts, interrupts-layout, twin-report, policies) | The host is now spawned by `bench.start()`, called at the end of `planOneRealBoard`, so the script clock starts when the plan is filled instead of when the bench was opened. Delays: interrupts 40/50 s -> 25/30 s, twin-report 20/30 -> 10/14, policies 20 -> 10 | The command is still due only after the device is announced, and for the approval tests after `Approve each command` is set; the margin is measured from a closer anchor. |
| launch "planned fault appears ... stays to the end" | After the fault is seen on its node at minute 5, `finishNow` picks "As fast as possible" in the Speed select; waits for the review | Pacing is not what it asserts; the fault and the final review still are. |
| launch "Evaluate ... until it is sealed" | Same `finishNow` after the unsealed assertions | Every unsealed assertion runs at the normal pace as before. |
| job-run "left running across a reload still ends by itself" | After the reload and reopening the panel, `finishNow` | The run is still left running across the reload and ends without Stop. |
| launch "top-bar run indicator ... Stop now keeps the shorter run" | Plan 6 h -> 1 h (unit still hours; "of 01:00:00") | The binary's pre-start dry-run check grows with the plan (measured locally: 20 min 0.22 s, 60 min 0.37 s, 6 h 1.5 s, 24 h 5.7 s; CI is 3 to 4 times slower, which is why the 6 h plan overran the 5 s expect on CI). 1 h gives a 30 s window at 120 s per s, shows "N s per s" and exercises the hours unit; nothing in it needs more than 1 h. |
| layout-stability "notices of w-edited" | The first count wait gets `timeout: 60_000` like the other waits for the binary's verification in the review | It waits for the binary's replay, not a render. |

Genuinely real-time tests kept as they were: "holds the mark for 3 s of wall-clock time" (launch), the pause/step/resume
polling tests (job-controls), the 500 ms poll based ones, and `Watch` plan tests that wait for the run clock to reach a
minute at 60 or 120 s per s (3 to 6 s each). Other specs use plans of 60 min or less (the Watch default).

No test was deleted or merged. No fixed sleeps added. Workers were not raised: at 8 workers on this machine 5 tests
timed out on load, and CI has 2 cores.

Unchanged config: `playwright.config.ts`, coverage thresholds, architecture register, scripts.

## What is left (proposals; nothing here is implemented)

Local e2e is still about 175 s, so the whole check is about 245 s, above 2 minutes. On the 2-core CI runner, e2e sum of
test time (about 810 s) over 2 workers is still about 7 minutes.

1. Split e2e into its own CI job (parallel with lint/vitest) and shard it (`--shard=1/2`, `2/2`): wall time of the job
   becomes max(e2e shard, the rest). Needs a workflow edit.
2. Per-test workspace copy: the template workspace is 395 MB (11 runs of 34 MB, `w-hall` 98 MB) and is copied for every
   test. Options: record shorter fixture runs (needs the assertions on "30 min" and the playback times changed), or
   hardlink the recorded runs if the binary is shown never to write inside an existing run folder. Probably the largest
   remaining per-test cost on a Linux runner without reflinks.
3. The host-scripted tests still wait 25 s (interrupts) and 10 to 14 s by design, because the plant wire host issues a
   scripted command at an absolute wall-clock time and drops it if the device is not yet announced. A host script step
   that fires "after the first announced device" or "on request" (a Go change in `plantwire host`) would remove the wait.
4. Split into smoke (PR) and full (nightly) e2e; this lowers coverage per PR and so is the owner's call.
5. `eslint --cache` was not used: its cache is keyed on file content, which can miss a type change in another file for
   type-aware rules.
