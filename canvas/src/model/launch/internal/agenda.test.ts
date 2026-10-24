import { describe, expect, it } from "vitest";
import { agendaOf } from "./agenda";
import type { LaunchPlan } from "./launch";

const plan = (overrides: Partial<LaunchPlan> = {}): LaunchPlan => ({
  seed: 1,
  durationS: 1800,
  explore: true,
  faults: [],
  perturbations: [],
  calls: [],
  ...overrides,
});

const categoryOf = (_entity: string, fault: string) => (fault === "sensor_drift" ? "measurement" : "world");

describe("the plan in time order", () => {
  it("lists faults, clears, delivery windows and calls together, earliest first", () => {
    const items = agendaOf(
      plan({
        faults: [
          { entity: "CT-1", fault: "fan_failure", atS: 300, untilS: 720 },
          { entity: "CT-2", fault: "sensor_drift", atS: 180 },
        ],
        perturbations: [{ kind: "drop", params: { rate: 0.5 }, atS: 240, untilS: 600 }],
        calls: [{ entity: "CH-1", effector: "chiller__set_setpoint", atS: 900, args: { setpoint_c: 9 } }],
      }),
      categoryOf,
    );
    expect(items.map((item) => [item.time, item.lane, item.text])).toEqual([
      ["03:00", "fault", "CT-2 sensor drift"],
      ["04:00", "delivery", "drop starts (rate 0.5)"],
      ["05:00", "fault", "CT-1 fan failure"],
      ["10:00", "delivery-end", "drop ends"],
      ["12:00", "clear", "CT-1 fan failure ends"],
      ["15:00", "call", "CH-1 set setpoint (setpoint c 9)"],
    ]);
  });

  it("labels each lane, and says which layer a fault acts on", () => {
    const items = agendaOf(
      plan({
        faults: [
          { entity: "A", fault: "fan_failure", atS: 0 },
          { entity: "B", fault: "sensor_drift", atS: 1 },
        ],
      }),
      categoryOf,
    );
    expect(items.map((item) => item.label)).toEqual(["Process fault", "Sensor fault"]);
    const unknown = agendaOf(plan({ faults: [{ entity: "A", fault: "x", atS: 0 }] }), () => null);
    expect(unknown[0]?.label).toBe("Fault");
  });

  it("keeps the order the simulator applies at one moment: faults, clears, delivery, then calls", () => {
    const items = agendaOf(
      plan({
        calls: [{ entity: "CH-1", effector: "start", atS: 60, args: {} }],
        perturbations: [{ kind: "drop", params: {}, atS: 60, untilS: 120 }],
        faults: [
          { entity: "CT-1", fault: "a", atS: 0, untilS: 60 },
          { entity: "CT-2", fault: "b", atS: 60 },
          { entity: "CT-3", fault: "c", atS: 60 },
        ],
      }),
      categoryOf,
    );
    expect(items.map((item) => `${item.time} ${item.text}`)).toEqual([
      "00:00 CT-1 a",
      "01:00 CT-2 b",
      "01:00 CT-3 c",
      "01:00 CT-1 a ends",
      "01:00 drop starts",
      "01:00 CH-1 start",
      "02:00 drop ends",
    ]);
  });

  it("gives each row the editor it belongs to", () => {
    const items = agendaOf(
      plan({
        faults: [{ entity: "CT-1", fault: "a", atS: 0, untilS: 60 }],
        perturbations: [{ kind: "drop", params: {}, atS: 10, untilS: 30 }],
        calls: [{ entity: "CH-1", effector: "start", atS: 20, args: {} }],
      }),
      categoryOf,
    );
    expect(items.map((item) => [item.lane, item.editor])).toEqual([
      ["fault", "launch-fault-1"],
      ["delivery", "launch-delivery-1"],
      ["call", "launch-call-1"],
      ["delivery-end", "launch-delivery-1-end"],
      ["clear", "launch-fault-1-end"],
    ]);
  });

  it("reads hours when the plan runs that long, and holds a row that is not finished until the end", () => {
    const items = agendaOf(
      plan({
        durationS: 7200,
        faults: [
          { entity: "", fault: "", atS: Number.NaN },
          { entity: "CT-1", fault: "a", atS: 3700 },
        ],
        perturbations: [{ kind: "", params: {}, atS: 10 }],
        calls: [{ entity: "", effector: "", atS: 20, args: {} }],
      }),
      categoryOf,
    );
    expect(items.map((item) => [item.time, item.text])).toEqual([
      ["00:10", "a delivery condition to choose"],
      ["00:20", "an effector call to choose"],
      ["1:01:40", "CT-1 a"],
      ["–", "a fault to choose"],
    ]);
  });

  it("is empty for a plan with nothing in it", () => {
    expect(agendaOf(plan(), categoryOf)).toEqual([]);
  });
});
