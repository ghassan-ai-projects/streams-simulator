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
