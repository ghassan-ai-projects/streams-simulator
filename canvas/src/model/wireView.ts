import type { PlantView, ViewNode } from "./plantView";
import { headerHeight, nodeSize } from "./plantView";
import type { Rect, Tier } from "./viewport";
import type { Point, Side, WireLink } from "./wireRoute";
import { routeWires, wireClearance } from "./wireRoute";

export type WireMode = "pins" | "rows" | "centre";

const titleText = { height: 16, line: 17, above: 4, inside: 6 } as const;
const titleChar = 6.3;
const firstPinRow = 5;
const firstReviewRow = nodeSize.note + 1;

interface End {
  readonly id: string;
  readonly role: "from" | "to";
  readonly point: Point;
  readonly box: Rect;
  readonly side: Side;
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

const noBox: Rect = { x: 0, y: 0, width: 0, height: 0 };

function stacked(a: Rect, b: Rect): boolean {
  const aligned = Math.abs(a.x - b.x) < 0.5 && Math.abs(a.width - b.width) < 0.5;
  return aligned && (a.y + a.height <= b.y || b.y + b.height <= a.y);
}

function endOf(view: PlantView, edge: PlantView["edges"][number], role: "from" | "to", mode: WireMode): End {
  const id = role === "from" ? edge.from : edge.to;
  const handle = role === "from" ? edge.fromHandle : edge.toHandle;
  const box = view.layout.nodes.get(id) ?? noBox;
  const node = view.nodes.find((entry) => entry.id === id);
  const row = rowOf(node, handle, role === "from" ? "out" : "in", mode);
  const y = node === undefined || row === null ? box.y + box.height / 2 : rowCentre(node, box, row, mode);
  const side = role === "from" ? "right" : "left";
  return { id: edge.id, role, point: { x: role === "from" ? box.x + box.width : box.x, y }, box, side };
}

function columnEnd(end: End, other: Rect): End {
  const lower = other.y > end.box.y;
  const y = lower ? end.box.y + end.box.height : end.box.y;
  const side = lower ? "down" : "up";
  return { ...end, point: { x: end.box.x + end.box.width / 2, y }, side };
}

function pairOf(view: PlantView, edge: PlantView["edges"][number], mode: WireMode): readonly [End, End] {
  const from = endOf(view, edge, "from", mode);
  const to = endOf(view, edge, "to", mode);
  if (!stacked(from.box, to.box)) return [from, to];
  return [columnEnd(from, to.box), columnEnd(to, from.box)];
}

function sideOf(end: End): "vertical" | "horizontal" {
  return end.side === "up" || end.side === "down" ? "vertical" : "horizontal";
}

function spread(ends: readonly End[]): ReadonlyMap<string, Point> {
  const groups = new Map<string, End[]>();
  for (const end of ends) {
    const key = `${end.side}:${String(end.point.x)}:${String(end.point.y)}`;
    groups.set(key, [...(groups.get(key) ?? []), end]);
  }
  const placed = new Map<string, Point>();
  for (const group of groups.values()) {
    const sorted = [...group].sort((a, b) => a.id.localeCompare(b.id) || a.role.localeCompare(b.role));
    sorted.forEach((end, index) => {
      const shift = (index - (sorted.length - 1) / 2) * wireClearance.lane;
      placed.set(`${end.id}:${end.role}`, shifted(end, shift));
    });
  }
  return placed;
}

function shifted(end: End, shift: number): Point {
  const { box, point } = end;
  const lane = wireClearance.lane;
  if (sideOf(end) === "vertical") {
    const x = Math.min(Math.max(point.x + shift, box.x + lane), Math.max(box.x + lane, box.x + box.width - lane));
    return { x, y: point.y };
  }
  const y = Math.min(Math.max(point.y + shift, box.y + lane), Math.max(box.y + lane, box.y + box.height - lane));
  return { x: point.x, y };
}

function stubOf(from: End, to: End): number | undefined {
  if (sideOf(from) !== "vertical") return undefined;
  const gap = Math.abs(to.point.y - from.point.y);
  return gap < 2 * wireClearance.stub ? Math.max(1, gap / 2) : undefined;
}

function linksOf(view: PlantView, mode: WireMode): WireLink[] {
  const pairs = view.edges.map((edge) => pairOf(view, edge, mode));
  const placed = spread(pairs.flat());
  return view.edges.map((edge, index) => {
    const [from, to] = pairs[index] ?? [];
    if (from === undefined || to === undefined) return { id: edge.id, from: { x: 0, y: 0 }, to: { x: 0, y: 0 } };
    const stub = stubOf(from, to);
    return {
      id: edge.id,
      from: placed.get(`${edge.id}:from`) ?? from.point,
      to: placed.get(`${edge.id}:to`) ?? to.point,
      leaves: from.side,
      enters: to.side,
      ...(stub === undefined ? {} : { stub }),
    };
  });
}

export function titleBox(frame: Rect, tier: Tier, pixel: number, title = ""): Rect {
  if (tier === "full") {
    return { x: frame.x, y: frame.y + titleText.inside, width: frame.width, height: titleText.height };
  }
  const words = (title.length * titleChar + 4) * pixel;
  const lines = Math.max(1, Math.ceil(words / Math.max(1, frame.width)));
  const height = titleText.line * lines * pixel;
  return { x: frame.x, y: frame.y - (titleText.above * pixel + height), width: Math.min(words, frame.width), height };
}

export function routeView(
  view: PlantView,
  tier: Tier,
  review: boolean,
  pixel = 1,
): ReadonlyMap<string, readonly Point[]> {
  return routeWires({
    links: linksOf(view, wireModeOf(tier, review)),
    nodes: [...view.layout.nodes.values()],
    titles: [...view.layout.groups].map(([title, frame]) => titleBox(frame, tier, pixel, title)),
    titleClearances: [9 * pixel, 5 * pixel, 3 * pixel, 1.5 * pixel, 0],
  });
}
