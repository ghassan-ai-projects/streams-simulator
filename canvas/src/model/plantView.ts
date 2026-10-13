import type { Layout, LayoutNode } from "./layout";
import { layout } from "./layout";
import type { Instance, Link, Plant } from "./plant";
import { areaOf } from "./plant";
import type { Category, ComponentSpec, Family, Pins } from "./ports";
import { familyOf, pinsOf } from "./ports";

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
    detail: instance.kind,
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
}

function nodeBox(node: ViewNode, readout: number): LayoutNode {
  if (node.instance === null) {
    return { id: node.id, group: node.group, width: nodeSize.groupWidth, height: nodeSize.groupHeight };
  }
  const rows = Math.max(1, node.pins.ins.length, node.pins.outs.length);
  const height = nodeSize.header + nodeSize.padding + rows * nodeSize.pin + readout;
  return { id: node.id, group: node.group, width: nodeSize.width, height };
}

function finish(context: Context, nodes: readonly ViewNode[], edges: readonly ViewEdge[]): PlantView {
  const order = ["Halls", ...context.plant.areas.map((area) => area.name)];
  const boxes = nodes.map((node) => nodeBox(node, context.readout));
  return { nodes, edges, layout: layout(boxes, edges, order), halls: context.halls };
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
  return finish(context, nodes, bundle(plant, plant.links, key, specs, false));
}

function instanceLevel(context: Context, keep: (instance: Instance) => boolean): PlantView {
  const kept = context.plant.instances.filter(keep);
  const ids = new Set(kept.map((instance) => instance.id));
  const nodes = kept.map((instance) => instanceNode(context.plant, instance, context.specs));
  const key = (id: string) => (ids.has(id) ? id : null);
  return finish(context, nodes, bundle(context.plant, context.plant.links, key, context.specs, true));
}

export function plantView(plant: Plant, specs: Specs, open: string | null, readout = 0): PlantView {
  const context = { plant, specs, halls: hallsOf(plant), readout };
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
  return instanceLevel(
    context,
    (instance) => hallOf(instance.id) === open || (hallOf(instance.id) === null && touching.has(instance.id)),
  );
}

export function describeNode(node: ViewNode, selected: boolean): string {
  const ports =
    node.instance === null
      ? `${String(node.members)} instances, Enter opens the hall`
      : `${String(node.pins.ins.length)} inputs, ${String(node.pins.outs.length)} outputs`;
  return `${node.label}, ${node.detail}, ${ports}${selected ? ", selected" : ""}`;
}
