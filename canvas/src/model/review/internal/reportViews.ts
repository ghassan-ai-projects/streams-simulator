import { clockOf, durationOf } from "../../shared/duration";
import { seedWords } from "../../shared/words";
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

export function kpiFigure(value: number): string {
  const size = Math.abs(value);
  return size >= 100 ? value.toFixed(0) : size >= 10 ? value.toFixed(1) : value.toFixed(2);
}

export function kpiSteady(row: KpiRow): boolean {
  return kpiFigure(row.min) === kpiFigure(row.max);
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

export function reportFacts(report: Report): string {
  return [seedWords(report.seed), durationOf(spanSeconds(report))].filter((part) => part !== "").join(" · ");
}

export function terminationWords(report: Report): string {
  const { termination, requestedEndNs, startNs } = report;
  if (termination?.reason !== "stopped_by_operator") return "";
  const at = clockOf((termination.atNs - startNs) / 1e9);
  if (requestedEndNs === null) return `Stopped by operator at ${at}`;
  return `Stopped by operator at ${at} of ${clockOf((requestedEndNs - startNs) / 1e9)}`;
}

export const unfinishedTruthWords =
  "This run did not finish. These are the plant's recorded values up to where it stopped, not the result of a completed run.";

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
