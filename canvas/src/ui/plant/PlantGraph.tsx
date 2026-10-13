import { Background, ReactFlow, useReactFlow, useStore } from "@xyflow/react";
import "@xyflow/react/dist/base.css";
import type { CSSProperties } from "react";
import { useEffect, useMemo } from "react";
import type { Reading } from "../../model/playback";
import type { PlantView } from "../../model/plantView";
import { centreOf, inView, tierOf } from "../../model/viewport";
import type { FlowNode } from "./flow";
import { flowEdges, flowNodes } from "./flow";
import { FrameNode, HallNode, InstanceNode } from "./nodes";
import { cx } from "../cx";
import styles from "./PlantScreen.module.css";

const nodeTypes = { instance: InstanceNode, hall: HallNode, frame: FrameNode };

export interface PlantGraphProps {
  readonly view: PlantView;
  readonly viewKey: string;
  readonly selected: string | null;
  readonly readings?: (id: string) => readonly Reading[];
  readonly onSelect: (id: string | null) => void;
  readonly onOpenHall?: (hall: string) => void;
}

function useFollowSelection(selected: string | null) {
  const flow = useReactFlow<FlowNode>();
  const width = useStore((state) => state.width);
  const height = useStore((state) => state.height);
  useEffect(() => {
    const node = selected === null ? undefined : flow.getInternalNode(selected);
    if (node === undefined) return;
    const box = {
      ...node.internals.positionAbsolute,
      width: node.measured.width ?? 0,
      height: node.measured.height ?? 0,
    };
    const viewport = flow.getViewport();
    if (!inView(box, viewport, { width, height })) {
      const centre = centreOf(box);
      void flow.setCenter(centre.x, centre.y, { zoom: viewport.zoom, duration: 200 });
    }
  }, [flow, selected, width, height]);
}

export function PlantGraph({ view, viewKey, selected, readings, onSelect, onOpenHall }: PlantGraphProps) {
  const nodes = useMemo(() => flowNodes(view, selected, readings ?? null), [view, selected, readings]);
  const edges = useMemo(() => flowEdges(view), [view]);
  const zoom = useStore((state) => state.transform[2]);
  useFollowSelection(selected);
  return (
    <div className={cx(styles.graph, `tier-${tierOf(zoom)}`)} style={{ "--zoom": String(zoom) } as CSSProperties}>
      <ReactFlow
        key={viewKey}
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        nodesDraggable={false}
        nodesConnectable={false}
        onlyRenderVisibleElements
        fitView
        fitViewOptions={{ padding: 0.04, maxZoom: 1 }}
        minZoom={0.05}
        onNodeClick={(_, node: FlowNode) => {
          onSelect(node.type === "frame" ? null : node.id);
        }}
        onNodeDoubleClick={(_, node: FlowNode) => {
          if (node.type === "hall") onOpenHall?.(node.id);
        }}
        onPaneClick={() => {
          onSelect(null);
        }}
        proOptions={{ hideAttribution: true }}
        aria-label="Plant graph"
      >
        <Background gap={16} />
      </ReactFlow>
    </div>
  );
}
