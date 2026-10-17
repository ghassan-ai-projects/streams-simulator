import type { LayoutNode, PlantView, ViewEdge, ViewNode, ViewPresentation } from "./types";
import { layout } from "./layout";
import { headerReserve } from "./frameTitle";
import { tally } from "../../shared/collect";
import { hallOf, hallsOf } from "./halls";
import type { ComponentSpec, Family, Instance, Link, Pin, Pins, Plant } from "../../plant";
import { areaOf, familyOf, pinsOf } from "../../plant";
import { kindLabel, shortChannel } from "../../shared/words";

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

function plainRows(pins: Pins): number {
  return Math.max(2, pins.ins.length, pins.outs.length);
}

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
  return tally(view.nodes, (node) => node.group);
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
  const counts = tally(
    plant.instances.flatMap((instance) => hallOf(instance.id) ?? []),
    (hall) => hall,
  );
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

interface PlantViewOptions {
  readonly open: string | null;
  readonly view?: ViewPresentation;
  readonly aspect?: number | undefined;
  readonly notes?: Notes;
}

export function plantView(plant: Plant, specs: Specs, options: PlantViewOptions): PlantView {
  const { open, view = plainView, aspect, notes = new Map() } = options;
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
