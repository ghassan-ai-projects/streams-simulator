import { EditorStore } from "../app/editor";
import { PlantStore } from "../app/plant";
import type { EditorGateway, PlantGateway } from "../app/ports";
import { createOperationIdentity } from "../app/shared";

export function editorStore(gateway: EditorGateway): EditorStore {
  return new EditorStore(gateway, createOperationIdentity());
}

export function plantStore(gateway: PlantGateway): PlantStore {
  return new PlantStore(gateway, createOperationIdentity());
}
