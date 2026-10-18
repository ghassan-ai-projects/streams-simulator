import type { InstanceOrigin } from "../../graph";
import type { BindingNote } from "../../launch";
import { shortChannel } from "../../shared/words";
import type { SessionLive } from "./sessionLive";

interface Observation {
  readonly atNs: number;
  readonly entity: string;
  readonly channel: string;
  readonly value: number | string | boolean;
  readonly unit: string;
}

export type JournalEntry =
  | { readonly kind: "session"; readonly atNs: number; readonly simulated: readonly string[] }
  | { readonly kind: "live"; readonly live: SessionLive; readonly atNs: number }
  | { readonly kind: "observation"; readonly observation: Observation }
  | { readonly kind: "opened"; readonly atNs: number; readonly mode: string; readonly adapter: string }
  | { readonly kind: "closed"; readonly atNs: number; readonly incomplete: boolean }
  | { readonly kind: "origin"; readonly atNs: number; readonly origin: InstanceOrigin }
  | { readonly kind: "binding"; readonly atNs: number; readonly binding: BindingNote }
  | { readonly kind: "other"; readonly atNs: number };

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
