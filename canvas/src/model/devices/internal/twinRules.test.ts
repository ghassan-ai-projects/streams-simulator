import { describe, expect, it } from "vitest";
import { coolerPath, startingChoices } from "../../../test/twinFixture";
import {
  addDrive,
  changeDrive,
  chooseResponse,
  renameInstance,
  renameTwin,
  retitleTwin,
  setArgument,
} from "./twinEdits";
import { choiceProblems, twinDomainPath, twinPath } from "./twinRules";

const start = startingChoices();
const at = { kind: coolerPath, command: "c1", index: 0 };

describe("twin paths", () => {
  it("keeps the choices beside the device and the plant in domains", () => {
    expect(twinPath("cooler")).toBe("twins/cooler.twin.json");
    expect(twinDomainPath("cooler-twin")).toBe("domains/cooler-twin.domain.json");
  });
});

describe("what must be filled in before the binary is asked", () => {
  it("accepts the suggested choices", () => {
    expect(choiceProblems(start)).toEqual([]);
  });

  it("asks for a name, a title and a name for each device, and leaves the pattern to the binary", () => {
    const problems = choiceProblems(renameInstance(retitleTwin(renameTwin(start, " "), " "), coolerPath, " "));
    expect(problems.map((problem) => [problem.scope, problem.field])).toEqual([
      ["twin", "id"],
      ["twin", "title"],
      ["kind", "instance"],
    ]);
    expect(choiceProblems(renameTwin(start, "Plant A.1"))).toEqual([]);
  });

  it("asks which reading a command sets, and a complete effect for a first order response", () => {
    const unset = setArgument(start, coolerPath, "c1", "duty", "");
    expect(choiceProblems(unset)).toMatchObject([{ scope: "command", key: "c1", field: "sets" }]);
    const first = chooseResponse(start, coolerPath, "c1", "first_order");
    expect(choiceProblems(first)).toMatchObject([{ field: "drives" }]);
    const full = changeDrive(first, at, { kind: coolerPath, metric: "m1", gain: 1, timeConstantS: 5 });
    expect(choiceProblems(full)).toEqual([]);
    const partial = changeDrive(full, at, { gain: null });
    expect(choiceProblems(partial)).toMatchObject([{ field: "drives" }]);
    expect(choiceProblems(addDrive(full, coolerPath, "c1"))).toMatchObject([{ field: "drives" }]);
    const bare = changeDrive(full, at, { from: "" });
    expect(choiceProblems(bare)).toHaveLength(1);
  });

  it("says a time constant at or below zero must be above zero, on the effect that has it", () => {
    const first = chooseResponse(start, coolerPath, "c1", "first_order");
    const full = changeDrive(first, at, { kind: coolerPath, metric: "m1", gain: 1, timeConstantS: -5 });
    expect(choiceProblems(full)).toMatchObject([
      { scope: "command", key: "c1", field: "time_constant_s", spot: "0", text: "Must be above zero" },
    ]);
    expect(choiceProblems(changeDrive(full, at, { timeConstantS: 0 }))).toHaveLength(1);
    expect(choiceProblems(changeDrive(full, at, { timeConstantS: 0.5 }))).toEqual([]);
  });

  it("does not ask for a first order effect when there are none", () => {
    const first = chooseResponse(start, coolerPath, "c1", "first_order");
    const kind = first.kinds[0];
    const command = kind?.commands.c1;
    if (kind === undefined || command === undefined) throw new Error("missing");
    const empty = { ...first, kinds: [{ ...kind, commands: { c1: { ...command, drives: [] } } }] };
    expect(choiceProblems(empty)).toMatchObject([{ field: "drives" }]);
  });
});
