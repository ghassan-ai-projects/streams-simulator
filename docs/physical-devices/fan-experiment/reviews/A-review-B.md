# A rounds review B (adversarial and safety)

Range `git -C W log 6a8ba9cb..f577ece --no-merges`: A0 e28b1ee4, H0-fix be5f237a, A1a/A1b c39e4800, A2 b44ff80e, A3a e965dd9a,
A3b f35a5faa, A3b2 ce02d38a, A3c c851548f, A3d 7cb7c90c, A4 f577ece8. Read as
`git -C W diff 6a8ba9cb..f577ece -- . ':!internal/spec' ':!internal/replay'`. W =
`/private/tmp/claude-501/-Users-ghassan-my-projects-streams-simulator/517cecf1-e1a3-4525-abf0-112e2348451e/scratchpad/wt-fan-a`
(branch `fan-a` of `/Users/ghassan/my-projects/agentic-stream`). Paths below are relative to W. Read-only review, 2026-10-20.

Scope note: the worktree moved while this review ran (HEAD `20312a1e`, plus uncommitted edits to `effector.go`,
`internal/authority/{operations,internal/app/service}.go` and a new `app/waiting.go` that renews a target claim during a long wait).
Everything below is against the stated tip `f577ece8` (`git show f577ece8:<path>` where the tree was dirty); the post-tip work is
excluded, and it is the fix for review A's H1, so that High is not repeated here.

Verdict: one High (consent), six Medium, seven Low. The safety spine is genuinely good: the origin rule, read-back verification, the
unresolved-command gate on reconciliation clearing, the not-sent/unknown split and the contract conformance are enforced in the right
layers and have tests that would fail if a guard were deleted. The High is not a missing rule; it is a rule evaluated on a snapshot
taken before a multi-transaction claim/bind path, i.e. not at the moment of transmission.

## Findings

### H1. High: consent and `--hardware-boards` are checked against an origin snapshot read before the claim/bind, and the DCMD is sent later without re-checking; a rebirth in between moves the device to another board

- Evidence: `internal/plantwire/internal/app/effector.go:58-72` (`prepare` reads `view` and `view.Origin` through the session loop,
  then returns), `effector.go:75-80` (`checkOrigin` uses that snapshot), `effector.go:48-52` (the claim and bind run after the check,
  the send after them), `internal/plantwire/internal/domain/command.go:95-108` (`Command` re-checks only the boot at send time),
  `internal/plantwire/internal/domain/conversation.go:141-150` (a rebirth is applied without `Admit`), and
  `internal/plantwire/internal/domain/ingest.go:45-54` (`rebirth` replaces `g.boards` and clears every device, so a device can be
  re-born on a different board while keeping the boot).
- Scenario: host started `--effect-profile emulator --plant-wire-socket ...` (so `PhysicalConsent == false`, no `--live-actuation`,
  no `--owner-authorized`) and the node admitted as `simulated` on board `sim-1`. Dispatch `set_indicator` for `led-01`: `prepare`
  reads `Origin{sim-1, simulated}`, `checkOrigin` passes. While `admit` runs (device-state write, command digest, claim, bind, assert —
  five DB transactions of a few ms), the node sends `NBIRTH`+`DBIRTH` moving `led-01` to board `hw-9` with `environment: hardware`
  (same `boot_id`, so the send-time boot check passes) and the loop applies it. `h.command` transmits the `DCMD` with no further
  `Authorizes` call: a process without the physical profile, without `--live-actuation`, without `--owner-authorized` and without
  `hw-9` in `--hardware-boards` has just commanded a hardware board. The same window exists for a mixed host with consent when the
  target moves to a hardware board that is *not* listed. A cooperative gateway that re-announces its inventory can hit the window;
  an adversarial one can spam rebirths to make it likely.
  Honest caveat: plant wire v1 has no authentication and the SPEC says a node only *claims* its environment, so a node that simply
  labels `hw-9` `simulated` bypasses consent too. The defect is that the host does not enforce its own rule on the live fact it
  already tracks, and its ADR (`documentation/design/plant-wire-adapter.md:95-104`) states the rule unconditionally. The boot is
  re-checked at send time precisely because the birth can change; the origin must be re-checked the same way.
- Fix: re-evaluate `settings.Authorizes` inside `session.order` (the session goroutine, from the device's current board) immediately
  before stamping the record, and answer the mailbox with `SettlementNotSent` + `originRefused` on failure; keep `checkOrigin` as the
  cheap pre-check. Add a test that flips a device's board between `deviceOf` and `submit` and asserts nothing is written.

### M1. Medium: a rebirth is never re-admitted, so the environment, contract, version and missing-device rules apply only to the first `NBIRTH` of a connection

- Evidence: `internal/plantwire/internal/domain/conversation.go:141-150` (`birth` calls `handshake`/`Admit` only when `!c.accepted`),
  `internal/plantwire/internal/domain/admission.go:27-34` (the whole rule list), `internal/plantwire/internal/domain/ingest.go:45-54`.
- Scenario: a host with `--plant-wire-environment simulated` accepts the first birth (all boards simulated). The node then sends a
  rebirth with `environment: hardware` and hardware boards; there is no `environment_mismatch`, `contract_unsupported` or
  `missingDevice` re-check, the boards map is replaced, and the session continues. Only the per-device `Authorizes` at the next command
  still protects (it refuses), so the outcome is fail-closed — but the operator's environment setting has silently stopped describing
  the session, and the same path is what makes H1 reachable. Rebirth is also the only place the host does not verify that `bd_seq`
  advanced (`conversation.go:142` assigns it blindly), so a retransmitted `NBIRTH` is applied as a rebirth.
- Fix: run `Admit` on every `NBIRTH` (the handshake-only parts become the same rules), or at least re-check `environmentAgrees` and
  require `bd_seq` to increase; add a test whose rebirth contradicts the admission and expects a refusal.

### M2. Medium: a `DCMD` that could not be written is reported as an unknown outcome and opens a reconciliation, although nothing left the host

- Evidence: `internal/plantwire/internal/app/orders.go:76-88` (`Command` registers `c.commands[id]` with phase `"sent"` *before*
  `s.send`; on a write error it returns `endWriteFailed` without answering `req.answers`), `internal/plantwire/internal/app/session.go:101-104`
  (the session ends), `session.go:162-170` and `internal/plantwire/internal/domain/command.go:119-133` (`Abandon` sees phase `"sent"` and
  settles `Unknown`), `internal/plantwire/internal/app/device_authority.go:51-73` (which opens the barrier).
- Scenario: the node has stopped reading (its receive buffer is full) or the socket was closed a moment ago; `Conn.Write` fails
  (`internal/plantwire/internal/transport/conn.go:50-57`), the session ends, `Abandon` reports "the session ended before the command's
  terminal answer", the dispatcher records `outcome_unknown`, and `RequireReconciliation` blocks the device until a good reading of
  every output clears it. A command that provably did not reach the node (nothing was written, or a line without `\n` that the node
  discards) is treated as possibly executed: fail-closed, but it costs a manual-review row and a device barrier on a main-path event
  (session churn).
- Fix: on a write error, deliver `Settlement{Phase: SettlementNotSent, Cause: ...}` to the mailbox (or record a `sent` flag that
  `Abandon` reads), so the outcome is a known "not sent".

### M3. Medium: one silent peer holds the only session slot forever; there is no handshake or idle deadline

- Evidence: `internal/plantwire/internal/app/session.go:68-88` (`step` selects on lines, `beats()` — nil before acceptance — calls and
  ctx), `internal/plantwire/internal/app/host.go:117-127` (`acceptSessions` serves one connection at a time, synchronously),
  `internal/plantwire/internal/transport/conn.go:39-48` (`ReadLine` has no read deadline).
- Scenario: a node (or a stuck gateway, or a test harness killed at the wrong moment with the socket still open) dials and writes
  nothing, or writes a partial line. `serve` never returns; `acceptSessions` never accepts the real node; the barriers, readings and
  commands for the whole experiment are stalled until `serve` is cancelled. The SPEC allows only one live session, so there is no
  second slot to fall back on.
- Fix: a handshake deadline before `accepted` (the SPEC's `STATE` answer is bounded) and an idle deadline tied to
  `3 × heartbeat_interval` on the data path, both cancelling the connection from the session loop.

### M4. Medium: the per-command answer trail is unbounded, and the node controls it with repeated `queued` receipts

- Evidence: `internal/plantwire/internal/domain/command.go:141-152` (only a receipt for an unknown command or one after `accepted` is
  dropped; unlimited `queued` receipts are accepted and settle the command as `queued` again), `internal/plantwire/internal/app/effector.go:115-127`
  (`await` appends every answer to `trail`), `internal/plantwire/internal/app/effects.go:40-58` (the whole trail is embedded in
  `ProviderResult`, which the dispatcher persists).
- Scenario: a node answers one `DCMD` with 200 000 `DRECEIPT {status: queued}` in the ~54 s dispatch call. `trail` grows with one map
  per answer (hundreds of MB in the worst case) and the provider-result document written to `outcomes` grows with it. The 8-slot
  mailbox bounds the channel, not the slice. The SPEC says at most one `queued` receipt per command, but the host does not enforce it.
- Fix: keep the first `queued` answer (or cap the trail at a small number and count the rest in the tally).

### M5. Medium: `deadline_ns` is always host wall time, but SPEC §8 puts a simulated node in world time; only a wall-paced plan makes it work

- Evidence: `internal/plantwire/internal/app/effector.go:91` (`h.clock.Now()` — the host's physical clock), `internal/plantwire/internal/domain/command.go:135-139`
  (`deadline_ns` from that value), `internal/plantwire/internal/wire/plantwire-contract/SPEC.md:151-153` ("A simulated node runs in
  virtual time: `deadline_ns`, `not_before_ns`, ... are world time"), and the ADR's contrary decision
  (`documentation/design/plant-wire-adapter.md:106-113`, "The deadline is UTC").
- Scenario: a simulated domain started without `pace: wall` (a fast, unpaced session). The node's `nowNS` is its world start (e.g.
  2026-01-01) and the host sends `now + 20 s` of wall time (Oct 2026): the node's `Freshness` refuses `deadline > now + 24 h` with
  `out_of_range` for every command. The fan experiment's plan is `"pace": "wall"` (`examples/.../simulate.plan.json`), and
  streams-simulator's `pacedNow()` takes `max(world, wall)`, which is why the experiment passes. Any other plan silently loses all
  commands, and the host has no branch for the node's time base.
- Fix: send world time for a simulated node (the host already receives `world_time_ns` in every `NSYNC`), or record the constraint in
  the contract/ADR that only a wall-paced simulated node may be commanded.

### M6. Medium: a timeout/ctx/session-loss unknown outcome can never be verified from readings; it stays in manual review and keeps the device barrier until an operator acts

- Evidence: `internal/plantwire/internal/app/effector.go:123-126` (the deadline and ctx paths return `unknownEffect` directly),
  `internal/plantwire/internal/domain/verification.go:33-44` (a watch exists only when `c.settle` sees a settlement), and
  `internal/plantwire/internal/app/verification.go:30-58` (`expectVerification` stores the target, `awaitReadings` waits on
  `t.verified`, which nothing can ever send). The module guide claims the opposite: `internal/plantwire/README.md` ("The conversation
  starts a watch when it settles a command as succeeded or unknown").
- Scenario: the node applies `set_cooling 45`, then dies before its `DRESULT` (a session loss, or a result lost on the socket). The
  host reports an unknown outcome, opens the reconciliation, and `VerifyDeviceCommand` waits the full 6 s on a channel no watch feeds,
  even though the device's own `level_percent = 45` reading arrives. The outcome stays `manual_review`/`reconciling` and the authority
  barrier is cleared only after an operator resolves the command, although the evidence to settle it was on the wire. Conservative
  (no false success), but it defeats the documented "unknown outcomes are settled by readings" recovery on the most likely unknown
  case.
- Fix: create the watch when the app settles a timeout/ctx/session loss as unknown (the session loop can do it with a call), or fold
  `unknownEffect` through `Conversation.Settle`.

### L1. Low: a session that never sends `NSYNC` accumulates `command_evidence` forever

- Evidence: `internal/plantwire/internal/domain/verification.go:48`, `:74` (evidence appended on verification and on withdrawal),
  `:80-87` (drained only by `barrier`). A hardware node has no barrier duty in the SPEC (section 8 gives the barrier to a simulated
  world), so a long hardware session with many commands grows one map per numeric expectation with nothing to send them.
- Fix: cap the pending evidence (drop oldest, count the loss) or drain it on the heartbeat.

### L2. Low: a command that never receives any settlement keeps its mailbox and watcher until the session ends

- Evidence: `internal/plantwire/internal/app/orders.go:82-83` (`mailboxes`/`watchers` entries created at send), `:95-116` (`forget`
  only on a settlement; `delete` only after a terminal one). The deadline path in `await` leaves the entry in place; the mailbox and its
  buffered channel stay referenced for the session's lifetime. Bounded by the number of unanswered commands, never reclaimed earlier.
- Fix: have `expectVerification`/the wait withdraw the mailbox when it gives up.

### L3. Low: every integer reading is narrowed to `float64` for the event document

- Evidence: `internal/plantwire/internal/domain/events.go:54-63` (`numeric` turns any `json.Number` into `float64`), against
  `SPEC.md:51-54` ("Integers are exact 64-bit ... a receiver that cannot parse them exactly must refuse"). The verification path keeps
  the exact `json.Number`; the event log does not, so an integer above 2^53 is recorded rounded.
- Fix: keep `json.Number`/`int64` when the number has no fraction, and only convert reals.

### L4. Low: architecture — the consent boolean is computed in `cmd`, and the socket is created by the worker module's helper

- Evidence: `cmd/agentic-stream/serve_plantwire.go:54-57` decides `PhysicalConsent` from three flags, duplicating the rule already
  enforced by `internal/device/internal/domain/profile.go:76-83` (a second place that must change together); `serve_plantwire.go:121`
  creates the listener with `worker.ListenEvidenceSocket`, and `internal/worker` was added to `cmd`'s import allow-list
  (`internal/architecture/import_rules_test.go:98`) for it. The plant wire transport wraps a listener it does not create, so its
  owner-only creation rules are borrowed from an unrelated module.
- Fix: expose the "profile grants physical consent" fact from `device` and let the plant wire transport create its own owner-only
  listener (or move the helper to `kernel`), then drop `internal/worker` from `cmd`.

### L5. Low: `Host.Close` is documented as returning a running `Run`, which a live session prevents

- Evidence: `internal/plantwire/host.go:109-110` versus `internal/plantwire/internal/app/host.go:101-127`: `Close` only closes the
  listener; a live session keeps `Run` blocked in `serve` until the caller's context is cancelled. `serve` happens to cancel it, so the
  behaviour is right and the doc is wrong.
- Fix: say `Close` refuses new sessions and the caller must cancel `Run`.

### L6. Low: the bench kind-copy test can skip every later file

- Evidence: `internal/plantwire/internal/app/kind_copies_test.go:35-45`: when one kind's command set differs it calls `t.Skipf` inside
  the loop, which aborts the whole test, so a byte drift in the remaining kind files is never compared; a missing checkout skips all of
  them. The intent is reasonable (the external master lags), but the scope is wider than the one file.
- Fix: `t.Logf` + `continue` for the tolerated mismatch, and compare the rest.

### L7. Low: a full ingest queue loses the rest of the record and the log does not say how many events were lost

- Evidence: `internal/plantwire/internal/app/feed.go:27-35` (`enqueue` pushes event by event and returns false at the first full
  slot), `internal/plantwire/internal/app/session.go:98-100` (the session ends), `session.go:63-65` (the tally reports `Events`,
  `Historical` and record-level `Dropped`, not the events of the half-enqueued record). The design accepts ending the session; it does
  not say a record may be split. The loss is bounded by one record and the session end is logged.
- Fix: enqueue a record atomically (check `cap - len` first) and count the events discarded with the session.

## Checked and found correct

- Consent composition: a hardware or mixed host cannot be opened without `physical` + `--live-actuation` + `--owner-authorized`
  (`internal/plantwire/internal/domain/settings.go:52-59`, `internal/plantwire/internal/app/host.go:70-81`,
  `cmd/agentic-stream/serve_plantwire.go:54-57`); a simulated host refuses `--hardware-boards`; `Dispatch` without a final
  authorization check always refuses (`effector.go:27-29`); an unauthorized origin is refused *before* the claim and names the board in
  the provider result (`effector.go:75-80`, `effects.go:60-65`).
- Verification cannot pass without a real reading: the watch starts only on a succeeded/unknown settlement, and only a good,
  non-historical `DDATA` of the same device on the command's boot, arriving after the result, matching every expected writable metric
  (word exactly, number within the kind's `resolution`) verifies (`internal/plantwire/internal/domain/verification.go:33-105`,
  `outputs.go:27-56`). Expectations come from the commissioned kind, never the birth. No expectation means "immediately unverified".
  A verified outcome seals the readings as device-state evidence, and the dispatcher checks that evidence against the binding's target,
  normalized target and boot (`internal/authority/internal/domain/binding.go:63-88`, `internal/actions/internal/store/reconciliation.go:43-51`).
  I found no path that reports a command verified without such a reading.
- Reconciliation clearing: the opening is read per device *and* boot (`internal/authority/internal/store/openings.go:13-31`), a reading
  must be good, live, of that boot and newer than the opening (any reading of a rebooted boot), every output of the commissioned kind
  is required, and the clear is refused while a command bound to that boot is unresolved in the action ledger
  (`internal/authority/internal/domain/readings.go:49-75`, `internal/authority/internal/app/readings.go:36-46`,
  `internal/authority/internal/app/resolution.go:47-72`, `internal/actions/internal/store/unresolved.go:13-33`); a pending-verification
  command is `manual_review`, which counts as unresolved. I could not construct a clear that contradicts an unresolved command.
- Unknown vs failure: the ADR table is implemented literally — refusal/`DRESULT failed`/queued-at-session-end are known failures,
  `DRESULT unknown`, a receipt or result on another boot, and a deadline/ctx loss are unknown and open the barrier; a not-sent refusal
  (no session, unborn device, boot changed, no site device, denied origin, claim refused) never opens one (`effects.go:13-38`,
  `command.go:119-133`, `device_authority.go:51-73`). `Command()` re-checks the boot at send time, so a reboot between prepare and send
  is a known "not sent" (`command.go:95-108`). `Abandon` settles every registered command when the session ends; a submit that loses the
  race with the session end returns "not sent" (`orders.go:37-46`). A claim lost after an executed command becomes unknown and opens the
  barrier, including after owner/epoch loss (`settleAuthority` + `authority.RequireReconciliation`, and `AuthorityLost` is wired in
  `internal/authority/service.go:47`).
- Protocol conformance: `seq` assigned at transmission with control-before-data priority, session/version stamped on every outgoing
  record, `HSYNC.through_seq` = the `NSYNC` seq, `DCMD` carries every required member (idempotency key unchanged, `expected_boot_id`,
  wall/UTC `deadline_ns`, `not_before_ns: 0`, `authority_epoch`, `policy_digest`, `fallback_on_lease_loss: safe_off`), `STATE` carries
  the epoch and an increasing heartbeat every 1000 ms, a heartbeat starts only after acceptance, the handshake refusal order matches
  SPEC §6, and every outgoing record is re-validated by `Encode` before it is written (`internal/plantwire/internal/domain/{records,conversation,command}.go`,
  `internal/plantwire/internal/wire/encode.go`). The contract digest, `VERSION`, the required safety members and enums, and the exchange
  fixtures are all pinned and tested (`internal/plantwire/internal/wire/contract_test.go`, `exchanges_test.go`).
- Socket security: the plant wire socket is 0600 in a 0700 directory, symlinks and occupied paths are refused, and the socket is
  removed only if it is still the one this process created (`internal/worker/internal/transport/uds.go:19-161`), meeting SPEC §2.
- Memory and goroutines: one reader goroutine per session which exits on `done` or on connection close; the session loop is single and
  never writes under a lock; the ingest queue is bounded and a full queue ends the session rather than blocking a barrier; a writer is
  bounded by a 2 s write deadline. No per-command or per-record goroutine leaks.
- Tests that would fail if the guard they claim were deleted: `internal/plantwire/internal/app/origin_test.go:42-54` (nothing written
  when the board is not listed, refusal names the board), `effector_authority_test.go:49-76` (no command after an unknown / after a
  reboot, nothing written), `reconciliation_test.go:75-115` (reboot clears only on the new boot's good reading; unknown clears only on
  a reading after it, not historical, not suspect, not another boot), `internal/authority/internal/domain/readings_test.go:24-69` (the
  same rules as a table) and `internal/authority/internal/app/readings_test.go:71-84` (no clear while a bound command is unresolved),
  `verification_test.go:39-95` (verified only by a matching good reading; a historical/suspect reading and a lapsed window give no
  status). `session_test.go` pins the barrier/heartbeat/one-session rules. The wire tests run the real conversation over the contract's
  exchanges.
- Determinism of the ingest mapping: event ids are a digest of origin/device/boot/metric/event time, the session's time is world time
  only, and the host clock is never read for the heartbeat (`internal/plantwire/internal/domain/{events,ingest}.go`).
- Architecture: the new module has the prescribed shape (facade layer 26 over app 25, domain 13, wire 14, transport 15), the facade
  operations are single delegations, no domain/board word is in the module, and the gate changes in `internal/architecture` tighten
  rather than loosen (`internal/watch` and `internal/device/internal/app` each *lose* an allowed import, a stale error-assertion
  exemption is removed, and a vocabulary floor is added in `domain_vocabulary_test.go:26`). `cmd` holds flag parsing and wiring only,
  with the one exception in L4.
