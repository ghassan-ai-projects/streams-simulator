# Hardening program — 2026-10-10

Follows [modularity-20261010](../modularity-20261010/README.md). That program
restructured the code and recorded what it found; this one fixes it. Unlike
the structural rounds, a round here **changes behaviour on purpose**.

Rules for every item:

1. Reproduce first: a test that fails on `main` and names the defect. An item
   that does not reproduce is closed as "not reproducible" with the reason.
2. One fix per commit, with its regression test and the pin impact stated.
   `scripts/behaviour-pin` must still match `BEHAVIOUR_PIN.txt`; if a fix
   moves a pinned hash the commit says which and why, and updates the pin.
3. A fix that changes a digest, stored artifact format or any number a
   consumer could depend on is a **design decision**: it is listed under
   "Needs a decision" with a recommendation and is not applied unprompted.
4. Source of truth for the findings is
   [DEFERRED.md](../modularity-20261010/DEFERRED.md); its status column here
   is [PLAN.md](PLAN.md).

## Outcome (2026-10-10)

- Fixed with a regression test: D-01…D-13, D-15…D-27, D-30…D-34, D-36, D-38,
  D-40, D-41, D-42, D-43, D-44, D-45, D-46, D-47, D-48, D-51, T-01, P-01,
  P-03 (failure modes), P-05, P-06, and the unified delivery path (D-18/P-07).
- Closed without a code change, with the reason in DEFERRED.md: D-28 (the
  synchronous push is the design, and the response body needs no drain),
  D-35 and D-39 (the contract choices are kept and now documented), D-50 (not
  a defect: the caller updates `lastSent`), D-37 (already deleted).
- Decided and not adopted: P-02 (comment policy) and P-04 (typed device
  records). The run facade's test-harness hooks stay public because
  cross-package tests need them; each is documented as a harness hook.
- Found while fixing: D-52 (a cold-chain detector channel its fault never
  changes) is left to the domain owner because changing it changes the
  domain digest.
- Simulator 0.1.0 → 0.2.0: read-schedule-independent integration (D-14), the
  oracle solver (D-11/12/13), the reorder draw (D-26) and the canonical
  adapter digest (D-36) change numbers or digests; the pin file was
  regenerated and says which outputs moved.
- Independent review of the branch (PR #19) found six further defects, fixed
  with regression tests as R-1 … R-6 in [PLAN.md](PLAN.md): a `run.begin` /
  `truth.reveal` lock-order deadlock, replay aborting on a logged refusal,
  the idempotency key, the version-before-digest check on replay,
  `DestroyWorld` after a failed `End`, and director reads outside the run's
  command lock. The decisions that change numbers, digests or stored formats
  are listed in PLAN.md under "Decisions taken" for the reviewer to confirm.
