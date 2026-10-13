# P7l evidence — factory utilities world, scale and the last pre-UI corrections

P7l lands on `pre-ui-hardening` in steps, one commit each. Domain-spec v0.2 §21 (several sources), §27 (lint drivers),
§28 (sampled ports in the library) and §29 (KPI truth) hold the rules.

## The factory utilities world

`examples/compositions/factory-utilities.composition.json` composes `examples/factory-utilities.domain.json` from the
built-in library only: weather and a site calendar; two molders, a CNC line and a packaging line; an air handler; lead and
lag compressors on a receiver and a dryer; a chiller, a cooling tower, three chilled-water pumps and a condenser pump; air,
cooling and process sub-meters under a main meter; a transformer and a standby generator. 21 kinds, 82 tags.

| Check | Result |
|---|---|
| Compose reproduces the committed example byte for byte | `TestComposingTheFactoryUtilitiesCompositionReproducesTheExampleByteForByte` |
| Every instance set comes from the built-in library | `TestEveryKindOfTheFactoryComesFromTheBuiltInLibrary` |
| The air and cooling loops close by sampled reads (compressors read the header every 1 s, lines read air, coolant and dew point every 10 s, the chiller its condenser water, the air handler its chilled water) | `TestTheFactoryClosesItsAirAndCoolingLoopsWithSampledReads` |
| A nominal day lints clean (0 errors, 0 warnings) | `TestTheFactoryUtilitiesWorldValidatesLintsRunsVerifiesReplaysAndReports` (binary) |
| Validate, run two hours with two faults and a perturbation, `verify` and `replay` match, the JSON and Markdown reports carry the two faults and the KPI truth | same test |
| A cooling tower fan failure warms the basin (+13.7 °C), the condenser (+12.7 °C), loads the chiller (+32 %) and shows on the cooling and main meters | `TestAUtilityFaultCascadesThroughTheLinksToEveryMeterAndConsumerDownstream` |
| A receiver drain stuck open keeps the lag compressor loaded longer (duty 0.575 → 0.644) and shows on the air and main meters | same test |
| At least one utility fault is not caught by a one-line baseline (header leak and airend wear are hard, fill fouling is trivial) | `TestTheFactoryUtilityFaultsAreNotAllTrivialForAOneLineBaseline` |

Day-long plausibility (seed 1, no fault): header pressure 6.3–7.7 bar with lead/lag load and unload, chilled water
7.0 ± 0.2 °C, main meter peak 958 kW on a 2.5 MW feeder, transformer winding at most 36 °C, dryer dew point −40 °C.

## Scale: the 2,032-tag campus and its speed budget

`testsupport.Campus` repeats every factory instance in 26 halls (`h01/`…`h26/`), sharing only weather and the calendar, and
connects each hall like the factory: the fixture follows the reference world by construction.

| Check | Result |
|---|---|
| The campus has at least 2,000 tags and the factory's kinds | `TestTheCampusFixtureCarriesTwoThousandTagsComposedOnlyFromTheFactory` (2,032 tags) |
| Ten simulated minutes through the run pipeline (world, perturbation, adapter, ledger, journal) at 30× real time or better | `TestTheCampusRunsTenMinutesThroughTheRunPipelineWithinItsSpeedBudget`: 109× measured with KPI tracking (121,210 events in 5.5 s) |

The world was 2.3× faster for it (one simulated campus hour: 14.2 s → 6.1 s): state reads reuse the entity in hand and
the state's cached dynamics and declaration, integrated steps are listed once per entity, and products and curves no
longer allocate. Lookups only: every trace, ledger and state history in the behaviour pin is unchanged.

## Library and engine corrections found by the world

| Change | Proof |
|---|---|
| A `follow` takes several sources, so one header sums every kind connected into it (R7k-02) | `TestAFollowSumsEveryLinkTypeItNamesAcrossKinds` (world, against the reference), `TestAConsumingStateFollowsTheSumOverEveryKindConnectedIntoIt`, `TestOneLinkTypeCarriesOneProvidedState` (compose), validation cases for a repeated link type and a missing source state |
| Compressors load and unload on a sampled header pressure, with lead and lag bands as parameters; the receiver's relief is a safety relief above the bands | Catalog harness over the library; the factory day |
| Chiller, pumps and tower provide their power to a meter; feeder meters start their billing state at zero | Catalog harness; the factory day |
| The lint follows the states a form's conditions read, so a compressor staged on a cycling header is cyclic | `TestAStateStagedByASequencerOnACyclingConditionCyclesAndIsNotExpectedToSettle` |

## Truth

| Change | Proof |
|---|---|
| The oracle worlds run from the scenario's world seed and are advanced to every sampled instant, so births, retirements, lease lapses and responder arrivals happen in them (R7h-01) | `TestTheOracleWorldsReplayTheChurnOfTheScenarioWorldSeed`; fails with either half of the fix removed |
| A pulse fault is judged only at the instants its detector channels record (R7e-05) | `TestAPulseIsObservableOnlyWhereItsDetectorChannelRecordsIt`; without the fix a 60 s periodic channel is labelled observable inside a 0.3 s sag |
| KPI truth per component and the ISA-18.2 alarm KPIs, sealed by every run (§29) | `TestTheKPITruthCountsAlarmActivationsAndActiveTimeOnTheAlarmScanGrid`, `TestTheKPITruthAveragesAKPIStateOverItsSamples`, `TestAlarmSystemKPIsFollowISA182`, `TestARunSealsItsKPITruthIntoTheDirectorJournalOnceBeforeItCloses`, `TestTrackingKPITruthChangesNoEmittedEvent`, `TestKPITruthRendersAsKPIAlarmAndAlarmSystemTables` |

## Findings closed on this branch

R7k-02, R7h-01, R7g-02, R7e-05, R7e-08, R7e-10, R7e-11, R7e-13 (the last four by two delegated branches, reviewed and merged:
report truth only for the run of its journals, torn journal tails, onsets before the clock, Markdown escaping, framing in the
observer journal, journal file digests; barrier, watermark, epoch and window tests, the host fuzz target, the operator
endpoint close race, SimVersion 0.3.1 and a day-long faulted, perturbed behaviour pin).

## Behaviour pin

Re-recorded twice, each time for a stated reason: the observer journal's framing records (R7e-08 d) change journal and
`run.json` lines, and the two journal file digests change `run.json` lines. No trace, ledger or state-history line changed
in either.
