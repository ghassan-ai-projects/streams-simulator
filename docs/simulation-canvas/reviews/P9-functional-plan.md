# P9 functional completion

Owner direction, 2026-10-14: finish functionality before optimizing visual details. Implement directly;
retain Luna code and Chrome reviewers. Every finding and disposition is persisted in the reviews folder.

## Ordered batches

1. R9-24: publish alarm activation instants from the existing world KPI tracker, through sealed KPI truth and
   the report; show alarm chips and a jump-to-alarm control in playback. Prove exact instants against the existing
   independent alarm scan, report preservation, and installed-Chrome end-to-end navigation.
2. R9-13/R9-05/R9-14: bounded simulator-owned playback/true-state queries and streaming report/verify evidence;
   canvas displays those results. Measure campus memory/time and browser responsiveness, not estimates.
3. R9-22: correct measurement semantics using declared instrumentation bounds without hiding faults or
   changing the observer/oracle relationship. Specify the rule before implementation and prove RNG/replay determinism.
4. Remaining functional findings: U6-1 geometric hall-fit reproduction, U6-2 direct graph keyboard route,
   E6-boundary strict typed result admission, and all findings assigned to P9 before marking Done.
5. Full gates and fresh frozen Chrome review; then P10 live control and P11 builder, with A1–A7 against the real binary.

## Alarm activation contract (R9-24)

Existing module/layer: `world` pure-core domain owns alarm scan samples and activation recognition;
`model` is the shared foundation for the emitted record. No new package. The existing `AlarmKPI` gains
`activation_times_ns`, an ordered array of absolute simulation instants, including exactly the samples already
counted as activations (codes 1/2, after the initial sample, with unshelving counted by the existing rule).
The returned slice is an owned copy. Initial standing alarms are not newly counted.

The existing sealed `kpi.truth` record carries the field to `report.truth.kpis.entities[].alarms[]`.
The report forwards it; it does not reconstruct activations. No new bridge privilege or observer truth record.
Evaluate does not expose it before sealing; Explore follows the existing truth gate. Canvas flattens the
reported instants into display/navigation marks and computes no alarm transition or KPI.
Alarm marks distinguish activation from active-now state; no cleared instant or duration is inferred.

Adding the field changes sealed journal and run digests: deliberately bump simulator version to 0.3.2,
document the change in CHANGELOG, and regenerate the behavior pin with the exact differences reviewed.
Existing aggregate counts and alarm dynamics remain unchanged. Record new review runs after this artifact change.
