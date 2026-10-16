import type { Rect } from "./viewport";
import type { Grid } from "./wireGrid";

const bendCost = 24;
const overlapCost = 60;
const crossingCost = 18;
const hugCost = 2;

export interface Occupancy {
  readonly usedH: Uint8Array;
  readonly usedV: Uint8Array;
  readonly passH: Uint8Array;
  readonly passV: Uint8Array;
}

export function emptyOccupancy(grid: Grid): Occupancy {
  const size = grid.nx * grid.ny;
  return {
    usedH: new Uint8Array(size),
    usedV: new Uint8Array(size),
    passH: new Uint8Array(size),
    passV: new Uint8Array(size),
  };
}

class Heap {
  private keys: number[] = [];
  private items: number[] = [];

  get size(): number {
    return this.items.length;
  }

  push(key: number, item: number): void {
    let at = this.items.length;
    this.keys.push(key);
    this.items.push(item);
    while (at > 0) {
      const parent = (at - 1) >> 1;
      if ((this.keys[parent] ?? 0) <= key) break;
      this.swap(at, parent);
      at = parent;
    }
  }

  pop(): number {
    const top = this.items[0] ?? -1;
    const lastKey = this.keys.pop() ?? 0;
    const lastItem = this.items.pop() ?? -1;
    if (this.items.length > 0) {
      this.keys[0] = lastKey;
      this.items[0] = lastItem;
      this.sift();
    }
    return top;
  }

  private sift(): void {
    let at = 0;
    for (;;) {
      const least = this.smallest(at);
      if (least === at) return;
      this.swap(at, least);
      at = least;
    }
  }

  private smallest(at: number): number {
    let least = at;
    for (const child of [at * 2 + 1, at * 2 + 2]) {
      if (child < this.keys.length && (this.keys[child] ?? 0) < (this.keys[least] ?? 0)) least = child;
    }
    return least;
  }

  private swap(a: number, b: number): void {
    [this.keys[a], this.keys[b]] = [this.keys[b] ?? 0, this.keys[a] ?? 0];
    [this.items[a], this.items[b]] = [this.items[b] ?? 0, this.items[a] ?? 0];
  }
}

interface Move {
  readonly to: number;
  readonly cost: number;
}

class Search {
  private readonly cost: Float64Array;
  private readonly parent: Int32Array;

  constructor(
    private readonly grid: Grid,
    private readonly occupancy: Occupancy,
    private readonly corridor: Rect | null,
  ) {
    this.cost = new Float64Array(grid.nx * grid.ny * 2).fill(Number.POSITIVE_INFINITY);
    this.parent = new Int32Array(grid.nx * grid.ny * 2).fill(-1);
  }

  run(start: number, goal: number, vertical: { readonly start: boolean; readonly goal: boolean }): number[] | null {
    const heap = new Heap();
    const target = goal * 2 + (vertical.goal ? 1 : 0);
    const first = start * 2 + (vertical.start ? 1 : 0);
    this.cost[first] = 0;
    heap.push(this.estimate(start, goal), first);
    while (heap.size > 0) {
      const state = heap.pop();
      if (state === target) return this.path(state);
      for (const move of this.moves(state)) {
        const total = (this.cost[state] ?? 0) + move.cost;
        if (total >= (this.cost[move.to] ?? 0)) continue;
        this.cost[move.to] = total;
        this.parent[move.to] = state;
        heap.push(total + this.estimate(move.to >> 1, goal), move.to);
      }
    }
    return null;
  }

  private estimate(vertex: number, goal: number): number {
    const { xs, ys, ny } = this.grid;
    const dx = Math.abs((xs[Math.floor(vertex / ny)] ?? 0) - (xs[Math.floor(goal / ny)] ?? 0));
    return dx + Math.abs((ys[vertex % ny] ?? 0) - (ys[goal % ny] ?? 0));
  }

  private moves(state: number): Move[] {
    const vertex = state >> 1;
    const vertical = (state & 1) === 1;
    const turn: Move = { to: vertex * 2 + (vertical ? 0 : 1), cost: bendCost };
    const along = vertical ? this.vertical(vertex) : this.horizontal(vertex);
    return [turn, ...along.filter((move) => this.allowed(move.to >> 1))];
  }

  private horizontal(vertex: number): Move[] {
    const { xs, ny, nx, freeH } = this.grid;
    const ix = Math.floor(vertex / ny);
    const moves: Move[] = [];
    if (ix > 0 && freeH[vertex - ny] === 1) {
      const length = (xs[ix] ?? 0) - (xs[ix - 1] ?? 0);
      moves.push(this.step(vertex - ny, 0, length, vertex - ny));
    }
    if (ix < nx - 1 && freeH[vertex] === 1) {
      const length = (xs[ix + 1] ?? 0) - (xs[ix] ?? 0);
      moves.push(this.step(vertex + ny, 0, length, vertex));
    }
    return moves;
  }

  private vertical(vertex: number): Move[] {
    const { ys, ny, freeV } = this.grid;
    const iy = vertex % ny;
    const moves: Move[] = [];
    if (iy > 0 && freeV[vertex - 1] === 1) {
      const length = (ys[iy] ?? 0) - (ys[iy - 1] ?? 0);
      moves.push(this.step(vertex - 1, 1, length, vertex - 1));
    }
    if (iy < ny - 1 && freeV[vertex] === 1) {
      const length = (ys[iy + 1] ?? 0) - (ys[iy] ?? 0);
      moves.push(this.step(vertex + 1, 1, length, vertex));
    }
    return moves;
  }

  private allowed(vertex: number): boolean {
    const box = this.corridor;
    if (box === null) return true;
    const { xs, ys, ny } = this.grid;
    const x = xs[Math.floor(vertex / ny)] ?? 0;
    const y = ys[vertex % ny] ?? 0;
    return x >= box.x && x <= box.x + box.width && y >= box.y && y <= box.y + box.height;
  }

  private step(to: number, orientation: 0 | 1, length: number, edge: number): Move {
    const used = (orientation === 0 ? this.occupancy.usedH : this.occupancy.usedV)[edge] === 1;
    const hugging = (orientation === 0 ? this.grid.hugH : this.grid.hugV)[edge] === 1;
    const across = orientation === 0 ? this.occupancy.passV[to] : this.occupancy.passH[to];
    const extra = (used ? overlapCost : 0) + (across === 1 ? crossingCost : 0) + (hugging ? hugCost * length : 0);
    return { to: to * 2 + orientation, cost: length + extra };
  }

  private path(end: number): number[] {
    const vertices: number[] = [];
    for (let state = end; state >= 0; state = this.parent[state] ?? -1) {
      const vertex = state >> 1;
      if (vertices[0] !== vertex) vertices.unshift(vertex);
    }
    return vertices;
  }
}

export interface Ends {
  readonly start: number;
  readonly goal: number;
}

export function findPath(
  grid: Grid,
  occupancy: Occupancy,
  ends: Ends,
  vertical: { readonly start: boolean; readonly goal: boolean },
  corridor: Rect | null = null,
): number[] | null {
  return new Search(grid, occupancy, corridor).run(ends.start, ends.goal, vertical);
}

export function occupy(grid: Grid, occupancy: Occupancy, path: readonly number[], pitch = 0): void {
  const { ny } = grid;
  for (let index = 1; index < path.length; index += 1) {
    const from = path[index - 1] ?? 0;
    const to = path[index] ?? 0;
    claimLine(grid, occupancy, from, to, Math.floor(from / ny) !== Math.floor(to / ny), pitch);
  }
}

function claimLine(
  grid: Grid,
  occupancy: Occupancy,
  from: number,
  to: number,
  horizontal: boolean,
  pitch: number,
): void {
  const { ny } = grid;
  const used = horizontal ? occupancy.usedH : occupancy.usedV;
  const pass = horizontal ? occupancy.passH : occupancy.passV;
  const step = horizontal ? ny : 1;
  const last = Math.max(from, to);
  for (let vertex = Math.min(from, to); vertex <= last; vertex += step) {
    if (vertex < last) {
      used[vertex] = 1;
      crowd(grid, occupancy, vertex, horizontal, pitch);
    }
    pass[vertex] = 1;
  }
}

function crowd(grid: Grid, occupancy: Occupancy, vertex: number, horizontal: boolean, pitch: number): void {
  if (pitch <= 0) return;
  const { ny, ys, xs } = grid;
  const ix = Math.floor(vertex / ny);
  const iy = vertex % ny;
  const lines = horizontal ? ys : xs;
  const at = horizontal ? iy : ix;
  const used = horizontal ? occupancy.usedH : occupancy.usedV;
  crowdLines(lines, at, pitch, (other: number): void => {
    used[horizontal ? ix * ny + other : other * ny + iy] = 1;
  });
}

function crowdLines(lines: readonly number[], at: number, pitch: number, close: (other: number) => void): void {
  for (const direction of [-1, 1]) {
    let other = at + direction;
    while (other >= 0 && other < lines.length && Math.abs((lines[other] ?? 0) - (lines[at] ?? 0)) < pitch) {
      close(other);
      other += direction;
    }
  }
}
