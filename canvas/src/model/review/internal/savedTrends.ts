import { isRecord, text } from "../../shared/json";
import type { Plant } from "../../plant";
import type { Watched } from "./watchList";
import { watchLimit } from "./watchList";

export const trendSizes = ["compact", "normal", "tall", "full"] as const;

export type TrendSize = (typeof trendSizes)[number];

export function defaultTrendSize(pinned: number): TrendSize {
  return pinned >= 3 ? "tall" : "normal";
}

export function resizedTrend(size: TrendSize, direction: "larger" | "smaller"): TrendSize {
  const index = trendSizes.indexOf(size) + (direction === "larger" ? 1 : -1);
  return trendSizes[Math.min(trendSizes.length - 1, Math.max(0, index))] ?? size;
}

export interface SavedTrends {
  readonly open: boolean;
  readonly size: TrendSize | null;
  readonly watched: readonly Watched[];
}

export function savedTrendsText(saved: SavedTrends): string {
  return JSON.stringify(saved);
}

function pinOf(value: unknown): Watched | null {
  if (!isRecord(value)) return null;
  const entity = text(value.entity);
  const channel = text(value.channel);
  return entity === "" || channel === "" ? null : { entity, channel };
}

function parsed(raw: string): unknown {
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function savedTrendsOf(raw: string | null): SavedTrends | null {
  const body = raw === null ? null : parsed(raw);
  if (!isRecord(body)) return null;
  const pins = Array.isArray(body.watched) ? body.watched.map(pinOf) : [];
  return {
    open: body.open === true,
    size: trendSizes.find((size) => size === body.size) ?? null,
    watched: pins.filter((pin) => pin !== null).slice(0, watchLimit),
  };
}

export function stillExists(plant: Plant, watched: Watched): boolean {
  const instance = plant.instances.find((each) => each.id === watched.entity);
  const channels = instance === undefined ? [] : (plant.kinds.get(instance.kind)?.channels ?? []);
  return channels.some((channel) => channel.name === watched.channel);
}
