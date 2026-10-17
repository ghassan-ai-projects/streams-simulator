import { Undo2 } from "lucide-react";
import { useState } from "react";
import type { EditorSession } from "../../../app/editor";
import { shared } from "../../primitives";
import styles from "./Editor.module.css";

export function Revert({ disabled, editor }: { readonly disabled: boolean; readonly editor: EditorSession }) {
  const [confirming, setConfirming] = useState(false);
  return (
    <>
      <span className="sr-only" aria-live="polite">
        {confirming ? "Press Discard changes again to drop every unsaved change" : ""}
      </span>
      <RevertButton disabled={disabled} editor={editor} confirming={confirming} setConfirming={setConfirming} />
    </>
  );
}

function RevertButton({
  disabled,
  editor,
  confirming,
  setConfirming,
}: {
  readonly disabled: boolean;
  readonly editor: EditorSession;
  readonly confirming: boolean;
  readonly setConfirming: (confirming: boolean) => void;
}) {
  return (
    <button
      type="button"
      className={confirming ? styles.danger : shared.secondary}
      aria-disabled={disabled}
      onBlur={() => {
        setConfirming(false);
      }}
      onClick={() => {
        if (disabled) return;
        if (confirming) editor.revert();
        setConfirming(!confirming);
      }}
    >
      <Undo2 size={14} aria-hidden="true" />
      {confirming ? "Discard changes" : "Revert"}
    </button>
  );
}
