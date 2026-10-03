# Refactoring completion — 2026-10-03

The strict clean-code and architecture bar is complete for this refactoring
scope. Every production Go function, method and anonymous function is at most
15 physical body lines, including braces, comments and blanks. Test functions
are exempt; no production exceptions or frozen-body allowances remain.

## Completed structure and review

- All 284 Go files, including tests, are within 300 physical lines; maximum 294.
- Responsibility-based files and named workflow steps cover all production
  packages. Public operations delegate admission, execution and reporting to
  concrete helpers. No extra runtime module/package or external dependency was
  introduced for size; the existing business ownership map remains appropriate.
- The AST checker includes build-tagged source and callbacks. It is enforced by
  `make function-length`, `make ci-check`, GitHub CI and the local pre-commit hook.
- Disposable checker probes reject 16-line named functions, methods, callbacks
  and build-tagged functions; 15-line equivalents pass. Invalid test source is
  excluded as required. No repository tests were edited for these probes.
- The existing import allowlist protects downward business dependencies and the
  deviceworld integration bridge. Go compilation rejects import cycles.
- Enola uses retained, comparable pre-change baselines and enforced cycle/layer
  checks at confidence 0.8. Generated state stays local and ignored. Its inferred
  command/internal layers complement the detailed Go business import guard.
- Self-review checked intent, abstraction levels, stepdown helpers, error
  priority, random draws, arithmetic order, JSON/digest representations, locks,
  command/effect/delivery sequence, cancellation and resource close behavior.
- Test files are byte-for-byte unchanged since round 11 (`7ee077a`), before the
  stricter source-only migration. Earlier rounds include test partitioning and
  regression tests for separately identified behavior corrections.
- Each round is committed after focused checks and a single retained-baseline
  Enola comparison. Current source migration has zero overlength bodies.

## Final validation

| Check | Result |
| --- | --- |
| Strict production function check | Pass; zero bodies above 15 lines |
| Go file inventory | Pass; 284 files, largest 294 lines |
| `make ci-check` | Pass, including build, vet, lint, race/shuffle, determinism, fuzz and vulnerability checks |
| Full shuffled race suite with coverage | Pass; longest package (suite) 446.058 seconds |
| Repository coverage | 69.9%; every modified runtime package has nonzero coverage |
| Enola retained-baseline comparison | Clean and comparable; zero added findings; cycles/layers enforced at confidence 0.8 |
| Documentation, links and whitespace | Pass |
| Source-only migration | Test files unchanged since `7ee077a` |

Enola's final graph delta contains documentation declarations and name links;
there are no new runtime dependency edges. The vulnerability check reports no
vulnerabilities. The passing deadcode command still reports four existing unused
symbols: Capabilities.TargetNames, Device.Advance, Device.SetFaultSchedule and
wall.Now; passing CI does not establish that all unused APIs are eliminated.

Pre-commit is not installed, so its optional invocation was skipped. Its new
hook runs the separately verified source checker. Local logs and coverage stay
outside the repository; remote CI status is reported separately at publication.

## Publication scope

PR [#16](https://github.com/ghassan-ai-projects/streams-simulator/pull/16) contains
round-scoped commits and this completion evidence. Publication requires exact
local, remote and PR-head verification after pushing. Remote CI status is
reported separately from local results.

A concurrent publication commit (`b6851c6`) included pre-existing docs/README.md
and docs/reviews/ changes. They were excluded from earlier refactoring commits
and retained as published. Historical review claims do not replace current
validation; REVIEW.md explicitly labels the former 60-line completion evidence.

## Improvement scan: follow-up correctness work

Items 1 and 3 were reproduced with a disposable external Go harness using the
current production APIs. Items 2, 4 and 5 are source-confirmed risks, not newly
reproduced failures. None was fixed in this readability round. Separate behavioral corrections from refactoring and
obtain regression evidence before claiming resolution.

1. **Adapter hash-suffix bounds.** `adapter/engine.go:rewriteOverflow` slices at
   `MaxLength - len(hash)` without checking that the maximum can fit the suffix.
   The probe loaded the shipped native adapter with max_length=1 and
   on_violation=hash_suffix through LoadBytes, then rendered a long entity ID:
   `runtime error: slice bounds out of range [:-15]`. Reject unsupported lengths
   or define a safe result.
   This is a small, high-impact candidate because it can panic during delivery.
2. **Recovery scoring scope.** `score/loop.go:firstSuccessfulCall` selects by
   effector without entity, while deadline scoring uses both. It also uses zero
   as the absence sentinel. Prove multi-entity and zero-time cases before changing
   recovery metrics and scoring-bundle semantics.
3. **Authority-log ownership.** `world/effector_log.go:recordCall` stores the
   supplied Args map, and `EffectorCalls` only copies the outer slice. The probe invoked
   start_aerator with level=1, changed the supplied map to 0.25, then changed a
   returned call map to 0.75; both changes appeared in subsequent log reads. Establish deep-copy
   ownership, including nested JSON values, and idempotent replay behavior.
4. **Failed-constructor cleanup.** `run/initialize.go:openPipeline` acquires the
   ledger before adapter/sink initialization; a later error returns without
   closing acquired resources. Add staged cleanup with failure-path evidence.
5. **Artifact publication failure.** `run/finalize.go:finishRun` marks the run
   finished before publishing evidence. Publication failure can leave partial
   output and subsequent End calls cannot retry. Define publication/retry
   semantics without replaying already completed effects or closing twice.

Further investigation remains for offline ledger availability, ignored JSONL
serialization/close errors and canonical JSON edge cases. These are not covered
by Enola's inferred layers. Do not turn heuristic performance warnings into defect
claims without measuring the actual path; action matching remains a nested scan.

## Probe reproduction

No repository test files were edited. A disposable external Go module used the
repository module through a local replacement and called its public APIs:

- Parse the shipped native-jsonl adapter as JSON; set entity_id_rewrite to
  `{"max_length":1,"on_violation":"hash_suffix"}`; call adapter.LoadBytes,
  adapter.NewEngine, then RenderRecord with EntityID `long-entity-id` and valid
  event/observation timestamps. Admission succeeds; rendering panics above.
- Load aquaculture-pond; create a seeded world with entity pond-1, the default
  start time and forced ModeOK. Invoke start_aerator with a unique command and
  `{"pond_id":"pond-1","level":1}`. Mutate the caller's level, then the level
  in EffectorCalls()[0].Args. A fresh EffectorCalls read exposes each mutation.

The behavior predates this round. The source-only migration intentionally
preserves it; these reproductions establish follow-up priorities, not a claim
that the simulator is release-ready or that all prior review issues are fixed.
