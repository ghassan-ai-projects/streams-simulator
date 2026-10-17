import { instanceIdTaken } from "../../plant";
import type { Plant } from "../../plant";
import { counted } from "../../shared/words";
import { instanceNameFits, instanceNameRule } from "./instanceNameRule";

const fewestRows = 2;
const mostRows = 6;

export function libraryRows(found: number): number {
  return Math.min(mostRows, Math.max(fewestRows, found));
}

export function libraryWords(found: number, shown: number, query: string): string {
  if (found === 0) return `Nothing in the library matches "${query.trim()}".`;
  const components = counted(found, "component");
  return found > shown ? `${components} · scroll the list for the rest` : components;
}

export function suggestedWords(suggested: string): string {
  return suggested === ""
    ? "Leave empty to use the suggested name."
    : `Leave empty to use the suggested name, ${suggested}.`;
}

export function addProblem(use: string, name: string, area: string, plant: Plant): string | null {
  if (use === "") return "Choose equipment from the library.";
  if (name === "") return "Give the instance a name.";
  if (!instanceNameFits(name)) return `"${name}" is not a valid name. ${instanceNameRule()}`;
  if (instanceIdTaken(plant, name)) return `${name} is already used by another instance.`;
  return area === "" ? "Choose an area." : null;
}
