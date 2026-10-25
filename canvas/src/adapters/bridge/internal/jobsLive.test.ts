import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";

const wire = {
  world_id: "w-9",
  run_id: "r-9",
  domain: "factory",
  seed: 3,
  state: "running",
  start_ns: 100e9,
  reached_ns: 160e9,
  requested_end_ns: 1900e9,
  rate: 60,
  explore: true,
  calls: [],
};

function answering(result: unknown) {
  return bridgeTransport(() => respond(200, { result }));
}

describe("the plan a job status carries", () => {
  it("reads the operator's planned calls, the live faults and the recovery of a planned fault", async () => {
    const reading = answering({
      ...wire,
      faults: [
        { entity_id: "CT-1", fault: "fan", at_s: 120, until_s: 540, fault_id: "f-0", cleared_at_s: 540 },
        { entity_id: "CH-1", fault: "fouling", at_s: 600, fault_id: "f-1", immediate: true },
      ],
      planned_calls: [{ entity_id: "CH-1", effector: "set_setpoint", args: { setpoint_c: 7 }, at_s: 60 }],
    });
    const outcome = await bridgeAdapter(reading).launch.jobStatus("w-9");
    expect(outcome).toMatchObject({
      ok: true,
      value: {
        faults: [
          { entity: "CT-1", fault: "fan", atS: 120, untilS: 540, faultId: "f-0", clearedAtS: 540 },
          { entity: "CH-1", fault: "fouling", atS: 600, faultId: "f-1", immediate: true },
        ],
        plannedCalls: [{ entity: "CH-1", effector: "set_setpoint", atS: 60, args: { setpoint_c: 7 } }],
      },
    });
  });

  it("refuses a status whose planned call, live flag or recovery second is malformed", async () => {
    const bad = [
      { planned_calls: [{ entity_id: "CH-1", effector: "start", at_s: "soon" }] },
      { planned_calls: [{ entity_id: "CH-1", effector: "start", at_s: 1, args: [1] }] },
      { planned_calls: [{ entity_id: "CH-1", effector: "start", at_s: 1, args: { a: { b: 1 } } }] },
      { planned_calls: "none" },
      { faults: [{ entity_id: "CT-1", fault: "fan", at_s: 1, immediate: "yes" }] },
      { faults: [{ entity_id: "CT-1", fault: "fan", at_s: 1, cleared_at_s: "late" }] },
    ];
    for (const change of bad) {
      const outcome = await bridgeAdapter(answering({ ...wire, ...change })).launch.jobStatus("w-9");
      expect(outcome.ok).toBe(false);
    }
  });
});
