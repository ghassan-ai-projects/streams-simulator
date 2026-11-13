import type { Family, Link } from "../../plant";
import type { ViewEdge } from "./types";

interface Way {
  readonly from: string;
  readonly to: string;
  readonly families: (Family | null)[];
}

interface Pair {
  readonly first: Way;
  readonly second: Way;
}

function carries(family: Family | null): boolean {
  return family !== null && family !== "signal";
}

function leads(way: Way, other: Way): boolean {
  const material = way.families.filter(carries).length - other.families.filter(carries).length;
  return material > 0 || (material === 0 && way.families.length > other.families.length);
}

function edgeOf(pair: Pair): ViewEdge {
  const [main, back] = leads(pair.second, pair.first) ? [pair.second, pair.first] : [pair.first, pair.second];
  const families = new Set([...main.families, ...back.families]);
  const [only = null] = families;
  return {
    id: `${main.from}.>${main.to}.`,
    from: main.from,
    to: main.to,
    fromHandle: null,
    toHandle: null,
    family: families.size === 1 ? only : null,
    count: main.families.length,
    reverse: back.families.length,
  };
}

export function groupedEdges(
  links: readonly Link[],
  shown: (id: string) => boolean,
  familyOf: (link: Link) => Family | null,
): ViewEdge[] {
  const pairs = new Map<string, Pair>();
  for (const link of links) {
    if (link.from === link.to || !shown(link.from) || !shown(link.to)) continue;
    const key = [link.from, link.to].sort().join("\u0000");
    const pair = pairs.get(key) ?? {
      first: { from: link.from, to: link.to, families: [] },
      second: { from: link.to, to: link.from, families: [] },
    };
    pairs.set(key, pair);
    (pair.first.from === link.from ? pair.first : pair.second).families.push(familyOf(link));
  }
  return [...pairs.values()].map(edgeOf);
}
