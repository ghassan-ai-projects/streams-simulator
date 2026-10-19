import type { ReactFlowInstance } from "@xyflow/react";
import { useReactFlow, useStore } from "@xyflow/react";
import { useEffect, useRef } from "react";
import type { Viewport } from "../../../model/graph";
import {
  boundsWithTitles,
  centreOf,
  followZoom,
  openingViewport,
  sameViewport,
  screenPixel,
  steppedViewport,
  tierOf,
} from "../../../model/graph";
import type { FlowNode } from "./flow";

export const lowestZoom = 0.05;

interface Size {
  readonly width: number;
  readonly height: number;
}

function titleTopsOf(nodes: readonly FlowNode[]): number[] {
  return nodes.flatMap((node) => (node.type === "frameTitle" ? [node.position.y] : []));
}

export function fittedViewport(
  flow: ReactFlowInstance<FlowNode>,
  screen: Size,
  minZoom: number,
  anchor: string | null = null,
): Viewport {
  const nodes = flow.getNodes();
  const bounds = flow.getNodesBounds(nodes);
  const limits = { pad: 16, headroom: 24, minZoom, maxZoom: 1 };
  const focus = anchor !== null && flow.getNode(anchor) !== undefined ? flow.getNodesBounds([anchor]) : null;
  const plain = openingViewport(bounds, screen, limits, focus);
  if (tierOf(plain.zoom) === "full") return plain;
  const lifted = boundsWithTitles(bounds, titleTopsOf(nodes), screenPixel(plain.zoom));
  return openingViewport(lifted, screen, limits, focus);
}

interface Limits extends Size {
  readonly minZoom: number;
  readonly maxZoom: number;
}

export function steppedZoom(flow: ReactFlowInstance<FlowNode>, screen: Limits, factor: number): Viewport {
  const current = flow.getViewport();
  const zoom = Math.min(screen.maxZoom, Math.max(screen.minZoom, current.zoom * factor));
  const bounds = flow.getNodesBounds(flow.getNodes());
  return steppedViewport(current, screen, zoom, bounds, 16);
}

function useViewportSize(): Size & { readonly ready: boolean } {
  const ready = useStore((state) => state.nodes.length > 0);
  const width = useStore((state) => state.width);
  const height = useStore((state) => state.height);
  return { ready: ready && width > 0, width, height };
}

export function useFitted(fitKey: string, minZoom: number, anchor: string | null) {
  const flow = useReactFlow<FlowNode>();
  const { ready, width, height } = useViewportSize();
  const last = useRef<{ key: string; viewport: Viewport | null }>({ key: "", viewport: null });
  useEffect(() => {
    if (!ready) return;
    const { key, viewport } = last.current;
    if (key === fitKey && (viewport === null || !sameViewport(viewport, flow.getViewport()))) return;
    const fitted = fittedViewport(flow, { width, height }, minZoom, anchor);
    last.current = { key: fitKey, viewport: fitted };
    void flow.setViewport(fitted).then(() => {
      if (last.current.key === fitKey && last.current.viewport === fitted) {
        last.current = { key: fitKey, viewport: flow.getViewport() };
      }
    });
  }, [flow, ready, width, height, fitKey, minZoom, anchor]);
}

export function useFollowSelection(selected: string | null, focusZoom: number) {
  const flow = useReactFlow<FlowNode>();
  const width = useStore((state) => state.width);
  const height = useStore((state) => state.height);
  useEffect(() => {
    const node = selected === null ? undefined : flow.getInternalNode(selected);
    if (node === undefined) return;
    const box = {
      ...node.internals.positionAbsolute,
      width: node.width ?? node.measured.width ?? 0,
      height: node.height ?? node.measured.height ?? 0,
    };
    const zoom = followZoom(box, flow.getViewport(), { width, height }, focusZoom);
    if (zoom === null) return;
    const centre = centreOf(box);
    void flow.setCenter(centre.x, centre.y, { zoom });
  }, [flow, selected, focusZoom, width, height]);
}
