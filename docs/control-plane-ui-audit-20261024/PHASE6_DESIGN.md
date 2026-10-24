# Phase 6 design: engineering and provenance (round R2, Go)

Findings F22, F24, F25 (read side), F26; capability OP08. The canvas renders what the binary says; it computes none of this
(AGENTS.md MUST 1 and 2). Two read-only use cases, one new contract each.

## Placement

No new module, package or top-level folder, and no allow-list entry.

| Use case | Owner | Why |
|---|---|---|
| Assumptions of a domain or twin | `internal/component` (domain layer) | The module owns the twin scaffold (`ScaffoldTwin`), the instrument library and `Compose`; the assumptions of a twin are read from the same build that writes the twin, so they cannot drift from it. Already imports `domain`, `model`, `schemas`, `canonical`. |
| Lineage of a plant | `internal/component` (domain layer; `files` edge reads the workspace folders) | The relation joins what `component` already owns: compositions (`Compose`), twin choices (`ScaffoldTwin`) and device kind files (read by `ScaffoldTwin`'s callers). Runs enter only as the domain digest in `run.json` or in the first observer line, decoded with the shared `model.RunArtifact` record. `report` and `run` cannot be used: neither imports `component`, and adding the edge would grow an allow-list. |

`internal/kernel` gains one generic read, `DirectoryNames` (the kernel lists plain files only and a run folder is a directory).
It knows no file name. `internal/cli` calls one `component` facade function per verb and prints the document;
`internal/canvasbridge` registers two CLI-backed actions.

Alternatives rejected: a `lineage` module (grows `internal/cli`'s allow-list and needs the owner); `report` (imports neither
`domain` nor `component`); `director` (a world and job registry, not a file relation).

## Contracts

Both are embedded (`internal/schemas`) and validated in production by the producing use case through
`model.ValidateDomainAssumptions` and `model.ValidateDomainLineage` (as round 4c does for `run-comparison-v0.1`), so no test
hook is needed. A failure to validate is a defect of the binary, reported as `domain_failed`.

### `domain-assumptions-v0.1`

`domain` (`id`, `version`, `digest`), `origin` (`kind` `twin` | `composition` | `domain`; for a twin file given with the
domain, `twin` `{id, version, status}` where `status` is `in_sync` or `out_of_sync`), `kinds[]`, `not_modeled[]`.

Per kind: `kind`, `implements`, `instances[]`, `readings[]`, `dynamics[]`, `commands[]`, `faults[]`.

- A reading states what the consumer receives: `channel`, `unit`, `value_type`, `range`, `resolution`, `initial`, `noise`,
  `cadence`, `availability`, `absence`, `tag`, and where it came from: `basis` (`library_instrument`, `plain_channel`,
  `domain_declared`), the `instrument` `{id, version, digest, ref}` and `recorded_in` (`composition` | `twin_choices` |
  `domain_file`).
- `equals_binary_suggestion` is present only for a twin that is in sync and whose kind files were given. It says the recorded choice
  equals what `kind twin --suggest` proposes. The twin-choices file does not record whether a person chose or accepted a
  default, so the document never says "chosen" or "accepted": "equals the suggestion" is a fact, the rest would be a guess.
- A command states `response` (`direct`, `first_order`, `none`, `fixed_safe_stop` for a twin; `declared` otherwise), `recorded_in`,
  `sets`, `drives` (gain and time constant, twin only), `effect`, `acknowledgement`, `lease` and a `statement`.
- `not_modeled[]` are `{code, scope, sentence}` with stable codes: the twin allow-lists (`kind_file_members_not_reproduced`), the
  response model (`first_order_response_only`, `device_behavior_not_modeled`, `safe_stop_is_a_step`, `command_changes_nothing`,
  `acknowledgement_never_fails`), `saturation_not_declared` per dynamics without a clamp, `no_measurement_noise` per reading
  without noise, `instrument_semantics_unchecked` and `no_calibration_evidence`.

### `domain-lineage-v0.1`

`plants[]` (one per domain file, by id), `device_kinds[]` (one per kind file), `unreadable[]`.

- Plant: `domain` (`id`, `version`, `title`, `digest`, `file`), `composition` and `twin` (the file of the same id, `status`
  `in_sync` | `out_of_sync` | `does_not_compose` | `does_not_scaffold` | `kind_files_missing`, the digest it produces now),
  `device_kinds[]` (each `implements` name with its kind file and digest, or `status: missing`), `runs[]` (id, mode, whether
  `run.json` exists, the domain version and digest the run recorded and `same_as_current`) and `editable`
  `{verdict, reason, route:{code, sentence}}` with reasons `composition`, `composition_out_of_sync`, `twin`,
  `twin_out_of_sync`, `example`, `hand_authored`.
- Device kind: `kind`, `file`, `digest`, `twins_including[]` (`{twin, file, kinds}`: how many kinds that twin holds),
  `has_own_twin` (a twin of this kind alone) and `implemented_by[]` (domain ids). `twins_including` is the explicit relation
  F24 found missing: "included in another twin" is a list entry, "has its own twin" is a separate flag.
- Reads `run.json` and the first line of `journal.jsonl` only; sealed files are never opened.
- Order is by id, file name and run id; nothing carries a wall time.

## Surfaces

`streamsim domain assumptions <domain-file> [--twin <choices>] [--kinds-dir D]... [--components-dir D]` and
`streamsim domain lineage --domains-dir D [--kinds-dir D] [--twins-dir D] [--compositions-dir D] [--runs-dir D]
[--components-dir D] [--domain ID]`. Bridge actions `domain.assumptions` (`{path, twin?}`; kinds and component directories from the
workspace) and `domain.lineage` (`{domain?}`; every workspace folder that exists).

## Known limits

- A twin or composition whose id matches no domain file is not listed (the document is about plants); its kind files still appear
  in `device_kinds` and in `twins_including`.
- Which kind file a name belongs to is read from the file's `kind` member, not from its file name.
- Whether a person chose or accepted a default is not recorded anywhere and is not reported.
- "Example" is a domain whose id is a shipped example composition and which equals that composition composed now.

## Round record

Deviations from the brief:

- The list is `not_modeled` and the codes use American spelling (`device_behavior_not_modeled`, `open_device_behavior`): the repository's
  `misspell` lint is set to the US locale and no `//nolint` may be added. The canvas maps `open_device_behavior` to its own
  "Behaviour" step; no sentence of the binary names that step.
- `equals_binary_suggestion` replaces the brief's "suggested by the binary versus chosen": the choices file records no choice, so only
  the comparison with `kind twin --suggest` is a fact.
- The edge package calls the domain layer for `Lineage` (`files.Lineage` reads the workspace, then `DescribeLineage`), like
  `files.Load` builds the library: the facade rules allow one delegating call, and the module has no `app` layer.
- Gate registry (owner to add, `test/architecture/cli_wiring_test.go`, two new CLI wiring files, no cap or allow-list loosened):
  `"domain_assumptions.go": {"command: domain assumptions", []string{cliF, "internal/component", "internal/plantwire"}, 85}` and
  `"domain_lineage.go": {"command: domain lineage", []string{"internal/component"}, 60}`.

# Round R5 (Go): the domain diff (OP08, F24, F25)

Acceptance: change the thermal model's gain, see the affected relation and model version, run a variant, and relate both recordings
to their exact model assumptions. R2 answered "what does this model assume" for one side; R5 answers "what changed between two".
One read-only use case, one new contract (`domain-diff-v0.1`).

## Placement

No new module, package or allow-list entry. `internal/component` (domain layer) owns the diff for the same reason it owns
assumptions: the diff compares the assumptions views (`assemble` without a twin overlay) of two domains, so a fact has one reading
in both documents. The `files` edge reads a side: a domain file, or a run folder (`run.json` only, `domain_spec` member; sealed files
are never opened). Facade functions `component.DiffDomains(a, b)` and `component.RenderDiff(doc, format)`; the CLI calls one each.
The diff needs no library and no twin files: it compares what the two domain documents declare.

## Document

`schema_version`, `a` and `b` (`id`, `version`, `title`, `digest`, `source {kind: file|run, label}`), `identity` (`same_id`,
`same_version`, `same_digest`, `same_id_and_version_different_digest`: the model changed and its version did not), `kinds[]`
(`kind`, `status` changed|added|removed, `instances {before, after}` each `{count, names, id_templates}`, `relations[]`
`{type, from_kind, to_kind, links}` (the relation a change runs over, from the side that still has the kind), `changes[]`),
`domain_changes[]` (links, link types and every other top-level member), `summary` (`changes`, per-status kind counts) and
`unchanged_summary` (items present on both sides and equal, per section). A change is `{code, subject, section, field, before,
after, sentence}`; `before` / `after` are JSON values (`null` when absent); `subject` is path-like (`<kind>/<section>/<key>[/<child>/<key>]`).

Codes are stable. Fields of an item are compared one by one on the assumptions view (readings, dynamics, commands, faults) or on the
declared state / link; a field maps to `<field>_changed` where it names a modelled quantity (`gain`, `time_constant`, `dead_time`,
`range`, `resolution`, `noise`, `cadence`, `availability`, `instrument`, `unit`, `inputs`, `clamp`, `safe_stop`, `risk_class`, `sets`,
`lease`, `latency`, `failure_modes`, `state_changes`) and otherwise to `<noun>_changed`. Items are matched by key and reported
`<noun>_added` / `<noun>_removed`. `gain_changed` is both a dynamics' own `gain` and the `coef` of one of its inputs (the strength of a
coupling; for a twin, the gain of a command's drive). A declaration that differs where the view found nothing is reported
`<noun>_changed` on field `definition` (before / after: the raw declaration), so a differing digest never shows an empty list; every
top-level member not itemised (profiles, ground truth, axes, site, boards, title, kind flows and boundaries) is `section_changed` /
`kind_changed`. Order: kinds by name; within a kind readings, dynamics, commands, faults, states, instances, kind members; items by key.

## Surfaces

`streamsim domain diff <a> <b> [--format json|markdown]`: `a` and `b` each a domain file or a run folder. Markdown prints
`{format, markdown}` like `report`. Bridge action `domain.diff {a, b}`: each a workspace domain file (`domains/x.domain.json`) or a run
id; CLI-backed, read-only.

## Known limits

- Dynamics, command and fault members the view does not model are compared only as a whole declaration (`definition`) and only when
  nothing else in that item differed.
- Which instance set a change affects is the kind's instance declarations; the diff does not resolve counts against a scenario.
- A run recorded before its domain file was edited is compared by the `domain_spec` embedded in `run.json`, never by the workspace file.

## Round record (R5)

Deviations from the brief:

- `kinds[]` carries `relations[]` besides `instances`: "the affected relation" of OP08 is the link type and its instance links
  (`fan_drives_thermal_sensor`, `fan-01>zone-01`), not only the subject path of the change.
- The inputs of a dynamics are a child group keyed `state@link type`; a coefficient change is `gain_changed` on field `coef`
  (for a twin that is the gain of the command's drive), so the thermal-gain case of OP08 is one `gain_changed`, not an opaque
  `inputs_changed`. `inputs_changed` does not exist; an input is `input_added`, `input_removed` or `gain_changed`.
- The diff needs neither library nor twin choices; it compares the assumptions views without the twin overlay (`assemble(compiled,
  nil)`), so `instrument_changed` appears for composed domains (provenance) and, for a twin, the instrument swap shows as
  `range_changed`, `noise_changed`, `resolution_changed` or `cadence_changed`. The twin choices stay `domain assumptions`' input.
- A run folder is checked: `run.json` must satisfy `run-artifact-v0.1`, hold a `domain_spec`, and the spec must hash to the
  digest the run recorded (`artifact_invalid` otherwise).
- Bridge sides: a workspace domain path (`domains/<id>.domain.json`) or a bare run id (as `run.compare`); a new argument shape
  `plant` renders either.
- Markdown prints `{format, markdown}` and reuses the `report_markdown` definition of `cli-output-v0.1`.
- Known limit D-119 in `governance/DEFERRED.md`.
- Gate registry (owner process as for R2, `test/architecture/cli_wiring_test.go`, no cap or allow-list loosened):
  `"domain_diff.go": {"command: domain diff", []string{"internal/component"}, 45}`.
