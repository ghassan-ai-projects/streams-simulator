# Findings register — simulation canvas goal

Every review finding has exactly one row. A gate in `test/architecture` fails when a finding
whose phase is marked Done in [PLAN.md](CANVAS_PLAN.md) is still `open`.

- **Phase**: a phase id from PLAN.md, or `next-goal` (Agentic Stream / gateway goal), or `—`.
- **Status**: `open`, `done` (fixed, evidence in the phase review), `next-goal`, `dropped`
  (removed by an owner decision), `not-a-defect` (evidence in the review record).
- Sources: `REVIEW` = [REVIEW-2026-10-12.md](../docs/simulation-canvas/REVIEW-2026-10-12.md) ids; `R0` = [reviews/P0.md](../docs/simulation-canvas/reviews/P0.md); `R1` = [reviews/P1.md](../docs/simulation-canvas/reviews/P1.md); `R2` = [reviews/P2.md](../docs/simulation-canvas/reviews/P2.md); `R2b` = [reviews/P2b.md](../docs/simulation-canvas/reviews/P2b.md); `R2d` = [reviews/P2d.md](../docs/simulation-canvas/reviews/P2d.md); `R5a` = [reviews/P5a.md](../docs/simulation-canvas/reviews/P5a.md); `R5b` = [reviews/P5b.md](../docs/simulation-canvas/reviews/P5b.md); `R6a` = [reviews/P6a.md](../docs/simulation-canvas/reviews/P6a.md); `R6b` = [reviews/P6b.md](../docs/simulation-canvas/reviews/P6b.md); `R6c` = [reviews/P6c.md](../docs/simulation-canvas/reviews/P6c.md); `R6d` = [reviews/P6d.md](../docs/simulation-canvas/reviews/P6d.md); `R6e` = [reviews/P6e.md](../docs/simulation-canvas/reviews/P6e.md); `R7c` = [reviews/P7c.md](../docs/simulation-canvas/reviews/P7c.md); `R7d` = [reviews/P7d.md](../docs/simulation-canvas/reviews/P7d.md); `R7e` = [reviews/P7e.md](../docs/simulation-canvas/reviews/P7e.md); `R7g` = [reviews/P7g.md](../docs/simulation-canvas/reviews/P7g.md); `R2c` = [reviews/P2c.md](../docs/simulation-canvas/reviews/P2c.md); `R3` = [reviews/P3.md](../docs/simulation-canvas/reviews/P3.md); `self` = found by the implementer.

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
| I3 | REVIEW | Med | Customer workspace marker, ignore rules and check; done in P8a: the STREAMSIM_CUSTOMER_WORKSPACE marker, `/workspaces/` ignored, and TestNoCustomerWorkspaceInsideTheRepository in ci-check | P8a | done |
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
| R0-14 | R0 | Med | Technology decision record (UI, bridge language and protocol): [canvas-technology.md](../docs/simulation-canvas/canvas-technology.md) | P8a | done |
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
| R3-03 | R3 | Low | Session records (`consumer.reported`, `pacing`) are outside the stream digest and unverified; validate them against the contract and never let scoring trust them; the digest part is done in P7l (journal file digests the run report checks, see R7e-08), contract validation and the scoring rule remain | P10 | open |

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
| R6a-01 | R6a | Low | Store stiffness and flow sign are unchecked: nothing bounds `conductance / capacity × dt` (RK4 diverges near 2.8) and a negative scale state reverses a flow; validate the bound and keep health scales non-negative | P7i | done |
| R6b-01 | self | Low | Calendar billing periods (months of unequal length) and the projected demand of the open block are not modelled; `peak` uses a fixed `period_s` | P12 | open |
| R6b-02 | R6b | Low | A read of an integrated state further in the past than the entity's longest step returns the current value (the snapshot history covers one longest step); runs read forward, so no output is affected; return a located error or keep a longer history | P13 | open |
| R6c-01 | R6c | Low | Oracle and audit worlds apply profile setup calls before injecting faults, so a setup call dated after a fault onset is checked (interlock, preconditions) without the fault; inject faults first or refuse setup dated after the onset | P12 | open |
| R6c-02 | R6c | Low | A precondition refusal of a safe-stop effector becomes `ErrPlantUnavailable` in the device bridge (`completeSafeStop`); decide whether a safe-stop effector may carry `requires` | P7c | done |
| R6e-01 | self | Med | Aggregation of dynamics inputs over typed links (`sum`, `mean`, vocabulary table §7) is in no phase: links are structure only; needed before the reference scenario models a shared header or bus across instances | P7h | done |
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
| R7e-04 | R7e | Med | World: an assign effect with a time constant starts from 0, not from the current value; one-sided state bounds clamp to 0; dt/τ stability is unvalidated for lag, dead time, rc network and regime; overlapping id templates pass validation; a negative lease is accepted | P7f | done |
| R7e-05 | R7e | Med | Truth labels a pulse observable on a channel that cannot record it; fixed in P7l: a pulse is judged only at the instants its detector channels record | P7f | done |
| R7e-06 | R7e | High | Run records: a reused run folder appended to `ledger.jsonl` (doubled rows); `verify`, `report` and `score` trust `ledger.jsonl`, the history, `verdict.json` and `label.json`, which no digest covers | P7f | done |
| R7e-07 | R7e | High | Quiescence: a consumer can report quiescence through any time and switch off every later barrier; the watermark is not clamped to the world clock | P7f | done |
| R7e-08 | R7e | Med | Run records: session records were outside every digest; fixed in P7l with a second digest per journal file over every line as written (`journal_file_digest`, `sealed_journal_file_digest`), which the run report checks and replay does not (pacing holds wall-clock waits). Fixed in P7f: preamble and postamble lines are now in the observer journal (`framing` mark); the report drops a torn last line and names it in `torn_tail`; a fault onset before the clock of a world that has run is refused (`ErrFaultOnsetInPast`); the Markdown report escapes free text | P7f | done |
| R7e-09 | R7e | Med | Truth: the nameplate and `sim.effector.list` pass effector descriptions (scoring and fault hints) and full argument schemas to the operator | P7i | done |
| R7e-10 | R7e | Med | Truth: the report shows truth whenever a `run.json` exists, without checking it belongs to the run of the journals | P7f | done |
| R7e-11 | R7e | Med | Tests: no test that a met barrier resets the missed count, that the quiescence watermark never moves back, that a command with a higher epoch is refused, or for a perturbation window end; the host fuzz target is not in `make fuzz` | P7f | done |
| R7e-12 | R7e | Med | Findings phases (done in P7e: R3-02 to P9, R2-01 to P8b, R6a-01 to P12): R3-02 belongs to P9 (the Verified badge), R2-01 to P8b (the bridge reads run files), R6a-01 to P12 (factory stores) | P7e | done |
| R7e-13 | R7e | Low | `TestDirectorServesItsToolsUntilTheSessionEndsThenStopsTheOperatorEndpoint` failed once under a full parallel run (the operator endpoint still accepted a connection right after the session ended); it passes in isolation | P7f | done |
| R7g-01 | R7g | High | The engine treats an absent `observation_bias.noise_scale_at_full` as 0 although the schema documents 1, so any bias fault silences the channel's noise; compose writes 1 explicitly (COMPONENTS §3), the engine default and a regression test belong with the observation faults | P7i | done |
| R7g-02 | R7g | Med | Equipment start and stop need state products (power and load scale with `running`); the chiller ships without them until P7i; the forms exist since P7i (`product`, §23), the chiller gains start and stop with the catalog; done in P7k: the compressor, dryer, boiler, chiller, conveyor, generator, motor drive, pump and both lines have `running` and start and stop effectors, with power, flow and heat scaled by `product` | P7k | done |
| R7h-01 | R7h | Med | Truth oracle worlds start every instance at the world start and replay no births or retirements, so a presence-sensitive link aggregate (`sum`, `mean`, `count_running`) in a churning domain differs between the run and the solver's clean and faulted worlds; fixed in P7l: the oracle worlds run from the scenario's world seed and are advanced to every sampled instant, so churn and every other queued event happen in them | P7k | done |
| R7k-01 | R7k | Med | Kinds may not read each other in a cycle over links, so a consumer cannot read back the header it feeds (line air demand into a receiver, receiver pressure back into the line); a safe form needs sampled, delayed link reads and a history retention that terminates over cycles | P7k | done |
| R7k-02 | R7k | Low | A consuming port follows one providing kind; summing loads of several kinds needs one meter per kind under a main meter composed as its own kind (hierarchical metering); fixed in P7l: a `follow` takes several sources (one per link type) and compose gathers every kind connected into a consuming state | P7l | done |
| R7l-01 | R7l | High | Truth: the pulse recorder read the clean oracle world before advancing it, so a pulse in a churning domain was labelled observable at its onset against a clean world missing the births | P7l | done |
| R7l-02 | R7l | High | Truth: oracle worlds took the literal world id `w-solver`, so their random substreams (churn births and lifetimes, stochastic fault envelopes) differed from the run's; R7h-01 was only half fixed | P7l | done |
| R7l-03 | R7l | Med | KPI truth: a retirement less than one sample step before the close extended the retired entity's active alarm time to the clock, and its alarm counted as standing | P7l | done |
| R7l-04 | R7l | Low | KPI truth: the sample at the instant an advance ended missed a fault injected at that instant with the default onset; a fault clear before the clock rewrote sampled history | P7l | done |
| R7l-05 | R7l | Low | Truth: the pulse recorder reads the faulted oracle inside its emitter, which advances a `random_walk` drift channel's walk; no shipped pulse detector uses one, so it is latent; forbid `random_walk` on a pulse fault's detector channels or make the oracle reading side-effect free | P12 | open |
| R7l-06 | R7l | Low | Truth: advancing the oracle worlds processed every emission event; oracle worlds without a pulse now schedule none (`Unobserved`), as before P7l | P7l | done |
| R9-01 | R9 | Med | `streamsim run` derives the world id from the domain alone, so every CLI run of one domain carries the same `world_id` (e.g. `w-3noul`); the bridge names run folders by world id, so two runs of a plant cannot sit in one workspace under their real ids | P9 | open |
| R9-02 | R9 | Low | `GET /api/workspace` run entries are `{world_id, ended, explore}`, while CANVAS_BRIDGE.md §2 documents `{world_id, ended, mode}`; align the document with the code | P9 | done |
| R9-03 | R9 | Med | A composed domain keeps no port names on its links (`{from, to, type}` only), so a plant view of a domain without its composition cannot draw typed pins; compose should record each link's source and target port | P9 | open |
| R9-04 | R9 | Low | Run files over 8 MiB (`trace.jsonl`, `sealed/journal.jsonl`, `world_state_history.jsonl` of a long run) are refused by `GET /api/file`; playback must use the observer event stream, and true-versus-reported after sealing needs a bounded read of the sealed journal (by time window or entity) | P9 | open |
| R9-05 | R9 | Med | Playback of a long run has to stream the whole observer journal from the start (520 k records, about 180 MB, for an 18 h factory run; 3.8 M for the campus), because the bridge has no read of observations by time window; a seekable read (records between two times, or every channel's last value before a time) would make seek instant and bounded | P9 | open |
| R9-06 | R9 | Med | `streamsim verify` (and the canvas `run.verify` action) replays a run against the domain file currently in the domains directory, not the `domain_spec` recorded in its `run.json`; after a domain is edited or recomposed every earlier run of it fails with `replay_mismatch: domain digest mismatch`, so the badge can only say "Not checked". Fixed: verify and replay use the recorded domain and adapter specs and report `domain_changed` and `current_domain_digest` as information | P9 | done |
| R9-07 | R9 | High | Engineering review E3-2: when the artifact's own recorded digests did not hold (embedded spec not matching its digest, or the recorded specs replaying to another world digest), `verify` exited `replay_mismatch` with no `result`, so a consumer could not tell a verdict from a failure to run. Fixed: the replay reports `matches: false` with a `detail` naming the reason; a missing or unreadable `run.json` stays an error | P9 | done |
| R9-08 | R9 | Med | Engineering review E3-3: `verify` did not compare `trace.jsonl` on disk with the digest `run.json` recorded (only ledger, history and journals), so a trace with an appended newline verified as matching. Fixed: the trace file is checked and the `detail` names it | P9 | done |
| R9-09 | R9 | Med | Engineering review E3-6: `scenario from-run` and `score` loaded the current domain from the directory and failed with a digest mismatch for every earlier run once the domain file was edited. Fixed: both use the recorded specs through `run.ResolveInputs`, the function `replay` and `verify` use | P9 | done |
| R9-10 | R9 | Low | Engineering review E3-7: determinism.md and quickstart.md still said `verify` loads the current domain files. Fixed, with CLI_SURFACE.md and the CLI reference | P9 | done |
| R9-11 | R9 | Low | Engineering review E3-8: `domain_changed` is omitted when the catalog fails to load or the run's domain was removed. Pinned deliberately by a test and documented: an absent `domain_changed` means unknown | P9 | done |
| R9-12 | R9 | Low | Engineering review E3-10: two canvas bridges on `127.0.0.1` with different ports shared the cookie name `streamsim_canvas`, so pairing one unpaired the other. Fixed: the cookie is named `streamsim_canvas_<port>` | P9 | done |
| R9-13 | R9 | Med | Design review D3-12: an unblinded Explore run shows only observed values; the canvas cannot draw the true state in purple because no bridge action reads `world_state_history.jsonl` by entity and time. A truth read for Explore runs (refused for Evaluate runs until sealed) belongs in the simulator before the canvas can show it | P9 | open |
| R9-14 | R9 | Med | Design review D3-11: `streamsim report` and `verify` on the 1.45 M-observation campus run take 100–120 s and reach 7–8 GB of memory; the canvas also builds its playback index on the main thread (3.8 s blocked). Bounded memory in report and verify, and a worker or windowed read in the canvas (see R9-05) | P9 | open |
| R9-15 | R9 | High | Engineering review E4-1: `streamsim verify` skipped an evidence file when `run.json` had no recorded digest for it and a file that was absent, and the schema made every evidence digest optional; deleting a ledger or history row together with its digest, or deleting a file, verified as matching, and `report` never checked `world_state_history.jsonl`. Fixed: a run that ended complete must have `expected_trace_digest`, both journal digests and file digests, `ledger_digest` and `history_digest` (run artifact schema), `verify` requires every evidence file beside `run.json` and compares each with its recorded digest and with the replay (a missing file or digest is `matches: false` naming the file), and `report` checks the history file and all seven digests from one list (`RunArtifact.EvidenceDigests`). A trace sent to a `--sink-target` outside the run folder is also written to `trace.jsonl` there. The run-report schema gains the `journal_file`, `sealed_journal_file` and `history` check names | P9 | done |
| R9-16 | R9 | Med | Engineering review E4-3: `verify` said `matches: false` with a `detail` sentence only, so a consumer had to parse prose to tell a spec digest failure from a trace divergence or a changed file. Fixed: the replay and verify result carries `reason` (`match`, `recorded_spec`, `world_digest`, `trace_divergence`, `journal_divergence`, `evidence_divergence`, `unblinded_stamp`, `digest_missing`, `file_differs`, `file_missing`) and, for the file reasons, `file` | P9 | done |
| R9-17 | R9 | Low | Engineering review E4-8: `streamsim verify --help` printed "Usage of replay:", and `ResolveInputs` wrapped its error with a no-op `fmt.Errorf("%w")`. Fixed: the flag set is named after the command; the no-op wraps in `ResolveInputs` and `replayRecordedInputs` are removed (the CLI one stays: wrapcheck requires a wrap at the module boundary) | P9 | done |
| R9-20 | R9 | High | Engineering review E4-2: the canvas decided which faults were active from the report's command records, which cannot say it (the clear record has no entity or effective `at_ns`; pulse faults and responder `clear_fault` end faults with no command). The world now records each fault as an interval with how it ended, the run seals it as `fault.truth`, and the report's `truth.faults` gives `fault_id`, `entity_id`, `fault`, `injected_at_ns`, `onset_ns`, `cleared_ns`, `cleared_by` | P9 | done |
| R9-21 | R9 | Med | Engineering review E4-4: the canvas decoded annunciator values with a hard-coded table and recognised an annunciator by the id prefix. Instruments and channels now declare their state vocabulary (`states`: `value`, `name`, `attention`); compose copies it onto the channel, so a domain is self-describing. The annunciator declares the codes of `model.Alarm*` | P9 | done |
