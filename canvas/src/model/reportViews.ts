import type { LaunchPlan } from "./launch";
import { durationOf } from "./workspace";
import type { KpiRow, Report, VerifyState } from "./report";
import { badgeOf, replayLine, wallClock } from "./report";

export function kpisOf(report: Report, entity: string | null): readonly KpiRow[] {
  if (entity === null) return report.kpis;
  return report.kpis.filter((row) => row.entity === entity || row.entity.startsWith(`${entity}/`));
}

const kpiTokens: Readonly<Record<string, string>> = { pct: "", frac: "fraction", cop: "COP" };

export function kpiWords(state: string): string {
  const text = state
    .split("_")
    .map((word) => kpiTokens[word] ?? word)
    .filter((word) => word !== "")
    .join(" ");
  return text.charAt(0).toUpperCase() + text.slice(1);
}

export interface KpiGroup {
  readonly entity: string;
  readonly rows: readonly KpiRow[];
}

export function kpiGroups(rows: readonly KpiRow[]): readonly KpiGroup[] {
  const groups: { entity: string; rows: KpiRow[] }[] = [];
  for (const row of rows) {
    const last = groups.at(-1);
    if (last?.entity === row.entity) last.rows.push(row);
    else groups.push({ entity: row.entity, rows: [row] });
  }
  return groups;
}

export function runIdentity(report: Report): string {
  return [
    `${report.domainId} ${report.domainVersion}`.trim(),
    report.seed === null ? "" : `seed ${String(report.seed)}`,
    report.adapterId,
    report.simVersion === "" ? "" : `sim ${report.simVersion}`,
    report.endNs > report.startNs ? `run time 00:00:00 is ${wallClock(report.startNs)} UTC simulated` : "",
  ]
    .filter((part) => part !== "")
    .join(" · ");
}

export function runFacts(report: Report): string {
  const seconds = report.endNs > report.startNs ? Math.round((report.endNs - report.startNs) / 1e9) : 0;
  return [report.seed === null ? "" : `seed ${String(report.seed)}`, durationOf(seconds)]
    .filter((part) => part !== "")
    .join(" · ");
}

export interface Rerun {
  readonly plan: LaunchPlan;
  readonly left: readonly string[];
}

function count(n: number, one: string, many: string): string {
  return n === 0 ? "" : `${String(n)} ${n === 1 ? one : many}`;
}

function leftOut(report: Report): readonly string[] {
  const kinds = (kind: string) => report.disturbances.filter((d) => d.kind === kind).length;
  return [
    report.seed === null ? "the seed, which the run did not record (seed 1 is used)" : "",
    report.endNs - report.startNs < 60e9 ? "a length under one minute (one minute is used)" : "",
    count(kinds("perturbation"), "delivery perturbation", "delivery perturbations"),
    count(kinds("environment"), "environment change", "environment changes"),
  ].filter((part) => part !== "");
}

export function rerunOf(report: Report): Rerun {
  return { plan: planOfRun(report), left: leftOut(report) };
}

function planOfRun(report: Report): LaunchPlan {
  return {
    seed: report.seed ?? 1,
    durationS: Math.max(60, Math.round((report.endNs - report.startNs) / 1e9)),
    explore: report.mode === "explore",
    faults: report.disturbances
      .filter((disturbance) => disturbance.kind === "fault")
      .map((fault) => ({
        entity: fault.target,
        fault: fault.name,
        atS: Math.max(0, Math.round((fault.onsetNs - report.startNs) / 1e9)),
      })),
  };
}

export function integrityLine(report: Report, verify: VerifyState): string {
  switch (badgeOf(report, verify)) {
    case "verified":
      return "Verified: the files match and the replay reproduces.";
    case "verifying":
      return "Verifying: replaying the run…";
    case "incomplete":
      return `Incomplete: ${report.incompleteReason || "the run did not close."}`;
    case "not-verified":
      return report.verification === "files_differ"
        ? "Not verified: the files changed after the run."
        : `Not verified: ${replayLine(verify)}`;
    case "not-checked":
      return verify.phase === "failed" ? replayLine(verify) : "Not checked: the run recorded no file digests.";
  }
}
