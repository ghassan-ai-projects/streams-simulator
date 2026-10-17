import { holdEnter, panel } from "../../primitives";
import { FolderPlus } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useState } from "react";
import type { Draft, EditorSession } from "../../../app/editor";
import { areaNodesOf } from "../../../model/editor";
import { addArea, areaNameProblem } from "../../../model/editor";
import { Choice } from "./Choice";
import { SubmitRow } from "./SubmitRow";

interface AreaProps {
  readonly draft: Draft;
  readonly editor: EditorSession;
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
      <Choice
        label="Inside"
        value={parent}
        placeholder="Top level"
        options={nodes.map((node) => [node.name, node.path])}
        onChange={setParent}
      />
      <SubmitRow id="add-area-reason" reason={reason}>
        <FolderPlus size={14} aria-hidden="true" />
        Add area
      </SubmitRow>
    </form>
  );
}
