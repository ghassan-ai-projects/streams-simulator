# Plan — Simulation canvas goal

- **Goal**: [GOAL.md](GOAL.md) · **Bar**: [QUALITY_BAR.md](QUALITY_BAR.md) · **Design**: [canvas-design.md](canvas-design.md) · **Findings**: [FINDINGS.md](FINDINGS.md)
- **Branch**: `simulation-canvas`
- **Decisions in force**: REVIEW-2026-10-12 §1 and D1–D7 (accepted as recommended on
  2026-10-12); no backward compatibility (pre-release); the device path stays intact and gated.
- **Preflight (2026-10-12)**: DEFERRED D-08, D-15 and D-17 (director/run concurrency) are fixed
  with regression tests by the hardening program (`docs/refactoring/hardening-20261010/PLAN.md`);
  the register in `DEFERRED.md` is stale and is not relied on.

Order: simulator foundations the canvas consumes (P1–P4), then the data model (P5–P6), then
the plant wire (P7), then the canvas (P8–P11), then reference worlds and hardening (P12–P13).
The canvas starts after the structures it displays exist, so no view is built twice.

## Phases

| # | Phase | Spec step | Delivers | Exit criteria (in addition to the standing gates) |
|---|---|---|---|---|
| P0 | Goal, bar, plan | — | GOAL, QUALITY_BAR, PLAN, FINDINGS with a disposition for every review item, findings gate, P0 review record, status banners on superseded design text | Two reviews resolved; findings gate green |
| P1 | Structured CLI output | `docs/design/CLI_SURFACE.md` and `cli-output-v0.1`: success documents per command, one error document shape with stable codes and issues (JSON pointer when known), exit codes, exempt long-running commands (`mcp`, `device serve`), the deliberate stdout change | Every non-exempt command prints exactly one schema-valid JSON document on stdout for success and failure; typed validation errors in the `domain` and `adapter` modules; `godox` and `nolintlint`; `make phase-gate`; device-serve transcript and new stdout in the behaviour pin; consumer-knowledge gate | Every command and every error class tested against the schema; pin diff recorded as deliberate |
| P2 | Scenario file | `scenario-v0.1` and semantics; terminology in `UBIQUITOUS_LANGUAGE.md` (scenario file vs profile vs suite scenario) | `streamsim run --scenario <file>`; flags and file build the same run; `streamsim scenario from-run <run.json>` exports a run's commands as a scenario file; fuzz target | Equivalence test (flags vs file); round trip run → export → run gives an identical command-log digest |
| P3 | Run journal | Journal spec and `run-journal-v0.1`: every record type tagged truth-bearing or not; who may read an in-progress run folder; sealed truth stream; unblinded (Explore) stream; pacing and lifecycle records; a reserved, versioned wire-exchange record type that P7 fills; crash behaviour | Journal appended during CLI and director runs; truth-bearing records only in the sealed stream until sealing; `verify` checks the journal against replay; fuzz target for the journal reader | Journal identical across replays; partial journal after an injected crash valid to its end; **canary test**: a distinctive injected fault in an Evaluate run appears in no consumer-readable file before sealing; Explore run stamped unblinded and refused by scoring |
| P4 | Report | `run-report-v0.1` and Markdown rules; consumer-side digests included only when the run recorded them | `streamsim report <run-dir>` (JSON and Markdown) | Pure function of the run folder; schema-valid; golden tests |
| P5a | Domain-spec v0.2: kinds and instances | v0.2 semantics part 1; `DECISIONS.md` deviation entries (incl. D-03, D-04, D-14 where affected) | Schema, loader, compiler and digest for kinds and instances compiling to the runtime world; the eight domains rewritten to v0.2; `streamsim domain format` (canonical formatting); manifest, goldens and schema suite updated | Every rewritten domain compiles to a runtime world identical to its v0.1 one (oracles, determinism, replay unchanged); every validation issue carries a JSON pointer |
| P5b | v0.2: overrides, links, system entities, board | v0.2 semantics part 1 (cont.) | Add-or-narrow overrides; typed links; system entities; board (`boot_id` owner) | Override, link and board rules proven at the domain layer; device path gate green |
| P6a | Vocabulary: `store`, `flow` | v0.2 semantics part 2 for these forms | Both forms with an independent second implementation | Analytic cross-check across swept parameters; closed-grammar test (no free-text expression anywhere in the schema) |
| P6b | Vocabulary: `regime`, `schedule`, block `peak` (`peak_mtd`, ratchet) | Same | Same | Crossing times derived independently and agreeing; same |
| P6c | Vocabulary: `boundary`, `trip`, latched `lockout`, `anti_recycle`, LOTO isolation | Same | Same | Same |
| P6d | Vocabulary: `precondition`, `responder` (closed comparator set), `fallback_on_lease_loss`, plant sequencer entity | Same | Same | Same; no expression language |
| P6e | Event clock | Event-clock semantics with an RNG draw-order table | Sub-step events (trips, short sags) on an event clock | v0.1-equivalent worlds bit-identical; draw order proven |
| P7a | Plant wire: spec and contract | Plant-wire spec v2 applying REVIEW §4 (each item: node, reference host or next goal), D1–D5, M5 placement, the pipeline amendment (perturbation after sequencing), quiescence barrier and clock mode, `authority_epoch` decision, coexistence with the device path | `contracts/plant-wire/` schemas and fixtures (valid, invalid, scripted exchanges, lying results per command class, injection strings, must-ignore); `make contract-sync`, pinned digest, safety-field floor test | Contract reviewed; fixtures validate against the schemas |
| P7b | Plant wire: node sessions | — | Node module: node dials host (D2), handshake (version range, contract digest, environment marker), births from kinds with current values, one sequence per direction, lanes, `bd_seq`, heartbeat, rebirth; private socket | Session fixtures pass; owner-only socket; host refuses an environment mismatch |
| P7c | Plant wire: commands | — | Commands, receipts and results, freshness and dedup rules, perturbation placement, node-side digest gate for the bench (D4) | Command fixtures pass; replay reproduces the wire exchange |
| P7d | Plant wire: reference host and journal | — | Minimal conformant reference host; wire records in the journal; quiescence barrier | A slow reference host is not outrun; replay identical; device path gate green |
| P8a | Canvas foundation | Technology decision record (UI stack, bridge language, bridge protocol), `canvas/README.md` architecture | `canvas/` toolchain and every QUALITY_BAR §3 gate on an empty shell; `make canvas-check`; canvas CI job; Go isolation gate; `AGENTS.md` canvas section; customer-workspace marker, ignore rules and check (REVIEW I3) | All gates green on the shell |
| P8b | Bridge | Bridge protocol and threat model | Bridge: director MCP client, workspace-scoped file service with atomic writes and conflict detection, live stream, token and Origin/Host checks; e2e harness running the real binary; stress fixture | One test per threat; first e2e passes against the real binary |
| P9 | Playback and review | Layout and connections schemas; performance budgets | Workspace open, run list, plant view (kinds, instances, links), timeline, playback with seek, report view, verification badge, true vs reported after sealing, reload on file change | e2e with no `streamsim` on PATH plays back a run; crashed-run and failing-verify fixtures show the right badges; budgets enforced |
| P10 | Live run and control | Bridge command set, live-state model, action registry, `.mcp.json` exposure decision (which role, no unblind for coding agents) | Director-driven runs; play, pause, step, speed, reset; live view with wire traffic and party status from session records; fault, environment and perturbation panel; Explore / Evaluate; export session via `scenario from-run`; `.mcp.json` | Parity test: every bridge action maps to a documented CLI or MCP command; Evaluate canary through bridge and page; registered MCP server cannot reveal truth on an open run |
| P11a | Builder: overview, kinds, state, behaviour | Builder interaction spec (views, panels, honesty rule) | Kind view and these panels, validate as you edit | e2e: edits produce binary-valid domains |
| P11b | Builder: sensors, actions, faults | — | These panels | Same |
| P11c | Builder: scenarios and tools | — | Scenarios panel, preview run, fault-quality check, templates, diff before save, layout sidecar | e2e: a world built in the canvas is byte-identical to the authored JSON after `domain format` |
| P12 | Reference worlds | Reference scenario v2 (REVIEW §5 items in this goal, D3) | Factory utilities domain; thermal bench twin domain | Both validate, preview and run over the plant wire with the reference host; the canvas shows wire traffic, session status and environment marker; twin event stream compared with a recorded bench run; factory delivered as "validated and run, headroom not measured" |
| P13 | Hardening and documentation | — | Accessibility audit, final budgets, public documentation in `documentation/`, `limitations.md`, D6 open-decision note, idea.md dispositions (REVIEW §8) | Every GOAL item demonstrated; no open finding for this goal |

## Progress log

The commit hash of a phase is recorded by the first commit of the next phase.

| Phase | Status | Commit | Evidence | Review |
|---|---|---|---|---|
| P0 | Done | (recorded in P1) | [evidence/P0.md](evidence/P0.md) | [reviews/P0.md](reviews/P0.md) |
