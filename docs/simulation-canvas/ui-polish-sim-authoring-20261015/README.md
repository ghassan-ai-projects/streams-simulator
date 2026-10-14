# UI polish through simulation authoring — 2026-10-15

Goal: ten reasonable, useful, multi-level simulations authored through the canvas UI in Chrome; every comprehension or
functional problem met on the way is reported here, fixed one commit per fix, and the interface re-evaluated until it
scores 10/10.

Roles: an author (Haiku) only builds the simulations through the UI and reports blockers; a hunter (Haiku) separately
hunts for comprehension, functional and improvement issues, including real-world failure probes. Both append to
`issues/round-N-*.md` as they go; a Sonnet fixer fixes them one commit each while they work; the orchestrator reviews every
diff; an independent evaluator scores the interface in `evaluation-round-N.md`. The loop ends when ten simulations exist and the evaluation is 10/10.

The authored plants live in a scratch workspace outside the repository; each round's notes list their ids, areas, equipment,
instruments and the actions exercised.
