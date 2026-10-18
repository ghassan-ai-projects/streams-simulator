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
input came in R8 (below); clock pacing (D-71) and hold/approve/e-stop (R11) are not built.

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
effect unknown. The relay forwards host content unchanged and records only evidence matching the routed command and
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
  expectation. A receipt or result from another boot, a regressing answer, duplicate or premature evidence changes nothing.
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
  The relay forwards `HSYNC` content unchanged (only `through_seq` is mapped, as before).
- **Gateway watchdog (D-73 fixed).** The relay counts ticks since the last admitted gateway record; stray, forged, malformed
  and repeated records do not reset it. After 30 s it reports the devices `DDEATH link_down`, closes the gateway link and
  records the gateway offline. It arms at admission, without a host heartbeat.
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
