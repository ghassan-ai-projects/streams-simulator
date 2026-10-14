import { CircleAlert, Wrench } from "lucide-react";
import panel from "../panel.module.css";
import { AddEquipment } from "./AddEquipment";
import styles from "./Editor.module.css";
import { InstanceEditor } from "./InstanceEditor";
import type { PanelProps } from "./panelProps";

function Issues({ draft, editor }: Omit<PanelProps, "plant">) {
  const listed = draft.issues?.listed ?? [];
  if (listed.length === 0) return null;
  return (
    <section className={panel.group} aria-labelledby="issues-title">
      <h3 id="issues-title">Issues from the binary</h3>
      <ul className={styles.issues}>
        {listed.map((issue, index) => (
          <li key={`${String(index)}:${issue.message}`}>
            <CircleAlert size={14} aria-hidden="true" />
            {issue.target === null ? (
              <span>{issue.message}</span>
            ) : (
              <button
                type="button"
                className={styles.issue}
                onClick={() => {
                  editor.select(issue.target);
                }}
              >
                <strong>{issue.target}</strong>
                {issue.message}
              </button>
            )}
          </li>
        ))}
      </ul>
    </section>
  );
}

export function EditorPanel({ draft, plant, editor }: PanelProps) {
  const instance = plant.instances.find((entry) => entry.id === draft.selected);
  if (instance !== undefined) return <InstanceEditor draft={draft} plant={plant} editor={editor} instance={instance} />;
  return (
    <aside className={panel.panel} aria-labelledby="build-title">
      <header className={panel.head}>
        <Wrench size={16} aria-hidden="true" />
        <h2 id="build-title">Build</h2>
      </header>
      <div className={panel.body}>
        <Issues draft={draft} editor={editor} />
        <AddEquipment draft={draft} editor={editor} />
        <p className={panel.hint}>
          Select an instance to rename it, move it to another area, link its ports or delete it. Drag from an output pin
          to an input pin to link two instances.
        </p>
      </div>
    </aside>
  );
}
