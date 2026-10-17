import { isRecord, list, text } from "../../shared/json";
import { argumentsOf, schemaOf } from "./args";
import type { Command, Fields, KindFile, Metric, Range } from "./types";

const kindFields = new Set(["kind", "metrics", "commands"]);
const metricFields = new Set([
  "name",
  "data_type",
  "unit",
  "eu_range",
  "resolution",
  "scan_class",
  "writable",
  "description",
]);
const commandFields = new Set(["name", "risk_class", "safe_stop", "description", "args_schema"]);

function rest(source: Fields, known: ReadonlySet<string>): Fields {
  return Object.fromEntries(Object.entries(source).filter(([name]) => !known.has(name)));
}

function number(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

function metricOf(source: Fields, key: string): Metric {
  const range = isRecord(source.eu_range) ? source.eu_range : {};
  return {
    key,
    name: text(source.name),
    dataType: text(source.data_type),
    unit: text(source.unit),
    range: { min: number(range.min), max: number(range.max) },
    resolution: number(source.resolution),
    scanClass: text(source.scan_class),
    writable: source.writable === true,
    description: text(source.description),
    extra: rest(source, metricFields),
  };
}

function commandOf(source: Fields, key: string): Command {
  return {
    key,
    name: text(source.name),
    riskClass: text(source.risk_class),
    safeStop: source.safe_stop === true,
    description: text(source.description),
    args: argumentsOf(isRecord(source.args_schema) ? source.args_schema : {}),
    extra: rest(source, commandFields),
  };
}

function items<T>(value: unknown, prefix: string, build: (source: Fields, key: string) => T): T[] | null {
  if (value === undefined) return [];
  if (!Array.isArray(value)) return null;
  const built: T[] = [];
  for (const [index, entry] of list(value).entries()) {
    if (!isRecord(entry)) return null;
    built.push(build(entry, `${prefix}${String(index + 1)}`));
  }
  return built;
}

export function parseKind(source: string): KindFile | null {
  let parsed: unknown;
  try {
    parsed = JSON.parse(source);
  } catch {
    return null;
  }
  if (!isRecord(parsed)) return null;
  const metrics = items(parsed.metrics, "m", metricOf);
  const commands = items(parsed.commands, "c", commandOf);
  if (metrics === null || commands === null) return null;
  return { kind: text(parsed.kind), metrics, commands, extra: rest(parsed, kindFields) };
}

function present(entries: readonly (readonly [string, unknown])[]): Fields {
  return Object.fromEntries(entries.filter(([, value]) => value !== "" && value !== null && value !== false));
}

function rangeJson(range: Range): Fields | null {
  const bounds = {
    ...(range.min === null ? {} : { min: range.min }),
    ...(range.max === null ? {} : { max: range.max }),
  };
  return Object.keys(bounds).length === 0 ? null : bounds;
}

function metricJson(metric: Metric): Fields {
  return {
    name: metric.name,
    ...present([
      ["data_type", metric.dataType],
      ["unit", metric.unit],
      ["eu_range", rangeJson(metric.range)],
      ["resolution", metric.resolution],
      ["scan_class", metric.scanClass],
      ["writable", metric.writable],
      ["description", metric.description],
    ]),
    ...metric.extra,
  };
}

function commandJson(command: Command): Fields {
  return {
    name: command.name,
    ...present([
      ["risk_class", command.riskClass],
      ["safe_stop", command.safeStop],
      ["description", command.description],
    ]),
    args_schema: schemaOf(command.args),
    ...command.extra,
  };
}

export function serialiseKind(kind: KindFile): string {
  const document = {
    kind: kind.kind,
    metrics: kind.metrics.map(metricJson),
    commands: kind.commands.map(commandJson),
    ...kind.extra,
  };
  return `${JSON.stringify(document, null, 2)}\n`;
}

export function blankKind(name: string): KindFile {
  return { kind: name, metrics: [], commands: [], extra: {} };
}
