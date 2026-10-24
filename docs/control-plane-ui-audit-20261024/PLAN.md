# Remediation plan

Turns the audit ([findings](FINDINGS.md), [direction](CONTROL_PLANE_DIRECTION.md), [round 2](round-2/README.md)) into phases with quality bars. Revision 2: the P1 findings were traced to code (references below, checked at `3c6d78c8`), which changes the plan substantially.

## What the code says

The audit was a browser review. Reading the code shows that most of the trust findings share one cause.

**The browser runs the experiment.** `canvas/src/app/launch/internal/launchStore.ts:62-80` creates a world, then loops `inject` / `advance` / `invoke` over HTTP with a pacing delay, then calls `endRun`. The plan, the cursor, the stop request and the outcomes live only in browser memory. The binary already has the same use case: `scenario-v0.1` is an ordered command list (`clock.advance`, `fault.inject`, `fault.clear`, `perturb.apply`, `effector.invoke` with arguments) and `run` executes and exports it (`run.RunScenario`, `run.ExportScenarioFile`). The canvas re-implements a scenario runner. That breaks AGENTS.md rule 2 (one use case, one implementation) and the canvas rule that orchestration comes from `streamsim`.

| Finding | Root cause | Side |
|---|---|---|
| F02 reload loses the run | Loop dies with the page; `run.json` is only written on `run.end`; the world stays alive in the director with nobody advancing it | Canvas orchestration |
| F04 running shown as Did not finish | `RunList.tsx:37-43` shows Running only for connected sessions; the bridge knows `runIsRunning` (`canvasbridge/internal/app/director.go:85`) but emits it only for sessions | Both |
| F05 stop loses intent | Requested duration and stop reason never leave the browser; `RunArtifact` has no field for them | Both |
| F10 rerun drops calls | `reportDecoder.ts:105-116` drops `args`, then `reportViews.ts:109` hard-codes `calls: []`; the Go report carries the args | Canvas |
| F11 incomplete rerun resets | Prefill is rebuilt from a report that has no seed and sealed truth for an open run | Canvas reconstruction |
| F12, F13 no saved plan, no clears or delivery faults | The canvas plan model is narrower than `scenario-v0.1`, which already has clears and perturbations | Canvas model |
| F01 fault label in Evaluate | `showing.ts:12-20` → `planMarks.ts` → `Readout.tsx:46` and the accessible name in `InstanceHeader.tsx:35`, built from the user's own plan; mode is never checked. The bridge does not leak truth, but the page breaks `CANVAS_UI_ACCEPTANCE.md` G7/A4 | Canvas |
| F03 failed launch shows Verified | `plantwire/internal/app/connect_run.go:131-137` calls `r.End` even when `serveSession` failed, so the run is written as complete; the failure is held only in bridge memory. The canvas `badgeOf` (`model/review/internal/report.ts:121-128`) says Verified without checking for observations | Go plantwire, plus canvas |
| `/privatesockets/absent.sock` | `canvasbridge/internal/domain/scrub.go:9` does a substring `ReplaceAll` with no path boundary, and the root is not symlink-resolved (`/tmp` → `/private/tmp`). It mangles messages and only partly redacts the workspace path | Go canvasbridge |

**So the strongest improvement is structural:** the canvas builds a `scenario-v0.1`, and the binary runs it as a job it owns. That single change fixes F02, F04, F05, F10, F11, F12 and most of F13. It also gives F08 (pause, rate, live commands) a natural home, because they become job controls whose effects already land in the command log.

## Re-rating after the trace

| Item | Before | After | Why it changed |
|---|---|---|---|
| F01 | Very high | Very high, about 1 day | Pure canvas guard plus test |
| F03 | Very high | Very high, about 1 day | Two lines in plantwire plus an outcome check |
| Scrub bug | Unrated | High, under 1 day | Mangles errors and only partly redacts the workspace path |
| F02, F04, F05, F10–F13 | Separate fixes | One structural change | Same cause |
| F08 pause, step, live commands | Medium, expensive | Medium, cheap once the job exists | Job controls rather than a new subsystem |
| Saved plans, Repeat exact | High | High, nearly free | A saved plan is a scenario file; Repeat exact is `ExportScenarioFile` |
| Round 2 Plan view | Medium | Medium–high | Becomes an editor for a contract that already exists, not a new model |
| Operations console (OP02, OP04, F09 full, R19, R20) | Low | Low | Unchanged until the product direction is decided |

## Quality bars for every phase

| # | Bar | Check |
|---|---|---|
| Q1 | Each fixed finding has a test that fails on `3c6d78c8` and passes after; test name cites the finding id | Test names |
| Q2 | Rules and orchestration in the owning Go module; surfaces render. No new canvas code computes status, verification or prefill. | `test/architecture`; reviewer reads the canvas diff |
| Q3 | No gate loosened | `git diff main -- test/architecture scripts` justified line by line; `make gate-ratchet` |
| Q4 | Go bar: files ≤ 300 lines, functions ≤ 15, packages documented, coverage floors not lowered, no new `//nolint` | `make quick` |
| Q5 | Canvas bar: `make canvas-check`, axe, e2e against the real binary, no sleeps, no mocked binary | `make canvas-check` |
| Q6 | Determinism: replay reproduces; behaviour-pin diff empty or every change named | `scripts/behaviour-pin` |
| Q7 | Truth boundary: the A4 canary string is checked in rendered DOM text and accessible names as well as payloads | E2E canary |
| Q8 | Chrome design review per `CANVAS_UI_ACCEPTANCE.md` §3 | Phase review |
| Q9 | `make quick` per commit; `make phase-gate` once per PR | CI |
| Q10 | Phase review lists findings closed, deferred with reason, gate changes, and what was not verified | Phase review |

## Phase 1 · Quick trust fixes (days)

These do not depend on the structural change and close two P1s and a redaction bug.

1. **F01 Evaluate guard.** Plan-derived fault marks are shown only in Explore. In Evaluate the node shows observation quality only. Extend the A4 canary to the rendered page: put the canary in a fault name and assert it is absent from `document.body.innerText` and every accessible name until sealed.
2. **F03 failed start.** In `serveNode`, abandon the run with the session error when the session fails before the barrier, so `run.json` records `incomplete` and the error. The report leads with the failure. Remove `badgeOf`'s local Verified rule and use an outcome supplied by the report (see Phase 2).
3. **Scrub.** Replace whole path prefixes only, at a separator boundary, and scrub both the given root and its `EvalSymlinks` form. Table test: `/tmp` and `/private/tmp` roots, a root that is a prefix of a sibling, and the root at the end of a string.
4. **F20 copy.** One sentence per mode × lifecycle, from the G7 rule. No new behaviour.

**Bar.** Q1–Q10. The F01 canary fails today in the DOM and passes after. A failed connected launch reproduced as in the review log opens as Failed to start. The scrub test reproduces `/privatesockets` first.

## Phase 2 · The binary runs the experiment

**Findings.** F02, F04, F05, F10, F11, F12, F13 (clears and perturbations), plus the minimum of F09: Home lists active, interrupted and failed runs by status.

**Design first (AGENTS.md rule 4).** Write a short design note before any code. It answers:

- Owner. Proposed: `run` executes a scenario with an optional pacer (the wall clock sits in an edge); `director` owns the job registry (start, status, stop, list), because it already owns the world registry and run lifecycle; `canvasbridge` exposes those calls and holds no rule.
- Lifecycle: starting, running, stopping, completed, stopped by operator, interrupted, failed to start. Restart semantics: on bridge start, every open world without a job is closed as interrupted, with a reason.
- Record: the run artifact gains the scenario's planned end and a termination reason and time. Pre-release, so the schema changes freely.
- Outcome: the report emits separate execution, integrity, coverage and assessment fields. The canvas renders them and stops deriving Verified.

**Canvas changes.** Launch builds `scenario-v0.1` and submits it. After any reload the canvas asks the bridge for the active job. "Set up another run" loads the exported scenario of the source run, with the full command list, rather than rebuilding a plan from the report. A saved plan is a scenario file in the workspace.

**Bar.** Q1–Q10, plus:

- **One implementation:** the same scenario run from the canvas and from `streamsim` CLI produces the same run digest. This is the test that proves rule 2.
- Reload e2e: start a 30-minute Watch run, reload, regain identity, progress and Stop; let it finish; Review is complete.
- Stop e2e: a 60-minute run stopped at 28 shows 60 planned, 28 actual, Stopped by operator.
- Repeat exact of the factory run `w-1` reproduces its digest, including the CH-1 setpoint call.
- Bridge restart mid-run: the run shows Interrupted with a reason, never Running and never Verified.
- The browser no longer issues `fault.inject`, `clock.advance` or `effector.invoke` for a launched run (network assertion in e2e).
- `launchStore`'s loop, `planOfRun`, `leftOut` and `badgeOf` are deleted, not kept beside the new path.

## Phase 3 · See and steer a running plant

**Findings.** F06, F07, F08, F18, F28. Round 2: R05, R07, R08, R13, R15.

**Scope.**

- Inspector shows value, unit, quality, observation time, age and a mini-trend, from the observer stream the bridge already allows (`journal.jsonl`, `trace.jsonl`).
- Watch list stored in the scenario's description or a sidecar the binary owns; synchronized trends with gaps drawn as gaps.
- Absence explained: never instrumented, awaiting first reading, stale, bad quality.
- Job controls from Phase 2: pause, resume, rate, and an immediate command or fault appended to the job. Each lands in the command log.
- Unit symbols from a binary-supplied table, with the code in details.

**Bar.** Q1–Q10, plus:

- A run with pauses, rate changes and live commands replays to the same digest from `run.json` alone, and the analytic cross-check holds across a pause.
- Missing data is never drawn as zero and never interpolated across a loss (component test with a dropout fixture).
- Evaluate canary passes on every new surface, including watch-list labels and trend tooltips.
- Acceptance exercises A (cooling cascade) and E (sensor versus process failure) pass in Chrome without the injected truth.
- A performance budget for live trends on the 23-instance factory, measured before the phase starts and met at its end.

## Phase 4 · Plan, review and compare

**Findings.** F19, F21. Round 2: R01–R04, R09, R11, R14, R16–R18.

**Scope.** Chronological scenario editor with lanes (world, measurement, delivery, commands) over `scenario-v0.1`; name and purpose in `description`; readiness including scorer readiness or an explicit "unscored investigation"; question-led Review using the Phase 2 outcome fields; action story for each command; reference-run comparison aligned by event, computed by the binary.

**Bar.** Q1–Q10, plus:

- The editor round-trips any valid scenario: load, save, byte-equal canonical JSON. A scenario it cannot express is shown read-only, not silently reduced.
- Plan validation comes from the binary's scenario check; the canvas holds no duplicate rule.
- Comparison never ranks runs from one aggregate or one seed; interval statistics carry their interval.
- Exercise: build the nominal, disturbance and response references of round 2 as three named scenarios and compare a post-event window.

## Phase 5 · Connected operation (gated on a host)

**Findings.** F14, F15, F16, F27. Capabilities OP02, OP03, OP05.

Start only with a real or representative plant-wire host; the fan experiment is the first target. Before that, building approvals or emergency-stop UI would guess at a protocol.

**Bar.** Q1–Q10, plus acceptance exercises B (second browser, reconnect), C and D against a real host. Readiness lists each check as performed or unknown. Approval is bound to request, target and scope, a changed precondition invalidates it, and dispatch is at most once by the owning system's evidence. Receipt, application and observed effect remain separate facts.

## Phase 6 · Engineering and provenance

**Findings.** F17, F22–F26. Capability OP08. Round 2: R10.

**Bar.** Q1–Q10, plus: conformance, plausibility (`domain lint`), calibration and hardware validation shown as separate claims; converting a read-only domain never drops semantics; a thermal-gain change yields a new model version, a variant run, and both recordings linked to their exact assumptions.

## Held back

- Operations console (OP04 alarms, full F09 landing page, R19 notes, R20 handover, shift exercises): decide first whether the canvas is a test bench or a control plane.
- Dark mode, mobile, other sizes: desktop-only canvas rule.

## Order

```text
P1 quick fixes ──► P2 binary runs the experiment ──► P3 see and steer ──► P4 plan and compare
                                                  └─► P5 connected (when a host exists)
P6 after P4
```

## Open owner decisions

1. Phase 2 ownership: job registry in `director` (proposed) or elsewhere.
2. F01: Evaluate is observer-only (proposed), or an explicit, labelled director view is also wanted.
3. Test bench or control plane: decides whether the held-back items return.
