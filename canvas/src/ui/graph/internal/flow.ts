import type { NodeLive } from "./types";
import { MarkerType } from "@xyflow/react";
import type { Edge, Node } from "@xyflow/react";
import type { PlanMark, PlantView, Point, ViewEdge, ViewNode } from "../../../model/graph";
import { groupCounts } from "../../../model/graph";

type InstanceData = Readonly<{
  view: ViewNode;
  live: NodeLive | null;
  active: boolean;
  flagged: boolean;
  editable: boolean;
  plan: PlanMark | null;
}>;
type FrameData = Readonly<{ title: string; count: number }>;

export type InstanceNodeType = Node<InstanceData, "instance">;
export type HallNodeType = Node<InstanceData, "hall">;
type FrameNodeType = Node<FrameData, "frame">;
export type FrameTitleNodeType = Node<FrameData, "frameTitle">;
export type FlowNode = InstanceNodeType | HallNodeType | FrameNodeType | FrameTitleNodeType;

function frames(view: PlantView): FrameNodeType[] {
  const counts = groupCounts(view);
  return [...view.layout.groups.entries()].map(([title, box]) => ({
    id: `frame:${title}`,
    type: "frame",
    position: { x: box.x, y: box.y },
    data: { title, count: counts.get(title) ?? 0 },
    width: box.width,
    height: box.height,
    style: { width: box.width, height: box.height },
    selectable: false,
    draggable: false,
    focusable: false,
    zIndex: -1,
  }));
}

function frameTitles(view: PlantView): FrameTitleNodeType[] {
  const counts = groupCounts(view);
  return [...view.layout.groups.entries()].map(([title, box]) => ({
    id: `title:${title}`,
    type: "frameTitle",
    position: { x: box.x, y: box.y },
    data: { title, count: counts.get(title) ?? 0 },
    width: box.width,
    style: { width: box.width },
    selectable: false,
    draggable: false,
    focusable: false,
    zIndex: 2,
  }));
}

export function flowNodes(
  view: PlantView,
  selected: string | null,
  live: ((id: string) => NodeLive) | null = null,
  active: string | null = null,
  flagged: ReadonlySet<string> | null = null,
  interaction: "inspect" | "connect" = "inspect",
  marks: ReadonlyMap<string, PlanMark> | null = null,
): FlowNode[] {
  const nodes = view.nodes.map((node): InstanceNodeType | HallNodeType => {
    const box = view.layout.nodes.get(node.id) ?? { x: 0, y: 0, width: 0, height: 0 };
    return {
      id: node.id,
      type: node.instance === null ? "hall" : "instance",
      position: { x: box.x, y: box.y },
      data: {
        view: node,
        live: live === null ? null : live(node.id),
        active: node.id === active,
        flagged: flagged?.has(node.id) ?? false,
        editable: interaction === "connect",
        plan: marks?.get(node.id) ?? null,
      },
      selected: node.id === selected,
      draggable: false,
      width: box.width,
      height: box.height,
      style: { width: box.width },
    };
  });
  return [...frames(view), ...nodes, ...frameTitles(view)];
}

function touches(selected: string | null, from: string, to: string): boolean {
  return selected !== null && (from === selected || to === selected);
}

function emphasis(selected: string | null, from: string, to: string): string {
  if (selected === null) return "";
  return touches(selected, from, to) ? " wire-on" : " wire-off";
}

function familyColor(family: PlantView["edges"][number]["family"]): string {
  return family === null ? "var(--border-strong)" : `var(--port-${family})`;
}

function markerColor(
  family: PlantView["edges"][number]["family"],
  selected: string | null,
  from: string,
  to: string,
): string {
  if (selected !== null && !touches(selected, from, to)) {
    return "color-mix(in srgb, var(--text-tertiary) 30%, transparent)";
  }
  return familyColor(family);
}

function arrow(color: string) {
  return { type: MarkerType.ArrowClosed, width: 10, height: 10, color };
}

function wireWords(edge: ViewEdge): string {
  const total = edge.count + edge.reverse;
  if (total === 1) return `${edge.from} to ${edge.to}`;
  if (edge.reverse === 0) return `${edge.from} to ${edge.to}, ${String(total)} links`;
  return `${edge.from} and ${edge.to}, ${String(total)} links: ${String(edge.count)} to ${edge.to}, ${String(edge.reverse)} to ${edge.from}`;
}

function wireClass(edge: ViewEdge, selected: string | null): string {
  const grouped = edge.count + edge.reverse > 1 ? " wire-grouped" : "";
  return `wire family-${edge.family ?? "none"}${grouped}${emphasis(selected, edge.from, edge.to)}`;
}

export function flowEdges(
  view: PlantView,
  routes: ReadonlyMap<string, readonly Point[]>,
  selected: string | null = null,
): Edge[] {
  return view.edges.map((edge) => {
    const marker = arrow(markerColor(edge.family, selected, edge.from, edge.to));
    const total = edge.count + edge.reverse;
    return {
      id: edge.id,
      type: "wire",
      source: edge.from,
      target: edge.to,
      sourceHandle: "out",
      targetHandle: "in",
      data: { points: routes.get(edge.id) ?? [] },
      markerEnd: marker,
      ...(edge.reverse > 0 ? { markerStart: { ...marker, orient: "auto-start-reverse" } } : {}),
      className: wireClass(edge, selected),
      zIndex: 0,
      label: total > 1 ? `×${String(total)}` : undefined,
      ariaLabel: wireWords(edge),
      focusable: false,
    };
  });
}
