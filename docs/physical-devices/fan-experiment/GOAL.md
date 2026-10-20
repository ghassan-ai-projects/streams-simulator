# Goal: the fan experiment works, simulated first, then real

- **Set by the owner, 2026-10-20:** implement [PLAN.md](PLAN.md) until the real fan runs through Tamoz, Agentic Stream and
  streamsim, switching from simulated to real by the session plan only. Commit each round, reviewed before the commit.
- **Model roles (owner, 2026-10-20):** **Opus writes** (production code and its tests, documents); **Sonnet investigates**
  (reads code, traces behaviour, writes findings and briefs' facts); **Haiku runs** commands, gates, lint and lint fixes, and
  any small mechanical task. The main session briefs, reviews every diff and commits.
- **Branches:** streamsim rounds on a branch per phase cut from the previous tip, starting from `real-run`. Agentic Stream:
  `fan-h0` (H0), then `fan-plantwire-host` (A0 to A4). Gateway (`agent-research-lab`): `fan-gateway`. Tamoz: no branch (no
  code). Untracked files the owner keeps in those checkouts are left alone. Merge to `main` only when the owner says so.

## Outcome

[PLAN.md](PLAN.md) section 7's pass criteria hold on the simulated rung (P4) and on the real rung (P6), with Agentic Stream and
Tamoz not restarted between them.

## Rounds

The phases and work items of [PLAN.md](PLAN.md) section 6, in that order: P0, P0.5, P0.7 (H0.1 to H0.5), then S1 to S9,
A0 to A4 and G1 to G2 (in parallel where they touch different repositories), P4, P5 (the owner), P6. A round that turns out
too large is split; a deviation from the plan that improves it is made and recorded in the round log.

## Quality bar (every round)

1. **Brief.** The main session gives one Opus writer a self-contained brief: the deliverable, the files or modules, the rules
   below, what done looks like. When the facts are not yet known, a Sonnet investigator reads the code first and returns
   them; the brief is written from its findings. A new package or contract states kind, layer, responsibility and callers
   first (AGENTS.md MUST 4; the other repository's AGENTS.md when the round is there).
2. **Code rules.** The repository's AGENTS.md in full. For streamsim: logic in its module, files at most 300 lines, functions
   at most 15 lines, no consumer knowledge in the binary, domains and devices are data, no new dependency, minimal comments.
3. **Never loosen a gate**, and no `//nolint`: fix the cause. Lint findings are fixed by a Haiku agent that edits and re-runs
   the linter itself.
4. **Tests at the layer they prove**, written by the same Opus writer as the code. Contract changes have fixtures and schema
   tests; a change to what a run produces has a replay-determinism test.
5. **Every command by Haiku**, started by the main session with a fixed list (`make build` first, focused `go test` for the
   touched packages, the thermal-experiment tests, `make quick`), reporting failures only. Failures go back to the same Opus
   writer by `SendMessage`; then Haiku runs the gate again. Long runs (`make phase-gate`, joined runs) in the background.
6. **Review before commit** by the main session: `git diff --stat`, then every change to `test/architecture`, `scripts/`,
   `contracts/`, module facades, `AGENTS.md`, the Makefile, and the round's own diff against its brief.
7. **Thermal experiment** (`go test ./internal/device/... ./internal/deviceworld/... ./test/contract/...`) stays green.
8. **Commit.** One commit per round, naming the round (for example `fan P0.7 H0.1: …`) and listing any gate change with its
   reason.
9. **Phase end** (after P0.7, the S rounds, A4, G2, P4): `make phase-gate` (or the other repository's full gate) by Haiku in
   the background, and two Sonnet reviewers with different lenses: (A) correctness and architecture against this bar and
   PLAN.md; (B) adversarial: hostile input, bypassable checks, tests that would not fail if the code broke, the truth
   boundary. High findings fixed, Medium fixed or deferred with a reason in the round log.
10. **Failures in a joined run** are triaged by PLAN.md section 4.3 before anything is fixed.

## Done means

- PLAN.md section 7 holds on P4 and P6, with retained run folders and Agentic Stream `verify-run` passes.
- Every round committed with its review; the full gate green on the last commit of each repository's branch.
- PHYSICAL_DEVICES.md, `documentation/` and the other repositories' docs describe what was built; deviations in the round log.

## Waiting for the owner (found while working autonomously)

| # | Item | Where | Recommendation |
|---|---|---|---|
| W1 | Tamoz's prompt frame names the output protocol but never defines it (keys, allowed intents); the repair message repeats only the error code. The fan spec works around it in its own prompt (config) | P0.5-FINDINGS.md F1, F2 | A Tamoz change: put the output contract and the allowed intents in the frame and the repair message |
| W2 | Overlapping attempts: Agentic Stream redispatches 7 ms after a deadline while Tamoz still holds the episode lease, so attempt 2 fails `checkpoint_conflict` and the episode ends. Neither side's protocol forbids it | P0.5-FINDINGS.md F3 | Decide which side owns the fix; the fan run avoids it with a budget longer than a model call |
| W3 | `export-run`/`verify-run` say `pass` for a run where nothing happened (safety verdict only) | P0.5-FINDINGS.md F4 | Add a liveness or `inconclusive` state, or document it |
| W4 | Tamoz's terminal says only `graph_status: failed`; the real cause survives only in a checkpoint | P0.5-FINDINGS.md F7 | Surface the failure reason |
| W5 | D-116 (High): a command after an actuator fault's onset cancels the fault in the world core | governance/DEFERRED.md | Fix before scripted actuator faults and the A/B/C evaluation |
| W6 | `tamoz` (a consumer name) appears in Agentic Stream production code, including a key in the sealed shadow-comparison document | H0-FINDINGS.md section 7 | Backlog; renaming changes comparison digests |
| W8 | The I/O-edge inventory entry for `internal/plantwire/internal/uds/await.go` keeps its old reason text ("wall-time cap of the plant wire barrier…"), which no longer mentions the S1 pacer that now uses the same clock calls | `test/architecture` (I/O-edge inventory) | Reword the reason (text only; a gate file, so the owner's call) |
| W7 | The run artifact does not record the event-schema registry digest (logged only, H0.5) | PLAN.md 5.1 decision 4 | A versioned manifest change when wanted |

## Round log

| Round | Repository | Commit | Notes |
|---|---|---|---|
| P0 | streams-simulator | 9b95c33a | Plan and goal committed; the old demo canvas (PID 27234) stopped, its workspace left in place; commits at start: streams-simulator `real-run` 9b95c33a, agentic-stream `real-run` 91c2079d, tamoz `real-run` 9170538c, agent-research-lab `main` faf457b; `make build` and `make phase-gate` green on 9b95c33a (354 s; the Enola step reported coupling changes as information only) |
| P0.5 | streams-simulator (evidence) | — | **PASS** on 2026-10-20 at the pre-H0 commits (5951812d, 91c2079d, 9170538c) with `zai/glm-5.3-flash` and a labelled spec variant (prompt states the output contract, objective names the allowed action, budget 300 s / 20000 tokens; spec digest `sha256:5122056a…`, policy `sha256:40cde0fd…`): 61 events, 0 quarantined; 1 situation, decision (`over_temperature`, 0.85), intent `set_indicator alert` (R1, automatic), command executed, LED energized 1000, outcome and verification observed, reconciliation clear; `verify-run` pass; one model call, 17.9 s, 619 in / 852 out tokens. Run folder `.e2e-run/fan-p05-variant-20261020T153832`. The first attempts failed on the unstated output contract (P0.5-FINDINGS.md) |
| P0.7 H0 | agentic-stream `fan-h0` | e2747bf7, b234bbb5, 8f5afa06, 5fc2f120, f03a0f24, 41d57fd9, 6a8ba9cb | H0 complete: characterization pins; catalog-owned routing; strict start check; baseline from spec data; domain-word guard; event schemas from a directory with a root flag. `make ci-check` green (51 s). **After-H0 re-run PASS** (`.e2e-run/fan-p05-afterH0-20261020T160706`): same spec, policy and capability digests as before, same pass criteria, `set_indicator alert` executed, `event_schemas_digest` logged. Gate change (tightening): two slow-test register entries removed |
| S3, S6, S7, S2, S4a | streams-simulator `fan-sim` | a10bf5be, 85890993, 52000ba8, 19d8f099, a3cb4692 | Warm spell 4 °C, `zone_heat_load`, D-116 recorded; D-114 and D-115 fixed; an all-real session announces `hardware`; contract 1.4.0 (BOARD, VERSION, two fixture fixes; DIGEST `sha256:92cdaaa1…`) |
| G1a, G1b, G1c | agent-research-lab `fan-gateway` | 6201c4c, 589b49e, 240883d | Contract vendored with digests; the sans-IO node core; the command pipeline (25 contract exchanges pass in memory) |
| H0.6 | agentic-stream `fan-h0` | be5f237a | H0 review fixes (reviews in `reviews/H0-review-A.md`, `-B.md`, no High): one ownership decision for routing and the start check, schema directory validated at load, wiring tests, `DispatchedIntentTypes`, `byPhase` against writable fields, guard narrowed. Gate changes all tightening. `make ci-check` green |
| S1a, S5a, S4b1 | streams-simulator `fan-sim` | f9a0a5be, a9c35ff1, e29301cd | Paced live sessions (`pace: wall`, no new I/O edge); scripted disturbances in the session plan; semantic kinds on the wire through the board mapping (projection inside the implements check; DCMD translated before the effector; replay unchanged) |
| A0, A1a+A1b, A2 | agentic-stream `fan-a` | e28b1ee4, c39e4800, d497d8f0 (merge of `fan-h0`), b44ff80e | The ADR; contract 1.4.0 vendored, domain and wire, the session in `serve` (handshake, heartbeats, immediate HSYNC, a new session without restart); readings and per-zone heartbeat into the event log through a site file. Against the real streamsim node: two sessions, 658 events each, zero quarantines. `make ci-check` green |
| G1d, G1e, G1f-1, G2, G3a, G3b | agent-research-lab `fan-gateway` | ed44a13, 2f88264, 12e633c, ec54561, 5925033, 167931b | Board adapter, transport and CLI, contract 1.4.0 re-vendored, integration against the real relay (checks 1 to 4) and bench runbook; review fixes (reviews in `reviews/G-review-A.md`, `-B.md`; one High each: poll chains, nested-JSON crash): last-gasp stops, fallbacks, silence detection, caps, state lock, ledger pruning; 27 control mutants all caught. Disconnected-from-boot rule: the kind's `eu_range.min` with quality `disconnected` |
