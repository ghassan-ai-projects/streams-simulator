# Install and build

> Status: Current setup guide. Authority: `go.mod`, Makefile, CI workflow, and repository layout. Verified by: build/help smoke checks. Last verified: 2026-11-03.

Streams Simulator is a Go command-line application. The safest first setup is a clean checkout followed by the same build and test commands used by contributors.

## Prerequisites

- Go `1.25.12` or a compatible toolchain able to satisfy [`go.mod`](../../go.mod).
- Git.
- The repository’s declared Go modules. `go mod download` may need network access on a fresh machine.
- `golangci-lint`, `deadcode`, and `govulncheck` for the full `make ci-check` gate.
- Node 24 and npm, only to build the visual canvas or run `make canvas-check` (see [`canvas/README.md`](../../canvas/README.md)). The release checks intentionally fail closed when required tools are absent.

## Build

```bash
git clone https://github.com/ghassan-ai-projects/streams-simulator.git
cd streams-simulator
go run ./cmd/streamsim help
make build
```

The Makefile writes the binary under `bin/` using the repository directory name. `go run ./cmd/streamsim ...` avoids relying on that generated filename while experimenting.

## Verify the checkout

Start with the narrow checks:

```bash
go test ./...
go vet ./...
git diff --check
```

Then run the project gate when the required tools are installed:

```bash
make ci-check
```

The gate includes the function-length and gofmt checks, the documentation smoke check, module maintenance, build, vet, lint, the gate ratchet, one short shuffled test run with the wall clock removed and coverage floors, dead-code analysis, vulnerability analysis, and bounded fuzzing. For daily work use `make quick` (about 30 s with the test cache). `make phase-gate` adds the canvas gates, the architecture check and the behaviour pin; `make test-race`, `make test-long`, soak, performance and release-manifest checks are separate targets run on demand; see [release evidence](../benchmark/evidence.md).

## Repository inputs

- Domains are loaded from [`domains/`](../../domains/).
- Adapters are loaded from [`adapters/`](../../adapters/).
- Machine-readable contracts are maintained in [`contracts/`](../../contracts/) and embedded in [`internal/schemas/`](../../internal/schemas/).
- The component library and example plants are built in; `--components-dir` adds customer directories.
- The visual canvas is built separately: `npm ci --ignore-scripts && npm run build` in `canvas/`, then `streamsim canvas --workspace <dir> --ui-dir canvas/dist` (see the [CLI reference](../reference/cli.md)).
- Run artifacts are written to the output directory selected by the command or MCP configuration.

## Current checkout note

All shipped domains in [`domains/`](../../domains/) are committed and validated by the schema suite. New domain files must be added together with their test, inventory, and release-manifest updates; see [limitations](../limitations.md) for the current status boundary.

## Next reads

- [Quickstart](quickstart.md)
- [CLI reference](../reference/cli.md)
- [Contributor quality guide](../governance/quality.md)
