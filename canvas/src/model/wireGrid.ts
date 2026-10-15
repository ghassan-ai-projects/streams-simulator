import type { Rect } from "./viewport";

export interface Grid {
  readonly xs: readonly number[];
  readonly ys: readonly number[];
  readonly nx: number;
  readonly ny: number;
  readonly freeH: Uint8Array;
  readonly freeV: Uint8Array;
}

function canonical(value: number): number {
  return Math.round(value * 1000) / 1000;
}

function sortedUnique(values: readonly number[]): number[] {
  return [...new Set(values.map(canonical))].sort((a, b) => a - b);
}

function indexer(values: readonly number[]): ReadonlyMap<number, number> {
  return new Map(values.map((value, index) => [value, index] as const));
}

function blockedCells(xs: readonly number[], ys: readonly number[], obstacles: readonly Rect[]): Uint8Array {
  const columns = indexer(xs);
  const rows = indexer(ys);
  const cells = new Uint8Array(Math.max(0, (xs.length - 1) * (ys.length - 1)));
  for (const box of obstacles) {
    const x0 = columns.get(canonical(box.x)) ?? 0;
    const x1 = columns.get(canonical(box.x + box.width)) ?? 0;
    const y0 = rows.get(canonical(box.y)) ?? 0;
    const y1 = rows.get(canonical(box.y + box.height)) ?? 0;
    for (let ix = x0; ix < x1; ix += 1) cells.fill(1, ix * (ys.length - 1) + y0, ix * (ys.length - 1) + y1);
  }
  return cells;
}

function freeSegments(nx: number, ny: number, cells: Uint8Array): Pick<Grid, "freeH" | "freeV"> {
  const blocked = (ix: number, iy: number) =>
    ix >= 0 && iy >= 0 && ix < nx - 1 && iy < ny - 1 && cells[ix * (ny - 1) + iy] === 1;
  const freeH = new Uint8Array(nx * ny);
  const freeV = new Uint8Array(nx * ny);
  for (let ix = 0; ix < nx; ix += 1) {
    for (let iy = 0; iy < ny; iy += 1) {
      freeH[ix * ny + iy] = ix < nx - 1 && !(blocked(ix, iy - 1) && blocked(ix, iy)) ? 1 : 0;
      freeV[ix * ny + iy] = iy < ny - 1 && !(blocked(ix - 1, iy) && blocked(ix, iy)) ? 1 : 0;
    }
  }
  return { freeH, freeV };
}

export function lanes(edge: number, outward: 1 | -1, count: number, pitch: number): number[] {
  return Array.from({ length: count }, (_, index) => edge + outward * pitch * index);
}

export function buildGrid(obstacles: readonly Rect[], extraX: readonly number[], extraY: readonly number[]): Grid {
  const xs = sortedUnique([...obstacles.flatMap((box) => [box.x, box.x + box.width]), ...extraX]);
  const ys = sortedUnique([...obstacles.flatMap((box) => [box.y, box.y + box.height]), ...extraY]);
  const nx = xs.length;
  const ny = ys.length;
  return { xs, ys, nx, ny, ...freeSegments(nx, ny, blockedCells(xs, ys, obstacles)) };
}

export function nearestIndex(values: readonly number[], target: number): number {
  const wanted = canonical(target);
  let best = 0;
  values.forEach((value, index) => {
    if (Math.abs(value - wanted) < Math.abs((values[best] ?? 0) - wanted)) best = index;
  });
  return best;
}
