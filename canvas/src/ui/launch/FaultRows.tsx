import { Plus, Trash2 } from "lucide-react";
import type { Plant } from "../../model/plant";
import { humanise } from "../../model/report";
import type { Draft, DraftFault } from "./draft";
import { useRef, useState } from "react";
import panel from "../panel.module.css";
import { useFocusOnMount } from "../useFocusOnMount";
import styles from "./LaunchPanel.module.css";

export type Change = (draft: Draft) => Draft;

interface RowProps {
  readonly fresh?: boolean;
  readonly plant: Plant;
  readonly fault: DraftFault;
  readonly label: string;
  readonly update: (next: Partial<DraftFault>) => void;
}

function faultsOf(plant: Plant, entity: string): readonly string[] {
  const kind = plant.instances.find((instance) => instance.id === entity)?.kind ?? "";
  return plant.kinds.get(kind)?.faults ?? [];
}

function InstancePick({ plant, fault, label, update, fresh = false }: RowProps) {
  const first = useFocusOnMount<HTMLSelectElement>(fresh);
  return (
    <select
      ref={first}
      aria-label={`${label} instance`}
      value={fault.entity}
      onChange={(event) => {
        update({ entity: event.target.value, fault: "" });
      }}
    >
      <option value="">Choose instance</option>
      {plant.instances.map((instance) => (
        <option key={instance.id} value={instance.id}>
          {`${instance.id} · ${humanise(instance.kind)}`}
        </option>
      ))}
    </select>
  );
}

function FaultPick({ plant, fault, label, update }: RowProps) {
  return (
    <select
      aria-label={`${label} fault`}
      value={fault.fault}
      disabled={fault.entity === ""}
      onChange={(event) => {
        update({ fault: event.target.value });
      }}
    >
      <option value="">Choose fault</option>
      {faultsOf(plant, fault.entity).map((id) => (
        <option key={id} value={id}>
          {humanise(id)}
        </option>
      ))}
    </select>
  );
}

function MinutePick({ fault, label, update }: RowProps) {
  return (
    <label className={styles.at}>
      <span>Starts at minute</span>
      <input
        aria-label={`${label} starts at minute`}
        inputMode="numeric"
        value={fault.atMinutes}
        onChange={(event) => {
          update({ atMinutes: event.target.value });
        }}
      />
    </label>
  );
}

function FaultRow({
  plant,
  fault,
  index,
  change,
  fresh,
  onRemoved,
}: {
  readonly plant: Plant;
  readonly fault: DraftFault;
  readonly index: number;
  readonly change: (change: Change) => void;
  readonly fresh: boolean;
  readonly onRemoved: () => void;
}) {
  const label = `Fault ${String(index + 1)}`;
  const update = (next: Partial<DraftFault>) => {
    change((draft) => ({ ...draft, faults: draft.faults.map((f) => (f.key === fault.key ? { ...f, ...next } : f)) }));
  };
  const row = { plant, fault, label, update, fresh };
  return (
    <li className={styles.faultRow}>
      <InstancePick {...row} />
      <FaultPick {...row} />
      <MinutePick {...row} />
      <button
        type="button"
        className={panel.icon}
        aria-label={`Remove ${label.toLowerCase()}`}
        onClick={() => {
          change((draft) => ({ ...draft, faults: draft.faults.filter((f) => f.key !== fault.key) }));
          onRemoved();
        }}
      >
        <Trash2 size={14} aria-hidden="true" />
      </button>
    </li>
  );
}

function useAdding(draft: Draft, change: (change: Change) => void) {
  const [added, setAdded] = useState<number | null>(null);
  const addButton = useRef<HTMLButtonElement>(null);
  const add = () => {
    setAdded(draft.nextKey);
    change((d) => ({
      ...d,
      faults: [...d.faults, { key: d.nextKey, entity: "", fault: "", atMinutes: "0" }],
      nextKey: d.nextKey + 1,
    }));
  };
  return { added, addButton, add };
}

export function FaultRows({
  plant,
  draft,
  change,
}: {
  readonly plant: Plant;
  readonly draft: Draft;
  readonly change: (change: Change) => void;
}) {
  const { added, addButton, add } = useAdding(draft, change);
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
        Add fault
      </button>
    </fieldset>
  );
}
