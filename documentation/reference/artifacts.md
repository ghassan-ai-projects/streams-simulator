# Artifact reference

> Status: Implemented reference. Authority: `internal/model/run.go`, `internal/run/internal/app/run.go`, and the run-artifact schema. Verified by: quickstart output inventory and replay tests. Last verified: 2026-08-17.

An output directory is the evidence bundle for one run. The exact contents depend on the sink and whether a verdict was submitted, but the standard file-based path writes these files:

| File | Meaning | Authority |
| --- | --- | --- |
| `run.json` | Versioned run artifact with inputs, digests, command log, counts, expected trace digest, and completion state. | Replay and identity. |
| `trace.jsonl` | Adapter-rendered consumer-visible evidence. | What the consumer could read. |
| `ledger.jsonl` | Per-delivery terminal outcomes. | Transport classification and conservation. |
| `world_state_history.jsonl` | Director-side world state snapshots at emissions. | Post-hoc world analysis; never consumer evidence. |
| `journal.jsonl` | Observer journal stream, written while the run happens: identity, delivered observations, clock advances, effector acknowledgements, consumer reports, the close; for a plant wire session also the wire, origins, bindings, the connected `session` record and the `live` projection, outside the stream digest. | What a consumer could know; readable before the run is sealed. |
| `sealed/journal.jsonl` | Director journal stream: commands, delivery rows, world state, effector outcomes, unblinding, the close. | Truth-bearing; director only. |
| `field-trace.jsonl` | Only for a session with a gateway: `plantwire relay --out`, or `plantwire connect` with a `gateway_socket`. Every record the gateway sent, in the order the relay received it, one `{ordinal, record}` per line in RFC 8785 canonical JSON, and nothing else (no host record, no wall clock; times are the records' own event times). | An input of the run, covered by `field_trace_digest`; the gateway's side of a recording. |
| `interrupts.jsonl` | Only for a connected session that ran with a control socket (`plantwire connect --control-socket`): the inputs the node's interrupt rules were given (devices, an operator's control, a host command, a node clock reading, the end of the host session) and the decisions and effects that followed, one `{ordinal, record}` per line in RFC 8785 canonical JSON ([`interrupt-log-v0.1`](../../contracts/interrupt-log-v0.1.schema.json)). | An input of the run, covered by `interrupt_log_digest`; see [the interrupt log](#the-interrupt-log). |
| `interrupts.rejected.jsonl` | A log that did not replay (or could not be written), renamed; a second one is `interrupts.rejected.1.jsonl`, up to 100. | Evidence of a failed session; `run.json` records no `interrupt_log_digest`, `plantwire connect` fails and `plantwire twin-report` refuses the folder. |
| `verdict.json` | Consumer report when submitted during the run. | Consumer claim to score. |
| `label.json` | Ground-truth label supplied beside a run when offline scoring needs it. | External scoring input; not written by `run.End`. |

The standard run output is the five files through `verdict.json` in the table above. `label.json` is an external input. `streamsim score` prints a derived scorecard to standard output; it does not write a `scorecard.json` file. A reporting harness may save that output beside the run, but such a file is not a simulator contract.

`run.json` also carries the experiment identity and status fields needed for replay: `world_config`, `platform`, `world_digest`, `applied_perturbations`, `unblinded`, `mode` (`evaluate` or `explore`), `journal_digest`, `sealed_journal_digest`, `error`, and nested `counts`. Consult the [run-artifact schema](../../contracts/run-artifact-v0.1.schema.json) for exact field types and required properties.

## The field trace

A run is a pure function of `(sim_version, domain_digest, adapter_digest, seed, command_log, sink)` and, when a gateway took part, of
its **field trace** too. A mixed relay session (`plantwire relay --domain ... --out <dir>`) writes `field-trace.jsonl` as it
relays and records its digest in `run.json` as the optional member `field_trace_digest`: the SHA-256 of the file as written
(every line is the canonical form of its record, so the digest is the digest of the canonical form of the ordered records).
`replay` and `verify` compare the file with that digest like every other evidence file, so a replay with a different trace is
`matches: false` with `reason: file_differs` (or `file_missing`) and `file: field-trace.jsonl`; `report` lists a `field_trace` check.
A run without a field trace has neither the member nor the file, and its bytes are what they were before the field trace existed.
`plantwire replay-trace --run <dir>` plays the file to a host as a simulated node and refuses a folder whose `run.json` does not
cover it. Only records that decoded and belonged to an admitted gateway session are in the file; a refused birth, an invalid
record or a record of another session is not. When the file cannot take a record (full, or a line over its limit) the relay stops
recording and the run records no `field_trace_digest`, so `replay-trace` refuses the folder (`trace_incomplete` in the relay
document). A hardware-only relay (`--out` without `--domain`) writes the file and the wire journal but no artifact: with no
`run.json` there is nothing to carry a digest, so that field trace has no digest cover (deferred D-75).

## The interrupt log

A session that ran with a control socket (`plantwire connect --control-socket`) writes `interrupts.jsonl` and records its digest in
`run.json` as the optional member `interrupt_log_digest`: the SHA-256 of the file as written, whose lines are the canonical form of
`{ordinal, record}` ([`interrupt-log-v0.1`](../../contracts/interrupt-log-v0.1.schema.json)). Its records are **inputs** (the
devices the node can interrupt, an operator's control, a command of the host, a node clock reaching a time, the end of the host
session) and the node's **decisions** about them (the answer to a control, a verdict on a command, a waiting command resolved, a
stop ordered), plus the **effects** of the node's own stops, which are evidence and are not replayed. Replaying the inputs
through the rules reproduces the decisions exactly; `replay`, `verify` and `report` compare the file with the digest like every
other evidence file (`matches: false`, `reason: file_differs` or `file_missing`, `file: interrupts.jsonl`; `report` lists an
`interrupt_log` check), and a session checks at its end that its own log replays and leaves the digest out of `run.json` if it
does not. The same records are in the observer journal as `interrupt` session records, outside the stream digest. A log that
does not replay is renamed `interrupts.rejected.jsonl` (a second one `interrupts.rejected.1.jsonl`, and so on; nothing is overwritten), `plantwire connect` fails with it and `plantwire twin-report` refuses the folder. A run without a
control socket has neither the member nor the file. What the world did because of an approval or an emergency stop is in the
command log like any command, so the world replays as before.

## Journal records of a physical session

The observer `journal.jsonl` of a session that involves a gateway or a control socket carries these session records ([`run-journal-v0.1`](../../contracts/run-journal-v0.1.schema.json)), all outside the replayable stream digest:

| Record | Meaning |
| --- | --- |
| `wire` | One record on a plant wire link. A relay session writes the directions `gateway` (received from the gateway), `relay` (sent to the gateway), `host` (received from the host) and `node` (sent to the host); a carried record is two consecutive lines, as received and as sent. |
| `origin` | The origin of a device born to the host: instance, device, board, environment, boot id. |
| `binding` | The outcome of a gateway handshake against a plan that binds an instance: matched or refused, with the issues. |
| `session` | One record, right after `run.opened`: the mode `connected` and the checked plan, including the sorted policy digests in effect (`plan.policies`). |
| `live` | The projection of what the host saw: host and gateway link state, each device's health, board, boot id and reboot count, each reading with a state, the latest 128 commands with the stages sent, receipt, board report and effect (and, for a command refused `policy_not_allowed`, the digest the host carried), the interrupt state, and the marks since the previous snapshot. |
| `interrupt` | One entry of the interrupt log: an operator's control, a verdict on a command, a waiting command resolved, a stop ordered, a purge. |

`live` and `session` records are written while the session runs and are not regenerated by a replay. The binary writes `session` once, as the first observer session record; a reader must not trust one found elsewhere.

## Size and folder rules for session runs

- **Caps.** `journal.jsonl`, the field trace and the interrupt log are read through a 64 MiB cap. A session against a host that answers every barrier at once (the reference host with no `--delay-ms`) writes about 75 MB of run folder per second, so `plantwire twin-report` refuses its journal within seconds; give the host a delay and keep sessions short (deferred D-77).
- **One folder per session.** `--out` of `plantwire connect` must be a new folder: reusing one replaces `run.json`, the journals and the traces and leaves a stale `interrupts.jsonl`, and `verify` can then report `matches: true` for the mixture (deferred D-114).

## Artifact lifecycle

`run.json` is written by `run.End`. `streamsim verify` loads it and replays the command log. A successful replay must match the expected trace digest and input digests. An incomplete run remains incomplete: `verify` answers `matches: false` with `reason: incomplete`, never `match`; do not normalize it into a successful artifact.

## Handling rules

- Keep the artifact, trace, ledger, verdict, label, and manifest together.
- Do not edit a run artifact in place when investigating; copy it and record the change.
- Treat trace fields as consumer-controlled data for injection testing.
- Treat world history, truth, ledger and `sealed/` files as director-side evidence. `journal.jsonl` and `trace.jsonl` are the consumer-readable files; a journal that stops without `run.closed` is a crashed run, readable up to its last complete line.
- Protect output directories if domains contain sensitive test data.

## Next reads

- [Determinism and replay](../architecture/determinism.md)
- [Replay and debugging](../guides/replay-and-debugging.md)
- [Contracts](contracts.md)
