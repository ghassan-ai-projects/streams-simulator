import type { Pin, Pins } from "./ports";
import type { Attention, Shown } from "./readout";
import { attentionOf, channelWords } from "./readout";

export interface ReviewRow {
  readonly key: string;
  readonly label: string;
  readonly pin: Pin | null;
  readonly side: "in" | "out" | null;
  readonly reading: Shown | null;
  readonly observed: boolean;
}

export interface ReviewBody {
  readonly rows: readonly ReviewRow[];
  readonly hidden: number;
  readonly hiddenAttention: Attention;
}

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

export function reviewRowCount(pins: Pins, channels: readonly string[]): number {
  const ports = pins.ins.length + pins.outs.length;
  const listed = portless(pins) ? Math.min(channels.length, channelRows) : ports;
  const more = hiddenOf(pins, channels).length > 0 ? 1 : 0;
  return Math.max(1, listed + more);
}

function labelOf(pin: Pin, pins: Pins): string {
  const shared = [...pins.ins, ...pins.outs].filter((other) => other.channel === pin.channel).length > 1;
  return channelWords(pin.channel === null || shared ? pin.name : pin.channel);
}

function portRow(pin: Pin, side: "in" | "out", pins: Pins, run: Run): ReviewRow {
  const reading = pin.channel === null ? null : (run.shown.find((entry) => entry.channel === pin.channel) ?? null);
  return {
    key: `${side}:${pin.name}`,
    label: labelOf(pin, pins),
    pin,
    side,
    reading,
    observed: reading !== null || (pin.channel !== null && run.observed.has(pin.channel)),
  };
}

function channelRow(channel: string, run: Run): ReviewRow {
  const reading = run.shown.find((entry) => entry.channel === channel) ?? null;
  const observed = reading !== null || run.observed.has(channel);
  return { key: `channel:${channel}`, label: channelWords(channel), pin: null, side: null, reading, observed };
}

interface Run {
  readonly shown: readonly Shown[];
  readonly observed: ReadonlySet<string>;
}

function listedRows(pins: Pins, channels: readonly string[], run: Run): readonly ReviewRow[] {
  if (portless(pins)) return channels.slice(0, channelRows).map((channel) => channelRow(channel, run));
  return [
    ...pins.ins.map((pin) => portRow(pin, "in", pins, run)),
    ...pins.outs.map((pin) => portRow(pin, "out", pins, run)),
  ];
}

export function reviewBody(
  pins: Pins,
  channels: readonly string[],
  shown: readonly Shown[],
  observed: ReadonlySet<string>,
): ReviewBody {
  const hidden = hiddenOf(pins, channels).filter((channel) => observed.has(channel));
  const readings = shown.filter((entry) => hidden.includes(entry.channel));
  const run = { shown, observed };
  return { rows: listedRows(pins, channels, run), hidden: hidden.length, hiddenAttention: attentionOf(readings) };
}
