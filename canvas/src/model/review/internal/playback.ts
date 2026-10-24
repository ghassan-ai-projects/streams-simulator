import type { InstanceOrigin } from "../../graph";
import type { BindingNote } from "../../launch";
import { shortChannel } from "../../shared/words";
import type { InterruptRecord } from "./interrupts";
import type { SessionLive } from "./sessionLive";
import type { Quality } from "../../shared/quality";
import type { SampleValue, Series } from "./series";
import { append, firstAtOrAfter, lastAtOrBefore, newSeries } from "./series";

export interface Observation {
  readonly atNs: number;
  readonly entity: string;
  readonly channel: string;
  readonly value: SampleValue;
  readonly unit: string;
  readonly quality?: Quality;
  readonly lagNs?: number;
}

export type JournalEntry =
  | {
      readonly kind: "session";
      readonly atNs: number;
      readonly domain: string;
      readonly simulated: readonly string[];
      readonly policies: readonly string[] | null;
    }
  | { readonly kind: "live"; readonly live: SessionLive; readonly atNs: number }
  | { readonly kind: "live-unreadable"; readonly atNs: number }
  | { readonly kind: "interrupt-unreadable"; readonly atNs: number }
  | { readonly kind: "interrupt"; readonly atNs: number; readonly record: InterruptRecord }
  | { readonly kind: "observation"; readonly observation: Observation }
  | { readonly kind: "opened"; readonly atNs: number; readonly mode: string; readonly adapter: string }
  | { readonly kind: "closed"; readonly atNs: number; readonly incomplete: boolean }
  | { readonly kind: "origin"; readonly atNs: number; readonly origin: InstanceOrigin }
  | { readonly kind: "binding"; readonly atNs: number; readonly binding: BindingNote }
  | { readonly kind: "clock"; readonly atNs: number }
  | { readonly kind: "other"; readonly atNs: number };

export interface Opening {
  readonly atNs: number;
  readonly full: boolean;
}

export interface Reading {
  readonly channel: string;
  readonly value: SampleValue;
  readonly unit: string;
  readonly atNs: number;
  readonly quality: Quality;
  readonly lagNs: number;
}

export interface Window {
  readonly times: readonly number[];
  readonly values: readonly SampleValue[];
}

interface IndexOptions {
  readonly capacity?: number;
}

export class PlaybackIndex {
  private readonly series = new Map<string, Map<string, Series>>();
  private readonly capacity: number;
  private first = Number.POSITIVE_INFINITY;
  private last = 0;
  private total = 0;
  private clock = 0;
  private opened = 0;

  constructor(options: IndexOptions = {}) {
    this.capacity = options.capacity ?? Number.POSITIVE_INFINITY;
  }

  add(observation: Observation): void {
    const series = this.seriesFor(observation);
    const sample = {
      atNs: observation.atNs,
      value: observation.value,
      quality: observation.quality ?? "",
      lagNs: observation.lagNs ?? 0,
    };
    const trimmed = append(series, sample, this.capacity);
    this.first = trimmed ? this.earliest() : Math.min(this.first, observation.atNs);
    this.last = Math.max(this.last, observation.atNs);
    this.total += 1;
  }

  private seriesFor(observation: Observation): Series {
    const channels = this.series.get(observation.entity) ?? new Map<string, Series>();
    this.series.set(observation.entity, channels);
    const series = channels.get(observation.channel) ?? newSeries(observation.unit);
    channels.set(observation.channel, series);
    return series;
  }

  private earliest(): number {
    return Math.min(...this.firstReadings());
  }

  get clockNs(): number {
    return this.clock;
  }

  get openedNs(): number {
    return this.opened;
  }

  markOpened(atNs: number): void {
    this.opened = atNs;
  }

  advanceClock(atNs: number): void {
    this.clock = Math.max(this.clock, atNs);
  }

  samples(entity: string, channel: string, fromNs: number, toNs: number): Window | null {
    const series = this.series.get(entity)?.get(channel);
    if (series === undefined) return null;
    const start = firstAtOrAfter(series.times, fromNs);
    const end = lastAtOrBefore(series.times, toNs) + 1;
    return { times: series.times.slice(start, end), values: series.values.slice(start, end) };
  }

  get size(): number {
    return this.total;
  }

  get held(): number {
    return [...this.series.values()].reduce(
      (sum, channels) => sum + [...channels.values()].reduce((own, series) => own + series.times.length, 0),
      0,
    );
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
      if (value === undefined || at === undefined) return [];
      const quality = series.qualities[index] ?? "";
      return [{ channel, value, unit: series.unit, atNs: at, quality, lagNs: series.lags[index] ?? 0 }];
    });
  }
}

export const playbackSpeeds = [1, 10, 60, 600] as const;

export function advance(positionNs: number, elapsedMs: number, speed: number, endNs: number): number {
  return Math.min(endNs, positionNs + Math.max(0, elapsedMs) * 1e6 * speed);
}
