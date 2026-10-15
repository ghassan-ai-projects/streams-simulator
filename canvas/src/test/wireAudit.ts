import type { PlantView } from "../model/plantView";
import type { Rect } from "../model/viewport";
import type { Point } from "../model/wireRoute";

export interface Dump {
  readonly nodes: readonly {
    readonly id: string;
    readonly x: number;
    readonly y: number;
    readonly width: number;
    readonly height: number;
    readonly ins: readonly string[];
    readonly outs: readonly string[];
  }[];
  readonly frames: readonly {
    readonly title: string;
    readonly x: number;
    readonly y: number;
    readonly width: number;
    readonly height: number;
  }[];
  readonly edges: readonly string[];
}

export function viewOf(dump: Dump): PlantView {
  const pin = (name: string) => ({ name });
  const nodes = dump.nodes.map((node) => ({
    id: node.id,
    label: node.id,
    noteLines: 1,
    pins: { ins: node.ins.map(pin), outs: node.outs.map(pin) },
  }));
  const edges = dump.edges.map((id) => {
    const [from = "", rest = ""] = id.split(">");
    const [fromId = "", fromHandle = ""] = from.split(".");
    const [toId = "", toHandle = ""] = rest.split(".");
    return { id, from: fromId, to: toId, fromHandle, toHandle, family: null, count: 1 };
  });
  const box = (x: number, y: number, width: number, height: number): Rect => ({ x, y, width, height });
  return {
    nodes,
    edges,
    layout: {
      nodes: new Map(dump.nodes.map((n) => [n.id, box(n.x, n.y, n.width, n.height)] as const)),
      groups: new Map(dump.frames.map((f) => [f.title, box(f.x, f.y, f.width, f.height)] as const)),
    },
    halls: [],
  } as unknown as PlantView;
}

type Routes = ReadonlyMap<string, readonly Point[]>;

type Segment = readonly [Point, Point];

function segmentsOf(route: readonly Point[]): Segment[] {
  return route.flatMap((a, index) => {
    const b = route[index + 1];
    return b === undefined ? [] : [[a, b] as const];
  });
}

function samples(route: readonly Point[], step: number): Point[] {
  return segmentsOf(route).flatMap(([a, b]) => {
    const count = Math.max(1, Math.ceil((Math.abs(b.x - a.x) + Math.abs(b.y - a.y)) / step));
    return Array.from({ length: count }, (_, at) => ({
      x: a.x + ((b.x - a.x) * at) / count,
      y: a.y + ((b.y - a.y) * at) / count,
    }));
  });
}

export function notEnteringAtPin(view: PlantView, routes: Routes): string[] {
  return view.edges.flatMap((edge) => {
    const route = routes.get(edge.id) ?? [];
    const to = view.layout.nodes.get(edge.to);
    const last = route[route.length - 1];
    const before = route[route.length - 2];
    if (to === undefined || last === undefined || before === undefined) return [`${edge.id}: no route`];
    const level = Math.abs(before.y - last.y) < 1e-6 && before.x < last.x;
    const onSide = Math.abs(last.x - to.x) < 1e-6 && last.y >= to.y && last.y <= to.y + to.height;
    return level && onSide ? [] : [`${edge.id}: ends at ${String(last.x)},${String(last.y)}`];
  });
}

export function behindNodes(view: PlantView, routes: Routes): string[] {
  const found: string[] = [];
  for (const edge of view.edges) {
    for (const point of samples(routes.get(edge.id) ?? [], 1)) {
      const hit = [...view.layout.nodes].find(
        ([, box]) =>
          point.x > box.x + 0.5 &&
          point.x < box.x + box.width - 0.5 &&
          point.y > box.y + 0.5 &&
          point.y < box.y + box.height - 0.5,
      );
      if (hit !== undefined) {
        found.push(`${edge.id}: inside ${hit[0]}`);
        break;
      }
    }
  }
  return found;
}
