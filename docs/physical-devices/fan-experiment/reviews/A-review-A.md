# A rounds review A (correctness and architecture)

Range `git -C W log be5f237a..f577ece --no-merges` (A1 c39e4800, A2 b44ff80e, A3a e965dd9a, A3b f35a5faa, A3b2 ce02d38a, A3c c851548f,
A3d 7cb7c90c, A4 f577ece8) plus A0 e28b1ee4, read as `git diff 6a8ba9cb..f577ece -- . ':!internal/spec' ':!internal/replay'`.
W = `/private/tmp/claude-501/-Users-ghassan-my-projects-streams-simulator/517cecf1-e1a3-4525-abf0-112e2348451e/scratchpad/wt-fan-a` (branch `fan-a`
of `/Users/ghassan/my-projects/agentic-stream`). Read-only review, 2026-10-20. Paths below are relative to W.

Run here (no edit in W): `go test -count=1` of `internal/plantwire/...`, `internal/authority/...`, `internal/device/...`, `internal/actions/...`,
`internal/ingress/...`, `internal/runtime/...`, `internal/architecture/...`, `cmd/agentic-stream -short` (all pass); `go test -race
./internal/plantwire/...` (pass); `golangci-lint run` over the touched modules (0 issues); package coverage plantwire 100 %, app 92.5 %,
domain 98.2 %, transport 95.8 %, wire 91.7 %, plantwirefake 81 %. Two probes (below) ran in a throwaway copy of W under the session
scratchpad (`scratchpad/probe`), never in W.

Verdict: one High and four Medium findings. The structure is sound: one adapter behind the existing ports, the shared rules lifted
into their owning modules (`device.Bound`, `authority.ClaimForDelivery`, `authority.RequireReconciliation`, `actions.DispatchCallTimeout`)
and device-wire switched onto them, no business rule of substance in `cmd`, facade operations are single delegations, comments obey the
"none inside a module" rule. The High is a real defect on the experiment's main path (a slowly approved fan command).

## Findings

### H1. High: a command held at the node longer than the owner lease becomes an unknown outcome and opens a reconciliation although the node accepted and succeeded it

- Evidence: `internal/plantwire/internal/app/device_authority.go:51-63` (`settleAuthority`) asserts the target claim after an executed
  command and, when the assert fails, turns the outcome into `UnknownOutcomeError` and calls `reconcile` (opens the barrier). The claim
  is taken once, before the send (`device_authority.go:23`, `ClaimForDelivery`), and nothing renews it while the host waits for the
  node. Its life is the claim lease, which `cmd/agentic-stream/effect_profile.go:200` sets to `owner.Lease` (`--owner-lease`, default
  1 minute, `documentation/reference/cli.md`). The fan route is meant to wait up to 120 s for the node operator (PLAN D9, pass
  criterion 4; `expires_after_ms` 120000), and `serve_plantwire.go:114` only checks the route expiry against the dispatch call
  (`--dispatch-lease 3m`), never against the claim lease. Probe (copy of `app/effector_test.go` with `expires_after_ms` 200000, command
  held 90 s of virtual time, then `accepted` + `succeeded`): `err=action outcome is unknown: plantwire: authority lost during the
  command: assert target "P1": device target claim is not held by this owner`, `pending=true`. Consequences: a correct command is
  reported unknown, the device boot is under reconciliation (the next command is refused until the dispatcher's verification or an
  operator resolves it), and the run shows an authority loss that did not happen. Device-wire never meets this because its exchange is
  one short serial round trip. No test advances a clock past the claim lease.
- Fix: keep the claim live for the whole wait. Either renew it (`ClaimForDelivery`'s claim step, or a periodic re-claim by the owner)
  while `await` runs and before `settleAuthority`, or give the plant wire authority a claim lease of at least the dispatch call
  timeout, and make `serve` refuse to start when `--owner-lease` (the claim lease) is not above every route's expiry plus
  `answerGrace` (extend `plantWireCatalog`). Add a test that holds a command past the claim lease and expects `pending verification`,
  and a second that lets the claim really lapse and expects unknown (so the assert stays meaningful).

### M1. Medium: the handshake rules are not run again on a rebirth, so a second `NBIRTH` can change environment, contract digest and boards unchecked

- Evidence: `internal/plantwire/internal/domain/conversation.go:141-149` (`birth`) calls `handshake` (and so `Admission.Admit`) only
  `if !c.accepted`; a second `NBIRTH` goes straight to `c.ingest.rebirth(b, ...)` (`ingest.go:45-54`), which replaces `g.boards`, the
  source of every device's origin. Probe on a simulated host after an accepted birth: a rebirth `NBIRTH` with
  `contract_digest` of another contract, `environment: hardware` and a hardware board gave `verdict=<nil> end="" dropped=""` and no
  reply. So `environment_mismatch`, `contract_unsupported`, `version_unsupported`, `inventory_mismatch` and the stray-board rule are
  skipped for the rest of the session, and a rebirth can turn a hardware board into a `simulated` one, which `Settings.Authorizes`
  (`origin.go:21-28`) then commands without consent. SPEC section 6 says a rebirth keeps the session but is an `NBIRTH`; section 2 says
  the host checks the node against its boards at the handshake. `Conversation.Command` (`command.go:95`) re-checks only the boot, so
  the origin read earlier by `deviceOf` is not re-read at send time. The origin is only a claim (D3), but the admission rules are
  stated as invariants in the ADR and the README and are tested for the first birth only.
- Fix: run `Admit` on every `NBIRTH`; a refused rebirth sends `STATE{online:false, refusal}` and ends the session, as the first one does.
  Re-read the origin inside `Conversation.Command` (it has the device) and refuse if it differs from the one the order was checked
  with. Tests: a rebirth that changes the digest, the environment, or flips a board, in `conversation_test.go`.

### M2. Medium: a command that cannot be encoded ends the whole session, reports the command unknown and opens a reconciliation

- Evidence: `internal/plantwire/internal/app/orders.go:76-90` (`order`): `Conversation.Command` has already recorded the command as in
  flight (`command.go:95-107`, phase `sent`) when `s.send` runs `wire.Encode`; an encoding failure (the host record fails the DCMD
  schema: `command_id`, `idempotency_key`, `policy_digest`, `device`, `command` patterns) returns `endWriteFailed`, which ends the
  session. Probe with `PolicyDigest = "not-a-digest"`: session ended (`reason="a host record could not be written"`), the node saw a
  host loss (a hardware gateway then stops commands and lets leases fall back), `c-1` came back as an unknown outcome
  (`Abandon` settles `sent` as unknown, `command.go:131-146`, and `settleAuthority` opens the barrier), and `c-2` found no session.
  Nothing was sent, so the outcome is known. Real Agentic Stream values pass the schema today, so this is a misconfiguration or
  future-data failure, but it fails in the most expensive way. The existing "sends nothing it cannot send safely" test covers the
  catalog refusals only.
- Fix: encode before registering. In `session.order` build and `wire.Encode` the record first; on failure settle the mailbox with
  `SettlementNotSent` and return `""`. Better, make `Conversation.Command` return the encoded line (or an error) so a sent command is
  recorded only after its record exists. Add the probe as a test.

### M3. Medium: nothing checks the catalog, the site file and the commissioned kinds against each other at start, and two rules the ADR states are not implemented

- Evidence: (a) `internal/plantwire/internal/app/kinds.go:28` commissions with `domain.Commission(kinds, nil)`: the list of required
  devices is always nil, so `inventory_mismatch` for "a commissioned device missing from the birth" (`inventory.go:63-70`) can never
  fire in production, although the ADR (`documentation/design/plant-wire-adapter.md`, Session lifecycle 2) and A-FINDINGS section 6 list
  it as an admission rule. (b) The catalog's route targets are never checked against the site file's devices (`site.go:166`
  `DeviceFor` only answers per command), so a route whose target has no site device, or a host with a catalog and no site file,
  fails at the first dispatch with `ErrDeviceNotInSite`, not at start; H0.2 made "fail at start, not at dispatch" the rule for routes.
  (c) The ADR says "argument bounds come from the commissioned inventory and the catalog", and A-FINDINGS step 2 plans an `args_schema`
  check of a preset against the kind so drift fails at start. Neither exists: grep finds no use of the kind's `commands`, `args_schema`
  or `safe_stop` in the module. The node still refuses `out_of_range`, so this fails safe, but late and visibly only as a refused command.
- Fix: pass the site file's device names to `Commission`; at `New`, when a catalog is present require a site file and every route
  target to be a site device whose commissioned kind declares the route's operation, and check each preset's parameters against that
  command's argument schema. If the check is deferred, strike the "inventory" claims from the ADR and README and record it in `DEFERRED.md`.

### M4. Medium: the physical-consent rule is re-derived in `cmd` instead of owned by the device profile module

- Evidence: `cmd/agentic-stream/serve_plantwire.go:54-57`: `consent := options.profile() == device.EffectProfilePhysical &&
  options.LiveActuation && options.OwnerAuthorized`. That rule already lives once in `internal/device/internal/domain/profile.go:84-92`
  (`requireActuationConsent`) behind `device.CheckEffectProfile`. A-FINDINGS C5 planned the new rule ("boards only with physical +
  `--live-actuation` + `--owner-authorized`") as a `HardwareBoards` member of `EffectProfileConfig` in `device`, and AGENTS MUST 1
  and 2 say rules live in their module and a surface never re-computes them. Today the two copies agree; the next change to the
  consent rule (a new flag) will update one and silently leave the other granting hardware access.
- Fix: expose the derived fact from `device` (for example `device.EffectProfileConfig.PhysicalConsent()` or a `ProfileGrants` result
  from `CheckEffectProfile`) and have `plantWireFlags.settings` copy it; the host-side `Settings.Authorizes` and
  `CheckSettings` then consume that one value.

### L1. Low: the route-expiry start check ignores the verification window and the answer grace

`internal/device/internal/domain/bounded.go:50-58` refuses only when `expiry >= callTimeout`. A-FINDINGS C1 asked for expiry plus the
verification window below 0.9 of the lease, and the host waits `deadline + answerGrace` (2 s, `effector.go:13,107`) and then up to
`VerificationWindow` (6 s) on the same call context (`actions/internal/app/dispatch.go:55-58`). A route that expires at 53 s under the
default 54 s call passes the check, is answered at 53 s and cannot be verified (manual review). Subtract the plant wire grace and window
in `plantWireCatalog` (the check is device-neutral, so pass `callTimeout - grace - window`).

### L2. Low: the vendored `SPEC.md` is behind the master

`diff internal/plantwire/internal/wire/plantwire-contract/SPEC.md /Users/ghassan/my-projects/streams-simulator/contracts/plant-wire/SPEC.md`
differs in the section 15 `kinds.<kind>.commands` row, the "each board type has one mapping" sentence and the order of the 1.1.0 changelog
row (streamsim S11a). `DIGEST` and `VERSION` match, so the guard tests pass, but D5 says the folder is byte-identical. Re-run
`make plant-wire-sync PLANT_WIRE_CONTRACT_DIR=/Users/ghassan/my-projects/streams-simulator/contracts/plant-wire`.

### L3. Low: `AGENTS.md` does not list `internal/plantwire`

`documentation/architecture/repository-map.md` and the gates know the module (`TestModuleMapListsEveryPackage` passes) but the canonical
module list in `AGENTS.md` ("Repository structure") has no entry, and the architecture README gate index has no row for the new
`facadeSpecs` entry. The next agent reads AGENTS first. Add one line with kind, layer and callers.

### L4. Low: unbounded and lingering per-command state

(a) `Conversation.evidence` (`domain/verification.go:48,74`) only drains 128 per `HSYNC`; a hardware-only session sends no `NSYNC` (D-112),
so it grows by one entry per verified or withdrawn numeric command for the life of the session. Cap it (drop oldest).
(b) `Host.sent` and the session's `watchers` entries (`verification.go:30-46`, `orders.go`) are removed only by `VerifyDeviceCommand` or the
session end; an effect the dispatcher never verifies leaves them until the session closes. Bounded by command count, harmless today.
(c) `holdClaim` (`device_authority.go:75`) can register a claim in a session whose `close()` already released its claims (a dispatch in
`admit` racing the session end); it then waits out its lease. The claim lease is short and the owner is the same, so low impact.

### L5. Low: a full ingest queue ends the session at once

`app/feed.go:27-37` (`enqueue`) drops the whole session on the first full-queue push, as the README says. A-FINDINGS risk 5 recommended
"never drop live data, refuse the session only if the queue stays full for two barrier intervals". A 4096-slot queue makes it rare, but a
database stall of a few seconds under a busy hardware node would make the gateway see a host loss. Accept and keep (it is documented), or
wait two heartbeat intervals before ending.

### L6. Low: event ids carry no session identity

`domain/events.go:78-82` hashes origin, device, boot, name and event time. A simulated node's `boot_id` is deterministic and, unpaced,
its world time starts at the same instant each run, so two runs against one database would produce identical ids and the second run's
events would deduplicate silently. The plan requires paced sessions for Agentic Stream, so this does not bite the experiment; state it in
the README ingest section or add the node's session/run id.

### L7. Low: kind and site files are read in the `app` layer

`app/kinds.go:48,65` and `app/site.go:33` call `os.ReadDir`/`os.ReadFile`; A-FINDINGS section 3 and the ADR say `app` imports no `os`.
`TestApplicationLayersDoNotTouchInfrastructure` bans only `net` and the database, and `internal/spec/internal/app` has the same
precedent, so no gate is broken, but the design text should either say so or the reads should move to `cmd` or an edge package.

## Checked and found correct

- **One adapter, rules lifted.** `device.CapabilityCatalog.Bound` is the one bounding rule (device-wire `Materialize` now calls it);
  claim/bind/assert is `authority.ClaimForDelivery`, called by device-wire (`session_exchange.go:95-108`) and plant wire; the
  authority-loss recovery moved into `authority.RequireReconciliation` and `device/session_reconcile.go` lost its copy;
  `actions.DispatchCallTimeout` is one function with the dispatcher's own use. Evidence-based clearing is `authority.ResolveWithReadings`,
  called only by plant wire (`app/reconciliation.go:25`); device-wire is untouched (D4). Migration 038 and the `NormalizedTarget`
  comparison (`binding.go:53-82`) keep device-wire bindings unchanged (NULL means the old rule).
- **Layering and gates.** Layers 13/14/15/25/26 and the allow-lists match the A-FINDINGS table, except that `domain` imports less
  (`canonicaljson`, `contractsv1`), the facade imports `transport` only to wrap the injected listener, and the facade's six operations are
  single delegations. Every gate edit in `internal/architecture` in the A commits is an addition for the new module
  (`allowedImports`, `packageLayers`, `facadeSpecs`, `effectImplementationPackages`, `deviceAdapterPackages`, `plantwirefake` at layer 0 with
  no imports, `cmd` gaining `plantwire` and `worker`) or a shrink (`device/internal/app` lost `internal/control`; `watch` lost its domain import;
  one `unnamedErrorAssertions` row removed). No cap raised, no hook added, no `scripts/`, `.golangci.yml` or coverage-floor change. The
  `domain_knowledge_test`/`domain_vocabulary_test` edits belong to the merged H0 work and were not reviewed here.
- **Session lifecycle.** Handshake order and refusals match the SPEC and the exchange fixtures (`wire/exchanges_test.go` drives every file of the folder
  through the conversation: handshake, `HSYNC`/`NCMD` answers, stamped `DCMD`); `STATE` heartbeats on the injected clock; `HSYNC` answered
  from the session loop; a sequence gap asks one rebirth; `NDEATH` of an older birth is ignored; goodbye on stop; a second connection
  waits (the ADR says so; A-FINDINGS said "closed", the ADR wins); two sessions in a row without restart are tested, including against the
  real streamsim node (env-gated).
- **Ingest.** Site file schema, complete quality table, "no non-good code may map to the good event quality" check, readings only on
  non-historical `DDATA` of a born device on its current boot, heartbeat on event time with the freshness and `every_ms` rules, historical
  counted and dropped, events through `ingress.ServeFeed` (same admission and quarantine as a live line), bounded queue with a single feeder
  so a slow batch cannot miss a barrier. Unmapped kinds (LED, fan) give no events.
- **Commands.** Bounded by the catalog, final authorization check, origin read from the live birth (an unannounced board is refused),
  idempotency key forwarded unchanged, `deadline_ns` from the route expiry on the host clock, `authority_epoch` from the host epoch (default
  Unix seconds), policy digest, `safe_off`, boot pinned (`ErrBootChanged`). Receipt/result table matches the ADR including
  `session_ending`, `accepted` on another boot, `unknown` result, deadline plus 2 s, and a dispatch context end. Refusals are known failures
  with the reason in the provider result and no retry.
- **Authority, verification, consent.** Claim/bind/assert before the send, reconciliation barrier on reboot and unknown outcome, claims
  released at session end; verification needs a good, non-historical sample on the same boot after the result for every writable-metric
  argument, within resolution; evidence is sealed with `authority.SealReconciliationEvidence` and numeric observations go to the next
  `HSYNC`; hardware commands need `physical` + both consent flags + the board in `--hardware-boards` (`CheckSettings`, `Authorizes`), tested
  both ways; the switch from a simulated to a real boot clears on the first good reading.
- **Docs.** `cli.md`, `configuration.md`, `testing.md`, the ADR and the module README agree with the code on flags, the profile rule, the
  dispatch-lease check, the 6 s window and the outcome table; the exceptions are listed in M3 and L3.
