import type { Acknowledgement, ComposeSummary, DomainDocument, FileContent } from "../../../app/ports";
import type { Workspace } from "../../../model/workspace";
import type { ExamplePlant, LibraryEntry, Presets } from "../../../model/plant";
import type { ComponentSpec, PortSpec } from "../../../model/plant";
import type { Replay, Report } from "../../../model/review";
import { parseReportPayload, parseReplayPayload } from "./reportDecoder";
import { isRecord, record, text } from "../../../model/shared/json";
export { journalEntry } from "./journalDecoder";

export function paired(value: unknown): boolean | null {
  return isRecord(value) && typeof value.paired === "boolean" ? value.paired : null;
}

export function fileContent(value: unknown): FileContent | null {
  if (
    !isRecord(value) ||
    typeof value.path !== "string" ||
    typeof value.content !== "string" ||
    typeof value.etag !== "string"
  ) {
    return null;
  }
  return { path: value.path, content: value.content, etag: value.etag };
}

export function written(value: unknown): { readonly path: string; readonly etag: string } | null {
  if (!isRecord(value) || typeof value.path !== "string" || typeof value.etag !== "string") return null;
  return { path: value.path, etag: value.etag };
}

export function workspace(value: unknown): Workspace | null {
  if (!isRecord(value) || typeof value.name !== "string") return null;
  const domains = definitions(value.domains);
  const compositions = definitions(value.compositions);
  const scenarios = definitions(value.scenarios);
  const runs = rows(value.runs, runEntry);
  if (domains === null || compositions === null || scenarios === null || runs === null) return null;
  return { name: value.name, domains, compositions, scenarios, runs };
}

function definitions(value: unknown): Workspace["domains"] | null {
  return rows(value, definition);
}

function rows<T>(value: unknown, decode: (entry: Record<string, unknown>) => T | null): readonly T[] | null {
  if (!Array.isArray(value)) return null;
  const decoded: T[] = [];
  for (const raw of value) {
    const row = isRecord(raw) ? decode(raw) : null;
    if (row === null) return null;
    decoded.push(row);
  }
  return decoded;
}

function definition(entry: Record<string, unknown>): Workspace["domains"][number] | null {
  if (text(entry.path) === "") return null;
  return {
    path: text(entry.path),
    id: text(entry.id),
    title: text(entry.title),
    instances: typeof entry.instances === "number" ? entry.instances : 0,
  };
}

function runEntry(entry: Record<string, unknown>): Workspace["runs"][number] | null {
  if (text(entry.world_id) === "") return null;
  return {
    worldId: text(entry.world_id),
    ended: entry.ended === true,
    complete: entry.complete === true,
    incompleteReason: text(entry.incomplete_reason),
    explore: entry.explore === true,
    seed: typeof entry.seed === "number" ? entry.seed : null,
    plant: text(entry.plant),
    durationS: typeof entry.duration_ns === "number" ? Math.round(entry.duration_ns / 1e9) : 0,
    recordedAt: text(entry.recorded_at),
    faults: typeof entry.faults === "number" ? entry.faults : 0,
  };
}

export function library(value: unknown): readonly LibraryEntry[] | null {
  if (!isRecord(value) || !Array.isArray(value.components)) return null;
  return value.components
    .map(record)
    .filter((entry) => entry.kind === "equipment" && text(entry.id) !== "")
    .map((entry) => ({
      component: text(entry.id),
      ref: `${text(entry.id)}@${typeof entry.version === "number" ? String(entry.version) : "1"}`,
      title: text(entry.title) || text(entry.id),
      words: [text(entry.title), text(entry.id), text(entry.class), ...strings(entry.tags)].join(" ").toLowerCase(),
    }))
    .sort((a, b) => a.title.localeCompare(b.title));
}

export function presets(value: unknown): Presets | null {
  if (!isRecord(value) || !isRecord(value.document)) return null;
  const document = value.document;
  const rawPresets = record(document.presets);
  const options = Object.entries(rawPresets).map(([name, raw]) => ({
    name,
    title: text(record(raw).title) || name,
  }));
  return { options, preferred: text(document.default_preset), description: text(document.description) };
}

export function examples(value: unknown): readonly ExamplePlant[] | null {
  if (!isRecord(value) || !Array.isArray(value.examples)) return null;
  return value.examples
    .map(record)
    .filter((entry) => text(entry.id) !== "")
    .map((entry) => ({
      id: text(entry.id),
      title: text(entry.title) || text(entry.id),
      description: text(entry.description),
    }));
}

export function component(value: unknown): ComponentSpec | null {
  if (!isRecord(value) || !isRecord(value.entry) || !isRecord(value.document)) return null;
  const entry = value.entry;
  const document = value.document;
  const tags = strings(entry.tags);
  const title = nonblank(entry.title) ?? nonblank(document.title);
  const description = nonblank(document.description);
  return {
    ports: parsePorts(document),
    category: category(text(entry.class), tags),
    ...(title === undefined ? {} : { title }),
    ...(description === undefined ? {} : { description }),
  };
}

function parsePorts(document: Record<string, unknown>): readonly PortSpec[] {
  if (!Array.isArray(document.ports)) return [];
  return document.ports.map((raw) => {
    const port = record(raw);
    const channel = nonblank(port.state);
    return {
      name: text(port.name),
      type: text(port.type),
      direction: port.direction === "provides" ? "out" : "in",
      ...(channel === undefined ? {} : { channel }),
    };
  });
}

function category(componentClass: string, tags: readonly string[]): ComponentSpec["category"] {
  const byClass = [
    ["electrical/", "elec"],
    ["process/", "prod"],
    ["environment/", "env"],
  ] as const;
  const classCategory = byClass.find(([prefix]) => componentClass.startsWith(prefix));
  if (classCategory !== undefined) return classCategory[1];
  const byTag = [
    ["compressed-air", "air"],
    ["hvac", "cool"],
    ["fluids", "cool"],
    ["heat", "cool"],
    ["drives", "prod"],
    ["material-handling", "prod"],
  ] as const;
  return byTag.find(([tag]) => tags.includes(tag))?.[1] ?? "env";
}

export function composition(value: unknown): Record<string, unknown> | null {
  if (!isRecord(value) || !isRecord(value.composition)) return null;
  const doc = value.composition;
  if (doc.schema_version !== "0.1" || !["id", "version", "title", "stresses"].every((key) => text(doc[key]))) {
    return null;
  }
  if (!isRecord(doc.axes) || !Array.isArray(doc.profiles) || !isRecord(doc.ground_truth)) return null;
  if (doc.equipment !== undefined && !Array.isArray(doc.equipment)) return null;
  return value.composition;
}

export function composeSummary(value: unknown): ComposeSummary | null {
  if (!isRecord(value)) return null;
  const channels = finiteNumber(value.channels);
  const faults = finiteNumber(value.faults);
  const effectors = finiteNumber(value.effectors);
  return channels === null || faults === null || effectors === null ? null : { channels, faults, effectors };
}

function finiteNumber(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

export function acknowledgement(value: unknown): Acknowledgement | null {
  if (!isRecord(value) || typeof value.accepted !== "boolean") return null;
  if (value.reason !== undefined && typeof value.reason !== "string") return null;
  if (value.effect_eta_ns !== undefined && typeof value.effect_eta_ns !== "number") return null;
  return {
    accepted: value.accepted,
    reason: text(value.reason),
    effectEtaNs: typeof value.effect_eta_ns === "number" ? value.effect_eta_ns : null,
  };
}

export function advanceResult(value: unknown): true | null {
  return isRecord(value) && typeof value.clock === "string" && finiteNumber(value.emitted) !== null ? true : null;
}

export function faultInjected(value: unknown): true | null {
  return isRecord(value) && text(value.fault_id) !== "" && value.simulated === true ? true : null;
}

export function runEnded(value: unknown): true | null {
  return isRecord(value) && text(value.run_artifact_path) !== "" ? true : null;
}

export function destroyed(value: unknown): true | null {
  return isRecord(value) && value.destroyed === true ? true : null;
}

export function report(value: unknown): Report | null {
  return parseReportPayload(value);
}

export function replay(value: unknown): Replay | null {
  return parseReplayPayload(value);
}

export function runDomain(value: unknown): DomainDocument | null {
  if (!isRecord(value) || typeof value.content !== "string") return null;
  try {
    const artifact: unknown = JSON.parse(value.content);
    if (!isRecord(artifact) || !isRecord(artifact.domain_spec)) return null;
    return artifact.domain_spec as DomainDocument;
  } catch {
    return null;
  }
}

function strings(value: unknown): readonly string[] {
  return Array.isArray(value) ? value.map(text) : [];
}

function nonblank(value: unknown): string | undefined {
  const content = text(value);
  return content.trim() === "" ? undefined : content;
}
