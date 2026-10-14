# P9 continuation handoff — 2026-10-14

This is the current handoff for Claude. Read `AGENTS.md` first, then this file, the seven rounds and every
disposition in [ui.md](ui.md), [P9-functional-plan.md](P9-functional-plan.md), and `governance/FINDINGS.md`.

## Git and delivery state

- Continue **`codex/canvas-p9-playback`** in the root checkout. PR #32 was merged; do not push to its old branch.
- Started from merged PR #32 (`6721cdf`). Completed commits: `53aefb8` (round 6 integrity warnings and tests),
  `bf7ebc5` (simulator-owned alarm activation truth and canvas navigation).
- Owner authorized merging latest main. `856bcb9` cleanly merges `origin/main` at `02fe45a` (PR #33 test-speed).
  Source commit `762c7fa` carries its own `Owner-approved-gate-change:` line. No new gate adjustment here.
- The handoff/progress commit follows that merge. No push or new PR has occurred. Owner asks to commit and hand off;
  **push and open PRs only when explicitly asked**.
- Existing untracked owner files are untouched: `docs/product-strategy-20261011/`,
  `docs/simulation-canvas/mockups/data.json`, `docs/simulation-ideas/`. Never stage them opportunistically.

## Owner directions now in force

Finish functionality before optimizing visual details. Implement directly; do not wait for or request Sonnet.
Use three independent reviewer lenses each round: design director, desktop keyboard/input, process/test engineering.
The owner requested Luna reviewers; the last round used Luna. Chrome always, desktop 1280–1920, light and dark.
No separate accessibility audit is requested; existing automated checks remain intact.

Persist every finding and disposition in this folder; phase findings also belong in `governance/FINDINGS.md`.
The nine non-negotiables and architecture/quality bars remain binding. The canvas computes no simulator-owned truth,
validation, composition, score or verification. Missing capabilities go into the binary first. Never loosen a gate.
Real-binary Playwright acceptance, no binary mocks and no fixed sleeps. No new code comments beyond package docs.
No customer data/configuration in GitHub. Each reviewer owns and stops only its own bridge PID.

## Completed and evidenced

1. Round 6 frozen review: design 4/5, desktop input 3/5, engineering 2/5; **P9 not approved**.
   All checked incomplete runs never read Reproduces. An additional High was found: rejected incomplete truth lacked
   a caution. Fixed in full Report and KPI-only Inspect, including verify propagation through RunPanel.
   Component regressions and real-binary Chrome tests cover the incomplete stamp and missing ledger digest.
2. R9-24 done: simulator **0.3.2** publishes the existing alarm scan's `activation_times_ns` as an owned copy in
   sealed KPI truth and report. Canvas shows alarm chips and Jump to alarm, seeks, selects the source, opens Inspect,
   and wraps after the final activation. No alarm dynamics/counts/RNG changes. Initial standing alarms are not activations.
3. Round 7 is **scoped to alarm navigation and integrity propagation**, not all P9: design 5/5, desktop input 5/5,
   engineer 4/5, no High within the batch. Independent Chrome checks covered all three widths and both themes.
4. Before merging main, full `make phase-gate` passed, including then-current race and coverage tiers, strict lint,
   vulnerability/deadcode/fuzz checks, canvas gates, Enola enforcement and behavior pin. Canvas: **130 unit/component
   tests and 18 real-binary Chrome Playwright tests**. See [P9-alarm-evidence.md](P9-alarm-evidence.md).
5. Intentional pin differences from PR32 base: 37/201 entries (18 run artifacts, 18 observer journal identities,
   one manifest file). Simulator version/compiler metadata explain them; shipped fixture trace, ledger, history,
   sealed journal and replay/verify verdict entries are unchanged. Alarm-specific truth changes have dedicated tests.
6. After main merge, `make quick` passed. Post-merge phase-gate outcome is recorded below before the handoff commit.

## Ordered next steps

1. **Streaming report memory**: design only, no implementation started. Read [P9-report-memory.md](P9-report-memory.md).
   Whole-file evidence buffers, repeated full journal decoding and retained ledger rows drive report memory/time.
   Use existing kernel/journal/report layers, incremental digests and bounded retained report records, not new packages.
   Preserve every byte digest, committed-record validation, torn-tail behavior, identity/truth gate and error precedence.
   Test first and compare exact JSON/Markdown before/after; measure peak RSS and wall time on the 26-hall campus.
   Review every kernel/journal facade diff. This addresses only report portion of R9-14, not verify or all R2-01 inputs.
2. **Bounded simulator-owned playback and true-state queries**: R9-13, R9-04, R9-05 and remaining R9-14. Declare the
   protocol/design first, implement once in the owning binary module, expose through surfaces, then wire canvas.
   Refuse Evaluate truth until sealed, honor Explore. Binary owns channel/state/gain mapping; no adapter-specific
   consumer knowledge. Absolute int64 ns must cross JS exactly (decimal strings or safe relative cursors): R9-37.
3. **Bounded verify/replay evidence**: replay currently retains trace, ledger, state history and journals, then
   marshals complete arrays to hash. Verification must stream/hash without retaining unnecessary replay evidence;
   grading still needs its evidence. Preserve exact digests, command order, cancellation and first-divergence outcomes.
   Measure browser responsiveness and campus report/verify memory/time after both binary and playback changes.
4. **R9-22 negative air flow**: specify measurement bounds semantics first; do not hide faults, alter RNG order or
   substitute canvas clamping. Prove oracle, analytic cross-check and replay determinism with meaningful regressions.
5. Remaining P9 functional findings: R9-01 world-id collisions; R9-03 typed composition ports; R9-23 absent older
   fault truth represented as unavailable; R9-25 declared groups; R9-26 session-record evidence boundary disposition;
   R9-34 disputed hall Fit clipping (same hall/width geometric reproduction); R9-35 direct graph keyboard route;
   R9-36 strict adapter admission for report/replay responses; R9-37 exact cursor (above).
   Check all findings assigned to P9, including earlier-phase entries. Do not mark Done while any is open.
6. Fresh frozen **full P9** Chrome review with three reviewers, previous findings and every disposition, until each
   scores at least 4 and no High. Visual optimization follows functionality. D6-01 compact kind ellipsis, D5-12 hall
   summary figure and wires across frame titles are still deferred details; their dispositions need final review.
7. Close P9 with an evidence file, review record and CANVAS_PLAN Done row only when the findings gate permits it.
   P8b stays Built while R0-15, R0-16 and R2-01 are open. Then P10 live control, P11 builder, A1–A7 complex simulations
   built in the interface. No P10/P11 completion is claimed by this handoff.

## Worktrees and local review evidence

Root owns the continuation branch. An unused worktree exists at `/private/tmp/streamsim-p9-report-memory`, branch
`codex/p9-report-memory`, still at `6721cdf`, containing only an untracked design note now copied into this branch.
There are **no implementation edits there**. Align it with the continuation branch before using it; do not cherry-pick
from it as if work were completed. Other `.claude/worktrees/` checkouts belong to other work; leave them untouched.

Frozen round 6: `/private/tmp/streamsim-p9-r6-wu1LiP/manifest.json` (binary 0.3.1, UI dist, synthetic workspace).
It includes factory Evaluate/Explore/edited evidence, campus-26 (725,253 observations), and an incomplete-stamp case.
This is historical evidence only; its runs do not verify the 0.3.2 artifact contract. Round 6 campus report took 55 s,
verify 77 s and a browser theme action timed out; the older larger campus report reached approximately 8 GB.

Frozen round 7: `/private/tmp/streamsim-p9-alarm-review/manifest.json`, binary and `dist-final`, synthetic `w-alarm`.
Activation at 140 s, observed HEAT-1 85 °C. Hashes and width/theme observations are in the three alarm reviews.
Parent bridge PID 41151 has been stopped for handoff; reviewers already stopped theirs. Restart your own bridge
with a fresh pairing session. Pairing secrets are intentionally omitted. Scratch files are local and may disappear.

## Fresh full-review workspace recipe

Build current binary and UI, copy both to a new scratch directory, record commit/tree state and SHA256 hashes.
Compose `examples/compositions/factory-utilities.composition.json` and a 26-hall campus based on
`canvas/e2e/fixtures/campus-3.composition.json` into its private `domains/` directory, then record with the frozen binary:

- Factory Evaluate: seed 11, duration 28800, `AC-1=compressor_lead__oil_cooler_fouling@21600`.
- Factory Explore: same plus `CT-1=cooling_tower__fan_failure@10800`, with `-explore`.
- Edited factory: seed 5, duration 7200; append `{}` to trace afterwards; must be Not verified.
- Campus-26: seed 3, duration 3600, Explore; faults on h07/M-1, h07/M-2 and h19/M-1.
- Edit factory domain description afterwards; earlier run shows plant-file-changed information, while verification
  uses recorded specs. Include incomplete stamp and missing digest cases from `canvas/e2e/review-integrity.spec.ts`.

Name workspace run folders consistently with the recorded world ids; existing CLI collision R9-01 is still open.
The old recording helper `/private/tmp/streamsim-p9-freeze.mjs` is a local recipe aid, not a committed fixture.
Do not use old runs to assert new-build verification. No production code/fixtures belong in docs.

## Toolchain and gates

Use CI's patched Go 1.26.9; default local Go 1.27.2 cannot be consumed by the installed lint export reader.
Go 1.27.1 fails the vulnerability gate. Keep pinned golangci-lint 2.14.0 first on PATH: append Go's bin directory
rather than prepend it (an older linter also exists there). Keep the existing G703 pragma in bridge routes;
it is required by the pinned linter and was restored after a diagnostic trial.

```sh
GOTOOLCHAIN=go1.26.9 PATH="$PATH:$HOME/go/bin" make quick
GOTOOLCHAIN=go1.26.9 PATH="$PATH:$HOME/go/bin" make phase-gate
cd canvas && STREAMSIM_BIN=$(cd .. && pwd)/bin/streams-simulator npm run check
```

Main's owner-approved speed change runs one short simdet coverage tier. Long simulation/component sweeps and race
are on demand with `make test-long` and `make test-race`; run them when touching dynamics, components, goroutines
or locks. Coverage floors remain unchanged. A successful new phase gate does not imply those on-demand tiers ran.
[Main integration review](P9-main-integration.md) records the narrower required clock/horizon coverage;
post-merge quick exercised untagged tests. No additional gate modification was made.
Avoid concurrent Playwright and Go architecture file walks: Playwright cleans test-results during scans.

Root `.enola/` has a pre-change baseline from `6721cdf`; do not re-pin to hide findings. The unused report worktree
has its own old baseline; establish the correct pre-edit baseline after aligning its source, before actual work.
CLI Enola 0.4.29 is the enforced path; the older MCP server snapshot is stale. Do not claim report-only exit zero
as enforced architecture evidence. See [P9-gate-environment.md](P9-gate-environment.md).

## Handoff verification

Post-merge quick: **passed**. Post-merge full phase gate: **passed** (exit 0, Behavior pin matches, Phase gate passed).
This includes 130 unit/component tests and 18 real-binary Chrome Playwright tests under the new approved Go test tier.
`git diff --check` passed; `git diff origin/main -- test/architecture scripts` is empty.
Handoff documents were checked separately after the full gate; no further production changes.
Logs: `/private/tmp/streamsim-p9-handoff-quick.log` and `/private/tmp/streamsim-p9-handoff-phase-gate.log`.
