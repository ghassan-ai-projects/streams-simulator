import type { FileContent } from "../app/ports";
import { failure as failed, success } from "../app/ports";
import { failure } from "./fakeFailures";
import type { Canned } from "./fakeFailures";

type Actions = ReadonlyMap<string, Canned<unknown>>;
type Files = ReadonlyMap<string, Canned<FileContent>>;

/** `component.new` with no canned answer: the double composes the plant from its base, as the binary does. */
export function newPlant(actions: Actions, files: Files, args: Readonly<Record<string, unknown>>): Canned<unknown> {
  const base = basePlant(actions, files, args);
  if (!base.ok) return base;
  if (typeof base.value !== "object" || base.value === null || Array.isArray(base.value)) {
    return failed(failure("composition_invalid", 422, "the base is not a composition"));
  }
  const plant: Record<string, unknown> = { ...base.value, id: args.id, title: args.title };
  if (args.empty !== true) return success({ composition: plant });
  const frame = Object.fromEntries(
    Object.entries(plant).filter(([key]) => !["description", "equipment", "connections", "areas"].includes(key)),
  );
  return success({ composition: { ...frame, areas: [{ name: siteArea(plant.site) }] } });
}

function basePlant(actions: Actions, files: Files, args: Readonly<Record<string, unknown>>): Canned<unknown> {
  if (typeof args.from !== "string") {
    const shown = actions.get("component.example") ?? failed(failure("component_failed", 422, "no example"));
    return shown.ok ? success(record(shown.value).example) : shown;
  }
  const file = files.get(args.from) ?? failed(failure("not_found", 404, `no ${args.from}`));
  return file.ok ? success(parsed(file.value.content)) : file;
}

function parsed(content: string): unknown {
  try {
    return JSON.parse(content) as unknown;
  } catch {
    return null;
  }
}

function record(value: unknown): Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value) ? (value as Record<string, unknown>) : {};
}

function siteArea(site: unknown): string {
  if (typeof site !== "object" || site === null || !("name" in site)) return "Site";
  return typeof site.name === "string" ? site.name : "Site";
}
