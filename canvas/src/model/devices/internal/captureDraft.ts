import { isRecord, list, text } from "../../shared/json";
import { rangeWords } from "./instruments";
import { issueText, issueTitle, resolveIssues } from "./issues";
import { parseKind } from "./kindFile";
import type { Command, Fields, KindFile, Metric, RawIssue } from "./types";
import { commandFacts, readingFacts } from "./words";

interface Line {
  readonly name: string;
  readonly facts: string;
}

export interface DraftLines {
  readonly readings: readonly Line[];
  readonly commands: readonly Line[];
}

interface DraftIssue {
  readonly title: string;
  readonly text: string;
  readonly message: string;
}

const kindOrder = ["kind", "metrics", "commands"];
const metricOrder = ["name", "data_type", "unit", "eu_range", "resolution", "scan_class", "description", "writable"];
const commandOrder = ["name", "risk_class", "description", "safe_stop", "args_schema"];

function ordered(source: Fields, first: readonly string[]): Fields {
  const head = first.filter((name) => Object.hasOwn(source, name)).map((name) => [name, source[name]] as const);
  return Object.fromEntries([...head, ...Object.entries(source).filter(([name]) => !first.includes(name))]);
}

function orderedItems(value: unknown, first: readonly string[]): unknown {
  if (!Array.isArray(value)) return value;
  return value.map((entry: unknown) => (isRecord(entry) ? ordered(entry, first) : entry));
}

export function captureContent(draft: Fields): string {
  const members = { ...draft };
  if ("metrics" in draft) members.metrics = orderedItems(draft.metrics, metricOrder);
  if ("commands" in draft) members.commands = orderedItems(draft.commands, commandOrder);
  return `${JSON.stringify(ordered(members, kindOrder), null, 2)}\n`;
}

function kindOf(draft: Fields): KindFile | null {
  return parseKind(JSON.stringify(draft));
}

function rangePart(metric: Metric): string {
  const { min, max } = metric.range;
  return min === null && max === null ? "" : rangeWords(min, max);
}

function optionsPart(metric: Metric): string {
  const options = list(metric.extra.enum_values)
    .map(text)
    .filter((option) => option !== "");
  return options.length === 0 ? "" : `options ${options.join(", ")}`;
}

function joined(parts: readonly string[]): string {
  return parts.filter((part) => part !== "").join(" · ");
}

function readingLine(metric: Metric): Line {
  const setpoint = metric.writable ? "setpoint" : "";
  const facts = [readingFacts(metric), rangePart(metric), optionsPart(metric), setpoint];
  return { name: metric.name === "" ? "Unnamed reading" : metric.name, facts: joined(facts) };
}

function argumentWords(command: Command): string {
  if (command.args.kind === "custom") return "custom arguments";
  const names = command.args.rows.map((row) => row.name).filter((name) => name !== "");
  return names.length === 0 ? "no arguments" : `takes ${names.join(", ")}`;
}

function commandLine(command: Command): Line {
  const name = command.name === "" ? "Unnamed command" : command.name;
  return { name, facts: joined([argumentWords(command), commandFacts(command)]) };
}

export function captureLines(draft: Fields): DraftLines | null {
  const kind = kindOf(draft);
  if (kind === null) return null;
  return { readings: kind.metrics.map(readingLine), commands: kind.commands.map(commandLine) };
}

export function draftIssues(draft: Fields, issues: readonly RawIssue[]): readonly DraftIssue[] {
  const kind = kindOf(draft);
  if (kind === null) return issues.map((raw) => ({ title: "", text: issueText(raw), message: raw.message }));
  return resolveIssues(kind, issues).map((issue) => ({
    title: issueTitle(kind, issue),
    text: issue.text,
    message: issue.message,
  }));
}

export function differenceWords(draft: Fields, pointer: string): string {
  const kind = kindOf(draft);
  const placed = kind === null ? undefined : resolveIssues(kind, [{ pointer, message: pointer }]).at(0);
  if (kind === null || placed === undefined || placed.part === "file") {
    return "The first difference is somewhere in the file.";
  }
  return `The first difference is in ${issueTitle(kind, placed)}.`;
}
