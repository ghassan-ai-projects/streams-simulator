import type { Quality } from "../../shared/quality";
import { qualityWords } from "../../shared/quality";
import { channelWords } from "../../graph";
import { displayUnit, codeTitle } from "../../shared/units";
import { shortChannel } from "../../shared/words";
import type { PlaybackIndex, Window } from "./playback";
import type { Frame, Projection, Run, Span, TrendPoint } from "./trend";
import { medianInterval, numberText, numeric, plotRun, splitGaps } from "./trend";
import type { Watched } from "./watchList";
import { watchKey } from "./watchList";

export const plotFrame: Frame = { width: 600, height: 48 };

const pixelsPerBucket = 2;
const edge = 0.08;

interface QualityBand {
  readonly x0: number;
  readonly x1: number;
  readonly quality: Exclude<Quality, "">;
  readonly words: string;
}

interface Tick {
  readonly y: number;
  readonly text: string;
}

export interface Plot {
  readonly key: string;
  readonly entity: string;
  readonly channel: string;
  readonly label: string;
  readonly unit: string;
  readonly unitTitle: string;
  readonly state: "plotted" | "words" | "empty";
  readonly segments: readonly (readonly TrendPoint[])[];
  readonly dots: readonly TrendPoint[];
  readonly top: Tick | null;
  readonly bottom: Tick | null;
  readonly bands: readonly QualityBand[];
  readonly gaps: number;
  readonly medianNs: number;
}

export function xOf(ns: number, span: Span, width = plotFrame.width): number {
  return ((ns - span.startNs) / Math.max(1, span.endNs - span.startNs)) * width;
}

export function nsAtX(x: number, span: Span, width = plotFrame.width): number {
  const fraction = Math.min(1, Math.max(0, x / width));
  return span.startNs + fraction * (span.endNs - span.startNs);
}

function paddedRange(min: number, max: number): { readonly min: number; readonly max: number } {
  const pad = max === min ? Math.abs(min) * 0.05 || 1 : (max - min) * edge;
  return { min: min - pad, max: max + pad };
}

function levelOf(value: number, range: { readonly min: number; readonly max: number }): number {
  return plotFrame.height - ((value - range.min) / (range.max - range.min)) * plotFrame.height;
}

function frameOf(width: number): Frame {
  return { width, height: plotFrame.height };
}

function bandsOf(window: Window, span: Span, width: number): readonly QualityBand[] {
  const bands: { from: number; to: number; quality: Exclude<Quality, ""> }[] = [];
  window.times.forEach((time, at) => {
    const quality = window.qualities[at] ?? "";
    if (quality === "") return;
    const last = bands.at(-1);
    if (last?.quality === quality && last.to === (window.times[at - 1] ?? -1)) last.to = time;
    else bands.push({ from: time, to: time, quality });
  });
  return bands.map((band) => ({
    x0: xOf(band.from, span, width),
    x1: xOf(band.to, span, width),
    quality: band.quality,
    words: qualityWords(band.quality),
  }));
}

function identity(watched: Watched, unit: string) {
  return {
    key: watchKey(watched),
    entity: watched.entity,
    channel: watched.channel,
    label: `${watched.entity} · ${channelWords(shortChannel(watched.channel))}`,
    unit: displayUnit(unit),
    unitTitle: codeTitle(unit),
  };
}

function bare(watched: Watched, unit: string, state: Plot["state"]): Plot {
  const empty = { segments: [], dots: [], top: null, bottom: null, bands: [], gaps: 0, medianNs: 0 };
  return { ...identity(watched, unit), state, ...empty };
}

function drawn(watched: Watched, unit: string, window: Window, values: readonly number[], at: Placed): Plot {
  const { span, width } = at;
  const range = paddedRange(Math.min(...values), Math.max(...values));
  const medianNs = medianInterval(window.times);
  const runs = splitGaps({ times: window.times, values } satisfies Run, medianNs);
  const projection: Projection = { span, frame: frameOf(width), range, buckets: Math.floor(width / pixelsPerBucket) };
  const plotted = runs.map((run) => plotRun(run, projection));
  const [low, high] = [Math.min(...values), Math.max(...values)];
  return {
    ...identity(watched, unit),
    state: "plotted",
    segments: plotted.filter((points) => points.length > 1),
    dots: plotted.filter((points) => points.length === 1).flat(),
    top: { y: levelOf(high, range), text: numberText(high) },
    bottom: { y: levelOf(low, range), text: numberText(low) },
    bands: bandsOf(window, span, width),
    gaps: runs.length - 1,
    medianNs,
  };
}

interface Placed {
  readonly span: Span;
  readonly width: number;
}

export function plotOf(index: PlaybackIndex, watched: Watched, span: Span, width = plotFrame.width): Plot {
  const unit = index.unitOf(watched.entity, watched.channel) ?? "";
  const window = index.samples(watched.entity, watched.channel, span.startNs, span.endNs);
  if (window === null || window.times.length === 0) return bare(watched, unit, "empty");
  const values = numeric(window.values);
  return values === null ? bare(watched, unit, "words") : drawn(watched, unit, window, values, { span, width });
}

interface Cached {
  readonly size: number;
  readonly key: string;
  readonly plots: readonly Plot[];
}

const cache = new WeakMap<PlaybackIndex, Cached>();

export function plotsOf(index: PlaybackIndex, watched: readonly Watched[], span: Span, width: number): readonly Plot[] {
  const key = `${watched.map(watchKey).join("|")}@${String(span.startNs)}-${String(span.endNs)}x${String(width)}`;
  const known = cache.get(index);
  if (known?.size === index.size && known.key === key) return known.plots;
  const plots = watched.map((entry) => plotOf(index, entry, span, width));
  cache.set(index, { size: index.size, key, plots });
  return plots;
}
