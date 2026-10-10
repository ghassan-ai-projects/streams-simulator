# Reported defects and improvements (not fixed in this program)

Instruction for this program: **keep behaviour**; report bugs and
improvements for later. Nothing below is changed by a structure round. Items
come from the four read-only surveys (`survey/`) and were found by reading
code, not by running it; confidence is the surveyor's. Confirm each with a
failing test before fixing. A fix is its own commit, with its own regression
test and an entry in the release evidence, never part of a refactor round.

Severity: **H** wrong evidence/oracle or fail-open safety; **M** wrong result
in a reachable path or a race; **L** latent, cosmetic or hygiene.

## Correctness — oracle, evidence and safety

| ID | Sev | Where | Finding |
| --- | --- | --- | --- |
| D-01 | H | `truth/internal/domain/store.go`, `mcp/internal/app/director.go` | Sealed-oracle gate fails open: `Store.Reveal` refuses on an open run only `if OpenChecker != nil`, and the checker is wired after construction. Round R7 makes the constructor require it (production wiring is unchanged); the fail-open default of other call paths is the defect. |
| D-02 | H | `run/replay.go:145-155` | `ReplayResult.FirstDivergence` compares ledger length to `Counts.Emitted`; it never finds the first differing record, though `mcp verify` reports it. |
| D-03 | H | `cli/internal/app/replay.go:12-52` | `streamsim verify` exits 0 even when `matches=false` (`verifyOnly` ignored). |
| D-04 | H | `cli/internal/app/score.go`, `score/offline.go` | Offline `streamsim score` passes `calls=nil`: loop metrics zero, `ActionFidelity` false for any verdict with actions; `Reproducible`/`Unblinded` never set. Online scorecard differs from offline. |
| D-05 | H | `deviceworld/plant_execution.go:75` | World command id is the constant `"safe-stop/"+target`; a second safe stop inside `idempotency_window_s` replays the cached result and applies no kick while the device reports safe state. The instrument lies. |
| D-06 | H | `device/command_execution.go:62-66` | `rememberExecution` clears only `AckLost`; a retry with the same idempotency key re-injects `Disconnect`/`Duplicate` every time (livelock). No test covers the disconnect fault. |
| D-07 | M | `run/finalize.go:34-42`, `mcp/internal/app/run.go:143-153` | After a failed run `End` returns artifact + error, `finalizeRun` bails before `RunEnded=true`; the run looks open forever, so score/reveal are refused. |
| D-08 | M | `mcp/internal/app/run.go:29,153`, `mcp/internal/app/director.go:79-88` | Data race: `WorldRecord.Started`/`RunEnded` written outside `d.mu`, read under it; `BeginRun` check-then-set race. |
| D-09 | M | `run/artifact_metadata.go:24` | `Counts.FaultsInjected` is `ActiveFaultsCount`, i.e. faults still active at End; under-reports after any clear. |
| D-10 | M | `run/identity.go:41-43`, `replay.go:86-96` | Replay fidelity gap: `worldDigest` hashes `Noiseless`, `ForceFailureMode`, `ClockMultiplier`, but the artifact omits the first two and `replayConfig` never sets `ClockMultiplier`. Latent: production never sets them. |
| D-11 | M | `truth/internal/domain/solver_scan.go:48-58`, `record.go` | `FirstObservableNS == 0` means both "unset" and a legal epoch-zero time; wrong oracle with `start_time` 0. |
| D-12 | M | `truth/internal/domain/solver_scan.go:54`, `detector_math.go:52-58` | Sigma 0 (schema allows it) or unknown detector form makes every fault trivially observable instead of erroring. |
| D-13 | M | `truth/internal/domain/solver.go:135` | `peerSigma` uses `len(entityIDs)`; readings use the world's actual entities (differs when `entityIDs` is empty). |
| D-14 | M | `world/integration.go:54-67`, `world/dynamics.go:122-133` | Reads are not pure: `stateAt` takes a partial RK4 step and draws per read-bounded step, so a run world (reads every emission) and the oracle's world (no emission) partition steps differently. Fixing changes numbers: needs a design decision. |
| D-15 | M | `run/*` mutators | No `finished` guard on `Advance`, `InjectFault`, `ClearFault`, `ApplyPerturb`, `ClearPerturb`, `AddEntity`, `RetireEntity`, `EnvInject` (only `InvokeEffector`/`SubmitVerdict` check). |
| D-16 | M | `run/initialize.go:60-71`, `finalize.go:68-76` | File descriptors leak on partial failure in `New` and `End`; a publish failure after `finished=true` is unrecoverable. |
| D-17 | M | `run/verdict.go`, `history.go` | `Ledger`, `Verdict`, `History`, `Reproducible`, `Unblind` read/write without a lock; race with operator goroutines. |
| D-18 | M | `run/emission.go` vs `delivery.go` | Two near-identical delivery state machines diverge on `evidenceRec` timing, ledger `EventTimeNS` source, and continuing after sink failure. Intent must be decided before unifying. |
| D-19 | M | `cli/internal/process/process.go` (`time.Now`), `cli/internal/app/run_script.go:70` | `cli-<unixnano>` command ids enter the recorded command log: two identical `run --effector` runs are not bit-identical. |
| D-20 | M | `audit/evidence.go:20`, `mcp` audit | `sim.scenario.audit` with default `duration_ns=0` audits one sample against a 24 h world; MCP audits with seed 1 / 60 s, suite uses `Seed^0x5eed` / 120 s: they disagree for the same injection. |
| D-21 | M | `refconsumer/series.go:68,102` | Silence detections cite `seq:<lastSeq>`, 0 unless a prior detection set it; scoring flags unfounded evidence. |
| D-22 | M | `mcp/internal/app/world_config.go:23` | Seeds pass through `float64`: above 2^53 rounded, negatives wrap. |
| D-23 | M | `mcp` advance | `advanceToolError` maps every non-quiescence error to `clock_backwards`. |
| D-24 | M | `mcp` operator report | `OperatorView.Report` applies quiescence before validating the verdict and ignores `run_id`. |
| D-25 | M | `world/observe.go:10` | `World.Reading` is documented noise-free but draws the noise substream and mutates drift state on a non-Noiseless world. |
| D-26 | L | `perturb/internal/domain/helpers.go:83` | `reorder` swaps by `atNS%2==0`, not the layer RNG: whole-second events always swap. |
| D-27 | L | `world/effectors.go:86,121` | World idempotency key is `command_id` alone: same id with another effector/entity/args replays the first result. |
| D-41 | M | `cli/internal/app/run_script.go:invokeScriptedEffector` | `run --effector` always sends empty args, so every shipped effector with required args (all of them) is rejected: the flag is unusable except for argument-free custom domains. |
| D-42 | L | `world/effectors.go:callResult` | An idempotent replay returns the original result without `EffectETANS` (the call record does not keep it), so a replayed acknowledgement differs from the first one. |
| D-28 | L | `sink/sink.go:128-141` | HTTP push POSTs synchronously under the run lock (30 s timeout), response body not drained. |

## Correctness — loaders, adapters, schemas

| ID | Sev | Where | Finding |
| --- | --- | --- | --- |
| D-30 | M | `adapter/internal/files/verify.go` | `adapter verify` panics on an empty fixture file. |
| D-31 | M | `adapter/engine.go:127-128` | `hash_suffix`: `DigestBytes(...)[:16]` is `"sha256:"` + 9 hex (colon in id); panics when `1 ≤ max_length < 16`; untested, no shipped adapter uses it. |
| D-32 | M | `adapter/internal/domain/verify*.go`, `cli/internal/app/catalog.go` | `SchemaOK`/`GoldenMatch` unset when the field is absent so the CLI prints FAILED with an empty detail; `splitRecords` ignores `encoding` (`json-array` + output schema fails). |
| D-33 | M | `jsonschema/compile_keywords.go:143` | `Compile` panics on user schemas (`"uniqueItems": "x"`); `toInt` silently returns 0. Reachable from adapter `output_schema` and effector `args_schema`. |
| D-34 | L | `model/domain.go:111-115` | `F1Input` with explicit `"coef":0` marshals without coef and reloads as 1; `mcp/internal/app/director_resources.go:25` misreports it. |
| D-35 | L | `domain/compilation.go:85-90` | `observation_gain: 0` becomes 1; zero equals unset for `omitempty` floats. Behavioural contract: decide, do not "fix". |
| D-36 | L | `run/identity.go:10` | `adapterDigest` is `DigestBytes(json.Marshal(struct))`, not RFC 8785 of the document (AGENTS rule). Struct field order and `omitempty` tags are digest-bearing. Canonicalising changes every adapter digest and breaks stored artifacts. |
| D-37 | L | `randutil.Picker` | Sums `total` in map-iteration order (non-deterministic if used). Unused: deleted in R2 (a deletion, not a fix). |
| D-38 | L | `domain/validation_records.go:95,144`, `jsonschema/compile.go:44` | Which validation error surfaces varies with map order (pass/fail identical). |
| D-39 | L | `canonical/number.go:29-38` | Integer literals pass through verbatim, so integers beyond 2^53 are not RFC 8785 form. Digest-bearing; pin, do not touch. |
| D-40 | L | `adapter/expression.go:110-113` | Counter default width is dead: `checkCounter` rejects 0 while the engine pretends a default exists. |

## Design decisions needing an owner

| ID | Topic | Notes |
| --- | --- | --- |
| P-01 | `internal/wall` seam | Zero importers; `time.Now` is called directly at `run/artifact_metadata.go:34`, `run/verdict.go:39`, `cli/internal/process/process.go` (`time.Now`), `cli/internal/app/device_world.go:81`, `cli/internal/app/manifest_options.go:43`. Routing through `wall` makes `created_at`/`unblinded_at` zero under `-tags simdet`. Either inject a clock func at the edge or delete the package and amend DECISIONS D-13. |
| P-02 | Comment policy | `agentic-stream` forbids comments inside modules. Not adopted here (see STANDARD). |
| P-03 | Typed closed sets | `FailureMode`, `RejectCode`, cadence mode, detector form, transform op, encoding as named string types (JSON-neutral). Scheduled as R15 once layers settle. |
| P-04 | Typed device records | Receipt/Result/State as structs instead of `map[string]any`; canonical bytes must stay identical. |
| P-05 | `Config` fields set by no production caller | `ClockMultiplier`, `Noiseless`, `ForceFailureMode`, `WorldID`; `TimeMode` scaled/wall are labels only; `SinkBroker` is unsupported. Remove or implement. |
| P-06 | Suite profile-name branches | `suite/internal/domain/admission.go:30-55` branches on `correlated_cascade` and `sensor_pathology`, violating "domains are data" in spirit. |
| P-07 | Delivery-path unification | See D-18. |

## Hygiene and small improvements

- D-43 (M, FIXED in hardening): the gate was a name heuristic; a
  `go/types` scan finds 46 unmarked map ranges it passes. Order-sensitive
  ones: `truth/internal/domain/solver.go:168` and `audit/scenario_setup.go:65`
  (InjectFault in map order; fault ids `f-N` depend on it, latent because every
  caller passes one entry) and `world/internal/domain/faults.go:97`
  (`validateFaultParameters` names a map-order-dependent unknown parameter in
  its error). Upgrade the gate to type information and sort the three sites in
  a listed change. `device`/`deviceworld` have 15 ranges (message order,
  capability catalog order; `determinismDebt`).
- D-44 (L, see hardening): `world.Entity` returns the live `*Entity`; `ID`/`Type`/`BornNS` are
  writable by callers and `BornNS` feeds `InitialEntityIDs`. Return a value
  type or a `HasEntity` query in a listed change.
- D-45 (L): `EffectorCall.Args` is the caller's map (`recordCall` stores it and
  `EffectorCalls()` copies the slice shallowly); no layer code mutates it.

- `truth.Store.SealStatus` returns a `sealed` flag that is always true for a
  known run (the sealed map is set with the label); the `!sealed` branch in
  `mcp/internal/app/run.go` cannot trigger.

- `truth.Store.Reveal` holds `Store.mu` while calling `OpenChecker` (takes `Director.mu`): lock order `Store.mu → Director.mu`; fragile.
- `sink.File.Close` rereads the file from disk; not idempotent; `Inproc.Close` returns its internal slice; file perms `0o666&umask` vs run's `0o600`; doc promises a simdet "wall" sub-mode that does not exist; `NewFile` has an unreachable branch.
- `world`: `AddEntity(params)` is ignored; failure mode is a bare string; `recordCall` takes 11 positional params; dead `atNS`/`t` parameters threaded through many functions.
- `perturb`: unused `atNS`/receiver params; `delay_tail.sigma_s` is a second exponential mean; scattered default literals.
- `device`: `Listen` accept loop has no join and does not close the active connection; `ServeConn` loses a held swapped frame on injected disconnect; strict-decode copy-pasted 3× (device codec, capability loader, deviceworld bindings).
- `cli`/`mcp`/`domain`: error prefixes stutter (`streamsim: streamsim: domain:`, `adapter: adapter:`, `canonical: canonical: canonical:`); strings are contract-ish, change deliberately.
- Resolved in R1/M3: the legacy 60-line review table and the stale-root determinism test (now a repository gate, see D-43).
- Duplicates found for later: `asFloat` ×4, `mustAny` ×3, `formatErrs` ×2, `fnv` in `run` vs `randutil.Fnv1a64`, `audit.Perturbation` ≡ `suite.Perturbation`, `suite.renderID` ≡ `world.RenderID`, refconsumer name/version literal ×3, manifest hard-codes 0.4/0.9, strict-decode ×3.

- **D-46 (mcp)** `internal/mcp/internal/app` still mixes use cases with the
  MCP protocol wiring (server construction, schemas, tool registration).
  A protocol edge split needs the handler signatures to stop returning
  SDK types; coverage of the layer is 65.5 %, below the 70 % bar.
- **D-47 (device)** `Device.Advance` and `Device.SetFaultSchedule` in
  `internal/device/internal/domain/state.go` have no production caller.
  Removing `Advance` leaves the default manual clock unable to move; decide
  between removing both with the manual clock or keeping them as the
  documented test-clock API.
- **D-48 (run)** `Config.QuiescenceClock` is exported through the `Config`
  alias but typed by the private `quiesce.Clock`, so only in-module tests
  can implement it; inject through an unexported option instead.
  `durable.Ledger.Closed` exists only for the finalisation regression test.
- **D-49 (cli)** After M15 `replay` and `verify` share one handler that does not
  know which command invoked it, so the D-03 fix (exit non-zero when
  `matches=false` under `verify`) must give the handler the command name.
  `adapter verify` on an empty fixture file still panics (D-30), now in
  `adapter/internal/domain/verify.go`.
- **T-01 (tests)** `unnamedErrorAssertions` in `test/architecture/testbar_test.go`
  lists the 27 test files whose negative tests assert only that an error
  occurred (`if err == nil { t.Fatal }`). Each entry becomes an
  `errors.Is/As` or message assertion; the table only shrinks.
- **D-50 (world)** `exceptionCadence` (`world/internal/domain/cadence.go`)
  records `lastSent` only for the first reading, so a report-by-exception
  channel compares every later reading against the first value instead of the
  last one sent. Looks like a defect; not changed here (the behaviour pin and
  every golden trace depend on it).
- **D-51 (mcp)** `sim.world.create` does not return the run id, yet
  `sim.truth.seal` needs it before `sim.run.begin` (documented as returned by
  `run.begin`). Callers must derive `r-<n>` from `w-<n>`; the acceptance test
  does. Return `run_id` from `world.create`.
- **D-52 (domain data)** `cold-chain-transit` declares the `power_transfer_gap`
  detector on `reefer.link_state` (noise none, sigma 0), a channel the fault
  does not change. With the solver no longer treating a zero deviation as a
  detection (D-12) that scenario is correctly unobservable and the nominal
  suite excludes it; the domain should name the channel the fault affects.
  Changing the domain changes its digest, so it is left to the domain owner.

> **D-50 closed as not a defect.** `exceptionCadence` does not update
> `lastSent` itself, but its only caller (`numericReading`) sets
> `lastSent`/`hasSent` after every emission, so the deadband is measured from
> the last report. A through-the-world test would have passed before and after.

> **D-28 closed as designed.** The http-push sink posts one line at a time
> under the run lock because delivery order is part of the trace digest; the
> 30 s client timeout bounds a stalled receiver. A measured experiment showed
> the response body needs no explicit drain (Go's transport reuses the
> connection for small bodies), so there is nothing to fix without an
> asynchronous, order-preserving sink, which is a design change.

> **Decided in the hardening program.** D-35 (`observation_gain: 0` is read as
> unset): kept; zero and unset are the same value for the declared floats and
> the domain contract says so. D-39 (integers beyond 2^53 are written
> verbatim): kept; seeds are `uint64`, and RFC 8785's double form would map
> distinct seeds to one digest. Both are documented in the `canonical`
> package comment. D-36 is fixed (canonical adapter digest, simulator 0.2.0).
