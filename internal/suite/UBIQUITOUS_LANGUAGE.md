# Ubiquitous language — suite

A suite is a generated, audited set of scenarios for one domain and profile:
the graded benchmark a consumer is scored against.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Suite | A generated graded set of scenarios with their sealed labels | `Suite` | `<domain>-<profile>-suite.json` |
| Scenario | One executable recipe: entity, fault, onset, duration, perturbations, setup, command log. The entity is drawn from the entities of the fault's kind, and setup calls apply only when that kind declares the effector | `Scenario` | `scenarios[]` |
| Profile | A named scenario shape declared by the domain (nominal, correlated cascade, sensor pathology) | `Config.Profile` | `profile` |
| Negative class | A scenario whose fault expects no episode; its fraction is declared by the domain | `negative_class_fraction` | `is_negative_class` |
| Pre-degraded | A scenario that starts with the entity already degraded | `PreDegraded` | `pre_degraded` |
| Perturbation | A delivery corruption applied to the scenario; the same record the audit applies | `Perturbation` (alias of `audit.Perturbation`) | `perturbations[]` |
| Trivial case | A scenario the audit rejected; kept as a regression fixture, never counted | `TrivialCase` | `trivial_excluded[]` |
| Terminal state | The reported reason a domain could not produce the requested admitted scenarios | `TerminalState` | `terminal_state` |
| Attempt | One draw of the generate-audit-regenerate loop | `Attempts` | `attempts` |
| Publication | Generate the audited suite of a domain and profile and write it as `<domain>-<profile>-suite.json`; a profile the domain does not declare is refused first | `Publish`, `FileName`, `UnknownProfileError` | `suite --domain --profile --out` |
| Campaign | An ordered list of one-fault runs, each with a free phase label, generated from a campaign spec as a pure function of it; the consumer keeps its state across runs, the world starts fresh in each | `Campaign`, `CampaignSpec`, `BuildCampaign`, `CampaignSpecError` | `campaign.json` (`campaign-v0.1`), `suite campaign --spec --out` |
| Campaign phase | A stretch of a campaign with its domain, run count, negative share and positive-fault weights; the counts are exact (largest remainders), the order shuffled from the campaign seed | `CampaignPhase` | `phases[]` of the spec, `phase` of a run |
| Run plan | The `session-plan-v0.1` of one campaign run: the plant, the host socket and the one disturbance the session injects | `PlanFileName` | `run-NNNN.session-plan.json` |
| Sample period | The channel sampling period a published suite is generated with: two minutes | `DefaultSampleNS` | – |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `suite.Perturbation` (own struct) | alias of `audit.Perturbation` | identical record; the audit now receives the scenario's slice unchanged |
| `suite.renderID` | `world.RenderID` | identical output (nil params), one implementation |
| `suite.Perturbation` JSON tags | owned by `audit.Perturbation` | the scenario artifact's `perturbations[]` shape follows the audit record |
