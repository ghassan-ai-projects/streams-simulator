# The canvas

> Status: Implemented; desktop only. Authority: `internal/canvasbridge`, `canvas/`, and the [CLI reference](../reference/cli.md#canvas). Verified by: canvas gates (`make canvas-check`) and end-to-end tests against the real binary; `streamsim canvas` and the file layout checked against the CLI reference. Last verified: 2026-11-03.

The canvas is the visual client of the simulator: open a workspace, see a plant, build one from the component library, run it, disturb it and review what happened. It is one client of the binary, equal to the command line. It computes nothing the binary owns (validation, composition, simulation, truth, scoring, verification) and has no capability the CLI and the director MCP tools lack. A capability that is missing goes into the simulator first.

## Build and start

The built canvas is not embedded in the binary. You need Node 24 or later.

```bash
cd canvas && npm ci --ignore-scripts && npm run build && cd ..
go build -o bin/streamsim ./cmd/streamsim
bin/streamsim canvas --workspace ~/streamsim-workspace --ui-dir canvas/dist
```

- The command serves the canvas and its bridge on a loopback address until interrupted. A non-loopback `--addr` is refused; the default `127.0.0.1:0` picks a free port.
- The address and a one-time **pairing code** are printed on **stderr**; stdout carries no document. Open the address in a browser and enter the code once; the browser then holds a session cookie.
- `--adapters-dir` and `--components-dir` (repeatable) behave as for the other commands.
- Keep the workspace outside the repository when it holds customer data.

## Workspace layout

The workspace folder holds `domains/`, `compositions/`, `scenarios/`, `layouts/`, `kinds/`, `twins/`, `components/`, `settings/` and `runs/`. The canvas creates what it needs. A file changed on disk by a person or a coding agent reloads in place; an unsaved edit that conflicts is shown, never overwritten.

## Main flows

| Flow | What you do | What the binary does |
| --- | --- | --- |
| Try an example | On an empty workspace choose **Try it** beside an example plant, or make an editable copy of one. | Writes the copy's composition, puts the kind files and board mappings the example implements in `kinds/` (`component install-kinds`; a different file already there is never overwritten and stops the copy), then composes it. |
| Open a plant | Pick a domain from the workspace list. Instances appear as nodes grouped by area, typed links as edges; select one to see its kind (states, sensors, actions, faults, KPIs). A plant whose instances share many links (three or more one way between two of them) opens with **Group links** on: one counted wire per pair of instances, and each node lists its measured channels instead of its ports. The same switch is in a run's review. | Loads and describes the domain. |
| Build from the library | Browse equipment, instruments and presets; place and name instances, set parameters, connect ports. | Type-checks connections, composes, validates and lints; each issue is pinned to its element. Saving writes the composition and the composed domain. |
| Run | Start a run in **Explore** or **Evaluate** mode with a seed; play, pause, step, reset or end. Live values, charts and a timeline show observations, commands, faults and perturbations. | Runs the world; the run folder is the same one `run --out` writes. |
| Disturb | Inject or clear faults, environment events and perturbations from a panel that lists only what the domain declares. | Every disturbance lands in the command log; the session can be exported as a scenario file. |
| Review | The run list shows **Verified**, **Not verified** or **Incomplete** from `streamsim verify`. Open a run to play it back with seek and read its report. | `verify`, `report` and `report compare` produce the documents; playback needs no running world. |
| Connected session | In the Run panel choose **Connected session**, give the host and gateway sockets, add real boards and the allowed policy digests, **Check the plan**, then **Start the session**. | Runs `plantwire connect` as a child process, one at a time, paced to the wall clock. See [physical devices](physical-devices.md). |

In Evaluate mode no truth-bearing value reaches the page before the run is sealed. Explore runs are stamped unblinded and say so on every screen. See [truth and scoring](../architecture/truth-and-scoring.md).

## Bridge and agents

The bridge (`internal/canvasbridge`) maps each canvas action to one binary command or director tool and holds no rule; the actions for physical devices are listed in [physical devices: architecture](../architecture/physical-devices.md#bridge-actions). Anything the canvas does can be done with the [CLI](../reference/cli.md) or an agent through [MCP](../reference/mcp.md). Nothing under `internal/` imports `canvas/`.

## Gates

`make canvas-check` runs, from a clean install: format, types, lint, layer rules, the architecture register, duplication, dead code, unit and component tests with coverage floors, build, the emitted-stylesheet check and the end-to-end tests against the built binary. It is part of `make phase-gate`. `make canvas-e2e` runs only the end-to-end tests, on demand. Architecture and module rules are in [`canvas/README.md`](../../canvas/README.md).

## Limits

- Desktop only: the shell needs at least 1280 px; there is no mobile layout.
- The built canvas must be rebuilt after any change under `canvas/`; `streamsim canvas` does not build it.
- One connected session at a time; the bridge stops it when it closes.
- Connected sessions are verified against a scripted stand-in gateway, not against hardware. Read [limitations](../limitations.md) before trusting a bench result.
- The canvas is a loopback tool with a pairing code, not a multi-user server; see the [security model](../architecture/security-model.md).

## Next reads

- [Physical devices guide](physical-devices.md) for connected sessions with real boards
- [Campaigns](campaigns.md) for generating many sessions to run
- [CLI reference: `canvas`](../reference/cli.md#canvas)
- [Replay and debugging](replay-and-debugging.md)
