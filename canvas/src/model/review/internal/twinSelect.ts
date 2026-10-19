import { counted } from "../../shared/words";
import type { Connection } from "./origins";
import { runClock } from "./timeline";
import type { TwinDivergence, TwinMetric, TwinReport, TwinSeries, TwinTolerance } from "./twin";
import { notComparedWords, numberWords, toleranceSourceWords } from "./twinWords";

type TwinOffer =
  | { readonly kind: "absent" }
  | { readonly kind: "withheld"; readonly why: string }
  | { readonly kind: "offered"; readonly domain: string };

export type TwinCoverage =
  | {
      readonly kind: "series";
      readonly metric: TwinMetric;
      readonly divergences: readonly TwinDivergence[];
      readonly undrawn: number;
    }
  | { readonly kind: "empty"; readonly metric: TwinMetric; readonly why: string };

export function twinOffer(connection: Connection, domain: string): TwinOffer {
  if (connection !== "saved") return { kind: "absent" };
  if (domain === "") {
    return { kind: "withheld", why: "The session record does not name the domain the twin would run." };
  }
  return { kind: "offered", domain };
}

export function toleranceWords(tolerance: TwinTolerance, unit: string): string {
  const source = toleranceSourceWords[tolerance.source];
  if (tolerance.value === null) return `No tolerance (${source})`;
  return `±${[numberWords(tolerance.value), unit].filter((part) => part !== "").join(" ")} (${source})`;
}

export function ruleWords(report: TwinReport): string {
  return report.ruleDescription;
}

function reasonsWords(metric: TwinMetric): string {
  return metric.notCompared.map((entry) => `${notComparedWords[entry.reason]} (${String(entry.count)})`).join("; ");
}

function emptyWhy(metric: TwinMetric): string {
  if (metric.notCompared.length === 0) return `The board reported no reading of ${metric.metric} to compare.`;
  return `The twin cannot judge ${metric.metric}: ${reasonsWords(metric)}.`;
}

function intersects(divergence: TwinDivergence, firstNs: number, lastNs: number): boolean {
  return divergence.lastNs >= firstNs && divergence.firstNs <= lastNs;
}

function coverageFor(report: TwinReport, metric: TwinMetric): TwinCoverage {
  const first = metric.series.points[0];
  const last = metric.series.points.at(-1);
  if (first === undefined || last === undefined) return { kind: "empty", metric, why: emptyWhy(metric) };
  const mine = report.divergences.filter((d) => d.device === metric.device && d.metric === metric.metric);
  const drawn = mine.filter((d) => intersects(d, first.atNs, last.atNs));
  return { kind: "series", metric, divergences: drawn, undrawn: mine.length - drawn.length };
}

export function coverageOf(report: TwinReport, device: string): readonly TwinCoverage[] {
  return report.metrics.filter((metric) => metric.device === device).map((metric) => coverageFor(report, metric));
}

function undrawnWords(undrawn: number): string {
  if (undrawn === 0) return "";
  const stretches =
    undrawn === 1 ? "1 stretch where they part ways lies" : `${String(undrawn)} stretches where they part ways lie`;
  return ` ${stretches} beyond the readings shown; the Report tab lists ${undrawn === 1 ? "it" : "them"}.`;
}

export function truncationNote(series: TwinSeries, undrawn: number): string {
  if (!series.truncated) return "";
  const cut = `Showing the first ${String(series.points.length)} of ${String(series.total)} compared readings.`;
  return `${cut}${undrawnWords(undrawn)}`;
}

export function metricLine(metric: TwinMetric): string {
  const { compared, diverged } = metric.counts;
  return `${counted(compared, "reading")} compared · ${String(diverged)} beyond tolerance · ${toleranceWords(metric.tolerance, metric.unit)}`;
}

export function divergenceLine(divergence: TwinDivergence, unit: string, startNs: number): string {
  const at = (ns: number) => runClock(ns, startNs);
  const size = (value: number) => [numberWords(value), unit].filter((part) => part !== "").join(" ");
  const span =
    divergence.firstNs === divergence.lastNs
      ? at(divergence.firstNs)
      : `${at(divergence.firstNs)} to ${at(divergence.lastNs)}`;
  const worst = `worst error ${size(divergence.worstError)} at ${at(divergence.worstAtNs)}`;
  return `${span} · ${counted(divergence.readings, "reading")} · ${worst} (board ${size(divergence.boardValue)}, twin ${size(divergence.twinValue)})`;
}

export function inForceWords(divergence: TwinDivergence): string {
  return divergence.commands.length === 0 ? "none" : divergence.commands.join(", ");
}

export function unitOfMetric(report: TwinReport, divergence: TwinDivergence): string {
  const metric = report.metrics.find((m) => m.device === divergence.device && m.metric === divergence.metric);
  return metric?.unit ?? "";
}
