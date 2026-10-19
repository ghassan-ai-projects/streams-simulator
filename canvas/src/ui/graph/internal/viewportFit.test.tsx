import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Viewport } from "../../../model/graph";
import { useFitted } from "./viewportFit";

interface FakeScreen {
  nodes: { id: string }[];
  width: number;
  height: number;
}

const fakeScreen: FakeScreen = { nodes: [], width: 0, height: 0 };
const minFlowZoom = 0.5;
let applied: Viewport = { x: 0, y: 0, zoom: 1 };

const fakeFlow = {
  getNodes: () => fakeScreen.nodes,
  getNodesBounds: () => ({ x: 0, y: 0, width: 6000, height: 3000 }),
  getNode: () => undefined,
  getViewport: () => applied,
  setViewport: (viewport: Viewport) => {
    applied = { ...viewport, zoom: Math.max(viewport.zoom, minFlowZoom) };
    return Promise.resolve(true);
  },
};

vi.mock("@xyflow/react", () => ({
  useReactFlow: () => fakeFlow,
  useStore: (selector: (state: FakeScreen) => unknown) => selector(fakeScreen),
}));

function setScreen(width: number, height: number): void {
  fakeScreen.nodes = [{ id: "frame:wide" }];
  fakeScreen.width = width;
  fakeScreen.height = height;
}

function mountFit() {
  return renderHook(() => {
    useFitted("fit", 0.05, null);
  });
}

async function settle(): Promise<void> {
  await act(async () => {
    await Promise.resolve();
  });
}

describe("the fit of the plant canvas", () => {
  beforeEach(() => {
    applied = { x: 0, y: 0, zoom: 1 };
  });

  it("fits again when the pane changes height after React Flow clamped the fitted zoom", async () => {
    setScreen(1200, 800);
    const view = mountFit();
    await settle();
    const fitted = applied;
    expect(fitted.zoom).toBe(minFlowZoom);
    setScreen(1200, 400);
    view.rerender();
    await settle();
    expect(applied.y).not.toBe(fitted.y);
  });

  it("does not fit over a view the user has moved, even when the pane changes height", async () => {
    setScreen(1200, 800);
    const view = mountFit();
    await settle();
    applied = { x: 7, y: 9, zoom: minFlowZoom };
    setScreen(1200, 400);
    view.rerender();
    await settle();
    expect(applied).toEqual({ x: 7, y: 9, zoom: minFlowZoom });
  });
});
