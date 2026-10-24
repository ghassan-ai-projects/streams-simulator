import type {
  ChangedKind,
  DiffIdentity,
  DiffRelation,
  DiffSide,
  InstanceSet,
  ModelChange,
  ModelDiff,
} from "../../../model/workspace";
import { isRecord, list, number, record, text } from "../../../model/shared/json";
import { choice, requiredTexts, rows } from "./decodeKit";

function sideOf(value: unknown): DiffSide | null {
  if (!isRecord(value) || !isRecord(value.source)) return null;
  const given = requiredTexts(value, ["id", "version", "digest"]);
  const { source } = value;
  if (given === null || !choice(source.kind, ["file", "run"])) return null;
  return { ...given, title: text(value.title), source: { kind: source.kind, label: text(source.label) } };
}

function identityOf(value: unknown): DiffIdentity | null {
  const identity = record(value);
  const flags = ["same_id", "same_version", "same_digest", "same_id_and_version_different_digest"];
  if (!flags.every((flag) => typeof identity[flag] === "boolean")) return null;
  return {
    sameId: identity.same_id === true,
    sameVersion: identity.same_version === true,
    sameDigest: identity.same_digest === true,
    versionKeptDigestChanged: identity.same_id_and_version_different_digest === true,
  };
}

function changeOf(row: Record<string, unknown>): ModelChange | null {
  const given = requiredTexts(row, ["code", "subject", "sentence"]);
  if (given === null) return null;
  return {
    code: given.code,
    subject: given.subject,
    section: text(row.section),
    field: text(row.field),
    before: row.before ?? null,
    after: row.after ?? null,
    sentence: given.sentence,
  };
}

function instanceSetOf(value: unknown): InstanceSet {
  const set = record(value);
  return { count: number(set.count), names: list(set.names).map(text), idTemplates: list(set.id_templates).map(text) };
}

function relationOf(row: Record<string, unknown>): DiffRelation | null {
  const type = text(row.type);
  if (type === "") return null;
  return { type, fromKind: text(row.from_kind), toKind: text(row.to_kind), links: list(row.links).map(text) };
}

function kindOf(row: Record<string, unknown>): ChangedKind | null {
  const kind = text(row.kind);
  const [relations, changes] = [rows(row.relations ?? [], relationOf), rows(row.changes, changeOf)];
  if (kind === "" || !choice(row.status, ["changed", "added", "removed"]) || changes === null) return null;
  if (relations === null) return null;
  const instances = record(row.instances);
  const [before, after] = [instanceSetOf(instances.before), instanceSetOf(instances.after)];
  return { kind, status: row.status, before, after, relations, changes };
}

function unchangedOf(value: unknown): readonly (readonly [string, number])[] {
  return Object.entries(record(value)).filter((entry): entry is [string, number] => typeof entry[1] === "number");
}

export function modelDiff(value: unknown): ModelDiff | null {
  if (!isRecord(value) || value.schema_version !== "0.1") return null;
  const [a, b, identity] = [sideOf(value.a), sideOf(value.b), identityOf(value.identity)];
  const [kinds, domainChanges] = [rows(value.kinds, kindOf), rows(value.domain_changes ?? [], changeOf)];
  if (a === null || b === null || identity === null || kinds === null || domainChanges === null) return null;
  return {
    a,
    b,
    identity,
    kinds,
    domainChanges,
    changes: number(record(value.summary).changes),
    unchanged: unchangedOf(value.unchanged_summary),
  };
}
