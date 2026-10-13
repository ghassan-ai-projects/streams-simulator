# Findings register — simulation canvas goal

Every review finding has exactly one row. A gate in `test/architecture` fails when a finding
whose phase is marked Done in [PLAN.md](PLAN.md) is still `open`.

- **Phase**: a phase id from PLAN.md, or `next-goal` (Agentic Stream / gateway goal), or `—`.
- **Status**: `open`, `done` (fixed, evidence in the phase review), `next-goal`, `dropped`
  (removed by an owner decision), `not-a-defect` (evidence in the review record).
- Sources: `REVIEW` = [REVIEW-2026-10-12.md](REVIEW-2026-10-12.md) ids; `R0` = [reviews/P0.md](reviews/P0.md); `R1` = [reviews/P1.md](reviews/P1.md); `R2` = [reviews/P2.md](reviews/P2.md); `R2b` = [reviews/P2b.md](reviews/P2b.md); `R2d` = [reviews/P2d.md](reviews/P2d.md); `R5a` = [reviews/P5a.md](reviews/P5a.md); `R5b` = [reviews/P5b.md](reviews/P5b.md); `R6a` = [reviews/P6a.md](reviews/P6a.md); `R6b` = [reviews/P6b.md](reviews/P6b.md); `R6c` = [reviews/P6c.md](reviews/P6c.md); `R6d` = [reviews/P6d.md](reviews/P6d.md); `R6e` = [reviews/P6e.md](reviews/P6e.md); `R7c` = [reviews/P7c.md](reviews/P7c.md); `R7d` = [reviews/P7d.md](reviews/P7d.md); `R7e` = [reviews/P7e.md](reviews/P7e.md); `R2c` = [reviews/P2c.md](reviews/P2c.md); `R3` = [reviews/P3.md](reviews/P3.md); `self` = found by the implementer.

## Design review (REVIEW-2026-10-12)

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| M1 | REVIEW | High | Canvas design document missing | P0 | done |
| M2 | REVIEW | High | Gateway impact analysis missing | next-goal | next-goal |
| M3 | REVIEW | Med | Cross-repository build order missing | next-goal | next-goal |
| M4 | REVIEW | Med | Deployment topology (Purdue, zones, who dials whom) | next-goal | next-goal |
| M5 | REVIEW | High | Plant wire placement inside the simulator architecture | P7a | done |
| P1 | REVIEW | High | §3 records: node_birth kinds, device_birth instance values, versions, contract digest | P7a | done |
| P2 | REVIEW | High | Historical flag, values in births, `bd_seq`, death synthesis | P7a | done |
| P3 | REVIEW | High | One sequence per direction; lanes are transmit priority only | P7a | done |
| P4 | REVIEW | High | Command freshness, re-check at execution, queue flush, dedup window | P7c | done |
| P5 | REVIEW | High | `authority_epoch` is host/gateway-side only; decide its v1 status | P7a | done |
| P6 | REVIEW | High | Board level owns `boot_id` | P5b | done |
| P7 | REVIEW | High | Field names and refusal codes vs device-wire-v1; PLC refusal codes | P7a | done |
| P8 | REVIEW | High | Policy example and birth-declared safe-stop/risk class | P7d | done |
| P9 | REVIEW | High | Commissioning state machine and metadata changes (host side) | next-goal | next-goal |
| P10 | REVIEW | High | One quality-code mapping table; manual/out-of-service codes | P7a | done |
| P11 | REVIEW | High | Prompt injection through strings; injection fixtures | P7a | done |
| P12 | REVIEW | High | Quiescence barrier and clock mode with a live host | P7d | done |
| P13 | REVIEW | Med | Edge realism: time source, scan class, polling failures, buffering | P7a | done |
| P14 | REVIEW | Med | `host_state` offline semantics, version, heartbeat | P7b | done |
| P15 | REVIEW | Med | Capability digest definition | P7a | done |
| P16 | REVIEW | Med | `notify` alert event type and path to the session | next-goal | next-goal |
| P17 | REVIEW | Med | Contract sync supply chain: pinned digest, safety floor, minimum version | P7a | done |
| P18 | REVIEW | Med | Non-negotiables mapped to the wire: reference host, lying results | P7d | done |
| P19 | REVIEW | Med | Fault layer: existing vs new protocol faults (drop, duplicate, swap, garble done in P7c; split frames, reset mid-command and reconnect storm need reconnection) | P13 | open |
| P20 | REVIEW | Med | Gateway as a safety-relevant component | next-goal | next-goal |
| P21 | REVIEW | Med | `state` is self-reported; independent feedback for verification | P7a | done |
| P22 | REVIEW | Low | `device_death` "sensor outage" wording | P7a | done |
| P23 | REVIEW | Low | Batching, limits, session-boundary sequence, data-model fields | P7a | done |
| P24 | REVIEW | Low | Experiment naming ambiguity | P7a | done |
| S1 | REVIEW | High | Remove kill switch from the scenario; T1 as protection demands and margin | P12 | open |
| S2 | REVIEW | High | Supervisory-level actions; plant sequencer in the world | P6d | done |
| S3 | REVIEW | High | Fault ids collide with pump ids | P12 | open |
| S4 | REVIEW | High | Channel names as kind.metric; missing sensors; channel count | P12 | open |
| S5 | REVIEW | High | Cooling signatures and pump physics | P12 | open |
| S6 | REVIEW | High | Responders leak truth; availability; one approval path; work orders | P12 | open |
| S7 | REVIEW | High | Scoring gameability and statistics | next-goal | next-goal |
| S8 | REVIEW | High | Baselines and held-out discipline (D7) | next-goal | next-goal |
| S9 | REVIEW | Med | Demand charge block alignment, `peak_mtd`, ratchet | P6b | done |
| S10 | REVIEW | Med | Air pressure band, sequencer, generator transfer | P12 | open |
| S11 | REVIEW | Med | Time resolution and event clock | P6e | done |
| S12 | REVIEW | Med | Alarm management (priorities, ack, shelving, combined flood) | P12 | open |
| S13 | REVIEW | Med | Cause-code vocabulary authored on the host side; rotating decoys | P12 | open |
| S14 | REVIEW | Med | Vocabulary: `kind` row, `lockout` naming, closed comparators, crossing times | P6c | done |
| S15 | REVIEW | Med | DECISIONS D-03/D-04/D-14 revision recorded | P5a | done |
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
| R0-01 | R0 | High | Evaluate canary test over the journal (bridge and page: R0-01b) | P3 | done |
| R0-01b | R0 | High | Evaluate canary test over the bridge payloads and the page (the journal part of R0-01 is proven in P3) | P10 | open |
| R0-02 | R0 | High | `.mcp.json` exposure must not give coding agents unblind or reveal | P10 | open |
| R0-03 | R0 | High | `scenario from-run` export with round-trip digest test | P2 | done |
| R0-04 | R0 | High | Device-path gate and pinned `device serve` transcript (acceptance test golden) | P1 | done |
| R0-05 | R0 | High | Plant-wire spec states coexistence with the device path | P7a | done |
| R0-06 | R0 | Med | CLI output contract: default JSON, exempt commands, deliberate pin change | P1 | done |
| R0-07 | R0 | Med | `godox` and `nolintlint` in the Go lint set | P1 | done |
| R0-08 | R0 | Med | `make phase-gate` and evidence files | P1 | done |
| R0-09 | R0 | Med | Consumer-knowledge gate over Go and canvas sources | P1 | done |
| R0-10 | R0 | Med | `streamsim domain format` canonical formatter | P5a | done |
| R0-11 | R0 | Med | Closed-grammar test over the v0.2 schema | P6a | done |
| R0-12 | R0 | Med | RNG draw-order table for the event clock | P6e | done |
| R0-13 | R0 | Med | Contract sync target, pinned digest, safety-field floor | P7a | done |
| R0-14 | R0 | Med | Technology decision record (UI, bridge language and protocol): [canvas-technology.md](canvas-technology.md) | P8a | done |
| R0-15 | R0 | Med | Bridge threat model with one test per threat | P8b | open |
| R0-16 | R0 | Med | Stress fixture | P8b | open |
| R0-17 | R0 | Med | Numeric performance budgets enforced in CI | P9 | open |
| R0-18 | R0 | Med | Playback with no binary; crashed-run and failing-verify badges | P9 | open |
| R0-19 | R0 | Med | Action-registry parity test | P10 | open |
| R0-20 | R0 | Med | Thermal twin compared with a recorded bench run | P12 | open |
| R0-21 | R0 | Low | Terminology: scenario file, journal, in `UBIQUITOUS_LANGUAGE.md` | P2 | done |
| R0-22 | R0 | Low | Report includes consumer digests only when recorded | P4 | done |
| R0-23 | R0 | Med | Status banners on design text superseded by owner decisions | P0 | done |

## P1 review: work carried into later phases

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| R1-01 | R1 | Low | A "missing required property" issue points at the parent object; point at the missing member | P5a | done |
| R1-02 | R1 | Med | An unknown `--perturb` name is `run_failed` while malformed `--fault`/`--effector` are `usage_invalid`; unify through the scenario checks | P2 | done |
| R1-03 | R1 | Med | `run` and the director accept a profile the domain does not declare; validate in the run module for every surface (the CLI check added in P1 then moves there) | P2 | done |

## P2 review: work carried into later phases

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| R2-01 | R2 | Med | No file the CLI or director reads has a size bound (scenario, domain, adapter, artifact, verdict, label); bound every input read, with a typed failure, and test each | P8b | open |
| R2-02 | R2 | Low | Explicit `entity_ids` (scenario `world.entity_ids`, director `sim.world.create`) are not checked against the domain's entity count bounds or id template; decide and enforce with the v0.2 kinds-and-instances model | P5a | done |
| R2-03 | R2 | Low | Errors from a missing `--domains-dir` repeat the `streamsim:` prefix three times (pre-existing in `loadSimulatorInputs`); wrap once; the fix is the same as R2c-02, so it is carried there | P13 | open |
| R2-04 | R2 | Low | Perturbation windows with negative or pre-start times are accepted (negative `until` acts as "no end"); and the run-wide limits have no MCP-level test | P13 | open |
| R2-05 | R2 | Low | The effector idempotency key compares arguments by `json.Marshal` text, so `3` and `3.0` are different requests | P6a | done |
| R2b-01 | R2b | Low | The operator maps every invocation error to `effector_refused`; MCP_SURFACE lists `interlock_refused` and `invalid_args`, which are never emitted. Decide the taxonomy without revealing hidden outcomes, and add a protocol-level (wire) test of the acknowledgement | P10 | open |
| R2d-01 | R2d | Med | Director composition tests (`closed_loop_test.go`, `operator_surface_test.go`) import `mcp`, inverting layers in test code; move them to `test/acceptance` with a test-support seam instead of `export_test.go` `RunOf` | P3 | done |
| R2d-02 | R2d | Low | The director returns wire-shaped `map[string]any` results and messages naming tool arguments; return typed results and let `mcp/protocol` shape the wire (also removes protocol knowing result keys, and narrows the `wrapcheck` exemption for the director facade) | P10 | open |
| R2c-01 | R2c | Low | Each module that reads or writes files repeats the same `Read`/`Write` wrapper (`fmt.Errorf("%w", err)`, mode `0o600`, `#nosec`) in its own `files` edge (cli, grading, release, refconsumer, suite, sink, adapter, domain); decide whether one shared file edge in a foundation is worth the coupling, or record the duplication as the price of independent edges | P13 | open |
| R2c-02 | R2c | Low | Modules emit the CLI's `streamsim: ` error prefix themselves (`domain`, `run`, `grading`, `suite`, `refconsumer`, `device`), so library errors carry a command-line concern. R2-03 covers the doubled prefix on a missing `--domains-dir`; the fix for both is the same: modules return unprefixed errors and `cli` adds the prefix once, with the failure-document and stderr text updated together | P13 | open |

## P3 review: work carried into later phases

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| R3-01 | R3 | Med | Consumer-readable files (`journal.jsonl`, `trace.jsonl`) share a run directory with director-only files (`ledger.jsonl`, `world_state_history.jsonl`, `run.json` with the command log); separate them (for example an `observer/` folder) and have the bridge and the director hand out only observer paths | P10 | open |
| R3-02 | R3 | Low | Setting `incomplete: true` in `run.json` skips the journal comparison in `verify`; at least report that the journals were not compared, or compare up to the last replayable record | P9 | open |
| R3-03 | R3 | Low | Session records (`consumer.reported`, `pacing`) are outside the stream digest and unverified; validate them against the contract and never let scoring trust them | P10 | open |

## Found during implementation

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| S-01 | self | High | `sim.effector.invoke` returns the world's internal result (failure `mode`, `effect_applied`, latency) to the consumer, exposing `silent_no_effect` and other hidden outcomes (MCP_SURFACE §4.3, §4.5) | P2b | done |

## P5a review: work carried into later phases

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| R5a-01 | R5a | Med | v0.2 allows one kind; several kinds and per-instance overrides need a world change (per-kind runtime set, kind reference per entity, namespacing, solver and cross-check over kinds, draw order) | P5a2 | done |
| R5a-02 | R5a | Med | The explicit entity-id check matches every id against one `id_template`; named instances must be accepted (id in the set's names or matching its template) | P5b | done |

## P5b review: work carried into later phases

| ID | Source | Sev | Summary | Phase | Status |
|---|---|---|---|---|---|
| R5b-01 | R5b | Low | Peer sigma pooling is by kind: a `peer_residual` detector on a channel added by an override pools peers that lack the channel | P6a | done |
| R5b-02 | R5b | Low | Narrowing validation gaps: two narrow entries naming the same state, channel or argument are not refused; an argument `minimum` above the declared `maximum` (or the reverse) is not refused | P6a | done |
| R5b-03 | R5b | Low | A profile's `entity_count` is ignored when the first instance set has names only | P6a | done |
| R6a-01 | R6a | Low | Store stiffness and flow sign are unchecked: nothing bounds `conductance / capacity × dt` (RK4 diverges near 2.8) and a negative scale state reverses a flow; validate the bound and keep health scales non-negative | P12 | open |
| R6b-01 | self | Low | Calendar billing periods (months of unequal length) and the projected demand of the open block are not modelled; `peak` uses a fixed `period_s` | P12 | open |
| R6b-02 | R6b | Low | A read of an integrated state further in the past than the entity's longest step returns the current value (the snapshot history covers one longest step); runs read forward, so no output is affected; return a located error or keep a longer history | P13 | open |
| R6c-01 | R6c | Low | Oracle and audit worlds apply profile setup calls before injecting faults, so a setup call dated after a fault onset is checked (interlock, preconditions) without the fault; inject faults first or refuse setup dated after the onset | P12 | open |
| R6c-02 | R6c | Low | A precondition refusal of a safe-stop effector becomes `ErrPlantUnavailable` in the device bridge (`completeSafeStop`); decide whether a safe-stop effector may carry `requires` | P7c | done |
| R6e-01 | self | Med | Aggregation of dynamics inputs over typed links (`sum`, `mean`, vocabulary table §7) is in no phase: links are structure only; needed before the reference scenario models a shared header or bus across instances | P7h | open |
| R6e-02 | R6e | Low | The stochastic fault envelope stream is keyed by the fault declaration, not the injection, so injections of one stochastic fault on several entities share one stream in read order; key it by injection (a determinism-tuple change, pinned outputs move) | P7f | done |
| R6e-03 | R6e | Low | A direct read of an integrated state ahead of the clock (outside the look-ahead copy) commits integration and changes later emissions; production reads go through the look-ahead, so make the facade's StateValue use it too or refuse future reads | P13 | open |
| R7c-01 | R7c | Med | `fallback_on_lease_loss` is validated but not acted on: a simulated node ends its session on host loss (the run stops), and the domain's own effector leases lapse in the world; a node that keeps running after host loss (reconnection) must apply each live command's fallback | P13 | open |
| R7c-02 | R7c | Low | Wire-fault `swap` acts within one transmitted batch, so a swap on a batch's last record (a lone receipt, an `NSYNC`) is a no-op; apply faults over a persistent outbound buffer | P13 | open |
| R7d-01 | R7d | Med | The node never sends `DDEATH` or quality codes beyond `good` and `warming_up`, and models no reboot (a board's `boot_id` is constant): the world has no device retirement, sensor quality or reboot event to map from (`model.SimEvent` carries no quality); the contract's `reboot-reconcile` and `queued-across-reboot` exchanges run against the host only | P13 | open |
| R7d-02 | R7d | Low | Host loss is recorded only as three `missed` pacing records and the node's error; SPEC §8 also asks for a command-log entry, and no `NDEATH connection_lost` is synthesised: both come with reconnection | P13 | open |
| R7d-03 | R7d | Low | SPEC §12 limits are enforced on receipt (64 KiB, 1 MiB for births, 4096 records per barrier) but not checked on send (500 metrics per `DDATA`, kinds, boards and devices per `NBIRTH`) | P13 | open |
| R7d-04 | R7d | Low | The node checks the socket and its directory by path and does not verify the peer's credentials after dialling (`LOCAL_PEERCRED`/`SO_PEERCRED` need a platform package); the owner-only directory is the protection | next-goal | next-goal |
| R7d-05 | R7d | Low | The reference host sends one `STATE` and no heartbeats; a simulated node needs none (the barrier is its liveness), a hardware node will | next-goal | next-goal |
| R7d-06 | R7d | Low | `DSTATE` reports `safe_state: false`, no outputs and quality `good`, and a `DQUERY` for an unknown device is not answered | P13 | open |
| R7e-01 | R7e | High | World: snapshot history keeps only the longest step, so a chain of states with differing steps (1 s → 60 s → 45 s) reads trimmed history and depends on the read schedule (up to 0.48 off an independent recursion); `at()` falls back silently | P7f | done |
| R7e-02 | R7e | High | World: a stochastic fault walk lives on the active fault with one RNG, so a dated effector or boundary check, or a frequent channel read, changes earlier readings; the look-ahead copy does not cover it (also R6e-02) | P7f | done |
| R7e-03 | R7e | High | World: at one instant the first-issued assignment wins (strict `>`), and a lease renewed exactly at its lapse is overridden by the old fallback kick | P7f | done |
| R7e-04 | R7e | Med | World: an assign effect with a time constant starts from 0, not from the current value; one-sided state bounds clamp to 0; dt/τ stability is unvalidated for lag, dead time, rc network and regime; overlapping id templates pass validation; a negative lease is accepted | P7f | open |
| R7e-05 | R7e | Med | World: the P6a integration fix changed faulted hidden-state numbers (~1e-9) without a SimVersion bump; the behaviour pin is fault-free and short; truth labels a pulse observable on a channel that cannot record it; some oracles reuse production code (dead time, peak) | P7f | open |
| R7e-06 | R7e | High | Run records: a reused run folder appended to `ledger.jsonl` (doubled rows); `verify`, `report` and `score` trust `ledger.jsonl`, the history, `verdict.json` and `label.json`, which no digest covers | P7f | done |
| R7e-07 | R7e | High | Quiescence: a consumer can report quiescence through any time and switch off every later barrier; the watermark is not clamped to the world clock | P7f | done |
| R7e-08 | R7e | Med | Run records: preamble and postamble lines are in the trace but not the observer journal; a torn ledger fails the report of a crashed run; a scenario may inject a fault whose onset is before the clock; the Markdown report prints verdict strings unescaped; session records are outside every digest | P7f | open |
| R7e-09 | R7e | Med | Truth: the nameplate and `sim.effector.list` pass effector descriptions (scoring and fault hints) and full argument schemas to the operator | P7f | open |
| R7e-10 | R7e | Med | Truth: the report shows truth whenever a `run.json` exists, without checking it belongs to the run of the journals | P7f | open |
| R7e-11 | R7e | Med | Tests: no test that a met barrier resets the missed count, that the quiescence watermark never moves back, that a command with a higher epoch is refused, or for a perturbation window end; the host fuzz target is not in `make fuzz` | P7f | open |
| R7e-12 | R7e | Med | Findings phases (done in P7e: R3-02 to P9, R2-01 to P8b, R6a-01 to P12): R3-02 belongs to P9 (the Verified badge), R2-01 to P8b (the bridge reads run files), R6a-01 to P12 (factory stores) | P7e | done |
| R7e-13 | R7e | Low | `TestDirectorServesItsToolsUntilTheSessionEndsThenStopsTheOperatorEndpoint` failed once under a full parallel run (the operator endpoint still accepted a connection right after the session ended); it passes in isolation | P7f | open |
