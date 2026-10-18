import { isRecord, list } from "../../shared/json";
import { plainIssue } from "./issueWords";
import { choicesJson } from "./twinFile";
import { nameChoices } from "./twinKeys";
import type { TwinChoices, TwinIssue } from "./twinTypes";
import type { Fields, KindFile, Member, RawIssue } from "./types";

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

type Place = Pick<TwinIssue, "scope" | "kind" | "key" | "field" | "spot">;

function unescaped(part: string): string {
  return part.replaceAll("~1", "/").replaceAll("~0", "~");
}

function issueAt(place: Place, name: string, raw: RawIssue): TwinIssue {
  const keyword = raw.keyword ?? "";
  const member = raw.member ?? "";
  const label = fieldWords[member === "" ? place.field : member] ?? "This value";
  const text = plainIssue(keyword, { name, field: place.field, member, label });
  return { ...place, pointer: raw.pointer, keyword, message: raw.message, text };
}

function metricKey(kind: KindFile, name: string): string {
  return kind.metrics.find((metric) => metric.name === name)?.key ?? "";
}

function commandKey(kind: KindFile, name: string): string {
  return kind.commands.find((command) => command.name === name)?.key ?? "";
}

function commandPlace(path: string, key: string, tail: readonly string[]): Place {
  const [field = "", spot = "", sub = ""] = tail;
  const base = { scope: "command", kind: path, key } as const;
  if (field !== "drives") return { ...base, field, spot: field === "sets" ? spot : "" };
  const named = sub === "kind" || sub === "metric" ? "follows" : sub;
  return { ...base, field: named === "" ? "drives" : named, spot };
}

function kindPlace(member: Member, rest: readonly string[]): Place {
  const [section = "", name = "", ...tail] = rest;
  const base = { kind: member.path, spot: "" };
  if (section === "metrics" && name !== "") {
    return { ...base, scope: "metric", key: metricKey(member.kind, name), field: tail[0] ?? "" };
  }
  if (section === "commands" && name !== "") return commandPlace(member.path, commandKey(member.kind, name), tail);
  return { ...base, scope: "kind", key: "", field: section };
}

const nowhere: Place = { scope: "elsewhere", kind: "", key: "", field: "", spot: "" };

function placeIssue(choices: TwinChoices, members: readonly Member[], raw: RawIssue): TwinIssue {
  const [head = "", index = "", ...rest] = raw.pointer.split("/").slice(1).map(unescaped);
  if (head === "") return issueAt(nowhere, "", raw);
  if (head !== "kinds" || index === "") return issueAt({ ...nowhere, scope: "twin", field: head }, "", raw);
  const choice = /^\d+$/.test(index) ? choices.kinds[Number(index)] : undefined;
  const member = members.find((entry) => entry.path === choice?.kind);
  if (member === undefined) return issueAt(nowhere, "", raw);
  const place = kindPlace(member, rest);
  return issueAt(place, nameOf(member.kind, place), raw);
}

function nameOf(kind: KindFile, place: Place): string {
  if (place.scope === "metric") return kind.metrics.find((metric) => metric.key === place.key)?.name ?? "";
  if (place.scope === "command") return kind.commands.find((command) => command.key === place.key)?.name ?? "";
  return kind.kind;
}

export function resolveTwinIssues(
  choices: TwinChoices,
  issues: readonly RawIssue[],
  members: readonly Member[],
): readonly TwinIssue[] {
  return issues.map((issue) => placeIssue(choices, members, issue));
}

function child(found: unknown, part: string): unknown {
  if (Array.isArray(found)) return list(found)[Number(part)];
  return isRecord(found) ? found[part] : undefined;
}

function valueAt(document: Fields, pointer: string): string {
  const parts = pointer.split("/").slice(1).map(unescaped);
  const found = parts.reduce<unknown>(child, document);
  return found === undefined ? "" : JSON.stringify(found);
}

export function keepCurrentIssues(
  issues: readonly TwinIssue[],
  before: TwinChoices,
  after: TwinChoices,
  members: readonly Member[],
): readonly TwinIssue[] {
  if (issues.length === 0) return issues;
  const was = choicesJson(nameChoices(before, members));
  const is = choicesJson(nameChoices(after, members));
  return issues.filter((issue) => issue.pointer === "" || valueAt(was, issue.pointer) === valueAt(is, issue.pointer));
}

export function twinIssuesOn(issues: readonly TwinIssue[], where: Omit<Place, "spot"> & { readonly spot?: string }) {
  return issues
    .filter((issue) => issue.scope === where.scope && issue.kind === where.kind && issue.key === where.key)
    .filter((issue) => issue.field === where.field && issue.spot === (where.spot ?? ""))
    .map((issue) => issue.text);
}

export function twinIssueTitle(issue: TwinIssue, members: readonly Member[]): string {
  const member = members.find((entry) => entry.path === issue.kind);
  const item = member !== undefined && (issue.scope === "metric" || issue.scope === "command");
  const parts = [member?.kind.kind ?? "", item ? nameOf(member.kind, issue) : "", fieldWords[issue.field] ?? ""];
  const named = parts.filter((part) => part !== "");
  return named.length === 0 ? "Twin" : named.join(" · ");
}
