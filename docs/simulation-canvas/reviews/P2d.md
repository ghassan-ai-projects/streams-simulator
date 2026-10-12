# P2d review record — director out of mcp

- **Reviewed**: the P2d working-tree diff against `31b6a18` (2026-10-12)
- **Reviewer**: one independent subagent; MCP stdio sessions compared byte for byte against a
  binary built at HEAD.
- **Outcome**: 0 High, 1 Medium, 4 Low. Per the owner's direction to move fast, non-High
  findings are carried rather than iterated.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| Q1 | — | Behaviour | Verified: behaviour pin matches; MCP sessions identical (initialize, tools/list, world create, advance variants, refusals) |
| Q2 | Low | `wrapcheck` exemption for the director facade | Kept: refusal text is wire text; narrowing carried with R2d-02 |
| Q3 | Med | Director tests import `mcp` | Carried (R2d-01, P3) |
| Q4 | Low | Wire-shaped results and tool names in the director | Carried (R2d-02, P10) |
| Q5 | Low | The role gate checks imports only | Accepted: the dependency allow-list covers the rest |
| Q6 | — | Lost tests | None: every test name at HEAD exists; 10 added |
