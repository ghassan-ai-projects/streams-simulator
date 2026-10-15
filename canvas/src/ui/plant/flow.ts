import { MarkerType } from "@xyflow/react";
import type { Edge, Node } from "@xyflow/react";
import type { NodeStatus, Shown } from "../../model/readout";
import type { PlantView, ViewNode } from "../../model/plantView";

export interface NodeLive {
  readonly shown: readonly Shown[];
  readonly status: NodeStatus;
  readonly count: number;
  readonly fault: string;
}

type InstanceData = Readonly<{
  view: ViewNode;
  live: NodeLive | null;
  active: boolean;
  flagged: boolean;
  editable: boolean;
}>;
type FrameData = Readonly<{ title: string; count: number }>;

export type InstanceNodeType = Node<InstanceData, "instance">;
export type HallNodeType = Node<InstanceData, "hall">;
type FrameNodeType = Node<FrameData, "frame">;
export type FrameTitleNodeType = Node<FrameData, "frameTitle">;
export type FlowNode = InstanceNodeType | HallNodeType | FrameNodeType | FrameTitleNodeType;

function groupCounts(view: PlantView): Map<string, number> {
  const counts = new Map<string, number>();
  view.nodes.forEach((node) => counts.set(node.group, (counts.get(node.group) ?? 0) + 1));
  return counts;
}

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
  editable = false,
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
        editable,
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

function anchorHandles(view: PlantView, from: string, to: string): { sourceHandle: string; targetHandle: string } {
  const fromX = view.layout.nodes.get(from)?.x ?? 0;
  const toX = view.layout.nodes.get(to)?.x ?? 0;
  return toX >= fromX
    ? { sourceHandle: "out", targetHandle: "in" }
    : { sourceHandle: "out-left", targetHandle: "in-right" };
}

export function flowEdges(view: PlantView, anchors = false, selected: string | null = null): Edge[] {
  return view.edges.map((edge) => ({
    id: edge.id,
    source: edge.from,
    target: edge.to,
    ...(anchors
      ? anchorHandles(view, edge.from, edge.to)
      : { sourceHandle: edge.fromHandle ?? "out", targetHandle: edge.toHandle ?? "in" }),
    markerEnd: {
      type: MarkerType.ArrowClosed,
      width: 10,
      height: 10,
      color: markerColor(edge.family, selected, edge.from, edge.to),
    },
    className: `wire family-${edge.family ?? "none"}${emphasis(selected, edge.from, edge.to)}`,
    zIndex: touches(selected, edge.from, edge.to) ? 1 : 0,
    label: edge.count > 1 ? `×${String(edge.count)}` : undefined,
    ariaLabel: `${edge.from} to ${edge.to}${edge.count > 1 ? `, ${String(edge.count)} links` : ""}`,
    focusable: false,
  }));
}
