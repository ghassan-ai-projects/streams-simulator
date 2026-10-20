# Brief: the Arduino gateway's plant wire side

- **For:** the gateway work in `agent-research-lab` (`real-world-sensor/assessment/tools/arduino_gateway.py` and its successor), M2.
- **From:** the streamsim physical-devices goal (R13). Written against streamsim at the end of R13, plant wire contract 1.3.0.
- **Not in this brief:** device-wire-v1, the firmware, the serial protocol. They do not change.
- **Read first:** `contracts/plant-wire/SPEC.md` (whole), `contracts/plant-wire/fixtures/README.md`, and the kind files in `examples/devices/thermal-board/` (`thermal_sensor`, `indicator`, `fan`). `docs/design/PHYSICAL_DEVICES.md` sections 1 and 2 give the reasons.

## What you build

A plant wire **node** in front of the Arduino. The gateway dials a Unix socket, announces the board and its devices, reports readings, and executes the host's commands through device-wire-v1. The same gateway works in two shapes without change:

```
direct (production):  Arduino -- serial -- gateway --> plant wire --> Agentic Stream
lab (relay):          Arduino -- serial -- gateway --> plant wire --> streamsim --> plant wire --> Agentic Stream
```

In the lab shape the socket you dial belongs to streamsim (`plantwire relay` or `plantwire connect`, the "gateway socket"). streamsim is the gateway's host on that link and passes your records up and the real host's commands down, with their content unchanged. You do not need to know which shape you are in, with the exceptions listed under "Things that differ in the lab" below.

## 1. Connect and announce

1. Dial the socket as a node. Refuse to dial a socket that is not owner-only (the socket and its directory owned by you, no group or other access). Do not dial TCP; v1 is Unix sockets only.
2. Send `NBIRTH` as the first record, with `protocol_version` 1, `protocol_min` and `protocol_max`, `contract_version` **`1.3.0`**, `contract_digest` equal to the content of `contracts/plant-wire/DIGEST` (vendor the folder byte for byte), `environment` **`hardware`**, `bd_seq`, `prior_last_seq`, `capability_digest`, `kinds`, `boards` and `node_metrics`.
   - streamsim accepts **only** the contract version and digest it was built with. A different version or digest is refused `contract_unsupported`. (SPEC section 6 speaks of a minimum; the implementation requires equality, recorded as D-113 in `governance/DEFERRED.md`.)
   - Every board in `boards` is `{board, environment: "hardware", members}`; `members` are the device ids on the board. One board per physical controller (the Mega is one board). A board that says `simulated` or `mixed` is refused `environment_mismatch`, and so is a node that is not `hardware`.
3. **Announce the same kind files.** `kinds` holds one entry per kind, equal to the kind file plus `kind_digest`. Compute `kind_digest` as `sha256:` and the SHA-256 of the RFC 8785 canonical JSON of the kind file's content (the entry without `kind_digest`). `streamsim kind validate <file>` prints the digest it expects for each file; your digests must equal them. Announce exactly the kinds the plan binds:
   - `thermal_sensor` (`temperature` Cel 0 to 50, `humidity` % 20 to 90, no commands),
   - `indicator` (`brightness_permille` 0 to 1000, `pattern` code 0 to 2, writable; commands `set_led` and `stop_led`, the latter `safe_stop`),
   - `fan` (`duty_permille` 0 to 600, writable; commands `set_pwm_lease` with `duty_permille` 0 to 600 and `lease_ms` 5000 to 10000, and `stop_fan`, `safe_stop`).
   The gateway never invents metrics or commands and never edits a kind file's content to suit the hardware. The field mapping (which pin drives the fan) stays in your configuration. If you change a kind file, you change its digest everywhere.
4. The relay recomputes every `kind_digest` from the entry you sent. A digest that does not hold, or a kind announced twice (also under another case), is refused `inventory_mismatch`. A session that binds an instance to your device (see `session-plan-v0.1`) is checked further at the handshake: the bound board must be announced, must hold the device, and the device's kind must be announced with the digest of the kind file the plan names. Failures are `board_missing`, `device_missing`, `kind_missing`, `kind_mismatch`, and the reason arrives in the refusal's `STATE`.
5. After `NBIRTH`, send one `DBIRTH` per device: `device`, `kind`, `board`, `boot_id`, `metadata_version`, `birth_time_ns`, and the current value of every metric (`null` with quality `warming_up` before the first sample). `board` must be the board whose `members` hold the device and `kind` must be a kind your `NBIRTH` announced; otherwise the relay drops the birth. Use the ids you want to see in the host (`zone-01`, `led-01`, `fan-01`). Device and board ids must not equal ids in the simulated world of a mixed session, or the whole gateway is refused `inventory_mismatch`; a plan that binds an instance to your device makes the world withhold that instance, which is the intended way to put a twin and a board of one model in one session.
6. Wait for the host's `STATE` before assuming a session. If the host refuses (`STATE {online: false, refusal}`), the link closes. Do not reconnect without an operator's action.

## 2. Boards, boots and links

- `boot_id` belongs to the board. It changes on every reboot. A reboot is a `DDEATH` with reason `reboot` for every device on the board, then each `DBIRTH` with the new `boot_id` (SPEC section 7).
- **Unplugging the board** (serial lost): send `DDEATH` with reason `link_down` for each of the board's devices. When the board is back, send `DBIRTH` for each device. Acceptance check M2: unplugging shows `link_down` and a rebirth.
- A failed sensor on a live board is a quality code in `DDATA` (SPEC section 10), not a `DDEATH`. A non-`good` value never satisfies a control precondition.
- **Heartbeat to the relay.** If the relay admits your session but receives **no admitted record from you for 30 seconds**, it treats the gateway as lost: it reports every device `DDEATH link_down` to the host, closes your link and records the gateway offline. Send readings at least every 30 s, or other valid records. A stray, malformed, repeated or forged record does not reset this timer. The timer is suspended while the relay has told you its host is offline (before the first host and during a host loss), when SPEC sections 7 and 8 have you buffer and stay silent; it starts afresh when the returning host's first `STATE` reaches you, followed by `NCMD rebirth`.
- **Handshake.** A connection that does not send `NBIRTH` within 10 seconds is dropped. The relay serves one gateway at a time; a second connection waits until the first is gone.
- **Reads.** The relay writes with a 2 second deadline on every link. A gateway that does not read its socket for 2 seconds while the relay writes is treated as stalled: it is dropped and redialled. Read continuously and do not block the reader on serial I/O.

## 3. Data

- `DDATA`: one or more metrics `{alias, value, quality, event_time_ns, time_source, uncertainty_ms?}`. Aliases are those of your `DBIRTH`. `event_time_ns` is **wall time**, UTC nanoseconds, with `uncertainty_ms` covering your clock skew (SPEC section 8; a hardware node runs in wall time). Send no NaN or infinity. A `DDATA` has at most 500 metrics; a record is at most 64 KiB (`NBIRTH` and `DBIRTH` at most 1 MiB).
- Quality codes are those of SPEC section 10 (`good`, `warming_up`, `suspect`, `out_of_range`, `stale`, `disconnected`, `manual_override`, `out_of_service`). The canvas shows a reading that is not `good` as its state, never as a value. A value outside the metric's declared `eu_range`, or of the wrong type, is shown as `invalid`.
- **Staleness** in the host's live view is computed from the kind's `scan_class` when it parses as a duration (for example `2s`; the reading turns stale after three intervals) and otherwise after 30 s. The shipped kind files declare no `scan_class`, so a DHT11 that updates every 2 s is stale after 30 s of silence. Adding `scan_class` to a kind file changes its `kind_digest`.
- Store-and-forward is yours: data buffered while the host was away is sent with `historical: true`. The relay does not buffer. A record that arrives while no host session exists is dropped and counted; this is by design and deferred (D-70).
- A DHT11 reads once about every 2 s with whole-degree resolution. The twin report compares with a tolerance of 5 percent of the metric's range by default (2.5 degC for `temperature`); do not report faster than the sensor measures.

## 4. Commands

Implement the checks of SPEC section 9, in this order, at receipt and again at execution for a queued command: schema, device, command and argument bounds; deadline and `not_before_ns`; `expected_boot_id`; **`authority_epoch`**; **`policy_digest`**; host alive; idempotency; device preconditions and interlocks. Answer every `DCMD` with at most one `queued` and then one terminal `DRECEIPT` (`accepted` or `refused` with the reason), and every accepted command with exactly one `DRESULT` (`succeeded`, `failed`, `unknown`; self-reported).

- Map to device-wire-v1 as SPEC section 13 says (`expires_after_ms` from `deadline_ns`, `not_before_mono_us` for `not_before_ns`, the same names for `expected_boot_id`, `idempotency_key`, `policy_digest`).
- **Epoch.** The epoch floor is the highest `authority_epoch` any accepted `STATE` announced; a `STATE` with a lower epoch is ignored, and you persist the floor. A `DCMD` with another epoch is refused `stale_epoch`. This check exists only in the gateway (SPEC section 13).
- **Policy.** Keep a list of allowed policy digests in your own configuration and refuse a `DCMD` whose `policy_digest` is not on it with `DRECEIPT refused, reason policy_not_allowed`, in the order above. This is the **only** place a real board can refuse a policy: device-wire-v1 has no policy reject code (`device-receipt-v1.json`, `reject_code`), and the firmware receives `policy_digest` unchanged and does not refuse on it. Do not add a policy check to the firmware for this brief; it would need a device-wire change owned by Agentic Stream.
  - The list holds several digests (up to what you like; streamsim's own lists hold at most 8) so a narrowing can be rolled out: when the host's policy narrows, its digest changes; add the new digest, then remove the old.
  - Log the list in effect (the digests, not secrets) when the gateway starts.
  - The digest is a label the host copies into each command. It is not a signature and not a secret. The host's digest must be the same one the operator puts into the streamsim session plan's `policies` for the simulated devices. Ask Agentic Stream where an operator reads it (its brief asks them to expose it); until then the first refused command in the canvas commands table shows the digest the host used.
- **Safe stops** are checked only for schema, device, command and boot. They skip the deadline, epoch, policy and precondition checks, are never queued and never refused for idempotency. A command is a safe stop only if **your kind file declares `safe_stop: true`** for it (`stop_led`, `stop_fan`). Refuse to start if a safe-stop command has device preconditions or an interlock.
- **Idempotency.** Key on `idempotency_key`; the same key with the same content replays the stored receipt and result; the same key with other content is refused `duplicate`. Do **not** check that the key is the digest of the content: the node-originated stops below use another derivation.

### Commands the relay may send that the host never issued

With a control socket, an operator can order an **emergency stop**. For a device in scope the relay then sends the device's declared safe-stop command to you itself, as a `DCMD`:

| Member | Value |
| --- | --- |
| `command_id` | `node-stop-<N>` (N counts up in the session) |
| `command` | the device's declared safe-stop command only (`stop_fan`, `stop_led`) |
| `args` | `{}` (empty) |
| `expected_boot_id` | the last boot id you announced for the device |
| `deadline_ns` | the relay's wall-clock UTC now plus 30 s |
| `not_before_ns` | `0` |
| `authority_epoch` | the host's last epoch, kept when the host is lost (may be `0` if no host ever announced one) |
| `policy_digest` | `sha256:` and 64 zeros |
| `fallback_on_lease_loss` | `safe_off` |
| `idempotency_key` | a digest that includes the stop's id (not the content digest of SPEC section 9) |

Execute it like any safe stop (it skips the epoch and policy checks), answer it with `DRECEIPT` and `DRESULT` as usual, and let the stop take effect on the device. The relay consumes those answers; the host never sees a receipt or result for a stop it did not issue. A device that has rebooted since refuses the stop `wrong_boot`, which is safe and is reported to the operator. A stop for a device with no `DBIRTH` is not sent. A command whose id starts with `node-stop-` from the host is refused by the relay as `malformed`, so you will not see a forged one.

Because the stop uses wall-clock time for `deadline_ns`, **your gateway clock must be right to within seconds** (UTC).

### Commands you cannot be asked to cancel

The relay can hold, approve and refuse what the host sends **before** it reaches you. A command it has already forwarded is yours: if you queue it for its `not_before_ns` or have it in flight, a later hold or emergency stop cannot recall it (D-94). If you can cancel a queued command on request, say so; the relay would then be taught to ask. Until then, keep `not_before_ns` waits short, and make leases the way to bound a command (`set_pwm_lease` requires `lease_ms` 5000 to 10000).

## 5. Rebirth, sequences and host loss

- `NCMD rebirth` (from the host or the relay): send `NBIRTH` with the next `bd_seq`, then every `DBIRTH`, in the same session and without resetting sequences. The relay asks for one when a host session starts after you were born, and once per gap in your `seq`. It does not rate-limit a flapping link (D-74).
- `seq` runs per direction from 0 by 1 and is assigned **at transmission**. A repeated or missing `seq` makes the relay drop the record (and ask for a rebirth after a gap). `session_id` is yours for the connection.
- **Host loss.** When the relay loses the real host it tells you with `STATE {online: false}` (host id `relay`, authority epoch 0, no refusal) and keeps your session. Treat it as SPEC section 8 says: stop executing commands, flush the queue, let leases lapse into their fallback, buffer data. Do **not** lower your epoch floor: a `STATE` with a lower epoch is ignored. When the host returns, the relay forwards its `STATE` and asks you for a rebirth.
- **If streamsim stops,** your link ends. The firmware's leases and safe state are the last line of defence; nothing in streamsim stands in for them.
- `NDEATH` with `shutdown` or `connection_lost` is yours to send on a clean exit.

## 6. Things that differ in the lab shape

- In the lab the real host's `STATE`, `authority_epoch`, `DCMD` safety members and heartbeat reach you unchanged. You see the real host's epoch.
- A hardware gateway sends no `NSYNC`: you run in wall time and need no barrier. The relay maps barriers only if you send one.
- The relay drops, and counts, records that fail their schema or size limit, repeat a `seq`, belong to another session, are of a type that does not travel that way (a gateway `DCMD`, `STATE`, `NCMD`, `DQUERY` or `HSYNC`), name a device you did not announce, or are a `DRECEIPT` or `DRESULT` for a command it did not route to you. These are visible in the run's observer journal as `wire` records with direction `gateway`.
- Every record you send is journalled and the admitted ones are written to `field-trace.jsonl`, so a session can be replayed to a host without you. A replay never reaches a gateway.

## Acceptance checks (M2)

Run each against `plantwire relay` with the reference host (`plantwire host --environment hardware`), then against Agentic Stream. The detailed commands are in `bench-checks-m2-m5.md`.

1. The kind digests you announce equal the digests `streamsim kind validate examples/devices/thermal-board/*.kind.json` prints.
2. `kind capture <relay run>/journal.jsonl --kinds-dir <kinds>` reports `same` for every kind and no issues.
3. The reference host's report (with an inventory of your devices) has `quarantined: []`; Agentic Stream shows zero quarantines.
4. Unplugging the board gives `DDEATH link_down` for each device, and plugging it back gives a rebirth (`DBIRTH` with a new `boot_id`).
5. A `DCMD` with a wrong `authority_epoch` is refused `stale_epoch`; a `DCMD` with a digest not on your list is refused `policy_not_allowed`; a `DCMD` with the right digest executes; a safe stop with the wrong epoch and an unlisted digest still executes.
6. A `node-stop-1` `DCMD` as in the table above executes `stop_fan` and is answered with a receipt and a result.
7. After a stalled read, a host loss and a reconnect, the epoch floor is unchanged and the session recovers by rebirth.
8. Your log shows, for each command, its `command_id` and the UTC time at which it was handed to the firmware, and the policy list in effect at start. (The bench instrument brief uses these times.)
9. Conformance: your wire passes `contracts/plant-wire/fixtures` (valid, invalid, must-ignore, injection) and the exchanges `session-start`, `kind-digest-mismatch`, `stale-epoch`, `safe-stop-from-stale-epoch`, `policy-not-allowed`, `gap-rebirth`, `host-loss`, `reboot-reconcile`, `queued-across-reboot`, `duplicate-command`, `duplicate-different-content`, `expired-command`, `lying-result-*` and `environment-*`.

## Send back

The gateway's announced `NBIRTH` (one JSON line), its startup log with the policy list, the log lines of checks 4 to 8, the relay run folder of the unplug test (`journal.jsonl` and `field-trace.jsonl`), and any place the gateway cannot meet this brief, with the reason.
