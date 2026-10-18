import { describe, expect, it } from "vitest";
import { cooler, coolerPath, sensor, sensorPath, startingChoices } from "../../../test/twinFixture";
import { changeMetric, removeItem } from "./kindEdits";
import { changeDrive, chooseResponse, startAt } from "./twinEdits";
import { keyChoices, nameChoices, serialiseChoices } from "./twinKeys";
import type { TwinChoices } from "./twinTypes";
import type { Member } from "./types";

const at = { kind: coolerPath, command: "c1", index: 0 };

function driven(): TwinChoices {
  const first = chooseResponse(startingChoices([cooler, sensor]), coolerPath, "c1", "first_order");
  const follows = changeDrive(first, at, { kind: sensorPath, metric: "m1", gain: 2, timeConstantS: 30 });
  return startAt(follows, coolerPath, "m1", 30);
}

function renamed(): Member {
  return { ...cooler, kind: changeMetric(cooler.kind, "m1", { name: "air_temperature" }) };
}

describe("choices kept by the key of a reading, written by its name", () => {
  it("names every key for the file and finds every name again", () => {
    const choices = driven();
    const named = nameChoices(choices, [cooler, sensor]);
    expect(named.kinds[0]).toMatchObject({
      kind: "cooler",
      metrics: { temperature: { initial: 30 }, duty: {}, mode: {} },
      commands: {
        set_duty: {
          sets: { duty: "duty" },
          drives: [{ from: "duty", kind: "sensor", metric: "humidity", gain: 2, timeConstantS: 30 }],
        },
      },
    });
    expect(keyChoices(named, [cooler, sensor])).toEqual(choices);
  });

  it("keeps the instrument, the starting value and the effects of a reading that is renamed", () => {
    const choices = driven();
    const named = nameChoices(choices, [renamed(), sensor]);
    expect(Object.keys(named.kinds[0]?.metrics ?? {})).toEqual(["air_temperature", "duty", "mode"]);
    expect(named.kinds[0]?.metrics.air_temperature?.initial).toBe(30);
    expect(serialiseChoices(choices, [renamed(), sensor])).toContain('"air_temperature"');
  });

  it("drops what the device no longer has and blanks what an effect pointed at", () => {
    const choices = driven();
    const lost: Member = { ...sensor, kind: removeItem(sensor.kind, { part: "metric", key: "m1" }) };
    const named = nameChoices(choices, [cooler, lost]);
    expect(named.kinds[0]?.commands.set_duty?.drives[0]).toMatchObject({ kind: "sensor", metric: "" });
    expect(nameChoices(choices, [sensor]).kinds.map((kind) => kind.kind)).toEqual(["sensor"]);
    const gone = nameChoices(choices, [{ ...cooler, kind: removeItem(cooler.kind, { part: "metric", key: "m2" }) }]);
    expect(gone.kinds[0]?.commands.set_duty?.sets).toEqual({});
  });

  it("leaves a choice out when the file names a reading, a command or a device that does not exist", () => {
    const named = nameChoices(driven(), [cooler, sensor]);
    const kind = named.kinds[0];
    if (kind === undefined) throw new Error("no kind");
    const odd = {
      ...named,
      kinds: [
        {
          ...kind,
          metrics: { ...kind.metrics, ghost: { instrument: "none", initial: null } },
          commands: {
            ...kind.commands,
            ghost: { response: "direct" as const, sets: { duty: "nothing" }, drives: [] },
            set_duty: { response: "direct" as const, sets: { duty: "nothing" }, drives: [] },
          },
        },
        { ...kind, kind: "nowhere" },
      ],
    };
    const keyed = keyChoices(odd, [cooler, sensor]);
    expect(keyed.kinds).toHaveLength(1);
    expect(Object.keys(keyed.kinds[0]?.metrics ?? {})).toEqual(["m1", "m2", "m3"]);
    expect(Object.keys(keyed.kinds[0]?.commands ?? {})).toEqual(["c1"]);
    expect(keyed.kinds[0]?.commands.c1?.sets).toEqual({});
  });
});
