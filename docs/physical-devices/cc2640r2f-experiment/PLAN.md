# The CC2640R2F experiment: Bluetooth link and five one-byte sensors (plan, 2026-10-21)

Status: draft for the owner. Nothing in it has started. It follows the shape of the [fan experiment](../fan-experiment/PLAN.md)
and reuses its decisions (plant wire contract 1.4.0, kind files, board mapping, session plans, the gateway's node core).

## 1. Goal

Stage 1 (this plan): prove that a CC2640R2F board is read over Bluetooth Low Energy by the gateway, that its five one-byte
sensors reach the simulator through the plant wire, and that the board can be added to the simulator as data.

```
Simulate:  streamsim (twin of the five sensors)
Real:      streamsim <-- plant wire <-- gateway <-- BLE <-- CC2640R2F (5 sensors, 1 byte each)
```

Out of scope here: the firmware (done), actuation, Agentic Stream decisions, Tamoz.

## 2. What exists and what is new

| Piece | State |
|---|---|
| Plant wire contract 1.4.0, kind files, board mapping (`BOARD` schema), session plans, paced sessions | Exist; reused |
| Gateway node core (`plantwire_node`, `BoardPort`) | Exists; reused |
| Gateway board link | Serial only (`plantwire_serial.py`, device-wire-v1 lines). **New: a BLE transport behind the same port** |
| Simulator | Expected to need data only: a kind, a board mapping, a twin domain, two plans. A needed code change is recorded as a simulator defect |
| Branches | `streams-simulator` and `agent-research-lab` are on `CC2640R2F-wireless-microcontroller`, cut from the merged fan work. `agentic-stream` and `tamoz` not checked |

The five sensors are read-only: no commands, safe stops, leases or approvals.

## 3. Stages

### B0. Bench spike (gateway repo only)

A standalone script with `bleak` (CoreBluetooth on macOS; not installed, system Python is 3.9, so a venv): scan, connect,
list the GATT services and characteristics, read or subscribe until the five bytes print. Then unplug and replug the board.

Exit: a log with five values per update and the behaviour of a disconnect. macOS needs the Bluetooth permission for the
terminal app; it shows a UUID, not a MAC address.

### B1. Simulator data (`streams-simulator`)

- A kind file with `commands: []` holding the five one-byte metrics (like `thermal_sensor`).
- A board mapping `cc2640r2f.board.json`; scale only if the raw byte needs it.
- A twin domain for the five sensors.
- `simulate.plan.json` and `real.plan.json`, differing only in the gateway socket and the `real` list.

Exit: `kind validate` and `plantwire connect --check` pass; a dress rehearsal against the reference host with a retained run
folder and `verify` matching.

### B2. Gateway BLE transport (`agent-research-lab`)

- BLE transport: scan, connect, subscribe, parse the five bytes into readings.
- Link lost: `DDEATH link_down`; reconnect: rebirth, with backoff.
- A fake BLE peripheral for tests (like `fake_thermal_firmware.py`), so tests need no hardware.
- A `--ble` option beside `--serial`.

Exit: tests green; then the real node against streamsim's relay, as in G2.

### B3. Real rung with the stand-in host

Run `real.plan.json` with the board: readings in the canvas, steady values, reconnect works. A twin-versus-real report is
meaningful only once the sensors' meaning is known.

### B4. Agentic Stream ingest (configuration only)

A site file mapping readings to events and an event-schema folder, both outside the repository. Exit: readings arrive with
zero quarantines. Tamoz is untouched until it is decided what it should decide from these sensors.

## 4. Open questions for the owner

1. What are the five sensors (names, units, signed or unsigned byte)?
2. GATT layout: one characteristic with five bytes or five; notify or read; rate; advertised name. Where is the firmware
   source (not found under the home folder to depth 4)?
3. Does the gateway run on this Mac or on another machine (a Raspberry Pi)? It changes how Bluetooth is used.
4. One device with five metrics (recommended: one board, one link) or five devices?
5. Does the firmware expose a boot counter? BLE has none, so otherwise a new connection counts as a new boot.

## 5. Rules

Each repository's `AGENTS.md` in full. In streamsim: devices and domains are data, no consumer knowledge in the binary,
files at most 300 lines, functions at most 15 lines, never loosen a gate, minimal comments. One commit per improvement,
reviewed first; merge to `main` only when the owner says so.
