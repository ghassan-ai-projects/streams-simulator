# Physical devices — one device model, simulated and real implementations

- **Status**: design agreed in discussion with the owner, 2026-10-17, branch `simulator-new-vision`. Not built yet.
- **Source**: the owner's product direction "From Simulated Streams to the Physical World" (2026-10-17), and its principle:
  *the real device is not an addition to the simulator; it is another possible implementation of a device that the simulator
  already knows how to represent.*
- **Builds on**: [plant wire v1](../../contracts/plant-wire/SPEC.md), [the plant wire in the simulator](../simulation-canvas/plant-wire-in-simulator.md),
  [plant wire protocol decisions](../simulation-canvas/plant-wire-protocol.md), the device emulator (`internal/device`), the
  world bridge (`internal/deviceworld`), the canvas, and the thermal LED/fan experiment (acceptance test #1, green throughout).

## 0. Summary

The plant wire was designed so that a node is either a simulated world or an edge gateway in front of hardware (SPEC §1), and
the device emulator speaks device-wire-v1, which the Arduino firmware speaks. The direction therefore extends the existing
design rather than replacing it. Decisions taken:

| # | Decision | Section |
|---|---|---|
| D1 | The device model is the plant wire kind, written as a kind file with a `kind_digest`; both implementations must present it | §1 |
| D2 | **Relay mode**: in the lab the gateway talks to the simulator and the simulator to Agentic Stream; real devices pass through untouched | §2 |
| D3 | What a real board reports is a **recorded input** to the run, so a run stays a pure function of its inputs | §3 |
| D4 | The canvas becomes a **live control plane**: state, data, commands and results; pause, block, emergency stop | §4 |
| D5 | The Arduino thermal bench is the first proof; the co-founder's board is the second, as a new kind file with no workflow change | §6 |
| D6 | Bench only: one desktop, local, one user, no authentication, until the system is proven | §9 |

## 1. The unified device model: the plant wire kind

### 1.1 Today

| Description | Where | Says |
|---|---|---|
| Domain kind (channels, effectors, dynamics, faults) | `domains/*.domain.json`, `model.Kind` | Interface **and** behaviour |
| Component (equipment, instruments, vocabulary) | `internal/component/.../builtin` | Reusable parts that `Compose` turns into a domain |
| Device capability catalog | `internal/device/testdata/thermal_capability_catalog.json` | device-wire-v1 routes, presets, bounds, safe stops |
| Binding catalog | `internal/deviceworld/testdata/thermal.bindings.json` | Device target → world effector |
| Plant wire kind | `NBIRTH.kinds` (SPEC §5), built by `plantwire` `KindTemplates` | Metrics and commands as a host sees them |

Only the plant wire kind is what a host sees, and it is the only one a gateway and a simulated node both emit.

### 1.2 The model

**A kind is the interface; a domain is one implementation, a gateway with its hardware another.**

- **Kind file.** `*.kind.json` holds exactly one `NBIRTH.kinds` entry (metrics with type, unit, ranges, scan class, writable;
  commands with argument schema, `risk_class`, `safe_stop`), validated by the plant wire kind sub-schema. Its identity is
  `kind_digest`, the SHA-256 of its RFC 8785 canonical JSON. `capability_digest` covers only `{kind, commands}`, so two
  implementations with different metrics would share it; it is not enough for "same model".
- **Simulated implementation.** A domain kind declares `implements: "<kind file>"`. The domain module checks that
  `KindTemplates` of that kind equals the kind file after canonicalisation; a mismatch is a validation error, so "the same
  model" is a gate, not a promise.
- **Physical implementation.** The gateway is configured with kind files and its field mapping (for device-wire-v1, today's
  capability catalog). It announces the kind file's content; it never invents metrics or commands.
- **Board.** `boot_id`, leases and safe state stay board facts (SPEC §1). A lost link is `DDEATH link_down`; a failed sensor on a
  live board is a quality code (SPEC §10).

### 1.3 What the model never says

Births carry no faults, onsets, effect models, failure probabilities, latency distributions or truth (SPEC §5), so neither does a
kind file. Behaviour belongs to the implementation: the domain's dynamics and fault catalogue, or physics. The same kind can
therefore sit on both sides without leaking the oracle.

### 1.4 The other descriptions

Domain kinds keep behaviour and gain `implements`. Components stay the authoring path. The capability catalog stays the
gateway's field mapping below the plant wire (SPEC §13); device-wire-v1, its safety fields and the firmware do not change.
Bindings stay the emulator's bridge for field-side protocol-fault testing.

## 2. Relay mode

```
Direct (production shape):  Arduino ⇄ serial ⇄ gateway ⇄ plant wire ⇄ Agentic Stream ⇄ Tamoz
Relay (lab):                Arduino ⇄ serial ⇄ gateway ⇄ plant wire ⇄ streamsim ⇄ plant wire ⇄ Agentic Stream ⇄ Tamoz
                                                                         │ also runs simulated boards
                                                                         └→ canvas control plane
```

The simulator is the gateway's host on one side and one node to Agentic Stream on the other. It passes the gateway's records up
and Agentic Stream's commands down, records both, and adds its own simulated boards beside the real ones. Agentic Stream sees
one node and needs no change for mixed sessions.

Rules:

- **Real devices pass through untouched.** No perturbation, wire fault or fault injection acts on a hardware device's records or
  commands. Only the gateway touches hardware; the binary never speaks serial or BLE.
- **Replay never reaches hardware.** A replay or re-run never opens the gateway link; this is structural (replay has no gateway
  option), not a convention.
- **Origin per device.** `NBIRTH.environment` is per node today; a mixed node marks each board `simulated` or `hardware`
  (contract change). A host and the canvas never infer origin.
- **The firmware is the last line of defence.** If the simulator stops, leases expire and the board makes itself safe.
- **Independence.** The gateway and the simulator speak the same plant wire, so Agentic Stream can always connect to the gateway
  directly. Relay mode is a lab choice, not a dependency.

## 3. Real data as a recorded run input

A run is a pure function of `(sim_version, domain_digest, adapter_digest, seed, command_log, sink)`. Relay mode adds one input:
the **field trace**, every record the gateway sent, in order, as the simulator received it. A run becomes a function of
`(…, seed, command_log, field_trace)`.

- **Determinism holds without an exception**: replaying with the same field trace reproduces the same bytes, and one artifact
  reproduces an incident including its physical part.
- **Truth**: simulated devices keep sealed, generative truth and scoring. A hardware device has none; a `DRESULT` is
  self-reported and may lie (SPEC §9). A physical effect counts only when an independent instrument observes it.
- **Scoring**: for hardware devices, a consumer is measured on its response to what it observed, never on whether it was right
  about the world.
- **Twin versus board**: running the domain that implements the kind with the recorded command log and reporting where it
  diverges from the field trace beyond declared tolerances measures the twin, not a consumer. The analytic cross-check still
  guards the twin.

## 4. The control plane

The canvas gains a live screen for a running session, simulated, real or mixed, fed by the simulator over the existing bridge
(HTTP JSON and Server-Sent Events). It shows each device's current values and quality, its origin, births and deaths, the
commands sent with their receipts, results and observed effects, and a timeline that follows one reading from the board to
Tamoz's decision. A finished session opens in the same screen from its recording.

Interrupts act through the party that owns them:

| Action | Effect | Carried out by |
|---|---|---|
| Pause the agent | Tamoz's commands stop being executed | Agentic Stream, which holds the authority (`authority_epoch` revocation); the canvas asks it |
| Block a pending command | That command is refused before it reaches the device | Agentic Stream |
| Emergency stop | Selected devices go to their safe state | The canvas directly, through the relay: `safe_stop` is always accepted and always safe |

The canvas sends no ordinary command itself. For simulated devices the same buttons act on the simulated boards.

## 5. Safety

- **Read-only first.** A hardware board announces its kind's commands, but Agentic Stream's commissioned inventory denies them
  until M5.
- **Transport.** v1 is Unix socket only, owner-only (SPEC §2): gateway, simulator and Agentic Stream run on the bench machine.
- **Device-side safety.** Expiry, lease, `boot_id`, idempotency, policy digest and safe stop remain device-wire-v1 duties in the
  firmware; neither the gateway nor the simulator stands in for them.
- **Thermal experiment.** Every milestone keeps the emulated joined run and the physical Arduino bench green.

## 6. Reference devices

**First: the Arduino thermal board** (ELEGOO Mega, DHT11 on D2, LED, fan through an L293D; gateway
`agent-research-lab/real-world-sensor/assessment/tools/arduino_gateway.py`). Its kinds come from the existing thermal catalog and
device-wire-v1 operations; exact names are fixed in M1:

| Kind | Metrics | Commands |
|---|---|---|
| Thermal sensor | `temperature` (`Cel`), `humidity` (`%`) | – |
| Indicator | `brightness_permille`, `pattern` (writable) | `set_led`; `safe_stop` |
| Fan | `duty_permille` (writable, 0–600) | `set_pwm_lease` (`lease_ms` 5000–10000); `safe_stop` |

**Second: the co-founder's board.** A new kind file and a gateway mapping for its field link; a simulated twin built from the
kind file before the hardware is ready, so firmware and Tamoz workflows are developed in parallel against the same spec. The
proof is that M1–M5 repeat with no change to Tamoz or Agentic Stream workflows.

The model is behavioural: no instruction-level emulation of either microcontroller or its firmware.

## 7. Roadmap

Each milestone is shippable on its own and ends with the thermal experiment green.

| Milestone | Deliverables | Exit |
|---|---|---|
| **M1 — Arduino twin** | Kind file contract and `kind_digest`; `implements` check in the `domain` module; thermal kind files; a domain implementing them, as data | A simulated thermal board runs over the plant wire to Agentic Stream and Tamoz, deterministic on replay; a drifted domain kind is rejected |
| **M2 — Gateway on the plant wire** | The Arduino gateway gains a plant wire side announcing the same kind files; Agentic Stream commissions the board read-only | Same `kind_digest` from gateway and simulator; real data ingested with zero quarantines; unplugging the board shows `link_down` and a rebirth |
| **M3 — Relay and mixed sessions** | Relay mode (§2), per-board origin, field trace as run input (§3), simulated boards beside the real one | The real board plus N simulated boards reach Agentic Stream as one node; a recorded session replays byte-identically; a test proves no replayed or simulated command reaches the gateway |
| **M4 — Live control plane** | Canvas live screen (§4): values, quality, origin, commands and results, timeline | One reading followed from the board to Tamoz's decision on the live screen; the same session reopened from its recording |
| **M5 — Interrupts and real commands** | Emergency stop through the relay; pause and block through Agentic Stream; fan and LED commands commissioned; an independent instrument on the bench; twin-versus-board report | A Tamoz decision drives the real fan, confirmed by the instrument rather than `DRESULT`; pause, block and emergency stop each shown on the bench; the report names the divergences |
| **Second proof** | The co-founder's board through M1–M5 | No change to Tamoz or Agentic Stream workflows |

## 8. Changes by repository

| Repository | Change |
|---|---|
| streamsim | Kind files and `implements` (M1); relay, field trace, mixed node (M3); live screen (M4); emergency stop, twin report (M5) |
| Plant wire contract | `kind_digest` in `NBIRTH` (M1); per-board environment (M3) |
| Agentic Stream | Plant wire host for the thermal kinds (M1–M2); operator pause and block (M5) |
| Tamoz | Workflows on the thermal kinds, unchanged across simulated, real and mixed |
| Gateway and firmware | Plant wire side of the Arduino gateway (M2); an independent effect instrument (M5) |

## 9. Deferred

- **Time in mixed sessions.** Agentic Stream can hold a simulated world at each barrier; a real board cannot be held. How the
  world is paced when a real board is present is decided when M3 starts.
- Authentication, several users, customer sites and production reliability of the control plane (D6).
- Remote gateways: TCP needs mutual authentication in a new contract version.
- Incident reproduction with faults injected on a recording; more boards; fleet-scale mixed runs.

## 10. Not doing

- Instruction-level emulation of a microcontroller or its firmware.
- A serial or BLE stack, broker or device-specific code in the binary.
- The simulator as a policy authority, or as a sender of ordinary commands to hardware.
- Scoring a consumer against a device's self-reports.
- Customer devices, mappings or recordings in any repository.
