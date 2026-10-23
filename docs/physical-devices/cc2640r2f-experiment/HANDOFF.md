# Handoff: the CC2640R2F experiment (2026-10-21)

Read [PLAN.md](PLAN.md) first. This file says where the work stands and what the owner has to do next.

## 1. State

| Repo | Branch | What is there |
|---|---|---|
| `streams-simulator` | `CC2640R2F-wireless-microcontroller` | **B1 done** (commit `9154af97`): `examples/devices/cc2640r2f/` has the kind `ble_sensor_node` (`sensor_1` to `sensor_5`, integer 0 to 255, unit `1`), its twin domain `cc2640r2f-ble-node`, `ble-experiment/simulate.plan.json`, `real.plan.json` and a README. `kind validate`, `domain validate`, `domain lint` and `plantwire connect --check` pass; a 25 s dress rehearsal against the reference host wrote 60 readings and `verify` matched; `make quick` green. No gate changed |
| `agent-research-lab` (gateway) | `CC2640R2F-wireless-microcontroller` | **B2 done, not hardware-verified**: `tools/ble_plantwire_gateway.py` (the gateway), `plantwire_ble_link.py` (events, and `BleakLink`: bleak in its own thread, scan, connect, subscribe, reconnect with backoff), `plantwire_ble_board.py` (the board as the node's `BoardPort`: a new boot id per connection, `DDEATH link_down` on a lost link, `stale` after 6 s of silence, a notification of the wrong length quarantined), `plantwire_ble_profile.py` (how to find the board and read one notification), `ble_probe.py` (scan, list GATT, print notifications), `kits/cc2640r2f/` (the kind file and `cc2640r2f.ble-profile.json`), `RUNBOOK-BLE-GATEWAY.md`, `tools/requirements-ble.txt` (`bleak==0.22.3`) |
| `agentic-stream`, `tamoz` | not touched | B4 is configuration only and waits for the owner's answers |

Tests (gateway): 12 profile tests and 9 board tests need no hardware; 6 end-to-end tests run the gateway on a fake link against
the real streamsim binary (`STREAMSIM_BINARY=<path to bin/streams-simulator>` when the checkout's binary is older than its HEAD):
readings reach the host with zero quarantines, an unplug is `link_down` and a replug a rebirth on a new boot, `kind capture`
reports the kind `same`, the gateway ends with `NDEATH shutdown`. Two mutations (a constant boot id, no length check) are
caught. bleak 0.22.3 installs on this Mac's Python 3.9 and the API names the link uses exist.

## 2. What is a guess (until the board is probed)

- The five sensors' names and units: the kind has placeholders, so naming them changes the kind file and its digest in both
  repositories (the gateway's copy is `kits/cc2640r2f/ble_sensor_node.kind.json`, byte for byte the simulator's).
- The GATT layout: the UUIDs (`0000fff0` and `0000fff1` family), the advertised name `CC2640R2F`, one 5-byte notification.
- A new connection is a new boot (BLE has no boot counter). If the firmware exposes one, read it instead.
- No board mapping: a reading is the byte itself. Add one when a byte needs a scale into a physical unit.

## 3. What the owner does next

1. `RUNBOOK-BLE-GATEWAY.md` step 1: run `ble_probe.py`, write down the name, the UUIDs, the notification layout and rate.
2. Send those to the next session (or edit `cc2640r2f.ble-profile.json` and the kind file) and run step 2, the chain with
   `real.plan.json` and the reference host.
3. Then B3 (readings in the canvas, reconnect) and B4 (Agentic Stream site file and event schemas, outside its repository).

## 4. Rules

Each repository's `AGENTS.md`. Nothing is pushed or merged; merge to `main` only when the owner says so. The serial gateway
(`arduino_plantwire_gateway.py`) was not changed; the BLE gateway repeats its dial and send loop (about 100 lines) instead of
sharing it. Sharing is a later refactor with the serial gateway's tests as the net.
