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

const annunciatorStates = ["normal", "unacknowledged", "acknowledged", "returned, unacknowledged", "shelved"] as const;
const annunciatorAttention: readonly Attention[] = [null, "alarm", "alarm", "warn", null];

function isAnnunciator(instrument: string): boolean {
  return instrument.startsWith("streamsim/annunciator");
}

function number(value: number): string {
  const abs = Math.abs(value);
  return value.toFixed(abs >= 100 ? 0 : abs >= 10 ? 1 : 2);
}

export function showReading(reading: Reading, instrument: string): Shown {
  const channel = shortChannel(reading.channel);
  if (isAnnunciator(instrument) && typeof reading.value === "number") {
    const code = Math.round(reading.value);
    return {
      channel,
      value: annunciatorStates[code] ?? `code ${String(code)}`,
      unit: "",
      attention: annunciatorAttention[code] ?? null,
    };
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
