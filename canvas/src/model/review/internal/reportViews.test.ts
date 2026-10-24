import { describe, expect, it } from "vitest";
import { fileLine, kpiFigure, kpiSteady, terminationWords, unfinishedTruthWords } from "./reportViews";
import { report } from "../../../test/runFixture";

describe("file line", () => {
  it("says each file-check state in words a reader need not decode", () => {
    expect(fileLine(report)).toBe("Files match the record.");
    expect(fileLine({ ...report, verification: "files_differ" })).toBe("The files changed after the run.");
    expect(fileLine({ ...report, verification: "" })).toBe("The run recorded no file digests.");
  });
});

describe("kpi figures", () => {
  it("keeps more decimals for smaller values", () => {
    expect([123.456, 12.345, 1.2345, -123.4, -12.34, 0].map(kpiFigure)).toEqual([
      "123",
      "12.3",
      "1.23",
      "-123",
      "-12.3",
      "0.00",
    ]);
  });

  it("calls a row steady when its minimum and maximum read the same", () => {
    const row = { entity: "CH-1", state: "cop", mean: 5, min: 4.999, max: 5.001 };
    expect(kpiSteady(row)).toBe(true);
    expect(kpiSteady({ ...row, max: 5.2 })).toBe(false);
  });
});

describe("how a run ended", () => {
  const stopped = { reason: "stopped_by_operator", atNs: 1_680e9 } as const;

  it("says where an operator stopped the run, out of what was requested", () => {
    const run = { ...report, termination: stopped, requestedEndNs: 3_600e9 };
    expect(terminationWords(run)).toBe("Stopped by operator at 00:28:00 of 01:00:00");
    expect(terminationWords({ ...run, startNs: 600e9 })).toBe("Stopped by operator at 00:18:00 of 00:50:00");
  });

  it("says where it stopped even when the request was not recorded", () => {
    expect(terminationWords({ ...report, termination: stopped })).toBe("Stopped by operator at 00:28:00");
  });

  it("says nothing of a run that completed or recorded no ending", () => {
    expect(terminationWords({ ...report, termination: { reason: "completed", atNs: 3_600e9 } })).toBe("");
    expect(terminationWords(report)).toBe("");
  });

  it("warns that the true results of a run that did not finish are not a finished run's result", () => {
    expect(unfinishedTruthWords).toBe(
      "This run did not finish. These are the plant's recorded values up to where it stopped, not the result of a completed run.",
    );
  });
});
