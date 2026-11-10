# Predicting a failing conveyor drive before it strands the belt

## What

A closed-loop test of an AI operator on a simulated industrial belt conveyor. A 60 m conveyor is pulled by three
motor drives that share the load. One drive starts to fail. The operator reads the plant's sensors, works out what is
happening, and acts through the plant's own controls: it cuts or shapes the feed, runs the belt empty, and stops the
failing drive at a chosen moment, with a human approving every step that stops equipment.

Three systems work together:

- **Streams Simulator** runs the plant, keeps the true story sealed from the operator, and grades every run against
  what would have happened if nobody had acted.
- **Agentic Stream** turns the sensor stream into situations, applies the safety policy, and sends commands.
- **Tamoz** reasons over each situation with a language model (GLM 5.3 Flash in these runs) and recommends the step.

## Why

When one drive of a multi-drive conveyor weakens, its share of the load moves at once onto the others.

- If the drives have a large rating margin, the survivors carry the belt but age several times faster.
- If the margin is small, they overload and trip within about two minutes, and the belt stops full of material.
- Two drives cannot restart a loaded belt, so a failure that could have cost a planned stop costs hours of manual
  unloading.

The failing drive usually shows it first: vibration rises, its winding runs hotter than its peers at the same current.
The question is whether an operator can turn those early signs into the right action in time, without being fooled by
a feed surge or a dead sensor, and without stopping a healthy line.

## Data

All plant parameters are synthetic engineering estimates, not a customer's data.

| Plant | Value |
|---|---|
| Belt | 60 m, flat, 2 m/s, 150 t/h nominal feed (about 1.25 t on the belt) |
| Drives | 3 (motor, gearbox, pulley, VFD); torque load sharing; D-1 leads the speed |
| Plant A (large margin) | one drive out: survivors at 0.97× rating; safe limp feed 54 % |
| Plant B (small margin) | one drive out: survivors at 1.40× rating, inverter trip in about 2 min; safe limp feed 18 % |
| Sensors | per drive: current, winding temperature, vibration, pulley speed; belt speed; feed rate |
| Actions | feeder setpoint, stop a drive, stop the belt (stopping steps need human approval) |

| Story | What happens |
|---|---|
| S0 normal day | daily feed swings, no fault (must do nothing) |
| S1 bearing failure | D-3 loses lubrication; the trip comes 6-9 h later if nobody acts (plants A and B) |
| S2 sudden trip | D-3 trips under full load with no warning (plants A and B) |
| S3 gearbox seizure | D-3's gearbox loses oil and seizes; the belt stalls loaded |
| S4 feed surge | all drives load up together for a few minutes (must do nothing) |
| S5 frozen sensor | as S1, but D-3's vibration sensor freezes at a healthy value; S5b: it goes silent |
| S6 lead drive lost | D-1 trips; the plant either stops everything or hands the lead to D-2 |

Each story has 10 seeds, each drawing its own onset time, severity and sensor noise.

**How a run is graded.** Every run is replayed as it ran, with nobody acting (B0), and under four simple policies:

| Baseline | Policy |
|---|---|
| B0 | do nothing |
| B1 | vibration alarm: feed 0, then stop the drive |
| B2 | current threshold: feed 67 % |
| B3 | safe limp feed from the start |
| B4 | today's practice: on a drive trip, feed 67 % |

The cost J is in equivalent production hours: a stranded belt, protective trips, drive wear, tonnes not delivered and
unneeded actions each carry a weight. A negative difference means the run cost less than the baseline.

## Evidence

**Rules only, 10 seeds per story, 113 of 113 runs pass.** These runs use the operator's rules with a scripted reasoner,
no language model.

| Story | Result over 10 seeds | Cost vs doing nothing (h) | Detected |
|---|---|---|---|
| S1 bearing, plant A | beats all five baselines every seed | -1.0 .. -3.2 | 0.3-0.9 h after onset, correct cause |
| S1 bearing, plant B | beats all five baselines every seed | -5.6 .. -6.1 | 0.5-1.2 h after onset |
| S3 gearbox | beats all five every seed | -3.1 .. -3.5 | 0.5 h after onset |
| S5 frozen / silent sensor | beats all five; sensor fault reported first | -2.0 .. -4.2 | sensor in 15 min, bearing within hours |
| S6 lead drive handed over | beats all five every seed | -5.0 .. -5.5 | 10-13 s |
| S6 everything stops | nothing can prevent it; no harmful restart | 0 | 12-16 s |
| S0, S4 controls (40 runs) | no action, no false alarm | 0 | - |

**With the language model reasoning, three seeds per story, every story passes.**

| Story | Result | Cost vs doing nothing (h) | Model calls per run |
|---|---|---|---|
| S1 bearing, plant A | 3/3 | -2.4 .. -3.2 | 4 |
| S1 bearing, plant B | 3/3 | -5.6 .. -5.9 | 5 |
| S2 sudden trip, plant B | 3/3, feed cut within 14 s of the trip | -5.5 | 1 |
| S3 gearbox | 3/3 | -3.1 .. -3.4 | 3 |
| S5 frozen sensor / S5b silent sensor | 3/3 each | -2.2 .. -3.9 | 5-10 |
| S6 lead drive lost | 3/3 each variant | -5.1 .. -5.4 (handover), 0 (stop) | 1 |

**Predicting the trip.** On plant B the best move is to keep full feed and stop the failing drive just before it would
trip. A curve fitted to the drive's winding temperature predicts the trip 1.4-1.5 h ahead, within 6 minutes on every
seed. A straight-line fit warns only 1.0-1.2 h ahead. The operator stops the drive 1.34-1.37 h before its trip and beats
even the vibration alarm.

**What the language model needed.** Alone, the model held back when the cause was uncertain: it answered "unknown" and
only watched, 3 of 3 times on the bearing story. It succeeded once the loop around it named the symptom ("a drive
degrading against its peers"), let it take the steps that are right whatever the cause, and woke it again when the
evidence changed. The bearing story went from 0/3 to 3/3.

**Trust.** Every run replays bit for bit from its record. The true story stays sealed from the operator until grading.
Every claimed action is read back from the plant. A run that a simple baseline ties or beats is flagged as easy rather
than counted as a win.
