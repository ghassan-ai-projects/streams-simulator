# Brief: the independent effect instrument

- **For:** the owner, building the bench for the thermal LED/fan experiment (M5).
- **From:** the streamsim physical-devices goal (R13). Updated in fan S4c (2026-10-20): semantic kinds on the wire (D7), paced sessions (D2).
- **Why it exists:** a command is not done because the board says so. SPEC section 9: receipt is not result is not observation is not verification. A `DRESULT` is self-reported and may lie. The effect of a command on the physical world counts only when something that does not depend on the board's own report observes it.

## Why the host's observation is not enough

The reference host and Agentic Stream confirm an effect by reading a metric of the commanded device after it reported success, for example `level_percent` on the fan after `set_cooling` (the gateway reports the firmware's duty through the board mapping). That metric is the board's own report of its own state. If the firmware reports the commanded duty without driving the fan (a wiring fault, an unplugged supply, a stuck driver), the host sees a matching value and records the effect as observed. The live screen shows it as the host's decision; it is **recorded evidence, not independent proof** (PHYSICAL_DEVICES section 3, SPEC section 9). Only an instrument whose sensing, power, timekeeping and logging path shares nothing with the board and the gateway can say the fan really turned.

## What the instrument must be

1. **Independent sensing.** Its own sensor, its own ADC or microcontroller, its own power where the measurement allows. Not read through the Arduino, its firmware or the gateway. Not a second metric of the same board.
2. **Independent clock.** Timestamps in UTC from a host disciplined by NTP (or by a GPS/PPS source), recorded with the offset to the reference measured before and after each run.
3. **Raw, append-only log.** One file per run, written as it measures, hashed (SHA-256) when the run ends, never edited.
4. **Calibrated and controlled.** A baseline with nothing commanded, a positive control (the actuator driven by hand through a separate path) and a negative control (the board commanded and the actuator physically disabled).

## What to measure (thermal LED/fan bench)

The bench is an ELEGOO Mega, a DHT11 on D2, an LED, and a fan driven through an L293D. Kinds and commands on the plant wire: `thermal_sensor` (`temperature` Cel 0 to 50, `humidity` % 20 to 90), `indicator` (`state` `off`, `watch` or `alert`; `show_state`, `stop_led`), `fan` (`level_percent` 0 to 60; `set_cooling` with `level_percent` 0 to 60; `stop_fan`). The board mapping `examples/devices/thermal-board/arduino-thermal.board.json` turns them into the firmware's raw commands: `set_cooling` is `set_pwm_lease` with `duty_permille` = 10 × `level_percent` and `lease_ms` 5000; `show_state` is `set_led` with the brightness and pattern of that state.

| Effect to confirm | Suggested instrument | Reading | Notes |
| --- | --- | --- | --- |
| Fan runs at about the commanded duty | A tachometer on the fan (optical or hall) logged by a separate microcontroller or USB logger | Rotational speed against time | Fast (tens of hertz) and direct. It answers "did the fan turn" in a second. |
| Fan power reaches the fan | A current sensor in the fan supply line (for example a high-side shunt monitor) on a separate logger, or a scope/logic probe on the L293D output | Current or PWM duty at the load | Detects a driver that is not switching even when the firmware reports success. |
| The zone cools | A second temperature probe next to the DHT11 (a thermocouple with its own amplifier, or a digital probe on a separate logger) | Temperature against time | Slow: the twin's time constant is 180 s, and a DHT11 reads about once every 2 s at whole-degree resolution. Use it for the twin report's expectation, not for the instant effect. |
| LED lit at the commanded brightness | A light sensor (photodiode or lux sensor) facing the LED, or a camera with frame timestamps | Illuminance against time | The `pattern` (off, slow blink, solid) shows as a waveform. |
| Safe stops (`stop_fan`, `stop_led`) | The same instruments | The fan stops, the LED goes dark | The most important check for an emergency stop. |

Sampling: at least 20 Hz for tachometer, current and light; 1 Hz for temperature.

Log format (suggested), one CSV per run: `utc_ns,channel,value,unit,instrument_id`, with a header block recording the instrument ids, calibration points, the reference clock offset before and after, and the SHA-256 of the file added when the run ends.

## Timestamps and clock alignment with the run

- The board's readings carry `event_time_ns` in UTC wall nanoseconds, stamped by the gateway, with `uncertainty_ms`. They are in `field-trace.jsonl`.
- The host's commands are issued in the run's own clock. A session paced to the wall clock (`pace: "wall"`, which the canvas always sets) starts the world at the session's wall time and runs one world second per wall second, so its world time is UTC up to the barrier's lag; an unpaced session starts at the domain's start (2026-01-01T00:00:00Z by default) and runs faster than wall time. Either way, do not align the instrument to the command's `at_ns` or to the world clock: use the gateway's log below.
- Align with wall time on the gateway side: the gateway brief asks the gateway to log each command's `command_id` and the UTC time at which it handed the command to the firmware. Use that log as the time of the command. For a hardware device, the relay's `interrupts.jsonl` also records a wall-clock time (`node_ns`, clock `wall`) for each command it judged when the session ran with a control socket.
- Measure the clock offset between the instrument host, the gateway host and the machine running streamsim (for example with `chronyc tracking` or `ntpq -p` on each, before and after), and allow it in the comparison window. State the offset in the evidence.
- A `set_cooling` holds for the board's lease (5 s in the mapping; the firmware's `set_pwm_lease` accepts 5 to 10 s) and then falls back; the expected effect window is the lease plus the fan's spin-up and spin-down.

## How the readings are used

Streamsim does not read the instrument's file. In this goal the instrument is the owner's evidence beside the run, not an input of it. For each command in the run's commands table (or in the observer journal), the owner records:

| Column | Source |
| --- | --- |
| `command_id`, device, command, arguments | the run (`journal.jsonl` `wire` records with direction `host`, or the canvas commands table) |
| UTC time handed to the firmware | the gateway log |
| Receipt and board report | the run (`live` records, or the commands table) |
| Expected physical effect | from the command's arguments and the kind |
| Measured effect and its time window | the instrument file |
| Pass or fail | the owner, against a stated threshold |

Pass means the instrument shows the expected effect inside the window after the command's UTC time; fail means it does not, whatever the board reported. The live screen's effect stage is the host's own decision; the bench result is this table.

Two limits of the design, recorded so nobody expects more (D-111, D-112): the live effect stage accepts host evidence only for a metric of the commanded device, so an instrument modelled as another device cannot confirm a command on the live screen; and a hardware gateway raises no barrier, so a hardware-only session carries no `command_evidence` at all. If the owner wants the instrument's verdict to appear on the live screen, that is a design decision (the instrument's reading as a metric of the commanded device's kind, which changes the kind file and its digest; or an evidence member that names another device, which changes the contract).

## Acceptance

1. Baseline: with nothing commanded for 5 minutes, the fan channel stays at zero, the LED channel at its dark level, and the temperature drifts only as the room does. Record the noise.
2. Positive control: the fan driven by hand at the bench (not through the board) shows the expected tachometer and current response.
3. A lying result: with the fan supply disconnected, `set_cooling` returns `accepted` and `succeeded` on the live screen while the instrument shows no fan response. The bench result is "fail"; the live screen's effect reads unknown or, if the reference host's `verify` of `level_percent` is used, "observed" (the board's own report), which demonstrates why the host's observation is not independent.
4. A held command (operator hold, `local_control`): no physical effect is seen, and the row says "refused at the node".
5. An emergency stop during a fan run: the instrument shows the fan stop within the gateway's stop latency plus the fan's spin-down, and the gateway log shows a `node-stop-N` command for `stop_fan`.
6. For each commanded effect in the M5 run, the table above has a measured effect, a time window and a pass or fail.
7. The raw file is hashed and stored with the run folder; the clock offsets are stated.

## Send back

The instrument design (sensors, wiring, sampling, clock source), calibration points, the baseline and control recordings, the raw files with hashes, the completed table for the M5 run, the clock offsets, and photos of the rig. Do not include credentials or customer data.
