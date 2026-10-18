import type { DomainChecked, InstrumentSpec, LibraryInstrument, LintReport, TwinMade } from "../../../app/ports";
import { isRecord, list, number, record, text } from "../../../model/shared/json";

function bound(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

export function instrumentRefs(value: unknown): readonly LibraryInstrument[] | null {
  if (!isRecord(value) || !Array.isArray(value.components)) return null;
  return value.components
    .map(record)
    .filter((entry) => entry.kind === "instrument" && text(entry.id) !== "")
    .map((entry) => ({
      ref: `${text(entry.id)}@${typeof entry.version === "number" ? String(entry.version) : "1"}`,
      title: text(entry.title) || text(entry.id),
    }));
}

export function instrumentSpec(value: unknown): InstrumentSpec | null {
  if (!isRecord(value) || !isRecord(value.document)) return null;
  const document = value.document;
  const range = record(document.range);
  return {
    family: text(document.unit_family),
    unit: text(document.unit),
    min: bound(range.min),
    max: bound(range.max),
  };
}

export function twinMade(value: unknown): TwinMade | null {
  if (!isRecord(value) || text(value.id) === "") return null;
  return {
    id: text(value.id),
    digest: text(value.digest),
    channels: number(value.channels),
    faults: number(value.faults),
    effectors: number(value.effectors),
    implements: list(value.implements)
      .map(text)
      .filter((name) => name !== ""),
  };
}

export function domainChecked(value: unknown): DomainChecked | null {
  if (!isRecord(value) || value.valid !== true) return null;
  return { id: text(value.id), digest: text(value.digest) };
}

export function lintReport(value: unknown): LintReport | null {
  if (!isRecord(value) || !Array.isArray(value.findings)) return null;
  return {
    errors: number(value.errors),
    warnings: number(value.warnings),
    findings: value.findings.map(record).map((finding) => ({
      severity: text(finding.severity),
      entity: text(finding.entity),
      detail: text(finding.detail),
    })),
  };
}
