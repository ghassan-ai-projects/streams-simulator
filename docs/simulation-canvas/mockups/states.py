from bp import ic, page, write, node_html, wire, legend, Node, Pin, TYPE_NAMES


def cell(label, body, w=232, h=None):
    hh = f"height:{h}px;" if h else ""
    return (f'<figure style="margin:0;width:{w}px"><div style="position:relative;{hh}padding:12px;border:1px solid var(--line);border-radius:10px;background:var(--bg)">{body}</div>'
            f'<figcaption class="sub" style="margin-top:6px;font-size:11px">{label}</figcaption></figure>')


def row(title, cells, note=""):
    return (f'<section style="margin-bottom:24px"><h2 style="font-size:11px;letter-spacing:.05em;text-transform:uppercase;color:var(--tx-2);margin:0 0 4px">{title}</h2>'
            f'<p class="sub" style="margin:0 0 10px">{note}</p><div style="display:flex;flex-wrap:wrap;gap:16px">{"".join(cells)}</div></section>')


def controls():
    b = [("Default", '<button class="b solid">Save</button>'), ("Hover", '<button class="b solid is-hover">Save</button>'),
         ("Keyboard focus", '<button class="b solid is-focus">Save</button>'), ("Pressed / on", '<button class="b" aria-pressed="true">Run</button>'),
         ("Disabled · not available yet", f'<button class="b solid" disabled title="Not available yet: the bridge has no action for this">Compare halls</button><div class="tip" style="position:static;margin-top:8px;display:inline-block">Not available yet: the bridge has no action for this</div>'),
         ("Primary", f'<button class="b primary">{ic("play", 12)} Simulate</button>'), ("Primary · focus", f'<button class="b primary is-focus">{ic("play", 12)} Simulate</button>'),
         ("Primary · disabled", f'<button class="b primary" disabled>{ic("play", 12)} Simulate</button>')]
    i = [("Input", '<div class="input mono">CT-1</div>'), ("Input · focus", '<div class="input mono is-focus">CT-1</div>'),
         ("Read only", f'<div class="input ro">{ic("lock", 12)} streamsim/cooling-tower@1</div>'),
         ("Invalid", f'<div class="input bad mono">CT 1</div><div style="color:var(--alarm);font-size:11px;margin-top:4px;display:flex;gap:4px">{ic("x", 12)} Names may not contain spaces</div>'),
         ("Segmented", '<div class="seg"><button aria-pressed="false">small</button><button aria-pressed="true">medium</button><button aria-pressed="false">large</button></div>'),
         ("Tabs", '<div class="tabs" style="height:32px"><span class="tab" aria-selected="true">Details</span><span class="tab" aria-selected="false">Component</span></div>')]
    return row("Controls", [cell(l, x, 200) for l, x in b + i], "Focus is a 2 px accent ring with 2 px offset. Disabled is 50 % with a reason on hover. Read only is a dashed border with a lock.")


def nodes():
    def mk(state="", tier="full"):
        n = Node("CT-1", "cooling tower", "cool", [Pin("heat_in", "heat"), Pin("cw_flow_in", "cw")], [Pin("basin_out", "temp")], tier=tier, state=state)
        n.x, n.y = 16, 12
        return n
    specs = [("Default", "", None, None, "build"), ("Hover", "is-hover", None, None, "build"), ("Selected", "", "CT-1", None, "build"), ("Keyboard focus", "", None, "CT-1", "build"),
             ("Selected and focused", "", "CT-1", "CT-1", "build"), ("Run · neutral header, category stripe", "", None, None, "run"),
             ("Fault injected · dashed red", "fault", None, None, "run"), ("Alarm · solid red", "alarm", None, None, "run"),
             ("Stopped · amber", "stopped", None, None, "run"), ("Not on the trace · 55 %", "dim", None, None, "run")]
    cells = []
    for label, state, sel, focus, mode in specs:
        n = mk(state)
        if state == "alarm":
            n.foot = f'<span class="badge alarm">{ic("bell", 11)} basin high · unack</span>'
        if state == "fault":
            n.foot = f'<span class="badge alarm-soft">{ic("bolt", 11)} fan_failure</span>'
        if state == "stopped":
            n.foot = f'<span class="badge warn">{ic("square", 9)} stopped</span>'
        cells.append(cell(label, f'<div class="{mode}" style="position:relative;height:{n.height() + 28}px">{node_html(n, sel, focus)}</div>', 272))
    ro = mk()
    cells.append(cell("Read only · review", f'<div class="run" style="position:relative;height:{ro.height() + 28}px">{node_html(ro)}<span class="badge o" style="position:absolute;right:0;bottom:0;background:var(--panel)">{ic("lock", 11)} read only</span></div>', 272))
    return row("Nodes", cells, "Build mode tints the header by category. Run and Review use a neutral header with a 3 px category stripe, so only alarms and faults draw the eye.")


def pins():
    specs = [("Not connected · hollow", Pin("level_in", "water", on=False)), ("Connected · filled", Pin("flow_in", "water")),
             ("Hover · 14 px", Pin("flow_in", "water", state="hover")), ("Keyboard focus", Pin("flow_in", "water", state="focus")),
             ("Compatible target while wiring", Pin("flow_in", "water", state="compat")), ("Incompatible while wiring", Pin("heat_in", "heat", state="incompat")),
             ("Sampled every 10 s", Pin("cond_water_in", "temp", sampled="10 s")), ("Wire leaves the view", Pin("load_in", "heat", via="← M-1 – M-4")),
             ("Compile error", Pin("flow_in", "water", err=True))]
    from bp import pin_html
    cells = [cell(l, f'<div class="node full" style="position:relative;--nw:200px;--cat:var(--c-cool)"><div class="nb">{pin_html(p, "in")}</div></div>'
                  + ('<div class="tip" style="left:40px;top:44px">flow_in · chilled_water · water</div>' if p.state == "hover" else ""), 232, 96 if p.state == "hover" else None) for l, p in specs]
    return row("Pins", cells, "Shape names the family (○ flow, □ energy, ◇ state and signals); colour names the type; the legend and every tooltip name it in words.")


def wires():
    specs = [("Normal", "water", ""), ("Hover", "water", "hover"), ("Selected · accent halo", "water", "sel"), ("Not on the trace · 60 %", "water", "dim"),
             ("On the trace · 3.2 px", "power", "trace"), ("Compile error · red halo", "cw", "err"), ("Signal · dashed", "signal", ""), ("Weather · dotted", "env", "")]
    cells = []
    for label, typ, style in specs:
        w = wire((10, 30), (190, 30), typ, "sel" if style == "sel" else style, 3 if style == "hover" else 2)
        cells.append(cell(label, f'<svg width="200" height="60" aria-hidden="true">{w}</svg>', 224))
    return row("Wires", cells, "Dimmed wires keep at least 3:1 against the canvas. A selected wire is 3.2 px with an accent halo; Delete removes it.")


def compile_states():
    specs = [("Compiled", f'<button class="b solid compile-ok"><span class="st">{ic("check", 13)}</span> Compiled</button>'),
             ("Compiled with 1 warning · opens the results", f'<button class="b solid compile-warn"><span class="st">{ic("alert", 13)}</span> Compiled · 1 warning</button>'),
             ("1 error · Simulate disabled", f'<button class="b solid compile-err"><span class="st">{ic("x", 13)}</span> Compile · 1 error</button>'),
             ("Compiling", f'<button class="b solid" aria-busy="true"><span class="st" style="width:12px;height:12px;border:2px solid var(--line-3);border-top-color:var(--accent);border-radius:50%"></span> Compiling…</button>'),
             ("Edited since the last compile", f'<button class="b solid"><span class="st" style="width:8px;height:8px;border-radius:50%;background:var(--tx-2)"></span> Compile <span class="kbd">⌘B</span></button>')]
    return row("Compile button", [cell(l, x, 260) for l, x in specs])


def chips():
    specs = [("Live run", '<span class="badge live n"><span class="pulse"></span>live</span>'), ("Explore · true state shown", f'<span class="badge truth">{ic("eye", 11)} Explore</span>'),
             ("Evaluate · sealed", f'<span class="badge n">{ic("lock", 11)} Evaluate · sealed</span>'), ("Verified", f'<span class="badge ok">{ic("check", 11)} Verified</span>'),
             ("Not verified", f'<span class="badge n">{ic("x", 11)} Not verified</span>'), ("Incomplete", f'<span class="badge n">{ic("alert", 11)} Incomplete</span>'),
             ("Alarm", f'<span class="badge alarm">{ic("bell", 11)} alarm</span>'), ("Fault injected", f'<span class="badge alarm-soft">{ic("bolt", 11)} fault injected</span>'),
             ("Stopped", f'<span class="badge warn">{ic("square", 9)} stopped</span>'), ("Read only", f'<span class="badge o">{ic("lock", 11)} read only</span>')]
    return row("Status chips", [cell(l, x, 180) for l, x in specs], "Red is only alarms and faults, amber only warnings and stopped equipment, purple only truth and Explore, accent blue only selection, focus and the primary action. Every chip carries an icon and a word.")


def keyboard():
    keys = [("Tab / Shift Tab", "top bar → left panel → graph → details → results"), ("Arrows", "move to the nearest node in that direction (the graph is one tab stop)"),
            ("Enter", "go into the node's pins; Up and Down move between pins"), ("Left / Right on a pin", "jump along its wire to the connected pin"),
            ("C", "start a wire from the focused pin; Tab cycles compatible pins; Enter connects"), ("Delete", "remove the selected wire or node"),
            ("A", "search the library and add at the focus"), ("F6", "next compile issue"), ("/ or ⌘K", "search"), ("+ − 0", "zoom in, out, fit"),
            ("Shift Arrows", "nudge the selection 16 px"), ("Esc", "back out: pins → node → area → site")]
    t = "".join(f'<tr><td style="width:200px"><span class="kbd">{k}</span></td><td style="white-space:normal">{v}</td></tr>' for k, v in keys)
    note = ("<p class='sub' style='margin:8px 0 0'>An Outline tree is the non-graph view of every instance and wire. Compile results and new alarms are announced through a polite live region. "
            "Animations stop under prefers-reduced-motion. Pin hit areas are 24 px.</p>")
    return row("Keyboard", [f'<div style="width:760px;background:var(--panel);border:1px solid var(--line);border-radius:10px;overflow:hidden"><table class="r">{t}</table></div>{note}'])


def screen(theme="dark"):
    lg = f'<div style="position:relative;height:430px;width:400px">{legend("left:0;top:0")}</div>'
    body = (f'<header class="top"><span class="logo">{ic("run", 13)}</span><span class="title">Streams Simulator</span><span class="sub">State sheet · C3, C5, C12</span></header>'
            f'<div class="scroll" style="grid-row:2/4;overflow:auto;padding:24px 32px">{controls()}{nodes()}{pins()}{wires()}{compile_states()}{chips()}'
            f'{row("Port-type legend", [lg], "Nine colours for the 19 library port types. The canvas maps each library type to one colour; the tooltip shows the exact type.")}{keyboard()}</div>')
    return page("State sheet", body, theme)


write("states.html", screen())
