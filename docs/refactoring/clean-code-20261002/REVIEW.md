# Refactoring review and completion evidence

Scope: the clean-code and package-ownership bar in [BAR.md](BAR.md), starting at
`4e26cc1`. The original working-tree changes in `docs/README.md` and
`docs/reviews/` are preserved and excluded from this program's commits.

## Structure assessment

The repository has one Go module and 21 internal packages. Existing packages
already name business responsibilities. Splitting their files is sufficient;
there is no present need for another Go module, generic service/store layers or
an additional business package. `test/architecture` is an executable review guard,
not a runtime layer.

| Responsibility | Owner | Review decision |
| --- | --- | --- |
| Flags, application composition, CLI workflows | `cli`, `cmd/streamsim` | Name parsing/loading/execution steps; keep process entry thin. |
| Director/operator protocol and capability surfaces | `mcp` | Register tools by operation group; preserve schemas, ordering and roles. |
| Seeded state, time, observations, faults, effects | `world` | Keep mechanics here; preserve numerical and random-stream ordering. |
| Delivery corruption, buffering and terminal outcomes | `perturb` | Dispatch to named transforms; preserve short-circuit and draw order. |
| Projection expressions and byte rendering | `adapter` | Separate validation, expression evaluation, encoding and verification. |
| Record transport | `sink` | Retain the existing small destination implementations. |
| Commands, ledger, run lifecycle, artifacts and replay | `run` | Named lifecycle steps; retain locked mutation and unlocked quiescence. |
| Generated labels and independent solver | `truth` | Keep existing oracle implementation and tests. |
| Instrument, consumer, loop and judgment metrics | `score` | Share online/offline policy ownership; leave history-dependent metrics online. |
| Trivial detectors over delivered evidence | `audit` | Capture immutable delivery series separately from grid projection. |
| Scenario sampling, auditing, admission and composition | `suite` | A concrete generation context replaces a long workflow and parameter list. |
| Reference detection and operator-client integration | `refconsumer` | Name consumption, silence detection and reporting steps. |
| Capability admission and device wire exchange | `device` | Separate boot/freshness/safety checks and ordered frame exchange. |
| Device-to-world binding | `deviceworld` | Retain the leaf integration bridge; neither core imports it. |
| Domain data loading and validation | `domain` | Validate channels, dynamics, faults, effectors and profiles in named phases. |
| Records, schemas, validation, digests, random/time utilities | `model`, `schemas`, `jsonschema`, `canonical`, `randutil`, `wall` | Specific foundations; no dependency on application orchestration. |

All 226 Go files are at most 300 physical lines, including tests/comments/blanks;
the largest is `internal/model/domain.go` at 298. The [inventory](INVENTORY.md)
records each file. File names identify responsibilities; no numbered overflow
files, generated-file waivers or new runtime dependencies were introduced.

The dependency allowlist checks direct imports and rejects upward edges, unknown
packages and new external runtime dependencies. The root documentation package is
also checked. Build-tagged tool imports are explicitly limited to the three
existing development tools. Go rejects cycles. These structural checks complement
the runtime tests; they cannot establish semantic isolation on their own.

## Function decisions

The parsed production inventory changed from 545 functions with 37 bodies longer
than 60 lines to 673 functions with three such bodies. Body length includes braces,
comments and blanks. Named workflow steps account for the increase; no abstraction
framework was introduced.

| Retained function | Body lines | Rationale and evidence |
| --- | ---: | --- |
| `canonical.writeValue` in `encoding.go` | 66 | One closed JSON-type encoding operation. Integer-width cases are concrete peers; object encoding already delegates to `writeObject`. Keep the visible supported-type/error boundary together. Existing serialization/digest tests retain the current byte contract; known RFC edge cases remain separate correctness work. |
| `mcp.toolSchema` in `schemas.go` | 185 | One declarative public input-contract table. Its property/required/conditional definitions are reviewable beside the tool names. Wire validation, closed-schema and nested-contract tests protect it. Splitting this table would distribute a single public contract across callbacks. |
| `world.World.rk4Step` in `integration.go` | 72 | One declared integration step. Keep special-form updates, clamp/time updates and RK4 arithmetic visible together. Existing independent first-order, RC-network, integrator and dead-time oracles protect the arithmetic. This does not fill remaining oracle gaps. |

`functions_test.go` enforces the 60-line review threshold and these exact upper
bounds. A new long function, growth of a retained function or a stale decision
fails. Anonymous functions above the threshold must become named operations.
Length is a review trigger, not proof that every short function is cohesive.

Self-review inspected the changed public flows and private steps for intent,
abstraction level and order. It specifically checked validation error precedence,
PRNG draw order, emission/delivery identity, buffered versus immediate recorder
ordering, argument validation before idempotency, lease/freshness/safe-stop order,
receipt-before-result framing, command logging before consumer waits, and suite
rejected-attempt identity. Round-specific regressions are in [ROUNDS.md](ROUNDS.md).

## Deliberate behavior corrections

- `scorecard-bundle-v0.2`: offline judgment selects the earliest detection;
  identity-conflict gating and entity-specific action credit match online scoring.
  Online malformed citations fail evidence grounding, matching offline scoring.
  JSON field layout is unchanged. Nil unavailable offline ledgers remain distinct
  from available empty ledgers. History-only metrics remain online.
- `Run.End` now flushes, syncs and closes its owned durable ledger descriptor,
  whether or not an output directory is supplied. Close regressions first failed
  before the correction. Publication failure retains its existing terminal behavior.

Other rounds preserve behavior. The large relocation round compared all 1,099
baseline declarations after whitespace normalization; bodies, signatures and
literals were unchanged. Existing runtime oracles plus file/dependency guards
cover relocation; semantically changed workflows add boundary regressions.

## Final validation

Validation completed on 2026-10-03 with Go 1.27.1 and local socket access.

| Check | Result |
| --- | --- |
| `make ci-check` | Pass: docs smoke, tidy, build, vet, lint, short/race/shuffle tests, full simdet tests, deadcode invocation, vulnerability scan and bounded fuzzing. |
| `go test -race -timeout=30m -count=1 -shuffle=on -coverprofile=… ./...` | Pass across the full suite; suite generation completes in 475.9 s. Replay, process determinism and analytic-oracle tests execute. |
| Full coverage | 69.4% overall; every modified runtime package has nonzero coverage. |
| `golangci-lint run --new-from-rev=4e26cc1 --timeout=5m` | Zero issues across all rounds. |
| Final architecture guard additions | Race/simdet tests and broad lint pass after the repository gate; all 226 files stay under the cap. |
| Import formatting preservation | All non-import declarations in the 88 formatted files are identical under AST printing. |
| Documentation links / `git diff --check` | Changed local links resolve; whitespace checks pass. |
| `pre-commit` | Not installed; optional hook invocation skipped. |

Coverage by modified runtime package: adapter 69.1%, audit 92.9%, canonical 80.4%,
CLI 26.2%, device 84.1%, deviceworld 76.8%, domain 76.1%, jsonschema 68.8%, MCP
65.7%, model 14.3%, perturb 85.2%, refconsumer 66.8%, run 68.2%, score 83.7%,
suite 83.3%, world 64.5%. Untouched `wall` has no tests and 0% coverage; module and
entrypoint packages have no runtime tests of their own.

The deadcode tool exits successfully but reports four existing unused exported
functions: `Capabilities.TargetNames`, `Device.Advance`, `Device.SetFaultSchedule`
and `wall.Now`. They are retained to preserve existing signatures. The Makefile
does not fail on reported findings; a passing gate is not a claim of zero unused
APIs. No vulnerabilities were found by the current scan.

All seven criteria in BAR.md are met for this refactoring scope: file-size
enforcement; named responsibility files; reviewed stepdown workflows and bounded
exceptions; package ownership checks; preservation/regression evidence; completed
local gates; and preserved runtime boundaries with explicit corrections.
Per-round evidence is recorded in [ROUNDS.md](ROUNDS.md).

## Remaining correctness and release work

Readability completion does not close the earlier architecture findings. In
particular: atomic/retryable artifact publication and constructor-failure cleanup;
offline artifact/associated-evidence validation and availability signaling;
canonical JSON edge cases and adapter digest representation; complete analytic
oracle coverage; device-contract governance; large-run memory/lookup scaling; and
release-manifest/remote evidence remain separate work. This program adds no remote
release or physical-HIL evidence. Existing advisory review files remain untouched.
