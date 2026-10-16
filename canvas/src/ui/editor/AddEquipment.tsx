import { Plus } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useEffect, useState } from "react";
import type { Draft, EditorStore } from "../../app/editorStore";
import { areaNodesOf, suggestedName } from "../../model/composition";
import { instanceNameFits, instanceNameRule } from "../../model/instanceNameRule";
import type { LibraryEntry, Presets } from "../../model/library";
import { searchLibrary } from "../../model/library";
import type { Plant } from "../../model/plant";
import { instanceIdTaken } from "../../model/plant";
import panel from "../panel.module.css";
import styles from "./Editor.module.css";
import type { FocusRequest } from "./panelProps";
import { SubmitRow } from "./SubmitRow";

interface AddProps {
  readonly draft: Draft;
  readonly plant: Plant;
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
        aria-label={label}
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

interface NameProps {
  readonly typed: string;
  readonly suggested: string;
  readonly setTyped: (name: string) => void;
}

function suggestedWords(suggested: string): string {
  return suggested === ""
    ? "Leave empty to use the suggested name."
    : `Leave empty to use the suggested name, ${suggested}.`;
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

function libraryWords(found: number, shown: number): string {
  const counted = `${String(found)} ${found === 1 ? "component" : "components"}`;
  return found > shown ? `${counted} · scroll the list for the rest` : counted;
}

function LibraryPick({ library, use, setUse }: PickProps) {
  const [query, setQuery] = useState("");
  const found = searchLibrary(library, query);
  const shown = Math.min(6, Math.max(2, found.length));
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
      <select
        aria-label="Equipment"
        aria-describedby="equipment-count"
        size={shown}
        value={use}
        onChange={(event) => {
          setUse(event.target.value);
        }}
      >
        <option value="" hidden />
        {found.map((entry) => (
          <option key={entry.ref} value={entry.ref}>
            {entry.title}
          </option>
        ))}
      </select>
      <p className={panel.hint} id="equipment-count">
        {found.length === 0 ? `Nothing in the library matches "${query.trim()}".` : libraryWords(found.length, shown)}
      </p>
    </div>
  );
}

function sizes(presets: Presets): readonly (readonly [string, string])[] {
  return [["", "Default"], ...presets.options.map((option) => [option.name, option.title] as const)];
}

function missing(use: string, name: string, area: string, plant: Plant): string | null {
  if (use === "") return "Choose equipment from the library.";
  if (name === "") return "Give the instance a name.";
  if (!instanceNameFits(name)) return `"${name}" is not a valid name. ${instanceNameRule()}`;
  if (instanceIdTaken(plant, name)) return `${name} is already used by another instance.`;
  return area === "" ? "Choose an area." : null;
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
  const problem = missing(use, name, area, plant);
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
      <Choice label="Size" value={preset} options={sizes(presets)} onChange={setPreset} />
      <NameField typed={typed} suggested={suggested} setTyped={setTyped} />
      <Choice label="Area" value={area} options={options} onChange={setArea} />
      <p className={panel.hint}>Areas are listed in order, with the site first. Add more under Add area.</p>
      <SubmitRow id="add-reason" reason={problem}>
        <Plus size={14} aria-hidden="true" />
        Add to the plant
      </SubmitRow>
    </form>
  );
}
