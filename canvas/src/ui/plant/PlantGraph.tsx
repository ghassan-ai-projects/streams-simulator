import { Background, ReactFlow, useReactFlow, useStore } from "@xyflow/react";
import "@xyflow/react/dist/base.css";
import type { CSSProperties, KeyboardEvent } from "react";
import { useEffect, useMemo, useState } from "react";
import { directionOf, firstOf, nearest } from "../../model/graphKeys";
import type { PlantView } from "../../model/plantView";
import { describeNode } from "../../model/plantView";
import { anchorsOnly, centreOf, inView, tierOf } from "../../model/viewport";
import type { FlowNode, NodeLive } from "./flow";
import { flowEdges, flowNodes } from "./flow";
import { useZoomActions } from "./GraphControls";
import { FrameNode, HallNode, InstanceNode } from "./nodes";
import { cx } from "../cx";
import styles from "./PlantScreen.module.css";

const nodeTypes = { instance: InstanceNode, hall: HallNode, frame: FrameNode };

export interface PlantGraphProps {
  readonly view: PlantView;
  readonly viewKey: string;
  readonly selected: string | null;
  readonly live?: (id: string) => NodeLive;
  readonly onSelect: (id: string | null) => void;
  readonly onOpenHall?: (hall: string) => void;
  readonly onEscape?: () => void;
  readonly minFitZoom?: number;
}

function useFollowSelection(selected: string | null) {
  const flow = useReactFlow<FlowNode>();
  const zoomTo = useZoomActions();
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
      zoomTo.centre(centre.x, centre.y, viewport.zoom);
    }
  }, [flow, zoomTo, selected, width, height]);
}

interface Keys {
  readonly active: string | null;
  readonly announcement: string;
  readonly onKeyDown: (event: KeyboardEvent<HTMLDivElement>) => void;
  readonly onFocus: () => void;
  readonly point: (id: string | null) => void;
}

const zoomKeys: Readonly<Record<string, "in" | "out" | "fit">> = { "+": "in", "=": "in", "-": "out", "0": "fit" };

function useZoomKeys() {
  const zoomTo = useZoomActions();
  return (key: string): boolean => {
    const action = zoomKeys[key];
    if (action !== undefined) zoomTo[action]();
    return action !== undefined;
  };
}

function useViewActive(view: PlantView, viewKey: string) {
  const [active, setActive] = useState<string | null>(null);
  const [shownKey, setShownKey] = useState(viewKey);
  if (shownKey !== viewKey) {
    setShownKey(viewKey);
    setActive(view.nodes.some((entry) => entry.id === shownKey) ? shownKey : firstOf(view.layout.nodes));
  }
  const current = active !== null && view.nodes.some((entry) => entry.id === active) ? active : null;
  return [current, setActive] as const;
}

function useGraphKeys({ view, viewKey, selected, live, onSelect, onOpenHall, onEscape }: PlantGraphProps): Keys {
  const [active, setActive] = useViewActive(view, viewKey);
  const zoomKey = useZoomKeys();
  const node = view.nodes.find((entry) => entry.id === active);
  const announcement =
    node === undefined ? "" : describeNode(node, node.id === selected, live?.(node.id).status ?? null);
  const activate = (id: string) => {
    if (view.nodes.find((entry) => entry.id === id)?.instance === null) onOpenHall?.(id);
    else onSelect(id);
  };
  const move = (key: string): boolean => {
    const direction = directionOf(key);
    if (direction === null) return false;
    setActive(active === null ? firstOf(view.layout.nodes) : (nearest(active, direction, view.layout.nodes) ?? active));
    return true;
  };
  const choose = (key: string): boolean => {
    if ((key !== "Enter" && key !== " ") || active === null) return false;
    activate(active);
    return true;
  };
  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (move(event.key) || choose(event.key) || zoomKey(event.key)) {
      event.preventDefault();
    } else if (event.key === "Escape") {
      if (selected !== null) onSelect(null);
      else onEscape?.();
    }
  };
  const onFocus = () => {
    if (active === null) setActive(selected ?? firstOf(view.layout.nodes));
  };
  const point = (id: string | null) => {
    if (id !== null) setActive(id);
  };
  return { active, announcement, onKeyDown, onFocus, point };
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
  minZoom: 0.05,
  proOptions: { hideAttribution: true },
} as const;

function Flow({
  view,
  viewKey,
  selected,
  live,
  onSelect,
  onOpenHall,
  active,
  minFitZoom,
}: PlantGraphProps & { readonly active: string | null }) {
  const nodes = useMemo(() => flowNodes(view, selected, live ?? null, active), [view, selected, live, active]);
  const zoom = useStore((state) => state.transform[2]);
  const anchors = anchorsOnly(tierOf(zoom));
  const edges = useMemo(() => flowEdges(view, anchors), [view, anchors]);
  return (
    <ReactFlow
      key={viewKey}
      nodes={nodes}
      edges={edges}
      {...fixedProps}
      fitViewOptions={{ padding: 0.04, maxZoom: 1, minZoom: minFitZoom ?? 0.05 }}
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
      <Flow
        {...props}
        active={keys.active}
        onSelect={(id) => {
          keys.point(id);
          props.onSelect(id);
        }}
      />
    </div>
  );
}
