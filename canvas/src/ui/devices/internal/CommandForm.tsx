import { changeCommand, idRule, issuesOn, looksLikeId, riskChoices } from "../../../model/devices";
import type { Command } from "../../../model/devices";
import { cx, panel, useFocusOnMount } from "../../primitives";
import { ArgumentRows } from "./ArgumentRows";
import styles from "./Devices.module.css";
import { Field, ItemIssues, RemoveButton, Select, TextBox } from "./Field";
import type { CommandEdit, FormProps } from "./formProps";

interface CommandFields {
  readonly command: Command;
  readonly on: (field: string) => readonly string[];
  readonly change: (edit: CommandEdit) => void;
}

const safeStopWords =
  "A safe-stop command skips the usual timing and policy checks and is never queued. Mark one only if it puts the device in its safe state and is harmless to repeat.";

function Identity({ command, on, change }: CommandFields) {
  const ref = useFocusOnMount<HTMLInputElement>({ when: command.name === "" });
  const likely = command.name === "" || looksLikeId(command.name);
  const caution = likely ? "" : "The binary is likely to refuse this name.";
  return (
    <>
      <Field label="Name" hint={idRule} caution={caution} issues={on("name")}>
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
      <Field label="Description" hint="Up to 200 characters." issues={on("description")}>
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

function Safety({ command, on, change }: CommandFields) {
  return (
    <>
      <Field label="Risk class" hint="How much harm a wrong command could do." issues={on("risk_class")}>
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
      <Field label="Safe stop" hint={safeStopWords} issues={on("safe_stop")}>
        {(aria) => (
          <label className={styles.check}>
            <input
              {...aria}
              type="checkbox"
              checked={command.safeStop}
              onChange={(event) => {
                change({ safeStop: event.target.checked });
              }}
            />
            This command stops the device safely
          </label>
        )}
      </Field>
    </>
  );
}

export function CommandForm({ command, devices, issues }: FormProps & { readonly command: Command }) {
  const on = (field: string) => issuesOn(issues, "command", command.key, field);
  const change = (edit: CommandEdit) => {
    devices.edit((kind) => changeCommand(kind, command.key, edit));
  };
  return (
    <section className={cx(panel.body, styles.form)} aria-labelledby="command-title">
      <h2 id="command-title" className={styles.formTitle}>
        Command
      </h2>
      <ItemIssues messages={on("")} />
      <Identity command={command} on={on} change={change} />
      <Safety command={command} on={on} change={change} />
      <ArgumentRows command={command} devices={devices} issues={issues} />
      <RemoveButton
        words="Remove this command"
        onRemove={() => {
          devices.remove({ part: "command", key: command.key });
        }}
      />
    </section>
  );
}
