# Handoff — physical-devices goal, state at the end of R13 (2026-10-19)

Read first, in this order: [GOAL.md](GOAL.md) (outcome, rounds R1–R13, quality bar, round log, "Done means" status),
[../design/PHYSICAL_DEVICES.md](../design/PHYSICAL_DEVICES.md) (the design and every as-built note),
[POLICY-ALLOW-LIST.md](POLICY-ALLOW-LIST.md), `governance/DEFERRED.md` (D-56 to D-113), `AGENTS.md`, and the briefs in
[briefs/](briefs/README.md).

## Where everything is

| Place | Holds | State |
|---|---|---|
| `origin/main` (`2f58b7c0`) | M1 to M4: PRs 44 (`simulator-new-vision`), 45 (`physical-devices-m3`), 46 (`physical-devices-r9`), 47 (`physical-devices-m4`) | merged (the local `main` branch in this checkout is stale at PR 43: `git fetch` first) |
| `physical-devices-m5` (`f491fbbc`) | M5: R11a, R11b, R12a, R12b, the M5 reviews and fixes, the UI test-speed merge, D-94 to D-109 | pushed; no pull request known to the author of this note (the author could not query GitHub) |
| `policy-allow-list` and `physical-devices-r13` (`63bb6212`) | the policy allow-list (D-88): design, P1 (contract and binary), P2 (bridge and canvas) and the Chrome review fixes, on top of `physical-devices-m5` | pushed |
| `physical-devices-r13` (working tree) | R13: public documentation, briefs, limitations, this handoff | **uncommitted** when this note was written; the runner checks and the owner commits |

The owner merges at the end. Cut the next milestone's branch from the tip of the previous one (owner rule). Nothing after
`origin/main` is merged as far as this checkout shows.

## What is done

M1 to M5 and the policy allow-list are built in the binary, the bridge and the canvas, with two-lens reviews, a canvas
code-quality review and real-Chrome design reviews for each milestone (GOAL.md round log). R13 wrote:

- `documentation/guides/physical-devices.md`, `documentation/architecture/physical-devices.md`, a "Physical devices" section in
  `documentation/limitations.md`, `documentation/roadmap.md`, `documentation/architecture/security-model.md`, the claims matrix,
  and the reference updates (`canvas` command, twin-choices contract, `interrupts.jsonl`, journal records);
- the four briefs in `docs/physical-devices/briefs/`;
- D-111 to D-113 in `governance/DEFERRED.md`.

**What no one has done:** run anything against a real board. The simulator's tests use a scripted stand-in gateway. The
physical exits of M2 to M5 (PHYSICAL_DEVICES section 7) are the owner's bench checks (`briefs/bench-checks-m2-m5.md`), the
gateway's plant wire side, the Agentic Stream changes, Tamoz workflows and the independent instrument are other repositories'
work (`briefs/`). The second proof (the co-founder's board) is not started.

## Open from M4, as known from the round log

1. First "Open live view" right after Run plant: the last M4 fix (`e323b0c0`) was not re-checked in Chrome at M4; the M5
   "plant of a running session opens by address" fix (H3) and the M5 browser reviews cover live sessions. Not recorded as a
   separate check. **Confirm in bench check M4-1.**
2. Timeline pill de-collision (overlapping pills) and fit-zoom on scrub: still deferred, not in DEFERRED.md. The M5 refit change
   (`useFitted` records the applied viewport) is a different fix.
3. Refused, unconfirmed and confirmed command rows: since R11b the e2e host issues commands (`plantwire host --script`), so
   the refused (`local_control`, `policy_not_allowed`) and accepted rows are checked in a real browser; the "effect observed" row
   is covered by unit and component tests only. **Bench check M4-2.**
4. True 1280 px: the M5 design reviews ran at 1280 x 800 and 1440 x 900. Closed.
5. The stand-in gateway ends by itself after about 4 min 48 s (test support). Open, harmless.
6. D-77, D-78, D-79, D-71, D-72, D-74 to D-76: open (below).

## D-numbers still open (`governance/DEFERRED.md`)

D-56 to D-59 (M1 review), D-65 to D-72, D-74 to D-87, D-89 to D-115. Closed: D-73 (gateway watchdog, R10) and D-88 (policy
allow-list, P1 and P2). D-60, D-63 and D-64 are the four remaining `nolint` directives (owner: not acceptable; each has a design
note). Highest impact for the bench:

- D-94 (a command already at the gateway cannot be cancelled), D-81/D-106 (controls act at a barrier or tick), D-99 (a stop can
  wait one bounded host write), D-82/D-100/D-105 (what the relay's stop can and cannot order).
- D-71 (mixed sessions not wall-clock paced), D-72 (replay does not stream the world's part), D-90/D-91/D-95/D-96 (twin report).
- D-111 (an independent instrument cannot confirm on the live screen), D-112 (no carrier for `command_evidence` in a
  hardware-only session), D-113 (exact contract-version check) are new in R13 and need owner decisions; D-114 (`connect --out`
  accepts an existing run folder) and D-115 (the reference host exits 1 without a report when its node is stopped) came from the
  documentation review and are small product changes.

## Next for the owner

1. Run `make phase-gate` on the R13 commit (documentation only; `make quick` and `make docs-check` are the narrower checks).
2. Hand the briefs to the gateway, Agentic Stream and bench work; run the bench checks as the work lands.
3. Decide D-111 to D-113 (the instrument's place in the live view; hardware-only evidence; the version rule).
4. Open a pull request for `physical-devices-m5` plus the policy allow-list and R13 if not yet done.

## How we work (owner rules)

- Sonnet subagents write code and docs only, and run nothing. Haiku subagents run every command (build, tests, gates)
  and fix **all lint-class findings** (gofmt, golangci-lint, eslint, prettier, dupl/jscpd, knip, deadcode, function and
  file length). Opus for the hardest rounds. The main session briefs, reads gate-file diffs, commits, and keeps its own token
  use low.
- Never loosen a gate (AGENTS.md MUST 3): no cap raised, no allow-list grown, no floor lowered, no test hook. **No
  `//nolint` directives**. Registering a genuinely new file, edge file or package as the standard prescribes is allowed; name
  every gate edit in the commit message with `Owner-approved-gate-change: <reason>`.
- One commit per round; update the GOAL.md round log in a follow-up commit. Commit trailers:
  `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>` (or the model that did the work).
- The canvas computes nothing the binary owns. The binary owns validation, composition, naming rules, twin knowledge.

## Traps we hit (save yourself the loops)

- `make build` writes `bin/streams-simulator`; `bin/streamsim` is a stale binary. Start every gate list with `make build`; the
  canvas e2e runs the freshly built binary and `bin/standin-gateway` (a test binary of `internal/plantwire/internal/app`, run with
  `--standin-gateway`).
- Run the long `make phase-gate` yourself with `Bash run_in_background` writing to a log in the scratchpad, then read the log.
  A subagent's background process dies when the subagent ends.
- CI runs tests with `-test.shuffle` and a pinned Chromium: run `go test -shuffle=on` locally. Playwright locators must use
  `exact: true` when a label is a substring of another; avoid positional `nth()` and one-shot `evaluate`; use retrying `expect`.
- `test/architecture` gates are per file: new files with I/O need an `ioEdges` entry; new CLI files need a `cliAppFiles` entry;
  production functions are at most 15 body lines; facades only delegate; fixture paths are never written as strings in tests (use
  `internal/testsupport`); tests assert specific errors; no mention of `docs/` outside docs and Markdown files. The Markdown
  link check (`TestDocumentationLinksResolve`) resolves every relative link in `documentation/`, `governance/`, `.agents/` and the
  root pages.
- Behaviour pin: `scripts/behaviour-pin | diff test/pin/BEHAVIOUR_PIN.txt -` must be empty for existing runs unless a round
  claims a change the owner approved.
- Contract files under `contracts/plant-wire` are generated from `scripts/plant-wire/{schemas,fixtures}.py`; regenerate with the
  commands in `scripts/plant-wire/README.md`.
- A by-hand bench: an owner-only directory (`mkdir -m 700`), `bin/streams-simulator plantwire host --socket D/h.sock
  --environment mixed --delay-ms 100`, `bin/streams-simulator canvas --ui-dir canvas/dist --workspace W --addr 127.0.0.1:7788`
  (W holds `domains/` with `examples/devices/thermal-board/thermal-board.domain.json` as `arduino-thermal.domain.json` and
  `kinds/` with its kind files), then after the session starts `bin/standin-gateway --standin-gateway --socket D/g.sock
  --kinds-dir examples/devices/thermal-board --board uno-07 --device zone-01:thermal_sensor:61`.
- Under `-cover` (`make phase-gate`), a re-executed test binary prints a `GOCOVERDIR` warning on stderr: never compare a child's
  stderr exactly.
- Canvas: the emitted-CSS oracle checks module import order; keeping inactive panes mounted duplicates labels and breaks
  locators; unmount them and keep drafts in the store.
- The thermal LED/fan experiment must stay green (`go test ./internal/device/... ./internal/deviceworld/... ./test/contract/...`).

## Not in this repository

The Arduino gateway's plant wire side, Agentic Stream's plant wire host and `local_control` handling, Tamoz, firmware, the bench
instrument, and the owner's bench checks: see [briefs/](briefs/README.md).
