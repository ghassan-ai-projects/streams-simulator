# Multi-drive belt conveyor: mechanics and operation (research notes)

Date: 2026-11-07. Scope: equations, numbers and failure behaviour needed to build a defensible simulator of a ~60 m, three-drive belt conveyor
(see `../README.md`). Confidence tags: **[H]** consistent across several sources or textbook physics, **[M]** one decent source or
standard engineering practice, **[L]** forum/vendor claim or weak. **[EST]** = my own estimate or derivation, not from a source.

Honest limits of this research: the DIN 22101 / ISO 5048 / CEMA texts are paywalled; the equations below come from secondary sources
(Overland Conveyor "Belt Analyst" help pages, vendor glossaries, lecture decks, papers) and from textbook mechanics. No public
case study of "one of N drives failed, survivors overloaded" with numbers was found. Items that could not be sourced are flagged.

---

## 1. Belt resistance and required pull

### 1.1 Structure (ISO 5048 / DIN 22101) [H]

Total peripheral force on the drive pulleys to be transmitted to the belt (effective tension), in N:

    F_U = F_H + F_N + F_S1 + F_S2 + F_St

- F_H main resistance (idler/roller rotation, belt flexure, material flexure), F_N secondary resistance (acceleration of material at the
  loading point, skirt friction in the loading zone, pulley bearing drag, belt bending round pulleys), F_S1 special main resistance
  (idler tilt, chute/skirt friction over the length), F_S2 special secondary (belt cleaners, plows, skirts over part of the length),
  F_St slope resistance.
  Source: Overland Belt Analyst ISO 5048 help page (lists the five terms and variables, equations given only as images):
  https://overlandftp.s3.amazonaws.com/help/NetHelp/iso-calculation.html . Also describes DIN 22101 as "very similar".
- Sum structure with "main + secondary + slope + special" also in the DIN/ISO papers:
  https://actamont.fberg.tuke.sk/pdf/2017/n2/12gladysiewicz.pdf and the CEMA-vs-DIN comparison
  https://www.matec-conferences.org/10.1051/matecconf/201816601007 (DIN uses one global friction coefficient f, CEMA separate idler/belt/material coefficients).

### 1.2 Equations (units: N, m, kg, kg/m) [H for form, M for constants]

Main resistance of a flight of length l, inclination delta:

    F_H = f * l * g * [ q_RO + q_RU + (2*q_B + q_L) * cos(delta) ]        [N]

- f hypothetical friction coefficient (dimensionless); g = 9.81 m/s^2; q_RO, q_RU rotating idler mass per metre, carry and return side (kg/m)
  = (mass of rotating parts of one idler set) / (idler spacing); q_B belt mass per metre (kg/m); q_L material mass per metre (kg/m).
  Form cited from a lecture deck (secondary): https://www.scribd.com/doc/74670541/Belt-Conveyors ; variables confirmed on the Overland page above.
- Material load per metre: **q_L = Q / (3.6 * v)** [kg/m] with Q in t/h, v in m/s. (Capacity relation Q = 3600*A*v*rho: see
  https://www.firgelliauto.com/blogs/engineering-calculators/conveyor-belt-speed-and-capacity-calculator ; one other page prints 3.6 with rho in t/m^3 which is off by 1000, so check units.) [H]

Slope resistance (lift h positive upward, m): **F_St = g * q_L * h** for the loaded run, plus belt weights g*h*(q_B up - q_B down) which cancel when both strands
are in the same vertical span (belt weight cancels for a closed loop) [H, textbook]. For 20.8 kg/m: about **204 N per metre of lift** [EST from the formula].

Secondary resistance, ISO 5048 form [M, constants recalled, not re-verified]: F_N = C-factor approach: F_H + F_N is approximated as
**C * F_H** where C depends on conveyor length (ISO table, approx. 1.9 at 80 m, 1.78 at 100 m, 1.58 at 150 m, 1.31 at 300 m, ~1.1 at 1 km). For L = 60 m
extrapolate **C ~ 2.2 [EST, low]**. Short conveyors are dominated by secondary resistances: that is the key qualitative point and
is consistent with the Belt Analyst note that "main resistance dominates on long horizontal belts" (see f-table source below).
Acceleration of material onto the belt at the feed: **F_acc = (Q/3.6) * v = mdot * v** [N] (mdot kg/s; v m/s) [H, momentum].
Skirt friction, cleaners: lecture deck gives a scraping term proportional to belt width with ~1500 N/m "scraping factor" (unverified, [L]); treat as a few hundred N total [EST].

### 1.3 Friction factor f [M]

| Condition | f | Source |
|---|---|---|
| Clean, modern, well-aligned, low-rolling-resistance cover | 0.018 | https://bison-convey-website.pages.dev/glossary/friction-factor [L/M] |
| Normal industrial conveyor, good condition (DIN 22101 recommended default) | 0.020 | same; also open-source `conveyance` lib default 0.02, https://conveyance.readthedocs.io/en/latest/api.html |
| Older, dusty | 0.022-0.025 | same |
| Short, dirty, poor maintenance | 0.028-0.030 | same |
| Italian drum-motor catalogue | 0.025-0.030 | https://www.tecnicaindustriale.it/dugom_rulli_catalogues/belt_conveyors_drum_motors.pdf |

f rises in cold weather, with belt-speed deviations, with worn/seized idlers. Effect: a 10% error in f gives ~10% error in F_U on a long belt (same glossary).
A degrading idler set is therefore another slow load disturbance common to all drives. [M]

### 1.4 Worked example: ~60 m, flat, ~150 t/h, 2 m/s [all EST, arithmetic from the formulas above]

Assumed hardware (my picks for a light-duty aggregate/bulk belt): belt width 800 mm, q_B = 10 kg/m, carry idlers 3-roll at 1.2 m spacing, rotating mass
12 kg per set -> q_RO = 10 kg/m; return idlers 1 roll at 3 m, 10 kg -> q_RU = 3.3 kg/m; f = 0.020; L = 60 m.

- q_L = 150/(3.6*2) = **20.8 kg/m**; mdot = 41.7 kg/s.
- g*f*L = 0.02*60*9.81 = 11.77 N per (kg/m).
- Loaded: F_H = 11.77*(10 + 3.3 + 2*10 + 20.8) = 11.77*54.1 = **637 N**. Empty: 11.77*33.3 = **392 N**.
- Secondary with C = 2.2: loaded +764 N, empty +470 N. Special (skirts, cleaner) loaded 200 N, empty 100 N. Material acceleration 83 N (loaded only).
- **F_U loaded ~ 1.7 kN; F_U empty ~ 1.0 kN.** Empty/loaded ratio ~ 0.57 (so cutting feed to zero removes only ~40% of the pull on a short flat belt).
- Power at pulley P = F_U * v: loaded **3.4 kW**, empty 1.9 kW. Motor shaft with gearbox 0.95 and motor 0.93: **~3.9 kW loaded**, ~2.3 kW empty.
- Per drive (three equal shares): ~1.3 kW loaded, i.e. 560 N pulley force each. With a 0.4 m pulley: belt 2 m/s -> 95.5 rpm; total pulley torque 337 N*m, 112 N*m per drive.
- Each +1 m of lift adds ~204 N (12% of loaded pull); a 3 m incline would add 35%.
- Order-of-magnitude cross-check: MATEC paper reports DIN 22101 effective tension 22.2 kN and 66.5 kW for a 1000 t/h, 3 m/s belt
  (https://www.matec-conferences.org/10.1051/matecconf/201816601007). Power scales ~ with Q*v for this class; our 150 t/h, 2 m/s case is plausibly a 4-10 kW conveyor. [M]
  If the customer's three drives are each 7.5-15 kW the belt is heavier (steel cord/wider/inclined) than assumed: ask (see section 10).

Kinetic energy / equivalent mass [EST]: belt 2*60*10 = 1200 kg, material 20.8*60 = 1250 kg (carry side), idler rotating mass 13.3*60 = 800 kg, pulleys/gear reflected
~300-500 kg -> **m_eq ~ 3.5 t loaded, ~2.3 t empty**. Unpowered stop deceleration a = F_U/m_eq = 1700/3500 = 0.49 m/s^2 -> belt stops from 2 m/s in **~4 s** if all
drives trip (loaded), longer if empty-running resistance is lower relative to mass. On a short flat belt the loaded belt does not coast far.

---

## 2. Drive arrangements and why a short belt has several drives

- Head drive (drive at the discharge): standard, the carry strand is pulled so the return strand is the slack side; best traction. Tail drive: pushes the belt
  from the return side, slack more likely, and the return strand acts as an energy-storing spring (forum). Intermediate/booster: needs its own friction wrap
  (belt cannot be pushed). Options listed: single head, two head, single tail, head plus tail, etc. Max belt tension and belt size depend on number and location of drives.
  Sources (forums, vendor): https://www.bulk-online.com/en/node/61041 , https://bulk-online.com/en/node/64056 , https://bulk-online.com/de/node/61825 , patent https://image-ppubs.uspto.gov/dirsearch-public/print/downloadPdf/4930624 . [M]
- Dual-pulley head drive (two pulleys, belt wraps both, snub for ~210-420 deg total) is chosen to raise wrap angle and so the traction capacity; each
  pulley can have its own motor. Typical single-pulley wrap 180 deg, snub pulleys push to 210 deg or more
  (https://editions.mydigitalpublication.com/article/POWER+TRANSMISSION+COMPONENTS/4596595/794613/article.html). [M]
- Why a 60 m belt may have three drives although the power is only a few kW (my reasoning, partly source-backed): (a) standardisation on smaller drive sizes (drum
  motors or geared motors), (b) traction: low-tension short belts have a small T2, so the allowed T_e per pulley is small; sharing the effective tension over two or
  three pulleys keeps each pulley's T1/T2 below e^(mu*theta) (c) redundancy/limp mode (d) drives at both ends reduce peak belt tension (this is
  the "maximum tension depends on drive location" point). The customer's reason is unknown: **ask** (section 10). [EST/M]
- Take-up location: should be near the drive on the slack side, away from the point of maximum tension, minimise max tension
  (forum, bulk-online 61041). Gravity take-up weight about **2 * T2** (180 deg wrap on the take-up pulley)
  (https://mechanical-design-handbook.blogspot.com/2026/01/202601conveyor-belt-tension-t1-t2.html [L]).

---

## 3. Load sharing between drives

### 3.1 Passive and active schemes

- **Master/follower, speed + torque**: first drive in speed control, its torque output is sent (analogue/fieldbus) to the others in torque mode, so torque and power are
  shared; this is "the correct way" for mechanically coupled motors per forum guidance.
  https://invertekdrives.com/support/iknow/vfd-fundamentals/torque-control-and-master-slave-operation-15 ,
  https://industrialmonitordirect.com/blogs/knowledgebase/powerflex-525-torque-follower-load-sharing-workaround [M]
- Followers get a small fixed speed offset (2-5% suggested, or +-0.1-0.5% trim in another source) so the master "leads" and followers do not fight it.
  https://industrialmonitordirect.com/blogs/knowledgebase/multi-motor-conveyor-drive-systems-load-sharing-configuration [L]
- **Torque balancing**: each VFD holds a set share of total torque (25% each in a 4-drive example). Same page. [L]
- **Passive slip sharing (direct-on-line motors / no torque loop)**: motor with more load slips more and sheds load to the others. Nominal slip 1-5%, full-load slip 3-8%.
  Needs matched gear ratios and pulley diameters: pulley diameter wear differences (e.g. 1% diameter difference = 1% speed difference) produce large
  torque differences because the torque-speed slope is steep (a 1% speed mismatch at 4% rated slip moves about 25% of rated torque) [EST, derived: dT/T_rated = d(speed)/slip_rated]. [M]
- **Droop**: speed falls with load by a set percentage. Not covered by sources I found (the vendor articles above do not mention droop). Standard VFD feature in principle; model as
  speed_ref_i = speed_ref - droop*torque_i/torque_rated_i [EST, generic].
- **Power-proportional sharing**: forum expert on a six-drive conveyor with drive overcurrent: drives should share load in proportion to their power ratings and the
  control system must ensure that no drive is overloaded for any appreciable time; recommends a power-monitoring program. https://bulk-online.com/en/node/74015 (search-result summary; page itself returned 503) [L].
- Stall risk: a stalled motor in a multi-motor-on-one-VFD setup can overheat without tripping because the VFD overload is sized for the combined current
  (PLC/forum summary, industrialmonitordirect). Relevant if the customer shares one VFD across motors. [L]

### 3.2 What happens when a drive drops out

Sources do not describe this. This is the single biggest gap in public sources. What follows is physics + my reasoning. **[EST unless noted]**

1. A tripped VFD removes torque; the motor/gearbox/pulley freewheels. Passive drag is small (gearbox and bearing losses, order 1-3% of the drive's rated pull, i.e. tens of N here)
   but the reflected inertia (motor rotor * i^2) rides with the belt.
2. The belt is a stiff coupling: elastic wave speed c = sqrt(E*A/q_B). For a fabric belt EA of the order 3-10 MN (rated 400 N/mm * 0.8 m = 320 kN, EA ~ 10-30x breaking strength, [L]) -> c ~ 550-1000 m/s,
   one-way transit of 60 m ~ 0.06-0.1 s. So over 0.1 s time scales the belt is a rigid link; model as one speed for all drives, with the elastic stretch relevant only to tension sags.
3. Missing torque becomes a speed deficit: dv/dt = (sum T_drives - F_U)/m_eq. For a step of 1/3 of the pull (560 N) on m_eq 3.5 t: a = -0.16 m/s^2, i.e. 0.16 m/s in the first second (8%) if nothing reacts.
4. Survivors:
   - **Speed-master + torque followers**: if the tripped drive was the follower, the master's speed PI raises its torque and the others follow the master's torque signal. Response limited by the speed
     loop (order 0.1-1 s) -> speed dip of a few percent recovers in about a second or two. If the tripped drive was the master, followers lose their reference and typically trip or fall to a fallback; behaviour depends
     entirely on configuration -> **must ask the customer** (section 10).
   - Each survivor then carries 1.5x its previous torque (3/2 for identical drives, equal split). If it hits a current/torque limit (VFD typically 150% for 60 s of drive continuous current, the old fixed
     rule, newer drives use an I2t curve; vector drives ~200% for 30 s; variable-torque drives ~110%) the belt keeps slowing until the survivors stall or trip on overload.
     Sources: https://eng-tips.com/threads/frequent-motor-start-up.96292 , https://forums.mikeholt.com/threads/input-amps-vfd-overload-condition.140350/ , https://eng-tips.com/threads/vfd-i2t.395856 [M]. VFD max torque 150%, preferably design to 135%: https://overlandftp.s3.amazonaws.com/help/NetHelp/breakawaytorque.html [M].
5. The patent on belt energy efficiency states (search summary, PDF unreadable): when one drive is lost the remaining drives must pick up its power and may be overloaded, and that
   a motor is overloaded if its drive power rises proportionally by more than the smallest power reserve of the motors; slip operation can be detected early by comparing the drive powers of the drives.
   https://image-ppubs.uspto.gov/dirsearch-public/print/downloadPdf/11220403 [L, secondhand]
6. **Thermal consequence for survivors [EST, medium]**: copper loss ~ I^2. If a motor runs at 65% of rated load (share 1/3 of 3.9 kW with 2.0 kW rating) its winding rise is perhaps
   80 K*(0.4 + 0.6*0.65^2) = 52 K; at ~98% it is 78 K; delta ~26 K. With the "10 K halves insulation life" rule
   (https://www.motioncontroltips.com/?p=21658 ; https://www.designworldonline.com/?p=561515 ; rule of thumb, [M]) that is 2^2.6 ~ **6x faster insulation ageing**
   while in limp mode. Thermal time constants for small motors: winding ~10-30 min, frame ~30-60 min [EST]. Class F hot-spot limit 155 C, class B 130 C
   (fabrico.io/blog/motor-insulation-classes/ [M]). Bearings/gear: load-life exponent 3 for ball bearings -> life ~ (1/1.5)^3 = 0.30 [textbook, M]; gearbox oil temperature also rises.

---

## 4. Traction: Euler-Eytelwein limit and why it tightens with fewer drives

### 4.1 Equations [H]

    T1 / T2 <= e^(mu * theta)         (impending slip; theta wrap angle in radians; mu friction coefficient)
    T_e = T1 - T2   (pulley effective tension = drive pulley peripheral force F_U_i)
    T2_min = T_e / (e^(mu*theta) - 1)         (slack-side tension needed to avoid slip)
    T_e,max = T2 * (e^(mu*theta) - 1)         (pulley capacity at a given slack tension)

Sources: https://www.elastotec.com.au/a-deeper-understanding-of-conveyer-pulley-friction , https://www.tribonet.org/calculators/belt-drive-band-brake-friction-calculator/ ,
https://mechanical-design-handbook.blogspot.com/2026/01/202601conveyor-belt-tension-t1-t2.html .

### 4.2 Friction coefficient values [M]

- ISO 5048 / DIN 22101 use a conservative value "up to 0.45" depending on lagging and contamination; ceramic lagging kinetic values up to 0.8; Harrison FEM
  inputs 0.9 static / 0.7 kinetic, effective 0.32-0.84 depending on tension. (Elastotec link above.) Ceramic tile on rubber planar test 0.4-0.8 (https://issuu.com/primecreativemedia-2016/docs/abhr0921_lr/s/13679143).
- Rubber on metal 0.2-0.4 (Belt Friction Theory PDF, https://demo2.wcbi.com/Download_PDFS/u2C073/243494/Belt%20Friction%20Theory.pdf).
- Typical design values (my recollection of DIN/CEMA tables, **not verified**): bare steel dry 0.30-0.35, lagged dry 0.35-0.40, wet/muddy 0.10-0.25 [EST/L].
- Friction is not constant: real contact has a slipping arc and a stationary arc; the slipping arc grows with load; mu reaches the kinetic maximum only at/near slip
  (Elastotec; Pit & Quarry). Slip is progressive, not an on/off switch. [M]

### 4.3 Numbers for the best-guess plant [EST]

theta = 210 deg = 3.665 rad:

| mu | e^(mu*theta) | e^(mu*theta) - 1 | T2_min / T_e |
|---|---|---|---|
| 0.25 (wet) | 2.50 | 1.50 | 0.67 |
| 0.35 (dry lagged) | 3.61 | 2.61 | 0.38 |
| 0.45 (ISO ceiling) | 5.21 | 4.21 | 0.24 |

Per drive at equal shares: T_e = 560 N -> T2_min 216 N dry (373 N wet). After one drive loss: T_e,i = 842 N -> T2_min 324 N dry (561 N wet). The factor 1.5 in demand
is the factor 1.5 in required slack tension. **If the take-up/belt sag set T2 at e.g. 700 N, capacity per pulley = 1830 N dry or 1050 N wet: safe dry, thin when wet.** [EST]
Because capacity depends on T2 (set by take-up and where the drive sits relative to it), a gravity take-up keeps T2 constant while a screw take-up leaves it as set; a head drive on a gravity take-up behind the pulley keeps T2 from the weight,
but "driving through a gravity take-up loses slack-side tension and the pulley slips" (bulk-online forum summary https://bulk-online.com/en/node/61883 [L]).

### 4.4 What a lost/seized drive does to the traction budget [EST derivation]

- Drive that merely tripped (freewheels): its pulley no longer transfers force (T1 ~ T2 across it, apart from bearing drag); the belt's total required T_e must be carried by the remaining pulleys.
- **Seized (locked) pulley with the belt dragged over it**: sliding friction gives T_in/T_out = e^(mu_k*theta), so it resists with T_out*(e^(mu_k*theta) - 1).
  At T_out = 700 N, mu_k = 0.35, theta = 210 deg: **~1.8 kN of extra drag, as large as the entire running pull of the loaded belt (1.7 kN)**. Survivors (2 x 1.3 kN max at 150%, from section 6) cannot supply 3.5 kN,
  so the belt slows and stalls, with friction heat, lagging damage and fire risk at the locked pulley. This corrects the README note "a seized drive is a drag": it is a drag whose size
  depends on the slack tension and wrap and can be the dominant load. Real outcome depends on whether the gearbox actually locks the output shaft (motor rotor seized behind a gearbox of ratio i
  reflects a very large inertia/locked state; a belt cannot turn it). [EST, medium]
- Friction heat at a slipping pulley: P = F_slip * v_slip; belt-pulley slip heating ignites belts: MSHA alert on a drive that kept running with a stalled belt, drive rollers slipping on the belt, friction started a belt fire
  (Desperado Mine 1992, https://fmshrc.gov/sites/default/files/decisions/commission/westfuel.pdf and MSHA BP card https://www.msha.gov/sites/default/files/Alerts%20and%20Hazards/BP%20Cards/BP-92 Belt Conveyors Fire Protection.pdf). [M]

---

## 5. Starting, breakaway and the restart trap

- Starting force exceeds running force because of (a) breakaway/static friction of idlers and belt (rubber viscoelastic, grease cold-stiff), (b) acceleration F_a = m_eq * a_start, (c) material at rest in the loading zone.
  Rules of thumb: starting torque 1.5-2.5x running (https://mechanical-design-handbook.blogspot.com/2025/12/how-to-calculate-starting-torque-and.html [L]);
  unloaded conveyor can start at ~40% torque, loaded sometimes up to ~180% (Bulk-online / softstarter forum summaries, https://www.bulk-online.com/en/node/61430 (503 now; via search summary) and
  https://industrialmonitordirect.com/blogs/knowledgebase/configuring-softstarter-full-load-start-mode-for-150hp-conveyor-motors [L]); softstarter guide: motor should give >=125% of breakaway torque at reduced voltage. [L]
- Belt Analyst: breakaway torque computed with a "Breakaway Friction Multiplier" on the running friction; VFD max torque typically 150%, require breakaway under 150% and ideally under 135%;
  for direct-on-line starts the pull-up torque must exceed breakaway. https://overlandftp.s3.amazonaws.com/help/NetHelp/breakawaytorque.html [M]
- A forum voice argues "breakaway friction" is partly mythical and the extra draw is largely idler/belt stiction: measure instead (bulk-online 61430 summary) [L]. Starting friction decays within seconds of motion: model f_start(t) = f*(1 + (m_b - 1)*exp(-t/tau_b)), tau_b 2-5 s [EST].
- Restart of a stalled **loaded** belt: material is distributed along the whole belt (30 s of feed), idlers and belt are stiction-locked; required pull ~ m_b * F_U_loaded + F_a.

**Numbers for the best-guess plant [EST]**: m_b loaded 1.8 (range 1.5-2.2), empty 1.4; accelerate in 20 s at 0.1 m/s^2: F_a = 3500*0.1 = 350 N.
- Required to start loaded: 1.8*1700 + 350 = **3.4 kN**.
- Drive rated pulley force each (2.0 kW, eta 0.88, 2 m/s): 880 N continuous; at 150% limit 1320 N.
- Three drives: 3960 N at 150% -> starts. Two drives: **2640 N -> cannot start** (even at 1.8 x 1700 = 3060 alone it fails). Running with two: 1700 N vs 1760 N continuous = 97% -> "just carry".
  This reproduces the README's trap: a design with ~1.5x per-drive overload on loss of one drive that runs fine at steady state but cannot restart loaded. The stall is plausible and typical advice is
  "run out the belt (empty it) before stopping, or start empty". (Note the *ratio* matters: 3 drives gives a factor 1.5 when one drops, hence 3/2; "1.4x" in README is within estimation error.)
- Also slip on restart: T2 of a stationary belt is smaller if the take-up has relaxed; "slip is most likely at startup, usually insufficient T2 or worn lagging" (mechanical-design-handbook T1/T2 page above) [L].
  Check slip at 200% motor torque, since conveyors slip at up to ~200% of full motor torque (bulk-online 65778 snippet, https://www.bulk-online.com/en/node/65778 [L]).
- Fluid couplings and soft starters limit transmitted torque (delay-fill couplings give soft starts); VFD gives full control incl. ramp time. Coupling-based drives share load passively by slip. (General knowledge, [L].)

---

## 6. VFD limits useful for the simulation

- Overload ratings: heavy duty (constant torque) 150% for 60 s of the drive's continuous current; normal duty ~110% for 30-60 s; vector ~200% for 30 s. Current limit protects the semiconductors; the motor I2t protects the winding.
  Sources: eng-tips 96292, Mike Holt thread, industrialmonitordirect I2t guide https://industrialmonitordirect.com/blogs/knowledgebase/vfd-i2t-motor-overload-protection-configuration-guide ; one manual says 150% => fault in 60 s, e.g.
  https://www.manualslib.com/manual/1550791/Weg-Cfw100.html?page=112 [M].
- Model: trip when thermal state integral of ((I/I_n)^2 - 1) over time exceeds a limit; with 150%/60 s: I2t margin = (2.25-1)*60 = 75 (units of I_n^2*s). [EST generic]
- Speed/torque loop: torque loop 5-50 ms, speed loop 50-300 ms [EST, generic vector drive]. Ramp time 10-30 s for conveyors.

---

## 7. Seized drive, locked gearbox, failed motor back-driven; backstops and clutches

- Backstops (holdbacks): sprag/roller one-way clutches stop reverse rotation; required on inclined conveyors (gravity runaway), absent as a need on flat belts.
  https://www.warrenforensics.com/2021/09/22/conveyor-backstops-sometimes-one-isnt-enough-part-1/ ; https://www.engineeringnews.co.za/article/reverse-running-on-inclined-conveyors-prevented-2015-05-29 [M].
  Incident: loaded incline limestone conveyor, V-belts between motors and reducers broke, conveyor stopped held by a single backstop on one reducer; when the torque arm was removed the loaded
  belt ran backwards and the unlocked reducer injured a worker (Warren Forensics). Shows that holding against load goes through one gearbox. [M]
- When a gearbox is back-driven the speed-reducer becomes a speed-increaser; drivetrain parts designed for forward torque are overloaded (Renold/Engineering News above). Relevance to flat belt: if one motor is
  dead and its gearbox is back-driven by the belt, it adds a small drag (mechanical losses, see section 3.2) unless seized.
- Seized/locked gearbox or rotor: see 4.4: sliding of belt over a locked pulley is the worst case. Protections in practice: zero-speed/speed-differential switch on each drive pulley, torque-limiting
  couplings or shear pins, fluid couplings with fusible plugs (thermal) [general knowledge, L], VFD stall trip. None of the searched sources gave a failure-sequence study for this.
- One-way (overrunning) clutch on each motor/gearbox input ("freewheel") allows a non-driving drive to be dragged without being turned when the belt is faster; not found in sources, mention as option to ask. [L]

---

## 8. Feed control, material on belt, transport delay

- Capacity Q = 3.6*q_L*v (t/h, kg/m, m/s) and Q = 3600*A*v*rho (A cross-section area m^2, rho t/m^3). With 150 t/h at 2 m/s and rho = 1.0 t/m^3: A = 0.0208 m^2.
  A troughed 800 mm belt at 35 deg and moderate surcharge can carry roughly 0.04-0.06 m^2 [EST, low], so nominal fill is 35-50% and **the physical maximum surge is about 2-2.5x nominal** before spillage. (CEMA / manufacturer tables not accessed; see
  https://www.mekaglobal.com/en/blog/belt-conveyor-capacity-calculation for tables by width and speed.)
- Transport delay: a change at the feed point reaches the far end after L/v = 60/2 = **30 s**. Material on the carry strand at any time: m = q_L*L = 1250 kg. Pull responds as a ramp: F_U(t) = F_U,empty + (F_U,loaded - F_U,empty) * (fraction of carry strand loaded). Feed shut-off therefore reduces demand
  linearly over 30 s (1.7 kN -> 1.0 kN) while a step in feed rate shows up in the drives gradually. Drive currents therefore trail the feed by up to 30 s: a surge shows as a ramp in all drive currents together (all proportional to their torque shares), not as an individual drive deviating.
- Feeders: belt/vibratory/apron feeders set volumetric rate; a controllable feeder or VFD speed change is the primary lever. Surge sources: bin/hopper bridging release, feeder PID overshoot, wet or sticky material. Upstream lump size and moisture also change f of material. [general knowledge, L]
- Slope: if there is any lift, F_St = 204 N per metre of lift (loaded); proportion of pull due to material increases from 41% (flat) to 60-70% with a few metres of lift [EST].
- Chute blockage or buildup at the feed: raises skirt resistance, adds a pull that does not scale with feed. [L]

---

## 9. Incidents and case studies (what exists publicly)

No public case study of the exact scenario was found. What is available:

1. Slip friction fire, MSHA Desperado 1992: belt stalled while drive continued, rollers slipped on belt, friction ignited belt; the slippage switch failed to shut it down in another MSHA alert (rock wedged at take-up roller).
   https://fmshrc.gov/sites/default/files/decisions/commission/westfuel.pdf , https://www.msha.gov/sites/default/files/Alerts%20and%20Hazards/BP%20Cards/BP-92 Belt Conveyors Fire Protection.pdf [M].
2. MSHA Program Policy Letter on remote restarting: speed difference between drive roller and belt causes slip, friction, overheating, fire hazard. https://www.msha.gov/sites/default/files/Directive%20%26%20Guidance/PPL/PPL13-V-07_0.pdf [M].
3. Inco conveyor fire investigation: drive pulley friction is an ignition route; fires typically start after stop from heated defective rollers/drives with coal dust and grease. https://onemine.org/documents/investigation-into-conveyor-fires-and-experience-with-protective-devices-at-inco-limited [M].
4. Fort Knox mine fire Jan 2026: failed pulley bearing overheated and ignited the belt (not slip). https://fm.kuac.org/mining/2026-03-27/fort-knox-mine-fire-wasnt-caused-by-safety-violation-inspector-says [L, news].
5. Bulk-online thread "Drive Motor Overcurrent" on a six-drive conveyor with unequal sharing; patent on drive-power comparison for slip detection (section 3).
6. Reliability sources: "conveyor lines suffer more motor failures than other systems" (https://gesrepair.com/why-conveyor-lines-experience-more-motor-failures-than-other-systems) and maintenance guides list rising current, temperature and vibration as early indicators; overload and worn lagging cause slip
   (https://oxmaint.com/industries/manufacturing-plant/conveyor-belt-failure-troubleshooting-guide-for-manufacturing-plants , https://reliabilitysolutions.net/?p=1362) [L].
7. Warren Forensics backstop incident (section 7).

Gap to disclose: no quantitative incident record of surviving-drive burnout after the loss of one drive. This scenario needs the customer's own failure history. (Ask.)

---

## 10. Instrumentation, protections and interlocks (typical)

Typical conveyor protection set (vendor sources: Thermo Fisher blog https://www.thermofisher.com/blog/mining/lets-talk-about-safety-when-using-belt-conveyors , FineTek https://www.fine-tek.com/uploads/download/341/337c1012e48216b105de8f31554efd09.pdf , Kiepe/Vossloh https://www.mining-technology.com/contractors/electrical/vossloh/ , Pacific Automation https://pacificautomation.com.au/content/files/brochures/smart-series-brochure-web.pdf) [M]:

| Device | Function | Typical thresholds |
|---|---|---|
| Pull-cord emergency stop | trips and latches the starter; along both sides | on pull |
| Belt-speed / zero-speed switch, belt slip detector | compares belt (tail or bend pulley) speed to drive pulley / underspeed trip | alarm ~90-95%, trip 80-85% of nominal [EST; vendor range not retrieved] |
| Belt misalignment (belt drift) switch | roller lever, two stages | alarm 20 deg, trip 35 deg (vendor datasheets) |
| Blocked chute / plugged chute | stop feeders and belt | on level |
| Drive motor current/torque (from VFD) | overload, load-sharing watch | I2t trip as in section 6 |
| Motor winding temperature (PTC / PT100), bearing temperature | thermal trip | class F 155 C; typical alarm at ~130-140 C, trip 145-155 C [EST] |
| Vibration (accelerometer, retrofit) | bearing/gear degradation | ISO 10816/20816 zones, [general] |
| Tension (load cell on take-up), belt weigher | feed rate/material on belt | |
| Start-up warning horn, interlocks: upstream feeder permissive from belt-running + speed OK; downstream starts first | | |

- Slip detection via drive powers: the survivors' current rising together while the belt-speed sensor under-reads. Comparing drive powers of the drives, per the patent, detects slip early (section 3.2 item 5).
- Slippage switches can fail (see MSHA): do not assume protection exists; treat as a stuck/dropped sensor scenario (README scenario 5). [M]
- Shared-VFD stall warning: where several motors share one VFD, overload trips are sized for the combined current (industrialmonitordirect). [L]

---

## 11. Implications for the simulation

### 11.1 Must model (cheapest model that still produces the right behaviour)

1. **Rigid-body belt speed** v with one equation: m_eq * dv/dt = sum_i F_i - F_U(v, load, state) where F_i = pulley force of drive i (limited by VFD torque/current limits, thermal derate and slip capacity).
   Elastic belt waves are 0.1 s scale and can be neglected at 1 s steps; model slip as an extra state only when a pulley exceeds its Euler capacity.
2. **Load = f(feed history)**: a conveyor transport line (30 s) so F_U(t) ramps with the material on the belt, plus empty-belt pull (about 60% of loaded for a short flat belt), not material alone.
3. **Load sharing law**: torque-follower with master speed PI, per-drive share weights, optional droop; each drive with torque limit (150% continuous-rating, I2t 60 s), thermal derate, failure state (running, degraded, tripped-freewheel, seized).
4. **Slip/traction**: per-drive T_e,max = T2*(e^(mu*theta) - 1), mu variable (wet/dry), T2 from take-up; slip raises when F_i exceeds capacity, and the seized-pulley drag formula (section 4.4).
5. **Starting resistance**: breakaway multiplier m_b decaying with time, restart trap with fewer drives; stop transient ~4 s coast.
6. **Winding temperature** per motor (first-order thermal, copper loss ~ I^2 + fixed loss), ageing rate doubling per 10 K, bearing life ~ load^3, gearbox temperature; this is what makes overload damage cumulative.
7. **Degradation precursors** on the weak drive (rising drag torque, vibration, winding temp, falling delivered torque vs reference); and sensor fault injection (stuck/dropped). Tripping a VFD is a discrete event.

### 11.2 May simplify

- Belt elasticity, catenary/sag, idler-by-idler tracking, belt creep (0.1-1% elastic creep through the wrap arc; ignore except as speed difference signal), pulley lagging wear dynamics, material-specific flow.
- Equal weights for the three drives; one T2 for all pulleys; deterministic f per section.
- Fluid couplings and soft starters are out of scope unless the customer has them.

### 11.3 Suggested default parameters (best-guess plant; all EST unless noted)

| Parameter | Value |
|---|---|
| Length L, speed v, nominal feed Q | 60 m, 2.0 m/s, 150 t/h (q_L 20.8 kg/m); feed range 0 to 2.2x nominal |
| Belt width, q_B, q_RO, q_RU | 0.8 m, 10 kg/m, 10 kg/m, 3.3 kg/m |
| f (DIN default) | 0.020 [M], range 0.018-0.030; cold/dirty +20% |
| Secondary + special | C = 2.2 on F_H, +200 N special loaded, +100 N empty, +83 N acceleration of material (loaded) |
| F_U loaded / empty | ~1.7 kN / ~1.0 kN; P_pulley 3.4 / 1.9 kW |
| m_eq loaded / empty | 3.5 t / 2.3 t |
| Drives | 3 identical: 2.2 kW class (use rating 2.0 kW, eta 0.88, pulley force 880 N cont., torque limit 150% = 1320 N), two at head (210 deg wrap) one at tail |
| Pulley diameter, gear ratio | 0.4 m (95.5 rpm), 15.4:1 on 4-pole 50/60 Hz |
| Wrap theta, mu dry/wet | 210 deg; 0.35 / 0.25 (kinetic 0.30 / 0.20) |
| T2 (take-up) | 700 N (capacity per pulley 1.8 kN dry, 1.05 kN wet) |
| Breakaway multiplier m_b loaded / empty, tau_b | 1.8 / 1.4, 3 s; acceleration 0.1 m/s^2 |
| Motor thermal | winding tau 15 min, frame 45 min, rated rise 80 K (class F with B rise), ambient 35 C, age acceleration 2^((T - T_ref)/10) with T_ref 120 C winding |
| VFD | 150% for 60 s; torque loop 20 ms, speed PI bandwidth ~2 Hz; ramp 20 s |
| Failure events | tripped drive (freewheel drag 3% of rated); seized pulley (drag T2*(e^(mu_k*theta)-1)); weak drive (bearing drag +5% to +30% of rated over hours, vibration up, delivered torque down) |
| Protections | zero-speed trip at 80% for 5 s, overload I2t, winding temp trip 150 C, pull cord |
| Feed lever | feeder 0-100%; effect reaches the pull over 30 s |

Scenario checks the defaults give (all derived above): normal share 65% of rating each; one drive out -> 98% each, ~6x faster ageing, loaded restart impossible (2.64 kN < 3.4 kN);
feed cut to zero -> survivors drop to ~57% load factor of full-feed operation within 30 s; seized drive -> total demand ~3.5 kN, survivors cannot hold it, belt slows to stall in seconds unless stopped.

---

## 12. Open questions for the customer

1. Actual belt: width, belt type (fabric/steel cord), idler spacing and roller mass, lift/incline, material and bulk density, real feed rate and peak surge; is a belt scale or feed sensor present?
2. Drive layout: where is each drive (head pair + tail, three at head, intermediate), pulley diameters and lagging type, wrap angle, take-up type (gravity/screw) and where; any snub pulleys.
3. Drive hardware: motor kW and speed, gearbox type, VFD brand/model and overload rating, soft starters or fluid couplings, brakes/backstops/overrunning clutches, any shared VFD.
4. Load-sharing mode: master/follower torque, droop, or passive slip; which drive is master; what the control does when the master or a follower trips (stop all, continue, switch master, re-scale torque limits).
5. Actual design margin: installed vs running power, current at full feed with all drives and observed current share per drive; can two drives restart a loaded belt (have they tried)? What happened in past trips?
6. Observed failures: which component of the failing drive fails first (bearing, gearbox, motor winding, VFD), precursor timescales, sensor traces if any, past damage on surviving drives.
7. Sensors and their sample rates: per-drive current/torque, winding temp, vibration, pulley speed encoder, belt speed (tail), tension; what is already logged.
8. Allowed actions: feed rate, belt speed (VFD reference), per-drive stop/start, torque bias, alarms only; operating constraints (continuous process, cannot stop?), cost of a stop vs cost of motor damage.
9. Environment: wet/dust/cold season (friction and mu), lagging condition, belt cleaner/skirt configuration.
10. Existing protections: slip switch, zero-speed, pull cords, thermal trips and their set points.
