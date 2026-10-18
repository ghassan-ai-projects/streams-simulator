# Ubiquitous language — component

A *component library* here is a set of directories holding reusable equipment,
instruments and vocabulary as data. Composing a plant from it writes an
ordinary domain, so a world never depends on the library at run time.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Equipment | Reusable process equipment: a kind without channels, its measurement points, ports, presets and parameters | `Equipment` | `*.component.json`, component-v0.1 |
| Instrument | A reusable sensor that turns a measurement point into a channel | `Instrument` | `*.instrument.json`, instrument-v0.1 |
| Vocabulary | Units and conversions, port types and ISA-5.1 tag letters; merged entry by entry across libraries | `Vocabulary` | `vocabulary.json`, component-vocabulary-v0.1 |
| Library | The merged content of the embedded built-in library (loaded first) and the customer directories, each (id, version) once | `Library` | `--components-dir`, `STREAMSIM_COMPONENTS_DIRS` |
| Reference | `namespace/name@major`, the key of one library file | `reference` | `use`, `default_instrument` |
| Composition | A plant built from equipment, sizes, instruments and connections | `Composition`, `Entry`, `Connection` | `*.composition.json`, composition-v0.1 |
| Target | `<section>/<name>/<path>`, the one field a preset or parameter writes | `workspace.write` | `targets`, `cases` |
| Reading | An instrument at a point: the unit conversion, the converted range and the span that scales noise and fault magnitudes | `reading`, `conversion` | `observation_gain`, `observation_offset` |
| Follow | The dynamics entry of a connected consuming state reads the sum of the providing states over a link type | `follow` | `dynamics[].follow` |
| Twin | A runnable domain scaffolded from kind files: each of its kinds `implements` a kind file, so the simulated device presents the interface of the physical one | `ScaffoldTwin`, `Twin` | `streamsim kind twin` |
| Choices | What a twin needs beyond its kind files: per kind one instrument for every metric and one response for every command that is not a safe stop | `twinChoices` | `*.twin.json`, twin-choices-v0.1 |
| Response | One of three fixed templates for a command: `direct` (its arguments assign the metrics they name), `first_order` (as direct, and a read-only metric follows its rest value plus gain times the set metric through a first-order lag) or `none` (acknowledged, changes nothing); a safe stop and a lease come from the kind file | `responseDirect`, `responseFirstOrder`, `responseNone` | `response` |
| Drive | The coupling of `first_order`: a metric the command sets moves a read-only metric of the same or another kind (gain, time constant); across kinds it is a typed link `<driver>_drives_<driven>` | `driveChoice`, `coupling` | `drives[]` |
| Twin requirements | What a kind file must state for a twin to reproduce it (`resolution` and `writable` on a reading, `name`, `args_schema`, `risk_class` and `safe_stop` on a command, a unit on a number reading, name patterns, readable types); reported all at once, each issue with a `keyword` and the `member` | `TwinRequirements`, `Keyword…` | `twin_member_required`, `twin_unit_required`, `twin_name_pattern`, `twin_range`, … |
| Suggestion | The binary's default choices for kind files: an instrument per reading chosen by unit (`none` when nothing is compatible), a `direct` response only where unambiguous | `SuggestChoices`, `Suggest` | `kind twin --suggest` |
| Instruments for a unit | The instruments that measure in a unit or in one convertible to it (the conversion a twin uses), fitting ones first, then exact unit first | `InstrumentsForUnit`, `ListForReading`, `ListFor` | `component list --unit`, `match` |
| Fit | Whether a twin can read a metric (unit, optional `eu_range`) through an instrument, by the one comparison the scaffold uses: `ok`, or a stable `code` (`unit_incompatible`, `range_unstated`, `range_narrower`, `reading_not_number`) with the `member` of the metric at fault; a unit conversion is never a refusal | `Fit`, `fit` | `fit: {ok, code?, member?}` |
| Reading range | The `eu_range` of the metric an instrument is listed for (`--min`, `--max`: both or neither, min below max), compared with the instrument's range after conversion | `readingRange`, `ErrReadingRange` | `component list --min --max` |
| Library error | A file or composition that fails its schema or a library rule; every issue carries a JSON pointer | `model.ValidationError` | CLI `component_invalid`, `composition_invalid` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| plugin | equipment, instrument | a library is data, not code |
