import { spaced } from "../../shared/words";
import type { Plant } from "./plant";
import { endpointOf } from "./plant";
import type { ComponentSpec } from "./ports";
import { portType, specOf } from "./ports";

function typeAt(plant: Plant, specs: ReadonlyMap<string, ComponentSpec>, end: string): string {
  const { id, port } = endpointOf(end);
  return portType(
    specOf(
      specs,
      plant.instances.find((entry) => entry.id === id),
    ),
    port,
  );
}

function endWords(end: string): string {
  const { id, port } = endpointOf(end);
  return `${id}'s ${spaced(port)}`;
}

export function linkRefusal(
  plant: Plant,
  specs: ReadonlyMap<string, ComponentSpec>,
  from: string,
  to: string,
): string | null {
  const gives = typeAt(plant, specs, from);
  const takes = typeAt(plant, specs, to);
  if (gives === "" || takes === "" || gives === takes) return null;
  return `${endWords(from)} is ${spaced(gives)}, but ${endWords(to)} takes ${spaced(takes)}. A link joins an output to an input of the same type.`;
}
