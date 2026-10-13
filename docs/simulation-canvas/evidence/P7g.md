# P7g evidence — component library contracts

- **Date**: 2026-10-13
- **Change**: [COMPONENTS_V0_1](../../design/COMPONENTS_V0_1.md); schemas `component-v0.1`, `instrument-v0.1`, `composition-v0.1`,
  `component-vocabulary-v0.1`; `components/` with the vocabulary, two equipment components (chiller in three presets, AHU), three
  instruments (RTD, magnetic flow meter, power meter) and the pilot composition; domain-spec v0.2 §20 (`area`, `provenance` on
  instance sets, structure only); [`docs/examples/chiller-pilot.domain.json`](../../examples/chiller-pilot.domain.json), the domain
  the pilot composes to, produced by applying the documented rules.

## Exit evidence

| Check | Result |
|---|---|
| Every library file validates against its contract | `TestEveryLibraryFileFollowsItsContract` |
| Built-in ids stay in the `streamsim` namespace | `TestBuiltInLibraryFilesUseTheStreamsimNamespace` |
| The hand-composed pilot (two chiller sizes, two instrument sets, one AHU link, qualified names, bias states for instrument faults, provenance) passes the existing domain validator and runs deterministically | `streamsim domain validate`: valid, 8 channels, 20 faults; `TestTheHandComposedPilotValidatesAndRunsDeterministically` |
| Values are plausible per size | 2 h run from steady state: supply 6.5–7.2 °C at setpoints 6.5 and 7, condenser 33 °C, power proportional to size |
| Library consistency | `TestTheVocabularyIsConsistent`, `TestEveryMeasurementPointHasAMatchingDefaultInstrument`, `TestThePilotRecordsTheDigestsOfTheFilesItCameFrom`, `TestComposedInstrumentFaultsKeepTheChannelNoise` |

## Findings while composing by hand

- A preset must carry every size-dependent coefficient (the load-to-temperature gains), not only sizes: the first large preset
  drifted to 8.7 °C. Recorded in COMPONENTS §2 (presets are complete variants); `curve` and `product` (P7i) remove the need.
- Kind names prefix fault and effector names, which allow only `[a-z0-9_]`: a component name's hyphens become underscores.
- A fault whose detector names a removed measurement point is a compose error.
- Initial values that differ from steady state give a settling transient; the nominal-day test (P7k) will catch large ones.
