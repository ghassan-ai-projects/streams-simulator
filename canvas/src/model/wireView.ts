import type { PlantView, ViewNode } from "./plantView";
import { headerHeight, groupCounts, nodeSize, noteHeight } from "./plantView";
import { titleAbove, titleBandHeight, titleScale } from "./frameTitle";
import type { Rect, Tier } from "./viewport";
import type { End } from "./wireEnds";
import { runsAcross, spreadEnds } from "./wireEnds";
import type { Point, WireLink } from "./wireRoute";
import { routeWires, wireClearance } from "./wireRoute";

export type WireMode = "pins" | "rows" | "centre";

const titleText = { height: 16, inside: 6 } as const;
const firstPinRow = 5;

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
  const first = mode === "rows" ? noteHeight(node.noteLines) + 1 : firstPinRow;
  return box.y + headerHeight(node.label) + first + (row + 0.5) * nodeSize.pin;
}

const noBox: Rect = { x: 0, y: 0, width: 0, height: 0 };

function stacked(a: Rect, b: Rect): boolean {
  const aligned = Math.abs(a.x - b.x) < 0.5 && Math.abs(a.width - b.width) < 0.5;
  return aligned && (a.y + a.height <= b.y || b.y + b.height <= a.y);
}

function endY(view: PlantView, id: string, handle: string | null, role: "from" | "to", mode: WireMode): number | null {
  const node = view.nodes.find((entry) => entry.id === id);
  const box = view.layout.nodes.get(id);
  const row = rowOf(node, handle, role === "from" ? "out" : "in", mode);
  if (node === undefined || box === undefined || row === null) return null;
  return rowCentre(node, box, row, mode);
}

function endOf(view: PlantView, edge: PlantView["edges"][number], role: "from" | "to", mode: WireMode): End {
  const id = role === "from" ? edge.from : edge.to;
  const handle = role === "from" ? edge.fromHandle : edge.toHandle;
  const box = view.layout.nodes.get(id) ?? noBox;
  const y = endY(view, id, handle, role, mode) ?? box.y + box.height / 2;
  const point = { x: role === "from" ? box.x + box.width : box.x, y };
  const slot = role === "from" ? `${edge.from}.${edge.fromHandle ?? ""}` : edge.id;
  return { id: edge.id, role, point, box, side: role === "from" ? "right" : "left", slot, other: point };
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
  const [a, b] = stacked(from.box, to.box) ? [columnEnd(from, to.box), columnEnd(to, from.box)] : [from, to];
  return [
    { ...a, other: b.point },
    { ...b, other: a.point },
  ];
}

function sideOf(end: End): "vertical" | "horizontal" {
  return runsAcross(end) ? "vertical" : "horizontal";
}

function stubOf(from: End, to: End): number | undefined {
  if (sideOf(from) !== "vertical") return undefined;
  const gap = Math.abs(to.point.y - from.point.y);
  return gap < 2 * wireClearance.stub ? Math.max(1, gap / 2) : undefined;
}

function linksOf(view: PlantView, mode: WireMode, pixel: number): WireLink[] {
  const pairs = view.edges.map((edge) => pairOf(view, edge, mode));
  const placed = spreadEnds(pairs.flat(), pixel);
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

export function titleBox(frame: Rect, tier: Tier, pixel: number, title = "", count = 0): Rect {
  if (tier === "full") {
    return { x: frame.x, y: frame.y + titleText.inside, width: frame.width, height: titleText.height };
  }
  const scale = titleScale(pixel);
  const height = titleBandHeight(title, count, frame.width, scale);
  return { x: frame.x, y: frame.y - (titleAbove(scale) + height), width: frame.width, height };
}

export function routeView(
  view: PlantView,
  tier: Tier,
  review: boolean,
  pixel = 1,
): ReadonlyMap<string, readonly Point[]> {
  const counts = groupCounts(view);
  return routeWires({
    links: linksOf(view, wireModeOf(tier, review), pixel),
    nodes: [...view.layout.nodes.values()],
    titles: [...view.layout.groups].map(([title, frame]) =>
      titleBox(frame, tier, pixel, title, counts.get(title) ?? 0),
    ),
    titleClearances: [9 * pixel, 5 * pixel, 3 * pixel, 1.5 * pixel, 0],
  });
}
