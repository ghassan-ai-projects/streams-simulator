import type { Shown } from "../../graph";
import { channelWords, showReading } from "../../graph";
import type { Plant } from "../../plant";
import { channelOf } from "../../plant";
import { durationOf } from "../../shared/duration";
import { qualityWords } from "../../shared/quality";
import { codeTitle } from "../../shared/units";
import { shortChannel } from "../../shared/words";
import type { PlaybackIndex, Reading } from "./playback";
import { runClock } from "./timeline";
import type { Trend } from "./trend";
import { numberText, trendOf } from "./trend";

type Attention = Shown["attention"];

export interface ObservedRow {
  readonly key: string;
  readonly label: string;
  readonly present: boolean;
  readonly value: string;
  readonly unit: string;
  readonly unitTitle: string;
  readonly word: boolean;
  readonly attention: Attention;
  readonly outside: string;
  readonly observedAt: string;
  readonly measuredAt: string;
  readonly age: string;
  readonly quality: string;
  readonly absent: string;
  readonly trend: Trend | null;
  readonly trendNote: string;
}

export interface ObservedFacts {
  readonly index: PlaybackIndex;
  readonly plant: Plant | null;
  readonly entity: string;
  readonly atNs: number;
  readonly referenceNs: number;
  readonly startNs: number;
  readonly live: boolean;
}

export function ageWords(ns: number): string {
  const seconds = Math.round(Math.max(0, ns) / 1e9);
  return seconds < 2 ? "just now" : `${durationOf(seconds)} old`;
}

const measuredAfterNs = 1e9;
const later = 1e30;

function declaredChannels(plant: Plant | null, entity: string): readonly string[] {
  const kind = plant?.instances.find((instance) => instance.id === entity)?.kind;
  return plant?.kinds.get(kind ?? "")?.channels.map((channel) => channel.name) ?? [];
}

function absentWords(facts: ObservedFacts, channel: string): string {
  const reported = facts.index.samples(facts.entity, channel, facts.atNs + 1, later);
  if ((reported?.times.length ?? 0) > 0) return "No reading yet at this time";
  return facts.live ? "Awaiting first reading" : "Never observed in this run";
}

function gapWords(trend: Trend): string {
  if (trend.gaps === 0) return "No gaps.";
  const noun = trend.gaps === 1 ? "gap" : "gaps";
  const median = durationOf(Math.round(trend.medianNs / 1e9)) || "under 1 s";
  return `${String(trend.gaps)} ${noun} left blank: a gap is a pause longer than 3 times the median interval between readings (${median}), inferred from the observed intervals; nothing is drawn across it.`;
}

function trendNote(trend: Trend, unit: string): string {
  const span = durationOf(trend.windowNs / 1e9);
  const range = `from ${numberText(trend.min)} to ${numberText(trend.max)}${unit === "" ? "" : ` ${unit}`}`;
  return `Last ${span} of simulated time, ${range}. ${gapWords(trend)}`;
}

function presentRow(facts: ObservedFacts, channel: string, reading: Reading): ObservedRow {
  const shown = showReading(reading, facts.plant === null ? null : channelOf(facts.plant, facts.entity, channel));
  const trend = trendOf(facts.index, facts.entity, channel, facts.referenceNs);
  const measured = reading.lagNs >= measuredAfterNs ? runClock(reading.atNs - reading.lagNs, facts.startNs) : "";
  return {
    key: channel,
    label: channelWords(shortChannel(channel)),
    present: true,
    value: shown.value,
    unit: shown.unit,
    unitTitle: codeTitle(reading.unit),
    word: shown.word,
    attention: shown.attention,
    outside: shown.outside,
    observedAt: runClock(reading.atNs, facts.startNs),
    measuredAt: measured,
    age: ageWords(facts.referenceNs - reading.atNs),
    quality: qualityWords(reading.quality),
    absent: "",
    trend,
    trendNote: trend === null ? "" : trendNote(trend, shown.unit),
  };
}

function absentRow(facts: ObservedFacts, channel: string): ObservedRow {
  return {
    key: channel,
    label: channelWords(shortChannel(channel)),
    present: false,
    value: "",
    unit: "",
    unitTitle: "",
    word: false,
    attention: null,
    outside: "",
    observedAt: "",
    measuredAt: "",
    age: "",
    quality: "",
    absent: absentWords(facts, channel),
    trend: null,
    trendNote: "",
  };
}

export function observedRows(facts: ObservedFacts): readonly ObservedRow[] {
  const readings = facts.index.readings(facts.entity, facts.atNs);
  if (readings.length === 0 && facts.plant === null) return [];
  const declared = declaredChannels(facts.plant, facts.entity);
  const extra = readings.map((reading) => reading.channel).filter((channel) => !declared.includes(channel));
  const rows = [...declared, ...extra].map((channel) => {
    const reading = readings.find((entry) => entry.channel === channel);
    return reading === undefined ? absentRow(facts, channel) : presentRow(facts, channel, reading);
  });
  return rows.toSorted((a, b) => rank(a) - rank(b));
}

const attentionRank = { alarm: 0, warn: 1 } as const;

function rank(row: ObservedRow): number {
  if (!row.present) return 3;
  return row.attention === null ? 2 : attentionRank[row.attention];
}
