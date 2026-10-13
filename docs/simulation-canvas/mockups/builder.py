from bp import ic, page, write, node_html, frame, wire, legend, minimap, navigator, zoom_bar, topbar, doc_tabs, Node, Pin
from plant import build, place, links, type_of

DOCS = [("factory-utilities", "composition", "blocks"), ("factory-utilities", "plant", "plant"), ("w-3", "live run", "run")]
LAYOUT = {"P-1": (36, 60), "P-4": (36, 236), "CT-1": (316, 236), "CH-1": (600, 60),
          "MB-AIR": (36, 444), "MB-PROC": (36, 524), "MB-COOL": (316, 444), "MM-1": (600, 444)}
SHOWN = ["P-1", "P-4", "CT-1", "CH-1", "MB-AIR", "MB-PROC", "MB-COOL", "MM-1"]
VIA = {("CH-1", "load_in"): "← M-1 – M-4, AHU-1", ("CH-1", "tariff_in"): "← CAL-1", ("CT-1", "wet_bulb_in"): "← WX-1",
       ("CH-1", "supply_out"): "→ M-1 – M-4, AHU-1", ("MB-AIR", "loads_in"): "← AC-1, AC-2, D-1", ("MB-PROC", "loads_in"): "← M-1 – M-4",
       ("MM-1", "power_out"): "→ TX-1, G-1"}


def nodes(error=False, connect=False):
    n = build("full", SHOWN)
    n["P-1"].id, n["P-1"].kind = "P-1 – P-3", "pump · chw · ×3"
    for (nid, port), via in VIA.items():
        for p in n[nid].ins + n[nid].outs:
            if p.name == port:
                p.via = via
    if error:
        n["P-4"].outs[1].err = True
        n["CH-1"].ins[1].err = True
    if connect:
        for node in n.values():
            for p in node.ins:
                p.state = "compat" if (node.id, p.name) == ("CH-1", "flow_in") else ("incompat" if p.typ != "water" else "")
    if connect:
        new = build("full", ["P-4"])["P-4"]
        new.id, new.kind = "P-5", "pump · chw · new"
        for p in new.outs:
            p.on = False
        new.outs[0].state = "focus"
        n["P-5"] = new
    place(n, {**LAYOUT, "P-5": (316, 52)})
    return n


def graph_wires(n, error=False):
    svg = ""
    for a, ap, b, bp in links():
        src = "P-1" if a in ("P-1", "P-2", "P-3") else a
        if src in n and b in n and (a, b) not in (("P-2", "CH-1"), ("P-3", "CH-1"), ("P-2", "MB-COOL"), ("P-3", "MB-COOL")):
            style = "sel" if "CT-1" in (a, b) else ""
            svg += wire(n[src].out_xy(ap), n[b].in_xy(bp), type_of(n, src, ap), style)
    if error:
        svg += wire(n["P-4"].out_xy("cw_flow_out"), n["CH-1"].in_xy("flow_in"), "cw", "err")
    return svg


def stack(n):
    return "".join(f'<div class="node full" style="left:{n.x + 6 * k}px;top:{n.y + 6 * k}px;height:{n.height()}px;--nw:{n.w}px;opacity:.6" aria-hidden="true"></div>' for k in (2, 1))


def graph(error=False, connect=False, extra=""):
    n = nodes(error, connect)
    frames = frame("Chiller plant", 16, 16, 832, 380, "cool", "4 entries · 6 instances") + frame("Substation", 16, 404, 832, 196, "elec", "4 of 6 shown · TX-1, G-1 off-screen")
    svg = f'<svg class="wires" width="900" height="700" aria-hidden="true">{graph_wires(n, error)}{extra_wires(n, connect)}</svg>'
    body = stack(n["P-1"]) + "".join(node_html(v, sel="P-5" if connect else "CT-1") for v in n.values())
    flags = ""
    if error:
        flags = (f'<div class="badge alarm" style="position:absolute;left:330px;top:166px">{ic("x", 12)} /connections/29</div>')
    return f'<div class="graph build" tabindex="0" aria-label="Composition graph">{frames}{svg}{body}{flags}{extra}</div>'


def nav():
    return navigator([(10, 10, 54, 40, "color-mix(in srgb,var(--c-air) 45%,transparent)"), (70, 10, 60, 46, "color-mix(in srgb,var(--c-cool) 45%,transparent)"),
                  (10, 56, 54, 44, "color-mix(in srgb,var(--c-prod) 45%,transparent)"), (70, 62, 60, 40, "color-mix(in srgb,var(--c-elec) 45%,transparent)"),
                  (136, 10, 54, 30, "color-mix(in srgb,var(--c-env) 45%,transparent)")], (66, 6, 70, 100), "2 of 5 areas in view")


def extra_wires(n, connect):
    if not connect:
        return ""
    from bp import curve
    return f'<path d="{curve(n["P-5"].out_xy("flow_out"), n["CH-1"].in_xy("flow_in"))}" stroke="var(--t-water)" stroke-width="3" fill="none" stroke-dasharray="2 5" stroke-linecap="round"/>'


def graph_bar(crumbs, right=""):
    return (f'<div style="display:flex;align-items:center;gap:8px;height:40px;padding:0 12px;border-bottom:1px solid var(--line);background:var(--panel);white-space:nowrap">'
            f'<nav aria-label="Breadcrumb" style="display:flex;align-items:center;gap:6px;color:var(--tx-2)">{crumbs}</nav><span class="spacer" style="flex:1"></span>{right}</div>')


def crumbs_factory():
    return (f'{ic("plant", 13)}<button class="b" style="padding:0 4px">Factory</button><span class="muted">/</span>'
            f'<button class="b" style="padding:0 4px;color:var(--tx);font-weight:600" aria-current="location">Chiller plant + Substation</button>')


def left():
    tree = [("Factory", 0, "23"), ("Site", 1, "2"), ("Production hall", 1, "5"), ("Compressor room", 1, "4"), ("Chiller plant", 1, "6"),
            ("CH-1", 2, ""), ("CT-1", 2, "sel"), ("P-1 – P-3 (3)", 2, ""), ("P-4", 2, ""), ("Substation", 1, "6")]
    t = "".join(f'<div role="treeitem" aria-level="{d + 1}" aria-selected="{str(c == "sel").lower()}" style="padding-left:{12 + d * 14}px">'
                f'{ic("chev", 11) if d < 2 else ""}<span class="{"mono" if d == 2 else ""}">{n}</span><span class="muted" style="margin-left:auto">{"" if c == "sel" else c}</span></div>' for n, d, c in tree)
    lib = [("Cooling", "cool", [("chiller", 8), ("cooling tower", 5), ("heat exchanger", 6), ("pump", 5), ("air handler", 4)]),
           ]
    rows = ""
    for cat, c, items in lib:
        rows += f'<div role="treeitem" aria-expanded="true" style="color:var(--tx-2);font-weight:600">{ic("chev", 11)}{cat}<span class="muted" style="margin-left:auto;font-weight:400">{len(items)}</span></div>'
        rows += "".join(f'<div role="treeitem" draggable="true" style="padding-left:28px"><span style="width:10px;height:10px;border-radius:3px;background:var(--c-{c})"></span>{n}<span class="muted" style="margin-left:auto">{p} ports</span></div>' for n, p in items)
    rows += "".join(f'<div role="treeitem" aria-expanded="false" style="color:var(--tx-2);font-weight:600">{ic("chev", 11)}{cat}<span class="muted" style="margin-left:auto;font-weight:400">{k}</span></div>'
                    for cat, k in [("Compressed air", 3), ("Production", 4), ("Electrical", 3), ("Process", 3), ("Site", 2), ("Instruments", 17)])
    return (f'<aside class="panel" style="border-right:1px solid var(--line)"><div class="ph">Outline<span class="r">23 instances · 20 entries</span></div>'
            f'<div class="tree" role="tree" aria-label="Plant outline" style="padding:4px 0">{t}</div>'
            f'<div class="ph" style="border-top:1px solid var(--line)">Library<span class="r">drag onto the graph, or press A</span></div>'
            f'<div style="padding:8px 12px"><label class="input">{ic("search", 13)}<span class="muted">Search components</span></label></div>'
            f'<div class="tree" role="tree" aria-label="Component library">{rows}</div>{nav()}</aside>')


def right():
    params = (f'<div class="field"><label for="nm">Name</label><div class="input mono" id="nm">CT-1</div></div>'
              f'<div class="field"><label>Area</label><button class="input" aria-haspopup="listbox">Chiller plant<span style="margin-left:auto">{ic("chev", 11)}</span></button></div>'
              f'<div class="field"><label>Preset</label><div class="seg" role="group"><button aria-pressed="false">small</button><button aria-pressed="true">medium</button><button aria-pressed="false">large</button></div></div>'
              f'<div class="field"><label>fan_speed</label><div style="display:flex;align-items:center;gap:8px"><div style="flex:1;height:4px;border-radius:2px;background:var(--line-3);position:relative">'
              f'<div style="position:absolute;width:50%;height:4px;border-radius:2px;background:var(--accent)"></div><div role="slider" aria-valuenow="0.6" aria-valuemin="0.2" aria-valuemax="1" style="position:absolute;left:calc(50% - 7px);top:-5px;width:14px;height:14px;border-radius:50%;background:var(--tx)"></div></div>'
              f'<span class="mono">0.60</span><span class="muted">0.20–1.00</span></div></div>')
    ports = [("heat_in", "heat", "CH-1.heat_out"), ("wet_bulb_in", "env", "WX-1.wet_bulb_out"), ("cw_flow_in", "cw", "P-4.cw_flow_out"),
             ("basin_out", "temp", "CH-1.cond_water_in"), ("power_out", "power", "MB-COOL.loads_in")]
    p = "".join(f'<tr><td><span class="pd on" style="position:relative;display:inline-block;color:var(--t-{t});margin-right:8px;vertical-align:-1px"></span>{n}</td><td class="mono" style="color:var(--tx)">{c}</td></tr>' for n, t, c in ports)
    inst = [("basin_temp", "RTD", "TT"), ("fan_power", "Power meter", "JT"), ("hot_water_temp", "RTD · optional, off", "TT")]
    i = "".join(f'<tr><td class="mono">{n}</td><td>{a}</td><td><span class="badge o">{tg}</span></td></tr>' for n, a, tg in inst)
    faults = ('<div style="padding:4px 12px;display:flex;flex-wrap:wrap;gap:6px"><span class="badge o">fan_failure</span><span class="badge o">fill_fouling</span>'
              '<span class="badge o">humid_spell · negative class</span><span class="badge o">+ 10 instrument faults</span></div>')
    return (f'<aside class="panel" style="border-left:1px solid var(--line)"><div class="ph"><div class="tabs" role="tablist"><span class="tab" role="tab" aria-selected="true">Details</span><span class="tab" role="tab" aria-selected="false">Component</span></div></div>'
            f'<div style="padding:12px;display:flex;gap:10px;align-items:center;border-bottom:1px solid var(--line)"><div style="width:36px;height:36px;border-radius:8px;background:var(--panel-3);box-shadow:inset 0 3px 0 var(--c-cool);display:grid;place-items:center">{ic("plant", 16)}</div>'
            f'<div style="min-width:0"><div style="font-weight:600;font-size:14px" class="mono">CT-1</div><div class="sub">Cooling tower · streamsim/cooling-tower@1</div></div></div>'
            f'<div class="sect"><div class="sh">Properties</div>{params}</div>'
            f'<div class="sect"><div class="sh">Ports <span class="r">5 of 5 connected</span></div><table class="r">{p}</table></div>'
            f'<div class="sect"><div class="sh">Instruments</div><table class="r">{i}</table></div>'
            f'<div class="sect"><div class="sh">Faults it can suffer <span class="r">13</span></div>{faults}</div></aside>')


def go(label):
    return f'<button class="b solid">{label}</button>' if label else ""


def results(error=False):
    if error:
        rows = [("alarm", "x", "Compose failed: port P-4.cw_flow_out is of type condenser_water, port CH-1.flow_in of type chilled_water", "/connections/29", "Go to wire")]
    else:
        rows = [("ok", "check", "Composed factory-utilities.composition.json → domains/factory-utilities.domain.json", "20 kinds · 74 channels · 34 effectors · 399 faults", ""),
                ("ok", "check", "Lint · a nominal day, 1,441 samples, seed 1", "0 errors · 0 warnings", "")]
    r = "".join(f'<tr><td style="width:28px"><span class="ico-{s}">{ic(i, 14)}</span></td><td style="color:var(--tx);white-space:normal">{m}</td><td class="mono" style="text-align:right">{d}</td>'
                f'<td style="width:96px;text-align:right">{go(g)}</td></tr>' for s, i, m, d, g in rows)
    status = (f'<span class="badge alarm-soft">{ic("x", 11)} 1 error</span><span class="muted" style="text-transform:none;letter-spacing:0;font-weight:400">compose reports the first issue only</span>'
              if error else f'<span class="badge ok">{ic("check", 11)} 0 issues</span>')
    return (f'<section class="panel" style="border-top:1px solid var(--line)" aria-label="Compiler results"><div class="ph"><div class="tabs" role="tablist"><span class="tab" role="tab" aria-selected="true">Compiler results</span><span class="tab" role="tab" aria-selected="false">Output</span></div>{status}</div>'
            f'<table class="r" aria-live="polite">{r}</table></section>')


def toolbar(error=False):
    compile_btn = (f'<button class="b solid compile-err"><span class="st">{ic("x", 13)}</span> Compile · 1 error</button>' if error
                   else f'<button class="b solid compile-ok"><span class="st">{ic("check", 13)}</span> Compiled</button>')
    return (f'<button class="b solid">{ic("file")} Save <span class="kbd">⌘S</span></button>{compile_btn}'
            f'<button class="b primary" aria-haspopup="menu" {"disabled" if error else ""}>{ic("play", 12)} Simulate {ic("chev", 11)}</button>')


def simulate_menu():
    return (f'<div class="menu" role="menu" style="right:96px;top:46px;width:340px"><div class="mh">Start a run of factory-utilities</div>'
            f'<div class="field"><label>Mode</label><div class="seg" role="group"><button aria-pressed="true">Explore</button><button aria-pressed="false">Evaluate</button></div></div>'
            f'<div class="field" style="padding-top:0"><span></span><span class="sub" style="font-size:11px">Explore shows true state; Evaluate seals it until the run ends.</span></div>'
            f'<div class="field"><label>Profile</label><button class="input">nominal · 24 h<span style="margin-left:auto">{ic("chev", 11)}</span></button></div>'
            f'<div class="field"><label>Seed</label><div class="input mono">7</div></div>'
            f'<div class="field"><label>Adapter</label><button class="input">native-jsonl<span style="margin-left:auto">{ic("chev", 11)}</span></button></div>'
            f'<div class="field"><label>Clock step</label><div class="seg" role="group"><button aria-pressed="false">1 s</button><button aria-pressed="true">10 s</button><button aria-pressed="false">60 s</button></div></div>'
            f'<div class="sep"></div><div style="display:flex;justify-content:flex-end;gap:8px;padding:2px 12px"><button class="b">Cancel</button><button class="b primary">{ic("play", 12)} Start Explore run</button></div></div>')


def screen(theme="dark", error=False, connect=False, menu=False, show_legend=False):
    extra = legend() if show_legend else ""
    if connect:
        extra += '<div class="tip" style="left:600px;top:20px">chilled_water → CH-1.flow_in · release to connect</div>'
    g = graph(error, connect, extra)
    center = (f'<div style="display:grid;grid-template-rows:auto minmax(0,1fr) 168px;min-width:0;min-height:0;overflow:hidden">'
              f'{graph_bar(crumbs_factory(), zoom_bar(100, show_legend))}{g}{results(error)}</div>')
    body = topbar("Build", toolbar(error)) + doc_tabs(0, DOCS) + f'<div class="main" style="grid-template-columns:264px minmax(0,1fr) 320px">{left()}{center}{right()}</div>'
    if menu:
        body += simulate_menu()
    return page("Builder", body, theme)


write("builder.html", screen())
write("builder-error.html", screen(error=True))
write("builder-connect.html", screen(connect=True))
write("builder-simulate.html", screen(menu=True))
