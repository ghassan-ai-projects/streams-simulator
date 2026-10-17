import type { Workspace } from "../../../model/workspace";
import type { ExamplePlant } from "../../../model/plant";
import type { DomainDocument, Outcome, PlantFailure, PlantSource } from "../../ports";
import type { Store } from "../../shared";
import type { PlantState } from "./plantStore";

export type { PlantFailure } from "../../ports";
export type { PlantState, ReadyPlant } from "./plantStore";

export interface BaseChoice {
  readonly key: string;
  readonly label: string;
  readonly group: "plants" | "examples";
  readonly source: PlantSource;
}

export type MakeMode = "copy" | "empty";

export interface Made {
  readonly composition: string;
  readonly domain: string;
}

export interface PlantSession extends Store<PlantState> {
  load(path: string, workspace: Workspace): Promise<void>;
  show(key: string, domain: DomainDocument): Promise<void>;
  select(id: string | null): void;
  close(): void;
}

export interface PlantMaker {
  examples(): Promise<readonly ExamplePlant[]>;
  idProblem(source: PlantSource, id: string, title: string): Promise<string>;
  create(source: PlantSource, id: string, title: string, mode: MakeMode): Promise<Outcome<Made, PlantFailure>>;
  tryExample(example: ExamplePlant, id: string): Promise<Outcome<Made, PlantFailure>>;
}
