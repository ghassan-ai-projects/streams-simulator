from bp import ic, page, write, node_html, frame, wire, navigator, zoom_bar, topbar, doc_tabs, Node, Pin, check_layout
from plant import build, place, links, type_of

UP, DOWN = "▲", "▼"
COUNT = {"Site": 2, "Production hall": 5, "Compressor room": 4, "Chiller plant": 6, "Substation": 6}
MACROS = {
    "Site": ("env", [], [("shift_out", "signal"), ("tariff_out", "signal"), ("dry_bulb_out", "env"), ("wet_bulb_out", "env")], "WX-1, CAL-1"),
    "Production hall": ("prod", [("air_in", "pressure"), ("coolant_in", "temp"), ("dew_point_in", "temp"), ("shift_in", "signal"), ("outdoor_in", "env")], [("heat_out", "heat"), ("air_out", "air"), ("power_out", "power")], "M-1 – M-4, AHU-1"),
    "Compressor room": ("air", [("demand_in", "air")], [("pressure_out", "pressure"), ("dew_point_out", "temp"), ("power_out", "power")], "AC-1, AC-2, R-1, D-1"),
    "Chiller plant": ("cool", [("load_in", "heat"), ("tariff_in", "signal"), ("wet_bulb_in", "env")], [("supply_out", "temp"), ("power_out", "power")], "CH-1, CT-1, P-1 – P-4"),
    "Substation": ("elec", [("loads_in", "power")], [], "MB-AIR … G-1"),
}


def macro(name, readout, x, y, used=()):
    cat, ins, outs, members = MACROS[name]
    n = Node(name, f"▸ {COUNT[name]} instances", cat, [Pin(*p, on=p[0] in used) for p in ins], [Pin(*p, on=p[0] in used) for p in outs], readout, tier="compact")
    n.x, n.y = x, y
    return n


def explore_nodes():
    n = build("compact", ["P-1", "P-4", "CT-1", "CH-1", "MB-AIR", "MB-PROC", "MB-COOL", "MM-1"])
    n["P-1"].id, n["P-1"].kind = "P-1 – P-3", "pump · chw · ×3"
    ro = {"P-1": [("P-1 ●", "55", "m³/h", "", ""), ("P-2 ●", "55", "m³/h", "", ""), ("P-3 ●", "55", "m³/h", "", ""), ("power", "36.0", "kW", "", "")],
          "P-4": [("flow", "211", "m³/h", "", ""), ("power", "16.4", "kW", "", "")],
          "CT-1": [("basin", "33.4", "°C", "", f"{UP}6.4"), ("fan power", "4.0", "kW", "", "")],
          "CH-1": [("cond water", "33.1", "°C", "", f"{UP}6.1"), ("COP", "4.21", "", "", f"{DOWN}1.58"), ("power", "266", "kW", "", f"{UP}73"), ("supply", "7.12", "°C", "", f"{UP}0.12")],
          "MB-AIR": [("power", "106", "kW", "", "")], "MB-PROC": [("power", "520", "kW", "", "")],
          "MB-COOL": [("power", "322", "kW", "", f"{UP}73")], "MM-1": [("power", "948", "kW", "", f"{UP}73")]}
    for k, v in ro.items():
        n[k].readout = v
    n["CT-1"].state = "fault"
    n["CT-1"].foot = f'<span class="badge alarm-soft">{ic("bolt", 11)} fan_failure · injected</span><span class="mono">f-0 · 09:40:00</span>'
    place(n, {"P-4": (220, 48), "P-1": (220, 166), "CT-1": (424, 48), "CH-1": (628, 48), "MB-AIR": (220, 360), "MB-PROC": (220, 468), "MB-COOL": (424, 360), "MM-1": (628, 360)})
    m = {"Site": macro("Site", [("wet bulb", "22.9", "°C", "", ""), ("tariff", "peak", "", "", "")], 16, 16, ("wet_bulb_out", "tariff_out")),
         "Production hall": macro("Production hall", [("heat", "1,120", "kW", "", ""), ("power", "520", "kW", "", "")], 16, 152, ("coolant_in", "heat_out", "power_out")),
         "Compressor room": macro("Compressor room", [("header", "692", "kPa", "", ""), ("power", "106", "kW", "", "")], 16, 288, ("power_out",))}
    check_layout(list(n.values()) + list(m.values()), 580, 872)
    return n, m


TRACE = {("CT-1", "CH-1"), ("CH-1", "MB-COOL"), ("MB-COOL", "MM-1")}
MACRO_LINKS = [("Site", "wet_bulb_out", "CT-1", "wet_bulb_in", "env"), ("Site", "tariff_out", "CH-1", "tariff_in", "signal"),
               ("Production hall", "heat_out", "CH-1", "load_in", "heat"), ("Production hall", "power_out", "MB-PROC", "loads_in", "power"),
               ("Compressor room", "power_out", "MB-AIR", "loads_in", "power")]


def explore_wires(n, m):
    svg = ""
    for a, ap, b, bp in links():
        src = "P-1" if a in ("P-1", "P-2", "P-3") else a
        if src in n and b in n and not (a in ("P-2", "P-3")):
            style = "trace" if (a, b) in TRACE else "dim"
            svg += wire(n[src].out_xy(ap), n[b].in_xy(bp), type_of(n, src, ap), style)
    for a, ap, b, bp, t in MACRO_LINKS:
        svg += wire(m[a].out_xy(ap), n[b].in_xy(bp), t, "dim")
    return svg


def explore_graph():
    n, m = explore_nodes()
    frames = frame("Chiller plant", 204, 8, 660, 314, "cool", "6 instances") + frame("Substation", 204, 330, 660, 246, "elec", "4 of 6 · TX-1, G-1 off-view")
    svg = f'<svg class="wires" width="900" height="600" aria-hidden="true">{explore_wires(n, m)}</svg>'
    body = "".join(node_html(v) for v in m.values()) + "".join(node_html(v, sel="CT-1") for v in n.values())
    return f'<div class="graph run" tabindex="0" aria-label="Live plant graph">{frames}{svg}{body}</div>'


def spark(points, color="var(--tx-2)", w=64, h=20):
    pts = " ".join(f"{i * w / (len(points) - 1):.0f},{h - 2 - p * (h - 4):.0f}" for i, p in enumerate(points))
    return f'<svg width="{w}" height="{h}" aria-hidden="true"><polyline points="{pts}" fill="none" stroke="{color}" stroke-width="1.5"/></svg>'


def kpis(items):
    out = ""
    for label, value, unit, delta, pts, cls in items:
        d = f'<span class="dl">{delta}</span>' if delta else ""
        out += (f'<div class="kpi {cls}"><div class="lr"><span class="l" title="{label}">{label}</span>{d}</div><div class="vr"><span class="vv">{value}<span class="u">{unit}</span></span>'
                f'{spark(pts, "var(--alarm)" if cls else "var(--tx-2)", 56)}</div></div>')
    return f'<div class="kpis" role="region" aria-label="Key figures">{out}</div>'


def explore_kpis():
    rise, fall, flat = [.2, .2, .21, .2, .3, .45, .6, .72, .8, .86], [.9, .9, .89, .9, .8, .62, .48, .38, .32, .28], [.5, .52, .49, .51, .5, .5, .52, .5, .49, .5]
    return kpis([("Site power", "948", "kW", f"{UP}73", rise, ""), ("Header pressure", "692", "kPa", "R-1", flat, ""),
                 ("CHW supply", "7.12", "°C", "sp 7.00", flat, ""), ("Chiller COP", "4.21", "", f"{DOWN}1.58", fall, ""),
                 ("Active alarms", "0", "", "", [0] * 10, "")])


def runbar(mode, clock):
    chip = (f'<span class="badge truth" style="height:24px" title="Explore: true state is shown">{ic("eye", 12)} Explore</span>' if mode == "explore"
            else f'<span class="badge n" style="height:24px" title="Evaluate: truth is sealed until the run ends">{ic("lock", 12)} Evaluate · sealed</span>')
    return (f'<span class="badge live n" style="height:24px"><span class="pulse"></span>live</span>{chip}'
            f'<div class="tb" role="group" aria-label="Clock"><button class="b" aria-pressed="true" aria-label="Pause">{ic("pause", 12)}</button>'
            f'<button class="b" aria-label="Step 10 seconds">{ic("step", 12)} +10 s</button>'
            f'<button class="b" aria-haspopup="menu" title="The canvas calls clock.advance by 10 s every 250 ms">auto ≈ 40× {ic("chev", 10)}</button></div>'
            f'<div class="mono" style="font-size:13px"><span class="muted">Thu 1 Jan</span> {clock}</div>'
            f'<button class="b solid" aria-haspopup="menu">{ic("bolt", 13)} Disturb</button><button class="b solid">{ic("stop", 11)} End run</button>')


def trend():
    w, h, l = 316, 150, 40
    basin = [27.0, 27.0, 27.1, 27.0, 27.2, 27.6, 28.4, 29.3, 30.2, 31.0, 31.7, 32.3, 32.8, 33.1, 33.4]
    wb = [22.6, 22.7, 22.7, 22.8, 22.8, 22.8, 22.9, 22.9, 22.9, 22.9, 22.9, 22.9, 22.9, 22.9, 22.9]
    y = lambda t: h - 22 - (t - 20) / 20 * (h - 40)
    x = lambda i: l + i * (w - l - 10) / 14
    line = lambda d: " ".join(f"{x(i):.0f},{y(t):.0f}" for i, t in enumerate(d))
    grid = "".join(f'<line x1="{l}" x2="{w - 10}" y1="{y(t):.0f}" y2="{y(t):.0f}" stroke="var(--line)"/><text x="{l - 6}" y="{y(t) + 4:.0f}" text-anchor="end" class="ax">{t}</text>' for t in (20, 30, 40))
    xs = "".join(f'<text x="{x(i):.0f}" y="{h - 4}" text-anchor="{a}" class="ax">{t}</text>' for i, t, a in [(0, "09:25", "start"), (7, "09:55", "middle"), (14, "10:25", "end")])
    onset = x(3.6)
    return (f'<svg width="{w}" height="{h}" role="img" aria-label="CT-1 basin temperature, 09:25 to 10:25: 27.0 °C before the fault, 33.4 °C now">'
            f'<style>.ax{{font:11px var(--font-mono);fill:var(--tx-2)}}</style>{grid}{xs}<text x="4" y="12" class="ax">°C</text>'
            f'<line x1="{l}" x2="{w - 10}" y1="{y(27):.0f}" y2="{y(27):.0f}" stroke="var(--tx-3)" stroke-dasharray="2 3"/><text x="{w - 12}" y="{y(27) + 14:.0f}" text-anchor="end" class="ax">pre-fault 27.0</text>'
            f'<line x1="{onset:.0f}" x2="{onset:.0f}" y1="18" y2="{h - 22}" stroke="var(--alarm)" stroke-width="1.5"/><text x="{onset + 4:.0f}" y="28" class="ax" style="fill:var(--alarm)">f-0 09:40</text>'
            f'<polyline points="{line(wb)}" fill="none" stroke="var(--t-env)" stroke-width="1.5" stroke-dasharray="2 4"/><polyline points="{line(basin)}" fill="none" stroke="var(--t-temp)" stroke-width="2.2"/></svg>')


def inspector_head(nid, kind, badge):
    return (f'<div style="padding:12px;display:flex;gap:10px;align-items:center;border-bottom:1px solid var(--line)"><div style="width:36px;height:36px;border-radius:8px;background:var(--panel-3);box-shadow:inset 0 3px 0 var(--c-cool);display:grid;place-items:center">{ic("plant", 16)}</div>'
            f'<div style="min-width:0"><div style="font-weight:600;font-size:14px;display:flex;gap:8px;align-items:center"><span class="mono">{nid}</span>{badge}</div><div class="sub">{kind}</div></div></div>')


FAULT_BADGE = f'<span class="badge alarm-soft">{ic("bolt", 11)} fault injected</span>'


def explore_right():
    rows = [("basin_temp", "33.4", "°C", f"{UP}6.4", "TT · good"), ("fan_power", "4.0", "kW", "=", "JT · good")]
    r = "".join(f'<tr><td class="mono">{n}</td><td class="num">{v}<span class="u">{u}</span></td><td class="mono" style="text-align:right">{d}</td><td style="text-align:right">{q}</td></tr>' for n, v, u, d, q in rows)
    truth = [("fan_health", "0.33", "ramping to 0.10 at 10:34"), ("fan_eff", "0.20", "fan_speed × fan_health"), ("fan_speed", "0.60", "commanded"), ("approach", "10.5 K", "4.1 K before the fault")]
    t = "".join(f'<tr><td class="mono">{n}</td><td class="num" style="color:var(--truth)">{v}</td><td style="white-space:normal">{d}</td></tr>' for n, v, d in truth)
    return (f'<aside class="panel" style="border-left:1px solid var(--line)"><div class="ph"><div class="tabs" role="tablist"><span class="tab" role="tab" aria-selected="true">Inspect</span><span class="tab" role="tab" aria-selected="false">Watch · 4</span></div></div>'
            f'{inspector_head("CT-1", "Cooling tower · Chiller plant", FAULT_BADGE)}'
            f'<div class="sect"><div class="sh">Basin temperature <span class="r">last 60 min</span></div><div style="padding:0 12px">{trend()}</div>'
            f'<div style="padding:2px 12px 0;display:flex;gap:14px;color:var(--tx-2);font-size:11px"><span><i style="display:inline-block;width:14px;border-top:2px solid var(--t-temp);vertical-align:middle"></i> basin_temp</span><span><i style="display:inline-block;width:14px;border-top:2px dotted var(--t-env);vertical-align:middle"></i> wet bulb (WX-1)</span></div></div>'
            f'<div class="sect"><div class="sh">Readings <span class="r">observed</span></div><table class="r">{r}</table></div>'
            f'<div class="sect" style="background:var(--truth-soft)"><div class="sh" style="color:var(--truth)">{ic("eye", 12)} True state <span class="r">Explore only</span></div><table class="r">{t}</table></div>'
            f'<div class="sect"><div class="sh">Disturbances on CT-1</div><div style="margin:0 12px;padding:8px 10px;border-radius:8px;border:1px solid var(--alarm);display:flex;align-items:center;gap:10px">'
            f'<span class="ico-alarm">{ic("bolt", 14)}</span><div><div style="font-weight:600">fan_failure</div><div class="sub mono" style="font-size:11px">f-0 · ramp 1/h · 09:40:00</div></div><button class="b solid" style="margin-left:auto">Clear</button></div>'
            f'<div style="padding:8px 12px"><button class="b solid" style="width:100%;justify-content:center">{ic("bolt", 13)} Inject a fault on CT-1…</button></div></div></aside>')


def lane(name, inner, h=28):
    return (f'<div style="display:grid;grid-template-columns:120px 1fr;border-bottom:1px solid var(--line);height:{h}px;align-items:center">'
            f'<span style="padding:0 12px;color:var(--tx-2)">{name}</span><div style="position:relative;height:100%;margin-right:16px">{inner}</div></div>')


def timeline(lanes, axis, now_label):
    bars = "".join(f'<div style="position:absolute;left:{i * 1.17:.1f}%;bottom:4px;width:.7%;height:{30 + (i * 7) % 50}%;background:var(--line-3);border-radius:1px"></div>' for i, _ in enumerate(range(85)))
    ticks = "".join(f'<span class="mono" style="position:absolute;left:{p}%;transform:translateX({"-50%" if 0 < p < 100 else ("0" if p == 0 else "-100%")});font-size:11px;color:var(--tx-2)">{t}</span>' for p, t in axis)
    rows = "".join(lane(n, i) for n, i in lanes) + lane("Observations", bars, 24)
    return (f'<section class="panel" style="border-top:1px solid var(--line)" aria-label="Timeline"><div class="ph"><div class="tabs" role="tablist"><span class="tab" role="tab" aria-selected="true">Timeline</span><span class="tab" role="tab" aria-selected="false">Journal</span></div><span class="r">{now_label}</span></div>'
            f'<div style="position:relative">{rows}<div style="display:grid;grid-template-columns:120px 1fr"><span></span><div style="position:relative;height:22px;margin:4px 16px 0 0">{ticks}</div></div>'
            f'<div style="position:absolute;right:16px;top:0;bottom:26px;width:2px;background:var(--accent)" title="now"></div></div></section>')


def explore_timeline():
    dist = (f'<div style="position:absolute;left:25%;right:0;top:4px;height:20px;border-radius:4px;border:1px solid var(--alarm);background:var(--alarm-soft);color:var(--tx);font-size:11px;padding:0 8px;line-height:18px;white-space:nowrap;overflow:hidden">'
            f'{ic("bolt", 11)} f-0 fan_failure · CT-1 · 09:40:00 →</div>')
    return timeline([("Disturbances", dist), ("Alarms", '<span class="muted" style="line-height:28px">no alarms in this run</span>')],
                     [(0, "09:25"), (25, "09:40"), (50, "09:55"), (75, "10:10"), (100, "10:25")], "following live")


def explore_left():
    watch = [("CT-1 basin_temp", "33.4 °C", f"{UP}6.4"), ("CH-1 cond_temp", "38.9 °C", f"{UP}5.9"), ("CH-1 power", "266 kW", f"{UP}73"), ("MM-1 power", "948 kW", f"{UP}73")]
    w = "".join(f'<div class="li"><span class="mono" style="color:var(--tx)">{n}</span><span class="mono" style="margin-left:auto;color:var(--tx)">{v}</span><span class="delta" style="width:40px;justify-content:flex-end">{d}</span></div>' for n, v, d in watch)
    events = [("10:25:00", "clock", "auto-step running · 10 s steps"), ("09:40:00", "bolt", "you injected fan_failure on CT-1 (f-0)"), ("09:00:00", "run", "run w-3 opened · Explore · seed 7")]
    e = "".join(f'<div class="li" style="align-items:flex-start;padding-top:6px;padding-bottom:6px;white-space:normal"><span class="mono muted" style="flex:none">{t}</span><span>{m}</span></div>' for t, i, m in events)
    return (f'<aside class="panel" style="border-right:1px solid var(--line)"><div class="ph">Alarms<span class="r">none active</span></div>'
            f'<div class="li" style="color:var(--tx-2)">{ic("check", 13)} No alarm is active in this run</div>'
            f'<div class="ph" style="border-top:1px solid var(--line)">Watch<span class="r">change since 09:40</span></div><div class="list">{w}</div>'
            f'<div class="ph" style="border-top:1px solid var(--line)">Events</div><div class="list">{e}</div>'
            f'<div class="ph" style="border-top:1px solid var(--line)">Run</div><div class="kv"><div class="k2">world · seed · profile</div><div class="v2">w-3 · 7 · nominal</div></div><div class="kv"><div class="k2">observations delivered</div><div class="v2">12,480 of 12,480</div></div>'
            f'{navigator([(10, 10, 50, 60, "color-mix(in srgb,var(--c-env) 45%,transparent)"), (66, 10, 120, 56, "color-mix(in srgb,var(--c-cool) 45%,transparent)"), (66, 72, 120, 46, "color-mix(in srgb,var(--c-elec) 45%,transparent)")], (6, 6, 186, 118), "all in view")}</aside>')


def graph_bar(crumbs, right):
    return (f'<div style="display:flex;align-items:center;gap:8px;height:40px;padding:0 12px;border-bottom:1px solid var(--line);background:var(--panel);white-space:nowrap">'
            f'<nav aria-label="Breadcrumb" style="display:flex;align-items:center;gap:6px;color:var(--tx-2)">{crumbs}</nav><span style="flex:1"></span>{right}</div>')


def screen(left, kp, bar, g, tl, right, docs, clock, mode, theme):
    center = f'<div style="display:grid;grid-template-rows:auto auto minmax(0,1fr) 136px;min-width:0;min-height:0;overflow:hidden">{kp}{bar}{g}{tl}</div>'
    body = topbar("Run", runbar(mode, clock), workspace=False, compact=True) + doc_tabs(2, docs) + f'<div class="main" style="grid-template-columns:232px minmax(0,1fr) 336px">{left}{center}{right}</div>'
    return page("Live run", body, theme)


DOCS = [("factory-utilities", "composition", "blocks"), ("factory-utilities", "plant", "plant"), ("w-3", "live run", "run")]
TRACE_BAR = (f'<span class="badge o">{ic("bolt", 11)} f-0 effect: CT-1 → CH-1 → MB-COOL → MM-1</span><button class="b" aria-label="Clear trace">{ic("close", 11)}</button>')


def explore(theme="dark"):
    crumbs = f'{ic("plant", 13)}<span style="color:var(--tx);font-weight:600">Factory</span><span class="muted">· 3 areas collapsed</span>'
    return screen(explore_left(), explore_kpis(), graph_bar(crumbs, TRACE_BAR + zoom_bar(80)), explore_graph(), explore_timeline(), explore_right(), DOCS, "10:25:00", "explore", theme)




def eval_nodes():
    n = build("compact", ["AC-1", "AC-2", "R-1", "D-1", "MB-AIR", "MM-1"])
    ro = {"AC-1": [("discharge", "105.7", "°C", "hot", ""), ("power", "76.2", "kW", "", "")],
          "AC-2": [("discharge", "86.0", "°C", "", ""), ("power", "22.4", "kW", "", "")],
          "R-1": [("pressure", "692", "kPa", "", ""), ("flow", "1,153", "m³/h", "", "")],
          "D-1": [("dew point", "−40.0", "°C", "", ""), ("power", "7.1", "kW", "", "")],
          "MB-AIR": [("power", "106", "kW", "", "")], "MM-1": [("power", "948", "kW", "", "")]}
    for k, v in ro.items():
        n[k].readout = v
    n["AC-1"].state = "alarm"
    n["AC-1"].foot = f'<span class="badge alarm">{ic("bell", 11)} discharge high · unack</span>'
    place(n, {"AC-1": (220, 48), "AC-2": (220, 236), "R-1": (424, 48), "D-1": (424, 236), "MB-AIR": (668, 48), "MM-1": (668, 196)})
    m = {"Production hall": macro("Production hall", [("heat", "1,094", "kW", "", ""), ("power", "516", "kW", "", "")], 16, 48, ("air_in", "dew_point_in", "air_out", "power_out")),
         "Chiller plant": macro("Chiller plant", [("supply", "7.04", "°C", "", ""), ("power", "326", "kW", "", "")], 16, 236, ("power_out",))}
    check_layout(list(n.values()) + list(m.values()), 580, 872)
    return n, m


def eval_graph():
    n, m = eval_nodes()
    svg = ""
    for a, ap, b, bp in links():
        if a in n and b in n:
            svg += wire(n[a].out_xy(ap), n[b].in_xy(bp), type_of(n, a, ap), "trace" if "AC-1" in (a, b) else "")
    svg += wire(m["Production hall"].out_xy("air_out"), n["R-1"].in_xy("demand_in"), "air", "dim")
    svg += wire(n["R-1"].out_xy("pressure_out"), m["Production hall"].in_xy("air_in"), "pressure", "dim")
    svg += wire(n["D-1"].out_xy("dew_point_out"), m["Production hall"].in_xy("dew_point_in"), "temp", "dim")
    svg += wire(m["Chiller plant"].out_xy("power_out"), n["MM-1"].in_xy("loads_in"), "power", "dim")
    svg += wire(m["Production hall"].out_xy("power_out"), n["MM-1"].in_xy("loads_in"), "power", "dim")
    frames = frame("Compressor room", 204, 8, 412, 360, "air", "4 instances") + frame("Substation", 652, 8, 212, 320, "elec", "2 of 6")
    body = "".join(node_html(v) for v in m.values()) + "".join(node_html(v, sel="AC-1") for v in n.values())
    return f'<div class="graph run" tabindex="0" aria-label="Live plant graph">{frames}<svg class="wires" width="900" height="600" aria-hidden="true">{svg}</svg>{body}</div>'


def eval_kpis():
    flat = [.5, .52, .49, .51, .5, .5, .52, .5, .49, .5]
    return kpis([("Site power", "948", "kW", "", flat, ""), ("Header pressure", "692", "kPa", "R-1", flat, ""),
                 ("CHW supply", "7.04", "°C", "sp 7.00", flat, ""), ("Chiller COP", "5.83", "", "", flat, ""),
                 ("Active alarms", "1", "", "AC-1", [0, 0, 0, 0, 0, 1, 1, 1, 1, 1], "alarm")])


def discharge_trend():
    w, h, l = 316, 150, 40
    temp = [89.0 + max(0.0, (i * 180 / 14 - 20) / 60 * 6.25) for i in range(15)]
    y = lambda t: h - 22 - (t - 80) / 40 * (h - 40)
    x = lambda i: l + i * (w - l - 10) / 14
    line = " ".join(f"{x(i):.0f},{y(t):.0f}" for i, t in enumerate(temp))
    grid = "".join(f'<line x1="{l}" x2="{w - 10}" y1="{y(t):.0f}" y2="{y(t):.0f}" stroke="var(--line)"/><text x="{l - 6}" y="{y(t) + 4:.0f}" text-anchor="end" class="ax">{t}</text>' for t in (80, 100, 120))
    xs = "".join(f'<text x="{x(i):.0f}" y="{h - 4}" text-anchor="{a}" class="ax">{t}</text>' for i, t, a in [(0, "05:40", "start"), (7, "07:10", "middle"), (14, "08:40", "end")])
    return (f'<svg width="{w}" height="{h}" role="img" aria-label="AC-1 discharge temperature, 05:40 to 08:40: 89.0 °C rising to 105.7 °C; alarm limit 102 °C crossed at 08:05">'
            f'<style>.ax{{font:11px var(--font-mono);fill:var(--tx-2)}}</style>{grid}{xs}<text x="4" y="12" class="ax">°C</text>'
            f'<line x1="{l}" x2="{w - 10}" y1="{y(102):.0f}" y2="{y(102):.0f}" stroke="var(--alarm)" stroke-dasharray="5 3"/><text x="{l + 4}" y="{y(102) - 5:.0f}" class="ax" style="fill:var(--alarm)">limit 102</text>'
            f'<line x1="{x(1.56):.0f}" x2="{x(1.56):.0f}" y1="18" y2="{h - 22}" stroke="var(--tx-2)" stroke-dasharray="3 3"/><text x="{x(1.56) + 4:.0f}" y="28" class="ax">sent 06:00</text>'
            f'<line x1="{x(11.28):.0f}" x2="{x(11.28):.0f}" y1="18" y2="{h - 22}" stroke="var(--alarm)" stroke-width="1.5"/><text x="{x(11.28) - 4:.0f}" y="{h - 28}" text-anchor="end" class="ax" style="fill:var(--alarm)">alarm 08:05</text>'
            f'<polyline points="{line}" fill="none" stroke="var(--t-temp)" stroke-width="2.2"/></svg>')


ALARM_BADGE = f'<span class="badge alarm">{ic("bell", 11)} alarm</span>'


def eval_right():
    rows = [("discharge_temp", "105.7", "°C", "TT"), ("alarm_state", "1", "unack", "YA"), ("run_status", "1", "running", "YS"), ("power", "76.2", "kW", "JT")]
    r = "".join(f'<tr><td class="mono">{n}</td><td class="num">{v}<span class="u">{u}</span></td><td style="text-align:right"><span class="badge o">{q}</span></td></tr>' for n, v, u, q in rows)
    return (f'<aside class="panel" style="border-left:1px solid var(--line)"><div class="ph"><div class="tabs" role="tablist"><span class="tab" role="tab" aria-selected="true">Inspect</span><span class="tab" role="tab" aria-selected="false">Watch · 3</span></div></div>'
            f'{inspector_head("AC-1", "Air compressor · lead · Compressor room", ALARM_BADGE)}'
            f'<div class="sect"><div class="sh">Discharge temperature <span class="r">05:40–08:40</span></div><div style="padding:0 12px">{discharge_trend()}</div></div>'
            f'<div class="sect"><div class="sh">Readings <span class="r">observed</span></div><table class="r">{r}</table></div>'
            f'<div class="sect"><div class="sh">{ic("lock", 12)} True state</div><div style="padding:0 12px;color:var(--tx-2)">Sealed in Evaluate. It appears in Review when the run ends.</div></div>'
            f'<div class="sect"><div class="sh">Commands you sent</div><table class="r"><tr><td class="mono">06:00:00</td><td>fault.inject · AC-1 · oil_cooler_fouling</td></tr></table>'
            f'<div style="padding:8px 12px;color:var(--tx-2);font-size:11px;white-space:normal">Acknowledge and reset are AC-1 effectors. The plant host calls them, not the canvas.</div></div></aside>')


def eval_left():
    a = (f'<div class="li" style="align-items:flex-start;padding-top:8px;padding-bottom:8px;white-space:normal;background:var(--alarm-soft);box-shadow:inset 3px 0 0 var(--alarm)">'
         f'<span class="ico-alarm" style="margin-top:2px">{ic("bell", 13)}</span><div><div><b class="mono">AC-1</b> discharge temperature high</div><div class="sub mono" style="font-size:11px">08:05:12 · unacknowledged · 35 min</div></div></div>')
    events = [("08:05:12", "AC-1 alarm raised (alarm_state 1)"), ("06:00:00", "you sent fault.inject on AC-1"), ("06:00:00", "run w-4 opened · Evaluate · seed 11")]
    e = "".join(f'<div class="li" style="align-items:flex-start;padding-top:6px;padding-bottom:6px;white-space:normal"><span class="mono muted" style="flex:none">{t}</span><span>{m}</span></div>' for t, m in events)
    return (f'<aside class="panel" style="border-right:1px solid var(--line)"><div class="ph">Alarms<span class="badge alarm" style="margin-left:auto">1 active</span></div><div class="list">{a}</div>'
            f'<div class="ph" style="border-top:1px solid var(--line)">Events</div><div class="list">{e}</div>'
            f'<div class="ph" style="border-top:1px solid var(--line)">Run</div><div class="kv"><div class="k2">world · seed · profile</div><div class="v2">w-4 · 11 · nominal</div></div>'
            f'{navigator([(10, 10, 60, 50, "color-mix(in srgb,var(--c-prod) 45%,transparent)"), (76, 10, 110, 70, "color-mix(in srgb,var(--c-air) 45%,transparent)"), (10, 66, 60, 50, "color-mix(in srgb,var(--c-cool) 45%,transparent)"), (130, 86, 56, 34, "color-mix(in srgb,var(--c-elec) 45%,transparent)")], (6, 6, 186, 118), "all in view")}</aside>')


def eval_timeline():
    cmd = f'<div style="position:absolute;left:11%;top:4px;height:20px;border-radius:4px;border:1px solid var(--line-3);color:var(--tx);font-size:11px;padding:0 8px;line-height:18px;white-space:nowrap">{ic("bolt", 11)} fault.inject AC-1 · sent 06:00</div>'
    al = f'<div style="position:absolute;left:80.6%;right:0;top:4px;height:20px;border-radius:4px;background:var(--alarm);color:var(--on-alarm);font-size:11px;font-weight:600;padding:0 8px;line-height:20px;white-space:nowrap">{ic("bell", 11)} AC-1 08:05 →</div>'
    return timeline([("Commands sent", cmd), ("Alarms", al)], [(0, "05:40"), (25, "06:25"), (50, "07:10"), (75, "07:55"), (100, "08:40")], "following live")


def evaluate(theme="dark"):
    crumbs = f'{ic("plant", 13)}<span style="color:var(--tx);font-weight:600">Factory</span><span class="muted">· 2 areas collapsed</span>'
    bar = graph_bar(crumbs, f'<span class="badge o">{ic("bell", 11)} following AC-1 alarm</span>' + zoom_bar(80))
    docs = [("factory-utilities", "plant", "plant"), ("w-3", "live run", "run"), ("w-4", "live run · Evaluate", "run")]
    return screen(eval_left(), eval_kpis(), bar, eval_graph(), eval_timeline(), eval_right(), docs, "08:40:00", "evaluate", theme)


if __name__ == "__main__":
    write("live-run.html", explore())
    write("live-evaluate.html", evaluate())
