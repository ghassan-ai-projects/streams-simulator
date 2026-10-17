import { shared, cx } from "../../primitives";
import type { SyntheticEvent } from "react";
import { useState } from "react";
import type { PlantMaker, PlantSource } from "../../../app/plant";
import { idProblemOf } from "../../../app/plant";
import type { NavigationStore } from "../../../app/shell";
import type { Definition } from "../../../model/workspace";
import type { ExamplePlant } from "../../../model/plant";
import { suggestedId } from "../../../model/workspace";
import { Field } from "./NewPlantField";
import styles from "./HomeScreen.module.css";

export interface NewPlantProps {
  readonly bases: readonly Definition[];
  readonly examples: readonly ExamplePlant[];
  readonly taken: readonly string[];
  readonly maker: PlantMaker;
  readonly navigation: NavigationStore;
  readonly blank?: boolean;
}

function BasePick({
  bases,
  examples,
  base,
  setBase,
}: Pick<NewPlantProps, "bases" | "examples"> & { readonly base: string; readonly setBase: (base: string) => void }) {
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
        {bases.length > 0 && (
          <optgroup label="Your plants">
            {bases.map((item) => (
              <option key={item.path} value={`file:${item.path}`}>
                {item.title || item.id}
              </option>
            ))}
          </optgroup>
        )}
        {examples.length > 0 && (
          <optgroup label="Examples">
            {examples.map((example) => (
              <option key={example.id} value={`example:${example.id}`}>
                {example.title}
              </option>
            ))}
          </optgroup>
        )}
      </select>
    </div>
  );
}

function sourceOf(base: string): PlantSource {
  return base.startsWith("example:") ? { kind: "example", id: base.slice(8) } : { kind: "file", path: base.slice(5) };
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
          <strong>An empty plant</strong> with the site, areas, profiles and axes of the start plant and no equipment;
          you add the equipment.
        </span>
      </label>
    </fieldset>
  );
}

function BlankNote() {
  return (
    <p className={styles.meta}>
      A blank plant: you add its equipment, areas and connections in the editor, then compose and run it.
    </p>
  );
}

function Choices(
  props: NewPlantProps & {
    readonly base: string;
    readonly setBase: (base: string) => void;
    readonly copy: boolean;
    readonly setCopy: (copy: boolean) => void;
  },
) {
  if (props.blank === true) return <BlankNote />;
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
  const create = async (base: string, id: string, title: string, copy: boolean) => {
    const made = await maker.create(sourceOf(base), id, title, copy);
    if ("composition" in made) {
      navigation.go({ screen: "edit", composition: made.composition });
      return;
    }
    reportId(id, idProblemOf(made));
    setFailure(made.diagnostic.message);
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
          className={styles.create}
          aria-disabled={!ready}
          aria-describedby={ready ? undefined : "new-plant-reason"}
        >
          Create and open the editor
        </button>
        <button type="button" className={cx(styles.open, shared.button)} onClick={onCancel}>
          Cancel
        </button>
        {!ready && (
          <span id="new-plant-reason" className={styles.meta}>
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

function useIdProblem(maker: PlantMaker, base: string, name: string, id: string, followAgain: () => void) {
  const [found, setFound] = useState<{ readonly id: string; readonly message: string } | null>(null);
  const reportId = (value: string, message: string) => {
    setFound({ id: value, message });
  };
  const check = async () => {
    const value = id.trim();
    if (value === "") return;
    reportId(value, await maker.idProblem(sourceOf(base), value, name.trim() === "" ? value : name.trim()));
  };
  const leave = () => {
    followAgain();
    void check();
  };
  return { problem: found !== null && found.id === id.trim() ? found.message : "", leave, reportId };
}

function readiness(name: string, id: string, base: string): string {
  if (name.trim() === "") return "Name the plant first.";
  if (id.trim() === "") return "Give it an id.";
  return base === "example:" ? "Choose what to start from." : "";
}

export function NewPlantForm(props: NewPlantProps & { readonly onCancel: () => void }) {
  const { bases, examples, taken, blank = false, onCancel } = props;
  const { name, setName, id, setId, followAgain } = useNaming(taken);
  const first = bases[0] === undefined ? `example:${examples[0]?.id ?? ""}` : `file:${bases[0].path}`;
  const [base, setBase] = useState(first);
  const [copy, setCopy] = useState(!blank);
  const { problem, leave, reportId } = useIdProblem(props.maker, base, name, id, followAgain);
  const { failure, create } = useCreate(props, reportId);
  const reason = readiness(name, id, base);
  const ready = reason === "";
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
      <Field
        label="Id"
        hint="Lower case letters, digits and hyphens, starting with a letter."
        value={id}
        onChange={setId}
        onLeave={leave}
        selectOnFocus
        problem={problem}
      />
      <Choices {...props} base={base} setBase={setBase} copy={copy} setCopy={setCopy} />
      <Actions failure={failure} reason={reason} onCancel={onCancel} />
    </form>
  );
}
