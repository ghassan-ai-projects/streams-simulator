# P9 report memory — design and verification plan

R9-14, report portion. **Implemented (report portion only)**; see Results. Prepared in an isolated worktree
and copied here for the next engineer. This is behavior-preserving for the recorded report documents: no simulation, scoring, KPI or digest
rule changes, and no simulator-version bump. No new Go package or dependency.

## Root cause

Report reads trace, two journals, ledger and history into byte slices, repeatedly decodes all journal records into
retained maps, then retains every ledger row. It needs only file digests, counts, identity/control records and sealed
truth. Memory and repeated parsing grow with observation count, rather than with the returned report.

## Existing owning layers

- `kernel` OS-edge facade over `internal/files`: add a cancellable open-for-read operation returning the standard
  `io.ReadCloser` interface. Keep the facade delegating. Confine resolution to the file's parent directory through
  `os.Root`; callers close the opened file. No record name or digest policy in kernel.
- `journal` pure reader domain: expose its existing one-record decoder through a delegating facade and a pure rule
  deciding whether a committed line participates in the replay digest. Reuse the rules in the existing whole-buffer APIs.
- `report/internal/domain`: incrementally digest and summarize evidence, retaining only records needed by report
  assembly and counts. Domain owns record classification, tally rules, replay/file digest selection, identity and truth.
- `report/internal/files`: bounded chunk/line reading with context cancellation; feeds the domain accumulator and owns
  the reader lifecycle. A single committed line is bounded to 8 MiB; a complete evidence file is streamed without an
  8 MiB total-file cap. Trace/history are raw digests; journal/ledger lines are validated as before.
- `report/internal/app`: pass context into evidence reads and assemble the existing report from the summaries.

Memory target: independent of observation, delivery and state-history record counts; proportional to a bounded
read buffer, the largest allowed record, and metadata/truth included in the returned report. The artifact's own
boundedness remains separately tracked in R2-01; this batch makes no claim to close all input-bound findings.

## Preserved behavior and deliberate edges

File digests include every byte, including a torn tail. Replay journal digests include complete non-session lines
only. A torn final journal/ledger record is ignored for assembly and named in `torn_tail`. All committed records are
decoded/validated even when not retained. First record, seed/identity and last clock semantics remain unchanged.
Missing files, malformed committed records, missing digests, unclosed Evaluate truth and foreign-run evidence retain
their current outcomes. Exact JSON/Markdown before/after comparisons are required on factory and campus runs.

New reads honor context cancellation. The OS reader refuses a symlink escaping the file's parent directory, an
intentional tightening of file resolution with an edge test. No new linter suppression or architecture allowance.

## Tests and measurements

Tests first for the open reader, exact incremental journal digest behavior, summary parity and bounded retained
records, torn/malformed/error cases and cancellation. Existing report package tests stay as oracles. Run focused
tests, strict function length, architecture tests, lint and `make quick`. Compare report output bytes and peak RSS/
wall time with the frozen 26-hall campus and factory, using `/usr/bin/time -l`. Review every facade change before
integrating; run the full phase gate after integration. Record actual results here, not estimates.

## Results (2026-10-14)

Measured on the 26-hall campus run (725,253 observations; 1.02 GB of evidence: journal 260 MB, sealed journal 719 MB,
ledger 191 MB, history 399 MB, trace 169 MB), `/usr/bin/time -l bin/streams-simulator report <run>`, same machine:

| | before | after |
|---|---|---|
| wall time | 31.5 s | 20.7 s |
| peak resident memory | 8.83 GB | 26 MB |
| report JSON | reference | byte-identical (`cmp`) |

As built: `kernel.OpenIfPresent` (cancellable reader, facade delegating); `journal.DecodeLine` and `journal.Replayable`
(the existing per-line rules, exposed); `report/internal/domain.Scanner` summarizes a file one committed line at a time
(raw digest, replay digest, delivered count, ledger tally, the first record and only the record types the report reads);
`report/internal/files.Scan` streams the file (lines bounded to 8 MiB, context checked per line). `journal.Read` was
production-dead after this change and is removed; tests decode through `DecodeLine`. Whole-buffer `Document.Raw` remains
only for the small verdict file and for tests, which share the scanner so parity holds by construction.

Open from this design: the `os.Root` confinement of the reader (a symlink escaping the run folder) was not added; the
existing read path never had it, so nothing is tightened or loosened. Verify, replay, bridge file reads and playback
queries remain tracked under R9-14, R2-01, R9-04 and R9-05.
