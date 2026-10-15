import type { Connection } from "@xyflow/react";
import { Background, ReactFlow, useStore, useUpdateNodeInternals } from "@xyflow/react";
import "@xyflow/react/dist/base.css";
import type { CSSProperties, KeyboardEvent } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import { directionOf, firstOf, nearest } from "../../model/graphKeys";
import type { PlantView } from "../../model/plantView";
import { describeNode, hallCount } from "../../model/plantView";
import { tierOf } from "../../model/viewport";
import { routeView } from "../../model/wireView";
import type { FlowNode, NodeLive } from "./flow";
import { flowEdges, flowNodes } from "./flow";
import { useZoomActions } from "./GraphControls";
import { plainKey } from "../keys";
import { FrameNode, FrameTitleNode, HallNode, InstanceNode } from "./nodes";
import { WireEdge } from "./WireEdge";
import { useFitted, useFollowSelection } from "./viewportFit";
import styles from "./PlantScreen.module.css";

const nodeTypes = { instance: InstanceNode, hall: HallNode, frame: FrameNode, frameTitle: FrameTitleNode };
const edgeTypes = { wire: WireEdge };

export interface PlantGraphProps {
  readonly view: PlantView;
  readonly viewKey: string;
  readonly selected: string | null;
  readonly live?: (id: string) => NodeLive;
  readonly onSelect: (id: string | null) => void;
  readonly onOpenHall?: (hall: string) => void;
  readonly onEscape?: () => void;
  readonly minFitZoom?: number;
  readonly focusZoom?: number;
  readonly flagged?: ReadonlySet<string>;
  readonly fitKey?: string;
  readonly onConnect?: (from: string, to: string) => void;
}

interface Keys {
  readonly active: string | null;
  readonly announcement: string;
  readonly onKeyDown: (event: KeyboardEvent<HTMLDivElement>) => void;
  readonly onFocus: () => void;
  readonly onPointerDown: () => void;
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

function useSelectionMoves(view: PlantView, selected: string | null, setActive: (id: string) => void) {
  const [seen, setSeen] = useState(selected);
  if (seen !== selected) {
    setSeen(selected);
    if (selected !== null && view.nodes.some((entry) => entry.id === selected)) setActive(selected);
  }
}

function useViewActive(view: PlantView, viewKey: string) {
  const [active, setActive] = useState<string | null>(null);
  const [shownKey, setShownKey] = useState(viewKey);
  const [entered, setEntered] = useState("");
  if (shownKey !== viewKey) {
    const back = view.nodes.some((entry) => entry.id === shownKey);
    setShownKey(viewKey);
    setActive(back ? shownKey : firstOf(view.layout.nodes));
    setEntered(back ? "" : `${viewKey}, ${hallCount(view, viewKey)}. `);
  }
  const current = active !== null && view.nodes.some((entry) => entry.id === active) ? active : null;
  const move = (id: string | null) => {
    setEntered("");
    setActive(id);
  };
  return { active: current, setActive: move, entered };
}

function announce(
  view: PlantView,
  active: string | null,
  selected: string | null,
  live: ((id: string) => NodeLive) | undefined,
): string {
  const node = view.nodes.find((entry) => entry.id === active);
  if (node === undefined) return "";
  const now = live?.(node.id);
  return describeNode(node, node.id === selected, now?.status ?? null, now?.count ?? 0, now?.fault ?? "");
}

function enterHall(id: string, onSelect: (id: string | null) => void, onOpenHall?: (hall: string) => void) {
  if (onOpenHall === undefined) {
    onSelect(id);
    return;
  }
  onSelect(null);
  onOpenHall(id);
}

function useGraphFocus(
  view: PlantView,
  active: string | null,
  selected: string | null,
  setActive: (id: string | null) => void,
) {
  const pointed = useRef(false);
  const onPointerDown = () => {
    pointed.current = true;
  };
  const onFocus = () => {
    const fromPointer = pointed.current;
    pointed.current = false;
    if (active !== null) return;
    setActive(fromPointer ? selected : (selected ?? firstOf(view.layout.nodes)));
  };
  return { onPointerDown, onFocus };
}

function useGraphKeys({ view, viewKey, selected, live, onSelect, onOpenHall, onEscape }: PlantGraphProps): Keys {
  const { active, setActive, entered } = useViewActive(view, viewKey);
  useSelectionMoves(view, selected, setActive);
  const zoomKey = useZoomKeys();
  const said = announce(view, active, selected, live);
  const announcement = said === "" ? "" : entered + said;
  const activate = (id: string) => {
    if (view.nodes.find((entry) => entry.id === id)?.instance !== null) onSelect(id);
    else enterHall(id, onSelect, onOpenHall);
  };
  const move = (key: string): boolean => {
    const direction = directionOf(key);
    if (direction === null) return false;
    setActive(active === null ? firstOf(view.layout.nodes) : (nearest(active, direction, view.layout.nodes) ?? active));
    return true;
  };
  const choose = (key: string): boolean => {
    if ((key !== "Enter" && key !== " ") || active === null) return false;
    if (key === " ") onSelect(active);
    else activate(active);
    return true;
  };
  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (!plainKey(event)) return;
    if (move(event.key) || choose(event.key) || zoomKey(event.key)) {
      event.preventDefault();
    } else if (event.key === "Escape") {
      if (selected !== null) onSelect(null);
      else onEscape?.();
    }
  };
  const focus = useGraphFocus(view, active, selected, setActive);
  const point = (id: string | null) => {
    if (id !== null) setActive(id);
  };
  return { active, announcement, onKeyDown, onFocus: focus.onFocus, onPointerDown: focus.onPointerDown, point };
}

const graphNames = {
  "aria-roledescription": "plant graph",
  "aria-label": "Plant graph",
  "aria-describedby": "graph-keys",
  tabIndex: 0,
} as const;

const fixedProps = {
  nodeTypes,
  edgeTypes,
  nodesDraggable: false,
  nodesFocusable: false,
  edgesFocusable: false,
  disableKeyboardA11y: true,
  zoomOnDoubleClick: false,
  onlyRenderVisibleElements: true,
  minZoom: 0.05,
  proOptions: { hideAttribution: true },
} as const;

function connection(onConnect: PlantGraphProps["onConnect"]) {
  if (onConnect === undefined) return { nodesConnectable: false };
  return {
    nodesConnectable: true,
    onConnect: (made: Connection) => {
      if (made.sourceHandle && made.targetHandle) {
        onConnect(`${made.source}.${made.sourceHandle}`, `${made.target}.${made.targetHandle}`);
      }
    },
  };
}

type FlowProps = PlantGraphProps & { readonly active: string | null; readonly keys: Keys };

function useFlowContent({ view, selected, live, flagged, onConnect, active }: FlowProps) {
  const editable = onConnect !== undefined;
  const nodes = useMemo(
    () => flowNodes(view, selected, live ?? null, active, flagged ?? null, editable),
    [view, selected, live, active, flagged, editable],
  );
  const zoom = useStore((state) => state.transform[2]);
  const tier = tierOf(zoom);
  const shown = view.nodes.some((entry) => entry.id === selected) ? selected : null;
  const review = live !== undefined;
  const routes = useMemo(() => routeView(view, tier, review), [view, tier, review]);
  const edges = useMemo(() => flowEdges(view, routes, shown), [view, routes, shown]);
  return { nodes, edges, zoom };
}

function useRemeasuredHandles(nodes: readonly FlowNode[], tier: string) {
  const update = useUpdateNodeInternals();
  useEffect(() => {
    update(nodes.map((node) => node.id));
  }, [update, nodes, tier]);
}

function Flow(props: FlowProps) {
  const { onSelect, onOpenHall, onConnect, keys } = props;
  const { nodes, edges, zoom } = useFlowContent(props);
  useRemeasuredHandles(nodes, tierOf(zoom));
  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      {...fixedProps}
      {...connection(onConnect)}
      className={`tier-${tierOf(zoom)}`}
      style={{ "--zoom": String(zoom) } as CSSProperties}
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
    </ReactFlow>
  );
}

export function PlantGraph(props: PlantGraphProps) {
  const keys = useGraphKeys(props);
  useFitted(props.fitKey ?? props.viewKey, props.minFitZoom ?? 0.05);
  useFollowSelection(keys.active ?? props.selected, props.focusZoom ?? 0);
  return (
    <div className={styles.graph} onPointerDown={keys.onPointerDown}>
      <p id="graph-keys" className="sr-only">
        Arrow keys move between instances, Space selects, Enter selects an instance or opens a hall, Escape clears the
        selection, question mark lists every key. The outline lists every instance.
      </p>
      <p id="graph-said" className="sr-only" aria-live="polite">
        {keys.announcement}
      </p>
      <Flow
        {...props}
        keys={keys}
        active={keys.active}
        onSelect={(id) => {
          keys.point(id);
          props.onSelect(id);
        }}
      />
    </div>
  );
}
