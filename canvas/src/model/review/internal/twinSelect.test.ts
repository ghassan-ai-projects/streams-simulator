import { describe, expect, it } from "vitest";
import { firstOf, twinModel } from "../../../test/twinReportFixture";
import type { TwinMetric } from "./twin";
import {
  coverageOf,
  divergenceLine,
  inForceWords,
  metricLine,
  ruleWords,
  toleranceWords,
  truncationNote,
  twinOffer,
  unitOfMetric,
} from "./twinSelect";

describe("when the twin comparison is offered", () => {
  it("offers it for a saved session that names a domain, and lets the binary refuse a run with no board", () => {
    expect(twinOffer("saved", "mini")).toEqual({ kind: "offered", domain: "mini" });
  });

  it("shows nothing for a run that is not a saved session", () => {
    expect(twinOffer("none", "mini")).toEqual({ kind: "absent" });
    expect(twinOffer("running", "mini")).toEqual({ kind: "absent" });
  });

  it("says why it is withheld when the session record names no domain", () => {
    expect(twinOffer("saved", "")).toEqual({
      kind: "withheld",
      why: "The session record does not name the domain the twin would run.",
    });
  });
});

describe("how tolerances and the rule are worded", () => {
  it("says the value, the unit and where the tolerance came from", () => {
    expect(toleranceWords({ value: 2.5, source: "default_span" }, "Cel")).toBe("±2.5 Cel (default share of the range)");
    expect(toleranceWords({ value: 3, source: "declared" }, "")).toBe("±3 (declared)");
    expect(toleranceWords({ value: null, source: "none" }, "Cel")).toBe("No tolerance (none)");
  });

  it("shows the binary's own description of the rule and builds no words of its own", () => {
    expect(ruleWords(twinModel({ ruleDescription: "Absolute, 5 % of the range." }))).toBe(
      "Absolute, 5 % of the range.",
    );
  });
});

describe("what covers a device in the report", () => {
  it("gives a series with its divergences for a metric that has compared readings", () => {
    const [covered] = coverageOf(twinModel(), "sensor");
    expect(covered).toMatchObject({ kind: "series", metric: { metric: "temperature" } });
    expect(covered?.kind === "series" ? covered.divergences.map((d) => d.name) : []).toEqual(["sensor.temperature#1"]);
    expect(covered).toMatchObject({ undrawn: 0 });
  });

  it("leaves out a divergence that lies beyond the readings the truncated series draws, and counts it", () => {
    const base = twinModel();
    const metric: TwinMetric = {
      ...firstOf(base.metrics),
      series: { ...firstOf(base.metrics).series, truncated: true, total: 3000 },
    };
    const beyond = { ...firstOf(base.divergences), name: "sensor.temperature#2", firstNs: 900e9, lastNs: 950e9 };
    const straddling = { ...firstOf(base.divergences), name: "sensor.temperature#3", firstNs: 450e9, lastNs: 950e9 };
    const report = twinModel({ metrics: [metric], divergences: [...base.divergences, beyond, straddling] });
    const [covered] = coverageOf(report, "sensor");
    expect(covered).toMatchObject({ kind: "series", undrawn: 1 });
    expect(covered?.kind === "series" ? covered.divergences.map((d) => d.name) : []).toEqual([
      "sensor.temperature#1",
      "sensor.temperature#3",
    ]);
  });

  it("keeps the divergences of other metrics and devices out of a series", () => {
    const base = twinModel();
    const other = { ...firstOf(base.divergences), name: "other.temperature#1", device: "other" };
    const report = twinModel({ divergences: [...base.divergences, other] });
    const [covered] = coverageOf(report, "sensor");
    expect(covered?.kind === "series" ? covered.divergences : []).toHaveLength(1);
  });

  it("explains why a metric with no compared reading cannot be judged, with the binary's reasons", () => {
    const metric: TwinMetric = {
      ...firstOf(twinModel().metrics),
      series: { points: [], total: 0, truncated: false },
      notCompared: [
        { reason: "before_first_command", count: 3 },
        { reason: "quality_not_good", count: 1 },
      ],
    };
    const [covered] = coverageOf(twinModel({ metrics: [metric] }), "sensor");
    expect(covered).toMatchObject({
      kind: "empty",
      why: "The twin cannot judge temperature: readings before the first command was applied to the twin (3); readings whose quality was not good (1).",
    });
  });

  it("says the board reported nothing when no reason is given", () => {
    const metric: TwinMetric = {
      ...firstOf(twinModel().metrics),
      series: { points: [], total: 0, truncated: false },
      notCompared: [],
    };
    expect(coverageOf(twinModel({ metrics: [metric] }), "sensor")).toMatchObject([
      { kind: "empty", why: "The board reported no reading of temperature to compare." },
    ]);
  });

  it("covers nothing for a device the report does not name", () => {
    expect(coverageOf(twinModel(), "fan")).toEqual([]);
  });
});

describe("lines about a metric and a divergence", () => {
  it("counts the readings and names the tolerance", () => {
    const metric = firstOf(twinModel().metrics);
    expect(metricLine(metric)).toBe("5 readings compared · 2 beyond tolerance · ±2.5 Cel (default share of the range)");
    expect(metricLine({ ...metric, counts: { ...metric.counts, compared: 1, diverged: 0 } })).toContain(
      "1 reading compared · 0 beyond",
    );
  });

  it("gives the interval, the worst error and both values on the run clock", () => {
    const divergence = firstOf(twinModel().divergences);
    expect(divergenceLine(divergence, "Cel", 0)).toBe(
      "00:06:40 to 00:08:20 · 2 readings · worst error 8 Cel at 00:08:20 (board 30 Cel, twin 22 Cel)",
    );
  });

  it("names one time when the stretch is one reading, and omits an empty unit", () => {
    const single = { ...firstOf(twinModel().divergences), firstNs: 400e9, lastNs: 400e9, readings: 1 };
    expect(divergenceLine(single, "", 0)).toBe("00:06:40 · 1 reading · worst error 8 at 00:08:20 (board 30, twin 22)");
  });

  it("lists the commands in force or says there were none", () => {
    const divergence = firstOf(twinModel().divergences);
    expect(inForceWords(divergence)).toBe("cmd-1");
    expect(inForceWords({ ...divergence, commands: ["a", "b"] })).toBe("a, b");
    expect(inForceWords({ ...divergence, commands: [] })).toBe("none");
  });

  it("finds the unit of the metric a divergence belongs to", () => {
    const divergence = firstOf(twinModel().divergences);
    expect(unitOfMetric(twinModel(), divergence)).toBe("Cel");
    expect(unitOfMetric(twinModel({ metrics: [] }), divergence)).toBe("");
  });

  it("says when the series was cut short", () => {
    const { series } = firstOf(twinModel().metrics);
    expect(truncationNote(series, 0)).toBe("");
    expect(truncationNote({ ...series, truncated: true, total: 3000 }, 0)).toBe(
      "Showing the first 5 of 3000 compared readings.",
    );
    expect(truncationNote({ ...series, truncated: true, total: 3000 }, 1)).toBe(
      "Showing the first 5 of 3000 compared readings. 1 stretch where they part ways lies beyond the readings shown; the Report tab lists it.",
    );
    expect(truncationNote({ ...series, truncated: true, total: 3000 }, 3)).toContain(
      "3 stretches where they part ways lie beyond the readings shown; the Report tab lists them.",
    );
  });
});
