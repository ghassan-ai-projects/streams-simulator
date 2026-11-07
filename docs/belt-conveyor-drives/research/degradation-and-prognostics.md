# Degradation, failure precursors and prognostics for conveyor drives (motors, gearboxes, VFDs)

Status: research note, 2026-11-07. Read with [../README.md](../README.md). No code.

## Reading guide and honesty about sources

- Search results were dominated by vendor blogs (fabrico.io, industrialmonitordirect.com, oxmaint.com, viox.com) that often
  paraphrase each other. Where I could not reach a primary source I say so. Numbers are tagged
  **[high] / [medium] / [low]** confidence. **[est]** = my own estimate or arithmetic, not from a source.
- Standards (ISO 20816-3, IEC 60947-4-1, IEC 60034-1, NEMA MG-1, ISO 281) are paywalled; figures below come from secondary
  summaries plus engineering memory and need checking against the standard before they become acceptance values.
- Nothing here is customer data. Everything is a default to be replaced by the customer's drive, motor and gearbox datasheets.

## 1. Failure mode statistics

### 1.1 Induction motors

| Survey | Population | Bearing | Stator winding | Rotor | Other |
|---|---|---|---|---|---|
| EPRI (1982; IEEE updates 1986-87) | about 4,800 utility motors, >= 100 hp | 41 % | 36 % | 9 % | 14 % |
| IEEE-IAS (Petro-chemical, 1985) | industrial/commercial, > 200 hp | 44 % | 26 % | 8 % | 22 % |

Confidence **[medium]**: the two tables are quoted identically in many secondary sources; I did not open the primary report
(EPRI EL-2678). Sources: https://reliamag.com/guides/electric-motor-failure-statistics/ ,
https://dspace5old.zcu.cz/bitstream/11025/50446/1/Leffler_Failures_of_Electrical_Machines_-_Review.pdf .
The commonly seen "37 % stator / 10 % rotor" variant is a one-point shift in some reproductions.

Takeaways:
- Bearings are the largest single share (about 40 %), stator insulation the second (about 25-35 %). Rotor faults are
  under 10 % (and mostly in large, frequently started or hot-cycled machines).
- These are utility/petrochemical motors with mixed duty. For a VFD-fed conveyor motor, additional contributors are
  inverter-induced bearing currents (EDM fluting) and voltage-stress on insulation (dv/dt). **[low, from engineering memory]**
- Failure *shares* are not detectability: bearing faults usually give a precursor (section 2); stator inter-turn and
  ground-wall failures often do not, or only minutes to days of it. **[est]**

### 1.2 Gearboxes

- Best public dataset is wind-turbine gearboxes (NREL Gearbox Reliability Database, 750 damage records 2009-2015): bearings 76 %,
  gears 17.1 %, other (mainly lubrication/filtration) 6.9 %. Sandia 2013 slides: bearings about 70 %, gears about 26 %.
  Confidence **[medium]**, and wind gearboxes are a poor proxy for a parallel/planetary conveyor drive
  (very different load spectrum). https://www.energy.gov/node/1246711 ,
  https://pt.slideshare.net/sandiaecis/sheng-glinsky-sandia-2013-wind-plant-reliability-workshop
- Swedish wind data: gearbox = 9.8 % of failures but 19.4 % of downtime (long repair). Same source as above family.
  Lesson for the simulation: a gearbox failure is rare per event but costly, and typically a bearing inside the gearbox first.
- For industrial/conveyor gearboxes I found no public share table. Use "bearing-dominated plus oil/seal/overheat" as a prior **[est]**.
- Conveyor-specific failure modes (my engineering knowledge, no source): loss of oil / overheated oil, input-shaft seal
  leak, backstop failure (on inclined belts), coupling failure (fluid coupling fusible plug, elastomer element wear), seized
  bearing leading to a locked output (the README's "drag" scenario).

### 1.3 VFDs

- No quantitative share found from a primary source in this pass. Commonly cited weak parts: DC-link electrolytic
  capacitors (wear-out, life halves per 10 C, same Arrhenius form), cooling fans, IGBT power cycling. **[low]**
- Drive trips (overcurrent, overload, over-temperature, DC bus, ground fault) are the *fast* failure class: a tripped
  drive is a step change, no ramp. That is the README scenario 2.

## 2. Precursor progressions and time scales

### 2.1 Bearing stages and the P-F interval

Order of detectability (consistent across sources): ultrasound / high-frequency (acoustic emission, enveloped
acceleration) -> vibration bearing-defect frequencies -> sidebands/harmonics, overall velocity RMS rises -> audible noise
-> temperature -> seizure. Sources: https://reliamag.com/articles/early-failure-detection/ ,
https://industrialmonitordirect.com/blogs/knowledgebase/predictive-maintenance-for-bearings-vibration-thermal-ultrasound-methods ,
https://www.isa.org/intech-home/2019/march-april/features/improving-maintenance-by-adopting-a-p-f-curve-meth .

Vendor-quoted "P-F" timings (oxmaint.com, vendor, **[low]**): vibration about months before failure, oil debris 3-6 months,
audible noise 1-4 weeks, heat 1-5 days. Reliamag **[medium]**: a high-speed cage failure may give only hours; slower
cage wear days to weeks with ultrasound. ISA states the P-F curve "is a conceptual representation only and does not have
any units or scale"; the interval varies with failure mode, speed, load and sensitivity. Rule of thumb: inspect/sample at
1/2 to 1/3 of the P-F interval, 1/4 to 1/5 for critical assets.

Important contradiction-in-waiting with the README: the README says "hours of precursor" for scenario 1. For a *slowly
rotating* conveyor drive the realistic envelope is wider: the **motor** bearing (1450-1800 rpm) behaves like the vendor
numbers above (days-weeks from first detectable defect to a hot bearing), but the whole range from detectable to failure
is not hours unless a lubrication loss or contamination event drives it. "Hours" is right for the *thermal* end of the
curve (temperature rising to trip). **Use both: a long weak early phase and a short steep late phase** (section 5.4).
Low-speed (output-shaft, < 60 rpm) gearbox bearings give almost no vibration precursor until damage is advanced. **[est]**

Staged model used widely (SKF-style, 4 stages; description paraphrased, **[medium]**): (1) ultrasonic/high-frequency
energy rises, no change in overall vibration; (2) bearing frequencies appear; (3) harmonics and sidebands, overall
velocity rises, temperature starts to move; (4) broadband noise floor rises, temperature and overall vibration rise
fast, then seizure. Progression is usually *accelerating*: most of the lifetime consumed in stage 1-2, the last 5-10 %
of life contains the visible heat. **[est]**

### 2.2 Run-to-failure evidence on timing (public datasets)

- IMS / NASA (Univ. Cincinnati): 3 run-to-failure tests, samples every 10 min, longest about 34.5 days.
  https://arxiv.org/pdf/2201.01769 . Detection of the outer-race fault in a comparable run-to-failure test: about 3.5-3.8
  running days via envelope/spectral methods; IMS set 1 outer race "for certain from day 28" (of about 34).
  https://arxiv.org/pdf/1511.03174 , https://mdpi.com/2076-3417/13/10/5977 **[medium, different tests mixed]**.
  So detectable-to-failure was days (tens of percent of life), then a short steep end phase.
- PRONOSTIA/FEMTO-ST: 17 accelerated tests, max runtime 126 min, lives vary by a wide spread; bearing failure type is not
  labelled; some test-set bearings are truncated before failure. https://arxiv.org/pdf/2201.01769 ,
  https://yanncalec.github.io/dpmhm/datasets/femto , https://arxiv.org/pdf/2109.11579 .
- Sparsity/health indices on IMS: classical features (RMS, kurtosis) show the trend but not monotonically; engineered
  indices are smoother. https://arxiv.org/pdf/2205.00178 . One RMS-trend extrapolation study reported 8.6 % error in
  predicted failure time (in-sample, so optimistic) **[low]**.

### 2.3 Vibration severity (ISO 10816-3, now ISO 20816-3)

Group 2 (15-300 kW or shaft height 160-315 mm), RMS velocity 10-1000 Hz, **[medium]** from vendor summary
https://www.fabrico.io/blog/iso-10816-3-vibration-severity/ and Siemens/DIN note
https://cache.industry.siemens.com/dl/files/877/38709877/att_108911/v1/38709877_76.pdf :

| Zone | Meaning | Rigid support (mm/s RMS) | Flexible support (mm/s RMS) |
|---|---|---|---|
| A | new machine | < 1.4 | < 2.3 |
| B | unrestricted long-term | 1.4 - 2.8 | 2.3 - 4.5 |
| C | restricted/short-term, plan maintenance | 2.8 - 4.5 | 4.5 - 7.1 |
| D | damage likely | > 4.5 | > 7.1 |

Larger machines (Group 1, > 300 kW) have higher limits (about 2.3 / 4.5 / 7.1 rigid) **[low, memory]**. The standard
also defines a *change* criterion: a sustained step of more than about 25 % of the zone B/C boundary is
significant regardless of absolute level **[low, memory]**. Equation: v_rms = sqrt( (1/T) * integral(v(t)^2 dt) ), units mm/s.

### 2.4 Temperature

- Bearing temperature: normal 50-80 C operating; alarm often +20 K over baseline, trip about 90-100 C for greased motor
  bearings **[low, est]**. It is a late indicator (hours to days before seizure, https://industrialmonitordirect.com/blogs/knowledgebase/predictive-maintenance-for-bearings-vibration-thermal-ultrasound-methods).
- Winding temperature (PT100 or PTC in slots) tracks load (I^2) with thermal lag; a rise at *constant* current means
  cooling loss, fouling, bearing friction, or inter-turn fault. See section 3.
- Gearbox oil sump temperature: slow (tens of minutes to hours time constant) **[est]**; a rise at constant load =
  oil level/viscosity/bearing friction.

### 2.5 Current signature (MCSA) and current unbalance

- Broken rotor bars: sidebands at f_b = (1 +/- 2s) * f_line (f_line Hz, s slip, unitless). Needs near-rated load: at light
  load s -> 0 and sidebands merge with the line peak. Progression to multiple bars and cage failure takes months.
  https://www.fabrico.io/tr/blog/broken-rotor-bar-detection/ , https://tractian.com/en/glossary/motor-current-signature-analysis **[medium]**.
- Bearing faults appear in current at f_line +/- m * f_defect (BPFO, BPFI, BSF, FTF) but MCSA is less sensitive than
  vibration for bearings, especially above 100 kW. Same sources. Vendor dB thresholds disagree (> -54 dB "monitor",
  > -40 dB "urgent", > -30 dB severe, relative to fundamental); calibrate per motor, do not hard-code **[low]**.
- Under VFD supply the line frequency is variable and the spectrum carries PWM harmonics; MCSA on VFD-fed motors is
  materially harder. The VFD's own estimated torque/current is the practical channel. **[est]**
- Current unbalance: voltage unbalance of 1 % can give about 4-10 % current unbalance (blogs; ratio 5-10x depends on
  load and impedance) **[low]**: https://industrialmonitordirect.com/blogs/knowledgebase/three-phase-motor-current-unbalance-monitoring-derating .
  NEMA MG-1: operate within 1 % voltage unbalance at rated load; derate above, about 0.75 at 5 % (published factors vary,
  0.76-0.92 at 2-5 %). A *developing stator inter-turn fault* shows as growing negative-sequence current at constant supply.
  Unbalance % = 100 * max|I_k - I_avg| / I_avg.

### 2.6 How much warning is typical (summary)

| Fault | Typical warning from first reliable detection to failure | Confidence |
|---|---|---|
| Motor bearing, lubrication/wear (spall growth) | days to weeks (hours at high speed/cage) | medium |
| Motor bearing, contamination or lube loss | hours to days | low [est] |
| Gearbox gear/bearing pitting | weeks to months; hours at the very end | low [est] |
| Stator insulation, gradual (thermal aging, partial discharge) | months, rarely observable on conveyor sensors | low |
| Stator inter-turn short | minutes to days | low [est] |
| Rotor bar | months | medium |
| VFD trip (IGBT, DC bus, ground fault) | none | medium |
| Fan/cooling loss | tens of minutes to hours via temperature | est |
| Overload from material/feed | immediate, no degradation | n/a |

## 3. Thermal models, insulation life, protection curves

### 3.1 Insulation classes and the 10 degree rule

- IEC 60034-1 / NEMA classes (max winding temperature): B 130 C, F 155 C, H 180 C (A 105 C). Class F dominates; common
  practice is class F insulation with class B temperature rise ("about 25 K margin").
  https://www.fabrico.io/blog/motor-insulation-classes/ **[medium]**.
- Arrhenius: aging rate k = A * exp(-E_a / (k_B * T)), T in kelvin. Over normal ranges this is approximated by the
  "10 degree rule": life L(T) = L_0 * 2^((T_0 - T)/10) with T, T_0 in C. Equivalent aging rate (relative to design)
  a(T) = 2^((T - T_0)/10). Rated life of order 20,000 h at class limit (IEEE 117 type tests; I recall 20,000 h as the common
  basis; **[low]**). Sources: same fabrico page; one blog table is internally inconsistent, so stick to the halving rule
  rather than any published percent table. Real insulation systems have halving intervals of 8-12 K (E_a dependent) **[est]**.
- Cumulative damage (Miner / life fraction): D = integral( a(T(t)) dt ) / L_design, failure at D = 1. This is the
  README's "cumulative, not threshold" statement, correct for insulation.

### 3.2 Winding thermal model (first-order, lumped)

Equations (units: K, s, A):
- Steady rise: dT_ss = dT_fe + dT_cu * (I/I_r)^2 where dT_fe is the iron/friction part and dT_cu the copper part of the
  rated rise. Typical split copper:iron about 70:30 **[est]**.
- Dynamics: tau * dT/dt = (T_amb + dT_ss) - T, i.e. T(t) = T_ss - (T_ss - T_0) * exp(-t/tau).
- Time constant tau: manufacturers rarely publish it. Relay-modelling papers use 1370 s and 3720 s (stator, SF 1.15);
  protection relays allow 1-60 min; knowledge-base articles quote 5-30 min for large motors.
  https://selinc.com/api/download/2600 (Schweitzer paper), https://industrialmonitordirect.com/blogs/knowledgebase/motor-overload-duration-thermal-time-constant-calculation
  **[medium for range]**. Default for a 55-250 kW fan-cooled TEFC motor: winding tau about 15-25 min, frame about 60-90
  min **[est]**. Stalled-rotor/locked time (t_E, 5-15 s) is a separate, much faster, rotor-bar limit.
- Rotor losses grow as slip rises with load; the (I/I_r)^2 approximation underestimates heating at high load for
  line-fed motors, and is roughly right for VFD vector control. **[est]**
- Cooling: a shaft-mounted fan loses airflow at low speed. If the VFD slows the belt, cooling falls with speed (about
  N^2-N^3 for flow-driven cooling) **[est]**. At constant torque this cuts the allowed continuous torque
  (typical VFD-duty derating curves) - important for a "run slower to protect survivors" reaction.

### 3.3 Service factor

NEMA SF 1.15 = permitted occasional load 1.15 x rated with winding rise allowed about +10 K above rated-load rise; SF
is "for short-term or occasional overloads", and running there shortens life. IEC motors usually have SF 1.0; many IEC
datasheets quote S1 continuous only. https://industrialmonitordirect.com/blogs/knowledgebase/running-ac-induction-motors-at-service-factor-life-expectancy-analysis
(blog) **[medium]**; datasheet example 80 K rise at SF 1.15 vs 105 K at SF 1.0 (class F). Sources from the earlier search.

### 3.4 I^2t overload relay (IEC 60947-4-1) and VFD overload

- Trip classes defined at 7.2 x I_r from cold, trip-time windows (IEC 60947-4-1 Table 2 per Schneider community,
  **[medium]**): class 5: 0.5-5 s; 10A: 2-10 s; 10: 4-10 s; 20: 6-20 s; 30: 9-30 s.
  https://community.se.com/t5/Expert-on-Demand-for-Standards/Di-Thermal-Overload-Relay-TOR-ada-spesifikasi-class-10-20-dan-30/m-p/377408
- Other points of the standard (from engineering memory, **[low-medium]**, not confirmed in this session): no trip within
  2 h at 1.05 x I_r cold; trip within 2 h at 1.2 x I_r (hot); at 1.5 x I_r hot: class 10A < 2 min, 10 < 4 min,
  20 < 8 min, 30 < 12 min. The viox.com page lists 10A <= 2 min at 1.5x.
- Model (fits the above): thermal capacity relay, with I_p = pre-load, I_th = trip threshold current (about 1.05-1.1 I_r):
  t_trip = tau_r * ln( (I^2 - I_p^2) / (I^2 - I_th^2) ), I in per-unit of I_r, tau_r in s.
  Fit tau_r to the class: class 20 at 7.2x from cold (I_p = 0) -> t = 20 s -> tau_r = 20 / ln(51.84/(51.84-1.1)) about
  1000 s **[est, my arithmetic]**. Then at 1.4 x I_r from cold: about 12 min; from hot, with I_p = 0.7 (survivor sharing
  3 drives at 70 % before the loss): t about 9 min **[est]**. So a survivor carrying 1.4x would be tripped by an
  I^2t relay in **under 10-15 min**, long before the damage the README fears accumulates - unless the relay is set high
  or replaced by VFD-limited operation. **This is a key correction: the survivor either trips quickly or the VFD current
  limit holds it; "hours of overload" happen only if protection is permissive (or the trip is ignored).**
- VFD overload: ratings commonly "normal duty 110 % for 60 s per 5 min; heavy duty 150 % for 60 s per 5 min" and an I^2t
  model on the inverter output, plus motor thermal model trip. **[low, memory]**; take the real values from the customer's drive
  manual. Typical drive current limit 150-200 % transient.

### 3.5 Hot-spot to winding offset

Winding temperature measured by embedded PT100 reads average/near hot spot; hot spot typically 5-15 K above the
measured sensor reading **[est]**. Use a hot-spot margin of +10 K when computing aging.

## 4. Damage accumulation for the survivors

### 4.1 Insulation, worked estimate **[est, my arithmetic]**

Assumptions: class F, rated winding rise 80 K, ambient 40 C, rated hot-spot = 40 + 80 + 10 = 130 C (design point,
aging rate 1), copper share of rise 70 % (56 K), iron 24 K. Steady state at load L (per unit of rated torque ~ I):

| Load L | dT_ss = 24 + 56 L^2 (K) | Hot-spot (C) = 40 + dT_ss + 10 | Aging rate a = 2^((dT_ss - 80)/10) | 1 h at that load consumes |
|---|---|---|---|---|
| 0.7 (3 drives sharing) | 51 | 101 | 0.14 | 0.14 h |
| 1.0 | 80 | 120 | 1 (by definition) | 1 h |
| 1.15 | 98 | 148 | 3.5 | 3.5 h |
| 1.3 | 119 | 169 | 14.5 | 14.5 h |
| 1.4 | 134 | 184 | 42 | 42 h |
| 1.5 | 150 | 200 | 128 | 128 h |

(Hot-spot above 155 C
exceeds the class F limit, so the 10 K rule gets less reliable and a safety trip (PTC at 155/160 C) comes first.
Equivalent wording: **1 h at 1.3x costs about 14 h of life, 1 h at 1.5x about 128 h** if steady state is reached, which
takes 3 tau (about 45-75 min) because of the thermal lag; a 15 min overload at 1.4x reaches only about 50 % of the
steady rise. Integrate T(t), not I(t). If designed with class B rise and F insulation, the margin of 25 K divides the
rates above by about 5.6.

### 4.2 Bearings

L10 = (C/P)^p million revolutions, p = 3 for ball bearings, 10/3 for roller bearings (ISO 281);
L10h = 10^6 / (60 n) * (C/P)^p h, n in rpm. https://www.fabrico.io/blog/l10-bearing-life/ **[medium]**.
Life ratio = (P_new/P_old)^-p: 1.3x load -> 0.46 (life 2.2x shorter), 1.5x -> 0.30, 2x -> 0.125 (ball, **[est]**
arithmetic). But motor bearing P is dominated by rotor weight, belt/pulley/gearbox side loads and preload, not by
torque. Higher torque affects the *coupled/output* bearings through gear tooth force (gearbox) and belt pull. So the
torque-to-bearing-load coupling is partial. **[est]** Cumulative form: Miner D = sum( t_i / L_i ), with L_i = L_10 *
(P_r/P_i)^p.

### 4.3 Gearbox gears

Gear contact stress sigma_H ~ sqrt(F_t) (F_t tangential force, proportional to torque); endurance S-N slope for
pitting is steep. Life exponent on torque about 5-8 (derived from sigma^-(10-16) and sigma ~ T^0.5; ISO 6336 / AGMA type;
**[low, est]**) so 1.3x torque -> life 0.13-0.27 of nominal for pitting; bending fatigue even more sensitive. AGMA service
factors (1.25-1.5 for conveyors) exist because of this. Conveyor gearboxes are sized with service factor >= 1.25-1.5 so a
temporary 1.4x on a healthy gearbox is within design margin for transient, not for sustained operation **[est]**.

### 4.4 Other survivor effects

- Repeated overload trips and restarts: motor starting thermal stress; each start from standstill under belt load
  (README restart trap) gives locked-rotor heating (stator at about 6 x I_r for seconds, with VFD limited to about 1.5-2 x).
- VFD: junction temperature swings (power cycling) shorten IGBT life by a power law (Coffin-Manson, exponent about 3-5 of
  delta-T_j, **[low]**); overload increases delta-T_j.
- So *do not* model the survivor's damage as only thermal: split it into winding aging (Arrhenius), bearing and gear
  fatigue (power law), plus an immediate protective-trip channel. Real outcome distribution: most survivor "damage"
  appears as protective trips and shortened life, not immediate failure.

## 5. Remaining useful life methods for streaming data

### 5.1 Hierarchy of practical methods

1. **Threshold / zone alarm** on a health indicator: zone C, T > limit, unbalance > x %. Cheapest; lead time is whatever
   the margin gives. (Used by all products, section 6.2.)
2. **Trend extrapolation**: fit a robust line or exponential to a health index over a sliding window, RUL = (threshold -
   x_now) / slope (linear) or ln(threshold / x_now) / b (exponential). Needs a threshold, a smooth index, and a regime
   where the trend stays in form.
3. **Health index fusion**: combine several features (RMS, kurtosis, envelope band energy, temperature rise over load
   baseline) into 1 index; engineered indices are more monotonic than raw RMS/kurtosis (https://arxiv.org/pdf/2205.00178).
4. **Statistical degradation models**: Wiener process with drift: X(t) = x_0 + mu * t + sigma * B(t); first passage to a
   threshold w follows an inverse Gaussian: f(t) = w / sqrt(2 pi sigma^2 t^3) * exp( -(w - mu t)^2 / (2 sigma^2 t) ).
   Mean time to threshold (w - x_0)/mu, variance (w - x_0) sigma^2 / mu^3. Review: Zhang, Si, Hu, Lei, EJOR 271(3), 2018
   (doi 10.1016/j.ejor.2018.02.033); Si et al., EJOR 213(1), 2011 (doi 10.1016/j.ejor.2010.11.018) - cited from the
   bibliographic listings, content not read. Gamma process (monotone) and exponential-growth (log-linear Wiener) are the
   other standard forms; Bayesian updating with random-effects drift handles unit-to-unit spread.
5. **Particle filter / Kalman** on a state model of fault size (Paris-law crack growth, Li-Kang spall growth):
   da/dN = C (delta-K)^m (Paris law, a crack length, N cycles). Requires physics calibration.
6. **Data-driven deep RUL** (LSTM, CNN) on PRONOSTIA/IMS: strong on benchmarks, fragile out of distribution and
   needs many run-to-failure examples; not needed for a simulator's ground-truth oracle.

### 5.2 Exponential bearing health index (typical behaviour)

Empirical: health index h(t) = h_0 + a * (exp(b (t - t_0)) - 1) for t > t_0 (onset), noise lognormal, with abrupt jumps
when a spall propagates. Time-to-threshold from onset T_f = ln(1 + (h_f - h_0)/a)/b. Parameters are fit from run-to-failure
data (section 8); I have no validated population values. RUL error from naive linear extrapolation is large when the
true curve is accelerating. **[est]**

### 5.3 PHM performance metrics (Saxena, Celaya, Goebel et al.)

Source: Saxena et al., "Metrics for evaluating performance of prognostic techniques" (PHM 2008) and "On Applying the
Prognostic Performance Metrics" (PHM Society 2009); NASA Prognostics Metrics Library.
https://github.com/nasa/prognosticsmetricslibrary , https://ntrs.nasa.gov/citations/20100023445 ,
https://papers.phmsociety.org/index.php/phmconf/article/view/1621 . **[high for definitions]**

- **Prognostic horizon (PH)**: PH = t_EoL - t_i where t_i is the first time the predicted RUL enters the accuracy band
  (+/- alpha * t_EoL around the true RUL, typical alpha = 0.05-0.2) with probability >= beta and stays there.
- **alpha-lambda performance**: binary: at fractions lambda of the way from first detection to EoL (e.g. lambda = 0.5, 0.75),
  does predicted RUL lie within the cone [(1 - alpha) RUL*, (1 + alpha) RUL*], with probability mass >= beta.
- **Relative accuracy (RA)**: RA(t_lambda) = 1 - |RUL*(t_lambda) - RUL_hat(t_lambda)| / RUL*(t_lambda).
- **Cumulative relative accuracy** and **convergence** (distance from the origin to the centroid of the area under a
  metric curve; smaller means faster convergence).
- Detection-side metrics (not in the PHM set but standard): false-alarm rate per unit time, missed-detection rate,
  **lead time** (alarm to failure), earliest correct alarm, and for repeated events, the lead-time distribution.
  Often pared with cost-based scoring (asymmetric: late prediction costs more than early; e.g. the C-MAPSS scoring
  function, exponential penalty with different slopes, **[medium, memory]**).

### 5.4 Use in the simulator

Truth = known degradation parameters and a known failure time per drive (generative). Grade the consumer on lead time
(alarm to failure), false alarms (on healthy drives and on feed surges), RUL error vs ground truth at lambda fractions,
and *reaction* quality (survivor damage avoided). Sealed ground truth must include onset time, health trajectory, and
threshold crossing.

## 6. Fault-tolerant and coordinated operation of multi-motor conveyors

### 6.1 Control architectures

- Master-slave torque following: one drive in speed control, others in torque control from the master's torque reference
  (analog or fieldbus). Common on long conveyors. https://invertekdrives.com/support/iknow/vfd-fundamentals/torque-control-and-master-slave-operation-15 ;
  load sharing via slip (droop) alternative: sharing depends on slip match (5 % vs 5.5 % example), fragile.
  https://industrialmonitordirect.com/blogs/knowledgebase/multi-motor-conveyor-drive-systems-load-sharing-configuration
  (blog, **[low-medium]**). Practitioners report master-slave working "quite well" and slip sharing hard to tune
  (https://control.com/forums/threads/conveyor-load-sharing-using-ac-drive.7089/post-7089).
- Belt-coupled drives must have consistent speed, else belt tension distribution is inconsistent; coordinated MPC
  proposed for dual PM drives in coal mining belts. https://eprints.whiterose.ac.uk/206718/1/MPC-Based_Coordination_Control_of_Dual_Direct-Drive_Permanent_Magnet_Motors_Used_in_Coal_Mining_Belt_Conveyors.pdf
- Master-slave torque following has a failure sensitivity: if the master (speed-control) drive trips, the torque followers
  have no speed reference. Mission-critical designs reassign the master role or fall back to common speed reference
  **[est, engineering knowledge]**.

### 6.2 What deployed products actually do

- ABB Ability Smart Sensor (motor-mounted, BLE/WirelessHART, IP66/67, ATEX): measures vibration, surface temperature,
  (magnetic field for rotor/stator), default one measurement per hour; computes overall vibration, bearing condition,
  misalignment, unbalance, looseness, rotor winding health as traffic lights; alerts by app/email/webhook; regreasing
  reminders. https://new.abb.com/motors-generators/service/advanced-services/smart-sensor ,
  https://www.baldor.com/brands/baldor-reliance/advanced-services/smart-sensors-for-motors/ **[medium]**. It *alarms*; it
  does not act on the process.
- Siemens SIMOTICS CONNECT 400, SKF Enlight / Rockwell condition monitoring: not retrieved in this pass; known from memory
  to be the same class (sensor, KPI, threshold and trend alerts, cloud analytics), **[low]**.
- VFD built-ins (ABB ACS, Siemens, Rockwell PowerFlex): motor thermal model, I^2t overload, torque limit, stall protection,
  load-curve / "underload/overload" monitoring, "ride-through" on loss of supply, and fault-reaction choices
  (ramp stop, coast, fault). **[low, memory]**.
- None found: an industrial product that, on predicted failure of drive k, automatically redistributes load, slows or
  starves the belt and *notifies the other drives*. In practice this is done in the PLC/DCS (conveyor control) with fixed
  logic: on drive fault -> reduce belt speed or inhibit feeder, trip chute interlock, "reduced capacity mode" with a
  derated feed rate. The simulation use case is therefore genuinely "new coordination logic" not an existing product feature.
  **[medium; absence-of-evidence from a limited search]**

### 6.3 Reaction levers (engineering reasoning, **[est]**)

| Lever | Effect | Cost / risk |
|---|---|---|
| Reduce feed rate (feeder) | Load falls after the belt-transit delay (belt length / speed) | Throughput loss |
| Lower belt speed at same feed | Torque per metre of belt rises (more material per metre); power = F*v falls | Spillage, if feed not cut too |
| Torque bias (shift share) | Gives the degrading drive less, survivors more | Survivors' headroom |
| Stop failing drive in a planned way | Replaces an uncontrolled trip, may avoid a drag if the gearbox is seizing | Restart trap (needs a free belt) |
| Run belt empty then stop (clear the belt) | Avoids the loaded-restart trap | Takes belt transit time (60 m / 2 m/s = 30 s) |
| Alarm only | No action | Operator latency |

Belt-pull basis (DIN 22101 / CEMA, **[medium, memory]**): F_U = C * f * L * g * [ m_R' + (2 m_B' + m_L') cos(delta) ] +
m_L' * g * H (N), C secondary resistance factor (1.05-1.5 for long belts, larger for short), f artificial friction
factor (about 0.02), L length (m), m' masses per metre (kg/m), H lift (m). Drive power P = F_U * v / eta. Another
research file will cover this; here only note the load is quasi-static, so overload on survivors is tied to feed history.

## 7. Sensor failure modes that mask degradation

Taxonomy used in NASA fault modeling: bias (offset), drift (time-varying offset), scaling (gain), dropout (zero or no data),
stuck-at (frozen), intermittent (noise bursts). Y_f = X + S(t) + noise for drift.
https://ntrs.nasa.gov/citations/20205008007 , https://ntrs.nasa.gov/api/citations/20110016122/downloads/20110016122.pdf **[medium]**.

Vibration (accelerometer) specific:
- Detachment / loosened mounting: resonance shift, high-frequency attenuation, bearing signature disappears; looks like
  "a healthy quiet machine". (NASA SDA work monitors baseline resonant frequency to detect attachment health.)
- Cable / connector open circuit: flat line at 0 or at bias voltage; short: saturated.
- Clipping: signal larger than dynamic range truncated, RMS plateaus while real severity rises.
- Environmental: moisture, heat, vibration-induced bias drift and noise growth in MEMS.
- Smart wireless sensors (ABB): battery depletion, slow sample rate (hourly), missed packets - a time-sampling masking
  effect rather than a signal fault.

RTD / PT100 / thermocouple (from engineering knowledge, no source found this pass, **[low]**): lead-wire resistance
and connector corrosion add positive bias (2/3-wire: about 0.4 C/ohm for PT100), open circuit reads "over-range" or
"under-range" (system dependent), moisture drift, thermal contact loss giving slow response (large lag, under-reading hot
spot), slow calibration drift of order 0.05-0.1 C per year (low).

How faults mask degradation (design for the sim):
- A stuck channel at a healthy value hides a rising trend; a dropped channel gives a flat or NaN series that a consumer
  may treat as "no alarm".
- A degrading machine can also kill its own sensor (heat, vibration), correlating failures (common cause). **[est]**
- A slow drift on temperature offsets absolute thresholds but not relative (vs. peers, vs. load model) comparisons.
- Cross-sensor consistency is the main defence: three drives share load, so currents, speeds and temperatures should
  agree up to the sharing law; vibration on healthy peers gives a baseline. Residual vs. peer is a good health index.

## 8. Public datasets for calibration and validation

| Dataset | Content | Use here | Limits |
|---|---|---|---|
| CWRU Bearing Data Center | seeded (EDM) faults, 12k/48k Hz, 4 loads | diagnostic classification of fault frequencies, amplitude vs defect size | not run-to-failure; artificial faults |
| IMS / NASA (Univ. Cincinnati) | 3 run-to-failure tests, 4 bearings, 20 kHz, 1-s snapshot every 10 min, up to about 34.5 d | shape of late-life trend, onset-to-failure time scale | 3 tests, loads constant, failure mode seen in only a few bearings; snapshots sparse |
| PRONOSTIA / FEMTO-ST (IEEE PHM 2012 challenge) | 17 accelerated run-to-failure, vibration + temperature, 3 operating conditions, up to about 2 h | spread of lives, RUL scoring | accelerated; failure type unlabeled; some test sets truncated |
| NASA Prognostics Center repository | many sets (bearings, batteries, C-MAPSS turbofan) | metrics code, RUL benchmarks | not conveyors |
| XJTU-SY, Paderborn, MFPT, Ottawa (other) | run-to-failure and real/artificial damage | extras; I have not verified details | **[low, memory]** |

Sources: https://data.phmsociety.org/nasa/ , https://github.com/VictorBauler/awesome-bearing-dataset ,
https://mdpi.com/2624-6120/2/4/40 (review: few run-to-failure examples, wide spread of lives, no failure-mode labels).
No public dataset found with *induction-motor winding temperature degradation*, *gearbox on belt-conveyor duty*, or
*multi-drive load redistribution events*; those would have to come from the customer.

What the datasets can calibrate: (a) shape of the health-index curve (long flat, then exponential), (b) ratio of lifetime
spent before detectable onset, (c) noise level and non-monotonic jumps, (d) spread of life (Weibull shape about 1.5-3,
IMS/PRONOSTIA lives vary by order of magnitude **[est]**). They do not calibrate absolute time scales for a 60 m conveyor
drive; scale time to the customer's operating hours.

## Implications for the simulation

### Degradation models to use (each as data, not code branches, per AGENTS.md)

1. **Bearing wear (motor and gearbox)**: hidden health h in [0,1] consumed at rate r(t) with a latent onset time. Model
   the *observable* features as a function of h: vibration RMS v = v_0 * (1 + k1 * exp(b * (1 - h))); temperature rise
   delta_T_b = dT_0 + k2 * (1 - h)^2 + load term; current ripple at defect frequency small. Fault life: Weibull(shape
   2, scale per drive); onset at 60-90 % of life then exponential growth. Load coupling: effective consumption rate scaled by
   (torque / torque_rated)^p with p = 3 for motor/ball bearing, p about 5-6 for gears.
2. **Winding insulation aging**: lumped thermal model (section 3.2) driving D = integral a(T) dt / L, a = 2^((T - T_0)/10);
   failure at D = 1. Produces the survivor's cumulative damage; also winding-temperature observations. Optional
   inter-turn fault as a hazard: rate multiplied by a(T), no precursor except growing negative-sequence current.
3. **Protective trips as a first-class channel**: I^2t relay model (class parameter), VFD overload (e.g. 110 % / 60 s), PTC
   trip at 155 C, torque limit. These explain why survivors trip faster than they age.
4. **Sudden failures**: VFD/drive trip, coupling break, seized gearbox as step events with configurable zero-lead-time.
   The seized gearbox adds drag torque on the belt (README scenario 3).
5. **Thermal cooling loss** (fan fouling/blockage): tau and dT_ss shift; precursor of tens of minutes.
6. **Sensor fault overlay** (stuck, dropout, drift, bias, clipping, detachment) independent of the machine state, with
   a correlation option (sensor fails because machine fails).
7. **Stationary regime for the feed**: material on belt is a delayed copy of feed (transit = L / v about 30 s), so the load
   response to a feed change lags by up to transit time - a surge raises all three currents together (README scenario 4).

### Time scales for scenarios (defaults, all **[est]** unless noted)

| Scenario | Time scale |
|---|---|
| Slow motor-bearing failure | simulate 3-14 days wall-equivalent compressed; detectable onset at day 3-10 of 14; steep final phase 4-12 h of rising temperature; failure/trip at end |
| Fast bearing (lube loss) | onset to trip 2-12 h |
| Survivors after a loss | thermal tau 15-25 min; I^2t trip in 5-15 min at 1.3-1.5x; aging rate x14 at 1.3x, x128 at 1.5x |
| Gearbox oil overheat | 1-4 h from alarm to seizure |
| Seizure | step; drag within seconds |
| Sampling | vibration RMS every 1-60 s (smart sensor 1/h default is a *limit*); temperature 1-10 s; current 0.1-1 s |

Compress days into minutes only with a clock-rate factor; keep the chain onset -> alarm -> reaction -> failure order and
the ratios (thermal tau vs. precursor) identical to avoid unrealistic lead times.

### Parameter defaults (replace with customer datasheets)

- Motors: 160 kW class (power = F_U v / eta, to be sized from the conveyor note), class F insulation, rated rise 80 K,
  ambient 35-40 C, copper fraction 0.7, winding tau 20 min, frame tau 75 min, SF 1.0, I^2t class 10 or 20 at the
  relay, VFD overload 110 % / 60 s.
- Bearings: ball bearing p = 3; L10 = 40,000-100,000 h; Weibull shape 1.5-3 for individual failure time; precursor
  fraction 60-90 % of life in the pre-onset phase.
- Gearbox: sump tau 30-60 min; seizure hazard multiplied by (T_oil / T_rated)^n; service factor 1.25-1.5.
- Vibration baseline 1.0-1.8 mm/s RMS (zone A/B), alarm B/C at 2.8, trip at 4.5 (Group 2, rigid) or 7.1 (flexible).
- Health index noise: lognormal sigma 10-20 % on RMS; kurtosis baseline 3.

### Grading metrics (offline, using sealed truth)

- **Lead time** of the first correct alarm for the failing drive (hours), distribution over seeds; PHM prognostic horizon
  with alpha = 0.2, beta = 0.5.
- **False alarms**: per 1000 h on healthy drives, and *specifically at feed surges* (scenario 4) and sensor faults.
- **Missed events**: sudden trips cannot be predicted - grade "reaction time after the trip" instead of lead time.
- **RUL accuracy**: relative accuracy at lambda = 0.5, 0.75; cumulative RA; convergence.
- **Consequence metrics** (what the customer cares about): survivor life consumed (insulation D, bearing Miner) during the
  event; number of protective trips; peak per-drive load; belt-stop count and loaded restarts; throughput lost; time with
  any drive above 1.2x.
- **Trivial-baseline audit** (non-negotiable): fixed-threshold-on-vibration, "always shed load", and "never react"
  baselines; the consumer must beat them. A baseline that alarms on any current rise should fail the feed-surge scenario.
- **Silent no-effect test**: every reaction command (feed cut, torque bias, stop) must demonstrably change survivor load
  in the sim, otherwise it's flagged.
- Digital-twin cross-check: analytic solution of the first-order thermal ODE (section 3.2) and of the Wiener first-passage
  distribution as independent oracles for the thermal and RUL simulation code.

## Contradictions and corrections to the README's working understanding

1. "Survivors age faster, long-run damage" - partly true but the **time scale is minutes, not days**: at 1.3-1.5x a
   standard I^2t relay trips in about 5-15 min, and insulation aging runs x14 (1.3x) to x128 (1.5x) while it lasts. The
   dominant survivor outcome is therefore a *cascading trip* (second drive), unless the protection is permissive. **[est]**
2. "The failing motor shows degradation before it stops" holds for bearings (about 40 % of failures) and thermal paths
   but **not for sudden stator or VFD faults**; the simulator needs a realistic share of unpredictable failures.
3. Bearing life is a cube-law in *bearing load*, but motor-torque coupling to bearing load is weak; gears and insulation are
   the torque-sensitive parts. Bearings are mostly a cause (the failing drive) rather than a victim.
4. Precursors span days to weeks for slow bearing wear; "hours" applies only to the thermal end or to
   lubrication-loss events. Calibrate scenario 1 as multi-phase.
5. No off-the-shelf product coordinates the drives on predicted failure; ABB-class monitoring only alarms.
6. Torque master-slave: losing the speed master is itself a failure mode to include.
7. Slowing the belt at constant feed *raises* load per metre; and fan-cooled motors lose cooling at low speed, so
   "slow down" is not automatically protective.

## Open questions for the customer

1. Motor rating, insulation class, rated rise, thermal class, SF, cooling (TEFC/forced), PTC/PT100 setting; motor
   thermal time constant or heat-run data; VFD make and model with its overload/torque-limit configuration.
2. Overload protection actually configured: relay class, thermal model, trip temperatures, VFD overload (110/150 %).
3. Historic failures: which components, how often, what warning was seen and when (hours/days), and which sensors showed it.
4. What happened in past single-drive losses: did survivors trip, at what load, how long did they run, any damage found?
5. Sensors: vibration (type, location, sampling interval, wireless?), winding RTD (how many per motor), bearing
   temperature, gearbox oil temperature/level, torque signal source (VFD estimate vs. meter), sampling and logging rates.
6. Gearbox type (helical/planetary), service factor, oil system, backstop, fluid couplings.
7. Current load-sharing mode (master-slave torque, droop), what happens to the speed reference if the master trips.
8. Allowed automatic actions and latency (feed cut, belt speed, torque bias, planned stop), permission model.
9. Cost model: cost of a loaded restart failure, of belt stop, of a damaged motor, of lost throughput per hour.
10. Duty: hours per year, starts per day, ambient range, dust/fouling, VFD carrier frequency and cable length (bearing currents).
11. Will they give access to historic data to calibrate Weibull/degradation parameters (beyond public bearing datasets)?
