import { CircleAlert, Wrench } from "lucide-react";
import panel from "../panel.module.css";
import { AddEquipment } from "./AddEquipment";
import styles from "./Editor.module.css";
import { InstanceEditor } from "./InstanceEditor";
import type { PanelProps } from "./panelProps";
import { useRequestedFocus } from "./panelProps";

function Issues({ draft, editor, request }: Omit<PanelProps, "plant">) {
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
                  request.ask("instance");
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

function Build({ draft, editor, request }: Omit<PanelProps, "plant">) {
  const heading = useRequestedFocus<HTMLHeadingElement>(request, "build");
  return (
    <aside className={panel.panel} aria-labelledby="build-title">
      <header className={panel.head}>
        <Wrench size={16} aria-hidden="true" />
        <h2 id="build-title" ref={heading} tabIndex={-1}>
          Build
        </h2>
      </header>
      <div className={panel.body}>
        <Issues draft={draft} editor={editor} request={request} />
        <AddEquipment draft={draft} editor={editor} request={request} />
        <p className={panel.hint}>
          Select an instance to rename it, move it to another area, link its ports or delete it. Link ports with Link an
          output in the instance panel, or at 100 % zoom drag from an output pin to an input pin.
        </p>
      </div>
    </aside>
  );
}

export function EditorPanel(props: PanelProps) {
  const instance = props.plant.instances.find((entry) => entry.id === props.draft.selected);
  if (instance !== undefined) return <InstanceEditor {...props} instance={instance} />;
  return <Build draft={props.draft} editor={props.editor} request={props.request} />;
}
