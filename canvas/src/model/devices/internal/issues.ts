import { isTwinKeyword, plainIssue } from "./issueWords";
import type { FieldIssue, KindFile, RawIssue, Selection } from "./types";

interface Owner {
  readonly part: "metric" | "command";
  readonly key: string;
  readonly name: string;
}

const fieldWords: Readonly<Record<string, string>> = {
  kind: "Device name",
  name: "Name",
  data_type: "Type",
  unit: "Unit",
  eu_range: "Range",
  resolution: "Resolution",
  scan_class: "Scan class",
  enum_values: "Options",
  writable: "Setpoint",
  description: "Description",
  risk_class: "Risk class",
  safe_stop: "Safe stop",
  args_schema: "Arguments",
  metrics: "Readings",
  commands: "Commands",
};

function unescaped(part: string): string {
  return part.replaceAll("~1", "/").replaceAll("~0", "~");
}

function ownerOf(kind: KindFile, collection: string, index: string): Owner | null {
  if (!/^\d+$/.test(index)) return null;
  const at = Number(index);
  const metric = collection === "metrics" ? kind.metrics[at] : undefined;
  const command = collection === "commands" ? kind.commands[at] : undefined;
  if (metric !== undefined) return { part: "metric", key: metric.key, name: metric.name };
  return command === undefined ? null : { part: "command", key: command.key, name: command.name };
}

function issueOf(where: Pick<FieldIssue, "part" | "key" | "field">, name: string, raw: RawIssue): FieldIssue {
  const keyword = raw.keyword ?? "";
  const member = raw.member ?? "";
  const label = fieldWords[member === "" ? where.field : member] ?? "This value";
  const text = plainIssue(keyword, { name, field: where.field, member, label });
  return { ...where, keyword, message: raw.message, text, twin: isTwinKeyword(keyword) };
}

function knownField(field: string): string {
  return Object.hasOwn(fieldWords, field) ? field : "";
}

function unplaced(collection: string, index: string): Pick<FieldIssue, "part" | "key" | "field"> {
  if (index !== "") return { part: "file", key: "", field: "" };
  return { part: "kind", key: "", field: knownField(collection) };
}

function resolveIssue(kind: KindFile, raw: RawIssue): FieldIssue {
  const [collection = "", index = "", field = ""] = raw.pointer.split("/").slice(1).map(unescaped);
  const member = raw.member ?? "";
  if (collection === "") return issueOf({ part: "file", key: "", field: "" }, kind.kind, raw);
  if (collection === "kind") return issueOf({ part: "kind", key: "", field: "kind" }, kind.kind, raw);
  const owner = ownerOf(kind, collection, index);
  if (owner === null) return issueOf(unplaced(collection, index), kind.kind, raw);
  const named = knownField(field === "" ? member : field);
  return issueOf({ part: owner.part, key: owner.key, field: named }, owner.name, raw);
}

export function resolveIssues(kind: KindFile, issues: readonly RawIssue[]): readonly FieldIssue[] {
  return issues.map((issue) => resolveIssue(kind, issue));
}

export function issueText(raw: RawIssue): string {
  const member = raw.member ?? "";
  const label = fieldWords[member] ?? "This value";
  return plainIssue(raw.keyword ?? "", { name: "", field: "", member, label });
}

export function localIssue(part: FieldIssue["part"], key: string, field: string, text: string): FieldIssue {
  return { part, key, field, keyword: "", message: text, text, twin: false };
}

export function issuesOn(
  issues: readonly FieldIssue[],
  part: FieldIssue["part"],
  key: string,
  field: string,
): readonly string[] {
  return issues
    .filter((issue) => issue.part === part && issue.key === key && issue.field === field)
    .map((issue) => issue.text);
}

export function issueSelection(issue: FieldIssue): Selection {
  if (issue.part === "metric" || issue.part === "command") return { part: issue.part, key: issue.key };
  return { part: "kind" };
}

export function schemaIssues(issues: readonly FieldIssue[]): readonly FieldIssue[] {
  return issues.filter((issue) => !issue.twin);
}

export function twinNeeds(issues: readonly FieldIssue[]): readonly FieldIssue[] {
  return issues.filter((issue) => issue.twin);
}

export function issueTitle(kind: KindFile, issue: FieldIssue): string {
  const field = fieldWords[issue.field] ?? "";
  const metric = kind.metrics.find((entry) => entry.key === issue.key);
  const command = kind.commands.find((entry) => entry.key === issue.key);
  const owner = issue.part === "metric" ? metric?.name : command?.name;
  const subject = { kind: "Device", file: "File", metric: "Reading", command: "Command" }[issue.part];
  const named = owner === undefined || owner === "" ? subject : `${subject} ${owner}`;
  return field === "" ? named : `${named} · ${field}`;
}
