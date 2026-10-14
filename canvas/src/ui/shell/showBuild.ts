import type { EditorStore } from "../../app/editorStore";
import type { PanelStore } from "../../app/panelStore";

export function showBuild(editor: EditorStore, panels: PanelStore): void {
  const state = editor.snapshot();
  if (state.phase === "ready" && state.selected !== null) {
    editor.select(null);
    panels.show("right");
    return;
  }
  panels.toggle("right");
}
