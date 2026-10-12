# Impact analysis — Tamoz

- **Date**: 2026-10-12
- **Status**: analysis; **no code changed in Tamoz or here**
- **Scope**: what the simulation-canvas direction ([plant-wire-protocol.md](plant-wire-protocol.md),
  [reference-scenario-factory-utilities.md](reference-scenario-factory-utilities.md)) means for
  `/Users/ghassan/my-projects/tamoz`
- **Evidence**: Tamoz ADR-024, 036, 038, 039, 055, 058; `gems/tamoz-stream/lib/tamoz/stream/situation_request.rb`;
  `gems/tamoz-agent-kernel/lib/tamoz/agent/intent_catalog.rb`; `test/fixtures/domains/`

## 1. Verdict

**Tamoz needs no code or data change for the factory scenario, the plant wire or entity
templates.** It never sees devices, events or the wire. Everything scenario-specific reaches it
inside each episode request from Agentic Stream. Two Tamoz rules do constrain the scenario
design (§4), and a few existing Tamoz issues are worth fixing separately (§5).

## 2. Tamoz's role (from its ADRs)

| ADR | Rule | Consequence here |
|---|---|---|
| 055 | Agentic Stream is the authority; Tamoz is its episode worker. "Tamoz proposes; the stream disposes." | Tamoz is reached only through Agentic Stream, never by the simulator |
| 036 | Cognition sees only a sealed, digest-checked Situation snapshot, never raw evidence | Plant-wire records, births and sequences never reach Tamoz |
| 038 | The model proposes typed intents from a digest-pinned intent catalog; deterministic policy outside dispatches | No device commands, no actuator tools in Tamoz |
| 039 | Tamoz is supervisory; e-stops, interlocks and PLC loops stay external and authoritative; R4 control is advisory only | Shapes the kill-switch design (§4.1) |
| 058 | Domain knowledge is digest-pinned data, never code | Scenario knowledge is data, owned outside Tamoz's code |

## 3. What reaches Tamoz per episode

`situation_request.rb` reads these from the gRPC `EpisodeRequest`, each digest-checked:

| Field | Meaning |
|---|---|
| `snapshot_json` + `snapshot_sha256` | The sealed Situation |
| `objective`, `prompt`, `prompt_version` (+ digests) | What to do and how |
| `intent_catalog_json` + digest, `allowed_intent_types`, `risk_ceiling` | The action vocabulary and its limits |
| `diagnosis_catalog_json` + digest | The diagnosis vocabulary |
| `tool_catalog_json` + digest, `evidence_tools_endpoint`, `capability_token` | Read-only evidence tools |
| `model_policy`, `dispatch_policy`, `budget`, `skill_refs_json` | Model choice, dispatch rules, limits |

So the full scenario pack for the agent side is authored and pinned in **Agentic Stream**
(its `SituationSpec`) and delivered at runtime. Tamoz holds none of it in production.

## 4. Constraints the scenario must respect

### 4.1 Kill switch is advisory for Tamoz (ADR-039)

`emergency_shutdown` is R4. Under ADR-039 Tamoz may only **recommend** it. The scenario's
kill-switch test therefore measures a chain, not one agent action:

| Layer | Who | In the simulation |
|---|---|---|
| Plant safety system | Hardwired trips and interlocks | Simulator `trip` / `lockout` forms; authoritative, cannot be overridden |
| Deterministic host | Agentic Stream policy and interlock | May block the action plane; may route an R4 recommendation for human execution |
| Human | Operator or supervisor | Simulated responder executes the shutdown after a delay |
| Agent | Tamoz | Recommends shutdown with evidence |

Scoring must attribute timing to each layer: when Tamoz recommended, when policy routed it,
when the simulated human pressed it. This is more realistic than the agent pressing the
button, and it is the only design compatible with ADR-039.

### 4.2 The diagnosis catalog must not mirror the simulator's faults

The diagnosis catalog arrives from Agentic Stream. It must be written as a plant engineer
would write it, independently of the simulator's fault list, including causes the simulator
never injects and `unknown`. A catalog generated from the simulator's faults hands the agent
the answer set and invalidates the benchmark.

## 5. Existing Tamoz issues found (out of scope; fix separately)

| Issue | Evidence | Why it matters |
|---|---|---|
| Real domain packs live as test fixtures | `test/fixtures/domains/{aquaculture,climate,cold-chain,shipment,thermal-lab}.json`; loaded only from `test/support/` | Domain knowledge belongs to the deployment (Agentic Stream); Tamoz tests should use small synthetic catalogs that test the mechanism |
| Cross-repo digest pins on fixtures | ADR-058: the aquaculture intent-catalog digest is also pinned in the Go repository | Editing a Tamoz test fixture can break Agentic Stream; ownership is inverted |
| Domain branches in the loader | ADR-058 status: `test/support/domain_loader.rb` still carries domain-specific branches | ADR-058 is only partly implemented |
| No health endpoint | No health route found in Tamoz | The canvas cannot show Tamoz "up"; see §6 |
| Worker transport unauthenticated | ADR-055 status: insecure gRPC, umask-dependent socket, TCP on all interfaces | Not a canvas concern, but a live demo should run on the Unix socket only |

## 6. Liveness for the canvas

Agentic Stream dials Tamoz, so Agentic Stream already knows whether the worker is reachable
(handshake success, last episode). The cleanest source of Tamoz's status is therefore
**Agentic Stream reporting its worker state**, not a new Tamoz endpoint:

| Canvas status | Source |
|---|---|
| Up | Agentic Stream's last successful `Handshake` with the worker |
| Connected | Worker handshake within the last interval |
| Active | Last episode start/end, from Agentic Stream |

## 7. Benchmark overlap

Tamoz has its own benchmark protocol (ADR-024, `documentation/benchmark/BENCHMARK_PROTOCOL.json`,
hold-out pins). Proposed split, to confirm:

- **The simulator scores the world**: safety, regulatory, production, energy, wear, from sealed truth.
- **Tamoz's protocol scores the reasoning**: decision quality against its own criteria, on the
  episodes the run produced.
- One run artifact links both, so a report can say "unsafe outcome at t, caused by this decision,
  made on this snapshot."

## 8. Changes required in Tamoz

| Change | Needed for MVP? |
|---|---|
| None for the plant wire, templates or the factory scenario | — |
| Move real domain packs out of test fixtures; synthetic catalogs in tests | No; recommended cleanup |
| Align ADR-024 benchmark protocol with simulator scoring (§7) | Before the first published benchmark |
