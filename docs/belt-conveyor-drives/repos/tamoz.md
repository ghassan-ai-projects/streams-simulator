# tamoz for the belt-conveyor use case

Status: repo note, 2026-11-07. Read-only investigation of `/Users/ghassan/my-projects/tamoz` (branch `belt-conveyor-drives`),
no builds or runs. Paths are relative to tamoz unless prefixed. "FAN" means
`/Users/ghassan/my-projects/streams-simulator/docs/physical-devices/fan-experiment`, "RWS" means
`/Users/ghassan/my-projects/agent-research-lab/real-world-sensor`.

Convention: **Facts** carry file:line. **Inference** is marked as such and is not verified.

---

## 1. What tamoz is

### Facts

- **Purpose.** "Ruby-native durable agent framework for checkpointed, interruptible, observable, and evaluation-governed AI
  workflows" (README.md:3-4). An agent turn is a graph run over a SQLite checkpoint store that survives `kill -9`, and "nothing
  acts without a reviewed plan bound to its digest" (README.md:6-10). Pre-release `0.1.0.alpha.1` (README.md:12).
- **Users.** Operators running a personal or team agent, Ruby developers building durable agent workflows, engineers
  integrating models/MCP/scheduled work (documentation/overview/product.md:26-30).
- **Explicit non-goals.** "Not a multi-tenant hosted platform. Single operator, many sessions. No auth server, no billing, no team
  workspace" (product.md:22). "Not a certified or hard real-time controller ... emergency stops and interlocks remain external
  and authoritative" (product.md:23). Same in documentation/limitations.md:42-47.
- **Tech stack.** Ruby 3.3 (AGENTS.md:12), ~35 gems under `gems/` (README.md:16-51): graph engine `tamoz-graph`, SQLite
  adapter `tamoz-sqlite`, agent layers `tamoz-agent-*`, gRPC episode worker `tamoz-stream` (deps: gRPC, protobuf, README.md:25),
  approval engine `tamoz-approval` (policy as YAML), channels `tamoz-comms`/`tamoz-comms-gateway`/`tamoz-telegram`/`tamoz-talk`,
  `tamoz-scheduler`, `tamoz-observability`, `tamoz-otel`. Reference app and CLI: `tamoz` (`tamoz-agent-cli`).
- **Architecture of one stream episode.** `tamoz-stream` implements the `agenticstream.runtime.v1.EpisodeWorker` gRPC service
  (Handshake, Execute, optional DraftPlaybookRule); the Agentic Stream runtime dials the worker, normally over a Unix socket, and
  streams one episode against a sealed, digest-verified Situation snapshot (documentation/design/streaming.md:9-19, :28-30).
  Launcher: `bin/tamoz-stream-worker` (bin/tamoz-stream-worker:1-17; flags `--profile --database --tenant --socket`).
- **What the worker emits.** Typed Decisions: observed facts separated from inferences, evidence cited, uncertainty, validity
  interval, zero or more typed `ActionIntent`s (streaming.md:34). Every DIAGNOSE decision carries a `diagnosis_code`
  (streaming.md:36).
- **Reads only.** The episode runs in a containment host with a fixed read-only tool allowlist: `features.query`,
  `evidence.get`, `situations.related`, `history.prior_incidents`, `knowledge.search`, `forecast.run` (streaming.md:26).
  Worker events are "proposals and telemetry; the runtime remains authoritative" (streaming.md:19).
- **No actuator.** "The supervised worker proposes typed Decisions (intents); it holds no effector at all. Actuation belongs to
  the stream ... No actuator adapter exists on the Tamoz side" (limitations.md:228-236 region, section "Real physical
  actuation").
- **Domain knowledge is data.** "Catalogs, prompts, intents and risk classes ... live only in `test/fixtures/domains/*.json`"
  (AGENTS.md "Domain knowledge is data, never code (ADR-058)"). Fixtures present: aquaculture, climate, cold-chain, pump-station,
  shipment, thermal-lab (`ls test/fixtures/domains`). At run time tamoz holds no domain knowledge: intents, catalog, prompt and
  objective arrive from Agentic Stream on every request (FAN/PLAN.md:35).

### Role relative to the other two repos

| Repo | Role | Source |
|---|---|---|
| streamsim | The world (twin) and plant wire node; produces the telemetry and executes commands in the simulated plant | FAN/PLAN.md:11-12 |
| agentic-stream | Continuous plane: ingest, event time, situations, snapshots, risk classes, approval authority, action plane, effectors, verification, run artifact | RWS/assessment/WORK-tamoz.md; streaming.md:63; limitations.md (P14 retirement) |
| tamoz | The "brain": supervisory cognition over one sealed snapshot; also the optional human-approval delivery bridge and the learning loop | WORK-tamoz.md:3-5; approval_relay.rb:5-12 |

Research invariant quoted verbatim in tamoz's own docs: "Do not give Tamoz the serial port" (docs/real-world-sensor-tamoz/README.md:75).

---

## 2. The thermal LED/fan experiment: what tamoz did

### Facts

- **Plan statement.** FAN/PLAN.md rule R2: "No work in Tamoz", configuration only (worker profile, database path, model key,
  pinned Ruby), verified (FAN/PLAN.md:18, "Tamoz" section ~322-324). Chain: `Tamoz <-> Agentic Stream <-> plant wire <-> streamsim
  <-> plant wire <-> gateway <-> serial <-> Arduino` (FAN/PLAN.md:11-12). Tamoz has no branch in the experiment (FAN/GOAL.md:9) and
  FAN/HANDOFF.md:17 records tamoz `real-run` at `9170538c`, "no code".
- **Prior tamoz work for the sensor program** (2026-08-28): a tamoz-scoped program, WP-T0 to WP-T5, mostly domain data
  (`test/fixtures/domains/thermal-lab.json`), test support and one Ruby change (`metrics.rb`, +44 lines)
  (docs/real-world-sensor-tamoz/README.md:20-27). Thermal intents and risk classes: `install_watch_condition` R0,
  `request_evidence` R0, `set_indicator` R1, `request_bounded_cooling` R2, `downgrade_cooling` R1, `withdraw_cooling` R1
  (test/fixtures/domains/thermal-lab.json:15-21).
- **Real-model evidence.** `script/thermal_real_run` (never in CI) drove 8 trial cells with `z-ai/glm-5.3-flash`; abstention
  quality ~0.50 baseline vs ~0.875 model; the disconnected-sensor cell went from false-alarm to `request_evidence`; the
  ambient-driven cell was a real miss; the statistical verdict was `inconclusive` over 8 cells
  (docs/real-world-sensor-tamoz/README.md:29-56).
- **What tamoz did in the joined run.** It was the supervisory worker: receives the snapshot via Execute, returns a Decision
  with an intent (e.g. `set_indicator`, `select_thermal_mode bounded_cooling`); Agentic Stream classified risk, executed through
  the plant wire host, verified from the twin's own reading. Simulated-rung result (attempt 2): the fan command was forwarded,
  `DRECEIPT accepted`, `DRESULT succeeded`, fan `level_percent` 45 for its 5 s lease then 0, verified by the fan's reading;
  3 episodes, 3 decisions, 3 commands (FAN/GOAL.md, row P4). Model: `zai/glm-5.3-flash` from `~/.tamoz/config.yaml` (FAN/PLAN.md D11).
- **Earlier joined software run.** RWS/assessment/CURRENT-STATUS.md:71: "The active Tamoz worker produced the command path in the
  joined run" for one scenario; round-007 run: 61 accepted events, one Tamoz decision, one `set_indicator` command, reconciled
  (CURRENT-STATUS.md:41-46).
- **Findings logged against tamoz** (fan owner-waiting list, FAN/GOAL.md:59-62):
  - W1: the prompt frame names the output protocol but never defines it (keys, allowed intents); the repair message only repeats
    the error code. The fan spec worked around it in its own prompt (FAN/P0.5-FINDINGS.md:25-45).
  - W2: overlapping attempts. Agentic Stream redispatches 7 ms after a deadline while tamoz still holds the episode lease, so
    attempt 2 fails `checkpoint_conflict`.
  - W4: tamoz's terminal says only `graph_status: failed`; the real cause is only in a checkpoint.
  - W6: the consumer name `tamoz` is in Agentic Stream production code (the other direction of the same coupling).
- **Not used in the fan run: operator UI, approvals, reports.** Approval was deliberately placed at the node, not in tamoz (D9);
  in the simulated rung "no human approval" was an owner-decided deviation (FAN/GOAL.md row P4; FAN/PLAN.md:125-131). The
  approval relay "has no production implementation (only test fakes)" (FAN/PLAN.md:130-131), and a repo grep finds
  `build_approval_relay` only in `bin/tamoz-stream-subscriber` (the file loader) plus tests (`test/stream_approval_relay_test.rb`,
  `test/stream_invariants_test.rb:166`). Telegram-as-R2-approver is listed as a "later goal" (FAN/PLAN.md:198-199).
- **Learning loop** (`bin/tamoz-stream-subscriber`): reads Agentic Stream's Channel-B SSE feed, records outcomes into verification
  and memory (`live_learning_handlers.rb:46-100`), and can deliver/withdraw/resolve approvals if a relay file is supplied
  (live_learning_handlers.rb:119-176; flags at bin/tamoz-stream-subscriber:17-35). Whether it ran in the fan P4 is not stated in
  the files I read.
- **Real rung not yet done.** P5/P6 need the owner at the board (FAN/HANDOFF.md:6-8, :39-45). So the real-rung claim "tamoz drove a
  real fan" does not yet exist; the simulated rung passed.
- **Later fan goals that touch tamoz** (FAN/PLAN.md:376-391): A/B/C "does tamoz earn its cost", rules-first with tamoz only when
  hard, and tamoz drafting rule changes from learned outcomes (human-approved in Agentic Stream).

---

## 3. What tamoz would do for the conveyor

Everything below that is not a quoted capability is **inference** from the capabilities listed in sections 1 and 2.

### 3.1 Customer onboarding of a plant

- Facts: tamoz holds no plant knowledge; it receives intent catalog, prompt and objective with every request (FAN/PLAN.md:35).
  The knowledge it does need, the diagnosis codes and intent catalog, are authored as data in a domain JSON; `pump-station.json`
  already has `bearing_wear`, a vibration-creep lubrication code, a "probe reads high while healthy" sensor-fault code, a
  "pins high" stuck-sensor code, and intents `install_watch_condition` R0, `schedule_maintenance` R1, `reduce_load` R2,
  `stop_motor` R3 with preset `operator_stop` (test/fixtures/domains/pump-station.json:10-31, :36-60).
- Inference: onboarding is mostly Agentic Stream work (SituationSpec, feature definitions, effector catalog) and streamsim twin
  work. Tamoz's part is a worker profile (model roles, authority), a SQLite path, a tenant id, and a conveyor domain/catalog in
  the same shape as the pump-station one for tests and benchmarks. The domain file must stay generic or synthetic in the repo (see 5).

### 3.2 Showing the drives' state and predictions

- Facts: tamoz has no live dashboard or plant UI. Its operator surfaces are the CLI (`tamoz status --json`, `observe`), Telegram,
  and the browser "talk" page, which is a chat/voice page, not a plant view (documentation/guides/talk.md:1-5;
  limitations.md "Operator visibility is partial": status reports pending work, approvals, blocked effects, no recent
  completions). "No second UI" is a stated non-goal (limitations.md, "Deliberate non-goals").
- Facts: it can produce predictions as reasoning: Decision with a validity interval, uncertainty and cited evidence
  (streaming.md:34), and it has a read-only `forecast.run` evidence tool (streaming.md:26).
- Inference: the drive state view and failure-probability trace belong in streamsim's canvas (twin view) and Agentic Stream
  (situation/feature facts). Tamoz would supply the explanatory text (why drive 3 is judged degrading, what is uncertain) which
  can be shown in chat on request.

### 3.3 Approving actions and rules

- Facts: Agentic Stream classifies R2 and creates the request; tamoz's approval relay delivers the prompt to the bound channel,
  submits the human's signed answer to Channel C, handles escalation, with a signer whose key tamoz never sees
  (approval_relay.rb:5-12, :36-60). The relay requires injected delivery/submission/nonce-store/signer ports
  (approval_relay.rb:62-75). Telegram approval is evidence-gated and the paired chat can Approve/Deny
  (documentation/operations/operations.md:174-181; documentation/roadmap.md "Telegram as an operator channel").
- Facts: rule drafting. `DraftPlaybookRule` turns 1 to 20 verified cases into a restricted-CEL condition; the proposal has no
  mode, envelope or activation field; the stream replays it and a human approves there (streaming.md:38-47). Measured weakly:
  13 of 25 on one fixture after prompt revision, discrimination between bearing wear and sensor drift is the weak point
  (limitations.md, "Playbook-rule drafting is unproven beyond one fixture").
- Inference: for the conveyor the R-class ladder is roughly: alarm/watch R0, cut feed or shift torque bias R1/R2,
  stop drive 3 at a chosen time R2/R3. The stop of a drive is the pump-station `stop_motor` R3 analogue, and Agentic Stream's
  contract denies R3 and R4 outright (FAN/PLAN.md:35-ish table: "R0, R1 automatic; R2 approval; R3, R4 denied", citing
  `agentic-stream/contractsv1/internal/domain/risk.go`). A conveyor stop decision would therefore need a risk-class decision in
  Agentic Stream's catalog.

### 3.4 Alerts to operators

- Facts: Telegram is "a real, configured operator surface"; delivery is durable outbox with ambiguous sends kept `:unknown`, never
  blindly retried (limitations.md "Channel communications"). Pairing and allowlist model: operations.md:124-143. Alerting and
  automated response "are intentionally outside this phase and require separate authorization under phase 5 / ADR-048"
  (limitations.md "Observability remains partial").
- Facts: the subscriber currently handles only outcome and approval notifications from the stream feed
  (`LiveLearningHandlers#callables`, live_learning_handlers.rb:29-40, :46-176). No handler turns a decision or a prediction into
  an operator message.
- Inference: a "drive 3 degrading" push alert needs either a new handler on a decision/situation event or Agentic Stream-side
  notification routing. It is not there today.

### 3.5 Reports of avoided damage

- Facts: tamoz has run reports of its own kinds (a deep-research report, an evaluation scorecard, a run artifact manifest with
  prompt/tool/memory digests, streaming.md:59) but nothing that reports "damage avoided". The stream's run artifact and
  `export-run`/`verify-run` live in Agentic Stream (FAN/GOAL.md W3: they report safety verdicts only, a no-action run says
  `pass`). Streamsim has `internal/report` (run report) and sealed ground truth, per its AGENTS.md.
- Inference: "avoided damage" as a number is only definable against a twin counterfactual (same seed with and without the
  action), which is a streamsim/Agentic Stream product (A/B/C idea at FAN/PLAN.md:379-383). Tamoz can narrate it, not compute it.

---

## 4. Gaps for this use case

Sizes: S = under a week of one engineer, M = 1 to 3 weeks, L = a month or more. All are inference except where cited.

| # | Gap | Where it lives | Size | Notes |
|---|---|---|---|---|
| 1 | Prompt frame does not define the output protocol or allowed intents (W1) | tamoz `tamoz-agent-kernel` `episode_frame_builder.rb` (FAN/P0.5-FINDINGS.md:25-45 cites lines 118-134) | S | Already a known open owner item; matters more with ten sensors per drive and more intents |
| 2 | Overlapping attempts / `checkpoint_conflict` on redispatch (W2); failure reason lost (W4) | tamoz worker + Agentic Stream dispatcher; owner chooses side | S to M | A decision miss on a degrading drive must not be lost to a lease race |
| 3 | Conveyor domain data: diagnosis codes (bearing, seized gearbox, sudden trip, feed surge, stuck/dropped sensor), intent catalog, risk classes, rule-proposal features, benchmark families | `test/fixtures/domains/conveyor.json` plus loader fixtures, synthetic only (ADR-058) | M | Pump-station fixture is the template (pump-station.json:10-60). Intents: cut feed, run empty, torque bias, stop drive at time, alarm |
| 4 | Multi-drive, load-redistribution reasoning (feed surge is not drive 3 failing; survivors' cumulative aging) | Mostly Agentic Stream features and prompt/objective, plus tamoz eval cells | M to L | Tamoz only reasons over the facts the snapshot carries; fleet-level facts like survivor thermal margin must be computed upstream |
| 5 | Production approval relay: Telegram delivery adapter, Channel C client, key-custody signer | tamoz `tamoz-stream` ports, `tamoz-telegram`/`tamoz-comms`; Agentic Stream needs no change (FAN/PLAN.md:198-200) | M | No production implementation exists (FAN/PLAN.md:130-131). Needed to put a drive stop behind a named human |
| 6 | Proactive operator alerts from decisions/predictions | New subscriber handler in `tamoz-stream` plus a comms outbound route; or Agentic Stream notification | S to M | Existing handlers cover only outcomes and approvals (live_learning_handlers.rb:29-40) |
| 7 | Time-bounded actions ("stop drive 3 at 03:00", "empty the belt before restart") | Agentic Stream effector/intent preset with a timer; tamoz proposes the time as a preset or parameter | M | Tamoz has `at` and `interval` schedules but no cron or timezone (limitations.md "Cron and civil-time scheduling"); a planned shift end needs civil time or an absolute UTC instant |
| 8 | Real-model measurement for the conveyor: A/B/C vs a thresholds-plus-trend baseline, sealed truth from streamsim | tamoz `script/*_eval` style runner plus streamsim scenarios | M | Thermal result was `inconclusive` over 8 cells (docs/real-world-sensor-tamoz/README.md:48-52); the trivial-baseline question "does tamoz beat a trend-on-vibration rule" must be answered per scenario |
| 9 | Rules-first, model only when hard | Agentic Stream (`decide: rules / tamoz`, FAN/PLAN.md:384-386) | L | For a slow bearing failure with hours of precursor, a rule probably suffices; tamoz earns its place on conflicting sensors and look-alikes |
| 10 | Tamoz proposes rule changes from learned outcomes | tamoz `DraftPlaybookRule` plus Agentic Stream replay gate | M | Drafting quality is weak on look-alike pairs (limitations.md "Playbook-rule drafting") |
| 11 | Plant UI for drive state, prediction and history | Not tamoz (streamsim canvas / Agentic Stream) | L | Tamoz has no UI by design |
| 12 | Avoided-damage report | streamsim twin counterfactual + report module; tamoz narration | M | See 3.5 |
| 13 | Real-hardware rung for a conveyor | owner-side gateway and PLC/VFD interface | L | The fan real rung is not even done yet (FAN/HANDOFF.md:6-8). Tamoz has no actuator adapter and must not (limitations.md) |
| 14 | Spend cap, durable model-usage persistence, recent-completion visibility | tamoz runtime | M | `cost_usd`, tokens, steps are recorded only, not enforced (limitations.md "Only two budgets are enforced"). Needed for unattended 24/7 operation |

Release gaps that apply to any customer deployment: limitations.md:14-155 lists release-blocking evidence gaps (kill-window
evidence environment-bound, versioned-record refusal, single-writer takeover, verified artifact store tamper evidence). Tamoz
states v0.1 "is not released" (limitations.md "Release readiness").

### Honest scoping of tamoz's value here (inference)

- The prediction itself (drive 3 degrading) is most simply done by a trend/anomaly feature in Agentic Stream plus a rule. Tamoz
  adds value where rules are brittle: separating drive degradation from a feed surge, discounting a stuck or dropped
  vibration channel, explaining and asking for evidence. That mirrors where it beat the thermostat baseline in the thermal work
  (disconnected sensor, door spike, conflicting sensors; docs/real-world-sensor-tamoz/README.md:39-45).
- Tamoz must not act in the fast loop. Torque sharing and trip response belong to the VFD/PLC and a safety system, which is
  its own stated boundary (limitations.md:42-47).

---

## 5. Customer deployment and data isolation

### Facts

- Owner rule (memory note `customer-data-isolation.md`, 2026-10-12): per-customer configs and data (SituationSpecs, policies,
  gateway mappings, twin domains, run artifacts, logs) never in GitHub; customers must not know about each other; each customer gets
  its own deployment (Agentic Stream + DB, Tamoz worker + memory, gateway); nothing learned at one customer flows to another.
  Details are pointed to `docs/simulation-canvas/plant-wire-protocol.md` section 16 in streamsim.
- Tamoz fits the "one isolated deployment" model: product docs say single operator, no hosted multi-tenant platform
  (product.md:22). State is one SQLite file per worker/thread (operations.md:3, :13-24) plus a runtime directory outside any
  repository (`--runtime-dir ~/.tamoz`, `~/.tamoz/config.yaml`, channels, backups exist on this machine: `ls ~/.tamoz`).
  Memory is per tenant in that database (`Memory::Engine.new(tenant:, adapter:)` at bin/tamoz-stream-subscriber:60).
- A `--tenant` flag is mandatory for the worker and subscriber (bin/tamoz-stream-worker:62; bin/tamoz-stream-subscriber:38).
  The subscriber refuses notifications whose tenant does not match (live_learning_handlers.rb:178-181).
- Trusted profiles: project authority lives outside the repository being worked on, and a file in an untrusted checkout never
  becomes authority without import and preview (README.md:91-93). Secrets are config, never printed or committed (AGENTS.md
  "Commands"; FAN/HANDOFF.md:67).
- Production worker intent: a UDS socket with mTLS "at the socket boundary", profile under `/etc/tamoz`, database under
  `/var/lib/tamoz` (bin/tamoz-stream-worker:12-17). Backup uses the online backup API, not `cp` (operations.md:13-29). Backups and
  deletion semantics: operations.md:78-87.
- Tamoz does not survive loss of the SQLite file without a backup (limitations.md "Boundaries you should plan around").
- Egress and data to the model provider: tamoz calls an external OpenAI-compatible model API (README.md quick start; FAN/PLAN.md
  D11 uses Z.ai GLM). A customer's snapshot content goes to that provider. Inference: the customer must accept that, or a local
  model route must be configured; the docs I read do not establish a local-model path as tested.
- Repo hygiene: tamoz untracked `PROFILE.md` and `outreach/` are the owner's; I did not read or touch them. Repos must carry only
  generic or synthetic conveyor content (the memory note).

### Implications for the conveyor (inference)

1. Keep the conveyor domain fixture synthetic (invented plant numbers) in tamoz `test/fixtures/domains/`; the customer's real
   ratings, positions and sensor maps go in a per-customer folder outside every repository, passed by path.
2. One tamoz worker, one SQLite file, one tenant id, one model key, one Telegram bot and allowlist per customer.
3. Memory ("what drive 3 looked like before it failed") is per customer by construction; nothing to share, which also means
   no cross-customer learning, as intended.
4. Before the first customer gateway is in the field, the no-backward-compatibility freedom (AGENTS.md ADR-059) ends; schema or
   profile changes then need migrations that assume existing databases. Inference: plan the first customer cut-over date
   against that.
5. The model provider leg is the one data path that leaves the customer site; it needs a customer decision and an egress
   allowlist (tamoz has an egress concept in profiles, README.md:91-93, but I did not verify its coverage of the model transport).

---

## 6. Summary

Tamoz would be the supervisory reasoner for the conveyor: it reads a sealed snapshot of the drives, separates real degradation
from look-alikes (feed surge, bad vibration channel), cites evidence, and proposes typed, risk-classed intents. It never actuates,
never owns the trip response, and holds no plant knowledge of its own. Operator-facing pieces (state view, alerts, avoided-damage
report) are mostly not tamoz; what tamoz can host is the Telegram approval path and the explanation text, and that approval
relay has no production implementation yet. The thermal fan experiment proved the simulated-rung chain with tamoz in it, with
open tamoz items (prompt frame, redispatch race, failure reason) that matter more for a real customer.

Unverified in this pass: whether the subscriber and relay ran in the fan P4 run; whether the model transport honours the profile
egress allowlist; any local-model path; the exact line numbers of the "Real physical actuation" paragraph in limitations.md
(cited by section name above).
