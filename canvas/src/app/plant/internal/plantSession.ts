import type { PlantGateway } from "../../ports";
import { createOperationIdentity } from "../../shared";
import type { PlantSession } from "./api";
import { PlantStore } from "./plantStore";

export function createPlantSession(gateway: PlantGateway): PlantSession {
  return new PlantStore(gateway, createOperationIdentity());
}
