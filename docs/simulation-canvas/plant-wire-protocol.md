# Plant wire protocol — industrial birth/data/death pattern

> **Superseded by [contracts/plant-wire/SPEC.md](../../contracts/plant-wire/SPEC.md) (spec v2, phase P7a).** This proposal is
> kept as the design record.
>
> **Status note (2026-10-12):** parts of this document are superseded by the owner decisions in [REVIEW-2026-10-12.md](REVIEW-2026-10-12.md) §1–§2 and tracked in [FINDINGS.md](../../governance/FINDINGS.md). The `emergency_shutdown` policy line in §7 is dropped; authentication is deferred; the node coexists with the existing device path in this goal; the document is rewritten as spec v2 in phase P7a.

- **Date**: 2026-10-12
- **Status**: design proposal; replaces the simulator ↔ Agentic Stream integration for the
  simulation canvas MVP. **No code or contract changed.**
- **Depends on**: [reference-scenario-factory-utilities.md](reference-scenario-factory-utilities.md)
- **Compatibility**: none required (pre-release). The current `trace-record-v0.1` adapter,
  Agentic Stream's `simulator_data.json` channel table and the device capability catalogs
  may be removed or rewritten.

## 1. Decision

The simulator and Agentic Stream talk the way industrial equipment talks to a host:
**devices announce themselves, then report; the host discovers rather than being configured.**
The model follows Sparkplug B's session semantics (birth, data, death, command, host state,
rebirth) and OPC UA's idea of a browsable, self-describing address space.

**Pattern, not protocol.** Sparkplug runs over an MQTT broker and protobuf; this repository
forbids a broker and keeps zero external dependencies. The records are JSON Lines over the
transports that already exist: a Unix socket for live runs and a file for recorded ones.

**Hard constraint: the real-world-sensor experiments always work.** The thermal LED/fan
experiment (`agent-research-lab/real-world-sensor`, HIL-0) is acceptance test #1 for this
design and stays green on both rungs: the emulated joined run (G1) and the physical Arduino
bench. See §12.

## 2. Roles and layering

The plant wire is the **host side** of an edge gateway, as in real plants: a field protocol
between gateway and device, Sparkplug-style messaging between gateway and host.

```
device / firmware ⇄ field protocol (device-wire-v1, arduino-telemetry-v1) ⇄ edge gateway ⇄ plant wire ⇄ Agentic Stream
```

- **Firmware does not change.** It keeps device-wire-v1 with its safety semantics.
- **The gateway** (today `arduino_gateway.py`) gains a plant-wire side: it announces its
  devices in births and maps commands down to device-wire records. It stays transport
  plumbing, never a policy engine.
- **The simulator emulates both sides**: the plant wire (canvas, factory, any domain) and the
  field protocol (protocol-fault testing for HIL), from the same world.

| Role | Here | Sparkplug analogue |
|---|---|---|
| Node | One simulated world, or one edge gateway in front of real hardware | Edge node |
| Device | One entity instance (`M1`, `AC1`, `P2`, …) | Device |
| Host | Agentic Stream | Primary host application |
| Observer | Canvas, coding agents (read the run journal) | Secondary host, read-only |

## 3. Records

| Record | Direction | When | Carries |
|---|---|---|---|
| `node_birth` | node → host | Session start, and on rebirth | Session id, node metrics (clock, run state), entity kinds (templates), sequence reset, accepted policy and capability digests |
| `device_birth` | node → host | Per entity after `node_birth`; when an entity appears; after a device reboot | Entity kind and id; **`boot_id`**; clock capability; **metrics**: name, alias, data type, unit, range, resolution, description, writable; **commands**: name, argument schema, risk class, whether it is the dedicated safe-stop path |
| `device_data` | node → host | Each observation | Sequence, `boot_id`, alias, value, device monotonic time and mapped event time, quality |
| `device_death` | node → host | Entity offline (sensor outage, link down, entity retired, reboot) | Entity id, reason class, last sequence, last `boot_id` |
| `node_death` | node → host | Session ends or connection lost (last-will semantics) | Session id |
| `command` | host → node | Agent action approved by Agentic Stream policy | `command_id` (one dispatch), `idempotency_key` (one semantic effect), entity, command name, arguments (incl. **lease**), **`expected_boot_id`**, **`authority_epoch`**, **`expires_after_ms`**, `not_before`, **`policy_digest`** |
| `command_receipt` | node → host | Accepted, refused or queued | Command id, status, refusal reason (`expired`, `wrong_boot`, `stale_epoch`, `replayed`, `out_of_bounds`, `policy_not_allowed`, …) |
| `command_result` | node → host | Outcome as reported by the device | Command id, reported outcome (may disagree with reality), `boot_id` |
| `state_query` / `state` | host → node / node → host | Reconciliation after unknown outcome or reboot | Device state, `safe_state` flag, energized outputs, `boot_id` |
| `host_state` | host → node | Host online/offline; carries the current `authority_epoch` | Host id, online flag, epoch, timestamp |
| `rebirth_request` | host → node | Host detects a sequence gap or unknown alias | Session id |

**Deviation from Sparkplug:** Sparkplug commands are fire-and-observe metric writes. This
design keeps explicit `command_receipt` and `command_result` because the benchmark needs
acknowledgements that can lie (`silent_no_effect`, `partial`) and the existing device
contract already has them.

## 4. What a birth may and may not say

| Announced (nameplate facts a real device would publish) | Never announced (sealed or simulator-internal) |
|---|---|
| Units, ranges, resolution, descriptions | Fault catalogue, onset, affected states |
| Command names, argument bounds, risk class | Expected effector, deadlines, counterfactuals |
| Entity kinds and topology links | Effect models, failure-mode probabilities, ack-latency distributions |
| Writable flags | True state, ground truth, profile weights |

The split is enforced by a test: a birth is built only from fields on an allow-list of the
domain schema, so a new sealed field can never leak by default.

## 5. Sequence, gaps and rebirth

- `device_data` and deaths carry a per-session sequence (64-bit, no wrap).
- The host detects gaps, duplicates and reordering from the sequence; on a gap or an unknown
  alias it sends `rebirth_request`, and the node republishes births before further data.
- This is where the **perturbation layer** bites: drops, duplicates, reordering, stale
  segments and store-and-forward bursts exercise the host's session handling, which is part
  of the system under test.
- `rebirth_request` and `host_state` are director-visible commands in the run's command log,
  so a run that included them replays exactly.
- A device reboot is a new `device_birth` with a new `boot_id`. Until the host reconciles
  (`state_query` → `state`), commands carrying the old `expected_boot_id` are refused
  (`wrong_boot`): the reconciliation barrier of the HIL program.

## 5a. Command safety (from device-wire-v1, kept end to end)

| Mechanism | Enforced by | Purpose |
|---|---|---|
| `expires_after_ms`, `not_before` | Gateway (server-expired work) and device (monotonic) | No stale command executes |
| Lease in arguments (e.g. `lease_ms`) | Device firmware | Output returns to safe state when the lease lapses |
| `expected_boot_id` | Device | No command crosses a reboot without reconciliation |
| `authority_epoch` | Gateway and device | A killed or replaced controller cannot act (split-brain rejection) |
| `idempotency_key` + bounded dedup ledger | Device / gateway | One semantic effect, even under duplicate delivery |
| `policy_digest` and capability-digest **allow-lists** | Gateway | The device side refuses a host running an unapproved policy |
| Hard limits | Firmware | Bounds narrower than any policy |
| Dedicated safe-stop path | Gateway and device | Safe stop never queues behind ordinary work |
| Receipt ≠ result ≠ observation ≠ verification | Host | Success is never inferred from an acknowledgement |

The simulator's field-side emulator enforces the same rules, and its protocol fault layer
(split/corrupt frames, reset mid-command, delayed receipt, repeated boot id, reconnect storm,
ack-without-effect) exercises them.

## 6. Liveness

The canvas party status (up / connected / active) comes from these records, not from
configuration:

| Status | Source |
|---|---|
| Node or device **connected** | Birth seen, no death since |
| Host **connected** | `host_state` online |
| **Active** | Last `command` or `device_data` time |
| **Up** (process) | Health endpoint from the canvas `connections.json` |

When the primary host goes offline (`host_state` offline), devices whose links declare
`store_and_forward` buffer and replay on reconnect, matching the existing link-delay model.

## 7. Permissions stay with the host

Births describe **what exists**. They grant nothing. Agentic Stream keeps a small,
human-owned policy that refers to discovered commands by name or pattern:

```yaml
default: deny
allow:
  - command: "*/emergency_shutdown"      # always permitted
  - command: "M*/set_line_rate"
    bounds: { pct: [50, 100] }
  - command: "P*/transfer_pump"
    approval: supervisor
```

Discovered but unmentioned commands are denied. Because Agentic Stream is part of the
system under test, the policy file is a **test input**: recorded with the run and shown in
the report.

**Defence in depth.** Host policy is not the only gate. The node (gateway or simulator)
allow-lists the host's `policy_digest` and capability digest and refuses commands from any
other policy; firmware hard limits sit below both. Births announce capabilities but never
lower these device-side gates.

## 8. What this replaces

| Today | After |
|---|---|
| `adapters/agentic-stream.adapter.json` projecting to `trace-record-v0.1` | Plant wire records; the adapter is deleted or reduced to a recorded-file view |
| Agentic Stream `internal/ingress/internal/domain/simulator_data.json` | Removed: data is self-describing via `device_birth` |
| Matching capability catalogs on both sides (digest-pinned) | Births announce commands; Agentic Stream keeps routes and policy; the node keeps the digest allow-lists (§7) |
| Device operations `set_led`, `set_pwm_lease`, `safe_stop` on the host link | Generic `command` by name on the host side; the same operations remain on the **field side** (device-wire-v1), unchanged |
| Gateway mapping file (`sensor_id` → schema ref, entity) | The gateway announces mapped sensors in `device_birth`; the mapping stays gateway configuration |

## 9. Determinism

Births, data and deaths are emitted by the world in deterministic order and recorded in the
delivery ledger; host-originated records (`command`, `host_state`, `rebirth_request`) are in
the command log. Replay reproduces the full wire exchange of a run made by the same binary.

## 10. Contract (decided 2026-10-12)

- **Neutral contract.** The plant wire is a generic plant protocol any host could implement,
  not "Agentic Stream's protocol". This removes the consumer-specific device contract from
  the simulator binary.
- **Two parts, used differently.**

  | Part | Used in | Purpose |
  |---|---|---|
  | Schemas (one per record type) | Production and tests | Embedded in each binary; every received record is validated and refused or quarantined if invalid |
  | Fixtures (valid/invalid records, scripted exchanges with expected reactions) | Tests only | Each side proves conformance offline: the simulator that it produces them, Agentic Stream that it accepts and reacts to them |

- **No shared code.** Each repository implements its own side; only schemas and fixtures are
  shared. One canonical copy, vendored into each repository and held byte-identical by a test
  (the `contracts/` ↔ `internal/schemas/` pattern).
- **Schemas describe shape, not inventory.** New sensors, device kinds and domains are data
  inside births: no contract change and no Agentic Stream change. A new command the agent
  should use needs only a policy line (deny by default). Only a new protocol feature (record
  type, value type, quality code) changes the contract.
- **Value types fixed from the start:** number, integer, boolean, enum, string.
- **Quality codes** taken from the bench telemetry contract (`arduino-telemetry-v1`), since they
  are already proven on hardware: `good`, `warming_up`, `suspect`, `stale`, `disconnected`,
  `out_of_range`.
- **Location (decided 2026-10-12):** for now, a self-contained folder
  `streams-simulator/contracts/plant-wire/` (schemas, fixtures, short spec; no code, no
  references to the rest of the repository). It moves to its own repository when there is a
  reason (e.g. the first customer gateway), by copying the folder and changing one env line.
- **Finding the master: env file.** Each repository has a committed `.env.example` and a
  git-ignored `.env` with `PLANT_WIRE_CONTRACT_DIR=<path>`. `make contract-sync` copies the
  schemas from that path; a drift test compares the local copy with the master when the
  variable is set and is skipped when it is not, so unit tests never need another checkout.
- **Contract handshake.** `node_birth` carries the plant-wire contract version and digest;
  the host refuses the session when it does not support them, and the node refuses a
  `host_state` that names an unsupported version. A stale vendored copy fails at connect time.
- **Registry later, if needed.** When several contract versions are live at once (e.g. customers
  on different gateway versions), a registry can serve versions and digests in place of folder
  sync. Programs still embed the versions they support; the registry is for discovery and
  checking, never for changing runtime behaviour.
- **Running programs never read the path.** Each binary embeds its copy: runs stay
  reproducible from the binary, a site's behaviour cannot change with a folder's contents,
  and each release is pinned to the contract it was tested against.

## 11. Open questions

1. ~~**Entity templates.**~~ Decided, see §14.
2. ~~**Alerts.**~~ Decided, see §15.
3. ~~**Live transport.**~~ Decided, see §13.

## 12. Acceptance tests from the real-world-sensor program

Source: `/Users/ghassan/my-projects/agent-research-lab/real-world-sensor` (round 1 demos,
round 2 `EXPERIMENTS.md`, `assessment/CURRENT-STATUS.md`).

**#1, always green:** the thermal LED/fan experiment (HIL-0, Experiment 02) on both rungs:

| Rung | Must keep passing |
|---|---|
| Emulated joined run (G1) | Simulator ⇄ plant wire ⇄ Agentic Stream ⇄ Tamoz: telemetry accepted with zero quarantine, one decision, bounded `set_indicator` / `select_thermal_mode`, executed exchange, reconciled state, artifact verdict `pass` |
| Physical bench | Arduino ⇄ device-wire-v1 (unchanged) ⇄ gateway ⇄ plant wire ⇄ Agentic Stream: the same exchange against real hardware |

It runs in CI on every change in any of the three repositories (emulated rung) and before
every hardware session (physical rung).

**Round 2 experiments as protocol acceptance:**

| Experiment | Plant-wire feature it proves |
|---|---|
| 2 Serial protocol torture | Field-side framing; nothing invalid crosses the gateway |
| 3 Time, reboot, offline buffer | `boot_id`, device monotonic time, sequence, store-and-forward |
| 5 One bounded verified effect | Command safety fields, lease, receipt/result/observation/verification |
| 6 False success | `command_result` that lies; verification against independent feedback |
| 7 Authority loss | `host_state`, `authority_epoch`, safe degradation |
| 8 Stale race, duplicates | Expiry, idempotency, reconciliation barrier |
| 9 Human override, interlock | Authority ordering: hard interlock > e-stop > local safety > approved command > recommendation |
| 10 Chaos soak | All of the above under scheduled faults |

**Round 1 demos as canvas worlds:** `temp-regulation`, `tank-level`, `leak-isolation`,
`motor-health` and `fault-injection-lab` are small worlds with a physical twin. They are the
natural first worlds for the builder, alongside the factory: each can run fully simulated or
with some devices bound to real hardware behind a gateway, because both speak the plant wire.

## 13. Transport (decided 2026-10-12)

- **One full-duplex session per node** carries everything: births, data, deaths, commands,
  receipts, results, state, host state and rebirth requests. It replaces today's two sockets
  (live ingress socket for data, device socket for commands).
- **Why:** one liveness (no half-connected state where the host acts blind or believes it is
  in control when it is not); one ordered stream, so a result and the observations that
  verify it arrive in true order; the whole session on one connection; one journal per
  session for exact replay. This is industry practice: Sparkplug B, OPC UA, Modbus TCP and
  DNP3 use one connection per node or device.
- **Priority lane:** each writer keeps two queues, control before data, so a safe stop,
  receipt or result never waits behind a data burst (e.g. a store-and-forward backlog).
- **Privilege separation:** Agentic Stream has one session component that owns the
  connection, hands data to ingress in-process and accepts commands only from the actions
  module; node-side digest allow-lists and the authority epoch back this up.
- **Safety stays off this link**, as in industry (hardwired e-stop, safety PLC, IEC 61850
  GOOSE): firmware hard limits, leases and safe state, and the simulator's hardwired trips
  work with the session down.
- **One mode, several carriers.** No two-socket mode. The records are carrier-independent:
  a local Unix socket now; TCP with mutual TLS later for a remote gateway.
- **Field side unchanged:** gateway ⇄ Arduino stays USB serial with device-wire-v1.
- **Migration:** the gateway (`arduino_gateway.py`) and Agentic Stream change together, landing
  with acceptance test #1 (§12) green on both rungs.

## 14. Entity templates (decided 2026-10-12)

- `node_birth` declares each **kind** once (metrics with units and ranges, commands with
  argument schemas); `device_birth` names its kind and carries only instance values.
- **Overrides may add or narrow, never remove or retype**, so "every `pump` has
  `motor_current`" always holds; the same rule applies in domain-spec v0.2.
- **System entities**: plant-wide measurements belong to devices of their own kind
  (`air_header`, `cooling_loop`, `power_bus`, `weather`, `tariff`, `control_room`); every
  metric belongs to exactly one device.
- **Expected inventory**: Agentic Stream's policy lists the devices and kinds it expects; a
  birth claiming an unexpected kind or id is quarantined and its commands denied, so a device
  cannot gain permissions by claiming a more permissive kind.
- **Quality codes** are the bench set (§10); a single failed sensor inside a live device is a
  quality code, not a `device_death`.
- Policy matches by kind first (`pump/transfer_pump`), by instance only for exceptions.

## 15. Alerts (decided 2026-10-12)

- Tamoz proposes `raise_alert` as an **intent**: same policy, rate limits, risk class and
  evidence trail as any action.
- Agentic Stream routes it to **`notify`**, its single path for messages to people, extended to
  carry agent alerts.
- Only the last hop differs: real recipients in production; in simulation, a `command` to the
  node-level **`people`** device, whose simulated responders react (see the scenario doc §5.3).
  The benchmark therefore exercises the production alert path.

## 16. Customers: shared contract, private configuration (decided 2026-10-12)

The contract and the three products are the same for every customer. Everything about a
customer's plant is configuration that belongs to that customer.

| | Shared (in the product repositories) | Per customer |
|---|---|---|
| Contract (`plant-wire`) | Record shapes, rules, fixtures | Nothing |
| Plant | — | Devices and metrics, announced in births (no configuration) |
| Edge gateway | Program | Sensor mapping file, allowed policy and capability digests |
| Agentic Stream | Program | `SituationSpec`, permission policy, expected inventory, approvers |
| Tamoz | Program | Nothing at rest; catalogs arrive with each episode |
| Simulator and canvas | Programs, generic example domains | Optional digital-twin domain, layouts, scenario files, run artifacts |

**Isolation rules**

1. **Customers never know about each other.** Each customer gets its own deployment: its own
   Agentic Stream process and database, its own Tamoz worker and memory, its own gateway.
   `tenant_id` remains as a second guard, not the isolation boundary.
2. **Customer data and configuration are never in GitHub**, public or private: no specs,
   policies, mappings, twin domains, run artifacts, journals, logs or screenshots. They live
   in a customer workspace outside every repository (on the customer's machines or a private
   store), and the tools take its path (`--domains-dir`, the canvas workspace, Agentic Stream
   config paths).
3. **Nothing learned at one customer flows to another.** Tamoz memory, self-improvement
   candidates and proposed rule changes are per customer.
4. **Published evidence uses only synthetic worlds.** Benchmarks, examples and fixtures in the
   repositories are generic; nothing is derived from a customer's plant.
5. **Contract versioning starts with the first customer.** The no-compatibility rule ends when
   customer gateways are in the field; records already carry `protocol_version`.

Guardrail to add when the first customer exists: a repository check (pre-commit and CI) that
refuses customer workspace markers, so a misplaced file cannot be committed.
