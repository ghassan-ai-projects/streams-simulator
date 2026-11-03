# Core concepts

> Status: Implemented conceptual model. Authority: current world, run, perturbation, and scoring packages. Verified by: architecture review and package tests. Last verified: 2026-11-03.

Streams Simulator is easier to use when a few distinctions stay clear: world truth, delivered evidence, commands, artifacts, and verdicts. Whatever drives a world (a person in the canvas, a script, a stream processor, an agent, a controller, a real board) meets the same ones.

## World

A world is a seeded discrete-event system compiled from one domain specification. A domain is written by hand or composed from the component library (equipment, instruments and vocabulary as data); a composition is written out as an ordinary domain. It owns entities, state variables, natural dynamics, faults, effectors, and the simulation clock. The world emits native `sim-event-v0.1` records describing what happened at a simulation time.

## Delivery

The delivery path is deliberately separate from the world:

1. The world emits a native event.
2. The perturbation layer may delay, duplicate, drop, reorder, mangle, or otherwise transform delivery according to the domain and scenario.
3. The adapter renders the delivered event into a consumer wire format.
4. The sink writes or pushes the rendered record.
5. The ledger records a terminal delivery state.

The consumer must reason from step 3. The scorer can reason from both the consumer-visible delivery and the director-only ledger.

## Closed loop

An effector invocation is a command with an idempotency key. The command is accepted against the loaded domain’s declared effectors and scheduled at a world time. The effector may change hidden state immediately or through a declared time constant. Later events expose the consequence—or its absence.

`silent_no_effect` is the important case: an acknowledgement does not prove that the world changed. A consumer that reports success without evidence should be distinguishable from one that waits for the consequence.

## Run artifact

A run artifact records the inputs and command history needed to replay a run: simulator version, domain and adapter identities/digests (and embedded specs when available), seed, world configuration, command log, expected trace digest, counts, and completion status. Replay re-executes the command log against a fresh world; it does not re-issue the original MCP conversation.

## Devices, runs and campaigns

A real board can join a session beside simulated ones over the plant wire; a node can hold, approve or stop commands, and a twin-versus-board report compares the board with its simulated twin. Every run writes a run folder (artifact, trace, ledger, journals) that `report` summarises and `score` grades offline. `suite campaign` generates one-fault runs with session plans, so arms can be compared over a campaign. See [physical devices](../guides/physical-devices.md) and the [CLI reference](../reference/cli.md).

## Truth and verdict

Ground truth is sealed for the consumer-facing phase. The consumer submits a verdict describing detections, actions, and optional resource evidence. The scorer compares that verdict with sealed truth, the delivery ledger, and effector history. Truth reveal is a one-way unblind transition for a run.

## Vocabulary

| Term | Meaning |
| --- | --- |
| Domain | JSON specification of entities, channels, state, dynamics, faults, effectors, and profiles. |
| Component library | Equipment, instruments and vocabulary as data, composed into a domain. |
| Kind file | Declaration of a device kind (readings, commands, safe stops) shared by a twin and a real board. |
| Run folder | The files one run leaves: artifact, trace, ledger, state history and journals. |
| Adapter | JSON projection from native simulator events to a consumer wire format. |
| Perturbation | A declared transformation of delivery; it does not rewrite world history. |
| Sink | In-process, file, or HTTP-push destination for rendered evidence. |
| Ledger | Immutable per-delivery evidence distinguishing delivered records, transport misses, mangling, omission, and sink errors. |
| Effector | Domain-declared command that can change world state. |
| Nameplate | Operator-visible static description of entities, channels, and effectors for one world. |
| Verdict | Consumer report submitted for scoring. |
| Quiescence | The barrier used when advancing time while awaiting consumer-side processing. |

## Next reads

- [Pipeline](../architecture/pipeline.md)
- [Determinism and replay](../architecture/determinism.md)
- [Truth and scoring](../architecture/truth-and-scoring.md)
