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
| Engineer | F15 | Medium | Accepted with reason: `sim.run.begin` seals truth for every run, so Evaluate from the canvas is sealed; scoring needs a consumer verdict (`sim.score`), which is the consumer's, so Review shows "not scored" |
| Engineer | F16 | Medium | Fixed: the 12-hour claim is now true (bridge cookie `Max-Age` 12 h since P8b); empty, wrong-code and locked states mocked |
| Engineer | F17 | Low | Fixed: legend with glyphs and the 19-to-9 mapping |

Walkthrough targets after round 1 (to be re-timed by the reviewers): W1 drag pump, drag flow_out to the highlighted
`CH-1.flow_in`, drag power_out, rename (4). W2 click "Go to wire" (1). W3 Simulate, Start Explore run (2). W4 select
CT-1, "Inject a fault on CT-1…", choose `fan_failure`, inject (3–4); the effect shows on MB-COOL and MM-1 without
navigation. W5 click Hall 07 in Needs attention, Open hall (2). W6 run row "Open review" (1).
