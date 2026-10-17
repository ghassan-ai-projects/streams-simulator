import type { Composition, NewEquipment } from "../../../model/editor";
import type { Presets } from "../../../model/plant";
import type { EditorGateway } from "../../ports";
import { createOperationIdentity } from "../../shared";
import { EditorStore } from "./editorStore";
import type { EditorState } from "./editorDraft";

export interface EditorSession {
  readonly snapshot: () => EditorState;
  readonly subscribe: (listener: () => void) => () => void;
  open(path: string): Promise<void>;
  close(): void;
  reload(): Promise<void>;
  select(id: string | null): void;
  edit(change: (doc: Composition) => Composition, selected?: string | null): void;
  revert(): void;
  undo(): void;
  redo(): void;
  add(entry: NewEquipment): Promise<void>;
  presets(use: string): Promise<Presets>;
  save(): Promise<void>;
}

export interface EditorModule {
  createSession(): EditorSession;
}

export function createEditorModule(gateway: EditorGateway): EditorModule {
  return { createSession: () => new EditorStore(gateway, createOperationIdentity()) };
}
