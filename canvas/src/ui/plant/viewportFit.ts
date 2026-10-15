import type { ReactFlowInstance } from "@xyflow/react";
import { useReactFlow, useStore, useStoreApi } from "@xyflow/react";
import { useEffect, useRef } from "react";
import type { Viewport } from "../../model/viewport";
import { centreOf, followZoom, openingViewport, sameViewport, steppedViewport } from "../../model/viewport";
import type { FlowNode } from "./flow";

interface Size {
  readonly width: number;
  readonly height: number;
}

const opened = new WeakMap<object, Viewport>();

export function openingOf(store: object): Viewport | null {
  return opened.get(store) ?? null;
}

export function fittedViewport(flow: ReactFlowInstance<FlowNode>, screen: Size, minZoom: number): Viewport {
  const bounds = flow.getNodesBounds(flow.getNodes());
  return openingViewport(bounds, screen, { pad: 16, headroom: 24, minZoom, maxZoom: 1 });
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

function useScreen(): Size & { readonly ready: boolean } {
  const ready = useStore((state) => state.nodes.length > 0);
  const width = useStore((state) => state.width);
  const height = useStore((state) => state.height);
  return { ready: ready && width > 0, width, height };
}

export function useFitted(fitKey: string, minZoom: number) {
  const flow = useReactFlow<FlowNode>();
  const store = useStoreApi<FlowNode>();
  const { ready, width, height } = useScreen();
  const last = useRef<{ key: string; viewport: Viewport | null }>({ key: "", viewport: null });
  useEffect(() => {
    if (!ready) return;
    const { key, viewport } = last.current;
    if (key === fitKey && (viewport === null || !sameViewport(viewport, flow.getViewport()))) return;
    const fitted = fittedViewport(flow, { width, height }, minZoom);
    last.current = { key: fitKey, viewport: fitted };
    opened.set(store, fitted);
    void flow.setViewport(fitted);
  }, [flow, store, ready, width, height, fitKey, minZoom]);
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
