import { humanise } from "../../model/report";
import { Box, Trash2, X } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useState } from "react";
import type { Draft } from "../../app/editorStore";
import type { Instance } from "../../model/plant";
import type { Connection } from "../../model/composition";
import { areasOf, connectionsOf, endpointOf, linkEnd } from "../../model/composition";
import { disconnect, isTemplate, moveInstance, removeInstance, renameInstance } from "../../model/compositionEdits";
import panel from "../panel.module.css";
import { ConnectForm } from "./ConnectForm";
import styles from "./Editor.module.css";
import type { PanelProps } from "./panelProps";
import { useRequestedFocus } from "./panelProps";

type EditorProps = PanelProps & { readonly instance: Instance };

function Rename({
  instance,
  editor,
  request,
  flagged,
}: Pick<EditorProps, "instance" | "editor" | "request"> & { readonly flagged: boolean }) {
  const [name, setName] = useState(instance.id);
  const input = useRequestedFocus<HTMLInputElement>(request, "name");
  const ready = name.trim() !== "" && name.trim() !== instance.id;
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!ready) return;
    request.ask("name");
    editor.edit((doc) => renameInstance(doc, instance.id, name.trim()), name.trim());
  };
  return (
    <form className={panel.field} onSubmit={submit} noValidate>
      <label htmlFor="instance-name">Name</label>
      <span className={panel.pair}>
        <input
          id="instance-name"
          ref={input}
          aria-describedby={flagged ? "instance-issues" : undefined}
          value={name}
          onChange={(event) => {
            setName(event.target.value);
          }}
        />
        <button type="submit" className={panel.secondary} aria-disabled={!ready}>
          Rename
        </button>
      </span>
    </form>
  );
}

function AreaPick({ draft, instance, editor }: Pick<EditorProps, "draft" | "instance" | "editor">) {
  return (
    <label className={panel.field}>
      <span>Area</span>
      <select
        value={instance.area}
        onChange={(event) => {
          const area = event.target.value;
          editor.edit((doc) => moveInstance(doc, instance.id, area));
        }}
      >
        {areasOf(draft.doc).map((area) => (
          <option key={area} value={area}>
            {area}
          </option>
        ))}
      </select>
    </label>
  );
}

function LinkRow({ wire, instance, editor }: { readonly wire: Connection } & Pick<EditorProps, "instance" | "editor">) {
  const end = linkEnd(wire, instance.id);
  const words = `${end.port} ${end.outgoing ? "to" : "from"} ${end.other} ${end.otherPort}`;
  return (
    <li>
      <span className={styles.linkText}>
        <span>{end.port}</span>
        <span aria-hidden="true">{end.outgoing ? "→" : "←"}</span>
        <button
          type="button"
          className={styles.linkOther}
          title={`Select ${end.other}`}
          onClick={() => {
            editor.select(end.other);
          }}
        >
          {end.other}
        </button>
        <span>{end.otherPort}</span>
      </span>
      <button
        type="button"
        className={panel.icon}
        aria-label={`Remove link: ${words}`}
        title="Remove this link"
        onClick={() => {
          editor.edit((doc) => disconnect(doc, wire.index));
        }}
      >
        <Trash2 size={14} aria-hidden="true" />
      </button>
    </li>
  );
}

function Links({ draft, instance, editor }: Pick<EditorProps, "draft" | "instance" | "editor">) {
  const touching = connectionsOf(draft.doc).filter(
    (wire) => endpointOf(wire.from).id === instance.id || endpointOf(wire.to).id === instance.id,
  );
  return (
    <section className={panel.group} aria-labelledby="links-title">
      <h3 id="links-title" className={styles.counted}>
        Links <span className={styles.count}>{touching.length}</span>
      </h3>
      {touching.length === 0 && <p className={panel.hint}>Not linked yet.</p>}
      <ul className={styles.links}>
        {touching.map((wire) => (
          <LinkRow key={wire.index} wire={wire} instance={instance} editor={editor} />
        ))}
      </ul>
    </section>
  );
}

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
      <AreaPick draft={draft} instance={instance} editor={editor} />
      <ConnectForm key={`connect:${instance.id}`} {...{ draft, plant, editor, request, instance }} />
    </>
  );
}

function kindLine(draft: Draft, instance: Instance): string {
  const title = draft.library.find((entry) => entry.component === instance.component)?.title;
  return title === undefined ? humanise(instance.kind) : `${title} · ${humanise(instance.kind)}`;
}

export function InstanceEditor(props: EditorProps) {
  const { draft, editor, request, instance } = props;
  return (
    <aside className={panel.panel} aria-labelledby="instance-title">
      <Head instance={instance} editor={editor} request={request} />
      <div className={panel.body}>
        <Problems id={instance.id} messages={draft.issues?.instances.get(instance.id) ?? []} />
        <p className={panel.hint} title={instance.component ?? undefined}>
          {kindLine(draft, instance)}
        </p>
        <Editing {...props} />
        <Links draft={draft} instance={instance} editor={editor} />
        {!isTemplate(draft.doc, instance.id) && (
          <button
            type="button"
            className={styles.delete}
            onClick={() => {
              request.ask("build");
              editor.edit((doc) => removeInstance(doc, instance.id), null);
            }}
          >
            <Trash2 size={14} aria-hidden="true" />
            {`Delete ${instance.id}`}
          </button>
        )}
      </div>
    </aside>
  );
}
