import { useEffect } from "react";
import type { EditorSession } from "../../../app/editor";
import { inTextField } from "../../primitives";

function command(event: KeyboardEvent, editor: EditorSession): boolean {
  const key = event.key.toLowerCase();
  if (key === "s") {
    if (editor.savable()) void editor.save();
    return true;
  }
  if (inTextField(event.target)) return false;
  if (key === "z" && event.shiftKey) editor.redo();
  else if (key === "z") editor.undo();
  else if (key === "y") editor.redo();
  else return false;
  return true;
}

export function useEditorKeys(editor: EditorSession) {
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (!(event.ctrlKey || event.metaKey) || event.altKey) return;
      if (command(event, editor)) event.preventDefault();
    };
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("keydown", onKey);
    };
  }, [editor]);
}
