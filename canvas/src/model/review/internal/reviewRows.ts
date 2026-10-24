import type { ReadoutBody, ReadoutRow, RowCount, Shown, ViewPresentation } from "../../graph";
import { attentionOf, channelWords, noteHeight } from "../../graph";
import type { Pin, Pins } from "../../plant";

const channelRows = 4;

function portChannels(pins: Pins): ReadonlySet<string> {
  const channels = [...pins.ins, ...pins.outs].flatMap((pin) => (pin.channel === null ? [] : [pin.channel]));
  return new Set(channels);
}

function portless(pins: Pins): boolean {
  return pins.ins.length + pins.outs.length === 0;
}

function hiddenOf(pins: Pins, channels: readonly string[]): readonly string[] {
  if (portless(pins)) return channels.slice(channelRows);
  const ported = portChannels(pins);
  return channels.filter((channel) => !ported.has(channel));
}

const reviewRowCount: RowCount = (pins, channels) => {
  const ports = pins.ins.length + pins.outs.length;
  const listed = portless(pins) ? Math.min(channels.length, channelRows) : ports;
  const more = hiddenOf(pins, channels).length > 0 ? 1 : 0;
  return Math.max(1, listed + more);
};

export const reviewView: ViewPresentation = { rows: reviewRowCount, note: noteHeight };

function labelOf(pin: Pin, pins: Pins): string {
  const shared = [...pins.ins, ...pins.outs].filter((other) => other.channel === pin.channel).length > 1;
  return channelWords(pin.channel === null || shared ? pin.name : pin.channel);
}

interface Readings {
  readonly shown: readonly Shown[];
  readonly observed: ReadonlySet<string>;
  readonly phase: ReadoutPhase;
}

const blank: Omit<Shown, "channel"> = {
  value: "",
  unit: "",
  word: false,
  attention: null,
  outside: "",
  quality: "",
};

export type ReadoutPhase = "live" | "ended";

function absentWords(channel: string | null, observed: boolean, phase: ReadoutPhase): string {
  if (channel === null) return "No measured channel";
  if (observed) return "No reading yet at this time";
  return phase === "live" ? "Awaiting first reading" : "Never observed in this run";
}

function rowOf(key: string, label: string, channel: string | null, readings: Readings): ReadoutRow {
  const reading = channel === null ? null : (readings.shown.find((entry) => entry.channel === channel) ?? null);
  const observed = reading !== null || (channel !== null && readings.observed.has(channel));
  const { value, unit, word, attention, outside, quality } = reading ?? blank;
  const absent = reading === null ? absentWords(channel, observed, readings.phase) : "";
  return { key, label, present: reading !== null, value, unit, word, attention, outside, observed, quality, absent };
}

function portRow(pin: Pin, side: "in" | "out", pins: Pins, readings: Readings): ReadoutRow {
  return rowOf(`${side}:${pin.name}`, labelOf(pin, pins), pin.channel, readings);
}

function channelRow(channel: string, readings: Readings): ReadoutRow {
  return rowOf(`channel:${channel}`, channelWords(channel), channel, readings);
}

function listedRows(pins: Pins, channels: readonly string[], readings: Readings): readonly ReadoutRow[] {
  if (portless(pins)) return channels.slice(0, channelRows).map((channel) => channelRow(channel, readings));
  return [
    ...pins.ins.map((pin) => portRow(pin, "in", pins, readings)),
    ...pins.outs.map((pin) => portRow(pin, "out", pins, readings)),
  ];
}

export function readoutBody(
  pins: Pins,
  channels: readonly string[],
  shown: readonly Shown[],
  observed: ReadonlySet<string>,
  phase: ReadoutPhase = "ended",
): ReadoutBody {
  const hidden = hiddenOf(pins, channels).filter((channel) => observed.has(channel));
  const hiddenReadings = shown.filter((entry) => hidden.includes(entry.channel));
  const rows = listedRows(pins, channels, { shown, observed, phase });
  return { rows, hidden: hidden.length, hiddenAttention: attentionOf(hiddenReadings) };
}
