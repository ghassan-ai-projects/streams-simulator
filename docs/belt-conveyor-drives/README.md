# Multi-drive belt conveyor: one drive degrading, the others overloaded

Status: deep-dive and planning only (2026-11-07). No code, no runs. Branch `belt-conveyor-drives` in streams-simulator,
agentic-stream and tamoz, cut from `origin/main`.

Scope (owner decision, 2026-11-08): this work changes streams-simulator and agentic-stream, and tamoz only if a
milestone needs it. The Retrofit Designer (`agent-research-lab/plant-retrofit-designer/`), which grew out of this use
case, is not built now; it is the next big milestone after this one.

## The customer problem (as told by the co-founders, best guesses marked)

A real industrial customer runs a belt conveyor about 60 m long driven by several motors (say three) that share the
load. When one motor stops working for any reason, its load moves onto the others; running overloaded, they age faster
and may be damaged in the long run. The sensors show the failing motor degrading before it stops, so the goal is to
predict the failure and react early: in effect the failing motor "notifies" the others so the system can get ready.

Owner direction (2026-11-07): upgrading our systems for this use case is fine. Gaps it exposes in streamsim,
agentic-stream or tamoz are to be planned as upgrades, not worked around, within the architecture rules.

## Working understanding

- The drives are mechanically coupled through the belt: same speed, the pulling force is shared by torque. Losing one
  puts its share on the survivors at once.
- The load is the material on the belt: belt pull is roughly friction x (belt + material mass) + lift. The material on
  the belt is the last ~belt-length/belt-speed seconds of feed (60 m at 2 m/s is about 30 s).
- Damage on the survivors is cumulative (insulation aging roughly doubles per 10 C, bearings, gearbox), not a threshold.
- A seized drive (locked gearbox) is a drag, not just a missing drive.
- Each drive pulley can transmit only so much force before the belt slips.
- Restart trap: starting a loaded belt needs much more torque than running it; two drives may keep a loaded belt running
  but be unable to restart it after a stop.
- The strongest lever is the feed: cut feed, run the belt empty, shift load off the weak drive, stop it at a chosen time.

## Best-guess assumptions (unknown, to be replaced with customer answers)

- 60 m, flat, about 150 t/h, about 2 m/s; three identical drives with VFDs (two at the head, one at the tail).
- Torque load sharing; two drives can just carry the running belt at full feed (about 1.4x each) but cannot restart a
  fully loaded belt.
- Controllable feeder at the loading point.
- Sensors: per-drive current/torque, winding temperature, vibration, pulley speed; belt speed; feed rate.

## Candidate scenarios

1. Slow bearing failure on drive 3 (hours of precursor).
2. Sudden trip of drive 3 under full load (no precursor; restart trap).
3. Seized gearbox on drive 3 (drag, slip).
4. Feed surge (all currents up together; must not be read as drive 3 failing).
5. Sensor faults hiding a degrading drive (stuck or dropped vibration channel).

## Open questions for the customer

Drive positions and ratings, incline, VFDs and load-sharing mode, which actions are allowed (feed, belt speed, per-drive
stop/start, torque bias, alarm only), which sensors exist, what failure costs them most.

## Review of the preparation notes (coordinator, 2026-11-07)

- The two research notes look contradictory but are one model: the survivors' fate depends on the drives' rating
  margin. Large margin: survivors near rating, slow aging (the co-founders' story). Small margin: 1.3-1.5x, I²t trip in
  minutes, cascade and a stranded loaded belt. The plan must cover both; the margin is the first customer question.
- "Feed is the strongest lever" is weaker on a short flat belt (empty-belt pull about 60 % of loaded pull) and acts as a
  ~30 s ramp, not a step. It is strong when there is lift.
- A seized pulley is far worse than a drag (adds about the whole loaded running pull); survivors may stall.
- Sanity check of the worked example (about 1.7 kN, about 3.4 kW at 2 m/s): plausible for a light 800 mm belt; real
  drives may be much larger. Parameters must be data, not baked in.
- Verified: composition connections aggregate only by `sum`
  (`internal/component/internal/domain/compose_follow.go:48`).
- Cross-repo blocker: agentic-stream is single-entity; the cross-drive correlation that is the heart of "motor 3
  notifies the others" does not exist there yet.
- Market: no product found that coordinates drives on a predicted failure; monitors only alarm.

## Folder layout

- `research/` - internet research (conveyor mechanics, degradation and prognostics)
- `repos/` - what each of the three repositories can and cannot do today for this use case
- `PLAN.md` - the deep plan, written after the research and repo notes
