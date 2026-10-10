# Plan and status

Branch `harden-weak-spots`, from `main` after PR #18. Tiers run in order;
within a tier, items are independent.

| Tier | Scope | Items |
| --- | --- | --- |
| 1 | Oracle, evidence and safety (severity H) | D-01, D-02, D-03, D-04, D-05, D-06 |
| 2 | Run lifecycle and concurrency | D-07, D-08, D-09, D-10, D-15, D-16, D-17 |
| 3 | Truth, world, consumer and surface correctness | D-11, D-12, D-13, D-19, D-20, D-21, D-22, D-23, D-24, D-25, D-27, D-41, D-42, D-50, D-51 |
| 4 | Loaders, adapters, schemas, sinks | D-26, D-28, D-30, D-31, D-32, D-33, D-34, D-38, D-40 |
| 5 | Structure and test weak spots | D-46 (mcp protocol edge), D-47, D-48, T-01 (error assertions), run facade test hooks, determinism-gate gaps D-43 |

## Needs a decision (not applied unprompted)

| ID | Why | Recommendation |
| --- | --- | --- |
| D-14 | Making reads pure changes simulated numbers | Pin the current numbers; make reads pure behind a new `sim_version` |
| D-18 | Two delivery paths diverge; intent unclear | Choose the ledger-first path, unify, record as a version change |
| D-35 | Zero equals unset for omitempty floats | Keep; document in the domain contract |
| D-36, D-39 | Changing digests invalidates stored artifacts | Keep the legacy digest, add the RFC 8785 digest beside it |
| P-01 | `wall` seam has no importers | Delete the package; amend DECISIONS D-13 |

## Status

| Item | State | Commit | Regression test |
| --- | --- | --- | --- |
| D-01 | done | `see git log` | `TestRevealRefusesALabelForARunTheDirectorDoesNotKnow` |
| D-02 | done (truthful, see note) | `see git log` | `TestReplayDivergenceSaysWhatIsAndIsNotKnown` |
| D-03 | done | `see git log` | `TestVerifyFailsTheCommandWhenTheReplayDoesNotReproduceTheArtifact` |

D-02 note: the artifact stores only the trace digest, so the position of the
first differing record cannot be known when the counts agree. The result now
says so in `detail` instead of implying a position; per-record digests in the
artifact would be a format change (decision).
| D-04 | done | `see git log` | `TestOfflineScoreEqualsTheOnlineScoreOfTheSameRun` |

D-04 note: `streamsim score` now needs `--domains-dir`/`--adapters-dir` (as
`replay` does); it builds the same `score.Evidence` the director builds and
calls the same `score.Score`. `score.Offline` remains for callers without a
replayable artifact.
| D-05 | done | `see git log` | `TestEverySafeStopAppliesItsOwnWorldEffect` |
| D-06 | done | `see git log` | `TestWireFaultsAreNotReplayedWithTheEvidenceOfARetry` |

D-05 note: world command ids for safe stops are now `safe-stop/<target>/<n>`
(was `safe-stop/<target>`), visible in the effector-call log.
| D-07 | done | `see git log` | `TestARunThatEndedWithAFailureIsStillClosed` |
| D-08 | done | `see git log` | `TestConcurrentBeginRunOpensTheRunOnce` (-race) |
| D-09 | done | `see git log` | `TestArtifactCountsEveryInjectedFaultEvenOnceCleared` |
| D-10 | done | `see git log` | `TestReplayRebuildsTheWorldIdentityInputsOfTheArtifact` |
| D-15 | done | `see git log` | `TestEveryCommandIsRefusedOnceTheRunIsFinished` |
| D-16 | done | `see git log` | `TestFailedNewLeaksNoFileDescriptor`, `TestEndKeepsTheArtifactWhenItsEvidenceCannotBePublished` |
| D-17 | done | `see git log` | `TestEvidenceReadsDoNotRaceWithAnAdvancingRun` (-race) |

Tier 2 notes: the run-artifact contract gained optional `noiseless` and
`force_failure_mode` in `world_config` (additive; the world digest already
hashed them). `End` now returns the artifact together with a publication
error, and a run whose end failed is finished.
| D-11, D-12, D-13 | done | `see git log` | `TestAnOnsetObservableAtTimeZeroIsObservable`, `TestAnUnknownDetectorFormIsRefusedNotTreatedAsObservable`, `TestAFaultWithNoEffectIsNotObservableEvenWithoutNoise`, `TestPeerResidualSigmaCountsTheEntitiesTheWorldHolds` |
| D-23 | done | `see git log` | `TestAdvanceErrorsKeepTheirOwnCodes` |
| D-24 | done | `see git log` | `TestReportChangesNothingWhenItsVerdictIsRefused` |
| D-22 | done (refuse) | `see git log` | `TestWorldCreateRefusesSeedsThatJSONCannotCarryExactly` |
| D-19, D-41 | done | `see git log` | `TestScriptedEffectorsCarryTheirArgumentsAndDeterministicIds` |

D-22 note: the SDK re-marshals arguments through float64 before the handler,
so exact seeds above 2^53-1 cannot be carried; MCP now refuses them in the
schema instead of rounding. The CLI `--seed` keeps the full uint64 range.

Pin impact: `suite/cold-chain-transit` and its suite file changed (D-12): the
`power_transfer_gap` scenario is no longer claimed observable because its
declared detector channel has zero noise and zero deviation (D-52). The pin
file was regenerated in this commit.
| D-21 | done | `see git log` | `TestSilenceDetectionCitesTheLastDeliveredRecordOfTheSeries` |
| D-27, D-42 | done | `see git log` | `TestCommandIDReusedForADifferentRequestIsRefused`, `TestIdempotentReplayReturnsTheFirstResultIncludingItsEffectETA` |
| D-51 | done | `see git log` | `TestWorldCreateReturnsTheRunIdTruthIsSealedAgainst` |

Not yet started: D-20, D-25, D-26, D-28, D-30..D-34, D-38, D-40, D-50, tier 5
(D-46, D-47, D-48, T-01). The last round-check was green at `92ace18`; the
D-21/D-27/D-42/D-51 commit below ran focused package tests only.
| D-25 | done | `see git log` | `TestReadingNeverAltersWhatARunEmits` |
| D-20 | done | `see git log` | `TestAuditScenarioWithoutAWindowAuditsTheSuiteScenarioLength` |
| D-26 | done | `see git log` | `TestReorderSwapsAboutHalfOfTheWholeSecondPairs`; pinned reorder/combined digests updated |
| D-30 | done | `see git log` | `TestVerifyRefusesAnEmptyFixtureInsteadOfPanicking` |
| D-31 | done | `see git log` | `TestHashSuffixKeepsTheLimitAndStaysHexWhateverTheLimit` |
| D-32 | done | `see git log` | `TestSplitRecordsReadsJSONArrayOutputElementwise`, `TestAdapterVerdictFailsOnlyForADeclaredCheckOrForNothingToCheck` |
| D-33 | done | `see git log` | `TestCompileRefusesMalformedCountAndUniqueKeywords` |
| D-34 | done | `see git log` | `TestF1InputExplicitZeroCoefficientSurvivesARoundTrip` |
| D-38 | done | `see git log` | `TestCompileReportsTheFirstBrokenDefinitionByName`, `TestProfileValidationNamesTheSortedFirstUndeclaredFault` |
| D-40 | done | removed the dead default | (no behaviour) |
| D-28, D-50 | closed, not defects | see DEFERRED | measured / traced |
