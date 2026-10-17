import { Plus, TriangleAlert } from "lucide-react";
import type { ReactNode } from "react";
import type { DeviceSession, Draft } from "../../../app/devices";
import { commandFacts, readingFacts } from "../../../model/devices";
import type { FieldIssue, Selection } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import styles from "./Devices.module.css";

interface OutlineProps {
  readonly draft: Draft;
  readonly devices: DeviceSession;
  readonly issues: readonly FieldIssue[];
}

interface ItemProps {
  readonly name: string;
  readonly facts: string;
  readonly current: boolean;
  readonly flagged: boolean;
  readonly onSelect: () => void;
}

function Item({ name, facts, current, flagged, onSelect }: ItemProps) {
  return (
    <li>
      <button type="button" className={styles.item} aria-current={current ? "true" : undefined} onClick={onSelect}>
        <span className={styles.itemLine}>
          <span className={styles.itemName}>{name}</span>
          {flagged && <TriangleAlert className={styles.flag} size={13} role="img" aria-label="Has issues" />}
        </span>
        {facts !== "" && <span className={styles.itemFacts}>{facts}</span>}
      </button>
    </li>
  );
}

interface SectionProps {
  readonly title: string;
  readonly addWords: string;
  readonly empty: string | null;
  readonly onAdd: () => void;
  readonly children: ReactNode;
}

function Section({ title, addWords, empty, onAdd, children }: SectionProps) {
  return (
    <>
      <div className={styles.listHead}>
        <h2 className={shared.sectionTitle}>{title}</h2>
        <button type="button" className={shared.secondary} aria-label={addWords} onClick={onAdd}>
          <Plus size={14} aria-hidden="true" />
          Add
        </button>
      </div>
      {empty === null ? (
        <ul className={styles.items}>{children}</ul>
      ) : (
        <p className={cx(panel.hint, styles.addRow)}>{empty}</p>
      )}
    </>
  );
}

function chosen(selected: Selection, part: Selection["part"], key: string): boolean {
  return selected.part === part && (selected.part === "kind" || selected.key === key);
}

function flaggedItem(issues: readonly FieldIssue[], part: FieldIssue["part"], key: string): boolean {
  return issues.some((issue) => issue.part === part && issue.key === key);
}

interface ListProps extends OutlineProps {
  readonly select: (selection: Selection) => () => void;
}

function Readings({ draft, devices, issues, select }: ListProps) {
  const { doc, selected } = draft;
  const none = "No readings yet. Add what the device reports, such as a temperature.";
  return (
    <Section
      title="Readings"
      addWords="Add a reading"
      empty={doc.metrics.length === 0 ? none : null}
      onAdd={() => {
        devices.addMetric();
      }}
    >
      {doc.metrics.map((metric) => (
        <Item
          key={metric.key}
          name={metric.name === "" ? "Unnamed reading" : metric.name}
          facts={readingFacts(metric)}
          current={chosen(selected, "metric", metric.key)}
          flagged={flaggedItem(issues, "metric", metric.key)}
          onSelect={select({ part: "metric", key: metric.key })}
        />
      ))}
    </Section>
  );
}

function Commands({ draft, devices, issues, select }: ListProps) {
  const { doc, selected } = draft;
  const none = "No commands yet. Add what the device can be told to do.";
  return (
    <Section
      title="Commands"
      addWords="Add a command"
      empty={doc.commands.length === 0 ? none : null}
      onAdd={() => {
        devices.addCommand();
      }}
    >
      {doc.commands.map((command) => (
        <Item
          key={command.key}
          name={command.name === "" ? "Unnamed command" : command.name}
          facts={commandFacts(command)}
          current={chosen(selected, "command", command.key)}
          flagged={flaggedItem(issues, "command", command.key)}
          onSelect={select({ part: "command", key: command.key })}
        />
      ))}
    </Section>
  );
}

export function DeviceOutline(props: OutlineProps) {
  const { draft, devices, issues } = props;
  const select = (selection: Selection) => () => {
    devices.select(selection);
  };
  return (
    <nav className={styles.outline} aria-label="Device outline">
      <ul className={styles.items}>
        <Item
          name={draft.doc.kind === "" ? "Unnamed device" : draft.doc.kind}
          facts="Device"
          current={chosen(draft.selected, "kind", "")}
          flagged={flaggedItem(issues, "kind", "")}
          onSelect={select({ part: "kind" })}
        />
      </ul>
      <Readings {...props} select={select} />
      <Commands {...props} select={select} />
    </nav>
  );
}
