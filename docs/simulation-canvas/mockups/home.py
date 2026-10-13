from bp import ic, page, write, topbar

CODE = "3S2QK7"


def boxes(chars, state=""):
    out = ""
    for i in range(12):
        c = chars[i] if i < len(chars) else ""
        border = "var(--alarm)" if state == "wrong" else ("var(--accent)" if i == len(chars) and state == "" else "var(--line-3)")
        ring = "outline:2px solid var(--accent);outline-offset:2px;" if i == len(chars) and state == "" else ""
        out += (f'<div style="width:40px;height:52px;border:1px solid {border};{ring}border-radius:8px;display:grid;place-items:center;font:600 22px var(--font-mono);'
                f'background:var(--panel-2);{"opacity:.5;" if state == "locked" else ""}">{c}</div>')
        if i in (3, 7):
            out += '<span class="muted" style="font-size:20px" aria-hidden="true">–</span>'
    return f'<div role="group" aria-label="Pairing code, 12 characters" style="display:flex;align-items:center;gap:6px">{out}</div>'


def pairing(state="", theme="dark"):
    chars = {"": CODE, "wrong": "3S2QK7MPLXQ4", "locked": ""}[state]
    msg = {"": f'<span class="sub">{len(chars)} of 12 characters</span>',
           "wrong": f'<span class="ico-alarm">{ic("x", 14)}</span><span style="color:var(--alarm);font-weight:600">That code did not match. 4 tries left.</span>',
           "locked": f'<span class="ico-alarm">{ic("lock", 14)}</span><span style="color:var(--alarm);font-weight:600">Pairing is locked after 5 wrong codes. Restart streamsim canvas for a new code.</span>'}[state]
    disabled = "disabled" if state in ("", "locked") else ""
    card = (f'<main style="width:640px;background:var(--panel);border:1px solid var(--line);border-radius:14px;box-shadow:var(--shadow);padding:32px">'
            f'<div style="display:flex;align-items:center;gap:10px"><span class="logo" style="width:30px;height:30px">{ic("run", 16)}</span><span style="font-weight:600;font-size:16px">Streams Simulator</span></div>'
            f'<h1 style="font-size:22px;margin:24px 0 8px;font-weight:600">Pair this browser</h1>'
            f'<p style="margin:0 0 20px;color:var(--tx-2);font-size:13px">Type the 12-character code that <span class="mono" style="color:var(--tx)">streamsim canvas</span> printed in your terminal. '
            f'The code works once; this browser then stays paired for 12 hours.</p>{boxes(chars, state)}'
            f'<div style="display:flex;align-items:center;gap:8px;margin-top:12px;min-height:20px" aria-live="polite">{msg}</div>'
            f'<div style="display:flex;align-items:center;gap:12px;margin-top:20px"><button class="b primary" style="height:36px;padding:0 18px" {disabled}>Pair</button>'
            f'<span class="sub">Five wrong codes lock pairing until the bridge restarts.</span></div>'
            f'<div style="margin-top:24px;padding-top:16px;border-top:1px solid var(--line);color:var(--tx-2);font-size:12px">Workspace <span class="mono" style="color:var(--tx)">lab-workspace</span> · bridge '
            f'<span class="mono" style="color:var(--tx)">127.0.0.1:47811</span> · loopback only</div></main>')
    return page("Pairing", f'<div class="graph" style="grid-row:1/4;display:grid;place-items:center">{card}</div>', theme)


PLANTS = [("factory-utilities", "20 kinds · 23 instances · 82 tags", "composed from factory-utilities.composition.json", ["air", "cool", "prod", "elec"]),
          ("factory-utilities-campus-26", "20 kinds · 548 instances · 2,032 tags", "26 halls of the factory", ["env"] * 4),
          ("chiller-pilot", "3 kinds · 4 instances · 11 tags", "composed from chiller-pilot.composition.json", ["cool", "cool", "env", "cool"]),
          ("aquaculture-pond", "1 kind · 1 instance · 7 tags", "hand-written domain", ["env", "env", "env", "env"])]


def thumb(cats):
    cells = "".join(f'<div style="position:absolute;left:{12 + (i % 2) * 96}px;top:{12 + (i // 2) * 40}px;width:84px;height:30px;border-radius:5px;background:var(--panel-3);box-shadow:inset 0 3px 0 var(--c-{c})"></div>' for i, c in enumerate(cats))
    return f'<div aria-hidden="true" style="position:relative;height:96px;border-radius:8px;background:var(--bg);border:1px solid var(--line);overflow:hidden">{cells}</div>'


def card(name, meta, src, cats):
    return (f'<article style="background:var(--panel);border:1px solid var(--line);border-radius:12px;padding:12px;min-width:0">{thumb(cats)}'
            f'<h3 class="mono" style="margin:10px 0 2px;font-size:13px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="{name}">{name}</h3><div class="sub">{meta}</div>'
            f'<div class="muted" style="font-size:11px;margin-top:2px">{src}</div><div style="display:flex;gap:6px;margin-top:10px">'
            f'<button class="b solid">{ic("plant", 13)} Open</button><button class="b solid">{ic("play", 11)} Simulate</button></div></article>')


RUNS = [("w-4", "factory-utilities", "Evaluate", ("n live", '<span class="pulse"></span>', "live"), "running · 08:40", "Open live run"),
        ("w-2", "factory-utilities", "Evaluate", ("ok", ic("check", 11), "Verified"), "12 h · 57,908 events", "Open review"),
        ("w-1", "factory-utilities", "Evaluate", ("n", ic("x", 11), "Not verified"), "files differ from the run", "Open review"),
        ("p-4", "chiller-pilot", "Evaluate", ("n", ic("alert", 11), "Incomplete"), "stopped at 01:12", "Open review")]


def h2(text, right=""):
    return f'<h2 style="font-size:11px;letter-spacing:.05em;text-transform:uppercase;color:var(--tx-2);margin:0 0 10px;display:flex">{text}<span class="muted" style="margin-left:auto;text-transform:none;letter-spacing:0;font-weight:400">{right}</span></h2>'


def home(theme="dark"):
    cards = "".join(card(*p) for p in PLANTS)
    rows = "".join(f'<tr><td class="mono" style="color:var(--tx)">{i}</td><td>{d}</td><td>{m}</td><td><span class="badge {c}">{icn} {l}</span></td><td>{x}</td>'
                   f'<td style="text-align:right"><button class="b solid">{a}</button></td></tr>' for i, d, m, (c, icn, l), x, a in RUNS)
    comps = "".join(f'<tr><td class="mono" style="color:var(--tx)">{n}</td><td>{d}</td><td style="text-align:right"><button class="b solid">{ic("blocks", 13)} Edit</button></td></tr>'
                    for n, d in [("factory-utilities", "20 entries · 59 connections"), ("chiller-pilot", "3 entries · 1 connection")])
    main = (f'<div style="padding:24px 32px;overflow:auto;max-width:1600px"><div style="display:flex;align-items:flex-end;gap:12px"><h1 style="margin:0;font-size:20px;font-weight:600">lab-workspace</h1>'
            f'<span class="sub">4 plants · 2 compositions · 4 runs</span><span style="flex:1"></span>'
            f'<button class="b solid">{ic("blocks")} New composition</button><button class="b primary">{ic("plus", 12)} Build a plant</button></div>'
            f'<section style="margin-top:22px">{h2("Plants", "domains/")}<div style="display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:14px">{cards}</div></section>'
            f'<div style="display:grid;grid-template-columns:minmax(0,1.6fr) minmax(0,1fr);gap:16px;margin-top:22px"><section>{h2("Runs", "runs/")}'
            f'<div style="background:var(--panel);border:1px solid var(--line);border-radius:12px;overflow:hidden"><table class="r">{rows}</table></div></section>'
            f'<section>{h2("Compositions", "compositions/")}<div style="background:var(--panel);border:1px solid var(--line);border-radius:12px;overflow:hidden"><table class="r">{comps}</table></div></section></div></div>')
    top = topbar("", workspace=True).replace('<div class="tb" role="tablist" aria-label="Mode" style="margin-left:12px">', '<div class="tb" role="tablist" aria-label="Mode" style="margin-left:12px;display:none">')
    return page("Workspace", top + f'<div style="grid-row:2/4;overflow:hidden;background:var(--bg)">{main}</div>', theme)


def home_empty(theme="dark"):
    steps = [("1", "Build a plant", "Drag equipment from the library and wire it. The canvas writes compositions/<name>.composition.json."),
             ("2", "Compile", "streamsim composes it into domains/<name>.domain.json and checks it."),
             ("3", "Simulate", "Start an Explore run and watch it live.")]
    s = "".join(f'<li style="display:flex;gap:12px;align-items:flex-start;padding:10px 0"><span class="badge n" style="flex:none">{n}</span><div><div style="font-weight:600">{t}</div><div class="sub">{d}</div></div></li>' for n, t, d in steps)
    main = (f'<div style="display:grid;place-items:center;height:100%"><main style="width:560px;text-align:left"><h1 style="margin:0 0 8px;font-size:20px;font-weight:600">lab-workspace is empty</h1>'
            f'<p class="sub" style="margin:0 0 16px;font-size:13px">No plants, compositions or runs yet. Start with a plant built from the component library.</p>'
            f'<ol style="list-style:none;margin:0 0 20px;padding:0;border-top:1px solid var(--line);border-bottom:1px solid var(--line)">{s}</ol>'
            f'<button class="b primary" style="height:36px;padding:0 18px">{ic("plus", 12)} Build a plant</button>'
            f'<p class="sub" style="margin:16px 0 0">Domain files you copy into <span class="mono" style="color:var(--tx)">domains/</span> appear here on their own.</p></main></div>')
    top = topbar("", workspace=True).replace('<div class="tb" role="tablist" aria-label="Mode" style="margin-left:12px">', '<div class="tb" role="tablist" aria-label="Mode" style="margin-left:12px;display:none">')
    return page("Workspace", top + f'<div style="grid-row:2/4;overflow:hidden;background:var(--bg)">{main}</div>', theme)


write("pairing.html", pairing())
write("pairing-wrong.html", pairing("wrong"))
write("pairing-locked.html", pairing("locked"))
write("home.html", home())
write("home-empty.html", home_empty())
