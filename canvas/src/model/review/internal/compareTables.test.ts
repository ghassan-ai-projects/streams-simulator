import { describe, expect, it } from "vitest";
import { comparison } from "../../../test/compareFixture";
import {
  alarmSystemTable,
  alarmTable,
  kpiTable,
  scopeCaption,
  signedFigure,
  signedPercent,
  windowedWords,
} from "./compareTables";

const units = (_entity: string, state: string) => ({ cop: "ratio", fan_pct: "pct", approach_c: "degC" })[state] ?? "";

describe("the scope of a comparison", () => {
  it("says the numbers cover each run's whole simulated length", () => {
    expect(scopeCaption({ aLengthNs: 1800e9, bLengthNs: 1680e9 })).toBe(
      "Whole run: reference 30:00, this run 28:00 simulated",
    );
    expect(scopeCaption({ aLengthNs: 3_725e9, bLengthNs: 60e9 })).toBe(
      "Whole run: reference 1:02:05, this run 01:00 simulated",
    );
  });

  it("says when a run recorded no whole-run results", () => {
    expect(scopeCaption({ aLengthNs: 1800e9, bLengthNs: null })).toBe(
      "Whole run: reference 30:00 simulated; this run recorded no whole-run results",
    );
  });

  it("says why event windows cannot be compared, in the binary's words", () => {
    expect(windowedWords(comparison)).toBe(
      "Event-window comparisons are not available: the run report holds whole-run KPI truth only, so no interval statistic is computed.",
    );
  });
});

describe("a difference keeps its sign", () => {
  it.each([
    [-0.69, "−0.69"],
    [0.69, "+0.69"],
    [0, "0"],
    [-0.0004, "−0.00040"],
    [0.0004, "+0.00040"],
    [-12.34, "−12.3"],
    [1234.5, "+1235"],
  ])("writes %s as %s", (delta, shown) => {
    expect(signedFigure(delta)).toBe(shown);
  });

  it.each([
    [-0.1356, "−13.6 %"],
    [0.0001, "+0.01 %"],
    [2.5, "+250 %"],
    [0, "0 %"],
    [null, "–"],
  ])("writes the relative difference %s as %s", (relative, shown) => {
    expect(signedPercent(relative)).toBe(shown);
  });
});

describe("the whole-run KPI table", () => {
  const table = kpiTable(comparison, units);

  it("is captioned with the scope of the numbers", () => {
    expect(table.caption).toBe("Whole run: reference 30:00, this run 28:00 simulated");
  });

  it("shows the reference, this run and the binary's difference with the unit symbol", () => {
    expect(table.rows[0]).toEqual({
      key: "CH-1/cop",
      entity: "CH-1",
      kpi: "COP",
      unit: "",
      reference: "5.09",
      thisRun: "4.40",
      delta: "−0.69",
      relative: "−13.6 %",
      note: "",
    });
    expect(table.rows[2]).toMatchObject({ unit: "°C", delta: "+0.00040", relative: "+0.01 %" });
  });

  it("marks a row only one run has and shows no difference for it", () => {
    expect(table.rows[1]).toMatchObject({
      kpi: "Fan",
      unit: "%",
      reference: "60.0",
      thisRun: "–",
      delta: "–",
      relative: "–",
      note: "Only in reference",
    });
    const flipped = comparison.kpis.slice(1, 2).map((row) => ({ ...row, presence: "only_in_b" as const }));
    const other = kpiTable({ ...comparison, kpis: flipped }, units);
    expect(other.rows[0]?.note).toBe("Only in this run");
  });

  it("takes its caption from the alarm system when there are no KPI rows", () => {
    expect(kpiTable({ ...comparison, kpis: [] }, units)).toMatchObject({
      caption: "Whole run: reference 30:00, this run 28:00 simulated",
      rows: [],
    });
  });

  it("never says which run is better", () => {
    const words = JSON.stringify([table, alarmTable(comparison), alarmSystemTable(comparison)]);
    expect(words).not.toMatch(/better|worse|winner|improv|degrad|best/i);
  });
});

describe("the alarm tables", () => {
  it("shows activations and time active for each run with the binary's differences", () => {
    expect(alarmTable(comparison).rows).toEqual([
      {
        key: "CH-1/high_temp",
        entity: "CH-1",
        alarm: "high temp",
        activations: { reference: "1", thisRun: "3", delta: "+2" },
        active: { reference: "120 s", thisRun: "300 s", delta: "+180 s" },
        note: "",
      },
    ]);
    expect(alarmTable(comparison).caption).toBe("Whole run: reference 30:00, this run 28:00 simulated");
  });

  it("shows the alarm system measures in the words of a single report", () => {
    const rows = alarmSystemTable(comparison).rows;
    expect(rows[0]).toEqual({
      key: "activations",
      measure: "Alarms raised",
      reference: "1",
      thisRun: "3",
      delta: "+2",
    });
    expect(rows.at(-1)).toMatchObject({ measure: "Still active at the end", delta: "+1" });
    expect(rows).toHaveLength(6);
  });

  it("shows nothing for an alarm system a run did not record", () => {
    const none = alarmSystemTable({ ...comparison, alarmSystem: { ...comparison.alarmSystem, a: null, delta: null } });
    expect(none.rows[0]).toMatchObject({ reference: "–", thisRun: "3", delta: "–" });
  });
});
