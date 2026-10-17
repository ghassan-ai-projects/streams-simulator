import type { LaunchPlan } from "../../launch";
import { durationOf } from "../../shared/duration";
import { counted, seedWords } from "../../shared/words";
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

interface KpiGroup {
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

function spanSeconds(report: Report): number {
  return report.endNs > report.startNs ? Math.round((report.endNs - report.startNs) / 1e9) : 0;
}

export function runIdentity(report: Report): string {
  const seconds = spanSeconds(report);
  return [
    `${report.domainId} ${report.domainVersion}`.trim(),
    seedWords(report.seed),
    report.adapterId,
    report.simVersion === "" ? "" : `sim ${report.simVersion}`,
    seconds === 0 ? "" : `${durationOf(seconds)} of plant time from ${wallClock(report.startNs)} UTC`,
  ]
    .filter((part) => part !== "")
    .join(" · ");
}

export function runFacts(report: Report): string {
  return [seedWords(report.seed), durationOf(spanSeconds(report))].filter((part) => part !== "").join(" · ");
}

interface Rerun {
  readonly plan: LaunchPlan;
  readonly left: readonly string[];
}

function count(n: number, one: string, many: string): string {
  return n === 0 ? "" : counted(n, one, many);
}

function clearedFaults(report: Report): number {
  return report.disturbances.filter((d) => d.kind === "fault" && d.clearedNs !== null).length;
}

function leftOut(report: Report): readonly string[] {
  const kinds = (kind: string) => report.disturbances.filter((d) => d.kind === kind).length;
  return [
    report.seed === null ? "the seed, which the run did not record (seed 1 is used)" : "",
    report.endNs - report.startNs < 60e9 ? "a length under one minute (one minute is used)" : "",
    count(kinds("perturbation"), "delivery perturbation", "delivery perturbations"),
    count(kinds("environment"), "environment change", "environment changes"),
    count(report.effectorCalls.length, "effector call", "effector calls"),
    count(clearedFaults(report), "fault clear (the fault stays on)", "fault clears (the faults stay on)"),
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
    calls: [],
  };
}

export function integrityLine(report: Report, verify: VerifyState): string {
  switch (badgeOf(report, verify)) {
    case "verified":
      return "The files match and the replay reproduces.";
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

export function fileLine(report: Report): string {
  if (report.verification === "files_match") return "Files match the record.";
  if (report.verification === "files_differ") return "The files changed after the run.";
  return "The run recorded no file digests.";
}
