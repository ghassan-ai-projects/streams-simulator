# Canvas mockups — review record

Benchmark: [mockups/BRIEF.md](../mockups/BRIEF.md). Reviewers: design director (craft, ISA-101, Blueprints grammar),
accessibility (WCAG 2.2 AA, keyboard), process and test engineer (task fit, honesty against the binary).

## Round 1

Scores: director 2/5, accessibility 2/5, engineer 3/5. Verdict: changes required (17 + 7 + 8 High).

### Systemic changes

| Change | Answers |
|---|---|
| New token set; all 125 pairs in `tools/pairs.json` pass in both themes (`tools/contrast.mjs`) | A11y F1, F2; director C2-a…e |
| Port palette re-chosen against the reserved colours (CIEDE2000 ≥ 11 between ports, ≥ 13 from red, amber, purple, accent; `tools/deltae.mjs`); signal is neutral dashed, weather neutral dotted; pin shape per family | Director C4-a, C4-c, C5; a11y F9 |
| Neutral node headers with a 3 px category stripe in Run and Review; Build tints the header at 22 %; macro headers neutral | Director C4-c, "calm" |
| Reserved colours enforced: Explore and truth purple, `live` neutral with a green dot, Not verified and Incomplete neutral with an icon, perturbations neutral, fault onset red | Director C4-b; a11y F5 |
| Dimmed wires: 60 % in dark (≥ 3:1), not faded in light (thin instead) | Director C2-a |
| Text floor 11 px everywhere (sample tag 11 px mono, badges 11 px); zoom never scales text, semantic zoom switches tiers | A11y F4; director C1 |
| One node width per tier (232 full, 176 compact), pin rows split by content, nothing truncated but kind labels | Director C7, C8; a11y F6, F10 |
| Real `button`, `role=tree`, `tab`, `slider` elements; 2 px accent focus ring; disabled 50 % with reason; read only dashed with lock; state sheet `states.html` | A11y F3, F8; director C12 |
| Fluid `.app` (100 vw, min 1280); minimap docked as Navigator in the left panel; zoom and legend in the graph header | Director "1920", S2/W4 minimap |
| `tools/layoutcheck.mjs`: no clipped text, no overlapping nodes, nothing under 11 px, at 1440 × 900 and 1920 × 1080 (new C13) | A11y F11; director "Clip" |
| Data rebuilt from the repository: the real composition (20 entries, 23 instances, 59 connections), real port types and `sample_s`, real compose error, real lint output, real alarms (compressor, transformer, boiler only), real report fields | Engineer F2, F3, F12, F14 |

### Findings and dispositions

| Reviewer | Id | Sev | Disposition |
|---|---|---|---|
| A11y | F1, F2 | High | Fixed: token set, `--accent-strong` with white for primary buttons, accent kept for focus and selection |
| A11y | F3 | High | Fixed: real elements, focus ring, focused pin, disabled and read-only specimens (`states.html`) |
| A11y | F4 | High | Fixed: 11 px floor, checked by `layoutcheck.mjs` |
| A11y | F5 | High | Fixed as director C4-b |
| A11y | F6 | High | Fixed: id never wraps (`flex:none`), kind truncates; "P-1 – P-3" with ×3 stack |
| A11y | F7 | High | Fixed: `home-empty.html`, `pairing-wrong.html`, `pairing-locked.html`; Pair disabled until 12 of 12 |
| A11y | F8 | Medium | Fixed: hover, selected, focus for pins and wires in `states.html` |
| A11y | F9 | Medium | Fixed: shape per family, legend naming 9 types and the 19 library types, alarm chips carry icon and word |
| A11y | F10, F11 | Medium | Fixed: 232 px nodes, content-sized pin rows, no clipping (C13) |
| A11y | F12 | Medium | Fixed: Navigator docked, never over the graph |
| A11y | F13 | Medium | Fixed: boundary portals on the hall screen, breadcrumb moved into the graph header |
| A11y | F14 | Medium | Fixed: compile states incl. "Compiled · 1 warning" in amber (`states.html`) |
| A11y | F15 | Medium | Fixed: °C axis units, fixed decimals per channel (COP 2 decimals everywhere) |
| A11y | F16, F17 | Low | Fixed: scrub time 09:45 after every marker; runs have "Open review" and "Open live run" |
| Director | C4-a, C4-b, C4-c | High | Fixed (systemic) |
| Director | C2-a…e | High | Fixed (systemic); measured, not estimated |
| Director | C1, C7, C8 | High | Fixed (systemic) |
| Director | S2/W4 | High | Fixed: MB-COOL and MM-1 on screen with change since the fault; trace chip "f-0 effect: CT-1 → CH-1 → MB-COOL → MM-1". CH-1 is not drawn amber: the binary raises no warning on the chiller, so amber would be invented (C9); the change markers carry the cascade instead |
| Director | S3 | High | Fixed: Navigator minimap; hall inputs fed by the shared Site bus. No substation summary node: the campus has none (each hall has its own MM-1), so it would be invented |
| Director | S4 | High | Fixed: portals for the Site inputs, live state and alarm kept, breadcrumb in the header |
| Director | S6, C12, C5/S1 | High | Fixed |
| Director | C6, Clip, Pins, Data, Review, 1920 | Medium | Fixed (COP 2 decimals; Hall 07 one figure everywhere; review wires solid, markers labelled) |
| Director | C7 tree, Calm | Low | Fixed: "P-1 – P-3 (3)"; frame fill 6 %, header 18 % |
| Engineer | F1 | High | Fixed as director S2/W4 |
| Engineer | F2 | High | Fixed: the CT-1 alarm is gone. Explore shows an injected fault (dashed red, purple truth). The alarm story moved to `live-evaluate.html`: AC-1 `oil_cooler_fouling` ramps the discharge temperature over `alarm_limit_c` 102 °C at 08:05, read from the `alarm_state` annunciator |
| Engineer | F3 | High | Fixed: the real compose error (`/connections/29`, condenser_water against chilled_water) pinned on the wire, "Go to wire", Simulate disabled |
| Engineer | F4 | High | Fixed: speed is "+10 s" steps and "auto ≈ 40×", labelled as the canvas calling `clock.advance` by 10 s every 250 ms; Simulate menu mocked (`builder-simulate.html`) |
| Engineer | F5 | High | Fixed: one timeline (fault 09:40, basin lag after it), °C axis, red onset |
| Engineer | F6 | High | Fixed as director S4 |
| Engineer | F7 | High | Fixed (systemic) |
| Engineer | F8 | High | Fixed: `live-evaluate.html` is an Evaluate run with truth sealed; the campus is labelled Explore and its fault chip counts only faults the user injected |
| Engineer | F9 | Medium | Fixed: true state shows fan_health 0.33 ramping, fan_eff 0.20 = speed × health, fan_power unchanged |
| Engineer | F10 | Medium | Fixed: MB-AIR 106 + MB-COOL 322 + MB-PROC 520 = MM-1 948 kW |
| Engineer | F11 | Medium | Fixed: per-pump rows in the P-1 – P-3 node; in the builder it is one composition entry, as in the file |
| Engineer | F12 | Medium | Fixed: 20 kinds, 23 instances, 20 entries, 59 connections, 548 campus instances, chiller 8 ports, pump 5, no `fan_status` |
| Engineer | F13 | Medium | Fixed: one Hall 07 figure; "Compare" removed (and shown as a "not available yet" specimen) |
| Engineer | F14 | Medium | Fixed: effector calls section, units on KPI truth, mean and min–max only |
| Engineer | F15 | Medium | Accepted with reason (corrected in round 2): the canvas never calls `run.begin`; `world.create` → `clock.advance` → `run.end` writes `sealed/journal.jsonl` in both modes, and Review reads the run folder and `streamsim report`. Scoring needs a consumer verdict (`sim.score`), which is the consumer's, so Review shows "not scored" |
| Engineer | F16 | Medium | Fixed: the 12-hour claim is now true (bridge cookie `Max-Age` 12 h since P8b); empty, wrong-code and locked states mocked |
| Engineer | F17 | Low | Fixed: legend with glyphs and the 19-to-9 mapping |

Walkthrough targets after round 1 (to be re-timed by the reviewers): W1 drag pump, drag flow_out to the highlighted
`CH-1.flow_in`, drag power_out, rename (4). W2 click "Go to wire" (1). W3 Simulate, Start Explore run (2). W4 select
CT-1, "Inject a fault on CT-1…", choose `fan_failure`, inject (3–4); the effect shows on MB-COOL and MM-1 without
navigation. W5 click Hall 07 in Needs attention, Open hall (2). W6 run row "Open review" (1).

## Round 2

Scores: director 3/5, accessibility 3/5, engineer 3/5. Verdict: changes required (director 6 High, accessibility 3 High,
engineer 2 High).

### Systemic changes

| Change | Answers |
|---|---|
| Every value on every screen comes from real runs of the binary: `tools/runs.sh` runs five scenarios (Explore with `fan_failure`, its fault-free twin, Evaluate with `oil_cooler_fouling`, a 18 h review run with report and verify, the 26-hall campus) and `tools/extract.py` writes `data.json`, which the generators read | Engineer R2-1, R2-2, R2-3, R2-6, R2-7, R2-8 |
| Node readouts show observed channels only; true values appear only in Explore, in purple with the eye icon | Engineer R2-1, R2-4 |
| Change markers are Δ against a fault-free twin world with the same seed, which the canvas runs beside the live world with `world.create` and the same `clock.advance` steps; the trace chip says so | Engineer R2-2 |
| Layout gate: tooltips inside the graph are checked, text wrapped in a fixed box is caught, chart text must stay inside its chart, overlays may not cover nodes or frame titles | Director Gate-1; accessibility gate gaps |
| Contrast gate: 229 pairs, now with every port colour on every frame tint (normal and dimmed), dimmed-node text, and text on the soft badge fills | Director Gate-2; accessibility R2-1 |
| Heat is a double line (a 1 px core in the canvas colour), so the two pink families differ in shape as well as colour | Director Gate-3 |
| Back edges are routed out right, along a lane below both nodes and in from the left, instead of under node bodies | Director "Wires" |
| C4 amended: red also marks a validation error that blocks the next step, in Build and Pairing only, always with an icon and words. Reason: an error that blocks compile or pairing needs "look here" exactly as an alarm does, and Blueprints marks compile errors red; Run and Review chrome stay alarm-only | Director C4; accessibility R2-2 |

### Findings and dispositions

| Reviewer | Id | Sev | Disposition |
|---|---|---|---|
| A11y | R2-1 | High | Fixed: dimmed nodes at 80 % (`--dim-node`), pair added |
| A11y | R2-2 | High | Fixed by amending C4 (above) |
| A11y | R2-3 | High | Fixed: hint repositioned inside the graph; gate now checks `.tip` |
| A11y | R2-4 | Medium | Fixed: `aria-disabled` keeps focus, reason inline and referenced by `aria-describedby` ("Fix 1 compile error to simulate") |
| A11y | R2-5 | Medium | Fixed: full-tier pin rows 24 px; the compact tier shows readings only and is not wired (zoom in or use the Details ports table) |
| A11y | R2-6 | Medium | Fixed: F8 / Shift F8, ⌘Z / ⇧⌘Z, Space, Esc cancels a wire, single keys only with the graph focused, Enter opens Inspect in Run and Review |
| A11y | R2-7 | Medium | Fixed: escaped `&lt;name&gt;` |
| A11y | R2-8 | Low | Fixed: ellipsis with a title |
| A11y | R2-9 | Low | Fixed: real inputs for the code, `aria-invalid`, `role=alert`, labelled slider, disabled boxes when locked or used |
| A11y | R2-10 | Low | Fixed: `<main>` landmark, tabs carry `aria-selected` only |
| A11y | R2-11, R2-12 | Low | Fixed: label offset; recovery hint added (generic, the canvas does not reason about types) |
| Director | W4 | High | Fixed: `live-inject.html`, select CT-1, "Inject a fault on CT-1…", pick `fan_failure` (3); the 13 faults come from `catalog.describe`, parameters are edited afterwards |
| Director | C13-a, C13-b | High | Fixed (as accessibility R2-3, R2-8) |
| Director | C2 | High | Fixed: light `--t-env` #737b8b, light `--t-signal` #3f4552 to keep separation, dark `--t-env` #97a0ae for dimmed wires on frame tints; all in the gate |
| Director | C4 | High | Fixed by the amendment |
| Director | Gate-1 | High | Fixed (systemic) |
| Director | Gate-2, Gate-3 | Medium | Fixed (systemic) |
| Director | S3 | Medium | Fixed: one trunk across the top with a stub per column and a tick into each hall; the pill sits 8 px below the Site node |
| Director | Wires | Medium | Fixed: lane routing for back edges |
| Director | C7 | Medium | Fixed: collapsed areas show the bare count (▸ 5) |
| Director | S7 | Medium | Fixed: node hover lifts 1 px with a `--line-3` border; wire hover is 3 px with enlarged ends |
| Director | C12 | Medium | Fixed: selection is always a 2 px ring 2 px out; keyboard focus is the same ring dashed; both together add a soft halo; selected + fault and selected + alarm specimens added |
| Director | 1920 | Medium | Fixed in part: graph content centres on wide viewports (container units). Accepted with reason for the rest: fit-to-view at load is a P9 behaviour; text never scales with zoom (C1), so a wider viewport shows more context rather than larger nodes |
| Director | Copy, Trend | Medium, Low | Fixed |
| Director | Compact | Low | Fixed: no kind label at the compact tier (it is in the tooltip) |
| Director | Nav | Low | Fixed: Hall 07 cell outlined in the accent |
| Director | Legend | Low | Accepted with reason: every pin tooltip names the type and the legend is one click in the graph header; opening it on first use covers the graph the user came to edit |
| Engineer | R2-1 | High | Fixed: Evaluate shows observed channels only; COP and line heat removed |
| Engineer | R2-2 | High | Fixed: real values; Δ against the twin; no main-meter claim beyond the measured Δ |
| Engineer | R2-3 | Medium | Fixed: fan power from the run (it falls); `fan_eff` = speed × health × trim |
| Engineer | R2-4 | Medium | Fixed (systemic) |
| Engineer | R2-5 | Medium | Fixed as director W4 |
| Engineer | R2-6 | Medium | Fixed: times from the runs; review shows alarm, acknowledge, trip and return |
| Engineer | R2-7 | Medium | Fixed: KPI rows from the real report and its names |
| Engineer | R2-8 | Low | Fixed: counts from the runs |
| Engineer | R2-9 | Low | Fixed: no counter (the bridge does not report one); `pairing-used.html` added |
| Engineer | R2-10 | Low | Fixed: w-3 Explore listed |
| Engineer | R2-11 | Low | Fixed: F15 reason corrected above |

## Decision after round 2 (2026-10-13)

The owner asked why the screens were not built on the real workspace files. Agreed: the direction is settled (all three
reviewers accept it), and the remaining findings are about numbers that a static page has to invent. The mockups are
frozen at this commit as the visual spec — `bp.css` tokens, `states.html`, the screen layouts, the keyboard model — and
formal approval moves to the real UI: the same BRIEF, the same three gates run against the live pages, the same three
reviewers in Chrome, on a real workspace that `tools/runs.sh` builds from real runs and the bridge serves. The round 2
fixes above are in the frozen spec; the data findings (engineer R2-1 to R2-8) are closed by construction, because the UI
shows only what the bridge returns.
