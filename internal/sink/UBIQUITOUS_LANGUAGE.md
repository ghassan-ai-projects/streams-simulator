# Ubiquitous language — sink

Evidence leaves the simulator on a sink; the MCP surface carries control and
actuation, never the stream.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Sink | Receives rendered lines in order and returns the full byte stream at close, which the trace digest is computed from | `Sink` | `sink` |
| Inproc | The sink that buffers everything in memory | `Inproc` | `inproc` |
| File sink | A buffered, byte-reproducible sink backed by one file; flushed at command boundaries, synced at End | `File`, `NewFile` | `file`, `trace.jsonl` |
| HTTP push | A sink that posts each line to an endpoint in stepped, in-order delivery | `HTTPPush`, `NewHTTPPush` | `http-push` |
| Line | One rendered record as the adapter emitted it | `Write(line)` | – |
| Stream | The concatenation of every line, newline-terminated | `Close` result | trace bytes |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| wall sub-mode of http-push | – | never existed in code; the doc promised a simdet-tagged wall mode that is not there |
