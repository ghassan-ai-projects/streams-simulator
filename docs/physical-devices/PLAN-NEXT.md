# Plan for the next working session (written 2026-10-19)

Where we are: M1 to M5, the policy allow-list and R13 are built, gated and merged to `main` (PRs 48, 49 and 50; `main` is at
`8070602e`). Nothing was tested on hardware. Work continues on the local branch `real-run`, cut from `main`. The goal for the next
session is the real run: the thermal LED/fan experiment driven by Tamoz through Agentic Stream.

## 0. First ten minutes: decisions and housekeeping

1. **Confirm `main` is green.** The PRs were merged in order; run `make phase-gate` once on `main` (or on `real-run`, which equals
   it plus this plan) as the starting evidence. Push `real-run` when you want it on GitHub (not pushed yet).
2. **Decide D-111 to D-113** (they change what the bench can prove):
   - D-111: should an independent instrument's reading reach the live "effect" stage? (needs a plant wire 1.4.0 evidence member, or
     the instrument's verdict stays in the bench table only)
   - D-112: how does a hardware-only session carry host evidence (a gateway-raised `NSYNC`, a relay barrier, or a new record)?
   - D-113: keep the exact-match contract version check or implement the minimum-version rule?
3. **Stop the demo bridge.** `kill 27234` (the canvas on `127.0.0.1:7788`, workspace in the scratchpad `fan-bench/`). The "hot
   afternoon" scenario is `fan-bench/hot-afternoon.scenario.json`; copy it into the repository's `examples/` only if you want it kept.

## 1. Baseline: repeat last time (the G1 joined emulated run)

Goal: prove acceptance test #1 still holds on today's checkouts before building anything new. It uses `device-wire-v1` through
`streamsim device serve`, not the plant wire, so M1 to M5 should not affect it.

1. Record the commits of `streams-simulator`, `agentic-stream` and `tamoz`; use a clean artifact folder under `.e2e-run/`.
2. Follow `agent-research-lab/real-world-sensor/assessment/RUNBOOK-G1.md`: terminal 1 `streamsim device serve --socket ...
   --capabilities ... --world ... --world-domain ... --world-entity ... --device-id dev-01 --boot-id boot-A`; terminal 2 the Tamoz
   worker (pinned Ruby via rbenv); terminal 3 `agentic-stream serve` with a continuous telemetry feed (send an extra heartbeat if the
   feed pauses through the scheduler cooldown).
3. **Use the thermal bench twin as the plant** (`examples/devices/thermal-board/`, domain id `arduino-thermal`). The device-wire
   capability catalog (`internal/device/testdata/thermal_capability_catalog.json`) already names `zone-01`, `led-01`, `fan-01` with
   the routes Tamoz uses (`set_indicator`, `select_thermal_mode`) and bounds that match the twin (duty 0 to 600, lease 5 to 10 s).
   The only existing bindings file (`internal/deviceworld/testdata/thermal.bindings.json`) maps to the cold-chain `reefer` domain
   (`reefer_id`, `set_fan_duty`), so write a bindings file for `arduino-thermal` (data only): `fan-01` to `set_pwm_lease`
   (`duty_permille`, `lease_ms`) with safe stop `stop_fan`, `led-01` to `set_led` (`brightness_permille`, `pattern`) with safe stop
   `stop_led`, one entity per binding (the binding model has an `entity` field; confirm `device serve --world` accepts a
   multi-entity twin, else bind the entities one device each).
   Script only the **disturbances** (the warm spell, the stuck sensor, the clock) from `fan-bench/hot-afternoon.scenario.json`, not
   the fan and LED commands: Tamoz decides those. Fix the warm spell magnitude first (see section 4, it raises the zone 16 degC).
   Find out what fed Agentic Stream's telemetry last time (the runbook only says it needs a continuous source of normalized events).
4. Pass criteria (from the runbook): telemetry accepted with zero quarantines, one Tamoz decision, one `set_indicator`/fan command
   executed on the device, state feedback reconciled, artifact `verdict.json` = `pass`.
5. Also run the simulator's own checks: `go test ./internal/device/... ./internal/deviceworld/... ./test/contract/...`.

Exit: a retained run folder with the three commits recorded, or a precise list of what broke.

## 2. The new path: Agentic Stream as a plant wire host

Facts: Agentic Stream has no plant wire code (a search for `NBIRTH` and "plant wire" in its checkout finds nothing). Tamoz needs no
change: it only talks to Agentic Stream. The brief is `docs/physical-devices/briefs/agentic-stream-host.md`; the contract is
`contracts/plant-wire/SPEC.md` (1.3.0, pinned `DIGEST`).

Run this as its own session in the `agentic-stream` repository. Suggested rounds:

| Round | Deliverable | Test peer (already in this repository) |
|---|---|---|
| H1 | Plant wire host skeleton: owner-only UDS listener, `STATE` handshake, version and digest negotiation, births (`NBIRTH`, `DBIRTH`) validated against the commissioned inventory and kind digests | `plantwire connect` / `plantwire relay` with the stand-in gateway; `plantwire host` as the reference to copy behaviour from |
| H2 | Reading ingestion: `DDATA` with quality into the normalized events (zero quarantines), `DDEATH`/rebirth, `boot_id` changes, stale and disconnected handling | stand-in gateway; a simulated node (`plantwire node`) |
| H3 | Commands: `DCMD` with `expected_boot_id`, `idempotency_key`, deadline, `authority_epoch`, `policy_digest`; `DRECEIPT`/`DRESULT`; `queued`, `expired`, `local_control` handled as ordinary refusals; the barrier (`NSYNC`/`HSYNC`) and optional `command_evidence` | the canvas interrupts (hold, approve, e-stop) against a connected session |
| H4 | The policy digest hand-over: expose the digest Agentic Stream signs with, and document how the simulator operator enters it (the session plan's `policies`) | `plantwire connect --check` echo |
| H5 | Joined run: Tamoz decision to a plant wire command to the simulated fan, effect read back from the twin | canvas live view, run folder, `plantwire twin-report` |

Exit: the "emulated" rung of the experiment through the plant wire, with the same pass criteria as section 1 plus: one hold, one
approval and one emergency stop shown on the canvas and reproduced on replay.

## 3. The physical rung (after section 2)

1. **Gateway plant wire side** (`agent-research-lab`, the Arduino gateway): brief `briefs/gateway-plant-wire.md`. The gateway is the only
   place a real board can refuse on policy; its allow-list must name the same digest as section 2 H4.
2. **Independent instrument**: brief `briefs/bench-instrument.md`. Needed for the exit "confirmed by the instrument rather than
   `DRESULT`". Decide D-111 first.
3. **Bench checks M2 to M5**: `briefs/bench-checks-m2-m5.md`, with exact commands and what to send back. Use `--delay-ms 250` on every
   reference host and a new `--out` folder per attempt.

## 4. Small simulator tasks that can run in parallel (any time)

| Item | What | Where |
|---|---|---|
| D-114 | `plantwire connect --out` should refuse an existing run folder | plantwire app + CLI + tests |
| D-115 | The reference host should exit cleanly with its report when its node is stopped | plantwire host |
| D-107, D-108 | Per-device badge for a stop that could not be sent; lead the e-stop outcome with the failure and agree the number | canvas |
| D-109, D-110 | Transient plant clip after the Interrupts bar resizes; wrap a 71-character digest in the narrow panel | canvas |
| Domain data | `ambient_warm_spell` raises the zone by 16 °C, not the 4 °C its description says (fault delta 4 and onset magnitude 4 multiply): fix the example domain and its description | `examples/devices/thermal-board/` |
| Open from M4 | Timeline pill de-collision and keeping fit-zoom on scrub; a real-browser look at the "unconfirmed" and "confirmed" command stages (needs a host that confirms effects, so after section 2 H5) | canvas |

## 5. Working rules for the next session

Writers (Sonnet) write production code and their own tests; Haiku runs gates and fixes lint-class findings; browser reviews run in the
user's real Chrome at its own window size; one commit per round and `make phase-gate` green before each push; never loosen a gate;
no pause for owner testing (the owner tests hardware at the end).

## Suggested order for tomorrow

1. Section 0 (green check on `main`, three decisions).
2. Section 1 (baseline run), about the first working block.
3. Brief and start the Agentic Stream session for H1 and H2 (section 2) while the simulator-side items in section 4 run alongside.
