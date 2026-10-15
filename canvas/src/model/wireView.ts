import type { PlantView, ViewNode } from "./plantView";
import { headerHeight, nodeSize } from "./plantView";
import type { Rect, Tier } from "./viewport";
import type { Point, WireLink } from "./wireRoute";
import { routeWires, wireClearance } from "./wireRoute";

export type WireMode = "pins" | "rows" | "centre";

const frameTitle = 28;
const titleBand = { compact: 36, overview: 72, map: 72 } as const;
const firstPinRow = 5;
const firstReviewRow = nodeSize.note + 1;

interface End {
  readonly id: string;
  readonly role: "from" | "to";
  readonly point: Point;
  readonly box: Rect;
}

export function wireModeOf(tier: Tier, review: boolean): WireMode {
  if (tier === "overview" || tier === "map") return "centre";
  return review ? "rows" : "pins";
}

function rowOf(node: ViewNode | undefined, handle: string | null, side: "in" | "out", mode: WireMode): number | null {
  if (node === undefined || handle === null || mode === "centre") return null;
  const own = side === "in" ? node.pins.ins : node.pins.outs;
  const index = own.findIndex((pin) => pin.name === handle);
  if (index < 0) return null;
  return mode === "rows" && side === "out" ? node.pins.ins.length + index : index;
}

function rowCentre(node: ViewNode, box: Rect, row: number, mode: WireMode): number {
  const first = mode === "rows" ? firstReviewRow : firstPinRow;
  return box.y + headerHeight(node.label) + first + (row + 0.5) * nodeSize.pin;
}

function endOf(view: PlantView, edge: PlantView["edges"][number], role: "from" | "to", mode: WireMode): End {
  const id = role === "from" ? edge.from : edge.to;
  const handle = role === "from" ? edge.fromHandle : edge.toHandle;
  const box = view.layout.nodes.get(id) ?? { x: 0, y: 0, width: 0, height: 0 };
  const node = view.nodes.find((entry) => entry.id === id);
  const row = rowOf(node, handle, role === "from" ? "out" : "in", mode);
  const y = node === undefined || row === null ? box.y + box.height / 2 : rowCentre(node, box, row, mode);
  return { id: edge.id, role, point: { x: role === "from" ? box.x + box.width : box.x, y }, box };
}

function spread(ends: readonly End[]): ReadonlyMap<string, number> {
  const groups = new Map<string, End[]>();
  for (const end of ends) {
    const key = `${end.role}:${String(end.point.x)}:${String(end.point.y)}`;
    groups.set(key, [...(groups.get(key) ?? []), end]);
  }
  const offsets = new Map<string, number>();
  for (const group of groups.values()) {
    const sorted = [...group].sort((a, b) => a.id.localeCompare(b.id));
    sorted.forEach((end, index) => {
      const shift = (index - (sorted.length - 1) / 2) * wireClearance.lane;
      const low = end.box.y + wireClearance.lane;
      const high = end.box.y + end.box.height - wireClearance.lane;
      const y = Math.min(Math.max(end.point.y + shift, low), Math.max(low, high));
      offsets.set(`${end.id}:${end.role}`, y);
    });
  }
  return offsets;
}

function linksOf(view: PlantView, mode: WireMode): WireLink[] {
  const ends = view.edges.flatMap((edge) => [endOf(view, edge, "from", mode), endOf(view, edge, "to", mode)]);
  const placed = spread(ends);
  const pointOf = (end: End): Point => ({ x: end.point.x, y: placed.get(`${end.id}:${end.role}`) ?? end.point.y });
  return view.edges.map((edge, index) => {
    const from = ends[index * 2];
    const to = ends[index * 2 + 1];
    return {
      id: edge.id,
      from: from === undefined ? { x: 0, y: 0 } : pointOf(from),
      to: to === undefined ? { x: 0, y: 0 } : pointOf(to),
    };
  });
}

export function titleBox(frame: Rect, tier: Tier): Rect {
  if (tier === "full") return { x: frame.x, y: frame.y, width: frame.width, height: frameTitle };
  const band = titleBand[tier];
  return { x: frame.x, y: frame.y - band, width: frame.width, height: band };
}

export function routeView(view: PlantView, tier: Tier, review: boolean): ReadonlyMap<string, readonly Point[]> {
  return routeWires({
    links: linksOf(view, wireModeOf(tier, review)),
    nodes: [...view.layout.nodes.values()],
    titles: [...view.layout.groups.values()].map((frame) => titleBox(frame, tier)),
  });
}
