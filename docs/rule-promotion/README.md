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

## 2. Decisions (owner, 2026-10-25)

| # | Decision |
|---|---|
| 1 | **A human approves every new rule**, including rules that only recommend. There is no automatic promotion. Tamoz never promotes its own rule (Tamoz invariant 34, `docs/P12_SELF_HEALING_PLAN.md`). |
| 2 | **Rule shape**: a playbook rule has a `when` condition in the restricted, deterministic CEL Agentic Stream already uses (`internal/spec/internal/domain/cel.go`); an `envelope` of the feature ranges seen in its source cases, computed by the compiler (widening it is a separate approved change); a `recommend` limited to an operator-declared intent and preset (no free parameter values); a `mode` (shadow or active); and provenance (source episodes, approval digest). When two rules match, none fires and the case goes to the LLM. A sampled share of rule-handled cases still runs an audit episode in shadow. |
| 3 | **Kind pinning**: each rule is pinned to the device `kind_digest` it was learned on. A device that is born again with another kind digest puts every rule pinned to the old digest back to shadow until a human approves it again. |
| 4 | **The playbook is its own file**, `<site>.playbook.yaml`, next to the site's SituationSpec in the customer workspace, compiled with it into one spec digest. Engineers author the spec; Tamoz proposes changes to the playbook only. |
| 5 | **Per-rule circuit** in Agentic Stream: a rule is switched off automatically when its kind digest changes, when a sampled audit episode disagrees, when its recommended action had no effect on the device's own readings, or when most cases matching its condition fall outside its envelope. Switching it back on needs a human. The circuit state lives in the deployment's database, never in the playbook file. The audit rate is set per rule at approval; the default is 1 in 10. |
| 6 | **When Tamoz may propose**: 3 verified cases by default (a per-site setting) with the same situation type, kind digest, intent and preset, each outcome verified by the device's readings; no contradicting case inside the proposed envelope; a replay report over the site's history attached; one pending proposal per situation type; no new proposal after a rejection without new evidence. |
| 7 | **The success bar is a complex simulation**, not the real board: the rotating-machinery world as a 10-week campaign, with four arms (see [SCENARIOS.md](SCENARIOS.md)). The nine scenarios there were agreed before any solution design. |
| 8 | **Harness split**: the simulator owns a generic phased scenario suite and its grading; Agentic Stream owns the campaign driver, the scripted approvers and the campaign score; Tamoz tests its proposal step on its own in `tamoz-evals`; a real-board run, if any, stays a runbook. Model responses are recorded so a campaign replays exactly. |

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

## 4. Open points (solution design, after the scenarios)

Proposed during the discussion, not yet decided:

1. **Who nominates a pattern.** Recommendation: both. Agentic Stream nominates when its ledgers show 3 identical
   verified cases; Tamoz may nominate a pattern it found in its memory by citing episode ids. Either way Agentic
   Stream checks the cited cases against its own ledgers and computes the envelope itself.
2. **The key of "the same case".** Recommendation: add Tamoz's diagnosis code (from the spec's
   `diagnosisCatalog`) to situation type, kind digest, intent and preset.
3. **One approval or two.** Recommendation: approval 1 puts the rule in shadow; after enough agreeing shadow
   matches, a one-click approval 2 with the shadow report makes it active. Alternative: one approval with a
   shadow period set at approval, after which the rule goes active on its own.
4. **Human edits before approval.** Recommendation: allowed; the edited rule is compiled and replayed again and
   the approval binds the edited digest; the playbook records the rule as human-edited.
5. **Hand edits to the playbook file.** Recommendation: Agentic Stream writes the file only if its digest is
   still the one it last loaded; otherwise it refuses and checks the proposal again against the new file.
6. **Proposal flow.** Proposed: Tamoz has no access to the workspace. Agentic Stream detects eligibility, sends a
   bounded "draft a rule" episode with the cases and the current playbook (read only), and receives a typed
   rule proposal; it then compiles it, computes the envelope from the recorded features, replays it, stores it as
   pending and notifies the approvers. This amends decision 6, which placed the trigger in Tamoz's offline
   improvement pipeline.
7. **Verified is not correct.** A rule can be learned from cases whose readings improved while the fault went on
   (scenario 1, the masking trap). Should a case count as verified only when the fault does not return before its
   deadline?

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
