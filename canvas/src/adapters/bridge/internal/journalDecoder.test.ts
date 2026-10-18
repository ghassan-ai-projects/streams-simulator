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

  it("reads where a device comes from and how a gateway matched the session plan", () => {
    const envelope = { stream: "observer", seq: 3, at_ns: 5e9 };
    const origin = {
      ...envelope,
      type: "origin",
      instance: "zone-01",
      device: "zone-01",
      board: "uno-07",
      boot_id: "61",
    };
    expect(journalEntry(JSON.stringify({ ...origin, environment: "hardware" }))).toEqual({
      kind: "origin",
      atNs: 5e9,
      origin: { instance: "zone-01", kind: "real", board: "uno-07", device: "zone-01", bootId: "61" },
    });
    expect(journalEntry(JSON.stringify({ ...origin, environment: "simulated", instance: undefined }))).toMatchObject({
      origin: { instance: "", kind: "simulated" },
    });
    const issue = { pointer: "/real/0/device", message: "no device", keyword: "device_missing", member: "device" };
    const refused = {
      ...envelope,
      type: "binding",
      outcome: "refused",
      refusal: "inventory_mismatch",
      reason: "no device",
      issues: [issue],
    };
    expect(journalEntry(JSON.stringify(refused))).toEqual({
      kind: "binding",
      atNs: 5e9,
      binding: { matched: false, refusal: "inventory_mismatch", reason: "no device", issues: [issue] },
    });
    expect(
      journalEntry(JSON.stringify({ ...envelope, type: "binding", outcome: "matched", issues: [] })),
    ).toMatchObject({
      binding: { matched: true, refusal: "", issues: [] },
    });
  });

  it("refuses an origin or a binding that is not one", () => {
    const envelope = { stream: "observer", seq: 3, at_ns: 5e9 };
    expect(journalEntry(JSON.stringify({ ...envelope, type: "origin", device: "d", environment: "mixed" }))).toBeNull();
    expect(
      journalEntry(JSON.stringify({ ...envelope, type: "origin", device: "", environment: "hardware" })),
    ).toBeNull();
    expect(journalEntry(JSON.stringify({ ...envelope, type: "binding", outcome: "maybe", issues: [] }))).toBeNull();
    expect(journalEntry(JSON.stringify({ ...envelope, type: "binding", outcome: "matched" }))).toBeNull();
  });

  it("refuses a line that is not JSON, or not a journal record", () => {
    expect(journalEntry("not json")).toBeNull();
    expect(journalEntry(JSON.stringify({ type: "run.opened", at_ns: 0, seq: 0 }))).toBeNull();
    expect(journalEntry(JSON.stringify({ ...journalWire[5], at_ns: "soon" }))).toBeNull();
    expect(journalEntry(JSON.stringify({ ...journalWire[5], seq: 1.5 }))).toBeNull();
    expect(journalEntry(JSON.stringify({ ...journalWire[4], line: undefined }))).toBeNull();
  });
});
