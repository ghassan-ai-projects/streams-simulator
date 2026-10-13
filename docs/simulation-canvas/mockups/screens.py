from bp import Node, node_html, wire, comment, page, ic, W

def factory_nodes(live=False):
    v = (lambda x: x) if live else (lambda x: None)
    N = {}
    def add(n): N[n.id] = n
    add(Node("AC-1", "compressor · lead", "air", 30, 92, ins=[("pressure_in", "pressure", v("6.92 bar"), "1 s")], outs=[("air_out", "air", v("700 m³/h")), ("power_out", "power", v("76.2 kW"))], readout=[("air","700","m³/h",False),("power","76.2","kW",False)]))
    add(Node("AC-2", "compressor · lag", "air", 30, 222, ins=[("pressure_in", "pressure", v("6.92 bar"), "1 s")], outs=[("air_out", "air", v("0 m³/h")), ("power_out", "power", v("22.4 kW"))], foot=(f'<span class="badge warn">unloaded</span><span>band 650–700 kPa</span>' if live else ""), readout=[("air","0","m³/h",False),("power","22.4","kW",False)]))
    add(Node("R-1", "air receiver", "air", 270, 104, ins=[("air_in", "air", v("700 m³/h")), ("demand_in", "air", v("1,153 m³/h"))], outs=[("pressure_out", "pressure", v("6.92 bar")), ("air_out", "air", v("1,153 m³/h"))], readout=[("pressure","6.92","bar",False)]))
    add(Node("D-1", "air dryer", "air", 270, 250, ins=[("air_in", "air", v("1,153 m³/h"))], outs=[("dew_point_out", "temp", v("−40.0 °C")), ("power_out", "power", v("7.1 kW"))], readout=[("dew point","−40.0","°C",False)]))
    add(Node("CH-1", "chiller · medium", "cool", 510, 92, ins=[("load_in", "heat", v("1,118 kW")), ("flow_in", "water", v("495 m³/h")), ("cond_water_in", "cw", v("38.4 °C"), "10 s"), ("tariff_in", "signal", v("peak"), None, True, False, "← CAL-1")], outs=[("supply_out", "temp", v("7.31 °C")), ("return_out", "temp", v("9.2 °C")), ("heat_out", "heat", v("1,212 kW")), ("power_out", "power", v("92.7 kW"), None, True, False, "→ MB-COOL")], readout=[("supply","7.31","°C",False),("COP","3.6","",False)]))
    add(Node("CT-1", "cooling tower", "cool", 744, 92, ins=[("heat_in", "heat", v("1,212 kW")), ("wet_bulb_in", "env", v("22.9 °C"), None, True, False, "← WX-1"), ("cw_flow_in", "cw", v("165 m³/h"))], outs=[("basin_out", "cw", v("38.4 °C")), ("power_out", "power", v("0.0 kW"), None, True, False, "→ MB-COOL")], state="alarm" if live else "ok", foot=(f'<span class="badge alarm">{ic("bolt",11)} fan failure</span><span>since 09:40</span>' if live else ""), readout=[("basin","38.4","°C",True),("fan","0.0","kW",True)]))
    add(Node("P-4", "condenser pump", "cool", 744, 268, outs=[("cw_flow_out", "cw", v("165 m³/h")), ("power_out", "power", v("16.4 kW"), None, True, False, "→ MB-COOL")], readout=[("flow","165","m³/h",False)]))
    add(Node("P-1…3", "chw pump ×3", "cool", 510, 286, outs=[("flow_out", "water", v("495 m³/h")), ("power_out", "power", v("49.2 kW"), None, True, False, "→ MB-COOL")], stack=2, readout=[("flow","495","m³/h",False)]))
    for nid, x in [("M-1", 30), ("M-2", 270)]:
        add(Node(nid, "molder · medium", "prod", x, 438, readout=[("power", "208" if nid == "M-1" else "207", "kW", False), ("rate", "96", "%", False)], ins=[("air_in", "pressure", v("6.92 bar"), "10 s"), ("coolant_in", "temp", v("7.31 °C"), "10 s"), ("dew_point_in", "temp", v("−40.0 °C"), "10 s", nid == "M-1", nid != "M-1"), ("shift_in", "signal", v("day"), None, True, False, "← CAL-1")], outs=[("heat_out", "heat", v("486 kW")), ("air_out", "air", v("246 m³/h")), ("power_out", "power", v("208 kW"), None, True, False, "→ MB-PROC")]))
    add(Node("MB-AIR", "feeder meter · small", "elec", 510, 470, ins=[("loads_in", "power", v("105.7 kW"))], outs=[("power_out", "power", v("105.7 kW"))], readout=[("power","105.7","kW",False)]))
    add(Node("MM-1", "feeder meter · large", "elec", 744, 470, ins=[("loads_in", "power", v("887 kW"))], outs=[("power_out", "power", v("887 kW"))], readout=[("power","887","kW",False)]))
    return N

LINKS = [("AC-1", "air_out", "R-1", "air_in", "air"), ("AC-2", "air_out", "R-1", "air_in", "air"), ("R-1", "air_out", "D-1", "air_in", "air"),
         ("R-1", "pressure_out", "AC-1", "pressure_in", "pressure"), ("R-1", "pressure_out", "AC-2", "pressure_in", "pressure"),
         ("R-1", "pressure_out", "M-1", "air_in", "pressure"), ("R-1", "pressure_out", "M-2", "air_in", "pressure"),
         ("M-1", "air_out", "R-1", "demand_in", "air"), ("M-2", "air_out", "R-1", "demand_in", "air"),
         ("D-1", "dew_point_out", "M-1", "dew_point_in", "temp"),
         ("CH-1", "supply_out", "M-1", "coolant_in", "temp"), ("CH-1", "supply_out", "M-2", "coolant_in", "temp"),
         ("M-1", "heat_out", "CH-1", "load_in", "heat"), ("M-2", "heat_out", "CH-1", "load_in", "heat"),
         ("P-1…3", "flow_out", "CH-1", "flow_in", "water"), ("CH-1", "heat_out", "CT-1", "heat_in", "heat"),
         ("CT-1", "basin_out", "CH-1", "cond_water_in", "cw"), ("P-4", "cw_flow_out", "CT-1", "cw_flow_in", "cw"),
         ("AC-1", "power_out", "MB-AIR", "loads_in", "power"), ("AC-2", "power_out", "MB-AIR", "loads_in", "power"), ("D-1", "power_out", "MB-AIR", "loads_in", "power"),
         ("MB-AIR", "power_out", "MM-1", "loads_in", "power")]

def portal(xy, label, typ, side="in", scale=1):
    x, y = xy
    if side == "in":
        return f'<div class="portal" style="left:{x-8-len(label)*7.2-22}px;top:{y-10}px;border-color:var(--t-{typ});color:var(--t-{typ})">◀ {label}</div>'
    return f'<div class="portal" style="left:{x+10}px;top:{y-10}px;border-color:var(--t-{typ});color:var(--t-{typ})">{label} ▶</div>'

def graph(live=False, selected="CT-1", scale=0.86, extra=""):
    N = factory_nodes(live)
    for n in N.values():
        n.ro_h = 46 if (live and n.readout) else 0
        n.width = 268 if (live and n.id == selected) else 196
    frames = (comment("Compressor room", 14, 44, 476, 326, "air", "4 nodes") + comment("Chiller plant", 496, 44, 470, 378, "cool", "4 nodes · 6 pumps")
              + comment("Production hall", 14, 390, 476, 300, "prod", "2 of 5 shown") + comment("Substation", 496, 432, 470, 150, "elec", "2 of 6 shown"))
    related = lambda a, b: selected is None or selected in (a, b)
    wires = "".join(wire(N[a].out_xy(o), N[b].in_xy(i), t, live, dim=not related(a, b)) for a, o, b, i, t in LINKS)
    portals = ""
    nodes = "".join(node_html(n, live, selected) for n in N.values())
    world = f'<div style="position:absolute;left:0;top:0;width:980px;height:720px;transform:scale({scale});transform-origin:0 0">{frames}<svg class="wires" width="980" height="720">{wires}</svg>{nodes}{portals}</div>'
    return world + extra

def minimap():
    blocks = [(8, 10, 60, 40, "--h-air"), (74, 10, 60, 46, "--h-cool"), (8, 56, 60, 38, "--h-prod"), (74, 62, 60, 30, "--h-elec"), (140, 10, 40, 84, "--h-env")]
    rects = "".join(f'<div style="position:absolute;left:{x}px;top:{y}px;width:{w}px;height:{h}px;border-radius:3px;background:color-mix(in srgb,var({c}) 55%,transparent)"></div>' for x, y, w, h, c in blocks)
    return f'<div class="minimap">{rects}<div style="position:absolute;left:4px;top:6px;width:134px;height:92px;border:1.5px solid var(--accent);border-radius:4px"></div></div>'

def zoom():
    return f'<div class="zoom"><span class="b">{ic("minus")}</span><span class="b mono" style="padding:0 4px">87 %</span><span class="b">{ic("plus")}</span><span class="b">{ic("fit")}</span></div>'

def topbar(right=""):
    return (f'<header class="top"><span class="logo">{ic("run",13)}</span><span class="title">Streams Simulator</span><span class="muted">lab-workspace</span>'
            f'<div class="tb" style="margin-left:12px"><span class="b">{ic("plant")} Plant</span><span class="b on">{ic("blocks")} Build</span><span class="b">{ic("run")} Run</span><span class="b">{ic("check")} Review</span></div>'
            f'<span class="spacer"></span>{right}<span class="b">{ic("search")} Search <span class="muted mono" style="margin-left:6px">⌘K</span></span><span class="b">{ic("moon")}</span></header>')

def docs(active, tabs=None):
    tabs = tabs or [("factory-utilities", "composition", "blocks"), ("factory-utilities", "plant", "plant"), ("w-3", "live run", "run")]
    return '<div class="docs">' + "".join(f'<div class="doc {"on" if i == active else ""}">{ic(icn,13)}<span>{n}</span><span class="muted">{k}</span><span class="x">×</span></div>' for i, (n, k, icn) in enumerate(tabs)) + "</div>"
