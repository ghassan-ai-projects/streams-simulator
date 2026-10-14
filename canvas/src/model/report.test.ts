import { describe, expect, it } from "vitest";
import { report } from "../test/runFixture";
import { kpiGroups, kpisOf, kpiWords, rerunOf, runFacts, runIdentity } from "./reportViews";
import { badgeLabels, badgeOf, humanise, parseReport, runSpan, parseReplay, replayLine, wallClock } from "./report";

describe("run report", () => {
  it("reads the run, counts, verification and truth sections", () => {
    const parsed = parseReport(report);
    expect(parsed).toMatchObject({
      runId: "r-1",
      mode: "evaluate",
      complete: true,
      domainId: "mini",
      adapterId: "native-jsonl",
      seed: 7,
      verification: "files_match",
    });
    expect(parsed.counts).toEqual({ emitted: 12, delivered: 10, dropped: 2 });
    expect(parsed.dropped).toEqual({ dropped_by_perturbation: 2 });
    expect(
      parsed.disturbances.map((d) => [d.onsetNs / 1e9, d.injectedNs / 1e9, d.kind, d.action, d.target, d.name]),
    ).toEqual([
      [300, 0, "perturbation", "apply", "delivery", "drop"],
      [600, 0, "fault", "inject", "CT-1", "tower__fan_failure"],
      [900, 900, "environment", "inject", "weather", "heat_wave"],
    ]);
    expect(parsed.effectorCalls.map((call) => call.outcome)).toEqual([
      "applied · 300 ms",
      "refused: interlock",
      "accepted, no effect · silent_no_effect",
      "refused",
    ]);
    expect(parsed.kpis).toEqual([{ entity: "CH-1", state: "cop", mean: 5.95, min: 4.65, max: 8.85 }]);
    expect(parsed.alarms).toEqual({ activations: 1, standing_at_end: 0 });
    expect(parsed.scoring).toBeNull();
    expect(parsed.ended).toBe(true);
  });

  it("keeps what a scored or truthless report says", () => {
    const parsed = parseReport({ scoring: { status: "scored", reason: "verdict r-1" }, truth_unavailable: "sealed" });
    expect(parsed.scoring).toEqual({ status: "scored", reason: "verdict r-1" });
    expect(parsed.truthUnavailable).toBe("sealed");
    expect(parseReport({ scoring: { status: "failed" } }).scoring).toEqual({ status: "failed", reason: "" });
    expect(parseReport(null).disturbances).toEqual([]);
  });

  it("derives the badge from the report and verify, never from its own checks", () => {
    const parsed = parseReport(report);
    const ready = (matches: boolean) =>
      ({ phase: "ready", value: { matches, domainChanged: false, detail: "", reason: "", file: "" } }) as const;
    expect(badgeOf(parsed, ready(true))).toBe("verified");
    expect(badgeOf(parsed, { phase: "loading" })).toBe("verifying");
    expect(badgeOf(parsed, { phase: "failed" })).toBe("not-checked");
    expect(badgeOf(parsed, ready(false))).toBe("not-verified");
    expect(badgeOf({ ...parsed, verification: "files_differ" }, { phase: "loading" })).toBe("not-verified");
    expect(badgeOf({ ...parsed, verification: "not_checked" }, ready(true))).toBe("not-checked");
    expect(badgeOf({ ...parsed, complete: false }, ready(true))).toBe("incomplete");
    expect(badgeOf({ ...parsed, ended: false }, ready(true))).toBe("incomplete");
    expect(badgeLabels["not-verified"]).toBe("Not verified");
    expect(parseReplay({ matches: true, reason: "match" })).toEqual({
      matches: true,
      domainChanged: false,
      detail: "",
      reason: "match",
      file: "",
    });
    expect(parseReplay({ matches: false, domain_changed: true, detail: "trace differs", file: "trace.jsonl" })).toEqual(
      {
        matches: false,
        domainChanged: true,
        detail: "trace differs",
        reason: "",
        file: "trace.jsonl",
      },
    );
    expect(parseReplay({ matches: "yes" })).toBeNull();
    expect(parseReplay(null)).toBeNull();
    expect(replayLine({ phase: "loading" })).toMatch(/^Replaying/);
    expect(replayLine({ phase: "failed" })).toMatch(/^Not checked/);
    expect(replayLine(ready(true))).toMatch(/^Reproduces/);
    expect(replayLine(ready(false))).toMatch(/^Does not verify: streamsim verify reported/);
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
  });

  it("filters KPI truth to an instance, or to every instance of a hall", () => {
    const rows = ["h03/M-1", "h03/M-2", "h13/M-1", "CH-1"].map((entity) => ({
      entity,
      state: "s",
      mean: 1,
      min: 0,
      max: 2,
    }));
    const report = { ...parseReport({}), kpis: rows };
    expect(kpisOf(report, null)).toHaveLength(4);
    expect(kpisOf(report, "h03").map((row) => row.entity)).toEqual(["h03/M-1", "h03/M-2"]);
    expect(kpisOf(report, "CH-1").map((row) => row.entity)).toEqual(["CH-1"]);
  });

  it("names the run by its plant, seed, adapter and simulator", () => {
    const parsed = parseReport({
      run: {
        domain: { id: "factory", version: "0.1.0" },
        adapter: { id: "native-jsonl" },
        seed: 7,
        sim_version: "0.3.1",
      },
    });
    expect(runIdentity(parsed)).toBe("factory 0.1.0 · seed 7 · native-jsonl · sim 0.3.1");
    expect(runIdentity(parseReport({ run: { domain: { id: "factory" } } }))).toBe("factory");
  });

  it("says a run's facts in words and plans the same run again", () => {
    const parsed = parseReport(report);
    expect(runFacts(parsed)).toBe("seed 7 · 1 h");
    expect(rerunOf(parsed).plan).toEqual({
      seed: 7,
      durationS: 3600,
      explore: false,
      faults: [{ entity: "CT-1", fault: "tower__fan_failure", atS: 600 }],
      calls: [],
    });
    expect(rerunOf(parsed).left).toEqual(["1 delivery perturbation", "1 environment change", "4 effector calls"]);
    const cleared = parsed.disturbances.map((d) => (d.kind === "fault" ? { ...d, clearedNs: 900e9 } : d));
    expect(rerunOf({ ...parsed, disturbances: cleared, effectorCalls: [] }).left).toContain(
      "1 fault clear (the fault stays on)",
    );
    expect(rerunOf(parseReport(null))).toEqual({
      plan: { seed: 1, durationS: 60, explore: false, faults: [], calls: [] },
      left: [
        "the seed, which the run did not record (seed 1 is used)",
        "a length under one minute (one minute is used)",
      ],
    });
    expect(runFacts(parseReport(null))).toBe("");
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
