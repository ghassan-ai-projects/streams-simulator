# P4 runbook: the simulated rung of the fan experiment

Written 2026-10-20 by reading code, docs and the earlier run folders. **Nothing in this runbook was run as a joined
system**: no Tamoz worker, no model call, and no `plantwire connect` session was started to write it. What was run is
listed in section 11; everything else that cannot be seen from the code is marked UNVERIFIED.

Topology (PLAN section 1, "Simulate"):

```
Tamoz worker (zai/glm-5.3-flash)  <-gRPC unix socket->  Agentic Stream `serve` (branch fan-a)
                                                          = the plant wire HOST (host.sock)
                                                                      ^
                                                                      | plant wire 1.4.0
                                                                      v
        streams-simulator `plantwire connect` (branch fan-sim): paced zone + LED + fan twin, zone_heat_load,
        control socket for approve-each / hold
```

Nothing is fed to Agentic Stream by hand: the twin produces the readings (temperature and humidity every 2 s), the site
file turns them into `zone.temp.observed` / `zone.humidity.observed`, and the host emits `zone.heartbeat.observed`
every 10 s. Do not use the P0.5 `feed_telemetry.py`, `--live-socket` or `--trace`.

Kit: `/Users/ghassan/my-projects/.e2e-run/fan-p4-kit/` (called `$K`). Run folder: `$K/../fan-p4-<timestamp>/` (called
`$RUN`, created by `scripts/prep-run.sh`; one new folder per attempt). A copy of this file is in the repository at
`docs/physical-devices/fan-experiment/P4-RUNBOOK.md`; the kit copy is the one to edit.

## 1. What is in the kit

| File | What it is |
| --- | --- |
| `agentic-stream/zone-thermal-bench-fan.situation.yaml` | Copy of the bench spec (`examples/real-world-sensor/zone-thermal-bench.situation.yaml` at `fan-a`) with the experiment's edits; `spec-edits.diff` is the exact diff. Validates: spec digest `sha256:ce35becf...fc476`, policy digest `sha256:c6ed0ed6...e95aa9` |
| `agentic-stream/thermal-bench.catalog.json` | Capability catalog: routes `set_indicator` -> `show_state` (led-01, presets off/watch/alert) and `select_thermal_mode` -> `set_cooling` (fan-01, presets hold = 0 %, bounded_cooling = 45 %, bounds 0..60), safe stops `stop_led`/`stop_fan` resolved by the host to each kind's declared safe-stop command (catalog says `operation: safe_stop`, PLAN 3.2 correction). Identical to the repository fixture except fan `expires_after_ms` 120000 (D9) |
| `agentic-stream/thermal-bench.site.json` | Site file: `zone-01` -> entity `thermal_zone/zone-01`, temperature and humidity readings, heartbeat every 10 s, quality table. Byte-identical to the repository fixture (digest `sha256:34025cf9...3dce`) |
| `streamsim/session.plan.template.json` | `simulate.plan.json` with `@RUN@`, `@POLICY_DIGEST@`, `@HEAT_S@` placeholders |
| `tamoz/worker-profile.template.yml` | Worker profile: role `thermal_reasoner` -> provider `zai`, model `glm-5.3-flash` (no `base_url`, no key in the file) |
| `scripts/prep-run.sh` | Creates `$RUN`, builds both binaries into it, freezes the configuration, fills the plan, checks spec and plan. Starts nothing |
| `scripts/confirm-policy.sh` | Compares the plan's policy list with the `policy_digest` the running `serve` logged; `--fix` rewrites the plan |
| `scripts/ctl.sh` | The node's control socket from the shell (state, pending, approve-mode, approve, deny, hold, release, estop) |
| `scripts/watch.sh` | Live monitor for the operator |
| `scripts/check-run.sh`, `scripts/journal-summary.py` | Evidence for each PLAN section 7 criterion, read-only |
| `bin/` | The two binaries built while validating the kit (not used by a run: `prep-run.sh` builds fresh ones) |

No event-schema directory is needed (`--event-schemas` unused): the three event types the site file maps are in the
embedded registry (`validate` and the A4 tests use them). No `principals.yaml` is needed: the fan is R1/automatic (D9), so
nobody signs an approval; the node's approve-each is the human gate.

### The spec edits (all of them; `agentic-stream/spec-edits.diff`)

| Edit | Why |
| --- | --- |
| `select_thermal_mode`: `risk: R1`, `policy: automatic` | D9. The policy digest changes; the plan lists the new one. `riskCeiling: R2` is left as it is |
| `budget.wallTime: 300s`, `outputTokens: 20000` | P0.5 F5: `glm-5.3-flash` always reasons (20 to 80 s per call); 30 s caused `checkpoint_conflict` (F3). Kept `modelCalls: 5`, `providerRetries: 2` |
| `promptVersion: zone-thermal-bench-fan-v1` | Label |
| `executor.prompt`: the output contract of `tamoz.episode-diagnosis/v2`, the allowed actions with their parameters, one complete example | P0.5 F1, F2, F6: the model guessed a different shape in 3 of 3 answers when the frame did not state it. States: exact top-level keys, `diagnosis_probabilities` as an array covering the three catalog codes, `evidence_refs` as `fact:<id>` copied from the user message, at most one `recommended_intents` entry, values in `parameters` (Tamoz accepts only a preset named `default`, `decision_builder.rb:255-259`), no `tool_requests`/`evidence_gaps` |
| `executor.objective`: an escalation ladder | First confirmed rise (mean > 30 and slope > 6 C/h, the spec's own open condition) -> `set_indicator alert`; mean > 38 C (the spec's own `over_ceiling` threshold) or an earlier alert with the rise continuing -> `select_thermal_mode bounded_cooling`; falling -> `watch` / `hold`; link missing -> no action. It uses no twin-specific number, so the same spec serves the real rung (R1). The "never set a duty, level or duration" sentence is kept |

Nothing else in the spec changed (inputs, windows, operators, situation, triggers, diagnosis catalog, intents' schemas).

### One deliberate deviation from the committed plan

`simulate.plan.json` heats the zone for 900 s. This kit uses **1800 s** (`@HEAT_S@`, override with
`HEAT_S=900 scripts/prep-run.sh`). Reason: criterion 4 needs up to three fan proposals (approve, deny, an unanswered one that
waits 120 s). The decisions come in the first 8 to 10 minutes (section 5), but an operator needs room to act and the 120 s
expiry must fit before the heat clears. Seed, domain, `at_s` 120, `pace: wall` and everything else equal the committed plan.
Record this in the PLAN as a deviation when the run is reported.

## 2. Preconditions

- macOS, Go 1.27.2 on `PATH`, `git`, `sqlite3`, `python3`, `nc` (macOS netcat, supports `-U`), `rbenv` with Ruby 3.3.11 for Tamoz
  (`bundle check` ok in `$TAMOZ`; P0.5 recorded this on the same checkout 9170538c, `real-run`).
- Outbound network from the worker to the `zai` endpoint. The key is in `~/.tamoz/service.env` (variable `ZAI_API_KEY`, and the
  endpoint variable the worker reads). **Never print, copy or `cat` that file.** The worker shell below sources it and only
  tests that the variable is set.
- Sleep prevention for the roughly 40 minute session: run `caffeinate -i -w <pid>` against the connect process, or keep the lid
  open with power connected (a paced world, three missed barriers end it, PLAN 3.1).
- Unix socket paths stay under 103 bytes (`prep-run.sh` checks `$RUN/control.sock`).
- The streams-simulator checkout `fan-sim` has uncommitted work (47 modified files at the time of writing, among them
  `internal/plantwire/**`, `contracts/plant-wire/SPEC.md`). `prep-run.sh` builds the **committed** ref (`git archive`), so a
  fix that only exists in the working tree is not in the run. Commit first when the run must include it, and set `SIM_REF` if
  the branch name differs. `$K/bin/` was built from `fan-sim` 04439426 and `fan-a` f577ece8.
- Optional preflight (no model, no key; about a minute): the adapter's own end-to-end tests against the binary the run will use.

```bash
K=/Users/ghassan/my-projects/.e2e-run/fan-p4-kit; . "$(cat $K/LAST_RUN)/env.sh"        # after section 3
cd "$RUN/src-stream" && STREAMSIM_BIN="$SIMBIN" STREAMSIM_REPO="$SIM" go test ./cmd/agentic-stream -count=1 \
  -run 'TestThePlantWireEffectorCommandsAPacedStreamsimSession|TestTheFanExperimentCloses|TestServeHostsTwoSessions'
```

  `STREAMSIM_REPO` is the real checkout because the test compares the binary's age with the repository's HEAD commit time
  (the archive has no `.git`, the test would skip). The first test sends one command under a foreign policy digest and expects
  `refused policy_not_allowed`: that is the **negative half of criterion 5**, which a normal run cannot produce (Agentic Stream
  only ever carries its own digest).

## 3. Prepare the run folder (once per attempt)

```bash
K=/Users/ghassan/my-projects/.e2e-run/fan-p4-kit
"$K/scripts/prep-run.sh"                  # prints the new run folder and the line to source
. "$(cat $K/LAST_RUN)/env.sh"             # in EVERY shell of the attempt: sets RUN AS SIMBIN KINDS CTL POLICY_DIGEST ...
```

What `prep-run.sh` does (so it can be done by hand):

```bash
RUN=/Users/ghassan/my-projects/.e2e-run/fan-p4-$(date +%Y%m%dT%H%M%S); mkdir -m 700 "$RUN" "$RUN"/{bin,worker-root,config,domains,src-sim,src-stream}
SIM=/Users/ghassan/my-projects/streams-simulator; STREAM=/Users/ghassan/my-projects/agentic-stream
git -C "$SIM" archive fan-sim | tar -x -C "$RUN/src-sim"            # branch fan-sim, committed state
git -C "$STREAM" archive fan-a | tar -x -C "$RUN/src-stream"        # branch fan-a, committed state
(cd "$RUN/src-sim"    && go build -ldflags="-X main.Version=p4 -X main.Commit=$(git -C $SIM rev-parse --short fan-sim)" -o "$RUN/bin/streams-simulator" ./cmd/streamsim)
(cd "$RUN/src-stream" && go build -ldflags="-X main.Version=p4 -X main.Commit=$(git -C $STREAM rev-parse --short fan-a)" -o "$RUN/bin/agentic-stream" ./cmd/agentic-stream)
cp "$K"/agentic-stream/{zone-thermal-bench-fan.situation.yaml,thermal-bench.catalog.json,thermal-bench.site.json} "$RUN/config/"
cp "$RUN/src-sim/examples/devices/thermal-board/thermal-board.domain.json" "$RUN/domains/arduino-thermal.domain.json"   # the domains dir may hold only domain files
sed "s#@RUN@#$RUN#g" "$K/tamoz/worker-profile.template.yml" > "$RUN/worker-profile.yml"; chmod 600 "$RUN/worker-profile.yml"
"$RUN/bin/agentic-stream" validate "$RUN/config/zone-thermal-bench-fan.situation.yaml"   # prints digest and policy_digest (on stderr)
# plan: template with @RUN@, @POLICY_DIGEST@ (= the policy_digest line above), @HEAT_S@ (1800)
"$RUN/bin/streams-simulator" plantwire connect --plan "$RUN/session.plan.json" --check --kinds-dir "$RUN/src-sim/examples/devices/thermal-board" \
  --domains-dir "$RUN/domains" --adapters-dir "$RUN/src-sim/adapters"
```

Expected from the script (checked on 2026-10-20): `ok: zone_over_temp`, the digests in section 1, and a `--check` document with
`environment: simulated`, three simulated instances (`thermal_sensor` `7d0921b1...`, `indicator` `f10fe1d8...`, `fan` `a33eb3a8...`),
board mapping `arduino-thermal` digest `25c7a642...`, and the plan's policy list. `commits.txt` in `$RUN` records the commits, the
checkout branches and the dirty-line counts of the three repositories.

**The policy digest.** `agentic-stream validate` prints it and it is a function of the spec alone, so the script fills the plan
from there. `serve` logs the same value at start (`msg="spec loaded" ... policy_digest=...`); section 4 step 3 compares the plan
with that log line before the node starts, and `--fix` rewrites the plan if they ever differ. A plan that lists another
digest makes every non-stop command `policy_not_allowed`.

## 4. Start order and every command

Four shells, `. "$(cat $K/LAST_RUN)/env.sh"` in each. Order: worker, serve, confirm policy, node, operator.

### Shell 1: Tamoz worker (D11: `zai/glm-5.3-flash`; budget in the spec: 300 s, 20000 tokens)

```bash
set -a; . ~/.tamoz/service.env; set +a; : "${ZAI_API_KEY:?missing}"      # never echo the value; keep `set -x` off
cd "$TAMOZ" && nohup rbenv exec ruby bin/tamoz-stream-worker --profile "$RUN/worker-profile.yml" \
  --database "$RUN/tamoz.db" --tenant default --socket "$RUN/worker.sock" \
  --lane-config fast=glm-5.3-flash,deep=glm-5.3-flash,batch=glm-5.3-flash > "$RUN/worker.log" 2>&1 & echo $! > "$RUN/worker.pid"
sleep 3; ls -l "$RUN/worker.sock"
```

Fresh `tamoz.db` and `stream.db` per attempt: Tamoz memory from an earlier attempt would change what the model recalls.

### Shell 2: Agentic Stream `serve` = the plant wire host

```bash
export AGENTIC_STREAM_SUBSCRIBER_TOKEN=$(openssl rand -hex 16)      # mandatory, any value; not printed
cd "$RUN" && nohup "$AS" serve --db "$RUN/stream.db" \
  --spec "$RUN/config/zone-thermal-bench-fan.situation.yaml" \
  --worker-socket "$RUN/worker.sock" --worker-name tamoz \
  --effect-profile physical --live-actuation --owner-authorized \
  --plant-wire-socket "$RUN/host.sock" --plant-wire-environment mixed \
  --hardware-boards mega-01 \
  --plant-wire-kinds "$KINDS" --plant-wire-site "$RUN/config/thermal-bench.site.json" \
  --device-catalog "$RUN/config/thermal-bench.catalog.json" \
  --dispatch-lease 4m --listen 127.0.0.1:0 --poll-interval 1s > "$RUN/stream.log" 2>&1 & echo $! > "$RUN/stream.pid"
sleep 3; grep -E "plant wire host ready|spec loaded|serving runtime API|error" "$RUN/stream.log" | cut -c1-250
```

| Flag | Value and why |
| --- | --- |
| `--effect-profile physical --live-actuation --owner-authorized --hardware-boards mega-01 --plant-wire-environment mixed` | **D3: one `serve` for both rungs.** The consent for the named hardware board is given once, at start, so the switch to the real board is only the session plan and needs no restart (PLAN D3, R1). The simulated rung commands only simulated devices, which need no consent, so the flags change nothing for it; `mega-01` appears only in the real rung. `--plant-wire-environment mixed` accepts a node that announces `simulated` (this rung) and one that announces `hardware` (the real rung); `simulated` would refuse the real board, `hardware` would refuse this rung. The consent rule is the `device` module's (`device.GrantsActuationConsent`): the physical profile with both flags |
| `--dispatch-lease 4m` | A dispatch call ends after 0.9 of the lease: 216 s, less the answer grace (2 s) and the verification window (6 s): 208 s. The fan command can be held at the node up to its 120 s `expires_after_ms`, plus 2 s for the node's terminal answer; a call shorter than that is an unknown outcome. `serve` refuses to start when a route expires at or after the call's end: the default 1 min lease (54 s call) was refused with `route "select_thermal_mode" expires after 2m0s, a dispatch call ends after 54s` (smoke test, section 11). The minimum that starts is about 134 s |
| `--plant-wire-kinds "$KINDS"` | The commissioned kind files: `$RUN/src-sim/examples/devices/thermal-board` (the same files the node announces; the folder also holds the board mapping and domain, which the host ignores) |
| `--plant-wire-site` | The site file (outside the repository); its digest is logged |
| `--device-catalog` | The catalog copy with the fan at 120000 ms; without it the host would send no commands |
| `--plant-wire-epoch` | Not given: Unix seconds at start |
| `--worker-socket`, `--worker-name tamoz` | As in P0.5 |

Expected log lines (seen in the smoke test with a different spec path): `plant wire kind file read` x3 with the kind digests above,
`plant wire host ready contract_version=1.4.0 contract_digest=sha256:92cdaaa1... environment=mixed hardware_boards=[mega-01]
site_digest=sha256:34025cf9...`, `spec loaded ... policy_digest=...`, `serving runtime API`.

### Step 3 (any shell): confirm the policy digest

```bash
"$K/scripts/confirm-policy.sh"           # "match"; on a mismatch: "$K/scripts/confirm-policy.sh" --fix, then re-run the --check of section 3
```

### Shell 3: the node (streams-simulator `plantwire connect`)

```bash
cd "$RUN" && nohup "$SIMBIN" plantwire connect --plan "$RUN/session.plan.json" \
  --kinds-dir "$KINDS" --domains-dir "$RUN/domains" --adapters-dir "$RUN/src-sim/adapters" \
  --control-socket "$CTL" --out "$RUN/sim-run" > "$RUN/connect.log" 2>&1 & echo $! > "$RUN/connect.pid"
caffeinate -i -w "$(cat $RUN/connect.pid)" &                          # macOS: no sleep while the node runs
sleep 5; ls -l "$CTL" "$RUN/sim-run"; tail -3 "$RUN/connect.log"
```

- No `gateway_socket` in the plan: all three devices are simulated, announced as one board of environment `simulated`.
- No `--policy`: the policy list is only in the plan. No consent flags exist on this command.
- `--out "$RUN/sim-run"` must not exist yet (a used folder is refused). `--control-socket` gives the node the operator's hold /
  approve-each / emergency stop and writes `interrupts.jsonl`.
- `at_s` 120 of the disturbance counts from the host's first `STATE` (attach), so the heat load starts about 2 minutes after
  this command and lasts `HEAT_S` (1800 s).

### Shell 4: the operator (watch only: the simulated rung needs no approval)

The simulated rung needs **no human approval**: the node forwards every command to the twin at once (the approve mode
stays off), so a command is executed as soon as the model proposes it. The operator only watches. The approve, deny
and hold rehearsals belong to the real rung (P5/P6, with the owner at the board) or to a separate rehearsal run
against the simulated node.

Immediately after the node is up (first fan command is at least three minutes away):

```bash
"$K/scripts/ctl.sh" state                           # {"op":"state",...,"approve":[],"holds":[],"pending":[]}
"$K/scripts/watch.sh"                               # live: pending approvals, latest temperature, phase, episodes, commands
```

Do **not** run `ctl.sh approve-mode on fan-01` for the simulated rung: a command that waits at the node for an
approval that nobody gives expires after its 120 s deadline instead of running (the first attempt of 2026-10-20 did
that: `queue` then `refused expired`).

## 5. What to expect and what the operator does

Estimates from the domain (zone first-order lag, tau 180 s, heat load +8 C, the bench spec's 5 min windows), **UNVERIFIED** as a
timeline: times are from the host's attach.

| T+ | Event |
| --- | --- |
| 0 | Host accepts the birth; temperature about 34 C (rest value), heartbeat events every 10 s |
| 2:00 | Heat load on: zone moves toward 42 C. The window slope passes 6 C/h within about a minute; the DHT11 twin resolves 1 C |
| about 3 to 4 | Situation opens (`candidate`), `watch` one minute later (minDuration 1m), `cooling` after 30 s of slope > 9 C/h |
| about 7 | 5 min mean > 38 C -> `over_ceiling` after 15 s |
| 4 to 10 | Decisions: one per material change (watch, cooling, over_ceiling), each a Tamoz episode of 20 to 80 s. Expect LED alert first, the fan at the later ones. Total number is **not predictable**: expect 2 to 4 |
| 32 | Heat load ends (1800 s after its start); stop the session by about T+35 when the evidence is in |

Dispatch is serial (`DispatchOnce` handles one command at a time). No command waits at the node in this rung, so a
later command is never held behind an earlier one.

| Fan command | What happens | Expected |
| --- | --- | --- |
| #1 (and any later one) | nothing to do: the node forwards it | node `DRECEIPT accepted`; `DCMD set_cooling {level_percent: 45}` reaches `fan-01`, `DRESULT succeeded`; the fan reads `level_percent` 45 for the 5 s lease, then 0; Agentic Stream verifies it from that reading (6 s window) |
| any | `ctl.sh hold fan-01` before it arrives (a rehearsal, not this rung) | `refused local_control`, no queueing (criterion 6). `ctl.sh release fan-01` afterwards |

Do not feed anything, and do not touch the model's output. If a model answers in a wrong shape, that is a finding (section 9), not
something to patch around (PLAN 4.3).

## 6. Stop (order matters: node, host, worker)

```bash
kill -INT "$(cat $RUN/connect.pid)"; sleep 5; cat "$RUN/connect.log" | tail -5     # prints {out, stopped}; the run folder is complete
kill -TERM "$(cat $RUN/stream.pid)"; sleep 3                                       # the host says goodbye; the database closes
kill -TERM "$(cat $RUN/worker.pid)"; sleep 2
pgrep -fl "agentic-stream serve|tamoz-stream-worker|plantwire connect" || echo none-left
```

The worker may print a gRPC "can't be called from trap context" / `completion_queue.cc` message on stop (cosmetic, P0.5).
Stopping the node ends the session: any command still waiting at the node is `session_ending` and Agentic Stream records it
as failed; stop only after the last evidence is in.

## 7. Export and verify

```bash
AS_SHA=$(git -C "$STREAM" rev-parse --short fan-a)
"$AS" export-run --db "$RUN/stream.db" --tenant default --output "$RUN/artifact" --run-id "$(basename $RUN)" \
  --provider zai --model glm-5.3-flash --git-commit "$AS_SHA" --policy-digest "$POLICY_DIGEST" --scenario-seed 1 \
  --device-id zone-01 | tee "$RUN/export.log"
"$AS" verify-run "$RUN/artifact" | tee "$RUN/verify.log"; cat "$RUN/artifact/verdict.json"
"$SIMBIN" verify "$RUN/sim-run/run.json" | tee "$RUN/sim-verify.log"
"$SIMBIN" report "$RUN/sim-run" --format markdown --out "$RUN/sim-report.md"
"$K/scripts/check-run.sh" | tee "$RUN/check.txt"
du -sh "$RUN/sim-run"
```

`--device-id` names the device the artifact describes: several devices reported state in this experiment, and the
export refuses to guess (`manifest device_id is required when multiple device states exist`). `verdict: pass` is a
safety verdict only and is `pass` with zero decisions and zero commands (P0.5 F4): cite the criteria below, never the
verdict alone. `export-run`/`verify-run` have only been seen on device-wire runs: UNVERIFIED on plant wire commands.
Open the database without `-readonly`, and only after `serve` stopped (a WAL database with no `-shm` file cannot be opened
read-only).

## 8. PLAN section 7 criteria: where to look and what passes

`scripts/check-run.sh` prints one numbered block per criterion (`C1` to `C6`) and a liveness block; `C7` is the verify output.

| # | Criterion | Evidence (command) | Pass |
| --- | --- | --- | --- |
| 1 | Host accepts the birth; zero quarantines; readings every 2 s or faster | `grep "plant wire" $RUN/stream.log`; `select count(*) from event_quarantine`; the gap query in C1; `journal-summary.py $RUN/sim-run` wire counts | log shows the session admitted (no `environment_mismatch`/`contract_unsupported` refusal); `quarantined` 0; `zone.temp.observed` mean gap about 2.0 s (DHT11 dropouts add a few gaps of 6 to 8 s, `availability.mttr_s` 6); `zone.heartbeat.observed` present; every `node NSYNC` has a `host HSYNC` and there is one `node NBIRTH` |
| 2 | The disturbance raises the zone | C2: mean temperature per minute from `event_log`, and the `situations` phase | about 34 C before T+2:00, rising to about 40 C by T+8 and 42 C at the plateau; the situation reaches `watch`/`cooling` (a rise of more than 4 C) |
| 3 | `set_indicator` executed (R1 automatic), receipt accepted, result succeeded | C3: `intents` (`set_indicator`, `R1`, `approved`, `automatic_r0_r1`), the `commands`/`outcomes` row (`succeeded`, `plantwire.reading` verification, metric `state`, observed `alert`), the DCMD/DRECEIPT/DRESULT list in C4 | one row each; the wire shows `DRECEIPT accepted` then `DRESULT succeeded` for `led-01 show_state {state: alert}` |
| 4 | `select_thermal_mode bounded_cooling` waits at the node, then `queued` -> `accepted`, `DCMD set_cooling {level_percent: 45}` reaches `fan-01`, result follows, reading 45 then 0; a denied one is `refused local_control`; an unanswered one is `expired` after 120 s | C4: the DCMD list with node answers and timestamps, the fan `level_percent` change list, and the `interrupts.jsonl` verdict/resolve lines (`queue`, then `forward` or `refuse` with reason `local_control` or `expired`) | approved: answers `queued`, `accepted`, `DRESULT succeeded`; the fan goes 0 -> 45 -> 0 about 5 s later; `outcomes` row verified by `level_percent` 45. Denied: `refused` reason `local_control`. Unanswered: `refused` reason `expired` at the command's own `deadline_ns` (120 s after Agentic Stream created it), command `failed` in Agentic Stream |
| 5 | The policy digest is accepted by the plan's list; another digest is refused `policy_not_allowed` | C5: plan list, `policy_evaluations.policy_digest`, `serve`'s log value; `grep -c policy_not_allowed` in the journal | all three equal, no `policy_not_allowed` in this run. The negative half is the preflight test of section 2 (`TestThePlantWireEffectorCommandsAPacedStreamsimSession`), because Agentic Stream cannot be made to send a foreign digest |
| 6 | A hold gives `refused local_control`; Agentic Stream does not retry in a loop | C6: `ctl.sh hold fan-01` in the transcript, the `interrupts.jsonl` verdict `refuse`/`local_control`, then commands per route and DCMDs on the wire | exactly one `DCMD` per fan command (no repeated sends); `commands` rows per route equal the DCMD count; the fan command is `failed`, not retried; no new fan command appears within a minute of the refusal other than from a new decision (rate limit is 6 per hour per intent type) |
| 7 | `streams-simulator verify` on the run folder matches | `$RUN/sim-verify.log`; `sim-report.md` (checks `interrupt_log`, `journal`) | `matches: true`. The operator's controls and the node's decisions replay from `interrupts.jsonl` |

Always also record the liveness block: situations, episodes, attempts, decisions, intents, commands, outcomes, verifications,
device reconciliations; the episode attempts' statuses (a `timed_out` attempt or `checkpoint_conflict` means the budget or the
model, P0.5 F3/F5); `decisions.validation_status` (`accepted`); `episode_rejections` 0; `grep -ci error worker.log stream.log`.
Also keep `$RUN/commits.txt`, `validate.log`, `plan-check.json`, `export.log`, `verify.log`.

Not claimed (PLAN section 7): a measured temperature drop (the 5 s fan burst lowers the zone by about 0.25 C), independent
confirmation of the fan, any fault matrix.

## 9. Findings to expect and how to classify them (PLAN 4.3)

| Symptom | Class |
| --- | --- |
| `serve` refuses to start (`--dispatch-lease`, site, kinds, socket not owner-only) | Our setup: fix the flag |
| `environment_mismatch`, `contract_unsupported` at the birth | Contract folder drift between the two builds: record the three commits; not a model matter |
| `quarantined` > 0 | Site file or event schema: read `event_quarantine.reason_code` |
| Episode `failed` with a worker terminal, `decisions` empty | Model shape or budget (P0.5 F1/F5): read the Tamoz checkpoint `repair_directive` (`tamoz.db`, `tamoz_checkpoints`) and the response in `tamoz_effect_attempts.result`. Record the answer; do not patch the prompt inside the same attempt |
| The model proposes only `set_indicator` (never the fan) or only the fan | Model outcome with the ladder in the spec; try the attempt again once and record both |
| A command refused `policy_not_allowed` | The plan's list lacks the digest the host sent; run `confirm-policy.sh` |
| The fan verified `pending`/manual review although `DRESULT succeeded` | Verification window (6 s) missed the 45 reading; compare `journal-summary.py` fan changes with the `DRESULT` time |
| `session_ending`, unknown outcome | The node was stopped with a command waiting or in flight |
| `checkpoint_conflict` | A model call outlived `wallTime` (300 s): P0.5 F3, Tamoz/Agentic Stream latent defect |

## 10. What a later owner run with the canvas approve-each adds

The canvas is the same node plus a human at the buttons. It does not change Agentic Stream or Tamoz.

- **Start the canvas** (after `npm ci --ignore-scripts && npm run build` in `$SIM/canvas`, or use the existing `canvas/dist`):

  ```bash
  mkdir -p $W/{domains,kinds,runs}; cp "$KINDS"/*.kind.json "$KINDS/arduino-thermal.board.json" $W/kinds/
  cp "$RUN/domains/arduino-thermal.domain.json" $W/domains/
  "$SIMBIN" canvas --workspace $W --ui-dir "$SIM/canvas/dist" --adapters-dir "$RUN/src-sim/adapters"   # address and pairing code on stderr
  ```

  Pair the browser once; **Connected session** in the Run panel takes the host socket (`$RUN/host.sock`), no gateway socket,
  the allowed policy digest (`$POLICY_DIGEST`), **Check the plan**, **Start the session**. The canvas starts `plantwire connect`
  itself, paced, and keeps its own run folder (`c-<8 hex>`), so Shell 3 and `ctl.sh` are not used.
- **The Interrupts region** replaces `ctl.sh`: **Approve each command** on `fan-01`, then Approve / Deny per queued command (it
  shows the args and the content digest), **Hold**, **Emergency stop**; the commands table shows "Refused at the node: local
  control" distinct from other refusals; the live view shows the readings, link health and the `level_percent` stage by stage.
  Everything is recorded in `interrupts.jsonl` as with `ctl.sh`.
- **Gap for the simulated rung with the canvas (UNVERIFIED, from reading the bridge and canvas sources):** the canvas plan builder
  has no `disturbances` member (no match for "disturbance" in `internal/canvasbridge` or `canvas/src` outside the report view), so a
  canvas-started simulated session has **no scripted heat load** and the situation never opens. Options for the owner: run the
  simulated rung by hand as above (this runbook), or have the simulator add the member to the Connect panel first (a simulator
  item, PLAN 8: the canvas computes nothing the binary owns). On the real rung the disturbance is physical (O5), so the canvas
  fits there.
- **Switch to the real board (D3), the only differences:** the plan gains `gateway_socket` and `real`
  (`real.plan.json`). Agentic Stream keeps running: its consent for the named board was given once, at its start
  (`--effect-profile physical --live-actuation --owner-authorized --plant-wire-environment mixed --hardware-boards mega-01`,
  section 4), so **no `serve` restart is needed**; the simulated rung is unaffected by the consent because simulated
  devices need none. `--device-firmware-digest` is not used with plant wire; the gateway's allow-list takes the same
  `$POLICY_DIGEST`; fan `expires_after_ms` stays 120000; the 5 s lease is the firmware's.
- The owner also decides the `hold` and `estop` rehearsals on the real rung and the e-stop (O2); the canvas lists
  `stopped`/`skipped`/`unreachable` for an emergency stop and does not show a clean stop when a device is skipped.

## 11. What was verified while writing this, and what was not

Verified (2026-10-20, nothing of the joined run was started):

- Both binaries build from the committed refs by `git archive` (`fan-sim` 04439426 streams-simulator, `fan-a` f577ece8 agentic-stream).
- `agentic-stream validate` accepts the spec copy; `streams-simulator plantwire connect --plan ... --check` accepts the plan
  (done by `prep-run.sh` into a throwaway folder, then removed).
- A smoke start of `serve` alone (6 to 8 s, no worker, no node, throwaway database in `/tmp`, ended by `timeout`) with the flags
  of section 4 as they stood before the D3 change (`--effect-profile emulator --plant-wire-environment simulated`, no consent
  flags): it started, read the three kind files and the site file, announced `plant wire host
  ready` (contract 1.4.0, digest `sha256:92cdaaa1...`, `simulated`), logged `spec loaded` with a `policy_digest` equal to the
  `validate` value (`confirm-policy.sh` printed `match`), and the default `--dispatch-lease` was refused with the message
  quoted in section 4. **The D3 flag combination of section 4 (`--effect-profile physical --live-actuation
  --owner-authorized --plant-wire-environment mixed --hardware-boards mega-01`) is UNVERIFIED**: it is the combination the
  first run uses, so the run itself is its check (the host must announce `environment=mixed hardware_boards=[mega-01]`).
- `ctl.sh` and `journal-summary.py` against a fake control socket and a synthetic journal; `check-run.sh` against the SQL
  of an old P0.5 database (syntax and columns).

**UNVERIFIED** (the experiment itself, never run together):

1. The whole chain: Tamoz <-> `serve` <-> node. No model call was made; whether `glm-5.3-flash` returns the contract shape with this
   prompt, in how many calls and in how many seconds (P0.5 afterH0 with the sim-spec variant: one call, about 20 s).
2. Tamoz accepting the changed intent catalog (the fan at R1) and the longer prompt: its catalog digest check is against the wire the
   spec produces, so it should; not run.
3. The number and order of decisions, the escalation from LED to fan (it relies on the facts rule `mean > 38` and, for earlier
   escalation, on recalled experience that may be empty), and therefore how many fan commands one attempt gives.
4. The timeline table in section 5 (derived from the domain model, not observed).
5. `export-run`/`verify-run` over plant wire commands; `streams-simulator verify` of a session with a control socket and a paced
   world (documented as supported).
6. The exact `provider_result_json` and `interrupts.jsonl` shapes used by `check-run.sh` (read from code and the schemas: reason
   words `local_control`, `expired`, `session_ending`; `plant_wire.answers`), and the journal `at_ns` stamps used by
   `journal-summary.py`: adjust the greps after the first run.
7. The negative half of criterion 5 (the preflight test was not run by the author).
8. The Tamoz checkout had 2 uncommitted lines when last read (`commits.txt` records the count per run); the worker runs from the
   repository, not from an archive (Ruby bundle).
9. The canvas path (section 10): the Connect panel fields and the missing `disturbances` member were read, not exercised.
10. `caffeinate` against a lid-closed sleep; a 30 to 40 minute paced session length on this Mac.
