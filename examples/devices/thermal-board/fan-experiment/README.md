# Fan experiment: session plans

Two session plans (`session-plan-v0.1`) for the thermal LED/fan experiment on the Arduino thermal bench. Both run the
domain `arduino-thermal` (`../thermal-board.domain.json`) with the kind files and the board mapping beside it, both are
paced to the wall clock (`"pace": "wall"`), and both list the same policy digest.

| Plan | Instances | What it is for |
| --- | --- | --- |
| `simulate.plan.json` | `zone-01`, `led-01` and `fan-01` all simulated | Rehearsing the host without the board: a heat source is added inside the zone (`zone_heat_load` on `zone-01`) 120 s after the session starts, for 900 s, so the zone settles about 8 degC warmer and cooling it is the right answer |
| `real.plan.json` | `zone-01`, `led-01` and `fan-01` all real, on the gateway's board `board-01` | The experiment on the board |

**What differs, exactly:** `real.plan.json` adds `gateway_socket` (`/tmp/fan-experiment/gateway.sock`) and `real` (the three
bindings to the devices of `board-01`), and has no `disturbances` (a fault can strike only a simulated instance). Every other
member is the same: `schema_version`, `domain`, `seed`, `host_socket`, `pace` and `policies`.

**Replace the policy digest.** `policies` holds a placeholder, `sha256:` followed by 64 `b`. Before a session with Agentic
Stream, replace it in both files with the policy digest Agentic Stream logs for its commands (a command refused
`policy_not_allowed` in the canvas commands table also shows it). Leave it when the host is the reference host started with
the same `--policy`. The gateway checks the digest of a real board's commands against its own list.

## Run

From the repository root, after `make build`. The domains folder may hold only domain files, so copy the domain into one:

```bash
B=bin/streams-simulator
E=examples/devices/thermal-board
mkdir -m 700 -p /tmp/fan-experiment/domains
cp "$E/thermal-board.domain.json" /tmp/fan-experiment/domains/arduino-thermal.domain.json
POLICY=sha256:$(printf 'b%.0s' $(seq 64))
```

Check a plan (starts nothing):

```bash
$B plantwire connect --plan "$E/fan-experiment/simulate.plan.json" --check \
  --kinds-dir "$E" --domains-dir /tmp/fan-experiment/domains
```

**Simulated.** Start the host on `/tmp/fan-experiment/host.sock` (Agentic Stream, or the reference host
`$B plantwire host --socket /tmp/fan-experiment/host.sock --environment simulated --policy "$POLICY"`), then:

```bash
$B plantwire connect --plan "$E/fan-experiment/simulate.plan.json" \
  --kinds-dir "$E" --domains-dir /tmp/fan-experiment/domains --out runs/fan-sim-1
```

**Real.** Start the host as above (the reference host with `--environment hardware`), then the node, then the gateway on
`/tmp/fan-experiment/gateway.sock` with board `board-01`:

```bash
$B plantwire connect --plan "$E/fan-experiment/real.plan.json" \
  --kinds-dir "$E" --domains-dir /tmp/fan-experiment/domains --out runs/fan-real-1
```

`--out` must be a new folder for every session. Stop a session with an interrupt (Ctrl-C) or from the canvas, which builds
the same plan from its Connect panel. The tests in `internal/plantwire/internal/app/fan_experiment_test.go` check both plans,
rehearse the simulated one against the reference host (the heat load raises the zone; `show_state alert` and
`set_cooling {level_percent: 45}` are accepted and read back as `state alert` and `level_percent 45`) and drive the real one
with the stand-in gateway.

## The bench-calibrated twin

`arduino-thermal.bench.domain.json` is `../thermal-board.domain.json` with the zone's rest state fitted to three real runs (2026-10-25):
the zone rests at ambient (22 degC, no board heat at the DHT11: `heat_load` 0, `zone_temperature` 22) and the humidity intercept is 97
(53 percent at 22 degC). The cooling time constant stays 180 s: the real decays gave 126 to 354 s depending on how the heat source was
removed. Copy it as `arduino-thermal.domain.json` into the domains folder to compare a real session with this twin. A person warming the
sensor by hand is a disturbance the twin does not model, so the comparison still diverges while the sensor is warmed. The shipped
`../thermal-board.domain.json` is unchanged (tests pin its values).

