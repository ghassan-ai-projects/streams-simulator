import { describe, expect, it } from "vitest";
import { codeTitle, displayUnit, grouped, groupedFigure, presentedUnits, withUnit } from "./units";

describe("unit grouping", () => {
  it("groups thousands with a thin space", () => {
    expect([grouped(7), grouped(14_483), grouped(1_234_567)]).toEqual(["7", "14 483", "1 234 567"]);
  });
});

describe("unit symbols", () => {
  it.each([
    ["Cel", "°C"],
    ["degC", "°C"],
    ["degF", "°F"],
    ["1", ""],
    ["rel", ""],
    ["frac", ""],
    ["ratio", ""],
    ["%", "%"],
    ["percent", "%"],
    ["%RH", "% RH"],
    ["permille", "‰"],
    ["m3/h", "m³/h"],
    ["m3", "m³"],
    ["kWh/m3", "kWh/m³"],
    ["N.m", "N·m"],
    ["kW", "kW"],
    ["rpm", "rpm"],
    ["a-unit-nobody-listed", "a-unit-nobody-listed"],
  ])("shows %s as %j", (code, symbol) => {
    expect(displayUnit(code)).toBe(symbol);
  });

  it("keeps the code for the tooltip only where the symbol differs from it", () => {
    expect([codeTitle("Cel"), codeTitle("1"), codeTitle("kW"), codeTitle("")]).toEqual([
      "unit code: Cel",
      "unit code: 1",
      "",
      "",
    ]);
  });

  it("joins a value and its symbol, and leaves a dimensionless value bare", () => {
    expect([withUnit("128", "1"), withUnit("21.5", "Cel"), withUnit("on", ""), withUnit("3", "kW")]).toEqual([
      "128",
      "21.5 °C",
      "on",
      "3 kW",
    ]);
  });

  it("lists every code the table presents, each once", () => {
    const codes = presentedUnits();
    expect(new Set(codes).size).toBe(codes.length);
    expect(codes).toEqual(expect.arrayContaining(["Cel", "1", "kW", "%", "rpm", "Hz"]));
  });

  it("groups the thousands of a printed figure, and leaves its sign, decimals and small values alone", () => {
    expect(["180013", "-1234567.5", "999", "1000", "12.345", "0.5", "", "state 1"].map(groupedFigure)).toEqual([
      "180 013",
      "-1 234 567.5",
      "999",
      "1 000",
      "12.345",
      "0.5",
      "",
      "state 1",
    ]);
  });
});
