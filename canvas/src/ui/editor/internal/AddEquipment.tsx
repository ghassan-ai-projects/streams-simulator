import { panel } from "../../primitives";
import { Plus } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useEffect, useState } from "react";
import type { Draft, EditorSession } from "../../../app/editor";
import { areaNodesOf, suggestedName } from "../../../model/editor";
import { addProblem, instanceNameRule, libraryRows, libraryWords, suggestedWords } from "../../../model/editor";
import type { LibraryEntry, Presets } from "../../../model/plant";
import { searchLibrary } from "../../../model/plant";
import type { Plant } from "../../../model/plant";
import { Choice } from "./Choice";
import styles from "./Editor.module.css";
import type { FocusRequest } from "./panelProps";
import { SubmitRow } from "./SubmitRow";

interface AddProps {
  readonly draft: Draft;
  readonly plant: Plant;
  readonly editor: EditorSession;
  readonly request: FocusRequest;
}

const noPresets: Presets = { options: [], preferred: "", description: "" };

function usePresets(editor: EditorSession, use: string) {
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

interface NameProps {
  readonly typed: string;
  readonly suggested: string;
  readonly setTyped: (name: string) => void;
}

function NameField({ typed, suggested, setTyped }: NameProps) {
  return (
    <>
      <label className={panel.field}>
        <span>Name</span>
        <input
          value={typed}
          placeholder={suggested}
          aria-describedby="add-name-rule"
          onChange={(event) => {
            setTyped(event.target.value);
          }}
        />
      </label>
      <p className={panel.hint} id="add-name-rule">
        {`${suggestedWords(suggested)} ${instanceNameRule()}`}
      </p>
    </>
  );
}

interface PickProps {
  readonly library: readonly LibraryEntry[];
  readonly use: string;
  readonly setUse: (use: string) => void;
}

function EquipmentList({
  found,
  use,
  setUse,
}: { readonly found: readonly LibraryEntry[] } & Omit<PickProps, "library">) {
  const shown = libraryRows(found.length);
  return (
    <select
      aria-label="Equipment"
      aria-describedby="equipment-count"
      size={shown}
      value={use}
      onChange={(event) => {
        setUse(event.target.value);
      }}
      onKeyDown={(event) => {
        if (event.key === "Enter") event.preventDefault();
      }}
    >
      <option value="" hidden />
      {found.map((entry) => (
        <option key={entry.ref} value={entry.ref}>
          {entry.title}
        </option>
      ))}
    </select>
  );
}

function LibraryPick({ library, use, setUse }: PickProps) {
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
        onKeyDown={(event) => {
          if (event.key === "Enter") event.preventDefault();
        }}
      />
      <EquipmentList found={found} use={use} setUse={setUse} />
      <p className={panel.hint} id="equipment-count">
        {libraryWords(found.length, libraryRows(found.length), query)}
      </p>
    </div>
  );
}

function sizes(presets: Presets): readonly (readonly [string, string])[] {
  return [["", "Default"], ...presets.options.map((option) => [option.name, option.title] as const)];
}

function useNaming(draft: Draft, use: string) {
  const [typed, setTyped] = useState("");
  const title = draft.library.find((entry) => entry.ref === use)?.title ?? "";
  const suggested = use === "" ? "" : suggestedName(draft.doc, use, title);
  return {
    typed,
    suggested,
    name: typed.trim() === "" ? suggested : typed.trim(),
    setTyped,
    reset: () => {
      setTyped("");
    },
  };
}

function useArea(draft: Draft) {
  const nodes = areaNodesOf(draft.doc);
  const [picked, setArea] = useState("");
  const area = nodes.some((node) => node.name === picked) ? picked : (nodes[0]?.name ?? "");
  return { area, setArea, options: nodes.map((node) => [node.name, node.path] as const) };
}

export function AddEquipment({ draft, plant, editor, request }: AddProps) {
  const [use, setUse] = useState("");
  const { typed, suggested, name, setTyped, reset } = useNaming(draft, use);
  const { area, setArea, options } = useArea(draft);
  const { presets, preset, setPreset } = usePresets(editor, use);
  const problem = addProblem(use, name, area, plant);
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (problem !== null) return;
    request.ask("instance");
    void editor.add({ use, preset, name, area });
    reset();
  };
  return (
    <form className={panel.group} onSubmit={submit} aria-labelledby="add-title" noValidate>
      <h3 id="add-title">Add equipment</h3>
      <LibraryPick library={draft.library} use={use} setUse={setUse} />
      {presets.description !== "" && <p className={styles.describe}>{presets.description}</p>}
      <Choice named label="Size" value={preset} options={sizes(presets)} onChange={setPreset} />
      <NameField typed={typed} suggested={suggested} setTyped={setTyped} />
      <Choice named label="Area" value={area} options={options} onChange={setArea} />
      <p className={panel.hint}>Areas are listed in order, with the site first. Add more under Add area.</p>
      <SubmitRow id="add-reason" reason={problem}>
        <Plus size={14} aria-hidden="true" />
        Add to the plant
      </SubmitRow>
    </form>
  );
}
