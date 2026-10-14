# P9 gate environment findings

Date: 2026-10-14. These findings are recorded separately from feature and UI approval.

| Finding | Evidence | Resolution |
|---|---|---|
| Parallel artifact cleanup interrupted the architecture walk | First phase-gate race run reported `openat canvas/test-results: no such file or directory` while a separately launched Playwright run removed its old results. | Finish independent Chrome jobs before running the full gate; no architecture check changed. |
| Release tools absent from default PATH | `make phase-gate` reached deadcode after both Go suites and coverage, then reported the tool missing; both release tools already exist in the user's Go bin directory. | Append that directory to PATH, preserving the pinned linter first. |
| Prepending Go bin shadowed the pinned linter | Default linter is 2.14.0; Go bin also contains 2.12.2. Prepending it ran the older analyzer, which reported an unused G703 suppression. | Do not shadow the pinned 2.14.0 executable. |
| Default Go 1.27.2 export format exceeds installed linter support | Pinned linter is built with Go 1.27.1; its typecheck could not decode Go 1.27.2 standard-library export data. | Use CI's pinned Go 1.26.9 toolchain. Do not alter gate configuration. |
| Go 1.27.1 fails the vulnerability gate | With the compatible older compiler, the full gate passed Go tests and deadcode, but govulncheck reported 11 reachable standard-library vulnerabilities. | Keep the vulnerability gate intact; use the patched CI toolchain instead. |
| Apparent unused G703 suppression was another analyzer-shadowing symptom | A trial removal caused the pinned 2.14.0 analyzer to report G703, proving that the original suppression is required by that analyzer; the older 2.12.2 analyzer had called it unused. | Restore the original suppression and keep the original static-serving code unchanged. Append release tools to PATH, never prepend them. No suppression or gate change remains in the diff. |
| Behavior pin includes compiler platform metadata | The script masks wall-clock fields but includes `platform.go_version` in run.json and the manifest; changing the compiler changes 18 artifact hashes and the manifest-file hash. | Record the compiler alongside the intentional simulator-version pin update. Review every difference; trace, ledger, history, sealed journal and replay/verify verdict hashes remain unchanged for the shipped pin fixtures. Do not mask new fields or change the script. |

The enforced Enola baseline was pinned with the local CLI before edits. The MCP server has an older/stale snapshot
and cannot substitute for this CLI baseline. Keep the existing baseline; do not re-pin it after edits.
No full phase-gate pass is claimed until the final command with the pinned toolchain completes.
