import type { Layout, LayoutNode } from "./layout";
import { layout } from "./layout";
import { headerReserve } from "./frameTitle";
import { areaTree } from "./areaTree";
import type { Instance, Link, Plant } from "./plant";
import { areaOf } from "./plant";
import type { Category, ComponentSpec, Family, Pin, Pins } from "./ports";
import { plainRows } from "./nodePresentation";
import type { ViewPresentation } from "./nodePresentation";
import { familyOf, pinsOf } from "./ports";
import { kindLabel, shortChannel } from "./shared/words";

export interface ViewNode {
  readonly id: string;
  readonly label: string;
  readonly detail: string;
  readonly group: string;
  readonly instance: Instance | null;
  readonly category: Category;
  readonly members: number;
  readonly pins: Pins;
  readonly channels: readonly string[];
  readonly noteLines: number;
}

interface ViewEdge {
  readonly id: string;
  readonly from: string;
  readonly to: string;
  readonly fromHandle: string | null;
  readonly toHandle: string | null;
  readonly family: Family | null;
  readonly count: number;
}

export interface PlantView {
  readonly nodes: readonly ViewNode[];
  readonly edges: readonly ViewEdge[];
  readonly layout: Layout;
  readonly halls: readonly string[];
}

export const nodeSize = {
  width: 232,
  header: 40,
  kind: 24,
  pin: 24,
  padding: 16,
  groupWidth: 176,
  groupHeight: 72,
  idChars: 15,
  idLine: 36,
  note: 24,
} as const;

export const plainView: ViewPresentation = { rows: plainRows, note: () => 0 };

export function idExtra(label: string): number {
  return (Math.ceil(label.length / nodeSize.idChars) - 1) * nodeSize.idLine;
}

export function noteHeight(lines: number): number {
  return lines * nodeSize.note;
}

export function headerHeight(label: string): number {
  return nodeSize.header + nodeSize.kind + idExtra(label);
}

export function groupCounts(view: PlantView): Map<string, number> {
  const counts = new Map<string, number>();
  view.nodes.forEach((node) => counts.set(node.group, (counts.get(node.group) ?? 0) + 1));
  return counts;
}

export function hallOf(id: string): string | null {
  const slash = id.indexOf("/");
  return slash > 0 ? id.slice(0, slash) : null;
}

export function hallsOf(plant: Plant): readonly string[] {
  const halls = new Set(
    plant.instances.map((instance) => hallOf(instance.id)).filter((hall): hall is string => hall !== null),
  );
  return halls.size > 1 ? [...halls].sort() : [];
}

type Specs = ReadonlyMap<string, ComponentSpec>;
type Notes = ReadonlyMap<string, number>;

function instanceNode(plant: Plant, instance: Instance, specs: Specs, notes: Notes): ViewNode {
  const spec = instance.component === null ? undefined : specs.get(instance.component);
  return {
    id: instance.id,
    label: instance.id,
    detail: kindLabel(instance.kind),
    group: areaOf(plant, instance),
    instance,
    category: spec?.category ?? "env",
    members: 1,
    pins: pinsOf(plant, instance, spec?.ports),
    channels: (plant.kinds.get(instance.kind)?.channels ?? []).map((channel) => shortChannel(channel.name)),
    noteLines: notes.get(instance.id) ?? 1,
  };
}

function hallNode(hall: string, members: number): ViewNode {
  return {
    id: hall,
    label: hall,
    detail: `${String(members)} instances`,
    group: "Halls",
    instance: null,
    category: "env",
    members,
    pins: { ins: [], outs: [] },
    channels: [],
    noteLines: 1,
  };
}

function linkFamily(plant: Plant, link: Link, specs: Specs): Family | null {
  const source = plant.instances.find((instance) => instance.id === link.from);
  const spec = source?.component
    ? specs.get(source.component)?.ports.find((port) => port.name === link.fromPort)
    : undefined;
  return spec ? familyOf(spec.type) : null;
}

function pinNamed(pins: readonly Pin[], port: string | null): string | null {
  if (port !== null && pins.some((pin) => pin.name === port)) return port;
  return (pins.find((pin) => pin.connection === "unknown") ?? pins[0])?.name ?? null;
}

function handleOf(pins: ReadonlyMap<string, Pins>, id: string, side: "ins" | "outs", port: string | null) {
  return pinNamed(pins.get(id)?.[side] ?? [], port);
}

function bundle(
  plant: Plant,
  links: readonly Link[],
  key: (id: string) => string | null,
  specs: Specs,
  detailed: ReadonlyMap<string, Pins> | null,
): ViewEdge[] {
  const edges = new Map<string, ViewEdge>();
  for (const link of links) {
    const from = key(link.from);
    const to = key(link.to);
    if (from === null || to === null || from === to) continue;
    const fromHandle = detailed === null ? null : handleOf(detailed, from, "outs", link.fromPort);
    const toHandle = detailed === null ? null : handleOf(detailed, to, "ins", link.toPort);
    const id = `${from}.${fromHandle ?? ""}>${to}.${toHandle ?? ""}`;
    const known = edges.get(id);
    edges.set(
      id,
      known
        ? { ...known, count: known.count + 1 }
        : { id, from, to, fromHandle, toHandle, family: linkFamily(plant, link, specs), count: 1 },
    );
  }
  return [...edges.values()];
}

interface Context {
  readonly plant: Plant;
  readonly specs: Specs;
  readonly halls: readonly string[];
  readonly view: ViewPresentation;
  readonly aspect: number | undefined;
  readonly notes: Notes;
}

function nodeBox(node: ViewNode, view: ViewPresentation): LayoutNode {
  if (node.instance === null) {
    return { id: node.id, group: node.group, width: nodeSize.groupWidth, height: nodeSize.groupHeight };
  }
  const rows = view.rows(node.pins, node.channels);
  const note = view.note(node.noteLines);
  const height = headerHeight(node.label) + nodeSize.padding + note + rows * nodeSize.pin;
  return { id: node.id, group: node.group, width: nodeSize.width, height };
}

function finish(context: Context, nodes: readonly ViewNode[], edges: readonly ViewEdge[]): PlantView {
  const order = ["Halls", ...context.plant.areas.map((area) => area.name)];
  const boxes = nodes.map((node) => nodeBox(node, context.view));
  const rows = (group: string, count: number) => (group === "Halls" ? Math.ceil(Math.sqrt(count)) : 3);
  return {
    nodes,
    edges,
    layout: layout(boxes, edges, order, rows, context.aspect, headerReserve),
    halls: context.halls,
  };
}

function hallLevel(context: Context): PlantView {
  const { plant, specs, halls } = context;
  const counts = new Map<string, number>();
  plant.instances.forEach((instance) => {
    const hall = hallOf(instance.id);
    if (hall !== null) counts.set(hall, (counts.get(hall) ?? 0) + 1);
  });
  const shared = plant.instances.filter((instance) => hallOf(instance.id) === null);
  const nodes = [
    ...shared.map((instance) => instanceNode(plant, instance, specs, context.notes)),
    ...halls.map((hall) => hallNode(hall, counts.get(hall) ?? 0)),
  ];
  const key = (id: string) => hallOf(id) ?? id;
  const edges = bundle(plant, plant.links, key, specs, null);
  const feedsAll = new Set(
    shared
      .map((instance) => instance.id)
      .filter((id) => halls.every((hall) => edges.some((edge) => edge.from === id && edge.to === hall))),
  );
  const marked = nodes.map((node) =>
    feedsAll.has(node.id) ? { ...node, detail: `${node.detail} · feeds all ${String(halls.length)} halls` } : node,
  );
  return finish(
    context,
    marked,
    edges.filter((edge) => !(feedsAll.has(edge.from) && halls.includes(edge.to))),
  );
}

function instanceLevel(context: Context, keep: (instance: Instance) => boolean): PlantView {
  const kept = context.plant.instances.filter(keep);
  const ids = new Set(kept.map((instance) => instance.id));
  const nodes = kept.map((instance) => instanceNode(context.plant, instance, context.specs, context.notes));
  const key = (id: string) => (ids.has(id) ? id : null);
  const pins = new Map(nodes.map((node) => [node.id, node.pins] as const));
  return finish(context, nodes, bundle(context.plant, context.plant.links, key, context.specs, pins));
}

function withinHall(view: PlantView, hall: string): PlantView {
  const prefix = `${hall}/`;
  return {
    ...view,
    nodes: view.nodes.map((node) =>
      node.label.startsWith(prefix) ? { ...node, label: node.label.slice(prefix.length) } : node,
    ),
  };
}

export function plantView(
  plant: Plant,
  specs: Specs,
  open: string | null,
  view: ViewPresentation = plainView,
  aspect?: number,
  notes: Notes = new Map(),
): PlantView {
  const context = { plant, specs, halls: hallsOf(plant), view, aspect, notes };
  if (context.halls.length === 0) {
    return instanceLevel(context, () => true);
  }
  if (open === null || !context.halls.includes(open)) {
    return hallLevel(context);
  }
  const touching = new Set(
    plant.links
      .filter((link) => hallOf(link.to) === open || hallOf(link.from) === open)
      .flatMap((link) => [link.from, link.to]),
  );
  const keep = (instance: Instance) =>
    hallOf(instance.id) === open || (hallOf(instance.id) === null && touching.has(instance.id));
  return withinHall(instanceLevel(context, keep), open);
}

export function hallMembers(plant: Plant, hall: string): readonly (readonly [string, number])[] {
  const counts = new Map<string, number>();
  for (const instance of plant.instances) {
    if (hallOf(instance.id) === hall) counts.set(instance.kind, (counts.get(instance.kind) ?? 0) + 1);
  }
  return [...counts].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]));
}

export function areaOrder(plant: Plant): readonly (readonly [string, number])[] {
  const declared = areaTree(plant.areas).map((node) => [node.name, node.depth] as const);
  const loose = [...new Set(plant.instances.map((instance) => areaOf(plant, instance)))].filter(
    (area) => !declared.some(([name]) => name === area),
  );
  return [...declared, ...loose.map((area) => [area, 0] as const)];
}

export function areaCounts(plant: Plant): readonly (readonly [string, number])[] {
  const tally = new Map<string, number>();
  for (const instance of plant.instances) {
    const area = areaOf(plant, instance);
    tally.set(area, (tally.get(area) ?? 0) + 1);
  }
  return areaOrder(plant).flatMap(([name]) => (tally.has(name) ? [[name, tally.get(name) ?? 0] as const] : []));
}

export function hallCount(view: PlantView, hall: string): string {
  const own = view.nodes.filter((node) => hallOf(node.id) === hall).length;
  const shared = view.nodes.filter((node) => node.instance !== null).length - own;
  return `${String(own)} instances${shared > 0 ? ` plus ${String(shared)} shared` : ""}`;
}
