import { describe, expect, it } from "vitest";
import { humidity, instruments, metricOf, tachometer, thermistor } from "../../../test/twinFixture";
import {
  chosenInstrumentWords,
  instrumentCount,
  instrumentOf,
  instrumentWords,
  noInstrument,
  rangeWords,
  searchInstruments,
  sortInstruments,
  suggestInstrument,
  takesInstrument,
} from "./instruments";

describe("instruments", () => {
  it("is built from a library entry and what the binary says about it, or from the entry alone", () => {
    const known = instrumentOf({ ref: "a/b@1", title: "Probe" }, { family: "speed", unit: "rpm", min: 0, max: 9 });
    expect(known).toMatchObject({ ref: "a/b@1", title: "Probe", unit: "rpm", min: 0, max: 9 });
    expect(known.words).toBe("probe a/b@1 speed rpm");
    const bare = instrumentOf({ ref: "a/c@1", title: "Bare" }, null);
    expect(bare).toMatchObject({ family: "", unit: "", min: null, max: null });
  });

  it("sorts by title and searches every term in the words", () => {
    expect(sortInstruments([thermistor, humidity]).map((entry) => entry.title)).toEqual([
      "Capacitive humidity sensor",
      "Thermistor",
    ]);
    expect(searchInstruments(instruments, "")).toHaveLength(3);
    expect(searchInstruments(instruments, "TEMPERATURE therm").map((entry) => entry.title)).toEqual(["Thermistor"]);
    expect(searchInstruments(instruments, "nothing here")).toEqual([]);
  });

  it("states a range in words whatever bounds it has", () => {
    expect(rangeWords(0, 10)).toBe("0 to 10");
    expect(rangeWords(5, null)).toBe("from 5");
    expect(rangeWords(null, 5)).toBe("up to 5");
    expect(rangeWords(null, null)).toBe("no stated range");
  });

  it("describes an instrument and a choice, and counts a search", () => {
    expect(instrumentWords(thermistor)).toBe("Thermistor · degC · -40 to 125");
    expect(instrumentWords(instrumentOf({ ref: "x@1", title: "Bare" }, null))).toBe("Bare · no stated range");
    expect(chosenInstrumentWords(noInstrument, instruments)).toBe("No instrument: a plain reading");
    expect(chosenInstrumentWords("streamsim/thermistor@1", instruments)).toContain("Thermistor");
    expect(chosenInstrumentWords("gone@1", instruments)).toBe("gone@1");
    expect(instrumentCount(0, " zz ")).toBe('Nothing in the library matches "zz".');
    expect(instrumentCount(1, "")).toBe("1 instrument");
    expect(instrumentCount(3, "")).toBe("3 instruments");
  });

  it("suggests an instrument that measures the unit and covers the range, and otherwise none", () => {
    expect(suggestInstrument(metricOf(), instruments)).toBe("streamsim/thermistor@1");
    expect(suggestInstrument(metricOf({ unit: "%" }), instruments)).toBe("streamsim/humidity@1");
    expect(suggestInstrument(metricOf({ unit: "degC" }), instruments)).toBe("streamsim/thermistor@1");
    expect(suggestInstrument(metricOf({ eu_range: { min: -90, max: 50 } }), instruments)).toBe(noInstrument);
    expect(suggestInstrument(metricOf({ eu_range: { min: 0 } }), instruments)).toBe(noInstrument);
    expect(suggestInstrument(metricOf({ unit: "rpm" }), [tachometer])).toBe("streamsim/tachometer@1");
    expect(suggestInstrument(metricOf({ unit: "furlong" }), instruments)).toBe(noInstrument);
    expect(suggestInstrument(metricOf({ writable: true }), instruments)).toBe(noInstrument);
    expect(suggestInstrument(metricOf({ data_type: "boolean" }), instruments)).toBe(noInstrument);
    expect(suggestInstrument(metricOf(), [instrumentOf({ ref: "a@1", title: "A" }, null)])).toBe(noInstrument);
  });

  it("takes an instrument for numbers only", () => {
    expect(takesInstrument(metricOf())).toBe(true);
    expect(takesInstrument(metricOf({ data_type: "integer" }))).toBe(true);
    expect(takesInstrument(metricOf({ data_type: "enum" }))).toBe(false);
  });
});
