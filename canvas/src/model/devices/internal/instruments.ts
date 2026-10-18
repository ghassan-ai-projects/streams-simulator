import type { Instrument } from "./twinTypes";
import type { Metric } from "./types";

interface Entry {
  readonly ref: string;
  readonly title: string;
  readonly unit: string;
  readonly min: number | null;
  readonly max: number | null;
  readonly match: "exact" | "convertible" | "";
}

export const noInstrument = "none";

export const conversionNote = "different unit — needs conversion";

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

export function instrumentWords(instrument: Instrument): string {
  const note = instrument.match === "convertible" ? conversionNote : "";
  return [instrument.title, instrument.unit, rangeWords(instrument.min, instrument.max), note]
    .filter((part) => part !== "")
    .join(" · ");
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
