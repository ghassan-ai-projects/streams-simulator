import type { BaseChoice, MakeMode } from "../../../app/plant";
import workspaceStyles from "./WorkspaceScreen.module.css";
import styles from "./NewPlant.module.css";

interface BasePickProps {
  readonly choices: readonly BaseChoice[];
  readonly base: BaseChoice | null;
  readonly setBase: (base: BaseChoice) => void;
}

function BaseGroup({
  label,
  group,
  choices,
}: {
  readonly label: string;
  readonly group: BaseChoice["group"];
  readonly choices: readonly BaseChoice[];
}) {
  const shown = choices.filter((choice) => choice.group === group);
  if (shown.length === 0) return null;
  return (
    <optgroup label={label}>
      {shown.map((choice) => (
        <option key={choice.key} value={choice.key}>
          {choice.label}
        </option>
      ))}
    </optgroup>
  );
}

function BasePick({ choices, base, setBase }: BasePickProps) {
  return (
    <div className={styles.newField}>
      <label htmlFor="new-plant-base">Start from</label>
      <select
        id="new-plant-base"
        value={base?.key ?? ""}
        onChange={(event) => {
          const picked = choices.find((choice) => choice.key === event.target.value);
          if (picked !== undefined) setBase(picked);
        }}
      >
        <BaseGroup label="Your plants" group="plants" choices={choices} />
        <BaseGroup label="Examples" group="examples" choices={choices} />
      </select>
    </div>
  );
}

function CopyChoice({ mode, setMode }: { readonly mode: MakeMode; readonly setMode: (mode: MakeMode) => void }) {
  return (
    <fieldset className={styles.choice}>
      <legend className="sr-only">What to keep</legend>
      <label>
        <input
          type="radio"
          name="new-plant-copy"
          checked={mode === "copy"}
          onChange={() => {
            setMode("copy");
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
          checked={mode === "empty"}
          onChange={() => {
            setMode("empty");
          }}
        />
        <span>
          <strong>An empty plant</strong> with the site, areas, profiles and axes of the start plant and no equipment;
          you add the equipment.
        </span>
      </label>
    </fieldset>
  );
}

function BlankNote() {
  return (
    <p className={workspaceStyles.meta}>
      A blank plant: you add its equipment, areas and connections in the editor, then compose and run it.
    </p>
  );
}

export function Choices(
  props: BasePickProps & {
    readonly blank: boolean;
    readonly mode: MakeMode;
    readonly setMode: (mode: MakeMode) => void;
  },
) {
  if (props.blank) return <BlankNote />;
  return (
    <>
      <BasePick {...props} />
      <CopyChoice {...props} />
    </>
  );
}
