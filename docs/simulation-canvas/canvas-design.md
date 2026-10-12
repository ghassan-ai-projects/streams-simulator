# Simulation canvas — design

- **Date**: 2026-10-12
- **Status**: design proposal; consolidates the canvas decisions taken with the owner on
  2026-10-12. **No code changed.**
- **Related**: [idea.md](idea.md) (product idea), [reference-scenario-factory-utilities.md](reference-scenario-factory-utilities.md)
  (acceptance scenario), [plant-wire-protocol.md](plant-wire-protocol.md) (wire between plant
  and host), [REVIEW-2026-10-12.md](REVIEW-2026-10-12.md) (open findings)
- **Compatibility**: none required (pre-release).
- **Goal and phases**: [GOAL.md](GOAL.md), [PLAN.md](PLAN.md). Where this document and PLAN.md
  differ on sequencing or acceptance, PLAN.md is current.

## 1. Purpose and principles

The canvas is where a person builds a simulated world, runs it with an agent system (Tamoz +
Agentic Stream) connected, watches it live, disturbs it, and reviews what happened.

1. **The binary and the files are the product; the canvas is one client.** Coding agents are
   another, equal client. The canvas has **no exclusive capability**: every action maps to a
   file edit, a CLI command or a director MCP tool (§9). If the canvas needs something no CLI
   command or tool offers, that logic goes into the simulator, not the canvas.
2. **The canvas computes nothing the binary owns.** Validation, simulation, truth, scoring,
   replay and verification come only from `streamsim`.
3. **Files for definitions, commands for live changes, outputs read-only.** Domains, layouts and
   scenario files are edited as JSON. A running world changes only through director commands, so
   every change lands in the command log and replays exactly. Run artifacts are evidence and are
   never edited.
4. **The canvas observes and directs; it never acts as the host.** It controls the *world*
   (clock, faults, perturbations) through the director role. It never sends plant-wire commands
   to devices: that is Agentic Stream's job, under its policy.
5. **Sealed truth stays sealed** (§7).

## 2. Placement and dependency rules

| Rule | Detail |
|---|---|
| Location | Top-level `canvas/` in this repository, with its own toolchain |
| One-way dependency | Go code never imports from `canvas/`; the canvas reaches the simulator only through files, the CLI and director MCP. Enforced by a check in `test/architecture` |
| Dependencies | All JavaScript dependencies live under `canvas/`; `go.mod` is unchanged; each dependency pinned and justified |
| Gates | `make canvas-check`; `make ci-check` stays Go-only or calls it when `canvas/` changed |
| Rules | `canvas/` gets its own section in `AGENTS.md` for TypeScript conventions |
| Shipping | First a local dev server; bundling the built UI into `streamsim` (`go:embed`, `streamsim canvas`) is deferred until the viewer has proven itself |

## 3. Files the canvas reads and writes

| File | Owner | Canvas | Notes |
|---|---|---|---|
| `domains/*.domain.json` (or a customer workspace) | Simulator | Read, write (builder) | Validated only by `streamsim domain validate --json` |
| `canvas/layouts/<domain-id>.layout.json` | Canvas | Read, write | Sidecar, §4 |
| `canvas/connections.json` | Canvas | Read | External parties and health URLs, §6 |
| Scenario files (`scenarios/*.json`) | Simulator | Read, write | Scripted runs, §5.4 |
| Run folder: `run.json`, journal (§7.1), `ledger.jsonl`, trace, state history, verdict, report | Simulator | **Read only** | Evidence |
| `adapters/*.adapter.json` | Simulator | Read | Choose and inspect output formats |

### 3.1 Layout sidecar

- Lives in `canvas/layouts/`, **not** in `domains/`: the domain loader treats every `*.json` in
  that folder as a domain and fails the whole catalog on one bad file.
- References domain elements by their existing names (kinds, instances, states, channels,
  effectors, faults, links); copies no domain content.
- **Optional**: without a layout, the canvas auto-places elements.
- **Drift-tolerant**: new elements are auto-placed; entries for removed elements are dropped on
  the next save; a layout never blocks opening a domain.
- **Presentation only**: never part of the domain digest, the run artifact or the release
  manifest. Moving a box never changes a run.
- May declare **visual groups** ("Cooling unit") that mean nothing to the simulator.

### 3.2 Workspaces and customers

The canvas opens a **workspace**: a folder with domains, layouts, scenarios and runs. The
repository holds only generic, synthetic example worlds. A customer workspace lives outside every
repository and is passed by path (`CUSTOMER_WORKSPACE_DIR` in a git-ignored `.env`). Nothing from
a customer workspace is ever written into the repository (see plant-wire §16).

## 4. Views

Domain-spec v0.2 (see the scenario doc §7) adds entity kinds, instances and typed links, so the
canvas has two complementary views of the same domain:

| View | Nodes | Edges | Used for |
|---|---|---|---|
| **Plant view** | Instances (`M1`, `AC1`, `P2`, system entities such as `air_header`, `power_bus`), external parties (Agentic Stream, Tamoz) | Typed links (topology); live wire traffic | Understanding and watching the plant |
| **Kind view** | Inside one kind: states, sensors (channels), actions (effectors), faults | Dynamics inputs (state → state), `observes` (sensor → state), effects (action → state), `affects` (fault → state), detector (fault → sensor) | Building behaviour (cause and effect) |

A kind is a palette item; an instance is a placed node. Selecting an instance shows its kind's
definition plus its overrides (add or narrow only).

## 5. World builder (MVP)

### 5.1 Panels, mapped one-to-one to the domain schema

| Panel | Edits |
|---|---|
| Overview | id, version, title, description, stresses, axes (dropdowns from the schema's allowed values) |
| Kinds | Kind definitions: states, sensors, actions, faults; add-or-narrow override rule enforced |
| Instances and links | Placing instances of kinds, system entities, typed links |
| State | name, unit, initial value, min/max, description |
| Behaviour | Dynamics from the fixed vocabulary only (`lag`, `store`, `flow`, `regime`, `trip`, `schedule`, `peak`, …); never free formulas |
| Sensors | resolution, range, cadence, noise, drift, availability, link delay, absence |
| Actions | argument schema, risk class, ack latency, failure modes (`ok`, `reject`, `partial`, `silent_no_effect`; probabilities sum to 1), effect, dead time, idempotency window, preconditions |
| Faults | onset, affected states, detector, SNR thresholds, expected action, deadline, counterfactual, negative class |
| Scenarios | profiles: instance counts, duration, fault weights, setup |
| Ground truth | negative and pre-degraded fractions, minimum scenarios (advanced, collapsed) |

### 5.2 Builder tools, all backed by the binary

| Tool | Backed by |
|---|---|
| Validate as you edit, errors pinned to the node (JSON pointer) | `streamsim domain validate --json` |
| Preview: short run without an agent, plotting state trajectories | `streamsim run` (short duration) |
| Fault check: is a fault too easy or impossible to detect | Trivial-baseline audit (`sim.scenario.audit` / `suite`) |
| Start from a template | Clone a shipped or workspace domain |
| Diff before save | JSON diff of the domain file |

### 5.3 Honesty rule

The builder exposes exactly what the schema can express. Anything else is shown as "not
expressible yet" and never stored as a UI-only field. Gaps become evidence for schema changes.

### 5.4 Scenario file

A JSON file for repeatable scripted runs: domain, profile, seed, duration, timed faults, timed
actions, perturbations. It replaces long `run --fault … --effector …` flag lists, gives coding
agents a pure edit-and-run path, and lets the fault panel save a manual session as a repeatable
test.

## 6. Run, watch and control

### 6.1 Bridge

Browsers cannot speak stdio MCP, so a small local backend sits between them. The browser ⇄
bridge protocol, the bridge's language and its threat model are decided in the P8a/P8b specs.

```
Browser UI ⇄ local HTTP + event stream ⇄ canvas bridge ⇄ stdio MCP ⇄ streamsim mcp --role director
                              │
                              ├─ follows the run journal files (§7.1)
                              └─ polls health URLs from connections.json
```

- Launches and talks to `streamsim mcp --role director`; relays commands.
- Owns the **pacing timer**: the world clock is virtual and moves only on `sim.clock.advance`.
- Follows the journal files and pushes new records to the UI.
- Holds no simulation logic; anything it cannot do through a CLI command or MCP tool goes into
  the simulator.

### 6.2 Control bar

| Control | Implementation |
|---|---|
| Play at N× | Advance N simulated seconds per real second |
| Pause | Stop advancing |
| Step | One advance |
| Reset | Destroy and recreate the world with the same seed |
| Status | Elapsed simulated time, speed, faults active, protection trips |

Pacing respects the simulator's quiescence barrier, so a slow agent is not outrun (clock mode is
an open review item, REVIEW P12).

### 6.3 Live view

- Plant view with live values; edges light up from wire traffic (observations out, commands in).
- Timeline of observations, commands, receipts and results, injected faults and perturbations.
- **Fault panel**: inject or clear faults (`sim.fault.inject` / `clear`), environment changes
  (`sim.env.inject`), perturbations (`sim.perturb.apply` / `clear`). All go through the director,
  land in the command log, and can be saved as a scenario file.
- In worlds that bind some devices to real hardware behind a gateway, director and fault commands
  affect **simulated devices only**; the environment marker in births (plant-wire) makes the
  distinction visible.

### 6.4 Party status

External parties (Agentic Stream, Tamoz) appear as boxes with three states, taken from records,
not configuration:

| State | Source |
|---|---|
| Up | Health URL in `connections.json` (Agentic Stream `/health/*`) |
| Connected | Plant-wire session state (births, `host_state`) |
| Active | Last command or data time; Tamoz's last episode as reported by Agentic Stream |

Tamoz has no health endpoint; its status comes from Agentic Stream, which dials it. In playback,
the boxes show the **recorded** states at the playback time, never today's health checks.

## 7. Truth, modes and evidence

### 7.1 Run journal

Today `trace.jsonl`, `world_state_history.jsonl` and `run.json` are written only at `End`. The
simulator gains a **run journal**, appended during the run:

| Journal content | Truth? |
|---|---|
| Plant-wire exchange as delivered (births, data, deaths, commands, receipts, results, host state) | No |
| Clock advances, pacing and lifecycle events | No |
| Injected faults, environment changes, perturbations | **Yes**: they are the labels the oracle scores against |
| Ledger records | No |
| True state history | **Yes** |

Truth-bearing records go to a sealed stream that consumers and agents cannot read before
sealing (or to an explicitly unblinded stream in Explore mode). In Evaluate mode the canvas
knows its own injected faults only from the commands it sent, never from a consumer-readable
file. The exact record list and read rules are fixed in the P3 spec.

The journal is general simulator evidence (also useful after a crash), not a UI feature. Its
schema is published with the other contracts.

### 7.2 Modes

| Mode | Live view shows | Truth | Scored |
|---|---|---|---|
| **Explore** | Observations, actions, injected faults **and true state** | Unblinded from the start; run stamped unblinded | No |
| **Evaluate** | Observations, actions, injected faults | Revealed after `sim.run.end` and sealing, shown in the report and playback | Yes |

This follows the existing design rule: revealing truth on an open run requires `unblind` and
excludes the run from scorecards. The same applies to coding agents: an agent that builds or tunes
the system under test reads only plant-wire records during a run and sees truth only after the
run is sealed.

### 7.3 Playback vs replay

| | Playback (canvas) | Replay / verify (binary) |
|---|---|---|
| What | Reads recorded files | Re-runs from seed and command log and compares |
| Needs the binary | No | Yes |
| Proves anything | No | Yes |
| Seek, scrub, backwards | Yes | No |

**Rule: play back files, prove them by replay.** Every run shows a badge: **Verified**
(`streamsim verify` passed), **Not verified**, or **Incomplete** (crashed run, playable up to its
end). A sealed run shows true vs reported side by side, with violation and trip moments marked.

### 7.4 Report

Timeline, actions and their outcomes, protection trips and boundary violations, diagnosis
results, true vs reported after sealing, and the system configuration used (Agentic Stream policy
and SituationSpec digests). Produced headless by the binary (`streamsim report <run-dir>`, JSON
and Markdown); the canvas renders it.

## 8. Agent-native operation

| Canvas action | Agent equivalent |
|---|---|
| Edit a world | Edit the domain JSON (and layout) |
| Validate | `streamsim domain validate --json` |
| Run, pause, step, speed | Director MCP (`sim.run.begin`/`end`, `sim.clock.advance`) |
| Inject a fault | `sim.fault.inject`, `sim.env.inject`, `sim.perturb.apply` |
| Watch live | Read the journal |
| Play back | Read the run folder |
| Verify | `streamsim verify` |
| Report | `streamsim report` |

Additions that make this real:

| # | Addition | MVP |
|---|---|---|
| A1 | Project `.mcp.json` registering a director subset that cannot reveal or unblind truth (decided in P10) | Yes |
| A2 | `--json` output on every CLI command; validation errors with a JSON pointer | Yes |
| A3 | JSON Schemas for journal, layout, connections and scenario files | Yes |
| A4 | Headless report (`streamsim report`) | Yes |
| A5 | Run query command ("first violation", "events between t1 and t2", "what the agent saw at t") | Later |
| A6 | Skills/prompts: author a domain, run and evaluate, investigate a failed run | Later |
| A7 | Canvas reloads on file change, so a person and a coding agent can work on the same files | Yes |

## 9. Quality bar (from the start)

| Area | Bar |
|---|---|
| Architecture | Layered: domain model (types from the JSON Schemas), application state, adapters (files, bridge, director client, journal follower), UI components. Dependency direction enforced by a tool in CI |
| Code | TypeScript strict; zero lint warnings; file and function size limits in the spirit of the Go gates; minimal, pinned, justified dependencies |
| Truth | No simulation, validation, scoring or verification logic in the canvas |
| Tests | Unit tests on the model; component tests; end-to-end tests against the real binary and the reference worlds; visual regression on key screens |
| UX | Design tokens, light and dark themes, full keyboard support, WCAG AA; a professional engineering tool |
| Performance | Budgets set from the reference scenario: smooth canvas at its size, live view keeps up with the journal's peak rate |
| Process | Same gates locally and in CI (`make canvas-check`) |

## 10. Components and MVP scope

| # | Component | MVP |
|---|---|---|
| Simulator | Run journal (§7.1) | Yes |
| Simulator | Plant-wire node with connection and session records | Yes |
| Simulator | `--json` CLI output, `report` command | Yes |
| Simulator | Domain-spec v0.2 (kinds, instances, links, new forms) | Yes |
| Canvas | Bridge | Yes |
| Canvas | Plant and kind views | Yes |
| Canvas | World builder (§5) | Yes |
| Canvas | Control bar, live view, fault panel, party status | Yes |
| Canvas | Playback with verification badge, report view | Yes |
| Canvas | Compare two runs side by side | Later |
| Canvas | Counterfactual branching (re-run from a command-log prefix with a different command) | Later |
| Canvas | Permission-policy editor next to the builder | Later |

**MVP acceptance** (in this goal met against the minimal reference host; the legs with Agentic
Stream and Tamoz live belong to the next goal, see GOAL.md)

1. The real-world-sensor thermal LED/fan experiment is visible end to end in the canvas
   (emulated rung), with Agentic Stream and Tamoz boxes live, and still passes its gates.
2. The factory utilities world is built in the canvas **and** by a coding agent editing JSON, both
   validate, preview, run live with Tamoz + Agentic Stream, take a fault injected mid-run, and play
   back with a Verified badge and a report.
3. All §9 gates are green.

## 11. Open questions

1. **Technology stack.** Decided: TypeScript, React + Vite + React Flow
   ([canvas-technology.md](canvas-technology.md)).
2. **Bridge language.** Decided: Go, `streamsim canvas`, HTTP JSON + Server-Sent Events
   ([canvas-technology.md](canvas-technology.md)); the rule in §6.1 holds.
3. **Clock mode with a live LLM host** (REVIEW P12): lockstep, real time, or virtual time with
   recorded agent latency.
4. **Where the plant-wire node sits in the simulator** (REVIEW M5), which the journal depends on.
