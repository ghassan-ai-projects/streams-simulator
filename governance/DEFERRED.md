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
| D-41 | M | `world` observation noise | Observations are not bounded by the channel's declared `range`: gaussian noise around a near-zero true value reports negative flows (factory run w-1: AC-1 `air_flow` −4.60 m³/h against `range` 0–1120). Real transmitters saturate or report out of range. Clamping changes every trace digest: needs a design decision (saturate, or emit an out-of-range quality flag). The canvas marks such readings meanwhile (UX goal, UXD-4). |

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

- **D-46 (mcp), resolved in P2d** `internal/mcp/internal/app` mixed use cases with the
  MCP protocol wiring (server construction, schemas, tool registration); the use cases are now
  the `director` module. Original note: the layer mixed use cases with the protocol wiring.
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
- **D-53 (domain data)** `domain lint` (P7j) finds `discrete-line-oee` out of range on a
  nominal day: `cycle_time_s` observes `line_speed` with `observation_gain: 60` and `observation_offset: 1`
  (3601 at 60 ppm) but declares `range` `[0.5, 10]`; the channel means seconds per part, which is
  `60 / line_speed`, and a linear observation cannot express that. The `units` counter also saturates at its
  `1000000` maximum within the day (a warning). **Fixed in P7j:** `cycle_time_s` observes a `curve` state
  (seconds per part against line speed), and the counter integrates parts per minute at gain 1/60; the domain lints clean.

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

> **Found while authoring simulations through the canvas (2026-10-15).**
> D-54 (`canvasbridge/internal/domain/runsummary.go`): the bridge decodes `run.json`
> itself (run seed, fault count, completeness) because `test/architecture/dependencies_test.go`
> lets `internal/canvasbridge` import only `kernel`; one implementation of the run
> summary would need an allow-list change, which needs the owner's approval.
> D-55 (`internal/component`): the motor drive consumes `load_torque` but no pump or fan
> offers it, so a pump cannot be linked to its drive; production and packaging lines
> declare `air_out` for air demand with no description saying so. The simulator backlog
> of the same round lists the failure behaviours (alarm never raised, frozen reading
> without a quality flag, no feeder trip, no late or duplicated readings).

> **Left with a directive after the nolint sweep (2026-10-18).**
> D-60 (`test/contract/factory_scale_test.go`, `//nolint:paralleltest`): the campus speed budget compares the wall clock
> of a 2,000-tag run with 20 s, so any parallel test in the package distorts it. Hidden dependency: the wall clock and the
> CPU shared with other tests. Design change that would make it fixable: report the run's own work (events emitted and
> a CPU-time or allocation counter taken by the run module on the run's goroutine) and budget that, or move the budget
> to a benchmark (`go test -bench`) that is run alone by design. Trigger: a second wall-clock budget appears, or the
> package gains enough parallel tests that the budget flakes.
> D-61 (real `SIGINT` delivery): the device serve test no longer sends a real signal; `uds.serveInterruptedBy` takes the
> interrupt source, so the test proves the source is registered before the socket is bound. That `kernel.UntilInterrupted`
> turns a real `SIGINT` into cancellation is proven by no test, because a signal is process-wide and a re-executed helper
> process needs `exec` of the test binary (gosec G204). Trigger to fix: a test-helper in `internal/testsupport` that
> re-executes the test binary with a fixed, validated argument list.
> D-62 (error text `End: ...`, `run/internal/durable`): the messages begin with the capitalised operation name `End`
> (the `Run.End` method); the text is kept because the comment that stood on the old directive called it part of the
> error contract and was not verified against `contracts/`, `test/pin/` and the docs in this sweep. The capital is
> built from the constant `endOperation` so staticcheck ST1005 no longer fires. Trigger to fix: a check of those
> places shows no contract, then lower-case the text in `endFailed` and update the two prefix assertions.
> D-63 (`canvasbridge/internal/transport/admission.go`, `//nolint:gosec` G124): the pairing cookie has no `Secure`
> attribute because the bridge serves plain HTTP on loopback; Safari and Go's cookie jar do not send a Secure cookie over
> `http://`, which would unpair the canvas. Hidden dependency: the transport scheme is decided nowhere. Design change:
> the server knows whether it is served over TLS and sets `Secure` from that fact. Trigger: the bridge serves over TLS or
> to a remote client.
> D-64 (`canvasbridge/internal/process/command.go` and `director.go`, `//nolint:gosec` G204): the bridge runs its own
> executable, resolved once at start, with an argv it built; gosec only accepts literal arguments. Design change: none
> that keeps an exec of a dynamic path; the bridge would have to embed streamsim as a library instead of running it as
> a child process. Trigger: the canvas bridge and the binary share one process.

> **Found while reviewing the physical-devices milestone M1 (2026-10-18).**
> D-56 (`device serve`, `internal/device`): it runs a domain without the `implements` check. It serves the device emulator
> for field-side protocol testing, where the domain is a fixture, not a node announced to a host. Trigger to fix: the
> emulator announces a kind to a host, or a domain with `implements` is served for a hardware-in-the-loop run; then it
> takes `--kinds-dir` and the check as `plantwire node` does.
> D-57 (loaders that skip the check on purpose): `domain lint` judges a nominal day of the domain's behaviour, not its
> interface; `suite` and `release` generate and digest scenarios from shipped domains, which declare no `implements`;
> `replay`, `verify`, `scenario from-run` and `run.ExportScenarioFile` reproduce a recorded run and send nothing to a
> host, so the recorded `implements_checked` and the domain digest are the evidence. Trigger to fix: a shipped domain
> implements a kind file (then `suite` and `release` must refuse a drifted one), or a replay is sent to a host.
> D-58 (error text outside the workspace): a message about a file given through `--components-dir` keeps its absolute path,
> because the bridge relativises only paths inside the workspace. Trigger to fix: the bridge serves a remote client, or a
> message crosses a trust boundary; then the binary reports a component by id, not by path.
> D-59 (KIND schema, `contracts/plant-wire/schemas/KIND.schema.json`): it constrains `unit` to a length of 32 and no more
> (no UCUM check), and does not limit how many commands of a kind are `safe_stop`. A kind with several safe stops is
> legal today; the host takes safe stops from its commissioned inventory, not from the birth. Trigger to fix: a host
> or gateway needs exactly one safe stop per kind, or a unit check is added to the contract.
> D-65 (default instrument by unit): `kind twin --suggest` picks the first instrument that fits the unit and range of a
> reading, exact unit first, then by id; units shared by several quantities (`%` is a fraction for a valve position and,
> by conversion, a relative humidity) are told apart only by a reading named after the instrument. Trigger to fix: a
> reading is suggested an instrument of another quantity that a user cannot tell from the unit; then an instrument
> declares the quantity (a kind file would name it) and the match is by quantity, not by unit.
> D-66 (twin from a kind directory): `kind twin --kinds-dir` (and `plantwire.WithKindFiles` with directories) still loads
> every `*.kind.json` of the directory and refuses the whole command for one malformed file, even a kind the choices do not
> name; only the `implements` check of a domain reads the kinds it implements alone. The canvas bridge passes the kind files
> of the twin as arguments and never the directory, so no surface depends on it. Trigger to fix: a surface hands the twin
> commands a directory that holds unfinished kinds; then the choices' kind names select the files, as a domain's
> `implements` names do (`SelectKindSources`).
> D-67 (where a capture differs): `kind capture` gives the first differing place (`differs_at`, a JSON pointer) but not the
> two values; the canvas resolves the pointer in the saved file it already fetched and shows "Saved ... The board says ...".
> That is pointer resolution the binary should own. Trigger to fix: a second surface needs the values, or the canvas
> shows a wrong value; then `comparison` carries `saved` and `announced` at `differs_at`.
> D-68 (runs that hold an announcement): the workspace listing does not say which recorded runs hold an NBIRTH, so the
> capture screen's run picker lists every run and the binary answers "No announcement found" for the others. Trigger to
> fix: more than a handful of runs per workspace; then the listing carries `announces` read from the observer journal's
> first records.
> D-69 (relay liveness against a host that sends no heartbeat): the relay counts three missed heartbeats only after the host has
> sent a `STATE` with `heartbeat` above zero, because the reference host sends one `STATE` at the handshake and never a
> heartbeat; a host that dies without closing its socket and never beat is not noticed until the stream ends. Trigger to fix:
> a host that is silent but connected in a lab run; then the reference host gains heartbeats and the relay's watchdog arms at
> the handshake.
> D-70 (relay: one gateway, no store-and-forward): the relay serves one gateway at a time (a second connection waits until
> the first is gone), and a gateway record that arrives while no host session exists is dropped and counted, not held, because
> store-and-forward with `historical: true` is the gateway's job (SPEC section 7). R8 kept this for the mixed node: it still
> serves one gateway, and the simulated world's records are held back until the host session has the merged birth. Trigger to
> fix: a gateway that does not buffer, or two gateways behind one mixed node; then the relay's slot and queue are redesigned.
> D-71 (mixed session: no clock pacing, the world ends with its host): a mixed session advances the simulated world exactly as
> `plantwire node` does, one slice and one barrier at a time in virtual time, as fast as the host answers `HSYNC`; it is not
> paced to the wall clock, so a simulated board beside a real one runs ahead of the hardware's wall time. If the host is away
> for three barrier caps the world's session ends (its devices are reported `DDEATH offline`) and the relay keeps passing the
> gateway's records. Trigger to fix: a lab run where the world and the board must be compared in real time; then a paced
> session design (a wall-clock slice schedule owned by `run`, with the barrier as a cap and not as the pace) replaces the
> free-running slices, and the world is reborn when the host returns.
> D-72 (replay of a mixed recording): `plantwire replay-trace` plays the field trace, the gateway's side of a recording, to a
> host as a simulated node; the simulated part of a mixed recording is reproduced and verified by `replay`/`verify` (domain,
> seed, command log and trace digest), not streamed to a host beside the trace, and a replay answers no command (every
> `DCMD` is refused `not_ready`). Trigger to fix: a consumer that must be shown a recorded mixed session end to end; then the
> recorded node plays the world's part from the command log in the same barrier model.
> D-73 (relay: no gateway liveness watchdog): a gateway that stays connected but sends nothing is never detected; the relay
> watches the host's heartbeat (D-69) and the gateway's link ending, not the gateway's silence, so its devices stay born and
> a command for them waits for an answer that never comes. Owner: R10 (connection health). Trigger to fix: a board that hangs
> with its serial or socket open in a lab run. **Fixed in R10:** the relay counts its ticks since the last admitted gateway
> record (a stray, forged, malformed or repeated record does not count) and after 30 s of silence reports its devices
> `DDEATH link_down`, closes the gateway link and records the gateway offline; the watchdog arms at admission, without a
> host heartbeat (`relay_liveness.go`).
> D-74 (relay: no rebirth rate limit): a gateway that sends a sequence gap on every record makes the relay ask it to rebirth
> once per gap, and a host that sends a rebirth `NCMD` in a loop makes the mixed node rebirth the gateway and the world each
> time. Each request is one record and the birth is bounded, but nothing limits the rate. Trigger to fix: a flapping link in a
> lab run; then a minimum interval between rebirth requests per side.
> D-75 (hardware-only relay field trace has no digest cover): a relay without `--domain` writes `field-trace.jsonl` and the wire
> journal but no `run.json`, so nothing records a digest of the field trace and a later edit of the file is not detected;
> `replay-trace` cannot refuse it. A mixed relay and a connected session are covered (`field_trace_digest`). Trigger to fix: a
> hardware-only recording that must be evidence; then the relay writes a minimal artifact or a sidecar digest.
> D-76 (replay does not regenerate origin and binding records): `replay` and `verify` reproduce the world, the command log
> and the trace digest; the observer `origin` and `binding` journal records of a connected or mixed session are written while
> it runs from what the gateway announced, sit outside the stream digest, and are not regenerated by a replay. Trigger to fix:
> a consumer that must show the origin of every device from a replay alone; then the field trace is replayed through the
> relay's routing and the notes are rederived.
> D-77 (live projection: full snapshots, kept whole by the canvas): a connected session journals the plant wire live projection
> (`live`, `run-journal-v0.1`) as a full snapshot of link states, devices and the latest 128 commands, at most once per barrier
> (simulated session) or relay tick (250 ms) when something changed; like `origin` and `binding` it sits outside the stream digest
> and a replay does not regenerate it (D-76). The canvas keeps every snapshot of a visit in memory so the cursor can rewind.
> Each record is bounded, but a session of many hours with many devices grows the journal and the canvas's history linearly.
> Trigger to fix: a multi-hour lab session whose review lags or whose journal is too large; then the journal writes changes
> (per device and per command) with a periodic full snapshot, and the canvas keeps a bounded index of snapshots.
