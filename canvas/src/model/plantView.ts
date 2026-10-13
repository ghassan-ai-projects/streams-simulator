import type { Layout, LayoutNode } from "./layout";
import { layout } from "./layout";
import type { Instance, Link, Plant } from "./plant";
import { areaOf } from "./plant";
import type { Family, Pins, PortSpec } from "./ports";
import { familyOf, pinsOf } from "./ports";

export interface ViewNode {
  readonly id: string;
  readonly label: string;
  readonly detail: string;
  readonly group: string;
  readonly instance: Instance | null;
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

type Specs = ReadonlyMap<string, readonly PortSpec[]>;

function instanceNode(plant: Plant, instance: Instance, specs: Specs): ViewNode {
  const pins = pinsOf(plant, instance, instance.component === null ? undefined : specs.get(instance.component));
  return {
    id: instance.id,
    label: instance.id,
    detail: instance.kind,
    group: areaOf(plant, instance),
    instance,
    members: 1,
    pins,
  };
}

function hallNode(hall: string, members: number): ViewNode {
  return {
    id: hall,
    label: hall,
    detail: `${String(members)} instances`,
    group: "Halls",
    instance: null,
    members,
    pins: { ins: [], outs: [] },
  };
}

function linkFamily(plant: Plant, link: Link, specs: Specs): Family | null {
  const source = plant.instances.find((instance) => instance.id === link.from);
  const spec = source?.component ? specs.get(source.component)?.find((port) => port.name === link.fromPort) : undefined;
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

function nodeBox(node: ViewNode): LayoutNode {
  if (node.instance === null) {
    return { id: node.id, group: node.group, width: nodeSize.groupWidth, height: nodeSize.groupHeight };
  }
  const rows = Math.max(1, node.pins.ins.length, node.pins.outs.length);
  return {
    id: node.id,
    group: node.group,
    width: nodeSize.width,
    height: nodeSize.header + nodeSize.padding + rows * nodeSize.pin,
  };
}

function finish(
  plant: Plant,
  nodes: readonly ViewNode[],
  edges: readonly ViewEdge[],
  halls: readonly string[],
): PlantView {
  const order = ["Halls", ...plant.areas.map((area) => area.name)];
  return { nodes, edges, layout: layout(nodes.map(nodeBox), edges, order), halls };
}

function hallLevel(plant: Plant, halls: readonly string[], specs: Specs): PlantView {
  const shared = plant.instances.filter((instance) => hallOf(instance.id) === null);
  const counts = new Map<string, number>();
  plant.instances.forEach((instance) => {
    const hall = hallOf(instance.id);
    if (hall !== null) counts.set(hall, (counts.get(hall) ?? 0) + 1);
  });
  const nodes = [
    ...shared.map((instance) => instanceNode(plant, instance, specs)),
    ...halls.map((hall) => hallNode(hall, counts.get(hall) ?? 0)),
  ];
  const key = (id: string) => hallOf(id) ?? id;
  return finish(plant, nodes, bundle(plant, plant.links, key, specs, false), halls);
}

function instanceLevel(
  plant: Plant,
  keep: (instance: Instance) => boolean,
  specs: Specs,
  halls: readonly string[],
): PlantView {
  const kept = plant.instances.filter(keep);
  const ids = new Set(kept.map((instance) => instance.id));
  const nodes = kept.map((instance) => instanceNode(plant, instance, specs));
  return finish(
    plant,
    nodes,
    bundle(plant, plant.links, (id) => (ids.has(id) ? id : null), specs, true),
    halls,
  );
}

export function plantView(plant: Plant, specs: Specs, open: string | null): PlantView {
  const halls = hallsOf(plant);
  if (halls.length === 0) {
    return instanceLevel(plant, () => true, specs, halls);
  }
  if (open === null || !halls.includes(open)) {
    return hallLevel(plant, halls, specs);
  }
  const touching = new Set(
    plant.links
      .filter((link) => hallOf(link.to) === open || hallOf(link.from) === open)
      .flatMap((link) => [link.from, link.to]),
  );
  return instanceLevel(
    plant,
    (instance) => hallOf(instance.id) === open || (hallOf(instance.id) === null && touching.has(instance.id)),
    specs,
    halls,
  );
}
