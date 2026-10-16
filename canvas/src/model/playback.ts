import { shortChannel } from "./shared/words";

export interface Observation {
  readonly atNs: number;
  readonly entity: string;
  readonly channel: string;
  readonly value: number | string | boolean;
  readonly unit: string;
}

export type JournalEntry =
  | { readonly kind: "observation"; readonly observation: Observation }
  | { readonly kind: "opened"; readonly atNs: number; readonly mode: string; readonly adapter: string }
  | { readonly kind: "closed"; readonly atNs: number; readonly incomplete: boolean }
  | { readonly kind: "other"; readonly atNs: number };

type Json = Record<string, unknown>;

function json(text: string): Json {
  try {
    const value = JSON.parse(text) as unknown;
    return typeof value === "object" && value !== null && !Array.isArray(value) ? (value as Json) : {};
  } catch {
    return {};
  }
}

const field = (value: unknown, key: string): unknown =>
  typeof value === "object" && value !== null ? (value as Json)[key] : undefined;
const str = (value: unknown): string => (typeof value === "string" ? value : "");
const num = (value: unknown): number => (typeof value === "number" ? value : 0);

function reading(value: unknown): number | string | boolean {
  return typeof value === "number" || typeof value === "boolean" ? value : str(value);
}

function observation(record: Json): JournalEntry {
  const line = json(str(record.line));
  return {
    kind: "observation",
    observation: {
      atNs: num(record.at_ns),
      entity: str(line.entity_id),
      channel: str(line.channel),
      value: reading(line.value),
      unit: str(line.unit),
    },
  };
}

export function parseJournalRecord(text: string): JournalEntry {
  const record = json(text);
  const atNs = num(record.at_ns);
  switch (record.type) {
    case "observation":
      return observation(record);
    case "run.opened":
      return { kind: "opened", atNs, mode: str(record.mode), adapter: str(field(record.adapter, "id")) };
    case "run.closed":
      return { kind: "closed", atNs, incomplete: record.incomplete === true };
    default:
      return { kind: "other", atNs };
  }
}

interface Series {
  readonly times: number[];
  readonly values: (number | string | boolean)[];
  readonly unit: string;
}

export interface Opening {
  readonly atNs: number;
  readonly full: boolean;
}

export interface Reading {
  readonly channel: string;
  readonly value: number | string | boolean;
  readonly unit: string;
  readonly atNs: number;
}

export class PlaybackIndex {
  private readonly series = new Map<string, Map<string, Series>>();
  private first = Number.POSITIVE_INFINITY;
  private last = 0;
  private total = 0;

  add(observation: Observation): void {
    let channels = this.series.get(observation.entity);
    if (channels === undefined) {
      channels = new Map();
      this.series.set(observation.entity, channels);
    }
    let series = channels.get(observation.channel);
    if (series === undefined) {
      series = { times: [], values: [], unit: observation.unit };
      channels.set(observation.channel, series);
    }
    series.times.push(observation.atNs);
    series.values.push(observation.value);
    this.first = Math.min(this.first, observation.atNs);
    this.last = Math.max(this.last, observation.atNs);
    this.total += 1;
  }

  get size(): number {
    return this.total;
  }

  bounds(): { readonly startNs: number; readonly endNs: number } | null {
    return this.total === 0 ? null : { startNs: this.first, endNs: this.last };
  }

  settled(): number | null {
    return this.opening()?.atNs ?? null;
  }

  opening(): Opening | null {
    if (this.total === 0) return null;
    const cap = this.first + (this.last - this.first) / 10;
    const complete = Math.max(this.first, ...this.firstReadings());
    return { atNs: Math.min(cap, complete), full: complete <= cap };
  }

  private firstReadings(): number[] {
    return [...this.series.values()].flatMap((channels) =>
      [...channels.values()].map((series) => series.times[0] ?? this.first),
    );
  }

  observed(entity: string): ReadonlySet<string> {
    const channels = this.series.get(entity);
    return new Set([...(channels?.keys() ?? [])].map(shortChannel));
  }

  readings(entity: string, atNs: number): readonly Reading[] {
    const channels = this.series.get(entity);
    if (channels === undefined) return [];
    return [...channels.entries()].flatMap(([channel, series]) => {
      const index = lastAtOrBefore(series.times, atNs);
      const value = series.values[index];
      const at = series.times[index];
      return value === undefined || at === undefined ? [] : [{ channel, value, unit: series.unit, atNs: at }];
    });
  }
}

function lastAtOrBefore(times: readonly number[], atNs: number): number {
  let low = 0;
  let high = times.length - 1;
  let found = -1;
  while (low <= high) {
    const middle = (low + high) >> 1;
    if ((times[middle] ?? Number.POSITIVE_INFINITY) <= atNs) {
      found = middle;
      low = middle + 1;
    } else {
      high = middle - 1;
    }
  }
  return found;
}

export const playbackSpeeds = [1, 10, 60, 600] as const;

export function advance(positionNs: number, elapsedMs: number, speed: number, endNs: number): number {
  return Math.min(endNs, positionNs + Math.max(0, elapsedMs) * 1e6 * speed);
}
