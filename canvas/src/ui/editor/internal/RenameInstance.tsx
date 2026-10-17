import { TriangleAlert } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useState } from "react";
import { nameCheck, renameInstance } from "../../../model/editor";
import { cx, panel, shared } from "../../primitives";
import type { PanelProps } from "./panelProps";
import { useRequestedFocus } from "./panelProps";
import type { Instance } from "../../../model/plant";
import styles from "./Editor.module.css";

type EditorProps = PanelProps & { readonly instance: Instance };

function NameRule({ check }: { readonly check: ReturnType<typeof nameCheck> }) {
  return (
    <>
      <span />
      <span className={check.fits ? panel.hint : styles.nameError} id="instance-name-rule">
        {check.fits ? null : <TriangleAlert size={13} aria-hidden="true" />}
        {check.hint}
      </span>
    </>
  );
}

export function Rename({
  instance,
  editor,
  request,
  flagged,
}: Pick<EditorProps, "instance" | "editor" | "request"> & { readonly flagged: boolean }) {
  const [name, setName] = useState(instance.id);
  const input = useRequestedFocus<HTMLInputElement>(request, "name");
  const check = nameCheck(name.trim());
  const { fits } = check;
  const ready = fits && name.trim() !== instance.id;
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!ready) return;
    request.ask("name");
    editor.edit((doc) => renameInstance(doc, instance.id, name.trim()), name.trim());
  };
  return (
    <form className={panel.field} onSubmit={submit} noValidate>
      <label htmlFor="instance-name">Name</label>
      <span className={panel.pair}>
        <input
          id="instance-name"
          ref={input}
          aria-invalid={!fits}
          aria-describedby={flagged ? "instance-issues instance-name-rule" : "instance-name-rule"}
          value={name}
          onChange={(event) => {
            setName(event.target.value);
          }}
        />
        <button type="submit" className={cx(shared.secondary, panel.secondary)} aria-disabled={!ready}>
          Rename
        </button>
      </span>
      <NameRule check={check} />
    </form>
  );
}
