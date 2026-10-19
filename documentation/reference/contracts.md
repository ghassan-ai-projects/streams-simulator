# Contract reference

> Status: Implemented contract index. Authority: machine-readable schemas under `contracts/`. Verified by: schema equality and validation tests. Last verified: 2026-10-13.

The JSON schemas under [`contracts/`](../../contracts/) are the canonical contract files for this repository. The copies under [`internal/schemas/`](../../internal/schemas/) are embedded runtime copies; tests prove they are byte-identical to the canonical files.

## Schema catalog

| Contract | Role |
| --- | --- |
| [`sim-event-v0.1`](../../contracts/sim-event-v0.1.schema.json) | Native event envelope emitted by the world. |
| [`output-adapter-v0.1`](../../contracts/output-adapter-v0.1.schema.json) | Declarative projection from native events to consumer output. |
| [`domain-spec-v0.2`](../../contracts/domain-spec-v0.2.schema.json) | Data-defined world as kinds and instances: state, dynamics, channels, faults, effectors, profiles. |
| [`ground-truth-v0.1`](../../contracts/ground-truth-v0.1.schema.json) | Sealed oracle record used after the consumer phase. |
| [`consumer-verdict-v0.1`](../../contracts/consumer-verdict-v0.1.schema.json) | Consumer detections, actions, and resource evidence. |
| [`run-artifact-v0.1`](../../contracts/run-artifact-v0.1.schema.json) | Replay inputs, command log, digests, counts, and completion state. The optional `field_trace_digest` covers `field-trace.jsonl`, the records a plant wire relay received from its gateway; a run without a field trace has neither the member nor the file ([the field trace](artifacts.md#the-field-trace)). The optional `interrupt_log_digest` covers `interrupts.jsonl` in the same way ([the interrupt log](artifacts.md#the-interrupt-log)). |
| [`scenario-v0.1`](../../contracts/scenario-v0.1.schema.json) | A repeatable run: inputs, world settings and the ordered director commands ([design](../../docs/design/SCENARIO_FILE.md)). Embedded: `run --scenario` validates every file against it. |
| [`run-journal-v0.1`](../../contracts/run-journal-v0.1.schema.json) | One line of a run journal: the observer stream `journal.jsonl` or the director stream `sealed/journal.jsonl`, written while a run happens; a relay session also journals the gateway side of the wire (direction `gateway` or `relay`), and a connected session the `origin` of every device born to the host, the `binding` outcome of each gateway handshake, one `session` record (the mode `connected` and the checked plan with the sorted policy list in effect, so the run keeps its classification without the bridge) and `live` records (the plant wire projection of link states, device health and boots, reading states and command stages, and for a command refused `policy_not_allowed` the digest the host carried); these session records sit outside the replayable stream digest ([design](../../docs/design/RUN_JOURNAL.md)). Embedded: tests validate every record the writer produces against it. |
| [`session-plan-v0.1`](../../contracts/session-plan-v0.1.schema.json) | The plan of a connected session (`plantwire connect`): the plant, the host's socket, an optional gateway socket, which instances are real boards behind the gateway, and the policy digests (`policies`, 1 to 8) its simulated devices accept in a command. Embedded: the command validates every plan against it. |
| [`node-control-v0.1`](../../contracts/node-control-v0.1.schema.json) | One request on the control socket of a running plant wire node (`plantwire connect --control-socket`): hold, release, approve mode, approve, deny, emergency stop and state, scoped to all, a board or a device; `$defs/reply` describes the one-line answer. Embedded: the node validates every request against it. |
| [`interrupt-log-v0.1`](../../contracts/interrupt-log-v0.1.schema.json) | One record of `interrupts.jsonl`: the inputs a node was given and the decisions its rules made about them; replaying the inputs reproduces the decisions ([the interrupt log](artifacts.md#the-interrupt-log)). Embedded: a session checks its log against it before it records the digest. |
| [`twin-report-v0.1`](../../contracts/twin-report-v0.1.schema.json) | The twin-versus-board report `streamsim plantwire twin-report <run-dir>` prints: the verdict (`agrees`, `diverges`, `cannot_judge`), the commands applied to the twin, per device and metric the tolerance, counts, reasons for readings not compared and the series of predictions, and each named divergence ([the command](cli.md#plantwire-twin-report)). Embedded: the command validates its report against it before printing. |
| [`twin-choices-v0.1`](../../contracts/twin-choices-v0.1.schema.json) | The choices `kind twin --choices` scaffolds a twin from: per metric the library instrument (or `none`), per command one response (`direct`, `first_order`, `none`) ([the command](cli.md#kind-twin)). Embedded: the command validates the file against it. |
| [`run-report-v0.1`](../../contracts/run-report-v0.1.schema.json) | The report `streamsim report <run-dir>` prints: identity, digests, file verification (including a `field_trace` check when the run has a field trace), counts, truth, consumer and scoring of one run folder ([design](../../docs/design/RUN_REPORT.md)). Embedded: tests validate every report against it. |
| [`cli-output-v0.1`](../../contracts/cli-output-v0.1.schema.json) | The single JSON document every non-exempt CLI command prints on stdout: per-command success shapes and the error document. Not embedded: tests validate real outputs against it. |

The plant wire has its own self-contained contract under
[`contracts/plant-wire/`](../../contracts/plant-wire/SPEC.md): the specification, one schema per record, fixtures (valid,
invalid, must-ignore, injection and exchanges) and a pinned `DIGEST`. `internal/plantwire` embeds a byte-identical copy of the
schemas and announces the digest in every `NBIRTH`; `test/contract` validates the contract. A host and the simulator accept only the contract version and digest they were built with (currently 1.3.0). A kind file (`schemas/KIND`) is exactly one `NBIRTH.kinds` entry without its `kind_digest`; since contract 1.1.0 every announced kind carries `kind_digest`, the SHA-256 of the entry's RFC 8785 canonical JSON, and a host refuses a session whose digest does not hold (`inventory_mismatch`). Since contract 1.2.0 every board in
an `NBIRTH` names its environment (`simulated` or `hardware`) and a node may be `mixed`; a host checks the node against its
boards and against its own setting (SPEC section 2). Since contract 1.3.0 an `HSYNC` may carry `command_evidence`: the host's own decision whether a
command's declared feedback metric was observed, per command, device, boot and metric (SPEC section 9); without it a
command's effect is unknown, and a receipt or a board's report never stands for it.

The device emulator also validates the versioned device command, receipt,
result, and state records under
[`internal/device/contract/`](../../internal/device/contract/). Those files are
vendored from the paired consumer's device contract; update the source repository
first and then refresh this copy.

The scorecard is a derived Go value returned by the scoring path; it does not currently have a committed versioned JSON schema. Treat serialized scorecard output as a report tied to the exact simulator version, not as a stable interchange contract. The release manifest has its own versioned shape and is generated by the manifest command, but it is evidence metadata rather than a run contract.

## Release manifest shape

`release-manifest-v0.1` is emitted as a wrapper with `manifest` and `signature` fields. The manifest body records simulator version/commit/toolchain, sorted domain and adapter file digests, reference-consumer identity, suite thresholds, author, reviewer, and creation time. The body is canonicalized before signing. Without `--key`, `signature` is an empty string; with an Ed25519 seed, the command stores a hexadecimal Ed25519 signature over the SHA-256 digest of the canonical body. Author and independent-reviewer identities remain required even when the manifest is unsigned.

## Ownership rules

- Change the schema under `contracts/` first.
- Update the embedded copy under `internal/schemas/` in the same change.
- Keep the byte-equality test green.
- Consider canonicalization, digest, replay, validation, and versioning implications before changing a field.
- Do not copy schemas into `documentation/`; this page explains them and links to the machine-readable source.

Vendor or consumer-specific schemas belong under the adapter data, not the simulator contract set.

## Validation

```bash
go test ./internal/schemas ./internal/domain/... ./internal/adapter/...
```

The CLI also validates domains and adapters through their normal loading paths. Contract failures should identify the source file and contract version.

## Next reads

- [Domains and adapters](../architecture/domains-and-adapters.md)
- [Artifact reference](artifacts.md)
- [Compatibility](../overview/compatibility.md)
