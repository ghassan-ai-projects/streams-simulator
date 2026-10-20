# G review A: correctness and architecture of the plant wire gateway

Reviewer A, 2026-10-20. Scope: `agent-research-lab` branch `fan-gateway`, `git diff faf457b..fan-gateway -- real-world-sensor`
(7 commits, G1a to G2). Bar: the brief `briefs/gateway-plant-wire.md`, `contracts/plant-wire/SPEC.md` sections 2 to 13 and 15,
`G-FINDINGS.md` (Q1 to Q8 as recommended), `PLAN.md` D4, D5, D7. Read-only. Paths below are relative to
`agent-research-lab/real-world-sensor/assessment/tools/` unless they start with `assessment/`.

## What I ran

- `python3 -m pytest -q` in `tools/`: 239 passed, 10 skipped (Python 3.9.6). Nine of the skips are
  `test_plantwire_streamsim.py` because `bin/streams-simulator` is older than its HEAD.
- The same file against a binary I built from the streams-simulator working tree into the scratchpad
  (`STREAMSIM_BINARY=...`): 8 passed, 1 expected failure (the reference host never beats). So checks 1 to 4 of the brief hold
  against the real relay and reference host today.
- `diff -r assessment/contracts/plant-wire /Users/ghassan/my-projects/streams-simulator/contracts/plant-wire`: byte-identical.
  The four kit files equal `examples/devices/thermal-board/*` byte for byte. `sync_plant_wire_contract.py --check`: "in sync".
- `arduino_gateway.py` and `test_arduino_gateway.py`: no diff in the range. Every file in the range is an addition.
- Runbook step 0 and 1 commands re-run with the fresh binary (`kind validate`, `kind twin`, `plantwire connect --check`): all
  work and print the digests the runbook says. The `control.sock` estop line matches `bench-checks-m2-m5.md`.
- Four probe scripts (scratchpad, not committed) against `Bench` and `Harness`: they confirm findings H1, M1, M2, L1.
- Files: every module is at most 284 lines, every test file at most 221, no function over 21 lines. No module-boundary problem
  beyond L6.

## Verdict

The pipeline, the node core, the ledger, the epoch floor and the mapping are in good shape and the tests assert behaviour
(order of refusals, replay with no second effect, floor kept across restart, reboot and link-down sequences, rejected-result
suppression, clamp). One High defect in the board adapter (query polling multiplies with every command), two Medium behaviours
that contradict the brief or hide a real-world failure, and a set of Low items.

## Findings

### H1 (High): every command result adds a permanent extra `query_state` poll chain

- File: `plantwire_board.py:64-67` (scheduler), `:122-124` (`_result` appends `now` and `now + lease + margin`).
- Evidence: `tick()` removes only the entries that are due and adds one `now + 2 s`; the regular chain's pending entry stays in the
  list. A result then appends an immediate entry, which, once due, is replaced by a second `+2 s` entry that never merges with
  the first chain. The lease-end poll does the same. I measured on `Bench` (fake clock, 10 s windows):
  0 commands, 5 queries; after 1, 13; after 2, 23; after 4, 35; after 8, 55. The count rises with each command and never falls.
  Each answer to `query_state` also emits a `DDATA` for the LED and for the fan (`_state`, `:127-133`).
- Why it matters: the experiment's main workflow is repeated fan commands with lease renewal. After a few dozen renewals the
  gateway queries the firmware many times a second and floods the relay and the host with actuator `DDATA`, which also
  distorts the evidence the experiment is meant to produce. No test measures the poll rate, which is how it passed.
- Fix: keep one regular deadline (`_next_poll_ns`) plus a set of one-shot deadlines (`_extra_polls`); a due one-shot sends a
  query and does not reschedule. Add a test that issues N commands and asserts the queries per 10 s stay at the baseline.

### M1 (Medium): the gateway goes silent during host loss (and while waiting for the first `STATE`), so the relay drops the link after 30 s

- File: `plantwire_node.py:111-118` (`tick`), `plantwire_devices.py:readings` (buffers when not live),
  `arduino_plantwire_gateway.py:135-141` (`send_gap` only logs a warning at 25 s).
- Evidence: in `HOST_LOST` and `AWAITING_STATE` `readings()` buffers everything and nothing else is emitted. Probe: 40 s after host
  loss, 0 records sent. The relay resets its 30 s timer on any admitted record from the gateway, whatever the host's state
  (`internal/plantwire/internal/domain/relay_gateway.go:26-31`, `relay_liveness.go:12-20`). So after 30 s of host loss the relay
  reports every device `link_down` and closes the gateway's link.
- Contradicts: brief section 2 ("send readings at least every 30 s, or other valid records") and section 5 ("keeps your
  session"; acceptance check 7: "the session recovers by rebirth"). A host outage over 30 s ends in a reconnect and a new
  session instead of a rebirth. Q8 accepted the drop only for the unplugged-board case (no live device to send for), not for this.
- Fix: while the host is lost or not yet heard, send a harmless admitted control record every 10 to 15 s for a born device (an
  unsolicited `DSTATE` is schema-valid and is not a reading), or send live-lane data and let the relay count it as dropped.
  Keep buffering for replay. Add a node test with the fake clock: no host for 40 s keeps the longest send gap under 30 s.

### M2 (Medium): a sensor that is disconnected from the first frame is invisible to the host

- File: `plantwire_telemetry.py:44-55` (`_sample`), `:84-89` (`tick`), `:91-97` (`_report_stale`).
- Evidence: a `disconnected` frame takes the last good value; with none, the sample is dropped (`sample_dropped`, "no good value
  yet"). Probe with NaN humidity and temperature from boot, 12 s: zero `DDATA` for `zone-01`; `DBIRTH` stays `null` with
  `warming_up`. The stale rule does not fire either, because frames keep arriving and `_report_stale` needs a last good value.
  The host sees a device that is born and silent, with no quality saying why.
- Why it matters: this is the "broken sensor" case the experiment is meant to probe, and the one most likely on the bench (a
  loose DHT11 wire at start). The same loss exists for a reading that goes `out_of_range` and is later replaced by `disconnected`:
  `_last_good` is set before the range check (`:48`), so the carried value can itself be out of range.
- Fix: set `_last_good` only for in-range good values. When there is no value, either emit `disconnected` with the kind's
  `eu_range.min` (a documented placeholder) or add an explicit rule to G-FINDINGS Q4 and the runbook limits and make the host
  side show "no data" for a born, silent device. Add a test for the from-boot case.

### L1 (Low): an accepted command whose result was not stored before a restart never gets its `DRESULT`

- File: `plantwire_commands.py:250-256` (`_replay`), `plantwire_ledger.py` (entry `outcome` is `None` until the result).
- Evidence: probe: accept (firmware silent), restart, repeat the same key: only `DRECEIPT accepted`, no `DRESULT`, forever. SPEC
  section 9: an accepted command receives exactly one `DRESULT`.
- Fix: in `_replay`, an entry with `status == accepted` and `outcome is None` that is not in flight is answered `unknown` and
  stored as such.

### L2 (Low): the ledger and the closed-id set only grow while the process runs; a full ledger refuses all commands

- File: `plantwire_ledger.py:48,51` (`full`, no pruning outside `load_ledger`), `plantwire_commands.py:56,241` (`_closed`).
- Evidence: retention is applied only at load. At `MAX_ENTRIES` 65536 every non-safe command is `not_ready` ("ledger full")
  until a restart; at one renewal every 5 s that is about 3.8 days. `_closed` keeps every dispatch id for the process lifetime.
- Fix: prune entries older than the retention on `settle` when full (and bound `_closed` to the ids of recent commands).

### L3 (Low): the LED never falls back on host loss, and nothing says so

- File: `plantwire_board.py:194-196` (only lease commands log `fallback_is_safe_off`); runbook section 3, "Lease fallback".
- Evidence: SPEC section 9: "every command falls back on host loss". `set_led` has no lease on the firmware, so after a host loss
  or an unlisted-policy session the LED stays at its last pattern. The runbook limit speaks only of the fan.
- Fix: say it in the runbook limits and G-FINDINGS F9, or send the node's own `stop_led` safe stop when the host is lost and the
  last LED command carried `safe_off`.

### L4 (Low): refusal class for a wrong argument type

- File: `plantwire_args.py:28-31`.
- Evidence: SPEC section 9 table: schema failure is `malformed`, bounds are `out_of_range`. A string where an integer is declared,
  a missing required argument and an unknown argument all come back `out_of_range`.
- Fix: return `malformed` for a type, missing or unknown-member failure and `out_of_range` for `minimum`/`maximum`/`enum`.

### L5 (Low): a birth after a reboot or reopen carries the pre-reboot readings as current

- File: `plantwire_telemetry.py:29-31` (`reset` clears `_pending` and `_heard_ns` but not `_latest`/`_last_good`),
  `plantwire_node.py` / `plantwire_devices.py:board_linked`.
- Evidence: the new `DBIRTH` takes `current_readings` from `_latest`, so the old value is announced `good` with a new
  `birth_time_ns`. SPEC section 3 says a birth carries `null` with `warming_up` before the first sample.
- Fix: clear `_latest` (not `_last_good`) in `_note_boot`.

### L6 (Low): coupling and structure

- `plantwire_serial.py:9` imports private names (`_configure_serial`, `_consume_serial_lines`) from `arduino_gateway.py`. The file
  is unchanged as required, but the private names are now a dependency; a later extraction should export them.
- `plantwire_commands.py` and `plantwire_board.py` hold `node: Any` / `pipeline: Any` back-references and reach into
  `node.devices.boot_of`. A small `NodeView` protocol in `plantwire_ports.py` (session_id, epoch_floor, host_alive, boot_of, emit)
  would make the cycle explicit and typed.
- `sync_plant_wire_contract.py:22`: `CONTRACT_PARTS` omits `VERSION`, which `load_contract` requires. It is copied anyway, but
  `check_contract_source` would accept a source without it.
- `LedgerError` and `StateFileError` raised at run time (disk full, changed mode) are not caught in `main()` (`arduino_plantwire_gateway.py:221-232`):
  the process dies with a traceback and exit 1, no `NDEATH`. Failing closed is right; map them to a logged exit.

### L7 (Low): runbook accuracy

- Section 3 says checks 5, 6 and 7 need Agentic Stream because the reference host never beats. Check 6 does not: a node stop is
  a safe stop and skips the host-alive check, so it runs against the reference host (test `test_the_relays_node_stop_executes`).
  Check 7 (host stops, returns) also works with the reference host because its one `STATE` is forwarded on return.
- The `handed_to_board` log line (`plantwire_board.py:200`) is written when the line is queued, not when the serial write
  completes; the runbook's hand-off time is the `command_dispatched` line. Both are before the write. State that in check 8.

## Requirement-by-requirement

| Item | Status |
|---|---|
| Brief 1: owner-only dial (`lstat` on socket and parent, type, uid, mode), no TCP, 103-byte limit | Met (`plantwire_socket.py:23-31`) |
| NBIRTH fields, `hardware`, boards with members, kinds with `kind_digest`, `capability_digest`, `node_metrics` | Met; `1.4.0` per D5 (brief says 1.3.0) |
| Kinds equal the kit; digests equal `kind validate` | Met and re-checked |
| DBIRTH only after `STATE`; `null`+`warming_up`; board and kind consistent | Met |
| Unplug: `DDEATH link_down` per device, `DBIRTH` with new boot on replug | Met (tests and relay run) |
| Heartbeat to relay: record at least every 30 s | Met while a host is live; not met in host loss (M1) |
| Reads never blocked by serial I/O (one select loop, non-blocking writes, bounded queues) | Met |
| DDATA wall time, uncertainty, qualities, historical replay, buffer bounds, stale after overflow | Met (L5 aside) |
| SPEC 9 check order at receipt and again at execution | Met (`plantwire_checks.py`, `_recheck`); test pins the order |
| Idempotency: same key and content replays; other content `duplicate`; no digest check; persistent 25 h | Met (L1, L2 aside) |
| Queued repeat answered `queued`, same terminal pair; lost-ack retry served from ledger | Met |
| Safe stops: schema/device/command/boot only, never queued, never `duplicate`, start check, fresh firmware key (Q1) | Met |
| Node stops (`node-stop-N`) execute and are answered | Met (test and runbook check 6) |
| Epoch floor: highest accepted `STATE`; lower ignored; goodbye honoured at any epoch; persisted, atomic 0600 | Met (`plantwire_node.py:152-186`) |
| Host loss: three missed beats, queue flushed `session_ending`, buffer, safe stops still run, floor kept | Met; LED fallback L3 |
| Birth/rebirth: `bd_seq` +1 per session and per rebirth, `prior_last_seq`, seq not reset on rebirth | Met |
| Reboot reconcile: `DDEATH reboot` then `DBIRTH` new boot; queued command refused `wrong_boot` at execution | Met |
| Quality mapping: firmware good/disconnected, gateway `stale` (6 s), `out_of_range` (eu_range), `warming_up` | Met; M2 |
| Firmware mapping: expiry clamp (safe stop 1000, LED 60000, fan 20000), rejected-result suppression, reject-code mapping, `executed`/`safe_state` to `succeeded`, key rewrite | Met (`plantwire_checks.py`, `plantwire_commands.py:firmware_receipt`) |
| Board mapping both directions (command to raw, raw to kind) incl. scale, table, lease, raw_enums | Met; I ran a synthetic 1.4.0 kit through `BoardMapping` (`set_cooling 45` to `duty 450 + lease 5000`; `450` to `45`; `slow_blink` to `watch`) |
| Mapping vs kind file check (SPEC 15) | Met; falls back to a raw mode while the kit still has the 1.3.0 kinds. Deferred and recorded in commit 12e633c ("raw 1.3.0 kinds until S4b"). Not in the runbook |
| `not_before_ns` mapped to `not_before_mono_us` (SPEC 13) | Deviation by design: queued in the gateway, firmware gets 0 (G-FINDINGS). Not in the runbook limits |
| Exchanges, fixtures (valid/invalid/must-ignore/injection) | Met (26 exchanges, 19 node expectations; schema and hand validators agree) |
| `arduino_gateway.py` unchanged | Met |
| Vendored contract byte-identical | Met |
| Runbook commands | Correct (re-run); see L7 |

Deferred and acceptable: G1f part 2 (semantic kind files after S4b), instrument_range (kits declare only `eu_range`, so
`out_of_range` uses it), no keep-alive during a long unplug (Q8), the bench itself (nothing has been run on hardware).
