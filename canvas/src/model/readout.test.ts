import { describe, expect, it } from "vitest";
import { attentionOf, showReading } from "./readout";
import { displayUnit } from "./units";

const at = (channel: string, value: number | string | boolean, unit = "") => ({ channel, value, unit, atNs: 0 });

describe("readouts", () => {
  it("decodes an annunciator into its documented states and attention", () => {
    const of = (code: number) => showReading(at("air_compressor.alarm_state", code, "on"), "streamsim/annunciator@1");
    expect([0, 1, 2, 3, 4, 9].map((code) => of(code).value)).toEqual([
      "normal",
      "unacknowledged",
      "acknowledged",
      "returned, unacknowledged",
      "shelved",
      "code 9",
    ]);
    expect([0, 1, 2, 3, 4].map((code) => of(code).attention)).toEqual([null, "alarm", "alarm", "warn", null]);
  });

  it("shows numbers with fixed decimals and the unit apart, and other values as words", () => {
    expect(showReading(at("chiller.cond_temp", 33.14, "degC"), "streamsim/rtd@1")).toEqual({
      channel: "cond_temp",
      value: "33.1",
      unit: "°C",
      attention: null,
    });
    expect(showReading(at("meter.power", 852.4, "kW"), "")).toMatchObject({ value: "852", unit: "kW" });
    expect(showReading(at("x.cop", 5.954), "")).toMatchObject({ value: "5.95", unit: "" });
    expect(showReading(at("x.run_status", true), "streamsim/run-status@1")).toMatchObject({ value: "true", unit: "" });
  });

  it("raises a node's attention to its worst readout", () => {
    const shown = (attention: "alarm" | "warn" | null) => ({ channel: "c", value: "v", unit: "", attention });
    expect(attentionOf([shown(null), shown("warn")])).toBe("warn");
    expect(attentionOf([shown("warn"), shown("alarm")])).toBe("alarm");
    expect(attentionOf([])).toBeNull();
  });

  it("writes units the way people read them", () => {
    expect(["degC", "m3/h", "kWh/m3", "kPa"].map(displayUnit)).toEqual(["°C", "m³/h", "kWh/m³", "kPa"]);
  });
});
