# Core concepts

> Status: Implemented conceptual model. Authority: current world, run, perturbation, and scoring packages. Verified by: architecture review and package tests. Last verified: 2026-08-17.

Streams Simulator is easier to use when five distinctions stay clear: world truth, delivered evidence, commands, artifacts, and verdicts.

## World

A world is a seeded discrete-event system compiled from one domain specification. It owns entities, state variables, natural dynamics, faults, effectors, and the simulation clock. The world emits native `sim-event-v0.1` records describing what happened at a simulation time.

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

## Truth and verdict

Ground truth is sealed for the consumer-facing phase. The consumer submits a verdict describing detections, actions, and optional resource evidence. The scorer compares that verdict with sealed truth, the delivery ledger, and effector history. Truth reveal is a one-way unblind transition for a run.

## Vocabulary

| Term | Meaning |
| --- | --- |
| Domain | JSON specification of entities, channels, state, dynamics, faults, effectors, and profiles. |
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
