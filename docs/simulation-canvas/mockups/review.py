from bp import connect, ic, page, write, node_html, frame, wire, navigator, zoom_bar, topbar, doc_tabs, check_layout
from liverun import explore_nodes, links, type_of, MACRO_LINKS, graph_bar, UP, DOWN

MARKS = [(17.4, "alarm", "bell", "AC-1 alarm 08:05", 0, "l"), (27.8, "n", "bolt", "drop 5 % 09:20", 1, "l"), (30.6, "alarm", "bolt", "f-1 CT-1 09:40", 0, "r")]


def review_graph():
    n, m = explore_nodes()
    values = {"CT-1": [("basin", "27.6", "°C", "", f"{UP}0.6"), ("fan power", "4.0", "kW", "", "")],
              "CH-1": [("cond water", "27.3", "°C", "", f"{UP}0.3"), ("COP", "5.71", "", "", f"{DOWN}0.08"), ("power", "195", "kW", "", f"{UP}2"), ("supply", "7.01", "°C", "", "")],
              "MB-COOL": [("power", "251", "kW", "", f"{UP}2")], "MM-1": [("power", "877", "kW", "", f"{UP}2")]}
    for k, v in values.items():
        n[k].readout = v
    n["CT-1"].foot = (f'<span class="badge alarm-soft">{ic("bolt", 11)} fan_failure · f-1</span>'
                      f'<span class="badge truth">{ic("eye", 11)} true fan_health 0.92</span>')
    svg = ""
    for a, ap, b, bp in links():
        src = "P-1" if a in ("P-1", "P-2", "P-3") else a
        if src in n and b in n and a not in ("P-2", "P-3"):
            svg += connect(n[src], ap, n[b], bp, type_of(n, src, ap))
    for a, ap, b, bp, t in MACRO_LINKS:
        svg += connect(m[a], ap, n[b], bp, t)
    frames = frame("Chiller plant", 204, 8, 660, 314, "cool", "6 instances") + frame("Substation", 204, 330, 660, 246, "elec", "4 of 6")
    lock = f'<div class="badge o lock fixed" style="position:absolute;right:12px;bottom:12px;background:var(--panel);height:28px">{ic("lock", 12)} Read only · finished run</div>'
    body = "".join(node_html(v) for v in m.values()) + "".join(node_html(v, sel="CH-1") for v in n.values())
    return f'<div class="graph run" tabindex="0" aria-label="Plant at 09:45:00, read only">{frames}<svg class="wires" width="900" height="600" aria-hidden="true">{svg}</svg>{body}{lock}</div>'


def scrubber():
    marks = "".join(f'<div style="position:absolute;left:{p}%;top:0;height:{24 + row * 24}px;width:2px;background:var({"--alarm" if c == "alarm" else "--line-3"})"></div>' for p, c, i, t, row, side in MARKS)
    labels = "".join(f'<span class="badge {"alarm-soft" if c == "alarm" else "o"}" style="position:absolute;left:calc({p}% {"- 1px" if side == "r" else "+ 1px"});{"transform:translateX(-100%);" if side == "l" else ""}top:{26 + row * 24}px;background:var(--panel);border-color:{"var(--alarm)" if c == "alarm" else "var(--line-3)"}">{ic(ic_, 11)} {t}</span>'
                     for p, c, ic_, t, row, side in MARKS)
    ticks = "".join(f'<span class="mono" style="position:absolute;left:{p}%;top:2px;font-size:11px;color:var(--tx-2);transform:translateX({"0" if p == 0 else ("-100%" if p == 100 else "-50%")})">{t}</span>' for p, t in [(0, "06:00"), (25, "09:00"), (50, "12:00"), (75, "15:00"), (100, "18:00")])
    return (f'<div style="background:var(--panel);border-bottom:1px solid var(--line);padding:8px 16px 4px"><div style="display:flex;align-items:center;gap:12px">'
            f'<div class="tb" role="group" aria-label="Playback"><button class="b" aria-label="Step back">{ic("step", 12)}</button>'
            f'<button class="b" aria-pressed="false" aria-label="Play">{ic("play", 12)}</button><button class="b mono">10× {ic("chev", 10)}</button></div>'
            f'<div class="mono" style="font-size:14px;white-space:nowrap">09:45:00 <span class="muted" style="font-size:12px">of 06:00 – 18:00</span></div><span style="flex:1"></span>'
            f'<span class="sub">Jump to</span><button class="b solid">{ic("bell", 12)} AC-1 alarm</button><button class="b solid">{ic("bolt", 12)} f-1</button></div>'
            f'<div style="position:relative;height:84px;margin-top:8px" role="slider" aria-label="Playback time" aria-valuetext="09:45:00">'
            f'<div style="position:absolute;left:0;right:0;top:20px;height:16px">'
            f'<div style="position:absolute;left:0;right:0;top:6px;height:4px;border-radius:2px;background:var(--line-3)"></div>'
            f'<div style="position:absolute;left:0;width:31.25%;top:6px;height:4px;border-radius:2px;background:var(--accent)"></div>{marks}'
            f'<div style="position:absolute;left:calc(31.25% - 8px);top:0;width:16px;height:16px;border-radius:50%;background:var(--tx);box-shadow:0 0 0 3px var(--accent-soft)"></div>{labels}</div>{ticks}</div></div>')


def runs_left():
    items = [("w-4", "factory-utilities", "live", "n", "Evaluate · running", False), ("w-2", "factory-utilities", "Verified", "ok", "Evaluate · 12 h · ended", True),
             ("w-1", "factory-utilities", "Not verified", "n", "files differ from the run", False), ("p-4", "chiller-pilot", "Incomplete", "n", "stopped at 01:12", False)]
    icons = {"Verified": ic("check", 11), "Not verified": ic("x", 11), "Incomplete": ic("alert", 11), "live": '<span class="pulse"></span>'}
    r = "".join(f'<div class="li{" on" if on else ""}" style="align-items:flex-start;padding-top:8px;padding-bottom:8px"><div style="flex:1;min-width:0"><div style="display:flex;gap:8px;align-items:center">'
                f'<b class="mono">{i}</b><span class="badge {c}{" live" if l == "live" else ""}" style="margin-left:auto">{icons[l]} {l}</span></div>'
                f'<div class="sub" style="font-size:11px;margin-top:2px">{d}</div><div class="sub" style="font-size:11px">{meta}</div></div></div>' for i, d, l, c, meta, on in items)
    return (f'<aside class="panel" style="border-right:1px solid var(--line)"><div style="padding:8px 12px"><label class="input">{ic("search", 13)}<span class="muted">Filter runs</span></label></div>'
            f'<div class="ph">Runs<span class="r">4</span></div><div class="list" role="listbox" aria-label="Runs">{r}</div>'
            f'{navigator([(10, 10, 50, 60, "color-mix(in srgb,var(--c-env) 45%,transparent)"), (66, 10, 120, 56, "color-mix(in srgb,var(--c-cool) 45%,transparent)"), (66, 72, 120, 46, "color-mix(in srgb,var(--c-elec) 45%,transparent)")], (6, 6, 186, 118), "all in view")}</aside>')


def table(rows, cols):
    return '<table class="r">' + "".join("<tr>" + "".join(f'<td class="{c}">{v}</td>' for v, c in zip(row, cols)) + "</tr>" for row in rows) + "</table>"


def report():
    checks = table([[f'<span class="ico-ok">{ic("check", 12)}</span>', n, "match"] for n in ("trace", "journal", "sealed_journal", "ledger")], ["", "mono", "num"])
    dist = table([["06:00:00", "scenario · oil_cooler_fouling · AC-1"], ["09:20:00", "you · perturb drop 5 % · until 09:30"], ["09:40:00", "you · fault fan_failure · CT-1 · f-1"]], ["mono", ""])
    eff = table([["08:07:30", "AC-1 · acknowledge_reset", "applied"], ["08:10:05", "AC-2 · set_loaded", "applied"]], ["mono", "mono", "num"])
    kpi = table([["CH-1 COP", "5.31", "4.02–6.12"], ["CT-1 approach K", "6.8", "3.9–10.9"], ["AC-1 kWh/m³", "0.108", "0.104–0.113"], ["MM-1 power kW", "903", "861–951"]],
                ["", "num", "num"])
    alarms = table([["activations", "1"], ["peak per 10 min", "1"], ["flood intervals", "0"], ["chattering", "0"], ["stale", "0"], ["standing at end", "1 · AC-1"]], ["", "num"])
    sect = lambda title, right, body: f'<div class="sect"><div class="sh">{title}<span class="r">{right}</span></div>{body}</div>'
    head = (f'<div style="padding:12px;border-bottom:1px solid var(--line)"><div style="display:flex;align-items:center;gap:8px"><span style="font-weight:600;font-size:14px" class="mono">w-2 · r-2</span>'
            f'<span class="badge ok">{ic("check", 11)} Verified</span></div><div class="sub">factory-utilities 0.1.0 · seed 7 · native-jsonl · sim 0.3.1</div>'
            f'<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:8px;margin-top:10px">'
            + "".join(f'<div class="kpi"><div class="l">{l}</div><span class="vv" style="font-size:15px">{v}</span></div>' for l, v in [("emitted", "57,908"), ("delivered", "57,783"), ("dropped", "125")])
            + '</div></div>')
    return (f'<aside class="panel scroll" style="border-left:1px solid var(--line);overflow:auto"><div class="ph"><div class="tabs" role="tablist"><span class="tab" role="tab" aria-selected="true">Report</span><span class="tab" role="tab" aria-selected="false">Deliveries</span><span class="tab" role="tab" aria-selected="false">Journal</span></div></div>'
            f'{head}{sect("Verification", "files match", checks)}{sect("Disturbances · truth", "revealed at end", dist)}{sect("Effector calls", "by the consumer", eff)}'
            f'{sect("KPI truth", "mean · min–max", kpi)}{sect("Alarm system", "ISA-18.2", alarms)}'
            f'{sect("Scoring", "not scored", NO_VERDICT)}</aside>')


NO_VERDICT = '<div class="sub" style="padding:0 12px">No consumer verdict was submitted for this run.</div>'


def topbar_right():
    return (f'<span class="badge ok" style="height:24px">{ic("check", 11)} Verified · files match</span><span class="badge n" style="height:24px">Evaluate · ended</span>'
            f'<button class="b solid">{ic("file")} Export scenario</button><button class="b solid">Markdown report</button>')


def screen(theme="dark"):
    crumbs = (f'{ic("plant", 13)}<span style="color:var(--tx);font-weight:600">Factory</span><span class="muted">· at 09:45:00</span>'
              f'<div class="seg" role="group" style="margin-left:12px"><button aria-pressed="false" style="padding:0 10px">Reported</button><button aria-pressed="true" style="padding:0 10px">True vs reported</button></div>')
    center = f'<div style="display:grid;grid-template-rows:auto auto minmax(0,1fr);min-width:0;min-height:0;overflow:hidden">{scrubber()}{graph_bar(crumbs, zoom_bar(80))}{review_graph()}</div>'
    docs = [("factory-utilities", "plant", "plant"), ("w-2", "review", "check")]
    body = topbar("Review", topbar_right(), workspace=False, compact=True) + doc_tabs(1, docs) + f'<div class="main" style="grid-template-columns:200px minmax(0,1fr) 368px">{runs_left()}<main style="display:contents">{center}</main>{report()}</div>'
    return page("Review", body, theme)


write("review.html", screen())
