import { Plus } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useEffect, useState } from "react";
import type { Draft, EditorStore } from "../../app/editorStore";
import { areasOf, suggestedName } from "../../model/composition";
import type { LibraryEntry, Presets } from "../../model/library";
import { searchLibrary } from "../../model/library";
import panel from "../panel.module.css";
import styles from "./Editor.module.css";
import type { FocusRequest } from "./panelProps";
import { SubmitRow } from "./SubmitRow";

interface AddProps {
  readonly draft: Draft;
  readonly editor: EditorStore;
  readonly request: FocusRequest;
}

const noPresets: Presets = { options: [], preferred: "", description: "" };

function usePresets(editor: EditorStore, use: string) {
  const [presets, setPresets] = useState<Presets>(noPresets);
  const [preset, setPreset] = useState("");
  useEffect(() => {
    let current = true;
    if (use === "") return undefined;
    void editor.presets(use).then((found) => {
      if (!current) return;
      setPresets(found);
      setPreset(found.preferred);
    });
    return () => {
      current = false;
    };
  }, [editor, use]);
  return { presets: use === "" ? noPresets : presets, preset, setPreset };
}

function Choice({
  label,
  value,
  options,
  onChange,
}: {
  readonly label: string;
  readonly value: string;
  readonly options: readonly (readonly [string, string])[];
  readonly onChange: (value: string) => void;
}) {
  return (
    <label className={panel.field}>
      <span>{label}</span>
      <select
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      >
        {options.map(([key, text]) => (
          <option key={key} value={key}>
            {text}
          </option>
        ))}
      </select>
    </label>
  );
}

function NameField({ name, setName }: { readonly name: string; readonly setName: (name: string) => void }) {
  return (
    <label className={panel.field}>
      <span>Name</span>
      <input
        value={name}
        onChange={(event) => {
          setName(event.target.value);
        }}
      />
    </label>
  );
}

function LibraryPick({
  library,
  use,
  setUse,
}: {
  readonly library: readonly LibraryEntry[];
  readonly use: string;
  readonly setUse: (use: string) => void;
}) {
  const [query, setQuery] = useState("");
  const found = searchLibrary(library, query);
  return (
    <div className={styles.library}>
      <input
        type="search"
        aria-label="Find equipment"
        placeholder="Find equipment: pump, chiller, meter…"
        value={query}
        onChange={(event) => {
          setQuery(event.target.value);
        }}
      />
      <select
        aria-label="Equipment"
        size={6}
        value={use}
        onChange={(event) => {
          setUse(event.target.value);
        }}
      >
        {found.map((entry) => (
          <option key={entry.ref} value={entry.ref}>
            {entry.title}
          </option>
        ))}
      </select>
      {found.length === 0 && <p className={panel.hint}>{`Nothing in the library matches "${query.trim()}".`}</p>}
    </div>
  );
}

function sizes(presets: Presets): readonly (readonly [string, string])[] {
  return [["", "Default"], ...presets.options.map((option) => [option.name, option.title] as const)];
}

function missing(use: string, name: string, area: string): string | null {
  if (use === "") return "Choose equipment from the library.";
  if (name.trim() === "") return "Give the instance a name.";
  return area === "" ? "Choose an area." : null;
}

function useNaming(draft: Draft, use: string) {
  const [typed, setTyped] = useState<string | null>(null);
  const title = draft.library.find((entry) => entry.ref === use)?.title ?? "";
  const name = typed ?? (use === "" ? "" : suggestedName(draft.doc, use, title));
  return {
    name,
    setName: setTyped,
    reset: () => {
      setTyped(null);
    },
  };
}

export function AddEquipment({ draft, editor, request }: AddProps) {
  const [use, setUse] = useState("");
  const { name, setName, reset } = useNaming(draft, use);
  const areas = areasOf(draft.doc);
  const [area, setArea] = useState(areas[0] ?? "");
  const { presets, preset, setPreset } = usePresets(editor, use);
  const problem = missing(use, name, area);
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (problem !== null) return;
    request.ask("instance");
    void editor.add({ use, preset, name: name.trim(), area });
    reset();
  };
  return (
    <form className={panel.group} onSubmit={submit} aria-labelledby="add-title" noValidate>
      <h3 id="add-title">Add equipment</h3>
      <LibraryPick
        library={draft.library}
        use={use}
        setUse={(next) => {
          setUse(next);
          reset();
        }}
      />
      {presets.description !== "" && <p className={styles.describe}>{presets.description}</p>}
      <Choice label="Size" value={preset} options={sizes(presets)} onChange={setPreset} />
      <NameField name={name} setName={setName} />
      <Choice label="Area" value={area} options={areas.map((entry) => [entry, entry] as const)} onChange={setArea} />
      <SubmitRow id="add-reason" reason={problem}>
        <Plus size={14} aria-hidden="true" />
        Add to the plant
      </SubmitRow>
    </form>
  );
}
