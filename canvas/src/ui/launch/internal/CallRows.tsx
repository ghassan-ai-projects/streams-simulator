import { cx, panel, shared } from "../../primitives";
import { Plus } from "lucide-react";
import type { CallProblem, Draft, DraftCall, PlanProblem } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { effectorNamesOf, hasEffectors } from "../../../model/plant";
import { humanise } from "../../../model/shared/words";
import { CallArgs } from "./CallArgs";
import type { Change } from "./RowParts";
import { AtMinute, InstancePick, invalid, RowTail, useAdding } from "./RowParts";
import styles from "./LaunchPanel.module.css";

interface CallRowProps {
  readonly plant: Plant;
  readonly call: DraftCall;
  readonly index: number;
  readonly problem: CallProblem | null;
  readonly change: (change: Change) => void;
  readonly fresh: boolean;
  readonly onRemoved: () => void;
}

interface EffectorPickProps {
  readonly plant: Plant;
  readonly call: DraftCall;
  readonly label: string;
  readonly id: string;
  readonly problem: string;
  readonly update: (next: Partial<DraftCall>) => void;
}

function EffectorPick({ plant, call, label, id, problem, update }: EffectorPickProps) {
  return (
    <span className={styles.sentence}>
      calls{" "}
      <select
        aria-label={`${label} effector`}
        value={call.effector}
        disabled={call.entity === ""}
        {...invalid(problem, id)}
        onChange={(event) => {
          update({ effector: event.target.value, values: {} });
        }}
      >
        <option value="">Choose an effector</option>
        {effectorNamesOf(plant, call.entity).map((name) => (
          <option key={name} value={name}>
            {humanise(name)}
          </option>
        ))}
      </select>
    </span>
  );
}

function useCallEdit(change: (change: Change) => void, key: number, onRemoved: () => void) {
  const update = (next: Partial<DraftCall>) => {
    change((draft) => ({ ...draft, calls: draft.calls.map((c) => (c.key === key ? { ...c, ...next } : c)) }));
  };
  const remove = () => {
    change((draft) => ({ ...draft, calls: draft.calls.filter((c) => c.key !== key) }));
    onRemoved();
  };
  return { update, remove };
}

function CallRow(props: CallRowProps) {
  const { plant, call, index, problem, fresh } = props;
  const label = `Effector call ${String(index + 1)}`;
  const id = `launch-call-${String(index + 1)}`;
  const text = problem?.text ?? "";
  const part = (name: string) => (problem?.part === name ? text : "");
  const { update, remove } = useCallEdit(props.change, call.key, props.onRemoved);
  return (
    <li className={styles.faultRow} aria-label={label}>
      <AtMinute
        value={call.atMinutes}
        label={label}
        id={id}
        problem={part("minute")}
        onChange={(atMinutes) => {
          update({ atMinutes });
        }}
      />
      <InstancePick
        plant={plant}
        entity={call.entity}
        label={label}
        id={id}
        problem={part("instance")}
        fresh={fresh}
        onPick={(entity) => {
          update({ entity, effector: "", values: {} });
        }}
      />
      <EffectorPick plant={plant} call={call} label={label} id={id} problem={part("effector")} update={update} />
      <CallArgs {...props} label={label} id={id} update={update} />
      <RowTail label={label} id={id} problem={text} onRemove={remove} />
    </li>
  );
}

const emptyWords = (hasEffectors: boolean) =>
  hasEffectors
    ? "None: nothing is commanded. A call is made when the run reaches its minute; the simulator may refuse it, and the result says why."
    : "This plant declares no effectors to call.";

const problemOf = (problems: readonly PlanProblem[], index: number): CallProblem | null =>
  problems.find((problem): problem is CallProblem => problem.field === "call" && problem.index === index) ?? null;

const appended = (draft: Draft, key: number, selected: string | null): Draft => ({
  ...draft,
  calls: [...draft.calls, { key, entity: selected ?? "", effector: "", atMinutes: "0", values: {} }],
});

interface CallRowsProps {
  readonly plant: Plant;
  readonly draft: Draft;
  readonly problems: readonly PlanProblem[];
  readonly selected: string | null;
  readonly change: (change: Change) => void;
}

export function CallRows({ plant, draft, problems, selected, change }: CallRowsProps) {
  const { added, addButton, add } = useAdding(change, (d, key) => appended(d, key, selected));
  return (
    <fieldset className={panel.group}>
      <legend>Effector calls</legend>
      {draft.calls.length === 0 && <p className={panel.hint}>{emptyWords(hasEffectors(plant))}</p>}
      <ol className={styles.faults}>
        {draft.calls.map((call, index) => (
          <CallRow
            key={call.key}
            plant={plant}
            call={call}
            index={index}
            problem={problemOf(problems, index)}
            change={change}
            fresh={call.key === added}
            onRemoved={() => {
              addButton.current?.focus();
            }}
          />
        ))}
      </ol>
      {hasEffectors(plant) && (
        <button
          type="button"
          ref={addButton}
          className={cx(shared.secondary, panel.secondary)}
          onClick={() => {
            add(draft.nextKey);
          }}
        >
          <Plus size={14} aria-hidden="true" />
          {selected === null ? "Add effector call" : `Add effector call on ${selected}`}
        </button>
      )}
    </fieldset>
  );
}
