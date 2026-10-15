import type { ChannelRange, ChannelState } from "./plant";
import type { Reading } from "./playback";
import { shortChannel } from "./playback";
import { displayUnit } from "./units";

export type Attention = "alarm" | "warn" | null;
export type NodeStatus = Attention | "fault";

export const statusWords: Readonly<Record<NonNullable<NodeStatus>, string>> = {
  alarm: "Alarm",
  fault: "Fault",
  warn: "Warning",
};

export interface Shown {
  readonly channel: string;
  readonly value: string;
  readonly unit: string;
  readonly attention: Attention;
  readonly word: boolean;
  readonly outside: string;
}

export interface ChannelSpec {
  readonly states: readonly ChannelState[];
  readonly range: ChannelRange | null;
}

const noSpec: ChannelSpec = { states: [], range: null };

function outsideOf(value: number, range: ChannelRange | null, unit: string): string {
  if (range === null || (value >= range.min && value <= range.max)) return "";
  const shown = unit === "" ? "" : ` ${unit}`;
  return `Outside the instrument's range of ${number(range.min)} to ${number(range.max)}${shown}`;
}

function number(value: number): string {
  const abs = Math.abs(value);
  return value.toFixed(abs >= 100 ? 0 : abs >= 10 ? 1 : 2);
}

function namedState(channel: string, value: number, states: readonly ChannelState[]): Shown {
  const state = states.find((entry) => entry.value === value);
  const name = state?.name ?? `state ${String(value)}`;
  return { channel, value: name, unit: "", attention: state?.attention ?? null, word: true, outside: "" };
}

function measured(channel: string, value: number, rawUnit: string, range: ChannelRange | null): Shown {
  const unit = displayUnit(rawUnit);
  return { channel, value: number(value), unit, attention: null, word: false, outside: outsideOf(value, range, unit) };
}

export function showReading(reading: Reading, spec: ChannelSpec | null = noSpec): Shown {
  const { states, range } = spec ?? noSpec;
  const channel = shortChannel(reading.channel);
  if (typeof reading.value === "number") {
    return states.length > 0
      ? namedState(channel, reading.value, states)
      : measured(channel, reading.value, reading.unit, range);
  }
  const value = typeof reading.value === "boolean" ? (reading.value ? "on" : "off") : reading.value;
  return { channel, value, unit: "", attention: null, word: true, outside: "" };
}

export function attentionOf(shown: readonly Shown[]): Attention {
  if (shown.some((entry) => entry.attention === "alarm")) return "alarm";
  return shown.some((entry) => entry.attention === "warn") ? "warn" : null;
}

const attentionRank = { alarm: 0, warn: 1 } as const;

export function abnormalFirst(shown: readonly Shown[]): readonly Shown[] {
  return [...shown].sort(
    (a, b) =>
      (a.attention === null ? 2 : attentionRank[a.attention]) - (b.attention === null ? 2 : attentionRank[b.attention]),
  );
}

export function channelWords(name: string): string {
  return name.replaceAll("_", " ");
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

export function withFault(shown: readonly Shown[], fault: string): readonly Shown[] {
  return fault === ""
    ? shown
    : [{ channel: "fault", value: fault, unit: "", attention: "alarm", word: true, outside: "" }, ...shown];
}

export function sharedTimes(times: readonly number[]): {
  readonly latest: number;
  readonly own: readonly (number | null)[];
} {
  const latest = Math.max(...times);
  return { latest, own: times.map((time) => (time === latest ? null : time)) };
}
