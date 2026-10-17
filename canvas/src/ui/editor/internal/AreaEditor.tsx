import { panel } from "../../primitives";
import type { SyntheticEvent } from "react";
import { useId, useState } from "react";
import type { Draft, EditorSession } from "../../../app/editor";
import type { AreaNode } from "../../../model/shared/areaTree";
import { areaNodesOf } from "../../../model/editor";
import { areaInUse, areaNameProblem, removeArea, renameArea } from "../../../model/editor";
import styles from "./Editor.module.css";

interface RowProps {
  readonly draft: Draft;
  readonly editor: EditorSession;
  readonly node: AreaNode;
}

function Reason({ id, text }: { readonly id: string; readonly text: string | null }) {
  if (text === null) return null;
  return (
    <p className={styles.reason} id={id}>
      {text}
    </p>
  );
}

function DeleteArea({
  editor,
  node,
  blocked,
  id,
}: Omit<RowProps, "draft"> & { readonly blocked: string | null; readonly id: string }) {
  return (
    <button
      type="button"
      className={panel.secondary}
      aria-label={`Delete ${node.path}`}
      aria-disabled={blocked !== null}
      aria-describedby={blocked === null ? undefined : `${id}-blocked`}
      onClick={() => {
        if (blocked === null) editor.edit((doc) => removeArea(doc, node.name));
      }}
    >
      Delete
    </button>
  );
}

function AreaRow({ draft, editor, node }: RowProps) {
  const [name, setName] = useState(node.name);
  const next = name.trim();
  const reason = areaNameProblem(draft.doc, next, node.name);
  const blocked = areaInUse(draft.doc, node.name);
  const id = useId();
  const ready = reason === null && next !== node.name;
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (ready) editor.edit((doc) => renameArea(doc, node.name, next));
  };
  return (
    <li>
      <form className={styles.areaRow} onSubmit={submit} noValidate>
        <span className={styles.areaPath}>{node.path}</span>
        <span className={panel.pair}>
          <input
            aria-label={`Name of ${node.path}`}
            aria-invalid={reason !== null}
            aria-describedby={reason === null ? undefined : `${id}-reason`}
            value={name}
            onChange={(event) => {
              setName(event.target.value);
            }}
          />
          <button type="submit" className={panel.secondary} aria-disabled={!ready}>
            Rename
          </button>
          <DeleteArea editor={editor} node={node} blocked={blocked} id={id} />
        </span>
        <Reason id={`${id}-reason`} text={reason} />
        <Reason id={`${id}-blocked`} text={blocked} />
      </form>
    </li>
  );
}

export function AreaEditor({ draft, editor }: { readonly draft: Draft; readonly editor: EditorSession }) {
  const nodes = areaNodesOf(draft.doc);
  return (
    <section className={panel.group} aria-labelledby="areas-title">
      <h3 id="areas-title">Areas</h3>
      <p className={panel.hint}>An area can be deleted when no instance is in it and no area is inside it.</p>
      <ul className={styles.areaList}>
        {nodes.map((node) => (
          <AreaRow key={node.name} draft={draft} editor={editor} node={node} />
        ))}
      </ul>
    </section>
  );
}
