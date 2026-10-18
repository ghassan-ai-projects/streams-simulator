# Handoff — physical-devices goal, state at the end of M4 (2026-10-18)

Read first, in this order: [GOAL.md](GOAL.md) (outcome, rounds R1–R13, quality bar, round log),
[../design/PHYSICAL_DEVICES.md](../design/PHYSICAL_DEVICES.md) (the design and as-built notes),
`governance/DEFERRED.md` (D-60 to D-72), `AGENTS.md`.

## Branches

| Branch | Holds | State |
|---|---|---|
| `simulator-new-vision` | M1, M2 (R1–R6b), review and CI fixes | [PR 44](https://github.com/ghassan-ai-projects/streams-simulator/pull/44), open, **not merged** |
| `physical-devices-m3` | R7a, R7b, R8 | pushed; `make phase-gate` green |
| `physical-devices-r9` | M3 complete: R9 and the M3 review fixes, on top of `physical-devices-m3` | [PR 46](https://github.com/ghassan-ai-projects/streams-simulator/pull/46) to `main`, open; `make phase-gate` green |
| `physical-devices-m4` | M4 complete: R10 (Codex WIP db17bd5e finished in c82fcb5e) and the M4 review fixes (cb327f45, e323b0c0) | pushed; `make phase-gate` green |

Nothing is merged to `main`; the owner merges at the end. **Start M5 by cutting `physical-devices-m5` from the tip of
`physical-devices-m4`** and push it.

## M4 — done (2026-10-18)

R10 and its reviews (A correctness/architecture, B adversarial, canvas code quality, real-Chrome design review with two
re-checks); every High fixed. See the GOAL.md round log (R10, M4 review fixes, M4 re-review) and the R10 sections of
PHYSICAL_DEVICES.md. Open from M4, to pick up in M5 or R13:

1. **Verify in a real browser first**: the first "Open live view" right after Run plant must show the plant at once
   (the last fix, e323b0c0, is covered by an app test and the e2e but was not re-checked in Chrome).
2. Timeline pill de-collision (overlapping pills) and canvas fit-zoom on scrub (deferred design pass).
3. The bench host issues no commands, so the commands table and the refused / unconfirmed / confirmed stages were
   checked only by unit, component and domain tests — the M4 exit item "told apart at a glance" still needs a browser
   look with a host that issues commands (`plantwire host --script`).
4. True 1280 px was never emulated in Chrome (the window would not shrink; 1440, 1471 and 1728 were tested, 1280 only
   by CSS zoom). Try a smaller Chrome window or the built-in browser pane.
5. The stand-in gateway exits by itself after about 4 min 48 s on the bench (test support; long bench sessions need a
   restart).
6. D-77 (live record growth), D-78, D-79 and D-71/D-72/D-74–D-76 in DEFERRED.md.

## M3 — done (2026-10-18)

R7a–R9 committed; two-lens review (A correctness/architecture, B adversarial), canvas code-quality review and a
real-Chrome design review (1280/1440, light/dark) with a re-review; every High fixed. See the GOAL.md round log
(R8, R9, "M3 review fixes") and the PHYSICAL_DEVICES.md as-built notes for R9 and the review fixes.

Carried into R10 (live mode, command lifecycle, connection health), owner-visible gaps found in the M3 reviews:

1. A reload during a live session gives no affordance: no top-bar session indicator, no "running" chip leading back to
   the live panel (design review M4/N9).
2. The plant screen shows no origin marks or live state while a session runs (M11).
3. The live review uses alarming copy: an amber "Running" badge, "Incomplete: run.json is absent", "Not checked"; it
   should read as live, with verification after Stop (N5).
4. "Connected session" and the plan's simulated instances are bridge memory (`RunEntry.connected`, `RunEntry.Plan`): after
   a bridge restart an old session lists as Evaluate and its simulated instances as "Origin unknown". Record the session
   mode (and the plan check) in the run itself, e.g. in `run.opened` or the artifact, and read it from there.
5. A blank row above the live status line in the Run panel; the Run panel's scroll area at 1280 is about 370 px.
6. D-73 (no gateway liveness watchdog: a silent but connected gateway is never noticed) belongs to connection health.

Also open: D-71, D-72, D-74–D-76 (DEFERRED.md). The canvas `schema_version: "0.1"` of the session plan is a canvas
constant because neither the bridge nor the binary supplies it.

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
- `make build` also builds `bin/standin-gateway` (a test binary of `internal/plantwire/internal/app`, run with
  `--standin-gateway`), which the canvas e2e uses as the board. A by-hand bench: an owner-only directory (`mkdir -m 700`),
  `bin/streams-simulator plantwire host --socket D/h.sock --environment mixed --delay-ms 100`, `bin/streams-simulator canvas
  --ui-dir canvas/dist --workspace W --addr 127.0.0.1:7788` (W holds `domains/` with
  `examples/devices/thermal-board/thermal-board.domain.json` as `arduino-thermal.domain.json` and `kinds/` with its kind
  files), then after the session starts `bin/standin-gateway --standin-gateway --socket D/g.sock --kinds-dir
  examples/devices/thermal-board --board uno-07 --device zone-01:thermal_sensor:61` (another device id gives a refusal).
- Under `-cover` (`make phase-gate`), a re-executed test binary prints a `GOCOVERDIR` warning on stderr: never compare a
  child's stderr exactly.
- Canvas: the emitted-CSS oracle checks module import order; a new component importing a graph stylesheet early puts
  `graph` before `vendor-react-flow`. Keeping inactive panes mounted duplicates labels (two "Seed" inputs) and breaks
  locators; unmount them and keep drafts in the store.
- `e2e/devices.spec.ts:115` (a device gets a twin) runs 7–12 s alone but once hit the 30 s test timeout in a full
  `make phase-gate` on the M3 tip; it passed 5/5 alone and the full e2e (114) passed on rerun. Watch it in R10; if it
  recurs, find what it waits on rather than raising the timeout.
- The thermal LED/fan experiment must stay green (`go test ./internal/device/... ./internal/deviceworld/... ./test/contract/...`).

## Not in this repository (handoff briefs are part of R13)

The Arduino gateway's plant wire side, Agentic Stream's plant wire host and `local_control` handling, Tamoz, firmware, the
bench instrument, and the owner's bench checks. Clock pacing for mixed sessions is deferred (D-71).
