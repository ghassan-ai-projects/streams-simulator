import { describe, expect, it } from "vitest";
import type { Point, Rect } from "./types";
import { routeWires, wireClearance } from "./wireRoute";

const node = (x: number, y: number, width = 100, height = 60): Rect => ({ x, y, width, height });

function cuts(points: readonly Point[], box: Rect, by: number): boolean {
  const left = box.x - by;
  const right = box.x + box.width + by;
  const top = box.y - by;
  const bottom = box.y + box.height + by;
  return points.slice(1).some((to, index) => {
    const from = points[index] ?? to;
    const x0 = Math.min(from.x, to.x);
    const x1 = Math.max(from.x, to.x);
    const y0 = Math.min(from.y, to.y);
    const y1 = Math.max(from.y, to.y);
    return x1 > left && x0 < right && y1 > top && y0 < bottom;
  });
}

function orthogonal(points: readonly Point[]): boolean {
  return points.slice(1).every((to, index) => {
    const from = points[index] ?? to;
    return from.x === to.x || from.y === to.y;
  });
}

type WireScene = Parameters<typeof routeWires>[0];

function sceneOf(partial: Partial<WireScene>): WireScene {
  return { links: [], nodes: [], titles: [], ...partial };
}

describe("wire router", () => {
  it("joins a source's right edge to a target's left edge with a horizontal stub at each end", () => {
    const nodes = [node(0, 0), node(300, 100)];
    const link = { id: "a", from: { x: 100, y: 30 }, to: { x: 300, y: 130 } };
    const route = routeWires(sceneOf({ links: [link], nodes })).get("a") ?? [];
    expect(route[0]).toEqual(link.from);
    expect(route.at(-1)).toEqual(link.to);
    expect(route[1]?.y).toBe(30);
    expect(route.at(-2)?.y).toBe(130);
    expect(orthogonal(route)).toBe(true);
  });

  it("goes round a node in the way, keeping its inflated rectangle clear", () => {
    const blocker = node(150, 0, 100, 400);
    const nodes = [node(0, 150), node(400, 150), blocker];
    const link = { id: "a", from: { x: 100, y: 180 }, to: { x: 400, y: 180 } };
    const route = routeWires(sceneOf({ links: [link], nodes })).get("a") ?? [];
    expect(orthogonal(route)).toBe(true);
    expect(cuts(route, blocker, wireClearance.node - 1)).toBe(false);
  });

  it("enters a target that lies left of its source on the target's left edge without crossing any node", () => {
    const source = node(400, 0);
    const target = node(0, 200);
    const link = { id: "a", from: { x: 500, y: 30 }, to: { x: 0, y: 230 } };
    const route = routeWires(sceneOf({ links: [link], nodes: [source, target] })).get("a") ?? [];
    expect(route.at(-2)).toMatchObject({ x: -wireClearance.stub, y: 230 });
    expect(orthogonal(route)).toBe(true);
    expect(cuts(route.slice(1), source, wireClearance.node - 1)).toBe(false);
    expect(cuts(route.slice(1, -1), target, 0)).toBe(false);
  });

  it("keeps clear of an area title box", () => {
    const title = { x: 150, y: 100, width: 200, height: 28 };
    const nodes = [node(0, 90), node(500, 90)];
    const link = { id: "a", from: { x: 100, y: 120 }, to: { x: 500, y: 120 } };
    const clearance = [8];
    const route =
      routeWires(sceneOf({ links: [link], nodes, titles: [title], titleClearances: clearance })).get("a") ?? [];
    expect(cuts(route, title, 8 - 1)).toBe(false);
  });

  it("is the same on every run and independent of the links' order", () => {
    const nodes = [node(0, 0), node(0, 100), node(300, 0), node(300, 100)];
    const links = [
      { id: "b", from: { x: 100, y: 30 }, to: { x: 300, y: 130 } },
      { id: "a", from: { x: 100, y: 130 }, to: { x: 300, y: 30 } },
    ];
    const first = routeWires(sceneOf({ links, nodes }));
    const second = routeWires(sceneOf({ links: [...links].reverse(), nodes }));
    expect(second.get("a")).toEqual(first.get("a"));
    expect(second.get("b")).toEqual(first.get("b"));
  });

  it("does not run two wires over each other along the same corridor", () => {
    const nodes = [node(0, 0), node(0, 100), node(300, 0), node(300, 100)];
    const links = [
      { id: "a", from: { x: 100, y: 20 }, to: { x: 300, y: 120 } },
      { id: "b", from: { x: 100, y: 40 }, to: { x: 300, y: 140 } },
    ];
    const routes = routeWires(sceneOf({ links, nodes }));
    const verticals = (id: string) =>
      (routes.get(id) ?? []).flatMap((point, index, all) => {
        const next = all[index + 1];
        return next?.x === point.x && next.y !== point.y ? [point.x] : [];
      });
    expect(verticals("a").filter((x) => verticals("b").includes(x))).toEqual([]);
  });

  it("leaves a node's bottom edge and enters another's top edge in the same column, round what lies between", () => {
    const upper = node(0, 0);
    const middle = node(0, 100);
    const lower = node(0, 200);
    const link = {
      id: "a",
      from: { x: 50, y: 60 },
      to: { x: 50, y: 200 },
      leaves: "down" as const,
      enters: "up" as const,
    };
    const route = routeWires(sceneOf({ links: [link], nodes: [upper, middle, lower] })).get("a") ?? [];
    expect(route[0]).toEqual(link.from);
    expect(route[1]?.x).toBe(50);
    expect(route[1]?.y).toBeGreaterThanOrEqual(76);
    expect(route.at(-2)?.x).toBe(50);
    expect(route.at(-2)?.y).toBeLessThanOrEqual(184);
    expect(orthogonal(route)).toBe(true);
    expect(cuts(route.slice(1, -1), middle, wireClearance.node - 1)).toBe(false);
  });
});
