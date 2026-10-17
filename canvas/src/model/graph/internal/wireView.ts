import type { PlantView, Point, Rect, ViewEdge, ViewNode } from "./types";
import { headerHeight, nodeSize, noteHeight } from "./plantView";
import { titleAbove, titleBandHeight, titleScale } from "./frameTitle";
import type { Tier } from "./viewport";
import type { End } from "./wireEnds";
import { runsAcross, spreadEnds } from "./wireEnds";
import type { WireLink } from "./wireRoute";
import { routeWires, wireClearance } from "./wireRoute";

type WireMode = "pins" | "rows" | "centre";

const titleText = { height: 16, inside: 6 } as const;
const firstPinRow = 5;

function wireModeOf(tier: Tier, detail: "pins" | "rows"): WireMode {
  if (tier === "overview" || tier === "map") return "centre";
  return detail;
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

function endY(view: PlantView, id: string, handle: string | null, role: "from" | "to", mode: WireMode): number | null {
  const node = view.nodes.find((entry) => entry.id === id);
  const box = view.layout.nodes.get(id);
  const row = rowOf(node, handle, role === "from" ? "out" : "in", mode);
  if (node === undefined || box === undefined || row === null) return null;
  return rowCentre(node, box, row, mode);
}

function endOf(view: PlantView, edge: ViewEdge, role: "from" | "to", mode: WireMode): End {
  const id = role === "from" ? edge.from : edge.to;
  const handle = role === "from" ? edge.fromHandle : edge.toHandle;
  const box = view.layout.nodes.get(id) ?? noBox;
  const y = endY(view, id, handle, role, mode) ?? box.y + box.height / 2;
  const point = { x: role === "from" ? box.x + box.width : box.x, y };
  const slot = role === "from" ? `${edge.from}.${edge.fromHandle ?? ""}` : edge.id;
  return { id: edge.id, role, point, box, side: role === "from" ? "right" : "left", slot, other: point };
}

function pairOf(view: PlantView, edge: ViewEdge, mode: WireMode): readonly [End, End] {
  const from = endOf(view, edge, "from", mode);
  const to = endOf(view, edge, "to", mode);
  return [
    { ...from, other: to.point },
    { ...to, other: from.point },
  ];
}

function stubOf(from: End, to: End): number | undefined {
  if (!runsAcross(from)) return undefined;
  const gap = Math.abs(to.point.y - from.point.y);
  return gap < 2 * wireClearance.stub ? Math.max(1, gap / 2) : undefined;
}

const corridorGutter = 48;

function corridorOf(view: PlantView, edge: ViewEdge): Rect | null {
  const from = view.layout.nodes.get(edge.from);
  const to = view.layout.nodes.get(edge.to);
  if (from === undefined || to === undefined) return null;
  const x = Math.min(from.x, to.x) - corridorGutter;
  const y = Math.min(from.y, to.y) - corridorGutter;
  const right = Math.max(from.x + from.width, to.x + to.width) + corridorGutter;
  const bottom = Math.max(from.y + from.height, to.y + to.height) + corridorGutter;
  return { x, y, width: right - x, height: bottom - y };
}

function wireLinks(view: PlantView, mode: WireMode, pixel: number): WireLink[] {
  const routed = view.edges.map((edge) => ({ edge, pair: pairOf(view, edge, mode) }));
  const placed = spreadEnds(
    routed.flatMap(({ pair }) => pair),
    pixel,
  );
  return routed.map(({ edge, pair: [from, to] }) => {
    const stub = stubOf(from, to);
    const corridor = corridorOf(view, edge);
    return {
      id: edge.id,
      from: placed.get(`${edge.id}:from`) ?? from.point,
      to: placed.get(`${edge.id}:to`) ?? to.point,
      leaves: from.side,
      enters: to.side,
      ...(corridor === null ? {} : { corridor }),
      ...(stub === undefined ? {} : { stub }),
    };
  });
}

export function titleBox(frame: Rect, tier: Tier, pixel: number): Rect {
  if (tier === "full") {
    return { x: frame.x, y: frame.y + titleText.inside, width: frame.width, height: titleText.height };
  }
  const scale = titleScale(pixel);
  const height = titleBandHeight(scale);
  return { x: frame.x, y: frame.y - (titleAbove(scale) + height), width: frame.width, height };
}

export function routeView(
  view: PlantView,
  tier: Tier,
  detail: "pins" | "rows",
  pixel: number,
): ReadonlyMap<string, readonly Point[]> {
  return routeWires({
    links: wireLinks(view, wireModeOf(tier, detail), pixel),
    nodes: [...view.layout.nodes.values()],
    titles: [...view.layout.groups.values()].map((frame) => titleBox(frame, tier, pixel)),
    titleClearances: [9 * pixel, 5 * pixel, 3 * pixel, 1.5 * pixel, 0],
    frames: [...view.layout.groups.values()],
    pixel,
  });
}
