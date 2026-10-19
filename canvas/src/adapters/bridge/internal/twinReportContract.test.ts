import { describe, expect, it } from "vitest";
import { cannotJudgeWire, twinReportWire } from "../../../test/twinReportFixture";
import { twinReport } from "./twinReportDecoder";

type Wire = ReturnType<typeof twinReportWire>;

function changed(change: (wire: Wire) => void): unknown {
  const wire = structuredClone(twinReportWire());
  change(wire);
  return wire;
}

describe("the reasons a report cannot judge", () => {
  it.each(["node_stop_not_applied", "node_stop_unconfirmed"])("reads %s", (reason) => {
    expect(twinReport(cannotJudgeWire({ reason }))).toMatchObject({ verdict: "cannot_judge", reason });
  });

  it("reads a report whose evidence is not bound to a run.json", () => {
    const wire = changed((w) => Object.assign(w.run, { evidence_bound: false }));
    expect(twinReport(wire)).toMatchObject({ evidenceBound: false });
  });
});

describe("the worked example of the binary's contract", () => {
  const example = {
    schema_version: "0.1",
    verdict: "diverges",
    summary: {
      devices: 1,
      metrics: 1,
      compared: 3,
      within: 2,
      diverged: 1,
      not_compared: 1,
      unattributed_readings: 0,
      metrics_cannot_judge: 0,
      divergences: 1,
      commands_applied: 1,
      commands_not_applied: 0,
    },
    run: {
      field_trace_digest: `sha256:${"a".repeat(64)}`,
      journal_digest: `sha256:${"b".repeat(64)}`,
      evidence_bound: true,
    },
    twin: { domain: "thermal-bench", domain_digest: `sha256:${"c".repeat(64)}`, kinds: [], start_ns: 0 },
    tolerance_rule: { default: "span_fraction", span_fraction: 0.05, description: "5 % of the declared range." },
    commands: [{ command_id: "cmd-1", device: "zone-01", command: "set_mode", at_ns: 1, status: "applied" }],
    metrics: [
      {
        device: "zone-01",
        kind: "thermal_sensor",
        metric: "temperature",
        unit: "Cel",
        tolerance: { value: 2.5, source: "default_span" },
        verdict: "diverges",
        summary: { compared: 3, within: 2, diverged: 1, not_compared: 1 },
        not_compared_reasons: [{ reason: "before_first_command", count: 1 }],
        series: { points: [{ at_ns: 1, board: 34, twin: 33.9 }], total: 3, truncated: false },
      },
    ],
    divergences: [
      {
        name: "zone-01.temperature#1",
        device: "zone-01",
        metric: "temperature",
        first_ns: 9,
        last_ns: 9,
        readings: 1,
        tolerance: 2.5,
        worst_error: 6.1,
        worst_at_ns: 9,
        board_value: 40,
        twin_value: 33.9,
        commands: ["cmd-1"],
      },
    ],
  };

  it("decodes in full, series and divergences included", () => {
    expect(twinReport(example)).toMatchObject({
      verdict: "diverges",
      twinDomain: "thermal-bench",
      evidenceBound: true,
      commands: [{ commandId: "cmd-1", origin: "host", atNs: 1 }],
      metrics: [{ series: { points: [{ atNs: 1, board: 34, twin: 33.9 }], total: 3, truncated: false } }],
      divergences: [{ name: "zone-01.temperature#1", worstError: 6.1, commands: ["cmd-1"] }],
    });
  });
});
