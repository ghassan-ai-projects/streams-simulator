import { describe, expect, it } from "vitest";
import { report } from "../test/runFixture";
import { badgeLabels, badgeOf, parseReport, runSpan, verifyMatches, wallClock } from "./report";

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
    expect(parsed.disturbances.map((d) => [d.atNs / 1e9, d.what, d.target, d.detail])).toEqual([
      [300, "perturbation apply", "delivery", "drop"],
      [600, "fault inject", "CT-1", "tower__fan_failure"],
      [900, "environment", "weather", "heat_wave"],
    ]);
    expect(parsed.effectorCalls.map((call) => call.outcome)).toEqual([
      "applied · 300 ms",
      "refused: interlock",
      "accepted, no effect · silent_no_effect",
      "refused",
    ]);
    expect(parsed.kpis).toEqual([{ entity: "CH-1", state: "cop", mean: 5.95, min: 4.65, max: 8.85 }]);
    expect(parsed.alarms).toEqual({ activations: 1, standing_at_end: 0 });
    expect(parsed.scoring).toBe("not scored");
  });

  it("keeps what a scored or truthless report says", () => {
    const parsed = parseReport({ scoring: { status: "scored", reason: "verdict r-1" }, truth_unavailable: "sealed" });
    expect(parsed.scoring).toBe("scored · verdict r-1");
    expect(parsed.truthUnavailable).toBe("sealed");
    expect(parseReport({ scoring: { status: "failed" } }).scoring).toBe("failed");
    expect(parseReport(null).disturbances).toEqual([]);
  });

  it("derives the badge from the report and verify, never from its own checks", () => {
    const parsed = parseReport(report);
    expect(badgeOf(parsed, true)).toBe("verified");
    expect(badgeOf(parsed, null)).toBe("verified");
    expect(badgeOf(parsed, false)).toBe("not-verified");
    expect(badgeOf({ ...parsed, verification: "files_differ" }, null)).toBe("not-verified");
    expect(badgeOf({ ...parsed, complete: false }, true)).toBe("incomplete");
    expect(badgeLabels["not-verified"]).toBe("Not verified");
    expect(verifyMatches({ matches: true })).toBe(true);
    expect(verifyMatches({ matches: "yes" })).toBe(false);
  });

  it("spans the run from the report, or from the observations while the report is missing", () => {
    expect(runSpan({ startNs: 0, endNs: 10 }, { startNs: 2, endNs: 8 })).toEqual({ startNs: 0, endNs: 10 });
    expect(runSpan({ startNs: 0, endNs: 0 }, { startNs: 2, endNs: 8 })).toEqual({ startNs: 2, endNs: 8 });
    expect(runSpan(null, null)).toEqual({ startNs: 0, endNs: 0 });
  });

  it("prints simulated time of day in UTC", () => {
    expect(wallClock(1_767_252_480e9)).toBe("07:28:00");
  });
});
