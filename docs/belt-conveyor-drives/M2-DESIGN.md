# M2 design: first closed loop, S2-B

Status: design for M2 (2026-11-09), from [PLAN.md](PLAN.md) section 7 (M2), section 4.2 (A1, A2, A6) and the owner
decisions O17 and O18 (section 9.6). The working closed loop to copy is the fan experiment
(`docs/physical-devices/fan-experiment/`, agentic-stream `cmd/agentic-stream/experiment_plantwire_*_test.go`).

## 1. The loop

`streamsim plantwire connect` runs plant B with a shortened S2-B plan (D-3's VFD trips at full feed; a downstream stop
later) at `pace: wall`; agentic-stream `serve` is the plant-wire host with a conveyor site, kinds, catalog and spec.
When a drive's current drops to about 0 A while the feed is high, the line enters `line_overload` and the decision
`reduce_feed` (R1, automatic) sends `set_feed` to F-1 at the limp feed (18 %, PLAN.md section 1.1, model M0). The feeder
reports its setpoint back (O17), so the host verifies the command. The survivors' inverters must not trip; the belt
must not strand. The run folder is graded with `streamsim outcome` and, with a verdict written by A6, `streamsim score`.

## 2. Work, by repository

| # | Repo | Work |
|---|---|---|
| 1 | streamsim | Compositions accept optional `implements` and `board_type` per equipment entry, copied onto the composed kind (generic; domain-spec already has both). The feeder gets a setpoint read-back measurement; a feeder kind file (`feed`, `feed_setpoint` writable; `set_feed`) and a board mapping under `examples/belt-conveyor/kinds/`; both plants use them; the other kinds' files are captured from a session NBIRTH (`streamsim kind capture`) |
| 2 | streamsim | Measure the latest time after D-3's trip at which a direct cut to the limp feed still saves the survivors (noiseless, plant B); M2's reaction bound is set from it with a margin |
| 3 | agentic-stream | A1: a site reading may name a `device`; lookup device first, then kind; schema, refusals, tests, docs. The fan's thermal site stays byte-identical |
| 4 | agentic-stream | Conveyor test data under `examples/belt-conveyor/testdata/` (outside the `examples/*/*.situation.yaml` gates): kind copies (guarded byte-equal, like the thermal ones), site with per-drive readings, event schemas, catalog (`reduce_feed` presets `limp_b` 18 %, target F-1), spec (`line_overload` from per-drive current and feed) |
| 5 | agentic-stream | Tests: a fast one against the fake node (the DCMD is `set_feed` 18 % to F-1, zero quarantines); a streamsim-driven one (skips without `STREAMSIM_BIN`) over 3 seeds: feed at the limp setting within the measured bound, verified by read-back, no survivor trip, not stranded, `verify-run` passes; A6 writes `verify.json`'s companion `verdict.json` (subject map: diagnosis to D-3, route to F-1) and `streamsim score` and `streamsim outcome` grade the run |

## 3. Rules that apply

- streamsim: AGENTS.md as in M0/M1; the component module change is generic (no conveyor word in Go).
- agentic-stream: its AGENTS.md (15-line functions, 300-line files, no domain words in production code, coverage and
  test-time budgets, `scripts/slow-tests.txt` changes called out); the fan experiment guards must stay green.
- Long tests run per package, `-p 1`, `GOMAXPROCS=4` (owner's machine).

## 4. Exit (PLAN.md section 7 M2, as changed by O17, O18)

3 seeds: the limp feed within the measured bound after the trip, read back; no survivor protective trip; not stranded;
graded by `streamsim score` with A6's verdict and by `streamsim outcome` (beats doing nothing); `verify-run` passes;
zero quarantines; fan re-check FR+ (streamsim `make phase-gate`, the device tests, agentic-stream `make ci-check` and its
Experiment tests, the P4 kit rung).
