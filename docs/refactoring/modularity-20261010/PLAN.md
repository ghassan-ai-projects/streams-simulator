# Plan

Baseline: `867026a` (`main` merge of PR #16), branch `improve-modularity`.
Tests, vet and `golangci-lint` are green at baseline; every package passes
`go test -short`. Production functions and files already meet the 15-line and
300-line bars; the program adds *structure*, not more splitting.

## Rules of every round

1. Plan the round here (scope, hazards, proof) before editing.
2. Change structure only (STANDARD M10). Intentional corrections are not in
   this program: they are in [DEFERRED.md](DEFERRED.md).
3. Proof, in this order, before the commit:
   `gofmt -l` → `go build ./...` → `go vet ./...` → `go test -short ./...` →
   `go test -tags simdet ./...` → `make function-length` → `golangci-lint run` →
   `scripts/behaviour-pin | diff docs/refactoring/modularity-20261010/BEHAVIOUR_PIN.txt -`.
4. An independent review subagent reads the diff against STANDARD.md and the
   behaviour hazards; its findings are fixed or recorded in `ROUNDS.md`.
5. One commit per round (`refactor(<pkg>): …`), staged by explicit path (the
   working tree carries unrelated user changes in `docs/README.md` and
   `docs/simulation-ideas/`). The commit hash goes into the status table.
6. A relocated declaration keeps its body byte-for-byte; `ROUNDS.md` records
   the source comparison. `make ci-check` runs before handoff, not each round.

## Rounds

Order is bottom-up so each round builds on cleaner dependencies. `Hazards`
are the behaviours the round could silently change.

| Round | Scope | Hazards to guard | Proof beyond the common list |
| --- | --- | --- | --- |
| R0 | This folder, standard, surveys, deferred list, behaviour pin, agent context and `AGENTS.md` update | none (docs, one script) | review |
| R1 | **Gates and lint ratchet** in `test/architecture`: `packages` kind+layer table with strict-lower-layer and no-stale-edge gates; `ioEdges` inventory (imports, clock calls and values, `go` statements) with scheduled debt; package-comment and package-map gates; `make coverage-check` with a per-package baseline table; lint adds `gocognit`, `gocyclo`, `nestif`, `dupl` (fix the one `nestif` in `jsonschema` and one `dupl` pair in `perturb`); remove the stale 60-line review table | exception tables must start equal to today's violations and may only shrink | each gate fails on an injected violation (recorded) |
| R2 | **Foundations**: delete `randutil.Picker` and the unused `model.TimeNS`; one `jsonschema.CompileJSON` and one `jsonschema.FormatErrors` replace three `mustAny` and two `formatErrs` copies; `model` keeps its two contract validators (foundation→foundation edge is allowed; error strings are contract); schema enum constants stay (they mirror contract enums) | error strings byte-identical, including the `streamsim: streamsim:` prefix stutter | run artifact tests; fuzz target |
| R3 | **domain**: file loading (`Load`, `LoadAll`, directory scan) isolated in `file.go`, the package's one declared I/O file (STANDARD: a package that is K3 only for one small reader stays one package; moving it to `cli` would force 28 call sites in 14 packages' tests through a new helper for no gate gain); delete unused `Compiled` name accessors | domain digest, error text including prefix stutter, sorted path order, `Compiled.Raw` copy | pin; `shipped_test`, replay tests; new `file_test.go` |
| R4 | **adapter**: `adapter/conformance` package (verify + fixture file branch); `Load` isolated in `adapter/file.go`; the `adapterDigest` recipe move waits for R11 (it is digest-bearing and lives with `run`'s identity code); `Engine.Meta` and `jsonEqualish` left (not worth a round) | `VerifyResult` JSON keys, `adapter verify` output, adapter digest bytes | pin (`adapter-verify/*`); golden tests |
| R5 | **world**: delete dead API, unexport never-read fields, `FailureMode` type, `invocation` value instead of 11 positional params, named `AdvanceResult`, split `clock.go` / `effector_log.go` / `availability.go` by concern | RNG substream names and draw order, idempotency window and interlock-before-replay order, effector call ledger bytes | `effector_order_test`, `determinism_test`, `oracle_test`, pin |
| R6 | **perturb**: name→transform table replaces the switch ladder; unexport `Active`; drop unused params; shared param defaults | per-perturbation RNG draws (short-circuit order), `nextID` allocation, `ActiveIDs` order | new fixed-seed golden over all 19 perturbations (committed first), `perturb_test` |
| R7 | **truth**: `store.go` (sealed `Store`) apart from solver/record; `NewStore` requires the open-run check (production wiring identical); `BuildRecord` takes an `Injection` value; `SetupCall` moves to `model` | label bytes (feed digests and scoring), reveal error precedence | analytic cross-check tests, suite golden, pin |
| R8 | **audit / suite / score / refconsumer**: shared `model.Perturbation`/`SetupCall`; `suite` uses `world.RenderID`; perturb-owned reason table read by score; `score.Evidence` so `Score` and `Offline` share one path and `score` stops importing `run`; dedupe `asFloat` and refconsumer identity literals; unexport symbols nobody uses | scorecard JSON byte-identical incl. omitempty; suite generation RNG order; nil-vs-empty ledger/calls semantics | `parity_test`, `shared_policy_test`, new golden scorecard, suite golden, pin |
| R9 | **device**: `device/contract` becomes a Go package (embed + codec + schemas); `device` core is pure (typed `commandView` parsed once, reject codes as typed consts); new `devicewire` for session loop, frame gate, `Listen`; `CapabilityCatalogDomain` moves to `device`; delete dead exports | admission order boot→freshness→target→operation→bounds; dedup-before-admit; fault-ordinal accounting; canonical frame bytes; lock scope (plant called under `mu`) | `conformance_test`, `uds_test`, `wire_test`, `transport_order_test`, `fault_schedule_test`; add a disconnect/duplicate characterization test first |
| R10 | **deviceworld + sink**: shared strict decode; delete duplicate `stateValue`; `sink` split into `inproc.go`/`file.go`/`httppush.go`, remove the unreachable branch and the false doc | sink byte streams, digests, `Close` return values | `sink_test`, `sink_equivalence_test`, pin |
| R11 | **run**: delete dead surface and test-only exports behind `export_test.go`; `internal/run/internal/durable` (ledger file + artifact publish/load); `internal/run/internal/quiesce` (barrier + timer); accessors regrouped by topic; hand-rolled FNV replaced by `randutil.Fnv1a64` | RunID value (pinned before), ledger flush points and fsync at End, write order trace→ledger→history→verdict→run.json, permissions, error prefixes, `select` precedence in await, both recording idioms keep their `AtNS` semantics | `golden_test`, `ledger_test`, `finalization_regression_test`, `quiescence_test` under `-race`, soak, fuzz, pin |
| R12 | **mcp**: surface reduced to what `cli` uses; lifecycle and audit policy leave handlers for named use-case files; constants shared with `suite`/`audit`; no race fix (D-08) | tool schemas, error codes and their precedence, resource shapes | mcp e2e tests, `prefix_test`, pin |
| R13 | **cli**: composition root only — manifest build/sign and artifact-layout loading as named functions taking an injected clock and readers; named policy constants; `cmd` unchanged | manifest bytes and signature input, flag defaults, exit codes, output text | `subprocess_test`, pin (`manifest`) |
| R14 | **Typed closed sets**: `DetectorForm`, `TransformOp`, `Encoding`, `CadenceMode`, `F1Form`, `RejectCode`, `FailureMode` as named string types | JSON/digest unchanged (named strings marshal identically); `model/adapter.go` field order and tags untouched | pin; digest pin tests from R3/R4 |
| R15 | **Test hygiene and close-out**: `paralleltest`/`tparallel`/`usetesting`; per-package coverage ≥ 70 % (cli, model, mcp, world, randutil, refconsumer, adapter below today); exception tables empty or justified; map and docs updated; full `make ci-check`, race suite, final rating | n/a | `make ci-check`, `make coverage-check` |

A round that turns out larger than one reviewable diff is split and the
extra rounds are appended with letters (`R5a`, `R5b`); the table is updated,
not rewritten.

## Not in this program

Everything in [DEFERRED.md](DEFERRED.md), in particular: any fix to D-01…D-40,
the wall-clock seam decision (P-01), RFC 8785 adapter digests (D-36), unifying
the two delivery paths (D-18), and world read-purity (D-14).

## Status

| Round | State | Commit | Notes |
| --- | --- | --- | --- |
| R0 | done | `3e81cf2` | |
| R1 | done | `52ea773` | gates, lint, coverage ratchet; review fixes folded in |
| R2 | done | `a3f6858` | review fixes amended |
| R3 | done | `33d757f` | `file.go` instead of a move to cli |
| R4 | in progress | | |
| R5 | pending | | |
| R6 | pending | | |
| R7 | pending | | |
| R8 | pending | | |
| R9 | pending | | |
| R10 | pending | | |
| R11 | pending | | |
| R12 | pending | | |
| R13 | pending | | |
| R14 | pending | | |
| R15 | pending | | |
