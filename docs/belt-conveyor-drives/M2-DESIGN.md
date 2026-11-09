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

## 5. As built, rows 1 and 2 (2026-11-09, streamsim)

Row 1:

- composition-v0.1 takes `implements` and `board_type` on an equipment entry (`board_type` needs `implements`, a schema
  rule, refused at `/equipment/<i>/implements`); `domain compose` copies both onto the composed kind and checks the domain
  against the kind files and board mappings of `--kinds-dir`, as `domain validate` does. Generic: no conveyor word in Go.
- The bulk feeder gains the point `feed_setpoint` (state `feed_set`, tag ZT, fraction 0-100 %) with a new library
  instrument `streamsim/setpoint-readback` (the setpoint read back from the controller's register: periodic 2 s, no noise,
  resolution 0.1 %, one `dropout` fault). No existing instrument fits: the valve position transmitter is a noisy analog
  feedback (σ 0.2 % of span), and a host verifies a write to within the metric's resolution. Composed channel
  `bulk_feeder.feed_setpoint`.
- `examples/belt-conveyor/kinds/`: `bulk_feeder.kind.json` (`feed` kg/s, `feed_setpoint` % writable, `set_feed` with the
  argument `feed_setpoint`), `feeder-controller.board.json` (`set_feed` → `bulk_feeder__set_feed` with `feeder_id` fixed to
  `F-1` and `percent` from `feed_setpoint`; `feed` → `bulk_feeder.feed`, `feed_setpoint` → `bulk_feeder.feed_setpoint`), and
  `belt_conveyor.kind.json`, `conveyor_drive.kind.json` captured from a 20 s session birth (`kind capture`). Plants A and B
  announce identical digests for all three kinds, so both compositions declare `implements` on all three entries (the
  feeder also `board_type: feeder-controller`): the committed files are now checked against the plants on every compose,
  run and session.
- Consequence: every `domain compose`, `run`, `plantwire node`, `plantwire connect` and `mcp` of a conveyor plant needs
  `--kinds-dir examples/belt-conveyor/kinds` (a domain that implements is refused without the check). A host sees the feeder
  as `feed`, `feed_setpoint` and `set_feed`; the world, its command log, the outcome weights and `b4` keep the raw names.
- Tests: compose copies and refuses (component domain layer); `domain compose` and `plantwire connect --check` of S2-B with
  and without `--kinds-dir` (cli); a 30 s plant-B session against the reference host: the feeder is born through its kind
  file with `feed_setpoint` writable, every born kind equals its committed file, and `set_feed {feed_setpoint: 18}` is
  accepted and verified by read-back within 6 s (test/contract).

Row 2 (`test/contract/belt_limp_deadline_test.go`, long tier): noiseless plant B, steady at full feed, D-3's `vfd_fault`;
one `bulk_feeder__set_feed` after a delay; saved means no protection trip of D-1 or D-2 and no `stranded` violation in the
30 min after the command. Bisected on a 1 s grid and checked at the delay ± 1 s:

| Command | Latest delay after the trip that still saves |
|---|---|
| limp feed 18 % | 98 s |
| cut to 0 | 103 s |

The survivors' inverters trip 121 s after the loss with nobody acting, so a limp command has about 98 s; M2's reaction
bound is set from 98 s with a margin.
