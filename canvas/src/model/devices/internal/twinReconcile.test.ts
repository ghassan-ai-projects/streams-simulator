import { describe, expect, it } from "vitest";
import { cooler, coolerPath, sensor, sensorPath, startingChoices, suggestion } from "../../../test/twinFixture";
import { addMetric, changeMetric, removeItem } from "./kindEdits";
import { choicesOfDocument } from "./twinFile";
import { keyChoices } from "./twinKeys";
import { addDrive, changeDrive, chooseInstrument, chooseResponse, setArgument } from "./twinEdits";
import { askedCommands, commandChoiceOf, metricChoiceOf, numericArguments, settleCommand } from "./twinMembers";
import { reconcileChoices } from "./twinReconcile";
import type { Member } from "./types";

const at = { kind: coolerPath, command: "c1", index: 0 };
const fresh = (members: readonly Member[]) => keyChoices(choicesOfDocument(suggestion), members);

function kindChoice(choices: ReturnType<typeof startingChoices>, path = coolerPath) {
  const found = choices.kinds.find((kind) => kind.kind === path);
  if (found === undefined) throw new Error("no kind");
  return found;
}

describe("the starting choices come from the binary's suggestion", () => {
  it("takes the name, the title, the instance and the response from the suggestion, keyed by reading and command", () => {
    const start = startingChoices();
    expect(start).toMatchObject({ id: "cooler-sensor-twin", title: "Twin of cooler, sensor" });
    expect(kindChoice(start)).toMatchObject({
      kind: coolerPath,
      instance: "cooler-01",
      metrics: { m1: { instrument: "none" }, m2: { instrument: "none" }, m3: { instrument: "none" } },
      commands: { c1: { response: "direct", sets: { duty: "m2" } } },
    });
    expect(Object.keys(kindChoice(start).commands)).toEqual(["c1"]);
  });

  it("is built from neutral choices when there is no suggestion", () => {
    const start = reconcileChoices({ id: "", title: "", kinds: [], extra: {} }, [cooler], null);
    expect(kindChoice(start)).toEqual({
      kind: coolerPath,
      instance: "cooler",
      metrics: {
        m1: { instrument: "none", initial: null },
        m2: { instrument: "none", initial: null },
        m3: { instrument: "none", initial: null },
      },
      commands: { c1: { response: "none", sets: {}, drives: [] } },
    });
  });
});

describe("choices follow the device as it changes", () => {
  it("keeps every choice of a reading that is renamed, and takes the suggestion for a new one", () => {
    const chosen = chooseInstrument(startingChoices(), coolerPath, "m1", "streamsim/thermistor@1");
    const grown = addMetric(changeMetric(cooler.kind, "m1", { name: "air" }));
    const members = [{ ...cooler, kind: changeMetric(grown.kind, "m4", { name: "humidity" }) }];
    const added = {
      ...suggestion,
      kinds: [{ kind: "cooler", instance: "x", metrics: { humidity: { instrument: "h@1" } } }],
    };
    const next = reconcileChoices(chosen, members, keyChoices(choicesOfDocument(added), members));
    expect(kindChoice(next).metrics.m1?.instrument).toBe("streamsim/thermistor@1");
    expect(kindChoice(next).metrics.m4).toEqual({ instrument: "h@1", initial: null });
    expect(kindChoice(next).instance).toBe("cooler-01");
    const plain = reconcileChoices(chosen, members, fresh(members));
    expect(kindChoice(plain).metrics.m4).toEqual({ instrument: "none", initial: null });
  });

  it("drops a reading that is removed and a setting that no longer fits", () => {
    const chosen = startingChoices();
    const members = [{ ...cooler, kind: removeItem(cooler.kind, { part: "metric", key: "m2" }) }];
    const next = reconcileChoices(chosen, members, fresh(members));
    expect(Object.keys(kindChoice(next).metrics)).toEqual(["m1", "m3"]);
    expect(kindChoice(next).commands.c1?.sets).toEqual({});
    const unwritable = [{ ...cooler, kind: changeMetric(cooler.kind, "m2", { writable: false }) }];
    expect(kindChoice(reconcileChoices(chosen, unwritable, fresh(unwritable))).commands.c1?.sets).toEqual({});
  });

  it("starts a device that joins the twin from the suggestion, and drops one that leaves", () => {
    const joined = reconcileChoices(startingChoices(), [cooler, sensor], fresh([cooler, sensor]));
    expect(kindChoice(joined, sensorPath)).toMatchObject({
      instance: "sensor-01",
      metrics: { m1: { instrument: "streamsim/humidity@1" } },
    });
    const left = reconcileChoices(joined, [cooler], fresh([cooler]));
    expect(left.kinds.map((kind) => kind.kind)).toEqual([coolerPath]);
  });

  it("keeps a half-filled effect when the device changes, and blanks what it pointed at that is gone", () => {
    const first = chooseResponse(startingChoices([cooler, sensor]), coolerPath, "c1", "first_order");
    const half = changeDrive(addDrive(first, coolerPath, "c1"), { ...at, index: 1 }, { gain: 3 });
    const pointing = changeDrive(half, at, { kind: sensorPath, metric: "m1", gain: 1, timeConstantS: 10 });
    const same = reconcileChoices(pointing, [cooler, sensor], fresh([cooler, sensor]));
    expect(kindChoice(same).commands.c1?.drives).toEqual([
      { from: "m2", kind: sensorPath, metric: "m1", gain: 1, timeConstantS: 10 },
      { from: "m2", kind: "", metric: "", gain: 3, timeConstantS: null },
    ]);
    const bare: Member = { ...sensor, kind: removeItem(sensor.kind, { part: "metric", key: "m1" }) };
    const blanked = reconcileChoices(pointing, [cooler, bare], fresh([cooler, bare]));
    expect(kindChoice(blanked).commands.c1?.drives[0]).toMatchObject({ kind: "", metric: "", gain: 1 });
  });
});

describe("what a command can set", () => {
  it("lists the number arguments and the commands the twin is asked about", () => {
    const [command, stop] = cooler.kind.commands;
    if (command === undefined || stop === undefined) throw new Error("no commands");
    expect(askedCommands(cooler.kind).map((entry) => entry.name)).toEqual(["set_duty"]);
    expect(numericArguments(command)).toEqual(["duty", "lease_ms"]);
    expect(numericArguments({ ...stop, args: { kind: "custom", schema: {} } })).toEqual([]);
    const schema = {
      oneOf: [{ type: "object" }],
      properties: { position: { type: "number" }, step: { type: "integer" }, label: { type: "string" }, odd: true },
    };
    expect(numericArguments({ ...stop, args: { kind: "custom", schema } })).toEqual(["position", "step"]);
  });

  it("answers for a reading or command it has no choice for", () => {
    const choice = kindChoice(startingChoices());
    expect(metricChoiceOf(choice, "m9")).toEqual({ instrument: "none", initial: null });
    expect(commandChoiceOf(choice, "c9")).toEqual({ response: "none", sets: {}, drives: [] });
  });

  it("names the reading an effect starts from when only one reading is set, and clears it when it is no longer set", () => {
    const first = chooseResponse(startingChoices(), coolerPath, "c1", "first_order");
    expect(kindChoice(first).commands.c1?.drives[0]?.from).toBe("m2");
    const two = setArgument(first, coolerPath, "c1", "lease_ms", "m2");
    expect(kindChoice(two).commands.c1?.drives[0]?.from).toBe("m2");
    const command = kindChoice(first).commands.c1;
    if (command === undefined) throw new Error("no command");
    const spread = settleCommand({ ...command, sets: { duty: "m2", lease_ms: "m1" } });
    expect(spread.drives[0]?.from).toBe("m2");
    const [drive] = spread.drives;
    if (drive === undefined) throw new Error("no drive");
    expect(settleCommand({ ...spread, drives: [{ ...drive, from: "m3" }] }).drives[0]?.from).toBe("");
    const none = settleCommand({ ...command, sets: {} });
    expect(none.drives[0]?.from).toBe("");
  });
});
