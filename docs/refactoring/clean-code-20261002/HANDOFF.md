# Refactoring handoff — 2026-10-03

This branch is an incremental refactoring, not completion of the stricter bar.
The current requirement is 15 physical body lines, including braces, comments
and blanks. The earlier 25-line goal text is superseded by the user's 15-line
selection and AGENTS.md. Tests are exempt from function length.

## Completed

- Responsibility-based files and a downward business-package dependency guard.
- All 243 Go files are within 300 lines; the largest is 298 lines.
- Shared online/offline scoring policies and named simulator workflows.
- Every production function in canonical, jsonschema, domain and score is within
  15 lines. Other packages have completed workflow subsets, not package-wide compliance.
- Strict AST source audit: `make function-length`, including anonymous functions, methods
  and build-tagged production sources, without exemptions.
- Enola baseline/change review with enforced cycle/layer checks at confidence 0.8.
  Generated architecture state remains local and ignored.
- Test files are unchanged since round 11, before the source-only migration.
  Earlier rounds on this branch include test partitioning and regression tests.

## Remaining function-length work

The strict audit still fails for **211 production bodies**:

| Package | Bodies over 15 lines |
| --- | ---: |
| device | 34 |
| world | 31 |
| mcp | 24 |
| cli | 19 |
| adapter | 18 |
| run | 18 |
| perturb | 15 |
| audit | 12 |
| suite | 12 |
| deviceworld | 8 |
| refconsumer | 8 |
| truth | 7 |
| model | 2 |
| sink | 2 |
| randutil | 1 |

Prioritize long shared admission/rendering/command workflows and duplication;
then remaining concrete helpers. Preserve command and delivery order, RNG draws,
locks, digest representations and error priority. No new package is justified
solely by length. Keep tests unchanged under the current instruction.

Once all bodies comply, add the strict check to CI and reconcile the legacy
60-line architecture-review test with the authoritative source checker. Today,
`make ci-check` does not include `make function-length`; a green CI run cannot
prove the new bar. Finish a readability review of stepdown order and abstraction
levels; length alone is insufficient.

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

## Validation and publication

`make ci-check` passes: documentation smoke checks, module tidy, build, vet,
lint, short race tests, deterministic tests, vulnerability scan and bounded
fuzzing. No vulnerabilities were reported. Deadcode reports four existing
unreachable symbols (Capabilities.TargetNames, Device.Advance,
Device.SetFaultSchedule and wall.Now); its successful exit is not zero findings.
Pre-commit is unavailable. The full shuffled race-and-coverage run (`go test -race -count=1 -shuffle=on
-coverprofile=... ./...`) passes with 69.4% total coverage; score coverage is
84.2%. The final Enola check is clean and comparable, enforcing cycles/layers at
0.8 with no new findings. Go dependency/file-size guards and whitespace checks
pass. The strict function checker still fails with the 211 bodies listed above.

The PR remains open with the 15-line migration and correctness follow-ups
explicitly incomplete. Local
Enola evidence complements, rather than replaces, the Go dependency guard and
behavioral oracles. Remote CI must be checked independently after publication.

A concurrent publication commit (`b6851c6`) included the pre-existing docs/README.md
update and docs/reviews/ architecture-review documents. They were excluded from
earlier refactoring commits; they are now part of this PR and retained as
published. Their historical claims do not replace this current validation.

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
