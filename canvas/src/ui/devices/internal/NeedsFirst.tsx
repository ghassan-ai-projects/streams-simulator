import { Info } from "lucide-react";
import type { Draft } from "../../../app/devices";
import { twinNeedsOf } from "../../../app/devices";
import { issueSelection, issueTitle } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import styles from "./Devices.module.css";
import { FormTitle } from "./FormTitle";
import notes from "./Notes.module.css";
import type { Reveal } from "./twinProps";

export function NeedsFirst({ device, reveal }: { readonly device: Draft; readonly reveal: Reveal }) {
  return (
    <section className={cx(panel.body, styles.form)} aria-label="What the twin needs first">
      <FormTitle id="needs-first-title" words="Before the behaviour" interim />
      <p className={panel.hint}>
        A twin cannot be started until the device states what it needs. Set these in the Interface step.
      </p>
      <ul className={notes.needList}>
        {twinNeedsOf(device).map((issue, index) => (
          <li key={`${String(index)}:${issue.message}`}>
            <Info size={14} aria-hidden="true" />
            <button
              type="button"
              className={cx(shared.secondary, panel.secondary)}
              onClick={() => {
                reveal(issueSelection(issue), issue.field);
              }}
            >
              {`${issueTitle(device.doc, issue)}: ${issue.text}`}
            </button>
          </li>
        ))}
      </ul>
    </section>
  );
}
