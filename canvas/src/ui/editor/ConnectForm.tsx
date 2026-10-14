import { Cable } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useState } from "react";
import { connect } from "../../model/compositionEdits";
import type { Instance, Plant } from "../../model/plant";
import type { ComponentSpec } from "../../model/ports";
import { portsMatching, portType } from "../../model/ports";
import panel from "../panel.module.css";
import { SubmitRow } from "./SubmitRow";
import type { PanelProps } from "./panelProps";

type FormProps = PanelProps & { readonly instance: Instance };

function specOf(specs: ReadonlyMap<string, ComponentSpec>, instance: Instance | undefined): ComponentSpec | undefined {
  return instance?.component ? specs.get(instance.component) : undefined;
}

function Pick({
  label,
  value,
  options,
  onChange,
}: {
  readonly label: string;
  readonly value: string;
  readonly options: readonly string[];
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
        <option value="">Choose…</option>
        {options.map((option) => (
          <option key={option} value={option}>
            {option.replaceAll("_", " ")}
          </option>
        ))}
      </select>
    </label>
  );
}

function targets(plant: Plant, specs: ReadonlyMap<string, ComponentSpec>, id: string, type: string): readonly string[] {
  return plant.instances
    .filter((entry) => entry.id !== id && portsMatching(specOf(specs, entry), "in", type).length > 0)
    .map((entry) => entry.id);
}

interface Choice {
  readonly output: string;
  readonly target: string;
  readonly input: string;
}

const none: Choice = { output: "", target: "", input: "" };

interface PicksProps extends Omit<FormProps, "editor"> {
  readonly choice: Choice;
  readonly setChoice: (choice: Choice) => void;
}

function LinkPicks({ draft, plant, instance, choice, setChoice }: PicksProps) {
  const type = portType(specOf(draft.specs, instance), choice.output);
  const targetInstance = plant.instances.find((entry) => entry.id === choice.target);
  return (
    <>
      <Pick
        label="Output"
        value={choice.output}
        options={portsMatching(specOf(draft.specs, instance), "out")}
        onChange={(output) => {
          setChoice({ ...none, output });
        }}
      />
      <Pick
        label="To"
        value={choice.target}
        options={targets(plant, draft.specs, instance.id, type)}
        onChange={(target) => {
          setChoice({ ...choice, target, input: "" });
        }}
      />
      <Pick
        label="Input"
        value={choice.input}
        options={portsMatching(specOf(draft.specs, targetInstance), "in", type)}
        onChange={(input) => {
          setChoice({ ...choice, input });
        }}
      />
    </>
  );
}

export function ConnectForm(props: FormProps) {
  const { editor, instance } = props;
  const [choice, setChoice] = useState(none);
  const [said, setSaid] = useState("");
  const ready = choice.output !== "" && choice.target !== "" && choice.input !== "";
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!ready) return;
    editor.edit((doc) => connect(doc, `${instance.id}.${choice.output}`, `${choice.target}.${choice.input}`));
    setSaid(`Linked ${choice.output.replaceAll("_", " ")} to ${choice.target} ${choice.input.replaceAll("_", " ")}`);
    setChoice({ ...choice, input: "" });
  };
  return (
    <form className={panel.group} onSubmit={submit} aria-labelledby="connect-title" noValidate>
      <h3 id="connect-title">Link an output</h3>
      <p className="sr-only" aria-live="polite">
        {said}
      </p>
      <LinkPicks {...props} choice={choice} setChoice={setChoice} />
      <SubmitRow
        id="link-reason"
        reason={ready ? null : "Choose an output, then an instance and an input that take what it gives."}
      >
        <Cable size={14} aria-hidden="true" />
        Link
      </SubmitRow>
    </form>
  );
}
