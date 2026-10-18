import { describe, expect, it } from "vitest";
import { convertibleThermistor, humidity, instruments, metricOf, thermistor } from "../../../test/twinFixture";
import {
  catalogueKey,
  chosenInstrumentWords,
  convertedNote,
  instrumentCount,
  instrumentOf,
  instrumentQuery,
  instrumentWords,
  noInstrument,
  queryKey,
  rangeWords,
  searchInstruments,
  takesInstrument,
  unfitWords,
  unstatedMember,
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
      fit: null,
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

  it("says when an instrument measures a different unit and is converted", () => {
    expect(instrumentWords(thermistor)).toBe("Thermistor · degC · -40 to 125");
    expect(instrumentWords(convertibleThermistor)).toBe("Thermistor · -40 to 125 · measured in degC, converted");
    expect(convertedNote(convertibleThermistor)).toBe("measured in degC, converted");
    expect(convertedNote(thermistor)).toBe("");
  });

  it("gives one plain reason for an instrument the binary says does not fit, and none for one that fits", () => {
    const unfit = (code: Parameters<typeof instrumentOf>[0]["fit"]) => instrumentOf({ ...thermistor, fit: code });
    expect(unfitWords(thermistor)).toBe("");
    expect(unfitWords(unfit({ ok: true, code: "", member: "" }))).toBe("");
    expect(unfitWords(unfit({ ok: false, code: "unit_incompatible", member: "unit" }))).toBe(
      "This instrument is in a different kind of unit",
    );
    expect(unfitWords(unfit({ ok: false, code: "range_unstated", member: "eu_range" }))).toBe(
      "Set a range for this reading in the Interface step",
    );
    expect(unfitWords(unfit({ ok: false, code: "range_narrower", member: "eu_range" }))).toBe(
      "This reading's range is wider than the instrument measures",
    );
    expect(unfitWords(unfit({ ok: false, code: "reading_not_number", member: "data_type" }))).toBe(
      "This instrument needs a number reading",
    );
    expect(unfitWords(unfit({ ok: false, code: "", member: "" }))).toBe(
      "The binary cannot use this instrument for this reading",
    );
  });

  it("names the member of the reading that an instrument waits for, when the range is unstated", () => {
    const wait = instrumentOf({ ...thermistor, fit: { ok: false, code: "range_unstated", member: "eu_range" } });
    const narrow = instrumentOf({ ...thermistor, fit: { ok: false, code: "range_narrower", member: "eu_range" } });
    expect(unstatedMember(wait)).toBe("eu_range");
    expect(unstatedMember(narrow)).toBeNull();
    expect(unstatedMember(thermistor)).toBeNull();
    expect(unstatedMember(undefined)).toBeNull();
  });

  it("asks the library by unit, and by range only when the reading states both ends in order", () => {
    expect(instrumentQuery(metricOf())).toEqual({ unit: "Cel", min: 0, max: 50 });
    expect(instrumentQuery(metricOf({ eu_range: { min: 0 } }))).toEqual({ unit: "Cel", min: null, max: null });
    expect(instrumentQuery(metricOf({ eu_range: { min: 5, max: 5 } }))).toEqual({ unit: "Cel", min: null, max: null });
    expect(catalogueKey(metricOf())).toBe("Cel|0|50");
    expect(queryKey({ unit: "%", min: null, max: null })).toBe("%||");
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
