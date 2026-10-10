# Ubiquitous language — adapter

An output adapter is data: a JSON file that projects native sim-event-v0.1
records into one consumer's wire format through a closed set of transforms.
The binary holds no consumer's field names or framing rules.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Adapter | The declarative projection of native events into a wire format | `model.Adapter` | `*.adapter.json`, output-adapter-v0.1 |
| Engine | The renderer for one run: framing, record template, transforms, id rewrite | `Engine` | – |
| Streaming session | `Begin`, then `RenderStreamRecord` per delivered event, then `End`; the lifecycle the run layer uses | `Begin`, `End` | – |
| Whole-run render | The same output produced in one call over an event list (conformance) | `RenderRun` | – |
| Record template | The field list rendered per event, with an optional when-guard | `model.RecordTemplate` | `record` |
| Preamble / postamble | Framing records before the first and after the last event | `Begin`, `End` | `preamble`, `postamble` |
| Run meta | Run-level bindings the framing reads (run id, versions, world times, seed) | `meta` | `run_meta` |
| Fixture | The committed 12-event native trace every adapter's golden is rendered from | `FixtureEvents` | `testdata/fixture.jsonl` |
| Golden | The committed expected output of an adapter over the fixture | – | `adapters/golden/*.jsonl` |
| Conformance | Rendering the fixture, validating against the declared output schema and byte-comparing the golden | `Verify`, `VerifyResult` | `adapter verify` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `Engine.Meta` | removed | no caller |
| `adapter.Load` inside the rules package | `files.Load` behind the facade | file reading is an edge |
| `separator` parameter (loader) | `DecodeFile` / `LoadBytes` | the multi-line versus single-line error layout is chosen by the entry point, not by callers (the layer still threads it privately) |
