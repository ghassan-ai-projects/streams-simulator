import type {
  DeviceLineage,
  Editable,
  Lineage,
  LineagePlant,
  LineageRun,
  LineageSource,
  PlantKind,
  Unreadable,
} from "../../../model/workspace";
import { isRecord, list, text } from "../../../model/shared/json";
import type { Given } from "./decodeKit";
import { choice, requiredTexts, rows } from "./decodeKit";

const statuses = ["in_sync", "out_of_sync", "does_not_compose", "does_not_scaffold", "kind_files_missing"] as const;
const reasons = [
  "composition",
  "composition_out_of_sync",
  "twin",
  "twin_out_of_sync",
  "example",
  "hand_authored",
] as const;
const verdicts = ["editable", "read_only"] as const;
const routes = [
  "edit_composition",
  "recompose",
  "open_device_behavior",
  "rescaffold_twin",
  "create_from_example",
  "copy_domain_file",
] as const;

function sourceOf(value: unknown): Given<LineageSource | null> | null {
  if (value === undefined) return { value: null };
  const given = isRecord(value) ? requiredTexts(value, ["file", "id", "version"]) : null;
  if (!isRecord(value) || given === null || !choice(value.status, statuses)) return null;
  return { value: { ...given, status: value.status, reason: text(value.reason) } };
}

function runOf(row: Record<string, unknown>): LineageRun | null {
  const given = requiredTexts(row, ["id", "domain_version", "domain_digest"]);
  if (given === null || typeof row.same_as_current !== "boolean") return null;
  return {
    id: given.id,
    mode: text(row.mode),
    recorded: row.recorded === true,
    domainVersion: given.domain_version,
    domainDigest: given.domain_digest,
    sameAsCurrent: row.same_as_current,
  };
}

function plantKindOf(row: Record<string, unknown>): PlantKind | null {
  const kind = text(row.kind);
  return kind === "" ? null : { kind, found: row.status === "found", file: text(row.file) };
}

function editableOf(value: unknown): Editable | null {
  if (!isRecord(value) || !isRecord(value.route)) return null;
  const { route } = value;
  if (!choice(value.verdict, verdicts) || !choice(value.reason, reasons) || !choice(route.code, routes)) return null;
  return { verdict: value.verdict, reason: value.reason, route: { code: route.code, sentence: text(route.sentence) } };
}

function plantOf(row: Record<string, unknown>): LineagePlant | null {
  const domain = isRecord(row.domain) ? requiredTexts(row.domain, ["id", "version", "title", "digest", "file"]) : null;
  const [composition, twin] = [sourceOf(row.composition), sourceOf(row.twin)];
  const [kinds, runs, editable] = [
    rows(row.device_kinds, plantKindOf),
    rows(row.runs, runOf),
    editableOf(row.editable),
  ];
  if (domain === null || composition === null || twin === null) return null;
  if (kinds === null || runs === null || editable === null) return null;
  return { ...domain, composition: composition.value, twin: twin.value, kinds, runs, editable };
}

function inclusionOf(row: Record<string, unknown>) {
  const given = requiredTexts(row, ["twin", "file"]);
  return given === null ? null : { twin: given.twin, file: given.file, kinds: Number(row.kinds) || 0 };
}

function deviceOf(row: Record<string, unknown>): DeviceLineage | null {
  const given = requiredTexts(row, ["kind", "file"]);
  const including = rows(row.twins_including, inclusionOf);
  if (given === null || including === null || typeof row.has_own_twin !== "boolean") return null;
  const implementedBy = list(row.implemented_by).map(text);
  return { kind: given.kind, file: given.file, twinsIncluding: including, hasOwnTwin: row.has_own_twin, implementedBy };
}

function unreadableOf(row: Record<string, unknown>): Unreadable | null {
  const given = requiredTexts(row, ["file", "code"]);
  return given === null ? null : { file: given.file, code: given.code, message: text(row.message) };
}

export function lineage(value: unknown): Lineage | null {
  if (!isRecord(value) || value.schema_version !== "0.1") return null;
  const plants = rows(value.plants, plantOf);
  const devices = rows(value.device_kinds, deviceOf);
  const unreadable = rows(value.unreadable, unreadableOf);
  return plants === null || devices === null || unreadable === null ? null : { plants, devices, unreadable };
}
