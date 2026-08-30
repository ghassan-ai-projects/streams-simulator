# Vendored device wire contract

These files are a **copy** of the device wire contract owned by Agentic Stream.
Agentic Stream is the single source of truth; this directory is a vendored copy
the device emulator tests against so the two repos cannot silently diverge.

- Source repo: `agentic-stream`
- Source path: `internal/contractsv1/schemas/v1/device-*.json` and
  `internal/contractsv1/conformance/v1/`
- Source commit: `e0fec2ee0747605d1ef82bb0cd698169eaa7e5c0`

## Do not edit these files here

To change the contract, change it in Agentic Stream, regenerate its fixtures
(`AGENTIC_STREAM_UPDATE_CONFORMANCE=1 go test ./internal/contractsv1/`), then
re-copy them here and update the commit above. The emulator's conformance test
(`internal/device/conformance_test.go`) proves this copy still decodes/rejects
exactly as the contract requires.
