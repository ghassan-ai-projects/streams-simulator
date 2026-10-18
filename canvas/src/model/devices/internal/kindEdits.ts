import { blankArguments, blankRow } from "./args";
import type { ArgRow, Command, KindFile, Metric, Selection } from "./types";

type Change<T> = Partial<Omit<T, "key">>;

function nextKey(prefix: string, keys: readonly string[]): string {
  const used = keys.map((key) => Number(key.slice(prefix.length))).filter((number) => Number.isFinite(number));
  return `${prefix}${String(Math.max(0, ...used) + 1)}`;
}

export function renameKind(kind: KindFile, name: string): KindFile {
  return { ...kind, kind: name };
}

export function addMetric(kind: KindFile): { readonly kind: KindFile; readonly selected: Selection } {
  const key = nextKey(
    "m",
    kind.metrics.map((metric) => metric.key),
  );
  const metric: Metric = {
    key,
    name: "",
    dataType: "number",
    unit: "",
    range: { min: null, max: null },
    resolution: null,
    scanClass: "",
    writable: false,
    description: "",
    extra: {},
  };
  return { kind: { ...kind, metrics: [...kind.metrics, metric] }, selected: { part: "metric", key } };
}

export function addCommand(kind: KindFile): { readonly kind: KindFile; readonly selected: Selection } {
  const key = nextKey(
    "c",
    kind.commands.map((command) => command.key),
  );
  const command: Command = {
    key,
    name: "",
    riskClass: "low",
    safeStop: false,
    description: "",
    args: blankArguments,
    extra: {},
  };
  return { kind: { ...kind, commands: [...kind.commands, command] }, selected: { part: "command", key } };
}

export function changeMetric(kind: KindFile, key: string, change: Change<Metric>): KindFile {
  return { ...kind, metrics: kind.metrics.map((metric) => (metric.key === key ? { ...metric, ...change } : metric)) };
}

export function changeCommand(kind: KindFile, key: string, change: Change<Command>): KindFile {
  return {
    ...kind,
    commands: kind.commands.map((command) => (command.key === key ? { ...command, ...change } : command)),
  };
}

export function removeItem(kind: KindFile, selection: Selection): KindFile {
  if (selection.part === "metric") {
    return { ...kind, metrics: kind.metrics.filter((metric) => metric.key !== selection.key) };
  }
  if (selection.part === "command") {
    return { ...kind, commands: kind.commands.filter((command) => command.key !== selection.key) };
  }
  return kind;
}

export function hasItem(kind: KindFile, selection: Selection): boolean {
  if (selection.part === "metric") return kind.metrics.some((metric) => metric.key === selection.key);
  if (selection.part === "command") return kind.commands.some((command) => command.key === selection.key);
  return true;
}

function withRows(kind: KindFile, key: string, edit: (rows: readonly ArgRow[]) => readonly ArgRow[]): KindFile {
  const args = kind.commands.find((entry) => entry.key === key)?.args;
  if (args?.kind !== "rows") return kind;
  return changeCommand(kind, key, { args: { ...args, rows: edit(args.rows) } });
}

export function addArgument(kind: KindFile, key: string): KindFile {
  return withRows(kind, key, (rows) => [...rows, blankRow]);
}

export function removeArgument(kind: KindFile, key: string, index: number): KindFile {
  return withRows(kind, key, (rows) => rows.filter((_row, at) => at !== index));
}

export function changeArgument(kind: KindFile, key: string, index: number, change: Partial<ArgRow>): KindFile {
  return withRows(kind, key, (rows) => rows.map((row, at) => (at === index ? { ...row, ...change } : row)));
}
