import type { Rect } from "./viewport";
import type { Grid } from "./wireGrid";
import { buildGrid, lanes, nearestIndex } from "./wireGrid";
import { emptyOccupancy, findPath, occupy } from "./wireSearch";

export interface Point {
  readonly x: number;
  readonly y: number;
}

export interface WireLink {
  readonly id: string;
  readonly from: Point;
  readonly to: Point;
}

export interface WireScene {
  readonly links: readonly WireLink[];
  readonly nodes: readonly Rect[];
  readonly titles: readonly Rect[];
}

export const wireClearance = { node: 8, title: 4, stub: 16, lane: 6 } as const;

const laneCount = { x: 4, y: 2 } as const;
const gridLimit = 400_000;

function inflate(box: Rect, by: number): Rect {
  return { x: box.x - by, y: box.y - by, width: box.width + 2 * by, height: box.height + 2 * by };
}

function stubs(link: WireLink): { readonly start: Point; readonly goal: Point } {
  return {
    start: { x: link.from.x + wireClearance.stub, y: link.from.y },
    goal: { x: link.to.x - wireClearance.stub, y: link.to.y },
  };
}

function obstaclesOf(scene: WireScene): Rect[] {
  return [
    ...scene.nodes.map((box) => inflate(box, wireClearance.node)),
    ...scene.titles.map((box) => inflate(box, wireClearance.title)),
  ];
}

function gridOf(scene: WireScene, obstacles: readonly Rect[]): Grid {
  const { lane } = wireClearance;
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
  return buildGrid(obstacles, xs, ys);
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

export function routeWires(scene: WireScene): ReadonlyMap<string, readonly Point[]> {
  const routes = new Map<string, readonly Point[]>();
  const grid = gridOf(scene, obstaclesOf(scene));
  const searchable = grid.nx * grid.ny <= gridLimit;
  const occupancy = emptyOccupancy(grid);
  const ordered = [...scene.links].sort((a, b) => weight(a) - weight(b) || a.id.localeCompare(b.id));
  for (const link of ordered) {
    const { start, goal } = stubs(link);
    const found = searchable ? findPath(grid, occupancy, vertexOf(grid, start), vertexOf(grid, goal)) : null;
    if (found !== null) occupy(grid, occupancy, found);
    routes.set(link.id, simplify(found === null ? detour(link) : polyline(link, grid, found)));
  }
  return routes;
}
