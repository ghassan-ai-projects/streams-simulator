export interface AreaRef {
  readonly name: string;
  readonly parent: string | null;
}

export interface AreaNode extends AreaRef {
  readonly depth: number;
  readonly path: string;
}

function childrenOf(areas: readonly AreaRef[], parent: string | null): readonly AreaRef[] {
  const names = new Set(areas.map((area) => area.name));
  return areas.filter((area) =>
    parent === null ? area.parent === null || !names.has(area.parent) : area.parent === parent,
  );
}

function visit(
  areas: readonly AreaRef[],
  area: AreaRef,
  trail: readonly string[],
  seen: Set<string>,
): readonly AreaNode[] {
  if (seen.has(area.name)) return [];
  seen.add(area.name);
  const path = [...trail, area.name];
  const below = childrenOf(areas, area.name).flatMap((child) => visit(areas, child, path, seen));
  return [{ ...area, depth: trail.length, path: path.join(" › ") }, ...below];
}

export function areaTree(areas: readonly AreaRef[]): readonly AreaNode[] {
  const seen = new Set<string>();
  const roots = childrenOf(areas, null).flatMap((area) => visit(areas, area, [], seen));
  const loose = areas.filter((area) => !seen.has(area.name)).flatMap((area) => visit(areas, area, [], seen));
  return [...roots, ...loose];
}
