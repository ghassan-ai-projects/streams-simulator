import { cx, panel, shared } from "../../primitives";
import { Plus, Trash2, TriangleAlert } from "lucide-react";
import type { Ref } from "react";
import { useRef, useState } from "react";
import type { Draft } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { kindLabel } from "../../../model/shared/words";
import { useFocusOnMount } from "../../primitives";
import styles from "./LaunchPanel.module.css";
import rows from "./PlanRows.module.css";

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
  readonly prompt?: string;
  readonly onPick: (entity: string) => void;
}

const graphPrompt = "Choose an instance, or click one on the graph";

export function InstancePick({
  plant,
  entity,
  label,
  id,
  problem,
  fresh,
  prompt = graphPrompt,
  onPick,
}: InstancePickProps) {
  const first = useFocusOnMount<HTMLSelectElement>({ when: fresh });
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
      <option value="">{prompt}</option>
      {plant.instances.map((instance) => (
        <option key={instance.id} value={instance.id}>
          {`${instance.id} · ${kindLabel(instance.kind)}`}
        </option>
      ))}
    </select>
  );
}

type When = "starts" | "ends";

interface MinutePickProps {
  readonly value: string;
  readonly label: string;
  readonly id: string;
  readonly problem: string;
  readonly when?: When;
  readonly placeholder?: string;
  readonly onChange: (text: string) => void;
}

function MinutePick({ value, label, id, problem, when = "starts", placeholder, onChange }: MinutePickProps) {
  return (
    <input
      id={`${id}-${when === "starts" ? "start" : "end"}`}
      className={styles.minute}
      aria-label={`${label} ${when} at minute`}
      inputMode="decimal"
      value={value}
      placeholder={placeholder}
      {...invalid(problem, when === "starts" ? id : `${id}-end`)}
      onChange={(event) => {
        onChange(event.target.value);
      }}
    />
  );
}

export function AtMinute({ lead = "At minute", ...pick }: MinutePickProps & { readonly lead?: string }) {
  return (
    <span className={styles.sentence}>
      {lead} <MinutePick {...pick} />
    </span>
  );
}

interface TimesProps {
  readonly label: string;
  readonly id: string;
  readonly atMinutes: string;
  readonly untilMinutes: string;
  readonly startProblem: string;
  readonly endProblem: string;
  readonly endPlaceholder: string;
  readonly update: (next: { readonly atMinutes?: string; readonly untilMinutes?: string }) => void;
}

export function RowTimes(props: TimesProps) {
  const { label, id, atMinutes, untilMinutes, startProblem, endProblem, endPlaceholder, update } = props;
  return (
    <span className={rows.times}>
      <AtMinute
        lead="Starts at minute"
        value={atMinutes}
        label={label}
        id={id}
        problem={startProblem}
        onChange={(next) => {
          update({ atMinutes: next });
        }}
      />
      <AtMinute
        lead="Ends at minute"
        when="ends"
        placeholder={endPlaceholder}
        value={untilMinutes}
        label={label}
        id={id}
        problem={endProblem}
        onChange={(next) => {
          update({ untilMinutes: next });
        }}
      />
    </span>
  );
}

interface RowTailProps {
  readonly label: string;
  readonly id: string;
  readonly problem: string;
  readonly onRemove: () => void;
}

export function ProblemNote({ id, text }: { readonly id: string; readonly text: string }) {
  if (text === "") return null;
  return (
    <small id={`${id}-problem`} className={styles.problem}>
      <TriangleAlert size={13} aria-hidden="true" />
      {text}
    </small>
  );
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
      <ProblemNote id={id} text={problem} />
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

interface AddRowProps {
  readonly label: string;
  readonly buttonRef: Ref<HTMLButtonElement>;
  readonly disabled?: boolean;
  readonly onAdd: () => void;
}

export function AddRow({ label, buttonRef, disabled = false, onAdd }: AddRowProps) {
  return (
    <button
      type="button"
      ref={buttonRef}
      disabled={disabled}
      className={cx(shared.secondary, panel.secondary)}
      onClick={onAdd}
    >
      <Plus size={14} aria-hidden="true" />
      {label}
    </button>
  );
}
