import { describe, expect, it } from "vitest";
import { parsePlant } from "../../plant";
import { PlaybackIndex } from "./playback";
import type { Observation } from "./playback";
import type { ObservedFacts } from "./observedRows";
import { qualityWords } from "../../shared/quality";
import { ageWords, observedRows } from "./observedRows";

const plant = parsePlant({
  id: "mini",
  site: { name: "Mini", areas: [{ name: "Mini" }] },
  kinds: [
    {
      name: "tower",
      channels: [
        { name: "tower.basin_temp", unit: "Cel" },
        { name: "tower.fan_power", unit: "kW" },
        { name: "tower.fill", unit: "1" },
      ],
    },
  ],
  instances: [{ kind: "tower", names: ["CT-1"], area: "Mini" }],
  links: [],
});

const at = (second: number) => 1_000e9 + second * 1e9;
const start = at(0);

function obs(second: number, channel: string, value: number, extra: Partial<Observation> = {}): Observation {
  return { atNs: at(second), entity: "CT-1", channel, value, unit: channel.endsWith("power") ? "kW" : "Cel", ...extra };
}

function facts(index: PlaybackIndex, overrides: Partial<ObservedFacts> = {}): ObservedFacts {
  return {
    index,
    plant,
    entity: "CT-1",
    atNs: at(300),
    referenceNs: at(300),
    startNs: start,
    live: true,
    ...overrides,
  };
}

function indexOf(...observations: Observation[]): PlaybackIndex {
  const index = new PlaybackIndex();
  observations.forEach((one) => {
    index.add(one);
  });
  return index;
}

describe("the observed rows of an instance", () => {
  it("lists every channel the kind declares, the readings first", () => {
    const rows = observedRows(facts(indexOf(obs(290, "tower.fan_power", 12))));
    expect(rows.map((row) => [row.label, row.present])).toEqual([
      ["fan power", true],
      ["basin temp", false],
      ["fill", false],
    ]);
  });

  it("gives a reading its value with the unit symbol, when it was observed and how old it is", () => {
    const rows = observedRows(facts(indexOf(obs(130, "tower.basin_temp", 24.1))));
    expect(rows[0]).toMatchObject({
      value: "24.1",
      unit: "°C",
      unitTitle: "unit code: Cel",
      observedAt: "00:02:10",
      age: "3 min old",
      quality: "",
      measuredAt: "",
    });
  });

  it("measures the age from the reached clock, not from the wall", () => {
    const index = indexOf(obs(130, "tower.basin_temp", 24.1));
    const later = observedRows(facts(index, { referenceNs: at(4000) }));
    expect(later[0]?.age).toBe("1 h 5 min old");
    expect(observedRows(facts(index, { referenceNs: at(130) }))[0]?.age).toBe("just now");
  });

  it("says when the instrument measured earlier than the reading arrived", () => {
    const rows = observedRows(facts(indexOf(obs(130, "tower.basin_temp", 24.1, { lagNs: 20e9 }))));
    expect(rows[0]).toMatchObject({ observedAt: "00:02:10", measuredAt: "00:01:50" });
    expect(observedRows(facts(indexOf(obs(130, "tower.basin_temp", 24.1, { lagNs: 0.4e9 }))))[0]?.measuredAt).toBe("");
  });

  it.each([
    ["suspect", "suspect"],
    ["out_of_range", "out of range"],
    ["stale", "stale"],
    ["disconnected", "disconnected"],
    ["", ""],
  ] as const)("words the quality %s in plain words", (quality, words) => {
    const rows = observedRows(facts(indexOf(obs(130, "tower.basin_temp", 24.1, { quality }))));
    expect(rows[0]?.quality).toBe(words);
    expect(qualityWords(quality)).toBe(words);
  });

  it("says a channel awaits its first reading while the run is live, and never claims a cause", () => {
    const rows = observedRows(facts(indexOf(obs(290, "tower.fan_power", 12))));
    expect(rows.find((row) => row.label === "basin temp")?.absent).toBe("Awaiting first reading");
    expect(JSON.stringify(rows)).not.toMatch(/broken|fault|failed|offline|sensor/i);
  });

  it("says a channel was never observed once the run has ended", () => {
    const rows = observedRows(facts(indexOf(obs(290, "tower.fan_power", 12)), { live: false }));
    expect(rows.find((row) => row.label === "basin temp")?.absent).toBe("Never observed in this run");
  });

  it("says no reading yet at this time when the channel reports later in the run", () => {
    const index = indexOf(obs(500, "tower.basin_temp", 24), obs(200, "tower.fan_power", 12));
    const rows = observedRows(facts(index, { atNs: at(300), referenceNs: at(300), live: false }));
    expect(rows.find((row) => row.label === "basin temp")?.absent).toBe("No reading yet at this time");
  });

  it("draws a trend over the last window and notes how its gaps are inferred", () => {
    const samples = [10, 20, 30, 40, 50, 200, 210, 220].map((second) => obs(second, "tower.basin_temp", second / 10));
    const row = observedRows(facts(indexOf(...samples), { atNs: at(230), referenceNs: at(230) }))[0];
    expect(row?.trend?.gaps).toBe(1);
    expect(row?.trendNote).toContain("1 gap left blank");
    expect(row?.trendNote).toContain("inferred from the observed intervals");
    expect(row?.trendNote).toContain("from 1 to 22 °C");
  });

  it("puts a channel in alarm before the others", () => {
    const alarmed = parsePlant({
      id: "p",
      site: { name: "P", areas: [{ name: "P" }] },
      kinds: [
        {
          name: "tower",
          channels: [
            { name: "tower.a", unit: "kW" },
            { name: "tower.b", unit: "", states: [{ value: 1, name: "tripped", attention: "alarm" }] },
          ],
        },
      ],
      instances: [{ kind: "tower", names: ["CT-1"], area: "P" }],
      links: [],
    });
    const index = indexOf(obs(10, "tower.a", 5), obs(11, "tower.b", 1));
    expect(observedRows(facts(index, { plant: alarmed })).map((row) => row.label)).toEqual(["b", "a"]);
  });

  it("has no rows for an entity that is not an instance and never reported", () => {
    expect(observedRows(facts(indexOf(), { entity: "hall-1" }))).toEqual([]);
    expect(observedRows(facts(indexOf(), { plant: null }))).toEqual([]);
  });

  it("lists a channel the kind does not declare when it reported", () => {
    const rows = observedRows(facts(indexOf(obs(10, "tower.surprise", 3))));
    expect(rows.map((row) => row.key)).toContain("tower.surprise");
  });
});

describe("the age of a reading", () => {
  it.each([
    [0, "just now"],
    [1.4e9, "just now"],
    [30e9, "30 s old"],
    [180e9, "3 min old"],
    [3900e9, "1 h 5 min old"],
    [-5e9, "just now"],
  ])("words %d ns as %s", (ns, words) => {
    expect(ageWords(ns)).toBe(words);
  });
});
