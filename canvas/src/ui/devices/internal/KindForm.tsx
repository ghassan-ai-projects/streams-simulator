import { idRule, issuesOn, kindFacts, looksLikeId, renameKind } from "../../../model/devices";
import type { KindFile } from "../../../model/devices";
import { cx, panel, useFocusOnMount } from "../../primitives";
import styles from "./Devices.module.css";
import { Field, TextBox } from "./Field";
import type { FormProps } from "./formProps";

export function KindForm({ kind, devices, issues }: FormProps & { readonly kind: KindFile }) {
  const ref = useFocusOnMount<HTMLInputElement>({ when: kind.kind === "" });
  const likely = kind.kind === "" || looksLikeId(kind.kind);
  const facts = kindFacts(kind.metrics.length, kind.commands.length);
  return (
    <section className={cx(panel.body, styles.form)} aria-labelledby="device-title">
      <h2 id="device-title" className={styles.formTitle}>
        Device
      </h2>
      <p className={panel.hint}>{`${facts}. Choose one in the outline to edit it, or add a new one.`}</p>
      <Field
        label="Name"
        hint={idRule}
        caution={likely ? "" : "The binary is likely to refuse this name."}
        issues={issuesOn(issues, "kind", "", "kind")}
      >
        {(aria) => (
          <TextBox
            aria={aria}
            inputRef={ref}
            value={kind.kind}
            onChange={(name) => {
              devices.edit((current) => renameKind(current, name));
            }}
          />
        )}
      </Field>
    </section>
  );
}
