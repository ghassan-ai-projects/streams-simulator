from bp import ic, page, write, node_html, frame, wire, navigator, zoom_bar, topbar, doc_tabs, Node, Pin, check_layout, curve
from liverun import kpis, runbar, graph_bar, macro, eval_nodes, eval_right, links, type_of, ALARM_BADGE
from plant import build, place

STATE = {7: "alarm", 19: "fault"}
COLS, X0, Y0, DX, DY = 6, 216, 24, 104, 108


def power(n):
    return 860 + (n * 37) % 90


def hall(n):
    st = STATE.get(n, "")
    badge = {"alarm": f'<span class="badge alarm">{ic("bell", 11)} 1</span>', "fault": f'<span class="badge alarm-soft">{ic("bolt", 11)} 1</span>'}.get(st, "")
    node = Node(f"Hall {n:02d}", "▸", "env", [Pin("signals", "signal")], [], [("power", f"{power(n)}", "kW", "", "")], st, tier="compact")
    node.foot = badge
    node.x, node.y = X0 + ((n - 1) % COLS) * DX * 1.0, Y0 + ((n - 1) // COLS) * DY
    return node


def campus_graph():
    halls = [hall(n) for n in range(1, 27)]
    for h in halls:
        h.tier = "hall"
    site = Node("Site", "▸ 2 instances", "env", [], [Pin("signals ×4", "signal")], [("wet bulb", "22.9", "°C", "", ""), ("shift", "day", "", "", "")], tier="compact")
    site.x, site.y = 16, 220
    sx, sy = site.out_xy("signals ×4")
    bus_x = X0 - 20
    svg = f'<path d="M{sx:.0f},{sy:.0f} L{bus_x},{sy:.0f}" stroke="var(--t-signal)" stroke-width="4" stroke-dasharray="7 5" fill="none"/>'
    top, bottom = halls[0].y - 12, halls[-1].y - 12
    svg += f'<path d="M{bus_x},{top} L{bus_x},{bottom}" stroke="var(--t-signal)" stroke-width="4" stroke-dasharray="7 5" fill="none"/>'
    for row in range(5):
        top_y = Y0 + row * DY - 12
        last = X0 + (min(COLS, 26 - row * COLS) - 1) * DX
        svg += f'<path d="M{bus_x},{top_y} L{last - 4},{top_y}" stroke="var(--t-signal)" stroke-width="2" stroke-dasharray="7 5" fill="none"/>'
        for col in range(min(COLS, 26 - row * COLS)):
            x = X0 + col * DX - 4
            svg += f'<path d="M{x},{top_y} L{x},{top_y + 46} L{x + 3},{top_y + 46}" stroke="var(--t-signal)" stroke-width="1.5" fill="none"/>'
    label = f'<div class="badge o" style="position:absolute;left:{bus_x - 150}px;top:{sy + 12:.0f}px;background:var(--panel)">×26 · shift, tariff, dry bulb, wet bulb</div>'
    body = node_html(site) + "".join(hall_html(h) for h in halls)
    return f'<div class="graph run" tabindex="0" aria-label="Campus graph, 26 halls">{svg_wrap(svg)}{body}{label}</div>'


def svg_wrap(svg):
    return f'<svg class="wires" width="900" height="600" aria-hidden="true">{svg}</svg>'


def hall_html(h):
    cls = f"node compact {h.state}" + (" sel" if h.id == "Hall 07" else "")
    foot = f'<div class="nf" style="padding:4px 10px">{h.foot}</div>' if h.foot else ""
    return (f'<div class="{cls}" role="button" aria-label="{h.id}, {h.readout[0][1]} kW{", 1 alarm" if h.state == "alarm" else ""}" style="left:{h.x:.0f}px;top:{h.y:.0f}px;--nw:96px;--cat:var(--c-env)">'
            f'<span class="pd on dia" style="left:-6px;top:34px;color:var(--t-signal)"></span>'
            f'<div class="nh" style="padding:0 8px"><span class="id">{h.id[5:]}</span><span class="k">hall</span></div>'
            f'<div style="padding:4px 8px 6px" class="ro1"><span class="mono" style="font:600 14px var(--font-mono)">{h.readout[0][1]}</span><span class="u">kW</span></div>{foot}</div>')


def campus_kpis():
    total = sum(power(n) for n in range(1, 27))
    flat = [.5, .52, .49, .51, .5, .5, .52, .5, .49, .5]
    return kpis([("Campus power", f"{total / 1000:.2f}", "MW", "", flat, ""), ("Halls normal", "24", "of 26", "", [1] * 10, ""),
                 ("Active alarms", "1", "", "Hall 07", [0, 0, 0, 0, 0, 0, 1, 1, 1, 1], "alarm"), ("Faults injected", "1", "", "Hall 19", [0, 0, 0, 1, 1, 1, 1, 1, 1, 1], ""),
                 ("Observations", "1.21", "M", "48 k/min", flat, "")])


def campus_left():
    att = [("alarm", "bell", "Hall 07 · h07/AC-1", "discharge temperature high", "08:05:12 · unacknowledged"),
           ("fault", "bolt", "Hall 19 · h19/CT-1", "fan_failure injected by you", "07:20:00 · f-3")]
    a = "".join(f'<div class="li{" on" if i == 0 else ""}" style="align-items:flex-start;padding-top:8px;padding-bottom:8px;white-space:normal"><span class="ico-alarm" style="margin-top:2px">{ic(icn, 13)}</span>'
                f'<div><div><b class="mono">{h}</b></div><div>{m}</div><div class="sub mono" style="font-size:11px">{t}</div></div></div>' for i, (s, icn, h, m, t) in enumerate(att))
    filters = ('<div style="display:flex;gap:6px;padding:8px 12px;flex-wrap:wrap"><button class="badge n" aria-pressed="true">All 26</button>'
               f'<button class="badge alarm-soft" aria-pressed="false">{ic("bell", 11)} Alarms 1</button><button class="badge alarm-soft" aria-pressed="false">{ic("bolt", 11)} Faults 1</button></div>')
    order = [7, 19] + [n for n in range(1, 27) if n not in (7, 19)]
    rows = "".join(f'<div class="li{" on" if n == 7 else ""}"><span class="mono" style="color:var(--tx)">Hall {n:02d}</span>{state_icon(n)}<span class="mono" style="margin-left:auto">{power(n)}<span class="u">kW</span></span></div>' for n in order[:7])
    blocks = [(8 + (i % 6) * 31, 8 + (i // 6) * 22, 26, 16, "var(--alarm)" if i == 6 else ("transparent;border:2px dashed var(--alarm)" if i == 18 else "color-mix(in srgb,var(--c-env) 55%,transparent)")) for i in range(26)]
    return (f'<aside class="panel" style="border-right:1px solid var(--line)"><div style="padding:8px 12px"><label class="input">{ic("search", 13)}<span class="muted">Filter halls, instances, tags</span></label></div>{filters}'
            f'<div class="ph" style="border-top:1px solid var(--line)">Needs attention<span class="r">2</span></div><div class="list">{a}</div>'
            f'<div class="ph" style="border-top:1px solid var(--line)">Halls<span class="r">attention first</span></div><div class="list">{rows}<div class="li muted">19 more · all normal</div></div>'
            f'{navigator(blocks, (4, 4, 192, 116), "all 26 in view")}</aside>')


def state_icon(n):
    return {"alarm": f'<span class="ico-alarm">{ic("bell", 12)}</span>', "fault": f'<span class="ico-alarm">{ic("bolt", 12)}</span>'}.get(STATE.get(n, ""), "")


def campus_right():
    rows = [("power", f"{power(7)}", "kW"), ("header pressure", "692", "kPa"), ("CHW supply", "7.04", "°C"), ("chiller COP", "5.83", "")]
    r = "".join(f'<tr><td>{n}</td><td class="num">{v}<span class="u">{u}</span></td></tr>' for n, v, u in rows)
    inside = [("Compressor room", "4", '<span class="ico-alarm">' + ic("bell", 12) + '</span> 1 alarm'), ("Chiller plant", "6", ""), ("Production hall", "5", ""), ("Substation", "6", "")]
    i = "".join(f'<tr><td>{n}</td><td class="num">{c}</td><td style="text-align:right">{s}</td></tr>' for n, c, s in inside)
    return (f'<aside class="panel" style="border-left:1px solid var(--line)"><div class="ph"><div class="tabs" role="tablist"><span class="tab" role="tab" aria-selected="true">Selection</span></div></div>'
            f'<div style="padding:12px;border-bottom:1px solid var(--line)"><div style="font-weight:600;font-size:14px;display:flex;gap:8px;align-items:center"><span class="mono">Hall 07</span>{ALARM_BADGE}</div>'
            f'<div class="sub">h07/ · 21 instances · 78 tags</div><div style="display:flex;gap:8px;margin-top:12px"><button class="b primary">{ic("chev", 11)} Open hall <span class="kbd" style="color:var(--on-accent);border-color:var(--on-accent)">↵</span></button></div></div>'
            f'<div class="sect"><div class="sh">Hall figures <span class="r">observed</span></div><table class="r">{r}</table></div>'
            f'<div class="sect"><div class="sh">Inside</div><table class="r">{i}</table></div>'
            f'<div class="sect"><div class="sh">Grouping</div><div style="padding:0 12px;color:var(--tx-2);white-space:normal">Halls are grouped by the instance prefix h01/ … h26/. The grouping is a canvas view; the domain has 548 instances.</div></div></aside>')


def campus(theme="dark"):
    crumbs = f'{ic("plant", 13)}<span style="color:var(--tx);font-weight:600">Campus</span><span class="muted">· 26 halls grouped · 548 instances · 2,032 tags</span>'
    center = (f'<div style="display:grid;grid-template-rows:auto auto minmax(0,1fr);min-width:0;min-height:0;overflow:hidden">{campus_kpis()}'
              f'{graph_bar(crumbs, zoom_bar(100))}{campus_graph()}</div>')
    docs = [("factory-utilities-campus-26", "plant", "plant"), ("c-1", "live run", "run")]
    body = topbar("Run", runbar("explore", "09:12:30"), workspace=False, compact=True) + doc_tabs(1, docs) + f'<div class="main" style="grid-template-columns:256px minmax(0,1fr) 320px">{campus_left()}{center}{campus_right()}</div>'
    return page("Campus overview", body, theme)


def hall_inside():
    n = build("compact", ["AC-1", "AC-2", "R-1", "D-1"])
    src, _ = eval_nodes()
    for k in n:
        n[k].readout, n[k].state, n[k].foot = src[k].readout, src[k].state, src[k].foot
        n[k].outs = [p for p in n[k].outs if p.name != "power_out"] + [p for p in n[k].outs if p.name == "power_out"]
    place(n, {"AC-1": (400, 48), "AC-2": (400, 236), "R-1": (604, 48), "D-1": (604, 236)})
    m = {"Production hall": macro("Production hall", [("heat", "1,094", "kW", "", ""), ("power", "516", "kW", "", "")], 196, 40, ("air_in", "dew_point_in", "shift_in", "outdoor_in", "air_out")),
         "Chiller plant": macro("Chiller plant", [("supply", "7.04", "°C", "", ""), ("power", "326", "kW", "", "")], 196, 268, ("tariff_in", "wet_bulb_in")),
         "Substation": macro("Substation", [("MM-1", "939", "kW", "", ""), ("TX-1 load", "62", "%", "", "")], 604, 424, ("loads_in",))}
    check_layout(list(n.values()) + list(m.values()), 700, 864)
    svg = ""
    for a, ap, b, bp in links():
        if a in n and b in n:
            svg += wire(n[a].out_xy(ap), n[b].in_xy(bp), type_of(n, a, ap), "trace" if "AC-1" in (a, b) else "")
    svg += wire(m["Production hall"].out_xy("air_out"), n["R-1"].in_xy("demand_in"), "air", "dim")
    svg += wire(n["R-1"].out_xy("pressure_out"), m["Production hall"].in_xy("air_in"), "pressure", "dim")
    svg += wire(n["D-1"].out_xy("dew_point_out"), m["Production hall"].in_xy("dew_point_in"), "temp", "dim")
    for k in ("AC-1", "AC-2", "D-1"):
        svg += wire(n[k].out_xy("power_out"), m["Substation"].in_xy("loads_in"), "power", "dim")
    portals = ""
    targets = [("shift_out", "signal", "Production hall", "shift_in"), ("dry_bulb_out", "env", "Production hall", "outdoor_in"),
               ("tariff_out", "signal", "Chiller plant", "tariff_in"), ("wet_bulb_out", "env", "Chiller plant", "wet_bulb_in")]
    for i, (port, typ, dest, dport) in enumerate(targets):
        y = 120 + i * 44 + (100 if i > 1 else 0)
        portals += f'<div class="portal" style="left:12px;top:{y - 14}px;width:156px;color:var(--t-{typ})"><span class="lbl">Site · {port}</span></div>'
        svg += wire((168, y), m[dest].in_xy(dport), typ, "dim")
    frames = frame("Compressor room", 384, 8, 412, 360, "air", "h07/ · 4 instances")
    note = '<div class="sh" style="position:absolute;left:4px;top:72px">Boundary · from Site</div>'
    body = "".join(node_html(v) for v in m.values()) + "".join(node_html(v, sel="AC-1") for v in n.values())
    body = body.replace('<span class="id">AC-', '<span class="id">h07/AC-').replace('<span class="id">R-1', '<span class="id">h07/R-1').replace('<span class="id">D-1', '<span class="id">h07/D-1')
    return f'<div class="graph run" tabindex="0" aria-label="Hall 07 graph">{frames}<svg class="wires" width="900" height="700" aria-hidden="true">{svg}</svg>{body}{portals}{note}</div>'


TRUTH_H07 = (f'<div class="sect" style="background:var(--truth-soft)"><div class="sh" style="color:var(--truth)">{ic("eye", 12)} True state <span class="r">Explore only</span></div>'
             '<table class="r"><tr><td class="mono">active fault</td><td class="num" style="color:var(--truth)">oil_cooler_fouling</td></tr>'
             '<tr><td>source</td><td style="text-align:right">scenario, profile nominal</td></tr><tr><td>onset</td><td class="num" style="color:var(--truth)">06:00:00 · ramp 0.25/h</td></tr>'
             '<tr><td>discharge offset</td><td class="num" style="color:var(--truth)">+16.7 °C</td></tr></table></div>'
             '<div class="sect"><div class="sh">Commands you sent</div><div style="padding:0 12px;color:var(--tx-2)">None on h07/AC-1.</div></div>')


def macro_screen(theme="dark"):
    crumbs = (f'<button class="b" style="padding:0 6px">{ic("plant", 13)} Campus</button><span class="muted">/</span>'
              f'<span style="color:var(--tx);font-weight:600" aria-current="location">Hall 07</span><span class="badge alarm">{ic("bell", 11)} 1 alarm</span>'
              f'<button class="b solid" style="margin-left:8px">↑ Back to campus <span class="kbd">Esc</span></button>')
    center = (f'<div style="display:grid;grid-template-rows:auto auto minmax(0,1fr);min-width:0;min-height:0;overflow:hidden">{campus_kpis()}'
              f'{graph_bar(crumbs, zoom_bar(80))}{hall_inside()}</div>')
    right = eval_right().replace('<span class="mono">AC-1</span>', '<span class="mono">h07/AC-1</span>').replace(" · Compressor room", " · Hall 07")
    sealed = right[right.index('<div class="sect"><div class="sh"><svg width="12" height="12" aria-hidden="true"><use href="#i-lock"/>'):right.index('</aside>')]
    right = right.replace(sealed, TRUTH_H07)
    docs = [("factory-utilities-campus-26", "plant", "plant"), ("c-1", "live run", "run")]
    body = topbar("Run", runbar("explore", "09:12:30"), workspace=False, compact=True) + doc_tabs(1, docs) + f'<div class="main" style="grid-template-columns:256px minmax(0,1fr) 336px">{campus_left()}{center}{right}</div>'
    return page("Hall 07", body, theme)


write("campus.html", campus())
write("macro.html", macro_screen())
