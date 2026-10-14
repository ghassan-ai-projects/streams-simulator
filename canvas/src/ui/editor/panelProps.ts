import type { Draft, EditorStore } from "../../app/editorStore";
import type { Plant } from "../../model/plant";

export interface PanelProps {
  readonly draft: Draft;
  readonly plant: Plant;
  readonly editor: EditorStore;
}
