import { Info, Trash2, TriangleAlert } from "lucide-react";
import type { ReactNode, RefObject } from "react";
import { useId, useState } from "react";
import { choicesFor } from "../../../model/devices";
import type { Choices } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import styles from "./Devices.module.css";
import notes from "./Notes.module.css";

export interface Aria {
  readonly id: string;
  readonly "aria-invalid": true | undefined;
  readonly "aria-describedby": string | undefined;
  readonly "data-field": string | undefined;
}

interface FieldProps {
  readonly label: string;
  readonly field?: string;
  readonly hint?: string | undefined;
  readonly caution?: string;
  readonly issues: readonly string[];
  readonly needs?: readonly string[];
  readonly group?: boolean;
  readonly children: (aria: Aria) => ReactNode;
}

interface NotesProps {
  readonly id: string;
  readonly hint: string | undefined;
  readonly caution: string;
  readonly issues: readonly string[];
  readonly needs: readonly string[];
}

function Notes({ id, hint, caution, issues, needs }: NotesProps) {
  return (
    <div id={id} className={styles.notes} aria-live="polite">
      {issues.map((message, index) => (
        <p key={`${String(index)}:${message}`} className={styles.issue}>
          <TriangleAlert size={12} aria-hidden="true" />
          {message}
        </p>
      ))}
      {needs.map((message, index) => (
        <p key={`${String(index)}:${message}`} className={notes.need}>
          <Info size={12} aria-hidden="true" />
          {message}
        </p>
      ))}
      {caution !== "" && <p className={styles.caution}>{caution}</p>}
      {hint !== undefined && <p className={panel.hint}>{hint}</p>}
    </div>
  );
}

function isNoted(hint: string | undefined, caution: string, issues: readonly string[], needs: readonly string[]) {
  return issues.length + needs.length > 0 || caution !== "" || hint !== undefined;
}

function ariaOf(id: string, field: string | undefined, issues: readonly string[], noted: boolean): Aria {
  return {
    id,
    "aria-invalid": issues.length > 0 ? true : undefined,
    "aria-describedby": noted ? `${id}-notes` : undefined,
    "data-field": field,
  };
}

function FieldLabel({ id, label, group }: { readonly id: string; readonly label: string; readonly group: boolean }) {
  return group ? <span id={`${id}-label`}>{label}</span> : <label htmlFor={id}>{label}</label>;
}

export function Field({ label, field, hint, caution = "", issues, needs = [], group = false, children }: FieldProps) {
  const id = useId();
  const aria = ariaOf(id, field, issues, isNoted(hint, caution, issues, needs));
  return (
    <div
      className={styles.field}
      role={group ? "group" : undefined}
      aria-labelledby={group ? `${id}-label` : undefined}
    >
      <FieldLabel id={id} label={label} group={group} />
      <div className={styles.control}>
        {children(aria)}
        <Notes id={`${id}-notes`} hint={hint} caution={caution} issues={issues} needs={needs} />
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
  readonly onCommit?: (() => void) | undefined;
}

export function TextBox({ aria, value, onChange, list, maxLength, inputRef, onCommit }: TextBoxProps) {
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
      onBlur={onCommit}
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
      aria-label={aria === undefined ? label : undefined}
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

interface ToggleProps {
  readonly aria?: Aria;
  readonly checked: boolean;
  readonly words: string;
  readonly onChange: (checked: boolean) => void;
}

export function Toggle({ aria, checked, words, onChange }: ToggleProps) {
  return (
    <label className={styles.check}>
      <input
        {...aria}
        type="checkbox"
        checked={checked}
        onChange={(event) => {
          onChange(event.target.checked);
        }}
      />
      {words}
    </label>
  );
}

export function ItemIssues({ messages }: { readonly messages: readonly string[] }) {
  if (messages.length === 0) return null;
  return (
    <div role="alert">
      <ul className={styles.issueList}>
        {messages.map((message, index) => (
          <li key={`${String(index)}:${message}`}>
            <TriangleAlert size={14} aria-hidden="true" />
            {message}
          </li>
        ))}
      </ul>
    </div>
  );
}

export function ItemNeeds({ messages }: { readonly messages: readonly string[] }) {
  if (messages.length === 0) return null;
  return (
    <ul className={notes.needList}>
      {messages.map((message, index) => (
        <li key={`${String(index)}:${message}`}>
          <Info size={14} aria-hidden="true" />
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
  readonly label?: string;
  readonly placeholder?: string;
  readonly onChange: (value: number | null) => void;
  readonly onCommit?: (() => void) | undefined;
}

function parsed(typed: string): number | null | "broken" {
  if (typed.trim() === "") return null;
  const value = Number(typed);
  return Number.isFinite(value) ? value : "broken";
}

export function NumberInput({ value, aria, label, placeholder, onChange, onCommit }: NumberProps) {
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
        aria-label={aria === undefined ? label : undefined}
        aria-invalid={broken ? true : aria?.["aria-invalid"]}
        value={shown}
        onChange={(event) => {
          const next = parsed(event.target.value);
          setTyped(event.target.value);
          if (next !== "broken") onChange(next);
        }}
        onBlur={() => {
          if (!broken) setTyped(null);
          onCommit?.();
        }}
      />
      {broken && (
        <p className={styles.issue} role="alert">
          Enter a number.
        </p>
      )}
    </>
  );
}
