import { Plus, Trash2 } from "lucide-react";
import type { ChangeEvent } from "react";
import type { RealRow, RowField } from "../../../model/launch";
import { rowIssue, withRealRow, withRowField, withoutRealRow } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { kindLabel } from "../../../model/shared/words";
import { cx, panel, shared } from "../../primitives";
import type { DraftEdit } from "./SessionFields";
import { IssueNote, flagged } from "./SessionFields";
import launch from "./LaunchPanel.module.css";
import styles from "./Session.module.css";

interface RowProps extends DraftEdit {
  readonly plant: Plant;
  readonly row: RealRow;
  readonly index: number;
}

const fieldWords: Readonly<Record<RowField, string>> = { instance: "Instance", board: "Board", device: "Device" };

function RowControl({ plant, draft, change, issues, row, index, field }: RowProps & { readonly field: RowField }) {
  const id = `real-${String(index)}-${field}-issue`;
  const issue = rowIssue(issues, index, field);
  const control = {
    "aria-label": `Real board ${String(index + 1)} ${field}`,
    value: row[field],
    ...flagged(issue, id),
    onChange: (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
      change(withRowField(draft, index, field, event.target.value));
    },
  };
  return (
    <label className={styles.rowField}>
      <span>{fieldWords[field]}</span>
      {field === "instance" ? (
        <select {...control}>
          {plant.instances.map((instance) => (
            <option key={instance.id} value={instance.id}>{`${instance.id} · ${kindLabel(instance.kind)}`}</option>
          ))}
        </select>
      ) : (
        <input {...control} />
      )}
      <IssueNote id={id} issue={issue} />
    </label>
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
          props.change(withoutRealRow(props.draft, props.index));
        }}
      >
        <Trash2 size={14} aria-hidden="true" />
      </button>
      <RowControl {...props} field="board" />
      <RowControl {...props} field="device" />
    </li>
  );
}

function freeInstance(plant: Plant, rows: readonly RealRow[]): string {
  return plant.instances.find((instance) => !rows.some((row) => row.instance === instance.id))?.id ?? "";
}

export function RealRows({ plant, draft, change, issues }: DraftEdit & { readonly plant: Plant }) {
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
            key={String(index)}
            plant={plant}
            draft={draft}
            change={change}
            issues={issues}
            row={row}
            index={index}
          />
        ))}
      </ol>
      <button
        type="button"
        className={cx(shared.secondary, panel.secondary)}
        onClick={() => {
          change(withRealRow(draft, freeInstance(plant, draft.real)));
        }}
      >
        <Plus size={14} aria-hidden="true" />
        Add a real board
      </button>
    </fieldset>
  );
}
