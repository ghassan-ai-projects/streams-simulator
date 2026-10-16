# Current-branch code and Chrome review — 2026-10-16

Reviewed `ui-score-loop` from `fab3eed831c5` against refreshed `origin/main` and the claims in [dimension-loop.md](../ui-polish-sim-authoring-20261015/dimension-loop.md). This review covers the Canvas application, its state transitions, panel layout and recent presentation changes. It is not a new whole-product 9/10 certification.

## Findings fixed

### P1 — a live review can retain an unfinished verdict and the current workspace plant

The listing-aware load guard correctly avoids missing-file and verification requests for unfinished runs. However, those skipped loads never resumed when the run ended. The graph could keep using the mutable workspace plant instead of the plant recorded with the run.

A real-bridge Chrome regression exposed a second part of the same transition: `run.closed` can arrive before `run.json` is published. The one stream-close refresh then reads an open report and an unavailable verdict and never refreshes again.

The store now resumes the deferred recorded-plant load from an ended report, verifies only after a complete report, and reacts to the workspace listing's ended-run update without reopening the event stream. An ended incomplete recording still supplies its recorded plant without requesting a verdict. A stale unfinished listing no longer overrides an authoritative completed report.

Four unit regressions cover these states. The real-binary browser regression creates a world, opens its live review, advances and ends it, then requires a successful recorded-plant request and the Verified status. It failed before the publication-race fix and passes afterwards. Existing abandoned-run browser checks also pass.

### P2 — overflow cue fails after native panel disclosures expand

`useCutBottom` watched scrolling and the outer element's size. Opening Faults changes its content height without resizing the fixed-height inspector or rerendering its parent. In Chrome, expansion changed `scrollHeight` from 856 to 1162 while `clientHeight` stayed 856, but the fade remained absent.

The hook now observes descendant content changes and native `open` attribute changes, with cleanup. It ignores its own `data-cut` attribute to avoid a feedback loop. Unit regressions cover disclosure and child-content changes; the Chrome regression expands and collapses Faults without scrolling or resizing and checks both the geometry and painted mask.

## Claim checks

| Claim | Evidence and result |
| --- | --- |
| Clear source-to-receiver links | View inspector shows `pmp-u1 flow out → ch-u1 flow in` and `ch-u1 heat out → ct-u1 heat in`; buttons identify the source or receiver they select. Confirmed in Chrome. |
| Parent areas explain children | Edit outline displays `Halls 0 · 15 below`, with an accessible explanation of the count. Confirmed. |
| Real duration in About | A new one-minute Explore run shows `1 min of plant time from 00:00:00 UTC`, with seed, adapter and simulator version. Confirmed. |
| Integrity sentence avoids repeating Verified | The header has the Verified badge; Report says “The files match and the replay reproduces.” Confirmed. This proves run integrity, not an outside system's success. |
| Keyboard dialog heading focus | Confirmed manually: the dialog heading receives focus and paints a 2px solid blue outline. Also covered by the existing real-Chrome keyboard regression; keep this claim scoped to that dialog, not every page heading. |
| Panel overflow hint | Initially disproved for native disclosures; fixed and checked in Chrome. |
| Unfinished reviews avoid doomed loads | Existing real-bridge checks pass; the completion transition required the P1 fix above. |

Manual Chrome review used an isolated copy of the workspace and a fresh build of this branch at `127.0.0.1:61312`. View, Edit, Run setup and completed Review were inspected in light mode. Desktop samples were 1280×800, 1440×900 and 1920×1080 across those flows; this is not a full cross-product breakpoint matrix. The 1280px editor and 1440px plant had document width equal to viewport width. A new campus run completed and verified. The review tab's captured warning/error log was empty at the inspection checkpoint. Existing editor domain-probe issue CB-41 remains documented separately; this report does not claim all editor requests are error-free.

## Remaining design limits

- Large-campus graphs still contain many outlined cards, small text and substantial unused space. Fitting the outer page to the viewport does not establish good use of its internal space. The broad responsiveness/comprehension concern is still open.
- Compact equipment labels and technical channel abbreviations still require domain knowledge. Better defaults and explanations need separate product work and novice task testing.
- The branch's earlier all-dimensions-9 result was a scoped review at one desktop viewport. It cannot establish whole-product, beginner, mobile or every-state quality.
- This round did not explore dark mode, zoom behavior or validation forms, as requested. The unchanged mandatory automated suite may exercise those paths.

## Evidence

- [View: expanded inspector with overflow cue](view-expanded-inspector.jpg)
- [Edit: campus layout and nested-area count](edit-campus.jpg)
- [Review: new completed run, About and integrity explanation](review-campus.jpg)

## Verification

- Both fixes were committed separately: `65331fc` (overflow cue) and `a9dc4ba` (live review completion).
- The first full run passed code CI, all 445 Canvas tests, all 70 real-Chrome tests, dependency/lint/duplication gates and Enola's structural check. The dependency audit reported zero vulnerabilities. Canvas statement coverage was 87.37%; the changed review store was 98.42% and overflow hook 95%.
- Its final behavior-pin step failed under the host Go 1.27.2. An untouched export of starting commit `fab3eed831c5` produced byte-identical output to the reviewed tree: the mismatch predates these fixes.
- The pin includes compiler/platform metadata. CI pins Go 1.26.9 in `.github/workflows/ci.yml`; the existing [P9 environment findings](../reviews/P9-gate-environment.md) document this cause. The compiler-dependent artifact hashes must be checked with CI's toolchain; they are not refreshed to host-specific values.
- Reproducible command: `GOTOOLCHAIN=go1.26.9 PATH="$PATH:$(go env GOPATH)/bin" make phase-gate`. Append the local release-tool directory, rather than prepending it and shadowing the compatible Homebrew linter.
- One additional stale pin was identified: `failure/unknown-command/stderr`. Commits `31693dc` and `92cc459` intentionally added `component examples|example` and `component new` to usage text after the previous pin. Source comparison against `bf7ebc5` confirms those two help lines; the command still exits 2 and its JSON output hash is unchanged. The user explicitly requested fixing the mismatch. Only this help-output hash was updated, from `2628313…` to `b99cc0b…`; no simulation-output pin or gate was relaxed.
- Final pinned-toolchain gate: **PASS**. `make phase-gate` completed with code CI, 96 test files / 445 Canvas unit and component tests, 70 real-Chrome tests, Enola no structural regression, and a matching behavior pin. The help-pin correction is committed as `0b6c538`. No compiler-dependent simulation hashes were changed.
- Existing build warnings remain: the PostCSS plugin omits `from`, and the primary JavaScript chunk exceeds Vite's 500 kB advisory threshold (207.52 kB gzip). These warnings do not establish a performance regression, but bundle splitting remains separate follow-up work.
