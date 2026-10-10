# Ubiquitous language — domain

A *domain* here is the data file that defines a whole simulated world:
entities, hidden state, dynamics, channels, faults, effectors and profiles.
The binary contains no domain behaviour; a domain that needs a code branch is
a bug in the simulator.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Domain spec | The declarative JSON document for one world | `model.DomainSpec` | `*.domain.json`, domain-spec-v0.1 |
| Compiled domain | A spec that passed schema validation and cross-checks, with its canonical digest and resolved defaults | `Compiled` | `domain_digest`, `domain_spec` in the run artifact |
| Digest | RFC 8785 canonical-JSON hash of the validated document | `Compiled.Digest` | `domain_digest` |
| Raw | The validated source document embedded in run artifacts | `Compiled.Raw` | `domain_spec` |
| State | A hidden physical quantity the world evolves | `HasState`, `StateNames` | state names |
| Channel | An observable stream derived from state with noise, cadence and absence | `Channel`, `ChannelNames` | channel names |
| Fault | A declared physical fault with onset, effects and observability | `Fault` | `fault_id` |
| Effector | A declared actuation with ack, failure modes and effect | `Effector` | `effector` |
| Profile | A named scenario shape the domain supports | `Profile` | `profile` |
| Catalog | The installed set of compiled domains, listed, described and checked for axis coverage | `Catalog` | `catalog list/describe/coverage` |
| Axis | A dimension a domain must declare so the catalog's coverage is measurable | `CoverageReport.ByAxis` | `axes` |
| Loading | Reading a spec from a file or directory; the only I/O in the module | `Load`, `LoadAll` (files edge) | – |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `Compiled.FaultNames`, `EffectorNames`, `ProfileNames` | removed | no caller (R3) |
| `domain.Load` inside the rules package | `files.Load` behind the facade | file reading is an edge |
