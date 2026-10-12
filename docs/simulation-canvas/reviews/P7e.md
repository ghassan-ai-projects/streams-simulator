# P7e review record — pre-UI review

- **Reviewed**: `main` at `79d4b85` (P0–P7d), 2026-10-13.
- **Reviewers**: four independent Sonnet subagents (world core and oracles; determinism, replay and run records; truth
  boundary; architecture, tests, gate and documentation) and a cleanliness hunt.
- **Owner decision**: this phase changes architecture and modularity only; behaviour findings are registered for P7f.

## Fixed in P7e

| Finding | Fix |
|---|---|
| Evidence files (`ledger.jsonl`, `world_state_history.jsonl`, the trace) written 0644; `run.json`, verdicts, manifests and suites written in place | `kernel`: every written file is owner-only and atomic |
| A second plant wire host's liveness probe ended the first host | The host keeps accepting until a connection speaks |
| The node's handshake used the barrier cap | A handshake waits at least 10 s |
| An operator could date `sim.effector.invoke` (`at_ns`) and probe hidden state at any time | The operator's call acts at the world clock; `at_ns` is not in the operator schema |
| `world_id` and `run_id` were short hashes of the domain and the seed, so the seed was recoverable by brute force | Both ids carry no seed; the world's random streams keep their private seed-derived key, so every world's numbers are unchanged |

## Registered for P7f

R7e-01 to R7e-12 in [FINDINGS.md](../FINDINGS.md): world snapshot depth, stochastic walks, equal-instant ties, effect and
bound edge cases, SimVersion and the pin, run-record digests and ledger reuse, the quiescence clamp, journal framing lines,
operator descriptions, the report's truth gate, missing tests, and findings phases.
