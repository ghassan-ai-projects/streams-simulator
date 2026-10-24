import type { InstanceOrigin } from "../../graph";
import type { BindingNote } from "../../launch";
import { shortChannel } from "../../shared/words";
import type { InterruptRecord } from "./interrupts";
import type { SessionLive } from "./sessionLive";
import type { Quality } from "../../shared/quality";
import type { SampleValue, Series } from "./series";
import { append, firstAtOrAfter, lastAtOrBefore, newSeries } from "./series";
import type { OperatorCall } from "./callLog";
import { CallLog } from "./callLog";

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
  | { readonly kind: "call"; readonly call: OperatorCall }
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
  readonly qualities: readonly Quality[];
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
  private readonly callLog: CallLog;

  constructor(options: IndexOptions = {}) {
    this.capacity = options.capacity ?? Number.POSITIVE_INFINITY;
    this.callLog = new CallLog(Number.isFinite(this.capacity) ? this.capacity / 4 : Number.POSITIVE_INFINITY);
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
    return {
      times: series.times.slice(start, end),
      values: series.values.slice(start, end),
      qualities: series.qualities.slice(start, end),
    };
  }

  trimmedBeforeNs(entity: string, channel: string): number | null {
    return this.series.get(entity)?.get(channel)?.trimmedNs ?? null;
  }

  addCall(call: OperatorCall): void {
    this.callLog.add(call);
  }

  calls(fromNs: number, toNs: number): readonly OperatorCall[] {
    return this.callLog.between(fromNs, toNs);
  }

  unitOf(entity: string, channel: string): string | null {
    return this.series.get(entity)?.get(channel)?.unit ?? null;
  }

  neighbour(entity: string, channel: string, atNs: number, direction: 1 | -1): number | null {
    const times = this.series.get(entity)?.get(channel)?.times;
    if (times === undefined) return null;
    const at = direction === 1 ? lastAtOrBefore(times, atNs) + 1 : firstAtOrAfter(times, atNs) - 1;
    return times[at] ?? null;
  }

  reading(entity: string, channel: string, atNs: number): Reading | null {
    const series = this.series.get(entity)?.get(channel);
    if (series === undefined) return null;
    return this.readingAt(channel, series, lastAtOrBefore(series.times, atNs));
  }

  private readingAt(channel: string, series: Series, at: number): Reading | null {
    const value = series.values[at];
    const atNs = series.times[at];
    if (value === undefined || atNs === undefined) return null;
    const quality = series.qualities[at] ?? "";
    return { channel, value, unit: series.unit, atNs, quality, lagNs: series.lags[at] ?? 0 };
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
      const found = this.readingAt(channel, series, lastAtOrBefore(series.times, atNs));
      return found === null ? [] : [found];
    });
  }
}

export const playbackSpeeds = [1, 10, 60, 600] as const;

export function advance(positionNs: number, elapsedMs: number, speed: number, endNs: number): number {
  return Math.min(endNs, positionNs + Math.max(0, elapsedMs) * 1e6 * speed);
}
