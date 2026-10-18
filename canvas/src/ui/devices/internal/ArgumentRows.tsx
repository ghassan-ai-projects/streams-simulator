import { Plus, X } from "lucide-react";
import type { ReactNode } from "react";
import { addArgument, argumentTypeChoices, changeArgument, issuesOn, removeArgument } from "../../../model/devices";
import type { ArgRow, Command } from "../../../model/devices";
import { cx, panel, shared, useFocusOnMount } from "../../primitives";
import styles from "./Devices.module.css";
import { ItemIssues, ItemNeeds, NumberInput, Select, Toggle } from "./Field";
import type { FormProps } from "./formProps";

interface RowProps {
  readonly row: ArgRow;
  readonly label: string;
  readonly change: (edit: Partial<ArgRow>) => void;
}

function Labelled({ words, children }: { readonly words: string; readonly children: ReactNode }) {
  return (
    <span className={styles.argField}>
      <span className={styles.argLabel} aria-hidden="true">
        {words}
      </span>
      {children}
    </span>
  );
}

function Limits({ row, label, change }: RowProps) {
  return (
    <>
      <Labelled words="Min">
        <NumberInput
          value={row.min}
          label={`${label} minimum`}
          placeholder="min"
          onChange={(min) => {
            change({ min });
          }}
        />
      </Labelled>
      <Labelled words="Max">
        <NumberInput
          value={row.max}
          label={`${label} maximum`}
          placeholder="max"
          onChange={(max) => {
            change({ max });
          }}
        />
      </Labelled>
    </>
  );
}

function ArgumentRow({ row, label, change, remove }: RowProps & { readonly remove: () => void }) {
  const name = useFocusOnMount<HTMLInputElement>({ when: row.name === "" });
  return (
    <li className={styles.argRow}>
      <Labelled words="Name">
        <input
          ref={name}
          type="text"
          aria-label={`${label} name`}
          placeholder="name"
          value={row.name}
          onChange={(event) => {
            change({ name: event.target.value });
          }}
        />
      </Labelled>
      <Labelled words="Type">
        <Select
          label={`${label} type`}
          value={row.type}
          choices={argumentTypeChoices}
          onChange={(type) => {
            change({ type });
          }}
        />
      </Labelled>
      <Limits row={row} label={label} change={change} />
      <Toggle
        checked={row.required}
        words="Required"
        onChange={(required) => {
          change({ required });
        }}
      />
      <button type="button" className={panel.icon} aria-label={`Remove ${label.toLowerCase()}`} onClick={remove}>
        <X size={16} aria-hidden="true" />
      </button>
    </li>
  );
}

function AddArgument({ onAdd }: { readonly onAdd: () => void }) {
  return (
    <button type="button" className={cx(shared.secondary, panel.secondary)} onClick={onAdd}>
      <Plus size={14} aria-hidden="true" />
      Add argument
    </button>
  );
}

export function ArgumentRows({ command, devices, issues, needs }: FormProps & { readonly command: Command }) {
  const { args } = command;
  if (args.kind === "custom") {
    return (
      <p className={panel.hint}>
        This command&apos;s arguments use a definition this editor cannot show. They are kept as they are.
      </p>
    );
  }
  return (
    <section aria-labelledby="arguments-title" className={panel.group}>
      <h3 id="arguments-title">Arguments</h3>
      <ItemIssues messages={issuesOn(issues, "command", command.key, "args_schema")} />
      <ItemNeeds messages={issuesOn(needs, "command", command.key, "args_schema")} />
      <ul className={styles.argRows}>
        {args.rows.map((row, index) => (
          <ArgumentRow
            key={String(index)}
            row={row}
            label={`Argument ${String(index + 1)}`}
            change={(edit) => {
              devices.edit((kind) => changeArgument(kind, command.key, index, edit));
            }}
            remove={() => {
              devices.edit((kind) => removeArgument(kind, command.key, index));
            }}
          />
        ))}
      </ul>
      <AddArgument
        onAdd={() => {
          devices.edit((kind) => addArgument(kind, command.key));
        }}
      />
    </section>
  );
}
