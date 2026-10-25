import { TriangleAlert } from "lucide-react";
import { useId, useState } from "react";
import type { ParameterSpec } from "../../../model/plant";
import { codeTitle, displayUnit } from "../../../model/shared/units";
import { cx, shared } from "../../primitives";
import styles from "./Parameters.module.css";

export interface FieldProps {
  readonly spec: ParameterSpec;
  readonly value: unknown;
  readonly refusals: readonly string[];
  readonly onChange: (value: number | string | null) => void;
}

function boundsWords({ min, max, default: given, unit }: ParameterSpec): string {
  const symbol = displayUnit(unit);
  const space = symbol === "" ? "" : ` ${symbol}`;
  const range = min !== null && max !== null ? `${String(min)} to ${String(max)}${space}` : "";
  const standard = given === null ? "" : `default ${String(given)}${space}`;
  return [range, standard].filter((part) => part !== "").join(" · ");
}

function Refusals({ id, messages }: { readonly id: string; readonly messages: readonly string[] }) {
  return messages.map((message) => (
    <p key={message} id={id} className={styles.refusal} role="alert">
      <TriangleAlert size={13} aria-hidden="true" />
      {message}
    </p>
  ));
}

function NumberInput({
  spec,
  value,
  onChange,
  id,
  described,
}: FieldProps & { readonly id: string; readonly described: string }) {
  const shown = typeof value === "number" ? String(value) : "";
  const [text, setText] = useState(shown);
  const [seen, setSeen] = useState(shown);
  if (seen !== shown) {
    setSeen(shown);
    setText(shown);
  }
  const commit = () => {
    if (text.trim() === "") onChange(null);
    else if (Number.isFinite(Number(text))) onChange(Number(text));
    else setText(shown);
  };
  return (
    <input
      id={id}
      type="number"
      step={spec.type === "integer" ? 1 : "any"}
      {...(spec.min === null ? {} : { min: spec.min })}
      {...(spec.max === null ? {} : { max: spec.max })}
      placeholder={spec.default === null ? undefined : String(spec.default)}
      aria-describedby={described}
      value={text}
      onChange={(event) => {
        setText(event.target.value);
      }}
      onBlur={commit}
      onKeyDown={(event) => {
        if (event.key === "Enter") commit();
      }}
    />
  );
}

function EnumInput({
  spec,
  value,
  onChange,
  id,
  described,
}: FieldProps & { readonly id: string; readonly described: string }) {
  const shown = typeof value === "string" ? value : String(spec.default ?? "");
  return (
    <select
      id={id}
      aria-describedby={described}
      value={shown}
      onChange={(event) => {
        onChange(event.target.value);
      }}
    >
      {spec.values.map((choice) => (
        <option key={choice} value={choice}>
          {choice}
        </option>
      ))}
    </select>
  );
}

function Control(props: FieldProps & { readonly id: string; readonly described: string }) {
  const unit = displayUnit(props.spec.unit);
  return (
    <span className={styles.control}>
      {props.spec.type === "enum" ? <EnumInput {...props} /> : <NumberInput {...props} />}
      {unit !== "" && (
        <span className={styles.unit} title={codeTitle(props.spec.unit) || undefined}>
          {unit}
        </span>
      )}
    </span>
  );
}

export function ParameterField(props: FieldProps) {
  const { spec, value, refusals, onChange } = props;
  const id = useId();
  const described = `${id}-bounds`;
  return (
    <div className={styles.parameter}>
      <label htmlFor={id} className={styles.name}>
        {spec.title}
      </label>
      <Control {...props} id={id} described={described} />
      <span id={described} className={styles.bounds}>
        {boundsWords(spec)}
      </span>
      <Refusals id={`${id}-refusal`} messages={refusals} />
      {value !== undefined && (
        <button
          type="button"
          className={cx(shared.secondary, styles.reset)}
          aria-label={`Use the default of ${spec.title}`}
          onClick={() => {
            onChange(null);
          }}
        >
          Use the default
        </button>
      )}
    </div>
  );
}
