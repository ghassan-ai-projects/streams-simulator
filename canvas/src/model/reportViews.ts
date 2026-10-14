import type { KpiRow, Report } from "./report";

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
