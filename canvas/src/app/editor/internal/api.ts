import type { Composition, NewEquipment } from "../../../model/editor";
import type { Presets } from "../../../model/plant";
import type { Store } from "../../shared";
import type { EditorState } from "./editorDraft";

export type { Draft } from "./editorDraft";

export interface EditorSession extends Store<EditorState> {
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
