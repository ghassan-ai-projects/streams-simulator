import { Box, Trash2, X } from "lucide-react";
import type { SyntheticEvent } from "react";
import { useState } from "react";
import type { Instance } from "../../model/plant";
import { areasOf, connectionsOf, endpointOf } from "../../model/composition";
import { disconnect, moveInstance, removeInstance, renameInstance } from "../../model/compositionEdits";
import panel from "../panel.module.css";
import { ConnectForm } from "./ConnectForm";
import styles from "./Editor.module.css";
import type { PanelProps } from "./panelProps";

type EditorProps = PanelProps & { readonly instance: Instance };

function Rename({ instance, editor }: Pick<EditorProps, "instance" | "editor">) {
  const [name, setName] = useState(instance.id);
  const ready = name.trim() !== "" && name.trim() !== instance.id;
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (ready) editor.edit((doc) => renameInstance(doc, instance.id, name.trim()), name.trim());
  };
  return (
    <form className={panel.field} onSubmit={submit} noValidate>
      <label htmlFor="instance-name">Name</label>
      <span className={panel.pair}>
        <input
          id="instance-name"
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

function Links({ draft, instance, editor }: Pick<EditorProps, "draft" | "instance" | "editor">) {
  const touching = connectionsOf(draft.doc).filter(
    (wire) => endpointOf(wire.from).id === instance.id || endpointOf(wire.to).id === instance.id,
  );
  return (
    <section className={panel.group} aria-labelledby="links-title">
      <h3 id="links-title">{`Links (${String(touching.length)})`}</h3>
      {touching.length === 0 && <p className={panel.hint}>Not linked yet.</p>}
      <ul className={styles.links}>
        {touching.map((wire) => (
          <li key={wire.index}>
            <span>{`${wire.from} → ${wire.to}`}</span>
            <button
              type="button"
              className={panel.icon}
              aria-label={`Remove link ${wire.from} to ${wire.to}`}
              onClick={() => {
                editor.edit((doc) => disconnect(doc, wire.index));
              }}
            >
              <Trash2 size={14} aria-hidden="true" />
            </button>
          </li>
        ))}
      </ul>
    </section>
  );
}

function Problems({ messages }: { readonly messages: readonly string[] }) {
  if (messages.length === 0) return null;
  return (
    <ul className={panel.problems} aria-label="Issues">
      {messages.map((message) => (
        <li key={message}>{message}</li>
      ))}
    </ul>
  );
}

function Head({ instance, editor }: Pick<EditorProps, "instance" | "editor">) {
  return (
    <header className={panel.head}>
      <Box size={16} aria-hidden="true" />
      <h2 id="instance-title">{instance.id}</h2>
      <button
        type="button"
        className={panel.icon}
        aria-label="Close the instance"
        onClick={() => {
          editor.select(null);
        }}
      >
        <X size={16} aria-hidden="true" />
      </button>
    </header>
  );
}

export function InstanceEditor({ draft, plant, editor, instance }: EditorProps) {
  return (
    <aside className={panel.panel} aria-labelledby="instance-title">
      <Head instance={instance} editor={editor} />
      <div className={panel.body}>
        <Problems messages={draft.issues?.instances.get(instance.id) ?? []} />
        <p className={panel.hint}>{`${instance.kind} · ${instance.component ?? ""}`}</p>
        <Rename key={instance.id} instance={instance} editor={editor} />
        <AreaPick draft={draft} instance={instance} editor={editor} />
        <ConnectForm key={`connect:${instance.id}`} draft={draft} plant={plant} editor={editor} instance={instance} />
        <Links draft={draft} instance={instance} editor={editor} />
        <button
          type="button"
          className={styles.delete}
          onClick={() => {
            editor.edit((doc) => removeInstance(doc, instance.id), null);
          }}
        >
          <Trash2 size={14} aria-hidden="true" />
          {`Delete ${instance.id}`}
        </button>
      </div>
    </aside>
  );
}
