# Prior Art — What to Steal, What to Avoid

Status: research input to the Streams Simulator design
Date: 2026-08-11
Method: literature and specification review across five fields that have each solved
part of this problem, plus a survey of MCP-driven simulators.

The simulator sits at the intersection of five mature traditions. None of them solves
the whole problem; each has solved one part better than we would on a first attempt.

| Field | The part it solved | What we take |
|---|---|---|
| Deterministic simulation testing | reproducing a distributed-systems bug from a seed | the entire determinism architecture (§1) |
| Co-simulation / digital twins | composing physical models across tools | fidelity tiers and the FMI escape hatch (§2) |
| Industrial telemetry standards | what an event *is* at the edge | envelope semantics, birth/death, report-by-exception (§3) |
| Alarm management | when a human should be interrupted | the cognition-rate target and its units (§4) |
| Anomaly-detection benchmarking | how synthetic benchmarks lie | four named failure modes to design against (§5) |

## 1. Deterministic simulation testing (DST)

The strongest prior art, and the one that determines the simulator's internal shape.

FoundationDB pioneered it around 2010: run the *real* software — not mocks — inside a
discrete-event simulator, with randomized workloads and aggressive fault injection, and
make every source of nondeterminism (clock, thread interleaving, RNG, I/O completion
order) a controlled input. Any bug found is then reproducible exactly from its seed.
The same team later built Antithesis, which does this at the hypervisor level for
arbitrary containers. TigerBeetle runs the technique at scale — reportedly two millennia
of simulated runtime per day across ~1,000 cores. WarpStream applied it to an entire
SaaS control plane.

**What we take, concretely:**

1. **Seed → single reproducible artifact.** A failing run must reduce to one
   `run.seed.json` that reproduces it. If a failure needs a log file plus a database
   plus "it happened on Tuesday," the architecture is wrong.
2. **Discrete-event core, no wall clock in the model.** The world advances by popping a
   priority queue keyed by `(event_time, tiebreak_seq)`. `time.Now()` appears nowhere in
   the world model. Any consumer with its own determinism boundary then composes with it
   cleanly, because neither side is reading a clock the other cannot see.
3. **A PRNG *tree*, not a PRNG.** The classic DST bug is a single global generator whose
   output depends on the order callers happen to draw from it — so adding one entity
   shifts every other entity's noise and every golden hash breaks. Instead: derive a
   named substream per `(entity_id, channel, purpose)` by hashing the name into a
   splitmix64 seed. Adding a channel then perturbs nothing else.
4. **Fault injection is a first-class input, not a test helper.** Faults are declared
   records in the run's command log, with the same identity and ordering guarantees as
   any other input.
5. **Randomized workload search, not hand-written cases.** Hand-written scenarios test
   what the author already imagined. The suite should include a seeded fuzz mode that
   composes perturbations at random and checks *invariants* rather than expected
   outputs.

**Where we differ:** DST frameworks typically test for crashes, deadlocks, and
invariant violations — properties with unambiguous truth. We additionally test
*judgment quality*, which has no such oracle. That difference is why §5 matters so much.

## 2. Co-simulation, digital twins, and FMI

The Functional Mock-up Interface (FMI, Modelica Association) is the de-facto standard
for exchanging simulation models between tools. An FMU is a zip containing
`modelDescription.xml` (typed variable declarations with causality and direction), a
platform-specific shared library implementing the FMI C API, and resources. FMI 3.0
added features aimed squarely at digital twins, virtual ECUs, and cloud/edge execution,
plus "layered standards" for embedding other standards' artifacts inside the container.

Open-source digital-twin frameworks (OpenTwins, SmartBuildSim, various IIoT toolkits)
converge on the same generation primitives: configurable **trend, seasonality, noise,
drift, and missingness**, over a semantic model of assets and points.

**What we take:**

1. **Fidelity is a declared tier, not an aspiration.** Four tiers, per channel:

   | Tier | Meaning | Cost | Where |
   |---|---|---|---|
   | `F0` statistical | trend + seasonality + noise + drift + missingness | trivial | transport and admission tests; any channel that is context, not subject |
   | `F1` phenomenological | first-order ODEs, thermal RC networks, dead time, saturation | low | **the default** for physical domains |
   | `F2` reference-model | a published equation (FAO-56 Penman-Monteith ET₀, IEC power curves, Arrhenius spoilage kinetics) | moderate | where the domain has an accepted standard and the test depends on it |
   | `F3` external co-simulation | an FMU stepped by the world clock | high | **deferred**; design the seam, build nothing |

   Declaring the tier per channel prevents the most common simulator failure: unbounded
   scope creep toward physical realism that no test actually needs.

2. **The FMI seam, unbuilt.** A domain channel may declare `fidelity: F3` with an `fmu`
   reference. The interface — step, get, set, with the world clock as the master — is
   specified in the technical design and returns `not_implemented`. Cost today: one
   paragraph. Value: the day someone has a real Modelica model of a chiller, the
   simulator is not the thing that has to be rewritten.

3. **Typed variable declarations with causality.** FMI's `input` / `output` / `parameter`
   distinction maps cleanly onto our channels (evidence out), effector arguments
   (control in), and world parameters (fixed at world creation).

**What we reject:** an actual physics engine. The simulator is a *stream* simulator. Its
job is to produce evidence with the right statistical and temporal character, not to be
right about bearings.

## 3. Industrial telemetry semantics

### Sparkplug B and OPC UA

Sparkplug B (Eclipse Foundation, now ISO/IEC) is MQTT plus a defined topic namespace,
protobuf payloads, and — the part that matters here — **device lifecycle via birth and
death certificates**. A node publishes a BIRTH containing every metric with its current
value; subsequent messages are **report-by-exception**, carrying only changed metrics; a
DEATH is delivered via MQTT Last Will and Testament when the node drops. Reported
bandwidth reduction is around 97%.

OPC UA supplies the complementary half: semantic information modelling through Companion
Specifications. Common practice is OPC UA at the machine, an edge gateway, Sparkplug B
northbound.

**What we take:**

1. **Report-by-exception is a channel mode, not an implementation detail.** A channel
   declared `report_by_exception` emits only on change beyond a deadband. This changes
   absence semantics fundamentally: silence means "unchanged", not "dead" — and the only
   thing distinguishing a healthy quiet channel from a dead one is a separate heartbeat.
   Several of the 25 domains exist specifically to make the runtime confront this.
2. **Birth/death as evidence.** A producer coming online should emit a birth burst — all
   channels at once, at the same timestamp, possibly with values hours stale. That burst
   is a nasty and completely realistic input: it looks like a spike, it violates
   monotonic event time, and it lands after a silence gap.
3. **The unified-namespace hierarchy** (`site/area/line/cell/device`) as the shape of
   `entity_id`, so the simulator's identifiers look like real plant identifiers and
   exercise a consumer's identifier alphabet at realistic length.

### ISO 13374 / OSA-CBM

ISO 13374, derived from MIMOSA's OSA-CBM, defines a six-block pipeline for condition
monitoring: **data acquisition → data manipulation → state detection → health assessment
→ prognostic assessment → advisory generation**.

The useful property is that it cleanly separates what the *simulator* owns from what a
*consumer* owns, along a line the industry settled decades ago:

| ISO 13374 block | Owner |
|---|---|
| data acquisition | **simulator** — it emits, with declared quality and timing |
| data manipulation | consumer — windows, aggregates, derived facts |
| state detection | consumer — deviation from normal |
| health assessment | consumer — the diagnosis |
| prognostic assessment | consumer — remaining useful life |
| advisory generation | consumer proposes; the simulator's effectors accept or refuse |

**What we take:** the vocabulary, and reassurance that the product boundary sits where an
established standard puts it — the simulator is block 1 and the effector end of block 6,
and everything between belongs to whoever is being tested. Domain specs name each
channel's role using these blocks, which makes the catalog legible to anyone from the
condition-monitoring world.

## 4. Alarm management — ISA-18.2 and EEMUA 191

The process industries have already had, and largely solved, the argument the agent
world is having now: *when should a system interrupt a human?*

Quantified findings worth borrowing:

- **Alarm flood**: >10 alarms in any 10-minute window — the point past which an operator
  cannot process each one.
- **EEMUA 191 target rates**: fewer than ~6 alarms per operator-hour in steady state;
  >12/hour unmanageable; >30/hour "seriously deficient."
- **Deadband and on/off-delay** alone have been measured to cut alarm load by 45–90%.
- **Rationalization** — justifying, prioritizing, and classifying every alarm before it
  exists — is the labour-intensive core of the practice, not an afterthought.

**What we take:**

1. **The interruption-rate metric gets real units and a real target.** A rate expressed
   as a fraction of events scales with event rate and is therefore incomparable across
   domains or consumers. Report the operator-facing form instead: **interruptions per
   entity per operator-hour**, targeted below 6, with a flood check (>10 in 10 minutes
   across the monitored fleet) as a hard failure.
2. **Deadband and on-delay are what every consumer reinvents.** The simulator should
   therefore be able to *check the 45–90% claim* on its own scenarios: run a domain
   against a detector with and without hysteresis and a sustained-duration requirement,
   and report the reduction. If it is not in that range, the noise model or the threshold
   placement is wrong — and that is a finding about the **simulator**, not the consumer.
3. **Alarm flood as a designed scenario, not an accident.** Every domain gets a
   `correlated_cascade` profile in which one root cause trips many entities at once
   (a power dip; a network partition; a chiller failure). This is the scenario where a
   naive per-entity trigger design produces 200 simultaneous episodes and a real cost
   ceiling gets hit — precisely what a consumer's aggregate cost ceiling and kill switch
   exist for, and a case such a mechanism is rarely tested against.

## 5. How synthetic benchmarks lie

Wu & Keogh, *Current Time Series Anomaly Detection Benchmarks are Flawed and are
Creating the Illusion of Progress* (arXiv 2009.13807), identifies four flaws in the
Yahoo, Numenta (NAB), and NASA datasets: **triviality, unrealistic anomaly density,
mislabeled ground truth, and run-to-failure bias**. Follow-on work (the UCR Anomaly
Archive; TSB-AD, >1,000 curated series) confirms and extends the critique; one study
found that trivial *model-selection* methods outperformed every anomaly detector tested.

This is the most important paper for this project, because our simulator will generate
its own benchmark, and generating your own benchmark is exactly how you get these flaws.

**Designed-in countermeasures, one per flaw:**

1. **Triviality → the trivial-baseline audit.** Every candidate scenario is run against
   a panel of one-line detectors (fixed threshold, z-score, first difference, moving-
   median residual). If any solves it, the scenario is stamped `trivial` and excluded
   from the headline metric — kept in the suite as a regression fixture, but never
   counted as evidence that reasoning helped. `sim scenario audit` refuses to admit a
   scenario to the graded suite without at least one non-trivial discriminator.
   *This makes the suite adversarial toward its own authors, which is the point.*

2. **Density → declared prevalence with a large negative class.** The existing evaluation
   design has one no-fault label out of seven, i.e. ~14% negatives. Real fleets are
   overwhelmingly healthy. Target **35–45% no-fault scenarios**, declared per suite, and
   report precision alongside recall so the cost of crying wolf is visible.

3. **Mislabeled truth → three timestamps, not one.** The injected onset is *not* when the
   fault becomes detectable. A slow bearing ramp at 0.01 mm/s per hour is invisible at
   injection and unmissable twelve hours later. Scoring detection latency against
   injection time is unfair and produces meaningless numbers. Every label therefore
   carries:

   - `injection_time` — when the world changed;
   - `first_observable_time` — computed, when the signal first exceeds the channel's
     noise floor at the declared SNR;
   - `unavoidable_time` — computed, when any competent detector must catch it.

   Detection before `first_observable_time` is *suspicious* (check for a leak, §8 of
   the transport analysis). Detection after `unavoidable_time` is a miss. The interval
   between them is where the actual comparison lives.

4. **Run-to-failure bias → randomized onset position, including before t₀.** Faults must
   not cluster at the end of traces. Onset position is sampled across the trace, and a
   fraction of scenarios begin with the asset **already degraded** — which also exercises
   a consumer's cold-start handling — the case where monitoring is deployed onto an
   already-degraded asset, with no healthy baseline to compare against.

A fifth flaw, ours rather than theirs: **the oracle leak** — the system under test
reading the answer from the simulator's control plane. Covered in the transport
analysis §8.

## 6. IoT simulators and MCP-driven simulators

**IoT/building simulators** (SmartBuildSim, OpenTwins, assorted device simulators that
target Azure IoT Hub or an MQTT broker) confirm the generation primitives in §2 and are
otherwise unremarkable for our purposes: they are open-loop trace producers with no
ground truth, no determinism contract, and no actuation. They generate data; they do not
constitute a test.

**MCP-driven simulators** are now common in robotics — Isaac Sim MCP servers (one exposes
~42 tools across scene, objects, lighting, robots, sensors, materials, assets, and
simulation), and Gazebo reached through ROS MCP servers. The pattern is well established
and the ergonomics are proven: agents drive simulators through MCP tool calls
comfortably.

Two observations from that ecosystem:

1. **They all use MCP for control, and none for the sensor firehose.** Isaac Sim's tool
   list is scene manipulation and simulation control. Camera and lidar data does not
   traverse MCP. This is convergent evidence for the split in the transport analysis —
   it is what everyone independently does.
2. **A commonly cited property is that "the AI agent doesn't know or care whether it's
   controlling a simulated or real robot — the MCP tool interface is identical."** That
   property is a *hazard* for us, not a feature. Our effector interface must be
   identical in shape but unmistakable in identity: every operator-role response carries
   a `simulated: true` marker and a world id, and the simulator has no code path reaching
   anything outside its own process except a configured sink endpoint. It must be
   impossible to later confuse a simulated effector with a real one.

## 7. Consolidated list of adopted patterns

| # | Pattern | Source | Where it lands |
|---|---|---|---|
| 1 | Seed → single reproducible artifact | FoundationDB / TigerBeetle | TECHNICAL_DESIGN §3 |
| 2 | Discrete-event core, no wall clock in the model | DST | TECHNICAL_DESIGN §3 |
| 3 | Named PRNG substream tree | DST | TECHNICAL_DESIGN §3.2 |
| 4 | Faults as logged inputs | DST | TECHNICAL_DESIGN §5 |
| 5 | Seeded fuzz over perturbation composition | DST | IMPLEMENTATION_PLAN M4 |
| 6 | Declared per-channel fidelity tiers | FMI / digital twins | DOMAIN_CATALOG §2 |
| 7 | Unbuilt FMI co-simulation seam | FMI 3.0 | TECHNICAL_DESIGN §9.4 |
| 8 | Report-by-exception channel mode | Sparkplug B | DOMAIN_CATALOG §2 |
| 9 | Birth/death bursts as evidence | Sparkplug B | TECHNICAL_DESIGN §6 |
| 10 | Hierarchical plant-style entity ids | Unified Namespace | contracts |
| 11 | Six-block role vocabulary | ISO 13374 / OSA-CBM | DOMAIN_CATALOG |
| 12 | Episodes per operator-hour, target < 6 | EEMUA 191 | GROUND_TRUTH_AND_SCORING §4 |
| 13 | Flood check: >10 in 10 minutes | ISA-18.2 | GROUND_TRUTH_AND_SCORING §4 |
| 14 | Prove the 45–90% suppression claim | ISA-18.2 literature | GROUND_TRUTH_AND_SCORING §4 |
| 15 | `correlated_cascade` profile per domain | alarm flood | DOMAIN_CATALOG §2 |
| 16 | Trivial-baseline audit, gated | Wu & Keogh | GROUND_TRUTH_AND_SCORING §3 |
| 17 | 35–45% negative class | Wu & Keogh | GROUND_TRUTH_AND_SCORING §2 |
| 18 | Three onset timestamps | Wu & Keogh | GROUND_TRUTH_AND_SCORING §2 |
| 19 | Randomized onset, incl. pre-degraded start | Wu & Keogh | GROUND_TRUTH_AND_SCORING §2 |
| 20 | MCP for control only; `simulated: true` marking | Isaac Sim / ROS MCP | MCP_SURFACE |

## Sources

- [Deterministic simulation testing — Antithesis](https://antithesis.com/docs/resources/deterministic_simulation_testing/)
- [Diving into FoundationDB's Simulation Framework](https://pierrezemb.fr/posts/diving-into-foundationdb-simulation/)
- [What's the big deal about Deterministic Simulation Testing?](https://notes.eatonphil.com/2024-08-20-deterministic-simulation-testing.html)
- [Deterministic Simulation Testing for Our Entire SaaS — WarpStream](https://www.warpstream.com/blog/deterministic-simulation-testing-for-our-entire-saas)
- [Release of FMI 3.0](https://fmi-standard.org/news/2022-05-10-fmi-3.0-release/)
- [Functional Mock-up Interface — Wikipedia](https://en.wikipedia.org/wiki/Functional_Mock-up_Interface)
- [OpenTwins: an open-source framework for compositional digital twins](https://www.sciencedirect.com/science/article/pii/S0166361523001574)
- [SmartBuildSim](https://www.ncbi.nlm.nih.gov/pmc/articles/PMC12693821/)
- [What is Sparkplug B? — Software Toolbox](https://softwaretoolbox.com/resources/what-is-sparkplug-b)
- [Key differences between OPC UA and MQTT Sparkplug — HiveMQ](https://www.hivemq.com/blog/iiot-protocols-opcua-vs-mqtt-sparkplug-digital-transformation/)
- [ANSI/ISA-18.2-2016 Management of Alarm Systems](https://18817087.s21i.faiusr.com/61/ABUIABA9GAAgyZfj5AUozIu7wwI.pdf)
- [ISA-18.2 alarm management guidelines — ProcessVue](https://www.processvue.com/resources/alarm-management-guidelines/)
- [ISO 13374: Condition Monitoring Data Standard](https://vibromera.eu/glossary/iso-13374/)
- [ISO/CD 13374-2 draft](http://alvarestech.com/temp/osacbm/N006%20-%20ISO%20CD%2013374-2.pdf)
- [Current Time Series Anomaly Detection Benchmarks are Flawed (arXiv 2009.13807)](https://arxiv.org/abs/2009.13807)
- [The Elephant in the Room: Towards A Reliable Time-Series Anomaly Detection Benchmark](https://openreview.net/forum?id=R6kJtWsTGy)
- [Isaac Sim MCP Extension and Server](https://github.com/omni-mcp/isaac-sim-mcp)
- [MCP and Robotics: bridging AI agents and robot systems via ROS](https://chatforest.com/guides/mcp-robotics-ros-integration/)
- [Coupled weather and crop simulation modeling for smart irrigation planning](https://iwaponline.com/ws/article/24/8/2844/103637/Coupled-weather-and-crop-simulation-modeling-for)
- [Agent Data Injection Attacks are Realistic Threats to AI Agents (arXiv 2607.05120)](https://arxiv.org/pdf/2607.05120)
