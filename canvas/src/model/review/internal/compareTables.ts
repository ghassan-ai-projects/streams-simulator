import { clockOf } from "../../shared/duration";
import { displayUnit } from "../../shared/units";
import { humanise } from "../../shared/words";
import type { ComparedAlarm, ComparedKpi, Comparison, Presence, Scope } from "./comparison";
import { kpiFigure, kpiWords } from "./reportViews";
import { alarmRow } from "./words";

const none = "–";

function length(ns: number): string {
  const whole = clockOf(ns / 1e9);
  return whole.startsWith("00:") ? whole.slice(3) : whole.replace(/^0/, "");
}

export function scopeCaption(scope: Scope): string {
  const { aLengthNs, bLengthNs } = scope;
  if (aLengthNs === null || bLengthNs === null) {
    const kept = aLengthNs === null ? `this run ${length(bLengthNs ?? 0)}` : `reference ${length(aLengthNs)}`;
    const lacking = aLengthNs === null ? "the reference" : "this run";
    return `Whole run: ${kept} simulated; ${lacking} recorded no whole-run results`;
  }
  return `Whole run: reference ${length(aLengthNs)}, this run ${length(bLengthNs)} simulated`;
}

export function windowedWords(comparison: Comparison): string {
  return `Event-window comparisons are not available: ${comparison.windowedReason}`;
}

function signOf(delta: number): string {
  return delta < 0 ? "−" : "+";
}

export function signedFigure(delta: number): string {
  if (delta === 0) return "0";
  const size = Math.abs(delta);
  const figure = kpiFigure(size);
  return `${signOf(delta)}${Number(figure) === 0 ? size.toPrecision(2) : figure}`;
}

export function signedPercent(relative: number | null): string {
  if (relative === null) return none;
  if (relative === 0) return "0 %";
  const percent = Math.abs(relative * 100);
  const figure = kpiFigure(percent);
  return `${signOf(relative)}${Number(figure) === 0 ? percent.toPrecision(2) : figure} %`;
}

function signedWhole(delta: number, suffix = ""): string {
  return delta === 0 ? `0${suffix}` : `${signOf(delta)}${String(Math.abs(Math.round(delta)))}${suffix}`;
}

const presenceNotes: Readonly<Record<Presence, string>> = {
  both: "",
  only_in_a: "Only in reference",
  only_in_b: "Only in this run",
};

export interface KpiLine {
  readonly key: string;
  readonly entity: string;
  readonly kpi: string;
  readonly unit: string;
  readonly reference: string;
  readonly thisRun: string;
  readonly delta: string;
  readonly relative: string;
  readonly note: string;
}

function kpiLine(row: ComparedKpi, unit: (entity: string, state: string) => string): KpiLine {
  return {
    key: `${row.entity}/${row.kpi}`,
    entity: row.entity,
    kpi: kpiWords(row.kpi),
    unit: displayUnit(unit(row.entity, row.kpi)),
    reference: row.a === null ? none : kpiFigure(row.a.mean),
    thisRun: row.b === null ? none : kpiFigure(row.b.mean),
    delta: row.delta === null ? none : signedFigure(row.delta),
    relative: signedPercent(row.relative),
    note: presenceNotes[row.presence],
  };
}

export interface Table<L> {
  readonly caption: string;
  readonly rows: readonly L[];
}

export function kpiTable(comparison: Comparison, unit: (entity: string, state: string) => string): Table<KpiLine> {
  const scope = comparison.kpis[0]?.scope ?? comparison.alarmSystem.scope;
  return { caption: scopeCaption(scope), rows: comparison.kpis.map((row) => kpiLine(row, unit)) };
}

interface Trio {
  readonly reference: string;
  readonly thisRun: string;
  readonly delta: string;
}

export interface AlarmLine {
  readonly key: string;
  readonly entity: string;
  readonly alarm: string;
  readonly activations: Trio;
  readonly active: Trio;
  readonly note: string;
}

function alarmLine(row: ComparedAlarm): AlarmLine {
  const seconds = (value: number | undefined) => (value === undefined ? none : `${String(Math.round(value))} s`);
  const count = (value: number | undefined) => (value === undefined ? none : String(value));
  return {
    key: `${row.entity}/${row.alarm}`,
    entity: row.entity,
    alarm: humanise(row.alarm),
    activations: {
      reference: count(row.a?.activations),
      thisRun: count(row.b?.activations),
      delta: row.delta === null ? none : signedWhole(row.delta.activations),
    },
    active: {
      reference: seconds(row.a?.activeS),
      thisRun: seconds(row.b?.activeS),
      delta: row.delta === null ? none : signedWhole(row.delta.activeS, " s"),
    },
    note: presenceNotes[row.presence],
  };
}

export function alarmTable(comparison: Comparison): Table<AlarmLine> {
  const scope = comparison.alarms[0]?.scope ?? comparison.alarmSystem.scope;
  return { caption: scopeCaption(scope), rows: comparison.alarms.map(alarmLine) };
}

export interface SystemLine extends Trio {
  readonly key: string;
  readonly measure: string;
}

const systemKeys = ["activations", "peak_per_10_min", "flood_intervals", "chattering", "stale", "standing_at_end"];

export function alarmSystemTable(comparison: Comparison): Table<SystemLine> {
  const { a, b, delta, scope } = comparison.alarmSystem;
  const shown = (counts: Readonly<Record<string, number>> | null, key: string) =>
    counts?.[key] === undefined ? none : String(counts[key]);
  return {
    caption: scopeCaption(scope),
    rows: systemKeys.map((key) => ({
      key,
      measure: alarmRow(key, "")[0],
      reference: shown(a, key),
      thisRun: shown(b, key),
      delta: delta?.[key] === undefined ? none : signedWhole(delta[key]),
    })),
  };
}
