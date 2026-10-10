# Survey A: internal/run, internal/sink, internal/truth (read-only; nothing in the repo was changed)

Conf = my confidence the item is real. Paths are relative to internal/.

## 0. One-paragraph verdict
run is a 1.7k-line orchestrator that also owns durable-ledger file I/O, artifact publishing/loading, a quiescence barrier, a wall timer, a replay command decoder and three grab-bag accessor files; its exported surface is ~2x what callers use. sink is small and mostly fine (dead branch, doc drift, blocking HTTP under the run lock). truth mixes a pure oracle (Solver/BuildRecord) with a stateful sealed Store whose reveal gate fails open when unwired. Real defects ranked in section 6.

## 1. Responsibility today vs should-own
- run: today "wire world->perturb->adapter->sink->ledger, log commands, publish artifact, replay" (run.go:1-8). SHOULD own only: command admission + ordering (commandMu), delivery recording, lifecycle (New/Advance/End). SHOULD NOT own: ledger file mechanics (initialize.go:102, delivery.go:15-39, finalize.go:90-112), artifact JSON/file layout (finalize.go:52-158, artifact*.go, replay.go:31), quiescence timer/channel mechanics (run.go:48-76, quiescence.go), digest recipes (identity.go, initialize.go:172-185).
- sink: today three line sinks behind `Sink{Write,Close}`. Right scope. Close() returning the full byte stream is the contract the digest needs; keep.
- truth: today (a) pure solver + label builder, (b) sealed label Store. SHOULD be two files/responsibilities (store.go vs record builder); Store is only used by mcp.

## 2. Exported surface vs real use (grep over repo, non-test unless noted)
run, used externally: Config, New, Run, ErrConsumerNotQuiesced, LoadArtifact, ReplayArtifact, ReplayResult.
Run methods used outside run: Advance, InjectFault, ClearFault, ApplyPerturb, ClearPerturb, InvokeEffector, RetireEntity, EnvInject, History, ReportQuiesced, Domain, Digest, UnblindedStamp, Unblind, AppliedPerturbations, SubmitVerdict, Verdict, Ledger, Reproducible, End.
Dead or test-only (candidates to unexport/delete; `deadcode -test` hides them because tests call them):
- RecordHistory (history.go:21) no caller anywhere. RenderRecord (emission.go:81; doc claims "MCP trace export path", false). Unblinded() (verdict.go:43; UnblindedStamp is the used twin; same state, two accessors).
- ConfigureEnvTarget (commands.go:111) only artifact_test.go:109. Production never calls it, so `sim.env.inject` can only ever return "not enabled" (commands.go:124). AddEntity: only replay (replay_commands.go:39) and tests; no MCP tool `sim.entity.add`.
- Test-only hooks exported: SetEvidenceRecorder (2 external test files), SetQuiesceParkedHook, SetFailureMode, Trace() (refconsumer tests, run tests).
- QuiescenceClock/QuiescenceTimer/DefaultQuiescenceTimeout and Config.QuiescenceClock: used only by in-package tests.
- Config fields set by no production caller: ClockMultiplier, Noiseless, ForceFailureMode, WorldID (mcp/cli never set them); ClockMultiplier only flows into digest/artifact. Model const SinkBroker unsupported (initialize.go:139).
- Fields Perturb, Engine, Sink: zero external uses -> unexport. Config: only `.Config.Seed` read externally (add Seed()). World: externally only read calls (Clock x17, EmittedCount, EffectorCalls, EntityIDs, Entity, StateValue, ListFaults, NextEventNS, PendingKicks, ActiveFaultsCount). No external code mutates through Run.World, so the world pointer could become a small read facade.
- History() returns unexported `[]stateSnapshot` (history.go:11): exported method, unnameable type; score reads its fields.
sink, used externally (only run): Sink, Inproc, NewFile, NewHTTPPush. File/HTTPPush methods beyond the interface: File.Flush used via anonymous interface assertion (delivery.go:33).
truth, used externally: BuildRecord (suite only), NewSolver (suite), NewStore (mcp), Solver, Store, SetupCall (audit, suite, tests). Store.OpenChecker is a mutable exported hook assigned in mcp/director.go:47. Result/Solve are used only by BuildRecord (+ truth tests) -> Solve and Result can be unexported. audit imports truth just for SetupCall (a record type) = upward-ish dependency smell: SetupCall belongs in model.

## 3. I/O, clock, env, globals, goroutines, locks (owner function)
- File I/O in run: ledger open/append/flush/sync/close initialize.go:102 openLedger; delivery.go:15 appendLedger; delivery.go:27 flushDurable; finalize.go:90/104 closeDurableLedger/flushDurableLedger; trace/ledger/history/verdict/run.json writes finalize.go:52,114-158; artifact.go:41 writeJSONL; replay.go:31 LoadArtifact (os.ReadFile). MkdirAll 0o700, files 0o600 (sink.File uses os.Create = 0o666&umask: inconsistent).
- Clock: run.go:75 time.NewTimer (realQuiescenceTimer; injected seam OK); artifact_metadata.go:34 time.Now (CreatedAt); verdict.go:39 time.Now (unblindedAt). Both bypass internal/wall, which has zero callers in the repo (wall.Now unused). model.CurrentPlatform() (runtime) at artifact_metadata.go:67.
- Env: none. Globals: none (only package-level error var ErrConsumerNotQuiesced).
- Goroutines: none spawned in run/sink/truth. HTTP client timeout 30s (sink.go:122).
- Locks (run): commandMu serializes world-mutating commands + End + SubmitVerdict; Advance drops it before waiting (advance.go:19-24) and retakes it to `fail` (advance.go:53). quiesceMu guards watermark + notify channel (quiescence.go:14,70). Reads WITHOUT lock: Ledger (verdict.go:29), Verdict (:26), History (history.go:11), Reproducible, UnblindedStamp, AppliedPerturbations, Trace, Unblind() writes unlocked (verdict.go:37).
- Locks (sink): File.mu, HTTPPush.mu (held across the network POST, sink.go:128-141). Inproc has none (relies on run's commandMu).
- Locks (truth): Store.mu; Reveal calls OpenChecker while holding it (truth.go:86) and OpenChecker takes Director.mu (mcp/director.go:80): order Store.mu -> Director.mu. No inverse path today; fragile.

## 4. Mixed responsibilities
- finalize.go: End = lock + sink close + digest + ledger close + directory creation + 4 file writes + artifact assembly, all in one chain (finishRun :30). publishTrace (:130) re-writes the file-sink's own file with the same bytes (redundant) and chooses the path by SinkName.
- delivery.go vs emission.go: two near-identical delivery state machines (deliver/deliverRendered/writePendingDelivery/failPendingDelivery vs onEmit/deliverEmission/renderEmission/writeEmission/failEmission). Divergences: evidenceRec called after render and only for non-empty line (delivery.go:97) vs before render, even for omitted/render-error (emission.go:55); ledger EventTimeNS from d.Event.EventTime (delivery.go:109) vs original ev (emission.go:61-ish `atNS`); ParseTime errors ignored in flush path; after a sink failure the emission path stops (returns false) but flush path keeps writing to the sink (advance.go:76).
- initialize.go: config defaulting + RunID derivation + hand-rolled FNV (:172-185; identical to randutil.Fnv1a64) + ledger file + adapter meta map + sink factory switch + preamble write.
- identity.go: worldDigest takes *Run to build a map; adapterDigest = DigestBytes(json.Marshal(struct)) (identity.go:10) NOT RFC 8785 canonical, against the AGENTS rule; adapter digest logic belongs with adapter (like domain.Compiled.Digest).
- artifact.go: writeJSONL(path, v any) type-switches on `any` and silently writes an empty file for unknown types (:47-53); json.Marshal errors dropped (:58, finalize.go:152, delivery.go:18).
- quiescence.go and verdict.go are grab-bags: Domain(), Digest(), SetFailureMode(), UnblindedStamp(), AppliedPerturbations() live in quiescence.go; Ledger(), TraceDigest(), Unblind(), Reproducible() in verdict.go.
- commands.go: two recording idioms (recordWorldCommand/recordCommandAt vs hand-built append in ClearFault/ClearPerturb/AddEntity/RetireEntity) with different AtNS semantics (ClearFault uses World.Clock(), AddEntity uses atNS).
- truth.go: BuildRecord (pure) + Store (stateful, locked) in one file; BuildRecord has 10 positional params, 4 same-typed adjacent (entityID,faultID strings; onsetNS,startNS int64).
- sink.File.Close reads the whole file back from disk (sink.go:99): I/O in a "close".

## 5. Vocabulary / untyped data
- `faultID` means fault TYPE in InjectFault but fault INSTANCE in ClearFault; command args "fault" vs "fault_id"; "perturbation"(name) vs "perturb_id" (commands.go:19,33,47,60).
- history / stateSnapshot / world_state_history / RecordHistory; evidence vs trace vs ledger vs artifact used loosely (SetEvidenceRecorder is a pre-render event hook, not evidence).
- Counts.FaultsInjected (artifact_metadata.go:24) = ActiveFaultsCount, i.e. faults still in force at End (world/faults.go:65-79 excludes cleared). Name lies.
- TimeMode scaled/wall are labels: no code scales or wall-drives delivery; scaled counted reproducible (initialize.go:75). Sink package doc (sink.go:1-4,106-109) promises a simdet-tagged wall sub-mode that does not exist.
- Error prefixes: "run:", "streamsim:", "End:", "InvokeEffector:", "Solve:", "buildWorld:" (visible via mcp errTool text; changing them changes output).
- Untyped: model.Command.Args map[string]any decoded ad hoc by commandString/commandTime/asMap (replay.go:112-138) with silent zero defaults; ReplayResult.FirstDivergence *int; op, reason, sink, time-mode, detector form are bare strings (consts exist in model); Config.SinkName/TimeMode strings.

## 6. Probable bugs / hazards (ranked)
1. HIGH, conf high: ReplayResult.FirstDivergence is meaningless (replay.go:145-155). It compares len(ledger) to Counts.Emitted, but ledger length includes duplicates/dropped; the comment promises "first differing line" and does not do it. Reported to users via mcp verify (run.go:176). Artifact lacks the trace, so a real index needs the original trace passed in or stored.
2. HIGH, conf high: sealed-oracle gate fails open. truth.Store.Reveal only refuses on an open run `if s.OpenChecker != nil` (truth.go:86); NewStore() returns nil checker, wired after construction (mcp/director.go:47). Any other constructor path reveals truth mid-run silently. Non-negotiable "sealed oracle".
3. MED-HIGH, conf medium: failed runs can never be closed through MCP. End returns (art, runErr) after setting finished (finalize.go:34,42); mcp finalizeRun returns on err before `w.RunEnded = true` (mcp/run.go:143-153). OpenChecker then says "open" forever: truth reveal/score refused, second End -> "already finished". Contract smell: End returns artifact AND error.
4. MED, conf high: Counts.FaultsInjected is wrong after any fault.clear (see 5).
5. MED, conf medium-high: replay fidelity gap. worldDigest hashes Noiseless, ForceFailureMode, ClockMultiplier (identity.go:41-43) but artifact world_config omits the first two (artifact_metadata.go:49-56) and replayConfig never sets ClockMultiplier although the artifact stores it (replay.go:86-96). Replay then fails "world digest mismatch" (replay_execute.go:17). Latent today since production never sets them, but tests/CLI flags that do cannot round-trip.
6. MED, conf medium: epoch-zero start. `FirstObservableNS == 0` is both "unset" and a legal t=0 (solver_scan.go:54-58,48; record.go:36; suite/scenario.go:21), while run.Config.StartTimeSet explicitly makes 0 legal (run.go:35). With start_time=0 an at-onset detection is treated as "never observable"/skipped. Oracle error, confident wrong answer.
7. MED, conf medium: zero/missing sigma makes every fault trivially observable: threshold = SNR*0 (solver_scan.go:54, detector_math.go:52-58) -> FirstObservable = scan start. Schema permits sigma 0 (domain-spec schema: minimum 0). Unknown detector form: quantity 0 and sigma 0 -> same (solver.go:94-106,128-140). Should error.
8. MED, conf medium: peer sigma uses len(entityIDs) argument (solver.go:135) while the world's real entity count may be the domain default when entityIDs is empty (InitialEntities falls back); readings use w.EntityIDs() (detector_math.go:42). Mismatch -> wrong effective_sigma. Verify with caller that passes nil.
9. MED, conf high: partial-failure fd leaks. New: openLedger opens file, later openAdapter/openSink/beginTrace failure returns without closing ledgerFile/sink (initialize.go:60-71). End: Sink.Close error returns before closeDurableLedger (finalize.go:68-76); finishRun sets finished=true before publish, so a publish failure is unrecoverable.
10. MED, conf medium: no `finished` guard on Advance, InjectFault, ClearFault, ApplyPerturb, ClearPerturb, AddEntity, RetireEntity, EnvInject (only InvokeEffector/SubmitVerdict check). mcp world handlers do not check RunEnded either. Post-End Advance writes to a closed File sink / mutates ledger and history after the artifact was written.
11. MED-LOW, conf medium: unsynchronized reads/writes (Ledger, Verdict, History, Unblind, Reproducible) race with commands/SubmitVerdict on operator goroutines.
12. MED-LOW, conf high: delivery-path divergence (section 4) = evidenceRec semantics and ledger EventTimeNS differ by path; flush path continues to write after sink failure. Decide intent before unifying.
13. LOW-MED: EnvInject never checks `target` against envTargets (commands.go:121-132); envTargets map write-only. Command args/params stored by reference (commands.go:19,47,75): caller mutation rewrites the log/artifact.
14. LOW: advanceThroughBoundary returns early on runErr WITHOUT logging clock.advance (advance.go:37-39) though the world moved; replay of an incomplete run cannot match.
15. LOW: HTTPPush POSTs synchronously under commandMu with a 30s timeout; response body not drained so keep-alive is lost (sink.go:135); File.Close not idempotent (second call errors at Sync); Inproc.Close returns its internal slice (no copy, unlike HTTPPush.Close).
16. LOW: sink.NewFile has a dead branch: inner `if err != nil` is always true, `return nil, nil` unreachable (sink.go:53-58); empty target gives an opaque os.Create error (http-push validates, file does not).
17. LOW: unused params (Solver.quantity entityIDs, effectiveSigma entityID); BuildRecord aliases caller's `perturbations` slice (truth.go/record.go:34; Store.Seal clones, so only pre-seal); Store.sealed duplicates presence in labels; Reveal(unblind=true) on a CLOSED run still stamps unblinded; Run.unblinded and Store.unblinded are two sources of truth reconciled by mcp (mcp/run.go:61-68).
18. LOW: internal/world/determinism_test.go lists a stale "../ledger" root (package no longer exists) and omits device packages: guard silently narrower than it reads.

## 7. Target shape and refactor rounds
Target (names follow "one named responsibility per package"; each new package needs an entry in test/architecture/dependencies_test.go):
- internal/run (facade + use cases): run.go, config.go, advance.go, commands.go (single `record` helper), delivery.go (single outcome pipeline), lifecycle/end.go, replay*.go, accessors split by topic (read_model.go: Ledger/History/Verdict/...). Imports adapter, domain, model, perturb, sink, world + the two helpers below.
- internal/run/durable (or `internal/evidence`): `LedgerFile{Append(rec), Flush(), Sync+Close()}` and the artifact directory publisher `Publish(dir, Bundle)` + `LoadArtifact(path)` + strict decode. Only file/os code in run lives here. Caller: run.Run.
- internal/run/quiesce: `Barrier{Report(ns), Await(ctx, ns)}` + `Clock` interface + real timer adapter (the only time.NewTimer). Caller: run.Run.
- Pure artifact assembly: function `buildArtifact(snapshot, now)` taking a plain struct; clock passed in from the edge (use internal/wall.Now so simdet pins it).
- adapter.Digest(a) moves next to adapter (value unchanged: DigestBytes(json.Marshal)); run/identity.go keeps only world identity.
- truth: store.go (Store, constructor takes required isOpen func; fail-closed), record.go (BuildRecord taking an `Injection` struct), solver*.go unchanged; SetupCall -> model.
- sink: keep one package; split inproc.go / file.go / httppush.go; fix dead branch + doc; no new packages.
Do NOT add: a consumer-side interface for score (no second implementation), generic "store"/"service" layers.

Rounds (each independently green on `make ci-check`; keep goldens untouched):
R1 Delete/unexport dead surface (RecordHistory, RenderRecord, Unblinded, Perturb/Engine/Sink fields, Seed() accessor, QuiescenceClock exports if tests adapt). Replace local fnv with randutil.Fnv1a64. Hazard: RunID value. Proof: golden_test + a new table test pinning RunID for 3 (domain, seed) pairs computed BEFORE the change.
R2 Regroup accessors into named files (pure moves, no code change). Proof: build + architecture size/func tests.
R3 Extract durable ledger writer. Hazards: flush at every command boundary, fsync only at End, bytes identical (json.Marshal + '\n'), O_APPEND|0600, error text "run: flush ledger". Proof: ledger_test.go, finalization_regression_test.go, soak/fuzz; add a test closing the file on New failure (fixes leak, separate commit).
R4 Extract artifact publisher/loader. Hazards: write order trace -> ledger -> history -> verdict -> run.json, 0600/0700, trace rewritten at SinkTarget, error prefixes ("End: ..."), ReplayArtifact(LoadArtifact) bytes. Proof: artifact_test.go, replay_params_test.go, golden.
R5 Extract quiesce barrier. Hazards: close-and-replace channel under quiesceMu, one timer per await, parked-hook ordering, ctx.Err precedence vs timeout (select is random when both ready; keep the same select). Proof: quiescence_test.go under `-race`.
R6 Unify delivery paths. FIRST add characterization tests for the 3 divergences (evidenceRec timing, ledger EventTimeNS under rewrite/delay, continue-after-sink-failure) as they are today; then refactor into shared steps with two thin entry points. Any chosen unification of semantics is a separate, flagged commit with a regression test. Hazards: ledger row order, delivery IDs, RNG untouched (perturb called only from onEmit/Flush).
R7 Commands: one recordCommand; clone params/args at record time; typed per-op decoding in replay with errors on missing/mistyped args (behaviour change, own commit; hazard: json.Number vs float64 vs int64 in commandTime; artifacts with legacy shapes). Fix FaultsInjected semantic only with golden artifact review (JSON counts change -> flagged).
R8 Truth: split store.go; NewStore(isOpen) (fail-closed); BuildRecord param struct; error on zero/unknown sigma; epoch-zero sentinel -> explicit Observable bool (JSON shape unchanged: keep first_observable_time_ns). Hazards: label bytes feed digests/scores; run suite generation golden before/after. Proof: truth_test (analytic cross-check), suite goldens, audit tests.
R9 Sink tidy: dead branch, doc, Close idempotent, drain body, 0o600 perms (perm change is the only observable edge). Proof: sink_test + sink_equivalence_test (run).
R10 Lifecycle guards: finished check on all mutators; End contract (set finished only after success or return artifact + typed error mcp handles; fix mcp RunEnded). Needs a design note: changes mcp-visible behaviour.
