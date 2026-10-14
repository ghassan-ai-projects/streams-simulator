# Goal — production-ready canvas UI

- **Date set**: 2026-10-14, by the owner: "create a goal to finish it, and make the UI production ready with all
  functionalities, use Sonnet sub agents for review and UI review".
- **Builds on**: [GOAL.md](GOAL.md) (the simulation-canvas goal), [CANVAS_PLAN.md](../../governance/CANVAS_PLAN.md)
  (phases), [CANVAS_UI_ACCEPTANCE.md](../../governance/CANVAS_UI_ACCEPTANCE.md) (acceptance tests, review checklist,
  budgets), [P9-handoff.md](reviews/P9-handoff.md) (current state) and [FINDINGS.md](../../governance/FINDINGS.md).
- **Branch**: `codex/canvas-p9-playback` until P9 closes; one branch per phase after that, from the latest `main`.

## Outcome

A person can build a plant, run it live, disturb it and review it entirely in the canvas, and every view is fast,
legible and truthful on a desktop (1280–1920 px, dark and light). Everything the canvas shows comes from the binary.

## Done means

Each item is proven by a test or a recorded review, not by a claim.

1. **P9 playback and review closed.** Every finding assigned to P9 is done or carries an owner-approved disposition;
   the CANVAS_PLAN row reads Done with an evidence file and review record. This includes:
   - bounded memory and time for `report`, `verify` and `replay` on the 26-hall campus (R9-14, R2-01);
   - a simulator-owned windowed read of observations and true state, so playback seeks without streaming the whole
     journal and Explore runs show true state next to reported values after sealing (R9-04, R9-05, R9-13, R9-37);
   - negative air flow fixed at its cause in the simulator (R9-22);
   - the remaining P9 items: R9-01, R9-03, R9-23, R9-25, R9-26, R9-34, R9-35, R9-36, D5-12, D6-01, frame titles.
2. **P10 live run and control.** Director-driven runs from the canvas: start, play, pause, step, speed, reset; live
   view with values, quality and wire traffic; fault, environment and perturbation injection mid-run; Explore and
   Evaluate; export of a session with `scenario from-run`; the Evaluate canary through bridge and page.
3. **P11 builder.** A world is built, validated and previewed in the canvas: overview, kinds, state and behaviour
   (P11a); sensors, actions and faults (P11b); scenarios, preview, templates, diff before save and the layout sidecar
   (P11c). The saved files equal hand-authored JSON after `streamsim domain format`.
4. **Acceptance A1–A7 green.** All seven Playwright acceptance tests in CANVAS_UI_ACCEPTANCE §2 run against the real
   binary on real files and pass; A6 meets the §4 budgets on the 2,032-tag campus.
5. **Approved by three Sonnet reviewers.** The final round over the whole UI scores at least 4 from each reviewer
   with no High finding (protocol below).
6. **Gates green.** `make phase-gate` and the canvas gate pass; every pipeline stays under two minutes; long
   simulations and the race detector run on demand (`make test-long`, `make test-race`) whenever a change touches
   dynamics, components, goroutines or locks.

## Order of work

1. P9: bounded verify and replay evidence → windowed observation and true-state reads in the simulator, then the
   canvas on top → R9-22 → remaining P9 findings → full P9 review round → close P9.
2. P10: design note for the live-control bridge surface and pacing, then implementation, then its review rounds.
3. P11a, P11b, P11c in turn, each with its design note, implementation and review rounds.
4. A1–A7 completed and the final whole-UI review round.

Function comes before visual polish in every phase. A missing capability goes into the binary first.

## Review protocol

- **Reviewers.** Each round launches three independent **Sonnet** subagents, read-only on the repository:
  1. **design director** (CANVAS_UI_ACCEPTANCE §3.2, the frozen mockups, Blueprints style, both themes, three widths);
  2. **desktop input and accessibility** (§3.3, keyboard model, focus, announcements, axe);
  3. **process and test engineer** (§3.1 truthfulness, the nine non-negotiables, tests, architecture, gates).
- **Frozen build.** Each round reviews a copied binary and UI build plus a freshly recorded workspace (recipe in the
  P9 handoff); each reviewer starts and stops only its own bridge, by PID.
- **Scoring.** 1–5 per reviewer; a round is approved when all three score at least 4 with no High finding.
- **Record.** Every finding and disposition goes into `docs/simulation-canvas/reviews/` (one file per phase) and,
  for phase findings, into `governance/FINDINGS.md`. Fixes are committed per round.
- **Implementation** is done in the main session; Sonnet subagents may take mechanical or simulator-side tasks in
  their own worktrees, and the main session reviews every diff they make to `test/architecture`, `scripts`,
  `AGENTS.md` and module facades before merging.

## Constraints (binding)

- AGENTS.md in full: the nine non-negotiables, determinism, logic in its owning module, no gate loosened without the
  owner's approval and an `Owner-approved-gate-change:` line, 300-line files, 15-line Go functions, no code comments
  beyond package docs.
- The canvas computes nothing the binary owns: validation, composition, simulation, truth, scoring, verification.
- Real-binary Playwright tests, no mocks of the binary, no fixed sleeps; desktop only.
- Customer data and configuration never enter GitHub; customers never learn about each other.
- Commit each round; push and open pull requests only when the owner asks.

## Out of this goal

Everything GOAL.md places in the next goal (Agentic Stream as plant-wire host, gateway, headroom measurement), the
reference worlds over the plant wire (P12) and hardening (P13), mobile layouts, and bundling the UI into the binary.
