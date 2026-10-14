# Goal: the canvas feels like a finished desktop product

- **Set by the owner, 2026-10-14**, after PR #34 merged: "continue on improving it, loop and improve the UI, the user
  experience … I am still not satisfied, use the director reviewer as always."
- **Branch:** `claude/canvas-ux`, from main at `2820170`.

## Outcome

The four flows of [GOAL-UI-PRODUCTION.md](GOAL-UI-PRODUCTION.md) work today. This goal makes them feel finished:

1. workspace;
2. build and edit a plant;
3. run a plant;
4. review a run.

In practice that means:

- A first-time user knows where to start and what each screen is for.
- Every state reads clearly (loading, empty, unsaved, running, refused, failed).
- Layout uses the screen well from 1280 to 1920 px.
- Wording is plain and consistent.
- Every action is fast to reach by mouse and keyboard.

## The loop

Each round runs these steps:

1. **Freeze a build.** Capture every screen of every flow against the real binary at 1280, 1440 and 1920 px in light and dark.
2. **Review.** Three Sonnet reviewers review read-only:
   - the design director;
   - desktop input and accessibility;
   - the process and test engineer.
3. **Fix.** Fix every High and Medium finding, and the Low findings that are cheap.
4. **Verify.** Re-check in Chrome myself.
5. **Commit and record.** Commit the fixes and record the round in `docs/simulation-canvas/reviews/UX-round-N.md`.

## Done means

- The design director scores **5**.
- The other two reviewers score **at least 4**.
- No High or Medium finding is open.
- At most five rounds; if the bar is not met by then, I report what still holds it back.
- `make phase-gate` and the canvas gate pass.
- Pipelines stay under two minutes.
- Every flow keeps its Playwright test against the real binary.

## Rules

- All of AGENTS.md applies.
- The canvas computes nothing the binary owns. A missing capability goes into the binary first, in its smallest form.
- The scope is the four flows' experience. The goal's postponed list still stands unless a round shows an item hurts a flow badly enough to bring it in, and the round record says so.
- Commit each round. Push and open pull requests only when the owner asks.
