# The owner's bench checks, M2 to M5

- **For:** the owner, at the thermal LED/fan bench.
- **From:** the streamsim physical-devices goal (R13). Updated in fan S4c (2026-10-20) for contract 1.4.0: semantic kinds on the wire through the board mapping (D7), the shipped twin domain, paced sessions (D2). The exits below are those of `docs/design/PHYSICAL_DEVICES.md` section 7. **Nothing here has been run against hardware.** The simulator's own tests use a scripted stand-in gateway; these checks are what turns each milestone's physical exit into evidence.
- **You need:** the built binary (`make build`, which writes `bin/streams-simulator`), `jq` (or any JSON Lines reader), the gateway with its plant wire side (`gateway-plant-wire.md`), Agentic Stream as the real host with the changes in `agentic-stream-host.md`, Tamoz, and the independent instrument (`bench-instrument.md`) for M5.
- **Convention.** Commands run from the repository root. `B=bin/streams-simulator`. A pass needs the expected output **and** the evidence file saved. A fail is a result: record it, do not retry until it passes. **`--out` must be a new folder for every session**: `connect` refuses a folder that already holds a run file (`plantwire_failed`; D-114, fixed). Name each attempt (`m5-session-1`, `m5-session-2`) and keep the failed ones; the canvas picks unique `c-<8 hex>` names. **Pace every session with a simulated world** (`"pace": "wall"` in the plan, as the canvas does) **and give the reference host `--delay-ms 250`** wherever it is started below: without either the world free-runs (about 75 MB of run folder per second), and `twin-report` refuses a `journal.jsonl` over 64 MiB. Stop sessions in minutes, not hours. The reference host prints its report and exits 0 when its node ends, also when `connect` is stopped during a `--delay-ms` wait, and when it is interrupted before any node arrived (D-115, fixed).

## 0. Preparation

```bash
git rev-parse HEAD                                   # record the commit
make build
B=bin/streams-simulator
$B help >/dev/null 2>&1; echo "exit $?"            # prints the usage; exit status 0
cat contracts/plant-wire/DIGEST                      # the contract digest the gateway must announce
$B kind validate examples/devices/thermal-board/*.kind.json | jq '.kinds[] | {kind, kind_digest}'
```

Write the three `kind_digest` values and the contract digest in the evidence sheet. The gateway must announce contract version `1.4.0` (`cat contracts/plant-wire/VERSION`) and exactly those digests.

Create the lab folders. Socket directories must be owner-only and the paths short (a socket path is at most 103 characters):

```bash
D=$(mktemp -d /tmp/bench.XXXX)                       # mode 0700
W=$(mktemp -d /tmp/benchws.XXXX)
mkdir -p "$W/domains" "$W/kinds" "$W/runs"
cp examples/devices/thermal-board/*.kind.json examples/devices/thermal-board/arduino-thermal.board.json "$W/kinds/"
cp examples/devices/thermal-board/thermal-board.domain.json "$W/domains/arduino-thermal.domain.json"
$B domain validate "$W/domains/arduino-thermal.domain.json" --kinds-dir "$W/kinds"
```

Expected: `domain validate` prints `implements_checked` with the three kinds and their digests (the twin's fan and indicator implement their semantic kind files through the board mapping). The twin's board is named `mega-01`: **give the gateway's board another id** (for example `mega-bench`), or the gateway is refused `inventory_mismatch`.

Fix the policy digest the host will carry. Ask Agentic Stream for it (`agentic-stream-host.md` asks them to expose it). For the reference host, any well-formed value works:

```bash
POLICY=sha256:$(printf 'b%.0s' $(seq 64))            # the reference host's; use Agentic Stream's when it is the host
```

## M2: the gateway on the plant wire

Exit: same `kind_digest` from gateway and simulator; real data with zero quarantines; unplugging shows `link_down` and a rebirth; capturing the connected board yields kind files equal to the defined ones.

For M2 the gateway announces all three devices (`zone-01`, `led-01`, `fan-01`; use your own ids in the inventory below). The reference host refuses a node that lacks a device of its inventory. Start the hardware-only relay and a reference host that accepts `hardware`:

```bash
echo '{"zone-01":"thermal_sensor","led-01":"indicator","fan-01":"fan"}' > "$D/inventory.json"   # your device ids
# terminal 1
$B plantwire host --socket "$D/host.sock" --environment hardware --inventory "$D/inventory.json" --delay-ms 250
# terminal 2
$B plantwire relay --gateway-socket "$D/gateway.sock" --host-socket "$D/host.sock" --out "$W/runs/m2-relay"
# terminal 3: start the gateway, dialling $D/gateway.sock
```

Run for a few minutes with the board attached, then interrupt the relay (Ctrl-C). It prints what it carried; the host prints its report when the session ends.

### M2-1 Same kind digests

```bash
$B kind capture "$W/runs/m2-relay/journal.jsonl" --kinds-dir "$W/kinds" \
  | jq '{environment, boards, issues, kinds: [.kinds[] | {kind, kind_digest, status: .comparison.status}]}'
```

- Expected: `environment` `"hardware"`; every board `hardware`; `issues` is `null` (not absent) when the capture is clean; each kind `status` `"same"`, and each `kind_digest` equal to the digests of step 0.
- Pass: three `same`. Fail: any `differs` (the first differing pointer is in `comparison.differs_at`), `new`, or an issue. Evidence: the full output.

### M2-2 Zero quarantines

- With the reference host: its report has `"quarantined": []` and no `refused`.
- With Agentic Stream as the host (repeat the relay with `--host-socket` set to its socket): its commissioning view shows the three devices discovered and no quarantine.
- Pass: no quarantine in either. Evidence: the report, a screenshot or export of Agentic Stream's inventory view.

### M2-3 Unplug shows `link_down` and a rebirth

Unplug the board's USB cable for about 10 seconds during a relay run, then plug it back. Interrupt the relay.

```bash
jq -c 'select(.type=="wire" and .direction=="node") | .record
       | select(.type=="DDEATH" or .type=="DBIRTH") | {seq, type, device, reason, boot_id}' \
   "$W/runs/m2-relay/journal.jsonl"
```

- Expected: the first `DBIRTH` of each device is the initial birth (three at the top of the output for three devices); after them, one `DDEATH` with reason `link_down` per device after the unplug, then one `DBIRTH` per device after the replug (with a new `boot_id` if the board rebooted).
- Pass: both present, in that order. Fail: devices keep reporting `good` readings while unplugged, or no rebirth. Evidence: the `jq` output and the gateway log.

### M2-4 Capture equals the defined kinds

M2-1 is this check (`kind capture` compares the board's announcement with the workspace kind files). Also run it from the canvas: `bin/streams-simulator canvas --workspace "$W" --ui-dir canvas/dist` (build `canvas/dist` first with `cd canvas && npm ci --ignore-scripts && npm run build`; the address and pairing code print on stderr), open **Capture from a board**, choose the run `m2-relay`, and read **Same** on every card. Pass: three **Same**.

## M3: relay and mixed sessions

Exit: the real board plus N simulated boards reach Agentic Stream as one node; a recorded session replays byte-identically; no replayed or simulated command reaches the gateway.

Plan B puts `zone-01` and `fan-01` on the real board and leaves `led-01` simulated (configure the gateway to announce only those two devices on board `mega-bench` for this run). Use the host's real socket (Agentic Stream) or the reference host with `--environment mixed`.

```bash
cat > "$D/plan.json" <<EOF
{"schema_version": "0.1", "domain": "arduino-thermal", "seed": 1,
 "host_socket": "$D/host.sock", "gateway_socket": "$D/gateway.sock", "pace": "wall",
 "real": [{"instance": "zone-01", "board": "mega-bench", "device": "zone-01"},
          {"instance": "fan-01", "board": "mega-bench", "device": "fan-01"}],
 "policies": ["$POLICY"]}
EOF
$B plantwire connect --plan "$D/plan.json" --check --domains-dir "$W/domains" --kinds-dir "$W/kinds" \
  | jq '{environment, instances, policies}'
```

Expected: `environment` `"mixed"`; `zone-01` and `fan-01` `real` with the kind digests of step 0, `led-01` `simulated`; `policies` holds `$POLICY`. A plan that does not hold prints the broken rule with a pointer. For M3-3 the reference host issues two scripted commands. `--script` needs a well-formed `--policy` (else exit 2, `usage_invalid`); `at_ns` is **world** time; the plan is paced, so world time is UTC from the session's start and the script below issues its commands about one minute after it is written (start the session within that minute); `id` is the command's `command_id`. Commands use the kinds' words and units; the gateway turns them into raw values through the board mapping. Then:

```bash
T0=$(( $(date +%s) + 60 ))
cat > "$D/script.json" <<EOF
[{"id": "c-led", "at_ns": ${T0}000000000, "device": "led-01", "command": "show_state",
  "args": {"state": "watch"}},
 {"id": "c-fan", "at_ns": $(( T0 + 3 ))000000000, "device": "fan-01", "command": "set_cooling",
  "args": {"level_percent": 30},
  "verify": {"metric": "level_percent", "equals": 30, "tolerance": 1, "within_ns": 5000000000}}]
EOF
# terminal 1 (reference host; or start Agentic Stream on $D/host.sock)
$B plantwire host --socket "$D/host.sock" --environment mixed --policy "$POLICY" --inventory "$D/inventory.json" \
  --script "$D/script.json" --delay-ms 250
# terminal 2
$B plantwire connect --plan "$D/plan.json" --domains-dir "$W/domains" --kinds-dir "$W/kinds" \
  --control-socket "$D/control.sock" --out "$W/runs/m3-mixed"
# terminal 3: start the gateway
```

### M3-1 One node, real and simulated devices

```bash
jq -c 'select(.type=="wire" and .direction=="node") | .record | select(.type=="NBIRTH")
       | {environment, boards: [.boards[] | {board, environment}]}' "$W/runs/m3-mixed/journal.jsonl" | head -1
jq -c 'select(.type=="origin")' "$W/runs/m3-mixed/journal.jsonl"
```

- Expected: environment `"mixed"`, one board `hardware` (`mega-bench`) and one `simulated`; one `origin` record per device with the right environment.
- In Agentic Stream: one node with all three devices, the two real ones marked hardware.
- Pass: both. Evidence: the output and a screenshot of the canvas plant with its origin marks.

### M3-2 Byte-identical replay

Stop the session (Ctrl-C on `connect`, or **Stop and save** in the canvas).

```bash
$B verify "$W/runs/m3-mixed/run.json" | jq '{matches, reason}'
$B report "$W/runs/m3-mixed" | jq '{status: .verification.status, checks: [.verification.checks[] | {name, status}]}'
```

- Expected: `matches` `true`; `files_match`; the `field_trace` check `match` (and `interrupt_log` `match` when controls were used).
- Then show the recording to a host, twice, with a host that logs the records it receives (the reference host prints only a report):

```bash
$B plantwire replay-trace --run "$W/runs/m3-mixed" --host-socket "$D/host.sock"
```

  The replay announces the node and every board `simulated`; a host set to `hardware` refuses it. Two replays against a recording host should give identical logs.
- Pass: `matches` true and the checks `match`. Fail: `matches` false (read `reason`: `file_differs` names the file). The replay-twice comparison is optional evidence; a host that cannot log is not a failure.

### M3-3 No replayed or simulated command reaches the gateway

1. During the live session, the host sends a command to the simulated `led-01` (`c-led` of the script above, or one from Agentic Stream). In the gateway log, **no** `DCMD` for `led-01` appears. In the canvas, the command is on the simulated side.
2. The host sends one command to the real `fan-01` (`c-fan`). The gateway log shows exactly that `DCMD`, with the members the host sent.
3. Stop the gateway. Run `plantwire replay-trace` as above and send a command from the host. The gateway log shows nothing, because the replay has no gateway path.
4. Pass: the gateway log has no `DCMD` for a simulated device and none during the replay.

## M4: live control plane

Exit: one reading followed from the board to Tamoz's decision on the live screen; a refused, an unconfirmed and a confirmed command told apart at a glance; the same session reopened from its recording.

Run the M3 session with Tamoz driving the host. In the canvas, open the plant, **Open live view**.

### M4-1 One reading to a decision

Pick one `temperature` reading of the real `zone-01` on the live screen (note its value and its time). Find the same reading in Tamoz's decision input and the decision that follows, and the command it leads to in the commands table.

```bash
jq -c 'select(.type=="wire" and .direction=="gateway") | .record
       | select(.type=="DDATA" and .device=="zone-01") | .metrics[]' "$W/runs/m3-mixed/journal.jsonl" | head
```

- Each `.metrics[]` entry shows the metric's **alias**, not its name: the alias maps to the metric through that device's `DBIRTH` (its `metrics` array pairs each `name` with its `alias`).
- Pass: the value and event time on the screen equal a `DDATA` in the journal and the input of a Tamoz decision, and the command that decision issued appears in the commands table with its four stages. Evidence: a screen recording or screenshots at 1440 x 900, the matching log lines.

### M4-2 Refused, unconfirmed and confirmed at a glance

Produce three commands to the real `fan-01` (`set_cooling` with `level_percent` 30 unless stated):

| Command | How | Expected stages |
| --- | --- | --- |
| Refused | Hold `fan-01` in the Interrupts region first | Receipt "Refused at the node: local control"; no board report; effect unknown |
| Unconfirmed | Send it with no verification script and no `command_evidence` | Receipt accepted; board report success; effect **unknown** |
| Confirmed | Send it from a host that verifies a metric of the fan and sends `command_evidence` in the next `HSYNC` (a mixed session: the simulated world's barrier carries it) | Receipt accepted; board report success; effect **observed** by the host's check |

- Pass: each reads differently without opening a detail, and a keyboard or screen-reader user gets the stages as words. Note that "confirmed" here is the host's own check of the board's reported value. The physical confirmation is M5.
- If the host cannot send evidence in a hardware-only session, record it as D-112; the "confirmed" row can then be shown only in a mixed session.

### M4-3 Reopened from the recording

After **Stop and save**, reopen the run from **Recent runs** (also after restarting the canvas).

- Pass: it opens labelled "Connected session", scrubbing shows the readings, the commands table and the health states of the time, and the interrupt marks if any; the plant and the origin marks are the same as live.

## M5: interrupts and real commands

Exit: a Tamoz decision drives the real fan, confirmed by the instrument rather than `DRESULT`; hold, approval and emergency stop each shown on the bench and reproduced on replay; the report names the divergences.

Start the instrument's log (`bench-instrument.md`) and the gateway log with command times. Run the M3 session again, with a new `--out "$W/runs/m5-session-1"` (a new folder for every attempt) and the same `--control-socket "$D/control.sock"`; if you use the reference host, start it with `--delay-ms 250`.

### M5-1 A Tamoz decision drives the real fan, confirmed by the instrument

- Let Tamoz issue a `set_cooling` to `fan-01` (through Agentic Stream). In the canvas the stages read accepted, board reports success.
- From the instrument file: the tachometer and the current channel show the fan running at about the commanded duty inside the lease window after the command's UTC time (from the gateway log).
- Pass: the instrument confirms the effect, recorded in the table of `bench-instrument.md`. Fail: the board reports success and the instrument shows nothing.

### M5-2 Hold, approval and emergency stop on the bench

Do each in the **Interrupts** region (or with the control socket) with a fan command from the host in flight, and watch the instrument.

1. **Hold** `fan-01`; send a fan command. Expected: the host receives `refused`, `local_control`; the fan does not run; the gateway log has no `DCMD`. **Release** the hold; the next command runs.
2. **Approve each command** for `fan-01`; send a command. Expected: the host receives `queued`; the command waits in the **waiting commands** list with its deadline; the fan does not run; **Approve**: the host then receives `accepted` and the fan runs. Send another and **Deny**: `refused`, `local_control`. Send a third and let its deadline pass: `refused`, `expired`.
3. **Emergency stop** (scope `fan-01`, then all devices) while the fan runs. Expected: the fan stops (the instrument shows it); the gateway log shows a `DCMD` `node-stop-<N>` with command `stop_fan`, empty arguments and policy digest of 64 zeros; the stop's outcome line names no device as `unreachable`; the hold is **latched**; a later host command is refused `local_control` until **Release**.

   Control socket equivalents (a client that can write a line to a Unix socket; `-w 2` closes an idle client):

   ```bash
   printf '%s\n' '{"op":"hold","scope":{"kind":"device","id":"fan-01"}}' | nc -U -w 2 "$D/control.sock"
   printf '%s\n' '{"op":"approve_mode","scope":{"kind":"device","id":"fan-01"},"enabled":true}' | nc -U -w 2 "$D/control.sock"
   printf '%s\n' '{"op":"estop","scope":{"kind":"device","id":"fan-01"}}' | nc -U -w 2 "$D/control.sock"
   printf '%s\n' '{"op":"release","scope":{"kind":"device","id":"fan-01"}}' | nc -U -w 2 "$D/control.sock"
   printf '%s\n' '{"op":"approve","command_id":"<id>"}' | nc -U -w 2 "$D/control.sock"
   printf '%s\n' '{"op":"deny","command_id":"<id>"}' | nc -U -w 2 "$D/control.sock"
   printf '%s\n' '{"op":"state"}' | nc -U -w 2 "$D/control.sock"
   ```

   `nc -U -w 2` works with the macOS and OpenBSD netcat; GNU netcat has no `-U`. Use `socat` instead: `printf '%s\n' '{"op":"state"}' | socat - UNIX-CONNECT:"$D/control.sock"`. Each request answers one line with `"result": "applied"` or `"refused"` and a `reason`. `approve` may carry the `content_digest` shown in the `pending` entry of the `state` reply. `release` must name the same scope as the hold or stop it clears: releasing `all` does not clear a hold on `fan-01`. In `state`, `devices` lists only devices with a non-default mode, so it is `[]` on a quiet session. Controls are applied at the next 250 ms tick, not instantly.
4. **Pass** per control: the instrument and the gateway log agree with the screen. Fail: the fan runs while the screen says held, or an emergency stop lists a device `unreachable` that is attached and born.

Known limits (do not count as a fail): a command the gateway had already accepted before the hold still runs (D-94); a stop is addressed to the last boot the gateway announced and a rebooted device refuses it `wrong_boot` (D-99).

### M5-3 Reproduced on replay

Stop the session. Then:

```bash
$B verify "$W/runs/m5-session-1/run.json" | jq '{matches, reason}'
$B report "$W/runs/m5-session-1" | jq '.verification.checks[] | select(.name=="interrupt_log" or .name=="field_trace")'
jq -c '.record | select(.event=="control" or .event=="verdict" or .event=="resolve" or .event=="stop" or .event=="effect")
       | {event, op, scope, enabled, command_id, device, verdict, resolution, reason}' "$W/runs/m5-session-1/interrupts.jsonl"
```

- Expected: `matches` `true`; the `interrupt_log` and `field_trace` checks `match`; the interrupt log shows your hold, release, approve-mode, approve, deny, emergency stop and the `node-stop` effect in the order you did them, with the verdicts (`pass`, `refuse`, `queue`) matching what the host received.
- Reopen the run in the canvas: the same controls appear as marks on the timeline, in the same order, and the commands table reads as it did live.
- Pass: all three. Fail: `matches` false, the log missing, or `interrupts.rejected.jsonl` in the folder (a log that did not replay; the session also failed with an error).

### M5-4 The twin report names the divergences

```bash
$B plantwire twin-report "$W/runs/m5-session-1" --domain arduino-thermal \
  --domains-dir "$W/domains" --kinds-dir "$W/kinds" \
  --tolerance zone-01.temperature=2 \
  | jq '{verdict, reason, summary, divergences: [.divergences[] | {name, first_ns, last_ns, worst_error, commands}]}'
```

- Expected: a verdict (`agrees`, `diverges`, or `cannot_judge` with a `reason`). A `diverges` verdict names each divergence as `device.metric#n` with its interval, worst error and the commands in force. For a fan run, the board cools more or less than the first-order twin predicts.
- To see a divergence deliberately, repeat with a tighter tolerance (`--tolerance zone-01.temperature=0.5`) or run the fan with the airflow blocked.
- Pass: the report runs, each divergence is named, and the owner agrees that the named intervals are where the temperature part ways with the twin. In the canvas, Report tab, **Compare with twin**, and the Inspect **Twin prediction** chart show the same.
- Read the caveats before using `agrees`: the twin starts from its domain's initial state (D-90); the default tolerance is 5 percent of the range; a run with no applied command is `cannot_judge`.

### M5-5 The policy digest check, on both sides

1. Put the host's digest in the plan's **Allowed policy digests** (or `policies`) **and** in the gateway's list. Start the session.
2. Send one command to the simulated `led-01` and one to the real `fan-01`. Both must read **Accepted** in the commands table.
3. If one reads `policy_not_allowed`, that side's list differs. The row shows the digest the host used (shortened; **Copy policy digest**). The simulated side's list is edited in the plan and needs a new session; the gateway's in its configuration.
4. After a narrowing: add the new digest on both sides, run step 2 again, then remove the old digest from both.
5. Pass: both Accepted with the right digest; both refused `policy_not_allowed` with a digest on neither list; the gateway logged its list at start.

## The evidence bundle to send back

For every check: the saved output, the pass or fail, and the time.

- The repository commit, `make build` output and the contract and kind digests (step 0).
- The gateway's announced `NBIRTH`, its start log with the policy list, and its command log (`command_id` and the UTC time handed to the firmware).
- Agentic Stream's commissioning view and the host log of the held, approved, denied and expired commands.
- The run folders `m2-relay`, `m3-mixed` and the M5 session folders (every attempt, failed ones too) as archives (`tar czf`), with `sha256sum` of each.
- The instrument raw files with hashes, calibration, the clock offsets, and the completed table from `bench-instrument.md`.
- Canvas screenshots or a short recording at 1440 x 900 of the plant with origin marks, the live view with the commands table, the Interrupts region during hold, approval and emergency stop, and the Twin vs board result.
- A list of anything that failed or surprised you, and which D-numbers in `governance/DEFERRED.md` it matches.

Do not send credentials or customer data. The run folders contain what the board reported; send them as the bench's own data.
