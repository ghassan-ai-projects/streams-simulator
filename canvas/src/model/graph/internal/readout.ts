import type { Attention, ChannelSpec, NodeStatus, ReadingValue, Shown } from "./types";
import type { ChannelRange, ChannelState } from "../../plant";
import { shortChannel, spaced } from "../../shared/words";
import { qualityWords } from "../../shared/quality";
import { displayUnit, groupedFigure } from "../../shared/units";

export const statusWords: Readonly<Record<NonNullable<NodeStatus>, string>> = {
  alarm: "Alarm",
  fault: "Fault",
  warn: "Warning",
};

const noSpec: ChannelSpec = { states: [], range: null };

function outsideOf(value: number, range: ChannelRange | null, unit: string): string {
  if (range === null || (value >= range.min && value <= range.max)) return "";
  const shown = unit === "" ? "" : ` ${unit}`;
  return `Outside the instrument's range of ${number(range.min)} to ${number(range.max)}${shown}`;
}

function number(value: number): string {
  const abs = Math.abs(value);
  return groupedFigure(value.toFixed(abs >= 100 ? 0 : abs >= 10 ? 1 : 2));
}

function namedState(channel: string, value: number, states: readonly ChannelState[]): Shown {
  const state = states.find((entry) => entry.value === value);
  const name = state?.name ?? `state ${String(value)}`;
  return { channel, value: name, unit: "", attention: state?.attention ?? null, word: true, outside: "", quality: "" };
}

function measured(channel: string, value: number, rawUnit: string, range: ChannelRange | null): Shown {
  const unit = displayUnit(rawUnit);
  const outside = outsideOf(value, range, unit);
  return { channel, value: number(value), unit, attention: null, word: false, outside, quality: "" };
}

export function showReading(reading: ReadingValue, spec: ChannelSpec | null): Shown {
  return { ...valueShown(reading, spec), quality: qualityWords(reading.quality ?? "") };
}

function valueShown(reading: ReadingValue, spec: ChannelSpec | null): Shown {
  const { states, range } = spec ?? noSpec;
  const channel = shortChannel(reading.channel);
  if (typeof reading.value === "number") {
    return states.length > 0
      ? namedState(channel, reading.value, states)
      : measured(channel, reading.value, reading.unit, range);
  }
  const value = typeof reading.value === "boolean" ? (reading.value ? "on" : "off") : reading.value;
  return { channel, value, unit: "", attention: null, word: true, outside: "", quality: "" };
}

export function attentionOf(shown: readonly Shown[]): Attention {
  if (shown.some((entry) => entry.attention === "alarm")) return "alarm";
  return shown.some((entry) => entry.attention === "warn") ? "warn" : null;
}

export const channelWords = spaced;

export function pinWords(name: string, side: "in" | "out"): { readonly full: string; readonly short: string } {
  const full = spaced(name);
  const suffix = ` ${side}`;
  const short = full.endsWith(suffix) && full.length > suffix.length ? full.slice(0, -suffix.length) : full;
  return { full, short };
}

const statusRank = { alarm: 0, fault: 1, warn: 2 } as const;

export function worstOf(statuses: readonly NodeStatus[]): { readonly status: NodeStatus; readonly count: number } {
  const ranked = statuses.flatMap((status) => (status === null ? [] : [status]));
  const worst = ranked.reduce<NodeStatus>(
    (best, status) => (best === null || statusRank[status] < statusRank[best] ? status : best),
    null,
  );
  return { status: worst, count: ranked.filter((status) => status === worst).length };
}
