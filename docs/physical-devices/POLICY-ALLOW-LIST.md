# Policy allow-list for simulated devices (D-88): open decisions

Status: open, written 2026-10-19 during M5. Nothing here is built. It is worked on separately from M5, after the M5 reviews close.

## What the allow-list is

Every plant wire `DCMD` carries a `policy_digest` (SPEC §5 table: "`policy_digest` is allowed (a bench allow-list; for a plant, an
OT-owned limits envelope that host policy can only narrow)" otherwise the node refuses `policy_not_allowed`). A node only
executes a command whose digest is on its list. The node is whoever executes the command:

| Device is | The node that checks is | Where the list is set today |
|---|---|---|
| Simulated | the simulator (`Session.authorize`, `app/commands.go`) | `plantwire connect --policy <digest>` (repeatable) or `plantwire node --policy` |
| Real board | the gateway and device firmware (SPEC §9 mapping: "`policy_not_allowed`: gateway only") | in the gateway and firmware, outside this repository |

The reference host signs with the single digest given to `plantwire host --policy` (`host/commands.go`). Agentic Stream, the real
host, computes its own digest from its own policy; it lives outside this repository, so what it exposes is unverified here.

## The gap (D-88)

A session started from the canvas never supplies `--policy`, so the simulator's list is empty: every host command to a SIMULATED
device is refused `policy_not_allowed`, even after an operator approves it. Real boards are unaffected (the relay passes the command
to the gateway). Consequence: the canvas cannot run "host command executes on a simulated device"; `canvas/e2e/interrupts.spec.ts`
asserts the refusal instead.

## Decisions to make

1. **Where is the allow-list declared?**
   - (a) In the session plan (`session-plan-v0.1`, optional `policies: [digest...]`), validated by the binary, passed by the bridge
     as `--policy` flags. The plan is already the contract for "what this session is". Recommended.
   - (b) In workspace settings read by the bridge. Hides a trust setting away from the session.
   - (c) Keep it CLI-only. Leaves the canvas unable to run the flow.
2. **Does the canvas let the operator type it, or discover it?**
   - Typed field in the Connect session panel (paste the digest from Agentic Stream). Simple, explicit.
   - Learned from the first command the host sends (show it, "trust this policy?"). More convenient but weakens the check: the
     operator would approve a digest after the fact; it must remain an explicit approval, never automatic.
3. **One digest or several?** The CLI is repeatable. Agentic Stream may narrow policy over time; the list should allow several, bounded.
4. **Per session or per workspace default?** A session-plan list is per session; a saved default per plant is possible later.
5. **What does the node show when it refuses?** Today the refusal reads "Policy not allowed". Should the commands table show the
   offending digest (shortened) so the operator can copy it into the list? The binary would have to expose it (additive).
6. **Real boards:** the gateway/firmware list is separate. The R13 handoff brief for the gateway must say how to set it, and the bench
   check must confirm both lists name the same digest.
7. **Agentic Stream side (outside this repo):** confirm how it exposes the digest it signs with (setting, log, API), and whether
   narrowing changes it. Needed for the handoff brief.

## Proposed shape if (1a) is accepted (for the next round)

- Contract: `session-plan-v0.1` gains optional `policies` (array of `sha256:<64 hex>`, at most 8), schema + embedded copy.
- Binary: `plantwire connect` takes them from the plan (the CLI `--policy` stays and is merged, deduplicated, sorted); the check
  lives in the existing `authorize`; `connect --check` echoes the list.
- Bridge: no logic; passes the plan to the binary as today.
- Canvas: a field "Allowed policy digests" in the Connect session panel, one per line, shown back from the binary's check result
  (the canvas validates nothing), kept in the plan draft.
- Tests: plan schema, authorize accept and refuse, check echo, canvas form, e2e with the scripted host's `scriptPolicy`
  (`canvas/e2e/support/session.ts`) turned into an accepted command and a reply `accepted`, which also gives the first end-to-end
  "approved command executes on a simulated device" proof.
- Docs: SPEC note unchanged (node-local list); handoff briefs (R13) updated.

## Out of scope

Policy authority itself (who signs, how policies narrow): Agentic Stream and the plant's OT envelope. The simulator never creates
or widens a policy; it only declares which digests it trusts, as a panel switch does.
