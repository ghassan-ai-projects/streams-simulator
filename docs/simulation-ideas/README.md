# Simulation ideas — candidate domain portfolio

- **Date**: 2026-10-05
- **Status**: documentation-only evaluation record — ten ideas scored against the portfolio quality bar; **none admitted**
- **Scope**: this folder only. No code, schema, or domain artifact was created or modified (`domains/` still holds exactly the eight shipped files).

## 1. Purpose

This folder is the evaluated portfolio of candidate **simulation ideas** for streams-simulator:
new data-defined domains — plus the few small, domain-agnostic capabilities some of them need —
proposed to extend the shipped eight-domain set. It records the full pipeline: project study,
four-angle research, consolidation, five-lens adversarial review, and scored improvement loops,
so a future admission decision can be made from evidence rather than taste.

It is an evaluation record, not a backlog commitment. On the recorded scores **no idea meets
the quality bar** (best: SI-02 `feeder-switching` at 80/100 against a bar of ≥85/100 total with
no dimension below 7/10), and none of the ten domain JSON artifacts exists. What survives is
the best-evidenced starting set for a future round, with every load-bearing feasibility claim
already checked against this repository and every unmet prerequisite named as a design change
first.

Two status words are kept strictly apart in this folder:

- **Met the bar** — a *scoring* outcome. Strictly, no idea's scores meet the bar. SI-02 and
  SI-09 carry the workflow's contrary "met" verdicts; both documents state the verdict as
  supplied **and** the arithmetic discrepancy against the stated bar rather than smoothing it
  over ([review-evidence.md](review-evidence.md) §4). SI-02's header originally read "idea
  accepted at 80/100"; it was corrected during this folder's documentation pass to separate
  the recorded verdict from admission.
- **Admitted** — a *portfolio* outcome. None is admitted: no domain artifact exists and
  nothing entered a graded suite.

## 2. Method

Every idea went through the same pipeline:

1. **Project study.** The repository was grounded first: `catalog list` and
   `catalog coverage` runs against the built binary, schema and runtime reads, and the G1–G9
   gate definitions in [QUALITY_BAR.md](../QUALITY_BAR.md). The measured thin and
   zero-coverage axis values are the demand signal the ideas answer to; the twelve-axis
   coverage model and the decoration rule are
   [DOMAIN_CATALOG.md](../design/DOMAIN_CATALOG.md).
2. **Four-angle research.** Four parallel angles, each combining external sources with repo
   grounding: (a) event-time semantics pain; (b) testing agentic and durable consumers;
   (c) new physical/operational worlds; (d) delivery-corruption frontiers. Findings, sources,
   and their limitations: [research-notes.md](research-notes.md).
3. **Consolidation.** A raw pool of 25 candidates across the four angles was consolidated
   into the 10 portfolio ideas. The pool, every cut candidate, and every recorded disposition
   (including the kill/absorb claims later withdrawn as unverifiable) are in
   [research-notes.md](research-notes.md).
4. **Five lens reviews.** Each idea was attacked by the same five lenses: determinism and
   analytic oracle; data-not-code expressibility; consumer testing value; delivery ledger and
   quiescence; novelty and portfolio balance. Per-idea verdicts and the portfolio-level lens
   notes: [review-evidence.md](review-evidence.md).
5. **Improvement loop.** Each idea was then revised against the quality bar for **at most 10
   rounds** (SI-07, SI-08 and SI-09 hit the cap; the others converged in 1–2 rounds). The
   portfolio quality bar: **total at least 85/100 across the seven dimensions, with no single
   dimension below 7/10**. The seven dimensions, each scored 0–10 and summed /70, then
   normalised to /100: `consumerValue`, `dataExpressibility`, `determinismOracle`,
   `ledgerQuiescence`, `novelty`, `sealedTruth`, `specificity`. This bar is distinct from the
   repository's G1–G9 completion bar in [QUALITY_BAR.md](../QUALITY_BAR.md); the feasibility
   sections of the idea documents address the gates per idea.
6. **Independent final check.** After the loop ended, an independent evaluation produced the
   recorded seven-dimension scores in each idea's *Quality bar result* section. Those scores —
   not the per-lens review scores, which grade the draft a lens reviewed — are the portfolio
   numbers in §3. The evaluator is the automated workflow that ran this study (its
   quality-bar evaluation and post-loop scoring step) — not a tool in this repository and not
   a human review board, and its runs are not reproducible from the tree. What is
   independently checkable is the code-level evidence each idea document records; see
   [review-evidence.md](review-evidence.md) §4 for the full provenance note.

## 3. Portfolio

| ID | Name | Score | Status | Document |
|---|---|---|---|---|
| SI-01 | `telemetry-estate` | 70/100 | not met | [SI-01 · telemetry-estate](ideas/SI-01-telemetry-estate.md) |
| SI-02 | `feeder-switching` | 80/100 | met as supplied¹ | [SI-02 · feeder-switching](ideas/SI-02-feeder-switching.md) |
| SI-03 | `district-metering` | 69/100 | not met | [SI-03 · district-metering](ideas/SI-03-district-metering.md) |
| SI-04 | `ami-headend-backpressure` | 73/100 | not met | [SI-04 · ami-headend-backpressure](ideas/SI-04-ami-headend-backpressure.md) |
| SI-05 | `tank-dip-replay-aliveness` | 77/100 | not met | [SI-05 · tank-dip-replay-aliveness](ideas/SI-05-tank-dip-replay-aliveness.md) |
| SI-06 | `ev-plaza-firmware-split` | 74/100 | not met | [SI-06 · ev-plaza-firmware-split](ideas/SI-06-ev-plaza-firmware-split.md) |
| SI-07 | `field-log-injection` | 66/100 | not met | [SI-07 · field-log-injection](ideas/SI-07-field-log-injection.md) |
| SI-08 | `amr-fleet` | 63/100 | not met | [SI-08 · amr-fleet](ideas/SI-08-amr-fleet.md) |
| SI-09 | `reservoir-release` | 73/100 | met as supplied¹ | [SI-09 · reservoir-release](ideas/SI-09-reservoir-release.md) |
| SI-10 | `tick-feed-integrity` | 77/100 | not met | [SI-10 · tick-feed-integrity](ideas/SI-10-tick-feed-integrity.md) |

¹ Verdict as recorded by the workflow's quality-bar evaluation and stated in the idea
document. Both documents explicitly record the discrepancy rather than smoothing it over:
taken literally against the stated bar, **no idea passes** — SI-02 (80/100) holds the
per-dimension floor but sits below the 85/100 total; SI-09 (73/100) fails both clauses
(novelty 6/10). The consolidated scores and the exact discrepancies are in
[review-evidence.md](review-evidence.md) §4.

Improvement rounds consumed: SI-01 1, SI-02 1, SI-03 1, SI-04 2, SI-05 2, SI-06 2, SI-07 10
(cap), SI-08 10 (cap), SI-09 10 (cap), SI-10 1.

## 4. What the portfolio found

The five lenses converged on findings that outlast any single idea; the full record is
[review-evidence.md](review-evidence.md) §3. In brief:

- **Machinery claimed in the wrong layer is the common failure mode.** The shipped oracle is a
  four-form SNR scan over a clean-vs-faulted world written into one closed label record; three
  ideas (SI-01, SI-08, SI-09) initially grounded their value in oracle semantics that have no
  detector form and no record field.
- **The one-entity detector rule.** Four ideas parked a headline mechanism somewhere the
  runtime cannot reach (a detector form outside the declared four, a cross-entity balance, a
  churn registry, a schedule truth). The admission rule the review imposed: name the ONE
  existing detector form, computed on the ONE labeled entity, that produces
  `first_observable_time` — anything beyond is an explicit domain-agnostic design change first.
- **Three inverted flagship cells.** Confirmation channels report the counterfactual shadow
  under `silent_no_effect` (`internal/world/reading.go:23-27`), so SI-09 and SI-10 initially
  specified their G8 cell backwards; absence-as-fault auto-trivials because the truth solver is
  emission-blind (SI-01); `batch_size`/`time_reference` are parsed-only (SI-03).
- **Shared gate-level prerequisites**, all named as design changes first, never hidden:
  `run.End` never flushes perturbation buffers (SI-01/SI-04/SI-05); a batch-emission capability
  (SI-03, with a periodic fallback); the injection-cell machinery — probe corpus, sampler case,
  stamping fix, fail-closed validation (SI-06/SI-07); a duplicate-command-log generator plus an
  exactly-once EffectorCall audit (SI-02); an optional settlement ground-truth block plus
  per-period scoring (SI-04); a churn/lifecycle truth home (SI-08, declined in-scope); a v2
  `release_calendar` schema extension (SI-09).
- **Portfolio balance.** The set concentrates on late/duplicated/replayed delivery resolved by
  the ledger's transport-vs-reasoning split (six ideas), with command/ack discipline again in
  three more — while the shipped zero-coverage frontier (rate=sparse, cardinality=singleton and
  large, value_shape=distribution, lateness=none, fidelity F2, batch-relative/cycle-relative/
  calendar time references) and catalog 25's adversarial-adaptive fault model stay unhosted.

## 5. How to read this folder

- **This README** — purpose, method, the portfolio table, the cross-idea findings, the
  recommendation (§6).
- **[PRODUCT-PROOF-REVIEW.md](PRODUCT-PROOF-REVIEW.md)** — a later review of this portfolio
  against proving Agentic Stream and Tamoz value: what to keep or fold, and a replacement idea
  set (P-1 to P-7) built on a baseline ladder and an outcome-regret oracle.
- **[research-notes.md](research-notes.md)** — the raw record: the 25-candidate pool with cut
  dispositions, the four research angles with sources and their limitations, and the three
  project briefs condensed.
- **[review-evidence.md](review-evidence.md)** — the five-lens verdict per idea, the
  portfolio-level lens notes, and the consolidated final score table with its provenance note.
- **[ideas/](ideas/)** — one document per idea, in a fixed shape: purpose; how it runs on the
  simulator; feasibility against G1–G9; quality-bar result; five-lens review; improvement
  history; open questions. Read an idea's *Quality bar result* and *Open questions* before its
  design sections — they carry the honest scope.

### Write order (all files dated 2026-10-05)

The folder was produced in passes on the same day, and every present-tense verification claim
("this session / this ask / this pass") belongs to the pass that wrote it:

1. **Research stage** — the pool and the four angles, preserved in
   [research-notes.md](research-notes.md).
2. **Per-idea writing passes** — each idea document's tests, greps and suite runs ran on a
   tree that did **not yet contain this folder**. Grep-hit claims in the idea documents are
   therefore true of their own pass's tree; names they report as absent (e.g.
   `fleet-telemetry-gateway`, `pump-station-relay-bank`) now also appear in
   [research-notes.md](research-notes.md) §1, which records the pool. The correction pass
   re-ran the killed-name greps with corrected flags: zero hits outside
   `docs/simulation-ideas/`.
3. **Index and correction passes** — this README, [research-notes.md](research-notes.md) and
   [review-evidence.md](review-evidence.md) frames; the correction pass re-read the runtime
   code where documents contradicted each other, re-ran the affected greps with corrected
   flags, quoted the coverage output SI-03 had promised, and annotated every claim whose
   truth was pass-relative (corrections summarized in
   [review-evidence.md](review-evidence.md) §5). It re-ran no test suites.

### Two schema paths, one schema

Idea documents cite the domain schema at both `docs/contracts/domain-spec-v0.1.schema.json`
and `internal/schemas/domain-spec-v0.1.schema.json`. They are one schema: `docs/contracts/`
holds the committed originals, `internal/schemas/` embeds byte-identical copies via `go:embed`
(`internal/schemas/schemas.go:12`), and `TestEmbeddedMatchCommitted`
(`internal/schemas/schemas_test.go:16-37`) fails the build if the copies ever drift (both
verified during the correction pass: `diff -q` exit 0). Cite either path — line numbers are
identical — and where two documents cite different line numbers for the same item, the higher
is the verified one (the detector-form enum is at line 408, grep-checked this pass).

Provenance: all checks cited in the idea documents were run in their own passes, and each
names its commands and what was not run. Documentation-only change throughout; nothing
outside `docs/` was modified.

## 6. What a next round should take

No admission decision is made here, but the record supports an ordering, and the folder
should end with it rather than with findings alone:

1. **Land the shared gate-carrying slice first** — the ledger/quiescence lens's first item
   ([review-evidence.md](review-evidence.md) §3.4): `run.End` buffer terminality, an
   injection-provenance record, and batch/burst-shaped >10k conservation tests. Five ideas
   (SI-01, SI-03, SI-04, SI-05, SI-07 — plus SI-10's G4 evidence) sit on these holes; nothing
   downstream is honest until they close.
2. **SI-02 `feeder-switching`** — highest score (80/100), the portfolio's reference G8
   design, and the smallest named gap: two generic prerequisites (a duplicate-command-log
   generator; the exactly-once EffectorCall audit) plus the cross-profile shortfall fix its
   evidence pass surfaced. Its headline dimension is orthogonal to the trivial-baseline
   audit, so G9 risk is low.
3. **One of the absence pair, not both** — SI-05 (77/100, its zero-new-machinery claim held
   best; the replay-mask lane) or SI-01 (70/100, live three-way separation under a dispatch
   deadline). The catalog's decoration rule says ship one and fold the other (SI-01 §7.6;
   SI-05 §1).
4. **SI-10 `tick-feed-integrity`** as the torrent-shaped G4-evidence vehicle (77/100,
   dataExpressibility 9) — gated on measuring the soak cost its own feasibility section
   declines to assert.
5. **Spend the last slot on the unhosted frontier, not a sixth delivery-path variant** (the
   novelty lens's recommendation, [review-evidence.md](review-evidence.md) §3.5): catalog
   25's adversarial-adaptive fault model, or an F2/distribution world — which also absorbs
   SI-07's fold rule (yield the slot to catalog 25 if the injection machinery ships under
   SI-06; SI-07 §7.1).
6. **Hold SI-03, SI-04, SI-06, SI-08 and SI-09** pending their named design changes: the
   batch-emission capability (SI-03), the settlement ground-truth block plus per-period
   scoring (SI-04), the seed-sweep admission proof (SI-06), the in-band tempo measurement —
   likely landing as a `k8s-cluster` profile rather than a new domain (SI-08 §7.6) — and the
   SI-02/SI-09 ownership split confirmation (SI-09 §7.6).

This ordering is the record's own arithmetic — scores, named prerequisites, fold rules — not
a new decision; the decision belongs to the run owner.
