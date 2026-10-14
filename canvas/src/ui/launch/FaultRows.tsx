import { Plus, Trash2, TriangleAlert } from "lucide-react";
import { useRef, useState } from "react";
import type { PlanProblem } from "../../model/launch";
import type { Plant } from "../../model/plant";
import { humanise } from "../../model/report";
import type { Draft, DraftFault } from "./draft";
import panel from "../panel.module.css";
import { useFocusOnMount } from "../useFocusOnMount";
import styles from "./LaunchPanel.module.css";

export type Change = (draft: Draft) => Draft;

interface RowProps {
  readonly fresh?: boolean;
  readonly plant: Plant;
  readonly fault: DraftFault;
  readonly label: string;
  readonly id: string;
  readonly problem: string;
  readonly update: (next: Partial<DraftFault>) => void;
}

function faultsOf(plant: Plant, entity: string): readonly string[] {
  const kind = plant.instances.find((instance) => instance.id === entity)?.kind ?? "";
  return plant.kinds.get(kind)?.faults ?? [];
}

const invalid = (problem: string, id: string) =>
  problem === "" ? {} : { "aria-invalid": true, "aria-describedby": `${id}-problem` };

function InstancePick({ plant, fault, label, id, problem, update, fresh = false }: RowProps) {
  const first = useFocusOnMount<HTMLSelectElement>(fresh);
  return (
    <select
      ref={first}
      id={id}
      className={styles.instance}
      aria-label={`${label} instance`}
      value={fault.entity}
      {...invalid(fault.entity === "" ? problem : "", id)}
      onChange={(event) => {
        update({ entity: event.target.value, fault: "" });
      }}
    >
      <option value="">Choose an instance, or click one on the graph</option>
      {plant.instances.map((instance) => (
        <option key={instance.id} value={instance.id}>
          {`${instance.id} · ${humanise(instance.kind)}`}
        </option>
      ))}
    </select>
  );
}

function FaultPick({ plant, fault, label, id, problem, update }: RowProps) {
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
      {faultsOf(plant, fault.entity).map((name) => (
        <option key={name} value={name}>
          {humanise(name)}
        </option>
      ))}
    </select>
  );
}

function MinutePick({ fault, label, id, problem, update }: RowProps) {
  const timing = fault.entity !== "" && fault.fault !== "" ? problem : "";
  return (
    <input
      id={`${id}-start`}
      className={styles.minute}
      aria-label={`${label} starts at minute`}
      inputMode="decimal"
      value={fault.atMinutes}
      {...invalid(timing, id)}
      onChange={(event) => {
        update({ atMinutes: event.target.value });
      }}
    />
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

function FaultRow({ plant, fault, index, problem, change, fresh, onRemoved }: FaultRowProps) {
  const label = `Fault ${String(index + 1)}`;
  const update = (next: Partial<DraftFault>) => {
    change((draft) => ({ ...draft, faults: draft.faults.map((f) => (f.key === fault.key ? { ...f, ...next } : f)) }));
  };
  const row = { plant, fault, label, id: `launch-fault-${String(index + 1)}`, problem, update, fresh };
  return (
    <li className={styles.faultRow} aria-label={label}>
      <span className={styles.sentence}>
        At minute <MinutePick {...row} />
      </span>
      <InstancePick {...row} />
      <span className={styles.sentence}>
        has a <FaultPick {...row} />
      </span>
      <button
        type="button"
        className={panel.icon}
        aria-label={`Remove ${label.toLowerCase()}`}
        title={`Remove ${label.toLowerCase()}`}
        onClick={() => {
          change((draft) => ({ ...draft, faults: draft.faults.filter((f) => f.key !== fault.key) }));
          onRemoved();
        }}
      >
        <Trash2 size={14} aria-hidden="true" />
      </button>
      {problem !== "" && (
        <small id={`${row.id}-problem`} className={styles.problem}>
          <TriangleAlert size={13} aria-hidden="true" />
          {problem}
        </small>
      )}
    </li>
  );
}

function useAdding(draft: Draft, change: (change: Change) => void, selected: string | null) {
  const [added, setAdded] = useState<number | null>(null);
  const addButton = useRef<HTMLButtonElement>(null);
  const add = () => {
    setAdded(draft.nextKey);
    change((d) => ({
      ...d,
      faults: [...d.faults, { key: d.nextKey, entity: selected ?? "", fault: "", atMinutes: "0" }],
      nextKey: d.nextKey + 1,
    }));
  };
  return { added, addButton, add };
}

interface FaultRowsProps {
  readonly plant: Plant;
  readonly draft: Draft;
  readonly problems: readonly PlanProblem[];
  readonly selected: string | null;
  readonly change: (change: Change) => void;
}

export function FaultRows({ plant, draft, problems, selected, change }: FaultRowsProps) {
  const { added, addButton, add } = useAdding(draft, change, selected);
  const problemOf = (index: number) => problems.find((problem) => problem.field === index)?.text ?? "";
  return (
    <fieldset className={panel.group}>
      <legend>Faults to inject</legend>
      {draft.faults.length === 0 && <p className={panel.hint}>None: the plant runs healthy.</p>}
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
      <button type="button" ref={addButton} className={panel.secondary} onClick={add}>
        <Plus size={14} aria-hidden="true" />
        {selected === null ? "Add fault" : `Add fault on ${selected}`}
      </button>
    </fieldset>
  );
}
