import type { ArgRow, Command, FieldIssue, KindFile, Range } from "./types";

const idPattern = /^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$/;
const fileStem = /^[A-Za-z0-9][A-Za-z0-9._-]{0,117}$/;

export const idRule = "Letters, digits and . _ : - only, starting with a letter or digit.";

export function looksLikeId(name: string): boolean {
  return idPattern.test(name);
}

export function rangeHint(range: Range): string {
  const inverted = range.min !== null && range.max !== null && range.min > range.max;
  return inverted ? "The minimum is above the maximum." : "";
}

export function devicePath(name: string): string {
  return `kinds/${name}.kind.json`;
}

export function looksLikeFileStem(name: string): boolean {
  return fileStem.test(name);
}

export function deviceNameProblem(name: string, taken: readonly string[]): string {
  const stem = name.trim();
  if (stem === "") return "Give the device a name.";
  if (!fileStem.test(stem)) return "Use letters, digits and . _ - only, starting with a letter or digit.";
  const same = (other: string) => other.toLowerCase() === stem.toLowerCase();
  return taken.some(same) ? `A device named ${stem} already exists.` : "";
}

function problem(part: FieldIssue["part"], key: string, field: string, message: string): FieldIssue {
  return { part, key, field, message };
}

function argumentProblems(command: Command): readonly FieldIssue[] {
  if (command.args.kind !== "rows") return [];
  const rows: readonly ArgRow[] = command.args.rows;
  const named = rows.map((row) => row.name.trim());
  const blank = named.includes("");
  const repeated = named.some((name, at) => name !== "" && named.indexOf(name) !== at);
  return [
    ...(blank ? [problem("command", command.key, "args_schema", "Give every argument a name.")] : []),
    ...(repeated ? [problem("command", command.key, "args_schema", "Two arguments share a name.")] : []),
  ];
}

export function draftProblems(kind: KindFile): readonly FieldIssue[] {
  return [
    ...(kind.kind.trim() === "" ? [problem("kind", "", "kind", "Give the device a name.")] : []),
    ...kind.metrics
      .filter((metric) => metric.name.trim() === "")
      .map((metric) => problem("metric", metric.key, "name", "Give this reading a name.")),
    ...kind.commands
      .filter((command) => command.name.trim() === "")
      .map((command) => problem("command", command.key, "name", "Give this command a name.")),
    ...kind.commands.flatMap(argumentProblems),
  ];
}
