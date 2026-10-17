# Goal — physical devices in streamsim (M1–M5)

- **Set by the owner, 2026-10-17:** finish the work in [PHYSICAL_DEVICES.md](../design/PHYSICAL_DEVICES.md), one commit per
  round, until all is done. Sonnet subagents write most of the code; Haiku subagents run gates, tests, lint and commands; the
  main session writes briefs, reviews boundaries and commits, and keeps its own token use low.
- **Branch:** `simulator-new-vision`.

## Outcome

Everything `streamsim` owns in M1–M5 works end to end: a device is defined in the canvas, runs as a twin over the plant wire,
is captured from a connected board, relayed beside simulated boards, recorded and replayed byte-identically, shown live with
its commands and health, held, approved and stopped at the node, and compared with its twin. Hardware is stood in for by a
scripted gateway in tests.

## Rounds (the whole scope)

| Round | Milestone | Deliverable |
|---|---|---|
| R1 | M1 | Kind file contract in `contracts/plant-wire` (one `NBIRTH.kinds` entry), `kind_digest` per kind in `NBIRTH`, `DIGEST` and safety-floor tests updated |
| R2 | M1 | Workspace kind files; domain kind `implements`; the "behaviour implements interface" check; `kind validate` |
| R3 | M1 | Command response templates as domain data; thermal kind files (sensor, indicator, fan); the Arduino twin as components and a domain implementing them |
| R4 | M1 | Bridge: workspace kinds and components read, written and validated through actions |
| R5 | M1 | Canvas: Devices section and the device editor (interface, behaviour, check) |
| R6 | M2 | `kind capture` from an `NBIRTH`; canvas "save as device"; the scripted stand-in gateway (test tool, `environment: hardware`) |
| R7 | M3 | Per-board environment in the contract; relay mode (gateway host side, node side to the host); pass-through rules; replay has no gateway option |
| R8 | M3 | Field trace as a run input; mixed sessions (simulated boards beside relayed ones); byte-identical replay |
| R9 | M3 | Canvas: Connect session, real or simulated per session, origin marks |
| R10 | M4 | Canvas: live mode, command lifecycle, connection health (and any journal record they need) |
| R11 | M5 | Node interrupts: hold, approve each command, emergency stop; `local_control` refusal; journal and command log; canvas controls |
| R12 | M5 | Twin-versus-board report; Inspect twin overlay |
| R13 | — | Public documentation (`documentation/`), limitations, handoff briefs for other repositories, final review |

A round that turns out too large is split (R3a, R3b); a round's order changes only when a dependency demands it.

## Quality bar (every round)

1. **Brief.** The main session gives one Sonnet subagent a self-contained brief: deliverable, files or modules, the rules
   below, and what done looks like. A new package or contract states kind, layer, responsibility and callers first
   (AGENTS.md MUST 4).
2. **Code rules.** AGENTS.md in full: logic in its module, one use case one implementation, files ≤ 300 lines, functions ≤ 15
   lines (canvas: ≤ 40 lines, complexity limits), no consumer knowledge in the binary, domains and devices are data, no new
   dependency, minimal comments.
3. **Never loosen a gate** (MUST 3). No cap, allow-list, coverage floor or test hook changes without the owner's approval in the
   conversation.
4. **Tests at the layer they prove**, for every changed behaviour; contract changes have fixtures and schema tests; any change
   to what a run produces has a replay-determinism test.
5. **Every command by Haiku, dispatched by the main session.** The Sonnet subagent writes code only: it runs and spawns
   nothing and returns when the code is written. The main session then starts one Haiku gate agent with a fixed command list
   (focused `go test` for the touched packages, the thermal-experiment tests, `make quick`; `npm run check` in `canvas/` for
   canvas rounds) that reports only failures and their output. Failures go back to the same Sonnet agent by `SendMessage`, so
   it keeps its context; then Haiku runs the gate again. The behaviour pin runs only in rounds that claim no behaviour change.
6. **Boundary review by the main session.** `git diff --stat` for the round, then a read of every change to
   `test/architecture`, `scripts/`, `contracts/`, module facades, `AGENTS.md` and the Makefile.
7. **Thermal experiment.** `go test ./internal/device/... ./internal/deviceworld/... ./test/contract/...` stays green every
   round.
8. **Commit.** One commit per round, naming the round and listing any gate change with its reason.
9. **Milestone end** (after R5, R6, R9, R10, R12): `make phase-gate` run by Haiku in the background; one Sonnet reviewer
   reads the milestone diff against this bar and PHYSICAL_DEVICES.md (High findings fixed before the next milestone); for
   canvas milestones, the main session checks the screens in the browser at 1280 and 1440 px, light and dark.

## Done means

- R1–R13 committed; `make phase-gate` green on the last commit.
- PHYSICAL_DEVICES.md and `documentation/` describe what was built; deviations are recorded in the round log below.
- Handoff briefs exist for the work outside this repository: the Arduino gateway's plant wire side, Agentic Stream's plant
  wire host and `local_control` handling, the bench instrument, and the owner's bench checks for M2–M5.

## Outside this goal

The gateway, firmware, Agentic Stream and Tamoz changes, and the physical bench exits; time in mixed sessions beyond the
existing barrier wall cap; authentication and several users; the canvas production-UI goal items not touched by a round.

## Round log

| Round | Commit | Notes |
|---|---|---|
| R1 | 7e52f2c3 | Kind file schema, `kind_digest` in `NBIRTH` (contract 1.1.0), host refuses a mismatch. Only the contract generators under `scripts/plant-wire` changed; no gate changed. The behaviour pin is stale at HEAD before this round (16 `run.json` hashes and `manifest-file` differ on a clean checkout); not touched, to be refreshed in a round that claims no behaviour change. |
| R2 | 8d501631 | `implements` on domain kinds; kind files from `--kinds-dir`; the check lives in plantwire (domain cannot import it) and reaches `run` as an injected `CheckDomain`; `kind validate`. Gate changes: new edge package `plantwire/internal/files` registered as prescribed; two CLI files registered, `component.go` cap lowered 130 to 70; plantwire domain coverage floor raised 93 to 95 (ratchet). |
