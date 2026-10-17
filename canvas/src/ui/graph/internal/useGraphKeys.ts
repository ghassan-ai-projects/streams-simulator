import type { KeyboardEvent } from "react";
import { useRef } from "react";
import { describeNode, directionOf, firstOf, nearest } from "../../../model/graph";
import type { PlantView } from "../../../model/graph";
import { plainKey } from "../../primitives";
import { enterHall } from "./enterHall";
import { useZoomActions } from "./GraphControls";
import type { NodeLive, PlantGraphProps } from "./types";
import { useSelectionMoves, useViewActive } from "./useViewActive";

export interface Keys {
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

function announce(
  view: PlantView,
  active: string | null,
  selected: string | null,
  live: ((id: string) => NodeLive) | undefined,
): string {
  const node = view.nodes.find((entry) => entry.id === active);
  if (node === undefined) return "";
  const now = live?.(node.id);
  return describeNode(node, {
    selected: node.id === selected,
    status: now?.status ?? null,
    count: now?.count ?? 0,
    fault: now?.fault ?? "",
  });
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

export function useGraphKeys({ view, viewKey, selected, live, onSelect, onOpenHall, onEscape }: PlantGraphProps): Keys {
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
