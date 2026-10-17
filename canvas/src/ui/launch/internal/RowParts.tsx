import { panel } from "../../primitives";
import { Trash2, TriangleAlert } from "lucide-react";
import { useRef, useState } from "react";
import type { Plant } from "../../../model/plant";
import { kindLabel } from "../../../model/shared/words";
import type { Draft } from "./draft";
import { useFocusOnMount } from "../../primitives";
import styles from "./LaunchPanel.module.css";

export type Change = (draft: Draft) => Draft;

export const invalid = (problem: string, id: string) =>
  problem === "" ? {} : { "aria-invalid": true, "aria-describedby": `${id}-problem` };

interface InstancePickProps {
  readonly plant: Plant;
  readonly entity: string;
  readonly label: string;
  readonly id: string;
  readonly problem: string;
  readonly fresh: boolean;
  readonly onPick: (entity: string) => void;
}

export function InstancePick({ plant, entity, label, id, problem, fresh, onPick }: InstancePickProps) {
  const first = useFocusOnMount<HTMLSelectElement>(fresh);
  return (
    <select
      ref={first}
      id={id}
      className={styles.instance}
      aria-label={`${label} instance`}
      value={entity}
      {...invalid(problem, id)}
      onChange={(event) => {
        onPick(event.target.value);
      }}
    >
      <option value="">Choose an instance, or click one on the graph</option>
      {plant.instances.map((instance) => (
        <option key={instance.id} value={instance.id}>
          {`${instance.id} · ${kindLabel(instance.kind)}`}
        </option>
      ))}
    </select>
  );
}

interface MinutePickProps {
  readonly value: string;
  readonly label: string;
  readonly id: string;
  readonly problem: string;
  readonly onChange: (text: string) => void;
}

function MinutePick({ value, label, id, problem, onChange }: MinutePickProps) {
  return (
    <input
      id={`${id}-start`}
      className={styles.minute}
      aria-label={`${label} starts at minute`}
      inputMode="decimal"
      value={value}
      {...invalid(problem, id)}
      onChange={(event) => {
        onChange(event.target.value);
      }}
    />
  );
}

export function AtMinute({ value, label, id, problem, onChange }: MinutePickProps) {
  return (
    <span className={styles.sentence}>
      At minute <MinutePick value={value} label={label} id={id} problem={problem} onChange={onChange} />
    </span>
  );
}

interface RowTailProps {
  readonly label: string;
  readonly id: string;
  readonly problem: string;
  readonly onRemove: () => void;
}

export function RowTail({ label, id, problem, onRemove }: RowTailProps) {
  return (
    <>
      <button
        type="button"
        className={panel.icon}
        aria-label={`Remove ${label.toLowerCase()}`}
        title={`Remove ${label.toLowerCase()}`}
        onClick={onRemove}
      >
        <Trash2 size={14} aria-hidden="true" />
      </button>
      {problem !== "" && (
        <small id={`${id}-problem`} className={styles.problem}>
          <TriangleAlert size={13} aria-hidden="true" />
          {problem}
        </small>
      )}
    </>
  );
}

export function useAdding(change: (change: Change) => void, append: (draft: Draft, key: number) => Draft) {
  const [added, setAdded] = useState<number | null>(null);
  const addButton = useRef<HTMLButtonElement>(null);
  const add = (nextKey: number) => {
    setAdded(nextKey);
    change((draft) => append({ ...draft, nextKey: draft.nextKey + 1 }, draft.nextKey));
  };
  return { added, addButton, add };
}
