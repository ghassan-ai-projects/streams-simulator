# Structural findings

Evidence: four read-only surveys in [survey/](survey/) (file:line references,
confidence tags). This page records the verdict per package against the
[standard](STANDARD.md). Defects are in [DEFERRED.md](DEFERRED.md).

## Where the repository already meets the bar

- 300-line files, 15-line production functions (AST checker, no exemptions),
  downward direct-import allowlist, 0 lint issues, all packages green.
- `world`, `perturb`, `truth`, `audit`, `score`, `suite`, `deviceworld`,
  `refconsumer`, `model`, `jsonschema`, `canonical`, `randutil`, `schemas`
  import no `os`, `net` or clock: the pure cores are already pure.
- Domains, adapters and effectors are data; no consumer knowledge in `internal/`.

## Where it does not (the program's targets)

| Package | Kind | Gap | Round |
| --- | --- | --- | --- |
| `domain` | K2→K3 | Reads files (`Load`, `LoadAll`) inside the compile/validate package; unused accessors (helper duplication fixed in R2) | R3 |
| `adapter` | K2→K3 | Conformance harness (`verify*`, fixture file branch) ships inside the render engine package; adapter digest recipe lives in `run` | R4 |
| `model` | K1 | Hosts two contract validators with two callers (allowed foundation edge; package doc drift) | R2 (doc, decision recorded) |
| `schemas`, `jsonschema` | K1 | `mustAny`/`formatErrs` copied across `domain`, `adapter`, `model`; embedded schemas recompiled per call (perf only, left) | R2 |
| `randutil`, `canonical` | K1 | Unused `Picker` (map-order sum); device constant in a foundation package | R2, R9 |
| `world` | K2 | Dead API and exported-but-unread fields; 11-positional-param call recorder; three mixed files | R5 |
| `perturb` | K2 | Switch ladder over one name string; unused params everywhere; exported unused `Active` | R6 |
| `truth` | K2 | Pure solver and stateful sealed store in one file; store gate optional at construction; `SetupCall` record owned by `truth` but used by `audit`/`suite` | R7 |
| `audit`, `suite`, `score` | K2 | Duplicate `Perturbation` record, own `renderID`, hard-coded perturbation names/reasons, `score` imports `run` only to read a plain data bundle | R8 |
| `refconsumer` | K2 | Identity literal ×3; unused exports | R8 |
| `device` | K3 | Five responsibilities in one package (codec/contract, capability catalog, state machine, fault schedule, transport); bare-string codes; `map[string]any` records | R9, R14 |
| `deviceworld` | K2 | Strict-decode copy; duplicate accessor | R10 |
| `sink` | K3 | Three sinks in one file; dead branch; false doc | R10 |
| `run` | K3 | Orchestration plus ledger file I/O, artifact publish/load, quiescence timer, replay decoding, grab-bag accessor files; ~2× exported surface actually used; direct `time.Now` | R11 |
| `mcp` | K4 | Handlers carry run-lifecycle rules, scoring text and audit defaults; most exported `Director` surface used only by tests | R12 |
| `cli` | K4 | Manifest policy, artifact layout knowledge and suite sampling constants in the composition root | R13 |
| `wall` | K1 | Declared clock seam with zero importers; guards nothing | decision P-01 |
| repo | – | No layer map, no purity/I-O gates, no per-package coverage floor, lint lacks complexity/dupl | R1, R15 |

## Packages that stay one package

`world`, `perturb`, `deviceworld`, `suite`, `score`, `audit`, `refconsumer`,
`canonical`, `jsonschema`, `randutil`, `schemas`: they share private state or
have no I/O edge, so a facade/domain split would only add forwarding. They are
protected by the purity gates instead.
