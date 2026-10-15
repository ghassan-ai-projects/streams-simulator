import type { FocusEvent } from "react";
import { useId } from "react";
import { useFocusOnMount } from "../useFocusOnMount";
import styles from "./HomeScreen.module.css";

function selectText(event: FocusEvent<HTMLInputElement>) {
  event.currentTarget.select();
}

interface FieldProps {
  readonly label: string;
  readonly hint?: string | undefined;
  readonly value: string;
  readonly onChange: (value: string) => void;
  readonly first?: boolean;
  readonly placeholder?: string;
  readonly onLeave?: () => void;
  readonly selectOnFocus?: boolean;
  readonly problem?: string;
}

export function Field({ label, hint, problem = "", ...input }: FieldProps) {
  const id = useId();
  return (
    <div className={styles.newField}>
      <label htmlFor={id}>{label}</label>
      <FieldInput id={id} hint={hint} problem={problem} {...input} />
      <FieldNotes id={id} hint={hint} problem={problem} />
    </div>
  );
}

function FieldInput({
  id,
  value,
  onChange,
  first = false,
  placeholder,
  onLeave,
  selectOnFocus = false,
  hint,
  problem,
}: Omit<FieldProps, "label" | "problem"> & { readonly id: string; readonly problem: string }) {
  const ref = useFocusOnMount<HTMLInputElement>(first);
  return (
    <input
      id={id}
      ref={ref}
      placeholder={placeholder}
      value={value}
      aria-invalid={problem === "" ? undefined : true}
      aria-describedby={describedBy(id, hint, problem)}
      onChange={(event) => {
        onChange(event.target.value);
      }}
      onBlur={onLeave}
      onFocus={selectOnFocus ? selectText : undefined}
    />
  );
}

function describedBy(id: string, hint: string | undefined, problem: string): string | undefined {
  const parts = [hint === undefined ? "" : `${id}-hint`, problem === "" ? "" : `${id}-problem`];
  const joined = parts.filter((part) => part !== "").join(" ");
  return joined === "" ? undefined : joined;
}

function FieldNotes({
  id,
  hint,
  problem,
}: {
  readonly id: string;
  readonly hint: string | undefined;
  readonly problem: string;
}) {
  return (
    <>
      {hint !== undefined && <small id={`${id}-hint`}>{hint}</small>}
      {problem !== "" && (
        <small id={`${id}-problem`} className={styles.idProblem}>
          {problem}
        </small>
      )}
    </>
  );
}
