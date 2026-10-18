import { describe, expect, it } from "vitest";
import { convertibleThermistor, humidity, instruments, metricOf, thermistor } from "../../../test/twinFixture";
import {
  chosenInstrumentWords,
  conversionNote,
  instrumentCount,
  instrumentOf,
  instrumentWords,
  noInstrument,
  rangeWords,
  searchInstruments,
  takesInstrument,
} from "./instruments";

describe("instruments", () => {
  it("is built from a library entry as the binary lists it for a unit", () => {
    const known = instrumentOf({
      ref: "a/b@1",
      title: "Probe",
      unit: "rpm",
      min: 0,
      max: 9,
      match: "exact",
    });
    expect(known).toMatchObject({ ref: "a/b@1", title: "Probe", unit: "rpm", min: 0, max: 9, match: "exact" });
    expect(known.words).toBe("probe a/b@1 rpm");
  });

  it("searches every term in the words and keeps the order the binary ranked", () => {
    expect(searchInstruments(instruments, "").map((entry) => entry.title)).toEqual([
      "Thermistor",
      "Capacitive humidity sensor",
      "Tachometer",
    ]);
    expect(searchInstruments(instruments, "DEGC therm").map((entry) => entry.title)).toEqual(["Thermistor"]);
    expect(searchInstruments(instruments, "nothing here")).toEqual([]);
  });

  it("states a range in words whatever bounds it has", () => {
    expect(rangeWords(0, 10)).toBe("0 to 10");
    expect(rangeWords(5, null)).toBe("from 5");
    expect(rangeWords(null, 5)).toBe("up to 5");
    expect(rangeWords(null, null)).toBe("no stated range");
  });

  it("says when an instrument measures a different unit and needs a conversion", () => {
    expect(instrumentWords(thermistor)).toBe("Thermistor · degC · -40 to 125");
    expect(instrumentWords(convertibleThermistor)).toBe(`Thermistor · degC · -40 to 125 · ${conversionNote}`);
    expect(conversionNote).toBe("different unit — needs conversion");
  });

  it("describes a choice and counts a search", () => {
    expect(chosenInstrumentWords(noInstrument, instruments)).toBe("No instrument: a plain reading");
    expect(chosenInstrumentWords("streamsim/humidity@1", instruments)).toBe(instrumentWords(humidity));
    expect(chosenInstrumentWords("gone@1", instruments)).toBe("gone@1");
    expect(instrumentCount(0, " zz ")).toBe('Nothing in the library matches "zz".');
    expect(instrumentCount(1, "")).toBe("1 instrument");
    expect(instrumentCount(3, "")).toBe("3 instruments");
  });

  it("takes an instrument for numbers only", () => {
    expect(takesInstrument(metricOf())).toBe(true);
    expect(takesInstrument(metricOf({ data_type: "integer" }))).toBe(true);
    expect(takesInstrument(metricOf({ data_type: "enum" }))).toBe(false);
  });
});
