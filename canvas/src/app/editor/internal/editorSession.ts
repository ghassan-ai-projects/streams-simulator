import type { EditorGateway } from "../../ports";
import { createOperationIdentity } from "../../shared";
import type { EditorSession } from "../contract";
import { EditorStore } from "./editorStore";

export function createEditorSession(gateway: EditorGateway): EditorSession {
  return new EditorStore(gateway, createOperationIdentity());
}
