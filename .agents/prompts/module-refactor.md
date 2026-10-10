# Module Refactor Prompt

Rebuild one package into the repository's module shape: a public facade over
private internal layers with no leaks (STANDARD v2 in
`docs/refactoring/modularity-20261010/STANDARD.md`). Use it when asked to
"migrate", "modularize" or "do the same" for a package.

Standing preferences for this work:

- Plan first in `docs/refactoring/modularity-20261010/PLAN.md`; one module per
  round; one commit per round; a reviewer subagent reads the commit.
- Behaviour does not change. Defects found go to `DEFERRED.md`, not into the
  round.
- Name things in the domain's own language and keep code, storage and wire
  names identical (`UBIQUITOUS_LANGUAGE.md` records them).
- Callers keep importing the same facade path with the same exported API;
  narrow only symbols nobody outside tests uses.

## 1. Survey

1. Read every production file of the package and its tests.
2. List what other packages use:
   `grep -rhoE '\b<pkg>\.[A-Z][A-Za-z]*' --include='*.go' . | grep -v '^./internal/<pkg>/' | sort | uniq -c | sort -rn`
   and the importing packages. Symbols used only by tests are not public API.
3. List I/O, clock, goroutine, lock and global-state sites (`ioEdges`).
4. Note the tests, their helpers and which layer each one really proves.

## 2. Choose the shape

From the module's row in STANDARD §1: pure core → facade · `internal/domain`;
core with an edge → facade · `app` · `domain` · one edge package per external
system; surface → facade · `app` · edges.

## 3. Build it (bodies unchanged)

1. Write `UBIQUITOUS_LANGUAGE.md` (terms, code names, retired words).
2. `git mv` files into `internal/<pkg>/internal/<layer>/`; change the package
   clause; move edge code (files, sockets, timers) into its own edge package.
3. Facade files: `doc.go` (package comment), `api.go` (aliases for value
   types, constants, sentinel errors), `service.go` (`Config`, `New`, the
   facade type holding the internal object), `operations.go` (one documented
   delegating line per operation). Required dependencies are constructor
   arguments. Exported signatures name facade types only.
4. Move each test to the layer it proves (STANDARD T1/T2); the facade keeps
   configuration and contract tests.
5. Register: `packages` and `moduleShapes` in `test/architecture`, the
   allowlist, `ioEdges` entries for edge files, `.agents/context/architecture.md`.

## 4. Prove

`gofmt`, `go vet`, `go test -short -race` for the module and its callers,
`go test -tags simdet`, `make function-length`, `golangci-lint run`,
`make coverage-check`, `scripts/behaviour-pin` against the baseline, the
architecture gates. Prove each new gate by injecting a violation and
recording the failure in `ROUNDS.md`.

## 5. Pitfalls

- A test that needs unexported state belongs in the layer package, not the
  facade; use `export_test.go` only for a facade test seam.
- Type aliases make a type public by name but not by location; do not alias a
  behaviour-heavy aggregate, wrap it in a facade type with explicit methods.
- Do not let app import `os`/`net`; add an edge package.
- Keep RNG substream names, draw order and digest inputs byte-identical.
