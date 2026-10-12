# Ubiquitous language — domain

A *domain* here is the data file that defines a whole simulated world:
entities, hidden state, dynamics, channels, faults, effectors and profiles.
The binary contains no domain behaviour; a domain that needs a code branch is
a bug in the simulator.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Domain spec | The declarative JSON document for one world: one or more kinds and instance sets. The validated document is also the runtime form every module reads | `model.DomainSpec`, `Compiled.Spec` | `*.domain.json`, domain-spec-v0.2, `catalog describe` spec |
| Kind | The template of an entity: its state, dynamics, channels, faults and effectors; its name is the `entity_type` of every instance. A domain may declare several; its references resolve inside the kind | `model.Kind`, `Compiled.Kind` | `kinds[]` |
| Instance set | A templated population of one kind: id template, count bounds, churn. A kind may have several; explicit entity ids a run is given must fit one of them | `model.InstanceSet`, `Compiled.InstanceSetFor`, `Compiled.CheckEntityIDs` | `instances[]` |
| Domain-wide name | A channel name, fault id or effector name: unique across all kinds, so it resolves to one declaration (state names are scoped to their kind) | `Compiled.Channel`, `Fault`, `Effector`, `FaultKind` | `channel`, `fault_id`, `effector` |
| Canonical formatting | The one layout of a domain file (member order, line-fitting objects) so a hand edit and a tool edit produce the same bytes; never changes the digest | `Format` | `domain format` |
| Compiled domain | A spec that passed schema validation and cross-checks, with its canonical digest and resolved defaults | `Compiled` | `domain_digest`, `domain_spec` in the run artifact |
| Digest | RFC 8785 canonical-JSON hash of the validated v0.2 document | `Compiled.Digest` | `domain_digest` |
| Validation error | A domain that fails its schema or a cross-reference check; every issue carries a JSON pointer into the v0.2 document, except that a document that is not JSON has none | `model.ValidationError` (contract `domain-spec-v0.2`) | CLI `domain_invalid` |
| Unknown domain | A catalog lookup for an id the catalog does not hold | `ErrUnknownDomain` | CLI `domain_not_found` |
| Raw | The validated source document embedded in run artifacts | `Compiled.Raw` | `domain_spec` |
| State | A hidden physical quantity the world evolves; named within its kind | `Kind.StateDecl`, `StateNames(kind)` | state names |
| Channel | An observable stream derived from state with noise, cadence and absence | `Channel`, `ChannelNames`, `KindChannelNames` | channel names |
| Fault | A declared physical fault with onset, effects and observability | `Fault` | `fault_id` |
| Effector | A declared actuation with ack, failure modes and effect | `Effector` | `effector` |
| Profile | A named scenario shape the domain supports | `Profile` | `profile` |
| Catalog | The installed set of compiled domains, listed, described and checked for axis coverage | `Catalog` | `catalog list/describe/coverage` |
| Axis | A dimension a domain must declare so the catalog's coverage is measurable | `CoverageReport.ByAxis` | `axes` |
| Loading and formatting | Reading a spec from a file or directory, and rewriting a file in canonical formatting; the only I/O in the module | `Load`, `LoadAll`, `Format` (files edge) | – |
| Catalog from a directory | Every domain in a directory, as a catalog; an empty directory name selects `domains`, and a directory with no domain is refused because its coverage report would claim nothing | `NewCatalogFromDir`, `files.LoadCatalog` | `--domains-dir` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `Compiled.FaultNames`, `EffectorNames`, `ProfileNames` | removed | no caller (R3) |
| `domain.Load` inside the rules package | `files.Load` behind the facade | file reading is an edge |
