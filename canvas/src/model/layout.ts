interface Box {
  readonly x: number;
  readonly y: number;
  readonly width: number;
  readonly height: number;
}

export interface LayoutNode {
  readonly id: string;
  readonly group: string;
  readonly width: number;
  readonly height: number;
}

export interface LayoutEdge {
  readonly from: string;
  readonly to: string;
}

export interface Layout {
  readonly nodes: ReadonlyMap<string, Box>;
  readonly groups: ReadonlyMap<string, Box>;
}

export const spacing = { pad: 24, header: 40, column: 96, row: 24, group: 48, maxWidth: 2600, maxRows: 3 } as const;

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

export function ranks(ids: readonly string[], edges: readonly LayoutEdge[]): Map<string, number> {
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

function columnsOf(nodes: readonly LayoutNode[], edges: readonly LayoutEdge[]): LayoutNode[][] {
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
  ).flatMap(wrap);
}

function wrap(column: LayoutNode[]): LayoutNode[][] {
  const parts: LayoutNode[][] = [];
  for (let start = 0; start < column.length; start += spacing.maxRows) {
    parts.push(column.slice(start, start + spacing.maxRows));
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

function placeGroup(columns: readonly LayoutNode[][], origin: { x: number; y: number }, out: Map<string, Box>): Box {
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

export function layout(
  nodes: readonly LayoutNode[],
  edges: readonly LayoutEdge[],
  groupOrder: readonly string[],
): Layout {
  const byGroup = new Map<string, LayoutNode[]>();
  nodes.forEach((node) => {
    byGroup.set(node.group, [...(byGroup.get(node.group) ?? []), node]);
  });
  const order = [
    ...groupOrder.filter((group) => byGroup.has(group)),
    ...[...byGroup.keys()].filter((group) => !groupOrder.includes(group)).sort(),
  ];
  const placed = new Map<string, Box>();
  const groups = new Map<string, Box>();
  let cursor = { x: 0, y: 0 };
  let rowHeight = 0;
  for (const group of order) {
    const box = placeGroup(columnsOf(byGroup.get(group) ?? [], edges), cursor, placed);
    if (box.x > 0 && box.x + box.width > spacing.maxWidth) {
      cursor = { x: 0, y: cursor.y + rowHeight + spacing.group };
      rowHeight = 0;
      groups.set(group, placeGroup(columnsOf(byGroup.get(group) ?? [], edges), cursor, placed));
    } else {
      groups.set(group, box);
    }
    const final = groups.get(group) ?? box;
    rowHeight = Math.max(rowHeight, final.height);
    cursor = { x: final.x + final.width + spacing.group, y: cursor.y };
  }
  return { nodes: placed, groups };
}
