# S review B: adversarial review of the S rounds (paced sessions, semantic kinds, disturbances)

- **Reviewer:** B (adversarial). Read-only on the repository.
- **Subject:** `fan-sim`, `git diff 5951812d..0065a1e5 -- . ':!docs'` (S1a, S2, S3, S4a, S4b1, S4b2, S4c, S5a, S6, S7, S10).
- **Method:** read the diff against PLAN (D2, D7, S5) and SPEC 1.4.0 section 15; ran the built binary (`plantwire host` and
  `plantwire connect`, node and `--check`) with hostile scripts and doctored board files; wrote one throwaway Go test in a
  scratch copy of the tree (late gateway attach under `pace: wall`); ran 49 hand-made mutants of the new checks against
  `go test ./internal/plantwire/...` in scratch copies (the repository was not edited). Scratch files:
  `/private/tmp/claude-501/-Users-ghassan-my-projects-streams-simulator/517cecf1-e1a3-4525-abf0-112e2348451e/scratchpad/{mutate.py,mutate*.out,p/,mut*/}`.
- In file references `P/` is `internal/plantwire/internal/`.
- Things I tried that held: a raw command name sent by a host is refused `unknown_command` (c1 `set_pwm_lease`, c7 `set_led`); a
  raw value in a kind word is refused (`show_state solid`); a bool, a string, a negative-zero and a subnormal level are refused or
  mapped to a safe raw value; the lease is injected by the mapping and a host-supplied `lease_ms` is refused; table words and
  raw bounds are checked at load (mutants M10, M11, M13, M14 die); disturbances on a real instance, on an unknown fault, with a
  duration on a pulse fault are refused; 64 disturbances, 30-day offsets, duplicate and overlapping disturbances, severity
  1e308 and a duration of 1e-10 s all run without a crash; the schedule is absent from the host's records, the check
  document and the observer journal.

## Summary

| # | Severity | Finding |
|---|---|---|
| 1 | High | The board mapping is in no announced, recorded or compared digest: a doctored mapping passes `--check` with the same kind digests and different behaviour |
| 2 | Medium | The semantic argument check is the only thing that stops extra and non-integer arguments, and no test fails without it; the translator silently drops what it does not consume |
| 3 | Medium | The wall pacer is anchored when the relay opens: a late host or gateway gets a catch-up burst and has the disturbance schedule partly consumed |
| 4 | Medium | The reference host cannot renew a lease: identical commands are replayed with no effect and the host reports `succeeded` |
| 5 | Medium | `sealed/journal.jsonl` holds the disturbance schedule in the open while the run is live |
| 6 | Medium | A reference host's 1 Hz heartbeat kills the host when the node is quiet for about 110 s (a paced slice of 200 s reproduces it) |
| 7 | Low | The wall pacer reads the wall clock, not a monotonic one (backward step stalls the world) and a late world accepts expired commands |
| 8 | Low | A scaled value is not rounded to the raw integer type: 4.35 x 100 is refused by the twin, and may be truncated by a gateway |
| 9 | Low | The silent-gateway watchdog restarts at every host re-handshake; a flapping host holds a dead gateway indefinitely |
| 10 | Low | The lease is invisible on the wire (it lives in the mapping only); a host cannot learn that `set_cooling` lapses after 5 s |
| 11 | Low | Smaller points: tiny durations, two metrics on one raw channel, one kind on two board types, `PeerGone` ends the host "cleanly", node-only e-stop waits a slice |
| 12 | Info | Test gaps: mutants that survive |

---

## 1. High: the board mapping is not bound into any digest

**Where.** `P/domain/board_files.go:148` computes `board_digest` (`withBoardDigest`) and `P/app/kind_files.go:42` throws the
summary away (`boards, _, err := domain.ParseBoardFiles(...)`). Nothing else uses it: not `NBIRTH`, not the check document
(`implements_checked` holds kind digests only, `P/app/connect.go:125`), not the `session` journal record, not `TwinInfo`
(`P/app/twin_report.go`), not the relay's `Collision` (`P/domain/relay_mixed.go:103`, kind digests only). SPEC section 15
defines the digest and then nothing reads it. The kind digest covers the kind file, and the kind file says *what* a command is;
what it *does* (the table rows, the lease, the scale) is only in the mapping.

**Reproduction (run).** Copy `examples/devices/thermal-board/*.kind.json` and the board file to a scratch directory, then edit
the board file: `show_state` table `off` -> `{brightness_permille: 1000, pattern: solid}` and `alert` -> `{0, off}` (the LED
is lit for "off" and dark for "alert"), and `set_cooling.lease.ms` 5000 -> 10000. `streamsim plantwire connect --plan ... --check
--kinds-dir <scratch>` passes and prints the *same* three `kind_digest`s as with the shipped file
(`7d0921b1...`, `f10fe1d8...`, `a33eb3a8...`). A real LED behind a gateway that holds the shipped mapping and the twin that
holds the doctored one are announced identically and merged without a collision.

**Why it matters.** The promise of D7 and SPEC 15 is "a twin and a gateway read the same file, so the twin and the real board
cannot drift apart". Nothing checks that they did read the same file: the gateway is another repository with its own copy.
The only cross-check, `board_scale_mismatch`, covers one case (a command scale against the metric scale of the same raw name);
the command table and the metric table of an enum (`off -> pattern, pattern -> state`) are never compared for a round trip, so
a mapping that reads back a different word than it wrote validates. The run is then not a pure function of its recorded
inputs either (AGENTS.md): the raw commands a run produced depend on a file the run does not name.

**Fix.** (a) Put `board_digest` of each used board type into `NBIRTH.kinds[]` (or `node_metrics`) next to `kind_digest`, make
`Collision` and the gateway admission compare it, and record it in the `session` record, `implements_checked` and
`TwinInfo`. (b) Add a round-trip rule: for every enum command table and metric table on the same raw name, the word written
by row `w` must read back as `w` (`board_roundtrip_mismatch`). (c) Test that two mappings differing only in a row or a lease
produce different check documents.

## 2. Medium: the kind-schema check is untested and the translator drops arguments

**Where.** `P/app/commands_mapped.go:18-28` (`checkArgs`) and `P/domain/board_translate.go:25-70` (`RawCommand`).
`RawCommand` reads only the arguments the mapping names (`from`, `by`); every other member of `args` is ignored. The only thing
that refuses `{"level_percent":30,"lease_ms":99999}`, `{"level_percent":30.5}` or `{"level_percent":30,"extra":{...}}` is
`wire.CheckArguments` against the kind file's `args_schema`.

**Reproduction (run).** A binary with that one call removed (scratch copy, mutant M02) accepts all three and the reference host
reports `accepted` and `succeeded` for each (c3, c5, c16); with the call present all are `refused:out_of_range`. The mutant
survives the whole `./internal/plantwire/...` suite (and so does M01, `identify` without `checkArgs`): no test sends a kind-
forbidden argument that still translates.

**Why it matters.** The host believes it set `lease_ms` 99999 (c5) or a half percent (c3); the node ran 5000 ms and 30 %. In a
mixed session the same gap in the gateway's Python translator would act on a real fan; this repository is the reference for
that behaviour and does not pin it.

**Fix.** `RawCommand` returns `ErrNotTranslated` for any argument not consumed by `from`, `by` or the table (defence in depth,
no schema needed); add node-level tests for an extra argument, a non-integer where the kind says integer, and a string for a
number, each expecting `refused:out_of_range`. Also see mutant M12 (section 12): a mapping that does not set a required raw
argument validates and only fails at the first command.

## 3. Medium: the pacer is anchored when the relay opens, not when the world starts

**Where.** `P/app/connect_run.go:45-46` (`p.pacer = uds.StartWallPacer()` at the top of `serve`, before the sockets, the run and
the relay exist); the simulated node only begins when the first host record is delivered (`P/app/relay_sim.go:171`,
`s.begin()`); SPEC/CLI text: "for a mixed session, when the relay opens".

**Reproduction (run, scratch test in a copy of the tree).** Paced mixed session, slice 500 ms, `time.Sleep(4s)` between
`startConnect` and the gateway dial: 12 barriers arrived in 1.99 s after the attach (6 s of world time in 2 s), and the first
barrier's world time lagged the wall by 5.50 s. The disturbance schedule is also measured from the anchor, so a plan with
`at_s: 120` and a host that attaches at 110 s shows the host a world whose disturbance starts 10 s after it connects.

**Why it matters.** (a) For the length of the catch-up the host sees simulated readings stamped in the past next to real
readings stamped now, in the one session whose purpose is to compare them. (b) `Freshness` runs on the world clock
(`P/app/commands.go:96`, `P/domain/commands.go:23`), so while the world lags by L a command whose `deadline_ns` expired up to
L ago is accepted by the twin and would be refused by the gateway. (c) Disturbance times become relative to "operator opened
the relay", not to anything the experiment controls.

**Fix.** Anchor on the first `Release` (set `WallStartNS` and the recorded world start together when the first host record
arrives; the run is created later or its start patched before the first slice), or announce the lag (`node_metrics.lag_ms`)
and refuse commands while it is above a bound; at least say it in the SPEC. Add a paced test with a late attach.

## 4. Medium: the reference host cannot renew a lease and says it did

**Where.** `P/host/commands.go:35`: `key := domain.Command{Device, Name, Args, Boot}.Content()`, i.e. the idempotency key is the
content digest, so two scripted commands with the same arguments have the same key. SPEC 9 (new text in 1.4.0) says a repeated
key with the same content replays the stored receipt and result with no second effect.

**Reproduction (run).** Script `r0..r3` = `set_cooling {level_percent: 30}` on `fan-01` at +3, +6, +9, +12 s against the
simulated fan. The host report: all four `accepted`, `succeeded`. The fan metric stream: 30 at +4, +6, +8, then 0 from +10 s
on (the 5 s lease of `r0` lapsed; `r1..r3` were replayed, not executed). With a real fan the same script lets the lease
lapse while the report says four successes.

**Why it matters.** The fan experiment is a leased actuator (D6 option A: bursts, renewal later), and the instrument's own
host reports success for renewals that did nothing. SPEC now documents the property; the reference host that ships with the
instrument still falls into it, and a script author has no way out (the key is not a script member).

**Fix.** Derive the reference host's key from the command id (a fresh key per intent), or add an optional script member
`key`; make the node's replay of a command whose effect carried a lease distinguishable (`result` stays, but the journal and the
ledger mark `replayed`), so the report cannot show a replay as an effect.

## 5. Medium: the sealed journal is readable while the run is live

**Where.** The director journal is `sealed/journal.jsonl` in the run folder; `P/app/connect_disturbances_test.go:63` asserts
it holds the disturbance (fault id, entity, onset, severity).

**Reproduction (run).** A paced-off connect with `{"at_s":5,"instance":"zone-01","fault":"zone_heat_load","severity":0.5}`:
8 s into the run `grep -l zone_heat_load -r <out>` lists `<out>/sealed/journal.jsonl` (mode 0600, directory 0700, same user).
`run.json` (command log) appears at the end.

**Why it matters.** S5 exists so that a consumer meets a disturbance it was not told about. Agentic Stream as the host on the
same machine and user, with a shell tool, can `cat` the schedule. "Sealed" is a path convention, not a boundary; the
documentation says the schedule is "never written to the observer journal", which is true, and does not say it is plain text
beside the run.

**Fix.** Either hold disturbance records out of the live director journal until the reveal (the command log keeps replay), or
state the threat model in the CLI reference (the host must not have read access to the run folder), and have the canvas
launcher create the run folder outside the host's workspace.

## 6. Medium: the host heartbeat kills the reference host when the node is quiet for about 110 s

**Where.** `P/uds/listen.go:93-104` (`turn`), `P/uds/link.go:76-90` (`Send`, write deadline `l.wait` = 10 s), `P/app/session.go:263`
(heartbeat 1 s), `P/uds/link.go:41` (incoming channel of 64). The node reads the host link only inside a barrier wait; during
a paced `Release` it reads into a 64-slot channel and then stops reading.

**Reproduction (run).** `plantwire host` plus `plantwire connect` with `pace: wall` and `--slice-s 200`: the host exits about 2
minutes in with `plantwire: write: write unix ...: i/o timeout` (`{"error":{"code":"plantwire_failed",...}}`), the node keeps
running with no host.

**Why it matters.** The default slice is 1 s, so the shipped canvas does not hit it; but any quiet stretch of the node (a long
slice, a stop of the process, a suspend) now ends the reference host, where before S10 it survived. The failure is a write
timeout from a *heartbeat*, not from anything the node owed the host.

**Fix.** Make a heartbeat send non-blocking (drop the beat when the previous write has not completed, never block the loop),
or bound `--slice-s` for paced sessions to the heartbeat window, or have the node drain and discard queued STATEs while
releasing.

## 7. Low: pacer clock source

`P/uds/await.go:109-120` and `P/domain/pace.go:17-21`: `WaitNS` is computed from `time.Now().UnixNano()` each slice. A backward
step of the system clock of S seconds makes the next release wait S seconds more with no cap (the barrier cap does not apply to
`Release`); a forward step releases every missed slice at once. Use `start + time.Since(startMonotonic)` for the elapsed time
and keep the wall value only for the recorded start. Not run (argument from the code).

## 8. Low: no rounding to the raw integer type

`P/domain/board_translate.go:63,88-93` (`exact`): `value * scale` becomes an integer only when exactly integral. With a kind
argument of type number and a raw integer argument, `4.35 * 100 = 434.99999999999994` and `0.07 * 100 = 7.000000000000001`
(computed) are non-integers, so the twin refuses a value the kind accepted (`out_of_range` after `identify` passed), and a
gateway that rounds or truncates would act on 435 or 434. The shipped fan uses an integer kind argument, so it is latent. Fix:
round to the raw argument's declared integer type when within 1e-9, refuse otherwise at load (`commandFits` could compare the
kind argument's `type`/`multipleOf` with the raw type).

## 9. Low: the watchdog restarts at every host handshake

`P/domain/relay_liveness.go:12` (not counted while the host is away or unshaken) and `P/domain/relay_host.go:148-149` (first
STATE of a new host session sets `gateway.silent = 0`). Intended, tested (relay_absence tests), and the host-less time is
bounded by the 10 s host handshake wait. But a host that reconnects every 20 s gives a silent gateway a new 30 s window each
time, so a hung gateway (connected, saying nothing) is never closed while the host flaps, and while the host is present the
real devices' last readings are served as current for up to 30 s. Fix: keep a monotonic gateway-silence counter that the host
cannot reset; reset only on a gateway record.

## 10. Low: the lease is not on the wire

PLAN 4.2 and S4 say "declared lease" in the kind; the shipped contract has none (the deviation is recorded: "the lease lives in
the board mapping only"). A host reading `NBIRTH.kinds` sees `set_cooling {level_percent}` with nothing saying the effect lapses
after 5 s (and `lease_ms`, which the raw interface showed, is gone). Together with finding 4 a host cannot tell that it must
renew. Fix before D6: a `lease_ms` (or range) member in `KIND` for commands, validated against the mapping's `lease.ms`.

## 11. Low, small points

- `P/app/disturbances.go:27` uses `ClearNS == 0` as "no clear"; `duration_s: 1e-10` rounds to 0 ns (`P/domain/disturbances.go:44`)
  so the fault is injected and cleared at the same instant, and the check accepts it (the schema only has `exclusiveMinimum 0`).
  Run: command log `fault.inject` then `fault.clear` at the same `at_ns`. Reject durations that round to 0 in `CheckPlan`.
- `P/domain/board_translate.go:108-116`: `SemanticReading` returns the first (sorted) metric for a raw channel; two metrics on
  one raw channel are valid, and the second never gets a sample (`warming_up` forever). Add a board rule.
- `P/domain/board_files.go:111-116`: a kind may appear in only one loaded board mapping, so the same `fan` on two board types is
  refused (`board mappings map one kind`). Fine for one board; blocks the second.
- `P/app/session.go:266-268` `endsHostCleanly`: a host whose node dies mid-session (EPIPE, ECONNRESET) exits 0 with a report that
  looks like a normal end. Add a report member `ended: node_gone`.
- Node-only sessions take controls at a barrier, so an e-stop of simulated devices waits up to a slice (documented in
  S-FINDINGS; with `--slice-s 60` it is a minute). The physical e-stop is the relay's 250 ms tick, which is unaffected.
- `P/domain/run_folder.go:27`: a folder whose only content is `sealed/` with unknown names passes `CheckFreshRunFolder`; the
  check and the create are not atomic (two concurrent connects to one new name).

## 12. Tests that pass with the check removed

49 mutants, one edit each, `go test -count=1 ./internal/plantwire/...` (45 s per run). Survivors (the suite is green with the
check gone):

| Mutant | File | What is gone | Consequence |
|---|---|---|---|
| M01 | `P/app/commands.go:87` | `identify` no longer calls `checkArgs` | see finding 2 |
| M02 | `P/app/commands_mapped.go:22` | kind-schema validation of a mapped command | see finding 2 (reproduced end to end) |
| M12 | `P/domain/board_fit_commands.go:96` | `covered`: every required raw argument is set by the mapping | a mapping that omits one validates and fails at the first command |
| M41 | `P/domain/board_project.go` (`unmapped`) | a mapping entry for a command or metric the kind file lacks is refused | a stale entry passes (unreachable, so low impact) |
| M48 | `P/app/devices.go:122` | `wire.Finite(value)` before a reading is carried | a NaN or Inf reading would reach `json.Marshal` of the `DDATA`; untested |
| M49 | `P/app/session.go:266` | `uds.PeerGone` in `endsHostCleanly` | the clean end on EPIPE/ECONNRESET is untested |
Every other mutant is killed (M03-M11, M13-M40, M43-M47, M50; M22, which removes the kind-file comparison, is killed by the safe-stop drift test): the raw-name path, the lease injection, scale, table, metric
division, drift comparison, bounds fit, the five board rule keywords, the pacer wait and the release call, the watchdog guard,
the real-instance, pulse and onset disturbance rules, injection, the fresh-folder check, the heartbeat, the beat-before-accepted
guard, the merged environment and clock mode, the semantic reading and the NBIRTH kinds.
