import type { LaunchPlan } from "./launch";
import { durationOf } from "./workspace";
import type { KpiRow, Report, VerifyState } from "./report";
import { badgeOf, replayLine } from "./report";

export function kpisOf(report: Report, entity: string | null): readonly KpiRow[] {
  if (entity === null) return report.kpis;
  return report.kpis.filter((row) => row.entity === entity || row.entity.startsWith(`${entity}/`));
}

export function runIdentity(report: Report): string {
  return [
    `${report.domainId} ${report.domainVersion}`.trim(),
    report.seed === null ? "" : `seed ${String(report.seed)}`,
    report.adapterId,
    report.simVersion === "" ? "" : `sim ${report.simVersion}`,
  ]
    .filter((part) => part !== "")
    .join(" · ");
}

export function runFacts(report: Report): string {
  const minutes = report.endNs > report.startNs ? Math.round((report.endNs - report.startNs) / 1e9) : 0;
  return [report.seed === null ? "" : `seed ${String(report.seed)}`, durationOf(minutes)]
    .filter((part) => part !== "")
    .join(" · ");
}

export function planOfRun(report: Report): LaunchPlan {
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
