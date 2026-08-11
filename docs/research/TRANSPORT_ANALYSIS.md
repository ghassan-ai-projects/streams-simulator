# Output and Transport Design

Status: research input to the Streams Simulator design
Date: 2026-08-11
Question: the simulator is MCP-first. What does MCP carry, and how does evidence reach a
consumer?

## 1. Two decisions

**MCP manages the simulator. It does not transmit the streams.**

MCP names, seeds, steps, perturbs, actuates and audits a run. Evidence leaves on a
data-plane sink. §3 is why.

**The simulator emits its own format. Consumers are adapters.**

Evidence is emitted as `sim-event-v0.1` — the simulator's native envelope, which owes
nothing to any consumer — and projected into a consumer's wire format by a declarative
[output adapter](../contracts/output-adapter-v0.1.schema.json). §5 is why, and it is the
decision that makes this a product rather than a test rig for one runtime.

## 2. What the MCP protocol actually offers

From the transport specification (`2025-11-25`, unchanged in these areas by `2026-07-28`):

| Mechanism | Shape | Server→client push? |
|---|---|---|
| stdio | newline-delimited JSON-RPC; **messages must not contain embedded newlines** | yes, unsolicited allowed |
| Streamable HTTP, POST | one JSON-RPC message per POST; server answers with `application/json` *or* opens an SSE stream | only within the originating request |
| Streamable HTTP, GET | client opens a listening SSE stream; server may push unrelated requests and notifications | yes |

Three properties of that GET stream decide the question:

1. **"The server MAY close the SSE stream at any time."** Closure is normal, not an error.
2. **"It MUST NOT broadcast the same message across multiple streams."** Each message
   lives on exactly one stream; if that stream dies the message is gone unless the server
   implemented resumability.
3. **Resumability is a MAY.** `Last-Event-ID` replay is per-stream, optional, and the
   spec's own words for what it addresses are "messages that *might otherwise be lost*."

There is also no flow control. JSON-RPC notifications are fire-and-forget; a client that
cannot keep up has no way to say so.

## 3. Why MCP cannot be the evidence plane

Four independent reasons. Any one is sufficient.

### 3.1 Best-effort delivery destroys what the instrument measures

A large share of the simulator's value is in *absence*: a producer going quiet, a channel
dropping out, a heartbeat stopping. Those are the hardest conditions for any stream
processor and the ones worth testing hardest.

A transport where a dropped message is a specification-blessed outcome makes silence
ambiguous forever. The consumer cannot distinguish *the pump stopped reporting* from *the
SSE stream closed between reconnects* — and neither can the simulator, after the fact. If
the delivery layer can lose events, every absence finding in every scenario carries an
asterisk that can never be resolved.

The instrument would be injecting its own worst failure mode into itself, undeclared.

### 3.2 Push has no back-pressure, and the simulator is a firehose by design

Several catalog domains emit tens of thousands of events per second in bursts.
Unsolicited notifications give the producer no signal to slow down. The queue grows in the
SSE writer, the socket buffer, or the client's parser — all places where overflow is
invisible and the resulting loss looks exactly like §3.1.

### 3.3 Timing becomes a function of TCP, and determinism is the whole point

A run must be reproducible from a seed. The moment delivery time depends on socket
scheduling, SSE reconnect backoff and the `retry` field, observed arrival order stops
being a function of the seed. Any consumer that derives ages, durations, cooldowns or
budgets from receipt time then produces a different history every run, and the
golden-comparison gate cannot exist.

### 3.4 The consumer of MCP is an LLM host, not a stream processor

An MCP tool result is content flowing toward a capability boundary designed for model
context — catalog epochs, redaction, attribution, bounded output. Every one of those is
correct for a tool result and absurd per sensor reading.

The cost is not just overhead. It puts raw producer-controlled strings one hop from a
prompt, which is precisely the property the injection probe exists to test. Routing
evidence through the same channel would make the probe untestable by construction.

### 3.5 Throughput, for completeness

A native event is ~200 bytes of JSON; wrapped in a JSON-RPC notification and SSE framing
it is ~320 bytes and one parse. At 10k eps that is ~3.2 MB/s — survivable, and entirely
beside the point next to §3.1–3.4. Nobody should reject MCP here on throughput. Reject it
on delivery semantics.

## 4. What MCP is genuinely excellent at

Everything that is *not* an event is naturally request/response, low rate, and much better
for being agent-callable:

| Operation | Why MCP fits |
|---|---|
| list domains, describe one | discovery; a resource read |
| create a world, seed 42, 8 entities | one call, returns a digest |
| advance the clock six hours | one call, returns counts and the new horizon |
| inject a fault at T+2h | one call, parameterized, logged |
| perturb 5% of delivery for an hour | one call |
| **invoke an effector** — called *by the consumer* | request/response with an ack, an idempotency key, and typed failure. The textbook shape. |
| export, reveal truth, score | one call each |

Note the direction reversal in row 6. Evidence flows out on a sink; commands flow back in
as tool calls. That is the shape that closes the loop, and it is where the MCP-first idea
pays off.

## 5. Why the simulator emits its own format

The tempting shortcut is to emit the first consumer's format natively. It is fewer moving
parts on day one and it is wrong, for three reasons.

**It embeds one consumer's decisions in the core.** That consumer's field names, its
identifier alphabet, its rule about who assigns receipt timestamps, its record framing,
its config header — all of that becomes simulator behaviour. A second consumer then
arrives and either gets a special case in the binary or is told to parse the first
consumer's format.

**It corrupts the domain model upward.** Once the core speaks a consumer's envelope,
domain specs start being written to satisfy it — identifier lengths chosen to fit its
alphabet, channels named to match its input declarations. The simulator's own model stops
being about the world and starts being about the client.

**It makes the consumer's contract untestable.** If the simulator only ever emits one
format, nothing proves that format is a *projection* rather than an *assumption*. An
adapter with a golden fixture and a schema check proves it.

So: native envelope, declarative adapters, adapters are data. This is the same injection
invariant the domain catalog holds, applied to output — the binary contains no
consumer-specific behaviour, and adding a consumer is a file, not a release.

Deliberate limit: v0.1 adapters cover text encodings (JSONL, JSON array, line, CSV).
Binary framings such as protobuf or Sparkplug B cannot be expressed declaratively and are
out of scope rather than faked with a plugin, because a plugin is consumer-specific code
by another name.

## 6. The transport profiles

A profile is a per-run parameter. The **native event is identical across all of them**;
only delivery differs. The adapter is chosen independently of the profile.

| Profile | Mechanism | Determinism | Practical ceiling | Use |
|---|---|---|---|---|
| **`inproc`** | Go callback, no serialization | total | — | simulator self-tests; an embedded consumer |
| **`file`** | adapter output written to a file | total, byte-reproducible | offline, unbounded | **default.** golden fixtures, graded suites, CI |
| **`http-push`** | one HTTP POST per record to a configured endpoint | reproducible under a stepped clock; not under a wall clock | ~2–5k eps loopback, to be measured | live-path testing, crash and restart, soak |
| **`broker`** | MQTT or Kafka | at-least-once | high | deferred until a consumer needs it |

`file` is the default and the workhorse: the only profile that is both high-rate and
totally deterministic.

### 6.1 Receipt timestamps, per profile

The simulator models a per-event link delay and sets `observed_time = event_time + delay`.
What a consumer does with that is the consumer's business, and it differs:

| Profile | How the link-delay model is realized |
|---|---|
| `file`, `inproc` | `observed_time` is carried in the record. Exact; the model is fully expressed. |
| `http-push` + stepped clock | Consumers that assign their own receipt time on ingest cannot be told one. So the simulator **withholds the POST** until the shared stepped clock reaches `observed_time`. Exact, because both sides step the same clock. |
| `http-push` + wall clock | The delay is approximate and differs run to run. **Not reproducible.** The run is stamped `reproducible: false` and hash-comparison metrics are refused for it. |

The wall-clock row is a real limitation, not a defect to engineer away. It is what "live"
means, and every report using it says so.

Deliberate violations are test material: the simulator can emit `observed_time <
event_time` (a producer with a skewed clock) and non-monotonic sequences. A consumer that
silently accepts those has a finding against it.

## 7. What this design does *not* require of any consumer

Stated as a contract, because it is the product claim:

- no change to the consumer's ingress;
- no simulator-specific code, library, or dependency in the consumer;
- no exposure of the consumer's database, logs, or internals;
- no shared process, language, or runtime.

A consumer integrates by doing three things, all of which it can already do:

1. **Read** the adapter's output on whichever profile suits it.
2. **Optionally call** effector tools over MCP, if it wants the loop closed.
3. **Optionally submit** a [consumer verdict](../contracts/consumer-verdict-v0.1.schema.json)
   — a neutral statement of what it concluded — if it wants to be scored.

All three are optional except the first. A consumer that only reads traces gets
deterministic, labelled, perturbed evidence and nothing else, which is already most of the
value.

## 8. Keeping the answer sealed

A consumer that can reach the control plane can read the ground truth. The evaluation then
measures nothing, and the dangerous part is that it will *look excellent*.

**One MCP server, two roles, disjoint capability sets.**

| Role | Sees | Held by |
|---|---|---|
| `director` | catalog, world lifecycle, clock, faults, perturbations, truth, export, scoring | the harness or a human operator |
| `operator` | effector list and invoke, static nameplate, verdict submission | the consumer under test |

Release-blocking invariant: **everything reachable from the `operator` role must be
computable from the delivered evidence alone.**

Two controls enforce it. A **compile-time separation** — the operator role's handlers are
constructed from a view struct with no reference to the truth store, the fault registry,
the perturbation log or hidden state, so there is no field to leak. And a **prefix
indistinguishability test**: build two worlds differing only in injected fault, advance
both to the last instant their delivered evidence is identical, replay the same operator
calls against each, and assert every response is byte-identical and every latency falls in
the same padded bucket. That catches leaks through error strings, field ordering, map
iteration and timing — none of which a checklist finds.

## 9. The shape, in one picture

```text
        ┌───────────────── MCP: director role ─────────────────┐
 harness│ catalog · world · clock · faults · perturb · truth    │
        └──────────────────────┬───────────────────────────────┘
                               │ command log
                               ▼
                     ┌───────────────────┐
                     │    world core     │  seeded · discrete-event
                     └────┬─────────┬────┘
      native sim events   │         │  effects, with time constants
                          ▼         │
                  ┌───────────────┐ │
                  │ perturbation  │ │
                  └───────┬───────┘ │
                          ▼         │
                  ┌───────────────┐ │
                  │    adapter    │ │   <- data, one per consumer
                  └───────┬───────┘ │
                          ▼         │
              file · http-push · inproc
                          ▼         │
                  ╔═══════════════╗ │
                  ║   consumer    ║ │
                  ╚═══════╤═══════╝ │
                          │         │
                          └─────────┘
                   MCP: operator role
             effector invoke · verdict submit
```

Evidence flows out through an adapter. Commands and verdicts flow back in over MCP under
the operator role. Control flows in under the director role, which the consumer never
holds.

## Sources

- [MCP Transports specification (2025-11-25)](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports)
- [Exploring the Future of MCP Transports](https://blog.modelcontextprotocol.io/posts/2025-12-19-mcp-transport-future/)
- [MCP architecture](https://modelcontextprotocol.io/docs/learn/architecture)
