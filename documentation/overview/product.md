# Product overview

> Status: Implemented product summary with conditional release posture. Authority: root README, current code/tests, and accepted decisions. Verified by: repository review and quickstart. Last verified: 2026-11-03.

Streams Simulator (`streamsim`) is a standalone, deterministic, closed-loop world simulator.

It builds worlds from data (domains, or plants composed from a component library), generates their event streams, applies declared delivery faults, carries sealed ground truth, accepts commands through declared effectors, and can grade a consumer’s verdict against what actually happened. A visual canvas, physical-device sessions over the plant wire, run reports and generated campaigns are built on the same core.

## What it is

A simulation platform whose credibility comes from making the observed evidence and the hidden truth precise enough to reproduce and audit. Typical uses: exploring and authoring plants, driving a stream processor, agent or controller in a closed loop, benchmarking consumers, running a real board beside simulated ones, and comparing learning against non-learning arms over a campaign.

- **Data-defined worlds.** Domains, channels, states, faults, effectors, and scenario profiles are JSON data loaded through common code.
- **Deterministic execution.** Seeds, versioned inputs, command logs, and sink choice define a replayable run.
- **Explicit delivery corruption.** Perturbations change what the consumer receives, not what the world did. The delivery ledger records the distinction.
- **Authoring and exploring.** Plants are composed from equipment, instruments and vocabulary, linted for plausibility, diffed, run, disturbed and reviewed in the canvas or from the CLI and MCP tools.
- **Physical devices.** A real board can join a session beside simulated ones over the plant wire, with interrupts, approvals and a twin-versus-board report.
- **Closed-loop actuation.** An effector command changes world state; later events can show whether the change took effect.
- **Sealed truth.** A consumer interacts through an operator surface without receiving the ground truth before the declared reveal transition.
- **Independent scoring.** The simulator scores verdicts using truth, delivery evidence, and action records rather than trusting the consumer’s own summary.

## What it is not

- It is not a message broker or a production event transport.
- It is not a physics engine, an industrial controller, or an emergency-stop system. With a real board attached ([physical devices](../guides/physical-devices.md)) it can hold, approve or stop commands at its node (a declared safe stop without arguments is never held), but it only refuses or stops, never originates an ordinary command, and the board's own firmware stays the last line of defence.
- It is not a consumer-specific benchmark binary. Consumer names, schemas, and behavior belong in adapters or external consumers.
- It is not a claim that arbitrary remote effects are exactly once. The simulator can make its own command log deterministic; it cannot make an external system transactional.
- It is not a license to publish benchmark numbers without a fresh, reproducible evidence bundle.

## Who it serves

- Plant and system designers who want to build, disturb and review a plant before or beside the real one.
- Bench engineers running a real board beside simulated ones.
- Stream-processor authors who need controlled faults, delayed or missing evidence, and replayable failures.
- Benchmark authors who need a world model, a delivery ledger, and an oracle independent of the consumer.
- Integration authors who need a stable native event contract and declarative output adapters.
- Maintainers who need adversarial tests for determinism, truth isolation, actuation, and transport semantics.

## The central boundary

The world describes what happened. Perturbation describes what was delivered. The adapter describes how that delivery is projected into a consumer format. The intended protocol evaluates the consumer’s verdict after the run reaches its declared end and before truth is unblinded; callers must preserve that sequence.

That separation is the product. If a domain or consumer requires a special branch in the simulator binary, the architecture has leaked.

## Current status

The current checkout contains an implemented Go CLI, deterministic run/replay paths, MCP director and operator surfaces, shipped domain and adapter data, schemas, a reference consumer, and extensive tests. It is not currently a clean release baseline: see [limitations](../limitations.md) for known status issues and [benchmark evidence](../benchmark/evidence.md) for the conditions required before publishing results.

## Next reads

- [Core concepts](concepts.md)
- [Architecture overview](../architecture/overview.md)
- [Quickstart](../getting-started/quickstart.md)
- [Limitations](../limitations.md)
