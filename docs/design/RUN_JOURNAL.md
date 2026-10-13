# Run journal

- **Status**: implemented in phase P3 of the simulation canvas goal
  (`governance/CANVAS_PLAN.md`); contract
  [`run-journal-v0.1`](../../contracts/run-journal-v0.1.schema.json); module `internal/journal`
  (a core with a `files` edge); evidence `docs/simulation-canvas/evidence/P3.md`.
- **Why**: today the trace, the state history and the run artifact are written only when a run
  ends. A person or agent watching a run, and anyone investigating a crashed one, needs a record
  that grows while the run happens, without ever exposing sealed truth to the consumer side.

## 1. Two streams

A run with a journal directory writes two append-only JSON Lines files while it runs:

| File | Stream | Readers before the run is sealed | Content |
|---|---|---|---|
| `journal.jsonl` | **Observer** | Anyone, including the consumer and a coding agent that tunes it | What a consumer could know: the run's identity, every delivered observation exactly as delivered, clock advances, the acknowledgement each effector call returned, consumer reports, the run's close |
| `sealed/journal.jsonl` | **Director** | The director only; in **Explore** mode also the person watching, and the run is stamped unblinded | Everything that makes the run true: every command (faults, perturbations, environment, entities, clock), every delivery-ledger row, every world-state snapshot, every effector call's real outcome |

A record is **truth-bearing** when knowing it would tell a consumer something the world has not
shown it: an injected fault, a perturbation, which deliveries were dropped or altered (so no
emission counts or delivery identifiers beside the delivered events), the true state, the real
outcome of an effector call. Truth-bearing records never go to the observer stream. The tests in
§6 enforce this.

An observation is the adapter's rendered line, exactly as the sink received it, not the native event:
the native event carries its emission number (a `drop` perturbation leaves gaps in it) and fields an
adapter may hide. The observer stream shows a consumer exactly what its trace shows, record for record.

The artifact files the run already publishes keep their meaning: `trace.jsonl` is the
consumer's delivered bytes (consumer-readable); `ledger.jsonl`, `world_state_history.jsonl` and
`run.json` are director-only. The consumer-readable files of a run directory are exactly
`journal.jsonl` and `trace.jsonl`.

## 2. Records

Every line is one JSON object with `stream`, `seq`, `type`, `at_ns` (the world clock when the
record was written; never a wall-clock read) and the type's fields. Keys are written in sorted
order. The envelope and every type's fields are closed by the contract.

**`seq`** numbers the stream's **replayable** records from 0 without gaps. A record is replayable
when replay can rebuild it from the command log. A **session record** (`consumer.reported`,
reserved `pacing`) reports something the command log does not hold (what a consumer said, how a
live session was paced): it is in the file, carries the `seq` of the replayable record that follows
it and the `at_ns` of the one before, and is outside the stream digest (§3).

### Observer stream

| `type` | Fields | Written |
|---|---|---|
| `run.opened` | `run_id`, `sim_version`, `domain {id, version, digest}`, `adapter {id, version, digest}`, `start_time_ns`, `entity_ids`, `mode` (`evaluate` or `explore`). **No seed**: MCP_SURFACE §4.5 never exposes it, and with the seed and the domain a consumer could regenerate the fault-free world and diff it against what it received | At creation |
| `observation` | `line`: the record exactly as the sink received it (the adapter's rendered line); `malformed: true` when that line was delivered broken; `framing: "preamble"` or `"postamble"` for an adapter framing line, which is in the trace but is no delivered event | Per delivered record, in delivery order; the preamble after `run.opened`, the postamble before `run.closed` |
| `clock.advanced` | `to_ns` | After each advance |
| `effector.acknowledged` | `command_id`, `effector`, `entity_id`, `response` (`model.EffectorAcknowledgement`, exactly what the operator received) | Per acknowledged invocation |
| `consumer.reported` (session) | `quiesced_through_ns`, `verdict_submitted` | Per consumer report |
| `wire` (session) | One plant wire record: `direction` (`node` sent it, `host` sent it to the node) and the `record` exactly as on the wire (phase P7d). Outside the stream digest; the wire is reproduced by replaying the host's records | - |
| `pacing` (session) | One plant wire barrier: `through_seq`, `waited_ms` (wall time), `missed` (phase P7d); live control adds its own pacing in P10 | - |
| `run.closed` | `incomplete` | At `End` |

An effector call the world declines is not acknowledged: the operator received a refusal, not an
acknowledgement, so it appears only in the director stream. An acknowledgement is the same for
every acknowledging failure mode (MCP_SURFACE §4.3), so this record cannot tell a consumer which
mode the world chose.

### Director stream

| `type` | Fields |
|---|---|
| `run.opened` | `run_id`, `seed` (the seed lives here and in the artifact, never in the observer stream) |
| `command` | `command {seq, at_ns, op, args}`: the command-log entry. A `clock.advance` carries only `to_ns`: whether a live session waited for its consumer (`await_consumer`) is pacing, which replay does not reproduce, so it is not part of the stream's replay identity (the artifact's command log keeps it) |
| `delivery` | `delivery`: the delivery-ledger row |
| `state` | `state`: the world-state snapshot taken at an emission |
| `effector.outcome` | `command_id`, `effector`, `entity_id` and either `result` (the world's `InvokeResult`: failure mode, whether the effect applied, latency) or `refused` (why the call was declined) |
| `run.unblinded` | `reason`: `explore` (written right after `run.opened`, at creation) or `revealed` (truth was revealed on the open run; written at the close, before `run.closed`, so replay can reproduce it). The wall-clock moment is in the artifact (`unblinded_at`), not in the stream |
| `kpi.truth` | `kpis`: the run's KPI truth (domain-spec v0.2 §29), written once at the close, before `run.unblinded` and `run.closed`, so replay reproduces it |
| `run.closed` | `incomplete`, `error` (when the run failed) |

## 3. Determinism and verification

Both streams are pure functions of the run (no wall-clock reads, maps encoded in sorted order).
A stream's **digest** is the SHA-256 (`sha256:` and hex) of its replayable lines, newline
included. The run artifact records `journal_digest` (observer) and `sealed_journal_digest`
(director). Computing a digest needs no file: a run with no journal directory keeps the digests
only. It also records `journal_file_digest` and `sealed_journal_file_digest`, the SHA-256 of every line as written, session
records (`consumer.reported`, `wire`, `pacing`) included (R7e-08): replay cannot reproduce session records, so `verify`
does not compare these, but the run report checks the files against them, so an edited session record is evident.

`replay` also compares the artifact's `unblinded` flag with the stamp the replay derives (an explore
run is unblinded by its `mode`): clearing the flag in `run.json` fails `verify`, and `score` refuses
an explore run by its `mode` as well as by its flag.

`replay` rebuilds both streams in memory. `verify` fails with `replay_mismatch` when either
rebuilt digest differs from the artifact's, and `verify` of an artifact file also compares the
replayable lines of `journal.jsonl` and `sealed/journal.jsonl` beside it with the rebuilt ones, so a
file edited after the run is caught. Either way the failure names the stream
(`ReplayResult.JournalDiverged`, `detail`) and, from the files, the first differing record. A run
that ended incomplete closed its journal in a way replay cannot reproduce, so its journals are not
compared; its trace still is.

## 4. Durability and crashes

Records are appended through a buffered writer, flushed at every recorded command and every
advance boundary, and synced at `End`, like the delivery ledger. Each record is one line. A crash
can leave a final line without its newline: the reader ignores an unterminated last line and
treats every complete line as authoritative, so a journal cut at any byte reads up to its last
complete line. A journal that stops without `run.closed` is an **incomplete** run, playable up to
its last complete record.

A record that cannot be encoded, or a file that cannot be written, is remembered by its stream,
stops the stream, and fails the run at the next boundary and at `End` (the artifact is returned
incomplete with the error); it is never silently dropped. A journal directory that cannot be created
fails run creation.

## 5. Where journals are written

| Surface | Directory |
|---|---|
| `streamsim run --out <dir>` | `<dir>` (no `--out`: no journal file) |
| Director (`sim.world.create`) | The world's run directory beside its ledger |
| Replay and verify | In memory only |

`run --explore` (and a director world created with `explore: true`) stamps the run unblinded
from its creation (`run.opened.mode = explore`, a `run.unblinded` record, artifact `mode: "explore"`
and `unblinded: true`), so scoring refuses it: `streamsim score` fails with `score_failed` and the
director's `sim.score` with `run_unblinded`. Scoring refuses every run stamped unblinded, including
one whose truth was revealed while it was open.

## 6. Proof

- **Canary**: an Evaluate run injects a fault and a perturbation; no consumer-readable file
  (`journal.jsonl`, `trace.jsonl`) contains the fault id, the fault name, the perturbation name,
  the `seed` key or the seed's value; the director stream does. With a `drop` perturbation the
  observer journal's observations are exactly the trace's lines, in order. Proven at the run layer, at the director layer and through the binary.
  Bridge payloads and the page are proven in P10 (finding R0-01b).
- **Replay identity**: both journals are byte-identical across runs and replays of the same
  inputs (a closed-loop run with a consumer, its acknowledgements and its reports included);
  `verify` detects a changed record in either stream, in the file and in the artifact's digest.
- **Crash**: a journal cut at any byte reads up to its last complete line, and the torn line does
  not change the digest; the files of a run that has not ended are readable.
- **Explore**: an Explore run is stamped unblinded and `score` refuses it; clearing the flag in
  `run.json` fails `verify` and `score` still refuses.
- **Contract**: every record the writer produces validates against `run-journal-v0.1`; the
  embedded copy is byte-identical to `contracts`.
- **Reader**: `FuzzJournalRead` (in `make fuzz` and `make fuzz-soak`).

## 7. Not in this phase

- The `pacing` records of live control (P10).
- Journal rotation or size bounds: a journal is as large as the run (a delivery produces one
  observation, one delivery row and one state row).
- Reading a journal while it grows (follow, offsets): the files are plain append-only JSON Lines,
  and the canvas follower is P9/P10 work.
