import type { FileContent } from "../app/ports";
import { failure as failed, success } from "../app/ports";
import { cannedFailure } from "./fakeFailures";
import type { Canned } from "./fakeFailures";

type Json = Record<string, unknown>;

const frame = { schema_version: "0.1", version: "1", stresses: "none", axes: {}, profiles: [], ground_truth: {} };
const bodyKeys = ["description", "equipment", "connections", "areas"];

export function newPlant(
  actions: ReadonlyMap<string, Canned<unknown>>,
  files: ReadonlyMap<string, Canned<FileContent>>,
  args: Readonly<Json>,
): Canned<unknown> {
  const base = typeof args.from === "string" ? fromFile(files, args.from) : fromExample(actions);
  if (!base.ok) return base;
  if (!isJson(base.value)) return failed(cannedFailure("composition_invalid", 422, "the base is not a composition"));
  const plant: Json = { ...frame, ...base.value, id: args.id, title: args.title };
  if (args.empty !== true) return success({ composition: plant });
  const kept = Object.fromEntries(Object.entries(plant).filter(([key]) => !bodyKeys.includes(key)));
  return success({ composition: { ...kept, areas: [{ name: siteName(plant) }] } });
}

function fromExample(actions: ReadonlyMap<string, Canned<unknown>>): Canned<unknown> {
  const shown = actions.get("component.example") ?? failed(cannedFailure("component_failed", 422, "no example"));
  return shown.ok ? success((shown.value as { example?: unknown }).example) : shown;
}

function fromFile(files: ReadonlyMap<string, Canned<FileContent>>, path: string): Canned<unknown> {
  const file = files.get(path) ?? failed(cannedFailure("not_found", 404, `no ${path}`));
  return file.ok ? success(parsed(file.value.content)) : file;
}

function parsed(content: string): unknown {
  try {
    return JSON.parse(content) as unknown;
  } catch {
    return null;
  }
}

function isJson(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function siteName(plant: Json): string {
  const name = isJson(plant.site) ? plant.site.name : undefined;
  return typeof name === "string" && name !== "" ? name : "Site";
}
