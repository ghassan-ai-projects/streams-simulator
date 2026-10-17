import { describe, expect, it } from "vitest";
import { blankReport, report } from "../../../test/runFixture";
import { kpiGroups, kpisOf, kpiWords, rerunOf, runFacts, runIdentity } from "./reportViews";
import { humanise, kindLabel } from "../../shared/words";
import { badgeLabels, badgeOf, replayLine, runSpan, wallClock } from "./report";

describe("run report", () => {
  it("derives the badge from the report and verify, never from its own checks", () => {
    const ready = (matches: boolean) =>
      ({ phase: "ready", value: { matches, domainChanged: false, detail: "", reason: "", file: "" } }) as const;
    expect(badgeOf(report, ready(true))).toBe("verified");
    expect(badgeOf(report, { phase: "loading" })).toBe("verifying");
    expect(badgeOf(report, { phase: "failed" })).toBe("not-checked");
    expect(badgeOf(report, ready(false))).toBe("not-verified");
    expect(badgeOf({ ...report, verification: "files_differ" }, { phase: "loading" })).toBe("not-verified");
    expect(badgeOf({ ...report, verification: "not_checked" }, ready(true))).toBe("not-checked");
    expect(badgeOf({ ...report, complete: false }, ready(true))).toBe("incomplete");
    expect(badgeOf({ ...report, ended: false }, ready(true))).toBe("incomplete");
    expect(badgeLabels["not-verified"]).toBe("Not verified");
    expect(replayLine({ phase: "loading" })).toMatch(/^Replaying/);
    expect(replayLine({ phase: "failed" })).toMatch(/^Not checked: this run could not be replayed/);
    expect(replayLine(ready(true))).toMatch(/^Reproduces/);
    expect(replayLine(ready(false))).toMatch(/^Does not verify: replaying the run gave a different/);
    const because = (reason: string, file = "") =>
      replayLine({ phase: "ready", value: { matches: false, domainChanged: false, detail: "", reason, file } });
    expect(because("file_missing", "ledger.jsonl")).toBe(
      "Does not verify: ledger.jsonl is missing from the run folder.",
    );
    expect(because("trace_divergence")).toMatch(/^Does not reproduce: .* different trace/);
    expect(because("recorded_spec")).toMatch(/nothing was replayed/);
    expect(because("incomplete")).toMatch(/ended incomplete/);
    expect(because("incomplete_stamp", "journal.jsonl")).toMatch(/journal.jsonl contradicts/);
  });

  it("spans the run from the report, or from the observations while the report is missing", () => {
    expect(runSpan({ startNs: 0, endNs: 10 }, { startNs: 2, endNs: 8 })).toEqual({ startNs: 0, endNs: 10 });
    expect(runSpan({ startNs: 0, endNs: 0 }, { startNs: 2, endNs: 8 })).toEqual({ startNs: 2, endNs: 8 });
    expect(runSpan(null, null)).toEqual({ startNs: 0, endNs: 0 });
  });

  it("prints simulated time of day in UTC, with the day after the first", () => {
    expect(wallClock(1_767_252_480e9)).toBe("07:28:00");
    expect(wallClock(1_767_252_480e9 + 86_400e9, 1_767_252_480e9)).toBe("+1d 07:28:00");
  });

  it("turns library ids into words and keeps the instrument", () => {
    expect(humanise("compressor_lead__oil_cooler_fouling")).toBe("oil cooler fouling");
    expect(humanise("cooling_tower__basin_temp__offset")).toBe("basin temp · offset");
    expect(humanise("discharge_temp")).toBe("discharge temp");
    expect(kindLabel("air_compressor_2")).toBe("air compressor");
    expect(kindLabel("compressor_lead__air_compressor_2")).toBe("air compressor");
    expect(kindLabel("chiller")).toBe("chiller");
  });

  it("filters KPI truth to an instance, or to every instance of a hall", () => {
    const rows = ["h03/M-1", "h03/M-2", "h13/M-1", "CH-1"].map((entity) => ({
      entity,
      state: "s",
      mean: 1,
      min: 0,
      max: 2,
    }));
    const table = { ...blankReport(), kpis: rows };
    expect(kpisOf(table, null)).toHaveLength(4);
    expect(kpisOf(table, "h03").map((row) => row.entity)).toEqual(["h03/M-1", "h03/M-2"]);
    expect(kpisOf(table, "CH-1").map((row) => row.entity)).toEqual(["CH-1"]);
  });

  it("names the run by its plant, seed, adapter and simulator", () => {
    const named = blankReport({
      domainId: "factory",
      domainVersion: "0.1.0",
      adapterId: "native-jsonl",
      seed: 7,
      simVersion: "0.3.1",
    });
    expect(runIdentity(named)).toBe("factory 0.1.0 · seed 7 · native-jsonl · sim 0.3.1");
    expect(runIdentity(blankReport({ domainId: "factory" }))).toBe("factory");
  });

  it("says how long the plant ran, not a placeholder, and when its clock started", () => {
    const identity = (seconds: number) =>
      runIdentity(
        blankReport({
          domainId: "factory",
          domainVersion: "0.1.0",
          adapterId: "native-jsonl",
          seed: 7,
          startNs: 0,
          endNs: seconds * 1e9,
        }),
      );
    expect(identity(1800)).toBe("factory 0.1.0 · seed 7 · native-jsonl · 30 min of plant time from 00:00:00 UTC");
    expect(identity(5400)).toContain("1 h 30 min of plant time");
    expect(identity(0)).toBe("factory 0.1.0 · seed 7 · native-jsonl");
  });

  it("says a run's facts in words and plans the same run again", () => {
    expect(runFacts(report)).toBe("seed 7 · 1 h");
    expect(rerunOf(report).plan).toEqual({
      seed: 7,
      durationS: 3600,
      explore: false,
      faults: [{ entity: "CT-1", fault: "tower__fan_failure", atS: 600 }],
      calls: [],
    });
    expect(rerunOf(report).left).toEqual(["1 delivery perturbation", "1 environment change", "4 effector calls"]);
    const cleared = report.disturbances.map((d) => (d.kind === "fault" ? { ...d, clearedNs: 900e9 } : d));
    expect(rerunOf({ ...report, disturbances: cleared, effectorCalls: [] }).left).toContain(
      "1 fault clear (the fault stays on)",
    );
    expect(rerunOf(blankReport())).toEqual({
      plan: { seed: 1, durationS: 60, explore: false, faults: [], calls: [] },
      left: [
        "the seed, which the run did not record (seed 1 is used)",
        "a length under one minute (one minute is used)",
      ],
    });
    expect(runFacts(blankReport())).toBe("");
  });

  it("keeps what a scored or truthless report says", () => {
    expect(blankReport({ scoring: { status: "scored", reason: "verdict r-1" } }).scoring).toEqual({
      status: "scored",
      reason: "verdict r-1",
    });
    expect(blankReport({ truthUnavailable: "sealed" }).truthUnavailable).toBe("sealed");
    expect(blankReport({ scoring: { status: "failed", reason: "" } }).scoring).toEqual({
      status: "failed",
      reason: "",
    });
    expect(blankReport().disturbances).toEqual([]);
  });

  it("names KPIs in words and groups them by instance in report order", () => {
    expect(["scrap_pct", "load_frac", "cop", "specific_power"].map(kpiWords)).toEqual([
      "Scrap",
      "Load fraction",
      "COP",
      "Specific power",
    ]);
    const row = (entity: string, state: string) => ({ entity, state, mean: 1, min: 1, max: 1 });
    const groups = kpiGroups([row("M-1", "a"), row("M-1", "b"), row("M-2", "a")]);
    expect(groups.map((group) => [group.entity, group.rows.length])).toEqual([
      ["M-1", 2],
      ["M-2", 1],
    ]);
  });
});
