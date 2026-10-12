# P6a review record — `store` and `flow`

- **Reviewed**: the P6a working-tree diff against `7758dfc` (2026-10-12)
- **Reviewer**: one independent Sonnet subagent (math, determinism, oracle independence, standards), with scratch
  experiments on the built binary.
- **Outcome**: 1 High, 2 Medium, 5 Low. Resolved below, accepted, or carried.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | High | An integrated state read ahead by one reader returned its later committed value to another state reading it at an earlier step boundary, so a lag fed by a store (and, before P6a, a lag fed by a lag) depended on read order: an emitting world and a silent oracle world could differ | Fixed in the integrator for every F1 form: all integrated states of an entity advance together, earliest step first, and each keeps the snapshots of its last steps back to the entity's longest step, so an input is read at its exact boundary time (`advance.go`). Shipped outputs unchanged (behaviour pin matches). `TestAStateFedByAnIntegratedStateDoesNotDependOnReadOrder` (store and lag sources, three read orders, bit-identical) |
| 2 | Med | A store's committed level was clamped only to `f1.clamp`, not the state's `min`/`max`, so it wound up past the bound and outflows used the wound-up level | Fixed: the committed and probed levels are clamped to both; `TestAStoreLevelStaysWithinItsStateBounds` |
| 3 | Med/Low | An override could add a store with a different `dt_ms` from the base kind's stores | Fixed: the shared-step rule compares against the first store of the effective kind; a case in `TestFlowsJoinStoresOfTheirKind` |
| 4 | Low | Stiffness (`conductance / capacity × dt` large) and negative scale states are not checked | Carried (R6a-01) |
| 5 | Low | A fault or kick between step boundaries acts on a flow from the next boundary; "exact" crossing agreement holds for grid-aligned onsets | Accepted and documented (spec section 13, hybrid rule) |
| 6 | Low | Argument numbers compared as float64 lose distinction above 2^53 | Accepted: effector arguments are physical quantities; recorded in spec section 15 |
| 7 | Low | The closed-grammar test passed on type arrays, trivial patterns, typeless members and stale allowlist entries, and had no self-test | Fixed: the walk reports type arrays allowing strings, trivial patterns and typeless members, every allowlist entry must still match, and `TestTheGrammarWalkFindsAnExpressionSlippedIn` proves it reports each slipped-in form |
| 8 | Low | A map range in `request_numbers.go` needs the determinism marker | Fixed during review (gate marker present) |
