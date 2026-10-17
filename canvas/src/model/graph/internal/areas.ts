import type { Plant } from "../../plant";
import { areaOf } from "../../plant";
import { areaTree } from "../../shared/areaTree";
import { tally } from "../../shared/collect";

export function areaOrder(plant: Plant): readonly (readonly [string, number])[] {
  const declared = areaTree(plant.areas).map((node) => [node.name, node.depth] as const);
  const loose = [...new Set(plant.instances.map((instance) => areaOf(plant, instance)))].filter(
    (area) => !declared.some(([name]) => name === area),
  );
  return [...declared, ...loose.map((area) => [area, 0] as const)];
}

export function areaCounts(plant: Plant): readonly (readonly [string, number])[] {
  const counts = tally(plant.instances, (instance) => areaOf(plant, instance));
  return areaOrder(plant).flatMap(([name]) => {
    const count = counts.get(name);
    return count === undefined ? [] : [[name, count] as const];
  });
}
