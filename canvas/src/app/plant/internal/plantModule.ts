import type { Workspace } from "../../../model/workspace";
import type { DomainDocument, PlantGateway } from "../../ports";
import { createOperationIdentity } from "../../shared";
import { PlantStore } from "./plantStore";
import type { PlantState } from "./plantStore";

export interface PlantSession {
  readonly snapshot: () => PlantState;
  readonly subscribe: (listener: () => void) => () => void;
  load(path: string, workspace: Workspace): Promise<void>;
  show(key: string, domain: DomainDocument): Promise<void>;
  select(id: string | null): void;
  close(): void;
}

export interface PlantModule {
  createSession(): PlantSession;
}

export function createPlantModule(gateway: PlantGateway): PlantModule {
  return { createSession: () => new PlantStore(gateway, createOperationIdentity()) };
}
