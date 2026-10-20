# Brief: Agentic Stream's plant wire host

- **For:** the Agentic Stream team (the plant wire host), M2 to M5.
- **From:** the streamsim physical-devices goal (R13). Written against streamsim at the end of R13, plant wire contract 1.3.0.
- **Read first:** `contracts/plant-wire/SPEC.md`; `docs/design/PHYSICAL_DEVICES.md` sections 2 and 4.2; `docs/physical-devices/POLICY-ALLOW-LIST.md`.
- **Streamsim needs no operator interface from you.** It handles the operator's controls itself, at the node, and reaches you only through ordinary plant wire records. Your changes are the ones below.

## What the host sees

In the lab, streamsim stands between the Arduino gateway and you. You see one node. Its birth announces:

- `environment` `mixed` when simulated boards run beside the real one, or `hardware` for a relay with no simulated world. Your setting must accept it (a host set to `mixed` accepts all three environments; one set to `hardware` accepts only `hardware`). Each board in `NBIRTH.boards` carries its own `environment`; the **origin** of a device is the environment of the board its `DBIRTH` names, and nothing else. A node whose environment contradicts its boards is refused by the reference host `environment_mismatch`, with the broken rule in the report; do the same.
- `kinds` with a `kind_digest` each. Recompute it from the entry and refuse the session `inventory_mismatch` when it does not hold, or when a kind is announced twice. The digest of a real board's kind equals the digest of the kind file the twin implements; that equality is what lets one commissioning cover the board and its twin.
- A simulated board and a real board of the same model announce the same kind with the same digest. A session that binds a plan instance to the real device withholds that instance from the simulated side, so no device id appears twice.

Commission read-only first (SPEC section 5): births are claims, not grants, and the real board's commands stay denied in your inventory until M5.

## Handling refusals from the node

Streamsim adds a local veto, the way a panel switch does. It may only refuse or stop; it never originates an ordinary command and never widens what you may do. Handle each of these like any refusal from a device:

| You see | Meaning | Handle |
| --- | --- | --- |
| `DRECEIPT refused`, reason **`local_control`**, `boot_id` `"none"` | An operator at the node holds the device, board or plant, denied the command, or an emergency stop latched a hold. | A terminal refusal. Do not count the command done. Do not retry in a loop: the hold lasts until the operator releases it, and nothing in the wire says when. Surface it to the planner as "refused locally" and let it wait or re-plan. |
| `DRECEIPT queued` (then, later, one terminal `DRECEIPT`) | The operator has asked to approve each command. The command waits at the node. | Wait. The terminal receipt supersedes `queued`: `accepted` (then `DRESULT`) or `refused` after approval, `refused local_control` after a denial, `expired` after the deadline. The queued receipt carries the device's boot id when known. |
| `DRECEIPT refused`, reason **`expired`** | The `deadline_ns` passed while the command waited, no later than the first barrier or tick after it. | Terminal. A command that needs a human to approve it needs a `deadline_ns` long enough for one; the reference host uses 120 s. |
| `DRECEIPT refused`, reason `session_ending` | The session ended while the command waited. | Terminal. The node may send nothing when the host session ends (D-97): count a command still queued at a session's end as refused `session_ending`, as SPEC section 9 says. |
| `DRECEIPT refused`, reason `duplicate` | A `command_id` is already waiting with other content, or the `idempotency_key` was reused with other content. | Use a new id. |
| `DRECEIPT refused`, reason `not_ready` | The gateway is not connected, the simulated node is gone, or the node's queue is full. | Retry after the device is reborn. |
| `DRECEIPT refused`, reason **`policy_not_allowed`** | See "Policy digest". | Operator action; not retryable by you. |
| `DRECEIPT refused`, reason `stale_epoch`, `wrong_boot`, `unknown_device` and the others in SPEC section 9 | As in the contract. `unknown_device` is also the relay's answer for a device on no board. | As today. |

**Node-originated safe stops are not visible to you as commands.** On an operator's emergency stop streamsim sends the declared safe-stop command to the device itself (`DCMD` with an id `node-stop-<N>`, empty `args`) and consumes the receipt and result: you see no `DCMD`, `DRECEIPT` or `DRESULT` for it, and a host command whose id starts with `node-stop-` is refused `malformed`. What you see afterwards is the device's own data and state (a simulated device also sends a `DSTATE`; a relay sends none for a real device) and `local_control` refusals of your later commands. Do not assume that "no `DCMD` from me" means "device state unchanged". A device can be stopped under you, and a stopped scope stays latched until the operator releases it. A device's declared safe-stop command is never held or queued when it carries no arguments (so the host's own safe stops get through a hold); a safe stop with arguments is held like any command (D-100).

An approved command is handed on **unchanged**, as if it had just arrived, and passes every normal check at the device or gateway. An approval is not a grant; it only lifts the node's own veto.

## Policy digest

Every `DCMD` carries `policy_digest`. The simulator, for simulated devices, and the gateway, for the real board, execute a command only when the digest is on their allow-lists, and refuse it `policy_not_allowed` otherwise. The check is the node's, not yours; the digest is a label, not a signature, and the lists are the operator's.

Please:

1. **Expose the digest you put in `DCMD`** where an operator can read it: a setting, a log line at start, or an API. Say where in your reply. Until then, the operator reads it from the first refused command in the streamsim canvas commands table, which shows the host's digest (shortened, with a copy button), or from the observer journal's `wire` records with direction `host`.
2. **Confirm that narrowing the policy gives a new digest.** The operator then adds the new digest to both lists and removes the old.
3. Do not rely on streamsim to hold or widen a policy. Its list for simulated devices is the `policies` of the session plan (1 to 8 digests), set by the operator, never automatically.

## `HSYNC` and `command_evidence` (contract 1.3.0)

Verification of an effect is yours (SPEC section 9: receipt is not result is not observation is not verification). Streamsim records your decision and shows it as the **effect** stage of the command, separate from the receipt and the board's own report.

- `HSYNC` may carry optional `command_evidence`, at most 128 entries per barrier: `{command_id, device, boot_id, metric, expected, tolerance, outcome, observed_value?}` with `outcome` `observed` (requires `observed_value`) or `not_observed`, and `tolerance` at least 0. Without it a command's effect is unknown.
- The relay consumes `command_evidence`: it forwards `HSYNC` to the gateway or the simulated node without it. It is host-internal data.
- The effect shows `observed` only when **all** hold: the entry names the same command, device and boot as the command you sent; the command's receipt was `accepted` and its `DRESULT` `succeeded` on the device's current boot; the `metric` is one the device's kind announced; `tolerance` is finite and no wider than the metric's `eu_range`; the observed value lies within the tolerance of the expectation; and it agrees with the latest `good` reading the relay holds for that metric. Otherwise the entry is recorded as unverified, or leaves a mark if the device went offline or rebooted.
- Send only a decision based on a good, non-historical sample from the same boot that is newer than the device's successful result. A timeout is `not_observed`, not proof of failure. The reference host does exactly this (`plantwire host --script` with `verify`).
- **Limit you should know.** An `HSYNC` carries evidence only if it answers a barrier. A simulated world raises one every slice; a **hardware gateway runs in wall time and sends no `NSYNC`**, so in a hardware-only relay no `HSYNC` is routed and the effect stays `unknown` (D-112). In a mixed session the simulated world's barrier carries it. Evidence is also per device and metric: it can confirm a metric of the commanded device only, so the reading of an independent instrument on another device cannot feed this stage (D-111). Independent confirmation of a physical effect is the owner's bench check (`bench-instrument.md`); a host's check of a reading the board itself reported is recorded evidence, not independent proof.

## Authority epoch and host loss

- Announce your `authority_epoch` in `STATE`. The gateway persists the highest epoch floor it has seen and refuses `stale_epoch` any command with another epoch. A host that restarts with a lower epoch than its floor must raise it before commanding.
- Streamsim forwards your `STATE` to the gateway unchanged. When it loses you (end of stream, a gap in your `seq`, or three missed heartbeats once you have sent a heartbeat above zero) it tells the gateway `STATE {online: false}` with epoch 0 and keeps the gateway's session, so the firmware's lease logic applies. The gateway ignores a lower epoch; its floor does not drop.
- Streamsim keeps your last epoch for the node's own emergency stops. An epoch 0 announced by a host that returned without one never replaces it.
- When you return, streamsim sends the gateway your `STATE` and asks it for a rebirth, which reaches you as a new `NBIRTH` and `DBIRTH`s in your new session. Commands that waited for approval when you were lost are refused `session_ending`, never kept for you (D-83).
- Heartbeats: send them. The relay arms its three-missed-heartbeats watchdog only after you have sent a `STATE` with a heartbeat above zero (D-69).
- Make your `STATE` envelope, `seq` and `session_id` valid. A record that fails its schema is dropped by the relay and counted, with no receipt and no reason in the evidence (D-89): an empty `policy_digest` or a missing safety member in a `DCMD` shows up as a command that simply never gets an answer.
- On a stop, you may see `NDEATH` with `shutdown` (the relay ending) or only the socket closing (a simulated node session). Treat the close as `connection_lost`.

## What streamsim guarantees to you

- Content passes unchanged. For real devices, every member of a record except the envelope is as the gateway sent it, and every member of your `DCMD`, including all safety members, reaches the gateway as you sent it. The relay owns the envelope only: `session_id`, `seq`, the birth's `bd_seq`, `protocol_max`.
- It never originates, answers, delays or alters an ordinary command. Its own records are: a rebirth request to the gateway, an offline `STATE` to the gateway on host loss, `DDEATH link_down` for each device when the gateway is lost, `NDEATH` when it ends, refusals (`not_ready`, `unknown_device`, `local_control`, `expired`, `duplicate`, `session_ending`) and `queued` receipts for the operator's hold and approval, and, only on an emergency stop, the declared safe stop of a device.
- It never perturbs, delays or fault-injects a real device's records or commands. A simulated device's records are the simulated world's and carry that world's faults.
- A side answers only its own commands: a gateway's receipt or result for a command routed to the simulated side (and the reverse), or for a command never routed, is dropped and counted as forged.
- A recording can be replayed to you: `plantwire replay-trace` announces the node and **every board `simulated`**, so a host set to `hardware` refuses it and no consumer can take it for hardware. It refuses every command (`not_ready`) and never reaches a gateway.
- Bounds you can rely on: a record is at most 64 KiB (1 MiB for births); the latest 128 commands are shown; the relay's links have a 2 s write deadline, so a host that does not read for 2 s is treated as lost and redialled; a gateway silent for 30 s while a host is attached is reported `link_down` (the timer is suspended while the relay has no host).

## Acceptance checks

1. A session with a real board, simulated boards and an unchanged workflow: Agentic Stream accepts the `mixed` birth, commissions the three thermal kinds with the digests of `examples/devices/thermal-board/*.kind.json`, and shows zero quarantines.
2. A command held by the operator returns `refused local_control`; the planner treats it as a refusal and does not loop.
3. With approve-each on, a command returns `queued`, then `accepted` and a result after approval, `local_control` after a denial, and `expired` after its deadline.
4. After an emergency stop, the planner is told the device is stopped by the device's own data and by `local_control` refusals, not by a command it issued.
5. The digest in your `DCMD` is shown to an operator, and the same digest on the session plan's list and the gateway's list makes both a simulated and a real command read "Accepted".
6. After the host restarts with the same or a higher epoch, commands are accepted; with a lower one they are refused `stale_epoch`.
7. Where it applies (a mixed session), an `observed` and a `not_observed` `command_evidence` appear as the effect stage of the right command, and a command with no evidence stays unknown.

## Send back

Where the digest is exposed, whether narrowing gives a new digest, which of the acceptance checks pass, a host log of one held, one approved, one denied and one expired command, and any place Agentic Stream cannot meet this brief.
