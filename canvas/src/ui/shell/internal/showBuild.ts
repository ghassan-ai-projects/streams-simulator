import type { EditorSession } from "../../../app/editor";
import type { PanelStore } from "../../../app/shell";

export function showBuild(editor: EditorSession, panels: PanelStore): void {
  const state = editor.snapshot();
  if (state.phase === "ready" && state.selected !== null) {
    editor.select(null);
    panels.show("right");
    return;
  }
  panels.toggle("right");
}
