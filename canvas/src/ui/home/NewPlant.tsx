import { Plus } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useId, useState } from "react";
import type { EditorStore } from "../../app/editorStore";
import type { NavigationStore } from "../../app/navigationStore";
import type { Definition } from "../../model/bridge";
import { cx } from "../cx";
import { useFocusOnMount } from "../useFocusOnMount";
import styles from "./HomeScreen.module.css";

interface NewPlantProps {
  readonly bases: readonly Definition[];
  readonly editor: EditorStore;
  readonly navigation: NavigationStore;
}

function Field({
  label,
  hint,
  value,
  onChange,
  first = false,
  placeholder,
}: {
  readonly label: string;
  readonly hint?: string;
  readonly value: string;
  readonly onChange: (value: string) => void;
  readonly first?: boolean;
  readonly placeholder?: string;
}) {
  const id = useId();
  const input = useFocusOnMount<HTMLInputElement>(first);
  return (
    <div className={styles.newField}>
      <label htmlFor={id}>{label}</label>
      <input
        id={id}
        ref={input}
        placeholder={placeholder}
        value={value}
        aria-describedby={hint === undefined ? undefined : `${id}-hint`}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      />
      {hint !== undefined && <small id={`${id}-hint`}>{hint}</small>}
    </div>
  );
}

function BasePick({
  bases,
  base,
  setBase,
}: {
  readonly bases: readonly Definition[];
  readonly base: string;
  readonly setBase: (base: string) => void;
}) {
  return (
    <div className={styles.newField}>
      <label htmlFor="new-plant-base">Based on</label>
      <select
        id="new-plant-base"
        aria-describedby="new-plant-base-hint"
        value={base}
        onChange={(event) => {
          setBase(event.target.value);
        }}
      >
        {bases.map((item) => (
          <option key={item.path} value={item.path}>
            {item.title || item.id}
          </option>
        ))}
      </select>
      <small id="new-plant-base-hint">Copies its site, areas, profiles and axes; starts with no equipment.</small>
    </div>
  );
}

function useCreate({ editor, navigation }: Omit<NewPlantProps, "bases">) {
  const [failure, setFailure] = useState("");
  const create = async (base: string, id: string, title: string) => {
    const made = await editor.create(base, id, title);
    if (typeof made === "string") navigation.go({ screen: "edit", composition: made });
    else setFailure(made.message);
  };
  return { failure, create };
}

function Actions({
  failure,
  ready,
  onCancel,
}: {
  readonly failure: string;
  readonly ready: boolean;
  readonly onCancel: () => void;
}) {
  return (
    <>
      {failure !== "" && (
        <p role="alert" className={styles.failure}>
          {failure}
        </p>
      )}
      <span className={styles.newActions}>
        <button type="submit" className={styles.create} aria-disabled={!ready}>
          Create and edit
        </button>
        <button type="button" className={styles.open} onClick={onCancel}>
          Cancel
        </button>
      </span>
    </>
  );
}

function NewPlantForm({ bases, editor, navigation, onCancel }: NewPlantProps & { readonly onCancel: () => void }) {
  const [id, setId] = useState("");
  const [title, setTitle] = useState("");
  const [base, setBase] = useState(bases[0]?.path ?? "");
  const { failure, create } = useCreate({ editor, navigation });
  const ready = id.trim() !== "" && title.trim() !== "" && base !== "";
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (ready) void create(base, id.trim(), title.trim());
  };
  return (
    <form
      className={styles.newPlant}
      onSubmit={submit}
      onKeyDown={(event) => {
        if (event.key === "Escape") onCancel();
      }}
      aria-label="New plant"
      noValidate
    >
      <Field
        first
        label="Id"
        placeholder="e.g. cooling-hall"
        hint="Lower case, digits and hyphens, at least three characters."
        value={id}
        onChange={setId}
      />
      <Field label="Title" value={title} onChange={setTitle} />
      <BasePick bases={bases} base={base} setBase={setBase} />
      <Actions failure={failure} ready={ready} onCancel={onCancel} />
    </form>
  );
}

export function NewPlant(props: NewPlantProps) {
  const [open, setOpen] = useState(false);
  const [closed, setClosed] = useState(0);
  const button = useFocusOnMount<HTMLButtonElement>(closed);
  if (open && props.bases.length > 0) {
    return (
      <NewPlantForm
        {...props}
        onCancel={() => {
          setOpen(false);
          setClosed((count) => count + 1);
        }}
      />
    );
  }
  return (
    <button
      type="button"
      ref={button}
      className={cx(styles.open, styles.newButton)}
      aria-disabled={props.bases.length === 0}
      onClick={() => {
        setOpen(props.bases.length > 0);
      }}
    >
      <Plus size={14} aria-hidden="true" />
      New plant
    </button>
  );
}
