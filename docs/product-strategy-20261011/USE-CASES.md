# Use cases: what works now, what needs an upgrade

- **Date**: 2026-10-11
- **Method**: a hands-on end-to-end CLI check (binary built from `main`, every command run and
  exit code recorded), four parallel research passes (agent benchmarks, stream-testing practice,
  standards, capability check), and my own verification of the two defects below.
- **Confidence labels**: **VERIFIED** = observed by running it or read on a primary source;
  **REPORTED** = from a research pass, primary source not re-read by me; **UNVERIFIED** =
  secondary or absent evidence. Nothing here is legal advice.

## Two findings from the hands-on check

Both were reproduced by me, not only reported.

| # | Finding | Evidence | Consequence |
| --- | --- | --- | --- |
| F1 | **The reference consumer cannot detect a persistent step at its default settings.** | A +22 A step on `pump-1` (about 70 sigma) gave 0 detections at the default threshold 4, 2 at 3.5, 2 at 3, 6 at 2. Cause: each anomalous reading is appended to the 30-reading baseline, so the third consecutive reading can reach at most about 3.74 sigma, below 4. The unit test sets `Threshold = 3` ("the crash is loud"), which hides it. | The shipped reference consumer is a weak instrument. Any scorecard that uses it at defaults reports "no detection" for loud faults, which is the confident-wrong-answer failure the project exists to prevent. |
| F2 | **`score` grades a run against a label that did not come from it.** | A label for `rotating-machinery/0000` (`pump-6`, `transient_none`) was accepted for a run that injected `overload` on `pump-1`; a full scorecard was produced with no error. | A mismatched run and label yield a plausible but meaningless scorecard. Run artifact and label need a binding (scenario id and seed check, or a label digest in the artifact). |

Other friction recorded (REPORTED): `suite` takes about 90 s for 100 scenarios with no
progress output and exits 0 while `terminal_state` reports unmet profile coverage; there is no
command that runs a suite's scenarios; flags after the positional argument are silently
ignored by `verify`; error text repeats the `streamsim:` prefix.

## Three use cases that work today

### N1 — Reproducible chaos workload for a pipeline's CI (VERIFIED)

**Job**: give a stream-processing team a seeded, replayable input with known delivery
corruption, so a regression in late, duplicate or reordered handling fails a build.

**What works**: `run --domain … --seed … --sink file` produced a 171-event run in under 0.1 s
and a one-hour faulted run (5,277 events) in 0.15 s, with `trace.jsonl`, `ledger.jsonl`, and a
`run.json` whose digest `replay` and `verify` reproduce. A tampered seed makes `verify` exit 1
("world digest mismatch"). The ledger states what was delivered, dropped, duplicated, delayed.
The `agentic-stream` and `native-jsonl` adapters pass `adapter verify` against their goldens.

**Why it has demand** (REPORTED, primary pages fetched):

- Kafka Streams' `TopologyTestDriver` does not emulate brokers, faults, duplicates or ordering,
  and Confluent says it "doesn't fully reflect all the aspects of real Kafka cluster behavior".
- Flink's testing page does not discuss out-of-order or late data, and FLIP-554 states
  harnesses are "insufficient for testing full pipelines".
- Jepsen found acknowledged-message loss in Redpanda 21.10.1 and write loss in Bufstream
  0.1.0, so delivery guarantees are an active failure area.

**Limits today**: the team compares its own pipeline output to the ledger by hand; the CLI
does not score pipeline output, and only 2 adapters exist.

### N2 — Offline scoring of a detector against sealed generative truth (VERIFIED, with F1 and F2)

**Job**: measure a rule set or model on worlds where the true onset, first-observable time
and unavoidable time are known, including a declared negative class.

**What works**: `suite` generates 100 audited scenarios with labels (40 % negative class in the
shipped profile, trivial-baseline verdict on each); `run --fault` injects a named fault;
`refconsumer` produces a verdict; `score` emits a scorecard with instrument, consumer, loop
and judgment sections. The three-timestamp truth is the differentiator against annotated
benchmarks: none of the nine surveyed benchmarks uses an independent analytic oracle for
detection timing (REPORTED).

**Limits today**: no command runs a suite's scenarios, so a user scripts run→refconsumer→score
by hand; F2 allows a mismatched label; F1 makes the shipped baseline blind to steps. Treat
results from this path as indicative until those are fixed.

### N3 — Closed-loop sandbox for an agent's action plane (implemented; not exercised in my check)

**Job**: test whether an agent runtime retries safely, respects idempotency windows, and does
not report success on a command that had no effect.

**What exists** (read in the domain files and test suites): effectors declare ack latency and
failure modes (`ok`, `slow`, `ack_lost`, `reject`, `confirmed_no_effect`,
`silent_no_effect`), idempotency windows, dead time and time constants; the operator MCP role
is token-scoped; scoring has `false_success`, `silent_no_effect_calls` and deadline
adherence; closed-loop and acceptance tests are in the suite.

**Why it has demand** (REPORTED): AIOpsLab, ITBench, tau-bench and AgentDojo are closed-loop but
score task success or injection rate; none scores actions taken on corrupted or stale input.

**Limits today**: the operator loop was not run by the hands-on check; action cost and
unsafe-call counts are not numbers yet (see the economics design).

## Three use cases that need an upgrade

### U1 — A quantified decision-quality and cost benchmark for agents on operational streams

**Job**: show, with intervals, that an agent runtime beats a strong rules engine and an
LLM-on-every-event baseline on regret, unsafe actions and cost, on held-out seeds.

**White space** (REPORTED, nine primary sources): none of AIOpsLab, ITBench, OpenRCA,
OpenRCA 2.0, AssetOpsBench, PHMForge, PHM-Bench, ADMITBench, AgentDojo or tau-bench models late,
duplicate, out-of-order or dropped delivery; ITBench states runs are non-deterministic;
AssetOpsBench uses a stochastic LLM judge (reported κ = 0.44); OpenRCA 2.0 reports exact root
cause in 20.7 % of cases and only 61.5 % of correct services tied to a verified causal path.
A seeded, replayable, delivery-corrupting world with sealed truth is unoccupied.

**Needs**: outcome economics, baseline ladder, held-out protocol, the alarm-storm and
priced-probe capabilities. See [DESIGN-outcome-economics.md](DESIGN-outcome-economics.md),
[DESIGN-baseline-ladder.md](DESIGN-baseline-ladder.md) and [ROADMAP.md](ROADMAP.md) phases 2–5.
This is the strongest differentiation and should be built first.

### U2 — A regulatory test-evidence pack for AI acting on sensor streams

**Job**: produce an artifact a provider can attach to a conformity file: metrics fixed before
the run, a logged and replayable trace, and unsafe-action counts.

**What the sources support** (REPORTED, EUR-Lex and the Commission service desk read):

- Art. 9 requires testing against metrics and probabilistic thresholds defined in advance.
- Art. 12 requires automatic event logs over the system's lifetime.
- Art. 15 requires accuracy, robustness and cybersecurity, with accuracy metrics declared.
- Application dates under Regulation (EU) 2026/1744: Annex III systems 2 December 2027,
  Annex I systems 2 August 2028. Verify against the Official Journal before relying.
- **No verified primary text uses "reproducible" or "deterministic" or mandates an
  independent test instrument.** Position the product as *evidence for* these duties, not as a
  requirement of them.

**Needs**: a pre-registration step (metric and threshold set committed by digest before the
run); a fresh signed manifest with reviewer identity; the ISA-18.2/EEMUA flood KPI
(episodes per 10 minutes, already specified in `docs/design/GROUND_TRUTH_AND_SCORING.md` §4)
as a scorer field; unsafe-call counts from the economics block; a documented mapping table
reviewed by counsel. ISA-18.2, EEMUA 191 and IEC 61508 contents were UNVERIFIED in this
research and must be read from the standards themselves.

### U3 — A streaming-semantics conformance suite for stream engines

**Job**: a CI gate that checks a pipeline's event-time, lateness, watermark and
exactly-once behaviour against what the world actually did.

**Why it has demand** (REPORTED): the strongest pain evidence in the research is here (Flink
and Kafka Streams docs silent on late data, FLIP-554, Jepsen results).

**Why it is not ready**:

- All eight shipped domains declare `time_reference = wall`; event-time semantics are
  untested by design. This is the largest functional gap for this use case.
- There is no expected-output oracle for windowed aggregates; truth is fault-centred, not
  query-centred.
- No Kafka or OpenTelemetry sink or adapter exists; HTTP-push and file are the only
  transports, and "no broker" is a project rule.
- Cross-architecture float determinism is unverified.
- Competition is real: Antithesis (reported $105M Series A, customer names company-reported),
  RisingWave's Madsim, TigerBeetle's VOPR, Jepsen, Kafka Trogdor.

**Needs**: event-time domains, a query-truth oracle (the analytic solver extended to window
aggregates), a connector story that respects the no-broker rule, and a clear position:
*realistic worlds with sealed truth about what happened*, not a general fault-injection
platform.

## Ranking

| Use case | Evidence of demand | Differentiation | Readiness | Order |
| --- | --- | --- | --- | --- |
| N1 chaos workload for CI | Strong | Medium | Ready | Use now |
| N2 detector scoring | Medium | High | Ready after F1, F2 | Use now, fix first |
| N3 action-plane sandbox | Medium | High | Ready, unexercised | Use now, exercise it |
| U1 agent decision benchmark | Medium | **Highest** | Phases 2–5 | **Build first** |
| U2 regulatory evidence pack | Medium (dates and wording verified) | High | Needs U1 pieces | Second |
| U3 stream-engine conformance | **Strongest** | Medium (crowded) | Furthest | Third, only with event-time |

## Sources

Benchmarks: [AIOpsLab](https://arxiv.org/abs/2501.06706), [ITBench](https://arxiv.org/abs/2502.05352),
[OpenRCA](https://github.com/microsoft/OpenRCA), [OpenRCA 2.0](https://arxiv.org/abs/2606.27154),
[AssetOpsBench](https://arxiv.org/html/2506.03828v4), [PHMForge](https://arxiv.org/html/2604.01532),
[ADMITBench](https://arxiv.org/html/2608.03866), [AgentDojo](https://arxiv.org/html/2406.13352v3),
[tau-bench](https://arxiv.org/html/2406.12045).
Stream testing: [Kafka Streams testing](https://kafka.apache.org/42/streams/developer-guide/testing/),
[Confluent](https://confluent.io/fr-fr/blog/testing-kafka-streams),
[Flink testing](https://nightlies.apache.org/flink/flink-docs-stable/docs/dev/datastream/testing/),
[FLIP-554](https://cwiki.apache.org/confluence/display/FLINK/FLIP-554+Pipeline+Testing+Framework+for+Flink),
[Jepsen Redpanda](https://jepsen.io/analyses/redpanda-21.10.1),
[Jepsen Bufstream](https://jepsen.io/blog/2024-11-12-bufstream-0.1.0),
[Antithesis](https://antithesis.com/docs/resources/deterministic_simulation_testing/),
[RisingWave Madsim](https://risingwave.com/blog/deterministic-simulation-a-new-era-of-distributed-system-testing/),
[TigerBeetle](https://docs.tigerbeetle.com/concepts/safety/).
Regulation: [Art. 9](https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-9),
[Art. 12](https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-12),
[Art. 15](https://ai-act-service-desk.ec.europa.eu/en/ai-act/article-15),
[Reg. (EU) 2026/1744](https://eur-lex.europa.eu/eli/reg/2026/1744/oj/eng).
