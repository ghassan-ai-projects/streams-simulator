# P9 report memory — design and verification plan

R9-14, report portion. **Design only; no implementation has started.** Prepared in an isolated worktree
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
