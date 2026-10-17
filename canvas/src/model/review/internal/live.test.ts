import { describe, expect, it } from "vitest";
import { parsePlant } from "../../plant";
import { domain } from "../../../test/plantFixture";
import type { LiveFacts } from "./live";
import { hallLiveOf, liveOf } from "./live";
import { PlaybackIndex } from "./playback";

const alarming = {
  ...domain,
  kinds: [
    {
      name: "chiller",
      channels: [
        {
          name: "chiller.alarm_state",
          states: [
            { value: 0, name: "normal" },
            { value: 1, name: "unacknowledged", attention: "alarm" },
          ],
        },
      ],
    },
  ],
  instances: [
    { kind: "chiller", names: ["A/CH-1", "A/CH-2", "B/CH-3"], area: "Cooling" },
    { kind: "chiller", names: ["CH-9"], area: "Cooling" },
  ],
  links: [],
};
const plant = parsePlant(alarming);

function indexed(): PlaybackIndex {
  const index = new PlaybackIndex();
  const base = { channel: "chiller.alarm_state", unit: "" } as const;
  index.add({ ...base, atNs: 1e9, entity: "A/CH-1", value: 1 });
  index.add({ ...base, atNs: 1e9, entity: "A/CH-2", value: 0 });
  index.add({ ...base, atNs: 1e9, entity: "CH-9", value: 0 });
  return index;
}

function facts(overrides: Partial<LiveFacts>): LiveFacts {
  return {
    index: indexed(),
    plant,
    view: null,
    faulted: new Map(),
    effectors: new Map(),
    positionNs: 2e9,
    ...overrides,
  };
}

describe("a node's live values", () => {
  it("shows the readings at the position and lets an alarm win over a fault", () => {
    const alarm = liveOf(facts({ faulted: new Map([["A/CH-1", ["Fouling"]]]) }), "A/CH-1");
    expect(alarm).toMatchObject({ status: "alarm", fault: "Fouling", count: 0, effector: "" });
    expect(alarm.shown.map((entry) => entry.value)).toEqual(["unacknowledged"]);
  });

  it("is faulted without an alarm, and plain when neither", () => {
    const faulted = new Map([["CH-9", ["Fouling", "Fan failure"]]]);
    expect(liveOf(facts({ faulted }), "CH-9")).toMatchObject({ status: "fault", fault: "Fouling, Fan failure" });
    expect(liveOf(facts({}), "CH-9").status).toBeNull();
  });

  it("carries the effector words, and an empty readout without a view", () => {
    const live = liveOf(facts({ effectors: new Map([["CH-9", "Start · accepted"]]) }), "CH-9");
    expect(live.effector).toBe("Start · accepted");
    expect(live.readout).toEqual({ rows: [], hidden: 0, hiddenAttention: null });
  });

  it("lets a position before the first reading show nothing", () => {
    expect(liveOf(facts({ positionNs: 0 }), "A/CH-1")).toMatchObject({ shown: [], status: null });
  });
});

describe("a hall's live values", () => {
  it("takes the worst status of its members and counts the members at that status", () => {
    const faulted = new Map([["A/CH-2", ["Fouling"]]]);
    expect(hallLiveOf(facts({ faulted }), "A")).toMatchObject({ status: "alarm", count: 1, shown: [] });
    expect(hallLiveOf(facts({ positionNs: 0, faulted }), "A")).toMatchObject({ status: "fault", count: 1 });
  });

  it("falls back to the node itself for an id that is no hall, and without a plant", () => {
    expect(hallLiveOf(facts({}), "CH-9")).toMatchObject({ status: null, count: 0 });
    expect(hallLiveOf(facts({ plant: null }), "A/CH-1").shown).toHaveLength(1);
  });
});
