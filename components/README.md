# Component library

Reusable equipment, instruments and example compositions, as data ([COMPONENTS_V0_1](../docs/design/COMPONENTS_V0_1.md)).

| Folder | Contract | Holds |
|---|---|---|
| `equipment/` | `component-v0.1` | Equipment: a kind without channels, measurement points, ports, presets and parameters |
| `instruments/` | `instrument-v0.1` | Sensors that turn a measurement point into a channel |
| `compositions/` | `composition-v0.1` | Plants built from the library; `streamsim domain compose` (P7j) writes them as domains |
| `vocabulary.json` | `component-vocabulary-v0.1` | Units and conversions, port types, ISA-5.1 tag letters |

Every id here is in the `streamsim` namespace. A custom library (your own equipment, instruments and vocabulary) uses its own
namespace and lives in a directory outside this repository; no code change is needed to add one.
