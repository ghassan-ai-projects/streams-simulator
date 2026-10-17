import { panel, shared } from "../../primitives";
import { kindLabel } from "../../../model/shared/words";
import { Box, X } from "lucide-react";
import type { Draft } from "../../../app/editor";
import type { Instance } from "../../../model/plant";
import { areaNodesOf, isTemplate, moveInstance } from "../../../model/editor";
import { cx } from "../../primitives";
import { useCutBottom } from "../../primitives";
import { Choice } from "./Choice";
import { ConnectForm } from "./ConnectForm";
import { DeleteInstance } from "./DeleteInstance";
import { Links } from "./InstanceLinks";
import { Rename } from "./RenameInstance";
import styles from "./Editor.module.css";
import type { PanelProps } from "./panelProps";
import { useRequestedFocus } from "./panelProps";

type EditorProps = PanelProps & { readonly instance: Instance };

function grouped(messages: readonly string[]): readonly (readonly [string, number])[] {
  const counts = new Map<string, number>();
  messages.forEach((message) => counts.set(message, (counts.get(message) ?? 0) + 1));
  return [...counts.entries()];
}

function Problems({ id, messages }: { readonly id: string; readonly messages: readonly string[] }) {
  if (messages.length === 0) return null;
  return (
    <section className={styles.flagged} id="instance-issues" aria-labelledby="instance-issues-title">
      <h3 id="instance-issues-title">{`Issues on ${id}`}</h3>
      <ul>
        {grouped(messages).map(([message, count]) => (
          <li key={message}>{count > 1 ? `${message} (×${String(count)})` : message}</li>
        ))}
      </ul>
    </section>
  );
}

function Head({ instance, editor, request }: Pick<EditorProps, "instance" | "editor" | "request">) {
  const heading = useRequestedFocus<HTMLHeadingElement>(request, "instance");
  return (
    <header className={panel.head}>
      <Box size={16} aria-hidden="true" />
      <h2 id="instance-title" ref={heading} tabIndex={-1}>
        {instance.id}
      </h2>
      <button
        type="button"
        className={panel.icon}
        aria-label="Close the instance"
        onClick={() => {
          request.ask("build");
          editor.select(null);
        }}
      >
        <X size={16} aria-hidden="true" />
      </button>
    </header>
  );
}

function Editing({ draft, plant, editor, request, instance }: EditorProps) {
  const flagged = (draft.issues?.instances.get(instance.id) ?? []).length > 0;
  if (isTemplate(draft.doc, instance.id)) {
    return <p className={panel.hint}>Generated from an id template; change it in the composition file.</p>;
  }
  return (
    <>
      <Rename key={instance.id} instance={instance} editor={editor} request={request} flagged={flagged} />
      <Choice
        label="Area"
        value={instance.area}
        options={areaNodesOf(draft.doc).map((area) => [area.name, area.path])}
        onChange={(area) => {
          editor.edit((doc) => moveInstance(doc, instance.id, area));
        }}
      />
      <ConnectForm key={`connect:${instance.id}`} {...{ draft, plant, editor, request, instance }} />
    </>
  );
}

function kindLine(draft: Draft, instance: Instance): string {
  const title = draft.library.find((entry) => entry.component === instance.component)?.title;
  return title ?? kindLabel(instance.kind);
}

export function InstanceEditor(props: EditorProps) {
  const { draft, editor, request, instance } = props;
  const { ref, cut } = useCutBottom<HTMLElement>();
  return (
    <aside
      ref={ref}
      className={cx(panel.panel, shared.cut)}
      data-cut={cut ? "true" : undefined}
      aria-labelledby="instance-title"
    >
      <Head instance={instance} editor={editor} request={request} />
      <div className={panel.body}>
        <Problems id={instance.id} messages={draft.issues?.instances.get(instance.id) ?? []} />
        <p className={panel.hint} title={instance.component ?? undefined}>
          {kindLine(draft, instance)}
        </p>
        <Editing {...props} />
        <Links draft={draft} instance={instance} editor={editor} />
        {!isTemplate(draft.doc, instance.id) && (
          <DeleteInstance draft={draft} editor={editor} request={request} instance={instance} />
        )}
      </div>
    </aside>
  );
}
