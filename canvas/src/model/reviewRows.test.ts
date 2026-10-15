import { describe, expect, it } from "vitest";
import type { Pin, Pins } from "./ports";
import type { Shown } from "./readout";
import { reviewBody, reviewRowCount } from "./reviewRows";

function pin(name: string, channel: string | null): Pin {
  return { name, type: "heat_rejection", family: "heat", connection: "connected", channel };
}

function reading(channel: string, attention: Shown["attention"] = null): Shown {
  return { channel, value: "7.0", unit: "kW", attention, word: false, outside: "" };
}

const pins: Pins = {
  ins: [pin("power_in", "power"), pin("heat_in", null)],
  outs: [pin("heat_out", "heat")],
};

describe("review rows", () => {
  it("lists every port of the kind as a row, in then out, whether or not it is observed", () => {
    const body = reviewBody(pins, ["power", "heat", "vibration"], [reading("power")]);
    expect(body.rows.map((row) => [row.key, row.label, row.reading?.value ?? null])).toEqual([
      ["in:power_in", "power", "7.0"],
      ["in:heat_in", "heat in", null],
      ["out:heat_out", "heat", null],
    ]);
  });

  it("names a row by its port when several ports report the same channel", () => {
    const shared: Pins = { ins: [], outs: [pin("flow_a", "flow"), pin("flow_b", "flow"), pin("heat_out", "heat")] };
    const body = reviewBody(shared, ["flow", "heat"], []);
    expect(body.rows.map((row) => row.label)).toEqual(["flow a", "flow b", "heat"]);
  });

  it("counts channels without a port as hidden and carries their worst attention", () => {
    const body = reviewBody(pins, ["power", "heat", "vibration", "noise"], [reading("noise", "alarm")]);
    expect(body.hidden).toBe(2);
    expect(body.hiddenAttention).toBe("alarm");
    expect(reviewBody(pins, ["power", "heat"], []).hidden).toBe(0);
  });

  it("falls back to the first four channels for an instance without ports", () => {
    const none: Pins = { ins: [], outs: [] };
    const channels = ["a", "b", "c", "d", "e", "f"];
    const body = reviewBody(none, channels, [reading("b")]);
    expect(body.rows.map((row) => row.label)).toEqual(["a", "b", "c", "d"]);
    expect(body.rows[1]?.reading?.value).toBe("7.0");
    expect(body.hidden).toBe(2);
  });

  it("reserves a row for the more line only when channels are hidden, and at least one row", () => {
    expect(reviewRowCount(pins, ["power", "heat"])).toBe(3);
    expect(reviewRowCount(pins, ["power", "heat", "vibration"])).toBe(4);
    expect(reviewRowCount({ ins: [], outs: [] }, [])).toBe(1);
  });
});
