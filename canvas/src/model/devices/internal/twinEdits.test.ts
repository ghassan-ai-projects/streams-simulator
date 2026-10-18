import { describe, expect, it } from "vitest";
import { cooler, coolerPath, kindOf, startingChoices } from "../../../test/twinFixture";
import { reconcileChoices } from "./twinReconcile";
import type { Response } from "./twinTypes";
import {
  addDrive,
  changeDrive,
  chooseInstrument,
  chooseResponse,
  pickResponse,
  removeDrive,
  renameInstance,
  renameTwin,
  retitleTwin,
  setArgument,
  startAt,
} from "./twinEdits";

const start = startingChoices();
const at = { kind: coolerPath, command: "c1", index: 0 };

function command(choices: typeof start) {
  const found = choices.kinds[0]?.commands.c1;
  if (found === undefined) throw new Error("no command");
  return found;
}

describe("editing the choices", () => {
  it("renames and retitles the twin and names a device in it", () => {
    expect(renameTwin(start, "plant-a").id).toBe("plant-a");
    expect(retitleTwin(start, "Plant A").title).toBe("Plant A");
    expect(renameInstance(start, coolerPath, "c-9").kinds[0]?.instance).toBe("c-9");
    expect(renameInstance(start, "other", "c-9")).toEqual(start);
  });

  it("chooses an instrument and a starting value for a reading, and ignores a reading it does not have", () => {
    const chosen = chooseInstrument(start, coolerPath, "m2", "streamsim/thermistor@1");
    expect(chosen.kinds[0]?.metrics.m2?.instrument).toBe("streamsim/thermistor@1");
    expect(startAt(chosen, coolerPath, "m2", 40).kinds[0]?.metrics.m2).toEqual({
      instrument: "streamsim/thermistor@1",
      initial: 40,
    });
    expect(chooseInstrument(start, coolerPath, "absent", "x")).toEqual(start);
  });

  it("changes the response: first order starts with one blank effect, anything else clears the effects", () => {
    const first = chooseResponse(start, coolerPath, "c1", "first_order");
    expect(command(first).drives).toEqual([{ from: "m2", kind: "", metric: "", gain: null, timeConstantS: null }]);
    expect(chooseResponse(first, coolerPath, "c1", "first_order")).toEqual(first);
    const direct = chooseResponse(first, coolerPath, "c1", "direct");
    expect(command(direct)).toMatchObject({ response: "direct", sets: { duty: "m2" }, drives: [] });
    expect(command(chooseResponse(direct, coolerPath, "c1", "none"))).toEqual({
      response: "none",
      sets: {},
      drives: [],
    });
    const blank = setArgument(chooseResponse(start, coolerPath, "c1", "none"), coolerPath, "c1", "x", "");
    expect(command(chooseResponse(blank, coolerPath, "c1", "first_order")).drives[0]?.from).toBe("");
  });

  it("sets and clears the reading an argument sets", () => {
    const set = setArgument(start, coolerPath, "c1", "lease_ms", "m2");
    expect(command(set).sets).toEqual({ duty: "m2", lease_ms: "m2" });
    expect(command(setArgument(set, coolerPath, "c1", "lease_ms", "")).sets).toEqual({ duty: "m2" });
    expect(command(setArgument(set, coolerPath, "c1", "duty", "")).sets).toEqual({ lease_ms: "m2" });
  });

  it("adds, changes and removes effects", () => {
    const first = chooseResponse(start, coolerPath, "c1", "first_order");
    const added = addDrive(first, coolerPath, "c1");
    expect(command(added).drives).toHaveLength(2);
    const changed = changeDrive(added, { ...at, index: 1 }, { gain: -0.5, timeConstantS: 30, kind: coolerPath });
    expect(command(changed).drives[1]).toMatchObject({ from: "m2", kind: coolerPath, gain: -0.5, timeConstantS: 30 });
    expect(command(changed).drives[0]?.gain).toBeNull();
    expect(command(removeDrive(changed, coolerPath, "c1", 0)).drives).toHaveLength(1);
    expect(command(removeDrive(changed, coolerPath, "c1", 0)).drives[0]?.gain).toBe(-0.5);
    expect(addDrive(start, coolerPath, "absent")).toEqual(start);
  });
});

describe("choosing a response for a command", () => {
  const valve = kindOf({
    kind: "valve",
    metrics: [
      {
        name: "opening",
        data_type: "number",
        unit: "%",
        eu_range: { min: 0, max: 100 },
        resolution: 1,
        writable: true,
      },
    ],
    commands: [{ name: "set_opening", args_schema: { type: "object", properties: { opening: { type: "number" } } } }],
  });
  const path = "kinds/valve.kind.json";
  const blank = reconcileChoices({ id: "v", title: "V", kinds: [], extra: {} }, [{ path, kind: valve }], null);
  const [command] = valve.commands;
  const [metric] = valve.metrics;
  if (command === undefined || metric === undefined) throw new Error("no valve");
  const ask = (choices: typeof blank, response: Response) =>
    pickResponse(choices, { member: path, kind: valve, command }, response);
  const setsOf = (choices: typeof blank) => choices.kinds[0]?.commands[command.key]?.sets;

  it("preselects the only reading the only number argument can set, once", () => {
    const direct = ask(blank, "direct");
    expect(direct.kinds[0]?.commands[command.key]).toMatchObject({ response: "direct", sets: { opening: metric.key } });
    expect(setsOf(ask(direct, "none"))).toEqual({});
    expect(setsOf(ask(setArgument(direct, path, command.key, "opening", ""), "first_order"))).toEqual({
      opening: metric.key,
    });
    expect(setsOf(ask(ask(direct, "first_order"), "direct"))).toEqual({ opening: metric.key });
  });

  it("leaves the choice to the person when there are several arguments", () => {
    const [twoArguments] = cooler.kind.commands;
    if (twoArguments === undefined) throw new Error("no command");
    const none = chooseResponse(start, coolerPath, twoArguments.key, "none");
    const asked = { member: coolerPath, kind: cooler.kind, command: twoArguments };
    const direct = pickResponse(none, asked, "direct");
    expect(direct.kinds[0]?.commands[twoArguments.key]?.sets).toEqual({});
  });
});
