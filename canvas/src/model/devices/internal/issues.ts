import type { FieldIssue, KindFile, RawIssue, Selection } from "./types";

interface Owner {
  readonly part: "metric" | "command";
  readonly key: string;
}

function ownerOf(kind: KindFile, collection: string, index: string): Owner | null {
  if (!/^\d+$/.test(index)) return null;
  const at = Number(index);
  const metric = collection === "metrics" ? kind.metrics[at] : undefined;
  const command = collection === "commands" ? kind.commands[at] : undefined;
  if (metric !== undefined) return { part: "metric", key: metric.key };
  return command === undefined ? null : { part: "command", key: command.key };
}

function resolveIssue(kind: KindFile, { pointer, message }: RawIssue): FieldIssue {
  const [collection = "", index = "", field = ""] = pointer.split("/").slice(1);
  if (collection === "kind") return { part: "kind", key: "", field: "kind", message };
  const owner = ownerOf(kind, collection, index);
  return owner === null ? { part: "file", key: "", field: "", message } : { ...owner, field, message };
}

export function resolveIssues(kind: KindFile, issues: readonly RawIssue[]): readonly FieldIssue[] {
  return issues.map((issue) => resolveIssue(kind, issue));
}

export function issuesOn(
  issues: readonly FieldIssue[],
  part: FieldIssue["part"],
  key: string,
  field: string,
): readonly string[] {
  return issues
    .filter((issue) => issue.part === part && issue.key === key && issue.field === field)
    .map((issue) => issue.message);
}

export function issueSelection(issue: FieldIssue): Selection {
  if (issue.part === "metric" || issue.part === "command") return { part: issue.part, key: issue.key };
  return { part: "kind" };
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
  writable: "Writable",
  description: "Description",
  risk_class: "Risk class",
  safe_stop: "Safe stop",
  args_schema: "Arguments",
};

export function issueTitle(kind: KindFile, issue: FieldIssue): string {
  const field = fieldWords[issue.field] ?? "";
  const metric = kind.metrics.find((entry) => entry.key === issue.key);
  const command = kind.commands.find((entry) => entry.key === issue.key);
  const owner = issue.part === "metric" ? metric?.name : command?.name;
  const subject = { kind: "Device", file: "File", metric: "Reading", command: "Command" }[issue.part];
  const named = owner === undefined || owner === "" ? subject : `${subject} ${owner}`;
  return field === "" ? named : `${named} · ${field}`;
}
