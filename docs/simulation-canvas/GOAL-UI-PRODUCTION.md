# Goal — the canvas UI, basic functionality working well

- **Set by the owner, 2026-10-14:** "set a goal to finish the UI and do not drift, I want to have the basic
  functionality working well, anything that holds us we can postpone to later or remove it completely." The owner
  designs, updates and runs plants in Chrome only.
- **Branch:** `codex/canvas-p9-playback` until the first flow lands, then one branch per flow from the latest `main`.

## Outcome

In Chrome, with no terminal, a person can **build** a plant, **run** it and **review** the run. Each of these works
end to end on real files through the real binary, and looks and behaves like a finished product on a desktop
(1280–1920 px, light and dark).

## The four flows (the whole scope)

1. **Workspace.** Open the workspace; see plants, compositions and runs; open any of them. *(Exists.)*
2. **Build and edit a plant.** Start a new composition or open an existing one; add equipment from the component
   library (component, preset, name, area); connect ports by dragging between pins; rename, move or delete; the
   binary composes and validates it and every issue shows on the element it names; save. The saved composition and
   its composed domain are what the CLI would produce.
3. **Run a plant.** From a plant: choose seed, duration, Explore or Evaluate, and faults to inject (instance, fault,
   time); start; see progress; when it ends the run appears in the workspace and opens in the review.
4. **Review a run.** Play back, seek, step through events, read the report, the verdict and its reason. *(Exists;
   only regressions and real defects are fixed.)*

## Done means

- Each flow has a Playwright test against the real binary on real files, and all pass.
- A Sonnet review round per flow (design director, desktop input and accessibility, process and test engineer)
  scores at least 4 from each with no High finding; the record is in `docs/simulation-canvas/reviews/`.
- I checked each flow myself in Chrome at 1280, 1440 and 1920 px in light and dark.
- `make phase-gate` and the canvas gate pass; pipelines stay under two minutes.

## Postponed (not in this goal; nothing here blocks it)

- Run storage and its replacement: decision D-16 (actions, not data), with R9-04, R9-05 and R9-14.
- Live control during a run: play, pause, step, speed, reset, injecting faults mid-run (P10 beyond "run to the end").
- Explore true state beside reported values (R9-13, R9-37); negative air flow (R9-22).
- The full builder of P11: kinds, behaviour, sensors, actions, scenario files, templates, diff, layout sidecar.
- Scale budgets for the 26-hall campus (A6), hall summary figures (D5-12), frame-title layout, compact-tier kind
  ellipsis (D6-01) and every other polish finding that does not break a flow.
- Acceptance tests A1–A7 as written; this goal's four flow tests replace them for now.
- P8b's open findings (R0-15, R0-16, R2-01), the plant wire, reference worlds and hardening.

- Compose returning its output path and validating the plant id in the binary (round 1, E-5 and E-6); the editor
  follows the `domains/<id>.domain.json` convention until then.

Anything else that turns out to hold a flow back is postponed here with one line, not worked around.

## Rules

- AGENTS.md in full; the canvas computes nothing the binary owns; a missing capability goes into the binary first,
  in its smallest form.
- Function first, then visual quality, flow by flow; no work outside the four flows.
- Commit each step; push and open pull requests only when the owner asks.
