import type { FitCode, Instrument, InstrumentFit, InstrumentQuery } from "./twinTypes";
import type { Metric } from "./types";

interface Entry {
  readonly ref: string;
  readonly title: string;
  readonly unit: string;
  readonly min: number | null;
  readonly max: number | null;
  readonly match: "exact" | "convertible" | "";
  readonly fit: InstrumentFit | null;
}

export const noInstrument = "none";

const unfitReasons: Readonly<Record<Exclude<FitCode, "">, string>> = {
  unit_incompatible: "This instrument is in a different kind of unit",
  range_unstated: "Set a range for this reading in the Interface step",
  range_narrower: "This reading's range is wider than the instrument measures",
  reading_not_number: "This instrument needs a number reading",
};

const unfitFallback = "The binary cannot use this instrument for this reading";

export function instrumentOf(entry: Entry): Instrument {
  return { ...entry, words: [entry.title, entry.ref, entry.unit].join(" ").toLowerCase() };
}

export function searchInstruments(list: readonly Instrument[], query: string): readonly Instrument[] {
  const terms = query.toLowerCase().split(/\s+/).filter(Boolean);
  return list.filter((instrument) => terms.every((term) => instrument.words.includes(term)));
}

export function rangeWords(min: number | null, max: number | null): string {
  if (min !== null && max !== null) return `${String(min)} to ${String(max)}`;
  if (min !== null) return `from ${String(min)}`;
  return max === null ? "no stated range" : `up to ${String(max)}`;
}

export function convertedNote(instrument: Instrument): string {
  return instrument.match === "convertible" ? `measured in ${instrument.unit}, converted` : "";
}

export function instrumentWords(instrument: Instrument): string {
  const converted = instrument.match === "convertible";
  const range = rangeWords(instrument.min, instrument.max);
  const parts = converted
    ? [instrument.title, range, convertedNote(instrument)]
    : [instrument.title, instrument.unit, range];
  return parts.filter((part) => part !== "").join(" · ");
}

export function unfitWords(instrument: Instrument): string {
  const { fit } = instrument;
  if (fit === null || fit.ok) return "";
  return fit.code === "" ? unfitFallback : unfitReasons[fit.code];
}

export function unstatedMember(instrument: Instrument | undefined): string | null {
  const fit = instrument?.fit;
  return fit?.code === "range_unstated" ? fit.member : null;
}

export function chosenInstrumentWords(ref: string, list: readonly Instrument[]): string {
  if (ref === noInstrument) return "No instrument: a plain reading";
  const found = list.find((instrument) => instrument.ref === ref);
  return found === undefined ? ref : instrumentWords(found);
}

export function instrumentCount(found: number, query: string): string {
  if (found === 0) return `Nothing in the library matches "${query.trim()}".`;
  return found === 1 ? "1 instrument" : `${String(found)} instruments`;
}

export function takesInstrument(metric: Metric): boolean {
  return metric.dataType === "number" || metric.dataType === "integer";
}

export function instrumentQuery(metric: Metric): InstrumentQuery {
  const { min, max } = metric.range;
  const stated = min !== null && max !== null && min < max;
  return { unit: metric.unit, min: stated ? min : null, max: stated ? max : null };
}

export function queryKey({ unit, min, max }: InstrumentQuery): string {
  return [unit, min ?? "", max ?? ""].join("|");
}

export function catalogueKey(metric: Metric): string {
  return queryKey(instrumentQuery(metric));
}
