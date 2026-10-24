import { counted, shortDigest, spaced } from "../../shared/words";
import { displayUnit } from "../../shared/units";
import { isRecord } from "../../shared/json";
import type { ChangedKind, DiffRelation, InstanceSet, ModelChange, ModelDiff } from "./diffTypes";

export interface ChangeRow {
  readonly key: string;
  readonly label: string;
  readonly subject: string;
  readonly sentence: string;
  readonly before: string;
  readonly after: string;
}

export interface KindGroup {
  readonly kind: string;
  readonly status: ChangedKind["status"];
  readonly affects: string;
  readonly relations: readonly string[];
  readonly changes: readonly ChangeRow[];
}

export interface DiffView {
  readonly warning: string;
  readonly headline: string;
  readonly sides: readonly (readonly [string, string])[];
  readonly groups: readonly KindGroup[];
  readonly domain: readonly ChangeRow[];
  readonly unchanged: string;
}

export const versionWarning = "The model changed but its version did not.";

function rangeWords(value: Record<string, unknown>): string | null {
  const { min, max } = value;
  return typeof min === "number" && typeof max === "number" ? `${String(min)} to ${String(max)}` : null;
}

function plainWords(value: unknown): string {
  if (value === null || value === undefined) return "none";
  if (typeof value === "string") return value;
  if (typeof value === "number") return String(value);
  return (isRecord(value) ? rangeWords(value) : null) ?? JSON.stringify(value);
}

export function valueWords(value: unknown, unit = ""): string {
  const words = plainWords(value);
  const measured = typeof value === "number" || (isRecord(value) && rangeWords(value) !== null);
  return measured && unit !== "" ? `${words} ${unit}` : words;
}

function unitOf(change: ModelChange): string {
  if (change.field.endsWith("_ms")) return "ms";
  return change.field.endsWith("_s") ? "s" : "";
}

function changeRow(change: ModelChange, index: number): ChangeRow {
  const unit = unitOf(change);
  const symbol = change.code === "unit_changed" ? displayUnit : (text: string) => text;
  return {
    key: `${String(index)}:${change.subject}:${change.code}`,
    label: spaced(
      change.code
        .replace(/_changed$/, " changed")
        .replace(/_added$/, " added")
        .replace(/_removed$/, " removed"),
    ),
    subject: change.subject,
    sentence: change.sentence,
    before: symbol(valueWords(change.before, unit)),
    after: symbol(valueWords(change.after, unit)),
  };
}

function affectsOf(group: ChangedKind): string {
  const set: InstanceSet = group.status === "removed" ? group.before : group.after;
  if (set.names.length > 0) return `Affects ${set.names.join(", ")}`;
  if (set.idTemplates.length > 0) return `Affects ${counted(set.count, "instance")} from ${set.idTemplates.join(", ")}`;
  return set.count === 0 ? "" : `Affects ${counted(set.count, "instance")}`;
}

function relationWords(relation: DiffRelation): string {
  const links = relation.links.length === 0 ? "" : ` (${relation.links.join(", ")})`;
  return `${relation.type}: ${relation.fromKind} to ${relation.toKind}${links}`;
}

function groupOf(group: ChangedKind): KindGroup {
  return {
    kind: group.kind,
    status: group.status,
    affects: affectsOf(group),
    relations: group.relations.map(relationWords),
    changes: group.changes.map(changeRow),
  };
}

function headlineOf(diff: ModelDiff): string {
  if (diff.identity.sameDigest) return "The two models are identical.";
  return `${counted(diff.changes, "change")} in ${counted(diff.kinds.length, "device kind")}`;
}

function sideWords(side: ModelDiff["a"]): string {
  const where = side.source.kind === "run" ? `recorded in ${side.source.label}` : `file ${side.source.label}`;
  return `${side.id} ${side.version} · ${shortDigest(side.digest)} · ${where}`;
}

function unchangedWords(diff: ModelDiff): string {
  const items = diff.unchanged
    .filter(([, count]) => count > 0)
    .map(([name, count]) => `${String(count)} ${spaced(name)}`);
  return items.length === 0 ? "" : `Unchanged: ${items.join(", ")}.`;
}

export function diffView(diff: ModelDiff): DiffView {
  return {
    warning: diff.identity.versionKeptDigestChanged ? versionWarning : "",
    headline: headlineOf(diff),
    sides: [
      ["Model A", sideWords(diff.a)],
      ["Model B", sideWords(diff.b)],
    ],
    groups: diff.kinds.map(groupOf),
    domain: diff.domainChanges.map(changeRow),
    unchanged: unchangedWords(diff),
  };
}
