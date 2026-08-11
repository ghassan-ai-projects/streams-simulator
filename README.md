# Streams Simulator

> **Codename:** `streamsim` · **Status:** design, ready to build · **Date:** 2026-08-11

A standalone, deterministic, closed-loop world simulator for testing stream processors.

It generates realistic event streams from data-defined domains, corrupts their delivery in specified ways, carries sealed ground truth, accepts commands back through declared effectors so the world actually changes, and scores whatever consumed it.

**It is an independent product with no knowledge of its consumers.** No consumer's name, schema, field, or behaviour appears in the binary. A consumer is a file — an [output adapter](docs/contracts/output-adapter-v0.1.schema.json) — plus, optionally, an MCP client that invokes effectors and submits a verdict.

It is a **test instrument**. An instrument less trustworthy than the system it measures is worse than no instrument, because it produces confident wrong answers. Determinism, generative ground truth, and analytic oracles are therefore non-negotiable.

## The three decisions that shape everything

1. **MCP manages the simulator; it does not transmit the streams.** Control, actuation, and audit are request/response over MCP; evidence leaves on a sink. A dropped message is a blessed outcome on a push channel, which would make every absence finding permanently ambiguous.
2. **The simulator emits its own format; consumers are adapters.** Events are `sim-event-v0.1`, projected into a consumer's wire format by a declarative adapter. Adapters are data, loaded through one path, exactly like domain specs.
3. **One MCP server, two roles, and no domain knowledge in either.** `director` builds and drives worlds and holds the truth; `operator` is what a consumer gets: read the nameplate, list effectors, invoke one, submit a verdict. Four tools.

## Shape

```text
        ┌──────────────── MCP: director role ─────────────────┐
 harness│ catalog · world · clock · fault · perturb · truth    │
        └──────────────────────┬──────────────────────────────┘
                               │ command log
   domain spec ───────────────►│
   (data)                      ▼
                      ┌──────────────────┐
                      │   world core     │  seeded · discrete-event
                      └───┬──────────┬───┘
        native events     │          │  effects, with time constants
                          ▼          │
                  ┌───────────────┐  │
                  │ perturbation  │  │
                  └───────┬───────┘  │
                          ▼          │
    adapter ─────►┌───────────────┐  │
    (data)        │    adapter    │  │
                  └───────┬───────┘  │
                          ▼          │
                inproc · file · http-push
                          ▼          │
                  ╔═══════════════╗  │
                  ║   consumer    ║  │
                  ╚═══════╤═══════╝  │
                          └──────────┘
                    MCP: operator role
            effector.invoke · consumer.report
```

## Non-negotiables

Nine controls separate a test instrument from a trace generator (see [docs/design/TECHNICAL_DESIGN.md](docs/design/TECHNICAL_DESIGN.md)):

1. Determinism — a run is a pure function of `(sim_version, domain_digest, adapter_digest, seed, command_log, sink)`
2. The analytic cross-check — implement the integrator twice, assert agreement
3. The reference consumer — the product must be complete on its own
4. The delivery ledger — transport misses vs reasoning misses
5. The quiescence barrier — closed-loop reproducibility
6. A sealed oracle — compile-time role separation + differential prefix-indistinguishability test
7. The injection probe — adversarial text changes no consumer conclusion
8. The `silent_no_effect` test — rate zero
9. The trivial-baseline audit — no scenario a one-line detector solves enters the graded suite

## Documents

The full specification lives in [docs/](docs/README.md):

| Document | Contents |
|---|---|
| [docs/research/TRANSPORT_ANALYSIS.md](docs/research/TRANSPORT_ANALYSIS.md) | Why MCP carries control but not evidence; sinks, receipt timestamps, role separation |
| [docs/research/PRIOR_ART.md](docs/research/PRIOR_ART.md) | Deterministic simulation testing, FMI 3.0, Sparkplug B, ISO 13374, ISA-18.2 / EEMUA 191, the anomaly-benchmark critique |
| [docs/design/DOMAIN_CATALOG.md](docs/design/DOMAIN_CATALOG.md) | Twelve property axes, 25 domains, coverage proof, build order |
| [docs/design/TECHNICAL_DESIGN.md](docs/design/TECHNICAL_DESIGN.md) | Determinism, injection surfaces, channels and dynamics tiers, sinks and adapters, the closed loop, security |
| [docs/design/MCP_SURFACE.md](docs/design/MCP_SURFACE.md) | One server, two roles, four operator tools, error taxonomy |
| [docs/design/CONSUMERS.md](docs/design/CONSUMERS.md) | The integration contract; the reference consumer |
| [docs/design/GROUND_TRUTH_AND_SCORING.md](docs/design/GROUND_TRUTH_AND_SCORING.md) | Four benchmark flaws and countermeasures, three onset timestamps, leak detection |
| [docs/design/IMPLEMENTATION_PLAN.md](docs/design/IMPLEMENTATION_PLAN.md) | Seven stages, ~58 days, gates, cut list, stop/go |
| [docs/design/GAP_ANALYSIS.md](docs/design/GAP_ANALYSIS.md) · [docs/design/CRITICAL_REVIEW.md](docs/design/CRITICAL_REVIEW.md) | Twelve gaps; eight adversarial findings |
| [docs/contracts/](docs/contracts/) | `sim-event`, `output-adapter`, `consumer-verdict`, `domain-spec`, `run-artifact`, `ground-truth` |
| [docs/adapters/](docs/adapters/) | One adapter per consumer. Deleting one changes no simulator behaviour. |
| [docs/examples/](docs/examples/) | `aquaculture-pond`, the closed-loop showcase domain |

## Technology

Go 1.26, one binary, SQLite WAL, RFC 8785 canonical JSON, the official MCP Go SDK. No physics engine, no broker, no ORM, no expression language, no plugin system. The statistics and reporting layer may be Python, post-hoc and outside the determinism boundary.

Start here: [docs/research/TRANSPORT_ANALYSIS.md](docs/research/TRANSPORT_ANALYSIS.md), then [docs/design/DOMAIN_CATALOG.md](docs/design/DOMAIN_CATALOG.md), then both reviews, then the plan. Begin with S0.

## Development

```bash
make ci-check        # tidy + build + vet + lint + test-short + deadcode + vulncheck
make test            # race + shuffle + coverage
make test-coverage   # coverage HTML report
make lint            # golangci-lint
make cross-compile   # linux/amd64 binary
```

For coding agents: read [AGENTS.md](AGENTS.md) before editing.

## License

MIT - see [LICENSE](LICENSE).
