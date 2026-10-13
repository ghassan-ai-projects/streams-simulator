import { describe, expect, it } from "vitest";
import { journal, observation } from "../test/runFixture";
import { formatReading, parseJournalRecord, PlaybackIndex, shortChannel } from "./playback";

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
  });

  it("counts observations per time bucket for the timeline", () => {
    expect(indexOf(journal).density(0, 40e9, 4)).toEqual([0, 1, 2, 1]);
  });

  it("formats readings with fixed decimals and units", () => {
    const at = { channel: "x", unit: "kW", atNs: 0 };
    expect(formatReading({ ...at, value: 892.4 })).toBe("892 kW");
    expect(formatReading({ ...at, value: 27.04 })).toBe("27.0 kW");
    expect(formatReading({ ...at, value: 5.954, unit: "" })).toBe("5.95");
    expect(formatReading({ ...at, value: true })).toBe("true");
    expect(shortChannel("cooling_tower.basin_temp")).toBe("basin_temp");
    expect(shortChannel("plain")).toBe("plain");
  });
});
