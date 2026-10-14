import type { SyntheticEvent } from "react";
import { useId, useState } from "react";
import type { EditorStore } from "../../app/editorStore";
import type { NavigationStore } from "../../app/navigationStore";
import type { Definition } from "../../model/bridge";
import { suggestedId } from "../../model/workspace";
import { useFocusOnMount } from "../useFocusOnMount";
import styles from "./HomeScreen.module.css";

export interface NewPlantProps {
  readonly bases: readonly Definition[];
  readonly taken: readonly string[];
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
      <label htmlFor="new-plant-base">Start from</label>
      <select
        id="new-plant-base"
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
    </div>
  );
}

function CopyChoice({ copy, setCopy }: { readonly copy: boolean; readonly setCopy: (copy: boolean) => void }) {
  return (
    <fieldset className={styles.choice}>
      <legend className="sr-only">What to keep</legend>
      <label>
        <input
          type="radio"
          name="new-plant-copy"
          checked={copy}
          onChange={() => {
            setCopy(true);
          }}
        />
        <span>
          <strong>A full copy</strong> of its equipment, connections and settings, ready to run.
        </span>
      </label>
      <label>
        <input
          type="radio"
          name="new-plant-copy"
          checked={!copy}
          onChange={() => {
            setCopy(false);
          }}
        />
        <span>
          <strong>An empty plant</strong> with only its site, areas, profiles and axes; you add the equipment.
        </span>
      </label>
    </fieldset>
  );
}

function useCreate({ editor, navigation }: Pick<NewPlantProps, "editor" | "navigation">) {
  const [failure, setFailure] = useState("");
  const create = async (base: string, id: string, title: string, copy: boolean) => {
    const made = await editor.create(base, id, title, copy);
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
          Create and open the editor
        </button>
        <button type="button" className={styles.open} onClick={onCancel}>
          Cancel
        </button>
      </span>
    </>
  );
}

function useNaming(taken: readonly string[]) {
  const [name, setName] = useState("");
  const [typed, setTyped] = useState<string | null>(null);
  const id = typed ?? (name.trim() === "" ? "" : suggestedId(name, taken));
  return { name, setName, id, setId: setTyped };
}

export function NewPlantForm(props: NewPlantProps & { readonly onCancel: () => void }) {
  const { bases, taken, onCancel } = props;
  const { name, setName, id, setId } = useNaming(taken);
  const [base, setBase] = useState(bases[0]?.path ?? "");
  const [copy, setCopy] = useState(true);
  const { failure, create } = useCreate(props);
  const ready = id.trim() !== "" && name.trim() !== "" && base !== "";
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (ready) void create(base, id.trim(), name.trim(), copy);
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
      <Field first label="Name" placeholder="e.g. Cooling hall" value={name} onChange={setName} />
      <Field label="Id" hint="Used for its file names: lower case, digits and hyphens." value={id} onChange={setId} />
      <BasePick bases={bases} base={base} setBase={setBase} />
      <CopyChoice copy={copy} setCopy={setCopy} />
      <Actions failure={failure} ready={ready} onCancel={onCancel} />
    </form>
  );
}
