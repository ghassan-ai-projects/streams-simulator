# Implementation rebaseline

Date: 2026-08-11  
Baseline commit: `943083c`  
Release level: **Level 1 — replayable prototype**  
Benchmark status: **blocked; do not publish scores**

This is the result of the implementation loop defined in [QUALITY_BAR.md](QUALITY_BAR.md).
Each coherent fix was tested, reviewed, compliance-checked, and committed before the next
slice. The original issue inventory remains in [IMPLEMENTATION_READINESS_REVIEW.md](IMPLEMENTATION_READINESS_REVIEW.md);
this file records what is now proven and what still prevents Level 2.

## Green evidence

- Exact replay identity: lossless seeds/timestamps, authenticated domain/adapter digests,
  embedded validated source documents, and replay-from-artifact tests.
- World correctness: entity-scoped effects and RNG streams, churn allocation, RC-network
  two-time-constant integration, fault onset magnitude, bounded stochastic onset, availability
  renewal, native boolean/counter values, dependency-cycle rejection, and fail-closed
  unsupported batch/pink declarations.
- Delivery integrity: monotonic delivery-instance IDs, duplicate/reorder/hold/drop ledger
  states, no history truncation, adapter lifecycle framing, and incomplete-run errors.
- Scoring/oracle safety: immutable truth copies and unblind state, one-to-one evidence/action
  matching, exact labels, negative false-positive handling, and defensive run views.
- Control plane: opaque capability tokens, distinct director world/run identities, explicit
  truth sealing before `run.begin`, request-context propagation, declared MCP object schemas,
  sink targets, and a usable CLI default sink.
- Suite/reference baseline: deterministic weighted selection, admitted-only composition
  counters and IDs, data-defined profile setup, end-of-run silence handling, and record counts.
- Validation: `go test ./...`, `go build ./...`, `go vet ./...`, `git diff --check`, and
  `make ci-check` pass. The CI command explicitly reports that `deadcode` and `govulncheck`
  are not installed; those are not represented as green evidence.

## Remaining Level 2 blockers

Status: **all resolved** (2026-08-11, slices A–G). The five blockers below each have
their own commits and evidence; see [QUALITY_BAR.md](QUALITY_BAR.md) current-bar-status
for the commit map. The resolved state of each:

1. **G3 reference consumer topology** — resolved by `1c8a62b`: the operator role is
   served over streamable HTTP from the director process (`--operator-addr`), the
   reference consumer runs out-of-process against it (`refconsumer --mcp/--token/--run`),
   and a golden closed-loop run (fault → evidence → detection → effector over MCP →
   world effect → verdict → score) resolves per shipped domain; director tools are
   provably unreachable from the operator endpoint.
2. **G2 independent oracles** — resolved by `2977d15`: analytic closed forms now verify
   rc_network, integrator, threshold_integrator, dead_time (discrete recurrence), the F0
   trend/seasonality sum, and the fault envelopes; gaussian/quantization moment oracles
   and availability sojourn statistics guard the noise and renewal machinery (the
   availability oracle caught and fixed a real seconds-as-nanoseconds bug); pink noise
   and batch cadence fail closed with rejection tests.
3. **G4 durable ledger** — resolved by `e30a059`: append-only ledger persistence with
   sink-consistent flushing at command boundaries, crash recovery at command granularity,
   and conservation identity proven beyond 10,000 records (and beyond 1,000,000 in the
   `make soak` target).
4. **G5 deterministic barrier proofs** — resolved by `9840200`: injectable quiescence
   clock, request cancellation, timeout/cancel/stale-report/fast-path tests with no real
   sleeps, and replayable timed-out advances (incomplete runs replay as incomplete).
5. **MCP contract strictness** — resolved by `20d4818` + `1c8a62b`: every tool advertises
   a closed typed schema; verdict/ground-truth arguments are validated against the
   committed contracts over the wire; params/args reject unknown keys; the operator
   process boundary is documented and deployable.
6. **Security/release evidence** — resolved by the release slice: `make ci-check` fails
   closed without `deadcode`/`govulncheck` (both installed and green), bounded fuzzing of
   the parsers runs in ci-check, `make soak` and `make perf` exist and pass, and
   `streamsim manifest` emits a signed release manifest with simulator/domain/adapter/
   suite/toolchain/consumer digests plus author and independent review identity.

## Next gated order

1. Build the out-of-process operator reference harness and golden closed-loop tests.
2. Add injectable quiescence time and durable append-only ledger recovery.
3. Add independent dynamic/noise/value oracles and strict typed MCP schemas.
4. Install security/dead-code tools, run fuzz/soak/performance checks, and publish a signed
   Level 2 release manifest only when all nine gates are green.
