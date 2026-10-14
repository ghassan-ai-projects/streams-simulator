#!/usr/bin/env python3
"""Extract the values the mockups display from real streamsim run folders.

Usage:
  python3 tools/extract.py                 write ../data.json from $SCRATCH/runs
  python3 tools/extract.py alarms RUN      print alarm/trip/return times of AC-1 as JSON (seconds of day)

Python 3.9 compatible. Observed values come from trace.jsonl, truth from world_state_history.jsonl.
"""
import json
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
OUT = os.path.join(os.path.dirname(HERE), "data.json")
SCRATCH = os.environ.get(
    "SCRATCH",
    "/private/tmp/claude-501/-Users-ghassan-my-projects-streams-simulator/0abb3216-3a51-47d8-8539-36e4038c8aa8/scratchpad/mockdata",
)
RUNS = os.path.join(SCRATCH, "runs")
MIN = 60
HOUR = 3600


def hms(sec):
    sec = int(sec)
    return "%02d:%02d:%02d" % (sec // 3600 % 24, sec // 60 % 60, sec % 60)


def tsec(text):
    return (int(text[8:10]) - 1) * 86400 + int(text[11:13]) * 3600 + int(text[14:16]) * 60 + float(text[17:-1])


def sec_of(hms_text):
    h, m, s = [int(x) for x in hms_text.split(":")]
    return h * 3600 + m * 60 + s


def short_key(entity, channel):
    return entity + "." + channel.split(".", 1)[1]


def rnd(unit, value):
    if isinstance(value, bool):
        return int(value)
    if not isinstance(value, (int, float)):
        return value
    if unit == "kW":
        return round(value, 1) if abs(value) < 100 else round(value)
    if unit in ("degC", "%RH"):
        return round(value, 1)
    if unit in ("mm/s", "m"):
        return round(value, 2)
    if unit == "on":
        return int(value)
    return round(value)


def scan_trace(path, snap_times, keep=None, watch=None, windows=()):
    """One pass over trace.jsonl. snap_times: seconds of day. watch(key): record value changes.
    windows: [(a, b)] count of records with a < t <= b. Returns snapshots, changes, total, window counts."""
    snaps, changes, last = {}, {}, {}
    pending = sorted(set(snap_times))
    counts = [0] * len(windows)
    total = 0
    latest = {}
    for line in open(path):
        d = json.loads(line)
        t = tsec(d["event_time"])
        while pending and t > pending[0]:
            snaps[pending[0]] = dict(latest)
            pending.pop(0)
        total += 1
        for i, (a, b) in enumerate(windows):
            if a < t <= b:
                counts[i] += 1
        key = short_key(d["entity_id"], d["channel"])
        if keep is not None and not keep(d["entity_id"], key):
            continue
        value = rnd(d.get("unit"), d["value"])
        latest[key] = value
        if watch is not None and watch(key) and last.get(key) != value:
            last[key] = value
            changes.setdefault(key, []).append((t, value))
    for t in pending:
        snaps[t] = dict(latest)
    return snaps, changes, total, counts


def truth_round(key, value):
    if key.endswith(("fan_health", "fan_eff", "fan_trim")):
        return round(value, 3)
    if key.endswith(("fan_speed", "cop", "approach")):
        return round(value, 2)
    return round(value, 3)


def scan_history(path, snap_times, wanted, track=None):
    """wanted: {entity: [state names]}. Returns snapshots keyed 'ENTITY.state' and the series of `track` keys."""
    snaps, series = {}, {}
    pending = sorted(set(snap_times))
    latest = {}
    needles = ['"entity_id":"%s"' % e for e in wanted]
    for line in open(path):
        if not any(n in line for n in needles):
            continue
        d = json.loads(line)
        ent = d["entity_id"]
        if ent not in wanted:
            continue
        t = d["time_ns"] / 1e9 % 86400
        while pending and t > pending[0]:
            snaps[pending[0]] = dict(latest)
            pending.pop(0)
        for name in wanted[ent]:
            if name in d["states"]:
                latest[ent + "." + name] = d["states"][name]
        for k in track or ():
            if k.split(".")[0] == ent:
                series.setdefault(k, []).append((t, d["states"][k.split(".", 1)[1]]))
    for t in pending:
        snaps[t] = dict(latest)
    return snaps, series


def ledger_counts(path):
    delivered = dropped = 0
    for line in open(path):
        if '"delivered":true' in line:
            delivered += 1
        else:
            dropped += 1
    return delivered, dropped


def alarm_times(run):
    watch = lambda k: k in ("AC-1.alarm_state", "AC-1.run_status")
    _, ch, _, _ = scan_trace(os.path.join(run, "trace.jsonl"), [], keep=lambda e, k: e == "AC-1", watch=watch)
    states = ch.get("AC-1.alarm_state", [])
    alarm = [t for t, v in states if v == 1]
    ack = [t for t, v in states if v == 2]
    back = [t for t, v in states if v in (0, 3) and alarm and t > alarm[0]]
    trip = [t for t, v in ch.get("AC-1.run_status", []) if v == 0]
    return {
        "alarm_s": alarm[0] if alarm else None,
        "ack_s": ack[0] if ack else None,
        "trip_s": trip[0] if trip else None,
        "return_s": back[0] if back else None,
    }


def world_id(run):
    with open(os.path.join(run, "trace.jsonl")) as f:
        return json.loads(f.readline())["world_id"]


def human(times):
    return {k.replace("_s", "_time"): (None if v is None else hms(v)) for k, v in times.items()}


FACTORY_ENTITIES = ["CT-1", "CH-1", "P-1", "P-2", "P-3", "P-4", "MB-AIR", "MB-COOL", "MB-PROC", "MM-1",
                    "R-1", "AC-1", "AC-2", "D-1", "WX-1", "CAL-1", "AHU-1", "M-1", "M-2", "M-3", "M-4"]
TRUTH_WANTED = {
    "CT-1": ["fan_health", "fan_eff", "fan_speed", "fan_trim", "approach"],
    "CH-1": ["cop"],
}
SERIES_KEYS = ["CT-1.basin_temp", "WX-1.wet_bulb_temp", "CT-1.fan_power", "MM-1.power",
               "R-1.header_pressure", "CH-1.chw_supply_temp"]


def factory_keep(entity, key):
    return entity in FACTORY_ENTITIES


def pick_series(snaps, keys, times):
    return {k: [[hms(t), snaps[t].get(k)] for t in times] for k in keys}


def truth_pack(snap):
    return {k: truth_round(k, v) for k, v in snap.items()}


def first_time(series, predicate):
    return next((t for t, v in series if predicate(v)), None)


def optional_hms(t):
    return None if t is None else hms(t)


def explore_section():
    run, twin = os.path.join(RUNS, "explore"), os.path.join(RUNS, "twin")
    pre, now = sec_of("09:39:50"), sec_of("10:25:00")
    series_t = [sec_of("09:25:00") + i * 5 * MIN for i in range(13)]
    snaps, _, total, _ = scan_trace(os.path.join(run, "trace.jsonl"), [pre, now] + series_t, factory_keep)
    tsnaps, _, ttotal, _ = scan_trace(os.path.join(twin, "trace.jsonl"), [now], factory_keep)
    hsnaps, hseries = scan_history(os.path.join(run, "world_state_history.jsonl"), [pre, now] + series_t,
                                   TRUTH_WANTED, track=["CT-1.fan_health"])
    series = pick_series(snaps, SERIES_KEYS, series_t)
    series["CH-1.cop_truth"] = [[hms(t), truth_round("cop", hsnaps[t]["CH-1.cop"])] for t in series_t]
    delivered, dropped = ledger_counts(os.path.join(run, "ledger.jsonl"))
    health = hseries["CT-1.fan_health"]
    floor = min(v for _, v in health)
    return {
        "prefault_time": hms(pre), "now": hms(now), "fault_onset": "09:40:00",
        "at": {hms(pre): snaps[pre], hms(now): snaps[now]},
        "twin_at": {hms(now): tsnaps[now]},
        "truth_at": {hms(pre): truth_pack(hsnaps[pre]), hms(now): truth_pack(hsnaps[now])},
        "fan_health_floor": {"value": round(floor, 3),
                             "time": hms(first_time(health, lambda v: v <= floor + 1e-9)),
                             "first_decline": hms(first_time(health, lambda v: v < 1))},
        "series": series,
        "counts": {"observed_at_now": total, "delivered": delivered, "dropped": dropped, "twin_observed": ttotal},
    }


def evaluate_section():
    probe, run = os.path.join(RUNS, "evaluate_probe"), os.path.join(RUNS, "evaluate")
    now = int(open(os.path.join(SCRATCH, "evaluate_now_s.txt")).read())
    series_t = [now - 3 * HOUR + (i + 1) * 15 * MIN for i in range(12)]
    snaps, _, total, _ = scan_trace(os.path.join(run, "trace.jsonl"), [now] + series_t, factory_keep)
    _, hseries = scan_history(os.path.join(run, "world_state_history.jsonl"), [now],
                              {"AC-1": ["healthy"]}, track=["AC-1.healthy"])
    delivered, dropped = ledger_counts(os.path.join(run, "ledger.jsonl"))
    out = human(alarm_times(probe))
    out.update({
        "now": hms(now),
        "seen_in_final_run": human(alarm_times(run)),
        "truth_trip_time": optional_hms(first_time(hseries.get("AC-1.healthy", []), lambda v: v == 0)),
        "at_now": snaps[now],
        "series": pick_series(snaps, ["AC-1.discharge_temp"], series_t),
        "counts": {"observed_at_now": total, "delivered": delivered, "dropped": dropped},
        "note": "alarm/trip/return times come from the longer probe run (same seed); the final run ends at now",
    })
    return out


def review_section():
    run = os.path.join(RUNS, "review")
    at = sec_of("09:45:00")
    snaps, _, total, _ = scan_trace(os.path.join(run, "trace.jsonl"), [at], factory_keep)
    times = alarm_times(run)
    hsnaps, hseries = scan_history(os.path.join(run, "world_state_history.jsonl"), [at],
                                   dict(TRUTH_WANTED, **{"AC-1": ["healthy"]}), track=["AC-1.healthy"])
    delivered, dropped = ledger_counts(os.path.join(run, "ledger.jsonl"))
    ack_cmd = int(open(os.path.join(SCRATCH, "review_ack_s.txt")).read())
    return {
        "report": json.load(open(os.path.join(SCRATCH, "review.report.json"))),
        "verify": json.load(open(os.path.join(SCRATCH, "review.verify.json"))),
        "alarm_time": hms(times["alarm_s"]), "ack_time": hms(times["ack_s"]), "trip_time": hms(times["trip_s"]),
        "return_time": optional_hms(times["return_s"]),
        "truth_trip_time": optional_hms(first_time(hseries["AC-1.healthy"], lambda v: v == 0)),
        "ack_commanded_at": hms(ack_cmd),
        "at_0945": snaps[at],
        "truth_0945": truth_pack({k: v for k, v in hsnaps[at].items() if not k.startswith("AC-1")}),
        "counts": {"observed": total, "delivered": delivered, "dropped": dropped},
    }


def campus_section():
    run = os.path.join(RUNS, "campus")
    now = sec_of("09:12:30")
    halls = ["%02d" % h for h in range(1, 27)]
    detail = {"07", "19"}

    def keep(entity, key):
        if "/" not in entity:
            return entity == "WX-1"
        hall, name = entity.split("/", 1)
        return name in ("MM-1", "AC-1") or hall[1:] in detail

    watch = lambda k: k == "h07/AC-1.alarm_state"
    snaps, changes, total, counts = scan_trace(os.path.join(run, "trace.jsonl"), [now], keep, watch,
                                               [(now - 600, now)])
    snap = snaps[now]
    rows = {h: {"power": snap.get("h%s/MM-1.power" % h), "alarm": snap.get("h%s/AC-1.alarm_state" % h),
                "fault": h in detail} for h in halls}

    def view(h, names):
        return {n: snap.get("h%s/%s" % (h, n)) for n in names}

    alarm_t = [t for t, v in changes.get("h07/AC-1.alarm_state", []) if v == 1]
    h07 = view("07", ["R-1.header_pressure", "CH-1.chw_supply_temp", "MB-COOL.power", "MM-1.power",
                      "AC-1.discharge_temp", "AC-1.alarm_state", "AC-1.run_status", "AC-1.power"])
    h07["alarm_time"] = optional_hms(alarm_t[0] if alarm_t else None)
    h19 = view("19", ["CT-1.basin_temp", "CT-1.fan_power", "CH-1.chw_supply_temp", "MB-COOL.power",
                      "MM-1.power", "AC-1.alarm_state"])
    h19["WX-1.wet_bulb_temp"] = snap.get("WX-1.wet_bulb_temp")
    spec = json.load(open(os.path.join(SCRATCH, "dom", "campus.domain.json")))
    channels = {k["name"]: len(k["channels"]) for k in spec["kinds"]}
    return {
        "now": hms(now), "halls": rows, "hall07": h07, "hall19": h19,
        "counts": {"observed": total, "per_min_last_10": counts[0] / 10.0},
        "instances": sum(len(i["names"]) for i in spec["instances"]),
        "tags": sum(len(i["names"]) * channels[i["kind"]] for i in spec["instances"]),
        "faults": {"h07/AC-1": "oil_cooler_fouling onset 07:30:00", "h19/CT-1": "fan_failure onset 07:20:00"},
        "run_window": {"start": "04:00:00", "end": "09:12:30"},
    }


def main():
    if len(sys.argv) >= 3 and sys.argv[1] == "worldid":
        print(world_id(sys.argv[2]))
        return
    if len(sys.argv) >= 3 and sys.argv[1] == "alarms":
        print(json.dumps(alarm_times(sys.argv[2])))
        return
    data = {
        "explore": explore_section(),
        "evaluate": evaluate_section(),
        "review": review_section(),
        "campus": campus_section(),
    }
    with open(OUT, "w") as f:
        json.dump(data, f, indent=1)
        f.write("\n")
    print("wrote", OUT)


if __name__ == "__main__":
    main()
