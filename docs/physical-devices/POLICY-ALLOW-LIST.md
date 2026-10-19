# Policy allow-list for simulated devices (D-88): design

Status: decided by the owner on 2026-10-19; P1 (contract and binary) and P2 (bridge and canvas) built. Two rounds (P1, P2) after the M5 reviews close. File and line
references are to `f491fbbc` (branch `physical-devices-m5`).

## Decisions

| # | Decision | By |
|---|---|---|
| 1 | The allow-list is declared **per session**, in the session plan (`session-plan-v0.1`, optional `policies`), validated by the binary and enforced by the node's existing check (`Session.authorize`). Not per device, not in domain files: a domain is shared versioned data and a digest belongs to a deployment; it would also enter the domain digest. | Owner |
| 2 | A **saved default** lives in the workspace (outside the repository) and only pre-fills the Connect session form. | Owner |
| 3 | A command refused `policy_not_allowed` shows the **offending digest** (shortened, copyable) in the commands table. The operator adds it deliberately; nothing is ever added automatically. | Owner |
| 4 | A **list** of at most 8 distinct digests, because the host's policy may be narrowed over time. | Owner |
| 5 | Real boards: the gateway keeps its own list outside this repository; the bench check confirms both lists name the host's digest. | Owner |
| 6 | `plantwire connect` loses `--policy`: the plan is the only source for a connected session. `plantwire node` and `plantwire relay --domain` keep `--policy` and pass it through the same rule. | This note |
| 7 | The list is a recorded session input (the `session` journal record), not part of the run's identity and not a replay input. | This note |
| 8 | The plant wire contract does not change; the digest reaches the commands table through the `live` journal record. | This note |

## 1. Problem and scope

Every plant wire `DCMD` carries `policy_digest`. A node executes a command only when the digest is allowed (SPEC §9, check
table: "`policy_digest` is allowed (a bench allow-list; for a plant, an OT-owned limits envelope that host policy can only
narrow)", refusal `policy_not_allowed`). For a simulated device the node is the simulator. Its list is
`Options.Policies`, read by `Session.authorize` (`internal/plantwire/internal/app/commands.go:111-119`), and today only the
command line fills it: `plantwire node --policy`, `plantwire relay --domain ... --policy`
(`internal/cli/internal/app/plantwire_options.go:41`) and `plantwire connect --policy`
(`internal/cli/internal/app/plantwire_connect.go:66`, passed on at `connect_run.go:144`).

A canvas session never passes `--policy`: the bridge runs `plantwire connect --stdin` with the plan on standard input
(`internal/canvasbridge/internal/domain/session.go:35-39`), and the plan has no member for a list. So the list is empty and
every host command to a simulated device is refused `policy_not_allowed`, even after the operator approves it, because
approval only clears the interrupt gate and then runs the normal checks (`interrupt_side.go:16-22` →
`commands.go:33-41`). `canvas/e2e/interrupts.spec.ts:49-51` asserts that refusal today. Real boards are not affected: the
relay passes their commands to the gateway.

In scope: the plan member, its validation and echo, the recorded list, the refused digest in the live record and the
commands table, the workspace default, the canvas form, the R13 brief wording. Out of scope: policy authority itself (who
issues policies, what they contain, how they narrow). That belongs to the host and the plant's OT envelope.

## 2. The policy digest and who owns it

- **What it is.** `sha256:` and 64 lower-case hex digits (`DCMD.schema.json`, `policy_digest`). It names the policy a host
  says it acts under. It is opaque to the node: the node does not compute it, does not know what it covers, and verifies
  nothing about it. It is not a signature and not a secret.
- **How a host sets it.** The reference host copies the single digest given to `plantwire host --policy` into every `DCMD`
  (`internal/plantwire/internal/host/commands.go:38`). It signs nothing. The CLI help text says "signs"
  (`plantwire_options.go:68`, `documentation/reference/cli.md:280`); P1 changes it to "carries". Agentic Stream computes its
  own digest. How it exposes that digest is not known in this repository (§8).
- **Who checks.**

| Device is | Node that checks | Where its list is set |
|---|---|---|
| Simulated | the simulator: `Session.authorize`, at receipt and again at execution for a queued command (`execution.go:83`) | the session plan (P1); `--policy` for `plantwire node` and `relay --domain` |
| Real board | the gateway (SPEC §13: "`authority_epoch`, `stale_epoch`, `policy_not_allowed` \| gateway only; device firmware is unchanged") | the gateway's own configuration, outside this repository |

- **Authority.** The simulator never creates or widens a policy. Its list only says which digests this bench trusts, the
  way a panel switch does. Safe stops skip the policy check (SPEC §9 "Safe stop"; `commands.go:26-31`). The relay's own
  emergency stops carry the reserved digest `sha256:` and 64 zeros (`stopPolicy`, `domain/relay_interrupts.go:12`) and
  are safe stops, so the list never affects them.

## 3. Contract changes

### 3.1 `session-plan-v0.1`: `policies`

Added to `properties` of `contracts/session-plan-v0.1.schema.json`, plus the embedded copy in `internal/schemas`. The
`additionalProperties: false` rule stays.

```json
"policies": {
  "type": "array",
  "minItems": 1,
  "maxItems": 8,
  "description": "The policy digests this session's simulated devices accept in a command's policy_digest. Absent: none, and every command to a simulated device except a safe stop is refused policy_not_allowed. Real boards are checked by the gateway's own list.",
  "items": {
    "type": "string",
    "pattern": "^sha256:[0-9a-f]{64}$",
    "x-expected": "a policy digest: sha256: and 64 lower-case hexadecimal digits"
  }
}
```

There is no `uniqueItems`. The jsonschema package reports it at the array (`internal/jsonschema/validate_array.go:29-41`),
and the canvas needs the line. The `plantwire` domain reports a repeat at the item instead (§4), as
`instance_bound_twice` does for `real`. `minItems: 1` makes "none" mean one thing: the member is absent.

### 3.2 `cli-output-v0.1`: `plantwire_connect_check.policies`

The check document gains a required member
`"policies": {"type": "array", "maxItems": 8, "items": {"$ref": "#/$defs/digest"}, "description": "The allow-list in effect, sorted; empty when the plan names none."}`.
`connect --check` echoes the list, so the canvas shows what the binary will enforce, not what the form holds.

### 3.3 `run-journal-v0.1`

- `$defs/session/properties/plan` adds required `policies`, with the same shape as 3.2. The `session` record's `plan` is the
  check document (`connect_run.go:73-81`; the mixed path reaches it through `relay_sim.go:75-80`), so the change is a single
  struct field.
- `$defs/live_command` adds an optional `"policy_digest": {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"}`. It is
  present only when the receipt is `refused` with reason `policy_not_allowed`.

### 3.4 Plant wire contract: unchanged

`DRECEIPT` gains nothing. The host already knows the digest it sent. The contract is vendored byte-for-byte, with a pinned
`DIGEST`, by the gateway and by Agentic Stream, and a new member would mean a new contract version across three
repositories. The node pairs the receipt with its own `DCMD` instead (§5).

### 3.5 Recording and replay

The list is recorded once in the `session` record. That record is on the observer stream and outside the replayable stream
digest, per its schema description. It is not part of `run.json`'s identity and not a replay input. Reason: a run is a pure
function of `(sim_version, domain_digest, adapter_digest, seed, command_log, sink)`. The node's admission checks
(freshness, boot, epoch, policy, interrupt gate) run before `run.InvokeEffector` (`execution.go:25-29`), and only a command
that passes them enters the command log. `replay` and `verify` replay that log, and the list cannot change the outcome.
`plantwire replay-trace` refuses every `DCMD` (`replay_answers.go:32-50`) and never consults a list. The interrupt log is
different: it records decisions that change over the session and is digested into the run (`SetInterruptLog`). The list is
fixed for a session, so the record is enough to start the same session again, and nothing needs replaying.

## 4. Binary (P1)

| Rule | Module, layer |
|---|---|
| Shape, pattern, 1 to 8 items | `model` (schema, `DecodeSessionPlan`; issues `pattern` at `/policies/<i>`, `maxItems` and `minItems` at `/policies`) |
| The allow-list: each entry a digest (`policy_malformed`), at most 8 (`too_many_policies`), no repeat (`policy_listed_twice`, at the second occurrence), not the reserved stop digest (`policy_reserved`); the result sorted ascending | `plantwire` domain, new `policies.go`: `NewAllowList(digests []string, pointer string) (AllowList, []model.Issue)`, `AllowList.Allows(digest) bool`, `AllowList.Digests() []string` |
| The plan's list is checked with the plan's other rules, and its issues are joined with theirs | `plantwire` domain `CheckPlan` (`connect_plan.go:51-61`); `ConnectPlan` gains `Policies AllowList` |
| Flags (`node`, `relay --domain`) use the same rule; any issue is `ErrBadOptions` with the issue's message | `plantwire` app `validate` (`session.go:46-57`) |
| Enforcement: `authorize` keeps the epoch check, then asks `s.allowed.Allows(c.Policy)`; the list is built once when the session starts | `plantwire` app (`commands.go:111-119`, `NewSession`) |
| The session's list comes from the plan: `nodeOptions` takes `p.plan.Policies.Digests()`, no longer `o.Policies` | `plantwire` app (`connect_run.go:142-145`) |
| Echo and record: `ConnectCheck.Policies []string \`json:"policies"\``, never nil | `plantwire` app (`connect.go:37-42, 66-69`) |
| The offending digest on the live command (§5) | `plantwire` domain (`live_commands.go`) |
| `plantwire connect` loses `--policy` and its `ConnectOptions.Policies` | `cli` (`plantwire_connect.go:23, 66, 89`), `plantwire` app (`connect.go:29`) |

The format check exists twice: in the schema (the contract other programs read) and in the domain rule (which also covers
flags). The socket rules already do this (`connect_plan.go:63-100`). `policy_reserved` exists because only the relay's own
safe stops use the zero digest, so trusting it would admit any host that copies it. Issues use `model.RuleIssue(pointer,
keyword, "policies", message)`. A plan that does not hold is `plantwire_failed`, as now. `cli` and the bridge gain no
logic.

## 5. The refused digest in the commands table

- The live view already pairs each host `DCMD` with its receipt by `command_id` (`Live.sent`, `Live.receipt`,
  `domain/live_commands.go:13-72`). `sent` keeps the command's `policy_digest` in an unexported field. When `receipt` sees
  `refused` with `policy_not_allowed`, it copies the digest to `LiveCommand.PolicyDigest`
  (`json:"policy_digest,omitempty"`). This is the same for a simulated device (refused by the world's node) and for a real
  board (refused by the gateway).
- The value is always well formed. In node mode, `journalLive` observes only records that pass the schema
  (`app/live.go:5-13`). In relay mode, `liveRecord` observes only admitted host records (`app/live.go:31-38`), and the relay
  drops a schema-invalid host record (D-89). A host can therefore place only a pattern-valid digest there (71 characters,
  no free text). A forged digest is still only the host's claim, and showing it reveals nothing the observer journal does
  not already hold: node-mode `wire` records include the full `DCMD`.
- Size: at most 128 live commands × about 90 bytes, only for refused ones. The canvas reads the field from the decoded
  `live` record and never parses text.

## 6. Bridge (P2)

- `session.check` and `session.start` do not change: the plan, with `policies`, goes on standard input unchanged. The status
  and the recorded-run listing carry the check document as raw JSON (`domain/session.go:53`,
  `domain/runsummary.go:88-97`), so `policies` reaches the canvas with no bridge code.
- **The default file**: `settings/default.policies.json`, content `{"policies": ["sha256:..."]}`. One new file class in
  `domain/paths.go` (`definitionFolders`: `"settings": {ClassPolicies, ".policies.json"}`). It is read and written through
  the existing `GET /api/file` and `PUT /api/file` (`app/files.go:5-32`): `etag` is `domain.Version`, and a stale one is
  `conflict`. The write is atomic (temporary file, sync, rename), owner-only `0600` in a `0750` folder (`files/write.go`),
  capped by the generic 8 MiB `MaxFileBytes`. No action is added, so there is no new registry entry.
- The file is not listed in the workspace listing and never reaches a command. Actions pass only `domains/`, `kinds/` and
  the component folders (`Catalog`, `Kinds`, `Library`), never `settings/`. It is never in `domains/`, `kinds/` or
  `components/`.
- The workspace is the customer's folder. The repository ships no default file, and tests write one only into temporary
  lab workspaces.

## 7. Canvas (P2)

- **Draft and plan** (`model/launch/internal/session.ts`): `SessionDraft.policies: string` (the field's text). `planOf`
  adds `policies` as the trimmed non-empty lines in the operator's order, and omits it when there are none, as it does for
  `real` (`session.ts:55-67`). The canvas does not deduplicate, sort or check a digest. An issue at `/policies/<i>` is shown
  on the i-th non-empty line, and an issue at `/policies` on the field (`policyIssues`, beside `rowIssues`).
- **Form** (`ui/launch/internal/SessionFields.tsx`): a `<textarea>` labelled "Allowed policy digests", placed after the
  gateway socket. Hint: "One per line, at most 8. A command to a simulated device runs only when the host's policy digest
  is listed; safe stops always run. Real boards are checked by the gateway's own list." Issues use the existing
  `IssueNote`/`flagged` (`aria-invalid`, `aria-describedby`), and focus goes to the first invalid control on refusal
  (`SessionForm.tsx:25-34`). The field is full width at 1280 px with 3 visible rows, and a 71-character digest fits on one
  line in the monospace token.
- **Words** (`model/launch/internal/sessionWords.ts`). Keywords: `pattern` at `/policies/<i>` → "A policy digest is
  sha256: and 64 lower-case hex digits."; `maxItems` → "List at most 8 policy digests."; `policy_listed_twice` → "This
  digest is already listed."; `policy_reserved` → "This digest is reserved for the node's own stops." Check result: "Allowed
  policies: none. Commands to simulated devices are refused, except safe stops." or "Allowed policies: N", followed by
  each digest shortened.
- **Default**: `ConnectGateway` gains `readPolicyDefault()` and `savePolicyDefault(list, etag)`, which the bridge adapter
  implements with the file routes. `ConnectStore.draftFor` pre-fills `policies` only for a draft it has never held. Drafts
  stay in memory per plant (`connectStore.ts:14-32`), and an edited draft is never overwritten. A pre-filled
  field shows "Pre-filled from the workspace default." An absent file pre-fills nothing. A file that does not decode
  pre-fills nothing and says "The workspace default could not be read". **Save as default** is an explicit button. It is
  enabled only after a check of the current draft that is not stale, and it saves the check document's `policies`, the
  binary's sorted list, so the default only ever holds a list the binary accepted. An empty list clears it. On `conflict`
  it reports that the default changed elsewhere and saves nothing.
- **List in effect**: the live session panel shows the check document's `policies` from `session.status` (the bridge's
  `plan`). The review decodes `plan.policies` from the recorded `session` record (`adapters/bridge/internal/journalDecoder.ts`);
  if absent, it reads "not recorded".
- **Commands table** (`ui/review/internal/CommandEvidence.tsx:36-71`; words in `model/review/internal/sessionWords.ts:37`):
  - For `policy_not_allowed`, the detail reads "Policy not allowed by this session" for a simulated device and "Policy not
    allowed by the gateway" for a real board.
  - When `policyDigest` is present, a line reads "Host's policy `sha256:bbbbbbbb…bbbbbbbb`" (first and last 8 hex digits,
    `shortDigest` in `model/review`, full value in `title`), with a button "Copy policy digest" described by that line.
  - Copying goes through a new browser port, `copyText` (`app/ports/browser.ts`, `adapters/browser`). `127.0.0.1` is a
    secure context.
  - A polite status reports "Copied the policy digest of <command id>." or "Could not copy; select the digest instead."
  - The table must not scroll horizontally at 1280 px.
- Registration only: new port methods and callers (`node scripts/sync-callers.ts`, `write-architecture-doc.ts`); no import
  allow-list grows.

## 8. Real boards and the R13 briefs

- **Gateway brief**:
  - The gateway keeps its own list of allowed policy digests, in its own configuration, and refuses a `DCMD` whose
    `policy_digest` is not on it with `DRECEIPT refused policy_not_allowed`, in the SPEC §9 check order.
  - It skips the check for its devices' safe stops. The relay's emergency stop carries `sha256:` and 64 zeros.
  - The list holds several digests, so a narrowing can be rolled out.
  - Log the list in effect when the gateway starts.
  - device-wire-v1 has no policy reject code (`device-receipt-v1.json`, `reject_code`). The firmware receives
    `policy_digest` unchanged and does not refuse on it.
- **Agentic Stream brief**: say where an operator reads the digest the host puts in every `DCMD` (a setting, a log line at
  start, or an API), and confirm that narrowing the policy gives a new digest. Until then the operator has two sources: the
  first refused command in the canvas commands table, or the observer journal's `wire` records with direction `host`.
- **Bench check step**:
  1. Start the mixed session with the host's digest in the plan's `policies` and in the gateway's list.
  2. Send one command to a simulated device and one to the real board.
  3. Both must read "Accepted". If one side is refused `policy_not_allowed`, its list differs; the row shows the digest the
     host used.
  4. After a narrowing, add the new digest on both sides, then remove the old one.

## 9. Tests

| Layer | Test |
|---|---|
| `model` | `DecodeSessionPlan` accepts 1 and 8 digests; refuses 0 (`minItems`), 9 (`maxItems`), an upper-case or short digest (`pattern` at `/policies/<i>`); the schemas-equal test covers the copy |
| `plantwire` domain | `NewAllowList`: sorted output; `policy_listed_twice` at the second index; `policy_reserved`; `policy_malformed` and `too_many_policies` for flag input; `Allows` true for a listed digest, false for others and for any digest when the list is empty; `CheckPlan` joins policy issues with binding issues. `Live`: `policy_digest` only on a `policy_not_allowed` refusal, absent on accepted and on other refusals |
| `plantwire` app | A listed digest is accepted and the command executes (`DRESULT succeeded`); an unlisted one is refused `policy_not_allowed`; an approved command (approve-each on) with a listed digest executes, with an unlisted one is refused; a queued command whose digest is allowed runs at execution; a safe stop with an unlisted digest runs; `CheckConnect` echoes the sorted list; the `session` record holds it in node and mixed mode (`connect_test.go`, `relay_mixed_test.go`, which now put the digest in the plan); the `policy-not-allowed` exchange fixture still passes |
| Replay determinism | A connected session in which a listed command executes: `run.ReplayFile` reproduces the trace digest and `verify` matches; the same session with an empty list has no effector call in its command log |
| `cli` | `plantwire connect --policy` is a usage error; `plantwire node --policy nope` is refused as bad options; the help text says "carries" |
| Contracts | Schema fixtures for the three schema changes; `test/contract` unchanged (plant wire untouched) |
| Bridge | `ParsePath` accepts `settings/default.policies.json` and refuses other names in `settings/` without the suffix; writing it creates the folder; a stale etag is `conflict`; no action argument resolves to `settings/` |
| Canvas unit | `planOf` with and without lines; issue placement per line; words; store pre-fills a new draft only, never an edited one; save is disabled until a fresh check, saves the check's list, reports conflict; decoders for `policies` and `policy_digest` (adapters ≥ 99 %); component tests with axe for the field, the check line, the digest line and copy status |
| e2e | `interrupts.spec.ts`: the plan lists `scriptPolicy` (export it from `e2e/support/session.ts:43`); the approved `ref-a` reads "Accepted" and "Board reports success", which is the first end-to-end proof that an approved command executes on a simulated device; `ref-b` stays denied; the saved run still `verify`s `matches: true` with an executed command in its log. New `policies.spec.ts`: a session with no list; the first command reads "Policy not allowed by this session" with `sha256:bbbbbbbb…bbbbbbbb`; Copy puts the full digest on the clipboard (granted clipboard permission); after a passing check, "Save as default" writes the file; a new form for the plant is pre-filled and says so |

## 10. Migration

None is needed: the project is pre-release. A plan without `policies` stays valid and means the same as today (trust
none). Callers of `plantwire connect --policy` (`connect_test.go:37`, `documentation/reference/cli.md:349`) move the digest
into the plan. Recorded sessions without `plan.policies` read "not recorded".

## 11. Rounds

| Round | Deliverable | Exit |
|---|---|---|
| **P1: contract and binary** | §3 schemas and embedded copies; `plantwire` domain `policies.go`; `CheckPlan`, `authorize`, `validate`, echo, session record, live digest; `connect --policy` removed; help text; `documentation/reference/cli.md` | Tests of §9 for model, plantwire, replay and cli; `make ci-check` green; `git diff main -- test/architecture scripts` empty or registration only |
| **P2: bridge and canvas** | §6 file class; §7 form, default, check line, list in effect, commands-table digest and copy; e2e; D-88 closed in `governance/DEFERRED.md`; an "As built" section in `docs/design/PHYSICAL_DEVICES.md`; §8 wording handed to R13 | `make phase-gate` green; Chrome design review at 1280 px against `governance/CANVAS_UI_ACCEPTANCE.md` §3, recorded in the round log |

## 12. Risks and security

- **An authorisation gate, not authentication.** The digest is a label the host copies into each command on an owner-only
  unix socket. The list protects a bench against a host acting under a policy it was not set up for. It does not protect
  against a hostile process that can reach the socket. Words never call it a signature.
- **The default must never silently widen trust.** It only pre-fills a new draft, visibly and with a note. It is saved only
  by an explicit action and only from a binary-checked list. It never changes a running session. A default that cannot be
  read pre-fills nothing, so it fails closed. The list in effect is shown in the check result, the live panel and the
  recorded session.
- **Old digests.** A narrowed policy gets a new digest. Leaving the old one listed keeps trusting the wider policy, so the
  procedure in §8 removes it.
- **Copying.** Copy is explicit. There is no "trust this digest" button, so learning a digest from a refusal is always a
  deliberate edit and a new session.
- **The record is not verified.** The `session` record is outside the stream digest, so `verify` does not detect an edit to
  it. That is acceptable, because it records configuration and determines no outcome (§3.5).

## 13. Remaining decisions

1. **The reference host without `--policy`.** `plantwire host --script` without `--policy` sends `policy_digest: ""`. The
   node refuses that as `malformed`, and the relay drops it silently (D-89). Recommendation: in P1, `plantwire host`
   refuses `--script` without a well-formed `--policy` (a usage check, through the same domain rule).
2. **Default per workspace or per plant.** Recommendation: per workspace, because a digest belongs to the host and the
   deployment, and a workspace is one deployment. A per-plant default can be added later as `settings/<plant>.policies.json`
   without a contract change.

## 14. Concerns

- **Firmware does not check policy.** Decision 5 says the gateway and firmware keep their own list. SPEC §13 and
  device-wire-v1 (no policy `reject_code`) make the check the gateway's only. `docs/design/PHYSICAL_DEVICES.md` §5 lists
  "policy digest" among the firmware's device-wire-v1 duties, which contradicts both. The briefs should name the gateway.
  A firmware list would need a device-wire change owned by Agentic Stream. PHYSICAL_DEVICES §5 should be corrected in P2.
- **Decision 6 goes beyond the owner's text.** The owner did not ask for `connect --policy` to be removed. I chose removal
  over a merge rule so that one session has one source and its record is unambiguous (AGENTS MUST 2). If the owner wants
  the flag kept, it can merge through `NewAllowList` with no other change.
- No conflict with determinism, the truth boundary, consumer knowledge or per-domain code was found. The list is operator
  data on the observer side, the binary names no host, and domains are untouched.

## As built (P1)

The binary half is built as §3 and §4 say; see "Policy allow-list P1 as-built (binary)" in `docs/design/PHYSICAL_DEVICES.md`.
Differences and additions:

- `Serve` checks the node's flag list (`policyOptions`) before it dials, so `plantwire node --policy nope` fails as bad options
  without a host; the relay's simulated node is checked at its start (`validate`), before the relay listens.
- Remaining decision 1 was built as recommended: `plantwire.CheckHost` (one `NewAllowList` rule) refuses a `--script` without a
  well-formed, non-reserved `--policy` as a usage error, and `ServeHost` applies it too (`ErrHostPolicy`).
- `sessionNote` also turns a nil `Policies` into `[]`, as it does for `instances`, so a hand-built check document still
  satisfies the journal schema.
- The `live` command's `policy_digest` is copied only when it is well formed (`refusedPolicy`), a second guard beside the
  observer's schema check.

## As built (P2)

The bridge and canvas half is built as §6 and §7 say; see "Policy allow-list P2 as-built (bridge and canvas)" in
`docs/design/PHYSICAL_DEVICES.md`. Differences and additions:

- The file class accepts exactly `settings/default.policies.json` (not a `*.policies.json` suffix in `settings/`), so nothing else
  can be written into `settings/` and no listing, change feed or action knows the folder.
- `ConnectStore.draftFor` stays synchronous. The pre-fill is `ConnectSession.prefill(domain)` (async, run when the panel
  opens), because the default is read through the bridge; the rule is the note's: only a draft the plant never held, never an
  edited one, and the note shows while the field is the default's.
- The workspace default state is `connect.defaults`; "Save as default" is also guarded in the store (a fresh passing check of the
  same plant), not only hidden in the form.
- "Set up another run" keeps the list and its source, so a digest added with "Use this digest" survives to the next plan.
- The review decodes `plan.policies` strictly; a `session` record whose `policies` is not a list of strings is dropped.
- Copy is a `Clipboard` port reached through the review session (`copyText`); the table also shows the full digest to select
  when the browser refuses to copy.
- R13 wording (§8) is recorded here for that round; the public docs are not written yet.

## Wording for the R13 briefs (from §8, unchanged)

- **Gateway brief**: the gateway keeps its own list of allowed policy digests in its own configuration and refuses a `DCMD` whose
  `policy_digest` is not on it with `DRECEIPT refused policy_not_allowed`, in the SPEC §9 check order; it skips the check for its
  devices' safe stops (the relay's emergency stop carries `sha256:` and 64 zeros); the list holds several digests so a narrowing
  can be rolled out; it logs the list in effect at start; device-wire-v1 has no policy reject code, so the firmware receives
  `policy_digest` unchanged and does not refuse on it.
- **Agentic Stream brief**: say where an operator reads the digest the host puts in every `DCMD` (a setting, a log line at start,
  or an API) and confirm that narrowing the policy gives a new digest. Until then the operator has two sources: the first refused
  command in the canvas commands table (shortened, with Copy), or the observer journal's `wire` records with direction `host`.
- **Bench check**: start the mixed session with the host's digest in the plan's allowed policy digests and in the gateway's list;
  send one command to a simulated device and one to the real board; both must read "Accepted"; if one is refused
  `policy_not_allowed`, its list differs (the row shows the digest the host used); after a narrowing add the new digest on both
  sides, then remove the old one.
