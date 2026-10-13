# Canvas technology decision (P8a, finding R0-14)

- **Date**: 2026-10-12
- **Status**: decided by the owner
- **Answers**: [canvas-design.md](canvas-design.md) §11 questions 1 and 2

## Decision

| Concern | Choice |
|---|---|
| Language | TypeScript, strict (QUALITY_BAR §3) |
| UI framework and build | React with Vite |
| Graph editor | React Flow (`@xyflow/react`) for the plant and kind views |
| Bridge language | Go, in this repository, as its own module (`internal/canvasbridge`) behind a `streamsim canvas` command |
| Browser ⇄ bridge protocol | HTTP JSON for commands and file operations; Server-Sent Events for the live journal stream and party status |
| Shared types | TypeScript types generated from `contracts/*.schema.json`; never written by hand |

## Why

- **React Flow** is the most mature node editor: keyboard navigation, ARIA labels, custom nodes,
  minimap and viewport control are built in, which meets WCAG AA without a hand-built editor.
  The factory has well under 200 instances, so a canvas/WebGL renderer buys nothing and costs
  accessibility.
- **TypeScript** catches drift against the contracts at build time; the truth boundary (sealed
  director records never reaching observer views) is enforced by types as well as tests.
- **Go bridge**: the bridge is the security boundary (token, Origin/Host checks, workspace file
  service), so it sits under the existing gates (dependency allow-list, race, fuzz, coverage
  floors). It uses the MCP Go SDK already in `go.mod`, ships in the one `streamsim` binary and
  needs no Node runtime. The dependency allow-list lets it import only the MCP client and the
  journal reader, so it cannot hold simulation logic (canvas-design §6.1).
- **HTTP + Server-Sent Events** fit the traffic: commands go up, the journal streams down.
  Reconnect is built in, and every request carries the token and is Origin-checked. WebSocket
  adds two-way state the canvas does not need.

## Consequences

- The Go build never depends on the canvas build: the bridge serves the built UI from a
  directory given at start (`--ui-dir`), not through `//go:embed`. The Go isolation gate (P8a)
  stays meaningful.
- Node is a build and test dependency of `canvas/` only (`make canvas-check`), not a runtime one.
- Rejected: Svelte/Vue flow libraries (younger, thinner ecosystem), WebGL graph renderers
  (accessibility), plain JavaScript (no contract checking), a TypeScript bridge (a second server
  codebase outside the Go gates, Node at runtime), Electron/Tauri (packaging and attack surface
  without need).
