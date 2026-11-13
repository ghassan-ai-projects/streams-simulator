import type { Pins, Plant } from "../../plant";
import { tally } from "../../shared/collect";
import type { RowCount, ViewNode, ViewPresentation, Wires } from "./types";

const groupedFrom = 3;
const listedChannels = 6;

function widestPair(plant: Plant): number {
  const pairs = tally(
    plant.links.filter((link) => link.from !== link.to),
    (link) => `${link.from}\u0000${link.to}`,
  );
  return Math.max(0, ...pairs.values());
}

export function canGroupWires(plant: Plant): boolean {
  return widestPair(plant) > 1;
}

export function openingWires(plant: Plant): Wires {
  return widestPair(plant) >= groupedFrom ? "grouped" : "each";
}

export function channelRows(count: number): { readonly listed: number; readonly more: number } {
  if (count <= listedChannels) return { listed: count, more: 0 };
  return { listed: listedChannels - 1, more: count - (listedChannels - 1) };
}

const noPins: Pins = { ins: [], outs: [] };

export function rowPins(node: ViewNode): Pins {
  return node.ported ? node.pins : noPins;
}

const groupedRows: RowCount = (_, channels) => {
  const { listed, more } = channelRows(channels.length);
  return Math.max(1, listed + (more > 0 ? 1 : 0));
};

export const groupedView: ViewPresentation = { rows: groupedRows, note: () => 0 };
