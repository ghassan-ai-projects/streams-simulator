# Rule promotion: work plan per repository

- **Date**: 2026-10-27
- **Branches**: `rule-promotion` in each repository, cut from the latest main: streams-simulator `74a0c77c`,
  agentic-stream `d9f87164`, tamoz `9170538c`
- **Inputs**: [README.md](README.md) (decisions 1-18), [SCENARIOS.md](SCENARIOS.md), and three read-only surveys of
  the repositories (2026-10-27) whose key claims were checked against the code

The plan builds the **first slice** (S1 repeat offenders, S2 cavitation look-alike, S7 new probes) end to end
before anything else, as README §4 item 10 recommends. Everything the other scenarios need comes after it.

## 1. Findings that change the design (owner decisions needed)

| # | Finding | Proposal |
|---|---|---|
| F1 | **Accepted 2026-10-27.** **One digest would break approvals.** Agentic Stream derives the policy digest from the spec digest (`policy.DigestForVersion(compiled.Digest)`, `cmd/agentic-stream/serve_source.go:53`), and approvals and session allow-lists bind to it. Folding the playbook into the spec digest (decision 4) would change the policy digest at every approved rule and invalidate approvals in flight. The spec is also loaded once at start; there is no reload. | Give the playbook **its own digest**; the policy digest stays over the spec alone, so no edge gateway allow-list or approval changes when a rule is approved. Every rule decision, proposal, approval and run artifact records the playbook digest. The first slice reads the playbook at start; a playbook-only reload comes later. Amends decision 4. |
| F2 | **Accepted 2026-10-27.** **One situation type per spec.** `CompiledSpec.Situation` is a single struct (`internal/spec/internal/domain/spec.go:15`), and the diagnosis code is free text on the Go side (`primary_hypothesis`). | Keep one situation type (renamed to a general `pump_condition` in the pump-station spec); add an optional, catalog-checked **`diagnosis_code` to the decision** (shared schema with Tamoz, which already derives `selected_code`). The case key (decision 10), the settle window and the remedy map (decision 15) are keyed by it. Several situation types per spec stay out of scope. |
| F3 | **Accepted 2026-10-27.** **A rule's output needs an episode row.** Intents reference decisions, decisions reference episodes, and only one live episode per situation is allowed (`migrations/003_lifecycle_fencing.sql:101`). | An active rule runs as an in-process **`rule` executor** episode (a fourth executor beside fixture, native and remote; no model call), chosen at admission when exactly one active rule matches within its envelope and caps; its decision follows the existing decision, intent, policy and verification path and records the rule id, version and playbook digest. A rule in shadow writes a `rule_evaluations` record compared with the Tamoz decision when it lands. The audit is a `dispatch_policy: shadow` Tamoz episode admitted after the rule episode ends. |
| F4 | **Drafting is not a situation episode.** Episodes require a situation and snapshot; a draft has neither. Tamoz's frame holds at most 64 facts of 4 KB. | A separate, feature-gated **`DraftPlaybookRule` call** in the worker proto (Agentic Stream owns the proto; Tamoz vendors it), carrying the cited cases and the current playbook. Tamoz names its type `PlaybookProposal`, because "rule" already means `HealingRule` in Tamoz. |
| F5 | **The simulator grades one fault per run.** `score.Score` takes one ground-truth record and `UnnecessaryAction` counts every call in the run (`internal/score/internal/domain/score.go:82`, `loop_metrics.go:19`). | The first slice's campaign is a **sequence of one-fault runs** with phase labels; Agentic Stream and Tamoz keep their state across runs, the world starts fresh each run. Multi-fault runs with per-fault labels come later (S6). |
| F6 | **Each pump does not have its own normal yet.** Every pump starts from the same initial values; only noise differs. `rotating-machinery` is also the pinned reference domain (re-pins in three places if edited). | All changes go into **new domain files** (a pump-station copy and its variants); the reference stays untouched. Per-pump baselines (S4) are a later simulator round. |
| F7 | **"Defer to a human" is not a truth the simulator can state.** `ExpectedEffector` is one string. | Later: an `expected_response: defer` (or forbidden effectors) in the domain schema, ground truth and scoring, for S6 and S10. Not in the first slice. |
| F8 | **Verification confirms set points, not recovery.** Agentic Stream reads back the commanded metric; "the fault stopped" and the settle window are new logic, and the pump's kind file must declare the metrics to read. | Agentic Stream round AS-5 below. |
| F9 | **I-13 is not built**, and there is no export from the database to a trace. | The first slice replays a proposal over the **campaign's own run traces**; I-13 (lint, test, diff) comes after the slice. |

## 2. Streams Simulator

| Round | Work | Owner | Kind | Size |
|---|---|---|---|---|
| SIM-1 | A `pump-station` domain file (a copy of rotating-machinery) with the D-121 fix (`schedule_maintenance` also lowers `vibration_bias`; failing test first). | domain data | data | S |
| SIM-2 | A probe variant of it: `vibration` with another gain and range, so the pump's `kind_digest` changes (S7). | domain data | data | S |
| SIM-3 | A `campaign-v0.1` contract (ordered runs with generic phase labels, seeds and scenario references; no labels the consumer can read) and its generator in `suite`, drawing one-fault runs per phase with the fault weights and the 40% negative share. | `suite`, `contracts/` | code | M |
| SIM-4 | Plant wire tests for the pump station: births carry the `kind_digest`, a second run with the probe variant is born with the new digest, `DCMD` for `schedule_maintenance` and `reduce_load` reaches the world. | `plantwire` (tests) | tests | S |

After the slice: per-pump baselines (S4, schema change, M); multi-fault runs with per-fault labels and an entity
and window filter in scoring (S6 recurrence, M-L); header coupling as data plus "defer" truth (S6, M + M); entity
scoped perturbations and a `motor_current` probe fault (S10, S + M); maintenance lowering a pump's normal (S11,
S); an operating-point fault (S8, S); a `maintenance_log` text channel open to injection (S9, S).

## 3. Agentic Stream

| Round | Work | Owner | Size |
|---|---|---|---|
| AS-0 | Design note for a new `playbook` module (kind, layer, responsibility, callers, per the repository's rule) and the F1-F4 decisions written into it. | docs | S |
| AS-1 | Spec: playbook schema and compiler with its own digest and the denylist (R3+, plus new `emergency` and `safety_function` flags on intents); a settle window per situation type; a typed diagnosis code in the decision schema. | `spec`, contracts | M |
| AS-2 | A pump-station spec with per-pump relative features (a long window against a short one), a diagnosis catalog, intents for `schedule_maintenance` and `reduce_load` (and `stop_motor` declared R3, never a rule target), plus the plant wire kind, site and capability files. | `examples/` | M |
| AS-3 | The `playbook` module: rule tables, evaluation before trigger admission, the rule executor, shadow records, the per-rule circuit (kind change, audit disagreement, out-of-cap) and rate cap, playbook reload. | new `playbook`, `cognition`, `runtime` | L |
| AS-4 | A kind-digest query on the `plantwire` facade and a notification when a device is born with another digest. | `plantwire` | S-M |
| AS-5 | Verification for rules: a case is verified when its action took effect and no situation of the same type reopens on that entity within the settle window. | `situations` read, `playbook` | M |
| AS-6 | The envelope from the source cases' recorded features, with cap and walls (growth and weekly coverage later). | `playbook` | M |
| AS-7 | The proposal flow: eligibility (3 verified cases, same key), the `DraftPlaybookRule` proto call, checks on the reply, replay over the campaign traces, a rule approval subject over the signed approval endpoint, and the playbook writer with its loaded-digest check. | `playbook`, `proto`, `api`, `cmd` | L |
| AS-8 | The campaign driver in `internal/testsupport/campaign`: runs a simulator campaign in sequence, plays the scripted approvers (careful, approve-all) through the real endpoint, records and replays model responses, and computes the campaign score with the simulator's grading. A `cmd` test skipped without `STREAMSIM_BIN`. | test support | L |

After the slice: fleet cap (S6), envelope growth and weekly coverage, the edit-before-approval flow and the
hand-edit ledger, I-13 (lint, test, diff), more situation types.

## 4. Tamoz

| Round | Work | Owner | Size |
|---|---|---|---|
| TZ-1 | Store the case key (diagnosis code, intent, preset, kind digest) with each admitted verified episode; today it is dropped before admission. | `tamoz-stream` | S-M |
| TZ-2 | The `PlaybookProposal` schema, a strict parser and a fixture reply: typed fields only (`when`, intent, preset, diagnosis code, cited episode ids), no mode and no free values. | `tamoz-agent-kernel`, `tamoz-stream` | M |
| TZ-3 | Vendor Agentic Stream's proto change; the worker serves `DraftPlaybookRule` through a draft path that reuses reasoning and validation, with no tools and no recall. | `tamoz-stream`, `tamoz-agent` | M |
| TZ-4 | A `rule_proposal` scorecard with synthetic pump-station fixtures: a consistent pattern, a contradicting case, cavitation that settles by itself, injected text, mixed keys, too few cases; with null, oracle and adversary controls. | `tamoz-evals-runner`, `test/fixtures` | M-L |
| TZ-5 | Adversarial test: no Tamoz code path writes a rule's mode, and a proposal carries no mode or activation field. | `tamoz-stream` | S |

After the slice: Tamoz nominating patterns from its memory, feedback from approved, edited and rejected
proposals, a rule-candidate record with provenance, and a tenant leak test.

## 5. Order of work for the first slice

```text
0  owner decides F4 (F1-F3 accepted)    -> AS-0 design note
1  SIM-1, SIM-2, SIM-4 | AS-1, AS-2 | TZ-1, TZ-2        (parallel, independent)
2  SIM-3 campaign      | AS-3, AS-4, AS-5
3  AS-7 proto first    -> TZ-3          | AS-6
4  TZ-4, TZ-5          | AS-8 campaign driver (can start on canned proposals after AS-3)
5  run S1, S2, S7 with arms 0, A, B and C; write the result
```

Each round follows its repository's own gates and commits on that repository's `rule-promotion` branch. Nothing
merges to main until the first-slice result is written.
