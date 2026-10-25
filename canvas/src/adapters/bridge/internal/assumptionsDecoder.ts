import type {
  AssumedCommand,
  AssumedDrive,
  AssumedDynamics,
  AssumedFault,
  AssumedKind,
  AssumedReading,
  Assumptions,
  NotModeled,
} from "../../../model/workspace";
import { isRecord, list, number, record, text } from "../../../model/shared/json";
import { choice, requiredTexts, rows } from "./decodeKit";

function numberOrNull(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

type Range = NonNullable<AssumedDynamics["clamp"]>;

function rangeOf(value: unknown): Range | null {
  const [min, max] = [numberOrNull(record(value).min), numberOrNull(record(value).max)];
  return min === null || max === null ? null : { min, max };
}

function readingOf(row: Record<string, unknown>): AssumedReading | null {
  const channel = text(row.channel);
  if (channel === "") return null;
  const [noise, cadence] = [record(row.noise), record(row.cadence)];
  return {
    channel,
    unit: text(row.unit),
    valueType: text(row.value_type),
    range: rangeOf(row.range),
    resolution: number(row.resolution),
    noise: { model: text(noise.model), sigma: number(noise.sigma) },
    cadence: { mode: text(cadence.mode), periodS: numberOrNull(cadence.period_s) },
    instrument: text(record(row.instrument).ref),
    basis: text(row.basis),
    recordedIn: text(row.recorded_in),
    suggestion: typeof row.equals_binary_suggestion === "boolean" ? row.equals_binary_suggestion : null,
  };
}

function driveOf(row: Record<string, unknown>): AssumedDrive {
  return {
    from: text(row.from),
    kind: text(row.kind),
    metric: text(row.metric),
    gain: number(row.gain),
    timeConstantS: number(row.time_constant_s),
  };
}

function commandOf(row: Record<string, unknown>): AssumedCommand | null {
  const name = text(row.name);
  if (name === "") return null;
  const effect = record(row.effect);
  return {
    name,
    riskClass: text(row.risk_class),
    safeStop: row.safe_stop === true,
    response: text(row.response),
    recordedIn: text(row.recorded_in),
    drives: list(row.drives).map(record).map(driveOf),
    deadTimeS: number(effect.dead_time_s),
    timeConstantS: number(effect.time_constant_s),
    statement: text(row.statement),
    suggestion: typeof row.equals_binary_suggestion === "boolean" ? row.equals_binary_suggestion : null,
  };
}

function dynamicsOf(row: Record<string, unknown>): AssumedDynamics | null {
  const target = text(row.target);
  if (target === "") return null;
  return {
    target,
    form: text(row.form),
    gain: numberOrNull(row.gain),
    timeConstantS: numberOrNull(row.time_constant_s),
    clamp: rangeOf(row.clamp),
  };
}

function faultOf(row: Record<string, unknown>): AssumedFault | null {
  const id = text(row.id);
  if (id === "") return null;
  return {
    id,
    description: text(row.description),
    onsetShape: text(row.onset_shape),
    affects: list(row.affects).map(text),
  };
}

function instancesOf(row: Record<string, unknown>): string {
  const names = list(row.names).map(text);
  const where = text(row.area) === "" ? "" : ` in ${text(row.area)}`;
  return `${names.length > 0 ? names.join(", ") : text(row.id_template)}${where}`;
}

function kindOf(row: Record<string, unknown>): AssumedKind | null {
  const kind = text(row.kind);
  const parts = [
    rows(row.readings, readingOf),
    rows(row.dynamics, dynamicsOf),
    rows(row.commands, commandOf),
    rows(row.faults, faultOf),
    rows(row.instances, (each) => ({ words: instancesOf(each) })),
  ] as const;
  const [readings, dynamics, commands, faults, instances] = parts;
  if (kind === "" || readings === null || dynamics === null || commands === null || faults === null) return null;
  if (instances === null) return null;
  return {
    kind,
    implements: text(row.implements),
    instances: instances.map((entry) => entry.words),
    readings,
    dynamics,
    commands,
    faults,
  };
}

function notModeledOf(row: Record<string, unknown>): NotModeled | null {
  const given = requiredTexts(row, ["code", "sentence"]);
  return given === null ? null : { code: given.code, scope: text(row.scope), sentence: given.sentence };
}

const origins = ["twin", "composition", "domain"] as const;

function originOf(value: unknown): Assumptions["origin"] | null {
  if (!isRecord(value) || !choice(value.kind, origins)) return null;
  const twin = isRecord(value.twin) ? requiredTexts(value.twin, ["id", "version", "status"]) : null;
  return { kind: value.kind, twin };
}

export function assumptions(value: unknown): Assumptions | null {
  if (!isRecord(value) || value.schema_version !== "0.1" || !isRecord(value.domain)) return null;
  const domain = requiredTexts(value.domain, ["id", "version", "digest"]);
  const [origin, kinds, notModeled] = [
    originOf(value.origin),
    rows(value.kinds, kindOf),
    rows(value.not_modeled, notModeledOf),
  ];
  return domain === null || origin === null || kinds === null || notModeled === null
    ? null
    : { domain, origin, kinds, notModeled };
}
