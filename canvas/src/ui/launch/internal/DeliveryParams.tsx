import type { DraftPerturbation, PerturbationKind, PerturbationParamSpec } from "../../../model/launch";
import { spaced } from "../../../model/shared/words";
import { invalid } from "./RowParts";
import rows from "./PlanRows.module.css";

interface ParamProps {
  readonly name: string;
  readonly spec: PerturbationParamSpec | undefined;
  readonly label: string;
  readonly value: string;
  readonly problem: string;
  readonly id: string;
  readonly onChange: (text: string) => void;
}

function rangeText(spec: PerturbationParamSpec | undefined): string {
  if (spec?.type === "string_list") return "a, b, c";
  if (spec?.minimum === undefined || spec.maximum === undefined) return "";
  return `${String(spec.minimum)} to ${String(spec.maximum)}`;
}

function Param({ name, spec, label, value, problem, id, onChange }: ParamProps) {
  const shared = { "aria-label": `${label} ${spaced(name)}`, value, ...invalid(problem, id) };
  return (
    <label>
      {spaced(name)}
      {spec?.type === "enum" ? (
        <select
          {...shared}
          onChange={(event) => {
            onChange(event.target.value);
          }}
        >
          <option value="">default</option>
          {(spec.values ?? []).map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </select>
      ) : (
        <input
          {...shared}
          inputMode={spec?.type === "string_list" ? "text" : "decimal"}
          placeholder={rangeText(spec) || "default"}
          onChange={(event) => {
            onChange(event.target.value);
          }}
        />
      )}
    </label>
  );
}

interface ParamsProps {
  readonly delivery: DraftPerturbation;
  readonly kind: PerturbationKind | undefined;
  readonly label: string;
  readonly id: string;
  readonly problem: string;
  readonly update: (next: Partial<DraftPerturbation>) => void;
}

function namesOf({ delivery, kind }: Pick<ParamsProps, "delivery" | "kind">): readonly string[] {
  const declared = kind?.params.map((spec) => spec.name) ?? [];
  return [...declared, ...Object.keys(delivery.values).filter((name) => !declared.includes(name))];
}

export function DeliveryParams(props: ParamsProps) {
  const { delivery, kind, label, id, problem, update } = props;
  const names = namesOf(props);
  if (names.length === 0) return null;
  return (
    <div className={rows.params} role="group" aria-label={`${label} settings`}>
      {names.map((name) => (
        <Param
          key={name}
          name={name}
          spec={kind?.params.find((spec) => spec.name === name)}
          label={label}
          value={delivery.values[name] ?? ""}
          problem={problem}
          id={id}
          onChange={(text) => {
            update({ values: { ...delivery.values, [name]: text } });
          }}
        />
      ))}
    </div>
  );
}
