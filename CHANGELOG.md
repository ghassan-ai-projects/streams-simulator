# Changelog

This project has not established a formal release stream yet. Until it does, changes are recorded in pull requests and the repository history.

## Unreleased

- **A relay between a gateway and the real host.** `plantwire relay --gateway-socket <path> --host-socket <path> [--out <dir>]`
  is the lab shape for hardware: the simulator is the host of one gateway and the one node the real host sees. The gateway is
  checked as the reference host checks a `hardware` node (one shared predicate, `domain.Admission`), then its birth reaches the
  host as the relay's own: kinds with their `kind_digest`, `capability_digest`, boards and their environments, node metrics
  unchanged. Device records go up, and the host's `STATE`, `NCMD`, `DCMD`, `DQUERY` and `HSYNC` go down, with every member
  except the envelope as received (values, qualities, event times, boot ids, command ids and every safety member of a command);
  `seq` is assigned by the relay at transmission, and a barrier is mapped between the two sequences through the pairs of
  records it carried. The relay never originates, answers, refuses or delays a command, rewrites no quality and has no fault,
  perturbation or script option (a gate keeps its files from naming that code; `replay`, `run`, `scenario` and `plantwire node`
  take no gateway option). A lost host is reported to the gateway as offline with its session kept, and a rebirth follows when
  the host is back (three missed heartbeats count as loss); a lost gateway has each born device reported `DDEATH link_down`
  while the host session is kept. With `--out` every record on both links is journaled as observer `wire` records (directions
  `gateway`, `node`, `host`, `relay`; a carried record is two consecutive lines); `kind capture` reads that journal. New CLI
  document `plantwire_relay`; the run module gains `WireTrace` (a journal without a world); the reference host's handshake
  checks moved to `domain.Admission` with no change in behaviour. Gate edits: two plant wire edge files registered as I/O edges
  (`uds/gateways.go`, `uds/ticks.go`), `plantwire_relay.go` registered as a CLI wiring file (cap 45), and a new gate that
  keeps the relay's files from naming perturbation or wire fault code.

- **Plant wire 1.2.0: every board names its environment.** `NBIRTH.boards[].environment` (`simulated` or `hardware`) is
  required and `NBIRTH.environment` gains `mixed`; a device's origin is the environment of its board, never inferred. The
  reference host checks that a `simulated` node has only simulated boards, a `hardware` node only hardware boards and a
  `mixed` node at least one of each, and that the node is one it is set to accept: `plantwire host --environment` now
  takes `simulated`, `hardware` or `mixed` (`mixed` accepts all three; any other value fails before the host listens).
  A violation is refused as `environment_mismatch` and the host report says which rule in `reason`. `kind capture` shares
  the node-versus-boards check, refuses a contradicting birth and prints `boards: [{board, environment}]` beside the
  node's `environment`, both claims. A simulated node announces `simulated` for itself and all its boards and cannot be
  made to announce `hardware` or `mixed`. The contract generators, the fixtures (eight `environment-*` exchanges, whose
  header may carry `host_environment`, and two invalid births), the contract digest and the embedded schemas moved
  together; the `NBIRTH` bytes of a wire journal change (`contract_version`, `boards[].environment`), and nothing else in a
  simulated run does (the behaviour pin holds no wire run).

- **A kind can be captured from a birth.** `kind capture (<file> | --stdin) [--kinds-dir <dir>]...` reads the first
  `NBIRTH` of a file (bare, or the `wire` records of a run's observer journal; streamed line by line, a line at most
  1 MiB, a scan at most 32 MiB) and prints `{environment, bd_seq, births_seen, taken, later_births?, issues?, kinds}`:
  per announced kind a draft kind file (the entry without `kind_digest`), the announced `kind_digest`, `claims` (the
  members that came from the board and that nothing verified: `risk_class`, `safe_stop`, the bounds of an
  `args_schema`), the issues the draft still has, and, with `--kinds-dir`, a `comparison` with the workspace kind file
  of the same name (`new`, `same`, `differs` at the pointer of the first differing member, `unreadable`) and its
  `target`, the path the draft would be saved to, derived by one rule in the binary (a name that differs only in case is
  a `kind_name_collision` aimed at the existing file). Provenance is reported, never silent: the first birth is taken,
  later ones are listed with whether their kinds are the same (`rebirth_differs`), a line that names a birth but cannot
  be read is an issue (`birth_line_unreadable`), a cut scan is `scan_truncated`. A birth is a claim: it is refused
  (`kind_failed`, issues with `keyword` and `member`) unless it validates against the plant wire schema, speaks this
  contract and every `kind_digest` is recomputed and holds; the board's `capability_digest` is checked beside it
  (`capability_digest_mismatch`) and `environment` is echoed as the board's claim. Only the observer stream is taken:
  records are read strictly (exact member names, no member twice) and the stream is checked before a record is accepted.
  The reference host now also refuses a birth announcing a kind twice (`inventory_mismatch`). It writes nothing.
  `cli-output-v0.1` gains `kind_capture` (and `issue`); the kernel gains `OpenRegular` (a relative path is walked and
  no component may be a symbolic link). The canvas bridge gains the read-only `kind.capture` (`{content}` on standard
  input, or `{run}` for the observer journal of a run of the workspace, exactly one of them; the sealed journal is
  never reachable). Test support: a scripted plant wire node (`internal/plantwire/internal/standin`, imported by tests
  only) stands in for a board behind a gateway. A kinds directory that does not exist holds no kind files for a capture, and the bridge always passes the workspace `kinds/`, so a first capture has a `target`.

- **One unfinished device no longer breaks another, and the picker tells the truth about fit.** The twin actions take the
  kind files that are in the twin: the canvas bridge's `kind.twin` and `kind.suggest` take `kinds: ["<file stem>", ...]`
  (the stems of `kinds/<stem>.kind.json`, one to 64, repeats counted once, an absent stem is `not_found`) and pass exactly
  those files, never `--kinds-dir`; `kind twin [--suggest]` already took files as arguments, and its default `id` and
  `title` come from the given kinds only (`fan-twin`, `Twin of fan`; a long board is named after its first kind and a count).
  A domain that `implements` kind files now reads only the kind files of the kinds it implements (`domain validate`,
  `domain compose`, `run`, `plantwire node`, `mcp`), found by the declared kind and, for a file that does not parse, by its
  file name: a malformed kind file of another kind never blocks it. `component list --unit <u> [--min n --max n]` gives
  every instrument `fit: {ok, code?, member?}` (`unit_incompatible`, `range_unstated`, `range_narrower`,
  `reading_not_number`), decided by the comparison `kind twin` uses, so `ok: true` is what the scaffold accepts; the range
  is compared after unit conversion and fitting instruments are listed first. A twin now reads through an instrument of
  another unit (kPa transmitter, Pa reading) in the unit of its kind file, converting the instrument's noise, drift and
  fault magnitudes, so `twin_conversion_needed` is gone; a refusal of an instrument carries the `member` at fault
  (`unit`, `eu_range`, `data_type`) with keyword `twin_instrument_mismatch`. `cli-output-v0.1` gains `fit`. The canvas
  bridge gains `min` and `max` on `component.list`.

- **The binary tells the canvas everything a twin needs in one pass.** Every issue of `kind validate`, `kind twin` and
  `domain validate` keeps `pointer` and `message` and gains `keyword` (the JSON Schema keyword that failed, or a stable
  rule code such as `twin_member_required`, `twin_unit_required`, `twin_instrument_mismatch`, `twin_name_pattern`,
  `twin_range`, `domain_rule`) and, for a rule about a named member, `member` (`resolution`, `writable`, `safe_stop`,
  `unit`); `jsonschema.Error` and `model.Issue` carry them. The twin requirements of a kind file are reported all at once
  (`component.TwinRequirements`, one implementation). New: `kind validate --stdin [--twin]` (one kind document from
  standard input, at most 1 MiB, strict, writes nothing), `kind twin --suggest` (the binary's default choices for kind
  files, `{choices}`, writes nothing), `component list --unit <u>` (instruments of that unit or a convertible one, exact
  unit first, with `unit`, `range` and `match`). `cli-output-v0.1` gains `keyword`/`member` on issues, `kind_suggest`, and
  `unit`/`range`/`match` on a component summary. The canvas bridge gains `kind.check` (`{content, twin?}`, the first
  action with an input on standard input), `kind.suggest`, and `unit` on `component.list`.

- **A kind file never carries `kind_digest`.** The `KIND` schema forbids the member (a stale digest used to be accepted and
  ignored), and `kind validate` and every `--kinds-dir` load refuse it at `/kind_digest`; a metric's `eu_range` and
  `instrument_range` need `min` below `max`. The plant wire contract digest changes (schema, one invalid fixture). The canvas
  bridge lists a kind's `kind`, a component's `id` and a `title` only when they are plain text of the expected form, and
  lists only the path of a file over 256 KiB.
- **A safe stop is the domain's.** An effector may carry `safe_stop: true` (domain spec). The plant wire node announces
  and runs the commands the domain flags; `plantwire node` loses `--safe-stop` and takes `--kinds-dir`, and a domain that
  implements a kind file is checked against it before the node connects, so the announced `kind_digest` and safe stops
  are the kind file's. The `implements` check compares the safe-stop flag of every command. The `kind twin` scaffold
  flags the safe stops of its kind files.
- **The `implements` check cannot be skipped.** `run.ScenarioRunOptions.CheckDomain` with no check refuses a domain that
  implements a kind file, as the director does (`run.CheckDomain`, one implementation). `domain validate`, `run` and
  `sim.world.create` print `implements_checked` (`{kind, kind_digest}` per kind file checked; empty when the domain
  implements none). Kind files are read only for a domain that implements one, only as regular files of at most 1 MiB,
  as one JSON document with no member named twice and no number beyond the range of a double; `kind validate` no longer
  prints a blank digest for such a file. `plantwire.CheckImplements` returns what it checked, and `plantwire.Scaffold`
  replaces `ReadKindFiles` and `CheckImplementsKinds`: `kind twin` is one facade call.
- **`streamsim kind twin`.** Scaffolds a runnable twin of kind files: `kind twin [<file.kind.json>...] --choices
  <choices.json> [--kinds-dir <dir>] [--out <file>]` writes a domain whose kinds `implements` the kind files, from a
  small choices document (`twin-choices-v0.1`): per metric the library instrument that measures it or `none`, per
  command one of three fixed responses (`direct`, `first_order`, `none`); a safe stop and a lease come from the kind
  file. The twin is checked against the same kind files before it is printed. The canvas bridge gains the `kind.twin`
  action and the `twins/` workspace folder; `cli-output-v0.1` gains `kind_twin`. The built-in vocabulary accepts `Cel`
  for `degC` and `%` for `%RH`, the spellings of a kind file.
- **The director checks `implements`.** `streamsim mcp --role director` takes `--kinds-dir` (repeatable), and
  `sim.world.create` runs the `implements` check on the domain it is asked for, as `run` does; a director without kind
  files refuses a domain that implements one. The canvas bridge starts its director with the workspace `kinds/` folder.
  `director.NewFromDirs` and `mcp.ServeOptions` take the check as an argument.

- **Kind files and `implements`.** A domain kind may declare `implements: "<kind>"`. `streamsim domain validate`,
  `domain compose` and `run` (with `--kinds-dir`, repeatable) check that the kind's plant wire interface equals the
  kind file of that name (`*.kind.json` in a kinds directory) after canonicalisation, naming the first differing member; a mismatch, an unknown
  kind file or no `--kinds-dir` is a `domain_invalid` error. `streamsim kind validate <file>...` checks kind files
  against the plant wire kind schema and prints their `kind_digest`. `cli-output-v0.1` gains `kind_validate` and
  the `kind_failed` code.

- **`streamsim component new`.** Prints a new plant's composition, copied from a composition file or a shipped example
  or emptied to its frame, with the id and title checked by the binary. The canvas creates plants through it, so the
  rule for what an empty plant keeps and the id check live in the binary.
- **Run listing fault counts.** Each run in the canvas bridge's workspace listing carries `faults`, the fault
  injections in its command log.

- **Validation messages in words.** A JSON Schema node with a `pattern` may carry `x-expected`, a phrase naming what the
  value must be; a value that fails the pattern is then reported as `"bad name" is not an instance name: …` instead
  of quoting the regular expression. The composition contract describes plant ids and instance names this way.
  Patterns without the annotation keep their message.
- **Canvas bridge listing.** Each domain in the workspace listing carries its `instances` count.

- **Starter examples in the binary.** `streamsim component examples` lists the example compositions the binary ships
  (`factory-utilities`, `chiller-pilot`) and `streamsim component example <id>` prints one, so a fresh workspace can start
  from a working plant. The canvas offers them as "Try an example" on an empty workspace and as a starting point in New
  plant. `cli-output-v0.1` gains both documents.

- **Alarm playback truth** (R9-24). Simulator 0.3.2 publishes each alarm's ordered `activation_times_ns`
  beside its aggregate KPI count in sealed KPI truth and the run report. These are the existing scan's counted
  transitions, including unshelving, with no activation attributed to an initially standing alarm. Playback
  can navigate the reported instants without reconstructing alarm transitions. The sealed journal and run
  digests deliberately change; alarm dynamics and aggregate KPI calculations are unchanged.

- **Incomplete runs and the report's loading** (findings E5-1, E5-2, E5-4; R9-30 to R9-32). Breaking: `streamsim verify` never
  reports a run that ended `incomplete` as matching. It answers `matches: false` with `reason: incomplete` (exit 1,
  `replay_mismatch`) even when the replay and every file agree; a more specific reason (a file that differs or is missing, a
  trace that diverges) wins when one applies. Every evidence file and digest an incomplete run does have is compared with
  each other; the replayed journals, ledger and history are not compared with the replay (the close of a failed run cannot be
  reproduced). The `incomplete` stamp is checked against the `run.closed` record of `journal.jsonl` and `sealed/journal.jsonl`
  and against `reproducible` (an incomplete run is never reproducible): a contradiction, which editing `incomplete` on a closed
  run's `run.json` produces, is `reason: incomplete_stamp` and names the journal. `cli-output-v0.1` lists both reasons and the
  `replay` document now carries `incomplete`, `emitted` and `journal_diverged`; the `sim.run.verify` tool carries `reason` and
  `incomplete`. An empty or malformed evidence digest (`"ledger_digest": ""`) is a verdict (`digest_missing`), not a schema
  rejection. `streamsim report` loads `run.json` as `verify` does (`run.LoadArtifactForVerification`): a run that ended
  complete with an evidence digest absent, empty or malformed is `verification.status: files_differ` with a `digest_missing`
  check (`recorded` absent), where it failed with `artifact_invalid`; `run-report-v0.1` lists the status. A run still open in
  evaluate mode reports `truth_unavailable: "sealed until the run ends"` (the same field, same vocabulary as the
  foreign-run reason). The canvas names the two new reasons.
- **Fault intervals in the run report and a state vocabulary on channels** (findings E4-2, E4-4; R9-20, R9-21). Breaking:
  `truth.faults` in `streamsim report` is no longer the inject and clear commands; each fault is an interval the world
  gave it: `fault_id`, `entity_id`, `fault`, `injected_at_ns`, `onset_ns`, `cleared_ns` (`null` while active at the end
  of the run) and `cleared_by` (`command`, `pulse_end`, `responder` or `null`). A clear command ends the fault at its
  effective `at_ns`, a pulse at onset plus its duration, an effector responder at its arrival; the earliest end wins.
  The world records each end (`world.FaultIntervals`, `FinalFaultIntervals`), the run seals them in the director journal
  as `fault.truth` records (after each fault inject, clear and effector call that changes them, and at the close), and
  the report takes the last record, so evaluate-mode truth stays sealed until the run closes. Every run with a fault
  has a new sealed journal digest; the pinned report run, its goldens and the run-report and run-journal schemas change.
  A channel may declare `states` (`[{value, name, attention}]`, `attention` is `alarm`, `warn` or `null`, `value`
  unique); an instrument declares the same list and `domain compose` copies it onto its channel. The annunciator
  instrument declares the five alarm states; `examples/factory-utilities.domain.json` is recomposed. The member is
  `value`, not `code`: the domain schema's closed-grammar gate rejects a member named like an expression. The report
  schema now lists the `journal_file` and `sealed_journal_file` checks the report already wrote.
- **Verify evidence review** (findings E4-1, E4-3, E4-8; R9-15 to R9-17). A run that ended complete must record a digest for every
  evidence file (`expected_trace_digest`, `journal_digest`, `sealed_journal_digest`, `journal_file_digest`,
  `sealed_journal_file_digest`, `ledger_digest`, `history_digest`; the run artifact schema requires them unless `incomplete`),
  and `streamsim verify` requires each file (`trace.jsonl`, `ledger.jsonl`, `world_state_history.jsonl`, `journal.jsonl`,
  `sealed/journal.jsonl`) beside `run.json`, comparing it with the recorded digest and with the replay; a missing file or
  digest, which verified as matching before, is `matches: false`. The replay and verify result carries `reason` (`match`,
  `recorded_spec`, `world_digest`, `trace_divergence`, `journal_divergence`, `evidence_divergence`, `unblinded_stamp`,
  `digest_missing`, `file_differs`, `file_missing`) and `file`; `cli-output-v0.1` requires `reason`. `streamsim report` checks
  `world_state_history.jsonl` too, from the same digest list (`RunArtifact.EvidenceDigests`), and its schema names all seven
  checks. A trace written to a `--sink-target` outside the run folder is also written to `trace.jsonl` in it. `verify --help`
  prints "Usage of verify". Breaking: older `run.json` files without the digests no longer pass `report` or `score`
  (`artifact_invalid`) and `verify` reports them as `digest_missing`. The report fixture (`internal/report/testdata/run`) gains
  the four digests and a three-row `world_state_history.jsonl`, with both goldens, and the behaviour pin is regenerated (the
  `verify`, `replay` and `failure/replay-mismatch` documents gain `reason`; the `unknown-command` usage line for `canvas` was
  already missing from it).

- **Verify review fixes** (findings E3-2, E3-3, E3-6, E3-8, E3-10). `streamsim verify` and `replay` report an artifact whose own
  recorded digests do not hold (the embedded domain or adapter spec does not hash to its digest, or the recorded specs replay
  to a different world digest) as `matches: false` with a `detail`, in the `result` of `replay_mismatch`, instead of an error
  with no result; a missing or unreadable `run.json` stays an error. `verify` also compares `trace.jsonl` beside the artifact
  with the digest `run.json` recorded. `scenario from-run` and `score` replay the specs recorded in `run.json`
  (`run.ResolveInputs`, shared with `replay`/`verify`), so a later domain edit does not break them. `domain_changed` is omitted
  when the current domain is unknown. The canvas bridge names its session cookie `streamsim_canvas_<port>`, so two bridges on
  one host no longer unpair each other.

- **Factory utilities world, scale and KPI truth** (phase P7l). `examples/factory-utilities` is composed from the library only;
  a 2,032-tag campus runs at 109x real time through the run pipeline. Breaking: `follow` takes `sources: [{link, state}]`
  (several link types into one state) instead of `link` and `state`. New: `kinds[].kpis`, the sealed `kpi.truth` director
  record and `truth.kpis` in the report (domain-spec v0.2 §29); `journal_file_digest` and `sealed_journal_file_digest` in the
  run artifact, checked by the report; framing records in the observer journal; `world.ErrFaultOnsetInPast` and
  `world.ErrFaultClearInPast` for an onset or clear before the clock of a world that has run. Truth oracle worlds are built
  as a run builds its world (the scenario's seed and the derived world identity), are advanced in time order (clean before
  faulted) to every sampled instant, schedule no emission unless a pulse needs recordings, and judge a pulse only at its
  detector channels' recordings. KPI samples at the instant an advance ends are taken at the next advance or at the close,
  and a retired entity's alarms end at its retirement.
- **Link ports in the domain** (finding R9-03). A domain link may carry optional `from_port` and `to_port`; `domain compose`
  writes them from each connection, so a client reading only the domain sees which port of an instance a link uses. The world
  ignores them and a link's identity stays `(type, from, to)`. The two composed examples are regenerated.

- **Simulator version 0.3.1** (review P7e). `sim_version` is `0.3.1`: the P6a integration fix changed faulted hidden-state
  numbers, which changes the digest inputs of a faulted run. The behaviour pin is regenerated (only the `run.json` and
  `journal.jsonl` lines change) and gains a day-long run with two faults and three perturbations. The operator endpoint of
  the MCP director is closed, and its serving goroutine finished, before the session returns.

- **Component and lint review fixes** (phase P7j). The built-in component library is embedded in the binary (package
  `components`): `component.NewLibrary(customerDirs)` loads it first, then `--components-dir` and
  `STREAMSIM_COMPONENTS_DIRS` directories (empty entries ignored); a customer file with a `streamsim/` id is refused, the MCP
  director always serves the library, and no `components` working-directory default remains. Link types are named
  `<to kind>__<to port>__<from kind>` (one per consuming port and providing kind), conservation detector `inputs` and
  `outputs` are checked and qualified like other channel references, annotation keywords are stripped at every depth of a
  copied `args_schema`, two parameters writing one target is a load error, an id-template connection endpoint must render a
  valid id, span scaling uses the converted point span, library listings carry no file path, `utc_offset_s` needs `period_s`,
  and the composition schema gives `not_applicable` a minimum length of 20. Lint: `pinned_at_bound` applies to integrated
  states only (an idle integrated state at its bound is reported) and the `not_settled` cyclic exemption follows `follow`,
  linked inputs and flows. `examples/chiller-pilot.domain.json` is regenerated for the new link type name.
- **Instruments, electrical equipment and process loads** (phase P7k). Ten instruments (`thermistor`, `pressure-transmitter`,
  `vortex-flow-meter`, `level-radar`, `accelerometer`, `current-clamp`, `valve-position`, `humidity`, `run-status` with a boolean
  channel, `pulse-counter` with a counter channel) and an `energy-meter` register join `rtd`, `mag-flow-meter` and `power-meter`,
  each with `stuck` (a quality where the device knows), `drift`, `dropout`, `noise` and `scale` faults where physical; the three
  existing instruments gain the same. Equipment `feeder-meter`, `transformer`, `generator`, `production-line` and
  `packaging-line`. The vocabulary gains the port types `pressure`, `temperature` and `shift`. Compose no longer writes a
  `unit` on a boolean channel, which the domain contract forbids. `examples/chiller-pilot.domain.json` is recomposed.
- **Catalog harness, cooling group and environment** (phase P7k). `test/contract` proves every equipment of the built-in library
  and every preset alone: it composes, validates, lints clean, starts at rest, replays byte for byte after a two-hour run with
  its effectors called and its faults injected, every effector moves a state it names, every fault is observable by the truth
  solver, and the trivial-baseline audit runs (and finds at least one fault hard to detect); every instrument attaches to a
  point, every port type has both ends, and every providing port composes with every consuming port of its type. New
  equipment `cooling-tower`, `pump`, `heat-exchanger`, and the system parts `weather` and `site-calendar`; the chiller gains
  `running`, start and stop, a part-load efficiency curve with COP, a return temperature and an energy cost rate; the air
  handling unit gains occupancy, outdoor-air load and coil limits. The vocabulary gains the tariff family and the port types
  `temperature`, `cooling_water_temp`, `dry_bulb_temp`, `wet_bulb_temp`, `shift` and `tariff`.
  `examples/chiller-pilot.domain.json` is regenerated.

- **Domain lint** (phase P7j). `streamsim domain lint <domain.json> [--hours H] [--seed S]` simulates a noiseless nominal day
  (default population, no faults, no effector calls, sampled every 60 s) and reports `non_finite` and `out_of_range` errors and
  `pinned_at_bound`, `not_settled` and `conservation_residual` warnings; an error finding exits 1 with the new code
  `domain_lint`. New module `internal/lint`, success document `domain_lint`. Domain spec section 27.
- **Component module, compose and CLI** (`docs/design/COMPONENTS_V0_1.md`; phase P7j). `internal/component` loads component
  libraries (equipment, instruments, vocabulary; built-in directory first, customer directories after, the `streamsim/`
  namespace reserved to the first) with every load-time check of the spec, and composes a composition into an ordinary domain:
  presets then parameters, keyed targets, instruments with unit conversion and span scaling, instrument faults including the
  observation forms (`stuck`, `freeze`, `dropout`, `scale`, `noise`), qualified names, provenance digests, site and areas, and
  connections as link types and links with `follow` (aggregate `sum`) on the consuming state. New commands `streamsim component
  list|show|validate` and `streamsim domain compose`, each with repeatable `--components-dir` and `STREAMSIM_COMPONENTS_DIRS`;
  new error codes `component_invalid`, `composition_invalid`. `examples/chiller-pilot.domain.json` is now what compose writes
  (the air-handling link carries the load into the small chiller).

- **Plant wire reference host, journal and CLI** (phase P7d). `streamsim plantwire host` is a minimal conformant host
  (owner-only socket, handshake checks and refusal, commissioned inventory with quarantine, rebirth on a gap, historical data
  kept out of live values, scripted commands, verification by observation, a configurable barrier delay); `streamsim plantwire
  node` runs a domain as a node and writes its artifact and journal. The journal gains `wire` and `pacing` session records.
  New CLI error code `plantwire_failed` and success documents `plantwire_node`, `plantwire_host`.

- **Plant wire commands** (phase P7c). The node executes `DCMD`: checks in the contract's order (device, command, arguments,
  boot, deadline, epoch floor, policy allow-list) at receipt and again at execution for a command queued until
  `not_before_ns`; an idempotency ledger replays the stored receipt and result for the same content and refuses other
  content as `duplicate`; safe-stop commands skip the node's freshness, epoch and policy checks. Commands run through the
  run's single effector path, so they share its command log and replay. World outcomes map to wire classes (`interlocked`,
  `permissive_not_met`, `not_ready`, `out_of_range`; a lost acknowledgement sends nothing; a lying device reports
  `succeeded`). An outbound wire-fault plan (`drop`, `duplicate`, `swap`, `garble`) acts after sequencing. Test domain
  `pump-board.domain.json` matches the contract fixtures, whose boot ids now follow `<board>-boot-<n>`. An effector `lease`
  may declare `unit: "ms"`.

- **Plant wire node** (`internal/plantwire`; phase P7b; contract `contracts/plant-wire/SPEC.md`). A run can speak the plant
  wire to a host: it dials the host's owner-only socket (refusing a group- or world-accessible one), sends `NBIRTH` (each kind
  once, the contract and capability digests, the environment marker) and every `DBIRTH` (board, boot, current values, add-or-
  narrow overrides), negotiates the version or stops on a refusal, advances the world slice by slice with `DDATA` and an
  `NSYNC`/`HSYNC` barrier capped in wall time, processes host records in sequence order (epoch floor, rebirth in the same
  session, state queries, graceful goodbye), ends on a host sequence gap or three missed barriers, and closes with `NDEATH`.
  Commands are refused `not_ready` until P7c. Entity ids are mapped to wire-safe ids; no domain description reaches the wire.
  New `run.SetDeliveryObserver` (every delivered event after perturbation, before the adapter). Contract changes: a birth's
  current value may be `null` only with quality `warming_up`; an invalid host record still consumes its `seq`; the barrier
  rule is explicit; the contract digest moves. A new fuzz target `FuzzWireParse`.

- **Event clock** (`docs/design/DOMAIN_SPEC_V0_2.md` section 19; phase P6e, review item S11, finding R0-12). A `pulse` fault
  shape acts for `duration_s` and its start and end are event instants: trips sample their condition there (a 0.3 s sag trips
  a 0.2 s ride-through trip, a 0.1 s sag does not), and event-driven channels are checked there, so a power-quality channel
  records a sag at its exact time. A trip now shows at the boundary where `for_s` has elapsed since its first true sample (one
  step earlier than in P6c). The spec gains the RNG draw-order table. Shipped domains and every pinned output are unchanged.

- **Supervisory control: setpoints, sequencer, responder, lease** (`docs/design/DOMAIN_SPEC_V0_2.md` section 18; phase P6d,
  review items D3 and S2). A condition may compare against a setpoint state (`threshold_state`, with `threshold` as an
  offset); `f1.form: "sequencer"` stages a plant up and down within a band; an effector `responder` applies a delayed effect
  chosen by whether its argument names an active fault, and may clear it; an effector `lease` falls back (`safe_off`,
  `revert`, `hold_last`) unless renewed. The utility fixture gains an air system with a pressure band, a compressor
  sequencer, a leased boost and a work-order responder. Shipped domains and every pinned output are unchanged.

- **Protection: trips, lockouts, anti-recycle, boundaries, preconditions** (`docs/design/DOMAIN_SPEC_V0_2.md` section 17;
  phase P6c, review item S14). A shared closed `condition` (`lt`, `lte`, `gt`, `gte`); `f1.form: "trip"` latches (a lockout,
  reset when `reset_on` changes) or clears; `f1.form: "anti_recycle"` is a timer after an edge; `kinds[].boundaries` name trip
  states with a severity; `effectors[].requires` refuses a call whose preconditions fail (locked out, recycling, isolated,
  hand mode) with the generic `effector_refused`, the director log recording `precondition_refused`. **Fix:** interlocks and
  preconditions of a call dated after the world clock are evaluated on a copy of the entity's state, so they no longer change
  earlier readings. The utility fixture gains a chiller. Shipped domains and every pinned output are unchanged.

- **Schedule, regime and peak** (`docs/design/DOMAIN_SPEC_V0_2.md` section 16; phase P6b, review item S9). `f0.schedule` adds
  a clock-aligned piecewise profile (step or linear) to an F0 state; `f1.form: "regime"` is a lag whose gain and time constant
  switch when a selector state crosses a threshold at a step boundary; `f1.form: "peak"` is the billing-period demand peak: the
  largest epoch-aligned block mean of one input, with a ratchet on the previous period's peak and seeded `initial` and
  `previous_peak`. New fixture `utility-plant.domain.json`. Shipped domains and every pinned output are unchanged.

- **Stores and flows** (`docs/design/DOMAIN_SPEC_V0_2.md` sections 13 to 15; phase P6a, findings R0-11, R2-05,
  R5b-01..03). `f1.form: "store"` with a `capacity` accumulates its inputs and flows; `kinds[].flows` move
  `conductance × Π scale × (from − to)` between stores of one entity, or drain a store without `to`. Every store of an
  entity is integrated jointly on one shared grid; every store of a kind shares one `dt_ms`. **Integrator fix:** every
  integrated state of an entity now advances together, earliest step first, and an input is read at its exact step
  boundary, so no value depends on the order states are read (before, a lag fed by a lag read ahead could see a later
  value, and an emitting world could differ from the silent oracle world). A test proves the domain schema holds no free-text expression. Narrowing refuses a target named twice and
  argument bounds that leave no value; a `peer_residual` fault added by an override may not watch a channel only the
  override adds; a profile's entity count sizes the first templated instance set; idempotency compares argument numbers
  by value. New fixture `transfer-plant.domain.json`. Shipped domains and every pinned output are unchanged.

- **Named instances, overrides, links, system entities and boards** (`docs/design/DOMAIN_SPEC_V0_2.md` sections 8 to 12;
  phase P5b, finding R5a-02). An instance set may list `names` instead of or beside an `id_template`; an explicit entity id
  is accepted when it is a name or fits a template. `override` on an instance set adds states, dynamics (for added states),
  channels, faults and effectors, or narrows a state's bounds, a channel's range or an effector argument's bounds; removing
  or retyping is refused with a located issue, and the world gives each entity the effective kind of its set. `kinds[].system`
  marks a plant-wide kind with exactly one named instance. `link_types` and `links` declare typed topology between
  instances; `boards` group the instances hosted on one device and own the `boot_id` (structure only, no run-time behaviour).
  The domain-spec contract moves its definitions of state, dynamics, channel, fault and effector into `$defs` so a kind and
  an override share them. Shipped domains and every pinned output are unchanged. New world methods `StateNamesOf` and
  `EffectorOf`; new fixture `structured-plant.domain.json`.

- **Multi-kind world** (`docs/design/DOMAIN_SPEC_V0_2.md` §2.3 and §3; phase P5a2, finding R5a-01). A domain may declare
  several **kinds** and several **instance sets**, each set following a kind with its own `id_template`, `count` and
  `churn`; the `maxItems: 1` limits on `kinds` and `instances` are lifted. Channel, fault and effector names are unique
  across the domain, state names are unique within a kind, and every reference (observes, dynamics, affects, effects,
  detector channels, expected effector, confirmation channels) resolves inside its own kind; a kind with no instance
  set is refused. The validated document is now the runtime form (`model.DomainSpec` holds `Kinds` and `Instances`; the
  P5a lowering is removed). Each world entity carries its kind, which selects its state, dynamics, channels, faults and
  effectors: injecting a fault or invoking an effector on an entity of another kind is refused, a peer-residual
  detector pools only the entity's kind, a suite scenario draws its entity from the fault's kind and the audit uses
  that kind's channels and control fault. Single-kind domains are bit-identical (runtime-equivalence digests,
  trace, ledger and history in the behaviour pin). **Behaviour changes:** `catalog describe` prints the v0.2 document
  (`kinds`, `instances`) instead of the flat runtime shape; an explicit entity-id list is matched to an instance set
  and counted per set; `world.RenderID` is replaced by `world.PopulationIDs`.

- **Domain spec v0.2, kinds and instances** (`docs/design/DOMAIN_SPEC_V0_2.md`, `domain-spec-v0.2`, replacing
  `domain-spec-v0.1`; phase P5a). A domain declares one **kind** (the entity template: state, dynamics, channels,
  faults, effectors; its name is the `entity_type`) and one **instance set** (`id_template`, `count`, `churn`) that
  follows it. The document compiles to the v0.1-shaped runtime form, so the world, perturbation, adapter, truth,
  score and run code are unchanged; the eight shipped domains are rewritten and each compiles to the runtime world
  its v0.1 form produced (the recorded trace, ledger, world-state history, replay and suite outputs are
  byte-identical). **Behaviour changes:** `domain_digest` is the digest of the v0.2 document (every shipped domain's
  digest, every run artifact's `domain_spec` and `world_digest` change once); `sim_version` is `0.3.0`; run artifacts
  recorded before this change no longer replay (their `domain_spec` is v0.1); `catalog describe` prints the runtime
  form and the new digest.
- `streamsim domain format <path>` rewrites a valid domain file in the canonical layout (stable member order,
  line-fitting objects) and prints `{formatted, changed, digest}`; idempotent, digest unchanged, an invalid
  document is refused and left untouched. `cli-output-v0.1` gains `domain_format`.
- **Validation issues are located** (findings R1-01, R0-10, S15): a "missing required property" or "additional
  property" issue from any contract points at the member, not its parent, and every domain cross-reference issue
  carries a JSON pointer into the v0.2 document (for example `/kinds/0/channels/1/observes`). Domain cross-checks
  also reject a `count.default` outside `count.min`..`count.max`.
- **Explicit entity ids are checked** (finding R2-02): a scenario's `world.entity_ids` and the director's world
  creation must fit the domain's instance id template and count bounds; a violation refuses the world (a scenario
  violation at `/world`).

- **Run report** (`docs/design/RUN_REPORT.md`, `run-report-v0.1`, new `internal/report` module):
  `streamsim report <run-dir> [--format json|markdown] [--out <file>]` summarises a run folder as one
  deterministic document, a pure function of its files: identity and inputs, recorded digests,
  `verification` (do the files still agree with the recorded digests; not a replay), delivery counts
  with drop reasons from the ledger, faults, perturbations, environment faults and effector calls
  with the world's real outcomes from the command log, mode (`evaluate` or `explore`), whether the
  run is complete and ended, the consumer from `verdict.json` (its `config_digest` only when the
  verdict recorded one) and the scorecard when `verdict.json` and `label.json` are present. The
  truth sections appear only for a run that has ended or an Explore run; the report of an open
  Evaluate run reads neither the ledger nor `sealed/`, so it cannot be used to read ahead. JSON is the command's one stdout document; Markdown
  requires `--out` and stdout then names the file (`{report_written, format}`), keeping stdout one
  JSON document. `cli-output-v0.1` gains `report` and `report_written` and the `report_failed` code.
- `grading.Options` with no domains or adapters directory now grades from the domain and adapter the
  artifact embeds (`run.EmbeddedInputs`); `streamsim score` still passes its directories.

- **Run journal** (`docs/design/RUN_JOURNAL.md`, `run-journal-v0.1`, new `internal/journal` module):
  two append-only JSON Lines streams written while a run happens. The observer stream
  (`journal.jsonl`) holds what a consumer could know: the run's identity, every delivered
  observation exactly as the sink received it (the adapter's rendered line), clock advances, the acknowledgement each effector call returned,
  consumer reports and the close; never the seed (MCP_SURFACE §4.5). The director stream (`sealed/journal.jsonl`) holds commands,
  delivery rows, world state, effector outcomes, unblinding and the close. Truth-bearing records
  never enter the observer stream (canary test). `run --out <dir>` and a director world write both in
  the run directory; no `--out`, no files. Both streams are pure functions of the run;
  `run.json` records `journal_digest` and `sealed_journal_digest`; `replay` rebuilds both in memory
  and `verify` fails with `replay_mismatch` naming the stream (and, from the files beside the
  artifact, the first differing record). Journals are buffered, flushed at command boundaries and
  synced at the end; a torn last line is ignored by the reader; an unencodable record fails the run.
  Consumer reports are session records: in the file, outside the digest (replay cannot reproduce
  them). A `clock.advance` is journaled with its target only (`await_consumer` is pacing).
  Fuzz target `FuzzJournalRead` in `make fuzz` and `make fuzz-soak`.
- **Explore mode:** `run --explore` and `sim.world.create {explore: true}` stamp the run unblinded
  from creation (`mode: "explore"` and `unblinded: true` in `run.json`, `run.opened.mode`, a
  `run.unblinded` record). **Behaviour change:** `streamsim score` refuses every run stamped
  unblinded (`score_failed`), as the director's `sim.score` already did; it used to print a
  scorecard carrying `unblinded: true`.
- `run.json` gains `mode`, `journal_digest` and `sealed_journal_digest` (behaviour pin: only
  `run.json` lines change, and the journal files are pinned). `sim_version` is unchanged: the trace,
  ledger, world digest and determinism tuple are identical.
- Director composition tests moved to `test/acceptance` (finding R2d-01); `Director.RunOf` is the
  documented test-harness handle on the run behind a world, replacing the `export_test.go` seam.

- **Fix (truth leak):** `sim.effector.invoke` returned the world's internal result to the
  consumer, including the failure `mode`, `effect_applied` and `ack_latency_ms`, which exposed
  `silent_no_effect`. The operator now returns only the acknowledgement fields of MCP_SURFACE
  §4.3 (`model.EffectorAcknowledgement`), and every acknowledging mode returns a byte-identical
  response (`confirmed_no_effect` now promises an effect time like the others). A lost
  acknowledgement reads as `reason: "acknowledgement_timeout"` with no effect time. The reference
  consumer reads the same acknowledgement.

- **Scenario files** (`docs/design/SCENARIO_FILE.md`, `scenario-v0.1`): `run --scenario <file>`
  runs a repeatable scenario; run flags build the same scenario in memory, so there is one
  execution path, and the behaviour pin shows every existing run unchanged. `scenario from-run`
  exports the scenario that reproduces a run (identical command log, trace and world digest).
- **Behaviour change:** a scenario is checked before anything is delivered (`scenario_invalid`
  for a file, located at the field; `usage_invalid` for flags, each issue naming the flag). A
  static check reports every effector, fault, perturbation and profile that does not exist,
  every refused effector argument or fault or perturbation parameter, every inexact time and
  every backwards or over-long advance. A dry run on the same world with emission off then finds
  the first command the run would refuse as it unfolds: entities that do not exist at that point
  (churning domains included), ids already in use, clears of ids never assigned. During a run, an
  effector call the world refuses through an interlock or a reused command id is a recorded
  outcome. Replay still keeps every recorded refusal.
- **Behaviour change (every surface):**
  - A run's clock may not move more than 30 days past its start (`run.MaxRunSpanNS`;
    `--duration` at most 2592000 s).
  - A perturbation window that ends before it starts is refused (`--perturb x@100@50` used to
    succeed and never act).
  - An entity may not be added before the world clock.
  - A world starts between 0 and `world.MaxStartNS` (`--start-time` likewise); flag offsets
    and `--duration` must be finite numbers within the run span.
  - Artifacts recorded with such commands no longer replay.
- **Behaviour change:** `scenario from-run` checks the scenario as `run --scenario` would
  (`--domains-dir`, `--adapters-dir`) and writes nothing for a run that cannot be expressed as a
  runnable scenario. Effector calls the world refused at admission are dropped and counted
  (`refused`); a changed domain or adapter is `replay_mismatch`.
- **Behaviour change:** a run refuses a profile its domain does not declare on every surface,
  replay included (the director answers `domain_invalid`); an artifact recorded with such a
  profile no longer replays. `force_failure_mode` must be an effector failure mode in the
  scenario and run-artifact contracts and when a world is created.
- **Behaviour change:** scenario and exported `clock.advance` commands carry only `to_ns`;
  `await_consumer` is pacing of a live session, and scenario runs record it as `false`.
- **Behaviour change:** flags may follow positional arguments in every command (`replay <run.json>
  --domains-dir d` now honours the flag); `--` ends flag parsing unless it is a flag's value. A
  command or verb given more positional arguments than it takes fails with `usage_invalid`
  instead of ignoring them.
- **CLI output contract** (`docs/design/CLI_SURFACE.md`, `cli-output-v0.1`): every command except
  `help` and `mcp` prints exactly one JSON document on stdout, for failures too. Failures carry a
  stable code, the file at fault, every contract issue with a JSON pointer when known, and for
  `verify`/`adapter verify`/`score` the result that diverged.
- **Behaviour change:** argument errors that are not flag-parse errors (missing or unknown verb,
  missing required argument, malformed `--fault`/`--effector` value) now exit 2 with
  `usage_invalid` instead of 1. A failed `verify` prints the error document with the replay
  report as `result` instead of the bare report.
- Domain, adapter, run-artifact, verdict, label and ledger failures are typed
  (`model.ValidationError`) instead of text; `domain.ErrUnknownDomain` names an unknown id.
- **Behaviour change:** argument values are checked before running: `--sink` must be a known sink,
  `http-push` needs `--sink-target`, `--duration`, `suite --n`, `refconsumer --window` and
  `--threshold` must be positive, `suite` needs `--domain`, and `--profile` must name one of the
  domain's profiles (a `run` with an unknown profile used to succeed and ignore it). All exit 2.
- **Behaviour change:** `score` checks the verdict and label against their contracts, decodes the
  ledger and state history strictly, and loads the run artifact first; `-h` works after a verb
  (`domain validate -h`, `device -h`); flags may precede a verb (`adapter --adapters-dir d verify f`).
- `run.ErrDigestMismatch`, `adapter.Lookup`/`adapter.ErrUnknownAdapter` and
  `model.ValidateGroundTruth` are new typed entry points.
- Gates: `godox` and `nolintlint` in lint; a consumer-knowledge gate with a stale-exception check; a
  pinned `device serve` transcript; failure documents in the behaviour pin (stdout and stderr
  separately); `make behaviour-pin-check`; `make phase-gate`.

- Introduced a curated public documentation set under `documentation/`.
- Documented the distinction between the public set and the engineering archive under `docs/`.
- Added explicit limitations and evidence guidance for the current working-tree inventory.

When a release process is established, each entry should name contract changes, behavior changes, migration notes, and evidence for the exact release commit.
