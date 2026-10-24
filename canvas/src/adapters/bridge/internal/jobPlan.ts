import type { JobStatus } from "../../../app/ports";
import { isRecord, text } from "../../../model/shared/json";

type Fault = JobStatus["faults"][number];
type PlannedCall = JobStatus["plannedCalls"][number];
type Perturbation = JobStatus["perturbations"][number];
type Params = Perturbation["params"];
type ParamValue = Params[string];

function finite(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

function untilOf(value: unknown): { readonly untilS?: number } | null {
  if (value === undefined) return {};
  const until = finite(value);
  return until === null ? null : { untilS: until };
}

function clearedOf(value: unknown): { readonly clearedAtS?: number } | null {
  if (value === undefined) return {};
  const cleared = finite(value);
  return cleared === null ? null : { clearedAtS: cleared };
}

function liveOf(value: Record<string, unknown>): { readonly immediate?: true; readonly faultId?: string } | null {
  if (value.immediate !== undefined && typeof value.immediate !== "boolean") return null;
  const faultId = text(value.fault_id);
  return { ...(value.immediate === true ? { immediate: true as const } : {}), ...(faultId === "" ? {} : { faultId }) };
}

function fault(value: unknown): Fault | null {
  if (!isRecord(value)) return null;
  const atS = finite(value.at_s);
  const until = untilOf(value.until_s);
  const cleared = clearedOf(value.cleared_at_s);
  const live = liveOf(value);
  const entity = text(value.entity_id);
  const name = text(value.fault);
  if (atS === null || until === null || cleared === null || live === null || entity === "" || name === "") return null;
  return { entity, fault: name, atS, ...until, ...cleared, ...live };
}

function plannedCall(value: unknown): PlannedCall | null {
  if (!isRecord(value)) return null;
  const atS = finite(value.at_s);
  const entity = text(value.entity_id);
  const effector = text(value.effector);
  const args = argsOf(value.args);
  if (atS === null || entity === "" || effector === "" || args === null) return null;
  return { entity, effector, atS, args };
}

function argsOf(value: unknown): PlannedCall["args"] | null {
  if (value === undefined) return {};
  if (!isRecord(value)) return null;
  const entries = Object.entries(value);
  if (!entries.every(([, item]) => typeof item === "number" || typeof item === "string" || typeof item === "boolean"))
    return null;
  return value as PlannedCall["args"];
}

function paramValue(value: unknown): ParamValue | null {
  if (typeof value === "number" || typeof value === "string") return value;
  if (Array.isArray(value) && value.every((item) => typeof item === "string")) return value;
  return null;
}

function params(value: unknown): Params | null {
  if (value === undefined) return {};
  if (!isRecord(value)) return null;
  const decoded = Object.entries(value).map(([name, raw]) => [name, paramValue(raw)] as const);
  if (decoded.some(([, item]) => item === null)) return null;
  return Object.fromEntries(decoded) as Params;
}

function perturbation(value: unknown): Perturbation | null {
  if (!isRecord(value)) return null;
  const atS = finite(value.at_s);
  const until = untilOf(value.until_s);
  const decoded = params(value.params);
  const kind = text(value.kind);
  if (atS === null || until === null || decoded === null || kind === "") return null;
  return { kind, params: decoded, atS, ...until };
}

function listOf<T>(value: unknown, item: (entry: unknown) => T | null): readonly T[] | null {
  if (value === undefined) return [];
  if (!Array.isArray(value)) return null;
  const items = value.map(item);
  return items.every((entry) => entry !== null) ? items : null;
}

export function jobFaults(value: unknown): readonly Fault[] | null {
  return listOf(value, fault);
}

export function jobPlannedCalls(value: unknown): readonly PlannedCall[] | null {
  return listOf(value, plannedCall);
}

export function jobPerturbations(value: unknown): readonly Perturbation[] | null {
  return listOf(value, perturbation);
}

export function faultArgs(planned: Fault): Readonly<Record<string, unknown>> {
  const { entity, fault: name, atS, untilS } = planned;
  return { entity_id: entity, fault: name, at_s: atS, ...(untilS === undefined ? {} : { until_s: untilS }) };
}

export function perturbationArgs(planned: Perturbation): Readonly<Record<string, unknown>> {
  const { kind, params: given, atS, untilS } = planned;
  return {
    kind,
    ...(Object.keys(given).length === 0 ? {} : { params: given }),
    at_s: atS,
    ...(untilS === undefined ? {} : { until_s: untilS }),
  };
}
