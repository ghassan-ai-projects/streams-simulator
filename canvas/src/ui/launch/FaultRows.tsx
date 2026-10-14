import { Plus, Trash2 } from "lucide-react";
import type { Plant } from "../../model/plant";
import { humanise } from "../../model/report";
import type { Draft, DraftFault } from "./draft";
import styles from "./LaunchPanel.module.css";

export type Change = (draft: Draft) => Draft;

interface RowProps {
  readonly plant: Plant;
  readonly fault: DraftFault;
  readonly label: string;
  readonly update: (next: Partial<DraftFault>) => void;
}

function faultsOf(plant: Plant, entity: string): readonly string[] {
  const kind = plant.instances.find((instance) => instance.id === entity)?.kind ?? "";
  return plant.kinds.get(kind)?.faults ?? [];
}

function InstancePick({ plant, fault, label, update }: RowProps) {
  return (
    <select
      aria-label={`${label} instance`}
      value={fault.entity}
      onChange={(event) => {
        update({ entity: event.target.value, fault: "" });
      }}
    >
      <option value="">Instance…</option>
      {plant.instances.map((instance) => (
        <option key={instance.id} value={instance.id}>
          {instance.id}
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
      <option value="">Fault…</option>
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
        aria-label={`${label} minute`}
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
}: {
  readonly plant: Plant;
  readonly fault: DraftFault;
  readonly index: number;
  readonly change: (change: Change) => void;
}) {
  const label = `Fault ${String(index + 1)}`;
  const update = (next: Partial<DraftFault>) => {
    change((draft) => ({ ...draft, faults: draft.faults.map((f) => (f.key === fault.key ? { ...f, ...next } : f)) }));
  };
  const row = { plant, fault, label, update };
  return (
    <li className={styles.faultRow}>
      <InstancePick {...row} />
      <FaultPick {...row} />
      <MinutePick {...row} />
      <button
        type="button"
        className={styles.icon}
        aria-label={`Remove ${label.toLowerCase()}`}
        onClick={() => {
          change((draft) => ({ ...draft, faults: draft.faults.filter((f) => f.key !== fault.key) }));
        }}
      >
        <Trash2 size={14} aria-hidden="true" />
      </button>
    </li>
  );
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
  const add = () => {
    change((d) => ({
      ...d,
      faults: [...d.faults, { key: d.nextKey, entity: "", fault: "", atMinutes: "0" }],
      nextKey: d.nextKey + 1,
    }));
  };
  return (
    <fieldset className={styles.group}>
      <legend>Faults to inject</legend>
      {draft.faults.length === 0 && <p className={styles.hint}>None: the plant runs healthy.</p>}
      <ol className={styles.faults}>
        {draft.faults.map((fault, index) => (
          <FaultRow key={fault.key} plant={plant} fault={fault} index={index} change={change} />
        ))}
      </ol>
      <button type="button" className={styles.secondary} onClick={add}>
        <Plus size={14} aria-hidden="true" />
        Add fault
      </button>
    </fieldset>
  );
}
