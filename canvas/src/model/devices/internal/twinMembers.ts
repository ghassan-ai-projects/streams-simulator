import type { CommandChoice, KindChoice, MetricChoice } from "./twinTypes";
import type { Command, KindFile, Metric } from "./types";

export const neutralMetric: MetricChoice = { instrument: "none", initial: null };

export const neutralCommand: CommandChoice = { response: "none", sets: {}, drives: [] };

function numeric(type: string): boolean {
  return type === "number" || type === "integer";
}

export function numericArguments(command: Command): readonly string[] {
  if (command.args.kind !== "rows") return [];
  return command.args.rows.filter((row) => numeric(row.type) && row.name !== "").map((row) => row.name);
}

export function writableReadings(kind: KindFile): readonly Metric[] {
  return kind.metrics
    .filter((metric) => metric.writable && numeric(metric.dataType) && metric.name !== "")
    .filter((metric) => metric.range.min !== null && metric.range.max !== null);
}

export function askedCommands(kind: KindFile): readonly Command[] {
  return kind.commands.filter((command) => !command.safeStop && command.name !== "");
}

export function metricChoiceOf(kind: KindChoice, key: string): MetricChoice {
  return kind.metrics[key] ?? neutralMetric;
}

export function commandChoiceOf(kind: KindChoice, key: string): CommandChoice {
  return kind.commands[key] ?? neutralCommand;
}

export function settleCommand(choice: CommandChoice): CommandChoice {
  if (choice.drives.length === 0) return choice;
  const sources = [...new Set(Object.values(choice.sets))];
  const [only] = sources;
  const from = (current: string) => (sources.length === 1 && only !== undefined ? only : keepSource(sources, current));
  return { ...choice, drives: choice.drives.map((drive) => ({ ...drive, from: from(drive.from) })) };
}

function keepSource(sources: readonly string[], current: string): string {
  return sources.includes(current) ? current : "";
}
