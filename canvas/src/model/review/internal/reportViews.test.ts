import { describe, expect, it } from "vitest";
import { fileLine, kpiFigure, kpiSteady } from "./reportViews";
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
