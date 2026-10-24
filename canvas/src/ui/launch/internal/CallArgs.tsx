import type { CallProblem, NowDraft } from "../../../model/launch";
import type { EffectorArg, Plant } from "../../../model/plant";
import { effectorOf, rangeHint } from "../../../model/plant";
import { spaced } from "../../../model/shared/words";
import { invalid } from "./RowParts";
import styles from "./LaunchPanel.module.css";

function choices(arg: EffectorArg): readonly string[] {
  return arg.type === "boolean" ? ["true", "false"] : arg.options;
}

interface ArgFieldProps {
  readonly arg: EffectorArg;
  readonly label: string;
  readonly value: string;
  readonly problem: string;
  readonly id: string;
  readonly onChange: (text: string) => void;
}

function ArgField({ arg, label, value, problem, id, onChange }: ArgFieldProps) {
  const options = choices(arg);
  const shared = { "aria-label": `${label} ${spaced(arg.name)}`, value, ...invalid(problem, id) };
  return (
    <label>
      {spaced(arg.name)}
      {options.length > 0 ? (
        <select
          {...shared}
          onChange={(event) => {
            onChange(event.target.value);
          }}
        >
          <option value="">Choose</option>
          {options.map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </select>
      ) : (
        <input
          {...shared}
          className={styles.minute}
          inputMode="decimal"
          placeholder={rangeHint(arg)}
          onChange={(event) => {
            onChange(event.target.value);
          }}
        />
      )}
    </label>
  );
}

export function CallArgs({
  plant,
  call,
  label,
  id,
  problem,
  update,
}: {
  readonly plant: Plant;
  readonly call: NowDraft;
  readonly label: string;
  readonly id: string;
  readonly problem: CallProblem | null;
  readonly update: (next: Partial<NowDraft>) => void;
}) {
  const spec = effectorOf(plant, call.entity, call.effector);
  const shown = (spec?.args ?? []).filter((arg) => arg.type !== "other");
  if (shown.length === 0) return null;
  return (
    <div className={styles.args}>
      {shown.map((arg) => (
        <ArgField
          key={arg.name}
          arg={arg}
          label={label}
          value={call.values[arg.name] ?? ""}
          problem={problem?.part === "args" ? problem.text : ""}
          id={id}
          onChange={(text) => {
            update({ values: { ...call.values, [arg.name]: text } });
          }}
        />
      ))}
    </div>
  );
}
