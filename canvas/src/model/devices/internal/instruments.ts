import type { Instrument } from "./twinTypes";
import type { Metric } from "./types";

interface Entry {
  readonly ref: string;
  readonly title: string;
}

interface Spec {
  readonly family: string;
  readonly unit: string;
  readonly min: number | null;
  readonly max: number | null;
}

const sameScale: Readonly<Record<string, readonly string[]>> = {
  Cel: ["degC"],
  degC: ["Cel"],
  "%": ["%RH"],
  "%RH": ["%"],
};

const unknown: Spec = { family: "", unit: "", min: null, max: null };

export const noInstrument = "none";

export function instrumentOf(entry: Entry, spec: Spec | null): Instrument {
  const known = spec ?? unknown;
  return { ...entry, ...known, words: [entry.title, entry.ref, known.family, known.unit].join(" ").toLowerCase() };
}

export function sortInstruments(list: readonly Instrument[]): readonly Instrument[] {
  return [...list].sort((a, b) => a.title.localeCompare(b.title));
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
  return [instrument.title, instrument.unit, rangeWords(instrument.min, instrument.max)]
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

function sameUnit(instrument: Instrument, unit: string): boolean {
  if (instrument.unit === "") return false;
  return instrument.unit === unit || (sameScale[unit] ?? []).includes(instrument.unit);
}

function covers(instrument: Instrument, metric: Metric): boolean {
  const { min, max } = metric.range;
  if (min === null || max === null) return false;
  return (instrument.min === null || instrument.min <= min) && (instrument.max === null || instrument.max >= max);
}

export function takesInstrument(metric: Metric): boolean {
  return metric.dataType === "number" || metric.dataType === "integer";
}

export function suggestInstrument(metric: Metric, list: readonly Instrument[]): string {
  if (!takesInstrument(metric) || metric.writable) return noInstrument;
  return (
    list.find((instrument) => sameUnit(instrument, metric.unit) && covers(instrument, metric))?.ref ?? noInstrument
  );
}
