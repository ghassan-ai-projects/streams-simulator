# Product strategy — from trustworthy instrument to world-class product

- **Date**: 2026-10-11
- **Status**: assessment and plan; **no code, schema or domain artifact changed by this folder**
- **Question**: what would make Streams Simulator a world-class product, not only a well-built instrument?
- **Evidence base**: repository state at `main` (83eb163), `go build/vet/test ./...` green,
  `streamsim catalog coverage`, the hardening program outcome, `docs/design/*`, and the
  untracked [`../simulation-ideas/`](../simulation-ideas/README.md) portfolio and its
  [product-proof review](../simulation-ideas/PRODUCT-PROOF-REVIEW.md).

## Verdict in one paragraph

As an **instrument** the project is already in the top tier: deterministic, replayable,
sealed ground truth, analytic oracles, a delivery ledger, enforced architecture gates and a
hardening program that closed the High-severity oracle defects. As a **product** it is
unproven: it cannot yet say *what a run cost*, *how a consumer compares with a strong
baseline*, or *which result a skeptical buyer should believe*. World-class means closing
that gap, in this order: **trust → outcomes → baselines → proof scenarios → reach**.

## Documents

| File | Purpose |
| --- | --- |
| [ASSESSMENT.md](ASSESSMENT.md) | Evidence-based rating by dimension, with what was verified and what was not |
| [USE-CASES.md](USE-CASES.md) | Three use cases that work now, three that need an upgrade, with research evidence and two reproduced defects |
| [GAPS.md](GAPS.md) | What separates this from world-class: trust, proof, coverage, reach, operability |
| [ROADMAP.md](ROADMAP.md) | Seven phases with deliverables, exit criteria, sizing, and what *not* to build |
| [DESIGN-outcome-economics.md](DESIGN-outcome-economics.md) | Design change proposal: loss, action cost, unsafe calls, regret, no-headroom audit |
| [DESIGN-baseline-ladder.md](DESIGN-baseline-ladder.md) | Design proposal: comparative proof protocol, fairness rules, statistics, held-out seeds |

## Correction to an earlier assessment

A chat assessment on 2026-10-11 listed D-01…D-06 from `DEFERRED.md` as likely still open.
That was wrong. The hardening program
([`../refactoring/hardening-20261010/README.md`](../refactoring/hardening-20261010/README.md))
records D-01…D-13 and most others as fixed with regression tests, and a code check confirms:
`verify` exits non-zero on mismatch (`internal/cli/internal/app/replay.go`), offline `score`
replays the artifact to recover effector calls, `FirstDivergence` is a record index, the
safe-stop command id carries a counter, and `rememberExecution` clears `Disconnect` and
`Duplicate`. The register in `DEFERRED.md` is stale relative to the code and should be
reconciled (ROADMAP phase 0).

## Decisions this folder asks the owner to take

1. Adopt **outcome economics** as the next design change ([design](DESIGN-outcome-economics.md)).
2. Decide where the **baseline harness** lives: in-repo reference policies plus an external
   bench repo for product and LLM baselines ([design](DESIGN-baseline-ladder.md) §3).
3. Freeze a **contract v1.0 policy** and a release channel (ROADMAP phase 6).
4. Pause new idea documents and new domains until phases 2–3 produce a scorecard.
