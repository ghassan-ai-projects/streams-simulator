# Goal: ten useful simulations, authored through the canvas, interface at 10/10

Branch `ui-polish-sim-authoring`. Folder of record: `docs/simulation-canvas/ui-polish-sim-authoring-20261015/`.
The bar is `UI-EXPECTATIONS.md`. Roles and loop are in that folder's README.

## Done when (all of these)
1. Ten simulations exist in the scratch workspace (`scratchpad/sims/lab`), each composed, run at least once with two or more
   faults and verified, and each reasonable and useful: at least two area levels, 10+ instances, 6+ instruments/channels,
   8+ links, 3+ kinds with actions; at least five also use effector calls once the UI offers them.
   Ids so far: office-tower-hvac, compressed-air-works, water-pumping-station; to build: data-centre-cooling,
   boiler-steam-loop, packaging-factory, power-distribution-standby, greenhouse-ventilation, food-cold-chain,
   multi-hall-campus.
2. The UI/UX director's overall score is 10/10 in the latest `evaluation-round-N.md` (overall = lowest of ten dimensions),
   with no High or Medium finding open in any `issues/*.md` and `evaluation-*.md`.
3. `cd canvas && npm run check` and `GOTOOLCHAIN=go1.26.9 PATH="$PATH:$HOME/go/bin" make ci-check` pass (a load-induced fuzz
   timeout is re-run alone).
4. Simulator gaps found by the failure probes are recorded in `simulator-backlog.md` (not built here).

## Loop (one round)
author (Haiku, builds sims, reports blockers) and hunter (Haiku, finds issues, failure probes) then fixer (Haiku, one commit
per fix) then orchestrator review of every diff, then director (Haiku, evaluation). Run agents one at a time in the single
Chrome tab (`http://127.0.0.1:8766`, paired). Restart the canvas server on a binary built from HEAD before each round.

## Constraints
- All delegated agents use model Haiku (owner's usage budget). Keep rounds small: at most 3 simulations per author run.
- One commit per fix. Never push or open a PR; the owner does that.
- Never loosen a gate. The behaviour-pin refresh needs the owner's `Owner-approved-gate-change:` line: do not touch it.
- Canvas computes nothing the binary owns; binary gaps go to the owning module with tests, no per-domain code.
- No code comments beyond package docs; customer data never in the repo; screenshots are not committed.
- Do not stage `docs/product-strategy-20261011/`, `docs/simulation-ideas/`, `docs/simulation-canvas/mockups/data.json`.

## Stop rules
Stop and write a status note in the folder README when: ten simulations and a 10/10 director score are reached; or six rounds
have run without the director's score rising; or a fix needs a gate change or a design decision (record it, do not decide).
Scores are reported as measured; never raise one to finish.
