import type { TwinChoices, TwinIssue } from "./twinTypes";
import type { KindFile, RawIssue } from "./types";

const fieldWords: Readonly<Record<string, string>> = {
  resolution: "Resolution",
  id: "Twin name",
  title: "Title",
  kind: "Device",
  instance: "Name in the twin",
  instrument: "Instrument",
  initial: "Starts at",
  response: "Response",
  sets: "Reading it sets",
  drives: "Effect on a reading",
  from: "Reading it sets",
  follows: "Reading that follows",
  gain: "Gain",
  time_constant_s: "Time constant",
};

function unescaped(part: string): string {
  return part.replaceAll("~1", "/").replaceAll("~0", "~");
}

function elsewhere(message: string): TwinIssue {
  return { scope: "elsewhere", kind: "", key: "", field: "", spot: "", message };
}

function commandIssue(kind: string, key: string, tail: readonly string[], message: string): TwinIssue {
  const [field = "", spot = "", sub = ""] = tail;
  const base = { scope: "command", kind, key, message } as const;
  if (field !== "drives") return { ...base, field, spot: field === "sets" ? spot : "" };
  const named = sub === "kind" || sub === "metric" ? "follows" : sub;
  return { ...base, field: named === "" ? "drives" : named, spot };
}

const needs = /\/(metrics|commands)\/(\d+) needs "([a-z_]+)"/;

const labels: Readonly<Record<string, string>> = { writable: "Setpoint" };

function needWord(field: string): string {
  return (fieldWords[field] ?? field.replaceAll("_", " ")).toLowerCase();
}

export function interfaceLabel(field: string): string {
  const words = field.replaceAll("_", " ");
  return labels[field] ?? `${words.charAt(0).toUpperCase()}${words.slice(1)}`;
}

function memberNeed(kinds: readonly KindFile[], kind: string, message: string): TwinIssue | null {
  const found = needs.exec(message);
  const file = kinds.find((entry) => entry.kind === kind);
  if (found === null || file === undefined) return null;
  const [, section, index, field] = found;
  const owner = (section === "metrics" ? file.metrics : file.commands)[Number(index)];
  if (owner === undefined || field === undefined) return null;
  const scope = section === "metrics" ? "metric" : "command";
  return { scope, kind, key: owner.name, field, spot: "", message: `Needs a ${needWord(field)}.` };
}

function kindIssue(kinds: readonly KindFile[], kind: string, rest: readonly string[], message: string): TwinIssue {
  const [member = "", key = "", ...tail] = rest;
  if (member === "metrics" && key !== "") {
    return { scope: "metric", kind, key, field: tail[0] ?? "", spot: "", message };
  }
  if (member === "commands" && key !== "") return commandIssue(kind, key, tail, message);
  const need = member === "kind" ? memberNeed(kinds, kind, message) : null;
  return need ?? { scope: "kind", kind, key: "", field: member, spot: "", message };
}

function placeIssue(choices: TwinChoices, kinds: readonly KindFile[], { pointer, message }: RawIssue): TwinIssue {
  const [head = "", index = "", ...rest] = pointer.split("/").slice(1).map(unescaped);
  if (head === "") return elsewhere(message);
  if (head !== "kinds") return { scope: "twin", kind: "", key: "", field: head, spot: "", message };
  const choice = /^\d+$/.test(index) ? choices.kinds[Number(index)] : undefined;
  if (index === "") return { scope: "twin", kind: "", key: "", field: "kinds", spot: "", message };
  return choice === undefined ? elsewhere(message) : kindIssue(kinds, choice.kind, rest, message);
}

export function resolveTwinIssues(
  choices: TwinChoices,
  issues: readonly RawIssue[],
  kinds: readonly KindFile[],
): readonly TwinIssue[] {
  return issues.map((issue) => placeIssue(choices, kinds, issue));
}

const interfaceFields = new Set([
  "name",
  "data_type",
  "unit",
  "eu_range",
  "resolution",
  "writable",
  "scan_class",
  "description",
  "args_schema",
  "risk_class",
  "safe_stop",
]);

export function interfaceNeeds(
  issues: readonly TwinIssue[],
  scope: "metric" | "command",
  kind: string,
  key: string,
): readonly TwinIssue[] {
  return issues.filter(
    (issue) => issue.scope === scope && issue.kind === kind && issue.key === key && interfaceFields.has(issue.field),
  );
}

export function twinIssuesOn(
  issues: readonly TwinIssue[],
  where: Pick<TwinIssue, "scope" | "kind" | "key" | "field"> & { readonly spot?: string },
): readonly string[] {
  return issues
    .filter((issue) => issue.scope === where.scope && issue.kind === where.kind && issue.key === where.key)
    .filter((issue) => issue.field === where.field && issue.spot === (where.spot ?? ""))
    .map((issue) => issue.message);
}

export function twinIssueTitle(issue: TwinIssue): string {
  const field = fieldWords[issue.field] ?? "";
  const subject = [issue.kind, issue.key].filter((part) => part !== "");
  const parts = [...subject, ...(field === "" ? [] : [field])];
  return parts.length === 0 ? "Twin" : parts.join(" · ");
}
