import { panel } from "../../primitives";

interface ChoiceProps {
  readonly label: string;
  readonly value: string;
  readonly options: readonly (readonly [string, string])[];
  readonly placeholder?: string;
  readonly named?: boolean;
  readonly onChange: (value: string) => void;
}

export function Choice({ label, value, options, placeholder, named = false, onChange }: ChoiceProps) {
  return (
    <label className={panel.field}>
      <span>{label}</span>
      <select
        aria-label={named ? label : undefined}
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
      >
        {placeholder !== undefined && <option value="">{placeholder}</option>}
        {options.map(([key, text]) => (
          <option key={key} value={key}>
            {text}
          </option>
        ))}
      </select>
    </label>
  );
}
