# Ubiquitous language — suite

A suite is a generated, audited set of scenarios for one domain and profile:
the graded benchmark a consumer is scored against.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Suite | A generated graded set of scenarios with their sealed labels | `Suite` | `<domain>-<profile>-suite.json` |
| Scenario | One executable recipe: entity, fault, onset, duration, perturbations, setup, command log | `Scenario` | `scenarios[]` |
| Profile | A named scenario shape declared by the domain (nominal, correlated cascade, sensor pathology) | `Config.Profile` | `profile` |
| Negative class | A scenario whose fault expects no episode; its fraction is declared by the domain | `negative_class_fraction` | `is_negative_class` |
| Pre-degraded | A scenario that starts with the entity already degraded | `PreDegraded` | `pre_degraded` |
| Perturbation | A delivery corruption applied to the scenario; the same record the audit applies | `Perturbation` (alias of `audit.Perturbation`) | `perturbations[]` |
| Trivial case | A scenario the audit rejected; kept as a regression fixture, never counted | `TrivialCase` | `trivial_excluded[]` |
| Terminal state | The reported reason a domain could not produce the requested admitted scenarios | `TerminalState` | `terminal_state` |
| Attempt | One draw of the generate-audit-regenerate loop | `Attempts` | `attempts` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `suite.Perturbation` (own struct) | alias of `audit.Perturbation` | identical record; the audit now receives the scenario's slice unchanged |
| `suite.renderID` | `world.RenderID` | identical output (nil params), one implementation |
| `suite.Perturbation` JSON tags | owned by `audit.Perturbation` | the scenario artifact's `perturbations[]` shape follows the audit record |
