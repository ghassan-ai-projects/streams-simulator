# Goal — Simulation canvas

- **Date set**: 2026-10-12
- **Owner direction**: build the simulation canvas as a production-ready part of a critical
  test instrument: no workarounds, no shortcuts, phase by phase, each phase independently
  reviewed and committed. Agentic Stream work follows this goal.
- **Design**: [canvas-design.md](canvas-design.md) · **Bar**: [QUALITY_BAR.md](../../governance/CANVAS_QUALITY_BAR.md) ·
  **Phases**: [PLAN.md](../../governance/CANVAS_PLAN.md) · **Open findings**: [FINDINGS.md](../../governance/FINDINGS.md) ·
  **Design review**: [REVIEW-2026-10-12.md](REVIEW-2026-10-12.md)

## Outcome

A person or a coding agent can **build** a simulated plant, **run** it with a consumer
connected, **watch** it live, **disturb** it, and **review** what happened with evidence the
binary can prove, from a visual canvas or from files and commands alone, with the same result
either way.

## Done means

Each item names the phase whose exit criterion proves it.

1. **Build** (P11, P12). A world is authored in the canvas builder (kinds, instances, links,
   behaviour, sensors, actions, faults, scenarios) and saved as domain JSON plus a layout
   sidecar. The same world authored by editing JSON and formatted by `streamsim domain format`
   is byte-identical. Every edit is validated by the binary with errors pinned to the element;
   preview runs work without an agent.
2. **Run and control** (P10). Runs start, pause, step, change speed and reset from the canvas
   through the director role. Clock advances, faults, environment changes and perturbations
   land in the command log and replay exactly; pause, speed and reset are pacing and lifecycle
   events recorded in the journal.
3. **Watch** (P3, P10). A run journal is written while the run is in progress; the canvas
   follows it live: plant view with live values and quality, timeline, wire traffic and party
   status from session records.
4. **Disturb** (P2, P10). Faults, environment changes and perturbations are injected mid-run
   from the canvas, recorded, and exported by the binary as a repeatable scenario file
   (`streamsim scenario from-run`) whose replay yields the same command-log digest.
5. **Review** (P4, P9). Any run folder plays back with no `streamsim` on the machine; the
   binary proves it (Verified, Not verified, Incomplete badge); the report shows timeline,
   actions, outcomes, protection trips, and true vs reported after sealing.
6. **Truth stays sealed** (P3, P10). In Evaluate mode no truth-bearing record (true state,
   injected faults, environment or perturbation commands) reaches any file a consumer or agent
   may read, any bridge payload or the page before sealing, proven by a canary test. Explore
   mode stamps the run unblinded and excludes it from scorecards, proven by a test.
7. **Agent-native** (P1, P10). Every canvas action has a documented CLI or MCP equivalent with
   structured JSON output, proven by a parity test over the bridge's action registry; every
   file the canvas reads or writes has a schema.
8. **Reference worlds** (P12). The factory utilities world (domain-spec v0.2) is built,
   validated, previewed and run over the plant wire against the reference host; it is
   delivered as "validated and run, headroom not measured" (headroom needs the agent system,
   next goal). The thermal bench world runs as a simulated twin whose event stream is compared
   with a recorded bench run.
9. **Plant wire, simulator side** (P7). The simulator is a plant-wire node with the contract's
   schemas and fixtures, proven against a minimal conformant reference host. The existing
   device path, vendored device contract and adapters stay intact and gated in every phase, so
   the real-world-sensor experiment keeps working until Agentic Stream migrates.
10. **Quality** (every phase, P13). Every gate in [QUALITY_BAR.md](../../governance/CANVAS_QUALITY_BAR.md) is green,
    every phase has an evidence file and an independent review with all findings resolved,
    no finding assigned to a finished phase is open, and documentation describes what exists.

### How the design's MVP acceptance maps here

canvas-design.md §10 lists two acceptance runs with Agentic Stream and Tamoz live. In this goal
both are met against the **reference host** (a minimal conformant plant-wire host shipped with
the simulator); the legs with the real agent system move to the next goal. The joined
real-world-sensor run (G1) stays in the sibling repositories, unchanged, until migration.

## Out of this goal

- **Next goal (Agentic Stream and gateway):** Agentic Stream as plant-wire host; the gateway's
  plant-wire side; the fan experiment over the plant wire; the factory with Tamoz + Agentic
  Stream; baseline headroom measurement and the benchmark governance in D7; scoring changes
  (REVIEW S1/S7 scoring side); gateway impact analysis, cross-repository build order and
  deployment topology (REVIEW M2–M4); customer-isolation extensions I1–I2.
- **Documented, not decided:** where the LLM runs and provider data terms (D6), written as an
  open product decision in P13.
- **Later:** wire authentication (required before any network carrier or customer
  deployment), authority-revocation test, counterfactual branching, run comparison, policy
  editor, run query command, agent skills, bundling the UI into the binary, a real
  Sparkplug/MQTT binding, a contract registry.
- **Dropped:** agent-involved plant emergency shutdown.
