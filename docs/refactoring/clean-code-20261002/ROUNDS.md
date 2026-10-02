# Refactoring rounds

## Round 1 — Bar and architecture guard

Established the 300-total-line rule, named ownership and stepdown criteria in
AGENTS.md and BAR.md. Inventoried all 103 Go files; 30 exceed the cap. The current
business package graph needs internal organization, not additional Go modules.
Added a source-based import guard that requires deliberate review for new edges.

Baseline short tests pass outside socket-dependent device/MCP/sink tests. Those
fail with `bind: operation not permitted` in the restricted environment; rerun with
local socket access before claiming the full gate.
