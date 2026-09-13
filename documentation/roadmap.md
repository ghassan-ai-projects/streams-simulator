# Roadmap and open work

> Status: Maintainer backlog, not a delivery promise. Authority: current limitations and accepted decisions. Verified by: documentation review. Last verified: 2026-09-13.

This is a status map, not a promise of dates. The implementation plan and historical reviews remain in [`docs/`](../docs/); this page names the work that matters to users and maintainers now.

## Present

- Deterministic world, dynamics, faults, perturbations, adapters, sinks, run artifacts, replay, and verification.
- Director MCP surface and token-scoped operator endpoint.
- Reference consumer, truth sealing/reveal, scoring, suite generation, baseline audit, and release-manifest command.
- Test coverage for the nine non-negotiables, including analytic cross-check, ledger, quiescence, oracle isolation, injection, silent no-effect, and trivial-baseline controls.

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
