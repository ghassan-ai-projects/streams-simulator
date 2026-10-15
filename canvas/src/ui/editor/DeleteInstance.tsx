import { Trash2 } from "lucide-react";
import { useState } from "react";
import { connectionsOf, endpointOf } from "../../model/composition";
import { removeInstance } from "../../model/compositionEdits";
import styles from "./Editor.module.css";
import type { PanelProps } from "./panelProps";
import type { Instance } from "../../model/plant";

type DeleteProps = Pick<PanelProps, "draft" | "editor" | "request"> & { readonly instance: Instance };

function linkCount(props: DeleteProps): number {
  const id = props.instance.id;
  return connectionsOf(props.draft.doc).filter(
    (wire) => endpointOf(wire.from).id === id || endpointOf(wire.to).id === id,
  ).length;
}

export function DeleteInstance(props: DeleteProps) {
  const { editor, request, instance } = props;
  const [confirming, setConfirming] = useState(false);
  const links = linkCount(props);
  const question = `Delete ${instance.id} and its ${String(links)} ${links === 1 ? "link" : "links"}?`;
  const remove = () => {
    request.ask("build");
    editor.edit((doc) => removeInstance(doc, instance.id), null);
  };
  return (
    <div className={styles.deleteRow}>
      <span className="sr-only" aria-live="polite">
        {confirming ? question : ""}
      </span>
      <button
        type="button"
        className={confirming ? styles.danger : styles.delete}
        title={`Delete ${instance.id}`}
        onBlur={() => {
          setConfirming(false);
        }}
        onClick={() => {
          if (links === 0 || confirming) remove();
          else setConfirming(true);
        }}
      >
        <Trash2 size={14} aria-hidden="true" />
        {confirming ? question : <span className={styles.deleteWords}>{`Delete ${instance.id}`}</span>}
      </button>
    </div>
  );
}
