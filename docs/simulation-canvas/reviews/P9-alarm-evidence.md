# P9 alarm-playback evidence

Date: 2026-10-14. Branch: `codex/canvas-p9-playback`. R9-24 completed; P9 remains In progress.

## Change and review

Simulator 0.3.2 publishes the existing alarm scan's counted activation instants in sealed KPI truth and the report.
The canvas presents the instants as alarm chips and Jump to alarm. Navigation seeks, selects the source and opens
Inspect. An incomplete verify rejection now also warns beside Inspect KPI truth, with the ready verdict propagated
through RunPanel. No simulation math, RNG draws, alarm counts or alarm dynamics changed.

Round 7 scoped scores: design 5/5, desktop input 5/5, engineering 4/5, no High in this batch. Review files:
[design](P9-alarm-design-review.md), [input](P9-alarm-input-review.md),
[engineering](P9-alarm-engineering-review.md). R9-37 tracks precise time input to the next bounded query;
the current UI makes no nanosecond-precision claim.

All three reviewers used installed Chrome and their own bridges, stopped by PID. The frozen one-heater run's
activation is at 140 seconds; chip and jump show 02:20, HEAT-1 and the observed 85 °C reading. Light/dark checks
cover 1280×900, 1440×900 and 1920×1080. Frozen artifact hashes are in each review record.

## Gates

- `make quick`: passed.
- Exact activation scan, owned-copy, initial-standing and report truth-boundary Go tests: passed.
- Model boundary test: exact nanoseconds beyond floating-point precision survive Go JSON serialization;
  also passed separately under the race detector.
- Canvas unit/component coverage and the real-binary Chrome suite: passed in the full phase gate.
- Full command: `GOTOOLCHAIN=go1.26.9 PATH="$PATH:$HOME/go/bin" make phase-gate`: **passed**.
  This ran the two Go test tiers, coverage floors, pinned lint, deadcode, vulnerability check, fuzz, canvas checks,
  enforced Enola delta and behavior-pin comparison. Toolchain/PATH failures and their resolution are recorded in
  [P9-gate-environment.md](P9-gate-environment.md); no gate was bypassed.
- `git diff --check`: passed. `git diff main -- test/architecture scripts`: empty. No facade or dependency changes.

## Behavior pin

201 entries; 37 differ from main: 18 run.json, 18 observer journal hashes and the manifest-file hash.
The simulator version deliberately changes run/observer identity. The compiler's platform metadata differs from
the prior local pin; final evidence uses CI's Go 1.26.9. Every trace, ledger, state history, sealed journal and
replay/verify verdict entry in the shipped pin fixtures is unchanged. The pin script and its masking rules are unchanged.
Alarm-bearing runs additionally carry activation instants in sealed KPI truth; dedicated tests prove that deliberate
contract change. Re-record review runs after this simulator change; old frozen runs are not new-build verification evidence.

No push or new PR has occurred. Remaining P9 functional findings stay open; visual optimization is deferred
until functionality per the owner's direction.
