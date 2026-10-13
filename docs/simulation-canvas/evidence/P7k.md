# P7k evidence — catalogs

The built-in library (`internal/component/internal/files/builtin`): 20 equipment components (cooling, air, fluids, drives,
heat, electrical, process loads, weather and a site calendar) and 17 instruments, authored to
[catalog-spec.md](../catalog-spec.md). One table-driven harness over the whole library, `test/contract/catalog_test.go`,
runs every component in every preset (small, medium, large), composed alone, with its optional instruments.

| Check (per component and preset) | Harness step |
|---|---|
| Composes, validates and lints a nominal day with no error | `lint` |
| Starts at rest: no state moves more than 3 % in its first ten minutes | `steady_start` |
| A two-hour run replays byte for byte | `replay` |
| Every effector moves the states it names | `effectors` |
| Every fault is observable above its first SNR by the truth solver | `faults_observable` |
| The trivial-baseline audit runs, and at least one own fault is not trivial | `trivial_baseline_audit` |

Library-wide: every file follows its contract, every built-in id is in the `streamsim/` namespace, presets are small,
medium and large with medium by default, every instrument attaches to a matching point, every port type has both ends
and every providing port connects to every consuming port of its type (`catalog_library_test.go`,
`components_test.go`). Feedback loops between kinds close through sampled reads (R7k-01, domain-spec v0.2 §28).

KPI truth per component, listed in P7k's plan row, landed in P7l (domain-spec v0.2 §29; [P7l evidence](P7l.md)).
