# Survey C: domain, adapter, model, jsonschema, canonical, schemas, randutil, wall

Read-only. Paths relative to internal/. "Certain" = verified from code; "likely" = inferred.
Naming trap: in this repo `domain` = simulated-world spec and `adapter` = pure declarative projection engine. Neither is the standard's "pure rules" / "I/O edge adapter". Keep the names (ubiquitous language in AGENTS.md, docs, digests) and do not rename.

## 1. Per-package summary

| pkg | Owns today | Should own | Used outside (non-test) |
|---|---|---|---|
| domain | Load file/dir, JSON-schema validate, decode, cross-check, digest, catalog/coverage | Pure: bytes -> Compiled (validate, compile, digest) + Catalog. No os. | Compiled(38), Catalog, Load(2), LoadAll(1), Parse(1: run/replay_inputs.go:15), NewCatalog |
| adapter | Load adapter, validate, render engine, AND conformance verify + fixture | Load/validate + render engine. Conformance is a separate responsibility. | Load, LoadBytes, NewEngine, Engine, Verify (cli/catalog.go:197 only) |
| model | Records + constants + time fmt + strict decode + schema validators | Records, constants, time helpers, strict decode. No schema engine. | broad; ValidateVerdict/ValidateRunArtifact only run/verdict.go:67, run/artifact_decode.go:17 |
| jsonschema | Subset JSON-Schema compile/validate | Same | Compile, Schema, Error (world, adapter, domain, model, device) |
| canonical | RFC 8785 marshal, digests | Same | Marshal, MarshalString, Digest, DigestBytes, DigestDomain, CapabilityCatalogDomain |
| schemas | go:embed of 6 contract schemas | Same plus compiled-schema access | DomainSpec, OutputAdapter, ConsumerVerdict, RunArtifact, GroundTruth |
| randutil | SplitMix64, Substream, Fnv1a64, Picker | Same minus Picker | SplitMix64, Substream, Fnv1a64, NewSplitMix64 |
| wall | build-tag clock seam | n/a: nobody imports it | none (see 6.2) |

Over-exposed or dead (production callers = none):
- domain: `Compiled.FaultNames/EffectorNames/ProfileNames` (domain.go:93-100, no callers at all); `HasProfile` (tests only).
- adapter: `Engine.Meta()` (engine.go:104, no callers); `RenderRun` (callers: tests + own Verify); `NewEngine` never returns a non-nil error (engine.go:31-42) yet every caller handles one.
- model: `TimeNS` (model.go:61), `SinkBroker` (run.go:89; AGENTS forbids a broker), `TimeScaled` (run.go:95; mcp/schemas_values.go:63 hardcodes "scaled"), `AdmissionAccepted`, `SolveAnalytic` (truth.go:54; verify the literal "analytic" is not used instead).
- schemas: `SimEvent()` has no production caller (only the byte-identity test).
- randutil: `Picker`/`NewPicker`/`Pick` (randutil.go:97-149) used only by own tests. suite/scenario_draw.go:66-75 has its own sorted weighted draw.
- canonical: `CapabilityCatalogDomain` (canonical.go:44) is a device-contract constant with one device caller. Wrong owner; a paired-system name sits in a foundation package.

## 2. I/O, clock, env, global state

| Site | Function | Note |
|---|---|---|
| domain/domain.go:19 `os.ReadFile` | `Load` | edge I/O in the "pure" package |
| domain/domain.go:40 `os.ReadDir` | `LoadAll` (+ loading.go:57-78 `domainPaths`/`loadPaths`) | cli already has an unsorted-by-us twin for adapters (cli/catalog.go:36,60) |
| adapter/adapter.go:23 | `Load` | |
| adapter/fixture.go:50 | `decodeJSONL` | verification-only |
| adapter/verify_conformance.go:14,59 | `loadOutputSchema`, `compareGolden` | verification-only |
| adapter/engine.go:13, domain templateFieldRE | package-level compiled regexp | immutable, safe |
| schemas/*.go, adapter/fixture.go:18 | go:embed vars | immutable |
| wall/*.go | `Now` | no importers |
| model/run.go:108 | `CurrentPlatform` reads `runtime.*` | harmless, not I/O |
No env reads, no time.Now in these 8 packages. No mutable package globals. Every `jsonschema.Compile(mustAny(...))` recompiles the embedded schema on each call (domain/loading.go:22, adapter/adapter.go:60, model/model.go:70,87); only device/contract.go caches.

## 3. Mixed responsibilities

- domain/loading.go mixes four things: schema validation (17-30), decode+digest+compile orchestration (32-55), directory scanning (57-78), spec decode (80-85). `formatErrs`/`mustAny` sit in dynamics.go:64-84, unrelated to dynamics.
- domain `Compiled` is Spec pointer + digest + raw + 5 name sets + gains; lookups are linear scans (domain.go:103-140) next to maps. `Spec`/`Raw` exported and mutable by callers.
- adapter: verify.go/verify_conformance.go/verify_inputs.go/fixture.go (~320 prod lines) are a conformance harness inside the production package; they hold the only os/jsonschema-output code besides Load. Real CLI feature (`adapter verify`), so it is a use case, not test code: it deserves its own package.
- adapter wire-number formatting borrows canonical's digest serializer (encoding.go:53,90,119 `canonical.MarshalString`): a canonical change silently alters adapter golden output. Coupling is by accident.
- model is not "records only": model.go:65-95 pulls in jsonschema+schemas to validate verdict/artifact (contradicts architecture.md "transport concerns leaking into model", and its own package doc which says validation lives in domain/adapter). model.Adapter carries loader state `Raw` (adapter.go:20), mirrored by `Compiled.Raw`. `GroundTruthRecord.IsPositive()` (truth.go:60) is the only behaviour on a record (fine; doc comment starts "Epithet", not the name).
- adapter digest is `canonical.DigestBytes(json.Marshal(*model.Adapter))` (run/identity.go:10-12): digest of a Go struct encoding, not RFC 8785 of the document as for domains (domain/loading.go:46). Field order and `omitempty` tags in model/adapter.go are digest-bearing.

## 4. Vocabulary / typing drift

- Closed sets as raw strings, repeated at 2-3 sites each (validation, evaluation, render, schema enum):
  - ValueExpr.Op: validation.go:47-55,182-194, expression.go:35-45,139-153.
  - time layouts: validation.go:88 and expression.go:165-175 (two copies of the list).
  - encoding: engine.go:55,85, render.go:89-95.
  - IDRewrite.OnViolation: engine.go:123-129.
  - F1 form: validation_dynamics.go:16-25, validation_effectors.go:52 (`"dead_time"`).
  - cadence mode: validation_records.go:26,33; noise model :36,44; tier "F2"/"F3" :75,78.
  - Only Detector forms and a few Delivery*/Op* are constants, and those are untyped consts too.
- `map[string]any` is legitimate for effector args_schema/params/run meta; `Engine.meta map[string]any` carries a fixed 7-key set (validation.go:117 `runMetaKey`): a small struct would do, but only if RunMeta callers in run/initialize.go:123 move too (low value).
- Names: `Catalog.List(group)` filters by id prefix (catalog.go:36-45) and `hasPrefix` re-implements `strings.HasPrefix` (86-88). `jsonEqualish` (adapter/encoding.go:52) duplicates `jsonschema.jsonEqual` (value.go:144). `checkDynamicsGraph` (validation.go:48-53) wraps a function that already returns error. `rejectDynamicsCycles(spec, c, src)` takes spec and c where c.Spec == spec.
- Three copies of `mustAny` (domain, adapter, model) and two of `formatErrs` (domain, adapter) are byte-identical.
- Error prefix stutter (strings are contract-ish; do not change casually): `streamsim: streamsim: domain: ...` from domain.go:148 + loading.go:52 (+ loading.go:73 in LoadAll, + validation_records.go:87,131); `adapter: adapter: ...` in adapter.go:45 + validation.go:41,62,161; `canonical: canonical: canonical:` from canonical.go:31 + encoding.go:44 + number.go.
- adapter.go:27,33,65: the `separator` parameter leaks a presentation difference ("\n  " vs " ") between Load and LoadBytes into the loader.

## 5. Probable bugs / hazards, ranked

1. (certain, high) adapter `verify` panics on an empty fixture file: verify_inputs.go:29 `fixture[len(fixture)-1]`, :40 `fixture[0]`; `validateStrictObservedOrder` passes on empty (verify.go:66-79). User-supplied path via `--fixture`.
2. (certain, medium) `hash_suffix` rewrite: engine.go:127-128. `canonical.DigestBytes` returns "sha256:<hex>", so `[:16]` is `"sha256:"` + 9 hex chars: a colon lands in an id whose alphabet is being narrowed. Also `out[:rw.MaxLength-len(h)]` panics when 1 <= max_length < 16 (schema minimum is 1, output-adapter schema `max_length`). `truncate` slices bytes, can split UTF-8. No test covers `hash_suffix` and no shipped adapter uses it.
3. (certain, medium) `adapter verify` semantic holes: verifyOutputSchema returns `true` without setting `res.SchemaOK` when no output_schema is declared (verify.go:39-42 vs verify_conformance.go:40); same for GoldenMatch. The CLI requires both (cli/catalog.go:198) and prints "FAILED:" with empty divergence. An adapter declaring only `golden` can never pass. `splitRecords(out, encoding)` ignores `encoding` (verify.go:102-113): a `json-array` adapter with `output_schema` fails on the "[" / ",{...}" lines.
4. (certain, medium) `randutil.Picker` sums `total` in map iteration order before sorting (randutil.go:131-138; comment at 130 claims determinism-safe). Float sum is order-dependent, so `total` can differ by ulps between runs. Dead today, but a trap if used. Fix by deleting it (suite already does the sorted version).
5. (certain, low-medium) jsonschema.Compile panics on `"uniqueItems": "x"` (compile_keywords.go:143 unchecked `v.(bool)`), and `toInt` silently returns 0 for non-integers (compile_reference.go:17-38). Compile is reachable from user data: adapter `output_schema` file (verify_conformance.go:26) and effector `args_schema` (world/effector_policy.go:83).
6. (certain, low) F1Input round-trip: `Coef float64 omitempty` + `CoefSet json:"-"` (model/domain.go:111-115) means explicit `"coef":0` marshals without coef, and re-reading gives 1. mcp/director_resources.go:25 marshals `compiled.Spec`, so the describe resource misreports an explicit zero coefficient.
7. (certain, low) zero-vs-unset: `observation_gain: 0` becomes 1 (domain/compilation.go:85-90); State.Min/Max, Range, etc. are `omitempty` floats, so 0 == unset. Behavioural contract: report, do not "fix" in a refactor.
8. (certain, low) nondeterministic *which* error: map iteration in validation_records.go:144 (`FaultWeights`), :95 (`F2.Inputs`), and jsonschema compile.go:44 (`$defs`). Pass/fail identical; message can differ when several defects exist.
9. (likely, low) canonical integer literals pass through verbatim (number.go:29-38), so integers beyond 2^53 are not RFC 8785 (ES Number) form. Intentional-looking, digest-bearing: pin, do not touch.
10. (low) `evalCounter` default width 6 (expression.go:110-113) is dead: schema requires 1..20 only when present, but checkCounter rejects 0 (validation.go:201). Omitted width therefore fails load while the engine pretends a default exists.
11. (low) jsonschema `multipleOf` uses `q != float64(int64(q))` (validate_primitive.go:57-63): 0.3 multipleOf 0.1 fails. No shipped schema uses multipleOf; ok for now.
12. (low) `Engine.End` mutates the caller's meta map (engine.go:80; `NewEngine` aliases it at :35). Catalog.List returns nil, so JSON `null` for no match (catalog.go:37): shape fixed, do not touch.
13. (low) `registerDefinition` (jsonschema/compile.go:97-104) unreachable: registry is pre-filled with every $defs name.

## 6. Cross-cutting findings

6.1 `internal/wall` is dormant: zero importers (known: docs/reviews/ARCHITECTURE_REVIEW.md:314). Direct `time.Now` at run/artifact_metadata.go:34, run/verdict.go:39, cli/arguments.go:36, cli/device_world.go:81, cli/manifest_options.go:43. The simdet tag therefore proves nothing about those sites.
6.2 Routing those through wall changes behaviour under simdet (Now() returns the zero time): `created_at`/`unblinded_at` become 0001-01-01 in simdet test output. Not a behaviour-neutral refactor; needs an explicit decision (inject a clock func at cli/run edge, or delete the package and amend docs/DECISIONS.md D-13).

## 7. Target shape and rounds

Target (no new speculative packages):
- domain: pure `Parse(raw, src)`, `Compiled`, `Catalog`. File reading moves to cli edge (cli already owns adapter dir scan, cli/catalog.go:36-60, so this removes the asymmetry).
- adapter: pure `LoadBytes` + `Engine`. `Load(path)` stays as thin os.ReadFile+LoadBytes (6 callers/tests); conformance moves to `adapter/conformance` (concrete caller: cli/catalog.go:197).
- model: records, constants, time, strict decode only.
- schemas: embed + one cached compile helper per contract (schemas -> jsonschema, downward).

Answers to the two asked decisions:
- Split file loading from compile/validate? Yes in both. Domain: split by file first (loading_fs.go holding the only `os` import, with an architecture test asserting that), then relocate `Load`/`LoadAll` to cli. Adapter: already split in shape (`LoadBytes` is pure); keep `Load` as the one 3-line reader and move everything else that reads files (golden, output schema, fixture file) to conformance.
- Move verify out of the production path? Move into its own package `internal/adapter/conformance` (not out of the binary, since `adapter verify` is a shipped command). Keep `FixtureEvents()` (embedded, pure) in adapter so in-package tests (adapter_test.go:68, golden_test.go:20, lifecycle_test.go:13) do not form an import cycle.

Ordered rounds (each: focused tests, then `make ci-check`):
1. Pin before moving: golden-digest test for every shipped domain/adapter (domain digest + `adapterDigest`), error-string snapshot tests for Parse/LoadBytes failures, `adapter verify` outputs for shipped adapters. Hazard: none; this is the safety net.
2. Dedupe + dead surface (zero behaviour): one `mustAny`/`formatErrs` (put `FormatErrors` in jsonschema, compiled-schema accessors with sync.Once in schemas); delete `FaultNames/EffectorNames/ProfileNames`, `Engine.Meta`, `TimeNS`, `SinkBroker`, `Picker`, `SimEvent()` (if unused), `hasPrefix`. Hazards: error text must stay byte-identical; shared compiled *Schema is safe because Validate is read-only. Proof: round-1 snapshots + existing schemas/domain/adapter tests.
3. model purity: move ValidateVerdict/ValidateRunArtifact next to their two run callers (or schemas helper); drop jsonschema/schemas from model, update test/architecture allowlist (`internal/model`). Hazard: error strings "model: verdict ..." are contract; keep verbatim. Proof: run/artifact_test.go:74, fuzz_test.go:31, refconsumer_test.go:126.
4. domain edge split: loading_fs.go, then move Load/LoadAll to cli (16 test call sites: add a test helper). Hazards: the `streamsim:` wrap counts; `Compiled.Raw` copy; sorted `domainPaths` order; LoadAll fails the whole load. Proof: shipped_test.go + replay (run) tests.
5. adapter conformance package: move verify*.go (+ `loadFixture` file branch). Add the empty-fixture guard here (bug 5.1, an intentional correction: needs its own regression test, report separately per AGENTS). Hazard: VerifyResult JSON keys; cli/catalog.go:198 pass/fail semantics (decision on 5.3 is a behaviour change; do separately). Proof: conformance_test.go + golden_test.go.
6. Typed closed sets (named string types: DetectorForm, TransformOp, TimeLayout, Encoding, F1Form, CadenceMode): one validation+evaluation switch per type. Hazard: JSON/digest unchanged only because named string types marshal identically; adapter digest is struct-marshal, so field order/tags in model/adapter.go must not change. Proof: round-1 digest pins, validation_steps_test.go.
7. Smaller correctness fixes, each its own commit with regression test: bug 5.2 (hash_suffix length guard; the "sha256:" prefix is an output change, decide separately), jsonschema compile type checks (5.5), F1Input marshal (5.6).
8. Decision items needing a design note, not refactor: wall seam (6.2), adapter digest canonicalisation (changes every adapter digest, breaks replay artifacts), observation_gain 0 semantics, `CapabilityCatalogDomain` move to device (safe but touches device).
