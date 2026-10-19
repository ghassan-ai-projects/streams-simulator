import { describe, expect, it } from "vitest";
import { cannotJudgeWire, twinMetricWire, twinReportWire } from "../../../test/twinReportFixture";
import { twinReport } from "./twinReportDecoder";

type Wire = ReturnType<typeof twinReportWire>;

function changed(change: (wire: Wire) => void): unknown {
  const wire = structuredClone(twinReportWire());
  change(wire);
  return wire;
}

describe("the twin report document", () => {
  it("reads a diverging report in full", () => {
    const report = twinReport(twinReportWire());
    expect(report).toMatchObject({
      verdict: "diverges",
      reason: null,
      twinDomain: "mini",
      spanFraction: 0.05,
      summary: { compared: 5, within: 3, diverged: 2, notCompared: 2, commandsApplied: 1, commandsNotApplied: 2 },
    });
    expect(report?.commands).toEqual([
      {
        commandId: "cmd-1",
        device: "sensor",
        command: "set_mode",
        atNs: 50e9,
        status: "applied",
        reason: null,
        detail: "",
      },
      {
        commandId: "cmd-2",
        device: "sensor",
        command: "set_mode",
        atNs: 60e9,
        status: "not_applied",
        reason: "refused_by_board",
        detail: "not_ready",
      },
      {
        commandId: "cmd-3",
        device: "sensor",
        command: "reset",
        atNs: null,
        status: "twin_refused",
        reason: null,
        detail: "unknown effector",
      },
    ]);
    expect(report?.metrics[0]).toMatchObject({
      device: "sensor",
      metric: "temperature",
      unit: "Cel",
      tolerance: { value: 2.5, source: "default_span" },
      verdict: "diverges",
      counts: { compared: 5, within: 3, diverged: 2, notCompared: 2 },
      notCompared: [{ reason: "before_first_command", count: 2 }],
      series: { total: 5, truncated: false },
    });
    expect(report?.metrics[0]?.series.points[3]).toEqual({ atNs: 400e9, board: 28, twin: 22 });
    expect(report?.divergences[0]).toEqual({
      name: "sensor.temperature#1",
      device: "sensor",
      metric: "temperature",
      firstNs: 400e9,
      lastNs: 500e9,
      readings: 2,
      tolerance: 2.5,
      worstError: 8,
      worstAtNs: 500e9,
      boardValue: 30,
      twinValue: 22,
      commands: ["cmd-1"],
    });
  });

  it("reads a report that cannot judge, with its reason and a metric that has no tolerance", () => {
    const wire = cannotJudgeWire();
    const report = twinReport(wire);
    expect(report).toMatchObject({
      verdict: "cannot_judge",
      reason: "no_commands_applied",
      commands: [],
      divergences: [],
    });
    const noTolerance = cannotJudgeWire({
      metrics: [
        twinMetricWire({
          verdict: "cannot_judge",
          tolerance: { value: null, source: "none" },
          not_compared_reasons: [{ reason: "no_tolerance_declared", count: 3 }],
          series: { points: [], total: 0, truncated: false },
        }),
      ],
    });
    expect(twinReport(noTolerance)?.metrics[0]).toMatchObject({
      tolerance: { value: null, source: "none" },
      notCompared: [{ reason: "no_tolerance_declared", count: 3 }],
    });
  });

  it.each<[string, unknown]>([
    ["a value that is not a document", null],
    ["a list", []],
    ["another schema version", { ...twinReportWire(), schema_version: "0.2" }],
    ["no schema version", changed((w) => Reflect.deleteProperty(w, "schema_version"))],
    ["a verdict nobody defined", { ...twinReportWire(), verdict: "mostly_agrees" }],
    ["a reason nobody defined", { ...twinReportWire(), reason: "tired" }],
    ["a summary that is not a document", { ...twinReportWire(), summary: 3 }],
    ["a summary count that is not a number", changed((w) => Object.assign(w.summary, { compared: "5" }))],
    ["a summary count that is not finite", changed((w) => Object.assign(w.summary, { within: Infinity }))],
    ["a summary count that is missing", changed((w) => Reflect.deleteProperty(w.summary, "devices"))],
    ["no twin", { ...twinReportWire(), twin: 1 }],
    ["a twin with no domain", changed((w) => Object.assign(w.twin, { domain: "" }))],
    ["no tolerance rule", { ...twinReportWire(), tolerance_rule: null }],
    ["a tolerance share that is not a number", changed((w) => Object.assign(w.tolerance_rule, { span_fraction: "x" }))],
    ["commands that are not a list", { ...twinReportWire(), commands: {} }],
    ["metrics that are not a list", { ...twinReportWire(), metrics: {} }],
    ["divergences that are not a list", { ...twinReportWire(), divergences: {} }],
  ])("drops a report with %s as unreadable", (_name, value) => {
    expect(twinReport(value)).toBeNull();
  });
});

function withCommand(command: unknown): unknown {
  return { ...twinReportWire(), commands: [command] };
}

const applied = { command_id: "cmd-1", device: "sensor", command: "set_mode", status: "applied" };

describe("the commands of the twin report", () => {
  it("reads a command that carries no time", () => {
    expect(twinReport(withCommand(applied))?.commands).toEqual([
      {
        commandId: "cmd-1",
        device: "sensor",
        command: "set_mode",
        atNs: null,
        status: "applied",
        reason: null,
        detail: "",
      },
    ]);
  });

  it.each<[string, unknown]>([
    ["a row that is not a document", 7],
    ["a status nobody defined", { ...applied, status: "done" }],
    ["a reason nobody defined", { ...applied, status: "not_applied", reason: "shy" }],
    ["a time that is not a number", { ...applied, at_ns: "50" }],
    ["no command id", { ...applied, command_id: "" }],
    ["no device", { ...applied, device: undefined }],
    ["a command that was not applied and gives no reason", { ...applied, status: "not_applied" }],
  ])("drops the report when a command has %s", (_name, command) => {
    expect(twinReport(withCommand(command))).toBeNull();
  });
});

function withMetric(change: (metric: Record<string, unknown>) => void): unknown {
  const wire = structuredClone(twinReportWire());
  change(wire.metrics[0] as Record<string, unknown>);
  return wire;
}

type Change = (record: Record<string, unknown>) => unknown;

describe("the metrics of the twin report", () => {
  it.each<[string, Change]>([
    ["a verdict nobody defined", (m: Record<string, unknown>) => (m.verdict = "fine")],
    ["a tolerance that is not a document", (m: Record<string, unknown>) => (m.tolerance = 2)],
    [
      "a tolerance source nobody defined",
      (m: Record<string, unknown>) => (m.tolerance = { value: 1, source: "guess" }),
    ],
    ["a negative tolerance", (m: Record<string, unknown>) => (m.tolerance = { value: -1, source: "declared" })],
    [
      "a tolerance that is not a number",
      (m: Record<string, unknown>) => (m.tolerance = { value: "1", source: "declared" }),
    ],
    ["counts that are missing", (m: Record<string, unknown>) => (m.summary = { compared: 1 })],
    ["a list of reasons that is not a list", (m: Record<string, unknown>) => (m.not_compared_reasons = "x")],
    [
      "a reason nobody defined",
      (m: Record<string, unknown>) => (m.not_compared_reasons = [{ reason: "bored", count: 1 }]),
    ],
    [
      "a reason count that is not a number",
      (m: Record<string, unknown>) => (m.not_compared_reasons = [{ reason: "historical", count: "1" }]),
    ],
    ["no series", (m: Record<string, unknown>) => (m.series = null)],
    [
      "a series total that is not a number",
      (m: Record<string, unknown>) => (m.series = { points: [], total: "0", truncated: false }),
    ],
    [
      "a series flag that is not a boolean",
      (m: Record<string, unknown>) => (m.series = { points: [], total: 0, truncated: "no" }),
    ],
    [
      "a point that is not a document",
      (m: Record<string, unknown>) => (m.series = { points: [1], total: 1, truncated: false }),
    ],
    [
      "a point with a value that is not a number",
      (m: Record<string, unknown>) =>
        (m.series = { points: [{ at_ns: 1, board: "x", twin: 1 }], total: 1, truncated: false }),
    ],
    ["no device", (m: Record<string, unknown>) => (m.device = "")],
    ["no metric name", (m: Record<string, unknown>) => Reflect.deleteProperty(m, "metric")],
  ])("drops the report when a metric has %s", (_name, change) => {
    expect(twinReport(withMetric(change))).toBeNull();
  });

  it("reads a metric with no unit and an empty list of reasons", () => {
    const wire = withMetric((m) => {
      Reflect.deleteProperty(m, "unit");
      m.not_compared_reasons = [];
    });
    expect(twinReport(wire)?.metrics[0]).toMatchObject({ unit: "", notCompared: [] });
  });
});

function withDivergence(change: (divergence: Record<string, unknown>) => void): unknown {
  const wire = structuredClone(twinReportWire());
  change(wire.divergences[0] as Record<string, unknown>);
  return wire;
}

describe("the divergences of the twin report", () => {
  it.each<[string, Change]>([
    ["a number that is missing", (d: Record<string, unknown>) => Reflect.deleteProperty(d, "worst_error")],
    ["a time that is not a number", (d: Record<string, unknown>) => (d.first_ns = "400")],
    ["commands that are not a list of ids", (d: Record<string, unknown>) => (d.commands = [1])],
    ["no name", (d: Record<string, unknown>) => (d.name = "")],
    ["no device", (d: Record<string, unknown>) => Reflect.deleteProperty(d, "device")],
  ])("drops the report when a divergence has %s", (_name, change) => {
    expect(twinReport(withDivergence(change))).toBeNull();
  });
});
