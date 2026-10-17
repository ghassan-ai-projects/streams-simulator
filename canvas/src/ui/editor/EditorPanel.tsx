import { CircleAlert, Wrench } from "lucide-react";
import { isDirty } from "../../app/editorStore";
import panel from "../panel.module.css";
import type { IssueGroup, ListedIssue } from "../../model/compositionEdits";
import { groupIssues, profileWeightOf, removeProfileWeight } from "../../model/compositionEdits";
import { cx } from "../cx";
import shared from "../shared.module.css";
import { useCutBottom } from "../useCutBottom";
import { AddArea } from "./AddArea";
import { AreaEditor } from "./AreaEditor";
import { AddEquipment } from "./AddEquipment";
import styles from "./Editor.module.css";
import { InstanceEditor } from "./InstanceEditor";
import type { PanelProps } from "./panelProps";
import { useRequestedFocus } from "./panelProps";

function WeightIssue({ issue, editor }: { readonly issue: ListedIssue; readonly editor: PanelProps["editor"] }) {
  const weight = profileWeightOf(issue);
  return (
    <span title={issue.message}>
      {`Profile "${weight?.profile ?? ""}" has a weight for ${weight?.fault ?? ""}, which no instance declares. `}
      <button
        type="button"
        className={styles.action}
        onClick={() => {
          editor.edit((doc) => removeProfileWeight(doc, issue.pointer));
        }}
      >
        Remove this weight
      </button>
    </span>
  );
}

function IssueRow({ issue, editor, request }: { readonly issue: IssueGroup } & Omit<PanelProps, "plant" | "draft">) {
  const times = issue.count > 1 ? ` (×${String(issue.count)})` : "";
  if (issue.target === null && profileWeightOf(issue) !== null) return <WeightIssue issue={issue} editor={editor} />;
  if (issue.target === null) {
    return (
      <span>
        {`${issue.message}${times}`}
        {issue.pointer !== "" && <code className={styles.where}>{`in ${issue.pointer}`}</code>}
      </span>
    );
  }
  const target = issue.target;
  return (
    <button
      type="button"
      className={styles.issue}
      onClick={() => {
        request.ask("instance");
        editor.select(target);
      }}
    >
      <strong>{target}</strong>
      {`${issue.message}${times}`}
    </button>
  );
}

function Issues({ draft, editor, request }: Omit<PanelProps, "plant">) {
  const groups = groupIssues(draft.issues?.listed ?? []);
  if (groups.length === 0) return null;
  return (
    <section className={panel.group} aria-labelledby="issues-title">
      <h3 id="issues-title">Issues from the binary</h3>
      {isDirty(draft) && <p className={panel.hint}>These are from the last save. Save again to check your changes.</p>}
      <ul className={styles.issues}>
        {groups.map((issue, index) => (
          <li key={`${String(index)}:${issue.message}`}>
            <CircleAlert size={14} aria-hidden="true" />
            <IssueRow issue={issue} editor={editor} request={request} />
          </li>
        ))}
      </ul>
    </section>
  );
}

function FirstStep() {
  return (
    <p className={panel.hint}>
      Add the first piece of equipment: search the library below, name it and choose an area. Then link an output to an
      input and save to compose the plant.
    </p>
  );
}

function Build({ draft, plant, editor, request }: PanelProps) {
  const heading = useRequestedFocus<HTMLHeadingElement>(request, "build");
  const { ref, cut } = useCutBottom<HTMLElement>([draft, plant]);
  return (
    <aside
      ref={ref}
      className={cx(panel.panel, shared.cut)}
      data-cut={cut ? "true" : undefined}
      aria-labelledby="build-title"
    >
      <header className={panel.head}>
        <Wrench size={16} aria-hidden="true" />
        <h2 id="build-title" ref={heading} tabIndex={-1}>
          Build
        </h2>
      </header>
      <div className={panel.body}>
        <Issues draft={draft} editor={editor} request={request} />
        {plant.instances.length === 0 && <FirstStep />}
        <AddEquipment draft={draft} plant={plant} editor={editor} request={request} />
        <AddArea draft={draft} editor={editor} />
        <AreaEditor draft={draft} editor={editor} />
        <p className={panel.hint}>
          Select an instance to rename, move, link or delete it. To link two instances, drag from an output pin to an
          input pin, or use “Link an output” on the selected instance.
        </p>
        <p className={panel.hint}>
          Sensors are not added by hand: every instance is measured on each of its ports, and the plant gets one channel
          for each.
        </p>
      </div>
    </aside>
  );
}

export function EditorPanel(props: PanelProps) {
  const instance = props.plant.instances.find((entry) => entry.id === props.draft.selected);
  if (instance !== undefined) return <InstanceEditor {...props} instance={instance} />;
  return <Build draft={props.draft} plant={props.plant} editor={props.editor} request={props.request} />;
}
