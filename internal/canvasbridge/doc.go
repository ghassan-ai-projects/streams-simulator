// Package canvasbridge is the canvas bridge: the local HTTP server behind
// `streamsim canvas` through which the visual canvas reaches the simulator.
// It pairs one browser, serves the built UI and the workspace files the canvas
// may use, relays each registered canvas action to the director (over MCP) or
// to a streamsim command, and streams a run's observer journal. It holds no
// simulation, validation or truth logic of its own.
package canvasbridge
