# Benchmark-validity summary

A benchmark can be reproducible and still be invalid. Streams Simulator therefore treats validity controls as product behavior.

## Threats to validity

- A one-line detector solves a scenario because the label is encoded in a constant or obvious field.
- A consumer is penalized for a transport miss it could not observe.
- The oracle leaks hidden truth through a role, error, timing, or resource channel.
- An action acknowledgement is scored as if it proved a physical effect.
- The author writes both a fault model and a consumer summary that is hand-fitted to it.

## Controls

The trivial-baseline audit excludes leaky scenarios. The delivery ledger separates transport from reasoning. Director/operator separation and prefix-indistinguishability tests protect the oracle. The `silent_no_effect` test checks action truth. The suite is data-defined and digest-bound for replay.

## Source

See [`docs/design/GROUND_TRUTH_AND_SCORING.md`](../../docs/design/GROUND_TRUTH_AND_SCORING.md) and [`docs/design/CRITICAL_REVIEW.md`](../../docs/design/CRITICAL_REVIEW.md). These are design and review records; current gate evidence is listed in [the invariants page](../architecture/invariants.md).

## Next reads

- [Benchmark methodology](../benchmark/methodology.md)
- [Nine non-negotiables](../architecture/invariants.md)
