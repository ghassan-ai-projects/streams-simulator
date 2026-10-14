import type { Edge, Node } from "@xyflow/react";
import type { NodeStatus, Shown } from "../../model/readout";
import type { PlantView, ViewNode } from "../../model/plantView";

export interface NodeLive {
  readonly shown: readonly Shown[];
  readonly status: NodeStatus;
  readonly count: number;
  readonly fault: string;
}

type InstanceData = Readonly<{ view: ViewNode; live: NodeLive | null; active: boolean }>;
type FrameData = Readonly<{ title: string; count: number }>;

export type InstanceNodeType = Node<InstanceData, "instance">;
export type HallNodeType = Node<InstanceData, "hall">;
export type FrameNodeType = Node<FrameData, "frame">;
export type FlowNode = InstanceNodeType | HallNodeType | FrameNodeType;

function frames(view: PlantView): FrameNodeType[] {
  const counts = new Map<string, number>();
  view.nodes.forEach((node) => counts.set(node.group, (counts.get(node.group) ?? 0) + 1));
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

export function flowNodes(
  view: PlantView,
  selected: string | null,
  live: ((id: string) => NodeLive) | null = null,
  active: string | null = null,
): FlowNode[] {
  const nodes = view.nodes.map((node): InstanceNodeType | HallNodeType => {
    const box = view.layout.nodes.get(node.id) ?? { x: 0, y: 0, width: 0, height: 0 };
    return {
      id: node.id,
      type: node.instance === null ? "hall" : "instance",
      position: { x: box.x, y: box.y },
      data: { view: node, live: live === null ? null : live(node.id), active: node.id === active },
      selected: node.id === selected,
      draggable: false,
      width: box.width,
      height: box.height,
      style: { width: box.width },
    };
  });
  return [...frames(view), ...nodes];
}

export function flowEdges(view: PlantView, anchors = false): Edge[] {
  return view.edges.map((edge) => ({
    id: edge.id,
    source: edge.from,
    target: edge.to,
    sourceHandle: anchors ? "out" : (edge.fromHandle ?? "out"),
    targetHandle: anchors ? "in" : (edge.toHandle ?? "in"),
    className: `wire family-${edge.family ?? "none"}`,
    label: edge.count > 1 ? `×${String(edge.count)}` : undefined,
    ariaLabel: `${edge.from} to ${edge.to}${edge.count > 1 ? `, ${String(edge.count)} links` : ""}`,
    focusable: false,
  }));
}
