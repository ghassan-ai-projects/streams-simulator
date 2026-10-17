import type { Point, Rect } from "./types";
import type { Grid } from "./wireGrid";
import { buildGrid, lanes, nearestIndex } from "./wireGrid";
import type { Occupancy, Turns } from "./wireSearch";
import { emptyOccupancy, findPath, occupy } from "./wireSearch";

export type Side = "right" | "left" | "down" | "up";

export interface WireLink {
  readonly id: string;
  readonly from: Point;
  readonly to: Point;
  readonly leaves?: Side;
  readonly enters?: Side;
  readonly stub?: number;
  readonly corridor?: Rect;
}

interface WireScene {
  readonly links: readonly WireLink[];
  readonly nodes: readonly Rect[];
  readonly titles: readonly Rect[];
  readonly titleClearances?: readonly number[];
  readonly frames?: readonly Rect[];
  readonly pixel?: number;
}

const frameMargin = 12;

export const wireClearance = { node: 8, stub: 16, lane: 6 } as const;

const laneCount = { x: 4, y: 2 } as const;
const gridLimit = 400_000;

function inflate(box: Rect, by: number): Rect {
  return { x: box.x - by, y: box.y - by, width: box.width + 2 * by, height: box.height + 2 * by };
}

const outward: Readonly<Record<Side, Point>> = {
  right: { x: 1, y: 0 },
  left: { x: -1, y: 0 },
  down: { x: 0, y: 1 },
  up: { x: 0, y: -1 },
};

function horizontal(side: Side): boolean {
  return side === "left" || side === "right";
}

function stubs(link: WireLink): { readonly start: Point; readonly goal: Point } {
  const length = link.stub ?? wireClearance.stub;
  const from = outward[link.leaves ?? "right"];
  const to = outward[link.enters ?? "left"];
  return {
    start: { x: link.from.x + from.x * length, y: link.from.y + from.y * length },
    goal: { x: link.to.x + to.x * length, y: link.to.y + to.y * length },
  };
}

function obstaclesOf(scene: WireScene, titleClearance: number | null): Rect[] {
  const titles = titleClearance === null ? [] : scene.titles.map((box) => inflate(box, titleClearance));
  return [...scene.nodes.map((box) => inflate(box, wireClearance.node)), ...titles];
}

function gridOf(scene: WireScene, obstacles: readonly Rect[]): Grid {
  const pixel = scene.pixel ?? 1;
  const lane = wireClearance.lane * pixel;
  const band = frameMargin * pixel;
  const frames = scene.frames ?? [];
  const xs = obstacles.flatMap((box) => [
    ...lanes(box.x, -1, laneCount.x, lane),
    ...lanes(box.x + box.width, 1, laneCount.x, lane),
  ]);
  const ys = obstacles.flatMap((box) => [
    ...lanes(box.y, -1, laneCount.y, lane),
    ...lanes(box.y + box.height, 1, laneCount.y, lane),
  ]);
  for (const link of scene.links) {
    const { start, goal } = stubs(link);
    xs.push(start.x, goal.x);
    ys.push(start.y, goal.y);
  }
  for (const frame of frames) {
    xs.push(frame.x - band, frame.x + band, frame.x + frame.width - band, frame.x + frame.width + band);
    ys.push(frame.y - band, frame.y + band, frame.y + frame.height - band, frame.y + frame.height + band);
  }
  return buildGrid(obstacles, xs, ys, { frames, band });
}

function collinear(a: Point, b: Point, c: Point): boolean {
  return (a.x === b.x && b.x === c.x) || (a.y === b.y && b.y === c.y);
}

function simplify(points: readonly Point[]): Point[] {
  const kept: Point[] = [];
  for (const point of points) {
    const last = kept[kept.length - 1];
    const before = kept[kept.length - 2];
    if (last?.x === point.x && last.y === point.y) continue;
    if (before !== undefined && last !== undefined && collinear(before, last, point)) kept.pop();
    kept.push(point);
  }
  return kept;
}

function detour(link: WireLink): Point[] {
  const { start, goal } = stubs(link);
  if (!horizontal(link.leaves ?? "right")) {
    const middle = (start.y + goal.y) / 2;
    return [link.from, start, { x: start.x, y: middle }, { x: goal.x, y: middle }, goal, link.to];
  }
  const middle = (start.x + goal.x) / 2;
  return [link.from, start, { x: middle, y: start.y }, { x: middle, y: goal.y }, goal, link.to];
}

function vertexOf(grid: Grid, point: Point): number {
  return grid.ny * nearestIndex(grid.xs, point.x) + nearestIndex(grid.ys, point.y);
}

function polyline(link: WireLink, grid: Grid, path: readonly number[]): Point[] {
  const points = path.map((vertex) => ({
    x: grid.xs[Math.floor(vertex / grid.ny)] ?? 0,
    y: grid.ys[vertex % grid.ny] ?? 0,
  }));
  return [link.from, ...points, link.to];
}

function weight(link: WireLink): number {
  return Math.abs(link.to.x - link.from.x) + Math.abs(link.to.y - link.from.y);
}

function pathOf(
  grid: Grid,
  occupancy: Occupancy,
  ends: { readonly start: number; readonly goal: number },
  turned: Turns,
  corridor: Rect | null,
): number[] | null {
  const within = corridor === null ? null : findPath(grid, occupancy, ends, turned, corridor);
  return within ?? findPath(grid, occupancy, ends, turned);
}

function routeOn(scene: WireScene, titleClearance: number | null, ordered: readonly WireLink[]): Map<string, Point[]> {
  const grid = gridOf(scene, obstaclesOf(scene, titleClearance));
  const routes = new Map<string, Point[]>();
  const occupancy = emptyOccupancy(grid);
  for (const link of ordered) {
    const { start, goal } = stubs(link);
    const ends = { start: vertexOf(grid, start), goal: vertexOf(grid, goal) };
    const turned = { start: !horizontal(link.leaves ?? "right"), goal: !horizontal(link.enters ?? "left") };
    const found = grid.nx * grid.ny <= gridLimit ? pathOf(grid, occupancy, ends, turned, link.corridor ?? null) : null;
    if (found !== null) occupy(grid, occupancy, found, wireClearance.lane * (scene.pixel ?? 1));
    routes.set(link.id, found === null ? [] : polyline(link, grid, found));
  }
  return routes;
}

export function routeWires(scene: WireScene): ReadonlyMap<string, readonly Point[]> {
  const ordered = [...scene.links].sort((a, b) => weight(a) - weight(b) || a.id.localeCompare(b.id));
  const routes = new Map<string, readonly Point[]>();
  const passes: (number | null)[] = [...(scene.titleClearances ?? [0]), null];
  let pending = ordered;
  for (const clearance of passes) {
    if (pending.length === 0) break;
    const found = routeOn(scene, clearance, pending);
    pending.forEach((link) => {
      const path = found.get(link.id) ?? [];
      if (path.length > 0) routes.set(link.id, simplify(path));
    });
    pending = pending.filter((link) => !routes.has(link.id));
  }
  pending.forEach((link) => routes.set(link.id, simplify(detour(link))));
  return routes;
}
