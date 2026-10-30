# Limitations and current status

> Status: Current working-tree report. Authority: repository state, tests, and implementation review. Verified by: full `go test ./...` run, `make ci-check`, and inventory review. Last verified: 2026-09-13.

This page is the honest counterpart to the product overview. It separates implemented capability from release evidence and from design intent.

## Current working-tree status

The current checkout implements the documented pipeline and passes the repository gates at its commit. The remaining observations are about release evidence and environment boundaries, not about a red suite:

| Observation | Evidence | Consequence |
| --- | --- | --- |
| Eight domains are shipped and validated together. | [`domains/`](../domains/) contains eight committed domain files; the embedded-schema suite validates each one. | New domains must land together with test, inventory, and manifest updates in the same change. |
| Some integration tests need local loopback sockets. | MCP operator endpoint tests and HTTP-push equivalence tests bind local listeners. | A restricted execution environment may fail those tests with `operation not permitted`; rerun in a network-capable environment before interpreting the result. |
| The checked-in release manifest is historical. | [`release-manifest.json`](../release-manifest.json) pins the commit and toolchain it was generated at. | Generate a fresh manifest for the exact commit before publishing evidence. |
| Design and implementation records disagree in places. | Earlier design text describes SQLite, Go 1.26, and commands not present in the current CLI; decisions and code record file artifacts and Go 1.25.12. | Treat archive design text as status-labeled source material, not an unqualified promise. |
| The repository lacks a formal compatibility matrix and release channel. | No current support matrix or release policy establishes one. | Validate the exact commit and toolchain for every external result. |
| Documentation checking is deliberately lightweight. | `make docs-check` checks required pages and executable smoke commands; it does not fetch links or parse every Markdown construct. | Review links, commands, fences, and diagrams manually as part of contribution review. |

## Scope boundaries

- The simulator is not a production broker, controller, physics engine, or hosted service.
- HTTP-push is an integration sink, not a deterministic local benchmark transport.
- The operator MCP role is not a standalone CLI mode; it is served by a director process.
- The score command reads `verdict.json` and `ledger.jsonl` beside the run artifact and `label.json` (or `labels.json`) beside it unless `--label` is provided. It does not accept a `--verdict` flag.
- The offline score command does not receive an effector-call log, so action-loop metrics are incomplete. Use the MCP director score path when action evidence is required.
- Suite generation is not a complete batch runner: it produces audited scenario definitions and labels, while a harness must choose an adapter, execute scenarios, and persist score output.
- The CLI replay and verify commands replay the domain and adapter specifications recorded in the run artifact, so editing a domain afterwards does not break earlier runs; the result reports `domain_changed` when the domains directory now holds a different digest.
- The current code uses file-based JSON/JSONL run artifacts; older SQLite language in the design archive is historical or superseded by the accepted decision.
- No actual external production consumer is shipped; the reference consumer is a test instrument and integration example.

The `streamsim device serve` command is a deterministic emulator gateway link,
not a physical-HIL claim. Its capability catalog is a simulator-side device
contract and must be kept aligned with the paired consumer's catalog and
allow-listed digest before a cross-repository handshake can be treated as
evidence. When `--world` is supplied, the CLI binds the emulator to a
simulated world plant through `internal/deviceworld`; without it, the device
uses its local catalog projection. World bindings are loaded from structured
configuration; the adapter does not contain per-domain argument callbacks.

## Physical devices


### What you cannot treat as hardware evidence

| Limitation | Why it matters | What to do |
| --- | --- | --- |
| No test here exercises a real microcontroller or a real gateway. The stand-in gateway is test support: it sends no readings and refuses commands, so the twin report's end-to-end test can only reach "cannot judge", and the "effect observed" stage of a command on the live screen is covered by unit and component tests with documents shaped by the contract, not by an end-to-end test. | A green test suite says the simulator handles the contract, not that a given board and gateway do. | Run the owner's bench checks and keep the evidence they ask for. |
| A `DRESULT` is the board's own report, and a receipt is the gateway's. The effect stage is `observed` only on host evidence about a metric of the commanded device, and the host's check of a value the board itself reported is not independent. A hardware gateway raises no barrier, so a hardware-only session carries no host evidence at all (D-111, D-112). | "Succeeded" and "observed" are not a physical effect. An independent instrument is needed to say the fan really turned. | Treat the effect stage as the host's recorded decision; confirm the physical effect on the bench with an instrument that does not share the board's sensing path. |
| A real device has no truth and no score. The plausibility bounds of a reading come from the gateway's own birth, so a hostile gateway sets its own (D-78). | A consumer cannot be graded on a board's world, and the live view shows what the gateway declared. | Use simulated devices for graded runs; use hardware for integration evidence. |
| The contract version and digest must match exactly (D-113). A gateway on another version is refused `contract_unsupported`. | A gateway built against an older contract cannot connect, even if it would otherwise work. | Vendor [`contracts/plant-wire/`](../contracts/plant-wire/SPEC.md) byte for byte. |

### Time and controls

| Limitation | Why it matters | What to do |
| --- | --- | --- |
| A mixed session whose plan has no `pace: "wall"` is not paced to the wall clock (the canvas always paces its connected sessions; `plantwire node` and a gateway-less `plantwire relay --domain` are never paced). The simulated world then advances barrier by barrier as fast as the host answers, so a simulated board beside a real one runs ahead of the hardware's wall time. In a paced session or not, if the host is away for three barrier caps (30 s from the canvas, which uses a 10 s cap; 90 s from the command line, whose default cap is 30 s) the world ends and its devices are reported offline while the real board keeps passing (D-71). | Unpaced, simulated and real time series cannot be compared in real time; paced or not, the world is not reborn when the host returns. | Pace live sessions (`pace: "wall"`); keep the host attached. |
| Controls are served at a barrier (the simulated world) or on a 250 ms tick (the relay), not at once. A stop can wait for the slice in progress, a stalled host for up to one bounded 2 s write, and a request the node cannot take within 10 s is answered `node_busy`, which may still be applied afterwards (D-81, D-98, D-99, D-106). | An emergency stop acts at the next barrier or tick. | Use short slices (`--slice-s`); read the state again after `node_busy`. The board's own leases and safe state are the last line of defence. |
| A command the gateway already accepted cannot be cancelled by the node: a hold or an emergency stop refuses what the host sends from then on, but a command the gateway queued for its `not_before_ns` or has in flight still runs (D-94). | A held plant may still act once. | Prefer short leases on commands; stop the board at the gateway if a queued command must not run. |
| The relay's emergency stop is the declared safe-stop command without arguments, sent to the last boot the gateway announced with the host's last authority epoch, a 30 s deadline and the reserved all-zero policy digest. A device that rebooted since refuses it `wrong_boot`; a device never born is skipped; a safe stop that needs arguments is held like any command (D-82, D-99, D-100, D-105). | A stop can fail, and the reply says which devices were `unreachable`. They may still be energised. | Read the reply's `unreachable` list; stop those devices at the gateway. |
| An emergency stop with nothing to stop reads as applied with empty lists, and after a partial stop a device that could not be ordered shows the same Latched badge as a stopped one (D-104, D-107, D-108). | A quick glance can read as a clean stop. | Read the outcome text and the `unreachable` list, not the badge. |
| A command waiting for approval is refused `session_ending` when the host is lost, never kept for a returning host; turning approval mode off does not release commands already waiting; a hardware command waits against the relay's wall clock, so clock skew with the host expires it early or late (D-83, D-97). | Approvals do not survive a reconnect. | Approve or deny before the host link drops. |
| Only one control is in flight at a time in the canvas, except an emergency stop, which is always sent (D-85). | A second control pressed while one waits does nothing. | Wait for the outcome line. |
| The scope select offers what the newest live snapshot names, so a board that has not been born cannot be held in advance from the canvas (D-84). | A hold cannot be set before a board connects. | Use the control socket directly. |
| Bounds of the interrupt rules: at most 64 scopes are set (holds and approval scopes together), at most 64 commands wait for approval, each with at most 8 repeats, and a node has at most 4096 devices; beyond that a control is refused `too_many_scopes` or a command `not_ready`. | A very large plant cannot be held scope by scope. | Hold by board or by `all`. |
| A `release` clears exactly the scope it names: releasing `all` does not clear a hold on a device or a board, or the latch of a stop on one. | A latched device can stay refused after a release of a wider scope. | Release the scope that was held or stopped; the state lists them. |
| Controls exist on `plantwire connect` only; `plantwire node` and `plantwire relay` take none (D-83). | The standalone relay has no interrupts. | Use `plantwire connect` with a gateway. |
| No authentication and no operator identity on the control socket or the bridge beyond owner-only directories and the bridge's pairing code (D-86). | Any process of the same user can hold or stop the node, and the log does not say who. | One user, one desktop. |

### Connections

| Limitation | Why it matters | What to do |
| --- | --- | --- |
| The relay serves one gateway at a time, and a gateway record that arrives while no host session exists is dropped, not held. Store-and-forward is the gateway's job (D-70). | Two gateways behind one node are not supported, and a host outage loses data unless the gateway buffers it with `historical: true`. | One gateway per session; buffer in the gateway. |
| One connected session runs at a time in the bridge. | A second `session.start` is refused `session_running`. | Stop the running session first. |
| The reference host sent no heartbeat, so a host that died without closing its socket was not noticed and a hardware gateway declared it lost after 3 s (D-69, fixed: it beats every second while a session is live). | Host loss from silence is detected only for hosts that send heartbeats. | Use a host that sends them. |
| Nothing limits the rate of rebirth requests a flapping gateway or a looping host causes (D-74). | A flapping link produces repeated births. | Fix the link. |
| A gateway that sends no admitted record for 30 seconds while its host is present is treated as lost (D-73, fixed). While the relay has no host (before the first and during a host loss) the watchdog is suspended; the silence counts on when the host returns (a host cannot reset it, so a flapping host does not keep a dead gateway), and the gateway is closed only after the host has been back 5 s, time to answer its rebirth request. | A slow gateway must still send an admitted record, readings for example, at least every 30 s while its host is present. | Keep the gateway sending. |

### Recording and replay

| Limitation | Why it matters | What to do |
| --- | --- | --- |
| A connected session's planned disturbances are recorded in `sealed/journal.jsonl` as they are injected, in plain text in the run folder (owner-only, mode 0600), so the same operating-system user can read the schedule while the session runs; "sealed" is a convention of the run folder, not an access boundary (review B finding 5). | A host process running as the same user, an agent with a shell tool for example, could learn a disturbance it is meant to meet unannounced. | Run the host as another user, or keep the run folder out of the host's reach (outside its workspace); the canvas writes runs into its own workspace. |
| `plantwire replay-trace` plays the gateway's side of a recording. The simulated part of a mixed recording is reproduced and verified by `verify`, not streamed to a host beside the trace, and a replay refuses every command (D-72). | A consumer cannot be shown a recorded mixed session end to end. | Use `verify` for the simulated side and `replay-trace` for the board's. |
| A hardware-only relay (`plantwire relay` without `--domain`) writes the field trace and journal but no `run.json`, so the trace has no digest cover (D-75). | A later edit of the trace is not detected. | Use `plantwire connect`, which writes `run.json`. |
| `origin`, `binding`, `session`, `live` and `interrupt` records are outside the stream digest, are not regenerated by a replay, and are not integrity-protected (D-76, D-79). The rules of `interrupts.jsonl` are replayed in-process by the session that wrote it; there is no command that re-runs them over a folder received elsewhere (D-80). | A recording is not proof of what a board did if the folder was edited, together with its `run.json`. | Keep recordings together with their digests; receive them through a trusted path. |
| An unpaced session (no `pace: "wall"`) against a host that answers every barrier at once (the reference host with no `--delay-ms`) writes about 75 MB of run folder per second; `plantwire twin-report` refuses a run whose `journal.jsonl` is over 64 MiB, and a field trace or an interrupt log is capped at 64 MiB too (D-77). | A run folder from such a session is too large to report on or replay in practice. | Pace the session, or give the reference host `--delay-ms` (for example 250); stop sessions in minutes. |
| The reference host exits 1 without a report when the node is stopped during a `--delay-ms` wait (`write ... broken pipe`), and when it is interrupted before any node arrived (D-115). | A bench script that waits for the host's report gets none. | Take the evidence from `journal.jsonl`. |
| A multi-hour session with many devices grows the journal and the canvas history linearly (D-77). The live projection lists the latest 128 commands and ignores a re-sent id only within the last 4096 (D-78). | Long sessions get slow to review; very old command ids can be listed twice. | Keep sessions short; restart for a long soak. |

### Twin versus board

| Limitation | Why it matters | What to do |
| --- | --- | --- |
| The twin report refuses a field trace whose readings span more than 7 days (`ErrTwinSpan`, because the twin integrates from its start to every reading) and a run whose `journal.jsonl` is over 64 MiB. | A long or runaway recording cannot be compared. | Split the session; keep sessions short. |
| The twin starts at the earliest reading in the trace with the initial state its domain declares, not the board's state then. Readings before the first applied command are not judged, and a run with no applied command is `cannot_judge` (D-90). | A warm board against a cold-start twin shows as a divergence or no judgement. | Make the twin's initial state match the bench start, and send a command first. |
| A command is placed at the latest board event time recorded before the gateway's receipt for it, so it is exact only to one reading interval, and a board that stamps its own events can move where a command lands (D-90, D-96). | Small timing differences between the twin and the board are noise. | Choose tolerances wider than one reading interval's change. |
| Only the host's commands to the board's devices are applied; commands to simulated devices of a mixed session and the history before the recording are not. A board device with no matching twin entity is `no_twin_entity`; a board that rebooted mid-session is not told apart (D-91). | The twin may miss causes. | Compare sessions where the board is driven only by commands in the recording. |
| `agrees` also holds with a tolerance as wide as the metric's whole range, and with one reading compared out of thousands. The report carries no coverage figure (D-95). | `agrees` is not proof. | Read the per-metric `compared`, `not_compared` and tolerance, not the verdict alone. |
| The Inspect overlay draws the report's series (at most 2000 points, the first ones), in the board's own time base, which the canvas treats as run time. The canvas offers no span-fraction field or per-metric tolerance form (D-92, D-93). | A truncated series draws only the start; a gateway whose clock is not the run clock shifts the chart. | Use the command line for other tolerances; stamp the board's events on the run clock. |

### Defining and capturing devices

| Limitation | Why it matters | What to do |
| --- | --- | --- |
| The capture screen shows where a draft differs from the saved kind file but not the two values, and its run picker lists every run (D-67, D-68). | A reviewer resolves the pointer in the file by eye. | Open both files. |
| The default instrument for a reading is chosen by unit and range, so units shared by several quantities (`%`) are told apart only by name (D-65). `kind twin --kinds-dir` loads the whole directory and refuses it for one malformed kind file (D-66). `device serve` runs a domain without the `implements` check, and some loaders skip it on purpose (D-56, D-57). The kind schema does not constrain a unit or the number of safe stops of a kind (D-59). | A twin may be given a poor default instrument; an unfinished kind in a directory blocks the twin. | Review the suggested instruments; pass kind files as arguments. |
| The canvas device list logs a 404 for each kind without a twin choices file (D-101). | Console noise only. | None needed. |
| The second proof, a second board through the same steps, has not been done. | The workflow is shown on one reference board. | The owner's bench checks and a second kind file. |

### Policy allow-list

| Limitation | Why it matters | What to do |
| --- | --- | --- |
| The simulator checks the policy digest of simulated devices only. For a real board, the gateway's own list is the only check; the firmware's wire has no policy refusal code. | A session that lists the digest still depends on the gateway's list for the real board. | Confirm both lists name the host's digest before the bench check. |
| The list is an authorisation gate on an owner-only socket, not authentication. A relay drops a schema-invalid host record, including a command with an empty digest, without a reason in the evidence (D-89). A policy digest wraps inside a 336 px panel and the footer shows only the first three (D-110). | A host that stalls with no visible reason, or a digest that is hard to compare by eye. | Use Copy policy digest; compare the full text in the field. |

### Canvas

| Limitation | Why it matters | What to do |
| --- | --- | --- |
| Desktop only, 1280 px or wider. | The shell does not adapt to narrower windows. | Use a wider window. |
| The live view is a projection of what the host saw; a snapshot that cannot be decoded is shown as unreadable, not skipped, and the note names a snapshot even when the dropped record was an interrupt record (D-87). A control reply is ordered against snapshots by content, not by time (D-102). Node tags on the plant use short words (D-103). The frame is transiently clipped for up to a second after the Interrupts bar resizes (D-109). Overlapping timeline pills are not de-collided, and while events arrive the live scrubber lane grows and moves the plant's top. | Cosmetic and ordering edge cases. | Reload the view. |
| The stand-in gateway used by the end-to-end tests ends by itself after about five minutes. It is test support, not a feature. | Long bench sessions with it need a restart. | Use a real gateway for long sessions. |

## What is not a release claim

The presence of tests for all nine non-negotiables is valuable evidence, but it is not itself a green release gate. A release claim requires a clean run of the required checks, a matching manifest, reproducible artifacts, and an explicit review identity.

## How to clear the limitations

1. ~~Decide whether `cold-chain-transit` is shipped~~ Done: the domain is committed and validated with the rest of the inventory.
2. ~~Reconcile stale design, README, context, and decision text~~ Done for the public set and agent instructions; archive text under `docs/` remains status-labeled by decision.
3. ~~Add or document compatibility, support, conduct, changelog, and release policy~~ Done: see [compatibility](overview/compatibility.md), [SUPPORT](../SUPPORT.md), [CODE_OF_CONDUCT](../CODE_OF_CONDUCT.md), [CHANGELOG](../CHANGELOG.md), and the [release procedure](operations/release.md).
4. Extend `docs-check` with local link/fragment and example checks, then keep it in CI; manual diagram and prose review will still be required.
5. Add a suite runner that records the selected adapter and persists per-scenario score evidence.
6. Generate fresh release evidence (gates, soak, perf, manifest) from a clean, intentional commit at publication time.

## Next reads

- [Roadmap](roadmap.md)
- [Release procedure](operations/release.md)
- [Benchmark evidence](benchmark/evidence.md)
- [Quality bar](QUALITY_BAR.md)
