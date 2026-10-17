import { shared } from "../../primitives";
import type { SyntheticEvent } from "react";
import { useState } from "react";
import type { BaseChoice, MakeMode, PlantMaker } from "../../../app/plant";
import { baseChoices, idProblemOf, newPlantReadiness, startingMode } from "../../../app/plant";
import type { NavigationStore } from "../../../app/shell";
import type { Definition } from "../../../model/workspace";
import type { ExamplePlant } from "../../../model/plant";
import { suggestedId } from "../../../model/workspace";
import { Field } from "./NewPlantField";
import workspaceStyles from "./WorkspaceScreen.module.css";
import styles from "./NewPlant.module.css";

export interface NewPlantProps {
  readonly bases: readonly Definition[];
  readonly examples: readonly ExamplePlant[];
  readonly taken: readonly string[];
  readonly maker: PlantMaker;
  readonly navigation: NavigationStore;
  readonly blank?: boolean;
}

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

function Choices(
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

function useCreate(
  { maker, navigation }: Pick<NewPlantProps, "maker" | "navigation">,
  reportId: (id: string, message: string) => void,
) {
  const [failure, setFailure] = useState("");
  const create = async (base: BaseChoice, id: string, title: string, mode: MakeMode) => {
    const made = await maker.create(base.source, id, title, mode);
    if (made.ok) {
      navigation.go({ screen: "edit", composition: made.value.composition });
      return;
    }
    reportId(id, idProblemOf(made.failure));
    setFailure(made.failure.diagnostic.message);
  };
  return { failure, create };
}

interface ActionsProps {
  readonly failure: string;
  readonly reason: string;
  readonly onCancel: () => void;
}

function Actions({ failure, reason, onCancel }: ActionsProps) {
  const ready = reason === "";
  return (
    <>
      <span className={styles.newActions}>
        <button
          type="submit"
          className={shared.primary}
          aria-disabled={!ready}
          aria-describedby={ready ? undefined : "new-plant-reason"}
        >
          Create and open the editor
        </button>
        <button type="button" className={shared.secondary} onClick={onCancel}>
          Cancel
        </button>
        {!ready && (
          <span id="new-plant-reason" className={workspaceStyles.meta}>
            {reason}
          </span>
        )}
      </span>
      {failure !== "" && (
        <p role="alert" className={styles.newFailure}>
          {failure}
        </p>
      )}
    </>
  );
}

function useNaming(taken: readonly string[]) {
  const [name, setName] = useState("");
  const [typed, setTyped] = useState<string | null>(null);
  const id = typed ?? (name.trim() === "" ? "" : suggestedId(name, taken));
  const followAgain = () => {
    if (typed === "") setTyped(null);
  };
  return { name, setName, id, setId: setTyped, followAgain };
}

function useIdProblem(maker: PlantMaker, base: BaseChoice | null, name: string, id: string, followAgain: () => void) {
  const [found, setFound] = useState<{ readonly id: string; readonly message: string } | null>(null);
  const reportId = (value: string, message: string) => {
    setFound({ id: value, message });
  };
  const check = async () => {
    const value = id.trim();
    if (value === "" || base === null) return;
    reportId(value, await maker.idProblem(base.source, value, name.trim() === "" ? value : name.trim()));
  };
  const leave = () => {
    followAgain();
    void check();
  };
  return { problem: found !== null && found.id === id.trim() ? found.message : "", leave, reportId };
}

export function NewPlantForm(props: NewPlantProps & { readonly onCancel: () => void }) {
  const { bases, examples, taken, blank = false, onCancel } = props;
  const { name, setName, id, setId, followAgain } = useNaming(taken);
  const choices = baseChoices(bases, examples);
  const [base, setBase] = useState<BaseChoice | null>(choices[0] ?? null);
  const [mode, setMode] = useState(startingMode({ blank }));
  const { problem, leave, reportId } = useIdProblem(props.maker, base, name, id, followAgain);
  const { failure, create } = useCreate(props, reportId);
  const reason = newPlantReadiness(name, id, base);
  const ready = reason === "";
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (ready && base !== null) void create(base, id.trim(), name.trim(), mode);
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
      <Field
        label="Id"
        hint="Lower case letters, digits and hyphens, starting with a letter."
        value={id}
        onChange={setId}
        onLeave={leave}
        selectOnFocus
        problem={problem}
      />
      <Choices choices={choices} base={base} setBase={setBase} blank={blank} mode={mode} setMode={setMode} />
      <Actions failure={failure} reason={reason} onCancel={onCancel} />
    </form>
  );
}
