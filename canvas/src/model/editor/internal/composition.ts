import type { AreaNode } from "../../shared/areaTree";
import { areaTree } from "../../shared/areaTree";
import type { Instance, Link, Plant } from "../../plant";

type Json = Readonly<Record<string, unknown>>;
export type Composition = Json;

export interface Equipment {
  readonly index: number;
  readonly use: string;
  readonly as: string;
  readonly preset: string;
  readonly names: readonly string[];
  readonly area: string;
}

export interface Connection {
  readonly index: number;
  readonly from: string;
  readonly to: string;
}

export interface NewEquipment {
  readonly use: string;
  readonly preset: string;
  readonly name: string;
  readonly area: string;
}

function record(value: unknown): Json {
  return typeof value === "object" && value !== null && !Array.isArray(value) ? (value as Json) : {};
}

function list(value: unknown): readonly unknown[] {
  return Array.isArray(value) ? value : [];
}

function text(value: unknown): string {
  return typeof value === "string" ? value : "";
}

export function parseComposition(content: string): Composition | null {
  try {
    const value: unknown = JSON.parse(content);
    return typeof value === "object" && value !== null && !Array.isArray(value) ? (value as Json) : null;
  } catch {
    return null;
  }
}

export function equipmentOf(doc: Composition): readonly Equipment[] {
  return list(doc.equipment).map((raw, index) => {
    const entry = record(raw);
    const names = list(entry.names).map(text).filter(Boolean);
    return {
      index,
      use: text(entry.use),
      as: text(entry.as),
      preset: text(entry.preset),
      names: names.length > 0 ? names : [text(entry.id_template)].filter(Boolean),
      area: text(entry.area),
    };
  });
}

export function connectionsOf(doc: Composition): readonly Connection[] {
  return list(doc.connections).map((raw, index) => ({
    index,
    from: text(record(raw).from),
    to: text(record(raw).to),
  }));
}

export function areaNodesOf(doc: Composition): readonly AreaNode[] {
  return areaTree(
    list(doc.areas)
      .map((raw) => ({ name: text(record(raw).name), parent: text(record(raw).parent) || null }))
      .filter((area) => area.name !== ""),
  );
}

export function componentOf(use: string): string {
  const at = use.lastIndexOf("@");
  return at < 0 ? use : use.slice(0, at);
}

export function endpointOf(value: string): { readonly id: string; readonly port: string } {
  const dot = value.lastIndexOf(".");
  return dot < 0 ? { id: value, port: "" } : { id: value.slice(0, dot), port: value.slice(dot + 1) };
}

function link(connection: Connection): Link {
  const from = endpointOf(connection.from);
  const to = endpointOf(connection.to);
  return { from: from.id, to: to.id, type: "", fromPort: from.port || null, toPort: to.port || null };
}

export function shortKind(component: string): string {
  return (component.split("/").at(-1) ?? component).replaceAll("-", "_");
}

export function compositionPlant(doc: Composition, titleOf: (component: string) => string): Plant {
  const instances: Instance[] = equipmentOf(doc).flatMap((entry) =>
    entry.names.map((id) => ({
      id,
      kind: entry.as || titleOf(componentOf(entry.use)),
      area: entry.area,
      component: componentOf(entry.use),
      instruments: new Map<string, string>(),
    })),
  );
  return {
    id: text(doc.id),
    title: text(doc.title),
    site: text(record(doc.site).name),
    areas: list(doc.areas).map((raw) => ({
      name: text(record(raw).name),
      parent: text(record(raw).parent) || null,
    })),
    instances,
    links: connectionsOf(doc).map(link),
    kinds: new Map(),
  };
}

export interface LinkEnd {
  readonly outgoing: boolean;
  readonly port: string;
  readonly other: string;
  readonly otherPort: string;
}

export function linkEnd(connection: Connection, id: string): LinkEnd {
  const from = endpointOf(connection.from);
  const to = endpointOf(connection.to);
  const words = (port: string) => port.replaceAll("_", " ");
  return from.id === id
    ? { outgoing: true, port: words(from.port), other: to.id, otherPort: words(to.port) }
    : { outgoing: false, port: words(to.port), other: from.id, otherPort: words(from.port) };
}

function initials(title: string): string {
  const letters = title
    .split(/[\s-]+/)
    .filter(Boolean)
    .map((word) => word[0]?.toUpperCase() ?? "")
    .join("");
  return letters === "" ? "X" : letters.slice(0, 3);
}

function shortPrefix(name: string | undefined): string | null {
  const prefix = name?.match(/^(.*?)-?\d+$/)?.[1] ?? "";
  return /^[A-Za-z]{1,4}$/.test(prefix) ? prefix : null;
}

export function suggestedName(doc: Composition, use: string, title: string): string {
  const names = equipmentOf(doc).flatMap((entry) => entry.names);
  const sibling = equipmentOf(doc).find((entry) => componentOf(entry.use) === componentOf(use))?.names[0];
  const prefix = shortPrefix(sibling) ?? initials(title);
  let n = 1;
  while (names.includes(`${prefix}-${String(n)}`)) n += 1;
  return `${prefix}-${String(n)}`;
}

export function titleOf(doc: Composition): string {
  return text(doc.title) || text(doc.id);
}

export function domainOf(doc: Composition): string {
  return `domains/${text(doc.id)}.domain.json`;
}

export function serialise(doc: Composition): string {
  return `${JSON.stringify(doc, null, 2)}\n`;
}
