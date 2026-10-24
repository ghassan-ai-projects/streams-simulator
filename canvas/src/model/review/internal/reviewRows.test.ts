import { describe, expect, it } from "vitest";
import type { Pin, Pins } from "../../plant";
import type { Shown } from "../../graph";
import { readoutBody, reviewView } from "./reviewRows";

function pin(name: string, channel: string | null): Pin {
  return { name, type: "heat_rejection", family: "heat", connection: "connected", channel };
}

function reading(channel: string, attention: Shown["attention"] = null): Shown {
  return { channel, value: "7.0", unit: "kW", attention, word: false, outside: "", quality: "" };
}

const none = new Set<string>();

const pins: Pins = {
  ins: [pin("power_in", "power"), pin("heat_in", null)],
  outs: [pin("heat_out", "heat")],
};

describe("review rows", () => {
  it("lists every port of the kind as a row, in then out, whether or not it is observed", () => {
    const body = readoutBody(pins, ["power", "heat", "vibration"], [reading("power")], none);
    expect(body.rows.map((row) => [row.key, row.label, row.present ? row.value : null])).toEqual([
      ["in:power_in", "power", "7.0"],
      ["in:heat_in", "heat in", null],
      ["out:heat_out", "heat", null],
    ]);
  });

  it("names a row by its port when several ports report the same channel", () => {
    const shared: Pins = { ins: [], outs: [pin("flow_a", "flow"), pin("flow_b", "flow"), pin("heat_out", "heat")] };
    const body = readoutBody(shared, ["flow", "heat"], [], none);
    expect(body.rows.map((row) => row.label)).toEqual(["flow a", "flow b", "heat"]);
  });

  it("counts the observed channels without a port as hidden and carries their worst attention", () => {
    const seen = new Set(["vibration", "noise"]);
    const body = readoutBody(pins, ["power", "heat", "vibration", "noise"], [reading("noise", "alarm")], seen);
    expect(body.hidden).toBe(2);
    expect(body.hiddenAttention).toBe("alarm");
    expect(readoutBody(pins, ["power", "heat"], [], none).hidden).toBe(0);
    expect(readoutBody(pins, ["power", "heat", "vibration"], [], none).hidden).toBe(0);
  });

  it("marks a port row observed when its channel is in the run, with or without a reading now", () => {
    const body = readoutBody(pins, ["power", "heat"], [reading("power")], new Set(["power", "heat"]));
    expect(body.rows.map((row) => [row.key, row.observed, row.present])).toEqual([
      ["in:power_in", true, true],
      ["in:heat_in", false, false],
      ["out:heat_out", true, false],
    ]);
  });

  it("falls back to the first four channels for an instance without ports", () => {
    const channels = ["a", "b", "c", "d", "e", "f"];
    const body = readoutBody({ ins: [], outs: [] }, channels, [reading("b")], new Set(["b", "e", "f"]));
    expect(body.rows.map((row) => row.label)).toEqual(["a", "b", "c", "d"]);
    expect(body.rows[1]?.value).toBe("7.0");
    expect(body.hidden).toBe(2);
  });

  it("reserves a row for the more line only when channels are hidden, and at least one row", () => {
    expect(reviewView.rows(pins, ["power", "heat"])).toBe(3);
    expect(reviewView.rows(pins, ["power", "heat", "vibration"])).toBe(4);
    expect(reviewView.rows({ ins: [], outs: [] }, [])).toBe(1);
  });

  it("says why a row has no reading without guessing at a cause", () => {
    const seen = new Set(["heat"]);
    const words = (phase: "live" | "ended") =>
      readoutBody(pins, ["power", "heat"], [], seen, phase).rows.map((row) => row.absent);
    expect(words("live")).toEqual(["Awaiting first reading", "No measured channel", "No reading yet at this time"]);
    expect(words("ended")).toEqual([
      "Never observed in this run",
      "No measured channel",
      "No reading yet at this time",
    ]);
  });

  it("carries the quality of a reading to its row", () => {
    const flagged: Shown = { ...reading("power"), quality: "suspect" };
    const body = readoutBody(pins, ["power", "heat"], [flagged], none);
    expect(body.rows[0]).toMatchObject({ present: true, quality: "suspect", absent: "" });
  });
});
