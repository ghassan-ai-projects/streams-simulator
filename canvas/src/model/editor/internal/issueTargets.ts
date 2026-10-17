import { groupBy } from "../../shared/collect";
import { record } from "../../shared/json";
import { counted } from "../../shared/words";
import type { Composition } from "./composition";
import { connectionsOf, equipmentOf } from "./composition";
import { rowsOf } from "./rows";
import { endpointOf } from "../../plant";

interface Issue {
  readonly pointer: string;
  readonly message: string;
}

export interface ListedIssue {
  readonly target: string | null;
  readonly message: string;
  readonly pointer: string;
}

export interface IssueTargets {
  readonly instances: ReadonlyMap<string, readonly string[]>;
  readonly listed: readonly ListedIssue[];
}

export interface IssueGroup extends ListedIssue {
  readonly count: number;
}

function groupKey(issue: ListedIssue): string {
  return `${issue.target ?? ""}\n${issue.message}\n${issue.target === null ? issue.pointer : ""}`;
}

export function issuesWords(count: number): string {
  return counted(count, "issue");
}

export function groupIssues(listed: readonly ListedIssue[]): readonly IssueGroup[] {
  return [...groupBy(listed, groupKey).values()].flatMap((same) => {
    const latest = same.at(-1);
    return latest === undefined ? [] : [{ ...latest, count: same.length }];
  });
}

function equipmentTargets(doc: Composition, at: number, field: string, index: string): readonly string[] {
  const names = equipmentOf(doc)[at]?.names ?? [];
  if (field !== "names" || !/^\d+$/.test(index)) return names;
  const name = names[Number(index)];
  return name === undefined ? [] : [name];
}

function targetsOf(doc: Composition, pointer: string): readonly string[] {
  const [, collection = "", index = "", field = "", item = ""] = pointer.split("/");
  if (!/^\d+$/.test(index)) return [];
  const at = Number(index);
  if (collection === "equipment") return equipmentTargets(doc, at, field, item);
  if (collection !== "connections") return [];
  const wire = connectionsOf(doc)[at];
  return wire === undefined ? [] : [endpointOf(wire.from).id, endpointOf(wire.to).id];
}

export function issueTargets(doc: Composition, issues: readonly Issue[]): IssueTargets {
  const targeted = issues.map((issue) => ({ issue, targets: targetsOf(doc, issue.pointer) }));
  const hits = targeted.flatMap(({ issue, targets }) => targets.map((id) => ({ id, message: issue.message })));
  const instances = new Map(
    [...groupBy(hits, (hit) => hit.id)].map(([id, same]) => [id, same.map((hit) => hit.message)] as const),
  );
  const listed = targeted.map(({ issue, targets }) => ({
    target: targets[0] ?? null,
    message: issue.message,
    pointer: issue.pointer,
  }));
  return { instances, listed };
}

const weightPointer = /^\/profiles\/(\d+)\/fault_weights\/(.+)$/;
const undeclared = /^profile "(.+)" weights undeclared fault "(.+)"$/;

function unescaped(token: string): string {
  return token.replaceAll("~1", "/").replaceAll("~0", "~");
}

export function profileWeightOf(issue: ListedIssue): { readonly profile: string; readonly fault: string } | null {
  const match = undeclared.exec(issue.message);
  return match?.[1] === undefined || match[2] === undefined ? null : { profile: match[1], fault: match[2] };
}

export function removeProfileWeight(doc: Composition, pointer: string): Composition {
  const match = weightPointer.exec(pointer);
  const profiles = rowsOf(doc.profiles);
  const at = Number(match?.[1] ?? -1);
  const profile = profiles[at];
  if (match?.[2] === undefined || profile === undefined) return doc;
  const fault = unescaped(match[2]);
  const weights = Object.entries(record(profile.fault_weights)).filter(([name]) => name !== fault);
  const edited = { ...profile, fault_weights: Object.fromEntries(weights) };
  return { ...doc, profiles: profiles.map((entry, index) => (index === at ? edited : entry)) };
}
