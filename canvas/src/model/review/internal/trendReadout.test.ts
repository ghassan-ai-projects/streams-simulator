import { describe, expect, it } from "vitest";
import { PlaybackIndex } from "./playback";
import { readoutOf } from "./trendReadout";

const watched = { entity: "CT-1", channel: "cooling_tower.basin_temp" };

function indexed(): PlaybackIndex {
  const index = new PlaybackIndex();
  index.add({ atNs: 10e9, ...watched, value: 24.5, unit: "degC" });
  index.add({ atNs: 20e9, ...watched, value: 25.04, unit: "degC", quality: "suspect" });
  return index;
}

describe("the exact reading under the cursor", () => {
  it("is the newest reading at or before the cursor with its unit, time, age and quality", () => {
    const readout = readoutOf({ index: indexed(), plant: null, atNs: 25e9, referenceNs: 90e9, startNs: 0 }, watched);
    expect(readout).toMatchObject({
      value: "25.0",
      unit: "°C",
      observedAt: "00:00:20",
      age: "1 min old",
      quality: "suspect",
    });
    expect(readout?.words).toBe("25.0 °C, suspect, observed at 00:00:20, 1 min old");
  });

  it("reads an earlier reading when the cursor is earlier and has none before the first one", () => {
    const facts = { index: indexed(), plant: null, referenceNs: 12e9, startNs: 0 };
    expect(readoutOf({ ...facts, atNs: 15e9 }, watched)).toMatchObject({ value: "24.5", quality: "", age: "2 s old" });
    expect(readoutOf({ ...facts, atNs: 5e9 }, watched)).toBeNull();
    expect(readoutOf({ ...facts, atNs: 15e9 }, { entity: "CT-9", channel: "x.y" })).toBeNull();
  });
});
