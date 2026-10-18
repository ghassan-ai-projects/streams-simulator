import { isRecord, list } from "../../shared/json";
import type { ArgRow, Arguments, Fields } from "./types";

const argTypes: readonly string[] = ["number", "integer", "boolean", "string"];
const schemaKeys: readonly string[] = ["type", "properties", "required", "additionalProperties"];
const propertyKeys: readonly string[] = ["type", "minimum", "maximum"];

function bound(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

function rowOf(name: string, spec: unknown, required: readonly unknown[]): ArgRow | null {
  if (!isRecord(spec) || !Object.keys(spec).every((key) => propertyKeys.includes(key))) return null;
  if (typeof spec.type !== "string" || !argTypes.includes(spec.type)) return null;
  const [min, max] = [bound(spec.minimum), bound(spec.maximum)];
  return { name, type: spec.type, min, max, required: required.includes(name) };
}

function simple(schema: Fields): boolean {
  const required = list(schema.required);
  const names = isRecord(schema.properties) ? Object.keys(schema.properties) : [];
  return (
    schema.type === "object" &&
    (schema.additionalProperties === undefined || schema.additionalProperties === false) &&
    Object.keys(schema).every((key) => schemaKeys.includes(key)) &&
    required.every((entry) => typeof entry === "string" && names.includes(entry))
  );
}

export function argumentsOf(schema: Fields): Arguments {
  if (Object.keys(schema).length === 0) return { kind: "rows", rows: [] };
  if (!simple(schema)) return { kind: "custom", schema };
  const required = list(schema.required);
  const rows: ArgRow[] = [];
  for (const [name, spec] of Object.entries(isRecord(schema.properties) ? schema.properties : {})) {
    const row = rowOf(name, spec, required);
    if (row === null) return { kind: "custom", schema };
    rows.push(row);
  }
  return { kind: "rows", rows, ...(schema.additionalProperties === false ? { closed: true as const } : {}) };
}

function propertyOf(row: ArgRow): Fields {
  return {
    type: row.type,
    ...(row.min === null ? {} : { minimum: row.min }),
    ...(row.max === null ? {} : { maximum: row.max }),
  };
}

export function schemaOf(args: Arguments): Fields {
  if (args.kind === "custom") return args.schema;
  const required = args.rows.filter((row) => row.required).map((row) => row.name);
  return {
    type: "object",
    properties: Object.fromEntries(args.rows.map((row) => [row.name, propertyOf(row)])),
    ...(required.length === 0 ? {} : { required }),
    ...(args.closed === true ? { additionalProperties: false } : {}),
  };
}

export const blankArguments: Arguments = { kind: "rows", rows: [] };

export const blankRow: ArgRow = { name: "", type: "number", min: null, max: null, required: false };
