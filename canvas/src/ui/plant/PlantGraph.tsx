import { Background, ReactFlow, useReactFlow, useStore } from "@xyflow/react";
import "@xyflow/react/dist/base.css";
import type { CSSProperties, KeyboardEvent } from "react";
import { useEffect, useMemo, useState } from "react";
import type { Reading } from "../../model/playback";
import { directionOf, firstOf, nearest } from "../../model/graphKeys";
import type { PlantView } from "../../model/plantView";
import { describeNode } from "../../model/plantView";
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
  readonly onEscape?: () => void;
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

interface Keys {
  readonly active: string | null;
  readonly announcement: string;
  readonly onKeyDown: (event: KeyboardEvent<HTMLDivElement>) => void;
  readonly onFocus: () => void;
}

function useGraphKeys({ view, selected, onSelect, onOpenHall, onEscape }: PlantGraphProps): Keys {
  const [active, setActive] = useState<string | null>(null);
  const node = view.nodes.find((entry) => entry.id === active);
  const announcement = node === undefined ? "" : describeNode(node, node.id === selected);
  const activate = (id: string) => {
    if (view.nodes.find((entry) => entry.id === id)?.instance === null) onOpenHall?.(id);
    else onSelect(id);
  };
  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    const direction = directionOf(event.key);
    if (direction !== null) {
      event.preventDefault();
      setActive(
        active === null ? firstOf(view.layout.nodes) : (nearest(active, direction, view.layout.nodes) ?? active),
      );
    } else if ((event.key === "Enter" || event.key === " ") && active !== null) {
      event.preventDefault();
      activate(active);
    } else if (event.key === "Escape") {
      if (selected !== null) onSelect(null);
      else onEscape?.();
    }
  };
  const onFocus = () => {
    setActive((current) => current ?? selected ?? firstOf(view.layout.nodes));
  };
  return { active, announcement, onKeyDown, onFocus };
}

const fixedProps = {
  nodeTypes,
  nodesDraggable: false,
  nodesConnectable: false,
  nodesFocusable: false,
  edgesFocusable: false,
  disableKeyboardA11y: true,
  onlyRenderVisibleElements: true,
  fitView: true,
  fitViewOptions: { padding: 0.04, maxZoom: 1 },
  minZoom: 0.05,
  proOptions: { hideAttribution: true },
} as const;

function Flow({
  view,
  viewKey,
  selected,
  readings,
  onSelect,
  onOpenHall,
  active,
}: PlantGraphProps & { readonly active: string | null }) {
  const nodes = useMemo(() => flowNodes(view, selected, readings ?? null, active), [view, selected, readings, active]);
  const edges = useMemo(() => flowEdges(view), [view]);
  return (
    <ReactFlow
      key={viewKey}
      nodes={nodes}
      edges={edges}
      {...fixedProps}
      onNodeClick={(_, node: FlowNode) => {
        onSelect(node.type === "frame" ? null : node.id);
      }}
      onNodeDoubleClick={(_, node: FlowNode) => {
        if (node.type === "hall") onOpenHall?.(node.id);
      }}
      onPaneClick={() => {
        onSelect(null);
      }}
    >
      <Background gap={16} />
    </ReactFlow>
  );
}

export function PlantGraph(props: PlantGraphProps) {
  const zoom = useStore((state) => state.transform[2]);
  const keys = useGraphKeys(props);
  useFollowSelection(keys.active ?? props.selected);
  return (
    <div
      className={cx(styles.graph, `tier-${tierOf(zoom)}`)}
      style={{ "--zoom": String(zoom) } as CSSProperties}
      role="application"
      aria-roledescription="plant graph"
      aria-label="Plant graph"
      aria-describedby="graph-keys"
      tabIndex={0}
      onKeyDown={keys.onKeyDown}
      onFocus={keys.onFocus}
    >
      <p id="graph-keys" className="sr-only">
        Arrow keys move between instances, Enter selects or opens a hall, Escape clears the selection, question mark
        lists every key. The outline lists every instance.
      </p>
      <p className="sr-only" aria-live="polite">
        {keys.announcement}
      </p>
      <Flow {...props} active={keys.active} />
    </div>
  );
}
