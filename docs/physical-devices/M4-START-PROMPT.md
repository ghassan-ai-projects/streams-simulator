# Starting prompt for M4 (R10), any coding agent

Continue the physical-devices goal in /Users/ghassan/my-projects/streams-simulator with milestone M4, round R10.

Read first, in this order: AGENTS.md (binding; MUST 1–5, the Go and Canvas rules), docs/physical-devices/HANDOFF.md
(state at the end of M3, the items carried into R10, the bench recipe and the traps), docs/physical-devices/GOAL.md
(rounds, quality bar, round log), docs/design/PHYSICAL_DEVICES.md (§4.1 and the R7–R9 as-built notes, including the M3
review fixes), governance/DEFERRED.md (D-60 to D-76), canvas/README.md, governance/CANVAS_UI_ACCEPTANCE.md,
docs/simulation-canvas/UI-EXPECTATIONS.md and docs/simulation-canvas/ui-architecture-20261016/README.md.

State: M3 is done on branch `physical-devices-r9` (PR 46 to main, open; `make phase-gate` green apart from one load
timeout noted in HANDOFF). Nothing is merged; the owner merges at the end. First cut `physical-devices-m4` from the tip of
`physical-devices-r9` and push it.

R10 deliverable (PHYSICAL_DEVICES.md §4.1, milestone M4): canvas live mode (review follows the newest record while a
session runs; scrubbing back pauses following; "Back to live" resumes; an ended session is an ordinary review), the
command lifecycle (sent, accepted or refused with the reason, the board's own report, the effect observed or unknown,
shown separately; a commands table and richer timeline marks; "acknowledged" never reads as "done"), and connection
health (gateway and host link state, a board going offline and returning, reboots by `boot_id`, stale and disconnected
readings never shown as normal values; covers V-8), plus any journal record they need. Also close the six items HANDOFF
carries into R10 (reload affordance and top-bar session indicator, origin and live state on the plant screen, live
review copy, the session mode and plan check recorded in the run instead of bridge memory, the blank row in the Run
panel, D-73 gateway liveness). M4 exit (§5): one reading followed from the board to the host's decision on the live
screen; a refused, an unconfirmed and a confirmed command told apart at a glance; the same session reopened from its
recording.

Rules that do not change: the binary owns every rule (validation, matching, origin, health states, command stages);
the canvas computes nothing the binary owns and parses no message text. Never loosen a gate: no cap raised, allow-list
grown, floor lowered or test hook added, and no `//nolint` or `eslint-disable`; registering a genuinely new file as
prescribed is allowed and is named in the commit. Go files ≤ 300 lines, production function bodies ≤ 15 lines, facades
only delegate, tests at the layer they prove and asserting specific errors. Canvas files ≤ 300 lines, functions ≤ 40,
complexity limits, coverage floors, e2e against the real binary with exact locators, no `nth()`, no fixed sleeps.
Run `go test -shuffle=on` locally (CI shuffles), start every gate list with `make build`, run `make phase-gate` before
pushing. One commit per round plus a round-log commit; record deviations in the round log. Milestone end: a green
`make phase-gate`, a correctness/architecture review, an adversarial review, a canvas code-quality review and a
real-browser design review at 1280 and 1440 px, light and dark; fix every High, fix or defer each Medium with a reason.
Do not pause for owner testing; the owner tests hardware and product feel at the end and merges to main.

Owner's model split for Claude Code (adapt to your tool if different): the main session briefs, reviews gate-file diffs
and commits; writers write code and docs and run nothing; a cheaper runner runs every command and fixes lint-class
findings; the hardest canvas work goes to the strongest model. Keep the main session's token use low.

Then continue with M5 (R11 node interrupts and `local_control` refusal, R12 twin-versus-board report and Inspect overlay)
and R13 (public documentation, handoff briefs for the other repositories, final review), cutting a branch per milestone
from the previous tip.
