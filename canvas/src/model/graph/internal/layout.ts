import type { Layout, LayoutEdge, LayoutNode, Rect } from "./types";

const spacing = {
  pad: 24,
  header: 40,
  column: 96,
  row: 24,
  group: 48,
  lane: 24,
  maxWidth: 2600,
  maxRows: 3,
} as const;

function successors(ids: readonly string[], edges: readonly LayoutEdge[]): Map<string, string[]> {
  const inside = new Set(ids);
  const next = new Map(ids.map((id) => [id, [] as string[]]));
  for (const edge of edges) {
    if (inside.has(edge.from) && inside.has(edge.to) && edge.from !== edge.to) {
      next.get(edge.from)?.push(edge.to);
    }
  }
  next.forEach((targets) => targets.sort());
  return next;
}

function acyclicOrder(
  ids: readonly string[],
  next: ReadonlyMap<string, readonly string[]>,
): { order: string[]; back: Set<string> } {
  const state = new Map<string, "open" | "done">();
  const order: string[] = [];
  const back = new Set<string>();
  const visit = (id: string): void => {
    state.set(id, "open");
    for (const target of next.get(id) ?? []) {
      if (state.get(target) === "open") back.add(`${id}\u0000${target}`);
      else if (!state.has(target)) visit(target);
    }
    state.set(id, "done");
    order.unshift(id);
  };
  [...ids].sort().forEach((id) => {
    if (!state.has(id)) visit(id);
  });
  return { order, back };
}

function ranks(ids: readonly string[], edges: readonly LayoutEdge[]): Map<string, number> {
  const next = successors(ids, edges);
  const { order, back } = acyclicOrder(ids, next);
  const rank = new Map(ids.map((id) => [id, 0]));
  for (const id of order) {
    for (const target of next.get(id) ?? []) {
      if (!back.has(`${id}\u0000${target}`)) rank.set(target, Math.max(rank.get(target) ?? 0, (rank.get(id) ?? 0) + 1));
    }
  }
  return rank;
}

function columnsOf(nodes: readonly LayoutNode[], edges: readonly LayoutEdge[], rows: number): LayoutNode[][] {
  const rank = ranks(
    nodes.map((node) => node.id),
    edges,
  );
  const columns: LayoutNode[][] = [];
  for (const node of [...nodes].sort((a, b) => a.id.localeCompare(b.id))) {
    const index = rank.get(node.id) ?? 0;
    (columns[index] ??= []).push(node);
  }
  return orderByPredecessors(
    columns.filter((column) => column.length > 0),
    edges,
  ).flatMap((column) => wrap(column, rows));
}

function wrap(column: LayoutNode[], rows: number): LayoutNode[][] {
  const parts: LayoutNode[][] = [];
  for (let start = 0; start < column.length; start += rows) {
    parts.push(column.slice(start, start + rows));
  }
  return parts;
}

function orderByPredecessors(columns: LayoutNode[][], edges: readonly LayoutEdge[]): LayoutNode[][] {
  const row = new Map<string, number>();
  return columns.map((column) => {
    const score = (node: LayoutNode): number => {
      const rows = edges
        .filter((edge) => edge.to === node.id && row.has(edge.from))
        .map((edge) => row.get(edge.from) ?? 0);
      return rows.length === 0 ? Number.POSITIVE_INFINITY : rows.reduce((sum, value) => sum + value, 0) / rows.length;
    };
    const sorted = [...column].sort((a, b) => score(a) - score(b) || a.id.localeCompare(b.id));
    sorted.forEach((node, index) => row.set(node.id, index));
    return sorted;
  });
}

function placeGroup(columns: readonly LayoutNode[][], origin: { x: number; y: number }, out: Map<string, Rect>): Rect {
  let x = origin.x + spacing.pad;
  let bottom = origin.y + spacing.header;
  for (const column of columns) {
    let y = origin.y + spacing.header;
    const width = Math.max(...column.map((node) => node.width));
    for (const node of column) {
      out.set(node.id, { x, y, width: node.width, height: node.height });
      y += node.height + spacing.row;
    }
    bottom = Math.max(bottom, y - spacing.row);
    x += width + spacing.column;
  }
  return {
    x: origin.x,
    y: origin.y,
    width: x - spacing.column + spacing.pad - origin.x,
    height: bottom + spacing.pad - origin.y,
  };
}

type RowsFor = (group: string, count: number) => number;

function groupsOf(nodes: readonly LayoutNode[], groupOrder: readonly string[]) {
  const byGroup = new Map<string, LayoutNode[]>();
  nodes.forEach((node) => {
    byGroup.set(node.group, [...(byGroup.get(node.group) ?? []), node]);
  });
  const known = groupOrder.filter((group) => byGroup.has(group));
  const rest = [...byGroup.keys()].filter((group) => !groupOrder.includes(group)).sort();
  return { byGroup, order: [...known, ...rest] };
}

class Shelf {
  readonly placed = new Map<string, Rect>();
  readonly groups = new Map<string, Rect>();
  private cursor = { x: 0, y: 0 };
  private rowHeight = 0;
  width = 0;

  constructor(
    private readonly maxWidth: number,
    private readonly rowGap: number = spacing.group,
  ) {}

  add(group: string, columns: readonly LayoutNode[][]): void {
    const trial = new Map<string, Rect>();
    let box = placeGroup(columns, this.cursor, trial);
    if (box.x > 0 && box.x + box.width > this.maxWidth) {
      this.cursor = { x: 0, y: this.cursor.y + this.rowHeight + this.rowGap };
      this.rowHeight = 0;
      trial.clear();
      box = placeGroup(columns, this.cursor, trial);
    }
    trial.forEach((value, key) => this.placed.set(key, value));
    this.groups.set(group, box);
    this.rowHeight = Math.max(this.rowHeight, box.height);
    this.width = Math.max(this.width, box.x + box.width);
    this.cursor = { x: box.x + box.width + spacing.group, y: this.cursor.y };
  }

  get height(): number {
    return this.cursor.y + this.rowHeight;
  }
}

type Columns = readonly (readonly [string, LayoutNode[][]])[];

function pack(columns: Columns, maxWidth: number, rowGap: number): Shelf {
  const shelf = new Shelf(maxWidth, rowGap);
  columns.forEach(([group, groupColumns]) => {
    shelf.add(group, groupColumns);
  });
  return shelf;
}

function fitOf(shelf: Shelf, aspect: number): number {
  return Math.min(aspect / Math.max(1, shelf.width), 1 / Math.max(1, shelf.height));
}

function widthsToTry(columns: Columns): readonly number[] {
  const widths = columns.map(([, groupColumns]) => placeGroup(groupColumns, { x: 0, y: 0 }, new Map()).width);
  return widths.map((_, index) => widths.slice(0, index + 1).reduce((sum, w) => sum + w + spacing.group, 0));
}

function bestShelf(columns: Columns, aspect: number | undefined, rowGap: number): Shelf {
  if (aspect === undefined) return pack(columns, spacing.maxWidth, rowGap);
  const shelves = widthsToTry(columns).map((width) => pack(columns, width, rowGap));
  return shelves.reduce(
    (best, shelf) => (fitOf(shelf, aspect) > fitOf(best, aspect) ? shelf : best),
    pack(columns, spacing.maxWidth, rowGap),
  );
}

type HeaderReserve = (group: string, members: number, width: number) => number;

function rowGapOf(columns: Columns, reserve: HeaderReserve | undefined): number {
  if (reserve === undefined) return spacing.group;
  const heights = columns.map(([group, groupColumns]) => {
    const width = placeGroup(groupColumns, { x: 0, y: 0 }, new Map()).width;
    const members = groupColumns.reduce((sum, column) => sum + column.length, 0);
    return reserve(group, members, width);
  });
  return Math.max(spacing.group, Math.max(0, ...heights) + spacing.lane);
}

export function layout(
  nodes: readonly LayoutNode[],
  edges: readonly LayoutEdge[],
  groupOrder: readonly string[],
  rowsFor: RowsFor = () => spacing.maxRows,
  aspect?: number,
  reserve?: HeaderReserve,
): Layout {
  const { byGroup, order } = groupsOf(nodes, groupOrder);
  const columns: Columns = order.map((group) => {
    const members = byGroup.get(group) ?? [];
    return [group, columnsOf(members, edges, rowsFor(group, members.length))] as const;
  });
  const shelf = bestShelf(columns, aspect, rowGapOf(columns, reserve));
  return { nodes: shelf.placed, groups: shelf.groups };
}
