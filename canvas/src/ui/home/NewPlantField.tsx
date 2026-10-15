import { TriangleAlert } from "lucide-react";
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
  if (problem !== "") return `${id}-problem`;
  return hint === undefined ? undefined : `${id}-hint`;
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
  if (problem !== "") {
    return (
      <small id={`${id}-problem`} className={styles.idProblem}>
        <TriangleAlert size={12} aria-hidden="true" />
        {problem}
      </small>
    );
  }
  return hint === undefined ? null : <small id={`${id}-hint`}>{hint}</small>;
}
