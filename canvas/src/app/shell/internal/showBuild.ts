import type { EditorSession } from "../../editor";
import type { PanelStore } from "./panelStore";

export function showBuild(editor: EditorSession, panels: PanelStore): void {
  const state = editor.snapshot();
  if (state.phase === "ready" && state.selected !== null) {
    editor.select(null);
    panels.show("right");
    return;
  }
  panels.toggle("right");
}
