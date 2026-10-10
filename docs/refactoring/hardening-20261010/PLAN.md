# Plan and status

Branch `harden-weak-spots`, from `main` after PR #18. Tiers ran in order;
within a tier, items are independent.

| Tier | Scope | Items |
| --- | --- | --- |
| 1 | Oracle, evidence and safety (severity H) | D-01, D-02, D-03, D-04, D-05, D-06 |
| 2 | Run lifecycle and concurrency | D-07, D-08, D-09, D-10, D-15, D-16, D-17 |
| 3 | Truth, world, consumer and surface correctness | D-11, D-12, D-13, D-19, D-20, D-21, D-22, D-23, D-24, D-25, D-27, D-41, D-42, D-50, D-51 |
| 4 | Loaders, adapters, schemas, sinks | D-26, D-28, D-30, D-31, D-32, D-33, D-34, D-38, D-40 |
| 5 | Structure and test weak spots | D-46 (mcp protocol edge), D-47, D-48, T-01 (error assertions), run facade test hooks, determinism-gate gaps D-43 |

## Decisions taken

The branch was asked to "fix all weak spots"; the items below changed a
number, a digest or a stored format, so they are recorded here as decisions
for the reviewer of the pull request to confirm or reverse. Each was taken
with the recommendation originally written for it, except D-36/D-39 (the
canonical digest replaced the legacy one instead of sitting beside it, because
two digests of one document would each need a meaning).

| ID | Decision | Consequence |
| --- | --- | --- |
| D-14 | Reads are pure; the integration grid no longer depends on the read schedule | Simulator 0.1.0 → 0.2.0; world-state histories and run artifacts moved |
| D-18, P-07 | One delivery path for live and released records | Ledger event time and sink-failure rows follow the unified path |
| D-35 | Zero equals unset for omitempty floats: kept | Documented in the domain contract |
| D-36 | The adapter digest is the RFC 8785 digest of the adapter's JSON form | Adapter digests in every run artifact changed; 0.1.0 artifacts replay with a version mismatch (note 8) |
| D-39 | Verbatim integers in canonical JSON: kept | Documented in `canonical` |
| P-01 | The `wall` seam is wired into the run clock edge rather than deleted | DECISIONS D-13 amended |

## Status

| Item | State | Regression test |
| --- | --- | --- |
| D-01 | done | `TestRevealRefusesALabelForARunTheDirectorDoesNotKnow` |
| D-02 | done (truthful, note 1) | `TestReplayDivergenceSaysWhatIsAndIsNotKnown` |
| D-03 | done | `TestVerifyFailsTheCommandWhenTheReplayDoesNotReproduceTheArtifact` |
| D-04 | done (note 2) | `TestOfflineScoreEqualsTheOnlineScoreOfTheSameRun` |
| D-05 | done (note 3) | `TestEverySafeStopAppliesItsOwnWorldEffect` |
| D-06 | done | `TestWireFaultsAreNotReplayedWithTheEvidenceOfARetry` |
| D-07 | done | `TestARunThatEndedWithAFailureIsStillClosed`, `TestDestroyWorldAfterAFailedEndCanBeRetried` |
| D-08 | done | `TestConcurrentBeginRunOpensTheRunOnce` (-race), `TestBeginRunAndRevealTruthNeverDeadlock` |
| D-09 | done | `TestArtifactCountsEveryInjectedFaultEvenOnceCleared` |
| D-10 | done | `TestReplayRebuildsTheWorldIdentityInputsOfTheArtifact` |
| D-15 | done | `TestEveryCommandIsRefusedOnceTheRunIsFinished` |
| D-16 | done | `TestFailedNewLeaksNoFileDescriptor`, `TestEndKeepsTheArtifactWhenItsEvidenceCannotBePublished` |
| D-17 | done | `TestEvidenceReadsDoNotRaceWithAnAdvancingRun`, `TestWorldReadsDoNotOverlapAnAdvance` (both -race) |
| D-11, D-12, D-13 | done (note 4) | `TestAnOnsetObservableAtTimeZeroIsObservable`, `TestAnUnknownDetectorFormIsRefusedNotTreatedAsObservable`, `TestAFaultWithNoEffectIsNotObservableEvenWithoutNoise`, `TestPeerResidualSigmaCountsTheEntitiesTheWorldHolds` |
| D-19, D-41 | done | `TestScriptedEffectorsCarryTheirArgumentsAndDeterministicIds` |
| D-20 | done | `TestAuditScenarioWithoutAWindowAuditsTheSuiteScenarioLength` |
| D-21 | done | `TestSilenceDetectionCitesTheLastDeliveredRecordOfTheSeries` |
| D-22 | done (refuse; note 5) | `TestWorldCreateRefusesSeedsThatJSONCannotCarryExactly`, `TestIntegerTypeAcceptsEveryWholeNumberInRange` |
| D-23 | done | `TestAdvanceErrorsKeepTheirOwnCodes` |
| D-24 | done | `TestReportChangesNothingWhenItsVerdictIsRefused` |
| D-25 | done | `TestReadingNeverAltersWhatARunEmits` |
| D-26 | done | `TestReorderSwapsAboutHalfOfTheWholeSecondPairs`; pinned reorder/combined digests updated |
| D-27, D-42 | done | `TestCommandIDReusedForADifferentRequestIsRefused`, `TestIdempotentReplayReturnsTheFirstResultIncludingItsEffectETA`, `TestOmittedArgumentsAndAnEmptyObjectAreOneRequest`, `TestReplayReproducesARefusedEffectorInvocation` |
| D-51 | done | `TestWorldCreateReturnsTheRunIdTruthIsSealedAgainst` |
| D-30 | done | `TestVerifyRefusesAnEmptyFixtureInsteadOfPanicking` |
| D-31 | done (note 6) | `TestHashSuffixKeepsTheLimitAndStaysHexWhateverTheLimit` |
| D-32 | done | `TestSplitRecordsReadsJSONArrayOutputElementwise`, `TestAdapterVerdictFailsOnlyForADeclaredCheckOrForNothingToCheck` |
| D-33 | done | `TestCompileRefusesMalformedCountAndUniqueKeywords` |
| D-34 | done | `TestF1InputExplicitZeroCoefficientSurvivesARoundTrip` |
| D-38 | done | `TestCompileReportsTheFirstBrokenDefinitionByName`, `TestProfileValidationNamesTheSortedFirstUndeclaredFault` |
| D-40 | done | removed the dead default (no behaviour) |
| D-28, D-50 | closed, not defects | see DEFERRED; measured / traced |
| D-35, D-39 | decided: kept, documented | DEFERRED.md; the domain and adapter contracts state the behaviour |
| D-14 | done (note 7) | `TestStateValueDoesNotDependOnTheReadSchedule`, `TestStochasticFaultEnvelopeDoesNotDependOnTheReadSchedule` |
| D-18, P-07 | done | `TestSinkFailureStillAccountsForEveryDelivery` |
| D-36 | done (note 8) | `TestAdapterDigestIsTheCanonicalDigestOfItsJSONForm` |
| D-43 | done | the determinism gate type-checks the module (`go/packages`) and flags every map range without a stated reason; 13 order-sensitive sites iterate sorted keys, the rest carry a `determinism-safe` reason |
| D-44, D-45 | done | `TestEntitySnapshotCannotChangeTheWorld`, `TestEffectorCallLogIsIsolatedFromCallersArguments` |
| D-46 | done | mcp split into `app` (use cases, errors, operator view, capability) and `protocol` (tools, schemas, SDK); gate rank `protocol` between app and facade |
| D-47, D-48 | done | device setters removed; quiescence clock private; no test-only ledger query |
| T-01 | done | 51 negative tests name the error they expect; `TestErrorAssertionsNameTheErrorTheyExpect` has no exceptions |
| P-01 | done | the run clock edge reads through `wall`; `clock_simdet_test.go` pins the zero time under `simdet`; DECISIONS D-13 amended |
| P-03 | done for failure modes | `world.FailureMode` is a named string type (JSON-neutral); the other closed sets keep their constants |
| P-05, P-06 | done | `SinkBroker`/`TimeScaled` removed and MCP refuses scaled time; `ClockMultiplier` documented as reserved; profile names are `model.Profile*` constants |
| P-02, P-04 | decided: not adopted | P-02: exported symbols stay documented, comments inside modules stay; P-04: the device's record maps are the wire form the contract tests pin byte for byte, and no defect traces to them |
| R-1 … R-6 | done (note 9) | `TestBeginRunAndRevealTruthNeverDeadlock`, `TestReplayReproducesARefusedEffectorInvocation`, `TestOmittedArgumentsAndAnEmptyObjectAreOneRequest`, `TestReplayOfAnotherVersionReportsItsInputDigestsInsteadOfFailing`, `TestDestroyWorldAfterAFailedEndCanBeRetried`, `TestWorldReadsDoNotOverlapAnAdvance` |

## Notes

1. **D-02.** The artifact stores only the trace digest, so the position of the
   first differing record cannot be known when the counts agree. The result
   says so in `detail` instead of implying a position; per-record digests in
   the artifact would be a format change.
2. **D-04.** `streamsim score` needs `--domains-dir`/`--adapters-dir` (as
   `replay` does); it builds the same `score.Evidence` the director builds
   and calls the same `score.Score`. `score.Offline` remains for callers
   without a replayable artifact. One input the files cannot carry: a
   `reveal` with `unblind:true` after the run ended is not in the artifact,
   so only the online scorer sees it.
3. **D-05.** World command ids for safe stops are `safe-stop/<target>/<n>`
   (was `safe-stop/<target>`), visible in the effector-call log.
4. **D-11…D-13.** `suite/cold-chain-transit` and its suite file changed: the
   `power_transfer_gap` scenario is no longer claimed observable because its
   declared detector channel has zero noise and zero deviation (D-52, left to
   the domain owner because changing it changes the domain digest).
5. **D-22.** The SDK re-marshals arguments through float64 before the handler,
   so exact seeds above 2^53-1 cannot be carried over MCP; MCP refuses them
   in the schema instead of rounding. The CLI `--seed` keeps the full uint64
   range (the schema `integer` check now accepts whole numbers of any
   magnitude, so an artifact of seed 2^64-1 loads and verifies).
6. **D-31.** `hash_suffix` ids are `sha256:` plus 16 hex digits (was 9). No
   shipped adapter uses the keyword, but a consumer that parsed the width
   would notice.
7. **Version.** Simulator 0.1.0 → 0.2.0. The integration grid (D-14), the
   oracle's solver (D-11/12/13), the reorder perturbation (D-26) and the
   canonical adapter digest (D-36) changed numbers or digests. Pinned traces
   and ledgers are unchanged; world-state histories, run artifacts
   (`sim_version`, world digest, adapter digest) and the cold-chain suite
   moved, and the pin file was regenerated. The MCP `serverInfo.version` now
   reports `model.SimVersion` instead of a constant that outlived the bump.
8. **Replaying a 0.1.0 artifact.** Its stored adapter digest was computed by
   the old rule, so the digests differ. `replay` and `verify` do not fail on
   that: the result carries `version_match: false` and a `detail` naming the
   digests that were not reproduced (a same-version artifact with a wrong
   digest is still refused before execution). Whether the trace digest
   reproduces is then decided by the replay itself.
9. **Review follow-ups (R-1 … R-6)**, found by the independent review of this
   branch: a lock-order deadlock between `run.begin` and `truth.reveal`
   (introduced by D-08); replay aborting on a logged refused invocation
   (a command_id reused for another request, new with D-27); the idempotency
   key telling omitted arguments from `{}` and sharing a key between
   unencodable arguments; the version check running after the digest check;
   `DestroyWorld` wedging a world whose `End` failed; and director reads of
   the world outside the run's command lock (D-17 was partial).
10. **Process.** Commits `765257e` and `2d0d6b7` leave the behaviour pin
    failing for the cold-chain suite; the pin is regenerated in `92ace18`.
    History is not rewritten; bisect across those two commits needs the pin
    from `92ace18`.
