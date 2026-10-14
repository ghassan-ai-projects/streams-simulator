# P8b evidence — canvas bridge

Design and threat model: `docs/design/CANVAS_BRIDGE.md`. Module: `internal/canvasbridge` (`domain`, `app`, `transport`,
`files`, `process`, `entropy`). Statement coverage with `-short`: facade 100 %, `app` 92.7 %, `domain` 93.8 %, `entropy`
100 %, `files` 90.7 %, `process` 91.4 %, `transport` 95.3 % (floors in `scripts/coverage-floors.txt`).

| Exit criterion | Tests and result |
|---|---|
| T1 loopback only | `transport`: `TestListenRefusesANonLoopbackAddress`, `TestListenBindsLoopbackAndReportsAnOccupiedPort`; `domain`: `TestTheBridgeListensOnLoopbackOnly`; `app`: `TestServingRefusesANonLoopbackAddressAfterPreparingTheWorkspace`; facade: `TestTheFacadeServesThroughTheModuleAndRefusesANonLoopbackAddress` |
| T2 host, origin, fetch site | `TestOnlyTheBoundHostIsServedAgainstDNSRebinding`, `TestCrossOriginRequestsAreRefusedOnEveryAPIRoute`, `TestSameOriginAndNavigationFetchesAreAdmitted`, `TestAnUnknownAPIRouteIsAdmittedThenNotFoundNeverTheUI` |
| T3 pairing and session | `TestEveryAPIRouteButSessionNeedsThePairedSession`, `TestPairingSetsAnHttpOnlyStrictSessionCookie`, `TestAPIPairingFailuresMapToUnauthorized`, `TestTheSessionIsReadFromAmongOtherCookies`; `domain`: `pairing_test.go` (five failures lock, single use, constant time) |
| T4 secrets never in bodies, URLs or logs | `TestTheCodeAndTheSessionNeverAppearInAResponseBody`; acceptance `TestTheCanvasBridgeDrivesAWorldFromPairingToTheReport` (`checkSecretsStayOut`: code and session id absent from every response and from stderr) |
| T5 traversal and symlinks | `files`: `TestReadsStayInsideTheWorkspaceAndWithinTheirLimit`; `transport`: `TestTheUIRootCannotBeLeftByTraversalOrSymlink`; `domain`: `TestOnlyTheWorkspaceFilesOfTheCanvasAreAccepted`; fuzz `FuzzBridgeInput` |
| T6 conditional atomic writes | `files`: `TestAWriteNamesTheVersionItStartedFromAndIsAtomic`; `transport`: `TestAConflictedWriteCarriesTheCurrentEtag`; fix 6 below |
| T7 body limits | `TestOversizedBodiesAreRefusedWith413BeforeTheServiceSeesThem`, `TestBodiesWithinTheLimitsAreAccepted`, `TestMalformedJSONIsRefusedWith400` |
| T8 truth gate | `app`: `TestFilesAreReadAndWrittenThroughTheirRules`; acceptance: sealed files `403 sealed` while open, readable after `run.end`; fixes 1, 2 and 5 below |
| T9 argument injection | `domain`: `TestCommandArgumentsRefuseInjectionTraversalAndUnknownKeys`, `FuzzBridgeInput` (30 s run, 2.4 M executions, clean) |
| T10 headers | `TestEveryResponseCarriesTheSecurityHeaders` (CSP exactly as specified, nosniff, no-referrer, X-Frame-Options, no-store) |
| Status mapping, SSE format, static files | `TestEachDomainErrorMapsToItsStatusAndCode`, `TestAToolOrCommandFailureKeepsItsCodeMessageAndIssues`, `TestARunIsStreamedAsServerSentEventsAndClosesAfterTheClosingRecord`, `TestAFollowFailureIsAnEventNotAStatus`, `TestLastEventIDResumesAfterThatRecord`, `TestTheWorkspaceStreamAnnouncesEachChange`, `TestTheBuiltUIIsServedWithAnIndexFallback` |
| Whole flow through the real binary | `test/acceptance` `TestTheCanvasBridgeDrivesAWorldFromPairingToTheReport` (about 2 s; skipped with `-short`): pair, list, compose with etag, validate, `world.create` (Evaluate), `clock.advance`, `fault.inject`, live journal, sealed refusal, `run.end`, `closed`, resume, `world.destroy`, `run.verify` matches, `run.report` shows the fault, no operator token in any response |
| Input parsers survive arbitrary bytes | `FuzzBridgeInput` (Makefile `FUZZ_NAMES`) |

## Defects fixed after independent review

| # | Defect | Fix | Regression tests |
|---|---|---|---|
| 1 | Critical: an encoded slash in the run id of `/api/runs/{id}/events` streamed `sealed/journal.jsonl` of an open Evaluate run | `Follow` and the new `OpenRun` take one validated run id (`domain.RunJournal`) and always the observer journal | `domain.TestARunJournalIsNamedByOneRunIdAndNeverReachesASealedFile`; `app.TestFollowingAndOpeningARunTakeOneRunIdAndNeverASealedPath`; `transport.TestAnEncodedSlashInTheRunIdIsRefusedBeforeAnyStreamOpens`; acceptance (`%2Fsealed` is 400) |
| 2 | Critical: a stale `run.json` unlocked sealed files of a new open run (world ids restart at `w-1`, folders were reused) | (a) the director takes the next world id whose folder does not exist (`run.FolderUsed`); (b) the bridge calls a run ended only when its observer journal ends with `run.closed` and the world is no longer running in the bridge | `director/internal/app.TestANewWorldNeverReusesTheFolderOfAnEarlierRun`; `run/internal/durable.TestAFolderIsUsedOnceItExists`; `domain.TestAJournalIsClosedOnlyWhenItsLastRecordClosesTheRun`; `app.TestAStaleRunArtifactDoesNotUnlockTheSealedFilesOfAnOpenRun`, `TestARunIsEndedOnlyOnceItsJournalClosedAndItsWorldStoppedRunning` |
| 3 | High: one mutex held during director calls blocked cookie checks, streams and saves behind a long `clock.advance` | Separate locks: director calls, pairing and world sets, file writes | `app.TestPairingFilesAndAdmissionAreNotBlockedByALongDirectorCall` |
| 4 | High: a dead director was never replaced; its stderr was dropped | `process.ErrGone` on a lost connection; the bridge drops the child, clears its worlds and retries once; the stderr tail is in the message | `process.TestADirectorThatDiesIsReportedGoneWithWhatItSaid`, `TestADirectorThatCannotStartSaysWhy`; `app.TestADeadDirectorIsReplacedAndTheCallRetriedOnce`, `TestADirectorThatKeepsFailingFailsTheCallWithoutLooping` |
| 5 | Medium: resuming at the closing record never sent `closed`; an unknown run followed forever | A skipped closing record sends `closed` alone (`Record.Seen`); `OpenRun` answers `404` before the stream opens | `app.TestResumingAfterTheClosingRecordStillTellsTheStreamItClosed`; `transport.TestAClosingRecordTheClientAlreadyHoldsOnlyTellsTheStreamItClosed`, `TestARunThatDoesNotExistIsNotFoundNotAnEndlessStream`; acceptance |
| 6 | Medium: `domain.compose` and `scenario.export` overwrote files without a version check | Optional `etag`; an existing `out` without the current `etag` is `409` with the current `etag`, checked under the write lock before the command | `domain.TestOnlyCompositionAndScenarioExportGuardTheirOutputWithAVersion`; `app.TestACommandThatWouldOverwriteAFileNeedsTheVersionItStartedFrom`, `TestScenarioExportIsGuardedLikeComposition`; `transport.TestAnActionRefusedForAStaleVersionCarriesTheCurrentEtag`; acceptance |
| 7 | Medium: a session cookie lost with the browser forced a bridge restart | The cookie lasts 12 hours (`Max-Age`); the code stays single use, the lockout stays | `transport.TestPairingSetsAnHttpOnlyStrictSessionCookie` |
| 8 | Low: a write could exceed the read limit | The write limit is the read limit | `files.TestAWriteLargerThanTheReadLimitIsRefused`; `app.TestAFileLargerThanAReadMayBeIsNotWritten` |
| 9 | Low: a symlink out of the workspace was a 500 | `400 bad_path` | `files.TestReadsStayInsideTheWorkspaceAndWithinTheirLimit`; `app.TestARunReadThroughASymlinkOutOfTheWorkspaceIsABadPath` |
| 10 | Low: errors and results named the absolute workspace path | Removed from every relayed error and result (`domain.Scrub`) | `domain.TestWorkspacePathsAreScrubbedFromMessages`; `app.TestNoErrorOrResultNamesTheAbsoluteWorkspacePath` |
| 11 | Found while writing the transport tests: any unknown `GET /api/...` answered `200` with `index.html` and skipped admission | `GET /api/` is `404 not_found` behind admission | `transport.TestAnUnknownAPIRouteIsAdmittedThenNotFoundNeverTheUI` |

Known limits, unchanged: a run file over 8 MiB (a long run's `trace.jsonl` or `sealed/journal.jsonl`) is `413` through
`GET /api/file`; the canvas reads such a run through `run.report`. The protocol table of the design now names `422` for a
refused tool or command and `500` for a bridge failure (the first draft said `502`).
