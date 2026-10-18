import { Plus, Trash2 } from "lucide-react";
import type { ChangeEvent } from "react";
import { useState } from "react";
import type { RealRow, RowField } from "../../../model/launch";
import { freeInstance, rowIssues, withRealRow, withRowField, withoutRealRow } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { kindLabel } from "../../../model/shared/words";
import { cx, panel, shared, useFocusOnMount } from "../../primitives";
import type { DraftEdit } from "./SessionFields";
import { IssueNote, flagged } from "./SessionFields";
import launch from "./LaunchPanel.module.css";
import styles from "./Session.module.css";

interface RowProps extends DraftEdit {
  readonly plant: Plant;
  readonly row: RealRow;
  readonly index: number;
  readonly fresh: boolean;
}

const fieldWords: Readonly<Record<RowField, string>> = { instance: "Instance", board: "Board", device: "Device" };

interface Control {
  readonly id: string;
  readonly "aria-label": string;
  readonly "aria-invalid"?: boolean;
  readonly "aria-describedby"?: string;
  readonly value: string;
  readonly onChange: (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void;
}

function InstanceSelect({ plant, fresh, ...control }: { readonly plant: Plant; readonly fresh: boolean } & Control) {
  const first = useFocusOnMount<HTMLSelectElement>({ when: fresh });
  return (
    <select ref={first} {...control}>
      {plant.instances.map((instance) => (
        <option key={instance.id} value={instance.id}>{`${instance.id} · ${kindLabel(instance.kind)}`}</option>
      ))}
    </select>
  );
}

function RowControl({
  plant,
  draft,
  change,
  issues,
  row,
  index,
  fresh,
  field,
}: RowProps & { readonly field: RowField }) {
  const id = `real-${String(row.key)}-${field}`;
  const placed = rowIssues(issues, index, field);
  const control = {
    id,
    "aria-label": `Real board ${String(index + 1)} ${field}`,
    value: row[field],
    ...flagged(placed, `${id}-issue`),
    onChange: (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
      change(withRowField(draft, row.key, field, event.target.value));
    },
  };
  return (
    <div className={styles.rowField}>
      <label htmlFor={id}>{fieldWords[field]}</label>
      {field === "instance" ? <InstanceSelect plant={plant} fresh={fresh} {...control} /> : <input {...control} />}
      <IssueNote id={`${id}-issue`} issues={placed} />
    </div>
  );
}

function RowItem(props: RowProps) {
  const label = `Real board ${String(props.index + 1)}`;
  return (
    <li className={cx(launch.faultRow, styles.realRow)} aria-label={label}>
      <RowControl {...props} field="instance" />
      <button
        type="button"
        className={panel.icon}
        aria-label={`Remove ${label.toLowerCase()}`}
        title={`Remove ${label.toLowerCase()}`}
        onClick={() => {
          props.change(withoutRealRow(props.draft, props.row.key));
        }}
      >
        <Trash2 size={14} aria-hidden="true" />
      </button>
      <RowControl {...props} field="board" />
      <RowControl {...props} field="device" />
    </li>
  );
}

function AddRow({
  plant,
  draft,
  change,
  onAdded,
}: DraftEdit & { readonly plant: Plant; readonly onAdded: () => void }) {
  const free = freeInstance(
    plant.instances.map((instance) => instance.id),
    draft.real,
  );
  return (
    <button
      type="button"
      className={cx(shared.secondary, panel.secondary)}
      aria-disabled={free === null}
      title={free === null ? "Every instance of the plant already has a real board." : undefined}
      onClick={() => {
        if (free === null) return;
        onAdded();
        change(withRealRow(draft, free));
      }}
    >
      <Plus size={14} aria-hidden="true" />
      Add a real board
    </button>
  );
}

export function RealRows({ plant, draft, change, issues }: DraftEdit & { readonly plant: Plant }) {
  const [fresh, setFresh] = useState<number | null>(null);
  return (
    <fieldset className={panel.group}>
      <legend>Real boards</legend>
      {draft.real.length === 0 && (
        <p className={panel.hint}>
          Every instance is simulated. Add a real board to make an instance the board a gateway announces.
        </p>
      )}
      <ol className={launch.faults} aria-label="Real boards">
        {draft.real.map((row, index) => (
          <RowItem
            key={row.key}
            plant={plant}
            draft={draft}
            change={change}
            issues={issues}
            row={row}
            index={index}
            fresh={row.key === fresh}
          />
        ))}
      </ol>
      <AddRow
        plant={plant}
        draft={draft}
        change={change}
        issues={issues}
        onAdded={() => {
          setFresh(draft.nextKey);
        }}
      />
    </fieldset>
  );
}
