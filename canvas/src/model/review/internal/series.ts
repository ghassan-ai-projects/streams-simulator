import type { Quality } from "../../shared/quality";
export type SampleValue = number | string | boolean;

export interface Series {
  readonly times: number[];
  readonly values: SampleValue[];
  readonly qualities: Quality[];
  readonly lags: number[];
  readonly unit: string;
}

export interface Sample {
  readonly atNs: number;
  readonly value: SampleValue;
  readonly quality: Quality;
  readonly lagNs: number;
}

export function newSeries(unit: string): Series {
  return { times: [], values: [], qualities: [], lags: [], unit };
}

export function append(series: Series, sample: Sample, capacity: number): boolean {
  series.times.push(sample.atNs);
  series.values.push(sample.value);
  series.qualities.push(sample.quality);
  series.lags.push(sample.lagNs);
  if (series.times.length < capacity * 2) return false;
  const drop = series.times.length - capacity;
  series.times.splice(0, drop);
  series.values.splice(0, drop);
  series.qualities.splice(0, drop);
  series.lags.splice(0, drop);
  return true;
}

export function lastAtOrBefore(times: readonly number[], atNs: number): number {
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

export function firstAtOrAfter(times: readonly number[], atNs: number): number {
  return lastAtOrBefore(times, atNs - 1) + 1;
}
