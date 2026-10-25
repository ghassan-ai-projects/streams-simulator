# Rule promotion: reasoning that hardens into rules

- **Date**: 2026-10-25
- **Status**: decisions and agreed acceptance scenarios; no solution design and no code yet
- **Products**: Tamoz (learns and drafts), Agentic Stream (checks, stores, runs), Streams Simulator (proves it)
- **Origin**: idea I-3 in Agentic Stream `docs/investor-challenges-2026-10-11/04-ideas.md`

The acceptance scenarios, the campaign and the four arms are in [SCENARIOS.md](SCENARIOS.md).

## 1. Goal

Today every decision about *what to do* goes to an LLM episode. A site's SituationSpec holds static rules for
when a situation opens, how it moves between phases, when to reason and which actions are allowed, but no static
rule picks the action (the only one, `intent.baseline.byPhase`, is used by replay and shadow, never live).

Tamoz has learning and self-improvement. When a kind of case has been solved and verified often enough, Tamoz
proposes a deterministic rule; a human approves it; from then on Agentic Stream handles that case with no LLM
call. The LLM handles a shrinking remainder, the approved rules become the site's playbook, and the cost per
decision falls.

The simulator proves that this is worth it and safe: the same world is run with and without learning, and the
sealed ground truth judges both.

## 2. Decisions (owner, 2026-10-25 and 2026-10-26)

| # | Decision |
|---|---|
| 1 | **A human approves every new rule**, including rules that only recommend. There is no automatic promotion. Tamoz never promotes its own rule (Tamoz invariant 34, `docs/P12_SELF_HEALING_PLAN.md`). |
| 2 | **Rule shape**: a playbook rule has a `when` condition in the restricted, deterministic CEL Agentic Stream already uses (`internal/spec/internal/domain/cel.go`); an `envelope` of the feature ranges seen in its source cases, computed by the compiler (widening it is a separate approved change); a `recommend` limited to an operator-declared intent and preset (no free parameter values); a `mode` (shadow or active); and provenance (source episodes, approval digest). When two rules match, none fires and the case goes to the LLM. A sampled share of rule-handled cases still runs an audit episode in shadow. |
| 3 | **Kind pinning**: each rule is pinned to the device `kind_digest` it was learned on. A device that is born again with another kind digest puts every rule pinned to the old digest back to shadow until a human approves it again. |
| 4 | **The playbook is its own file**, `<site>.playbook.yaml`, next to the site's SituationSpec in the customer workspace, compiled with it into one spec digest. Engineers author the spec; Tamoz proposes changes to the playbook only. |
| 5 | **Per-rule circuit** in Agentic Stream: a rule is switched off automatically when its kind digest changes, when a sampled audit episode disagrees, when its recommended action had no effect on the device's own readings, or when most cases matching its condition fall outside its envelope. Switching it back on needs a human. The circuit state lives in the deployment's database, never in the playbook file. The audit rate is set per rule at approval; the default is 1 in 10. |
| 6 | **When Tamoz may propose**: 3 verified cases by default (a per-site setting) with the same situation type, kind digest, intent and preset, each outcome verified by the device's readings; no contradicting case inside the proposed envelope; a replay report over the site's history attached; one pending proposal per situation type; no new proposal after a rejection without new evidence. How a proposal is made is decision 14; what verified means is decision 15. |
| 7 | **The success bar is a complex simulation**, not the real board: the rotating-machinery world as a 10-week campaign, with four arms (see [SCENARIOS.md](SCENARIOS.md)). The nine scenarios there were agreed before any solution design. |
| 8 | **Harness split**: the simulator owns a generic phased scenario suite and its grading; Agentic Stream owns the campaign driver, the scripted approvers and the campaign score; Tamoz tests its proposal step on its own in `tamoz-evals`; a real-board run, if any, stays a runbook. Model responses are recorded so a campaign replays exactly. |
| 9 | **Both may nominate a pattern.** Agentic Stream nominates when its ledgers show 3 matching verified cases; Tamoz nominates a pattern found in its memory by citing episode ids. Either way Agentic Stream checks the cited cases against its own ledgers and computes the envelope itself. |
| 10 | **The key of "the same case"** is situation type, kind digest, Tamoz's diagnosis code (from the spec's `diagnosisCatalog`), intent and preset. |
| 11 | **One approval** covers the rule and its exit from shadow, for example "active after 10 shadow matches that agree with the LLM and none that disagree". The move to active is a deterministic check against evidence the human already approved, not a decision by Tamoz or the rule. Any disagreement in shadow keeps the rule in shadow and asks the human again. |
| 12 | **The human may edit a rule before approving it.** The edited rule is compiled and replayed again, the approval binds the edited digest, the provenance records it as human-edited, and the edit goes back to Tamoz's memory as feedback. |
| 13 | **Hand edits to the playbook file are allowed and recorded.** On load a hand change is compiled and linted like a spec change and recorded in the approval ledger as changed outside the flow, with its diff and author; rules keep their state. Agentic Stream writes the file only if its digest is still the one it last loaded; otherwise it refuses and checks any pending proposal again against the new file. |
| 14 | **Proposal flow.** Tamoz has no access to the workspace. Agentic Stream sends a bounded "draft a rule" episode with the cases and the current playbook (read only) and receives a typed rule proposal; it compiles it, computes the envelope from the recorded features, replays it over the site's history, stores it as pending and notifies the approvers. |
| 15 | **Verified is not correct.** Each situation type declares a settle window in the spec (for example 24 h for bearing faults, 2 h for overload). A case is verified only when its action took effect and no situation of the same type reopens on that entity within the window; the replay report shows the reopen rate of every source case. Optionally, a site declares a remedy map (which intents are valid remedies for which diagnosis), and Tamoz cannot propose a rule that pairs a diagnosis with an undeclared remedy. |

## 3. Where things live

Customer data is never in a repository and customers never share anything
([plant wire protocol §16](../simulation-canvas/plant-wire-protocol.md)). Learned rules are customer data.

| What | Where |
|---|---|
| Playbook schema, compiler, rule evaluation, per-rule circuit, proposal checks, approval flow | Agentic Stream repository (the program) |
| The rules (`<site>.playbook.yaml`) and their version history | The customer workspace, next to the SituationSpec. Agentic Stream is the only process that writes the file, and only after approval |
| Pending proposals, replay reports, approval records, circuit state | The customer deployment's Agentic Stream database |
| What Tamoz learned, its rule candidates and their provenance | The customer deployment's Tamoz memory |
| Rules learned in a simulation campaign | That campaign's run workspace, never committed |
| The world, its faults and the sealed truth | The simulator. It never sees, reads or grades with a rule: grading the consumer with its own logic would prove nothing |

The simulator and Agentic Stream share only the plant wire contract and the device kinds it carries (vendored
byte-identically and pinned by digest). The rules never enter that contract.

## 4. Still to settle in the solution design

1. **Switch-off limits.** The limits in [SCENARIOS.md](SCENARIOS.md) §3 come from an earlier draft and need the
   owner's confirmation.
2. **Expected effect sizes.** Write down, before any run, the effect each value measure is expected to show and
   the spread across the 3 repeats that still counts as a pass.
3. **Per-pump features.** Scenario S4 holds only if the SituationSpec computes features relative to each pump;
   CEL over absolute values cannot express "warm for this pump".
4. **Prerequisite I-13.** Rule promotion assumes Agentic Stream's lint, test, diff and shadow path (I-13).
5. **A first slice.** S1, S2 and S7 need far less simulator work than the full campaign and show early whether the
   value is there.

## 5. Simulator work this needs

| Gap | Note |
|---|---|
| Campaign | `suite` publishes independent scenarios (one fault, 48 h each). A campaign needs an ordered list of runs with phase labels and seeds, generic and without consumer knowledge. |
| Recurrence | Repeated faults on different pumps over weeks need several faults per run or a sequence of runs; `correlated_cascade` is the only multi-fault profile today. |
| Domain variants | New vibration probes (another kind digest and gain) and a higher operating point (load 0.8 to 1.0). |
| Injection probe | Confirm rotating machinery carries text an injection can ride on. |
| Plant wire | Run rotating machinery over plant wire. Agentic Stream's current rotating-machinery spec uses ticket intents, not the three effectors. |
| Sensor faults | `schedule_maintenance` never clears `vibration_bias` although both sensor faults expect it (D-121). The lying-sensor scenario waits for the fix. |

## 6. Prior material

- Agentic Stream `docs/investor-challenges-2026-10-11/04-ideas.md`, I-3 (reasoning that hardens into rules) and
  I-13 (lint, test, diff, shadow, promote); `03-landscape.md` on LLM-written symbolic rules.
- Tamoz `docs/design-v0.1/SELF_HEALING_DESIGN.md` (three reliability mechanisms, rule lifecycle) and
  `docs/P12_SELF_HEALING_PLAN.md` (no self-promotion, the promotion lifecycle).
- Agentic Stream `internal/watch`: watch conditions, the model-installed, expiring deterministic rules a playbook
  rule resembles.
