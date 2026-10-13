# P2b review record — operator truth leak

- **Reviewed**: the P2b working-tree diff against `ed05e15` (2026-10-12)
- **Reviewer**: one independent subagent (correctness of the truth boundary), per the owner's
  direction to keep reviews light.
- **Outcome**: 0 High, 1 Medium, 3 Low. Resolved below or carried in [FINDINGS.md](../../../governance/FINDINGS.md).

| # | Sev | Finding | Disposition |
|---|---|---|---|
| M1 | Med | `ack_lost` still told the consumer the effect applied (`reason: "ack_lost"` plus an effect time) | Fixed: a lost acknowledgement reads as a timeout, `reason: "acknowledgement_timeout"` with no effect time; `TestAnUnacknowledgedCallRevealsNoOutcome` pins both unacknowledged responses; MCP_SURFACE §4.3 states it |
| L1 | Low | Tests exercise the view, not the MCP wire | Carried (R2b-01) |
| L2 | Low | The allow-list test admits any `reason` | Fixed by the exact-bytes test for both unacknowledged modes |
| L3 | Low | `interlock_refused` and `invalid_args` are documented but never emitted | Carried (R2b-01) |
