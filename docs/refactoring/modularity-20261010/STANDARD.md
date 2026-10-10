# Modularity standard

Adapted from the sibling repository `agentic-stream` (its quality bar Q1–Q8,
architecture bar A1–A12 and the "reference module" pattern of
`internal/authority`). That repository owns SQLite tables; this one owns
deterministic computation, JSON/JSONL artifacts, a wire transport and an MCP
surface. The *principles* carry over; the *layer names* are adapted. Nothing
here changes behaviour: the nine non-negotiables, determinism tuple and
digest inputs are fixed.

## What was taken, what was changed

| agentic-stream | This repository | Why |
| --- | --- | --- |
| facade · `internal/app` · `internal/domain` · `internal/store` | facade · `internal/app` · `internal/domain` · named edge packages (`internal/artifact`, `internal/transport`, `internal/files`, …) | There are no SQL tables. The edge is named after the external system it touches. |
| Every module gets the layers | Layers only where a package has **both** a real I/O edge **and** non-trivial rules | `AGENTS.md` forbids layers "for future flexibility". Pure cores stay one package. |
| No comments inside modules | Not adopted (see `DEFERRED.md` D-1) | Large churn, no behavioural or structural value here; exported symbols stay documented. |
| `durableOwners` table ownership gate | Artifact/ledger file ownership gate (one writer per file family) | The durable state here is run artifacts, the append-only ledger and the release manifest. |
| `packageLayers` numeric layer map, strict lower-layer imports | Adopted | Catches same-layer coupling the per-package allowlist permits. |
| Per-module dated planning folder | One program folder; `ROUNDS.md` holds the per-round record | Keeps tracking in one place. |

## Package kinds

Every production package is exactly one kind. The kind decides its shape and
which gates apply.

| Kind | Packages today | Shape | Must not |
| --- | --- | --- | --- |
| **K1 Foundation** | `canonical`, `randutil`, `schemas`, `jsonschema`, `model`, `wall` | One package, stdlib (+ lower foundations) only | Import any business package; read files, env, sockets or the clock (`wall` is the declared clock seam for pure kinds; edge and surface packages read the clock at declared `ioEdges` sites) |
| **K2 Pure core** | `world`, `perturb`, `truth`, `audit`, `score`, `suite`, `deviceworld`, `refconsumer`, `domain` rules | One package (or facade + `internal/domain` when an edge is split off) | Import `os`, `net`, `net/http`, `os/exec`, `os/signal`; read the wall clock; use global `math/rand`; hold goroutines or sockets |
| **K3 Core with an edge** | `run`, `device`, `adapter`, `domain` loading, `sink` | Facade · `internal/app` (orchestration) · `internal/domain` (pure rules) · edge package per external system | Put rules in the edge or I/O in domain/app |
| **K4 Surface / composition root** | `cmd/streamsim`, `cli`, `mcp` | Wiring, flag/protocol decoding, output formatting | Contain a business decision; build a result a lower package should own |

A package that is K3 only because of one small file reader keeps that reader in
its own file and is declared in the edge table; it is split into a package
only when a second caller or a second external system appears.

## Layer responsibilities (K3)

- **Facade** (`internal/<pkg>`): exported value types and errors callers need,
  `New(Config)`/constructors with every safety dependency required,
  one documented delegating line per operation. No loops, no I/O, no decisions.
- **app**: use cases as a sequence of domain verbs: validate → open resources →
  load → decide (domain) → persist/send (edge) → record. May not import
  `os`, `net`, `net/http`, `os/exec`.
- **domain**: vocabulary and every rule as pure functions over values. Time,
  randomness and sizes arrive as parameters. No `os`, `net`, clock reads.
- **edge**: the only code that touches files, sockets, processes, the wall
  clock or environment. Named after the external system. Methods are named
  after domain actions and decide nothing.

## Rules

Each rule has a named enforcement. "Gate" means a test in
`test/architecture`. A rule with no gate is review-enforced and listed as such.

| ID | Rule | Enforced by |
| --- | --- | --- |
| M1 | Every production package states its responsibility in a package comment and appears in the package map (`.agents/context/architecture.md`); a layered package has `UBIQUITOUS_LANGUAGE.md`. | Gates `TestEveryPackageDocumentsItsResponsibility` (comment begins `Package <name>`), `TestPackageMapListsEveryPackage`; the language file is review-only until a layered module exists |
| M2 | Imports point only to a strictly lower layer; same-layer edges are forbidden unless listed. The direct-import allowlist stays complete and has no stale edges. | Gates `TestPackageDependencies`, `TestAllowedImportsHaveNoStaleEdges` (allowlist) and `TestImportsPointToStrictlyLowerLayers` (`packages` table) |
| M3 | K1/K2 packages and every `internal/domain` layer are pure: no `os`, `net`, `os/exec`, `os/signal`, entropy imports, no wall-clock call or function value, no `go` statement, no `filepath` file-system call. Exceptions are `ioEdges` entries carrying `debt: "R<n>"`, naming a PLAN round, and burn down to empty. | Gates `TestIOStaysInDeclaredEdges`, `TestPureKindsHoldOnlyScheduledIODebt` |
| M4 | File, socket, process, entropy, goroutine and wall-clock use sits only in files declared in `ioEdges` with a reason; a declared use a file no longer has fails the gate. | Gate `TestIOStaysInDeclaredEdges` |
| M5 | `cli` and `mcp` hold wiring and protocol only: no scoring, truth, delivery, digest or world rule. Facades only delegate. | Review + Gate `TestFacadesOnlyDelegate` once a facade exists |
| M6 | Exported surface is what another package or the CLI/MCP contract uses. Symbols used by nobody are removed or unexported; test-only seams live in `export_test.go`. | Review + `make deadcode` (production reachability) |
| M7 | Records cross a boundary typed and parsed once, with a closed field set. A raw document is kept only where a digest depends on its exact bytes. | Review |
| M8 | Production functions ≤ 15 body lines at one abstraction level, entry points first (stepdown); files ≤ 300 lines; cognitive complexity ≤ 15, cyclomatic ≤ 20, nested-`if` ≤ 3; no token clone of 75+ tokens. | `make function-length`, `TestGoFileSize`, lint (`gocognit`, `gocyclo`, `nestif`, `dupl`) |
| M9 | Every package with statements has tests and ≥ 70 % statement coverage (`-short`); behaviour moved across a layer boundary keeps its original tests. Each new gate is proven by injecting a violation and watching it fail. | `make coverage-check` (floors file ratchets: below floor fails, 2 points of slack fails); injection proofs recorded in `ROUNDS.md` |
| M10 | A refactor round changes structure only: RNG draw order, digest inputs, JSON shapes, error precedence, locks, command and delivery order are identical. A deliberate change is listed in `PLAN.md` first and proven by a regression test. | Existing replay, analytic cross-check and golden fixtures run unchanged; round review |
| M11 | No duplicated implementation of a job the repo already does; wrappers whose body is one call are removed unless a gate requires them. | Lint `dupl` + the duplication scan below |

## Function and naming bar (M8 detail)

Unchanged from `docs/refactoring/clean-code-20261002/BAR.md`: names state
intent; one task at one abstraction level; public entry points read as a short
sequence of domain verbs; callees sit below their first caller. Extraction is
by named responsibility, never `partA`/`partB`.

## Duplication scan (run on every touched file)

1. Token clones: `golangci-lint` with `dupl` at 75; lower to 60 locally for files you touched.
2. Same job, different code: search the repo for the verb before adding a helper.
3. Wrappers: a function whose body is one call into another package is deleted unless a gate requires it.
4. Identical bodies: group `FuncDecl` bodies by printed form with a throwaway `go/ast` program (kept in the scratchpad, not the repo).

## Refactoring method

- Go-aware tooling for Go changes (`gopls rename`, `gofmt -r`, `goimports`, small `go/ast` programs). `sed` only for non-Go text.
- Move code with declaration bodies unchanged; prove with the existing oracles. Keep source-comparison evidence for pure relocations in the module record.
- One round = focused tests + `gofmt` + `go vet` + lint + architecture tests + independent review + one commit; the round's row in `PLAN.md` gets the commit hash.
- `make ci-check` before handoff of the program; blocked checks are reported, not skipped silently.

## Constraints carried over unchanged

`AGENTS.md` Forbidden Changes remain: domains/adapters/effectors are data; no
consumer knowledge in the binary; no broker, physics engine, ORM, expression
language or plugin system; zero external dependencies except the MCP Go SDK;
no per-domain code branches.
