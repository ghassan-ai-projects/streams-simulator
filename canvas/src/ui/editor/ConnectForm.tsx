import { Cable } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useState } from "react";
import { connect } from "../../model/compositionEdits";
import type { Instance, Plant } from "../../model/plant";
import { kindLabel } from "../../model/shared/words";
import type { ComponentSpec } from "../../model/ports";
import { noTargetReason, portsMatching, portType, specOf } from "../../model/ports";
import panel from "../panel.module.css";
import { SubmitRow } from "./SubmitRow";
import type { PanelProps } from "./panelProps";

type FormProps = PanelProps & { readonly instance: Instance };

function Pick({
  label,
  value,
  options,
  empty,
  words,
  onChange,
}: {
  readonly label: string;
  readonly value: string;
  readonly options: readonly string[];
  readonly empty: string;
  readonly words: (option: string) => string;
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
        <option value="">{options.length === 0 ? empty : "Choose…"}</option>
        {options.map((option) => (
          <option key={option} value={option}>
            {words(option)}
          </option>
        ))}
      </select>
    </label>
  );
}

const portWords = (port: string) => port.replaceAll("_", " ");

function instanceWords(plant: Plant, id: string): string {
  const instance = plant.instances.find((entry) => entry.id === id);
  return instance === undefined ? id : `${instance.id} · ${kindLabel(instance.kind)}`;
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

function withOutput(
  draft: PicksProps["draft"],
  plant: Plant,
  instance: Instance,
  choice: Choice,
  output: string,
): Choice {
  const type = portType(specOf(draft.specs, instance), output);
  const keeps = targets(plant, draft.specs, instance.id, type).includes(choice.target);
  const aimed = plant.instances.find((entry) => entry.id === choice.target);
  const inputs = portsMatching(specOf(draft.specs, aimed), "in", type);
  return {
    output,
    target: keeps ? choice.target : "",
    input: keeps && inputs.includes(choice.input) ? choice.input : "",
  };
}

function TargetNote({
  plant,
  draft,
  instance,
  output,
  count,
}: {
  readonly plant: Plant;
  readonly draft: PicksProps["draft"];
  readonly instance: Instance;
  readonly output: string;
  readonly count: number;
}) {
  if (output === "" || count > 0) return null;
  return <p className={panel.hint}>{noTargetReason(plant, draft.specs, instance, output)}</p>;
}

type TargetProps = Pick<PicksProps, "draft" | "plant" | "instance" | "choice" | "setChoice">;

function TargetPicks({ draft, plant, instance, choice, setChoice }: TargetProps) {
  const type = portType(specOf(draft.specs, instance), choice.output);
  const targetInstance = plant.instances.find((entry) => entry.id === choice.target);
  const options = targets(plant, draft.specs, instance.id, type);
  const noTarget = choice.output === "" ? "Choose an output first" : "No instance takes it";
  return (
    <>
      <Pick
        label="To"
        value={choice.target}
        options={options}
        empty={noTarget}
        words={(id) => instanceWords(plant, id)}
        onChange={(target) => {
          setChoice({ ...choice, target, input: "" });
        }}
      />
      <TargetNote plant={plant} draft={draft} instance={instance} output={choice.output} count={options.length} />
      <Pick
        label="Input"
        value={choice.input}
        options={portsMatching(specOf(draft.specs, targetInstance), "in", type)}
        empty={choice.target === "" ? "Choose an instance first" : "No matching input"}
        words={portWords}
        onChange={(input) => {
          setChoice({ ...choice, input });
        }}
      />
    </>
  );
}

function LinkPicks({ draft, plant, instance, choice, setChoice }: PicksProps) {
  return (
    <>
      <Pick
        label="Output"
        value={choice.output}
        options={portsMatching(specOf(draft.specs, instance), "out")}
        empty="This instance has no outputs"
        words={portWords}
        onChange={(output) => {
          setChoice(withOutput(draft, plant, instance, choice, output));
        }}
      />
      <TargetPicks draft={draft} plant={plant} instance={instance} choice={choice} setChoice={setChoice} />
    </>
  );
}

export function ConnectForm(props: FormProps) {
  const { editor, instance, plant } = props;
  const [choice, setChoice] = useState(none);
  const [said, setSaid] = useState("");
  const ready = choice.output !== "" && choice.target !== "" && choice.input !== "";
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!ready) return;
    editor.edit((doc) => connect(doc, `${instance.id}.${choice.output}`, `${choice.target}.${choice.input}`));
    setSaid(`Linked ${portWords(choice.output)} to ${instanceWords(plant, choice.target)} ${portWords(choice.input)}`);
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
