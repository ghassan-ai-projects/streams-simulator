import { describe, expect, it } from "vitest";
import { setpointPlant as plant } from "../../../test/setpointPlant";
import type { NowDraft } from "./nowCall";
import { emptyNowDraft, nowCallOf, nowCallProblem } from "./nowCall";

const draft = (next: Partial<NowDraft>): NowDraft => ({ ...emptyNowDraft, ...next });
const problem = (next: Partial<NowDraft>) => nowCallProblem(draft(next), plant);

describe("a call given now", () => {
  it("names what is missing, in the words of a single call", () => {
    expect(problem({})).toMatchObject({ part: "instance", text: "The call needs an instance and an effector." });
    expect(problem({ effector: "set_setpoint" })).toMatchObject({
      part: "instance",
      text: "The call needs an instance.",
    });
    expect(problem({ entity: "CH-1" })).toMatchObject({ part: "effector", text: "The call needs an effector." });
  });

  it("holds the arguments to the effector's declared bounds before anything is sent", () => {
    const base = { entity: "CH-1", effector: "set_setpoint" };
    expect(problem({ ...base, values: {} })?.text).toBe("The call: setpoint c needs a value.");
    expect(problem({ ...base, values: { setpoint_c: "20" } })).toMatchObject({
      part: "args",
      text: "The call: setpoint c is a number from 4 to 12.",
    });
    expect(problem({ ...base, values: { setpoint_c: "7" } })).toBeNull();
  });

  it("fills the argument naming the instance itself, as a planned call does", () => {
    const call = nowCallOf(draft({ entity: "CH-1", effector: "set_setpoint", values: { setpoint_c: "7" } }), plant);
    expect(call).toEqual({ entity: "CH-1", effector: "set_setpoint", args: { chiller_id: "CH-1", setpoint_c: 7 } });
  });

  it("names an effector the instance does not have", () => {
    expect(problem({ entity: "CT-1", effector: "start" })?.text).toBe("The call: CT-1 has no effector start.");
  });
});
