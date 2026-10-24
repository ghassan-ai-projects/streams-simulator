# Control-plane direction

This is a product proposal based on the browser audit. It does not prescribe unreviewed backend architecture or assert that unseen connected-session features are absent. Any new command, policy, validation, derived status, or orchestration capability must belong to its owning simulator module and be exposed consistently to the UI and other clients.

## The operating model

The UI should keep one operating context as the user moves between tasks:

`Workspace → system/plant → environment → active session → selected asset → selected evidence time`

The same context should answer six questions without visiting unrelated pages:

| Question | Persistent answer |
|---|---|
| What am I looking at? | System, plant version, session identity, simulated/hardware/mixed origin |
| Is this live? | Live or historical; wall time, simulation time, playback cursor, freshness |
| Can I trust it? | Connection health, observation quality, coverage, recording integrity |
| Who can act? | Current controller, user authority, policy scope, hold/interlock state |
| What needs me? | Unacknowledged conditions, pending approvals, failed/incomplete actions |
| What happened after the action? | Receipt, application/report, observed physical effect, unresolved evidence |

Keep the existing graph, outline, inspector, and timeline. Give them an operating purpose. A wiring diagram should be one view of the system, with area summaries, asset tables, trends, and incident views available in the same context.

## Suggested destinations

| Destination | Main user task | Primary content |
|---|---|---|
| Operations | Understand what needs attention now | Active systems, degraded conditions, data freshness, decision queue, recent failed operations |
| System | Monitor and act within one system | Plant/area view, asset state, observed trends, actions, health, persistent time and authority |
| Experiments | Prepare and run a reproducible test | Saved scenarios, objectives, complete run plans, comparisons, campaigns |
| History | Explain and hand over what happened | Runs, incidents, operator decisions, integrity, evidence, exports |
| Engineering | Change models and interfaces deliberately | Plant builder, device definitions, twins, assumptions, validation, version diffs |
| Connections | Establish and diagnose runtime links | Named host/gateway profiles, discovered boards, identity mapping, readiness, policies |

These are tasks, not necessarily six new pages. Use context-preserving tabs or panels where that reduces navigation. Engineering and Operations need visibly different states even if they share components.

## Proposed capabilities

### OP01 · Persistent, recoverable operation context

Carry lifecycle, origin, mode, session identity, authority, and freshness across navigation and browser reconnect. Show unknown state explicitly while reconciling. Provide a way to return to every active operation and every unresolved failed attempt.

Keep execution outcome separate from recording integrity, data sufficiency, and scoring. An integrity pass must never overwrite a failed launch or interrupted experiment. A simulation stop must retain planned duration, actual duration, and reason.

**Addresses:** F01–F05, F09, F20. **First acceptance:** reload a running job and regain its actual state and allowed controls; reproduce a failed launch and see its failure as the main outcome.

### OP02 · A human decision queue with evidence

No usable approve/reject workflow was discoverable in the reviewed normal simulation, workspace, device, and session-setup screens. A working connected host was unavailable, so the audit does not conclude that its live command queue has no relevant controls. Discoverability and full human decision semantics still need an explicit acceptance exercise.

For a proposed action, show:

- Who/what requested it, why, and the incident or objective it addresses.
- Exact target, environment, hardware/simulated origin, command, arguments, units, and current relevant values.
- Risk classification, authority/policy identity, preconditions, expiry, and the evidence timestamp.
- Expected effect, failure modes relevant to that action, permitted alternative, and available safe action.
- Approve once, reject with reason, or defer; clear distinctions from enabling a policy or releasing a timing queue.
- An immutable decision record linked to the eventual command and its observed outcome.

Recheck stale proposals against authoritative state before dispatch. Bind approval to the exact request, target and scope. Prevent duplicate dispatch when the user double-clicks, reconnects, or the response is lost. Expired/rejected requests remain auditable. Avoid approval fatigue by using explicit policy scopes, not a generic confirmation on every low-impact click.

Do not describe risk labels or an allowlisted policy digest as human approval. Do not show “done” when only receipt is known. Preserve receipt, application/board report, and independent effect evidence as distinct facts.

### OP03 · Explicit control ownership and operating modes

Show observer/operator/director privileges and which controller currently owns commands. Make manual override, hold, and release scope visible. Identify whether another browser, agent, or host is issuing commands. Expose what happens to authority when a lease expires or a connection disappears, using the actual backend semantics.

Use separate action vocabularies for simulation execution, connected-session lifecycle, and physical safe-stop requests. Pausing playback, pausing a simulation, ending a recording, and requesting equipment stop are different operations. Hardware-origin actions need an unmistakable target and scope summary.

**Validation required:** real connected host, command-capable device, competing controllers, denied policy, stale request, and lost reply. None was validated here.

### OP04 · An incident and alarm workspace

Provide an ordered list of actionable conditions with severity, onset, age, asset, current state, ownership, and evidence quality. Link it to plant focus and synchronized trends. Distinguish process alarms, device faults, missing observations, connection failures, approval timeouts, and recorder failures.

Support acknowledge/assign/comment and a history where the backend supports them. Do not let acknowledgment imply that the equipment recovered. Explain alarm coverage: no configured detection, no observation, and no active alarm must be different states. Provide event grouping and filters to manage cascades without hiding their consequences.

**Addresses:** F06–F09, F17–F19. **Acceptance:** an operator finds the cooling incident, sees downstream impact, acts, and demonstrates recovery from observations without guessing the injected truth.

### OP05 · Guided connection and hardware supervision

Save named connection profiles and show discovered board identity, firmware/capability information where available, expected-versus-announced devices, policy scope, and recording readiness. Make physical-instance mapping an explicit review step. Do not silently substitute simulated equipment for missing hardware.

After launch, retain host, gateway, board, device and recorder health as separate signals. Show last contact, observation age, reconnect/rebirth history, command backlog and unresolved commands. A stopped or disconnected session must not imply a de-energized physical device.

Emergency action should remain accessible in its operating context, identify scope and unresolved targets, and preserve the backend's stopped/skipped/unreachable distinctions. Successful receipt is not independent proof of a physical stop. This proposal needs hardware/host validation; the audit exercised only launch failures.

**Addresses:** F03, F14–F16, F18.

### OP06 · Evidence workbench and handover

Add channel pinning, synchronized trends, event-relative windows, quality overlays, before/after periods, and measured-versus-true views where permitted. Preserve the selected asset and time when switching between graph, report, commands, and incidents.

Provide an operator-facing export containing run intent, input/version identities, decisions, command lifecycle, evidence coverage, outcome, integrity and unresolved questions. Include comments and incident notes when supported. Avoid generating a confident causal narrative from temporal coincidence alone.

**Addresses:** F07, F12, F18–F21, F28.

### OP07 · Reproducible experiments and comparisons

Save the whole scenario, including commands and their arguments, disturbances and clears, environment/delivery conditions, mode, seed, input versions and objective. Offer exact repeat and deliberate variant. Show the difference before launch.

Support baseline versus failure versus mitigation experiments and selected seed sweeps. Keep scored blind evaluation separate from exploratory diagnosis. Surface the consumer/verdict/scoring readiness and the validity checks applicable to the experiment; do not invent a frontend-only scoring path.

**Addresses:** F10–F13, F19, F21. **Acceptance:** repeat the factory mitigation experiment without rebuilding its effector call; compare its defined post-fault window with a matched baseline.

### OP08 · Model provenance and engineering change review

Connect device definitions, model choices, generated twins, plant versions and deployed/connected instances. Show the model's assumptions, bounds, dynamics, noise and calibration evidence. Explain check results precisely: valid interface, matching capabilities, plausible model output, calibrated response, and hardware validation are different claims.

Provide parameter editing where supported, a version diff and affected assets, and a clear route from a read-only generated plant to its owning editor. Preserve original models when creating variants. Make unsupported settings and excluded devices explicit before creation.

**Addresses:** F22–F26. **Acceptance:** change the thermal model's gain, see the affected relation and model version, run a variant, and relate both recordings to their exact model assumptions.

## Delivery order

### Why navigation feels unnatural: five whys

This is a product-level causal hypothesis, bounded by the observed behavior. It is not a diagnosis of implementation internals.

1. **Why is it difficult to grasp what is happening?** The selected live asset lists model channels, while progress, lifecycle, historical values and report results occupy different contexts (F02, F06, F07).
2. **Why does moving between those contexts increase effort?** Selection competes with the run panel, Home emphasizes inventory, and an ordinary reload loses the active operation (F02, F09).
3. **Why can the operator not rely on the status shown in those contexts?** Active work is called Did not finish, failed starts can be Verified, and deliberate stops lose their original intent (F03–F05). The visible status system does not consistently distinguish execution, evidence and assessment.
4. **Why do these distinctions matter so much?** The user's task is a continuing decision loop: observe, diagnose, authorize, act, verify and recover. The tested flows mainly support author, configure, generate and review, with decisions and recovery not consistently represented (F08–F16).
5. **What design issue best explains the pattern?** The visible interaction model centers on engineering objects and completed artifacts more than on a durable, accountable operating session. That is the working design hypothesis. Validate it with the acceptance exercises below and operator testing; do not infer backend ownership or cancellation bugs without a separate authorized investigation.

Changing navigation labels alone will not resolve this pattern. The decisive changes are durable lifecycle state, clear truth and authority boundaries, connected evidence, and complete decision/action history.

| Order | Outcome | Scope | Evidence required |
|---|---|---|---|
| 1 | Trust the run state | Fix Evaluate label disclosure, browser recovery, failed/running/stopped classification | Reproduce F01–F05; prove the corrected visible behavior and truth boundary |
| 2 | See and diagnose a running plant | Live observed values/quality, shared trends, incident focus, stable mode/time/authority | Factory cascade and thermal response; operator completes diagnosis without sealed cause |
| 3 | Act with accountable authority | Immediate simulation controls; discoverable approvals; connected readiness and failures | Exact approval binding, rejection, expiry, duplicate/lost-reply tests; host-backed command evidence |
| 4 | Repeat and compare experiments | Saved complete plans, exact rerun, variants, consumer/scorer readiness | Matched factory baseline/fault/mitigation runs and complete provenance |
| 5 | Engineer and hand over confidently | Model assumptions and revisions, connections, evidence exports, history filters | Model revision and shift-handover exercises |

These are outcome dependencies, not an argument for a large rewrite. Retain the useful existing components and introduce the smallest coherent workflow at each stage. Do not implement policy, simulation rules, scoring, or derived health in the UI when the binary owns them.

## Operator acceptance exercises

### A · Cooling failure with an intervention

Use the existing factory utilities model. Establish a matched baseline with the same seed and declared conditions. Then inject CT-1 fan failure at minute 5 and AC-1 failure to load at minute 8. Schedule or deliberately request CH-1 setpoint 9°C at minute 12.

The operator must identify the relevant cooling and air-system evidence, see data quality and downstream production impact, compare the event periods, and explain what is supported versus uncertain. The UI must retain the exact plan, original duration, command arguments, observed outcome and independent replay-integrity result. No sealed cause may leak into the blind observer view.

### B · Reload, reconnect, and interruption

Start an Evaluate Watch run. Reload after the first command/fault. Reopen from a second browser session where authorized. Lose and restore the UI connection. Verify authoritative status reconciliation and permitted controls without duplicate operations. If execution is interrupted, show its reason, retained evidence and supported recovery. Do not silently switch mode on retry.

### C · Approval under changing conditions

With a real host or a documented representative host setup, present a risky action. Verify pending → approved/rejected/expired states. Change a relevant precondition before approval; the old decision must not authorize a different action. Repeat approval after a lost response and prove at-most-once dispatch by the owning system's evidence. Trace request → decision → command receipt → report/application → independently observed effect.

### D · Connection failure and ambiguous stop

Check a structurally valid plan against an unavailable host. The UI must say which readiness checks remain unknown or failed. Start and retain the failed attempt; later review must still lead with failed-to-start, even if recording integrity passes. With a host and representative device, disconnect a target during a stop request and show unresolved/skipped/unreachable devices accurately.

### E · Sensor failure versus process failure

Run the thermal twin with a known fan response, then add dropout/drift and delivery disruption separately. The operator must see last good reading, quality, age and gaps without being told the sealed injected cause. Missing evidence must not appear as a normal zero, a clean alarm result, or proof of plant recovery.

### F · Shift handover and experiment reproduction

Leave one active run, one pending decision, one failed launch, one deliberately stopped run, and one incomplete recording. A second operator must find each state from Home, understand ownership and next action, and export a concise evidence package. Repeat the mitigation experiment exactly, then create a clearly labeled variant. No hidden loss of commands or truth-mode change is acceptable.

The current audit did not run these future acceptance exercises as a suite. They are the criteria for implementing the proposed control plane.
