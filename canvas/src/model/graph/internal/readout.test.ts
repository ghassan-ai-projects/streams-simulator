import { describe, expect, it } from "vitest";
import { abnormalFirst, attentionOf, channelWords, pinWords, sharedTimes, showReading, worstOf } from "./readout";
import { displayUnit } from "../../shared/units";

const at = (channel: string, value: number | string | boolean, unit = "") => ({ channel, value, unit, atNs: 0 });

const annunciator = [
  { value: 0, name: "normal", attention: null },
  { value: 1, name: "unacknowledged", attention: "alarm" },
  { value: 2, name: "acknowledged", attention: "alarm" },
  { value: 3, name: "returned unacknowledged", attention: "warn" },
  { value: 4, name: "shelved", attention: null },
] as const;

describe("readouts", () => {
  it("names a channel value from the vocabulary the domain declares for it", () => {
    const of = (code: number) =>
      showReading(at("air_compressor.alarm_state", code, "on"), { states: annunciator, range: null });
    expect([0, 1, 2, 3, 4, 9].map((code) => of(code).value)).toEqual([
      "normal",
      "unacknowledged",
      "acknowledged",
      "returned unacknowledged",
      "shelved",
      "state 9",
    ]);
    expect([0, 1, 2, 3, 4].map((code) => of(code).attention)).toEqual([null, "alarm", "alarm", "warn", null]);
  });

  it("shows numbers with fixed decimals and the unit apart, and other values as words", () => {
    expect(showReading(at("chiller.cond_temp", 33.14, "degC"), null)).toEqual({
      channel: "cond_temp",
      value: "33.1",
      unit: "°C",
      attention: null,
      word: false,
      outside: "",
    });
    expect(showReading(at("meter.power", 852.4, "kW"), null)).toMatchObject({ value: "852", unit: "kW" });
    expect(showReading(at("x.cop", 5.954), null)).toMatchObject({ value: "5.95", unit: "" });
    expect(showReading(at("x.run_status", true), null)).toMatchObject({ value: "on", unit: "", word: true });
    expect(showReading(at("x.mode", "auto"), null)).toMatchObject({ value: "auto", word: true });
  });

  it("flags a number outside the instrument's declared range", () => {
    const flow = { states: [], range: { min: 0, max: 600 } };
    expect(showReading(at("molder.air_flow", -4.6, "m3/h"), flow).outside).toBe(
      "Outside the instrument's range of 0.00 to 600 m³/h",
    );
    expect(showReading(at("molder.air_flow", 12, "m3/h"), flow).outside).toBe("");
  });

  it("raises a node's attention to its worst readout", () => {
    const shown = (attention: "alarm" | "warn" | null) => ({
      channel: "c",
      value: "v",
      unit: "",
      attention,
      word: true,
      outside: "",
    });
    expect(attentionOf([shown(null), shown("warn")])).toBe("warn");
    expect(attentionOf([shown("warn"), shown("alarm")])).toBe("alarm");
    expect(attentionOf([])).toBeNull();
  });

  it("writes units the way people read them", () => {
    expect(["degC", "m3/h", "kWh/m3", "kPa"].map(displayUnit)).toEqual(["°C", "m³/h", "kWh/m³", "kPa"]);
  });

  it("puts alarmed and warned channels first and names channels in words", () => {
    const entry = (channel: string, attention: "alarm" | "warn" | null) => ({
      channel,
      value: "1",
      unit: "",
      attention,
      word: false,
      outside: "",
    });
    const order = abnormalFirst([entry("a", null), entry("b", "warn"), entry("c", "alarm"), entry("d", null)]);
    expect(order.map((shown) => shown.channel)).toEqual(["c", "b", "a", "d"]);
    expect(channelWords("chw_supply_temp")).toBe("chw supply temp");
    expect(pinWords("cond_water_in", "in")).toEqual({ full: "cond water in", short: "cond water" });
    expect(pinWords("power_out", "out")).toEqual({ full: "power out", short: "power" });
    expect(pinWords("power_out", "in")).toEqual({ full: "power out", short: "power out" });
    expect(pinWords("in", "in")).toEqual({ full: "in", short: "in" });
  });

  it("rolls a hall up to its worst status and how many members share it", () => {
    expect(worstOf(["warn", null, "fault", "alarm", "alarm"])).toEqual({ status: "alarm", count: 2 });
    expect(worstOf(["warn", "fault"])).toEqual({ status: "fault", count: 1 });
    expect(worstOf([null, null])).toEqual({ status: null, count: 0 });
  });
});

describe("shared observation times", () => {
  it("gives the latest time once and only the rows that differ their own time", () => {
    expect(sharedTimes([20e9, 20e9])).toEqual({ latest: 20e9, own: [null, null] });
    expect(sharedTimes([15e9, 20e9, 15e9])).toEqual({ latest: 20e9, own: [15e9, null, 15e9] });
  });
});
