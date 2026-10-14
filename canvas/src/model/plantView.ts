import type { Layout, LayoutNode } from "./layout";
import { layout } from "./layout";
import type { Instance, Link, Plant } from "./plant";
import { areaOf } from "./plant";
import type { Category, ComponentSpec, Family, Pins } from "./ports";
import { familyOf, pinsOf } from "./ports";
import type { NodeStatus } from "./readout";
import { statusWords } from "./readout";
import { kindLabel } from "./report";

export interface ViewNode {
  readonly id: string;
  readonly label: string;
  readonly detail: string;
  readonly group: string;
  readonly instance: Instance | null;
  readonly category: Category;
  readonly members: number;
  readonly pins: Pins;
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

export const nodeSize = { width: 232, header: 28, pin: 24, padding: 16, groupWidth: 176, groupHeight: 72 } as const;

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

function instanceNode(plant: Plant, instance: Instance, specs: Specs): ViewNode {
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
  };
}

function linkFamily(plant: Plant, link: Link, specs: Specs): Family | null {
  const source = plant.instances.find((instance) => instance.id === link.from);
  const spec = source?.component
    ? specs.get(source.component)?.ports.find((port) => port.name === link.fromPort)
    : undefined;
  return spec ? familyOf(spec.type) : null;
}

function bundle(
  plant: Plant,
  links: readonly Link[],
  key: (id: string) => string | null,
  specs: Specs,
  detailed: boolean,
): ViewEdge[] {
  const edges = new Map<string, ViewEdge>();
  for (const link of links) {
    const from = key(link.from);
    const to = key(link.to);
    if (from === null || to === null || from === to) continue;
    const fromHandle = detailed ? link.fromPort : null;
    const toHandle = detailed ? link.toPort : null;
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
  readonly readout: number;
  readonly aspect: number | undefined;
}

function nodeBox(node: ViewNode, readout: number): LayoutNode {
  if (node.instance === null) {
    return { id: node.id, group: node.group, width: nodeSize.groupWidth, height: nodeSize.groupHeight };
  }
  const rows = readout > 0 ? 0 : Math.max(2, node.pins.ins.length, node.pins.outs.length);
  const height = nodeSize.header + nodeSize.padding + rows * nodeSize.pin + readout;
  return { id: node.id, group: node.group, width: nodeSize.width, height };
}

function finish(context: Context, nodes: readonly ViewNode[], edges: readonly ViewEdge[]): PlantView {
  const order = ["Halls", ...context.plant.areas.map((area) => area.name)];
  const boxes = nodes.map((node) => nodeBox(node, context.readout));
  const rows = (group: string, count: number) => (group === "Halls" ? Math.ceil(Math.sqrt(count)) : 3);
  return { nodes, edges, layout: layout(boxes, edges, order, rows, context.aspect), halls: context.halls };
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
    ...shared.map((instance) => instanceNode(plant, instance, specs)),
    ...halls.map((hall) => hallNode(hall, counts.get(hall) ?? 0)),
  ];
  const key = (id: string) => hallOf(id) ?? id;
  const edges = bundle(plant, plant.links, key, specs, false);
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
  const nodes = kept.map((instance) => instanceNode(context.plant, instance, context.specs));
  const key = (id: string) => (ids.has(id) ? id : null);
  return finish(context, nodes, bundle(context.plant, context.plant.links, key, context.specs, true));
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

export function plantView(plant: Plant, specs: Specs, open: string | null, readout = 0, aspect?: number): PlantView {
  const context = { plant, specs, halls: hallsOf(plant), readout, aspect };
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

function counted(n: number, noun: string): string {
  return `${String(n)} ${noun}${n === 1 ? "" : "s"}`;
}

function stateOf(status: NodeStatus, count: number, fault: string): string {
  if (status === null) return "";
  const where = count > 0 ? ` in ${String(count)}` : "";
  if (fault === "") return `${statusWords[status]}${where}, `;
  return status === "fault" ? `${statusWords.fault}: ${fault}, ` : `${statusWords[status]}${where}, Fault: ${fault}, `;
}

export function describeNode(
  node: ViewNode,
  selected: boolean,
  status: NodeStatus = null,
  count = 0,
  fault = "",
): string {
  const ports =
    node.instance === null
      ? "Space selects, Enter opens the hall"
      : `${counted(node.pins.ins.length, "input")}, ${counted(node.pins.outs.length, "output")}`;
  const state = stateOf(status, count, fault);
  return `${node.label}, ${state}${node.detail}, ${ports}${selected ? ", selected" : ""}`;
}

export function hallMembers(plant: Plant, hall: string): readonly (readonly [string, number])[] {
  const counts = new Map<string, number>();
  for (const instance of plant.instances) {
    if (hallOf(instance.id) === hall) counts.set(instance.kind, (counts.get(instance.kind) ?? 0) + 1);
  }
  return [...counts].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]));
}

export function hallCount(view: PlantView, hall: string): string {
  const own = view.nodes.filter((node) => hallOf(node.id) === hall).length;
  const shared = view.nodes.filter((node) => node.instance !== null).length - own;
  return `${String(own)} instances${shared > 0 ? ` plus ${String(shared)} shared` : ""}`;
}
