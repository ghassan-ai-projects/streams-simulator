# Physical devices: a real board beside simulated ones

> Status: Implemented for the binary, the bridge and the canvas; the physical side is not verified by this repository. Authority: `internal/plantwire`, `internal/canvasbridge`, `canvas/`, the [plant wire contract](../../contracts/plant-wire/SPEC.md) and the session, node-control, interrupt-log and twin-report contracts. Verified by: unit, contract, component and end-to-end tests that run the real binary against a scripted stand-in gateway. No test or claim here involves a real microcontroller; the bench checks are the owner's. Last verified: 2026-10-19.

This guide is for someone who has to test a stream processor (a consumer, together with its plant wire host) against a real board that sits next to simulated ones, and who wants the session recorded so it can be replayed. It explains the concepts first, then one path through the canvas and the command line, then the limits.

## What this is for

A plant wire host talks to one node. The node can be a simulated world, an edge gateway in front of hardware, or, in the lab, the simulator standing between the two. The simulator can therefore:

- define a device once, as an interface, and run a simulated twin of it before any hardware exists;
- capture that interface from a board that announces itself;
- put the real board and simulated boards in front of the host as one node;
- show the session live: readings, command stages, connection health;
- let an operator hold, approve or stop commands at the node;
- record the session and replay it byte for byte;
- compare the twin with what the board reported.

It does not make the board part of the simulator's truth. A simulated device has sealed, generative truth. A real board has none, and what it reports is only what it says.

## Concepts

| Term | Meaning |
| --- | --- |
| Kind file | `<name>.kind.json`: one plant wire kind, which is the interface of a device (metrics with type, unit and range; commands with argument bounds, risk class and safe stop). Its identity is `kind_digest`. A kind file never carries that member. |
| `implements` | A domain kind that declares `implements: "<kind>"` is a simulated implementation of that kind file. The binary refuses a domain whose interface differs from the file and names the first differing member. With `board_type`, the kind is the board's raw model and implements the file through that board type's mapping (`<board_type>.board.json`). |
| Twin | The simulated implementation of a kind: a domain that implements the kind files of a board. `kind twin` scaffolds one from kind files and a choices document. |
| Gateway | The program in front of the hardware. It speaks the plant wire as a node and keeps the hardware protocol to itself. It is outside this repository. |
| Plant wire node | A program that dials a host's owner-only Unix socket and announces devices. `plantwire node` is the simulated one. |
| Relay | `plantwire relay`: the simulator between a gateway and the real host, passing the gateway's records up and the host's commands down with their content unchanged. |
| Mixed session | A relay (or `plantwire connect` with a gateway) that also runs the simulated world, so the host sees real and simulated boards in one node whose environment is `mixed`. |
| Field trace | `field-trace.jsonl`: every record the gateway sent in a session, in order. It is an input of the run, so a mixed run is a function of the simulator's usual inputs and of the trace. |
| Origin | `simulated` or `hardware`, the environment of the board a device was born on. It is never inferred from a name or a kind. The canvas marks it by shape, colour and words, and says "Origin unknown" when it cannot tell. |
| Session plan | `session-plan-v0.1`: which plant, which sockets, which instances are real boards, which policy digests simulated devices accept, which faults strike simulated instances when (`disturbances`), and whether the session is paced to the wall clock (`pace`). |
| Live mode | The review of a session that is still running; it follows the newest record. Scrubbing back pauses it; "Back to live" resumes. |
| Command stages | A command has four stages, kept apart: **sent** (the host issued it), **receipt** (`pending`, `accepted`, or `refused` with the reason), **board report** (`pending`, `succeeded`, `failed`, `unknown`; self-reported) and **effect** (`unknown`, or `observed` when the host's own evidence says so). A receipt or a board report never stands for an effect. |
| Connection health | Host link, gateway link, per-device health, boot id and reboot count, and a state for each reading: `good`, `stale`, `disconnected`, `invalid`, `untrusted` or, for a simulated device, `simulated`. A reading that is not `good` is shown as its state, never as a value. |
| Node interrupt | A control at the node: **hold**, **approve each command**, **emergency stop**. The node may only refuse or stop. |
| Latch | An emergency stop leaves a hold on its scope until the operator releases it. A release must name the same scope as the hold or stop: a hold on a device is not cleared by releasing `all`. |
| Twin report | The comparison of the twin with the board's recorded readings under declared tolerances. It measures the twin, not the consumer. |
| Policy allow-list | The policy digests a session accepts from the host in a command (see [the policy allow-list](#the-policy-allow-list)). |

## What you need

- A plant wire host (your consumer's host, or `plantwire host` as the reference host) that creates an owner-only Unix socket.
- For a real board: a gateway that speaks the plant wire version 1.4.0 and announces the same kind files. Its requirements are in the [plant wire contract](../../contracts/plant-wire/SPEC.md); the simulator accepts only the contract version and digest it was built with.
- A workspace folder: `domains/`, `kinds/`, `components/`, `twins/`, `settings/` and `runs/`. The canvas creates what it needs. Keep a workspace outside the repository when it holds customer data.
- All sockets in owner-only directories (mode `0700`). The wire has no authentication; the directory is the access control, and the simulator refuses a socket others can reach.

A Unix socket path is at most 103 characters (the session plan schema enforces it). Put the sockets in a short directory such as one made by `mktemp -d`.

## The path

The same steps work in the canvas and on the command line. The canvas calls the same binary commands and computes nothing itself.

### 1. Define a device

A device is an interface (kind files) plus, if you want to simulate it, a behaviour (a twin).

**Canvas.** Open **Devices**, choose **New device**, and fill the interface: each reading (name, type, unit, range, resolution) and each command (arguments, risk class, safe stop). The Interface step shows what a twin still needs before you reach **Behaviour**. In Behaviour, choose the library instrument for each reading and one response for each command that is not a safe stop (`direct`, `first_order` or `none`), then **Check**. Saving writes the kind files and the twin; a definition the binary refuses is never written.

**Command line.** The thermal bench ships as an example in `examples/devices/thermal-board/`: the sensor, indicator and fan
kind files the plant wire speaks (the fan takes `set_cooling {level_percent}` and reports `level_percent`; the indicator takes
`show_state {state}` and reports `state`), the board mapping `arduino-thermal.board.json` with the board's raw values (duty,
pattern, the 5 s fan lease), and the hand-written twin domain `thermal-board.domain.json`, the board's raw model, whose fan and
indicator declare `board_type: "arduino-thermal"` and implement their kind files through the mapping. A twin is scaffolded with
`kind twin` only from kind files whose metrics are numbers set by their commands; the thermal kinds are semantic (a word, a level
with a board lease), so their twin is the raw domain, checked through the mapping:

```bash
go run ./cmd/streamsim kind validate examples/devices/thermal-board/*.kind.json
go run ./cmd/streamsim domain validate examples/devices/thermal-board/thermal-board.domain.json \
  --kinds-dir examples/devices/thermal-board
```

`domain validate` prints `implements_checked` with the `kind_digest` of each kind file the twin was checked against. A twin that drifts from its kind file fails with `domain_invalid` and an issue at `/kinds/<i>/implements`.

The steps below work in a workspace `$W` that holds the example: the kind files and the board mapping in `kinds/`, the twin
domain (id `arduino-thermal`) in `domains/`, and the runs in `runs/`:

```bash
W=$(mktemp -d)
mkdir -p "$W/domains" "$W/kinds" "$W/runs"
cp examples/devices/thermal-board/*.kind.json examples/devices/thermal-board/arduino-thermal.board.json "$W/kinds/"
cp examples/devices/thermal-board/thermal-board.domain.json "$W/domains/arduino-thermal.domain.json"
```

### 2. Capture a device from a board

When a real board announces itself, its kinds can be saved as draft kind files. A birth is a claim, so the binary only drafts; a person reviews and saves.

**Canvas.** **Capture from a board** reads the `NBIRTH` of a recorded run and shows one card per kind: **New**, **Same** or **Differs** from the workspace kind file, with the members the board asserted (risk class, safe stop, argument bounds) listed for review. **Save as device** writes the draft exactly as announced; replacing a file needs confirmation.

**Command line.**

```bash
go run ./cmd/streamsim kind capture "$W/runs/<run>/journal.jsonl" --kinds-dir "$W/kinds"
```

Only the observer journal is read, never the sealed one. The command recomputes every `kind_digest` and refuses a birth that does not hold. The board's `environment` is its claim and is not verified.

### 3. Connect a session

The plan says which instances of the plant are real boards behind the gateway. Every other instance is simulated.

**Canvas.** On a plant, open the Run panel, choose **Connected session**, give the host socket and the gateway socket, **Add a real board** for each real instance (instance, board, device), list the **Allowed policy digests**, **Check the plan**, then **Start the session**. A refused plan names the field and the rule. Only one session runs at a time. The canvas paces every connected session to the wall clock (`pace: "wall"` in the plan it writes): the world starts at the session's wall time and advances one world second per wall second.

**Command line.** Write the plan, check it, then start the node.

```bash
D=$(mktemp -d)
POLICY=sha256:$(printf 'b%.0s' $(seq 64))
cat > "$D/plan.json" <<EOF
{"schema_version": "0.1", "domain": "arduino-thermal", "seed": 1,
 "host_socket": "$D/host.sock", "gateway_socket": "$D/gateway.sock", "pace": "wall",
 "real": [{"instance": "zone-01", "board": "uno-07", "device": "zone-01"}],
 "policies": ["$POLICY"]}
EOF
go run ./cmd/streamsim plantwire connect --plan "$D/plan.json" --check \
  --domains-dir "$W/domains" --kinds-dir "$W/kinds"
```

The check prints each instance as `real` (with the board, device and kind digest the gateway must announce) or `simulated`, and the policy list in effect. Start the host first (any plant wire host; the reference host is `go run ./cmd/streamsim plantwire host --socket "$D/host.sock" --environment mixed --policy "$POLICY" --delay-ms 250`), then:

```bash
go run ./cmd/streamsim plantwire connect --plan "$D/plan.json" \
  --domains-dir "$W/domains" --kinds-dir "$W/kinds" \
  --control-socket "$D/control.sock" --out "$W/runs/c-lab1"
```

The gateway dials `gateway_socket`. The simulator checks it against the plan at its handshake: the bound board and device must be announced, as the kind the plan expects, with the kind file's `kind_digest`. A gateway that does not match is refused `inventory_mismatch` with the first failing rule as the reason, and the simulator keeps listening. A refused binding shows as "Not bound", and the world does not announce that instance to the host either.

The fan experiment ships ready-made plans for both cases, a simulated rehearsal and the real board, in
[`examples/devices/thermal-board/fan-experiment/`](../../examples/devices/thermal-board/fan-experiment/README.md).

Without `gateway_socket` the session is a simulated node only; with it, a mixed relay. With `pace: "wall"` the world is anchored when the host attaches (a late host or gateway gets no burst of past slices) and each slice is released when the wall clock reaches it, so world time is UTC and a 10 minute session takes 10 minutes; the barrier stays a cap. Without `pace` (offline tests and benchmarks) the world advances barrier by barrier as fast as the host answers.

**Give the reference host `--delay-ms` in an unpaced session.** Against the reference host with no `--delay-ms` and no `pace` the world free-runs: a 6 second run wrote about 450 MB (about 75 MB of run folder per second). A run folder that large cannot be reported on: `plantwire twin-report` refuses a run whose `journal.jsonl` is larger than 64 MiB, and a field trace or an interrupt log is capped at 64 MiB as well. Use `--delay-ms 250` (or more), and stop sessions in minutes, not hours (see the [limitations](../limitations.md#physical-devices), D-77).

`connect` needs the host's socket: without it the command fails `file_not_found` after it has created the run folder (`journal.jsonl`, `ledger.jsonl`, `run.json`) and any missing parent directories (mode `0700`). The reference host prints its report and exits 0 when its node ends, also when you stop `connect` while the host is waiting out a `--delay-ms`, and when it is interrupted before any node arrived (then with an empty report).

### 4. Run live

In the canvas, **Open live view** shows the session as it runs.

- Each device carries its origin. A real device shows board and boot id; a simulated one shows "Simulated · no link".
- Connection health shows the host link and the gateway link, a board going offline and coming back, and reboots.
- A reading that is stale, disconnected, invalid or untrusted shows that state, not a value. A reading of a real device turns `stale` after three of its declared scan intervals without a sample, or after 30 seconds when the kind declares no scan interval.
- The commands table shows each command with its stages. A command the node refused locally reads "Refused at the node: local control", distinct from a refusal of the board or the host.
- Scrubbing back pauses following. **Back to live** resumes. The history is bounded to 2000 snapshots in the browser, and the bar says when it has trimmed.

A real board that stays connected but sends no admitted record for 30 seconds while the host is present is treated as lost: its devices are reported `link_down` and the gateway link is shown offline. While the host is away the board is told so and may stay silent; the 30 seconds start again when the host returns.

### 5. Interrupts: hold, approve, emergency stop

The **Interrupts** region is always visible while a connected session runs. The controls apply to all devices, a board or a device.

| Control | What the node does | What the host sees |
| --- | --- | --- |
| Hold | Refuses every command in scope | `DRECEIPT refused`, reason `local_control` |
| Approve each command | Holds each command in scope until the operator approves or denies it, or its deadline passes | `queued`, then the normal receipt after approval, `local_control` after a denial, `expired` after the deadline |
| Emergency stop | Sends the declared safe stop to every device in scope, then latches a hold on the scope | Later commands are refused `local_control` until the operator releases the scope |

Rules the node keeps:

- It only refuses or stops. It never originates an ordinary command, never changes a command it forwards, and never widens what the host may do.
- A hold outranks approval. A device's declared safe-stop command is never held or queued when it carries no arguments. A safe stop with arguments is held like any command (D-100).
- An approved command goes on unchanged and still passes the normal checks (freshness, boot, authority epoch, policy digest, interlocks). A hold set after the command was queued still refuses it.
- The emergency stop reports three lists: `stopped`, `skipped` (devices that declare no safe stop, or were never born; they are held but not stopped) and `unreachable` (a stop could not be ordered: the device died, the gateway is gone or the simulated node is gone). A device in `skipped` or `unreachable` may still be energised. The canvas says so and does not show a clean stop.
- A command the gateway already accepted cannot be recalled (see [limitations](../limitations.md#physical-devices)).
- A release names the scope it clears, exactly: a hold on `fan-01` is not cleared by releasing `all`, and the latch of a stop on a board is cleared by releasing that board.
- Controls are applied at the next barrier (the simulated world) or the next 250 ms tick (a relay), not instantly.
- At most 64 scopes are set, at most 64 commands wait for approval (each with at most 8 repeats), and a node has at most 4096 devices.

Every control and every decision is recorded in `interrupts.jsonl` and in the observer journal, so a reopened recording shows them in order. From the command line the control socket takes one JSON request per line and answers one line; for example, with a client that can talk to a Unix socket:

```bash
printf '%s\n' '{"op":"state"}' | nc -U -w 2 "$D/control.sock"
printf '%s\n' '{"op":"hold","scope":{"kind":"device","id":"fan-01"}}' | nc -U -w 2 "$D/control.sock"
printf '%s\n' '{"op":"release","scope":{"kind":"device","id":"fan-01"}}' | nc -U -w 2 "$D/control.sock"
printf '%s\n' '{"op":"approve","command_id":"<id>"}' | nc -U -w 2 "$D/control.sock"
printf '%s\n' '{"op":"deny","command_id":"<id>"}' | nc -U -w 2 "$D/control.sock"
printf '%s\n' '{"op":"estop","scope":{"kind":"all"}}' | nc -U -w 2 "$D/control.sock"
```

`nc -U -w 2` works with the macOS and OpenBSD netcat; GNU netcat has no `-U`. The alternative is `socat`: `printf '%s\n' '{"op":"state"}' | socat - UNIX-CONNECT:"$D/control.sock"`. The `state` reply lists in `devices` only the devices that have a non-default mode, so on a quiet session `devices` is `[]`. `approve` may carry the `content_digest` shown in the state's `pending` entry for the waiting command.

The request vocabulary is in [`node-control-v0.1`](../../contracts/node-control-v0.1.schema.json) and the [CLI reference](../reference/cli.md#plantwire-connect).

### 6. Stop and save

**Canvas.** **Stop and save** interrupts the session; the bridge waits up to 30 seconds for the run to close before it ends the process group. The run opens as an ordinary review labelled "Connected session". **Set up another run** keeps the policy list and clears the sockets and boards.

**Command line.** Interrupt `plantwire connect` (Ctrl-C or `SIGTERM`). It prints `{out, stopped, relay?}`, and the run folder is complete. **`--out` must be a new folder for every session.** A folder that already holds a run file is refused (`plantwire_failed`) and left as it was. Name each attempt (`m5-session-1`, `m5-session-2`) and keep the failed ones; the canvas picks a unique `c-<8 hex>` name itself.

```text
run.json  trace.jsonl  ledger.jsonl  world_state_history.jsonl  journal.jsonl  sealed/journal.jsonl
field-trace.jsonl   (when a gateway took part)
interrupts.jsonl    (when a control socket was given)
```

### 7. Replay

```bash
go run ./cmd/streamsim verify "$W/runs/c-lab1/run.json"
go run ./cmd/streamsim report "$W/runs/c-lab1"
```

`verify` replays the command log against a fresh world and checks every evidence file, including `field-trace.jsonl` and `interrupts.jsonl`, against the digests in `run.json`. The simulated part reproduces byte for byte, and the board's part is the trace. The commands the operator approved, and the emergency stops, are in the command log like any command, so they replay with the world. The rules of the interrupt log are replayed too; the session checks them when it ends.

To show a recording to a host again:

```bash
go run ./cmd/streamsim plantwire replay-trace --run "$W/runs/c-lab1" --host-socket "$D/host.sock"
```

A replay is not a board. It announces the node and every board as `simulated`, so a host set to `hardware` refuses it, it has no path to a gateway, and it refuses every command it is sent. It plays the gateway's side of a recording; the simulated world's side is reproduced by `verify`, not streamed (see [limitations](../limitations.md#physical-devices)).

### 8. Compare the twin with the board

**Canvas.** On a saved session, the Report tab has **Twin vs board**. Optionally type tolerances, then **Compare with twin**. The Inspect panel of a real board adds a **Twin prediction** chart: the board's compared readings solid, the twin's prediction dashed, and the stretches where they part ways shaded.

**Command line.**

```bash
go run ./cmd/streamsim plantwire twin-report "$W/runs/c-lab1" \
  --domain arduino-thermal --domains-dir "$W/domains" --kinds-dir "$W/kinds" \
  --tolerance zone-01.temperature=2
```

The twin is replayed in a fresh noise-free world with the commands the board accepted, and its prediction is compared with every good reading under a declared tolerance. Without `--tolerance` the default is 5% of the metric's `eu_range` (change it with `--span-fraction`). The verdict is `agrees`, `diverges` or `cannot_judge`. Read the caveats in [limitations](../limitations.md#physical-devices) before relying on `agrees`: the twin starts from its domain's initial state, a very wide tolerance agrees with anything, and a run with no applied command cannot be judged. The command refuses a run whose `journal.jsonl` is over 64 MiB and readings that span more than 7 days.

## The policy allow-list

Every `DCMD` carries a `policy_digest`, a label for the policy the host says it acts under. A node executes a command only when that digest is on its list, and refuses it `policy_not_allowed` otherwise. The digest is not a signature and not a secret; it is an authorisation gate on an owner-only socket, not authentication.

- **Simulated devices.** The list is the session plan's `policies`: 1 to 8 distinct digests of the form `sha256:` and 64 lower-case hexadecimal digits. Absent means none, and then every command to a simulated device is refused, except a safe stop. `plantwire connect` has no `--policy` option; `plantwire node --policy` and `plantwire relay --domain ... --policy` apply the same rule.
- **Real boards.** The gateway checks the digest against its own list, outside this repository. The simulator never adds to it. The gateway is the only place a real board can refuse a policy; the firmware's wire has no policy refusal code.
- **Both lists must name the host's digest.** If one side refuses `policy_not_allowed`, its list differs. The canvas commands table shows the digest the host used (shortened, with **Copy policy digest**). Adding it is a deliberate edit and a new session; nothing is added automatically.
- **The default.** The canvas can save the checked list as `settings/default.policies.json` in the workspace. It only pre-fills a new draft, visibly, and never changes a running session.
- **When the policy narrows,** the host's digest changes. Add the new digest on both sides, then remove the old one; leaving the old one listed keeps trusting the wider policy.

## What the simulator will not do

- **It only refuses or stops.** It never originates an ordinary command, never widens what the host may do, and is not a policy authority. Its one command of its own is the declared safe stop of an emergency stop.
- **It does not emulate a microcontroller.** The model is behavioural: no instruction-level emulation of a board or its firmware. The firmware stays the last line of defence; if the simulator stops, leases lapse and the board makes itself safe.
- **It does not speak serial or Bluetooth.** Only the gateway touches hardware.
- **It has no truth for hardware.** A `DRESULT` from a board is self-reported and may be wrong. A physical effect counts only when an independent instrument observes it. The host's own check of a reading the board itself reported is recorded evidence, not independent proof.
- **It does not score a consumer against a board's self-reports.** For a hardware device a consumer is measured on its response to what it observed, never on whether it was right about the world.
- **It never reaches hardware from a replay.** `run`, `replay`, `verify`, `scenario` and `plantwire node` have no gateway option.
- **It has no authentication.** One desktop, local, one user. The control socket and the bridge rely on owner-only directories and a pairing code. See [the security model](../architecture/security-model.md#physical-devices-and-the-control-socket).

## When something does not work

| Symptom | Likely cause | First check |
| --- | --- | --- |
| Gateway refused `inventory_mismatch` | A bound board, device or kind is not announced as the plan expects, or a kind digest differs | The `reason` in the relay document; `plantwire connect --check` shows the digest the plan expects |
| Gateway refused `contract_unsupported` | The gateway speaks another contract version or digest | The gateway must announce version 1.4.0 and the digest in `contracts/plant-wire/DIGEST` |
| Gateway refused `environment_mismatch` | The node or a board is not `hardware`, or the host's setting does not accept `mixed` | The `reason`; a gateway must announce `hardware` |
| Every command to a simulated device is `policy_not_allowed` | The plan lists no policy digest, or not the host's | The commands table row shows the host's digest |
| A real board's command is `policy_not_allowed` | The gateway's own list lacks the host's digest | The gateway's configuration |
| Commands read `local_control` | A hold or an emergency-stop latch is set | The Interrupts region; release the scope that was held or stopped (releasing `all` does not clear a hold on a device) |
| A device shows `stale`, `disconnected` or "Origin unknown" | The board is silent, offline, or never announced a board for it | Connection health; the observer `wire` records |
| `session_running` | One connected session at a time | Stop the running one |
| `control_gone`, `control_timeout`, `no_session`, `node_busy` | The session ended (`no_session`, `control_gone`); the bridge's 12 s wait ended (`control_timeout`); or the node's own 10 s answer was `node_busy`. In the last two the node may still apply the control afterwards | Read the node state again before repeating a control |
| `replay-trace` refuses a folder | The run records no `field_trace_digest`, or the trace differs from it | `report` lists the `field_trace` check |

## Next reads

- [Architecture of the physical-device support](../architecture/physical-devices.md)
- [CLI reference](../reference/cli.md)
- [Artifact reference](../reference/artifacts.md)
- [Limitations: physical devices](../limitations.md#physical-devices)
- [Security model](../architecture/security-model.md)
