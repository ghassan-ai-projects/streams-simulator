import type {
  FaultCategories,
  FaultCategory,
  LaunchGateway,
  PerturbationKind,
  PerturbationParamSpec,
} from "../../../app/ports";
import { isRecord, text } from "../../../model/shared/json";
import { launchFailure } from "./failures";
import type { Transport } from "./http";
import { action } from "./http";

type CatalogGateway = Pick<LaunchGateway, "perturbationCatalog" | "faultCategories">;

const paramTypes: readonly PerturbationParamSpec["type"][] = ["number", "integer", "enum", "string_list"];
const categories: readonly FaultCategory[] = ["world", "measurement", "world_and_measurement"];

function optionalNumber(value: unknown): { readonly ok: boolean; readonly value?: number } {
  if (value === undefined) return { ok: true };
  return typeof value === "number" && Number.isFinite(value) ? { ok: true, value } : { ok: false };
}

function valuesOf(value: unknown): readonly string[] | null | undefined {
  if (value === undefined) return undefined;
  return Array.isArray(value) && value.every((item) => typeof item === "string") ? value : null;
}

function param(value: unknown): PerturbationParamSpec | null {
  if (!isRecord(value)) return null;
  const type = paramTypes.find((known) => known === value.type);
  const minimum = optionalNumber(value.minimum);
  const maximum = optionalNumber(value.maximum);
  const values = valuesOf(value.values);
  if (text(value.name) === "" || type === undefined || !minimum.ok || !maximum.ok || values === null) return null;
  return {
    name: text(value.name),
    type,
    ...(minimum.value === undefined ? {} : { minimum: minimum.value }),
    ...(maximum.value === undefined ? {} : { maximum: maximum.value }),
    ...(values === undefined ? {} : { values }),
  };
}

function params(value: unknown): readonly PerturbationParamSpec[] | null {
  if (value === undefined) return [];
  if (!Array.isArray(value)) return null;
  const decoded = value.map(param);
  return decoded.every((item) => item !== null) ? decoded : null;
}

function kind(value: unknown): PerturbationKind | null {
  if (!isRecord(value) || text(value.kind) === "") return null;
  const decoded = params(value.params);
  const requiresEnd = value.requires_end ?? false;
  if (decoded === null || typeof requiresEnd !== "boolean") return null;
  return {
    kind: text(value.kind),
    category: text(value.category),
    description: text(value.description),
    requiresEnd,
    params: decoded,
  };
}

export function perturbationCatalog(value: unknown): readonly PerturbationKind[] | null {
  if (!isRecord(value) || !Array.isArray(value.perturbations)) return null;
  const kinds = value.perturbations.map(kind);
  return kinds.every((item) => item !== null) ? kinds : null;
}

function categoryOf(value: unknown): FaultCategory | undefined {
  return categories.find((known) => known === value);
}

function ofKind(value: unknown): Readonly<Record<string, FaultCategory>> {
  const known = Object.entries(isRecord(value) ? value : {}).flatMap(([fault, raw]) => {
    const category = categoryOf(raw);
    return category === undefined ? [] : [[fault, category] as const];
  });
  return Object.fromEntries(known);
}

export function faultCategories(value: unknown): FaultCategories | null {
  if (!isRecord(value)) return null;
  const kinds = isRecord(value.fault_categories) ? value.fault_categories : {};
  return Object.fromEntries(Object.entries(kinds).map(([name, faults]) => [name, ofKind(faults)]));
}

export function catalogGateway(transport: Transport): CatalogGateway {
  return {
    perturbationCatalog: () =>
      action(transport, "catalog.perturbations", {}, perturbationCatalog, launchFailure("catalog")),
    faultCategories: (domain) =>
      action(transport, "catalog.describe", { domain }, faultCategories, launchFailure("catalog")),
  };
}
