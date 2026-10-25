import { panel } from "../../primitives";
import type { Draft, DraftFault, FaultCategories, PlanProblem } from "../../../model/launch";
import { faultGroups } from "../../../model/launch";
import { faultsOf } from "../../../model/plant";
import type { Plant } from "../../../model/plant";
import { humanise } from "../../../model/shared/words";
import type { Change } from "./RowParts";
import { AddRow, InstancePick, invalid, ProblemNote, RowTail, RowTimes, useAdding } from "./RowParts";
import styles from "./LaunchPanel.module.css";

export type { Change } from "./RowParts";

interface FaultPickProps {
  readonly plant: Plant;
  readonly fault: DraftFault;
  readonly label: string;
  readonly id: string;
  readonly problem: string;
  readonly categories: FaultCategories;
  readonly update: (next: Partial<DraftFault>) => void;
}

function FaultOptions({ plant, fault, categories }: Pick<FaultPickProps, "plant" | "fault" | "categories">) {
  const kind = plant.instances.find((instance) => instance.id === fault.entity)?.kind ?? "";
  return faultGroups(faultsOf(plant, fault.entity) ?? [], categories[kind]).map((group) => {
    const options = group.faults.map((name) => (
      <option key={name} value={name}>
        {humanise(name)}
      </option>
    ));
    return group.label === "" ? (
      options
    ) : (
      <optgroup key={group.label} label={group.label}>
        {options}
      </optgroup>
    );
  });
}

function FaultPick({ plant, fault, label, id, problem, categories, update }: FaultPickProps) {
  return (
    <select
      aria-label={`${label} fault`}
      value={fault.fault}
      disabled={fault.entity === ""}
      {...invalid(fault.entity !== "" && fault.fault === "" ? problem : "", id)}
      onChange={(event) => {
        update({ fault: event.target.value });
      }}
    >
      <option value="">Choose a fault</option>
      <FaultOptions plant={plant} fault={fault} categories={categories} />
    </select>
  );
}

interface FaultRowProps {
  readonly plant: Plant;
  readonly fault: DraftFault;
  readonly index: number;
  readonly problem: string;
  readonly endProblem: string;
  readonly categories: FaultCategories;
  readonly change: (change: Change) => void;
  readonly fresh: boolean;
  readonly onRemoved: () => void;
}

function useFaultEdit(change: (change: Change) => void, key: number, onRemoved: () => void) {
  const update = (next: Partial<DraftFault>) => {
    change((draft) => ({ ...draft, faults: draft.faults.map((f) => (f.key === key ? { ...f, ...next } : f)) }));
  };
  const remove = () => {
    change((draft) => ({ ...draft, faults: draft.faults.filter((f) => f.key !== key) }));
    onRemoved();
  };
  return { update, remove };
}

type TargetProps = Omit<FaultPickProps, "update"> & {
  readonly fresh: boolean;
  readonly update: (next: Partial<DraftFault>) => void;
};

function FaultTarget({ fresh, ...pick }: TargetProps) {
  const { plant, fault, label, id, problem, update } = pick;
  return (
    <>
      <InstancePick
        plant={plant}
        entity={fault.entity}
        label={label}
        id={id}
        problem={fault.entity === "" ? problem : ""}
        fresh={fresh}
        onPick={(entity) => {
          update({ entity, fault: "" });
        }}
      />
      <span className={styles.sentence}>
        has a <FaultPick {...pick} />
      </span>
    </>
  );
}

function FaultRow({ plant, fault, index, problem, endProblem, categories, change, fresh, onRemoved }: FaultRowProps) {
  const label = `Fault ${String(index + 1)}`;
  const id = `launch-fault-${String(index + 1)}`;
  const { update, remove } = useFaultEdit(change, fault.key, onRemoved);
  const complete = fault.entity !== "" && fault.fault !== "";
  return (
    <li className={styles.faultRow} aria-label={label}>
      <RowTimes
        label={label}
        id={id}
        atMinutes={fault.atMinutes}
        untilMinutes={fault.untilMinutes}
        startProblem={complete ? problem : ""}
        endProblem={endProblem}
        endPlaceholder="never"
        update={update}
      />
      <FaultTarget {...{ plant, fault, label, id, problem, categories, fresh, update }} />
      <RowTail label={label} id={id} problem={problem} onRemove={remove} />
      <ProblemNote id={`${id}-end`} text={endProblem} />
    </li>
  );
}

const problemOf = (problems: readonly PlanProblem[], field: "fault" | "faultEnd", index: number) =>
  problems.find((problem) => problem.field === field && problem.index === index)?.text ?? "";

interface FaultRowsProps {
  readonly plant: Plant;
  readonly draft: Draft;
  readonly problems: readonly PlanProblem[];
  readonly categories: FaultCategories;
  readonly selected: string | null;
  readonly change: (change: Change) => void;
}

export function FaultRows({ plant, draft, problems, categories, selected, change }: FaultRowsProps) {
  const { added, addButton, add } = useAdding(change, (d, key) => ({
    ...d,
    faults: [...d.faults, { key, entity: selected ?? "", fault: "", atMinutes: "0", untilMinutes: "" }],
  }));
  return (
    <fieldset className={panel.group}>
      <legend>Faults to inject</legend>
      {draft.faults.length === 0 && <p className={panel.hint}>None: no fault is scheduled for this run.</p>}
      <ol className={styles.faults}>
        {draft.faults.map((fault, index) => (
          <FaultRow
            key={fault.key}
            plant={plant}
            fault={fault}
            index={index}
            problem={problemOf(problems, "fault", index)}
            endProblem={problemOf(problems, "faultEnd", index)}
            categories={categories}
            change={change}
            fresh={fault.key === added}
            onRemoved={() => {
              addButton.current?.focus();
            }}
          />
        ))}
      </ol>
      <AddRow
        label={selected === null ? "Add fault" : `Add fault on ${selected}`}
        buttonRef={addButton}
        onAdd={() => {
          add(draft.nextKey);
        }}
      />
    </fieldset>
  );
}
