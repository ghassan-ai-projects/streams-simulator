# Plan: multi-drive belt conveyor, one drive degrading, the others overloaded

Status: plan only (2026-11-07); scope confirmed 2026-11-08 (section 7: streams-simulator and agentic-stream, tamoz
only if needed, the Retrofit Designer later). No code, no runs. Inputs: [README.md](README.md) (problem, coordinator review),
`research/conveyor-mechanics.md` (CM), `research/degradation-and-prognostics.md` (DP), `repos/streams-simulator.md` (RS),
`repos/agentic-stream.md` (RA), `repos/tamoz.md` (RT), the three AGENTS.md files and the fan experiment
(`docs/physical-devices/fan-experiment/`). Claims the plan leans on were re-read in code (section 9.3). Every deviation
from a note is tagged **[DEV-n]** and listed in section 9.2.

Reading order: sections 1-2 are what the owner agrees first (scenarios before solution). Sections 3-6 are the design.
Section 7 is the order of work. Sections 8-9 are limits and decisions.

---

## 1. Problem and assumptions

**Problem.** A ~60 m flat belt conveyor is pulled by three drives (motor, gearbox, pulley, VFD) that share the belt pull
by torque. When one drive weakens or stops, its share moves at once onto the others. Whether that is a slow cost or a
fast disaster depends on the drives' rating margin: with a large margin the two survivors run near rating and age a few
times faster for as long as the plant limps (days); with a small margin they run at 1.3-1.5x and their I²t protection
trips them within minutes, the belt stops loaded and two drives cannot restart it (stranded belt). The customer wants the
failing drive's precursors (vibration, temperature, drag) turned into an early, correct, governed reaction: shed or shape
the load, run the belt empty, stop the weak drive at a chosen moment, keep the belt restartable, and never be fooled by a
feed surge or a dead sensor. We build this as a closed loop: streamsim (the plant, sealed truth, grading) ->
agentic-stream (situations, policy, actuation) -> tamoz (judgment) -> back to streamsim through the plant wire.

### 1.1 Assumption table (all synthetic; real answers replace them as data in a customer folder outside every repo)

Confidence: H = physics/textbook, M = one decent source, L = estimate. "Q#" = customer question in section 8.3.

| Parameter | Default | Why | Conf. | Replaced by |
|---|---|---|---|---|
| Length L, lift h | 60 m, 0 m | README; flat makes feed the weakest lever (CM 1.4) | L | Q8 |
| Belt speed v | 2.0 m/s, constant (VFD speed control) | README | L | Q8 |
| Nominal feed Q | 150 t/h = 41.7 kg/s; material on belt m_ss = Q·L/v = 1250 kg | README, CM 1.4 | L | Q8 |
| Pull model | F_U = F_empty + k_m·m + v·ṁ; F_empty 1000 N, k_m 0.494 N/kg (C·f·g, C 2.2, f 0.02) | CM 1.2-1.4 (loaded 1.7 kN, empty 1.0 kN) | M form, L numbers | Q1, Q8 |
| Drives | 3 identical; D-1, D-2 at head (210° wrap), D-3 at tail | README | L | Q2 |
| Load sharing | Torque follower: D-1 speed master, all motors give equal torque | CM 3.1 "the correct way"; DP 6.1 | M | Q2 (most plan-changing after margin) |
| **Rating margin (regime)** | **A: F_rated 880 N/drive, 1 out -> 0.97x; B: F_rated 607 N/drive, 1 out -> 1.40x** | Coordinator review: plan both | L | **Q1** |
| VFD torque limit k_lim | A 1.5x, B 1.75x (needed so 3 drives can start a loaded belt in B) [DEV-15] | CM 6 | M | Q5 |
| Breakaway | F_b = a0 + a1·m; A: a0 1630 N, a1 1.424 N/kg (m_b 1.8); B: a1 1.152 (m_b 1.6) | CM 5 | L | Q7 |
| Restart trap | A and B: 3 drives start loaded, 2 cannot, 2 start empty (checked as a data test) | README, CM 5 | M | Q7 |
| Motor thermal | ambient 35 °C; rise 80 K = 24 iron + 56·L² copper; τ_w 900 s; hot spot +10 K | DP 3.2, 4.1 | M | Q5 |
| Insulation aging | rate 2^((T_hot − 130)/10); life 20 000 h at rate 1 | DP 3.1 | M | Q5, Q6 |
| Gear wear | rate ∝ L^6 (pitting exponent 5-8) | DP 4.3 | L | Q6 |
| Motor protection | I²t relay class 20 (τ_r 1000 s, I_th 1.05); winding trip 150 °C for 10 s | DP 3.4 | M | Q5 |
| Seized pulley drag | 1830 N (T2 700 N, μ_k 0.35, 210°, capstan, computed offline) [DEV-14] | CM 4.4 | M form, L numbers | Q2, Q9 |
| Bearing failure mode (D-3) | lubrication loss: onset to protective trip 6-9 h, load-dependent wear (∝ L³) | DP 2.6 "hours to days" [DEV-3] | L | Q6 |
| Sensors | per drive: current (VFD), winding RTD, vibration RMS, pulley speed; belt speed; feed (belt scale) | README | L | Q4 |
| Sample periods | current, speed 5 s; temperatures 10 s; vibration RMS 60 s; feed 5 s | DP "sampling" | L | Q4 |
| Allowed actions | feed setpoint; per-drive planned stop/start; belt controlled stop; alarms. No torque bias at first | README | L | Q3 |
| Limp duration (repair wait) | 24 h after a drive is lost or isolated | DP 4.1 (damage matters over days) | L | Q6 |
| Cost model | in equivalent production hours (section 6.3) | owner placeholder | L | Q9 |
| Downstream stop | one routine belt stop per limp period (crusher interlock, shift end) | makes the restart trap bite | L | Q7 |

Derived checks the defaults give (arithmetic, kept as a data test in M0):

| Quantity | Regime A | Regime B |
|---|---|---|
| Drive load L, 3 drives, full feed | 0.646 (model, M0) | 0.964 (model, M0) |
| L with one drive out, full feed | 0.984 (model, M0) | 1.461 (model, M0) |
| Survivor trip after loss (from hot) | never: I² 0.973 < 1.1025 (model, M0) | inverter (τ 700 s) first: 119 s = 1 s + τ_inv·ln((1.996−0.938)/(1.996−1.1025)) closed form, 121 s in the model; the relay alone would take 170 s (model, M0) |
| Latest single `set_feed` after the loss that still saves the survivors and the belt (noiseless, full feed, 30 min watch, 1 s grid) | not needed: survivors never trip | limp feed 18 %: 98 s; cut to 0: 103 s after D-3 trips (model, M2) |
| Survivor hot spot / aging rate after loss | 123 °C / 0.63 (8.5x normal 0.074) (model, M0) | trips first (overtemp would take about 13 min) (model, M0) |
| Highest feed keeping 2 drives able to restart (m ≤ (2·k_lim·F_rated − a0 − drag)/a1) | 57.3 % (716 kg) (model, M0) | 29.3 % (366 kg) (model, M0) |
| Highest feed keeping 2 drives ≤ 1.0x (F_empty + 700.8·f + drag ≤ 2·F_rated) | 104 % (model, M0) | 20.1 % (model, M0) |
| **Safe limp feed** | **54 %** (model, M0) | **18 %** (model, M0) |
| Seized D-3: demand vs capacity of 2 survivors | 3877 N vs 2760 N: stall 3 s after seizure (model, M0) | 3759 N vs 1821 N: stall (model, M0) |

---

## 2. Scenarios (agree these first)

### 2.1 Common frame

- Plant: one composed domain `belt-conveyor-drives` (section 3), two presets `regime-a`, `regime-b`. Instances F-1
  (feeder), B-1 (belt), D-1 (master), D-2, D-3, instruments per drive.
- Seeds: 10 per scenario variant for acceptance, 3 for development. A seed sets the world seed (channel noise) and picks
  onset `at_s` and `severity` from the ranges below. Until the suite draws severity itself (U8), the 10 rows per variant
  are a committed table in the scenario folder (data, not code).
- Every scenario has a sealed **outcome reference**: the same plan with no consumer, advanced noiseless to the horizon
  (U2). Its failure times are the denominators for lead time; its consequences are the do-nothing baseline B0.
- Negative-class share: S0 and S4 runs are at least 35 % of every campaign (streamsim's `negative_class_fraction` rule).
- Pacing: scenarios whose pass criterion is a reaction inside minutes (S2, S3, S6) run `pace: wall` so the consumer's
  reasoning time costs world time. Barrier pacing would hide it (section 8.1, R5). S0, S1, S4, S5 may run barrier-paced.

### 2.2 Trivial baselines (each must lose where marked)

| ID | Policy | Where it is computed |
|---|---|---|
| B0 | Do nothing (no alarm, no action) | outcome reference (U2) |
| B1 | Vibration zone alarm: D-k vibration > 4.5 mm/s for 60 s -> feed 0, stop D-k after 60 s | policy baseline (U4) |
| B2 | Current-rise alarm: any drive current > 1.1x its 1 h mean for 60 s -> feed 67 % | policy baseline (U4) |
| B3 | Always shed: permanent feed 55 % (A) / 30 % (B) | policy baseline (U4) |
| B4 | Today's PLC practice: on any drive trip -> feed 2/3 (proportional derate) | policy baseline (U4) |
| Bd | Detection baselines: fixed threshold, z-score, first difference, moving median, channel silence (hindsight-fitted) | existing `internal/audit` |

### 2.3 Scenario table

Times are world time from run start. "Trip" = first protective trip of a drive; "stranded" = boundary
`B-1.stranded` (belt stopped, mass > 30 % of m_ss, restart margin < 0, for 10 s).

**S0 Nominal day (control, negative class).** Feed follows a diurnal profile 70-100 % with ±5 % noise; no fault.
Horizon 12 h. Truth: no failure, no boundary. Must lose: B2, B3 (cost). Pass: 0 drive-degradation detections; 0
throughput-reducing actions; J_run ≤ J_B0 + 0.1 h.

**S1 Slow bearing failure on D-3 (lubrication loss), regimes A and B.**
- Story: t_on ∈ U[0.5, 2] h: D-3 lube loss starts (ramp fault on `lube_deficit`, severity U[0.8, 1.25]). Damage grows
  with load (∝ L³). Vibration rises first (convex in damage), then D-3 bearing heat lifts its winding temperature against
  its peers at equal current (torque follower) [DEV-5]. Do nothing: D-3 overtemperature trip at t_on + 6-9 h.
  A: survivors run at 0.98x for the 24 h limp period, aging 8.5x (model, M0); the downstream stop at trip + U[1, 6] h
  strands the loaded belt (M0 draws it at t_on + U[9.5, 12] h, after the latest trip). B: survivors trip on the inverter
  100-110 s after D-3 (cascade) (model, M0), belt stops loaded, stranded.
- Truth: `first_observable` (vibration SNR, existing solver), D-3 trip time t_f, survivor trips, stranding time, life
  and gear-wear used per drive, tonnes delivered, all from the outcome reference.
- Must lose: B0, B1 (late: vibration zone C comes < 2 h before t_f), B2 (fires on feed swings), B4 (strands: 67 % feed
  > 57 % in A; cascades in B).
- Pass (≥ 9/10 seeds each): first D-3 detection labeled `bearing` with lead ≥ 2 h before t_f and not before
  `first_observable`; D-3 stopped by command before t_f (planned, not tripped); no survivor protective trip; never
  stranded; feed after isolation ≤ safe limp feed + 5 points; tonnes delivered ≥ 90 % of B3's; J ≤ 0.5·J_B0 and < J_B4.

**S2 Sudden trip of D-3 under full load (no precursor), regimes A and B.**
- Story: VFD fault (step on `vfd_ok`) at t_trip ∈ U[10, 40] min, full feed. A: survivors 0.97x; downstream stop at
  t_trip + U[10, 40] min. B: survivors' inverters trip, cascade at t_trip + 121 s (closed form 119 s) (model, M0). Horizon t_trip + 2 h.
- Truth: t_trip; cascade time (B); stranding time; per-drive damage.
- Must lose: B0; B4 (A: 2/3 feed strands at the downstream stop; B: 2/3 feed one second after the trip still trips the survivors at +201 s (model, M0)).
- Pass: lead time is not graded (unpredictable). Reaction: feed ≤ safe limp feed within 120 s (B; feed 0 saves the
  survivors up to 104 s after the trip (model, M0)) or before the downstream stop (A); no cascade; not stranded; no futile restart attempt;
  J ≤ J_B4 − 1 h. ≥ 9/10 seeds.

**S3 Seized gearbox on D-3 (short precursor, then drag).**
- Story: gearbox oil loss at t_on ∈ U[0.5, 1.5] h: D-3 internal loss rises 3 % -> 30 % of rating over 1-3 h (all
  currents up together at constant feed; D-3 temperature +5-15 K against peers); seizure when the gear temperature state
  crosses its limit. Seizure adds 1830 N drag: demand 3.53 kN > survivor capacity, stall within 3 s, belt stops loaded
  with a locked pulley (slip and fire risk, CM 4.4). Regime A only (B adds nothing new).
- Truth: t_seize; stall time; mass on belt at stall; label `gearbox_seizing` on D-3; look-alike negative: none.
- Must lose: B0 (stranded loaded), B1 (vibration rise too small), B2 (alarms but on the wrong cause; cuts feed only to
  67 %: still stranded loaded), B4 (acts after the stall).
- Pass: belt brought to a controlled stop (or run empty) before t_seize, mass at t_seize ≤ 15 % of m_ss; no loaded
  stall; D-3 named as the suspect (label `gearbox_seizing` or `drag`) ≥ 30 min before t_seize. ≥ 8/10 seeds.

**S4 Feed surge (negative class), regimes A and B.**
- Story: healthy plant; 1-3 feed pulses of 1.3-1.6x for 180-420 s (existing negative-class `surge` shape on the feeder).
  All currents and temperatures rise together, ~30 s after the feed signal. Chosen so that even the worst pulse does not
  trip B (one 300 s pulse of up to 1.67x does not trip B (model, M0); the feeder's pulse lasts 300 s). Horizon 6 h.
- Truth: `is_negative_class`; no failure in the outcome reference.
- Must lose: B2 (alarms and cuts feed every pulse).
- Pass: 0 detections labeled drive degradation; 0 actions that reduce throughput; ≤ 1 R0 watch per pulse; J within
  0.1 h of B0. 10/10 seeds.

**S5 Stuck vibration sensor hides S1 (regime A).**
- Story: S1-A, plus D-3 accelerometer frozen at its healthy value from t_on − U[0.5, 1] h (`stuck` observation fault,
  quality good: nothing in the wire marks it). Evidence left: D-3 winding temperature residual against peers at equal
  current, and the frozen channel itself (flat while the peers' vibration moves with load).
- Truth: two labels: `bearing` on D-3 and `sensor_stuck` on D-3's vibration point.
- Must lose: B0, B1 (never fires), Bd on vibration.
- Pass: D-3 bearing detection lead ≥ 60 min (≥ 8/10 seeds); the stuck channel reported (label `sensor_stuck`) before
  the bearing detection in ≥ 7/10; then S1-A's outcome criteria.
- Variant S5b (later, needs U-A3): the channel drops out (heartbeat lost) instead of freezing.

**S6 Loss of the torque master D-1 (later; worth it because it tests "do no harm" when prediction is impossible).**
- Story: D-1 VFD fault at t ∈ U[10, 40] min. Parameter `master_loss` = `stop_all` (followers lose their reference, the
  plant stops the belt in ~4 s, loaded) or `promote` (D-2 takes the speed loop after 2 s; then S2 applies). Regime A.
- Truth: stop time; stranded (stop_all, 2 drives cannot restart loaded).
- Must lose: B0 with `promote` (as S2-A); with `stop_all` nothing can prevent the stop, so the criterion is harm.
- Pass: no futile loaded restart attempt (each costs a stall and heat); a correct request for manual unloading or an
  empty restart plan; with `promote`, S2-A criteria.

---

## 3. World model

### 3.1 Components and composition

| Instance | Component (new unless noted) | Role | Key ports (type: direction) |
|---|---|---|---|
| F-1 | `streamsim/bulk-feeder@1` | feed setpoint, surge fault, permissive from belt | `feed_out` (bulk_flow: provides), `permit_in` (run_state: consumes) |
| B-1 | `streamsim/belt-conveyor@1` | material on belt, pull, sharing, capacity, stall, restart trap, throughput | `feed_in`, `share_out` (pulley_force), `enabled_in`, `drag_in`, `capability_in`, `capacity_in` (consumes, `sample_s` 1 s), `moving_out` |
| D-1..D-3 | `streamsim/conveyor-drive@1` [DEV-1] | motor + gearbox + pulley + VFD on a shared belt | `share_in`, `enabled_out`, `drag_out`, `capability_out`, `capacity_out` |
| per drive | existing instruments: current-clamp (VFD current), rtd (winding), accelerometer (RMS), tachometer (pulley) | sensors and their faults | |
| B-1 | tachometer (tail pulley), mass-flow / belt scale (feed) | | |

Composition: three drives in one equipment entry with `names: [D-1, D-2, D-3]` (one kind, one kind digest). All
many-to-one links aggregate by `sum`, which is the only aggregate a composition emits (verified,
`compose_follow.go:48`): the running count is the `sum` of 0/1 `enabled`, the drag is the `sum` of drags. The loop
drive -> belt -> drive is closed by `sample_s` on the belt's consuming ports (the one-sample lag is the VFD's reaction).
S6 needs two drive kinds (`drive-master` with one instance, `drive-follower` with two), from the same component with a
`role` parameter: data only.

### 3.2 States and equations (SI units; t in s unless noted)

| # | State (owner) | Equation | Units | Existing form | Oracle (independent, in the test) |
|---|---|---|---|---|---|
| 1 | feed_set (F) | F0, set by `set_feed` | % | F0 + effector assign | - |
| 2 | feed (F) | lag(Q_nom·feed_set/100 · moving, τ 3 s) ·(1 + surge) | kg/s | first_order_lag, product, fault pulse | exponential step response |
| 3 | m: mass on belt (B) | dm/dt = feed − m·v/L [DEV-2] | kg | integrator (signed inputs) + product | m(t) = m_ss + (m0 − m_ss)e^(−vt/L) |
| 4 | F_U: running pull (B) | F_empty + k_m·m + v·feed | N | sum | algebraic |
| 5 | F_req: required force (B) | 2-D curve over (v, m): F_b = a0 + a1·m at v = 0, F_U at v ≥ 0.2 m/s | N | curve (bilinear, exact: linear in m) | table check at nodes and midpoints |
| 6 | drag_sum (B) | Σ_i drag_i | N | follow sum | algebraic |
| 7 | n: drives enabled (B) | Σ_i enabled_i | - | follow sum | integer exact |
| 8 | inv_n (B) | curve n -> [0, 1, 1/2, 1/3] | - | curve (exact at integers) | table |
| 9 | share (B) | (F_req + drag_sum)·inv_n | N | product | algebraic |
| 10 | capacity (B) | Σ_i enabled_i·k_lim·F_rated | N | follow sum | algebraic |
| 11 | stalled (B) | latched trip: capacity − F_req − drag_sum < 0 for 3 s; reset by `start_belt` | 0/1 | sum + trip | time of first crossing + 3 s |
| 12 | v: belt speed (B) | lag(v_set·run_cmd·(1 − stalled)·(n > 0), τ 7 s up, 2 s down) | m/s | regime (two τ) | piecewise exponential |
| 13 | restart_margin (B) | Σ_i available_i·k_lim·F_rated − (a0 + a1·m) − drag_sum | N | sum | algebraic |
| 14 | stranded (B) | trip: (v < 0.1) ∧ (m > 0.3·m_ss) ∧ (restart_margin < 0) for 10 s, via curve indicators and a product | 0/1 | curve, product, trip; declared `boundary` (major) | truth table over the three conditions |
| 15 | delivered (B) | ∫ m·v/L dt / 1000 | t | integrator; KPI | ∫ of the closed-form m(t) |
| 16 | L_i: drive load (D) | share·enabled/F_rated | - | product | algebraic |
| 17 | I_i: current ratio (D) | curve(L): √(0.35² + (1 − 0.35²)L²) at 0.05 steps | - | curve | √ within 0.5 % |
| 18 | θ_i: I²t state (D) | dθ/dt = (I² − θ)/τ_r | - | product + first_order_lag | t_trip = τ_r·ln((I² − I_p²)/(I² − I_th²)) |
| 19 | relay_ok_i (D) | latched trip θ > I_th²; `reset_trip` allowed when θ < 0.9 (requires) | 0/1 | trip | as 18 |
| 20 | T_w,i: winding (D) | lag(T_amb + 24 + 56·L² + ΔT_b(d) + ΔT_g(g), τ_w) | °C | sum, product, curve, first_order_lag | T(t) = T_ss − (T_ss − T0)e^(−t/τ) |
| 21 | overtemp_ok_i (D) | latched trip T_w > 150 for 10 s | 0/1 | trip | as 20 |
| 22 | aging_i (D) | curve(T_w + 10): 2^((T − 130)/10), nodes every 5 K, 40-200 °C | 1 | curve [DEV-12] | exact exponential, ≤ 1.5 % node error |
| 23 | life_used_i (D) | ∫ aging dt / 3600 | h | integrator; KPI | Simpson quadrature of 22 over the closed-form T(t) |
| 24 | gear_used_i (D) | ∫ curve(L): L^6 dt / 3600 | h | curve, integrator; KPI | quadrature |
| 25 | lube_deficit_i (D) | F0 0; S1 ramp fault, severity scales | 0..1 | F0 + ramp fault | min(s·rate·h, 1) |
| 26 | d_i: bearing damage (D) | dd/dt = k_d·lube_deficit·curve(L): L³ | 0..1 | integrator, product, curve (clamp 0..1) | closed form at constant L |
| 27 | vib_i (D) | 1.2 + 0.3·L + curve(d) (convex, 1.2 -> 11 mm/s) | mm/s | sum, curve | table |
| 28 | loss_i: internal drag (D) | 0.03·F_rated + curve(d) + curve(g_i) | N | sum, curve | table |
| 29 | g_i: gear oil distress (D) | S3 ramp fault 0 -> 1 over 1-3 h | 0..1 | ramp fault | linear |
| 30 | seized_i (D) | latched trip g > 0.95 | 0/1 | trip | first crossing |
| 31 | vfd_ok_i (D) | F0 1; S2/S6 step fault −1 | 0/1 | F0 + step fault | step |
| 32 | enabled_i (D) | run_cmd·relay_ok·overtemp_ok·vfd_ok·(1 − seized) | 0/1 | product | algebraic |
| 33 | available_i (D) | relay_ok·overtemp_ok·vfd_ok·(1 − seized) | 0/1 | product | algebraic |
| 34 | drag_i (D) | loss_i + seized_i·1830 + (1 − enabled_i)·0.03·F_rated | N | sum, product | algebraic |
| 35 | stop_permit_i (D) | indicator: demand·curve(n − 1 -> 1/(n − 1)) ≤ F_rated and m ≤ restartable mass; `requires` of `set_running 0` | 0/1 | curve, product | truth table |

Torque follower: every enabled motor gives the same pulley-equivalent force `share`; each drive's own internal loss is
part of the shared demand (row 9). So a failing D-3 does **not** draw more current than its peers; all currents rise a
little, and D-3 shows itself in vibration and temperature residuals [DEV-5]. Droop or passive slip sharing would change
this (Q2) and needs the `ratio` form (U7).

Effectors (data, plant-side preconditions in `requires`):

| Effector | Kind | Args | Plant precondition | Effect |
|---|---|---|---|---|
| `set_feed` | feeder | percent 0-100 | belt moving for raising feed | assign feed_set |
| `set_running` | drive | 0/1 | stop: stop_permit = 1; start: available | assign run_cmd |
| `reset_trip` | drive | - | θ < 0.9, T_w < 120 | reset latched trips |
| `start_belt`, `stop_belt` | belt | - | start: none (a failed start stalls and is visible) | run_cmd, reset stalled |

The `requires` on a drive stop is the PLC permissive a real plant has or should have (Q3). It is defense in depth: the
consumer must order the steps itself; the plant refuses a stop that would overload or strand.

### 3.3 Where today's forms are insufficient

| Need | Today | Plan |
|---|---|---|
| Equal share 1/n | curve over an integer count: exact | use it |
| Unequal share, torque bias, droop: w_i/Σw | no division form | U7 `ratio` (later) |
| Plug-flow transport delay at variable speed | `dead_time` is a fixed delay and needs `time_constant_s`; time-based delay empties a stopped belt | well-mixed ODE now [DEV-2]; variable transport delay later (U11) |
| Arrhenius exponential | F2 refused (`validation_records.go:79-81`) | curve table, error bounded by its oracle [DEV-12] |
| Two-condition trip (row 14) | a condition is one comparison | curve indicators + product: closed forms |
| Per-instance parameters in a composition | one entry = one kind | separate entries only for S6 roles |
| Belt speed dynamics under torque limit | not modeled | stall after 3 s instead of a slowdown (known limit) |
| Seized drag depends on slack tension | constant parameter | capstan computed offline [DEV-14] |

---

## 4. Upgrades per repository

Size: S under a day, M a few days, L a design round plus implementation (agent work with tests and gates). Milestone
column: the first milestone that needs it. "Must" = needed for the first demonstrable scenario chain (M0-M3).

### 4.1 streamsim

| ID | Gap | Owner module / layer | Design sketch | Why not a workaround | Size | Deps | Rule risks | When |
|---|---|---|---|---|---|---|---|---|
| U1 | No feeder, shared belt, conveyor drive; no port types for them | `component` edge data (`internal/component/internal/files/builtin`), `vocabulary.json` | three component files, five port types (`bulk_flow`, `pulley_force`, `run_state`, `drag_force`, `force_capacity`), two presets (A, B), instruments reused; a synthetic composition in `examples/compositions/` | Domains are data: this is exactly the intended extension path | M | - | none (data); new library files are in the module's edge package (MUST 4) | M0 must |
| U1t | Analytic cross-check of the coupled loop | tests in `world` and `component` | oracle tests for rows 3, 9, 18, 20, 23, 26 and the restart table (section 1.1) against independent recurrences; `domain lint` passes; silent_no_effect test per lever (each changes survivor load) | non-negotiables 2 and 8 | M | U1 | test files ≤ 300 lines | M0 must |
| U2 | No consequence truth: "when would D-3 trip, would the belt strand, what would it cost if nobody acts" | `truth` domain (sibling of `solver_scan.go`), `model`, new sealed file + contract `outcome-reference-v0.1`, `plantwire` app (write at session start), `director` (seal/reveal) | advance one noiseless world with every planned disturbance and no consumer call, at the world step, to the horizon; record per entity the first trip time of each trip state and each declared `boundary`, and start/end/max of each declared KPI. Run-level, not per label [DEV-10]. Exact counterfactual because world states carry no process noise (only channels and stochastic faults, both seeded) | Truth is generative and sealed, computed by the truth module from sealed inputs before the consumer runs | M | U1 | sealed-oracle boundary: no operator-view field (`mcp_protocol` test); 15-line functions; contract fixture + schema test | M1 must |
| U3 | Scoring is one entity, one effector, one recovered state | `score` domain (new `outcome` section), `grading` (pass reference), `report` (render), contract bump of the scorecard bundle | replay the run artifact (grading already does), read the replayed world's boundaries, trips and KPIs at horizon, compare with U2; metrics of section 6; cost J with weights from the domain (`ground_truth.outcome_weights`, data) | Generic: names states and boundaries, not conveyors; online/offline parity kept by computing both from replay | L | U2 | no consumer knowledge; scoreBundleVersion bump; `ViolatedBoundaries` gets its first production caller (verified unused today) | M1 must |
| U4 | Trivial-baseline audit covers detection only, not policies | `audit` domain, domain contract (`ground_truth.baseline_policies`) | closed record: `{name, when: condition on a kind's channel or state, for_s, then: [{effector, target: matched or named instance, args, delay_s}]}`; one condition, a fixed call list, no boolean composition; each policy is a noiseless world run; result sealed beside U2 | Non-negotiable 9 applied to the thing this use case grades (reactions) | M | U2, U3 | must not grow into a rule language: owner decision D-B; oracle: a hand-written command log equals the policy run | M1 must (B0, B4); B1-B3 later |
| U5 | Plant-wire sessions run until stopped; campaigns need a fixed horizon | `plantwire` app, session-plan contract | `horizon_s`: the session ends itself, U2 uses it | needed to compare a run with its reference over the same window | S | - | contract bump; fan plans unaffected (field optional) | M1 must |
| U6 | `Grade` refuses more than one label | `grading` app | grade a multi-label run as one `RunScore` (already returned by `GradeRun`) plus the run-level outcome section | S5 has two labels; outcome scoring is run-level anyway | S | U3 | none | M2 must |
| U7 | No division: torque bias, droop | `model` enum, `domain` validation, `world` `algebraic.go`, schema, TECHNICAL_DESIGN equation | F1 form `ratio`: product of numerator inputs over product of denominator inputs, `floor` > 0 for the denominator (value = numerator/floor below it); algebraic; no `dt_ms` | closed named form like `curve`/`product`, not an expression | M | - | needs design note + independent oracle (non-negotiable 2); division-by-zero rule stated | M3, before prediction (owner decision 2026-11-08: real belts share unevenly) |
| U8 | Suite/campaign cannot draw severity, multi-fault scenarios hand-listed | `suite` domain | `severity_range` per fault weight; campaign phases already allow several faults | replaces the committed seed tables | S-M | - | determinism: draw order fixed | M4 |
| U9 | Consumer verdict has no predicted failure time | `consumer-verdict` contract, `score` judgment | optional `predicted_failure_at` per detection; PHM α-λ, relative accuracy, prognostic horizon (section 6.2) | RUL accuracy is what the customer buys; needs it graded | S-M | U2 | contract bump | M4 |
| U10 | "How early must one act" | `truth` | latest time a named remedial call (e.g. `set_feed` to x) still avoids a boundary: bisection over noiseless worlds (~12 runs) | gives a real deadline instead of `deadline_s` text | M | U2, U4 | runtime cost bounded by horizon/step | dropped (2026-11-08): each scenario has a fixed deadline |
| U11 | Variable-speed transport delay | `world` new form | plug-flow over belt distance | removes [DEV-2] approximation | M | - | new form + oracle | later, if Q8 says long belt |
| U12 | Plant-wire sessions take no perturbations (injection probe, drops) | `plantwire`, session plan | `perturbations` like the director's | delivery faults over the real wire path | M | - | ledger must record them | not in this plan (2026-11-08): no scenario needs it; before a real plant |

No change needed: plant wire protocol (`not_before_ns` exists unset), `peer_residual` detector, KPI truth, `boundaries`,
effector `requires`/`lease`, instrument faults (`stuck`, `dropout`, `drift`), canvas (draws and runs the plant; no belt
widget, states and charts suffice).

### 4.2 agentic-stream

| ID | Gap | Owner module / layer | Design sketch | Why not a workaround | Size | Deps | Rule risks | When |
|---|---|---|---|---|---|---|---|---|
| A1 | Three devices of one kind on one entity merge: readings keyed by (kind, metric), no device in event data (verified `site.go:53-73`, `events.go:50-52`) | `plantwire` domain (site file) | a site reading rule may name `device`; lookup device-first, then kind; so D-3's `current` emits `drive3.current.observed` (schemas added via `--event-schemas`) [DEV-6] | generic data mapping; keeps one kind digest for identical drives (playbook pins by kind digest) | S | - | thermal site file unchanged; site schema additive; experiment guards | M2 must |
| A2 | Conveyor spec, schemas, kinds, site, catalog, principals | data in a customer-style folder; synthetic copy under `examples/belt-conveyor/` for tests | line entity `conveyor-1`; per-position operators; CEL peer residuals (`d3_temp − mean(d1,d2)`), feed-normalized current, phases section 5; catalog routes `reduce_feed` (R1, presets `limp_a`, `limp_b`, `restore`), `stop_feed` (R2), `planned_stop_d{1,2,3}` (R2), `start_belt` (R2), `watch` (R0) | domains are data (AGENTS "per-domain code branches" forbidden) | M | A1 | do not edit the thermal spec/catalog | M2 must |
| A3 | Line-as-one-entity: one dead heartbeat makes the line `uncertain`, which denies every R2 intent on it (verified `rules.go:38-41`, `risk.go:68`) [DEV-8] | `spec` (input grammar), `engine` completeness | input-level `completenessRole: required | advisory`; an advisory input's loss becomes a fact (`input_stale.<name>`), not line uncertainty | a principled declaration of which evidence the decision rests on | M | A2 | compiled-digest pins: additive optional field must not change existing digests; invariant 2 proof updated | M4 (S5b) |
| A4 | Ordered steps (cut feed -> run empty -> stop D-3 -> restore limp feed) | none now | phase-driven: each verified step moves the phase, whose trigger admits the next episode; the plant's `requires` refuses an unsafe stop [DEV-9] | uses existing freshness rule as a feature (a resolved situation cancels later steps) | 0 (data) | A2 | W15: a refused command leaves verification `awaiting` (open owner item) | M3 |
| A4g | Generic `after`/`not_before` on intents | `contractsv1`, `policy`, `actions` | only if A4 proves too slow (each step one episode) | - | L | - | contract change on both sides of the proto | later, only if needed |
| A5 | Time-to-threshold feature | `operators` (new kind `extrapolate`), spec schema | `(limit − level)/slope` over a declared window, published with a validity flag; feeds transitions and Tamoz | CEL can compute it in a transition but cannot publish it as a fact | M | A2 | engine state/restore/replay tests; new operator kind is a design change (AGENTS) | M3 if S1 lead fails without it, else M4 |
| A6 | A streamsim consumer verdict for a plant-wire run, with the device as subject | `internal/testsupport/campaign` (test support, exists) | diagnosis-code -> device map from the scenario fixture; detections from accepted decisions; actions from commands | grading already authoritative on actions; detections need the subject | S | A2 | test-only; no production path | M2 must |
| A7 | Route target chosen by a selector (one `planned_stop` intent, not three) | `device` catalog | preset may name a bound target | removes triplication | S | - | catalog digest | dropped (2026-11-08): saves only repeated catalog entries |
| A8 | Safe stop over plant wire not implemented | `plantwire` | per `plant-wire-adapter.md:287` | - | M | - | contract on both sides | already built: the plant wire contract has `safe_stop` (checked 2026-11-08) |
| A9 | Operator `where` filter is compiled and validated but never evaluated (verified by grep: only `cel.go:58`, `spec.go:94`) [DEV-7] | `spec`/`operators` | report to owner: implement over an `event` variable or remove | a spec field that silently does nothing | S-M | - | - | report now, fix outside this plan |

Unchanged and reused: risk routes (R0/R1 automatic, R2 approval, R3/R4 denied), interlock, owner epoch, leases,
read-back verification from the device's own readings, rule promotion with human approval, run export and `verify-run`.

### 4.3 tamoz

| ID | Gap | Where | Design sketch | Size | When |
|---|---|---|---|---|---|
| T1 | Conveyor domain fixture (synthetic): diagnosis codes `bearing`, `gearbox_seizing`, `vfd_trip`, `feed_surge`, `sensor_stuck`, `idler_drag`; intents, risk classes, fixture responses, a benchmark family | `test/fixtures/domains/conveyor.json` (ADR-058) | pump-station as template | M | M3 (runtime needs none: intents and prompt arrive from agentic-stream) |
| T2 | W1: frame does not define output protocol and allowed intents | `tamoz-agent-kernel` episode frame | as the fan owner item | S | recommend before M3 (more intents) |
| T3 | W2/W4: redispatch overlap, failure reason lost | worker + dispatcher (owner chooses side) | as the fan owner item | S-M | before M4 campaigns |
| T4 | Production approval relay (Telegram) for R2 stops | `tamoz-stream` ports, `tamoz-telegram` | relay exists, adapters missing | M | already built: `tamoz-telegram` (owner, 2026-11-08) |
| T5 | Proactive alert ("D-3 predicted to fail in ~2 h; plan; approve?") | subscriber handler on a decision/phase event + comms route | the R2 approval request is the alert for the first slices | S-M | already covered (owner, 2026-11-08): the R2 approval request is the alert |
| T6 | Real-model A/B/C on the conveyor scenarios vs the baselines | `script/*_eval` style runner reading streamsim scores | 10 seeds per variant, background | M | M4 |
| T7 | `forecast.run` is on the tool allowlist but agentic-stream's EvidenceTools proto has no forecast (verified by grep) | both | leave out of the prompt; no forecast tool claimed | - | note |

---

## 5. End to end: S1-A through the three repositories

Data path (left to right) and control path (right to left). One entity `conveyor-1` in agentic-stream; the drive is a
position inside it.

| Step | World time | streamsim | agentic-stream | tamoz | Risk / approval |
|---|---|---|---|---|---|
| 1 | 0 | session plan `s1-a/seed-03.plan.json` (domain, seed, disturbances: D-3 lube loss at 4300 s, severity 1.1; downstream stop; `horizon_s`); writes sealed labels + outcome reference (U2) + baseline runs (U4) | `serve` with site, kinds, catalog, spec; host admits births of F-1, B-1, D-1..3 | worker idle | - |
| 2 | 0 -> | `DDATA` per device; `NSYNC` each slice | A1 maps D-3 readings to `drive3.*` events on `conveyor-1`; windows: vib 15 min mean and 2 h slope, temp residual 30 min, current sum vs feed | - | - |
| 3 | ~t_on+2.5 h | vibration of D-3 above peer band | transition `normal -> d3_watch` (`vib_d3 − max(vib_d1, vib_d2) > 0.8` for 10 min); trigger `watch` (fast lane) | Decision: `install_watch_condition` | R0 automatic |
| 4 | ~t_on+4.5 h | D-3 temp residual 6 K and rising, currents equal | `d3_watch -> d3_degrading` (temp residual > 5 K and slope > 1 K/h; feed-normalized currents flat, so not a surge); deep-lane episode | reads `features.query`, `evidence.get`; diagnoses `bearing` on D-3, uncertainty, proposes `reduce_feed: limp_a` | R1 automatic |
| 5 | +30 s | `DCMD set_feed 55` -> `DRECEIPT accepted`, `DRESULT succeeded`; feed reading 55 % (read-back) | verification from F-1's own reading; phase `shedding` | - | - |
| 6 | +2 min | D-3 load falls, its wear rate falls (∝ L³): failure moves later (the truth's reference does not move) | phase `shedding` stable; trigger plan step 2 | proposes `stop_feed` (run empty) and states why: restart trap with 2 drives | R2: approval request (scripted approver in sim, Telegram later T4) |
| 7 | +2 min + approval | feed 0; m decays with τ 30 s | verified; after m-proxy (feed 0 for 120 s, current sum at empty level) phase `belt_empty` | proposes `planned_stop_d3` | R2 approval |
| 8 | | `set_running 0` on D-3: `requires stop_permit` holds (empty belt, 2 drives at 0.57x) -> stopped; belt keeps running on D-1, D-2 | verified by D-3 run-status reading; phase `d3_isolated` | proposes `reduce_feed: limp_a` (restore to 55 %) | R1 automatic |
| 9 | until horizon | survivors at 0.79x; downstream stop at t_s: belt stops with 688 kg < 709 kg, restart with 2 drives succeeds (operator `start_belt`, R2) | situation stays `d3_isolated`; repeated case becomes a draft rule after 3 verified cases (human approves) | may draft the playbook rule | rule needs human approval |
| 10 | end | run folder: artifact, ledger, journals, labels, outcome reference | `export-run`, `verify-run`; A6 writes the streamsim consumer verdict | - | - |
| 11 | after | `streamsim score` (grading): replays, compares with U2 and U4, prints the outcome section; `report compare` against the B4 run | - | T6 aggregates across seeds | - |

Notes: the "motor 3 notifies the others" of the brief is steps 3-8: D-3's own evidence changes the line's phase, and the
line's plan protects D-1 and D-2 before they are loaded. The fan experiment's chain is reused unchanged (one `serve`, one
session plan, R1 automatic, R2 approval, verification by the device's own reading).

---

## 6. Grading

### 6.1 Outcome metrics (U3), computed by replaying the run and comparing with the sealed reference (U2)

| Metric | Definition | Source |
|---|---|---|
| Failure time t_f | first trip (or seizure) of the faulted entity's declared failure state in the outcome reference | U2 |
| Detection lead | t_f − first verdict detection on the faulted entity with a correct label; detections before `first_observable` are `suspicious` (existing rule) | verdict + label |
| Action lead [DEV-17] | t_f − first effector call that took effect and reduced the faulted drive's load or stopped it | run artifact (authoritative) |
| Planned vs unplanned | faulted drive stopped by a call before t_f (planned) or by protection (unplanned) | replay |
| Survivor life consumed | Σ over non-faulted drives of Δ`life_used` (h) and Δ`gear_used` (h), run vs reference, same window | replay KPIs vs U2 |
| Cascade | protective trips of non-faulted drives (count, first time) | replay trip states |
| Stranding | count and duration of `B-1.stranded` boundary violations | replay `ViolatedBoundaries` |
| Throughput given up | tonnes `delivered`: reference − run (negative = more delivered) | KPI |
| False actions | throughput-reducing calls in a run whose reference has no boundary violation and no failure (S0, S4); calls on a wrong entity | calls + U2 |
| False detections | detections labeled degradation on a negative-class label or a non-faulted entity | verdict + labels |
| Futile restarts | `start_belt` calls followed by `stalled` within 10 s | replay |

### 6.2 Prognostic metrics (later, with U9)

Prognostic horizon PH = t_f − t_i (first time the predicted failure stays inside ±α of truth, α = 0.2, β = 0.5);
α-λ at λ = 0.5, 0.75; relative accuracy RA(λ) = 1 − |RUL* − RÛL|/RUL*; asymmetric late penalty (late costs 2x early).
Truth RUL* uses the outcome reference **as if nobody acted**; once the consumer acts the real failure moves, so RUL is
graded only up to the first effective action (section 8.2, L6).

### 6.3 Cost and pass rule

J = 4·stranded + 4·cascade_trips + 0.5·unplanned_trip + (16/20000)·Δlife_h + (16/50000)·Δgear_h + tonnes_lost/150 +
0.25·false_actions + 0.5·futile_restarts, in equivalent production hours. Weights are domain data
(`ground_truth.outcome_weights`) and are the customer's (Q9). A run passes a scenario when its hard constraints hold
(section 2.3) and J meets the stated ratio to B0 and B4. A scenario is valid only if B0 fails it (else it tests
nothing), and is flagged `easy` if any of B1-B4 passes (the trivial-baseline audit for policies).

---

## 7. Milestones

Each milestone: its own branch cut from the previous milestone's tip in each repo it touches (the first from
`belt-conveyor-drives`); one commit per improvement; Opus writes code and tests, Sonnet investigates and reviews (two
lenses at milestone end: correctness/architecture, adversarial/truth boundary), Haiku runs gates; the coordinator reviews
every diff before each commit and lists every `test/architecture` / `scripts/` change. Merge to `main` only at the end.

**Fan re-check** ("FR" below): streamsim `go test ./internal/device/... ./internal/deviceworld/... ./test/contract/...`
and `make phase-gate`; agentic-stream `go test ./cmd/agentic-stream -run Experiment` and
`go test ./internal/notify -run Experiment` plus `make ci-check`; FR+ adds the simulated rung P4 from the kit
`/Users/ghassan/my-projects/.e2e-run/fan-p4-kit/` (`P4-RUNBOOK.md`), Haiku in the background, all seven criteria.

| M | Branch | Repos | Content | Exit criteria | Fan |
|---|---|---|---|---|---|
| **M0 Scenarios as data** | `belt-m0-scenarios` from `belt-conveyor-drives` (streamsim) | streamsim only | U1, U1t; composition + presets A/B; session plans for S0-S5 × seed tables; a scripted "reference policy" command log per scenario (what a good controller does, hand-written); from section 9.6: inverter overload trip (O6), drive count 2-6 (O7), starting boost (O8), seized drag as a preset (O10), downstream run permit (O11), scenario files for runs and validated session plans (O12) | Section 2 agreed (O1). Oracle tests green. Do-nothing runs reproduce the story table, with times from the model's own equations (O14) (S2-B cascade at the oracle's time ± 15 s; S3 stall < 5 s; S4 no trip; S1 trip window 6-9 h after onset in 10/10 seeds). Every lever changes survivor load (silent_no_effect). The reference policy avoids stranding in every positive scenario. `domain lint` clean; canvas opens and runs the plant; a 2-drive belt passes its test (O7). `make ci-check` | FR |
| **M1 Truth and grading** (done 2026-11-08, evidence in [M1-DESIGN.md](M1-DESIGN.md) section 7) | `belt-m1-truth` | streamsim | U2, U3, U4 (B0, B4), U5; contracts and fixtures; report rendering | Outcome reference computed at scoring time (O15) for plant-wire and scripted runs; grading the reference-policy log beats B0 and B4 on S1-S3 (and S5) and ties B0 on S0/S4; B0 fails every positive scenario (scenario validity); outcome determinism; `make phase-gate` | FR |
| **M2 First closed loop: S2-B** (reaction only) — done 2026-11-09, see M2-DESIGN.md §6 | `belt-m2-loop` | streamsim (kind files, plans), agentic-stream (A1, A2 minimal, A6), tamoz none | spec with `line_overload` phase, `reduce_feed: limp_b` R1; Tamoz executor; `pace: wall` | 3/3 seeds: feed ≤ 30 % within 120 s of the trip, no cascade, not stranded; graded by `streamsim score` with A6's verdict; `verify-run` pass; zero quarantines | FR+ |
| **M3 Prediction and the planned isolation: S1-A/B, S3, S4** — done 2026-11-09, see M3-DESIGN.md §8 | `belt-m3-predict` | agentic-stream (A2 full, A4 data, A5 if needed), streamsim (U6, U7 unequal sharing), tamoz (T1, T2 recommended) | phases section 5; R2 steps with scripted approver; surge discount | 3 seeds each: S1 pass criteria; S3 belt empty before seizure; S4 zero false actions. Decision on A5 recorded | FR+ |
| **M4 Hard cases and evidence** | `belt-m4-evidence` | all three | S5 (stuck), S5b with A3, S6 with master/follower kinds, U8, U9, B1-B3, T3, T6; 10-seed acceptance campaign (background); human-approved playbook rule for the S2 reaction | Acceptance: every scenario's pass rule over 10 seeds; A/B/C report (rules only, Tamoz, both) against baselines | FR+ |
| **M5 (dropped, 2026-11-08)** | - | - | Its production pieces exist (T4, T5, A8); U7 moved to M3; U10 and A7 dropped; U12 not in this plan | - | - |

Scope (owner decision, 2026-11-08): the whole plan is built before the Retrofit Designer
(`agent-research-lab/plant-retrofit-designer/`), which is the next big milestone: M0-M4 with all seven scenarios
(S0-S6), plus U7 (unequal load sharing) in M3. M5 is dropped: its production pieces already exist (Telegram approvals,
plant-wire safe stop), U10 and A7 are not needed, and U12 waits until a real plant. Changes go to streams-simulator and
agentic-stream; tamoz only where a milestone needs it (T1-T3, T6 confirmed when their milestone starts). Nothing is
built for the designer, and no repository gains designer-specific code.

Smallest demonstrable slice: M0 alone already demonstrates the customer's story in the canvas (do-nothing vs the
hand-written reference policy, side by side with `report compare`), with no cross-repo change.

---

## 8. Risks, known limits, customer questions

### 8.1 Risks

| # | Risk | Effect | Mitigation |
|---|---|---|---|
| R1 | The real margin is neither A nor B (e.g. a heavy steel-cord belt, 7.5-15 kW drives) | scenario emphasis wrong | margin is one preset; scenarios hold for any margin; Q1 first |
| R2 | Customer uses droop or passive slip sharing | D-3 shows in current; equal-share model wrong | U7 `ratio` + a sharing-mode parameter; S1 still valid with different evidence |
| R3 | The customer allows alarms only | control path moot | M0-M1 and the detection half of M2-M3 stand; actions become advisory intents (R0) |
| R4 | U4 policy baselines grow into a rule language | violates "no expression language" | one condition, a fixed call list, owner decision D-B, reviewed at M1 |
| R5 | Barrier pacing hides reasoning latency | reaction looks instant | `pace: wall` for S2, S3, S6; report the measured decision latency |
| R6 | Real-model runs are slow and costly | campaigns take hours | 3 seeds in development, 10 in the background at M4; rules-first after promotion |
| R7 | Line-as-one-entity grows unwieldy (three of every operator, phase, route) | spec size, review load | accepted for one line; A7 trims routes; cross-entity operators only for many lines (RA G2 route 2) |
| R8 | Fan experiment breaks through A1/A3 or contract bumps | acceptance test #1 | additive fields only; FR/FR+ per milestone; guards named in each handoff |
| R9 | Synthetic numbers taken as calibrated | false confidence with the customer | every number in section 1.1 tagged; report says "synthetic plant" |
| R10 | Customer data leaks into a repo | owner rule broken | real spec/site/catalog/plans only in a per-customer folder passed by path; repos carry the synthetic plant only |

### 8.2 Known limits (accepted, good enough for now)

- L1 Well-mixed material model: a feed cut empties the belt exponentially (τ 30 s), not as a 30 s ramp [DEV-2].
- L2 No belt speed dynamics under torque limit: overload stalls after 3 s instead of slowing; no speed-dip signal.
- L3 Seized drag is a constant; wet/dry μ and take-up tension are not dynamic.
- L4 Aging by a 5 K table (≤ 1.5 % error); bearing life load coupling partial (DP 4.2) and not modeled for survivors.
- L5 Sudden trips (S2, S6) cannot be predicted; graded on reaction and harm only.
- L6 RUL truth is the do-nothing failure time; once the consumer acts it is no longer the real failure time.
- L7 One line, one entity; no fleet of conveyors, no fleet caps on rules.
- L8 No delivery perturbations over the plant wire (U12, not in this plan).
- L9 Time scales of degradation are not calibrated on conveyor drives (no public dataset; DP 8).
- L10 Torque bias not modeled before U7 (M3); physics says it buys little for bearing faults (bearing heat is friction).

### 8.3 Customer questions, ranked by how much each answer changes the plan

| Rank | Question | What changes |
|---|---|---|
| Q1 | Installed drive ratings vs running load: currents at full feed with three drives; what happened in past single-drive trips (did the others trip, how fast)? | regime A or B; whether the story is aging or cascade; reaction windows |
| Q2 | Load-sharing mode (torque follower, droop, slip, one VFD per motor or shared); which drive is master; what the PLC does when master or a follower trips | sharing equations, evidence for D-3, need for U7, S6 relevance |
| Q3 | Which automatic actions are allowed and at what approval: feed, per-drive stop, belt stop, torque bias, alarms only; existing PLC permissives and sequences | catalog, risk classes, A4, whether control is advisory |
| Q4 | Which sensors exist, where, at what rate (vibration type, winding RTD, VFD current/torque, belt scale, pulley speed) | which scenarios are detectable; S5 |
| Q5 | Protection settings: relay class, VFD overload and torque limits, trip temperatures, zero-speed and slip switches | I²t window, cascade timing |
| Q6 | Failure history: which component fails first, warning seen and when, how long the plant limps before repair | S1/S3 time scales; value of aging |
| Q7 | Can two drives restart a loaded belt (tried)? How often does the belt stop for other reasons? | restart trap weight; downstream stop rate |
| Q8 | Geometry: length, lift, belt type and width, drive positions, take-up, lagging, wet season | pull numbers; feed lever strength; slip and seized drag |
| Q9 | Cost model: cost of a stop, a stranded belt, a motor, lost tonnes per hour | weights of J |
| Q10 | Data and deployment: historic data for calibration, model provider egress, site network | calibration (L9); Tamoz provider |

---

## 9. Decisions, alternatives rejected, deviations

### 9.1 Decisions

| # | Decision | Alternative rejected | Why |
|---|---|---|---|
| D1 | New `conveyor-drive` component | add ports to `motor-drive@1` (RS 1.5) | motor-drive's load is a torque profile in Nm and its current is power-based (no stall current); retrofitting changes every composition and digest that uses it |
| D2 | Torque-follower equal sharing first | droop/slip first; division form first | best-practice default (CM 3.1); needs only curve 1/n; U7 later on Q2 |
| D3 | Line as one entity in agentic-stream, per-device reading rules (A1) | one kind per drive position (RA G1a); cross-entity operators (RA G2 route 2) | A1 is small and generic and keeps one kind digest; cross-entity is L and a design change |
| D4 | Ordering by phases plus plant `requires` | a sequencer "PLC" in the twin; a new `after` field | no new mechanism; plant-side permissive is real defense in depth; the twin's sequencer is a stage counter and cannot target one drive of a shared kind |
| D5 | Outcome reference as a run-level sealed file | new fields on each ground-truth label (RS G9) | outcomes belong to the run when several disturbances coexist (S5) |
| D6 | Policy baselines in `audit` as closed records | baselines as extra consumers; no policy baseline | non-negotiable 9 for reactions; noiseless and deterministic |
| D7 | Planned drive stop is R2 (approval), emergency stops stay in the plant | R3 like pump-station `stop_motor` (RT 3.3) | R3 is never executable; a planned stop of one drive on an empty belt is operational, not a safety function |
| D8 | First demonstrable cross-repo slice is S2-B (reaction), then S1 (prediction) | S1 first | S2-B needs no prediction machinery and proves the full chain on the hardest timing |
| D9 | No F2 Arrhenius; curve with an oracle | implement F2 | table error bounded; F2 adds a module change for no scenario gain |
| D10 | Do not modify the conveyor@1 component | rewire its own motor | it is a belt with its own motor; the new belt shares several drives |

### 9.2 Deviation register

| Tag | Note and place | Deviation | Reason |
|---|---|---|---|
| DEV-1 | RS 1.5, G1 | new drive component instead of motor-drive ports | D1 |
| DEV-2 | RS 2.2, 2.4 (dead_time + integrator) | well-mixed ODE dm/dt = ṁ − m·v/L | a time-based delay keeps discharging a stopped belt and loses material on restart; the ODE is speed-aware, exact at steady state, has a closed-form oracle |
| DEV-3 | README "hours of precursor"; DP 5.4 (compress days) | S1 is a lubrication-loss failure lasting hours in honest world time | DP 2.6 supports hours for lube loss; no clock compression needed |
| DEV-4 | README "survivors age, long-run damage" | in regime A the insulation life lost is small (~14 eq. h per survivor per day of limp); the main value is avoiding stranding and unplanned trips | arithmetic in section 1.1; DP 4.1 |
| DEV-5 | RS G13 (peer residual over currents), RA G4 (current normalization), README sensor story | under torque following, D-3 does not draw more current; evidence is vibration and temperature residuals at equal current | equal motor torque by construction; currents identify line drag, not the drive |
| DEV-6 | RA G1 option (a) | site-file per-device rule (A1) | keeps one kind file and digest for identical drives |
| DEV-7 | RA F1.3 ("optional per-operator `where` CEL filter") | `where` is validated but never evaluated | grep: only `cel.go:58` and `spec.go:94` reference it |
| DEV-8 | RA section 5 (did not flag) | single-entity completeness denies R2 on the whole line when any input is stale | `rules.go:38-41`, `risk.go:68`, `stream-processing.md` completeness rule |
| DEV-9 | RA G5 options | phases, not a new ordering field | D4 |
| DEV-10 | RS G9 | run-level outcome file | D5 |
| DEV-11 | RS 5.2 (audit is detection only) | policy baselines added | D6 |
| DEV-12 | RS G7 | no F2 | D9 |
| DEV-13 | RS G4 (division needed) | deferred | equal sharing needs none |
| DEV-14 | CM 4.4 (drag depends on T_out) | constant seized drag | no tension dynamics; known limit L3 |
| DEV-15 | README "1.4x" with CM's 150 % VFD | regime B uses k_lim 1.75 and m_b 1.6 | with 150 % and m_b 1.8 three drives could not start a loaded belt in B (2731 N < 3410 N): an impossible plant |
| DEV-16 | RT 3.3 (drive stop R3) | R2 | D7 |
| DEV-17 | RS 5.5 (lead time from detections) | primary lead time from effective actions | the simulator's own call log is authoritative; detections need the consumer's subject mapping (A6) |

### 9.3 Claims verified in code for this plan

| Claim | Where read |
|---|---|
| Composition emits only `sum` | `internal/component/internal/domain/compose_follow.go:43-53` |
| `curve` clamps and interpolates linearly; no division form | `internal/world/internal/domain/algebraic.go:9-58` |
| F2 and F3 refused | `internal/domain/internal/domain/validation_records.go:74-81` |
| Truth solver = clean vs faulted noiseless worlds, 60 s step, 72 h horizon | `internal/truth/internal/domain/solver.go:1-35` |
| Plant-wire sessions seal one label per disturbance (sample 120 s) | `internal/plantwire/internal/app/connect_label.go` |
| `Grade` needs one label; `GradeRun` returns many | `internal/grading/internal/app/grade.go` |
| Scoring is entity-bound to one expected effector | `internal/score/internal/domain/loop_metrics.go:47-49`, `loop.go:48` |
| `ViolatedBoundaries` has no production caller | grep over `internal/` |
| Fault parameters: severity only; it scales the envelope | `internal/world/internal/domain/faults.go:59-127`, `dynamics.go:73` |
| Session-plan disturbances carry `at_s`, `duration_s`, `severity`; no horizon | `contracts/session-plan-v0.1.schema.json` |
| Effector `requires`, `lease` | `contracts/domain-spec-v0.2.schema.json` `$defs.effector` |
| agentic-stream risk routes, consequential classes, completeness denial | `internal/contractsv1/internal/domain/risk.go:25-31,68`, `internal/policy/internal/domain/rules.go:38-41` |
| Site readings keyed by (kind, metric); event data has no device | `internal/plantwire/internal/domain/site.go:19-73`, `events.go:44-52` |
| Catalog route resolves one target per entity binding | `internal/device/internal/domain/catalog.go:17-42`, `materialize.go:40-52` |
| CEL environment: features, situation, delta, event_time, watermark | `internal/spec/internal/domain/cel.go:13-27` |
| Fan work merged to main in streamsim (`8c2b11fe`, PR #54) and agentic-stream (`fan-a` contained in `belt-conveyor-drives`) | `git log`, `git branch --contains` |

### 9.4 Coordinator review (2026-11-07)

Re-derived from section 1.1: loads 0.64/0.97 (A) and 0.93/1.40 (B), restartable mass 709 kg (A) and ~430 kg (B),
seized demand 3530 N, limp load 0.79x at 55 % feed. All hold.

- **Correction, cascade time in B.** Row 61 uses I² = L² = 1.96, but row 17 defines current by the magnetizing curve:
  at L 1.40, I² = 0.35² + (1 − 0.35²)·1.96 = 1.842, so t = 1000·ln((1.842 − 0.87)/(1.842 − 1.10)) ≈ 270 s, not 238 s.
  The M0 exit criterion "238 ± 15 s" must come from the model's own equations (the oracle test), not from this table.
  Same for the S4 bound (649 s) and the S2 B4 time (+509 s): recompute with row 17.
- DEV-4 is the main message for the owner: in regime A the survivors' insulation loss is small; the value is avoiding
  unplanned trips and a stranded loaded belt, not long-run aging. The co-founders' story holds as told only if the real
  margin is small (regime B), where it becomes a cascade in minutes.
- Row 9 puts D-3's internal loss into the shared demand. That is a simplification (the loss sits between D-3's motor and
  its pulley); it is consistent with equal motor torque and is acceptable as a known limit.
- A9 (agentic-stream `where` filter never evaluated) and DEV-8 (one stale input denies every R2 action on the line) are
  real findings outside this use case's scope; report them to the owner separately.

### 9.5 For the owner to decide before M0 (answered 2026-11-08: section 9.6)

1. Agree the scenario set, horizons and pass numbers of section 2 (or change them): the success bar.
2. Both margin regimes as presets of one plant (recommended), and the synthetic numbers of section 1.1 in the repo.
3. A new `conveyor-drive` component rather than changing `motor-drive@1` (D1).
4. Torque-follower equal sharing as the default model until Q2 is answered (D2).
5. Before M1 (not M0): policy baselines as closed records in `audit` (D6/U4) and the outcome reference as a new sealed
   run-level file and contract (D5/U2).

### 9.6 Owner decisions (2026-11-08)

Items 1-4 of section 9.5 are agreed; item 5 is decided before M1. The answers below override earlier sections where
they differ.

| # | Topic | Decision |
|---|---|---|
| O1 | Scenarios | All seven stories S0-S6 are agreed as stories and horizons. Pass numbers and seed counts are placeholders, revisited when M1 starts |
| O2 | Plant presets | Regimes A and B are presets of one plant; the synthetic numbers of section 1.1 live in the repo, labelled synthetic |
| O3 | Drive component | New `conveyor-drive` component (D1); `motor-drive@1` is not changed |
| O4 | Load sharing | Equal sharing (torque follower) is the default (D2); unequal sharing (U7) comes in M3 |
| O5 | Branches | One branch per milestone: M0 on `belt-m0-scenarios` cut from `belt-conveyor-drives`, each next one from the previous tip; merge to `main` only at the end |
| O6 | Inverter overload | The drive gets an inverter overload trip beside the motor relay: overload level and allowed time are preset values (heavy duty about 150 % for 60 s). Whichever protection trips first sets the survivors' clock. In M0 |
| O7 | Drive count | The number of drives is a setting from 2 to 6 (the 1/n curve covers 0-6). The stories use 3; a test proves a 2-drive belt. In M0 |
| O8 | Starting boost | The torque limit used to break a loaded belt loose (1.75x in regime B) holds only for a short boost time (a preset value, a few seconds), then the normal limit applies. Three drives still start a loaded belt; two still cannot. In M0 |
| O9 | Belt slowdown (L2) | Not in this plan: an overload still stops the belt after 3 s. Speed under a torque limit is built at the start of the Retrofit Designer milestone, the first thing that needs it |
| O10 | Seized drag | The drag of a seized drive is a preset value; 0 models a drive that disengages when it fails (one-way clutch). The S3 story uses the full drag. In M0 |
| O11 | Downstream stop | The belt gets a run permit from the next machine (a state). A session or scenario plan stops the belt by switching the permit off with a step fault at a set time, like any other fault; plans never send commands |
| O12 | Running in M0 | M0 runs every story as scenario files (`scenario-v0.1`): a do-nothing run and a hand-written reference-policy run. Live session plans are written and validated in M0 and run in tests from M2 |
| O13 | Owner review | No owner review between milestones: each milestone checks itself through its tests and gates. After M4, one full end-to-end test of all seven stories across the three repositories, and everything it finds is fixed |
| O14 | Expected numbers | Exit numbers come from the model's own equations, computed independently in the oracle tests, not from the table in section 1.1, which M0 corrects (238 s -> about 270 s with the current curve; 649 s and +509 s likewise), and again after O6. The 6-9 h bearing failure window stays, labelled an unpublished assumption |
| O15 | Outcome reference (D5/U2) | Computed at scoring time, not sealed as a file: the run's own artifact is replayed with every effector call removed. It is a pure function of inputs fixed before the run (domain, seed, planned faults), the consumer cannot influence it, and it never exists while the consumer runs. No new sealed file, no import-gate change, one path for scripted, plant-wire and director runs. Replaces "sealed run-level file" in D5, U2 and the M1 exit |
| O16 | Policy baselines (D6/U4) | Closed records in the domain's `ground_truth`: one condition on one state of a kind, an optional hold time, then a fixed list of effector calls with target and delay. No boolean composition, no expressions, no chaining. Played at scoring time like O15 (replay without the consumer's calls, policy applied). M1: B0 (do nothing) and B4 (today's PLC practice); B1-B3 in M4 |
| O17 | Command read-back (M2) | The feeder reports its feed setpoint back through a device interface file (a plant-wire kind file reached through a board mapping, as the fan does), so the host verifies every feed command; compositions gain an optional `implements` / `board_type` per equipment entry (generic). Rejected: leaving commands unconfirmed (verify-run would fail, the fan's W15 family) |
| O18 | M2 detection and test shape | A tripped drive is seen through its current (about 0 A), not its boolean run status, which agentic-stream operators cannot read yet; per-device readings (A1) are built in M2; the closed-loop test runs a shortened copy of the S2-B plan (trip at about 30 s, about 6 min wall) and leaves the committed plan as it is; the reasoner is a scripted stand-in, as in the fan tests; M2's reaction bound is measured from the model before it is fixed |
| O19 | Examples carry their kind files (M2, decided autonomously while the owner was away) | O17 made the shipped conveyor examples implement kind files that a fresh canvas workspace lacks, so a plant created from them could not compose. Fix: an example carries the kind files and boards its entries implement, and creating a plant from an example writes them to the workspace's `kinds/` (one component use case for CLI and canvas bridge). Rejected: default kind-file lookup beside a domain (hidden logic in three surfaces) and plain-plus-wired duplicate compositions with one domain id |
| O20 | Boot fencing with several devices on one entity (M2, decided autonomously while the owner was away) | agentic-stream fenced device boots per entity, so the five conveyor devices on `conveyor-1` displaced each other and only the last device's readings reached the situation. Fix: fence per (entity, input event type), correct with per-device readings (A1), no contract change; a site file that binds a kind-level reading to several devices of one entity is refused at load. Rejected for now: the device in every event's data (a contract change that moves the fan's event schemas and pins) |
| O21 | Read-back of stop and start commands (M3, autonomous) | A drive or belt stop could never be verified (no writable metric), so verify-run would fail. The conveyor's commands carry no id argument (the plant wire addresses the device); every commandable part reports its run command back as a writable metric reached through a board mapping, like the feeder's setpoint (O17). The feeder's fixed `F-1` id in its board goes away |
| O22 | R2 needs settled features (M3, autonomous) | agentic-stream denies R2 while the situation is `provisional`; M3 specs emit situation features `on_close` and a test proves an R2 intent passes |
| O23 | Barrier pacing for hour-long stories (M3, autonomous) | S1, S3 and S4 run barrier-paced (`--slice-s 60`, `--barrier-cap-ms 240000`, `horizon_s`); agentic-stream's trigger timers still count wall time (its world-clock branch is not merged), so every step of the plan has its own phase trigger and nothing relies on a re-trigger inside a cooldown |
| O24 | Scoring actions on other entities (M3, autonomous) | `streamsim score` matches a verdict action to a world call by command id, effector and entity, without requiring `issued_at` after the call; a fault may declare `expected_actions` on other kinds (`set_feed` on the feeder, `stop_drive` on the drive), resolved to instances in its label; multi-label runs grade in `report` too. A5 (time-to-threshold) is not needed in M3; Tamoz T1/T2 move to M4 |
| O25 | Two labels on one entity (M4, autonomous) | A label is keyed by entity and fault; a detection on an entity with several labels is matched to the label its diagnosis maps to (the verdict's subject map). One label per entity scores as before. `share_skew` stays out of the committed seed tables |
| O26 | Lead of a frozen sensor (M4, autonomous) | Counted from injection: a freeze has no `first_observable`; S5's order rule needs none |
| O27 | B1-B3 shapes (M4, autonomous) | B1 vibration > 4.5 mm/s for 60 s, feed 0, stop the matched drive after 60 s; B2 an absolute current threshold (1.1 x nominal-feed current) for 60 s, feed 67 % (departs: no rolling mean in the closed record); B3 always, the plant's safe limp feed (54 % A, 18 % B) |
| O28 | Arms (M4, autonomous) | A = rules only (specs and scripted stand-in), B = Tamoz, C = both; B and C need the live model and the owner |
| O29 | T3 side (M4, autonomous) | agentic-stream: a retryable failed terminal counts as an attempt, the next dispatch waits at least the worker lease, the failure reason is stored and shown |
| O30 | Playbook rule approval (M4, autonomous) | Learning runs use the scripted approver through the signed relay (labelled); the proposal waits for the owner's live approval |
| O31 | U9 consumer side (M4, autonomous) | A5 `extrapolate` is built in M4 so a consumer can publish a predicted failure time |
| O32 | Throughput before evidence (M4, autonomous) | A 16 h barrier-paced story must take under 15 min before the 10-seed evidence; measured unloaded at about 10.5 min (M3-DESIGN §8), so no change is needed |
| O33 | S6 plant (M4, autonomous) | The master-follower split is a separate plant variant per `master_loss` (`belt-conveyor-drives-s6-stop`, `-s6-promote`: plant A with D-1 as kind `drive_master` and D-2, D-3 as `drive_follower`), built on a new component version `streamsim/conveyor-drive@2` (`@1` plus the speed-reference link, parameters `role`, `master_loss`, `promote_s`). Plants A and B, their kind digests, stories, outcomes and the behaviour pin stay byte for byte. Rejected: splitting plants A and B (moves every conveyor story's kind digests and outcome digests for a story only S6 needs) and changing `@1` in place (changes both plants' composed domains) |
