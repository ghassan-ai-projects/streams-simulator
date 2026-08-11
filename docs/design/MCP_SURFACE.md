# MCP Surface

Status: implemented core surface; deferred items are listed explicitly below
Date: 2026-08-11

One MCP server. Two roles. Every tool is domain-agnostic.

Read [TRANSPORT_ANALYSIS.md](../research/TRANSPORT_ANALYSIS.md) first: it establishes why
MCP carries control and actuation but not evidence, and why role separation is a
correctness requirement rather than tidiness.

## 1. Two design rules

**R1 · No tool name, argument, or enum value may mention a domain.**

The binary does not know that ponds have aerators. It knows that a loaded domain spec
declares effectors, and it exposes one generic tool for invoking them. Twenty-five domains
and the twenty-sixth all work through the same four operator tools.

The earlier draft generated one MCP tool per effector — `plant.aquaculture.start_aerator`
and so on. That is domain knowledge reaching the protocol surface. It also makes the tool
catalog mutate whenever a world is created, which breaks catalog-epoch immutability for
any host that caches it. Both problems disappear with a generic invoke.

**R2 · No tool name, argument, or enum value may mention a consumer.**

There is no "stream" anything. A consumer is whoever holds an operator token.

## 2. One server, two roles

| | `director` | `operator` |
|---|---|---|
| Purpose | build, drive, perturb and audit a world | act on the world, as a consumer of it |
| Held by | the harness, or a human operator | the system under test |
| Knows ground truth | yes | **the view struct has no such field** |
| Can mutate the world | arbitrarily | only through declared effectors, with their declared effects |
| Can move the clock | yes | no |

One process, one endpoint, two handler sets constructed from different view structs. The
role is fixed at session initialize:

```bash
streamsim mcp --role director            # stdio
```

The current prototype serves the director over stdio. The operator server is constructed
per world by the host from the capability token; there is not yet a standalone operator
CLI or HTTP `serve` command. A session's role never changes, and each server advertises
only its role's tools — so a consumer cannot discover the existence of a director tool,
let alone call it.

The role-specific servers advertise MCP protocol range `2025-11-25`–`2026-07-28`, pinned
per configuration, failing closed outside it. They are stateless in the MCP sense:
identity lives in the world and run ids the tools carry, not in a transport session, so a
dropped connection loses nothing.

## 3. Director tools

### 3.1 Catalog

| Tool | Args | Returns |
|---|---|---|
| `sim.catalog.list` | `{group?}` | domains: id, title, axes vector, `stresses` |
| `sim.catalog.describe` | `{domain}` | the full spec: channels, faults, effectors, profiles, fidelity tiers |
| `sim.catalog.coverage` | `{}` | the axis-coverage matrix; which axes are thin |
| `sim.adapter.list` | `{}` | installed output adapters |

`adapter verify` is currently a CLI command (`streamsim adapter verify`), not an MCP
tool. Keeping conformance verification out of the live director surface avoids implying
that a consumer's schema or golden fixture is part of a running world.

Resources: `sim://catalog` and `sim://domains/{id}/spec`. Adapter metadata is available
through `sim.adapter.list`; an adapter resource template is not exposed yet.

### 3.2 World lifecycle

| Tool | Args | Returns |
|---|---|---|
| `sim.world.create` | `{domain, seed?, entities?, scenario_profile?, sink?, sink_target?, adapter?, time_mode?, start_time?, label?}` | `{world_id, world_digest, entity_ids[], clock, token}` |
| `sim.world.describe` | `{world_id}` | domain, seed, clock, emitted count |
| `sim.world.destroy` | `{world_id}` | `{world_id, destroyed}`; final artifact is written under the configured output directory |
| `sim.entity.retire` | `{world_id, entity_id, reason}` | churn domains |

`seed`, `start_time`, and all clock arguments are integer nanoseconds or integer seed
values. `start_time` defaults to `2026-01-01T00:00:00Z`
(`1767225600000000000` ns), the same default used by the CLI. `time_mode` is `stepped`
(the clock moves only when told — the deterministic default), `scaled` (real time ×
multiplier), or `wall`.

`sink` defaults to `inproc`; `sink_target` is required for `file` and `http-push`.
`sink` and `adapter` are independent: any adapter can be written to any sink.

`sim.entity.add` is not exposed over MCP yet. The run package has the replay primitive,
but the public identity-allocation contract is not complete, so it remains deferred
rather than accepting an ambiguous empty entity id.

### 3.3 Clock

| Tool | Args | Returns |
|---|---|---|
| `sim.clock.advance` | `{world_id, by_ns \| to_ns, await_consumer?}` | `{emitted, emitted_total, clock, effects_applied}` |
| `sim.clock.state` | `{world_id}` | `{clock, next_scheduled_ns, pending_effects}` |

Exactly one of `by_ns` and `to_ns` is required. `by_ns` is a non-negative relative
nanosecond delta; `to_ns` is an absolute epoch nanosecond timestamp. The clock never moves
backwards; a request to do so is refused. `emitted` is the number produced by this call,
while `emitted_total` is cumulative for the world.

`sim.clock.run` is not built. Continuous mode is deliberately cut for this release and
maps to repeated `sim.clock.advance` calls in a harness.

`await_consumer` is the quiescence barrier ([GAP_ANALYSIS G-02](GAP_ANALYSIS.md)). In
`stepped` mode `advance` returns once every record up to the new instant has been written
to the sink. That is well-defined for the simulator and says nothing about whether the
consumer has finished reacting — so with `await_consumer: true` the call additionally
blocks until the consumer reports quiescence through §4.4. Without it, a closed-loop run
is not reproducible, because the effector call lands at an arbitrary world time.

For the `file` sink, writes use a buffered writer. The path is created at
`sim.world.create`, but newly emitted bytes become visible on disk when the run is closed
by `sim.run.end` or `sim.world.destroy`. Consumers should use the MCP count response while
the world is open and read the finalized file after close.

### 3.4 Faults, perturbations, environment

| Tool | Args | Returns |
|---|---|---|
| `sim.fault.inject` | `{world_id, entity_id, fault, onset_ns?, params?}` | `{fault_id}` |
| `sim.fault.clear` | `{world_id, fault_id}` | — |
| `sim.fault.list` | `{world_id}` | active faults — **director only** |
| `sim.perturb.apply` | `{world_id, perturbation, params?, from_ns?, until_ns?}` | `{perturb_id}` |
| `sim.perturb.clear` | `{world_id, perturb_id}` | — |
| `sim.env.inject` | `{world_id, target, fault, params?}` | `{env_id}` — record an environment fault against a configured consumer endpoint |

Three surfaces, three tools, never conflated: `fault` changes physics, `perturb` changes
delivery, `env` changes the consumer's environment.

`sim.env.inject` is the one place the simulator touches a consumer, and it does so through
a *configured target* — a process id, a URL, a container name given at world creation.
It has no idea what the target is.

### 3.5 Export, runs, truth, scoring

| Tool | Args | Returns |
|---|---|---|
| `sim.run.begin` | `{world_id, label?}` | `{run_id}` |
| `sim.run.end` | `{world_id}` | `{run_artifact_path, trace_digest}` |
| `sim.run.verify` | `{run_artifact_path}` | `{matches, first_divergence?}` |
| `sim.truth.seal` | `{run_id, ground_truth}` | `{run_id, sealed}` |
| `sim.truth.reveal` | `{run_id, unblind?}` | labels, three onset timestamps, hidden state history |
| `sim.truth.seal_status` | `{run_id}` | `{sealed, unblinded, unblinded_at}` |
| `sim.score` | `{run_id}` | scorecard, computed from the submitted verdict, sealed truth, the delivery ledger and the effector log |
| `sim.scenario.audit` | `{domain, entity_id, fault, onset_ns?, start_ns?, duration_ns?}` | per-scenario trivial-baseline verdict |

`sim.trace.export` is not built. Evidence is delivered through the configured sink and the
run artifact is finalized by `sim.run.end` or `sim.world.destroy`; exporting a second live
trace path would create another delivery contract.

`sim.score` takes no consumer artifact of any kind. Everything it needs was either
generated by the simulator or submitted through §4.4.

`sim.truth.reveal` refuses on an open run unless `unblind: true`, which permanently stamps
the run and excludes it from every scorecard. Fail loud, not closed.

Every tool advertises a closed, typed input schema (`additionalProperties: false`); the
SDK rejects unknown and missing arguments before a handler runs. The nested
`ground_truth` (on `sim.truth.seal`) and `verdict` (on `sim.consumer.report`) arguments
are validated against the committed ground-truth and consumer-verdict contracts, so an
extra or malformed field is refused, not silently ignored. The remaining free-form
objects — `params` on fault/perturb/env tools and `args` on `sim.effector.invoke` — are
domain- or effector-dependent by design; their handlers reject unknown keys against the
declared parameter set (per-perturbation, `severity` for faults, the effector's declared
argument schema), so a misspelled option is an error, never a silently approximated
default.

## 4. Operator tools

Everything here is reachable by the system under test. The rule governing this section,
and every future addition to it:

> **Everything reachable from the operator role must be computable from the delivered
> evidence alone.**

Four tools. No evidence surface — evidence arrives on the sink, never here.

### 4.1 `sim.nameplate.read`

`{token}` → entities (ids, types, hierarchy), channels (names, units, declared ranges,
resolutions), effectors (names, argument schemas, risk classes).

Static and time-invariant, exactly the information an instrument datasheet gives you. A
nameplate field that changes as the world evolves is a covert evidence channel with none
of the evidence plane's guarantees, and is forbidden.

### 4.2 `sim.effector.list`

`{token}` → the declared effectors and their argument schemas, read from the loaded
domain spec.

### 4.3 `sim.effector.invoke`

```json
{
  "token": "t-...",
  "effector": "start_aerator",
  "entity_id": "site-a/pond-3",
  "command_id": "cmd-91",
  "args": {"level": 1.0}
}
```

→ `{accepted, simulated: true, world_id, command_id, effect_eta_ns?, reason?}`

One tool for every effector in every domain. `effector` is a string validated against the
loaded spec; `args` is validated against that effector's declared schema. **The binary
contains no effector name.**

- **`command_id` is required.** It is the idempotency key. Repeating a call with the same
  `command_id` inside the declared window returns the original result and applies no
  second effect — the property that makes the `ack_lost` failure mode survivable, and the
  reason that mode is worth injecting.
- **A capability token is required**, minted at world creation. Together with
  `command_id` this is how the simulator declines to be actuated by anything other than a
  consumer's deliberate, identified dispatch.
- Any of the seven failure modes in [TECHNICAL_DESIGN §8.2](TECHNICAL_DESIGN.md) may be
  returned, per the injected configuration.
- Every call is recorded with full arguments and the resulting world delta. That log, not
  the consumer's claim, is the authority when scoring actions.

### 4.4 `sim.consumer.report`

```json
{"token": "t-...", "run_id": "r-114", "quiesced_through_ns": 1785315600000000000, "verdict": { ... }}
```

Two jobs, one tool, because they are the same statement at different granularities.

**Quiescence** — `quiesced_through_ns` asserts the consumer has processed every record
through that instant and holds no pending work. `sim.clock.advance` with
`await_consumer: true` blocks on it. A consumer that does not implement this can still run
open-loop; it simply cannot be used for reproducible closed-loop runs, and the run artifact
records that.

**The verdict** — a
[consumer verdict](../contracts/consumer-verdict-v0.1.schema.json): detections, per-event
admission outcomes, actions taken, counters, resource usage. Write-only. It returns an
acknowledgement and never a score, so submitting cannot be used to probe for the answer.

This is the whole scoring interface. The simulator never reads a consumer's database,
parses its logs, or links its code.

### 4.5 Never exposed to the operator role

Enumerated so review has a checklist — though the differential test in §5.2 is what
actually enforces it:

- fault ids, names, parameters, or their existence;
- any of the three onset timestamps;
- hidden world state, including the true value behind a noisy observation;
- future values or scheduled emissions;
- ground-truth labels or scenario metadata;
- which delivered events were perturbed, or that any were;
- the world's PRNG state or seed;
- run ids other than its own, scorecards, or suite membership;
- error text that differs by fault — a leak through the error taxonomy;
- response latency that differs by fault — plant responses are padded to a fixed schedule
  in graded runs.

The last three are the ones a reviewer misses.

## 5. Enforcing the separation

### 5.1 Compile-time

The operator handlers are constructed from an `OperatorView` holding the effector table,
the nameplate, and a verdict sink. It has no reference to the truth store, the fault
registry, the perturbation log, or hidden state. There is no field to leak, so a leak
requires adding one — a visible diff in the one file a reviewer must read.

### 5.2 Prefix indistinguishability

Static review misses side channels:

1. Build worlds `W_A` and `W_B` from the same domain and seed, differing only in injected
   fault.
2. Advance both to the last instant at which their **delivered evidence** is identical.
3. Replay the same operator calls against both.
4. Assert every response is byte-identical and every latency falls in the same padded
   bucket.

Run for every domain, in CI. Catches leaks through error strings, field ordering, map
iteration and timing.

### 5.3 Configuration guard

`streamsim mcp --role operator` refuses to start if the same process is also serving
director tools on the same transport. The distributed artifacts include a lint that fails
any consumer configuration listing a director endpoint, with a message naming the leak —
because that mistake is easy to make while debugging at 01:00 and impossible to notice in
the results.

## 6. Interaction patterns

### 6.1 Harness — deterministic, graded

```text
director: catalog.describe(aquaculture-pond)
director: world.create(seed=42, sink=file, adapter=<consumer>, time_mode=stepped)
director: truth.seal(run_id, ground_truth)                 # director-only, before begin
director: run.begin()
director: fault.inject(pond-3, do_probe_fouling, onset=T+4h)
director: perturb.apply(duplicate_burst, rate=0.02)
director: clock.advance(by=24h, await_consumer=true)
          -- consumer reads the trace, reasons, decides --
operator: effector.invoke(start_aerator, pond-3, command_id=cmd-91)
director: clock.advance(by=2h, await_consumer=true)     # the effect propagates
operator: consumer.report(verdict)
director: run.end() ; score(run_id)
director: truth.reveal(run_id)                          # after the run closes
```

### 6.2 Exploration — improvised, still reproducible

A human in Claude Code, or an agent, improvises against the director role. Every call
appends to the command log, so the session collapses to one run artifact and
`streamsim replay run.json` reproduces it exactly with no server running.

Worth being precise about what earns this: it is a property of the **command log**, not of
MCP. A CLI writing the same log would get it too. MCP's contribution is that the
improvisation can be conversational.

### 6.3 Live — the closed loop

`time_mode: scaled`, `sink: http-push`, multiplier 60. Not reproducible, and stamped as
such. The demo, the soak test, and the only configuration exercising everything at once.

## 7. Error taxonomy

Typed and stable. Operator-facing codes are deliberately coarse so the text cannot leak
state.

| Code | Role | Meaning |
|---|---|---|
| `domain_invalid` | director | spec failed validation |
| `adapter_invalid` | director | adapter failed validation or golden comparison |
| `world_not_found` | both | unknown or destroyed world |
| `clock_backwards` | director | refused |
| `consumer_not_quiesced` | director | `await_consumer` timed out |
| `profile_rate_exceeded` | director | emission rate exceeds the sink's cap |
| `truth_sealed` | director | reveal refused on an open run |
| `run_unblinded` | director | scoring refused; the run is stamped |
| `unknown_effector` | operator | not declared by the loaded domain |
| `invalid_args` | operator | failed the effector's declared schema |
| `missing_command_id` | operator | required idempotency key absent |
| `capability_denied` | operator | bad or absent token |
| `effector_refused` | operator | the effector declined — generic; see §4.5 |
| `interlock_refused` | operator | an independent safety system declined |
| `not_implemented` | both | F3/FMU dynamics, broker sink, binary adapter encodings |

`effector_refused` is intentionally uninformative. A helpful message — "cannot start
aerator, the oxygen probe is fouled" — hands the diagnosis to the system under test. The
detail goes to the director's log, where the harness can read it and the consumer cannot.
