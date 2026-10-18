import type { ArgRow, Command, FieldIssue, KindFile, Range } from "./types";
import { localIssue } from "./issues";

export function rangeHint(range: Range): string {
  const inverted = range.min !== null && range.max !== null && range.min > range.max;
  return inverted ? "The minimum is above the maximum." : "";
}

export function devicePath(name: string): string {
  return `kinds/${name}.kind.json`;
}

export function deviceNameProblem(name: string, taken: readonly string[]): string {
  const stem = name.trim();
  if (stem === "") return "Give the device a name.";
  const same = (other: string) => other.toLowerCase() === stem.toLowerCase();
  return taken.some(same) ? `A device named ${stem} already exists.` : "";
}

function argumentProblems(command: Command): readonly FieldIssue[] {
  if (command.args.kind !== "rows") return [];
  const rows: readonly ArgRow[] = command.args.rows;
  const named = rows.map((row) => row.name.trim());
  const blank = named.includes("");
  const repeated = named.some((name, at) => name !== "" && named.indexOf(name) !== at);
  return [
    ...(blank ? [localIssue("command", command.key, "args_schema", "Give every argument a name.")] : []),
    ...(repeated ? [localIssue("command", command.key, "args_schema", "Two arguments share a name.")] : []),
  ];
}

export function draftProblems(kind: KindFile): readonly FieldIssue[] {
  return [
    ...(kind.kind.trim() === "" ? [localIssue("kind", "", "kind", "Give the device a name.")] : []),
    ...kind.metrics
      .filter((metric) => metric.name.trim() === "")
      .map((metric) => localIssue("metric", metric.key, "name", "Give this reading a name.")),
    ...kind.commands
      .filter((command) => command.name.trim() === "")
      .map((command) => localIssue("command", command.key, "name", "Give this command a name.")),
    ...kind.commands.flatMap(argumentProblems),
  ];
}
