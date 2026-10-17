import { Trash2 } from "lucide-react";
import type { ReactNode } from "react";
import { Fragment } from "react";
import type { Connection, LinkEnd } from "../../../model/editor";
import { connectionsTouching, disconnect, linkEnd, linkParts, linkWords } from "../../../model/editor";
import type { Instance } from "../../../model/plant";
import { cx, panel, shared } from "../../primitives";
import type { PanelProps } from "./panelProps";
import styles from "./Editor.module.css";

type EditorProps = PanelProps & { readonly instance: Instance };

function Ends({ end, other }: { readonly end: LinkEnd; readonly other: ReactNode }) {
  return linkParts(end).map((part, at) => {
    if (part.kind === "other") return <Fragment key={at}>{other}</Fragment>;
    if (part.kind === "arrow")
      return (
        <span key={at} aria-hidden="true">
          →
        </span>
      );
    return <span key={at}>{part.text}</span>;
  });
}

function LinkRow({ wire, instance, editor }: { readonly wire: Connection } & Pick<EditorProps, "instance" | "editor">) {
  const end = linkEnd(wire, instance.id);
  return (
    <li>
      <span className={styles.linkText}>
        <Ends
          end={end}
          other={
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
          }
        />
      </span>
      <button
        type="button"
        className={cx(shared.secondary, panel.secondary)}
        aria-label={`Remove link: ${linkWords(end)}`}
        title="Remove this link"
        onClick={() => {
          editor.edit((doc) => disconnect(doc, wire.index));
        }}
      >
        <Trash2 size={14} aria-hidden="true" />
        Remove
      </button>
    </li>
  );
}

export function Links({ draft, instance, editor }: Pick<EditorProps, "draft" | "instance" | "editor">) {
  const touching = connectionsTouching(draft.doc, instance.id);
  return (
    <section className={panel.group} aria-labelledby="links-title">
      <h3 id="links-title" className={styles.counted}>
        Links <span className={styles.count}>{touching.length}</span>
      </h3>
      {touching.length === 0 && <p className={panel.hint}>No links yet. Use “Link an output” above.</p>}
      <ul className={styles.links}>
        {touching.map((wire) => (
          <LinkRow key={wire.index} wire={wire} instance={instance} editor={editor} />
        ))}
      </ul>
    </section>
  );
}
