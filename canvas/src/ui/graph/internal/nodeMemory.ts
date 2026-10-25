import { sameValue } from "../../../model/graph";
import type { FlowNode } from "./flow";

export class NodeMemory {
  private byId = new Map<string, FlowNode>();

  stable(next: readonly FlowNode[]): FlowNode[] {
    const kept = next.map((node) => {
      const before = this.byId.get(node.id);
      return before !== undefined && sameValue(before, node) ? before : node;
    });
    this.byId = new Map(kept.map((node) => [node.id, node]));
    return kept;
  }
}

function shapeOf(node: FlowNode): string {
  if (node.type === "frame" || node.type === "frameTitle")
    return `${node.type}:${node.data.title}:${String(node.data.count)}`;
  const rows =
    node.data.live === null
      ? "ports"
      : `${String(node.data.live.readout.rows.length)}:${String(node.data.live.readout.hidden > 0)}`;
  return `${node.data.view.id}:${rows}:${String(node.width)}:${String(node.height)}`;
}

export function shapesOf(nodes: readonly FlowNode[]): ReadonlyMap<string, string> {
  return new Map(nodes.map((node) => [node.id, shapeOf(node)]));
}

export function reshaped(before: ReadonlyMap<string, string>, now: ReadonlyMap<string, string>): readonly string[] {
  return [...now].filter(([id, shape]) => before.get(id) !== shape).map(([id]) => id);
}
