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

These are deliberately explicit rather than hidden behind a “mostly complete” claim.

1. **G3 reference consumer topology.** `internal/refconsumer` is a deterministic in-process
   baseline. It does not yet run out-of-process against the documented operator MCP endpoint,
   exercise reconnect/idempotency over that endpoint, or complete a closed-loop golden run for
   every shipped domain.
2. **G2 independent oracles.** The analytic cross-check covers the first-order lag. RC-network,
   availability, pink/quantization boundaries, and every supported noise/value form need
   independent oracle tests; pink noise and batch cadence currently fail closed as unsupported.
3. **G4 durable ledger.** Ledger rows are retained and written at run end. A streaming run
   needs append-only durable ledger persistence and crash-recovery/conservation evidence for
   every delivery instance beyond 10,000 records.
4. **G5 deterministic barrier proofs.** The mutex/channel barrier removes the data race and
   has an ack-during-wait race test, but timeout/cancel/reconnect/stale-generation behavior
   still needs an injectable clock and no-real-sleep deterministic tests.
5. **MCP contract strictness.** Tools advertise an object schema, but it is intentionally
   permissive. Typed per-tool schemas and rejection of unknown/missing arguments are required
   before benchmark release; the operator CLI topology also needs a documented deployable
   process boundary.
6. **Security/release evidence.** Install and run `deadcode` and `govulncheck`, add fuzz/property
   and soak/performance evidence, and produce a signed release manifest containing simulator,
   domain, adapter, suite, toolchain, and consumer digests plus independent review identity.

## Next gated order

1. Build the out-of-process operator reference harness and golden closed-loop tests.
2. Add injectable quiescence time and durable append-only ledger recovery.
3. Add independent dynamic/noise/value oracles and strict typed MCP schemas.
4. Install security/dead-code tools, run fuzz/soak/performance checks, and publish a signed
   Level 2 release manifest only when all nine gates are green.
