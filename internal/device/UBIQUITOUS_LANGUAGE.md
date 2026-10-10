# Ubiquitous language — device

A wire-faithful emulator of the serial device at the far end of the
Agentic Stream serial-effector boundary. It is a test instrument: the effect a
command has on the plant is the ground truth, and an acknowledgement is never
proof of that effect.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Device | The emulator: boot identity, lease, dedup ledger, safe state | `Device` | `device_id`, `boot_id` |
| Record | One of the four wire messages: command, receipt, result, state | `EncodeRecord`, `DecodeRecord` | `message_type` |
| Command | A request to operate a target, bound to a boot id, freshness window and idempotency key | `ApplyCommand` | `command` |
| Receipt | The device's acknowledgement of a command: accepted or rejected with a code | `Outcome.Receipt` | `receipt`, `reject_code` |
| Result | The terminal status of an accepted command | `Outcome.Result` | `result` |
| State | The device's self-reported identity and current output | `State` | `state` |
| Capability catalog | The data-defined set of targets, operations, bounds and safe stops the device admits | `Capabilities` | `*.capabilities.json` |
| Admission | The ordered checks a command passes: boot, freshness, target, operation, bounds | `admit` | `reject_code` |
| Idempotency key | The key under which a repeated command returns its first outcome | dedup ledger | `idempotency_key` |
| Plant | The physical process the device actuates; the ground truth | `Plant` | – |
| Safe stop | The declared safe state a target is driven to on lease expiry or reboot | `SafeStopper` | `safe_stop` |
| Protocol fault | A deterministic device-level fault by accepted ordinal: ack lost, duplicate, disconnect, stuck | `FaultInjection`, `Faults` | `--fault` |
| Wire fault | A deterministic transport fault by outbound frame index: drop, duplicate, swap | `WireFaults` | – |
| Gateway link | The newline-delimited stream a host speaks to the device, including the `query_state` control line | `uds` edge | unix socket |
| Contract | The vendored device wire-protocol schemas and conformance fixtures | `device/contract` | `contract/` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `TargetNames`, `Advance`, `SetFaultSchedule` (device) | removed or domain-layer only | no production caller |
| `ServeConn`, `EncodeRecord`, `DecodeRecord` (exported) | module test seams | used only by the module's own tests |
