# Consumers

Status: design
Date: 2026-08-11

The simulator is a standalone product. A **consumer** is anything that reads its output.
This document is the integration contract, and it is deliberately short — most of the work
is on the simulator's side, which is the point.

## 1. The contract

A consumer integrates by doing one required thing and two optional ones.

### Required · read the output

Point the simulator at a sink and read what comes out. An
[output adapter](../contracts/output-adapter-v0.1.schema.json) projects the native
`sim-event-v0.1` into whatever format the consumer already accepts.

Writing an adapter is a JSON file: field mappings drawn from a closed set of transforms,
optional preamble and postamble records, an identifier rewrite if the consumer's alphabet
is narrower than the simulator's, and a conformance block naming the consumer's published
schema and a golden fixture.

`streamsim adapter verify <id>` proves the adapter correct **with the consumer absent** —
it renders the shipped 12-event fixture, validates against the declared schema, and byte-
compares to the golden file. A consumer team can therefore accept or reject an adapter by
reading one file and running one command.

A consumer that stops here still gets deterministic, reproducible, labelled, perturbed
evidence with hidden ground truth. That is most of the value.

### Optional · close the loop

Connect an MCP client to the `operator` role and call `sim.effector.invoke` with a
`command_id` and a capability token. Effects propagate through the world with physical
time constants, and the consequences appear in the ordinary evidence stream.

This is what makes outcome measurement possible, including the false-success test that
nothing else can perform.

### Optional · be scored

Submit a [consumer verdict](../contracts/consumer-verdict-v0.1.schema.json) through
`sim.consumer.report`: detections with timestamps and evidence citations, per-event
admission outcomes, actions taken, counters, resource usage.

The simulator scores that against sealed truth, its own delivery ledger, and its own
effector log. It never reads a consumer's database, parses its logs, or links its code.

### Also optional · quiescence

For reproducible closed-loop runs, report `quiesced_through_ns` — "I have processed every
record through this instant and hold no pending work." `sim.clock.advance` with
`await_consumer: true` blocks on it, so the world does not run ahead of the consumer's
reaction.

A consumer that does not implement this can still run open-loop; the run artifact simply
records that it was not reproducible.

## 2. What the simulator never requires

- no change to the consumer's ingress;
- no simulator-specific library, SDK, or dependency;
- no exposure of the consumer's storage, logs, or internals;
- no shared process, language, or runtime;
- no vocabulary change — the verdict's seven admission outcomes are a neutral target the
  consumer maps its own taxonomy onto, and the mapping is the consumer's business.

## 3. What the simulator asks in return

Two things, both narrow:

1. **A capability token and a `command_id` on every effector call.** The simulator holds
   exactly one opinion about a consumer's internals — that an actuator is driven by a
   deliberate, identified dispatch — and no opinion whatsoever about who inside the
   consumer was permitted to decide it.
2. **Honest verdicts.** A consumer that reports actions it did not take is caught by the
   effector log, and one that cites evidence it was never delivered is caught by the
   delivery ledger. Both are findings, not errors.

## 4. The reference consumer

`streamsim refconsumer` ships with the product: reads an adapter's output, applies a
trivial threshold detector, invokes effectors, submits a verdict. A few hundred lines.

It exists so the simulator is complete on its own — the closed loop is demonstrable and
CI-gated with nothing else installed — and so a new consumer has something to copy rather
than a specification to interpret.

## 5. First consumer: Agentic Stream

Written up here as a worked example. Nothing in this section is a dependency of the
simulator, and deleting it changes no simulator behaviour.

### 5.1 The adapter

[`adapters/agentic-stream.adapter.json`](../adapters/agentic-stream.adapter.json).
Everything Agentic-Stream-specific lives in that file: its `trace-record-v0.1` framing,
its `runtime_config` preamble and `trace_end` postamble, its `evt-NNNNNN` identifier
convention, its narrower id alphabet (the simulator's `/` is rewritten to `.`), and its
`arrival_time` field, which is fed from the native `observed_time`.

### 5.2 What each side must do

| Side | Work |
|---|---|
| Simulator | none beyond the adapter file |
| Agentic Stream | nothing to read traces. To close the loop: an effector adapter that dispatches `sim.effector.invoke` with `command_id` as the idempotency key, and classifies the seven failure modes into its own outcome taxonomy — in particular distinguishing `ack_lost` (ambiguous, must not double-apply) from `reject` (definite) and `interlock_refused` (terminal, never retry, never route around). To be scored: emit a consumer verdict. To be reproducible in closed-loop: report quiescence. |

Everything in that right-hand column is work Agentic Stream would need for any simulated
plant. None of it is owed to this simulator specifically.

### 5.3 One recommendation for that project

**Drop `eval generate` from the Agentic Stream V0.1 plan.** It is a single-domain,
open-loop fault-model trace generator — a strict subset of this simulator. Building both
produces two generators that will drift and then disagree about what a trace is.

This is a recommendation to that project, not a requirement of this one.

### 5.4 Design questions this consumer is likely to hit

The [domain catalog](DOMAIN_CATALOG.md) predicts eight. They are properties of *that*
runtime, discovered by modelling domains, and they are listed there rather than here
because a second consumer will hit a different eight.

Four look like defects rather than scope choices: a single lateness allowance cannot serve
channels four orders of magnitude apart; entities are never retired, so churn leaks state
without bound; free-text inputs reach the reasoning snapshot by design; and an interlock
refusal has no representation distinct from a retryable error.

Confirming or refuting those is the first useful thing the simulator will do for its first
customer — and it is a *result the instrument produces*, not a dependency it carries.
