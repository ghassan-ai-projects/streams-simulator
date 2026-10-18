import { describe, expect, it } from "vitest";
import { coolerDocument, instruments, kindOf } from "../../../test/twinFixture";
import { startingChoices } from "./twinDefaults";
import {
  addDrive,
  changeDrive,
  chooseInstrument,
  chooseResponse,
  removeDrive,
  renameInstance,
  renameTwin,
  retitleTwin,
  setArgument,
  startAt,
} from "./twinEdits";

const start = startingChoices([kindOf(coolerDocument)], instruments);
const at = { kind: "cooler", command: "set_duty", index: 0 };

function command(choices: typeof start) {
  const found = choices.kinds[0]?.commands.set_duty;
  if (found === undefined) throw new Error("no command");
  return found;
}

describe("editing the choices", () => {
  it("renames and retitles the twin and names a device in it", () => {
    expect(renameTwin(start, "plant-a").id).toBe("plant-a");
    expect(retitleTwin(start, "Plant A").title).toBe("Plant A");
    expect(renameInstance(start, "cooler", "c-9").kinds[0]?.instance).toBe("c-9");
    expect(renameInstance(start, "other", "c-9")).toEqual(start);
  });

  it("chooses an instrument and a starting value for a reading, and ignores a reading it does not have", () => {
    const chosen = chooseInstrument(start, "cooler", "duty", "streamsim/thermistor@1");
    expect(chosen.kinds[0]?.metrics.duty?.instrument).toBe("streamsim/thermistor@1");
    expect(startAt(chosen, "cooler", "duty", 40).kinds[0]?.metrics.duty).toEqual({
      instrument: "streamsim/thermistor@1",
      initial: 40,
    });
    expect(chooseInstrument(start, "cooler", "absent", "x")).toEqual(start);
  });

  it("changes the response: first order starts with one effect, anything else clears the effects", () => {
    const first = chooseResponse(start, "cooler", "set_duty", "first_order");
    expect(command(first).drives).toEqual([{ from: "duty", kind: "", metric: "", gain: null, timeConstantS: null }]);
    expect(chooseResponse(first, "cooler", "set_duty", "first_order")).toEqual(first);
    const direct = chooseResponse(first, "cooler", "set_duty", "direct");
    expect(command(direct)).toMatchObject({ response: "direct", sets: { duty: "duty" }, drives: [] });
    expect(command(chooseResponse(direct, "cooler", "set_duty", "none"))).toEqual({
      response: "none",
      sets: {},
      drives: [],
    });
    const blank = setArgument(chooseResponse(start, "cooler", "set_duty", "none"), "cooler", "set_duty", "x", "");
    expect(command(chooseResponse(blank, "cooler", "set_duty", "first_order")).drives[0]?.from).toBe("");
  });

  it("sets and clears the reading an argument sets", () => {
    const set = setArgument(start, "cooler", "set_duty", "lease_ms", "duty");
    expect(command(set).sets).toEqual({ duty: "duty", lease_ms: "duty" });
    expect(command(setArgument(set, "cooler", "set_duty", "lease_ms", "")).sets).toEqual({ duty: "duty" });
    expect(command(setArgument(set, "cooler", "set_duty", "duty", "")).sets).toEqual({ lease_ms: "duty" });
  });

  it("adds, changes and removes effects", () => {
    const first = chooseResponse(start, "cooler", "set_duty", "first_order");
    const added = addDrive(first, "cooler", "set_duty");
    expect(command(added).drives).toHaveLength(2);
    const changed = changeDrive(added, { ...at, index: 1 }, { gain: -0.5, timeConstantS: 30, kind: "cooler" });
    expect(command(changed).drives[1]).toMatchObject({ from: "duty", kind: "cooler", gain: -0.5, timeConstantS: 30 });
    expect(command(changed).drives[0]?.gain).toBeNull();
    expect(command(removeDrive(changed, "cooler", "set_duty", 0)).drives).toHaveLength(1);
    expect(command(removeDrive(changed, "cooler", "set_duty", 0)).drives[0]?.gain).toBe(-0.5);
    expect(addDrive(start, "cooler", "absent")).toEqual(start);
  });
});
