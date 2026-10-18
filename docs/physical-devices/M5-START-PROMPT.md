# Starting prompt for M5 (R11, R12) and R13, any coding agent

Continue the physical-devices goal in /Users/ghassan/my-projects/streams-simulator with milestone M5 (rounds R11 and
R12), then R13.

Read first, in this order: AGENTS.md (binding; MUST 1–5, the Go and Canvas rules), docs/physical-devices/HANDOFF.md
(state at the end of M4, what is open from M4, the bench recipe and the traps), docs/physical-devices/GOAL.md (rounds,
quality bar, round log), docs/design/PHYSICAL_DEVICES.md (§3, §4.1, §4.2 interrupts, §5 milestones, and the R7–R10
as-built notes), docs/design/RUN_JOURNAL.md, governance/DEFERRED.md (D-60 to D-79), contracts/plant-wire/SPEC.md,
canvas/README.md, governance/CANVAS_UI_ACCEPTANCE.md, docs/simulation-canvas/UI-EXPECTATIONS.md and
docs/simulation-canvas/ui-architecture-20261016/README.md.

State: M4 is done on branch `physical-devices-m4` (pushed; `make phase-gate` green on e323b0c0). Nothing is merged; the
owner merges at the end. First cut `physical-devices-m5` from the tip of `physical-devices-m4` and push it. Before R11,
do item 1 of "Open from M4" in HANDOFF (check in a real browser that the first Open live view shows the plant).

R11 (M5): node interrupts as a local/remote switch at the node (PHYSICAL_DEVICES §4.2): hold, approve each command,
emergency stop; the node may only refuse or stop, never originate or widen; `local_control` refusal to the host;
journal and command-log records; canvas controls always visible on the live screen. Everything a hold or an approval
decides is reproduced on replay.
R12 (M5): the twin-versus-board report (where the twin's prediction and the board's readings diverge, named) and the
Inspect twin overlay (the twin's prediction drawn over the real trend).
M5 exit (§5): a host decision drives the real fan, confirmed by an independent instrument rather than DRESULT; hold,
approval and emergency stop each shown on the bench and reproduced on replay; the report names the divergences.
Hardware is the owner's to test at the end; use the scripted stand-in gateway and `plantwire host --script` in tests.
R13: public documentation in `documentation/` (what was built, limitations), handoff briefs for the work outside this
repository (the Arduino gateway's plant wire side, Agentic Stream's plant wire host and `local_control` handling, the
bench instrument, the owner's bench checks for M2–M5), and the final review; `make phase-gate` green on the last commit.

Rules that do not change: the binary owns every rule (validation, matching, origin, health, command stages, interrupt
decisions, divergence); the canvas computes nothing the binary owns and parses no message text. Never loosen a gate: no
cap raised, allow-list grown, floor lowered or test hook added, and no `//nolint` or `eslint-disable`; registering a
genuinely new file as prescribed is allowed and is named in the commit; raising a floor is a tightening. Go files ≤ 300
lines, production function bodies ≤ 15 lines, facades only delegate, tests at the layer they prove and asserting specific
errors, no map iteration in output code. Canvas files ≤ 300 lines, functions ≤ 40, complexity limits, no boolean flag
parameters, coverage floors, e2e against the real binary with exact locators, no `nth()`, no fixed sleeps. Contract
changes regenerate per scripts/plant-wire/README.md. Run `go test -shuffle=on` locally, start every gate list with
`make build`, run `make phase-gate` before pushing. One commit per round plus a round-log commit; record deviations in
the round log. Milestone end: a green `make phase-gate`, a correctness/architecture review, an adversarial review, a
canvas code-quality review and a real-browser design review at 1280 and 1440 px, light and dark, followed by a
re-check; fix every High, fix or defer each Medium with a reason. In M4 the real-browser review found Highs that every
test missed (a false alarm banner, a blank live view): do not skip it, and when it reports a failure, capture network
and console evidence before fixing. Do not pause for owner testing; the owner tests hardware and product feel at the
end and merges to main.

Owner's model split for Claude Code (adapt to your tool if different): the main session briefs, reviews gate-file diffs
and commits; writers write code and docs and run nothing; a cheaper runner runs every command and fixes lint-class
findings; the hardest canvas work goes to the strongest model; the browser review is done by a separate agent in the
user's Chrome. Keep the main session's token use low.
