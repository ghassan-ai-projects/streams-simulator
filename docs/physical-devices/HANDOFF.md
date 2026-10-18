# Handoff — physical-devices goal, state at the end of the first session (2026-10-18)

Read first, in this order: [GOAL.md](GOAL.md) (outcome, rounds R1–R13, quality bar, round log),
[../design/PHYSICAL_DEVICES.md](../design/PHYSICAL_DEVICES.md) (the design and as-built notes),
`governance/DEFERRED.md` (D-60 to D-72), `AGENTS.md`.

## Branches

| Branch | Holds | State |
|---|---|---|
| `simulator-new-vision` | M1, M2 (R1–R6b), review and CI fixes | [PR 44](https://github.com/ghassan-ai-projects/streams-simulator/pull/44), open, **not merged**; last full `make phase-gate` green |
| `physical-devices-m3` | R7a, R7b, R8 (work in progress) | pushed; R7a and R7b are committed with green Go gates; **R8 is committed as work in progress and its gates fail** (below) |

Nothing is merged to `main`. The owner merges everything when the whole goal is done and the owner has tested
(hardware and product feel are the owner's to test at the end; do not pause for owner tests mid-goal). Cut the next
branch (`physical-devices-m4`, ...) from the previous tip at each milestone and push it.

## R8 — what is written, what fails

Written (see the R8 row of GOAL.md and PHYSICAL_DEVICES.md as-built notes): the field trace as a run input
(`field-trace.jsonl`, `field_trace_digest` in the run artifact, optional so existing runs are byte-identical), a mixed
relay (`plantwire relay --domain ...`: one node, gateway boards `hardware` plus simulated boards, commands routed by the
board's origin, no code path from the simulated side to the gateway link), and `plantwire replay-trace` (a recording
replays as simulated boards only; every command is refused).

Open before R8 can be committed as done (from the last Haiku gate run):

1. `test/architecture` T4: `internal/plantwire/internal/app/relay_mixed_faults_test.go` has one test that only asserts
   that an error occurred — assert the specific error.
2. `test/architecture` facade rule: `internal/run/service.go` `OpenFieldTrace` does more than delegate — move the logic
   into the run module's app layer; the facade is one delegating call.
3. `test/architecture` function length: `internal/plantwire/internal/app/relay_loop.go` `relayLoop.step` has an 18-line
   body (a `select` with 7 cases); it needs a design decision (group the sim ready/finished channels, or split by
   event source into named steps) — do not compress statements.
4. A flaky test: `TestACommandDuringAReplayIsRefusedAndReachesNothing` (`internal/plantwire/internal/app/replay_trace_test.go:141`)
   fails 2 of 3 isolated runs: the `unknown_device` receipt is sometimes missing from the snapshot — a race in the replay
   player or in the test's wait; fix the cause, not the test timing alone.
5. Coverage floors (never lower them): `internal/plantwire` facade 92.3% vs 99; `internal/plantwire/internal/uds` 66.7%
   vs 90; `internal/run/internal/durable` 63.9% vs 70; `internal/plantwire/internal/app` unknown until the tests pass.
   Add tests for the new code.
6. Not done by the R8 agent: D-71 (mixed-session wall-clock pacing, deliberately deferred by the owner), D-72 (the
   world part of a mixed recording is reproduced by `replay`, not streamed by `replay-trace`), the canvasbridge run-file
   lists and `documentation/reference/contracts.md` for the new file/members, and the R8 commit hash in the GOAL.md round
   log (it says "(this round)").

Then the usual: Haiku gate run, fill the round log, commit, push, and run `make phase-gate` before M3's review.

## Remaining rounds (details in GOAL.md)

R9 canvas Connect session, real or simulated per session, origin marks (the owner suggested **an Opus subagent for R9**,
the hardest canvas part) · R10 live mode, command lifecycle, connection health · R11 node interrupts (hold, approve each
command, emergency stop, `local_control` refusal) and canvas controls · R12 twin-versus-board report and Inspect
overlay · R13 public documentation, handoff briefs for the other repositories, final review. Milestone ends get the
two-lens review (A correctness/architecture; B adversarial; canvas milestones add a code-quality lens and a real-browser
design review at 1280 and 1440 px, light and dark) and a green `make phase-gate`.

## How we work (owner rules)

- Sonnet subagents write code and docs only, and run nothing. Haiku subagents run every command (build, tests, gates)
  and fix **all lint-class findings** (gofmt, golangci-lint, eslint, prettier, dupl/jscpd, knip, deadcode, function and
  file length). Opus for the hardest rounds (owner request for R9). The main session briefs, reads gate-file diffs,
  commits, and keeps its own token use low.
- Never loosen a gate (AGENTS.md MUST 3): no cap raised, no allow-list grown, no floor lowered, no test hook. **No
  `//nolint` directives** (owner: not acceptable; four remain from before with `DEFERRED.md` entries D-60/D-63/D-64).
  Registering a genuinely new file, edge file or package as the standard prescribes is allowed; name every gate edit in
  the commit message with `Owner-approved-gate-change: <reason>`.
- One commit per round; update the GOAL.md round log in a follow-up commit. Commit trailers:
  `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>` (or the model that did the work).
- The canvas computes nothing the binary owns. The binary owns validation, composition, naming rules, twin knowledge.

## Traps we hit (save yourself the loops)

- `make build` writes `bin/streams-simulator`; `bin/streamsim` is a stale 2026-10-13 binary. Start every gate list with
  `make build`; the canvas e2e runs the freshly built binary.
- Run the long `make phase-gate` yourself with `Bash run_in_background` writing to a log in the scratchpad, then read
  the log. A subagent's background process dies when the subagent ends.
- CI runs tests with `-test.shuffle` and a pinned Chromium: run `go test -shuffle=on` locally. Playwright locators must
  use `exact: true` when a label is a substring of another ("Device" vs "Device details"); avoid positional `nth()` and
  one-shot `evaluate`; use retrying `expect`.
- `test/architecture` gates are per file: new files with I/O need an `ioEdges` entry; new CLI files need a
  `cliAppFiles` entry; production functions are at most 15 body lines; facades only delegate; fixture paths are never
  written as strings in tests (use `internal/testsupport`); tests assert specific errors; no mention of `docs/`
  outside docs.
- Behaviour pin: `scripts/behaviour-pin | diff test/pin/BEHAVIOUR_PIN.txt -` must be empty for existing runs; the owner
  approved refreshing only `failure/unknown-command/stderr` and `failure/domain-invalid/stdout` rows (usage text, issue
  fields).
- Contract files under `contracts/plant-wire` are generated from `scripts/plant-wire/{schemas,fixtures}.py`; regenerate
  with the commands in `scripts/plant-wire/README.md` (schemas.py, then fixtures.py > DIGEST, then copy the schemas to
  `internal/plantwire/contract/schemas`).
- The thermal LED/fan experiment must stay green (`go test ./internal/device/... ./internal/deviceworld/... ./test/contract/...`).

## Not in this repository (handoff briefs are part of R13)

The Arduino gateway's plant wire side, Agentic Stream's plant wire host and `local_control` handling, Tamoz, firmware, the
bench instrument, and the owner's bench checks. Clock pacing for mixed sessions is deferred (D-71).
