import { describe, expect, it } from "vitest";
import { journalWire, observation } from "../../../test/runFixture";
import { journalEntry } from "./journalDecoder";

const line = (index: number): string => JSON.stringify(journalWire[index]);

describe("journal lines", () => {
  it("reads each kind of journal record", () => {
    expect(journalEntry(line(0))).toEqual({
      kind: "opened",
      atNs: 0,
      mode: "evaluate",
      adapter: "native-jsonl",
    });
    expect(journalEntry(line(6))).toEqual({
      kind: "closed",
      atNs: 3_600e9,
      incomplete: false,
    });
    expect(journalEntry(line(5))).toEqual({ kind: "other", atNs: 30e9 });
  });

  it("reads a reading's scalar value and strings anything else", () => {
    expect(journalEntry(line(1))).toEqual({
      kind: "observation",
      observation: { atNs: 10e9, entity: "CH-1", channel: "chiller.power", value: 92.7, unit: "kW" },
    });
    expect(journalEntry(line(4))).toMatchObject({ observation: { value: true, unit: "" } });
    expect(journalEntry(JSON.stringify(observation(9, 5, "A", "a.b", { odd: 1 })))).toMatchObject({
      observation: { value: "" },
    });
  });

  it("refuses a line that is not JSON, or not a journal record", () => {
    expect(journalEntry("not json")).toBeNull();
    expect(journalEntry(JSON.stringify({ type: "run.opened", at_ns: 0, seq: 0 }))).toBeNull();
    expect(journalEntry(JSON.stringify({ ...journalWire[5], at_ns: "soon" }))).toBeNull();
    expect(journalEntry(JSON.stringify({ ...journalWire[5], seq: 1.5 }))).toBeNull();
    expect(journalEntry(JSON.stringify({ ...journalWire[4], line: undefined }))).toBeNull();
  });
});
