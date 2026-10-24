import { describe, expect, it } from "vitest";
import { comparison } from "../../../test/compareFixture";
import { planLanes } from "./comparePlan";

describe("what differs between the two plans", () => {
  const lanes = planLanes(comparison);
  const lane = (key: string) => lanes.find((each) => each.key === key);

  it("has one lane per kind of plan item, in the order a reader plans them", () => {
    expect(lanes.map((each) => each.title)).toEqual(["Faults", "Delivery conditions", "Environment", "Effector calls"]);
  });

  it("puts an item only the reference has under the reference, in human words", () => {
    expect(lane("environment")?.rows).toEqual([
      {
        key: "environment-only_in_a-0",
        change: "only_in_reference",
        label: "Only in reference",
        text: "weather heat wave at 00:15:00",
        differences: [],
      },
    ]);
  });

  it("puts an item only this run has under this run, with its window and settings", () => {
    expect(lane("perturbations")?.rows[0]).toMatchObject({
      change: "only_in_this",
      label: "Only in this run",
      text: "drop starts 00:05:00, ends 00:10:00 (rate 0.5)",
    });
  });

  it("shows a changed item as both versions and names the fields that differ in plain words", () => {
    expect(lane("faults")?.rows).toEqual([
      {
        key: "faults-changed-0",
        change: "changed",
        label: "Changed",
        text: "CT-1 fan failure",
        reference: "CT-1 fan failure from 00:10:00",
        thisRun: "CT-1 fan failure from 00:11:00 until 00:15:00",
        differences: ["clear time", "start time"],
      },
    ]);
  });

  it("counts identical items instead of listing them", () => {
    expect(lane("effectorCalls")?.identical).toBe("2 effector calls are the same in both runs");
    expect(lane("faults")?.identical).toBe("");
    expect(lane("effectorCalls")?.rows.map((row) => row.label)).toEqual(["Only in this run", "Changed"]);
  });

  it("describes a call with its target, arguments, time and what the binary reported", () => {
    expect(lane("effectorCalls")?.rows[0]?.text).toBe(
      "CH-1 set setpoint (setpoint c 7) at 00:04:00 · applied · 300 ms",
    );
    expect(lane("effectorCalls")?.rows[1]).toMatchObject({
      reference: "CH-1 set setpoint (setpoint c 7) at 00:03:00 · applied · 300 ms",
      thisRun: "CH-1 set setpoint (setpoint c 9) at 00:03:00 · applied · 300 ms",
      differences: ["arguments"],
    });
  });

  it("says nothing differs when a lane is the same in both runs", () => {
    const same = planLanes({
      ...comparison,
      plan: {
        ...comparison.plan,
        environment: {
          identical: comparison.plan.effectorCalls.identical.map(() => ({ atNs: 0, target: "t", fault: "f" })),
          different: [],
          onlyInA: [],
          onlyInB: [],
        },
      },
    });
    expect(same.find((each) => each.key === "environment")).toMatchObject({
      rows: [],
      identical: "2 environment faults are the same in both runs",
    });
  });

  it("reads a delivery condition that was cleared", () => {
    const cleared = {
      ...comparison.plan.perturbations,
      onlyInA: [{ action: "clear", kind: "drop", atNs: 600e9, fromNs: 0, untilNs: 0, params: {} }],
    };
    const shown = planLanes({ ...comparison, plan: { ...comparison.plan, perturbations: cleared } });
    expect(shown.find((each) => each.key === "perturbations")?.rows[0]?.text).toBe("drop cleared at 00:10:00");
  });
});
