import { describe, expect, it } from "vitest";
import { journal, observation } from "../test/runFixture";
import { parseJournalRecord, PlaybackIndex, shortChannel } from "./playback";

function indexOf(records: readonly unknown[]) {
  const index = new PlaybackIndex();
  for (const record of records) {
    const entry = parseJournalRecord(JSON.stringify(record));
    if (entry.kind === "observation") index.add(entry.observation);
  }
  return index;
}

describe("playback", () => {
  it("reads each kind of journal record", () => {
    expect(parseJournalRecord(JSON.stringify(journal[0]))).toEqual({
      kind: "opened",
      atNs: 0,
      mode: "evaluate",
      adapter: "native-jsonl",
    });
    expect(parseJournalRecord(JSON.stringify(journal[6]))).toEqual({
      kind: "closed",
      atNs: 3_600e9,
      incomplete: false,
    });
    expect(parseJournalRecord(JSON.stringify(journal[5]))).toEqual({ kind: "other", atNs: 30e9 });
    expect(parseJournalRecord("not json")).toEqual({ kind: "other", atNs: 0 });
    expect(parseJournalRecord(JSON.stringify(observation(9, 5, "A", "a.b", { odd: 1 })))).toMatchObject({
      observation: { value: "" },
    });
  });

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
    expect(index.readings("CH-1", 15e9)).toEqual([{ channel: "chiller.power", value: 92.7, unit: "kW", atNs: 10e9 }]);
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
