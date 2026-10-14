# Assessment

- **Date**: 2026-10-11
- **Method**: build, vet and full test run; coverage census; reading the scoring, truth and
  replay code paths; reading the hardening outcome and the open register.
- **Not done**: `make ci-check`, `make lint`, `make function-length`, soak, performance runs,
  cross-architecture runs. Ratings below say where that limits confidence.

## Rating

| Dimension | Score | Basis |
| --- | --- | --- |
| Engineering discipline | 9 / 10 | Facade-over-layers modularity, file-size and 15-line-function rules enforced as executable tests in `test/architecture`; 445 Go files, 163 test files, about 21.8k non-test lines; 178 commits in two months with a recorded behaviour pin |
| Correctness evidence | 8.5 / 10 | Nine non-negotiables each have tests; analytic cross-check, replay determinism and sealed oracle exist; hardening closed the High-severity register items and an independent review found and fixed six more (R-1…R-6). Held back by no cross-architecture run and a stale defect register |
| Instrument completeness | 8 / 10 | Pipeline world → perturb → adapter → sink, director/operator roles, ledger, quiescence, trivial-baseline audit, suite generation, device emulator |
| Domain and scenario breadth | 6 / 10 | Eight domains, 5–7 faults each; 14 thin coverage cells; see [GAPS.md](GAPS.md) §3 |
| Outcome measurement | 4 / 10 | Scores are detection timing, admission outcomes and loop flags; consequence is a tag and a 200-character string, not a number |
| Comparative proof | 3 / 10 | One reference consumer (threshold detector); no baseline ladder; nothing removes scenarios a strong baseline already wins |
| Adoption surface | 5 / 10 | Two adapters, CLI and MCP; no release channel, no container, no CI recipe, no domain-authoring kit |
| Documentation honesty | 8 / 10 | Status headers, explicit limitations; but several statements are now stale (below) |

**Overall: 8.5 / 10 as an instrument; about 5 / 10 as a product.** The gap is not quality,
it is the missing measurement layer that turns correct runs into believable claims.

## What was verified

- `go build ./...`, `go vet ./...`, `go test ./...` all pass; the slowest package is
  `suite/internal/domain` at about 20 s.
- Lowest per-package statement coverage seen was about 75 %.
- Hardening outcome: D-01…D-13, D-15…D-27, D-30…D-34 and others fixed with regression
  tests; simulator version 0.1.0 → 0.2.0 with the pin file regenerated.
- Spot-checked in code: verify exit status, offline score replay, divergence index,
  safe-stop idempotency counter, retry clearing of disconnect/duplicate injections.

## Stale or inconsistent statements found

| Where | Statement | Reality |
| --- | --- | --- |
| `docs/README.md` banner | "current working tree has a known domain-inventory test failure" | Suite is green; eight domains validate |
| `documentation/limitations.md` | "Last verified: 2026-09-13"; offline score loop metrics incomplete | Predates the hardening program; offline score now replays for calls |
| `docs/refactoring/modularity-20261010/DEFERRED.md` header | "not fixed in this program"; table lists D-01…D-06 as open | Fixed in hardening; the table needs a status column |
| `documentation/roadmap.md` | Verified 2026-09-13 | Needs a refresh after hardening |
| `CHANGELOG.md` | "has not established a formal release stream" | Version is 0.2.0; there is no tag or release |
| `docs/simulation-ideas/` | Untracked | Either commit as evidence or move out of the tree |

## Open items still real

- **D-52** — `cold-chain-transit` declares a detector on a channel its fault does not change;
  left to the domain owner because the domain digest changes.
- **Suite runner** absent: generation is audited, execution and score persistence are not.
- **Release evidence**: manifest is historical; no fresh gate, soak or perf run at the
  current commit.
- **Float determinism across architectures** is unverified (see [GAPS.md](GAPS.md) §1.2).
- **Test bar T-01**: 27 test files assert only that an error occurred.

## Strengths worth protecting

1. The refusal to add features that weaken determinism or sealed truth.
2. Domains as data with no per-domain code branches.
3. A consumer-agnostic binary: no consumer knowledge in the core.
4. Generative ground truth with three timestamps (injection, first observable, unavoidable).
5. A trivial-baseline audit that removes scenarios which prove nothing.

These are the moat. Every item in the roadmap is built on them and none may trade them away.
