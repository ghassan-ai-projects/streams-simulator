# CC2640R2F experiment: session plans

Two session plans (`session-plan-v0.1`) for a CC2640R2F board that reports five one-byte sensors over Bluetooth Low Energy.
Both run the domain `cc2640r2f-ble-node` (`../cc2640r2f-ble-node.domain.json`) with the kind file beside it, both are paced to
the wall clock and both list the same placeholder policy digest. The board has no commands, so the policy is never exercised.

| Plan | Instances | What it is for |
| --- | --- | --- |
| `simulate.plan.json` | `node-01` simulated | Rehearsing the host without the board: `sensor_1_shift` raises sensor 1 by 40 for 120 s, 60 s after the session starts |
| `real.plan.json` | `node-01` real, on the gateway's board `cc2640r2f-01` | The experiment on the board |

`real.plan.json` adds `gateway_socket` and `real`, and has no `disturbances`. Every other member is the same.

The five sensors are placeholders (`sensor_1` to `sensor_5`, `integer` 0 to 255, unit `1`) until their meaning is known; naming
them changes the kind file and its digest everywhere. There is no board mapping: a reading is the byte itself. Add one
(`../../thermal-board/arduino-thermal.board.json` is the model) when a byte needs a scale into a physical unit.

## Run

From the repository root, after `make build`. The domains folder may hold only domain files:

```bash
B=bin/streams-simulator
E=examples/devices/cc2640r2f
mkdir -m 700 /tmp/ble-experiment && mkdir -m 700 /tmp/ble-experiment/domains
cp "$E/cc2640r2f-ble-node.domain.json" /tmp/ble-experiment/domains/
POLICY=sha256:$(printf 'b%.0s' $(seq 64))
```

Check a plan (starts nothing):

```bash
$B plantwire connect --plan "$E/ble-experiment/simulate.plan.json" --check \
  --kinds-dir "$E" --domains-dir /tmp/ble-experiment/domains
```

Simulated: start the reference host, then the session; stop it with Ctrl-C, then `verify` the run folder.

```bash
$B plantwire host --socket /tmp/ble-experiment/host.sock --environment simulated --policy "$POLICY"
$B plantwire connect --plan "$E/ble-experiment/simulate.plan.json" \
  --kinds-dir "$E" --domains-dir /tmp/ble-experiment/domains --out /tmp/ble-experiment/runs/sim-1
$B verify /tmp/ble-experiment/runs/sim-1/run.json --domains-dir /tmp/ble-experiment/domains
```

Real: the reference host with `--environment hardware`, then `connect` with `real.plan.json`, then the gateway's BLE
transport on `/tmp/ble-experiment/gateway.sock` with board `cc2640r2f-01` (see the gateway repository).
`--out` must be a new folder for every session.
