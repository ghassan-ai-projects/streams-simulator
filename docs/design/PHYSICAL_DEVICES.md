# Physical devices — one device model, simulated and real implementations

- **Status**: design agreed in discussion with the owner, 2026-10-17, branch `simulator-new-vision`. M1 is built; §1.2, §1.5 and
  the "As built" note at the end of §1 say what exists. M2 onwards is not built.
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
| D4 | The canvas becomes a **live control plane**: state, data, commands and results; interrupts are a **local/remote switch at the node**, which may only refuse or stop, never originate or widen | §4 |
| D5 | The Arduino thermal bench is the first proof; the co-founder's board is the second, as a new kind file with no workflow change | §6 |
| D6 | Bench only: one desktop, local, one user, no authentication, until the system is proven | §9 |
| D7 | A device is **defined** as an interface (kind file) plus optional simulated behaviour (component), in a canvas device editor or captured from a connected board | §1.5 |

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
  `kind_digest`, the SHA-256 of its RFC 8785 canonical JSON of the entry **without** that member; a kind file never carries
  `kind_digest` (the schema forbids it, so a stale digest cannot be accepted and ignored), and a metric's ranges need `min`
  below `max`. `capability_digest` covers only `{kind, commands}`, so two implementations with different metrics would
  share it; it is not enough to identify a model.
- **Simulated implementation.** A domain kind declares `implements: "<kind file>"`. The check compares `KindTemplates` of
  that kind with the kind file after canonicalisation and names the first differing member; a mismatch is a validation
  error, so equality with the kind file is a gate, not a promise. A domain effector flags a safe stop with
  `safe_stop: true`; `KindTemplates` announces that flag, so safe stop is part of the equality, and a plant wire node takes
  its safe stops from the domain (it has no option of its own that could drift from the kind file). The check runs in
  `domain validate`, `domain compose`, `run`, the director (`sim.world.create`), `kind twin` and `plantwire node`; recorded
  runs and `domain lint` do not run it (see `governance/DEFERRED.md`).
- **What the kind file proves.** The kind file is the contract. `implements` proves that a domain equals the file. A device
  (a gateway in front of hardware) matches it only if it announces a `kind_digest` equal to the file's, which a host or the
  gateway's commissioning checks (M2); nothing else claims "the same model".
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

### 1.5 Defining a device

Today the canvas places equipment from the built-in library only; a user cannot define a new device. A device definition has two
parts:

| Part | Says | Stored as | Needed for |
|---|---|---|---|
| **Interface** | The board, its devices, each device's readings (name, type, unit, range, resolution, scan rate) and commands (arguments, bounds, risk class), safe stop | Kind files | Real **and** simulated: this is the device model |
| **Behaviour** (optional) | How the twin behaves: each reading's instrument (noise, drift, faults, cadence) and each command's response | A component built from library instruments | Simulated only |

A device with an interface only can be connected as hardware but not simulated; with behaviour it can also run as a twin. The
binary checks that the behaviour implements the interface (§1.2), so the two cannot drift.

**Command responses are templates, as data.** A first set covers the reference boards: a first-order change of a reading
toward a target (fan duty lowers temperature with a time constant), an on/off state, and a setpoint that a reading follows.
They are expressed in the domain spec's existing dynamics, not as new code; a response that needs more is a domain-spec
change first.

**Binary work** (the canvas computes nothing the binary owns):

- kind files in the workspace (`kinds/*.kind.json`) with a validate action (`kind.validate`);
- workspace components (`components/equipment/*.component.json`, `components/instruments/*.instrument.json`), read and
  written through the bridge, and handed to every command that loads the library as `--components-dir`;
- the "behaviour implements interface" check;
- the response templates, as `kind twin`: it scaffolds a twin from kind files and a small choices document in the
  workspace folder `twins/` (`twin-choices-v0.1`: per metric the library instrument or `none`, per command one of `direct`,
  `first_order`, `none`), through the bridge action `kind.twin`;
- structured issues: every refusal of a kind file, a choices file or the check is `{pointer, message, keyword, member?}` per
  violation, a pointer into the file (or `/kinds/<i>/implements` in a domain), the schema keyword or rule code that failed and
  the named member, so the canvas shows the member and parses no text; the twin requirements of a kind file are all reported at
  once (`kind validate --twin`, bridge `kind.check`), the binary's defaults for a twin come from `kind twin --suggest` (bridge
  `kind.suggest`) and the instruments of a unit from `component list --unit` (G3);
- capturing a kind from an `NBIRTH` as a draft kind file: `kind capture`, bridge `kind.capture` (see §1.6, R6a).

**Capture from a connected board.** When a real board announces itself, its kinds can be saved as draft kind files. The
author reviews them, then adds behaviour. This follows the contract's commissioning order (discovered → reviewed → accepted):
a birth is a claim, and a captured kind becomes the model only after review. It is the expected path for the co-founder's
board: connect it, capture its interface, build its twin.

The gateway's field mapping (which pin drives the fan) stays in the gateway's own configuration.

### 1.6 As built (M1)

Deviations from the plan above:

- The thermal board's LED `pattern` metric is an integer code (0 off, 1 slow blink, 2 solid) in the example twin
  (`examples/devices/thermal-board`), not a string: string channels draw random enum values, so a twin could not follow a
  commanded pattern. The gateway maps the code to the firmware's pattern names.
- `plantwire node` has no `--safe-stop` option; its safe stops are the domain's `safe_stop: true` effectors, and it takes
  `--kinds-dir` and runs the check before connecting.
- The director fails closed: a domain that implements a kind file is refused by `sim.world.create` unless the director was
  given the kind files (`--kinds-dir`); the bridge passes the workspace `kinds/` folder when it exists.
- Not yet built: the gateway (M2), and the loaders listed in `governance/DEFERRED.md` that skip the check on purpose.

### As built (R6a): capture from a birth

- `kind capture` (bridge `kind.capture`) turns one `NBIRTH` into draft kind files and compares them with the workspace's
  kind files (`new`, `same`, `differs` at the first differing pointer). It is read-only: reviewing and saving is the author's
  step, as the contract's commissioning order (discovered, reviewed, accepted) has it. A birth is a claim, so every
  `kind_digest` is recomputed and a birth that does not hold is refused; a birth of any environment is accepted and echoed,
  unless its node environment contradicts its boards (R7a, §2).
- The source is a bare `NBIRTH` or the first one of a JSON Lines file, including the `wire` records of a run's observer
  journal (`direction: node`); only the observer stream is taken, read strictly, and provenance is reported (`births_seen`,
  `taken`, `later_births`, `rebirth_differs`). `environment` is the board's claim and is not verified; `claims` lists the
  members the board asserted (`risk_class`, `safe_stop`, `args_schema` bounds) for the reviewer. Each `comparison` carries
  `target`, the path the draft would be saved to, so a surface guesses no file name. Deviation from the plan: a file is
  streamed (a journal is larger than a kind file), a line at most 1 MiB and a scan at most 32 MiB; standard input stays at 1 MiB.
- The scripted stand-in gateway is test support, not a feature: `internal/plantwire/internal/standin` plays a hardware node
  from kind files and a script (readings and qualities, link loss and rebirth, command receipts and results) in script time,
  and a test (`test/architecture`) keeps every production package from importing it. A board built from the thermal kind files
  is accepted by the reference host, and capturing its birth yields kind files equal to the originals (equal digests).

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

  *As built (R7a, contract 1.2.0):* every board in `NBIRTH.boards` carries `environment` (`simulated` or `hardware`, required)
  and the node's `environment` gains `mixed`. A device's origin is the environment of the board its `DBIRTH` names, and
  nothing else. The reference host checks, with one predicate it shares with `kind capture`, that a `simulated` node has only
  simulated boards, a `hardware` node only hardware boards and a `mixed` node at least one of each; a contradiction is refused
  with `environment_mismatch` even by a `mixed` host, and the host's report says which rule (`reason`). The host's own setting
  has the same three values: `simulated` and `hardware` accept only a node of exactly that environment, `mixed` accepts all
  three (`plantwire host --environment`). A simulated world is always `simulated`: its boards are, and `plantwire node` and the
  session options refuse any other environment, so a simulated node can never announce hardware or mixed. `kind capture`
  echoes each board's claim (`boards: [{board, environment}]`) beside the node's. Not built here: the relay and the mixed
  node itself (R7b onward); nothing in a simulated run's artifact or digest changed, only the `NBIRTH` bytes in a wire journal.
- **The firmware is the last line of defence.** If the simulator stops, leases expire and the board makes itself safe.
- **Independence.** The gateway and the simulator speak the same plant wire, so Agentic Stream can always connect to the gateway
  directly. Relay mode is a lab choice, not a dependency.

### As built (R7b): the relay

`streamsim plantwire relay --gateway-socket <path> --host-socket <path> [--out <dir>]` is the lab shape for **hardware only**:
gateway, relay and host, the relay being the host's one node. Simulated boards beside the gateway and the field trace as a run
input came in R8 (below); clock pacing is not built (D-71); hold, approve and e-stop came in R11 (the last section).

- **Two roles, one session each.** Toward the gateway the relay is a host: it creates the gateway socket (owner-only, the
  reference host's helper: a live socket, a regular file or a group/world-accessible directory is refused), reads the gateway's
  `NBIRTH` and applies the reference host's handshake checks, one predicate (`domain.Admission`) shared with `plantwire host`
  and fixed to a `hardware` host: protocol range, contract version and digest, node and board environments (a `simulated` or
  `mixed` gateway is refused `environment_mismatch`, the report says which rule), `kind_digest` recomputation and a kind named
  twice (`inventory_mismatch`). A refusal is a `STATE` with the refusal and a close; the relay keeps listening. Toward the real
  host the relay is a node: it dials the host's owner-only socket only once a gateway has been admitted (a socket others can
  reach ends the relay with `ErrSocketNotPrivate`) and presents itself as one node.
- **What passes through, with content unchanged.** The gateway's `NBIRTH` becomes the relay's own `NBIRTH`: kinds with their
  `kind_digest`, `capability_digest`, boards (each with its own `environment`), node metrics, contract and environment are the
  gateway's; so the host commissions the digests the gateway announced. `DBIRTH`, `DDATA`, `DDEATH`, `DSTATE`, `DRECEIPT`,
  `DRESULT` and `NSYNC` go up, and `STATE`, `NCMD`, `DCMD`, `DQUERY` and `HSYNC` go down, with every member except the envelope
  as received: values, qualities, event times, boot ids, command ids, reasons and all the safety members of a `DCMD`
  (`expected_boot_id`, `deadline_ns`, `not_before_ns`, `authority_epoch`, `policy_digest`, `idempotency_key`,
  `fallback_on_lease_loss`). The gateway therefore sees the real host's authority epoch and heartbeat.
- **What the relay owns.** Only envelope members and its own sessions: the `session_id` of each link (`relay-h<n>` toward the
  host; the gateway's own toward the gateway), `seq` assigned at transmission in the order records arrive, `bd_seq` and
  `prior_last_seq` of its `NBIRTH`, and the `protocol_max` it can speak (1). It originates exactly: `NCMD rebirth` to the
  gateway (when a host session starts after the gateway was born, and once per gap in the gateway's `seq`, the reference
  host's rule; the record after the gap is still carried), a `STATE {online: false}` to the gateway when the host is lost or
  was never there (the host's last `STATE` with `online` false, or its own with `host_id: relay`, epoch 0 and no refusal),
  a `DDEATH link_down` per born device when the gateway is lost, and `NDEATH shutdown` when the relay ends. It never originates,
  answers, refuses or delays a `DCMD` or `DQUERY`, sends no data of its own, rewrites no quality and keeps no record back.
- **Sequences and the barrier.** Per carried record the relay keeps the pair (gateway `seq`, relay `seq`), the last 4096 of
  the host session. A gateway `NSYNC` goes up as the relay's own with `through_seq` translated to the relay's `seq` of the last
  carried record at or below it; the host's `HSYNC` comes back with `through_seq` translated to the gateway `seq` of the last
  carried record whose relay `seq` is at or below it, so an answer to an earlier barrier never covers a later one. A barrier
  no carried record covers is dropped, never invented. A hardware gateway runs in wall time and sends none (SPEC §8); the
  mapping is for a gateway that does.
- **Failures.** Host lost (end of stream, a gap in its `seq`, an unnegotiable `STATE`, or three missed heartbeats, counted in
  250 ms ticks once the host has sent a heartbeat above zero): the gateway is told the host is offline and its session is
  kept, so the firmware's lease logic applies; its records meanwhile are dropped (store-and-forward is the gateway's job) and
  the relay redials on every tick. When the host is back the relay sends the gateway's cached `NBIRTH`, forwards the host's
  `STATE` and asks the gateway for a rebirth, which goes up in the same host session with the next `bd_seq`. A host that
  refuses the relayed session has its `STATE` forwarded to the gateway and ends the relay (`ErrRefused`, an operator's action).
  Gateway lost (end of stream, a handshake silent for 10 s, a line over 1 MiB, or its own `NDEATH`): a `DDEATH link_down` for
  each device it had born, in device order, and the host session is kept, because later boards share it. A returning gateway
  is answered with the host's last `STATE` and reborn upward as a rebirth.
- **Dropped, counted, never forwarded.** A record that fails its schema or its size limit, repeats a `seq`, has none, belongs
  to another session or version, or is of a type that does not travel that way (a gateway's `DCMD`, `STATE`, `NCMD`, `DQUERY`,
  `HSYNC`; a host's data) is dropped and counted in `dropped`; a bad `DCMD` is not answered with a receipt, the relay never
  answers a command.
- **Observation.** With `--out` the relay writes every record on both links as observer `wire` session records of the run
  journal (`journal.jsonl`; the director stream holds only its closing record, since hardware has no truth): direction
  `gateway` (received from the gateway) and `node` (sent to the host), `host` (received from the host) and `relay` (sent to the
  gateway). A carried record is two consecutive lines, as received and as sent, so its own `seq` and the relayed one are both
  there; a dropped record has only the first. `kind capture <out>/journal.jsonl` reads the relay's `NBIRTH`.
- **Structure.** The rules are in the `plantwire` domain layer (`relay*.go`, pure: a state machine of events and moves), the
  loop in the app layer, the sockets and ticks in the `uds` edge (`gateways.go`, `ticks.go`), the journal through
  `run.WireTrace`. `RelayOptions` has no fault, perturbation or script option, and `TestTheRelayImportsNoPerturbationOrFaultCode`
  keeps the relay's files from naming perturbation or wire fault code; `replay`, `run`, `scenario` and `plantwire node` have no
  gateway option (a CLI test), so replay can never reach hardware.

### As built (R8): mixed sessions

`streamsim plantwire relay ... --domain <id> --out <dir>` puts the simulated world beside the gateway in the one node the real
host sees. The hardware-only relay of R7b is unchanged (a relay without `--domain` has no world and no router).

- **One birth, two origins.** The relay holds the gateway's `NBIRTH` and the world's (the world's node `Session` builds its own
  before it runs; the relay takes it as data). The `NBIRTH` the host sees is the gateway's content with the world's boards
  appended, `environment: mixed`, the kinds of both (a kind both sides announce appears once and only when its `kind_digest`
  is equal; each entry keeps its own digest) and a `capability_digest` recomputed over the union. A gateway that announces a
  board, a device or a kind with another digest that the world also has is refused `inventory_mismatch` at the handshake, the
  rule in `reason` (and a gateway rebirth that now collides is dropped). A gateway that is itself `mixed` or `simulated` is
  refused as before. The origin of a device is the board that announced it; a record that names a device its side did not
  announce is dropped (a gateway cannot speak for a simulated device and the other way round).
- **Routing, down.** A `DCMD` or `DQUERY` goes to the side of its device: hardware, to the gateway with every member as
  received; simulated, to the world, which answers it as it answers any host (freshness, epoch, policy, effector, receipt and
  result; the command enters the run's command log); a device on no board is refused `unknown_device` by the relay (a query is
  dropped). `NCMD rebirth` goes to the gateway only; when the gateway's `NBIRTH` answers it the merged `NBIRTH` goes up and the
  world is asked to birth its devices again, so the host sees the births after the birth. An `HSYNC` goes to the side whose
  `NSYNC` it answers, mapped to that side's own `seq` (a second table of pairs for the world); an answer no barrier waits for is
  dropped. Every host `STATE` also goes to the world; the first one of a host session starts the world's session (or asks it to
  birth again).
- **Direction rules are types.** `Relay.FromSim` returns a `SimMoves` (records for the host and lines for the journal): it has
  no member for the gateway, and the loop applies it with a function that never touches the gateway link. The world's wire
  faults and perturbations stay inside the world's session; the relay's files may not name them (the R7b gate covers the new
  files too).
- **The world's session.** It is the node `Session` over an in-process link: mailboxes in both directions, the barrier wait
  capped in wall time, run on one goroutine (`uds/sim.go`). Its own wire journaling is off (`Options.NoWireJournal`); the relay
  journals the merged stream (`gateway`, `relay`, `node`, `host`) through the run's journal. The world is advanced by the
  barrier exactly as `plantwire node` does when the host sends `HSYNC`; **it is not paced to the wall clock** (deferred D-71).
- **The run.** `--out` holds an ordinary run folder (`run.json`, trace, ledger, history, both journals) plus
  `field-trace.jsonl`; `run.json` records `field_trace_digest`.

## 3. Real data as a recorded run input

*As built (R8):* the field trace is `field-trace.jsonl` in the run folder: one line per record the gateway sent, in the order the
relay received it, `{"ordinal": n, "record": {...}}` in RFC 8785 canonical JSON and nothing else (no host record, no wall clock;
the times in it are the records' own event times). The relay writes it as it relays, including records it dropped, so the file
is what the gateway said and not what the relay decided. The run module owns it (`run.NewFieldTrace`, `run.ReadFieldTrace`):
the artifact gains the optional `field_trace_digest` (the SHA-256 of the file, which is the digest of the canonical form of
the ordered records), `replay`, `verify` and `report` check the file against it with the other evidence files, and a run
without a trace has neither the member nor the file. A mixed run is therefore a function of
`(…, seed, command_log, field_trace)`: the world's side reproduces from the command log, the gateway's side is the trace, and
tampering with either fails verification. `plantwire replay-trace` plays a trace to a host. **It announces the node and every
board `simulated`**, because a replay is not a board and no consumer may take it for one, whatever the recording said; it
keeps the recorded device ids, kinds, kind digests and values, has no gateway option or code path, refuses every command it is
sent, and raises its own barrier after each stretch of records (the recorded `NSYNC`s only mark where) so that its stream is a
pure function of the trace and the host's answers. It refuses a trace whose digest `run.json` does not record. Not built: the
world's part of a mixed recording streamed to a host beside the trace (deferred D-72); it is reproduced and verified by
`replay`.

The design as first written:

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

The Review screen is already most of a control plane: the plant graph with readouts, a timeline with fault, alarm and
effector-call marks, Inspect, and a bridge that streams a run's observer journal live (`GET /api/runs/{id}/events`). The
control plane extends it rather than adding a dashboard beside it.

### 4.1 Canvas changes

| Change | What | Milestone |
|---|---|---|
| **Devices** | A Devices section in the workspace beside Plants (built-in and own devices); a device editor in three steps: interface, behaviour, check (§1.5); "save as device" for a connected board; Add equipment lists own devices | M1, capture M2 |
| **Connect session** | The Run panel gains an open-ended session: no duration, linked to Agentic Stream and optionally a gateway, until stopped; Stop ends and saves. Today a run always goes to completion (UI-EXPECTATIONS RUN-2) | M3 |
| **Real or simulated per session** | The plant does not change; the session plan says which instances are real boards ("TH-1 is the Arduino on the gateway"). The binary checks the kind match; the canvas shows the result | M3 |
| **Origin everywhere** | A real/simulated mark on nodes, outline rows, timeline marks and the inspector header ("Real · board mega-01 · boot 61"), by shape and colour, never colour alone; new origin tokens | M3 |
| **Live mode** | Review follows the newest record while a session runs; scrubbing back pauses following, "Back to live" resumes; an ended session is an ordinary review | M4 |
| **Command lifecycle** | Each command shows its stages separately: sent, accepted or refused with the reason, the board's own report, the effect observed or unknown. A commands table and richer timeline marks; "acknowledged" never reads as "done" | M4 |
| **Connection health** | Gateway and host link state, a board going offline and coming back, reboots (`boot_id` changes), stale and disconnected readings shown as such, never as normal values (covers V-8) | M4 |
| **Interrupts** | Hold, approve each command and emergency stop (§4.2), always visible on the live screen | M5 |
| **Twin overlay** | In Inspect, the twin's prediction drawn over the real trend (§3) | M5 |

M1 changed this plan only where Go output changed: the bridge workspace listing shows a kind's `kind`, a component's `id` and
a `title` only when they are plain text of the expected form, and only the path of a file over 256 KiB; the Devices
section reads the listing and the structured issues, nothing else.

End-to-end tests run against the real binary as today, with a scripted stand-in gateway; the repository forbids mocking the
binary, not the hardware. The real board is checked in each round's bench design review.

### As built (R9): connect session, real or simulated per session, origin marks

Design, written before the code (R9, milestone M3).

**The session plan** is a new document the binary reads, so it is a contract: `contracts/session-plan-v0.1.schema.json`
(copied to `internal/schemas`, kept equal by the schemas test; typed and schema-checked in `model.DecodeSessionPlan`,
a violation is the `plantwire_failed` failure with its issues, as kind files are `kind_failed`, so the CLI gains no new
failure code):

```json
{"schema_version": "0.1", "domain": "arduino-thermal", "seed": 1,
 "host_socket": "/abs/path/host.sock", "gateway_socket": "/abs/path/gateway.sock",
 "real": [{"instance": "zone-01", "board": "uno-07", "device": "zone-01"}]}
```

`gateway_socket` is optional; `real` binds a plant instance to the device a gateway board announces ("zone-01 is device
zone-01 on the board uno-07 behind the gateway"). The plant does not change: every instance not in `real` is simulated.
The plan travels on standard input (`--stdin`); nothing is written to the workspace, and the canvas builds the document
from its form only (it decides nothing in it).

**Where each rule lives** (the canvas computes none of them):

| Rule | Module, layer |
|---|---|
| The plan's shape | `model` (schema `session-plan-v0.1`) |
| An instance named in `real` exists in the domain (`unknown_instance`), its kind implements a kind file (`no_kind_file`: only an interface can be matched), an instance or a board's device bound twice (`instance_bound_twice`, `device_bound_twice`), `real` without a gateway (`gateway_required`), one socket for both (`same_socket`); each issue a pointer into the plan | `plantwire` domain (`connect_plan.go`) |
| The kind match at the gateway's handshake: the bound board is announced (`board_missing`), holds the device (`device_missing`), the gateway announces the instance's kind (`kind_missing`) with the kind file's `kind_digest` (`kind_mismatch`); a refused gateway gets `inventory_mismatch` with the first issue as reason, through the existing admission path, and the relay keeps listening | `plantwire` domain (`relay_bindings.go`, the existing R8 relay) |
| A bound device born as another kind is still carried unchanged (gateway content passes through after admission, R7b); it is reported as a refused `binding` (`device_kind_mismatch`) and its `origin` names no instance | `plantwire` domain |
| The bindings reach the relay through the world's node options (`Options.Bound`, inside `SimulatedOptions.Node`): they decide what the world withholds and what the handshake checks, and `RelayOptions` keeps no option that acts on a gateway's records | `plantwire` app |
| A bound instance is withheld from the world's births, data and boards (a board left without members is not announced); the world still simulates it, as the twin | `plantwire` app (`devices`) |
| The origin of every device born to the host: `origin` records `{instance?, device, board, environment, boot_id}` in the observer journal, and a `binding` record `{outcome, refusal?, reason?, issues}` per gateway handshake of a plan that binds anything | `plantwire` domain builds them, `run`/`journal` write them (contract `run-journal-v0.1`: two new observer session records, outside the stream digest) |
| Starting, stopping (interrupt, bounded wait, kill) and reporting the session process | `canvasbridge` app with the new `process` edge file `lasting.go` |

**Command.** `streamsim plantwire connect (--plan <file> | --stdin) [--check] --out <dir> [--domains-dir] [--adapters-dir]
[--kinds-dir]... [--slice-s 1] [--barrier-cap-ms]`. With `--check` it reads the domain and the kind files and prints
`{domain, environment, instances: [{instance, origin: real|simulated, kind, kind_digest?, board?, device?},
implements_checked]}` (`environment` is `mixed` when the plan has a gateway) or the `plantwire_failed` failure with its issues,
and starts nothing. Without it, it runs until interrupted: with no gateway the
world is a plant wire node (`plantwire node` with no end time); with a gateway it is the R8 mixed relay given the plan's
bindings. An interrupt is the operator's Stop: the run ends and `--out` holds an ordinary run folder (`run.json`, trace,
ledger, history, both journals, and `field-trace.jsonl` when a gateway was attached) that `replay` and `verify` accept; the
document is `{out, stopped, relay?}` (the folder is the run's name; a mixed run's world is built inside the relay).

**Bridge actions** (registry, all mapped to `plantwire connect`):

| Action | Runs | Arguments | Result | Errors |
|---|---|---|---|---|
| `session.check` | `plantwire connect --check --stdin` (+ workspace `domains/`, adapters, `kinds/`) | `content`: the plan | the check document | `plantwire_failed` with issues (422); the command's own failures |
| `session.start` | `plantwire connect --stdin --out runs/<id>`, a lasting child the bridge owns; `<id>` is `c-` and 8 hex digits from the bridge's entropy | `content` | the status once the child has opened its run journal (so the canvas can follow it at once) or ended, waiting at most 10 s: `{run, state: "running"}`, or `failed` when the command refused the plan | `session_running` when one lives (one at a time); a run folder that exists (`conflict`) |
| `session.status` | nothing | none | `{state: "none"}` or `{run, state: running\|ended\|failed, result?, failure?}`: a non-zero exit is `failed` with the command's error document, or its last stderr words | – |
| `session.stop` | an interrupt, then at most 20 s wait, then a kill | none | the final status | – |

While a session lives its run folder is an unfinished run (the truth gate keeps sealed files back; the bridge counts it as
running). The bridge stops a living session when it closes.

**Canvas.** No new module: the session is a kind of launch. `model/launch` holds the session draft, the plan document it
writes, the placement of a binary issue on a row by its pointer, and the words; `app/launch` holds the connect store
(check, start, follow the run's journal for `binding` records, poll the status, stop); `ui/launch` the Run panel's choice
"To completion" or "Connected session" (RUN-2), the session form and the live session. The origin of an instance is the
journal's `origin` record, decoded by the adapter: `model/graph` holds the origin value and its words, `model/review`
gathers them per instance, `app/review` keeps them in the review state, `ui/graph` marks nodes and outline rows, `ui/review`
the inspector header and the timeline chips. Real and simulated differ by shape (a chip icon) and colour (new tokens
`--origin-real`, `--origin-simulated`, light and dark), and say it in words to assistive technology.

**End-to-end.** The test drives the real binary: `plantwire host --environment mixed` as the host, the canvas's connect
session, and the stand-in gateway (`internal/plantwire/internal/standin`, test support) run from a test binary of the
plantwire app package (`go test -c -o bin/standin-gateway`, built by `make build` beside `bin/streams-simulator`; the e2e
global setup only checks that it is there, `STANDIN_BIN` overrides the path; its `TestMain` plays the gateway when given
`--standin-gateway`), so no production package imports the stand-in.

**Deviations from the plan.**

- The world keeps simulating a withheld instance (the plant does not change, and links into it still work); its values
  are the twin's, so the review shows no simulated reading for an instance whose origin is real (R10 brings the board's
  readings, R12 the twin overlay).
- No live origin in the plant screen: origin marks are drawn from the journal of a session run in the review, which follows
  a living run's journal as it grows (live mode is R10).
- A stopped node session sends no `NDEATH`: the interrupt closes its link and the host sees the socket close.
- The `--origin-real` tokens existed already (the capture screen of R6b); R9 adds `--origin-simulated` beside them.
- The canvas asks the bridge for a session that is already running (after a reload) when the Run panel is switched to
  "Connected session", not when the panel opens, so a run to completion makes no session call. The top bar's run
  indicator does not follow a session yet (R10).
- A connected session advances the world in one-second slices by default (`--slice-s 1`), as fast as the host answers each
  barrier (D-71 still applies).

**As built (R9 review fixes, M3).** What the review of the connect session found, and what changed:

- *A side answers only its own commands.* The mixed relay records which side each routed `DCMD` went to
  (`routeLedger`, 4096 commands). A `DRECEIPT` or `DRESULT` from the gateway for a command routed to the world, or from
  the world for one routed to the gateway, or for a command never routed, is dropped and counted in `dropped` and in the new
  `forged`. A gateway `DBIRTH` on a board other than the one whose members hold the device, or of a kind the `NBIRTH` never
  announced, is dropped the same way; a bound device born as another announced kind on the right board still passes with the
  `device_kind_mismatch` note.
- *The field trace holds only what was admitted.* A gateway record is marked for the trace (`Traced.Field`) only when it
  decoded and belongs to the admitted session (an accepted `NBIRTH`, or a valid record after it): refused births, invalid
  records, records of another session and duplicates are journalled on the wire but not in `field-trace.jsonl`, so
  `replay-trace` no longer replays them. A full trace or a line over its limit no longer ends the relay: it stops recording,
  disconnects the gateway (its devices die `link_down`, the host session stays), turns the next gateway away, sets
  `trace_incomplete` and leaves `field_trace_digest` out of `run.json`, so `replay-trace` refuses the folder. The record line
  limit of the trace is now 8 MiB, so that every record the link accepts (1 MiB, with numbers that canonicalise longer, such
  as `1E20`) fits a line.
- *Session plan sockets* are checked by the schema (an absolute path, no control character, no `..` component) and by the
  plan check (`socket_not_absolute`, `socket_control_character`, `socket_parent_component`, and `same_socket` on the cleaned
  paths); the session uses the cleaned paths.
- *A gateway that is gone.* A `DCMD` for a device of a gateway that is not admitted is answered `DRECEIPT refused
  not_ready` (the contract has no `link_down` reason; `not_ready` is what a device that cannot act answers). A lost or
  departed gateway clears its devices, its pending barriers and its sequence, so an `NDEATH` followed by an `NBIRTH` at seq 0
  on the same link is admitted. A host's rebirth `NCMD` in a mixed relay reaches the gateway and asks the world to be reborn
  too, whether or not the gateway is there.
- *A dead simulated node.* A full mailbox is `ErrSimOverflow`, a failure of the relay (it was swallowed as a closed link). When
  the simulated node ends before the relay, its devices are reported `DDEATH offline` to the host at once, `simulated_lost`
  counts it, its commands are refused `not_ready`, and the failure surfaces in the result at stop.
- *Stop is not an error, other errors are.* A node session discards only the errors the interrupt caused (a cancellation or a
  closed connection); a real failure while stopping is reported.
- Not changed: `DropMalformedEvidenceDigests` keeps a malformed `field_trace_digest` on purpose (a test pins it: dropping it
  would let a trace the artifact no longer covers pass the file check). `Gateways.Close` removes the socket path only when it
  is still the socket the relay made. The kind digests of the plan are keyed by the kind file name in one function.
  Deferred: D-73 to D-76 (gateway liveness, rebirth rate, the digest of a hardware-only field trace, origin records in replay).

**As built (R9 review fixes, M3): the bridge and the canvas.**

- *The bridge never blocks on a session.* `session.start` runs the binary's plan check first (a refused plan starts
  nothing), claims a fresh run folder under its state lock (a taken id is retried, five times), then launches the child
  outside any lock. `session.status` answers at once with `none`, `starting`, `running`, `stopping`, `ended` or `failed`,
  always with the session's `domain` and the binary's check document (`plan`), so a reloaded canvas knows the plant and which
  instance is real. A stop asked during the start is carried out as soon as the child runs; a second stop during a stop
  answers `stopping`.
- *A normal stop closes the run.* The child gets `--barrier-cap-ms 10000` and the stop waits up to 30 s after the interrupt,
  longer than the cap, before it kills the child's process group (the child runs in its own group; closing the bridge stops
  the session the same way). The child's standard output is kept up to 1 MiB, and a result document over 64 KiB is left out
  of the status.
- *A run the bridge started says so.* The workspace listing marks the runs of sessions this bridge started `connected`, and
  `running` while the child lives; the canvas labels them "Connected session" (and "running"), never Explore or Evaluate.
  This is bridge memory: after a bridge restart an old session run is labelled by its journal (an `origin` or `binding`
  record makes the review a connected session), and the workspace list shows it as Evaluate until the binary records the
  run mode (a binary change, not made here).
- *The canvas keeps a session it could not stop.* A failed stop leaves the session live with a note and Stop can be tried
  again; the follower and the poll end only when the session ended or failed. A journal stream that closes or fails is
  noted and followed again on the next poll. The panel resumes a running or starting session with the domain and plan from
  the status, and shows "A connected session of <plant> is running" on another plant.
- *Plain words, every issue.* Each plan issue is worded by its rule (`keyword`) or member through one table, with the
  binary's own text under "For developers"; every issue a field does not hold (including issues on rows that no longer
  exist) is listed. An edit after a check clears the field errors and marks the check out of date. The seed is sent as typed:
  a whole number as a number, anything else as text the binary refuses at `/seed`. `schema_version` stays a canvas constant:
  neither the bridge nor the binary hands the canvas the session plan's version (a `plantwire connect --check` that accepts a
  plan without it, or a version in the catalog, would remove it).
- *A refused gateway on the plan list.* Each real instance reads "Not bound: refused · not announced to the host" with a
  warning icon, and the refusal says the next step (stop, fix the board or device, start again). The world withholds a
  refused instance from the host, so "running simulated" would be wrong.
- *Origin marks.* One label map (`model/shared/origin.ts`, real, simulated, origin unknown) and one badge
  (`ui/primitives` `OriginBadge`) serve the nodes, the outline, the inspector, the timeline chips and the plan list. A node
  says its origin once (its group label); at the compact, overview and map tiers the badge is an icon of fixed screen size
  without its word. In a connected run every instance without an `origin` record gets the "Origin unknown" mark, and a
  missing or unknown origin in the plan check is unreadable, never simulated. The review header shows the gateway binding,
  and the timeline shows it as a mark at its time. Effector calls reach the timeline from the report's `effector_calls`
  (`eventMarks`), the same path as a run to completion.
- *Focus.* Nothing takes focus when the session form opens; the run mode stays mounted (disabled while a session holds the
  plant or a run is in progress) and keeps its arrow keys; focus moves to the live status after a start, to the saved run
  after a stop, to the first flagged field (or the refusal) after a refused check, to the new row's instance after "Add a
  real board", and to the error of a failed session. The mode fieldset no longer shrinks, so its heading is not clipped.
- Deferred to R10 (live mode, the top-bar run indicator, connection health): a reload affordance in the plant screen and
  live origin marks on the plant screen, because both need the live run view R10 builds; after a reload the session panel
  already shows the plan summary from the status.

### 4.2 Interrupts: a local/remote switch at the node

The simulator has no knowledge of its consumers, so the canvas cannot reach Agentic Stream's operator interface through the
bridge. Real plants already solve this with a **local/remote switch**: an operator at the panel takes equipment into local
control, and remote commands are refused. The relay is the node, and a node may already refuse commands under the contract
(`interlocked`, `not_ready`).

| Control | Scope | Effect at the node | What the host sees |
|---|---|---|---|
| **Hold** (local) | A device, a board or all | Every `DCMD` is refused | `DRECEIPT refused`, reason `local_control` |
| **Approve each command** | A device, a board or all | Each `DCMD` waits until the operator approves it or its deadline passes | `accepted` after approval; `expired` otherwise |
| **Emergency stop** | A device, a board or all | The node sends `safe_stop` down to the device | `DSTATE` with the safe state |

Rules:

- **The node may only refuse or stop; it never originates an ordinary command and never widens what the host may do.**
  Agentic Stream stays the authority that grants; the node adds a local veto, as a panel switch does.
- Every hold, approval and emergency stop is a record in the session's journal and command log, so replay reproduces it.
- The same controls act on simulated devices, so how Tamoz handles refusals and safe stops can be tested before hardware.
- Agentic Stream needs no operator interface for this; it handles `local_control` refusals as it handles any refusal.

## 5. Safety

- **Read-only first.** A hardware board announces its kind's commands, but Agentic Stream's commissioned inventory denies them
  until M5.
- **Transport.** v1 is Unix socket only, owner-only (SPEC §2): gateway, simulator and Agentic Stream run on the bench machine.
- **Device-side safety.** Expiry, lease, `boot_id`, idempotency and safe stop remain device-wire-v1 duties in the firmware;
  neither the gateway nor the simulator stands in for them. The policy digest is not one of them: the gateway checks it for a
  real board (SPEC §13, `policy_not_allowed`, from its own list) and the simulator for a simulated device (the session plan's
  `policies`); device-wire-v1 has no policy reject code and the firmware does not refuse on it.
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
| **M1 — Define the Arduino board and its twin** | Kind file contract and `kind_digest`; `implements` check in the `domain` module; workspace kind files and components with validate actions; command response templates; the thermal kind files and a twin implementing them, as data; canvas Devices section and device editor (§1.5, §4.1) | The Arduino board is defined in the canvas without editing a file; its twin runs over the plant wire to Agentic Stream and Tamoz, deterministic on replay; a twin whose kind drifts from the kind file is rejected |
| **M2 — Gateway on the plant wire** | The Arduino gateway gains a plant wire side announcing the same kind files; Agentic Stream commissions the board read-only; capture of a kind from a connected board | Same `kind_digest` from gateway and simulator; real data ingested with zero quarantines; unplugging the board shows `link_down` and a rebirth; capturing the connected board yields kind files equal to the defined ones |
| **M3 — Relay and mixed sessions** | Relay mode (§2), per-board origin, field trace as run input (§3), simulated boards beside the real one; canvas Connect session, real or simulated per session, origin marks | The real board plus N simulated boards reach Agentic Stream as one node; a recorded session replays byte-identically; a test proves no replayed or simulated command reaches the gateway |
| **M4 — Live control plane** | Canvas live mode, command lifecycle and connection health (§4.1) | One reading followed from the board to Tamoz's decision on the live screen; a refused, an unconfirmed and a confirmed command are told apart at a glance; the same session reopened from its recording |
| **M5 — Interrupts and real commands** | Hold, approve each command and emergency stop at the node (§4.2) and in the canvas; `local_control` refusal; fan and LED commands commissioned; an independent instrument on the bench; twin overlay and twin-versus-board report | A Tamoz decision drives the real fan, confirmed by the instrument rather than `DRESULT`; hold, approval and emergency stop each shown on the bench and reproduced on replay; the report names the divergences |
| **Second proof** | The co-founder's board through M1–M5 | No change to Tamoz or Agentic Stream workflows |

## 8. Changes by repository

| Repository | Change |
|---|---|
| streamsim binary | Kind files, `implements`, workspace components, response templates (M1); kind capture (M2); relay, field trace, mixed node (M3); hold, approval, emergency stop at the node, twin report (M5) |
| streamsim canvas | Devices and device editor (M1); save as device (M2); Connect session, per-session binding, origin marks (M3); live mode, command lifecycle, connection health (M4); interrupts, twin overlay (M5) |
| Plant wire contract | `kind_digest` in `NBIRTH` (M1); per-board environment (M3); `local_control` refusal (M5) |
| Agentic Stream | Plant wire host for the thermal kinds (M1–M2); handles `local_control` refusals like any refusal (M5) |
| Tamoz | Workflows on the thermal kinds, unchanged across simulated, real and mixed |
| Gateway and firmware | Plant wire side of the Arduino gateway (M2); an independent effect instrument (M5) |

## 9. Deferred

- **Time in mixed sessions.** Agentic Stream can hold a simulated world at each barrier; a real board cannot be held. How the
  world is paced when a real board is present is decided when M3 starts.
- Authentication, several users, customer sites and production reliability of the control plane (D6).
- Remote gateways: TCP needs mutual authentication in a new contract version.
- Incident reproduction with faults injected on a recording; more boards; fleet-scale mixed runs.
- The gateway's field mapping in the canvas.
- **Open**: the order of this work against the open canvas production-UI goal (P9–P11, A1–A7).

## 10. Not doing

- Instruction-level emulation of a microcontroller or its firmware.
- A serial or BLE stack, broker or device-specific code in the binary.
- The simulator as a policy authority, or as a sender of ordinary commands to hardware; it may only refuse or stop (§4.2).
- Knowledge of Agentic Stream's operator interface in the binary or the bridge.
- Scoring a consumer against a device's self-reports.
- Customer devices, mappings or recordings in any repository.

## R10 design decision: recorded live evidence

The live screen and a reopened session consume the same observer journal. Plant wire owns the interpretation of
admitted records into connection health, reading quality and command stages. The bridge transports that projection;
the canvas owns presentation and whether its cursor follows the newest record. Connected-session metadata and the
checked plan are recorded with the session so a bridge restart cannot change its classification or origins.

Receipt, board report and observed effect remain separate. The reference host already checks a declared feedback
metric against an expectation using a sample received after a successful board report. Its decision previously stayed
in its final report. R10 adds optional `command_evidence` to `HSYNC` to carry that decision through the existing host
channel. Each decision identifies the command, device, boot and metric, with the expectation, tolerance and observed
value when available. This is an additive plant-wire contract change; existing hosts may omit it, which leaves the
effect unknown. The relay forwards host content unchanged, except that it consumes `command_evidence` from `HSYNC`, and records only evidence matching the routed command and
boot. A receipt or successful board report alone never confirms an effect. A host observation is recorded evidence,
not independent proof of a physical effect on the bench.

The existing plantwire domain layer owns these rules; its app and socket edge supply records and deterministic ticks.
The existing run/journal modules persist observer records without interpreting plant-wire policies. No new package or
dependency is required. Gateway silence uses a bounded watchdog even before a host heartbeat has been seen. Reading
quality distinguishes implausible, stale and disconnected values from current readings using admitted capability data.

## R10 as-built (live mode, command lifecycle, connection health)

Built as the decision above, with these details and deviations:

- **Projection.** `plantwire/internal/domain` `Live` folds what the host saw (the records the node or relay sent up, and the
  host records the relay admitted) into link states (`host`, `gateway`: `unknown`, `online`, `offline`, or
  `not_applicable` without a gateway), devices (board, environment from the `NBIRTH` boards, `unknown` for a board the birth
  did not announce, boot id, reboot count by `boot_id` change, health) and readings with a state: `good`, `stale` (three scan
  intervals of relay ticks without a sample, the gateway idle limit when the kind declares no scan class), `disconnected`
  (device death, gateway link lost for hardware devices, host link lost for all), `invalid` (wrong declared type, enum or
  outside `eu_range`) and `untrusted` (any other quality). Commands keep four separate stages: sent, receipt (`pending`,
  `accepted`, `refused` with the reason; a `queued` receipt stays pending), the board's report, and the effect, which is
  `observed` only for host evidence that names the same command, device and boot, follows an accepted command whose board
  reported success on the device's current boot, and whose observed value lies within the stated tolerance of the stated
  expectation. A receipt or result from another boot (a `refused` receipt with boot id `none`, which is how the relay and the session refuse before any boot is involved, is accepted), a regressing answer, duplicate or premature evidence changes nothing. `sent` means the host issued the command; a relay that did not route it shows that as the receipt stage. Evidence counts only from an `HSYNC` the relay routed (a dropped one counts nothing), only for a metric the device announced, with a finite tolerance no wider than the metric's `eu_range`, and an `observed` decision must agree with the latest good reading the relay holds, otherwise it is recorded as unverified; evidence for a device that is offline or rebooted leaves a mark instead of vanishing. Simulated instances have health `simulated` and never age with wall time.
  The latest 128 commands are kept and truncation is explicit. Link changes are idempotent (one mark per change).
- **Journal cadence (deviation from the WIP).** The WIP journalled a full snapshot after every admitted record, which for a
  plant of n devices writes O(n) snapshots of O(n) devices per slice. The session now marks the projection changed and writes
  one snapshot per barrier (simulated session, after the barrier) or per 250 ms relay tick (mixed session), one at every link
  change, and a final one when the session ends; `marks` carries every change since the previous snapshot, so no mark is lost.
  `live` and `session` records are observer session records outside the stream digest; replay does not regenerate them
  (D-76, D-77).
- **Session record.** `plantwire connect` writes one `session` record (`mode: connected` and the `--check` document, empty
  lists as arrays) right after `run.opened`, for the simulated session and the mixed relay. The bridge lists a run as a
  connected session from it when its memory has no entry (bridge restart), and the canvas reads the plan's simulated
  instances from it. Plain `plantwire relay --domain` and `plantwire node` write neither `session` nor `live`.
- **Contract 1.3.0.** Optional `HSYNC.command_evidence` (at most 128 entries, `observed` needs `observed_value`, tolerance
  non-negative); schemas, the valid fixture and two invalid fixtures generated from `scripts/plant-wire`, the `DIGEST`
  regenerated and the embedded copy equal. The reference host sends each decision once, only for a good, non-historical
  sample of the boot the command was issued to, newer than the device's successful result; a timeout is `not_observed`.
  The relay consumes `command_evidence`: it forwards `HSYNC` to the gateway or the simulated node without it (host-internal data, not for firmware), and only `through_seq` is mapped, as before.
- **Gateway watchdog (D-73 fixed).** The relay counts ticks since the last admitted gateway record; stray, forged, malformed
  and repeated records do not reset it. After 30 s it reports the devices `DDEATH link_down`, closes the gateway link and
  records the gateway offline. It arms at admission, without a host heartbeat.
- **Review fixes (M4).** A link change no longer writes a snapshot of its own: it waits for the next tick (a connect loop
  would otherwise write one per connection), a repeated link state changes nothing, and the session ends with a final
  snapshot with the links offline and readings disconnected. Elapsed time for staleness and the gateway and host watchdogs is
  the measured time between delivered ticks, not a count of ticks (`Relay.TickBy`). Marks carry `subject` and event words.

- **Canvas.** The adapter decodes `live` and `session` records strictly (an unknown stage, state or link value drops the
  snapshot rather than guessing); the review store keeps the snapshots of a visit as an immutable chain, follows the newest
  while the session runs, pauses following on any seek, and `Back to live` resumes; after `run.closed` the session is an
  ordinary review and verification loads. The Run panel shows connection health and a commands table (Acceptance, Board
  report, Effect evidence) while live and above the report once saved; the timeline adds the binary's command and connection
  marks. Real-board readings come from the projection: a reading that is not `good` shows its state, never a value. The plant
  screen of the session's plant shows origin marks and live readings while the session runs; the top bar shows
  "Connected session · running · Open live view" after a reload. While a session holds the plant, the Run panel hides the
  run-mode chooser (the blank row above the live status).
- **Fix to the WIP: the plant screen stays mounted.** The WIP rendered `PlantScreen` directly when no session was live and
  inside `ConnectedPlant` once one was, so starting a session remounted the plant screen and closed the Run panel (the
  `connect-session` e2e lost "Connected session · run c-…"). `PlantRoute` now always renders `ConnectedPlant`, which opens
  the review only for a live session of this plant and passes no session otherwise; `Stores` moved to `shell/internal/stores.ts`
  to break the `App` ↔ `PlantRoute` import cycle, and `LiveHistory` to `reviewState.ts` to break the `liveState` ↔
  `reviewState` cycle.
- **Not done in R10.** The commands table distinguishes stages in words, not by colour; timeline marks from the session all
  use the neutral tone (the binary's marks carry `kind` but no severity). After a reload the Run panel's live status has no
  plan check (the connect state resumes from the bridge without it); the review and the workspace listing read it from the
  `session` record. The e2e covers live following, Back to live, connection health and the reopened recording; a host that
  issues commands (refused, unconfirmed and confirmed rows on the live screen) is covered by unit and component tests, not
  by an e2e, because the reference host in the bench issues no commands.

### M4 review fixes (canvas and bridge)

- **Live history.** The review store keeps a flat, time-ordered array of snapshots (binary search by `at_ns`), drops a
  snapshot whose state equals the previous one (the adapter gives each snapshot a key of its links, devices and commands),
  and keeps at most 2000 snapshots and 2000 marks; when it trims, the live bar says "History trimmed before hh:mm:ss".
  Origins under the cursor are the journalled `origin` records overlaid with the snapshot's devices (one mapping,
  `originOf` in the model), memoised per snapshot.
- **Unreadable snapshots.** A `live` record that does not decode, or whose command stages break the invariant (an
  observed effect needs an accepted receipt, a successful board report and evidence), becomes a `live-unreadable` entry;
  the live bar says "A live snapshot could not be read at hh:mm:ss; the view is not current" until the next good snapshot.
- **Catching up.** The bridge's run event stream sends `caught_up` once it has read the journal to its end; until then
  the live bar says "Catching up to live…" and the graph bar "catching up", not "following". The bar shows the age of the
  last snapshot ("Last live snapshot 12 s ago"; from 30 s, a display-only rule, "No live snapshot for N s"). End on the timeline of a
  running session is "Back to live"; focus stays on the live status after "Back to live". The end-of-load seek no longer
  clears following for a running session.
- **One live store.** The shell keeps a second review store open for the running session (`liveReview`), shared by the
  plant screen and the live view and reference-counted, so moving between them keeps the history and never shows an
  older state.
- **Words.** Health, link, receipt, report, refusal and reading words come from one model table; a simulated device
  reads "Simulated · no link"; a reading that is not current shows "—" and its state and quality ("untrusted · warming
  up"); command rows show the side they were routed to, the run-clock time sent and coloured stage chips; the plant
  screen's primary action while live is "Open live view", and Recent runs marks the running session "Live".
- **Bridge.** `RecordedSession` accepts the `session` record only as the first observer session record of the journal.
- **Re-review fixes.** The bench recording showed `marks: null` in 523 of 529 `live` records (the binary wrote a nil
  slice); the decoder reads a null `marks` as no marks, so those snapshots are no longer unreadable (the binary should
  still write `[]`). The live view of a session loads its plant from the session record's domain before any report
  exists, asks for a failed report again once caught up, follows the newer of the newest snapshot and the newest
  observation, and follows a running session again (after 1 s) when its event stream fails, for example while the run
  folder does not exist yet. Pills of session marks show the binary's event words and stay inside the timeline; a node
  card shows the short reading state (the reason is in the inspectors); simulated rows show their board but no boot.
- **Deferred.** De-collision of overlapping timeline pills, keeping fit-zoom across a scrub, the board-readings heading
  alignment and keeping the plan check in view are left to the next design review.

## R11 as-built (binary)

Hold, approve each command and emergency stop (section 4.2), built in the binary; the canvas side is R11b. Decisions and
deviations:

- **One rule set, two nodes.** The rules are one pure state machine in the `plantwire` domain layer (`interrupts*.go`:
  `Interrupts`, with `InterruptGate` wrapping it for the callers). It takes *entries* (records) and returns *entries*; it
  never reads a clock, a socket or a map in output order. A simulated node (`Session`, a world speaking the plant wire) and the
  relay (`Relay`, hardware only or mixed) each hold one gate and supply only records, the node's clock and the edge. A mixed
  relay holds the one gate for both origins (the world's inner `Session` takes no controls), so a scope of `all` or a board
  means the same for a gateway device and a simulated one.
- **Scopes and precedence.** A scope is `{kind: all | board | device, id}` (no id for `all`; a board or device the node does
  not have is refused `unknown_board` or `unknown_device`). Several scopes may match one device: **hold outranks approval
  outranks nothing**. At most 64 scopes (holds and approval scopes together) are set (`too_many_scopes`; repeating a scope
  already set is not a new one), and at most 64 commands wait, each with at most 8 repeats. Setting and clearing are explicit
  and idempotent: `hold`, `release` (clears the hold, including a latch, on exactly that scope), `approve_mode` with `enabled`
  true or false.
- **Hold.** A `DCMD` in scope is refused `DRECEIPT refused, local_control` (boot id `none`, as every refusal before a boot is
  involved) and never reaches the world or the gateway. A hold set while commands wait refuses them at once.
- **Approve each command.** A `DCMD` in scope is answered `queued` (boot id of the device when known) and waits. `approve`
  hands it on **unchanged**, exactly as if it had just arrived, so the normal checks (freshness, boot, epoch, policy, effector,
  interlock) still accept or refuse it, and a hold set since then refuses it `local_control`; `deny` refuses it
  `local_control`; a node clock past its `deadline_ns` refuses it `expired` (never later than the first barrier or tick after
  the deadline); a command already past its deadline when it arrives is refused `expired` at once; a repeat of a waiting command
  (same key and content) waits too and is answered with it, other content under the key is `duplicate`; a full queue refuses
  `not_ready`. The clock is the world clock for a simulated device and the relay's wall clock (the tick time) for a hardware
  device; an entry records which.
- **Emergency stop.** `estop` for a scope performs the safe stop of every device in scope that declares one (a node of the
  world invokes the domain's `safe_stop` effector through the run, so it is in the command log and replays with the world; a
  relay sends the declared safe-stop `DCMD` to the gateway or the world, see D-82), reports `stopped` and `skipped` devices,
  **and latches a hold on that scope** until the operator releases it: a real emergency stop does not release itself, and a host
  that would resume the plant is answered `local_control`. This choice is deliberate and differs from a bare "send safe_stop".
  A node of the world also sends the stopped device's `DSTATE`.
- **Safety exemption.** A command the device's kind declares `safe_stop` is never held, queued or refused by a hold or an
  approval mode.
- **Never widens.** The node originates exactly one kind of command, the declared safe stop of an emergency stop, and
  refusals and `queued` receipts. It never forwards what the host did not send, never alters a forwarded command, and consumes
  the answers to its own stops so the host sees no receipt for a command it never issued. Without a control socket the relay is exactly
  what "As built (R7b)" describes: it originates, answers and refuses no command.
- **Control socket.** `plantwire connect --control-socket <path>` creates an owner-only socket (the kernel helper: owner-only
  directory, no live socket or file at the path, removed at the end) at most 8 connections wide. One request per line
  (`node-control-v0.1`, validated by the binary with specific reasons; the bridge validates nothing):

  ```json
  {"op": "hold", "scope": {"kind": "board", "id": "mega-01"}}
  {"op": "approve_mode", "scope": {"kind": "all"}, "enabled": true}
  {"op": "approve", "command_id": "cmd-7"}
  {"op": "estop", "scope": {"kind": "device", "id": "fan-01"}}
  {"op": "state"}
  ```

  One answer per request (`$defs/reply`): `{"result": "applied" | "refused", "reason"?, "op", "command_id"?, "stopped"?,
  "skipped"?, "state": {...}}`. Reasons: `malformed`, `unknown_board`, `unknown_device`, `unknown_command` (approve or deny of an
  id that does not wait), `too_many_scopes`, `node_busy`. `state` is answered at once from the last published view; every other
  request is applied by the node and answered after (D-81). The state in every answer and in the live snapshot:

  ```json
  {"holds": [{"kind": "board", "id": "mega-01", "latched": true}], "approve": [{"kind": "all", "id": ""}],
   "pending": [{"command_id": "cmd-7", "device": "fan-01", "command": "set_pwm_lease", "deadline_ns": 1767225780000000000,
                "queued_at_ns": 1767225660000000000, "clock": "world"}],
   "devices": [{"device": "fan-01", "board": "mega-01", "mode": "hold", "latched": true}]}
  ```

  `holds`, `approve` and `devices` are sorted (all, board, device; then id), `pending` is in arrival order, `devices` lists only
  covered devices and `mode` is `hold` or `approve`.
- **Storage and replay.** The decisions depend only on the entries and the node's clock reading in them, so they are a pure
  function of a recorded input. The record is `interrupts.jsonl` in the run folder, owned by the `run` module like the field
  trace (`run.NewInterruptLog`, `run.ReadInterruptLog`, `Run.SetInterruptLog`): canonical `{ordinal, record}` lines
  (`interrupt-log-v0.1`), covered by the optional `interrupt_log_digest` of `run.json`. A run without a control socket has
  neither the member nor the file, so no existing golden, pin or digest changes. Records: **inputs** `inventory` (the devices,
  their boards, clocks and safe stops, written when it changes), `control`, `command` (every `DCMD` the node judged: id, key,
  content digest, device, command, deadline, whether it is a safe stop), `clock` (a clock reading that settled waiting commands;
  written only when it did) and `flush` (the host session ended with commands waiting; refused `session_ending`);
  **decisions** `reply`, `verdict` (`pass`, `refuse` with a reason, or `queue`), `resolve` (a waiting command forwarded or
  refused, with the repeats that waited with it) and `stop`; **evidence** `effect` (what a node's own stop produced). Every
  entry carries `clock` (`world` or `wall`) and `node_ns` on that clock. `domain.ReplayInterrupts` feeds the inputs to a fresh
  machine and requires the recorded decisions, in order, to equal what it decides (`ErrInterruptReplay`, naming the input's
  record); `app.VerifyInterruptLog` reads the folder (digest against `run.json`), validates every record against its schema and
  replays, and a session calls it when it ends: a log that does not replay is left out of `run.json`. `replay`, `verify` and
  `report` check the file against the digest (`file_differs`, `file_missing`, a `interrupt_log` check). Held and denied
  commands never reach the world; approved commands and emergency stops go through `InvokeEffector`, so they are in the command
  log and the world replays byte for byte. The rules replay is not a subcommand (D-80).
- **Journal.** Each entry is also an observer `interrupt` session record (outside the stream digest, like `live`), so a
  reopened recording shows every control and decision in order. The `live` snapshot gains the optional member `interrupts`
  (the state above), present from the first barrier or tick of a node that has a control socket and rewritten only when the state changes; a command refused
  `local_control` is an ordinary `refused` command with that `reason`, and a command that waits stays at receipt `pending`
  until its terminal receipt. The canvas adapter decodes strictly: R11b adds the member and the `interrupt` record type before
  a session with a control socket is opened in the canvas, and registers `interrupts.jsonl` among the bridge's observer files.
- **Deviations from section 4.2.** The table's "`accepted` after approval" is the normal terminal receipt of the forwarded
  command, which can still be a refusal. Controls are served at barriers and relay ticks, not at once (D-81). Controls exist on
  `plantwire connect` only (D-83). The relay's stop is described in D-82.

## R11b as-built (bridge and canvas)

The controls of section 4.2 on the live screen. The binary keeps every rule; the bridge forwards and the canvas shows what
the binary says, with no rule of its own and no parsing of message text.

- **Bridge.** `session.start` makes an owner-only directory for the session (`os.MkdirTemp`, in the `process` edge file
  `control.go`) and passes `--control-socket <dir>/node.sock` to `plantwire connect` before `--out`. The new action
  `session.control` takes the `node-control-v0.1` request as its arguments, encodes them as one line (at most 4096 bytes), dials the
  socket with `kernel.DialPrivate` (owner-only directory and socket checked), waits at most 12 s for one line and returns it
  unchanged. A refusal of the node is a reply (`result: "refused"` with its `reason`), not a bridge error. The bridge's own
  failures are the `422` codes `no_session`, `control_unavailable` (the session has no control socket), `control_gone`,
  `control_timeout` (the node may still apply the control afterwards) and `control_unreadable`. `session.status` carries
  `controls` (true while the live session's socket exists); the directory is removed when the session is seen to have ended
  or failed, when the next session claims the slot, and when the bridge closes. `interrupts.jsonl` is registered among the
  observer-readable run files (`paths.go`, `truth.go`). A bridge that cannot make the directory starts the session without
  controls, and `controls` stays false.
- **Decoder.** The journal's `interrupt` record becomes a `JournalEntry` of kind `interrupt` (event, node time, clock, op,
  scope, flag, command id, device, verdict, resolution, result, reason). An unknown event, op, clock, result, verdict,
  resolution or scope kind, a scope without an id (or `all` with one), a flag that is not a boolean or a node time that is not a
  number drops the record to the same `live-unreadable` entry an unreadable snapshot gives, so the live bar says "A live
  snapshot could not be read at hh:mm:ss; the view is not current." (the words say snapshot; the dropped record may be an
  interrupt record, which only costs a timeline mark). The optional `live.interrupts` is decoded strictly (holds with their
  latch, approval scopes, waiting commands with their clock, covered devices with their mode); absent or `null` means a node
  without controls, a malformed state drops the whole snapshot. The snapshot key includes it, so a change of the state is a new
  snapshot in the history. The reply of `session.control` is decoded the same way; a `reason` the canvas has no words for is
  kept and shown as its own words with underscores removed.
- **Words.** One table in the model, `reasonWords` (`sessionWords.ts`), words every refusal reason: the command table's
  (`local_control` reads "Refused at the node: local control", distinct from a refusal of the device or the host, which
  keeps "Refused" and the reason underneath) and the control replies' (`unknown_board`, `unknown_device`, `unknown_command`,
  `too_many_scopes`, `node_busy`, `malformed`, `expired`, `not_ready`, `duplicate`, `session_ending`). Modes, requests, outcomes,
  the deadline of a waiting command and the stop report are in `interruptWords.ts`; selectors (the state to show, what covers
  an instance at the cursor, scope choices, active scopes) in `interruptSelect.ts`; the timeline marks in `interruptMarks.ts`.
- **State shown.** The bar shows the node's newest state: the state of the last control reply while no newer snapshot arrived
  (the store keeps the key of the newest snapshot it was read against), the newest snapshot otherwise. It is not the state at
  the cursor: a control acts on the node now. Scrubbing back leaves the bar as it is; the node cards and the Inspect line show the
  state at the cursor.
- **Controls store.** `ReviewSession.controls` (`app/review`, `ControlStore`) sends one control at a time and has no optimistic
  state: the bar shows "sent, waiting for the node's answer…", then the binary's reply or a plain failure. After a failure
  that may have reached the node (`timeout`, `gone`, `unavailable`, `unreadable`) and after a `node_busy` refusal it reads the
  state again with `{op: "state"}` and says whether that worked; after `no_session` and `no_controls` it does not.
- **Screen.** An "Interrupts" region, always visible while a connected session runs: on the live view between the live bar and
  the timeline, and on the plant screen under the plant bar. A scope select (all devices, each board, each device of the
  newest snapshot), Hold (pressed while the scope is held), Release, Approve each command (pressed while on; the same
  button turns it off), and Emergency stop (filled alarm colour, one press, described as sending the safe stop and holding the
  scope, latched, until released). Chips list every hold, latch and approval scope in words and by icon; a list names the
  commands that wait, each with Approve and Deny and the node's own deadline (world clock: time left on the run clock from the
  newest snapshot; wall clock: ticking from the browser clock); after a decision focus moves to the waiting heading; the
  outcome of the last control is a status line. Controls are disabled until the node publishes its first state, and the
  panel says why (waiting, or no control socket). Nodes show a badge by icon and word (hand and "Held", octagon and "Latched",
  clipboard and "Approving"), the node's accessible name and the Inspect lines say it in full; the commands table and the
  timeline show the binary's decisions (marks of kind "Interrupt control": the operator's control, a refused control, an
  approval handed on or refused, a safe stop sent). The Run panel's session status says whether the controls are available.
- **Test.** `e2e/interrupts.spec.ts` drives the real binary: the bench host starts with `plantwire host --script` (new
  `openScriptedBench` in the e2e support, 250 ms per barrier so the world runs four times faster than the wall clock, two
  `set_pwm_lease` commands to the simulated fan scheduled 200 s and 240 s of world time after the start), the stand-in gateway
  is the real board, and the test holds and releases the fan, turns approve-each on, approves one waiting command (the commands
  table reads "Refused" and "Policy not allowed", D-88) and denies the other (the table reads "Refused at the node: local
  control"), emergency-stops all devices (the latch is
  shown and released), checks axe and the fit at 1280 px, then stops the session and verifies the saved run, which holds
  `interrupts.jsonl`.
- **Deviations.** (a) The live bar's unreadable note also covers an `interrupt` record that does not decode. (b) The deadline of
  a world-clock command counts from the newest snapshot's time, so it moves per snapshot, not per second. (c) No new tokens: hold
  and approval use the warn tokens, a latch and the emergency stop the alarm tokens. (d) Scopes the node reports but no device of
  the newest snapshot belongs to (a board that has not announced) are not offered in the select. (e) Deferred (D-88): a
  canvas-connected session has no policy allow-list, so every command to a simulated device is refused `policy_not_allowed`.
  (f) Deferred (D-89): the relay drops a schema-invalid host record without a reason in the evidence.

## R12 as-built (binary)

The twin-versus-board report of section 3, built in the binary; the Inspect overlay is R12b. The canvas computes nothing:
matching, tolerance, divergence and which commands count are the binary's.

- **Ownership.** No new package. The pure rules are in the `plantwire` domain layer (`twin_*.go`): reading the board's trace,
  classifying the host's commands, tolerances, the evaluation and the report. The module already owns the field trace's
  interpretation (`PlanReplay`), the interrupt records and the kind files, and `TwinModel`, an interface the domain declares and
  the app supplies, is the one thing the rules need from a world. The app layer (`twin_report.go`, `twin_world.go`) reads the run
  folder, loads the twin domain and its kind files, builds the world and calls the rules; `plantwire.TwinReport` only delegates;
  the CLI (`plantwire twin-report`) parses flags and makes that one call; the bridge action `run.twin_report` runs the command.
  The `world` facade is the independent noise-free oracle the truth module already uses (`Noiseless`, `EmitDisabled`,
  `ForceEffectorOK`, readings without a clock): the twin's prediction is `World.Reading`, the noise-free observation of the
  channel the metric names. A domain of the catalog is the twin (`--domain`); `internal/domain` loads it, so no adapter is needed.
- **Input.** A run folder with `field-trace.jsonl` (checked against `run.json` when there is one, `run.ReadFieldTrace`; a run
  without one is refused with `ErrFieldTraceMissing`), the observer `journal.jsonl` (required: the host's `DCMD` records are
  there; checked against `journal_file_digest` of `run.json` when it records one) and `interrupts.jsonl` when the session had a
  control socket (digest checked, `run.ReadInterruptLog`). The twin domain must implement at least one kind of the board's devices
  (`ErrTwinKinds`), and every implemented kind is checked against its kind file as everywhere (`CheckImplements`: drift and a
  missing file are refused). A device of the board binds to the twin entity whose wire id equals the device id and whose kind
  implements the device's kind; a metric binds to the channel of the same name.
- **Which commands count.** The commands are the host's `DCMD` records in the journal (first occurrence of an id) to devices
  the field trace announces, in the order sent. A command is **applied** to the twin only when the gateway accepted it: the field
  trace holds a `DRECEIPT accepted` for its id (the latest receipt of the id), its `DRESULT` did not report `failed`, and the
  interrupt log, when there is one, did not refuse it (`verdict refuse`, `resolve refuse`) or leave it waiting (`verdict queue`
  never resolved). Every other command is listed with `status: not_applied` and a `reason`: `interrupt_refused`,
  `interrupt_unresolved`, `never_forwarded` (no receipt of the gateway: a relay refusal, a hold, a command never sent on),
  `refused_by_board`, `not_accepted` (a `queued` receipt only), `board_reported_failure`; `detail` holds the node's or the
  board's reason. A command the twin's world refuses (an unknown effector, bad arguments, an interlock) is `twin_refused` with
  the world's words in `detail`. Commands to devices a mixed session simulates are not part of the board and are not listed.
  A command takes effect at `at_ns`: the latest board event time recorded in the field trace before the gateway's receipt
  (D-90). The twin starts at the earliest reading time of the trace (a birth's time is not used: a gateway may stamp births 0) with its
  domain's initial state; a trace whose readings span more than 7 days is refused (`ErrTwinSpan`), because the world integrates
  from its start to every reading and a bogus clock must not hang the report.
- **Which readings are compared.** Each `DDATA` sample of a device, named through the device's latest `DBIRTH` (a sample whose
  alias no birth named is only counted: `summary.unattributed_readings`), is judged in this order and is `not_compared` with the
  first reason that holds: `historical`, `quality_not_good`, `not_numeric`, `no_twin_entity`, `no_twin_channel`,
  `no_tolerance_declared`, `before_first_command` (earlier than the first applied command: the twin has no commanded input yet;
  a run with no applied command therefore cannot be judged), `twin_not_finite`. Otherwise the twin and the board value are
  compared: `|board - twin| <= tolerance` is `within`, anything else is `diverged`. Commands and readings are processed in time
  order, a command first at equal times, so a lease renewal reaches the twin before the reading it precedes.
- **Tolerance.** Absolute, per metric, never invented: `--tolerance selector.metric=value` where the selector is the device id
  (tried first) or the kind name (value finite and at least 0; a key that names no device or kind metric of the board is
  refused, `ErrTwinTolerance`). Without a declared value the default is **`span_fraction` times (`eu_range.max` minus
  `eu_range.min`)** of the metric in its kind file, `span_fraction` being 0.05 unless `--span-fraction` says otherwise (in
  (0, 1]); the report states the rule (`tolerance_rule`) and each metric states its `tolerance.value` and `source` (`declared`,
  `default_span`, `none`). With neither, the value is `null`, every reading is `no_tolerance_declared` and the metric is
  `cannot_judge`. For the thermal bench: temperature `eu_range` 0 to 50 gives 2.5 degrees, humidity 20 to 90 gives 3.5 percent.
- **Report** (`twin-report-v0.1`): `{schema_version, verdict, reason?, summary, run, twin, tolerance_rule, commands, metrics,
  divergences}`. `verdict` is `diverges` when any compared reading is beyond its tolerance, `agrees` when readings were compared
  and none diverged, otherwise `cannot_judge` with `reason` `no_commands_applied` or `no_readings_compared`. `metrics` is sorted by
  device and metric; each has `tolerance`, `verdict`, `summary: {compared, within, diverged, not_compared}`,
  `not_compared_reasons: [{reason, count}]` (sorted) and `series: {points: [{at_ns, board, twin}], total, truncated}`: the
  twin's prediction at each compared reading, at most 2000 points, the first 2000 when `truncated`. `divergences` is one entry per
  stretch of consecutive diverged readings of one metric (readings that were not compared do not end a stretch; a reading within
  tolerance does): `{name: "device.metric#n", device, metric, first_ns, last_ns, readings, tolerance, worst_error, worst_at_ns,
  board_value, twin_value, commands}` where `commands` are the ids in force at the worst reading (for each device and command,
  the latest applied one at or before that time, ordered by time). Example:

  ```json
  {"schema_version": "0.1", "verdict": "diverges",
   "summary": {"devices": 1, "metrics": 1, "compared": 3, "within": 2, "diverged": 1, "not_compared": 1, "unattributed_readings": 0,
               "metrics_cannot_judge": 0, "divergences": 1, "commands_applied": 1, "commands_not_applied": 0},
   "metrics": [{"device": "zone-01", "kind": "thermal_sensor", "metric": "temperature", "unit": "Cel",
                "tolerance": {"value": 2.5, "source": "default_span"}, "verdict": "diverges",
                "summary": {"compared": 3, "within": 2, "diverged": 1, "not_compared": 1},
                "not_compared_reasons": [{"reason": "before_first_command", "count": 1}],
                "series": {"points": [{"at_ns": 1, "board": 34, "twin": 33.9}], "total": 3, "truncated": false}}],
   "divergences": [{"name": "zone-01.temperature#1", "device": "zone-01", "metric": "temperature", "first_ns": 9, "last_ns": 9,
                    "readings": 1, "tolerance": 2.5, "worst_error": 6.1, "worst_at_ns": 9, "board_value": 40, "twin_value": 33.9,
                    "commands": ["cmd-1"]}]}
  ```

  (abbreviated: `run`, `twin`, `tolerance_rule` and `commands` are also present.)
- **Determinism.** The report is a pure function of the run folder, the domain, the kind files and the options: no wall clock,
  no random number, every list sorted or in the file's order, floating point values as Go prints them. The app validates the
  document against the embedded contract before printing it. The analytic cross-check still guards the twin: the thermal
  bench's first-order lag is checked against its closed form by the world's own tests and by the report's test of a known
  command history.
- **Deviations.** (a) The command-line form is `--tolerance selector.metric=value` where the selector is a device id or a kind
  name, resolved by exact match, so ids that contain a dot work; there is no separate options document and no `--json` flag
  (stdout is always one JSON document). (b) The report needs `--domain` (the brief named only the directories). (c) A command's time
  is placed by sequence in the field trace, not by a clock (D-90, D-91). (d) The bridge action passes the kinds folder always.

## R12b as-built (canvas)

The twin-versus-board report and the Inspect overlay on the canvas. The binary keeps every rule (which commands count,
tolerance, divergence, prediction); the canvas asks for the report, decodes it strictly and shows it. It parses no message
text and computes no judgement.

- **Request.** The bridge action `run.twin_report` (registered in the canvas `rules.actionNames`, the only gate edit) is
  called as `{run, domain, tolerance?}` by the new `ReviewGateway.twinReport`. `domain` is the domain the session recorded
  (`state.sessionDomain`, from the `session` record); `tolerance` is the text the user typed, passed unchanged, and left out
  only when it is blank. The binary validates it and its refusal is shown as the binary reports it (`message` and code). A
  report that diverges or cannot judge is a result, not an error. The request is made only by the "Compare with twin" button,
  never on open.
- **When it is offered.** `twinOffer` (model): only for a saved session (`connection` saved, never while running and never for
  a plain run, which show nothing), and only when the run report lists a `field_trace` check (the binary lists it only when
  `run.json` records a field trace digest) and the session record names a domain. Otherwise the section says why in one
  line ("This run recorded no field trace, so it has no real board to compare.").
- **Decoder.** `twinReportDecoder.ts` (adapters) reads `twin-report-v0.1` strictly: another schema version, an unknown verdict,
  reason, command status, not-applied reason, not-compared reason or tolerance source, a count that is not a finite number, a
  `not_applied` command with no reason, a metric or divergence with a missing name, or a malformed point drops the whole report
  as unreadable (the store shows "unreadable" as a failure). Digests, `kinds` and `start_ns` are not read: the canvas shows
  none of them. Extra members are ignored, as in the other decoders.
- **Words.** One table per set in `twinWords.ts`: verdicts (Agrees, Diverges, Cannot judge, each with a line), the reasons a
  report cannot judge, command statuses, not-applied reasons, not-compared reasons, tolerance sources, and the line shown when
  there is no divergence. A test checks that each table has a word for exactly the values the decoder accepts.
- **Store.** `ReviewSession.twin` (`TwinStore`, `app/review`): idle, loading, ready or failed, each with the request that made
  it. A press while the same run is loading sends nothing; an answer that arrives after another run was asked for is dropped.
  `twinFor(state, run)` shows nothing of another run's answer.
- **Twin vs board (Report tab, saved sessions).** A section with the tolerance field ("Tolerances (optional)", described by
  its format), the "Compare with twin" button (`aria-disabled` while comparing, status line "Comparing the board with the
  twin…"), and then the result: the verdict as a chip by word and shape (check circle, warning triangle, question circle; the
  colour is secondary), the verdict's line, the reason for "Cannot judge", eight counts, the tolerance rule with the binary's
  description, the twin domain and the tolerances as typed. Focus moves to the result heading, or to the refusal (an alert).
  Three more sections follow a result: "Twin commands" (every command with its status and, for a command not applied, the
  binary's reason in words and its detail), "Where the twin and the board part ways" (each divergence by the binary's name
  `device.metric#n`, its interval, worst error with the board and twin value and the commands in force; the name is a button
  that moves the review cursor to the first reading, `review.seek`) and "Twin metrics and tolerances" (each metric with its
  verdict, its tolerance with the source and the readings it could not compare, with counts).
- **Inspect overlay.** For a real board selected in a saved session that offers the comparison, a "Twin prediction · <board>"
  section. Before a comparison it says so and has a button to the Report tab; while comparing or after a refusal it says that.
  After a result it draws one chart per metric of that board: the board's compared readings as a solid line with circle
  markers and the twin's prediction as a dashed line with square markers, divergence stretches shaded with a dashed outline,
  the playback cursor, start and end run times, three value ticks, a legend (Board reading, Twin prediction, Where they part
  ways) and the metric line (readings compared, beyond tolerance, tolerance and its source). A metric with no compared reading
  shows the binary's reasons in words; a board the report does not cover says so; a series cut at 2000 points says so.
- **Time.** The report's times are the board's own event times in nanoseconds. The canvas shows them with the same
  `runClock` the rest of the review uses, against the run's start, and seeks to them as run times. The chart's coordinates
  (`twinChart.ts`, model) are presentation only.
- **Tokens and CSS.** No new tokens: ok, alarm, accent, text and border tokens in both themes. One new module stylesheet,
  `Twin.module.css` (under the default cap, no `position`, no `z-index`, no `!important`). Desktop only, nothing clipped at
  1280 px (the e2e checks the inspector's width).
- **Test.** `e2e/twin-report.spec.ts` runs the real binary: a scripted host issues two `set_pwm_lease` commands to `fan-01`,
  the real board of the session (the stand-in gateway announces a fan), the session is stopped and saved, the review is
  opened, a bad tolerance is refused with the binary's words, and "Compare with twin" gives the honest report of this stand-in:
  **Cannot judge**, because the stand-in gateway sends no reading and refuses every command (`not_ready`), so both commands are
  listed as not applied and nothing can be compared. A diverging or agreeing report is therefore covered by unit and component
  tests with documents shaped by the contract, not by the e2e; the chart and legend appear only when the board reported
  readings. The e2e checks axe and the inspector's width at 1280.
- **Deviations.** (a) The "real trend" of the brief does not exist in Inspect (a real board shows a readings table at the
  cursor), so the overlay draws both series from the report: the board's compared readings and the twin's prediction at the
  same times (at most 2000 points, the first ones). D-92. (b) No span-fraction field: the action accepts `span_fraction`, the
  canvas does not offer it. D-93. (c) Choosing a divergence moves the cursor only; it does not select the board.

## M5 review fixes (Go)

The M5 review found these in R11a and R12a; every item has a regression test that fails on the old code. Contracts changed:
`node-control-v0.1`, `interrupt-log-v0.1`, `twin-report-v0.1`, `run-journal-v0.1` (embedded copies equal).

- **Commands queued for a future start time (node of the world).** `hold` put a command with a `not_before_ns` in `Session.queued`,
  and `runQueue` ran it without asking the interrupt rules, so a command accepted before an emergency stop started the device
  again after the latch. Now a control that restricts (`hold`, `estop`, `approve_mode` with `enabled`) and is applied refuses every
  queued command its scope covers: `ledger.Release`, a terminal `DRECEIPT refused, local_control` (boot `none`) for the command and
  for each repeat that waited with it, and a `purge` record. `purge` is evidence like `effect`: replay checks that a verdict
  `pass` for that command id precedes it and does not recompute it. The twin report reads a purge as an interrupt refusal. A
  command already forwarded to a gateway cannot be recalled (D-94).
- **Pending approvals are unambiguous and shown.** A `DCMD` whose `command_id` is already waiting (as a command or as a repeat) is
  refused `duplicate` before the hold and approval rules (a declared safe stop still passes). Each `pending` entry of the state
  view and of the live snapshot gains `content_digest` (the command's content digest: device, command, arguments, expected boot),
  `args` (at most 2048 bytes of canonical JSON, else `{}`) and `args_truncated`; the logged `command` input carries the same `args`
  and `args_truncated`, so the view replays. `approve` and `deny` may carry `content_digest`; a waiting command with other
  content makes the request `refused, unknown_command` and changes nothing.
- **The safe-stop exemption is the machine's.** The `command` input no longer has `safe_stop`. The rules pass a command when
  its name is the inventory's declared safe-stop command of that device, it has no arguments and they were not left out as
  too large. The inventory (`safe_stop` per device) is the one source for the node of the world (the effector flagged safe_stop)
  and the relay (the kind's command flagged `safe_stop`); a host `halt` with `{"mode":"run"}` is held.
- **Emergency stop.** It is not refused for capacity: it adds one latch, and the latch does not count against the 64 scopes
  and nothing is dropped (see "M5 re-review fixes (Go)"). The inventory entry of a device gains `unreachable` (the relay: no boot ever known, its `DDEATH` seen, the gateway not admitted, or the world gone; see "Stop boots" below); the
  reply gains `unreachable` beside `stopped` and `skipped`: a device that declares a safe stop and could not be ordered is listed
  there and never in `stopped`, and no `stop` is recorded for it. A node of the world performs the stops first and answers the
  host second, joining both errors, so a failed answer no longer skips the stops.
- **Relay stop identity.** A host command with an id starting `node-stop-` is refused `malformed`; each stop has its own
  idempotency key (a digest that includes the stop's id), a table of issued ids that evicts the oldest instead of clearing, and
  an `effect` record `ordered` written when the stop is sent, so the interrupt log names every stop even if the gateway never
  answers. An `effect` must follow a `stop` of its device (replay checks it).
- **A stalled host.** The relay is one loop; a host that stops reading used to hold it for the 30 s write deadline, and with it
  the controls and the stop to the gateway. Relay links now have a 2 s write deadline (a host that stalls is lost and redialled
  as before), `apply` sends to the gateway and the world before the host, and a tick serves a control after its own moves.
  The wait that remains is one bounded write plus one tick (D-99).
- **Stop boots.** Losing the host must not unreach the gateway. `forgetGatewayBirths` (host lost: the births must be announced to
  the host again) used to empty the relay's table of the gateway's boots, so an emergency stop right after a stalled host was
  dropped reported every device `unreachable` and sent nothing. The relay now keeps a second table of the last boot each device
  announced, written from every `DBIRTH` and `DDEATH` the admitted gateway sends (also while the host is away, and not for a stray or
  forged record), and the stop path reads only that table: the stop is addressed to that boot (a device that has rebooted since
  refuses it `wrong_boot`, which is safe), with the host's last authority epoch, which is kept when the host is lost. An entry
  is removed by the device's `DDEATH` or when the gateway is lost or readmitted (and a simulated device's when the world is gone);
  a gateway rebirth request, a host loss and the host's rebirth leave it. The table is derived from the same recorded gateway
  records and shows only through `inventory.devices[].unreachable`, so the log and its replay did not change.
- **A log that fails.** A session closes its interrupt log and replays it; when it does not replay (or a record could not be
  written) the file is renamed `interrupts.rejected.jsonl`, `run.json` records no digest, a warning is logged and
  `plantwire connect` fails with `ErrInterruptLogRejected` (the replay error inside). A node of the world also logs the end of
  the session: `flush` and a `resolve` `session_ending` for each waiting command (no receipt is sent, D-97).
- **Twin report.**
  - *Node stops.* The report turns each `effect` command id of the interrupt log (`node-stop-N`, the relay's stop to the gateway)
    into a command with `origin: "node"`, the device's declared safe-stop command and arguments `{}`, classified from the gateway's
    receipt in the field trace like a host command and applied to the twin at the same time rule. When a stop the node ordered
    never got an answer (`never_forwarded`, `not_accepted`) the report is `cannot_judge` with `reason: "node_stop_unconfirmed"`,
    and when the twin refuses it, `"node_stop_not_applied"`; both take precedence over `diverges`. A board that refused the stop
    leaves the verdict to the readings.
  - *Evidence.* The interrupt log is read through `app.VerifyInterruptLog` (digest, schema, rules replay) and a log that does
    not replay is refused (`domain.ErrInterruptReplay`); a missing `interrupts.jsonl` that `run.json` records is `ErrTwinNoInterruptLog`;
    `run.evidence_bound` is true when `run.json` exists and recorded the journal digest the file was checked against, false for a
    folder with no `run.json` (the files are then evidence of nothing but themselves). `journal.jsonl` is read with the 64 MiB
    limit through `kernel.ReadRegular`, so a link or a device in its place is refused (`kernel.ErrNotRegular`, `ErrTooLarge`).
- **Control socket.** The server's line buffer is one byte larger so the bridge's longest line (4096 bytes) is accepted, and an
  idle connection is closed after 60 s. The bridge makes the socket directory under the temp directory and, when the path is
  longer than a unix socket allows (103 bytes), under `/tmp`; when no base is short enough the session starts without controls
  (`ErrControlPathTooLong`).

Exact shapes (new members only):

```json
{"op": "approve", "command_id": "cmd-7", "content_digest": "sha256:..."}
{"result": "applied", "op": "estop", "stopped": ["fan-01"], "skipped": ["zone-01"], "unreachable": ["fan-02"], "state": {...}}
{"command_id": "cmd-7", "device": "fan-01", "command": "set_pwm_lease", "deadline_ns": 1, "queued_at_ns": 1, "clock": "world",
 "content_digest": "sha256:...", "args": {"duty_permille": 300}, "args_truncated": false}
{"event": "purge", "command_id": "c-1", "device": "fan-01", "reason": "local_control", "clock": "world", "node_ns": 1}
{"event": "inventory", "devices": [{"device": "fan-01", "board": "mega-01", "clock": "wall", "safe_stop": "stop_fan", "unreachable": false}], ...}
```

Twin report: `commands[].origin` (`"node"`, absent for the host), `run.evidence_bound` (boolean, always present), and the `reason`
values `node_stop_not_applied` and `node_stop_unconfirmed`.

## M5 re-review fixes (Go)

A second review of the fixes above. Every item has a regression test that fails on the old code. Contract change:
`node-control-v0.1` reply and `interrupt-log-v0.1` require `unreachable` on an applied `estop` reply (embedded copies equal).
Shapes of records are otherwise unchanged.

- **Purge of an approved command (H1).** A command a hold, approval or stop released and a node of the world then queued for
  its `not_before_ns` can be purged by a later restriction. Replay now accepts a `purge` for a command id that a `pass` verdict
  or a `resolve` with `forward` (an approval) preceded, on the device it was sent to, when a hold, stop or approval mode that
  covers the device was applied after that point; any other purge is refused.
- **The stop epoch (H2).** The relay's stop carries the last authority epoch a host announced; an epoch of 0, announced by a host
  that returned without one or kept when it is lost, never replaces it.
- **Retry (M1).** A `DCMD` whose id is waiting for approval with the same content and idempotency key is answered `queued` again
  and adds nothing; other content under that id is `duplicate`.
- **Stops (M2).** A node of the world invokes every safe stop of an emergency stop first, then tells the host the state of each
  device, joining the errors; a failed send no longer leaves a device running.
- **Scopes (M3).** Nothing is dropped any more. The bound of 64 counts approval scopes and plain holds; the latch of an emergency
  stop has its own place and is never refused for capacity.
- **Control port (M5).** An emergency stop is queued in a list of its own that is served first and has its own bound (16), and a
  request whose caller gave up is dropped from the backlog.
- **Twin report (M6).** A run folder that holds `interrupts.rejected.jsonl` is refused with `ErrTwinInterruptLogRejected`: the
  commands the node stopped or refused are unknown.
- **Offline devices (H3).** The relay lists in the inventory every device announced in the session by a gateway `NBIRTH` board
  (kept after the gateway is lost, and after a later birth leaves it out) and by the simulated node, with `unreachable` computed
  per device. An emergency stop of `all`, its board or itself names an offline device in `unreachable`, and a hold or approval
  mode scoped to it is accepted. The canvas shows the same lists and needs no change.
- **Smaller.** A hold or stop does not purge a queued declared safe-stop command. `interrupts.rejected.jsonl` is never
  overwritten: a second one is `interrupts.rejected.1.jsonl`, up to 100 (`ErrTooManyRejectedLogs`). `content_digest` on `approve`
  and `deny` stays optional; sending it is how an operator is safe against a reused id. The stall test allows 4 s.

## M5 review fixes (canvas)

The real-browser, code-quality and adversarial reviews of the canvas side of R11b and R12b. The binary keeps every rule; the
canvas shows what it says. Every item has a regression test that fails on the old code, except the layout items, which the
e2e at 1280 x 800 checks.

- **The bar is always visible (H1).** On the live view the Interrupts region and the live bar sit above a part of the screen
  that scrolls (`ReviewScreen.module.css`: `.liveTop`, then `.body`), so scrolling can never move the controls out of the window.
  The graph row shrinks (`minmax(236px, 1fr)` for the plant with its bar) instead of pushing the page past the window. The
  bar is one row at 1280 px (scope select, Hold, Release, Approve each command and Emergency stop; the title is a hidden heading
  and the scope label is hidden text), and it holds a place for its status row (40 px) and its waiting list (72 px), so the plant
  does not move when chips, an outcome or a waiting command appear. The e2e (`interrupts-layout.spec.ts`) checks the region
  inside the window and fully visible after scrolling, the plant's bottom inside 800 px and its top within 2 px while a command
  waits and after a stop latches.
- **The emergency stop is never dropped (H2).** `ControlStore` keeps the controls in flight as a list with an order number. An
  emergency stop is always sent, also while another control waits for the node; any other control pressed while one is in
  flight is not sent, `send` returns `"not-sent"` and the bar says "not sent: another control is still waiting" (the note stays
  until the next control is sent). The Hold, Release and Approve each command buttons stay enabled and give that answer; only
  Approve and Deny of a waiting command are disabled while a control is in flight. Focus moves to the waiting heading only when a decision was sent. The stop is enabled whenever the
  control socket is offered, without waiting for the node's first state. The newest outcome (by order) stays: a stop's outcome
  is not replaced by an older control's, and while an older control is still waiting the stop's outcome is shown with a note.
- **The plant of a running session opens by address (H3).** `showRunPlant` takes the run's domain state and its connection;
  while the recorded plant is still loading for a running session it loads the plant the session record names, as it
  already did after a failure.
- **A session's controls belong to the session (H4).** `ControlState` carries the run it belongs to, `ControlSession.bind`
  resets the answers, the outcome, the blocked request and the picked scope when the review store's `worldId` changes, and
  `newestInterrupts` returns a state seen only for the run on screen. The picked scope lives in the controls store, so the plant
  screen and the live view share it. A control's answer that arrives after the run changed is dropped. After "Stop and save" the
  Run panel offers "Set up another run".
- **Controls follow the socket and the session (M1).** Hold, Release, Approve each command and the Approve and Deny of a
  waiting command need the state, the control socket and a session that is not stopping; the stop needs the socket and a
  session that is not stopping. Each missing condition has its note.
- **A picked scope that left the session stays picked (M2).** It is listed as "Board gone (not in this session)", Hold, Release
  and Approve each command are disabled for it with a note, and the stop stays enabled.
- **Answer against snapshot (M3).** The store remembers the states the answer superseded (`SeenState.older`); a snapshot equal to
  one of them and different from the answer is not shown (D-102). A state read after a failure carries a generation, and a read
  that finishes after a newer answer is dropped without a claim that it was read.
- **Release and coverage in words (M4).** A release says what the node's returned state still holds: "Released all devices. 2
  narrower holds remain: board uno-07, device fan." or "Released device sensor. Device sensor is still held through board
  uno-07." A scope covered by a wider hold or approval says so under the chips ("Device sensor is also covered by Board
  uno-07: Held."). Both read the returned state's holds and covered devices (`interruptCover.ts`); nothing else is computed.
- **Disabled controls look disabled (M5).** The bar's buttons use the disabled tokens (`--disabled-*`) and `cursor: not-allowed`;
  a pressed toggle also shows a check mark, not only a colour.
- **The command table fits (M6).** Command evidence has two columns, the command and its four stages as labelled lines (Sent
  at, Acceptance, Board report, Effect evidence), so nothing is hidden at 1280 px; its scroll region is a named region. Table
  cells in the report take 8 px of side padding, an instance id may wrap and the outcome column starts at 7 em, which keeps the
  Effector calls table inside the panel (`review.spec.ts` checks it at 1280 px).
- **Twin result focus (M8).** `TwinState` ready and failed carry `fresh`; the result heading (or the refusal) takes focus when
  `fresh`, and `TwinSession.markSeen` clears it, so a change of tab does not move focus again.
- **Truncated series (M9).** A divergence that does not intersect the readings drawn is not drawn and not counted in the chart's
  name; the note under the chart says how many lie beyond ("1 stretch where they part ways lies beyond the readings shown; the
  Report tab lists it.").
- **Interrupt marks (M10).** The timeline mark for an interrupt is an octagon (like the stop sign of the button), not a rounded
  square, so it differs from binding and effector marks by shape.
- **Announcements (M12).** A newly arrived waiting command is announced in a polite status line; the outcome line is a status
  region marked busy while a control is in flight or its state is read again, so only the final outcome is announced; the
  once-a-second tick runs only while a command waits and the deadline is worded "expires in about …"; the waiting command's name
  wraps instead of being cut; the arguments of a waiting command are shown (bounded to six pairs of forty characters, or "arguments
  too large to show").
- **Failure words (M13).** An unknown 4xx code of a control is `rejected` ("The bridge refused the request, so nothing was
  confirmed.", with the bridge's own message), not "The answer could not be read". The code is looked up in a map, so a code
  such as `constructor` is not read from the prototype; the refusal-reason table is a map for the same reason.
- **New binary members.** The decoder reads, and requires, `args`, `args_truncated` and `content_digest` of each waiting command;
  approve and deny send the `content_digest` the operator saw (`content_digest` in the request). The stop's reply `unreachable`
  (optional) is worded "The stop could NOT be sent to …: they may still be energised." and makes the outcome a failure. A twin
  command has `origin` (absent means host; `node` shows "Node safe stop"), the report requires `run.evidence_bound` (false shows
  "Evidence is not bound to a run.json …"), and the judge reasons `node_stop_not_applied` and `node_stop_unconfirmed` have
  words. `interrupt` journal records with the event `purge` are marked on the timeline ("c-1 refused: Refused at the node: local
  control"). Text members of every interrupt and twin record must be text: a number where a device, command or reason belongs
  drops the record. An interrupt record that does not decode has its own note in the live bar, not the one about a snapshot.
- **The twin comparison is offered more widely.** `twinOffer` no longer keys on the check named `field_trace`: any saved session
  whose record names a domain offers the comparison and the binary's refusal explains a run with no board. The rule description
  is the binary's `tolerance_rule.description`, shown as given.
- **Deviations and what is left.** (a) The `twins/*.twin.json` 404s on the device list need a bridge listing: D-101. (b) The
  snapshot-versus-answer guard is by content: D-102. (c) The plant node tags keep their short words: D-103. (d) The e2e pins the
  overlay of the stand-in session to "The board reported no reading of duty_permille to compare." (the stand-in announces a fan
  with one metric and sends no reading).

### M5 second pass (canvas)

- **One stream per run per tab.** A run the workspace lists as running is shown from the shared live review store
  (`ReviewRoute`), so the plant screen, the live view and a reload all hold the one run-events stream; the review store opens no
  second one. The connect store stops following the run journal as soon as it has the binding. A tab then holds `/api/events`
  and one run stream, so two tabs stay inside the browser's six connections per host. `FakeBridge.openFollows` counts open
  streams in the tests; `interrupts-layout.spec.ts` opens the same running run in a second page.
- **One row in every state.** The bar is a non-wrapping row: the scope select shrinks first (80 to 150 px), "Approve each command"
  reads "Approve each" with " command" as hidden text (the accessible name is unchanged), and the check mark of a pressed toggle
  is always laid out (hidden when off), so applying a hold moves nothing. The e2e checks Hold on plus Approve on at 1280 x 800:
  one row, the stop's position unchanged.
- **Refit.** `useFitted` recorded the viewport it asked for; when React Flow clamped it the canvas took the difference for a
  manual move and never refitted on a resize. It now records the viewport that was applied, so a pane that changes height is
  refitted while the user has not moved the view. A manual zoom or pan, or the Fit button, is not refitted over; a resize after
  the Fit button is not refitted until the fit key changes (not changed here).
- **One interrupt source.** Node tags, the inspector line and the bar read the same state: `withShownInterrupts` puts the state
  `newestInterrupts` selects into the newest snapshot (`useShownLive`); a snapshot at an earlier cursor keeps its own.
- **Another run.** "Set up another run" also empties the plan draft (sockets and boards).
- **Details.** The refusal reason `duplicate` reads "A command with different content already uses this command id or key". The
  waiting-commands list is not a tab stop (its buttons are). The "First: …" line under the timeline wraps and a timeline chip's
  title carries its full text. An applied emergency stop reply without `unreachable` is refused by the decoder (the contract
  requires it).

## Policy allow-list P1 as-built (binary)

Design: [POLICY-ALLOW-LIST](../physical-devices/POLICY-ALLOW-LIST.md) (D-88). P1 is the contract and the binary; the bridge
file class and the canvas form are P2.

- **The plan carries the list.** `session-plan-v0.1` gains optional `policies`: 1 to 8 strings of `sha256:` and 64 lower-case
  hex digits (`minItems`, `maxItems`, `pattern` at `/policies/<i>`). The `plantwire` domain rule `NewAllowList(digests, pointer)`
  (`internal/plantwire/internal/domain/policies.go`) is the one place the list is judged: `policy_malformed`,
  `too_many_policies`, `policy_listed_twice` (at the second occurrence), `policy_reserved` (the relay's all-zero stop digest);
  a list with an issue allows nothing. `CheckPlan` joins its issues (pointer `/policies`) with the binding issues; the flag
  sources (`plantwire node --policy`, `plantwire relay --domain ... --policy`) go through the same rule as `ErrBadOptions`,
  at `Serve` for a node and at the simulated node's start for a mixed relay, before anything connects.
- **Enforcement.** `Session` holds the built list (`allowed`); `authorize` keeps the epoch check and then asks
  `allowed.Allows(c.Policy)`. A queued, approved or waiting command is checked again at execution, as before. A safe stop skips
  the check, as before. The connected session takes `p.plan.Policies.Digests()` for its node (node mode and the simulated side
  of a mixed relay); `ConnectOptions.Policies` and `plantwire connect --policy` are gone.
- **Echo and record.** `plantwire connect --check` returns `policies`: the list in effect, sorted, `[]` when the plan names
  none (never null). The `session` journal record's `plan` is the same document, so it records the sorted list
  (`run-journal-v0.1` `session.plan.policies`, required, at most 8). The list is a recorded session input, not part of the
  run's identity: `run.json`, the trace digest and the command log are unchanged, and a connected session replays as before. A
  command the list let through is in the command log as any command is; one it refused never reached the world.
- **The refused digest.** `live_command.policy_digest` (optional, pattern `sha256:` and 64 hex) is present only when the receipt
  is `refused` with reason `policy_not_allowed`. `Live.sent` keeps the command's digest in an unexported field and
  `Live.receipt` copies it when the refusal arrives, if it is well formed. This holds for a simulated device (refused by the
  node) and a real board (refused by the gateway). The plant wire contract is untouched (`DRECEIPT` gains nothing).
- **The reference host.** `plantwire host` with a `--script` and no well-formed `--policy` (the reserved all-zero digest
  included) is a usage error before it listens (`plantwire.CheckHost`, the same `NewAllowList` rule; `ServeHost` applies it
  too). The `--policy` help text says the host "carries" the digest; it signs nothing.
- **Gate edits.** None. New files carry no I/O and need no register entry.
- **Deviations from the note.** None in behaviour. `Serve` checks the node's flag list before it dials, so `node --policy nope`
  fails as bad options without a host.

## Policy allow-list P2 as-built (bridge and canvas)

Design: [POLICY-ALLOW-LIST](../physical-devices/POLICY-ALLOW-LIST.md) (D-88, closed). P2 is the bridge file class, the canvas
form and the commands-table digest.

- **The default file.** `settings/default.policies.json`, `{"policies": [...]}`, is the one file of `settings/`
  (`FileClass` `policies`, `domain.ParsePath` accepts exactly that name; any other name in `settings/` and the file in `domains/`,
  `kinds/` or `components/` are `bad_path`). It goes through the existing `GET /api/file` and `PUT /api/file`: `etag` is
  `domain.Version`, a stale or missing etag on an existing file is `conflict`, the write is atomic and owner-only, the cap is the
  generic 8 MiB, an absent file reads `404 not_found`. The bridge checks path and size only: the canvas saves what the binary
  echoed, and the binary validates digests when the plan is checked. The file is not in the workspace listing or the change feed,
  and no action argument resolves to it. `session.check` and `session.start` need no bridge logic: the plan travels unchanged and
  the check document and the `session` record carry `policies`.
- **Form and default.** The Connect session panel has "Allowed policy digests" (a textarea, one digest per line, empty means
  none; the canvas deduplicates, sorts and checks nothing). `ConnectGateway` gains `readPolicyDefault` and `savePolicyDefault`
  (the bridge adapter implements them with the file routes; an absent file is `absent`, a file that does not decode as exactly
  `{"policies": [strings]}` is `unreadable` and pre-fills nothing). `ConnectSession.prefill(domain)` runs when the panel opens:
  it re-reads the default, learns the etag, and fills a draft the plant has never held (marked `policiesFrom: "default"`, shown
  with "Pre-filled from the workspace default."); typing marks it the operator's and removes the note; "Clear the list" empties it.
  A draft that exists is never overwritten, also when the read finishes after the operator typed. Deviation from the note:
  `draftFor` stays synchronous and the fill is the async `prefill`, because the default is read through the bridge.
- **Save as default.** The button exists only after a passing check of the current draft that is not stale; the store checks the
  same (`saveDefault(domain)` returns on any other state) and saves the check document's `policies` (the binary's sorted list),
  never the typed text, with the etag of the last read (read again first when that read failed; an unreadable default is
  overwritten only by this explicit action). An empty list is saved as `{"policies": []}`. A `conflict` reports that the default
  changed elsewhere, saves nothing, and re-reads the file so a second press saves from the new version. The result is
  `connect.defaults` (`PolicyDefaultState`), a store beside the connect state.
- **List in effect.** The check result and the running session show "Allowed policies: none. Commands to simulated devices are
  refused, except safe stops." or "Allowed policies: N" with each digest shortened (first and last 8 hex digits, the full digest
  for assistive technology) from the binary's check document. The review decodes `plan.policies` of the `session` record into
  `ReviewState.sessionPolicies` (strict: not a list of strings drops the record; absent reads "Allowed policies: not recorded.").
- **The commands table.** `live` command entries are decoded with an optional `policy_digest`, strict: it must be a well-formed
  digest and appear only on a `refused` command with reason `policy_not_allowed`, otherwise the record is dropped as unreadable.
  The detail reads "Policy not allowed by this session" for a simulated device and "Policy not allowed by the gateway" for a real
  board. A "Policy" row under the stages shows "Host's policy sha256:bbbbbbbb…bbbbbbbb" with "Copy policy digest" (a new
  `Clipboard` port, `copyText` on the review session, `browserClipboard`) and "Use this digest". Copy says "Copied the policy
  digest of <command id>." or, if the browser refuses, "Could not copy; select the digest instead." and shows the full digest.
  "Use this digest" is shown when the session record names its plant; it appends the digest to that plant's draft
  (`ConnectSession.addPolicy`, once per digest, marked the operator's) and says the running session keeps its list.
- **Set up another run.** Keeps the list (and its source) of the last draft; the sockets and the real boards are cleared as before.
  Deviation from the note: without it a digest added from a refusal would be lost before the next plan could use it.
- **Gate edits.** None. No new module, stylesheet or I/O edge; `architecture.json` allow-lists are unchanged (callers are
  regenerated).
- **E2E.** `interrupts.spec.ts` lists the scripted policy: the approved `ref-a` is `Accepted` and `Board reports success`, and the
  saved run's command log holds an `effector.invoke` and still verifies. `policies.spec.ts`: a session with no list shows the refusal
  with the digest, Copy puts the full digest on the clipboard, "Use this digest" fills the next plan; Save as default writes the
  sorted list and a reloaded panel is pre-filled and says so; the binary's refusals sit on the field by digest.
