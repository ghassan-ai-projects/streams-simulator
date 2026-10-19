# Roadmap and open work

> Status: Maintainer backlog, not a delivery promise. Authority: current limitations and accepted decisions. Verified by: documentation review. Last verified: 2026-09-13.

This is a status map, not a promise of dates. The implementation plan and historical reviews remain in [`docs/`](../docs/); this page names the work that matters to users and maintainers now.

## Present

- Deterministic world, dynamics, faults, perturbations, adapters, sinks, run artifacts, replay, and verification.
- Director MCP surface and token-scoped operator endpoint.
- Reference consumer, truth sealing/reveal, scoring, suite generation, baseline audit, and release-manifest command.
- Test coverage for the nine non-negotiables, including analytic cross-check, ledger, quiescence, oracle isolation, injection, silent no-effect, and trivial-baseline controls.
- Physical devices ([guide](guides/physical-devices.md)): kind files and twins, kind capture from a board, the relay and mixed sessions, the field trace, the live control plane, node interrupts, the twin-versus-board report and the session policy allow-list, all tested against a scripted stand-in gateway.

## Physical devices: open work

Nothing below changes a contract unless it says so. The deferred items with their triggers are in [`governance/DEFERRED.md`](../governance/DEFERRED.md) and summarised in [limitations](limitations.md#physical-devices).

- **Outside this repository, owner-led.** A plant wire side for the Arduino gateway, a plant wire host that handles `local_control` refusals and carries its policy digest, an independent effect instrument on the bench, and the owner's bench checks for the physical exits of milestones M2 to M5. Handoff briefs for these are part of the working archive. Until the bench checks pass, no physical claim is made.
- **Independent confirmation of an effect.** The live effect stage accepts host evidence only for a metric of the commanded device (D-111), and a hardware-only session has no barrier to carry it (D-112). A bench instrument that is a separate device needs a design decision and a contract change before it can feed that stage.
- **Pacing of mixed sessions** to the wall clock (D-71) and streaming the simulated half of a recording beside the field trace (D-72), when a consumer must be shown a recorded mixed session end to end.
- **Twin report**: an initial state from the board (D-90), a coverage figure and a partial verdict for vacuous tolerances (D-95), and a board trend in Inspect (D-92).
- **Control plane**: serving controls between steps (D-81, D-106), cancelling a command at a gateway that exposes a way to (D-94), operator identity and confirmation if a bench has several operators (D-86).
- **Second proof**: a second board through the same steps, with no change to the host's workflows.

## Next maintenance work

- Extend the lightweight documentation check with local link/fragment and example validation while retaining human diagram review.
- Reconcile any future design/archive drift with the implemented package layout and file-artifact decision.
- Add a suite runner that records adapter selection and persists per-scenario action evidence and score output.
- Keep the support, conduct, changelog, compatibility, and release policies actionable as the hosting setup evolves.
- Generate a fresh release manifest and rerun the full gate from a clean, intentional commit at publication time.

## Design horizon

- Expand the data-defined domain catalog only when each new domain adds a distinct property vector and survives the trivial-baseline audit.
- Add adapters and external consumers without adding consumer knowledge to the binary.
- Improve evidence reporting without weakening sealed truth, replay, or delivery classification.

## Change policy

Every roadmap item must identify its affected contract, invariant, test evidence, and documentation page before implementation. “More domains” or “more scenarios” is not progress if it makes the instrument less trustworthy.

## Next reads

- [Limitations](limitations.md)
- [Design index](design/README.md)
- [Contributor quality](governance/quality.md)
