import type { ExamplePlant } from "../../../model/plant";
import { plantTitle } from "../../../model/plant";
import type { Definition } from "../../../model/workspace";
import type { PlantSource } from "../../ports";
import type { MakeMode } from "./plantMaker";

export interface BaseChoice {
  readonly key: string;
  readonly label: string;
  readonly group: "plants" | "examples";
  readonly source: PlantSource;
}

export function baseChoices(bases: readonly Definition[], examples: readonly ExamplePlant[]): readonly BaseChoice[] {
  return [
    ...bases.map((base): BaseChoice => ({
      key: `file:${base.path}`,
      label: plantTitle(base),
      group: "plants",
      source: { kind: "file", path: base.path },
    })),
    ...examples.map((example): BaseChoice => ({
      key: `example:${example.id}`,
      label: example.title,
      group: "examples",
      source: { kind: "example", id: example.id },
    })),
  ];
}

export function startingMode({ blank }: { readonly blank: boolean }): MakeMode {
  return blank ? "empty" : "copy";
}

export function newPlantReadiness(name: string, id: string, base: BaseChoice | null): string {
  if (name.trim() === "") return "Name the plant first.";
  if (id.trim() === "") return "Give it an id.";
  return base === null ? "Choose what to start from." : "";
}
