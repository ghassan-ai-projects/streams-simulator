// Package cli implements the streamsim command-line interface: catalog,
// domain/adapter tooling, scripted runs, replay/verify, the MCP servers,
// suite generation, offline scoring, the release manifest and the device
// emulator. The entrypoint in cmd/streamsim is a thin call to Main.
//
// The package is a facade: the commands live in its private app layer, and
// the process, file-system and long-running-server edges below it.
package cli
