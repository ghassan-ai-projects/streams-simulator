#!/usr/bin/env bash
# Runs the real streamsim binary for every scenario the mockups display.
# Usage: bash tools/runs.sh      (then: python3 tools/extract.py)
# Env: SCRATCH (work dir), ONLY="explore twin evaluate review campus" (subset), SKIP_EXISTING=1.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$HERE/../../../.." && pwd)"
SCRATCH="${SCRATCH:-/private/tmp/claude-501/-Users-ghassan-my-projects-streams-simulator/0abb3216-3a51-47d8-8539-36e4038c8aa8/scratchpad/mockdata}"
ONLY="${ONLY:-explore twin evaluate review campus}"
BIN="$SCRATCH/bin/streamsim"
DOM="$SCRATCH/dom"
COMP="$SCRATCH/comp"
RUNS="$SCRATCH/runs"
TIMINGS="$SCRATCH/timings.txt"
EXTRACT="$HERE/extract.py"

START_04=1767240000000000000

cd "$ROOT"
mkdir -p "$SCRATCH/bin" "$DOM" "$COMP" "$RUNS"
: > "$TIMINGS"
: > "$SCRATCH/workspace_map.txt"

wants() { [[ " $ONLY " == *" $1 "* ]]; }

timed() {
  local name="$1"; shift
  local t0=$SECONDS
  "$@" > "$SCRATCH/$name.out"
  echo "$name $((SECONDS - t0))s" | tee -a "$TIMINGS"
}

run() {
  local name="$1"; shift
  if [[ "${SKIP_EXISTING:-}" == 1 && -f "$RUNS/$name/run.json" ]]; then echo "$name skipped"; return; fi
  rm -rf "$RUNS/$name"
  timed "$name" "$BIN" run -domains-dir "$DOM" -out "$RUNS/$name" "$@"
}

alarms() { python3 "$EXTRACT" alarms "$RUNS/$1"; }
field() { python3 -c "import json,sys; print(int(float(json.loads(sys.argv[1])[sys.argv[2]])))" "$1" "$2"; }

echo "== build"
go build -o "$BIN" ./cmd/streamsim

echo "== domains"
"$BIN" domain compose examples/compositions/factory-utilities.composition.json --out "$DOM/factory-utilities.domain.json" > /dev/null
python3 - "$ROOT/examples/compositions/factory-utilities.composition.json" "$COMP/campus.composition.json" <<'PY'
import json, sys

src, dst, halls = sys.argv[1], sys.argv[2], 26
shared = {"WX-1", "CAL-1"}


def name(h, n):
    return n if n in shared else "h%02d/%s" % (h, n)


def endpoint(h, text):
    cut = text.rindex(".")
    return name(h, text[:cut]) + text[cut:]


doc = json.load(open(src))
doc["title"] = "%s campus of %d halls" % (doc["title"], halls)
doc["id"] = "%s-campus-%d" % (doc["id"], halls)
for entry in doc["equipment"]:
    out = []
    for h in range(1, halls + 1):
        for n in entry["names"]:
            r = name(h, n)
            if h == 1 or r != n:
                out.append(r)
    entry["names"] = out
conns = []
for c in doc["connections"]:
    for h in range(1, halls + 1):
        f, t = endpoint(h, c["from"]), endpoint(h, c["to"])
        if h == 1 or f != c["from"] or t != c["to"]:
            conns.append({"from": f, "to": t})
doc["connections"] = conns
json.dump(doc, open(dst, "w"))
PY
"$BIN" domain compose "$COMP/campus.composition.json" --out "$DOM/campus.domain.json" > /dev/null

AC_FAULT="compressor_lead__oil_cooler_fouling"
CT_FAULT="cooling_tower__fan_failure"

if wants explore; then
  echo "== A explore"
  run explore -domain factory-utilities -seed 7 -explore -duration 37500 -fault "CT-1=${CT_FAULT}@34800"
fi

if wants twin; then
  echo "== B twin"
  run twin -domain factory-utilities -seed 7 -explore -duration 37500
fi

if wants evaluate; then
  echo "== C evaluate (probe finds alarm, trip, return)"
  run evaluate_probe -domain factory-utilities -seed 11 -duration 30000 -fault "AC-1=${AC_FAULT}@21600"
  A="$(alarms evaluate_probe)"
  echo "alarms: $A"
  ALARM_S="$(field "$A" alarm_s)"; TRIP_S="$(field "$A" trip_s)"
  NOW_S=$(( (ALARM_S + TRIP_S) / 2 / 60 * 60 ))
  echo "now = ${NOW_S}s (halfway between alarm ${ALARM_S}s and trip ${TRIP_S}s, rounded down to the minute)"
  echo "$NOW_S" > "$SCRATCH/evaluate_now_s.txt"
  run evaluate -domain factory-utilities -seed 11 -duration "$NOW_S" -fault "AC-1=${AC_FAULT}@21600"
fi

if wants review; then
  echo "== D review (probe finds the alarm, final run acknowledges it 150 s later)"
  run review_probe -domain factory-utilities -seed 7 -duration 27000 \
    -fault "AC-1=${AC_FAULT}@21600" -fault "CT-1=${CT_FAULT}@34800" -perturb "drop@33600@34200"
  A="$(alarms review_probe)"
  echo "alarms: $A"
  ACK_S=$(( $(field "$A" alarm_s) + 150 ))
  echo "$ACK_S" > "$SCRATCH/review_ack_s.txt"
  run review -domain factory-utilities -seed 7 -duration 64800 \
    -fault "AC-1=${AC_FAULT}@21600" -fault "CT-1=${CT_FAULT}@34800" -perturb "drop@33600@34200" \
    -effector "compressor_lead__acknowledge_reset@AC-1@${ACK_S}@{\"compressor_id\":\"AC-1\"}"
  timed review.report "$BIN" report "$RUNS/review"
  cp "$SCRATCH/review.report.out" "$SCRATCH/review.report.json"
  set +e
  timed review.verify "$BIN" verify -domains-dir "$DOM" -adapters-dir adapters "$RUNS/review/run.json"
  set -e
  cp "$SCRATCH/review.verify.out" "$SCRATCH/review.verify.json"
fi

if wants campus; then
  echo "== E campus (04:00 start so h07 AC-1 alarm is active at 09:12:30)"
  run campus -domain factory-utilities-campus-26 -seed 3 -explore -start-time "$START_04" -duration 18750 \
    -fault "h07/AC-1=${AC_FAULT}@12600" -fault "h19/CT-1=${CT_FAULT}@12000"
fi

echo "== workspace"
WS="$SCRATCH/workspace"
rm -rf "$WS"
mkdir -p "$WS/domains" "$WS/compositions" "$WS/runs"
cp "$DOM/factory-utilities.domain.json" "$WS/domains/factory-utilities.domain.json"
cp "$DOM/campus.domain.json" "$WS/domains/factory-utilities-campus-26.domain.json"
cp examples/compositions/factory-utilities.composition.json "$WS/compositions/factory-utilities.composition.json"
for name in review evaluate explore twin campus; do
  [[ -f "$RUNS/$name/run.json" ]] || continue
  wid="$(python3 "$EXTRACT" worldid "$RUNS/$name")"
  dest="$WS/runs/$wid"
  [[ -e "$dest" ]] && dest="$WS/runs/$wid-$name"
  cp -cR "$RUNS/$name" "$dest" 2> /dev/null || cp -R "$RUNS/$name" "$dest"
  echo "$name -> runs/$(basename "$dest")" | tee -a "$SCRATCH/workspace_map.txt"
done

echo "== bridge check"
mkdir -p "$SCRATCH/ui" && echo '<!doctype html><title>ui</title>' > "$SCRATCH/ui/index.html"
"$BIN" canvas --workspace "$WS" --ui-dir "$SCRATCH/ui" --addr 127.0.0.1:0 2> "$SCRATCH/bridge.err" > /dev/null &
BRIDGE_PID=$!
trap 'kill "$BRIDGE_PID" 2> /dev/null || true' EXIT
for _ in $(seq 1 50); do grep -q "pairing code" "$SCRATCH/bridge.err" 2> /dev/null && break; sleep 0.2; done
URL="$(sed -n 's/^canvas: \(http.*\)$/\1/p' "$SCRATCH/bridge.err" | head -1)"
CODE="$(sed -n 's/^canvas: pairing code \(.*\)$/\1/p' "$SCRATCH/bridge.err")"
curl -s -c "$SCRATCH/jar" -H "Content-Type: application/json" -d "{\"code\":\"$CODE\"}" "$URL/api/session" > /dev/null
curl -s -b "$SCRATCH/jar" "$URL/api/workspace" | python3 -m json.tool > "$SCRATCH/workspace.api.json"
cat "$SCRATCH/workspace.api.json"
kill "$BRIDGE_PID" 2> /dev/null || true

echo "== done; timings:"
cat "$TIMINGS"
