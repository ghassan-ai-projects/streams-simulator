import { changeCommand, issuesOn, riskChoices } from "../../../model/devices";
import type { Command } from "../../../model/devices";
import { cx, panel, useFocusOnMount } from "../../primitives";
import { ArgumentRows } from "./ArgumentRows";
import styles from "./Devices.module.css";
import { Field, ItemIssues, ItemNeeds, RemoveButton, Select, TextBox, Toggle } from "./Field";
import { FormTitle } from "./FormTitle";
import type { CommandEdit, FormProps } from "./formProps";

interface CommandFields {
  readonly command: Command;
  readonly on: (field: string) => readonly string[];
  readonly need: (field: string) => readonly string[];
  readonly change: (edit: CommandEdit) => void;
}

const safeStopWords =
  "A safe-stop command skips the usual timing and policy checks and is never queued. Mark one only if it puts the device in its safe state and is harmless to repeat.";

function Identity({ command, on, need, change }: CommandFields) {
  const ref = useFocusOnMount<HTMLInputElement>({ when: command.name === "" });
  return (
    <>
      <Field label="Name" field="name" issues={on("name")} needs={need("name")}>
        {(aria) => (
          <TextBox
            aria={aria}
            inputRef={ref}
            value={command.name}
            onChange={(name) => {
              change({ name });
            }}
          />
        )}
      </Field>
      <Field
        label="Description"
        field="description"
        hint="Up to 200 characters."
        issues={on("description")}
        needs={need("description")}
      >
        {(aria) => (
          <TextBox
            aria={aria}
            maxLength={200}
            value={command.description}
            onChange={(description) => {
              change({ description });
            }}
          />
        )}
      </Field>
    </>
  );
}

function RiskField({ command, on, need, change }: CommandFields) {
  return (
    <Field
      label="Risk class"
      field="risk_class"
      hint="How much harm a wrong command could do."
      issues={on("risk_class")}
      needs={need("risk_class")}
    >
      {(aria) => (
        <Select
          aria={aria}
          value={command.riskClass}
          choices={riskChoices}
          onChange={(riskClass) => {
            change({ riskClass });
          }}
        />
      )}
    </Field>
  );
}

function SafeStopField({ command, on, need, change }: CommandFields) {
  return (
    <Field label="Safe stop" field="safe_stop" hint={safeStopWords} issues={on("safe_stop")} needs={need("safe_stop")}>
      {(aria) => (
        <Toggle
          aria={aria}
          checked={command.safeStop}
          words="This command stops the device safely"
          onChange={(safeStop) => {
            change({ safeStop });
          }}
        />
      )}
    </Field>
  );
}

function Safety(fields: CommandFields) {
  return (
    <>
      <RiskField {...fields} />
      <SafeStopField {...fields} />
    </>
  );
}

export function CommandForm({ command, devices, issues, needs }: FormProps & { readonly command: Command }) {
  const on = (field: string) => issuesOn(issues, "command", command.key, field);
  const need = (field: string) => issuesOn(needs, "command", command.key, field);
  const change = (edit: CommandEdit) => {
    devices.edit((kind) => changeCommand(kind, command.key, edit));
  };
  const fields = { command, on, need, change };
  return (
    <section className={cx(panel.body, styles.form)} aria-labelledby="command-title">
      <FormTitle id="command-title" words="Command" />
      <ItemIssues messages={on("")} />
      <ItemNeeds messages={need("")} />
      <Identity {...fields} />
      <Safety {...fields} />
      <ArgumentRows command={command} devices={devices} issues={issues} needs={needs} />
      <RemoveButton
        words="Remove this command"
        onRemove={() => {
          devices.remove({ part: "command", key: command.key });
        }}
      />
    </section>
  );
}
