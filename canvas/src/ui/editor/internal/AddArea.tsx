import { panel } from "../../primitives";
import { FolderPlus } from "lucide-react";
import type { KeyboardEvent, SyntheticEvent } from "react";
import { useState } from "react";
import type { Draft, EditorSession } from "../../../app/editor";
import type { AreaNode } from "../../../model/shared/areaTree";
import { areaNodesOf } from "../../../model/editor";
import { addArea, areaNameProblem } from "../../../model/editor";
import { SubmitRow } from "./SubmitRow";

interface AreaProps {
  readonly draft: Draft;
  readonly editor: EditorSession;
}

interface ParentProps {
  readonly nodes: readonly AreaNode[];
  readonly value: string;
  readonly onChange: (parent: string) => void;
}

function ParentPick({ nodes, value, onChange }: ParentProps) {
  return (
    <label className={panel.field}>
      <span>Inside</span>
      <select
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      >
        <option value="">Top level</option>
        {nodes.map((node) => (
          <option key={node.name} value={node.name}>
            {node.path}
          </option>
        ))}
      </select>
    </label>
  );
}

function holdEnter(event: KeyboardEvent<HTMLFormElement>) {
  if (event.key === "Enter" && event.target instanceof HTMLSelectElement) event.preventDefault();
}

export function AddArea({ draft, editor }: AreaProps) {
  const [name, setName] = useState("");
  const [parent, setParent] = useState("");
  const nodes = areaNodesOf(draft.doc);
  const reason = areaNameProblem(draft.doc, name.trim(), null);
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (reason !== null) return;
    editor.edit((doc) => addArea(doc, name.trim(), parent));
    setName("");
  };
  return (
    <form className={panel.group} onSubmit={submit} onKeyDown={holdEnter} aria-labelledby="add-area-title" noValidate>
      <h3 id="add-area-title">Add area</h3>
      <label className={panel.field}>
        <span>Area name</span>
        <input
          value={name}
          onChange={(event) => {
            setName(event.target.value);
          }}
        />
      </label>
      <ParentPick nodes={nodes} value={parent} onChange={setParent} />
      <SubmitRow id="add-area-reason" reason={reason}>
        <FolderPlus size={14} aria-hidden="true" />
        Add area
      </SubmitRow>
    </form>
  );
}
