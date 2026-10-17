import { Trash2, TriangleAlert } from "lucide-react";
import type { ReactNode, RefObject } from "react";
import { useId, useState } from "react";
import { choicesFor } from "../../../model/devices";
import type { Choices } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import styles from "./Devices.module.css";

interface Aria {
  readonly id: string;
  readonly "aria-invalid": true | undefined;
  readonly "aria-describedby": string | undefined;
}

interface FieldProps {
  readonly label: string;
  readonly hint?: string | undefined;
  readonly caution?: string;
  readonly issues: readonly string[];
  readonly children: (aria: Aria) => ReactNode;
}

interface NotesProps {
  readonly id: string;
  readonly hint: string | undefined;
  readonly caution: string;
  readonly issues: readonly string[];
}

function Notes({ id, hint, caution, issues }: NotesProps) {
  return (
    <div id={id} className={styles.notes}>
      {issues.map((message, index) => (
        <p key={`${String(index)}:${message}`} className={styles.issue}>
          <TriangleAlert size={12} aria-hidden="true" />
          {message}
        </p>
      ))}
      {caution !== "" && <p className={styles.caution}>{caution}</p>}
      {hint !== undefined && <p className={panel.hint}>{hint}</p>}
    </div>
  );
}

export function Field({ label, hint, caution = "", issues, children }: FieldProps) {
  const id = useId();
  const noted = issues.length > 0 || caution !== "" || hint !== undefined;
  const aria: Aria = {
    id,
    "aria-invalid": issues.length > 0 ? true : undefined,
    "aria-describedby": noted ? `${id}-notes` : undefined,
  };
  return (
    <div className={styles.field}>
      <label htmlFor={id}>{label}</label>
      <div className={styles.control}>
        {children(aria)}
        {noted && <Notes id={`${id}-notes`} hint={hint} caution={caution} issues={issues} />}
      </div>
    </div>
  );
}

interface TextBoxProps {
  readonly aria: Aria;
  readonly value: string;
  readonly onChange: (value: string) => void;
  readonly list?: string | undefined;
  readonly maxLength?: number | undefined;
  readonly inputRef?: RefObject<HTMLInputElement | null> | undefined;
}

export function TextBox({ aria, value, onChange, list, maxLength, inputRef }: TextBoxProps) {
  return (
    <input
      {...aria}
      ref={inputRef}
      type="text"
      list={list}
      maxLength={maxLength}
      value={value}
      onChange={(event) => {
        onChange(event.target.value);
      }}
    />
  );
}

interface SelectProps {
  readonly value: string;
  readonly choices: Choices;
  readonly aria?: Aria;
  readonly label?: string;
  readonly onChange: (value: string) => void;
}

export function Select({ value, choices, aria, label, onChange }: SelectProps) {
  return (
    <select
      {...aria}
      aria-label={label}
      value={value}
      onChange={(event) => {
        onChange(event.target.value);
      }}
    >
      {choicesFor(choices, value).map(([key, text]) => (
        <option key={key} value={key}>
          {text}
        </option>
      ))}
    </select>
  );
}

export function ItemIssues({ messages }: { readonly messages: readonly string[] }) {
  if (messages.length === 0) return null;
  return (
    <ul className={styles.issueList}>
      {messages.map((message, index) => (
        <li key={`${String(index)}:${message}`}>
          <TriangleAlert size={14} aria-hidden="true" />
          {message}
        </li>
      ))}
    </ul>
  );
}

export function RemoveButton({ words, onRemove }: { readonly words: string; readonly onRemove: () => void }) {
  return (
    <button type="button" className={cx(shared.secondary, panel.secondary, styles.remove)} onClick={onRemove}>
      <Trash2 size={14} aria-hidden="true" />
      {words}
    </button>
  );
}

interface NumberProps {
  readonly value: number | null;
  readonly aria?: Aria;
  readonly name: string;
  readonly placeholder?: string;
  readonly onChange: (value: number | null) => void;
}

function parsed(typed: string): number | null | "broken" {
  if (typed.trim() === "") return null;
  const value = Number(typed);
  return Number.isFinite(value) ? value : "broken";
}

export function NumberInput({ value, aria, name, placeholder, onChange }: NumberProps) {
  const [typed, setTyped] = useState<string | null>(null);
  const shown = typed ?? (value === null ? "" : String(value));
  const broken = typed !== null && parsed(typed) === "broken";
  return (
    <>
      <input
        {...aria}
        type="text"
        inputMode="decimal"
        placeholder={placeholder}
        aria-label={name}
        aria-invalid={broken ? true : aria?.["aria-invalid"]}
        value={shown}
        onChange={(event) => {
          const next = parsed(event.target.value);
          setTyped(event.target.value);
          if (next !== "broken") onChange(next);
        }}
        onBlur={() => {
          if (!broken) setTyped(null);
        }}
      />
      {broken && <p className={styles.issue}>Enter a number.</p>}
    </>
  );
}
