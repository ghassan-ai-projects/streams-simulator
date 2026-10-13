from bp import Node, Pin

PORTS = {
    "compressor": ([("pressure_in", "pressure", "1 s")], [("air_out", "air"), ("power_out", "power")]),
    "receiver": ([("air_in", "air"), ("demand_in", "air")], [("pressure_out", "pressure"), ("air_out", "air")]),
    "dryer": ([("air_in", "air")], [("dew_point_out", "temp"), ("power_out", "power")]),
    "line": ([("shift_in", "signal"), ("air_in", "pressure", "10 s"), ("coolant_in", "temp", "10 s"), ("dew_point_in", "temp", "10 s")],
             [("heat_out", "heat"), ("air_out", "air"), ("power_out", "power")]),
    "ahu": ([("chw_in", "temp", "10 s"), ("outdoor_in", "env"), ("shift_in", "signal")], [("load_out", "heat")]),
    "chiller": ([("load_in", "heat"), ("flow_in", "water"), ("cond_water_in", "temp", "10 s"), ("tariff_in", "signal")],
                [("supply_out", "temp"), ("return_out", "temp"), ("heat_out", "heat"), ("power_out", "power")]),
    "tower": ([("heat_in", "heat"), ("wet_bulb_in", "env"), ("cw_flow_in", "cw")], [("basin_out", "temp"), ("power_out", "power")]),
    "pump": ([("level_in", "water")], [("flow_out", "water"), ("cw_flow_out", "cw"), ("liquid_out", "water"), ("power_out", "power")]),
    "meter": ([("loads_in", "power")], [("power_out", "power")]),
    "transformer": ([("load_in", "power")], []),
    "generator": ([("load_in", "power")], [("power_out", "power")]),
    "weather": ([], [("dry_bulb_out", "env"), ("wet_bulb_out", "env")]),
    "calendar": ([], [("shift_out", "signal"), ("tariff_out", "signal")]),
}

INSTANCES = [
    ("WX-1", "weather", "weather", "env"), ("CAL-1", "site calendar", "calendar", "env"),
    ("M-1", "production line · molder", "line", "prod"), ("M-2", "production line · molder", "line", "prod"),
    ("M-3", "production line · cnc · small", "line", "prod"), ("M-4", "packaging line", "line", "prod"), ("AHU-1", "air handler", "ahu", "prod"),
    ("AC-1", "air compressor · lead", "compressor", "air"), ("AC-2", "air compressor · lag", "compressor", "air"),
    ("R-1", "air receiver", "receiver", "air"), ("D-1", "air dryer", "dryer", "air"),
    ("CH-1", "chiller", "chiller", "cool"), ("CT-1", "cooling tower", "tower", "cool"),
    ("P-1", "pump · chw", "pump", "cool"), ("P-2", "pump · chw", "pump", "cool"), ("P-3", "pump · chw", "pump", "cool"), ("P-4", "pump · cw", "pump", "cool"),
    ("MB-AIR", "feeder meter · air · small", "meter", "elec"), ("MB-COOL", "feeder meter · cooling · small", "meter", "elec"),
    ("MB-PROC", "feeder meter · process", "meter", "elec"), ("MM-1", "feeder meter · main · large", "meter", "elec"),
    ("TX-1", "transformer · large", "transformer", "elec"), ("G-1", "generator", "generator", "elec"),
]

LINKS = """CAL-1.shift_out M-1.shift_in;CAL-1.shift_out M-2.shift_in;CAL-1.shift_out M-3.shift_in;CAL-1.shift_out M-4.shift_in;CAL-1.shift_out AHU-1.shift_in
CAL-1.tariff_out CH-1.tariff_in;WX-1.dry_bulb_out AHU-1.outdoor_in;WX-1.wet_bulb_out CT-1.wet_bulb_in;AC-1.air_out R-1.air_in;AC-2.air_out R-1.air_in
M-1.air_out R-1.demand_in;M-2.air_out R-1.demand_in;M-3.air_out R-1.demand_in;M-4.air_out R-1.demand_in;R-1.air_out D-1.air_in
R-1.pressure_out AC-1.pressure_in;R-1.pressure_out AC-2.pressure_in;R-1.pressure_out M-1.air_in;R-1.pressure_out M-2.air_in;R-1.pressure_out M-3.air_in
R-1.pressure_out M-4.air_in;D-1.dew_point_out M-1.dew_point_in;D-1.dew_point_out M-2.dew_point_in;D-1.dew_point_out M-3.dew_point_in;D-1.dew_point_out M-4.dew_point_in
M-1.heat_out CH-1.load_in;M-2.heat_out CH-1.load_in;M-3.heat_out CH-1.load_in;M-4.heat_out CH-1.load_in;AHU-1.load_out CH-1.load_in
P-1.flow_out CH-1.flow_in;P-2.flow_out CH-1.flow_in;P-3.flow_out CH-1.flow_in;CH-1.supply_out M-1.coolant_in;CH-1.supply_out M-2.coolant_in
CH-1.supply_out M-3.coolant_in;CH-1.supply_out M-4.coolant_in;CH-1.supply_out AHU-1.chw_in;CH-1.heat_out CT-1.heat_in;P-4.cw_flow_out CT-1.cw_flow_in
CT-1.basin_out CH-1.cond_water_in;AC-1.power_out MB-AIR.loads_in;AC-2.power_out MB-AIR.loads_in;D-1.power_out MB-AIR.loads_in;M-1.power_out MB-PROC.loads_in
M-2.power_out MB-PROC.loads_in;M-3.power_out MB-PROC.loads_in;M-4.power_out MB-PROC.loads_in;CH-1.power_out MB-COOL.loads_in;CT-1.power_out MB-COOL.loads_in
P-1.power_out MB-COOL.loads_in;P-2.power_out MB-COOL.loads_in;P-3.power_out MB-COOL.loads_in;P-4.power_out MB-COOL.loads_in;MB-AIR.power_out MM-1.loads_in
MB-COOL.power_out MM-1.loads_in;MB-PROC.power_out MM-1.loads_in;MM-1.power_out TX-1.load_in;MM-1.power_out G-1.load_in"""


def links():
    out = []
    for item in LINKS.replace("\n", ";").split(";"):
        a, b = item.split()
        out.append((*a.split("."), *b.split(".")))
    return out


def connected():
    ends = set()
    for a, ap, b, bp in links():
        ends.add((a, ap))
        ends.add((b, bp))
    return ends


def build(tier="full", only=None):
    on = connected()
    nodes = {}
    for nid, kind, comp, cat in INSTANCES:
        if only and nid not in only:
            continue
        ins, outs = PORTS[comp]
        pins_in = [Pin(p[0], p[1], (nid, p[0]) in on, p[2] if len(p) > 2 else None) for p in ins]
        pins_out = [Pin(p[0], p[1], (nid, p[0]) in on) for p in outs]
        nodes[nid] = Node(nid, kind, cat, pins_in, pins_out, tier=tier)
    return nodes


def type_of(nodes, nid, port):
    return next(p.typ for p in nodes[nid].outs if p.name == port)


def place(nodes, layout):
    for nid, (x, y) in layout.items():
        if nid in nodes:
            nodes[nid].x, nodes[nid].y = x, y


def wires_between(nodes, style_of=lambda link: ""):
    from bp import wire
    svg = ""
    for link in links():
        a, ap, b, bp = link
        if a in nodes and b in nodes:
            svg += wire(nodes[a].out_xy(ap), nodes[b].in_xy(bp), type_of(nodes, a, ap), style_of(link))
    return svg
