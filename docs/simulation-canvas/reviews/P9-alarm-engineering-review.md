# P9 alarm playback — round 7 engineering review

Date: 2026-10-14. Scope: R9-24 alarm activation truth and playback only. This is not approval of P9.

## Frozen build and method

Reviewed frozen binary SHA-256 `96e3450742b948dadb85931a54d6e254f2950000309d3b2da731d042c3c6ec0a`
and UI JavaScript SHA-256 `3f7b7e5758d1665598a0314792dde6d4059b0015790c9dd166afbd7268cfefc1` from
`/private/tmp/streamsim-p9-alarm-review/manifest.json`. Started my own real-binary bridge with the frozen UI and
workspace, PID 45559, then paired a separate installed Chrome tab. The bridge used the manifest's workspace root,
which contains the `w-alarm` run. Stopped only PID 45559 after review and verified it had exited. No mocked binary
or fixed sleeps were used.

## Evidence and result

- `w-alarm` opened with `Verified` and one activation. Chrome displayed the reported event as
  `00:02:20 HEAT-1 temp alarm activated`. Selecting it moved playback to 140 seconds, selected HEAT-1, opened
  Inspect, and showed the observed `85.0 °C` at `00:02:20`. Starting from 00:05:00, Jump to alarm returned to
  the same 00:02:20 sample and observation.
- The frozen `sealed/journal.jsonl` contains the sealed `kpi.truth` activation instant
  `1767225740000000000`, exactly start `1767225600000000000` plus 140 seconds. Report verification in Chrome
  says `Reproduces`; the report displays one activation. Alarm marks are created from reported activation times;
  no active interval or cleared instant is inferred.
- `TestAlarmActivationInstantsMatchTheIndependentScanAndAreOwnedCopies` checks exact absolute instants against
  the independent alarm scan, matches the activation count, and mutates the returned slice to prove ownership.
  `TestAnInitiallyStandingAlarmHasNoActivationInstant` checks that the first active sample contributes no
  activation. Both focused Go packages passed: `go test ./internal/world/internal/domain ./internal/report/internal/domain`.
- `TestReportForwardsAlarmActivationInstantsWithoutReconstructingThem` checks report data and serialized JSON.
  `TestAnOpenEvaluateReportDoesNotExposeAlarmActivationInstants` confirms an open Evaluate report has no truth
  and reports the sealed reason. The app passes the verify state into Inspect's KPI-only `ReportPanel`.
- The real-binary Playwright regression checks the activation chip and Jump to alarm, then asserts the observed
  85 °C value at the exact scan instant. The integrity E2E additionally selects an entity in the incomplete-stamp
  run and asserts the unverified-truth caution remains beside Inspect KPI truth.
- The behavior-pin diff reflects the version bump in `run.json` and observer `journal.jsonl` digests; trace,
  ledger, state history, sealed journal, and verify/replay verdict entries remain unchanged. The frozen run is
  from simulator 0.3.2. No gate, script, facade, or gate threshold change was found in this batch.

## Findings and disposition

| ID | Severity | Finding | Disposition |
|---|---|---|---|
| E8-1 (carry-over E7-3) | Medium | The canvas stores absolute nanoseconds in JavaScript `number`, which cannot represent every int64 epoch-nanosecond value. This review proves the fixture's whole-second 140-second activation, not sub-millisecond navigation precision for every allowed alarm grid. | Scoped R9-24 Chrome flow passes. Do not claim nanosecond UI precision. Preserve activation instants as decimal nanoseconds or integral relative milliseconds in the bounded playback-query batch (R9-13/R9-05); retain as an open contract item until that query is implemented and tested. |

No other R9-24 defect was found. The Go ownership and activation-time semantics, sealed report forwarding, Evaluate truth gate, Inspect verification propagation, and alarm navigation pass this bounded review.

**Scoped score: 4/5.** R9-24 passes for this second-grid fixture and frozen build, with E8-1 carried into the playback precision contract. **P9 is not signed off:** other functional batches and open P9 findings remain, including R9-13/R9-14, R9-22, U6-1/U6-2, and E6-boundary. The full phase gate was still running separately when this review was requested; this review does not claim it passed.
