# G review B: adversarial safety review of the plant wire gateway

- **Reviewer:** B (adversarial, safety). Read-only.
- **Subject:** `agent-research-lab`, branch `fan-gateway`, HEAD `ec54561` (G2). Files under
  `real-world-sensor/assessment/tools/`: `arduino_plantwire_gateway.py`, `plantwire_*.py`, `fake_thermal_*.py`; the firmware
  `real-world-sensor/examples/firmware/dht11_test/dht11_test.ino` was read for what it enforces.
- **Method:** read all of it against SPEC section 9 and the brief; wrote probes against the in-memory `Bench` and against the
  real `Gateway` with the pty board and the Python host; ran 63 hand-made mutants of the safety checks against the test suite
  (a copy of `assessment/` in the scratchpad; the repository was not edited. The first `pytest` run left an ignored
  `tools/.pytest_cache/`.). Probe and mutation scripts: `/private/tmp/claude-501/-Users-ghassan-my-projects-streams-simulator/517cecf1-e1a3-4525-abf0-112e2348451e/scratchpad/{mutate.py,mutate3.py,probes/}`.
- **Baseline:** 239 passed, 10 skipped (the streamsim and Agentic Stream cross-repository tests skip: no checkout or binary
  at the probed paths). Python on this machine is 3.9.6.
- In file references below `T/` is `real-world-sensor/assessment/tools/`.

## Summary

I could not make a command energize the fan without passing the epoch floor, the policy allow-list, the deadline, the
bounds or the boot check (see "What held up"; every mutant of those checks is killed by the tests, three of them only by a
hang). The real exposure is elsewhere: **availability** (one small record kills the gateway), **honesty** (it says
`refused` for commands that did or will run, and shows stale actuator state as current), **host-loss timing chosen by the
host**, **silent loss of the epoch floor**, and **test gaps in the controls that carry the authority model**. Everything
that goes wrong is bounded by the firmware lease (5 to 10 s, duty at most 600 permille) because the firmware enforces its
own limits; that is the only reason nothing below is rated higher.

| # | Severity | Finding |
|---|---|---|
| 1 | High | One nested-JSON `DCMD` (about 1.4 KB) kills the gateway, before any check, even a stale-epoch or safe-stop one |
| 2 | Medium | `refused` is reported for commands the board executed or will still execute |
| 3 | Medium | A silent board is never detected: stale actuator state is served as current, no `link_down`, no stop |
| 4 | Medium | The host sets how long a hung host goes unnoticed (up to 30 min) |
| 5 | Medium | A missing state file silently resets the epoch floor to 0; no lock; parent directory unchecked |
| 6 | Medium | The idempotency ledger never expires at run time; after 65 536 commands every command is refused |
| 7 | Medium | Other uncaught-exception triggers (serial boot id, disk errors, mapping); no last-gasp stop |
| 8 | Medium | Tests pass with these safety controls removed (list in section 8) |
| 9 to 14 | Low | Socket trust, bursts and floods, receipt content, no proactive stop, unpinned kit, replay semantics |

---

## 1. High: a single nested-JSON `DCMD` kills the gateway

- **Where:** `T/plantwire_commands.py:73` (`contains_float(record["args"])` is the first thing `on_command` does),
  `T/plantwire_args.py:34-41` (unbounded recursion), `T/plantwire_args.py:44-66` (`fits`, same); no handler in
  `T/arduino_plantwire_gateway.py:94-101,223-243` (`main` catches only `OSError` and `UnsafeSocketError`).
- **Argument:** `strict_loads` catches `RecursionError`, but only when the JSON parser itself overflows (around 990
  levels). A document nested 400 to 900 levels parses, and then `contains_float` (two Python frames per level) overflows the
  stack. The exception is not caught anywhere, so the process exits with a traceback. The float scan runs before the device,
  epoch, policy, host-alive or safe-stop logic, so **any** peer on the socket can send it: a stale-epoch host, an unlisted
  policy, an unknown device, a command while the host is not alive, in `awaiting_state`. In the lab shape the relay forwards
  `args` unchanged, so whatever an agent puts in `args` reaches this line.
- **Reproduction (verified):**
  - In memory (`probes/p1.py`): depth 50 is refused `out_of_range`; depths 400 and 900 raise `RecursionError` out of
    `node.receive`; depths 990 and 3000 are silently ignored.
  - Against the real `Gateway` (`probes/e1.py`): fan at 450 permille with a 10 s lease, then one 1405-byte `DCMD`
    (`stop_fan`, `authority_epoch` 99, unlisted policy, `args {"x": [[...450 deep...]]}`): the gateway thread dies with
    `RecursionError`, the pty fan stays at 450 until its lease lapses. Nothing restarts it; there is no supervisor in the runbook.
- **Why it matters:** the gateway is the only component that can answer `stop_fan` sooner than the lease and the only
  place that holds the authority model. One record from the least trusted party removes both.
- **Fix:**
  1. Reject on depth and size before any recursion: in `host_record_problem` (or `strict_loads`), walk `args` iteratively
     and refuse anything deeper than, say, 4 (kind argument schemas are flat) as `malformed`; make `contains_float`
     and `fits` iterative or depth-limited.
  2. In `Gateway.step` (or around `node.receive`), catch `Exception` per record: log `record_crashed`, drop the record,
     keep the session. A crash on one record must never be a process exit.
  3. Add a regression test with depth 400 and 900 through `node.receive`.

## 2. Medium: `refused` reported for commands that were, or will be, executed

- **Where:** `T/plantwire_commands.py:183-186,203-209` (`on_tick` and `_give_up`: a command still in the `dispatched` stage
  after 2 s is settled `refused not_ready`); `T/plantwire_commands.py:172-177` (a `result` that arrives while the stage is
  `dispatched` is dropped as `unexpected`); `T/plantwire_board.py:181-200` and `T/plantwire_serial.py:62-79` (the line is
  not withdrawn from the write buffer on give-up); `dht11_test.ino:176-185` (the firmware ignores `expires_after_ms`
  when `not_before_mono_us` is 0, which the gateway always sends, `T/plantwire_firmware.py:130`).
- **Reproduction (verified):**
  - `probes/p4.py`: the firmware's receipt line is garbled in transit, its result line arrives intact. The fan is
    applied (450 permille, 10 s lease). The host receives `DRECEIPT refused not_ready` and nothing else. The ledger stores
    `refused/not_ready` for 25 h.
  - `probes/p4b.py`: the command bytes sit unsent (board not reading) for 2.5 s; the host gets `refused not_ready`; the
    board then reads the stale line, applies the fan, and its answers are suppressed (`firmware_*_suppressed`), so the
    host never learns.
- **Why it matters:** the spec says a receipt is the node's decision. The host (and the agent) believes the fan was not
  commanded while it runs for up to 10 s; the identical command is then replayed `refused` for 25 h (ledger), the exact
  opposite of what happened. The order of effects is still FIFO, so a later stop works, which keeps this at Medium.
- **Fix:**
  1. A `result` in stage `dispatched` implies the receipt: treat it as `accepted` then settle the result.
  2. On the receipt timeout, do not answer `refused` while the line may still reach the board. First withdraw it from the
     write buffers; if it may already have been written, send the device's safe stop and refuse only after a `state` frame
     shows the output off (or keep the command open until a frame shows the output).
  3. Do not ledger a transient `not_ready` produced by the gateway's own timeout.
  4. Make the firmware honour `expires_after_ms` (or send a real `not_before_mono_us`); today the 20 s clamp
     (`T/plantwire_checks.py:96-99`) is cosmetic on this firmware.

## 3. Medium: a silent board is never noticed; stale actuator state is served as current

- **Where:** `T/plantwire_board.py:64-69` (a `query_state` every 2 s, no check that an answer arrives),
  `T/plantwire_board.py:127-133,153-169` (`_safe_state` and `_raw` are last-known values with no age),
  `T/plantwire_devices.py:133-139` (`answer_query`), `T/plantwire_telemetry.py:84-97` (staleness only for the sensor).
- **Reproduction (verified, `probes/p7.py`):** fan commanded to 450 permille, then the board goes silent for 60 s with the
  serial port open. No `DDEATH`, no stale marking for `fan-01`; `DQUERY` returns `DSTATE safe_state false, duty_permille
  450` (60 s old, `self_reported true`). Only the DHT sensor turns `stale`.
- **Why it matters:** a hung firmware loop keeps the AVR PWM running past the lease (the lease watchdog is `updateFan`
  in `loop()`, `dht11_test.ino:477-481`, no hardware watchdog), and the gateway then reports the fan on and good forever,
  neither stopping it nor going `link_down`. Commands to a hung board end as `refused not_ready` (finding 2).
- **Fix:** record the arrival time of the last `state` or `telemetry` frame; if none for, say, 3 poll intervals, treat the
  board as lost (`link_down`: close and reopen the port, which resets the Arduino through DTR and so de-energizes it).
  Stamp `DSTATE` and actuator `DDATA` with the age and mark them `stale` after that age. Also consider enabling the AVR
  watchdog in the firmware.

## 4. Medium: the host decides how long a hung host goes unnoticed

- **Where:** `T/plantwire_node.py:188-191` (`_beat_interval_ns` is taken from the host's `STATE`), `:111-116` (host loss =
  `heartbeat_misses` (3) x that interval), `T/plantwire_validate.py:93` (`heartbeat_interval_ms` 100 to 600000).
- **Reproduction (verified, `probes/p6.py`):** after one `STATE` with `heartbeat_interval_ms 600000`, a command queued
  with `not_before_ns` +20 min (deadline +23 h) runs 20 minutes later although the host said nothing for 19 min; the phase is still
  `live`. `on_host_lost` (flush the queue) never fires inside 30 min.
- **Why it matters:** SPEC section 8 defines liveness in the host's units, but the safety property "stop acting when the
  host is gone" is the gateway's. A hung (not closed) host keeps authority for half an hour, and every queued command it left
  gets a fresh 5 to 10 s lease. The brief itself says to keep `not_before` waits short; nothing enforces it.
- **Fix:** gateway-side ceilings from configuration: effective interval = `min(host interval, gateway cap)` (for the fan, a
  few seconds), and a maximum `not_before` wait (for example 60 s) refused as `out_of_range` for hardware nodes.

## 5. Medium: the epoch floor silently resets when the state file is absent; no lock; directory unchecked

- **Where:** `T/plantwire_state.py:32-37` (`raw is None` gives `NodeState()` with floor 0, no log), `T/plantwire_files.py:15-26`
  (`lstat` then a separate `read_bytes`; the parent directory is never checked), no `flock` anywhere.
- **Reproduction (verified, `probes/p6.py`):** floor 7 persisted; delete `state.json`; a new node starts with floor 0 and logs
  only `session_start epoch_floor 0`. The session id also restarts at `gw-1`. The runbook says "Deleting it resets the epoch floor, so a stale command
  could be accepted again" but nothing stops a cleaned `/tmp`, a changed `--state-file`, or a container restart from doing it
  silently. A second gateway on the same state file is also not prevented (two serial opens, each reset the board; both dial the host).
- **Fix:** if the ledger file or any earlier artefact exists but the state file does not, refuse to start (or require an
  explicit `--init-state`); log a `state_file_created` event at WARNING level on first creation; take an exclusive `flock` on
  the state file for the process lifetime; open with `O_NOFOLLOW`, `fstat` the descriptor, and require the parent directory to be
  owned by the user with mode 0700.

## 6. Medium: the ledger never expires at run time, then refuses everything

- **Where:** `T/plantwire_ledger.py:47-60` (`full` counts all entries; `settle` never prunes), `:62-73` (retention applied only in
  `load_ledger` at start), `T/plantwire_commands.py:105-106` (`ledger full` gives `not_ready`), `:154-159,241` (`_closed` grows
  without bound; `_free_dispatch_id` is O(n) per repeated `command_id`).
- **Argument:** a control loop that renews the lease every 5 s with fresh keys makes 17 280 entries a day; the 65 536 limit is
  reached in under four days of continuous running, after which every new non-safe command is refused until a restart prunes
  entries older than 25 h. The file grows two fsynced lines per command. Mutant `M19` (ignore `full`) survives all 235
  tests, so this path is untested. `probes/p3.py`: 2000 safe stops with one `command_id` cost 2.2 s (quadratic), and
  `_closed` keeps every dispatch id forever.
- **Fix:** prune entries older than the retention on `settle` (or periodically in `on_tick`); bound `_closed` (keep ids only
  while a late firmware answer is possible, for example 60 s); test the full ledger.

## 7. Medium: other uncaught-exception triggers, and no last-gasp stop

All end the process, because nothing catches `Exception` (`T/arduino_plantwire_gateway.py:223-243`):

- **Garbled boot id from the board.** `T/plantwire_devices.py:84-86` raises `ValueError` for a boot id that is not a wire id,
  called from `T/plantwire_board.py:88-96`, whose handler (`:78`) catches only `JSONRuleError`, `GatewayError`, `AttributeError`.
  `parse_device_record` and `parse_telemetry` accept any non-empty string up to 256 bytes. Verified (`probes/p1.py`): `boot_id`
  `"boot 9~"` and a 200-character id both crash. One flipped bit on the serial line in a boot id would do it. The adapter has
  already replaced `self.boot` and the shadow when the exception hits, so state is half updated.
- **Disk errors.** `LedgerError` (`T/plantwire_ledger.py:54-59`) is raised from `_settle_receipt` inside a firmware-line
  handler, after the fan was commanded and before any receipt is sent; `StateFileError` (`T/plantwire_state.py:60-65`) from
  `_raise_floor`/`_end`. Neither derives from `OSError`. A full disk therefore kills the gateway mid-command.
- **Latent, after the 1.4.0 kind files arrive.** `T/plantwire_mapping.py:92-95` indexes `args[arg["from"]]` and
  `args[spec["by"]]`; `semantic_problems` checks that the name is an argument of the kind but not that it is `required`.
  An optional argument left out gives `KeyError` inside `dispatch`, after the command is already in flight.
- **Fix:** one top-level policy: any exception in a record or frame handler is logged and the record dropped; any fatal
  error first queues the device's declared safe stops to the serial port and flushes (best effort, 200 ms) before exiting
  non-zero; require `from` and `by` arguments to be `required` at load; check the boot id with `is_wire_id` in
  `parse_*` and quarantine the frame instead of raising.

## 8. Medium: tests that pass with a safety control removed

Mutation run: each line was changed in a copy and the whole suite run (`-x`). **Killed** (test fails or hangs): epoch check,
policy check, deadline, 24 h cap, boot check at receipt, host-alive, re-check at execution, queue flush on host loss and on
session end, argument bounds, required, `additionalProperties`, bool-as-integer, duplicate with other content, ledger replay,
safe stop = declared only, safe stop skipping the boot check, expiry clamp and the 1000 ms safe-stop cap, floor persistence,
lower-epoch `STATE`, unsafe-config refusal, board-lost give-up, `not_before` honoured, strict `DCMD`, mode checks on the state
file and socket, heartbeat loss, goodbye, inbound gap, profile/kind safe-stop cross-check, boot of board frames, receipt timeout,
duplicate JSON keys, wire-id `fullmatch`, ledger entry validation, 4 MiB outbound cap, inbound oversize and 64 KiB limits.

**Survived** (all 235 tests still pass):

| Mutant | Line | What is untested |
|---|---|---|
| Session id check removed | `T/plantwire_node.py:142` | a `DCMD` carrying another session's id is executed |
| Owner uid check removed on the socket | `T/plantwire_socket.py:31` | a 0600 socket owned by another user is dialled (this is the only authentication in v1) |
| Owner uid check removed on state and ledger | `T/plantwire_files.py:30` | a file owned by another user is trusted (floor, ledger) |
| `if not self.linked` in `dispatch` removed | `T/plantwire_board.py:182` | a command dispatched with the board link down would sit in `_out` and be sent after reconnect (`link_up` does not clear it) |
| Ledger `full` ignored | `T/plantwire_commands.py:105` | finding 6 |
| Queue cap `MAX_QUEUED` removed | `T/plantwire_commands.py:124` | unbounded queue of `not_before` commands |
| Repeat cap `MAX_WAITING` removed | `T/plantwire_commands.py:111` | unbounded repeats of one key |
| `stage != DISPATCHED` guard in `firmware_receipt` | `T/plantwire_commands.py:163` | a second receipt re-emits `accepted` |
| `pending.waiting = []` at session end removed | `T/plantwire_commands.py:221` | a result of the old session is sent into the new one |
| Serial pending cap removed | `T/plantwire_serial.py:63` | finding 11 |
| Safe stop sends `parameters {}` forced off | `T/plantwire_firmware.py:129` | raw arguments reaching a safe stop |
| Safe stop skips `identify` | `T/plantwire_checks.py:59` | unknown device or argument on a safe stop (equivalent today) |
| Float scan in `on_command` removed | `T/plantwire_commands.py:73` | equivalent (`args_refusal` repeats it) |

Fix: add one test per row (the first four are the ones that carry the authority model).

---

## Low findings

### 9. Socket trust (`T/plantwire_socket.py:26-69`)

The owner-only check is `lstat` on the socket and its parent, then a separate `connect` by path: there is no peer-credential
check after connecting (`SO_PEERCRED` / `LOCAL_PEERCRED`), ancestors of the directory are not examined, and a swap between the
check and the connect is possible for anyone who can rename the directory. `connect` is made on a blocking socket
(`:59-61`): on Linux a full listen backlog blocks the whole loop, serial service included (macOS returns `ECONNREFUSED`;
I confirmed that and could not test Linux). Fix: check the peer uid after connect, `settimeout` before connecting, resolve and
check every ancestor.

### 10. Floods of hostile input

A line flood: 640 KB of one-byte lines cost 3.6 s CPU plus one JSON log line per record to stderr (`probes/p5.py`), so a
peer can starve the loop and fill the log (`T/plantwire_socket.py:87-100` is O(n) per line, `T/plantwire_node.py:121-145`). `NCMD rebirth` is not rate-limited (each one
re-emits `NBIRTH` and the births and fsyncs the state). Fix: rate-limit and log-sample ignored records; honour the "one rebirth outstanding" rule.

### 11. No in-flight cap: a burst closes the serial port and resets the board

`T/plantwire_commands.py:139-152` dispatches every admitted command at once; the write buffer is 64 KiB
(`T/plantwire_serial.py:62-65`). 400 stop commands queue 153 KB (`probes/p10.py`); `LinkLost` closes the port and the Arduino resets
(fan off, new boot id, rebirth). Safe stops are unthrottled by design (spec) and each uses a fresh firmware key. Fix: at most
a few in flight per device; extra non-safe commands `not_ready`, extra safe stops coalesced.

### 12. Receipt and result content

- `T/plantwire_commands.py:275-280`: a refusal for an unknown device or board carries the host's own `expected_boot_id` as
  `boot_id` (verified: `wrong_boot` with `boot-777` while the board has no boot). Fix: use `"unknown"` or omit.
- `T/plantwire_board.py:194-196`: `fallback_on_lease_loss` `hold_last` or `revert_to_local` is accepted and logged but the
  firmware always turns off. Fail-safe, but the receipt says `accepted`. Fix: refuse values the node cannot honour (`out_of_range`).
- A late `unknown` or `not_ready` after a reboot is labelled with the new boot id (verified, `probes/p8.py`), not the one the
  command ran under.
- `T/plantwire_commands.py:250-255`: after a crash between accept and result, a replay returns the stored `accepted` and never a
  `DRESULT` (the spec promises exactly one). Fix: emit `unknown`.
- `T/plantwire_commands.py:223-230`: a gateway-made transient refusal is stored 25 h and replayed for the same key (see 2).

### 13. No proactive stop

On host loss (`T/plantwire_node.py:214-217`), session end, `SIGTERM` (`T/arduino_plantwire_gateway.py:186-191`) and fatal errors
the gateway sends nothing to the board; the fan runs out its lease (at most 10 s). The LED has no lease at all
(`set_led`) and keeps its state after host loss although the spec says every command falls back on host loss. Fix: on those
events send the declared safe stops of devices whose last output is non-zero.

### 14. Kit and profile are not pinned

`T/plantwire_contract.py:88-94` loads whatever `--kit` points to and computes the digests at run time; argument bounds and
`safe_stop` flags come from files an operator can edit (the profile cross-check blocks a false `safe_stop`; widened bounds are
stopped only by the firmware). Fix: pin the kind digests in the gateway's configuration and refuse to start on a mismatch.

---

## What held up (attempted and failed)

- **Authority bypass.** Every path to `dispatch` goes through `at_receipt`, `_recheck` or `for_safe_stop`; the checks run in
  the specified order; a queued command is re-checked at execution (epoch bump, deadline, boot, host alive); host loss, session end
  and board loss flush or give up the queue; a command with the board link down gets `wrong_boot` because `boot_of` is cleared.
- **Safe stops.** A command is a safe stop only if the kit declares `safe_stop` and the profile maps it to the firmware's
  `safe_stop` operation (cross-checked, mutant killed); the firmware parameters are forced to `{}`; the name `safe_stop`,
  `node-stop-N`, unicode look-alikes and `command_id` prefixes confer nothing; a safe stop can only de-energize.
- **Parsing.** Duplicate keys, `NaN`/`Infinity`, floats (`5000.0`, `1e3`), booleans as integers, huge integers, lone surrogates, unknown
  members in `DCMD`, wrong `lane`, wrong protocol version, ids with spaces, newlines or unicode, and lines over 64 KiB are refused
  or ignored without effect (apart from finding 1).
- **Idempotency.** The same key with other content is `duplicate`; a replay has no second effect; a replay with other
  `deadline_ns`, `authority_epoch` or `policy_digest` is first re-checked on its own fields.
- **Leases.** The gateway never re-issues; `lease_ms` is required by the kind schema (raw mode) or injected from the mapping
  (semantic mode) and bounded again by the firmware (5000 to 10000 ms, duty at most 600); the expiry clamp for the fan
  (20 000 ms) and safe stops (1000 ms) matches the firmware's limits. Opening the port resets the Mega, so a gateway restart is a
  safe state. A host can renew a lease forever; that is its authority, but nothing at the gateway limits continuous run time.
- **Link and files.** The socket must be a socket, owner-only, with an owner-only real directory; symlinks are refused;
  state writes are atomic with fsync and `O_EXCL|O_NOFOLLOW` staging; a corrupt state or ledger refuses to start; a torn last ledger
  line is dropped; the policy allow-list is empty (everything refused) unless given.
