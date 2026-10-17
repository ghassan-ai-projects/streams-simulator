import type { Acknowledgement, ComposeSummary, DomainDocument } from "../app/ports";
import type { Composition } from "../model/editor";
import type { ExamplePlant, LibraryEntry, Presets } from "../model/plant";
import type { ComponentSpec, PortSpec } from "../model/plant";
import type { Replay, Report } from "../model/review";

type Json = Record<string, unknown>;

function isRecord(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function record(value: unknown): Json {
  return isRecord(value) ? value : {};
}

function text(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function list(value: unknown): readonly unknown[] {
  return Array.isArray(value) ? value : [];
}

function documentOf(value: Json): Json {
  return isRecord(value.document) ? value.document : value;
}

const categories = [
  ["electrical/", "elec"],
  ["process/", "prod"],
  ["environment/", "env"],
  ["compressed-air", "air"],
  ["hvac", "cool"],
  ["fluids", "cool"],
  ["heat", "cool"],
  ["drives", "prod"],
  ["material-handling", "prod"],
] as const;

export function specOf(value: unknown): ComponentSpec | null {
  if (!isRecord(value)) return null;
  const document = documentOf(value);
  const words = [text(record(value.entry).class), ...list(record(value.entry).tags).map(text)];
  const category = categories.find(([key]) => words.some((word) => word.startsWith(key)))?.[1] ?? "env";
  return { ports: list(document.ports).map(portOf), category };
}

function portOf(value: unknown): PortSpec {
  const port = record(value);
  const channel = text(port.state);
  return {
    name: text(port.name),
    type: text(port.type),
    direction: port.direction === "provides" || port.direction === "out" ? "out" : "in",
    ...(channel === "" ? {} : { channel }),
  };
}

export function presetsOf(value: unknown): Presets {
  const document = isRecord(value) ? documentOf(value) : {};
  const options = Object.entries(record(document.presets)).map(([name, preset]) => ({
    name,
    title: text(record(preset).title) || name,
  }));
  return { options, preferred: text(document.default_preset), description: text(document.description) };
}

export function libraryOf(value: unknown): readonly LibraryEntry[] {
  const entries = isRecord(value) ? list(value.components) : [];
  return entries.flatMap((raw) => {
    const entry = record(raw);
    return entry.kind === "equipment" && text(entry.id) !== "" ? [libraryEntry(entry)] : [];
  });
}

function libraryEntry(entry: Json): LibraryEntry {
  const component = text(entry.id);
  const title = text(entry.title) || component;
  const version = typeof entry.version === "number" ? String(entry.version) : "1";
  return {
    component,
    ref: `${component}@${version}`,
    title,
    words: [title, component, text(entry.class), ...list(entry.tags).map(text)].join(" ").toLowerCase(),
  };
}

export function examplesOf(value: unknown): readonly ExamplePlant[] {
  return (isRecord(value) ? list(value.examples) : []).flatMap((raw) => {
    const entry = record(raw);
    const id = text(entry.id);
    return id === "" ? [] : [{ id, title: text(entry.title) || id, description: text(entry.description) }];
  });
}

export function compositionOf(value: unknown): Composition | null {
  const composition = isRecord(value) ? value.composition : null;
  return isRecord(composition) ? composition : null;
}

function count(value: unknown): number {
  return typeof value === "number" ? value : 0;
}

function etaOf(value: unknown): number | null {
  return typeof value === "number" ? value : null;
}

export function summaryOf(value: unknown): ComposeSummary {
  const summary = record(value);
  return { channels: count(summary.channels), faults: count(summary.faults), effectors: count(summary.effectors) };
}

export function worldOf(value: unknown): { readonly worldId: string } {
  return { worldId: text(record(value).world_id) };
}

export function acknowledgementOf(value: unknown): Acknowledgement {
  const ack = record(value);
  const eta = ack.effect_eta_ns === undefined ? ack.effectEtaNs : ack.effect_eta_ns;
  return { accepted: ack.accepted === true, reason: text(ack.reason), effectEtaNs: etaOf(eta) };
}

export function isReport(value: unknown): value is Report {
  return isRecord(value) && typeof value.runId === "string" && typeof value.counts === "object";
}

function isReplay(value: unknown): value is Replay {
  return isRecord(value) && typeof value.matches === "boolean";
}

function flag(value: unknown): boolean {
  return value === true;
}

export function replayOf(value: unknown): Replay | null {
  if (!isReplay(value)) return null;
  return {
    matches: value.matches,
    domainChanged: flag(value.domainChanged),
    detail: text(value.detail),
    reason: text(value.reason),
    file: text(value.file),
  };
}

export function domainDocument(value: object): DomainDocument {
  return value as DomainDocument;
}

export function runDomainOf(content: string): DomainDocument | null {
  try {
    const artifact: unknown = JSON.parse(content);
    const domain = isRecord(artifact) ? artifact.domain_spec : null;
    return isRecord(domain) ? domainDocument(domain) : null;
  } catch {
    return null;
  }
}
