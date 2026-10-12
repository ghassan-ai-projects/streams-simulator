# Findings register — simulation canvas goal

Every review finding has exactly one row. A gate in `test/architecture` fails when a finding
whose phase is marked Done in [PLAN.md](PLAN.md) is still `open`.

- **Phase**: a phase id from PLAN.md, or `next-goal` (Agentic Stream / gateway goal), or `—`.
- **Status**: `open`, `done` (fixed, evidence in the phase review), `next-goal`, `dropped`
  (removed by an owner decision), `not-a-defect` (evidence in the review record).
- Sources: `REVIEW` = [REVIEW-2026-10-12.md](REVIEW-2026-10-12.md) ids; `R0` = [reviews/P0.md](reviews/P0.md); `R1` = [reviews/P1.md](reviews/P1.md); `R2` = [reviews/P2.md](reviews/P2.md); `R2b` = [reviews/P2b.md](reviews/P2b.md); `R2c` = [reviews/P2c.md](reviews/P2c.md); `self` = found by the implementer.

## Design review (REVIEW-2026-10-12)

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| M1 | REVIEW | High | Canvas design document missing | P0 | done |
| M2 | REVIEW | High | Gateway impact analysis missing | next-goal | next-goal |
| M3 | REVIEW | Med | Cross-repository build order missing | next-goal | next-goal |
| M4 | REVIEW | Med | Deployment topology (Purdue, zones, who dials whom) | next-goal | next-goal |
| M5 | REVIEW | High | Plant wire placement inside the simulator architecture | P7a | open |
| P1 | REVIEW | High | §3 records: node_birth kinds, device_birth instance values, versions, contract digest | P7a | open |
| P2 | REVIEW | High | Historical flag, values in births, `bd_seq`, death synthesis | P7a | open |
| P3 | REVIEW | High | One sequence per direction; lanes are transmit priority only | P7a | open |
| P4 | REVIEW | High | Command freshness, re-check at execution, queue flush, dedup window | P7a | open |
| P5 | REVIEW | High | `authority_epoch` is host/gateway-side only; decide its v1 status | P7a | open |
| P6 | REVIEW | High | Board level owns `boot_id` | P5b | open |
| P7 | REVIEW | High | Field names and refusal codes vs device-wire-v1; PLC refusal codes | P7a | open |
| P8 | REVIEW | High | Policy example and birth-declared safe-stop/risk class | P7a | open |
| P9 | REVIEW | High | Commissioning state machine and metadata changes (host side) | next-goal | next-goal |
| P10 | REVIEW | High | One quality-code mapping table; manual/out-of-service codes | P7a | open |
| P11 | REVIEW | High | Prompt injection through strings; injection fixtures | P7a | open |
| P12 | REVIEW | High | Quiescence barrier and clock mode with a live host | P7d | open |
| P13 | REVIEW | Med | Edge realism: time source, scan class, polling failures, buffering | P7a | open |
| P14 | REVIEW | Med | `host_state` offline semantics, version, heartbeat | P7a | open |
| P15 | REVIEW | Med | Capability digest definition | P7a | open |
| P16 | REVIEW | Med | `notify` alert event type and path to the session | next-goal | next-goal |
| P17 | REVIEW | Med | Contract sync supply chain: pinned digest, safety floor, minimum version | P7a | open |
| P18 | REVIEW | Med | Non-negotiables mapped to the wire: reference host, lying results | P7d | open |
| P19 | REVIEW | Med | Fault layer: existing vs new protocol faults | P7c | open |
| P20 | REVIEW | Med | Gateway as a safety-relevant component | next-goal | next-goal |
| P21 | REVIEW | Med | `state` is self-reported; independent feedback for verification | P7a | open |
| P22 | REVIEW | Low | `device_death` "sensor outage" wording | P7a | open |
| P23 | REVIEW | Low | Batching, limits, session-boundary sequence, data-model fields | P7a | open |
| P24 | REVIEW | Low | Experiment naming ambiguity | P7a | open |
| S1 | REVIEW | High | Remove kill switch from the scenario; T1 as protection demands and margin | P12 | open |
| S2 | REVIEW | High | Supervisory-level actions; plant sequencer in the world | P6d | open |
| S3 | REVIEW | High | Fault ids collide with pump ids | P12 | open |
| S4 | REVIEW | High | Channel names as kind.metric; missing sensors; channel count | P12 | open |
| S5 | REVIEW | High | Cooling signatures and pump physics | P12 | open |
| S6 | REVIEW | High | Responders leak truth; availability; one approval path; work orders | P12 | open |
| S7 | REVIEW | High | Scoring gameability and statistics | next-goal | next-goal |
| S8 | REVIEW | High | Baselines and held-out discipline (D7) | next-goal | next-goal |
| S9 | REVIEW | Med | Demand charge block alignment, `peak_mtd`, ratchet | P6b | open |
| S10 | REVIEW | Med | Air pressure band, sequencer, generator transfer | P12 | open |
| S11 | REVIEW | Med | Time resolution and event clock | P6e | open |
| S12 | REVIEW | Med | Alarm management (priorities, ack, shelving, combined flood) | P12 | open |
| S13 | REVIEW | Med | Cause-code vocabulary authored on the host side; rotating decoys | P12 | open |
| S14 | REVIEW | Med | Vocabulary: `kind` row, `lockout` naming, closed comparators, crossing times | P6a | open |
| S15 | REVIEW | Med | DECISIONS D-03/D-04/D-14 revision recorded | P5a | open |
| S16 | REVIEW | Med | Negative-class operations events | P12 | open |
| S17 | REVIEW | Low | Synthetic boundaries labelled; parameters sourced | P12 | open |
| A1 | REVIEW | High | Thermal example, replays and Tamoz fixtures rewrite | next-goal | next-goal |
| A2 | REVIEW | High | Agentic Stream replay is report-only; sources depend on ReplaySimulator | next-goal | next-goal |
| A3 | REVIEW | Med | SituationSpec derived features and unit checks | next-goal | next-goal |
| A4 | REVIEW | Med | Kill-switch and revocation text in the impact analyses | next-goal | next-goal |
| A5 | REVIEW | Med | Tamoz memory and improvement are per customer | next-goal | next-goal |
| A6 | REVIEW | Med | One-system scoring vs Tamoz protocol | next-goal | next-goal |
| A7 | REVIEW | Low | Route cross-check against recorded births | next-goal | next-goal |
| I1 | REVIEW | High | Isolation beyond GitHub (provider logs, transcripts, CI, dumps) | next-goal | next-goal |
| I2 | REVIEW | High | Improvement-loop review gate | next-goal | next-goal |
| I3 | REVIEW | Med | Customer workspace marker, ignore rules and check | P8a | open |
| X1 | REVIEW §8 | Low | Observation scoping from idea.md | P13 | open |
| X2 | REVIEW §8 | Low | "Tamoz plus one independent agent" from idea.md | P13 | open |
| X3 | REVIEW §8 | Low | Idea's 30 °C / 55 °C case mapping | P13 | open |
| D6 | REVIEW §2 | Med | LLM hosting and provider terms documented as an open decision | P13 | open |

## P0 review: work carried into phases

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| R0-01 | R0 | High | Evaluate canary test over journal, bridge and page | P3 | open |
| R0-02 | R0 | High | `.mcp.json` exposure must not give coding agents unblind or reveal | P10 | open |
| R0-03 | R0 | High | `scenario from-run` export with round-trip digest test | P2 | done |
| R0-04 | R0 | High | Device-path gate and pinned `device serve` transcript (acceptance test golden) | P1 | done |
| R0-05 | R0 | High | Plant-wire spec states coexistence with the device path | P7a | open |
| R0-06 | R0 | Med | CLI output contract: default JSON, exempt commands, deliberate pin change | P1 | done |
| R0-07 | R0 | Med | `godox` and `nolintlint` in the Go lint set | P1 | done |
| R0-08 | R0 | Med | `make phase-gate` and evidence files | P1 | done |
| R0-09 | R0 | Med | Consumer-knowledge gate over Go and canvas sources | P1 | done |
| R0-10 | R0 | Med | `streamsim domain format` canonical formatter | P5a | open |
| R0-11 | R0 | Med | Closed-grammar test over the v0.2 schema | P6a | open |
| R0-12 | R0 | Med | RNG draw-order table for the event clock | P6e | open |
| R0-13 | R0 | Med | Contract sync target, pinned digest, safety-field floor | P7a | open |
| R0-14 | R0 | Med | Technology decision record (UI, bridge language and protocol) | P8a | open |
| R0-15 | R0 | Med | Bridge threat model with one test per threat | P8b | open |
| R0-16 | R0 | Med | Stress fixture | P8b | open |
| R0-17 | R0 | Med | Numeric performance budgets enforced in CI | P9 | open |
| R0-18 | R0 | Med | Playback with no binary; crashed-run and failing-verify badges | P9 | open |
| R0-19 | R0 | Med | Action-registry parity test | P10 | open |
| R0-20 | R0 | Med | Thermal twin compared with a recorded bench run | P12 | open |
| R0-21 | R0 | Low | Terminology: scenario file, journal, in `UBIQUITOUS_LANGUAGE.md` | P2 | done |
| R0-22 | R0 | Low | Report includes consumer digests only when recorded | P4 | open |
| R0-23 | R0 | Med | Status banners on design text superseded by owner decisions | P0 | done |

## P1 review: work carried into later phases

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| R1-01 | R1 | Low | A "missing required property" issue points at the parent object; point at the missing member | P5a | open |
| R1-02 | R1 | Med | An unknown `--perturb` name is `run_failed` while malformed `--fault`/`--effector` are `usage_invalid`; unify through the scenario checks | P2 | done |
| R1-03 | R1 | Med | `run` and the director accept a profile the domain does not declare; validate in the run module for every surface (the CLI check added in P1 then moves there) | P2 | done |

## P2 review: work carried into later phases

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| R2-01 | R2 | Med | No file the CLI or director reads has a size bound (scenario, domain, adapter, artifact, verdict, label); bound every input read, with a typed failure, and test each | P13 | open |
| R2-02 | R2 | Low | Explicit `entity_ids` (scenario `world.entity_ids`, director `sim.world.create`) are not checked against the domain's entity count bounds or id template; decide and enforce with the v0.2 kinds-and-instances model | P5a | open |
| R2-03 | R2 | Low | Errors from a missing `--domains-dir` repeat the `streamsim:` prefix three times (pre-existing in `loadSimulatorInputs`); wrap once | P4 | open |
| R2-04 | R2 | Low | Perturbation windows with negative or pre-start times are accepted (negative `until` acts as "no end"); and the run-wide limits have no MCP-level test | P13 | open |
| R2-05 | R2 | Low | The effector idempotency key compares arguments by `json.Marshal` text, so `3` and `3.0` are different requests | P6a | open |
| R2b-01 | R2b | Low | The operator maps every invocation error to `effector_refused`; MCP_SURFACE lists `interlock_refused` and `invalid_args`, which are never emitted. Decide the taxonomy without revealing hidden outcomes, and add a protocol-level (wire) test of the acknowledgement | P10 | open |
| R2c-01 | R2c | Low | Each module that reads or writes files repeats the same `Read`/`Write` wrapper (`fmt.Errorf("%w", err)`, mode `0o600`, `#nosec`) in its own `files` edge (cli, grading, release, refconsumer, suite, sink, adapter, domain); decide whether one shared file edge in a foundation is worth the coupling, or record the duplication as the price of independent edges | P13 | open |
| R2c-02 | R2c | Low | Modules emit the CLI's `streamsim: ` error prefix themselves (`domain`, `run`, `grading`, `suite`, `refconsumer`, `device`), so library errors carry a command-line concern. R2-03 covers the doubled prefix on a missing `--domains-dir`; the fix for both is the same: modules return unprefixed errors and `cli` adds the prefix once, with the failure-document and stderr text updated together | P13 | open |

## Found during implementation

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| S-01 | self | High | `sim.effector.invoke` returns the world's internal result (failure `mode`, `effect_applied`, latency) to the consumer, exposing `silent_no_effect` and other hidden outcomes (MCP_SURFACE §4.3, §4.5) | P2b | done |
