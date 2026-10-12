# P6d review record — supervisory control

- **Reviewed**: the P6d working-tree diff against `f18b815` (2026-10-12)
- **Reviewer**: one independent Sonnet subagent, with overlay probes outside the repository.
- **Outcome**: 1 High, 2 Medium, 6 Low. Resolved below or accepted.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | High | A renewed `revert` lease withdrew the first call's revert, so the first call's additive effect stayed forever | Fixed: the withdrawn revert is carried into the renewed one, which undoes the chain's total; `TestALeaseFallsBackUnlessRenewed` (heat 4 after two calls, 0 after the lapse and at 1000 s) |
| 2 | Med | A renewal call without the lease argument withdrew the fallback and armed nothing | Fixed: such a call neither renews nor ends the lease (same test) |
| 3 | Med | A revert ignored the effect's dead time and time constant | Fixed: a lease counts from the effect start and the revert has the effect's time constant |
| 4 | Low | The responder decides at call time | Accepted and documented (spec section 18) |
| 5 | Low | Partial and no-effect calls arm no lease or responder | Accepted and documented |
| 6 | Low | A withdrawn kick's queue event remains (director-only `next_event`) | Accepted: skipped when it pops, nothing reaches the operator |
| 7 | Low | Delta arguments in responder and safe-state lists unchecked | Fixed: responder deltas must read numeric arguments; safe states may not read arguments |
| 8 | Low | Sequencer start-up hold undocumented | Documented |
| 9 | Low | Closed-grammar entries are classification only | No change: resolution is proven by the located violation tests |
