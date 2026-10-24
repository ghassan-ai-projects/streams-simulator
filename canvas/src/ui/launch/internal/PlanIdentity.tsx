import type { Draft } from "../../../model/launch";
import { panel } from "../../primitives";
import type { Change } from "./RowParts";
import styles from "./PlanIdentity.module.css";

interface TextFieldProps {
  readonly id: string;
  readonly label: string;
  readonly value: string;
  readonly max: number;
  readonly example: string;
  readonly hint: string;
  readonly onChange: (text: string) => void;
}

function TextField({ id, label, value, max, example, hint, onChange }: TextFieldProps) {
  return (
    <div className={styles.field}>
      <label htmlFor={id}>{label}</label>
      <input
        id={id}
        value={value}
        maxLength={max}
        placeholder={example}
        autoComplete="off"
        aria-describedby={`${id}-hint ${id}-count`}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      />
      <span className={styles.foot}>
        <small id={`${id}-hint`} className={styles.hint}>
          {hint}
        </small>
        <small id={`${id}-count`} className={styles.count}>{`${String(value.length)} of ${String(max)}`}</small>
      </span>
    </div>
  );
}

export function PlanIdentity({ draft, change }: { readonly draft: Draft; readonly change: (change: Change) => void }) {
  return (
    <fieldset className={panel.group}>
      <legend>Plan</legend>
      <TextField
        id="launch-name"
        label="Name"
        value={draft.name}
        max={80}
        example="Fan failure and recovery"
        hint="Optional. Names the run in lists, the top bar and the review."
        onChange={(name) => {
          change((d) => ({ ...d, name }));
        }}
      />
      <TextField
        id="launch-purpose"
        label="Purpose"
        value={draft.purpose}
        max={500}
        example="How does cooling respond when the tower fan fails?"
        hint="Optional. Without one, this is an investigation without a stated question."
        onChange={(purpose) => {
          change((d) => ({ ...d, purpose }));
        }}
      />
    </fieldset>
  );
}
