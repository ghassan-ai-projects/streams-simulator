import { Background, Panel, ReactFlow, useStore, useUpdateNodeInternals } from "@xyflow/react";
import "@xyflow/react/dist/base.css";
import type { CSSProperties } from "react";
import { useEffect, useMemo } from "react";
import { routeView, screenPixel, textFloorPx, tierOf } from "../../../model/graph";
import type { FlowNode } from "./flow";
import { flowEdges, flowNodes } from "./flow";
import { connection, fixedProps, graphNames, graphWords } from "./flowConfig";
import { enterHall } from "./enterHall";
import { PanCue } from "./PanCue";
import styles from "./graph.module.css";
import { tierClass } from "./tierClass";
import type { PlantGraphProps } from "./types";
import type { Keys } from "./useGraphKeys";

export type FlowProps = PlantGraphProps & { readonly active: string | null; readonly keys: Keys };

function useFlowContent({ view, selected, live, flagged, onConnect, active, marks }: FlowProps) {
  const editable = onConnect !== undefined;
  const nodes = useMemo(
    () =>
      flowNodes(view, selected, live ?? null, active, flagged ?? null, editable ? "connect" : "inspect", marks ?? null),
    [view, selected, live, active, flagged, editable, marks],
  );
  const zoom = useStore((state) => state.transform[2]);
  const tier = tierOf(zoom);
  const shown = view.nodes.some((entry) => entry.id === selected) ? selected : null;
  const review = live !== undefined;
  const pixel = screenPixel(zoom);
  const routes = useMemo(() => routeView(view, tier, review ? "rows" : "pins", pixel), [view, tier, review, pixel]);
  const edges = useMemo(() => flowEdges(view, routes, shown), [view, routes, shown]);
  return { nodes, edges, zoom };
}

function useRemeasuredHandles(nodes: readonly FlowNode[], tier: string) {
  const update = useUpdateNodeInternals();
  useEffect(() => {
    update(nodes.map((node) => node.id));
  }, [update, nodes, tier]);
}

export function FlowCanvas(props: FlowProps) {
  const { onSelect, onOpenHall, onConnect, keys } = props;
  const { nodes, edges, zoom } = useFlowContent(props);
  useRemeasuredHandles(nodes, tierOf(zoom));
  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      {...fixedProps}
      {...connection(onConnect)}
      ariaLabelConfig={onConnect === undefined ? graphWords.select : graphWords.edit}
      className={tierClass[tierOf(zoom)]}
      data-zoom-tier={tierOf(zoom)}
      style={{ "--zoom": String(zoom), "--floor": `${String(textFloorPx / zoom)}px` } as CSSProperties}
      {...graphNames}
      onKeyDown={keys.onKeyDown}
      onFocus={keys.onFocus}
      onNodeClick={(_, node: FlowNode) => {
        onSelect(node.type === "frame" || node.type === "frameTitle" ? null : node.id);
      }}
      onNodeDoubleClick={(_, node: FlowNode) => {
        if (node.type === "hall") enterHall(node.id, onSelect, onOpenHall);
      }}
      onPaneClick={() => {
        onSelect(null);
      }}
    >
      <Background gap={16} />
      {props.panCue === true && <PanCue />}
      {typeof props.notice === "string" && (
        <Panel position="bottom-center" role="alert" className={styles.notice}>
          {props.notice}
        </Panel>
      )}
    </ReactFlow>
  );
}
