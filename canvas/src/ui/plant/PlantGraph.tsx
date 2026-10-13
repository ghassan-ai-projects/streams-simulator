import { Background, Controls, MiniMap, ReactFlow } from "@xyflow/react";
import type { ReactFlowInstance } from "@xyflow/react";
import "@xyflow/react/dist/base.css";
import { useMemo } from "react";
import type { Reading } from "../../model/playback";
import type { PlantView } from "../../model/plantView";
import type { FlowNode } from "./flow";
import { flowEdges, flowNodes } from "./flow";
import { FrameNode, HallNode, InstanceNode } from "./nodes";
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

const readable = 0.7;

function frame(instance: ReactFlowInstance<FlowNode>): void {
  void instance.fitView({ maxZoom: 1 }).then(() => {
    if (instance.getZoom() < readable) {
      void instance.setViewport({ x: 24, y: 24, zoom: readable });
    }
  });
}

export function PlantGraph({ view, viewKey, selected, readings, onSelect, onOpenHall }: PlantGraphProps) {
  const nodes = useMemo(() => flowNodes(view, selected, readings ?? null), [view, selected, readings]);
  const edges = useMemo(() => flowEdges(view), [view]);
  return (
    <div className={styles.graph}>
      <ReactFlow
        key={viewKey}
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        nodesDraggable={false}
        nodesConnectable={false}
        onlyRenderVisibleElements
        onInit={frame}
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
        <Controls showInteractive={false} position="top-right" />
        <MiniMap
          pannable
          zoomable
          position="bottom-right"
          ariaLabel="Navigator"
          nodeClassName={(node) =>
            node.type === "frame" ? "minimap-frame" : node.selected === true ? "minimap-selected" : ""
          }
        />
      </ReactFlow>
    </div>
  );
}
