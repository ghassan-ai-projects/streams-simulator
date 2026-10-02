# Refactoring rounds

## Round 1 — Bar and architecture guard

Established the 300-total-line rule, named ownership and stepdown criteria in
AGENTS.md and BAR.md. Inventoried all 103 Go files; 30 exceed the cap. The current
business package graph needs internal organization, not additional Go modules.
Added a source-based import guard that requires deliberate review for new edges.

Baseline short tests pass outside socket-dependent device/MCP/sink tests. Those
fail with `bind: operation not permitted` in the restricted environment; rerun with
local socket access before claiming the full gate.

## Round 2 — Responsibility files and enforced size limit

Partitioned all 30 oversized production/test files inside their existing packages.
Files now name loading, validation, expressions, encoding, capabilities, admission,
clock, delivery, commands, artifacts, scoring policies and test scenarios.
No package ownership or import edges changed. The source-walking size test includes
all Go files and proves handling of blank/comment lines and a missing final newline.

Compared all 1,099 baseline declarations against the partitioned declarations:
identical text after whitespace normalization (two one-line functions were realigned
by gofmt). No function bodies, signatures or data literals changed. Existing runtime
regressions and oracles remain intact. This relocation round uses that preservation
proof and the new architecture tests; behavioral rounds add focused runtime tests.

Validation: `go test -short -cover ./...` passes with local socket access; every
modified production package except suite has nonzero short-test coverage (suite's
runtime tests are deliberately non-short; full-suite coverage is required at final
validation). All 199 Go files are at most 300 lines; largest is 298. Dependency and
size guards pass. Lint and whitespace review pass.
