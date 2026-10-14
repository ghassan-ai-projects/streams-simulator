import { Cable } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useState } from "react";
import { connect } from "../../model/compositionEdits";
import type { Instance, Plant } from "../../model/plant";
import type { ComponentSpec } from "../../model/ports";
import panel from "../panel.module.css";
import type { PanelProps } from "./panelProps";

type FormProps = PanelProps & { readonly instance: Instance };

function portsOf(
  specs: ReadonlyMap<string, ComponentSpec>,
  instance: Instance | undefined,
  direction: "in" | "out",
): readonly string[] {
  const spec = instance?.component ? specs.get(instance.component) : undefined;
  return (spec?.ports ?? []).filter((port) => port.direction === direction).map((port) => port.name);
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
            {option}
          </option>
        ))}
      </select>
    </label>
  );
}

function others(plant: Plant, id: string): readonly string[] {
  return plant.instances.map((entry) => entry.id).filter((other) => other !== id);
}

export function ConnectForm({ draft, plant, editor, instance }: FormProps) {
  const [output, setOutput] = useState("");
  const [target, setTarget] = useState("");
  const [input, setInput] = useState("");
  const targetInstance = plant.instances.find((entry) => entry.id === target);
  const ready = output !== "" && target !== "" && input !== "";
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!ready) return;
    editor.edit((doc) => connect(doc, `${instance.id}.${output}`, `${target}.${input}`));
    setInput("");
  };
  return (
    <form className={panel.group} onSubmit={submit} aria-labelledby="connect-title" noValidate>
      <h3 id="connect-title">Link an output</h3>
      <Pick label="Output" value={output} options={portsOf(draft.specs, instance, "out")} onChange={setOutput} />
      <Pick
        label="To"
        value={target}
        options={others(plant, instance.id)}
        onChange={(value) => {
          setTarget(value);
          setInput("");
        }}
      />
      <Pick label="Input" value={input} options={portsOf(draft.specs, targetInstance, "in")} onChange={setInput} />
      <button type="submit" className={panel.secondary} aria-disabled={!ready}>
        <Cable size={14} aria-hidden="true" />
        Link
      </button>
    </form>
  );
}
