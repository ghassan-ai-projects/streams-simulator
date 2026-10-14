import pathlib
ICONS = pathlib.Path("icons.inc").read_text()
HEAD, PADT, ROW = 28, 4, 24
NW = {"full": 232, "compact": 176}
FAMILY = {"air": ("circle", ""), "water": ("circle", ""), "cw": ("circle", ""), "heat": ("sq", ""), "power": ("sq", ""),
          "pressure": ("dia", ""), "temp": ("dia", ""), "signal": ("dia", "7 5"), "env": ("dia", "2 4")}
TYPE_NAMES = {"air": ("Compressed air", "compressed_air"), "pressure": ("Pressure", "pressure"),
              "water": ("Water", "chilled_water, hot_water, liquid_flow, level, steam"), "cw": ("Condenser water", "condenser_water"),
              "temp": ("Temperature", "temperature, cooling_water_temp, dew_point"), "heat": ("Heat", "thermal_load, heat_rejection"),
              "power": ("Power", "electrical, load_torque"), "signal": ("Signal", "shift, tariff"), "env": ("Weather", "dry_bulb_temp, wet_bulb_temp")}
CAT = {"air": "Compressed air", "cool": "Cooling", "prod": "Production", "elec": "Electrical", "env": "Site"}


def ic(name, size=14):
    return f'<svg width="{size}" height="{size}" aria-hidden="true"><use href="#i-{name}"/></svg>'


class Pin:
    def __init__(self, name, typ, on=True, sampled=None, via=None, err=False, state=""):
        self.name, self.typ, self.on, self.sampled, self.via, self.err, self.state = name, typ, on, sampled, via, err, state


class Node:
    def __init__(self, nid, kind, cat, ins=(), outs=(), readout=(), state="", foot="", tier="full"):
        self.id, self.kind, self.cat = nid, kind, cat
        self.ins, self.outs, self.readout, self.state, self.foot, self.tier = list(ins), list(outs), list(readout), state, foot, tier
        self.x = self.y = 0

    @property
    def w(self):
        return NW[self.tier]

    def ro_h(self):
        per = 48 if any(cell[4] for cell in self.readout) else 34
        return 0 if not self.readout else 8 + per * ((len(self.readout) + 1) // 2)

    def vis(self, pins):
        return pins if self.tier == "full" else [p for p in pins if p.on]

    def row(self):
        return ROW if self.tier == "full" else 12

    def pin_y(self, i):
        return self.y + HEAD + self.ro_h() + PADT + i * self.row() + self.row() / 2

    def in_xy(self, name):
        return self.x - 1, self.pin_y([p.name for p in self.vis(self.ins)].index(name))

    def out_xy(self, name):
        return self.x + self.w + 1, self.pin_y([p.name for p in self.vis(self.outs)].index(name))

    def height(self):
        return HEAD + self.ro_h() + PADT + max(len(self.vis(self.ins)), len(self.vis(self.outs))) * self.row() + 8 + (34 if self.foot else 0)


def pin_dot(p):
    shape = FAMILY[p.typ][0]
    cls = "pd" + (" on" if p.on else "") + ("" if shape == "circle" else " " + shape)
    return f'<span class="{cls}" style="color:var(--t-{p.typ})" title="{TYPE_NAMES[p.typ][0]}"></span>'


def pin_html(p, side):
    extra = (f'<span class="clock" title="sampled every {p.sampled}">{p.sampled}</span>' if p.sampled else "")
    extra += f'<span class="via" title="{p.via}">{p.via.split(" ")[0]}</span>' if p.via else ""
    extra += f'<span class="err" title="compile error">{ic("x", 12)}</span>' if p.err else ""
    name = f'<span class="nm">{p.name}</span>'
    body = f'{pin_dot(p)}{name}{extra}' if side == "in" else f'{extra}{name}{pin_dot(p)}'
    return f'<div class="pin {side} {p.state}" role="button" aria-label="{p.name}, {TYPE_NAMES[p.typ][0]}, {"connected" if p.on else "not connected"}">{body}</div>'


def readout_html(n):
    cells = ""
    for label, value, unit, cls, delta in n.readout:
        d = f'<div class="delta">{delta}</div>' if delta else ""
        cells += f'<div><div class="rl">{label}</div><div class="rv {cls}">{value}<span class="u">{unit}</span></div>{d}</div>'
    return f'<div class="ro">{cells}</div>'


def kind_html(n):
    if n.tier == "compact" and not n.kind.startswith("▸"):
        return ""
    cls = "cnt" if n.kind.startswith("▸") else "k"
    return f'<span class="{cls}" title="{n.kind}">{n.kind}</span>'


def node_html(n, sel=None, focus=None):
    ins, outs = n.vis(n.ins), n.vis(n.outs)
    rows = max(len(ins), len(outs))
    title = f' title="{n.id} · {n.kind}"' if n.tier == "compact" else ""
    body = "".join('<div class="prow">' + (pin_html(ins[i], "in") if i < len(ins) else "<div></div>")
                   + (pin_html(outs[i], "out") if i < len(outs) else "<div></div>") + "</div>" for i in range(rows))
    cls = f'node {n.tier} {n.state}' + (" sel" if n.id == sel else "") + (" focus" if n.id == focus else "")
    foot = f'<div class="nf">{n.foot}</div>' if n.foot else ""
    ro = readout_html(n) if n.readout else ""
    return (f'<div class="{cls}" role="group" aria-label="{n.id}, {n.kind}"{title} style="left:{n.x}px;top:{n.y}px;--nw:{n.w}px;--cat:var(--c-{n.cat})">'
            f'<div class="nh"><span class="id">{n.id}</span>{kind_html(n)}</div>{ro}<div class="nb">{body}</div>{foot}</div>')


def check_layout(nodes, height, width, gap=12):
    items = list(nodes)
    for n in items:
        if n.y + n.height() > height or n.x + n.w > width:
            raise SystemExit(f"{n.id} leaves the graph: bottom {n.y + n.height()} > {height} or right {n.x + n.w} > {width}")
    for i, a in enumerate(items):
        for b in items[i + 1:]:
            if a.x < b.x + b.w + gap and b.x < a.x + a.w + gap and a.y < b.y + b.height() + gap and b.y < a.y + a.height() + gap:
                raise SystemExit(f"{a.id} and {b.id} overlap")


def curve(a, b):
    (x1, y1), (x2, y2) = a, b
    dx = max(48, abs(x2 - x1) * 0.5)
    return f"M{x1:.0f},{y1:.0f} C{x1 + dx:.0f},{y1:.0f} {x2 - dx:.0f},{y2:.0f} {x2:.0f},{y2:.0f}"


def lane_path(a, b, lane, r=8):
    (x1, y1), (x2, y2) = a, b
    xo, xi = x1 + 16, x2 - 16
    down1 = 1 if lane > y1 else -1
    down2 = 1 if y2 > lane else -1
    return (f"M{x1:.0f},{y1:.0f} H{xo - r:.0f} Q{xo:.0f},{y1:.0f} {xo:.0f},{y1 + r * down1:.0f} V{lane - r * down1:.0f} Q{xo:.0f},{lane:.0f} {xo - r:.0f},{lane:.0f} "
            f"H{xi + r:.0f} Q{xi:.0f},{lane:.0f} {xi:.0f},{lane + r * down2:.0f} V{y2 - r * down2:.0f} Q{xi:.0f},{y2:.0f} {xi + r:.0f},{y2:.0f} H{x2:.0f}")


def connect(na, ap, nb, bp, typ, style="", lane=None):
    a, b = na.out_xy(ap), nb.in_xy(bp)
    if b[0] < a[0] + 32:
        lane = lane if lane is not None else max(na.y + na.height(), nb.y + nb.height()) + 10
    return wire(a, b, typ, style, lane=lane)


def wire(a, b, typ, style="", width=2, lane=None):
    d = curve(a, b) if lane is None else lane_path(a, b, lane)
    dash = FAMILY[typ][1]
    dash_attr = f' stroke-dasharray="{dash}"' if dash else ""
    op = ' class="wd"' if style == "dim" else ""
    w = 3.2 if style in ("sel", "trace") else width
    halo = ""
    if style == "sel":
        halo = f'<path d="{d}" stroke="var(--accent)" stroke-width="8" fill="none" opacity=".35"/>'
    if style == "err":
        halo = f'<path d="{d}" stroke="var(--alarm)" stroke-width="8" fill="none" opacity=".55"/>'
    if style == "hover":
        w = 3
        halo = "".join(f'<circle cx="{x:.0f}" cy="{y:.0f}" r="6" fill="var(--t-{typ})"/>' for x, y in (a, b))
    core = ""
    if typ == "heat":
        core = f'<path d="{d}" stroke="var(--bg)" stroke-width="{max(w - 2.4, 0.8):.1f}" fill="none"{op} stroke-linecap="round"/>'
        w += 1.6
    return f'{halo}<path d="{d}" stroke="var(--t-{typ})" stroke-width="{w}" fill="none"{dash_attr}{op} stroke-linecap="round"/>{core}'


def frame(title, x, y, w, h, cat, note=""):
    return (f'<div class="frame" style="left:{x}px;top:{y}px;width:{w}px;height:{h}px;--cat:var(--c-{cat})">'
            f'<div class="ft">{title}<span class="note">{note}</span></div></div>')


def dash_attr(dash):
    return f' stroke-dasharray="{dash}"' if dash else ""


def legend(pos="right:12px;top:8px"):
    rows = ""
    for typ, (name, lib) in TYPE_NAMES.items():
        shape, dash = FAMILY[typ]
        core = '<line x1="2" y1="6" x2="30" y2="6" stroke="var(--panel)" stroke-width="1"/>' if typ == "heat" else ""
        sw = (f'<svg width="44" height="12" aria-hidden="true"><line x1="2" y1="6" x2="30" y2="6" stroke="var(--t-{typ})" stroke-width="{4 if typ == "heat" else 2.4}"'
              f'{dash_attr(dash)}/>{core}</svg>')
        dot = f'<span class="pd on {"" if shape == "circle" else shape}" style="position:relative;display:inline-block;color:var(--t-{typ});margin-left:-14px"></span>'
        rows += f'<div class="lr"><span style="display:flex;align-items:center">{sw}{dot}</span><b>{name}</b><span>{lib}</span></div>'
    glyphs = ('<div class="lr" style="margin-top:6px;border-top:1px solid var(--line);padding-top:8px"><span><span class="clock">10 s</span></span><b>Sampled</b><span>the input reads its link every 10 s; this is what closes a loop</span></div>'
              '<div class="lr"><span><span class="via">→</span></span><b>Off-screen</b><span>the wire goes to an instance outside this view</span></div>'
              '<div class="lr"><span style="padding-left:2px"><span class="pd" style="position:relative;display:inline-block;color:var(--tx-2)"></span></span><b>Hollow pin</b><span>not connected; filled pins are connected</span></div>')
    return (f'<div class="legend" role="dialog" aria-label="Port types" style="{pos}"><div class="sh">Port types <span class="r">○ flow · □ energy · ◇ state and signals</span></div>{rows}{glyphs}</div>')


def navigator(blocks, vp, note=""):
    return f'<div class="nav"><div class="ph">Navigator<span class="r">{note}</span></div>{minimap(blocks, vp)}</div>'


def minimap(blocks, vp, pos=""):
    rects = "".join(f'<div style="position:absolute;left:{x}px;top:{y}px;width:{w}px;height:{h}px;border-radius:3px;background:{bg}"></div>' for x, y, w, h, bg in blocks)
    vx, vy, vw, vh = vp
    return f'<div class="minimap" role="img" aria-label="Minimap" style="{pos}">{rects}<div class="vp" style="left:{vx}px;top:{vy}px;width:{vw}px;height:{vh}px"></div></div>'


def zoom_bar(pct, legend_open=False):
    return (f'<div class="tb" style="flex:none"><button class="b" aria-label="Zoom out">{ic("minus")}</button><span class="b mono" aria-live="polite">{pct} %</span>'
            f'<button class="b" aria-label="Zoom in">{ic("plus")}</button><button class="b" aria-label="Fit to view">{ic("fit")}</button>'
            f'<span style="width:1px;height:20px;background:var(--line-2);margin:0 2px"></span><button class="b" aria-pressed="{str(legend_open).lower()}" aria-label="Port type legend">{ic("legend")} Legend</button></div>')


def theme_button():
    return f'<button class="b" aria-label="Switch theme">{ic("moon")}</button>'


def topbar(active, right="", workspace=True, compact=False):
    search = (f'<button class="b" aria-label="Search, ⌘K">{ic("search")}</button>' if compact
              else f'<button class="b">{ic("search")} Search <span class="kbd">⌘K</span></button>')
    tabs = [("plant", "Plant"), ("blocks", "Build"), ("run", "Run"), ("check", "Review")]
    t = "".join(f'<button class="b" role="tab" aria-selected="{str(name == active).lower()}">{ic(i)} {name}</button>' for i, name in tabs)
    ws = '<span class="sub">lab-workspace</span>' if workspace else ""
    return (f'<header class="top"><span class="logo">{ic("run", 13)}</span><span class="title">Streams Simulator</span>{ws}'
            f'<div class="tb" role="tablist" aria-label="Mode" style="margin-left:12px">{t}</div><span class="spacer"></span>{right}'
            f'{search}{theme_button()}</header>')


def doc_tabs(active, tabs):
    return '<div class="docs" role="tablist" aria-label="Open documents">' + "".join(
        f'<div class="doc" role="tab" aria-selected="{str(i == active).lower()}">{ic(icon, 13)}<span>{name}</span><span class="k">{kind}</span>'
        f'<button class="b" style="height:20px;padding:0 4px" aria-label="Close {name}">{ic("close", 11)}</button></div>' for i, (name, kind, icon) in enumerate(tabs)) + "</div>"


def page(title, body, theme="dark"):
    return (f'<!doctype html><html lang="en" data-theme="{theme}"><head><meta charset="utf-8"><title>{title}</title>'
            f'<link rel="stylesheet" href="bp.css"></head><body>{ICONS}<div class="app">{body}</div><div class="sr-only" aria-live="polite"></div></body></html>')


def write(name, html):
    pathlib.Path(name).write_text(html)
