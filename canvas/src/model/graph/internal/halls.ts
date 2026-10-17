import type { PlantView } from "./types";
import type { Plant } from "../../plant";
import { tally } from "../../shared/collect";

export function hallOf(id: string): string | null {
  const slash = id.indexOf("/");
  return slash > 0 ? id.slice(0, slash) : null;
}

export function hallsOf(plant: Plant): readonly string[] {
  const halls = new Set(
    plant.instances.map((instance) => hallOf(instance.id)).filter((hall): hall is string => hall !== null),
  );
  return halls.size > 1 ? [...halls].sort() : [];
}

export function hallMembers(plant: Plant, hall: string): readonly (readonly [string, number])[] {
  const members = plant.instances.filter((instance) => hallOf(instance.id) === hall);
  return [...tally(members, (instance) => instance.kind)].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]));
}

export function hallCount(view: PlantView, hall: string): string {
  const own = view.nodes.filter((node) => hallOf(node.id) === hall).length;
  const shared = view.nodes.filter((node) => node.instance !== null).length - own;
  return `${String(own)} instances${shared > 0 ? ` plus ${String(shared)} shared` : ""}`;
}
