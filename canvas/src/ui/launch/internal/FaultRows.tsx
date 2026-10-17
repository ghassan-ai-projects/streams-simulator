import { panel } from "../../primitives";
import { Plus } from "lucide-react";
import type { Draft, DraftFault, PlanProblem } from "../../../model/launch";
import { faultsOf } from "../../../model/plant";
import type { Plant } from "../../../model/plant";
import { humanise } from "../../../model/shared/words";
import type { Change } from "./RowParts";
import { AtMinute, InstancePick, invalid, RowTail, useAdding } from "./RowParts";
import styles from "./LaunchPanel.module.css";

export type { Change } from "./RowParts";

interface FaultPickProps {
  readonly plant: Plant;
  readonly fault: DraftFault;
  readonly label: string;
  readonly id: string;
  readonly problem: string;
  readonly update: (next: Partial<DraftFault>) => void;
}

function FaultPick({ plant, fault, label, id, problem, update }: FaultPickProps) {
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
      {(faultsOf(plant, fault.entity) ?? []).map((name) => (
        <option key={name} value={name}>
          {humanise(name)}
        </option>
      ))}
    </select>
  );
}

interface FaultRowProps {
  readonly plant: Plant;
  readonly fault: DraftFault;
  readonly index: number;
  readonly problem: string;
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

function FaultRow({ plant, fault, index, problem, change, fresh, onRemoved }: FaultRowProps) {
  const label = `Fault ${String(index + 1)}`;
  const id = `launch-fault-${String(index + 1)}`;
  const { update, remove } = useFaultEdit(change, fault.key, onRemoved);
  const complete = fault.entity !== "" && fault.fault !== "";
  return (
    <li className={styles.faultRow} aria-label={label}>
      <AtMinute
        value={fault.atMinutes}
        label={label}
        id={id}
        problem={complete ? problem : ""}
        onChange={(atMinutes) => {
          update({ atMinutes });
        }}
      />
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
        has a <FaultPick plant={plant} fault={fault} label={label} id={id} problem={problem} update={update} />
      </span>
      <RowTail label={label} id={id} problem={problem} onRemove={remove} />
    </li>
  );
}

interface FaultRowsProps {
  readonly plant: Plant;
  readonly draft: Draft;
  readonly problems: readonly PlanProblem[];
  readonly selected: string | null;
  readonly change: (change: Change) => void;
}

export function FaultRows({ plant, draft, problems, selected, change }: FaultRowsProps) {
  const { added, addButton, add } = useAdding(change, (d, key) => ({
    ...d,
    faults: [...d.faults, { key, entity: selected ?? "", fault: "", atMinutes: "0" }],
  }));
  const problemOf = (index: number) => problems.find((problem) => problem.field === index)?.text ?? "";
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
            problem={problemOf(index)}
            change={change}
            fresh={fault.key === added}
            onRemoved={() => {
              addButton.current?.focus();
            }}
          />
        ))}
      </ol>
      <button
        type="button"
        ref={addButton}
        className={panel.secondary}
        onClick={() => {
          add(draft.nextKey);
        }}
      >
        <Plus size={14} aria-hidden="true" />
        {selected === null ? "Add fault" : `Add fault on ${selected}`}
      </button>
    </fieldset>
  );
}
