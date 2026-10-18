import { describe, expect, it } from "vitest";
import { coolerDocument, instruments, kindOf, sensorDocument } from "../../../test/twinFixture";
import {
  askedCommands,
  commandChoiceOf,
  metricChoiceOf,
  numericArguments,
  reconcileChoices,
  startingChoices,
  writableReadings,
} from "./twinDefaults";
import type { TwinChoices } from "./twinTypes";

const cooler = kindOf(coolerDocument);
const sensor = kindOf(sensorDocument);

function twin(overrides: Partial<TwinChoices> = {}): TwinChoices {
  return { ...startingChoices([cooler], instruments), ...overrides };
}

describe("what a kind file offers to choose", () => {
  it("lists the numeric arguments, the readings a command can set and the commands that ask for a choice", () => {
    expect(numericArguments(cooler.commands[0] ?? never())).toEqual(["duty", "lease_ms"]);
    expect(writableReadings(cooler)).toEqual(["duty"]);
    expect(askedCommands(cooler).map((command) => command.name)).toEqual(["set_duty"]);
    const custom = kindOf({ kind: "c", commands: [{ name: "x", args_schema: { oneOf: [] } }] });
    expect(numericArguments(custom.commands[0] ?? never())).toEqual([]);
  });
});

describe("starting choices", () => {
  it("suggests an instrument by unit, direct sets by argument name, and names the twin after the first device", () => {
    const choices = startingChoices([cooler, sensor], instruments);
    expect(choices).toMatchObject({ id: "cooler-twin", title: "cooler twin", extra: {} });
    expect(choices.kinds[0]).toEqual({
      kind: "cooler",
      instance: "cooler-01",
      metrics: {
        temperature: { instrument: "streamsim/thermistor@1", initial: null },
        duty: { instrument: "none", initial: null },
        mode: { instrument: "none", initial: null },
      },
      commands: { set_duty: { response: "direct", sets: { duty: "duty" }, drives: [] } },
    });
    expect(choices.kinds[1]?.metrics).toEqual({ humidity: { instrument: "streamsim/humidity@1", initial: null } });
    expect(startingChoices([], instruments)).toMatchObject({ id: "-twin", kinds: [] });
  });

  it("pairs the only number argument with the only settable reading, and leaves anything unclear to the author", () => {
    const lone = kindOf({
      kind: "lamp",
      metrics: [{ name: "level", data_type: "number", eu_range: { min: 0, max: 1 }, writable: true }],
      commands: [{ name: "go", args_schema: { type: "object", properties: { target: { type: "number" } } } }],
    });
    expect(startingChoices([lone], []).kinds[0]?.commands.go).toMatchObject({
      response: "direct",
      sets: { target: "level" },
    });
    const two = kindOf({
      kind: "lamp",
      metrics: [{ name: "level", data_type: "number", eu_range: { min: 0, max: 1 }, writable: true }],
      commands: [
        { name: "go", args_schema: { type: "object", properties: { a: { type: "number" }, b: { type: "number" } } } },
        { name: "none", args_schema: { type: "object", properties: {} } },
      ],
    });
    const commands = startingChoices([two], []).kinds[0]?.commands;
    expect(commands?.go).toEqual({ response: "none", sets: {}, drives: [] });
    expect(commands?.none).toEqual({ response: "none", sets: {}, drives: [] });
  });

  it("falls back to a plain choice for a name it was not given", () => {
    const kind = twin().kinds[0] ?? never();
    expect(metricChoiceOf(kind, "absent")).toEqual({ instrument: "none", initial: null });
    expect(commandChoiceOf(kind, "absent")).toEqual({ response: "none", sets: {}, drives: [] });
    expect(metricChoiceOf(kind, "temperature").instrument).toBe("streamsim/thermistor@1");
  });
});

describe("reconciling saved choices with the kind files", () => {
  it("keeps what is still asked, adds what is new and drops what is gone", () => {
    const base = twin();
    const old = base.kinds[0] ?? never();
    const saved: TwinChoices = {
      ...base,
      title: "Mine",
      kinds: [
        {
          ...old,
          instance: "c-1",
          metrics: { temperature: { instrument: "none", initial: 20 }, gone: { instrument: "none", initial: null } },
          commands: {
            set_duty: { response: "direct", sets: { duty: "duty", old: "duty", lease_ms: "gone" }, drives: [] },
            removed: { response: "none", sets: {}, drives: [] },
          },
        },
        { ...old, kind: "left" },
      ],
    };
    const result = reconcileChoices(saved, [cooler], instruments);
    expect(result.title).toBe("Mine");
    expect(result.kinds).toHaveLength(1);
    expect(result.kinds[0]?.instance).toBe("c-1");
    expect(result.kinds[0]?.metrics.temperature).toEqual({ instrument: "none", initial: 20 });
    expect(Object.keys(result.kinds[0]?.metrics ?? {})).toEqual(["temperature", "duty", "mode"]);
    expect(result.kinds[0]?.commands).toEqual({ set_duty: { response: "direct", sets: { duty: "duty" }, drives: [] } });
  });

  it("starts a kind that has no choices yet, and gives a new command its default", () => {
    const base = twin();
    const old = base.kinds[0] ?? never();
    const result = reconcileChoices({ ...base, kinds: [{ ...old, commands: {} }] }, [cooler, sensor], instruments);
    expect(result.kinds.map((kind) => kind.kind)).toEqual(["cooler", "sensor"]);
    expect(result.kinds[0]?.commands.set_duty?.response).toBe("direct");
    expect(result.kinds[1]?.instance).toBe("sensor-01");
  });

  it("drops an effect whose reading is no longer in the twin", () => {
    const base = twin();
    const old = base.kinds[0] ?? never();
    const drive = { from: "duty", kind: "cooler", metric: "temperature", gain: -1, timeConstantS: 10 };
    const away = { ...drive, kind: "sensor", metric: "humidity" };
    const driven = {
      ...old,
      commands: { set_duty: { response: "first_order" as const, sets: { duty: "duty" }, drives: [drive, away] } },
    };
    const kept = reconcileChoices({ ...base, kinds: [driven] }, [cooler], instruments);
    expect(kept.kinds[0]?.commands.set_duty?.drives).toEqual([drive]);
    const both = reconcileChoices({ ...base, kinds: [driven] }, [cooler, sensor], instruments);
    expect(both.kinds[0]?.commands.set_duty?.drives).toEqual([drive, away]);
  });
});

function never(): never {
  throw new Error("missing");
}
