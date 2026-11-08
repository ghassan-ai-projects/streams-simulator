# M0 design: the conveyor as data

Status: as built in M0 rounds 1 to 3 (2026-11-08). It refines [PLAN.md](PLAN.md) section 3 with the owner decisions of
section 9.6 (O1-O14), a calibration run on an independent Python mirror of these equations (one-second Euler steps), and
what the component format allows (no state may depend on itself through the inputs of its kind, even through a lag).
Where this page and PLAN.md section 3 differ, this page wins; each difference is listed in section 7.

## 1. Parts

| Library id | Instance(s) | Role |
|---|---|---|
| `streamsim/bulk-feeder@1` | F-1 | Feed setpoint (`set_feed`), feed lag; feeds only while the belt runs; shift pattern and surge (negative class), chute build-up |
| `streamsim/belt-conveyor@1` | B-1 | Material on the belt, required force, load sharing, capacity, stall, run permit from downstream, restart margin, stranded boundary, tonnes delivered, stop permissive; `start_belt`, `stop_belt` |
| `streamsim/conveyor-drive@1` | D-1..D-n, n 2 to 6 (O7) | Motor, gearbox, pulley and inverter on the shared belt: load, current, motor relay and inverter overload (O6), winding temperature, aging, bearing and gear failures, seizure; `start_drive`, `stop_drive`, `reset_trip` |

`streamsim/conveyor@1` and `streamsim/motor-drive@1` are not changed (O3, D10). The parts use `parameters`, not
presets: library presets are reserved for small/medium/large.

New vocabulary: unit family `force` (base `N`; `kN`); port types `bulk_flow` (mass_flow), `pulley_force` (force),
`drag_force` (force), `force_capacity` (force), `run_state` (state).

Ports (consumer side `sample_s` 1 s where a loop closes; every integrated step is 1 s, so the sample is on the step grid):

| From | To | Type | Meaning |
|---|---|---|---|
| F-1 `feed_out` | B-1 `feed_in` | bulk_flow | feed onto the belt, kg/s |
| B-1 `moving_out` | F-1 `moving_in` (sampled) | run_state | belt commanded and turning: cmd · (v > 0.1 m/s) |
| B-1 `share_out` | each D `share_in` | pulley_force | force each running drive delivers, N |
| B-1 `stop_ok_out` | each D `stop_ok_in` | run_state | stopping one more drive is safe (section 3) |
| each D `enabled_out` | B-1 `drives_in` (sampled) | run_state | drive running and healthy (sum = n) |
| each D `drag_out` | B-1 `drag_in` (sampled) | drag_force | drive's own drag on the belt (sum) |
| each D `run_cap_out` | B-1 `run_cap_in` (sampled) | force_capacity | enabled · k_run · F_rated (sum) |
| each D `boost_cap_out` | B-1 `boost_cap_in` (sampled) | force_capacity | enabled · k_boost · F_rated (sum) |
| each D `ready_cap_out` | B-1 `ready_cap_in` (sampled) | force_capacity | available · k_boost · F_rated (sum): what a restart could use |
| each D `rated_out` | B-1 `rated_in` (sampled) | force_capacity | enabled · F_rated (sum) |

A consuming state keeps its own F0 baseline for the part composed alone: the value a 3-drive regime A plant delivers
at nominal (feed 41.6667 kg/s, n 3, drag 82.8 N, capacities 4140 N, rated 2760 N, share 594.5446 N, moving 1), except
the drive's `stop_ok`, which is 1 alone (the loaded plant gives 0) so that `stop_drive` can be proven on its own.

## 2. Feeder (F-1)

| State | Equation | Unit |
|---|---|---|
| feed_set | 100, assigned by `set_feed` (0-100) | % |
| supply | 1; carries the feeder's faults | ratio |
| moving | port `moving_in`; alone 1 | 0/1 |
| feed_target | 0.416667 · feed_set · moving · supply | kg/s |
| feed | lag(feed_target, τ 3 s) | kg/s |

Q_nom 41.6667 kg/s (150 t/h). Measurement point `feed` (FT, mass flow meter, 0-80 kg/s).

| Fault | Class | Form |
|---|---|---|
| `shift_pattern` | negative | supply × (1 − 0.3) for the first 12 h of each 24 h (intermittent): the S0 nominal-day swing |
| `feed_surge` | negative | supply × (1 + severity) for 300 s (pulse): severity 0.3-0.6 is a 1.3-1.6x surge |
| `chute_buildup` | fault | supply × (1 − 0.3 · min(0.005 · h, 1)) (ramp) |

## 3. Belt (B-1)

Constants: L 60 m, v_set 2 m/s, F_empty 1000 N, k_m 0.494 N/kg, a0 1630 N, a1 parameter `breakaway_n_per_kg`
(A 1.424 default, B 1.152 N/kg), boost time parameter `boost_s` (2 s).

| State | Equation | Unit |
|---|---|---|
| m | store: dm/dt = feed − flow `discharge` (m · v / L) | kg |
| out | m · v / L | kg/s |
| delivered | ∫ out dt / 1000, a totalizer from 100 000 t; KPI | t |
| f_run | F_empty + k_m · m + v · feed | N |
| f_break | a0 + a1 · m | N |
| slow | curve(v): 1 at 0, 0 at ≥ 0.2 m/s | 0..1 |
| f_req | f_run + slow · (f_break − f_run) | N |
| demand | f_req + drag | N |
| cap | run_cap + slow · (boost_cap − run_cap) (O8: the boost holds only while breaking loose) | N |
| short | demand − cap | N |
| go | run_cmd · permit | 0/1 |
| short_cmd | short · go: a shortfall while the belt is asked to run | N |
| stall_ok | trip: short_cmd > 0 for `boost_s`, latched, tripped 0, reset by `start_belt` (counter `start_count`) | 0/1 |
| excess | curve(short): max(0, short) | N |
| inv_n | curve(n): 0, 1, 1/2 … 1/6 at n = 0..6 | - |
| cmd | go · stall_ok | 0/1 |
| share | (demand − excess) · inv_n · cmd | N |
| has_drive | curve(n): 0 at 0, 1 at ≥ 1 | 0/1 |
| v_target | v_set · go · stall_ok · has_drive | m/s |
| v | regime on short_cmd (threshold 0.5 N): drives can pull → lag(v_target, τ 4 s); cannot → lag(0, τ 2 s) | m/s |
| turning | curve(v): 0 at ≤ 0.1, 1 at ≥ 0.11 | 0/1 |
| moving | cmd · turning (port `moving_out`) | 0/1 |
| permit | 1; fault `downstream_stop` (negative class, step −1) switches it off (O11) | 0/1 |
| restart_margin | ready_cap − f_break − drag | N |
| stranded | trip: stopped · loaded · trapped > 0.5 for 10 s, unlatched; stopped = v < 0.1, loaded = m > 375 kg (0.3 · m_ss), trapped = restart_margin < 0 (curve indicators); boundary, major | 0/1 |
| stop_ok | ind(boost_cap · after(n) − f_break − drag ≥ 0) · ind(rated · after(n) − demand ≥ 0), after(n) = curve(n): (n−1)/n | 0/1 |
| tail_speed | v · 60 / (π · 0.5 m) · tail_grip: tail pulley speed for the tachometer (ST, 0-100 rpm) | rpm |
| tail_grip | 1; fault `tail_slip` (ramp, −0.3 at 0.002/h): only the tail pulley reading falls | ratio |

m_ss = Q_nom · L / v_set = 1250 kg. At regime A nominal: share 594.54 N, restart margin +647 N, stop_ok 0 (two
drives could not restart the loaded belt).

## 4. Drive (D-k)

Constants: k_run 1.5, I_rated 4.4 A, τ_r 1000 s, τ_inv parameter `inverter_tau_s` (700 s: heavy duty 150 % for
60 s from rated load), trip at θ > 1.1025 (1.05²), T_amb 35 °C, τ_w 900 s, seized drag parameter `seized_drag_n`
(1830 N, O10; 0 = a drive that disengages). Enum parameter `margin` writes F_rated, k_boost, k_d and every value
derived from them, steady initial values and the `share` fallback included:

| margin | F_rated | k_boost | k_d | load at nominal | winding at nominal |
|---|---|---|---|---|---|
| `strong` (default, regime A) | 920 N | 1.5 | 0.55 | 0.646 | 82.4 °C |
| `weak` (regime B) | 607 N | 1.75 | 0.135 | 0.964 | 111.0 °C |

| State | Equation | Unit |
|---|---|---|
| load | share · enabled / F_rated | - |
| i2 | curve(load): 0.1225 + 0.8775 · load² (0.05 steps, 0-3), times enabled | - |
| current | curve(load): I_rated · √(0.1225 + 0.8775 · load²), times enabled | A |
| θ_r, θ_inv | lag(i2, τ_r), lag(i2, τ_inv) | - |
| relay_ok, inverter_ok | trip θ > 1.1025, latched, reset by `reset_trip` | 0/1 |
| winding | lag(T_amb + enabled · (24 + heat_b + 15 · g) + 56 · load², τ_w) | °C |
| overtemp_ok | trip winding > 150 °C for 10 s, latched, reset by `reset_trip` | 0/1 |
| aging, life_used | curve(winding): 2^((T + 10 − 130)/10), nodes every 5 K; ∫ aging / 3600 from 10 000 h in service; KPI | 1, h |
| gear_used | ∫ curve(load): load⁶ / 3600 from 10 000 h in service; KPI | h |
| lube | 0; fault `lube_loss` (ramp 1/h) | 0..1 |
| bearing_damage d | dd/dt = k_d · lube · load³ / 3600, clamped 0..1 | 0..1 |
| heat_b, vib_b | curve(d): 75 · d³ K; 9.6 · d⁴ mm/s | K, mm/s |
| wiggle | 0.25 · sin(2π t / 24 h) from the world start, the same on every drive | mm/s |
| vibration | 1.2 · enabled + 0.3 · load + (vib_b + wiggle) · enabled | mm/s |
| g | 0; fault `gear_oil_loss`: 2 · severity · min(0.25 · h, 1), held at most 1, so g rises at 0.5 · severity per hour and every severity seizes | 0..1 |
| free_ok | trip g > 0.95, latched; `reset_trip` cannot clear it while g stays above (seizure) | 0/1 |
| vfd_ok | 1; fault `vfd_fault` (step −1) | 0/1 |
| available | relay_ok · inverter_ok · overtemp_ok · vfd_ok · free_ok | 0/1 |
| enabled | run_cmd · available (port `enabled_out`, run status YS) | 0/1 |
| drag | 0.03 · F_rated · (2 − enabled) + 0.05 · F_rated · d² + 0.27 · F_rated · g + (1 − free_ok) · seized_drag | N |

Instruments per drive: current clamp (current, 0-15 A), RTD (winding, 0-200 °C), vibration sensor (vibration,
0-45 mm/s), run status (enabled). Belt: tachometer (tail pulley). Feeder: mass flow meter (feed).

`streamsim/vibration-sensor@1` is an IO-Link vibration transmitter in the style of an ifm VVB001: v-RMS 0-45 mm/s,
one reading a minute, σ 0.02 mm/s, resolution 0.01 mm/s. Faults: `dropout`, `loose_mount` (gain 0.7) and `frozen`
(repeats its last value with quality good: the S5 story). The wiggle makes a frozen reading drift away from the true
one (by 0.1 mm/s within about two hours), so the fault is observable; a peer comparison cancels it.
`streamsim/accelerometer@1` is unchanged.

Effectors (each with ack ok, 300 ± 50 ms):

| Effector | Part | Args | Requires | Effect |
|---|---|---|---|---|
| `set_feed` | feeder | `feeder_id`, `percent` 0-100 | - | assign feed_set |
| `start_belt` | belt | `belt_id` | run_cmd < 0.5 | run_cmd +1, start_count +1 (re-arms the stall trip) |
| `stop_belt` | belt | `belt_id` | run_cmd ≥ 0.5 | run_cmd −1 |
| `start_drive` | drive | `drive_id` | run_cmd < 0.5 | run_cmd +1 |
| `stop_drive` | drive | `drive_id` | run_cmd ≥ 0.5, stop_ok ≥ 0.5 | run_cmd −1 |
| `reset_trip` | drive | `drive_id` | θ_r < 0.9, winding < 120 °C | reset_count +1 |

After a stall, a restart is `stop_belt` then `start_belt`.

## 5. What the calibration gives

Python mirror, noiseless:

| Story | Regime A | Regime B |
|---|---|---|
| S1 do nothing: D-3 overtemperature trip after onset, severity 0.8 / 1.0 / 1.25 | 9.4 / 7.7 / 6.3 h (k_d 0.50; 0.52 centres it on 6-9 h) | 8.9 / 7.3 / 6.0 h |
| S1 after the trip | survivors at about 1.0x, no trip; the downstream stop strands the loaded belt | survivors' inverters trip about 100 s later (cascade), belt stranded |
| S2 sudden trip of D-3 at full feed | survivors carry it; the downstream stop strands the belt | inverter cascade at about +119 s |
| S3 gear oil loss, severity 1 | seizure at +1.9 h, stall 2 s later | (not run) |

Round 1's smoke runs of the composed plant (`streamsim run`, noiseless state history) agreed: both regimes start steady and
lint clean; in A a `vfd_fault` on D-3 leaves the survivors at load 0.98 and a later downstream stop strands the belt
10 s after it stops; in B the survivors' inverters trip about 120 s after D-3; three drives restart a stopped loaded
belt in about 20 s, two stall it 2 s after `start_belt`.

The inverter, not the motor relay, sets the clock in B: about 120 s against the relay's 170 s (O6). The expected
numbers in the tests come from the oracles of section 6, not from this table (O14).

Library contract (each part composed alone): lint, steady start, replay, effectors, observability and the
trivial-baseline audit pass. Audit best balanced accuracy: `chute_buildup` 0.78, `tail_slip` 0.80, `lube_loss` 0.86,
`gear_oil_loss` 0.87 (non-trivial); `vfd_fault` 0.997 (trivial, as a hard trip should be). The drive with
`margin: weak`, alone, also starts steady, lints clean, keeps its faults observable and passes the audit (`lube_loss`
0.86, `gear_oil_loss` 0.84); `reset_trip` is refused at its nominal θ_r 0.938 (> 0.9), so its effector check is not run.

What the oracle tests compute (round 2, noiseless, the composed plants):

| Check | Plant A | Plant B |
|---|---|---|
| Survivors after D-3's `vfd_fault` | load 0.9844, I² 0.9728 settles below 1.1025: no trip in 50 min; winding 82.39 → 113.26 °C | load 1.4610, I² 1.9956, θ0 0.9384: inverters trip 119.1 s after D-3 (closed form, survivors see the share 1 s late), 121 s in the model |
| Relay alone (`inverter_tau_s` 3600) | - | 169.7 s closed form, 171 s in the model |
| Breakaway margin, loaded belt | +648 N with three drives, −759 N with two | +63 N with three, −1017 N with two |
| Stop permissive | loaded, three drives: refused; empty, three: allowed; empty, two: refused | same |
| Feed step to 70 % | m 1250 → 875 kg, worst deviation from the closed form 4.9 kg (one step of feed lag) | - |
| Dry bearing, severity 0.8 | damage 0.0035 / 0.0140 / 0.0561 at 0.25 / 0.5 / 1 h, within 0.1 % of the closed form | - |
| Gearbox without oil, severity 4 | seizure at 1770 s; with 1830 N seized drag the belt is 1117 N short and stalls 3 s later; with 0 N two drives carry 2048 N of 2760 N and keep running | - |
| 2-drive belt at 30 % feed | 632.7 N each; one survivor carries 1293.1 N (2.04 ×) | - |

## 6. Tests and scenario files

- Plants: `examples/compositions/belt-conveyor-drives-a.composition.json` (`margin: strong`, `breakaway_n_per_kg`
  1.424) and `-b` (`margin: weak`, 1.152): F-1, B-1, D-1..D-3 and every port of section 1; shipped as built-in examples
  (embedded copies in `internal/component/internal/files/examples/`) and composed to
  `examples/belt-conveyor-drives-{a,b}.domain.json`; both lint clean and are in the lint list.
- Oracle tests in `test/contract/belt_conveyor_*_test.go`, each expected number computed in the test from the design's
  constants and closed forms: compose byte for byte; both plants start steady; material after a feed step;
  share against n; inverter and relay trip times in B and no trip in A; winding step response; bearing damage; the
  restart table in both plants (3 drives start a loaded belt, 2 cannot: stall within `boost_s` + 1 s, stranded 10 s
  after the third drive is lost on the stopped belt; 2 start an empty one); a seizure with and without seized drag
  (O10); the stop permissive truth table; a 2-drive belt composed inside the test (O7); a frozen vibration reading
  drifting from the truth; the weak drive's library checks.
- Every lever changes survivor load or what the belt carries (silent_no_effect), each against an uncalled world:
  `set_feed`, `stop_drive`, `stop_belt`, `start_belt`.
- Tests that simulate an hour or more (survivors in A for 50 min, the dry bearing, the seizure, the frozen sensor, the
  weak drive's library checks) skip under `-short`.
- Stories (O12): `examples/belt-conveyor/stories.json` holds S0-A, S1-A/B, S2-A/B, S3-A, S4-A/B and S5-A (S6 is M4):
  plant, horizon, disturbances, ten seed rows drawn once from a fixed generator, and the hand-written reference policy
  (README in that folder). Generated from it and checked byte for byte: `scenarios/<story>.do-nothing|reference.scenario.json`
  (seed row 1, `scenario-v0.1`, noisy, `native-jsonl`) and `sessions/<story>.session-plan.json` (seed row 1, wall pace
  for S2 and S3, placeholder host socket, no policies; checked by the plant wire's own plan check).
- Story checks in `test/contract/belt_stories_*_test.go`: the table and the files (CI tier, including one committed
  scenario run through `run.RunScenario` up to a minute after its first command; the whole file in the long tier),
  and every seed row of every story with and without its policy on noiseless worlds executing the generated commands
  (long tier).

| Story | Horizon | Draws (PLAN.md section 2.3 unless noted) | Reference policy |
|---|---|---|---|
| S0-A | 24 h | `shift_pattern` from 0 | none |
| S1-A, S1-B | 16 h | lube loss at U[0.5, 2] h, severity U[0.8, 1.25]; downstream stop at onset + U[9.5, 12] h for 30 min | onset + 4 h: feed to the safe limp feed; + 5 min: feed 0; + 3 min: `stop_drive` D-3; + 10 s: feed back to limp |
| S2-A, S2-B | 3 h | `vfd_fault` at U[10, 40] min; downstream stop at trip + U[10, 40] min for 30 min | A: trip + 30 s: limp feed. B: trip + 10 s: feed 0; trip + 250 s: limp feed |
| S3-A | 4.5 h | gear oil loss at U[0.5, 1.5] h, severity U[0.8, 1.25] | onset + 1 h: feed 0; + 3 min: `stop_belt` |
| S4-A, S4-B | 6 h | 1-3 `feed_surge` pulses at U[0.5, 5.5] h, at least 20 min apart, severity U[0.3, 0.6] | none |
| S5-A | 16 h | S1-A with onset U[1, 2] h (so the freeze starts after 0), plus `vibration__frozen` on D-3 from onset − U[0.5, 1] h | as S1-A |

Safe limp feed (two drives restart the belt they leave loaded, and stay at or under rating): A 57.3 % and 104 % →
54 %; B 29.3 % and 20.1 % → 18 % (rating binds in B, not the restart).

What the long tier found (every seed row):
- S1 do nothing: D-3 trips on overtemperature 6.96-8.17 h after onset in S1-A, 7.14-8.86 h in S1-B, 6.99-8.54 h in
  S5-A. Severities above 1 bunch at about 7.0 h in A: the grease deficit `lube` is held at 1, so a severity above 1
  only reaches it sooner. A: survivors carry on and the belt is
  stranded at the downstream stop. B: survivors' inverters trip 100-110 s after D-3 (hotter than in S2, D-3's bearing
  drag having loaded them), then stranded.
- S2-B do nothing: cascade 121 s after the trip in every row (closed form 119.1 s; section 5). Feed 0 saves the
  survivors if it comes within 104 s of the trip (S2-B's reference reacts at 10 s); B4 (feed 67 % one second after
  the trip) still trips them at 201 s.
- S3 do nothing: seizure at the closed-form time (1.61-2.16 h after onset), stall 3 s later with 1250 kg on the belt.
- S4-B: one 300 s pulse up to 1.67x does not trip plant B; the drawn severities stay at or below 0.6.
- S0, S4: no protective trip, no stall, never stranded. Every reference policy: never stranded, no survivor
  protective trip; S1 and S5 stop D-3 by command 4 h 8 min after onset, before its do-nothing trip; S3 leaves the
  belt empty and stopped when D-3 seizes.
- The long tier simulates about 1530 plant hours (91 800 minutes) in 8.5 min of wall time on this machine
  (40 CPU minutes, the variants in parallel); it skips under `-short`.

## 7. Differences from PLAN.md section 3

| # | Change | Why |
|---|---|---|
| M1 | Regime A F_rated 920 N (was 880) | With each drive's own drag (3 % of rating) and a damaged bearing's drag, 880 N put A's survivors over the relay threshold: A would not be the "survivors cope" regime |
| M2 | Bearing drag at full damage 5 % of rating (was open) | The failure sketch's range is 2-10 %; 10 % tipped A's survivors |
| M3 | Inverter overload as a second lagged I² trip (O6) | Same form as the motor relay; τ 700 s fits "150 % for 60 s from rated load" |
| M4 | Boost only while the belt is slower than 0.2 m/s; stall after `boost_s` (O8) | Breaking loose takes under a second; the boost time is the hold time of the stall trip |
| M5 | Drive force capped at the belt's capacity; drives stop pushing once the belt stalls or the permit is off | Without it a stalled or interlocked belt loaded the drives at breakaway force and tripped their inverters |
| M6 | A downstream stop is not a stall | The stall trip watches short · go |
| M7 | k_d per regime | Load³ wear made B fail three times faster than A; the story's 6-9 h window is a property of the story, not of the regime |
| M8 | Belt speed is a regime lag selected by the shortfall (τ 4 s while the drives can pull, decay τ 2 s while they cannot); no `free` indicator; one τ for start and stop (was 7 s up, 2 s down) | v may not depend on itself through the inputs of its kind, even through a lag; the regime selector is not an input |
| M9 | Material on the belt is a store drained by a flow of conductance 1/L scaled by v | An integrator of feed − m · v / L reads m through its inputs: a refused cycle |
| M10 | Regimes A/B are the drive's enum parameter `margin` and the belt's parameter `breakaway_n_per_kg`, not presets; `inverter_tau_s` and `seized_drag_n` are parameters (O6, O10) | Library presets must be small/medium/large |
| M11 | `set_running` split into `start_drive` and `stop_drive`; the belt and drive start and stop as ±1 steps guarded by `requires` | A `requires` applies to every call of an effector, and an effect cannot assign a literal |
| M12 | The feeder stops on the belt's command, not only its speed (`moving` = cmd · (v > 0.1)) | With speed alone a downstream stop piled about 490 kg more on the belt during the stop and three healthy drives could no longer restart it |
| M13 | S0's swing is the feeder's negative-class `shift_pattern` fault (70 % for 12 h of each 24 h), not a `swing_pct` parameter and a third composition | A fault needs no extra composition and is also the benign control the audit compares `chute_buildup` with |
| M14 | Alone, the drive's `stop_ok` fallback is 1; the connected loaded plant gives 0 | The library proves every effector on the part alone |
| M15 | `delivered` starts at 100 000 t | A totalizer; from 0 it moves 25 t in ten minutes and fails the steady-start check |
| M16 | `chute_buildup` ramps at 0.005/h and `tail_slip` at 0.002/h | Slow enough not to be caught by a one-line detector in 8 h, fast enough to be observable within 24 h |
| M17 | The drive's vibration has a daily wiggle and its own instrument, `vibration-sensor@1`, which carries the silent `frozen` fault; the accelerometer does not | A frozen reading of a steady signal is never observable to the truth oracle, and the accelerometer serves five other parts with steady vibration |
| M18 | `life_used` and `gear_used` start at 10 000 h in service | From 0, a weak drive's counters move more than the steady-start check allows in ten minutes |
| M19 | The 2-drive belt (O7) is composed inside its test, not shipped | Two shipped plants are enough; the variant only proves the sharing |
| M20 | Regime A k_d 0.55 (was 0.52) | In the Go model 0.52 put severity 0.8 at 9.09 h, outside S1's 6-9 h window; 0.55 gives 8.6 h at 0.8 and 7.0 h at 1.25 (k_d is the calibration M7 provides for this window) |
| M21 | `gear_oil_loss` severity sets the speed of the oil loss, not its final level (g = 2 · severity · 0.25/h, held at 1) | With g = severity · min(0.5 · h, 1) a severity below 0.95 never seized, so a third of S3's U[0.8, 1.25] rows had no story; severity 1 still seizes at 1.9 h |
| M22 | `feed_surge` adds severity × the feed (was 0.5 × severity) | S4's severities U[0.3, 0.6] then are the 1.3-1.6x surges of PLAN.md section 2.3 |
