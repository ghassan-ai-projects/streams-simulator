import { describe, expect, it } from "vitest";
import { coolerDocument, instruments, kindOf } from "../../../test/twinFixture";
import { startingChoices } from "./twinDefaults";
import {
  addDrive,
  changeDrive,
  chooseResponse,
  renameInstance,
  renameTwin,
  retitleTwin,
  setArgument,
} from "./twinEdits";
import { choiceProblems, twinDomainPath, twinNameLooksRight, twinPath } from "./twinRules";

const start = startingChoices([kindOf(coolerDocument)], instruments);
const at = { kind: "cooler", command: "set_duty", index: 0 };

describe("twin names and paths", () => {
  it("keeps the choices beside the device and the plant in domains", () => {
    expect(twinPath("cooler")).toBe("twins/cooler.twin.json");
    expect(twinDomainPath("cooler-twin")).toBe("domains/cooler-twin.domain.json");
  });

  it("knows the name the binary accepts", () => {
    expect(twinNameLooksRight("cooler-twin")).toBe(true);
    expect(twinNameLooksRight("Cooler")).toBe(false);
    expect(twinNameLooksRight("ab")).toBe(false);
    expect(twinNameLooksRight("1abc")).toBe(false);
  });
});

describe("what must be filled in before the binary is asked", () => {
  it("accepts the starting choices", () => {
    expect(choiceProblems(start)).toEqual([]);
  });

  it("asks for a name that can be a file, a title and a name for each device", () => {
    const problems = choiceProblems(renameInstance(retitleTwin(renameTwin(start, "bad name"), " "), "cooler", " "));
    expect(problems.map((problem) => [problem.scope, problem.field])).toEqual([
      ["twin", "id"],
      ["twin", "title"],
      ["kind", "instance"],
    ]);
    expect(choiceProblems(renameTwin(start, "")).map((problem) => problem.field)).toEqual(["id"]);
    expect(choiceProblems(renameTwin(start, "Plant_A.1"))).toEqual([]);
  });

  it("asks which reading a command sets, and a complete effect for a first order response", () => {
    const unset = setArgument(start, "cooler", "set_duty", "duty", "");
    expect(choiceProblems(unset)).toMatchObject([{ scope: "command", key: "set_duty", field: "sets" }]);
    const first = chooseResponse(start, "cooler", "set_duty", "first_order");
    expect(choiceProblems(first)).toMatchObject([{ field: "drives" }]);
    const full = changeDrive(first, at, { kind: "cooler", metric: "temperature", gain: 1, timeConstantS: 5 });
    expect(choiceProblems(full)).toEqual([]);
    const partial = changeDrive(full, at, { gain: null });
    expect(choiceProblems(partial)).toMatchObject([{ field: "drives" }]);
    expect(choiceProblems(addDrive(full, "cooler", "set_duty"))).toMatchObject([{ field: "drives" }]);
    const bare = changeDrive(full, at, { from: "" });
    expect(choiceProblems(bare)).toHaveLength(1);
  });

  it("does not ask for a first order effect when there are none", () => {
    const first = chooseResponse(start, "cooler", "set_duty", "first_order");
    const kind = first.kinds[0];
    const command = kind?.commands.set_duty;
    if (kind === undefined || command === undefined) throw new Error("missing");
    const empty = { ...first, kinds: [{ ...kind, commands: { set_duty: { ...command, drives: [] } } }] };
    expect(choiceProblems(empty)).toMatchObject([{ field: "drives" }]);
  });
});
