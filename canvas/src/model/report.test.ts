import { describe, expect, it } from "vitest";
import { report } from "../test/runFixture";
import { badgeLabels, badgeOf, humanise, parseReport, runSpan, verifyMatches, wallClock } from "./report";

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
    const ready = (value: boolean) => ({ phase: "ready", value }) as const;
    expect(badgeOf(parsed, ready(true))).toBe("verified");
    expect(badgeOf(parsed, { phase: "loading" })).toBe("verifying");
    expect(badgeOf(parsed, { phase: "failed" })).toBe("not-checked");
    expect(badgeOf(parsed, ready(false))).toBe("not-verified");
    expect(badgeOf({ ...parsed, verification: "files_differ" }, { phase: "loading" })).toBe("not-verified");
    expect(badgeOf({ ...parsed, verification: "not_checked" }, ready(true))).toBe("not-checked");
    expect(badgeOf({ ...parsed, complete: false }, ready(true))).toBe("incomplete");
    expect(badgeOf({ ...parsed, ended: false }, ready(true))).toBe("incomplete");
    expect(badgeLabels["not-verified"]).toBe("Not verified");
    expect(verifyMatches({ matches: true })).toBe(true);
    expect(verifyMatches({ matches: "yes" })).toBe(false);
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
});
