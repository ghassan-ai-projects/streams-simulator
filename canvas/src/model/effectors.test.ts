import { describe, expect, it } from "vitest";
import { argProblem, callArgs, parseEffectors } from "./effectors";

const declared = [
  {
    name: "chiller__set_setpoint",
    args_schema: {
      type: "object",
      required: ["chiller_id", "setpoint_c"],
      properties: { chiller_id: { type: "string" }, setpoint_c: { type: "number", minimum: 4, maximum: 12 } },
    },
  },
  { name: "chiller__start", args_schema: { required: ["chiller_id"], properties: { chiller_id: { type: "string" } } } },
  {
    name: "mixer__mode",
    args_schema: {
      required: ["mode", "on", "level", "extra"],
      properties: {
        cluster_id: { type: "string" },
        pod_id: { type: "string" },
        mode: { type: "string", enum: ["auto", "manual"] },
        on: { type: "boolean" },
        level: { type: "integer", minimum: 0 },
        extra: { type: "object" },
      },
    },
  },
  { name: "bare" },
  { args_schema: {} },
];

describe("effector specs from a domain kind", () => {
  const specs = parseEffectors(declared);

  it("finds the instance argument and keeps the others", () => {
    expect(specs.map((spec) => spec.name)).toEqual(["chiller__set_setpoint", "chiller__start", "mixer__mode", "bare"]);
    expect(specs[0]).toEqual({
      name: "chiller__set_setpoint",
      entityArg: "chiller_id",
      args: [{ name: "setpoint_c", type: "number", required: true, options: [], min: 4, max: 12 }],
    });
    expect(specs[1]).toEqual({ name: "chiller__start", entityArg: "chiller_id", args: [] });
    expect(specs[2]?.entityArg).toBeNull();
    expect(specs[3]).toEqual({ name: "bare", entityArg: null, args: [] });
    expect(parseEffectors(null)).toEqual([]);
  });

  it("builds the arguments of a call from the instance and the typed values", () => {
    expect(callArgs(specs[0], "CH-1", { setpoint_c: " 7.5 " })).toEqual({ chiller_id: "CH-1", setpoint_c: 7.5 });
    expect(callArgs(specs[0], "CH-1", {})).toEqual({ chiller_id: "CH-1" });
    expect(callArgs(specs[2], "P-1", { mode: "auto", on: "true", level: "2" })).toEqual({
      mode: "auto",
      on: true,
      level: 2,
    });
    expect(callArgs(undefined, "CH-1", {})).toEqual({});
  });

  it("names the first argument that is missing or out of bounds", () => {
    const spec = specs[2];
    if (spec === undefined) throw new Error("no spec");
    expect(argProblem(spec, {})).toBe("mode needs a value.");
    expect(argProblem(spec, { mode: "fast", on: true, level: 1 })).toBe("mode is one of auto, manual.");
    expect(argProblem(spec, { mode: "auto", level: 1.5 })).toBe("on needs a value.");
    expect(argProblem(spec, { mode: "auto", on: true, level: 1.5 })).toBe("level is a whole number of 0 or more.");
    expect(argProblem(spec, { mode: "auto", on: true, level: 1 })).toBe("extra cannot be entered here.");
    const setpoint = specs[0];
    if (setpoint === undefined) throw new Error("no spec");
    expect(argProblem(setpoint, { setpoint_c: 13 })).toBe("setpoint c is a number from 4 to 12.");
    expect(argProblem(setpoint, { setpoint_c: Number.NaN })).toBe("setpoint c is a number from 4 to 12.");
    expect(argProblem(setpoint, { setpoint_c: 5 })).toBe("");
  });
});
