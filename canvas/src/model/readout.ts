import type { ChannelState } from "./plant";
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
}

function number(value: number): string {
  const abs = Math.abs(value);
  return value.toFixed(abs >= 100 ? 0 : abs >= 10 ? 1 : 2);
}

export function showReading(reading: Reading, states: readonly ChannelState[]): Shown {
  const channel = shortChannel(reading.channel);
  if (states.length > 0 && typeof reading.value === "number") {
    const value = reading.value;
    const state = states.find((entry) => entry.value === value);
    return { channel, value: state?.name ?? `state ${String(value)}`, unit: "", attention: state?.attention ?? null };
  }
  if (typeof reading.value === "number") {
    return { channel, value: number(reading.value), unit: displayUnit(reading.unit), attention: null };
  }
  return { channel, value: String(reading.value), unit: "", attention: null };
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
  return fault === "" ? shown : [{ channel: "fault", value: fault, unit: "", attention: "alarm" }, ...shown];
}
