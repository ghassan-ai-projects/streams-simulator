import { Redo, Undo } from "lucide-react";
import { useState } from "react";
import type { Draft, EditorSession } from "../../../app/editor";
import bar from "./EditorBar.module.css";

function useHistoryAnnouncement(draft: Draft): string {
  const counts = { past: draft.past.length, future: draft.future.length };
  const [seen, setSeen] = useState(counts);
  const [said, setSaid] = useState("");
  if (seen.past !== counts.past || seen.future !== counts.future) {
    setSeen(counts);
    const left = `${String(counts.past)} more to undo`;
    if (counts.future === seen.future + 1) setSaid(`Undone. ${left}.`);
    else if (counts.future === seen.future - 1) setSaid(`Redone. ${left}.`);
    else setSaid("");
  }
  return said;
}

export function History({ draft, editor }: { readonly draft: Draft; readonly editor: EditorSession }) {
  const mac = navigator.platform.toLowerCase().includes("mac");
  const key = mac ? "⌘" : "Ctrl+";
  const said = useHistoryAnnouncement(draft);
  return (
    <span className={bar.history}>
      <span className="sr-only" aria-live="polite">
        {said}
      </span>
      <button
        type="button"
        className={bar.icon}
        aria-label="Undo"
        aria-keyshortcuts="Control+Z Meta+Z"
        title={`Undo (${key}Z)`}
        aria-disabled={draft.past.length === 0}
        onClick={() => {
          editor.undo();
        }}
      >
        <Undo size={14} aria-hidden="true" />
      </button>
      <button
        type="button"
        className={bar.icon}
        aria-label="Redo"
        aria-keyshortcuts="Control+Shift+Z Meta+Shift+Z"
        title={`Redo (${key}Shift+Z)`}
        aria-disabled={draft.future.length === 0}
        onClick={() => {
          editor.redo();
        }}
      >
        <Redo size={14} aria-hidden="true" />
      </button>
    </span>
  );
}
