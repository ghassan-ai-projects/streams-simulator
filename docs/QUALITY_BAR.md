# Streams Simulator completion bar

Status: **required release gate**
Defined: 2026-08-11
Owner: Streams Simulator maintainers

This document converts the implementation-readiness review into an executable definition of
done. A feature is not complete because its package compiles or its nominal test passes. It
is complete only when the production path, an invariant test, and the compliance checks below
agree.

## Release levels

### Level 0 — development-safe

The code may be iterated on, but it must not be used to judge a consumer.

- `go test ./...` passes.
- `go vet ./...` passes.
- `git diff --check` passes.
- No known P0 issue is introduced or regressed.
- New behavior has a focused test and a short design note when it changes a contract.

### Level 1 — replayable prototype

The simulator may be used for local debugging, not benchmark claims.

- Level 0 is green.
- A run artifact contains or immutably references every input needed for replay.
- Replay verifies every input digest before execution and reports the first differing record.
- World state and stochastic output are entity-isolated and deterministic under metamorphic
  tests.
- Every native emission has a terminal delivery-ledger state.
- Adapter begin/record/end framing is exercised by the real sink path.
- Scores fail closed when required evidence is absent or ambiguous.

### Level 2 — benchmark-ready

The simulator may publish consumer measurements.

- Level 1 is green.
- All nine non-negotiables in the implementation-readiness review are green on production
  paths, not only direct unit helpers.
- An external reference consumer completes a closed-loop run through operator capabilities.
- Truth is sealed, immutable, and inaccessible until the declared unblind transition.
- Online and offline scoring use the same versioned bundle and produce identical results.
- Every shipped domain and adapter has independent behavioral conformance evidence.
- Suite generation is byte-identical, data-driven, and counts admitted scenarios only.
- Required lint, race, vulnerability, dead-code, fuzz/property, soak, and performance checks
  are available and fail closed when unavailable.
- A release manifest records simulator, domain, adapter, suite, toolchain, and consumer
  digests, plus independent review and author identity.

## Non-negotiable gates

The following gates are binary for Level 2. “Partial”, “not applicable”, or “covered by a
similar test” is not green without a recorded design decision and an explicit test.

| ID | Gate | Required evidence |
| --- | --- | --- |
| G1 | Determinism | Byte-identical replay and metamorphic isolation for seeds, entities, channels, faults, perturbations, and command order. |
| G2 | Analytic cross-check | Independent oracle for every supported dynamic/noise form, including boundary and negative-recovery cases. |
| G3 | Reference consumer | Out-of-process consumer uses only the documented operator surface and completes a closed loop. |
| G4 | Delivery ledger | Every emission and delivery instance has an immutable terminal state and conservation checks pass beyond 10,000 records. |
| G5 | Quiescence | Concurrent ack/timeout/cancel/reconnect tests pass without wall-clock sleeps in deterministic tests. |
| G6 | Sealed oracle | Truth lifecycle is monotonic, immutable, capability-scoped, and prefix-equivalence is proven. |
| G7 | Injection probe | Producer-controlled strings cannot alter control/oracle evidence; every delivery path is probed. |
| G8 | `silent_no_effect` | A declared no-effect command is proven not to change scoped truth, with action scoring checking the full tuple. |
| G9 | Trivial baseline | Baseline consumes the actual immutable delivered stream in order and passes known trivial/non-trivial fixtures. |

## Per-change loop

Every implementation slice follows this loop and produces one commit:

1. **Select a gate.** State the issue IDs, invariant, affected contracts, and out-of-scope
   behavior before editing.
2. **Implement the smallest coherent change.** Keep data model, runtime, and error behavior
   aligned. Do not hide unsupported input behind defaults.
3. **Add proof.** Add a focused test that would fail before the change and an adversarial or
   metamorphic test where the risk warrants it.
4. **Review the diff.** Check dependency direction, entity scope, exact numeric behavior,
   error propagation, security boundaries, and documentation/contract drift.
5. **Validate.** Run the narrow package tests first, then `go test ./...`, `go vet ./...`,
   race tests where concurrency changed, `git diff --check`, and the relevant gate command.
6. **Compliance check.** Re-read the affected design/contract section and update the review,
   decision record, or status document if behavior changed.
7. **Commit.** Commit only a green slice with a message naming the gate and issue IDs. Do not
   carry unrelated formatting or opportunistic refactors.
8. **Re-baseline.** Record the commit, remaining P0/P1 items, and the next slice before
   continuing.

## Commit acceptance checklist

Before each commit:

- [ ] The change has a stated issue/gate and a bounded scope.
- [ ] The failure existed before the change or the new invariant is demonstrated by a test.
- [ ] No test relies on wall time, map iteration order, or a mutable historical query unless
      that is the subject under test.
- [ ] Errors are returned or represented in the declared incomplete state; no new panic or
      ignored error exists on a production path.
- [ ] Exact seeds, timestamps, IDs, and command arguments retain their declared precision.
- [ ] Every new field has schema, canonicalization, replay, and validation implications
      considered.
- [ ] Domain and consumer knowledge remains in data or test fixtures, not production code.
- [ ] Targeted tests, full tests, vet/race checks, and `git diff --check` pass.
- [ ] The commit is independently revertible and the working tree contains only intended
      changes.

## Current bar status

Current status after the correctness-hardening loop: **Level 1 — replayable prototype**.

Evidence is committed in the focused slices `472a110`, `1bfa50c`, `8d21927`, `91b7eed`,
`8945048`, `5366f5d`, `b1d58d1`, `36683c1`, `07b83e3`, `c2e5e16`, `52b3ade`, `85b8407`,
`af1bc17`, and `943083c`. The full `go test ./...`, `go build ./...`, `go vet ./...`,
`git diff --check`, and `make ci-check` gates are green; `make ci-check` reports the optional
`deadcode` and `govulncheck` tools as unavailable rather than silently claiming those checks.

Level 2 is intentionally **not claimed**. The remaining benchmark-release blockers are
tracked in [IMPLEMENTATION_REBASELINE.md](IMPLEMENTATION_REBASELINE.md): a real
out-of-process operator reference consumer, strict typed MCP schemas/cancellation proofs,
fully independent analytic/noise oracles, durable streaming-ledger persistence, and a
release manifest with vulnerability/dead-code evidence. No benchmark scores should be
published until those gates have their own commits and evidence.
