import { describe, expect, it } from "vitest";
import { journal } from "../../../test/runFixture";
import type { JournalEntry } from "./playback";
import { PlaybackIndex } from "./playback";
import type { Observation } from "./playback";
import { shortChannel } from "../../shared/words";

function indexOf(entries: readonly JournalEntry[]) {
  const index = new PlaybackIndex();
  for (const entry of entries) {
    if (entry.kind === "observation") index.add(entry.observation);
  }
  return index;
}

describe("playback", () => {
  it("names the channels an entity reports in the run, short, and none for an unknown entity", () => {
    const index = indexOf(journal);
    expect([...index.observed("CT-1")].sort()).toEqual(["basin_temp", "run_status"]);
    expect(index.observed("nobody").size).toBe(0);
  });

  it("answers the last reading at or before a time for every channel", () => {
    const index = indexOf(journal);
    expect(index.size).toBe(4);
    expect(index.bounds()).toEqual({ startNs: 10e9, endNs: 30e9 });
    expect(index.readings("CH-1", 5e9)).toEqual([]);
    expect(index.readings("CH-1", 15e9)).toEqual([
      { channel: "chiller.power", value: 92.7, unit: "kW", atNs: 10e9, quality: "", lagNs: 0 },
    ]);
    expect(index.readings("CH-1", 99e9)[0]?.value).toBe(95.1);
    expect(index.readings("CT-1", 30e9).map((reading) => reading.channel)).toEqual([
      "tower.basin_temp",
      "tower.run_status",
    ]);
    expect(index.readings("nobody", 30e9)).toEqual([]);
    expect(new PlaybackIndex().bounds()).toBeNull();
    expect(new PlaybackIndex().settled()).toBeNull();
    expect(index.settled()).toBe(12e9);
  });

  it("settles when every channel has reported once", () => {
    const index = new PlaybackIndex();
    const at = (atNs: number, channel: string) => ({ entity: "A", channel, value: 1, unit: "", atNs });
    [at(0, "a.x"), at(5e9, "a.y"), at(100e9, "a.x")].forEach((observation) => {
      index.add(observation);
    });
    expect(index.settled()).toBe(5e9);
  });

  it("says whether playback opened after every channel reported, or before it", () => {
    const index = new PlaybackIndex();
    const at = (atNs: number, channel: string) => ({ entity: "A", channel, value: 1, unit: "", atNs });
    [at(0, "a.x"), at(5e9, "a.y"), at(100e9, "a.x")].forEach((observation) => {
      index.add(observation);
    });
    expect(index.opening()).toEqual({ atNs: 5e9, full: true });
    const late = new PlaybackIndex();
    [at(0, "a.x"), at(5e9, "a.y"), at(200e9, "a.z")].forEach((observation) => {
      late.add(observation);
    });
    expect(late.opening()).toEqual({ atNs: 20e9, full: false });
    expect(new PlaybackIndex().opening()).toBeNull();
  });

  it("names a channel without its kind", () => {
    expect(shortChannel("cooling_tower.basin_temp")).toBe("basin_temp");
    expect(shortChannel("plain")).toBe("plain");
  });
});

const sample = (atNs: number, value: number, extra: Partial<Observation> = {}): Observation => ({
  entity: "A",
  channel: "a.x",
  value,
  unit: "kW",
  atNs,
  ...extra,
});

describe("playback samples carry what the instrument said about them", () => {
  it("keeps the quality and how long before arrival the value was measured", () => {
    const index = new PlaybackIndex();
    index.add(sample(10e9, 1));
    index.add(sample(20e9, 2, { quality: "suspect", lagNs: 4e9 }));
    expect(index.readings("A", 15e9)[0]).toMatchObject({ value: 1, quality: "", lagNs: 0 });
    expect(index.readings("A", 25e9)[0]).toMatchObject({ value: 2, quality: "suspect", lagNs: 4e9, atNs: 20e9 });
  });

  it("holds the run clock the journal reported and never lets it run backwards", () => {
    const index = new PlaybackIndex();
    expect(index.clockNs).toBe(0);
    index.advanceClock(60e9);
    index.advanceClock(30e9);
    expect(index.clockNs).toBe(60e9);
  });

  it("slices the samples of one channel inside a window, with the last one at or before its end", () => {
    const index = new PlaybackIndex();
    [10, 20, 30, 40, 50].forEach((second) => {
      index.add(sample(second * 1e9, second));
    });
    expect(index.samples("A", "a.x", 20e9, 40e9)).toEqual({ times: [20e9, 30e9, 40e9], values: [20, 30, 40] });
    expect(index.samples("A", "a.x", 41e9, 49e9)).toEqual({ times: [], values: [] });
    expect(index.samples("A", "nope", 0, 99e9)).toBeNull();
    expect(index.samples("B", "a.x", 0, 99e9)).toBeNull();
  });
});

describe("a bounded playback index", () => {
  function filled(capacity: number, count: number) {
    const index = new PlaybackIndex({ capacity });
    for (let at = 1; at <= count; at += 1) index.add(sample(at * 1e9, at));
    return index;
  }

  it("keeps at least the newest capacity samples of a channel and never more than twice that", () => {
    const index = filled(100, 1000);
    const kept = index.samples("A", "a.x", 0, 2000e9)?.times.length ?? 0;
    expect(kept).toBeGreaterThanOrEqual(100);
    expect(kept).toBeLessThan(200);
    expect(index.samples("A", "a.x", 0, 2000e9)?.times.at(-1)).toBe(1000e9);
  });

  it("still counts every record it was given and moves the start of what it holds", () => {
    const index = filled(100, 1000);
    expect(index.size).toBe(1000);
    expect(index.bounds()?.endNs).toBe(1000e9);
    expect(index.bounds()?.startNs).toBeGreaterThan(1e9);
    expect(index.readings("A", 1000e9)[0]?.value).toBe(1000);
    expect(index.readings("A", 5e9)).toEqual([]);
  });

  it("counts the readings it holds across every channel", () => {
    const index = filled(100, 1000);
    index.add({ ...sample(1001e9, 1), channel: "a.y" });
    expect(index.held).toBe((index.samples("A", "a.x", 0, 1e15)?.times.length ?? 0) + 1);
    expect(new PlaybackIndex().held).toBe(0);
  });

  it("holds the same memory after a hundred times more records", () => {
    const small = filled(100, 1000);
    const large = filled(100, 100_000);
    const held = (index: PlaybackIndex) => index.samples("A", "a.x", 0, 1e15)?.times.length ?? 0;
    expect(held(large)).toBeLessThan(200);
    expect(held(small)).toBeGreaterThanOrEqual(100);
  });

  it("is unbounded unless asked", () => {
    const index = new PlaybackIndex();
    for (let at = 1; at <= 5000; at += 1) index.add(sample(at * 1e9, at));
    expect(index.samples("A", "a.x", 0, 1e15)?.times.length).toBe(5000);
  });
});
