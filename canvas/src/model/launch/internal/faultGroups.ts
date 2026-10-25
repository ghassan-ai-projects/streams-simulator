import type { Plant } from "../../plant";
import type { FaultCategories, FaultCategory } from "./perturbation";
import { faultCategoryOrder, faultCategoryWords } from "./perturbation";

export interface FaultGroup {
  readonly label: string;
  readonly faults: readonly string[];
}

export function faultGroups(
  faults: readonly string[],
  categories: Readonly<Record<string, FaultCategory>> | undefined,
): readonly FaultGroup[] {
  const known = faults.filter((fault) => categories?.[fault] !== undefined);
  if (known.length === 0) return [{ label: "", faults }];
  const groups = faultCategoryOrder.map((category) => ({
    label: faultCategoryWords[category],
    faults: known.filter((fault) => categories?.[fault] === category),
  }));
  const other = { label: "Other faults", faults: faults.filter((fault) => !known.includes(fault)) };
  return [...groups, other].filter((group) => group.faults.length > 0);
}

export function faultCategoryOf(
  plant: Plant,
  categories: FaultCategories,
  entity: string,
  fault: string,
): FaultCategory | null {
  const kind = plant.instances.find((instance) => instance.id === entity)?.kind ?? "";
  return categories[kind]?.[fault] ?? null;
}
