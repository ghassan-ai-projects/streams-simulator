import { noInstrument, suggestInstrument } from "./instruments";
import type { CommandChoice, Drive, Instrument, KindChoice, MetricChoice, TwinChoices } from "./twinTypes";
import type { Command, KindFile } from "./types";

function numeric(type: string): boolean {
  return type === "number" || type === "integer";
}

export function numericArguments(command: Command): readonly string[] {
  if (command.args.kind !== "rows") return [];
  return command.args.rows.filter((row) => numeric(row.type) && row.name !== "").map((row) => row.name);
}

export function writableReadings(kind: KindFile): readonly string[] {
  return kind.metrics
    .filter((metric) => metric.writable && numeric(metric.dataType) && metric.name !== "")
    .filter((metric) => metric.range.min !== null && metric.range.max !== null)
    .map((metric) => metric.name);
}

export function askedCommands(kind: KindFile): readonly Command[] {
  return kind.commands.filter((command) => !command.safeStop && command.name !== "");
}

function twinNameOf(kindName: string): string {
  const stem = kindName
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return `${stem}-twin`;
}

export function metricChoiceOf(kind: KindChoice, name: string): MetricChoice {
  return kind.metrics[name] ?? { instrument: noInstrument, initial: null };
}

export function commandChoiceOf(kind: KindChoice, name: string): CommandChoice {
  return kind.commands[name] ?? { response: "none", sets: {}, drives: [] };
}

function defaultSets(kind: KindFile, command: Command): Readonly<Record<string, string>> {
  const writable = writableReadings(kind);
  const named = numericArguments(command).filter((name) => writable.includes(name));
  const [onlyArgument, ...moreArguments] = numericArguments(command);
  const [onlyReading, ...moreReadings] = writable;
  if (named.length === 0 && onlyArgument !== undefined && onlyReading !== undefined) {
    return moreArguments.length === 0 && moreReadings.length === 0 ? { [onlyArgument]: onlyReading } : {};
  }
  return Object.fromEntries(named.map((name) => [name, name]));
}

function defaultCommand(kind: KindFile, command: Command): CommandChoice {
  const sets = defaultSets(kind, command);
  return { response: Object.keys(sets).length > 0 ? "direct" : "none", sets, drives: [] };
}

function defaultKind(kind: KindFile, instruments: readonly Instrument[]): KindChoice {
  const named = kind.metrics.filter((metric) => metric.name !== "");
  return {
    kind: kind.kind,
    instance: `${kind.kind}-01`,
    metrics: Object.fromEntries(
      named.map((metric) => [metric.name, { instrument: suggestInstrument(metric, instruments), initial: null }]),
    ),
    commands: Object.fromEntries(askedCommands(kind).map((command) => [command.name, defaultCommand(kind, command)])),
  };
}

export function startingChoices(kinds: readonly KindFile[], instruments: readonly Instrument[]): TwinChoices {
  const first = kinds[0]?.kind ?? "";
  return {
    id: twinNameOf(first),
    title: `${first} twin`,
    kinds: kinds.map((kind) => defaultKind(kind, instruments)),
    extra: {},
  };
}

function keepSets(kind: KindFile, command: Command, choice: CommandChoice): CommandChoice {
  const arguments_ = numericArguments(command);
  const writable = writableReadings(kind);
  const kept = Object.entries(choice.sets).filter(
    ([argument, reading]) => arguments_.includes(argument) && writable.includes(reading),
  );
  return { ...choice, sets: Object.fromEntries(kept) };
}

function reconcileCommands(kind: KindFile, base: KindChoice, existing: KindChoice) {
  const commands: Record<string, CommandChoice> = {};
  for (const command of askedCommands(kind)) {
    const previous = existing.commands[command.name];
    const choice = previous === undefined ? base.commands[command.name] : keepSets(kind, command, previous);
    if (choice !== undefined) commands[command.name] = choice;
  }
  return commands;
}

function reconcileKind(
  existing: KindChoice | undefined,
  kind: KindFile,
  instruments: readonly Instrument[],
): KindChoice {
  const base = defaultKind(kind, instruments);
  if (existing === undefined) return base;
  const metrics = Object.entries(base.metrics).map(([name, fresh]) => [name, existing.metrics[name] ?? fresh] as const);
  return {
    kind: kind.kind,
    instance: existing.instance,
    metrics: Object.fromEntries(metrics),
    commands: reconcileCommands(kind, base, existing),
  };
}

function followsKnown(drive: Drive, kinds: readonly KindChoice[]): boolean {
  return kinds.some((kind) => kind.kind === drive.kind && drive.metric in kind.metrics);
}

function pruned(choice: CommandChoice, kinds: readonly KindChoice[]): CommandChoice {
  const drives = choice.drives.filter((drive) => followsKnown(drive, kinds));
  return drives.length === choice.drives.length ? choice : { ...choice, drives };
}

function dropDangling(kind: KindChoice, kinds: readonly KindChoice[]): KindChoice {
  const commands = Object.entries(kind.commands).map(([name, choice]) => [name, pruned(choice, kinds)] as const);
  return { ...kind, commands: Object.fromEntries(commands) };
}

export function reconcileChoices(
  choices: TwinChoices,
  kinds: readonly KindFile[],
  instruments: readonly Instrument[],
): TwinChoices {
  const before = new Map(choices.kinds.map((kind) => [kind.kind, kind]));
  const rebuilt = kinds.map((kind) => reconcileKind(before.get(kind.kind), kind, instruments));
  return { ...choices, kinds: rebuilt.map((kind) => dropDangling(kind, rebuilt)) };
}
