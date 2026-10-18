import { isRecord, list, record, text } from "../../shared/json";
import { noInstrument } from "./instruments";
import type { CommandChoice, Drive, KindChoice, MetricChoice, Response, TwinChoices } from "./twinTypes";

const known = new Set(["schema_version", "id", "title", "kinds"]);

function number(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

export function responseOf(value: unknown): Response {
  return value === "direct" || value === "first_order" ? value : "none";
}

function driveOf(source: unknown): Drive {
  const drive = record(source);
  return {
    from: text(drive.from),
    kind: text(drive.kind),
    metric: text(drive.metric),
    gain: number(drive.gain),
    timeConstantS: number(drive.time_constant_s),
  };
}

function commandOf(source: unknown): CommandChoice {
  const choice = record(source);
  const sets = Object.entries(record(choice.sets)).filter(
    (entry): entry is [string, string] => typeof entry[1] === "string",
  );
  return {
    response: responseOf(choice.response),
    sets: Object.fromEntries(sets),
    drives: list(choice.drives).map(driveOf),
  };
}

function metricOf(source: unknown): MetricChoice {
  const choice = record(source);
  return { instrument: text(choice.instrument) || noInstrument, initial: number(choice.initial) };
}

function kindOf(source: unknown): KindChoice {
  const choice = record(source);
  const metrics = Object.entries(record(choice.metrics)).map(([name, entry]) => [name, metricOf(entry)] as const);
  const commands = Object.entries(record(choice.commands)).map(([name, entry]) => [name, commandOf(entry)] as const);
  return {
    kind: text(choice.kind),
    instance: text(choice.instance),
    metrics: Object.fromEntries(metrics),
    commands: Object.fromEntries(commands),
  };
}

export function parseChoices(source: string): TwinChoices | null {
  let parsed: unknown;
  try {
    parsed = JSON.parse(source);
  } catch {
    return null;
  }
  if (!isRecord(parsed)) return null;
  const extra = Object.entries(parsed).filter(([name]) => !known.has(name));
  return {
    id: text(parsed.id),
    title: text(parsed.title),
    kinds: list(parsed.kinds).map(kindOf),
    extra: Object.fromEntries(extra),
  };
}

function metricJson(choice: MetricChoice) {
  return { instrument: choice.instrument, ...(choice.initial === null ? {} : { initial: choice.initial }) };
}

function driveJson(drive: Drive) {
  return {
    from: drive.from,
    kind: drive.kind,
    metric: drive.metric,
    ...(drive.gain === null ? {} : { gain: drive.gain }),
    ...(drive.timeConstantS === null ? {} : { time_constant_s: drive.timeConstantS }),
  };
}

function commandJson(choice: CommandChoice) {
  return {
    response: choice.response,
    ...(Object.keys(choice.sets).length === 0 ? {} : { sets: choice.sets }),
    ...(choice.drives.length === 0 ? {} : { drives: choice.drives.map(driveJson) }),
  };
}

function kindJson(choice: KindChoice) {
  return {
    kind: choice.kind,
    instance: choice.instance,
    metrics: Object.fromEntries(
      Object.entries(choice.metrics).map(([name, entry]) => [name, metricJson(entry)] as const),
    ),
    commands: Object.fromEntries(
      Object.entries(choice.commands).map(([name, entry]) => [name, commandJson(entry)] as const),
    ),
  };
}

export function serialiseChoices(choices: TwinChoices): string {
  const document = {
    schema_version: "0.1",
    id: choices.id,
    version: "0.1.0",
    title: choices.title,
    ...choices.extra,
    kinds: choices.kinds.map(kindJson),
  };
  return `${JSON.stringify(document, null, 2)}\n`;
}
