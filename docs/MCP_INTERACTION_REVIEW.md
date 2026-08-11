# Independent MCP Interaction Review

Status: external review — findings from live interaction
Date: 2026-08-11
Author: OpenClaw orchestrator (independent, not a code author)
Scope: run the binary, drive it end-to-end over MCP, validate the core claims against behaviour

> This is a hands-on review, not a code audit. Every finding below comes from actually
> launching `bin/streams-simulator`, connecting to its MCP server in the `director` role
> over stdio, and driving the documented interaction surface. No source code was modified.

---

## 1. What was exercised

| Step | Method | Result |
|---|---|---|
| Build | `go build ./...` (Go 1.26.5 darwin/arm64) | ✅ clean |
| Test suite | `go test ./...` | ✅ all packages pass |
| MCP handshake | `initialize` / `notifications/initialized` (protocol 2025-11-25) | ✅ server responds, advertises 24 tools + 1 resource |
| Catalog | `sim.catalog.list`, `sim.catalog.coverage`, `sim.catalog.describe`, `sim://catalog` resource | ✅ 6 domains, 12-axis matrix, 13 "thin" axes reported |
| World lifecycle | `sim.world.create` / `sim.world.describe` / `sim.world.destroy` | ✅ create (8 pumps, capability token, world_digest), describe, destroy |
| Clock | `sim.clock.advance` / `sim.clock.state` | ✅ tested; two quirks, see §4 |
| Faults | `sim.fault.inject` | ✅ `f-0` injected, effect applied on next advance |
| Evidence | file sink write | ✅ 10 events recovered after destroy (flush) |
| Determinism | two `streamsim run` invocations, same seed | ✅ same `run_id`, same `trace_digest`, **byte-identical** 5269-line traces |
| Replay verify | `streamsim verify <run.json>` | ✅ `matches: true`, `version_match: true` |

The single most important claim — byte-deterministic, verifiable reproduction — **holds under
direct test**. Two independent runs with `seed 42` produced identical `run_id`
(`r-2wehlpvbfa06r`), identical `trace_digest`
(`sha256:916aa26d…d8a770`), and byte-equal trace files. `verify` confirms the artifact
reproduces.

---

## 2. Management via MCP, evidence on the sink — confirmed working

The architectural premise (see `research/TRANSPORT_ANALYSIS.md` and `design/MCP_SURFACE.md`)
is that MCP carries *control* and the event stream leaves on a *sink*. This was validated
end-to-end:

1. `sim.world.create` → world `w-3`, domain `rotating-machinery`, `seed 9`,
   `sink=file`, `sink_target=/tmp/streamsim-e3.jsonl`.
2. `sim.clock.advance` twice → exactly 10 events (T+19s then T+19.3s).
3. `sim.world.destroy` → flushed the buffered file sink to disk.
4. The trace was read from the sink, not from MCP.

The 10 emitted events (healthy baseline, ~30 A motor current, gaussian σ=0.3 noise):

```
1.  pump-8  motor_current = 29.4 A @ 00:00:09.165
2.  pump-1  motor_current = 30.2 A @ 00:00:09.589
3.  pump-2  motor_current = 30.3 A @ 00:00:09.823
4.  pump-4  motor_current = 30.4 A @ 00:00:09.884
5.  pump-6  motor_current = 30.0 A @ 00:00:09.968
6.  pump-7  motor_current = 30.1 A @ 00:00:09.973
7.  pump-3  motor_current = 29.9 A @ 00:00:10.341
8.  pump-5  motor_current = 30.2 A @ 00:00:10.939
9.  pump-4  motor_current = 29.9 A @ 00:00:18.953
10. pump-6  motor_current = 30.0 A @ 00:00:19.070
```

---

## 3. Strong points observed

- **Determinism is real.** Independently verified byte-equivalence of two same-seed runs and
  a passing `verify`. This is the make-or-break property for a test instrument and it holds.
- **Self-aware coverage reporting.** `sim.catalog.coverage` does not present a falsely
  complete matrix; it explicitly flags 13 "thin" axes (e.g. `rate=torrent(1)`,
  `lateness=gross(1)`, `consequence=safety(1)`). Honest instrument self-description is rare.
- **Domain data quality.** `sim.catalog.describe(rotating-machinery)` returns 10 state vars,
  7 channels with noise/link-delay/jitter models, 7 fault types with counterfactual outcomes
  and deadlines, and 3 effectors with idempotency windows and ack-failure modes. Genuinely
  production-grade modeling, not a toy.
- **Role/tool separation is observable.** The `director` role advertises only director tools;
  operator-only surface (nameplate/effector/consumer.report) is not present, matching the
  separation-of-roles design.

---

## 4. Findings / issues (no code changed)

These are the friction points an MCP-first consumer will actually hit. They are integration-
and ergonomics-level, not correctness-of-the-core.

### 4.1 Documented MCP surface ≠ implemented tool list
`docs/design/MCP_SURFACE.md` §3 lists `sim.trace.export`, `sim.adapter.verify`, and
`sim.clock.run`. **None are present** in the live `tools/list` (calling `sim.trace.export`
returned `unknown tool`). The docs are aspirational; the binary is what it is. A consumer
dictated by the docs will fail at the first missing tool.

### 4.2 Tool input schemas are effectively open
All 24 tools advertise `inputSchema: {type: object, additionalProperties: true}` with no
declared argument shape. There is no machine-readable guidance for the exact `sink`,
`adapter`, `sink_target`, or time arguments. Concretely: my first `world.create` passed
`sink` as an object `{"type":"file","path":…}` and it was silently accepted but did not
persist; the correct form is flat `sink: "file"` + `sink_target: "/path"`. The only way to
discover this was reading `internal/mcp/director.go`. Tightening the JSON-schema on each
tool is the highest-value ergonomics fix.

### 4.3 File sink is buffered; flushes only on world destroy / run end
Events are counted (`emitted`) as they are generated, but the `file` sink only writes to
disk on `Close()`, which happens at `world.destroy` (→ `Run.End`). Advancing the clock does
not produce on-disk evidence. This is consistent with the trace-digest/byte-reproducibility
design, but it is surprising: after an `advance` reported `emitted: 11`, the sink file was
empty until the world was destroyed. Worth documenting explicitly for consumers so they do
not poll an empty file mid-run.

### 4.4 `sim.clock.advance` reports `emitted` as the per-call delta, not cumulative
Advancing from T to T+Δt returns the number of events emitted *by that call*, and repeated
advances are not additive in the response. To hit exactly 10 events I advanced to T+19s
(`emitted: 9`) then to T+19.3s (`emitted: 1`) and summed mentally. A consumer that treats
the response as a cumulative count will over/undershoot the trace length.

### 4.5 Time-base inconsistency and a confusing error message
When driving via MCP, `world.create` accepted but did not honor my `start_time` on first
attempt and produced a world whose clock read `1970-01-01`, while `sim.clock.advance` uses
`to_ns` as an absolute epoch-0-ns value. A naive `by_ns`-style call failed with
`clock_backwards: run: advance: world: clock would move backwards (1767225600000000000 -> 0)`
— a message that reads as if the clock would move backwards by 56 years, when the real cause
was passing `by_ns` (a key the handler does not read) instead of `to_ns`. The CLI `run`
defaults to `-start-time 2026-01-01` while the MCP path can land on epoch-0; the two
surfaces disagree on the default time base.

---

## 5. Bottom line

A serious, well-conceived test instrument whose **core engineering claim (byte-deterministic,
verifiable replay) genuinely holds up** under direct, independent test. The friction lives in
the **MCP integration layer** — underspecified tool schemas, docs–code drift on the tool
surface, a buffered sink whose flush point is easy to miss, and a time-base inconsistency
with a misleading error — not in the simulation core.

If the goal is MCP-first consumer integration (which `docs/design/MCP_SURFACE.md` clearly
states), the recommended next steps, in order of value:

1. Declare real JSON-schemas on every MCP tool (`internal/mcp`) so `additionalProperties`
   goes away and argument shapes are machine-checkable.
2. Reconcile `docs/design/MCP_SURFACE.md` §3.5 with the implemented tool list — either ship
   the missing tools or mark them explicitly as not-yet-implemented.
3. Document (or change) the file-sink flush semantics so consumers know evidence lands only
   on world destroy / run end.
4. Document that `sim.clock.advance.emitted` is a per-call delta, and align the MCP time
   base with the CLI default (or surface a clearer error than `clock_backwards`).

No source code was modified during this review.
