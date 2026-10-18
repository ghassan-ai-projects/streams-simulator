import { issuesOn, kindFacts, renameKind } from "../../../model/devices";
import type { KindFile } from "../../../model/devices";
import { cx, panel, useFocusOnMount } from "../../primitives";
import styles from "./Devices.module.css";
import { Field, ItemIssues, ItemNeeds, TextBox } from "./Field";
import { FormTitle } from "./FormTitle";
import type { FormProps } from "./formProps";

export function KindForm({ kind, devices, issues, needs }: FormProps & { readonly kind: KindFile }) {
  const ref = useFocusOnMount<HTMLInputElement>({ when: kind.kind === "" });
  const facts = kindFacts(kind.metrics.length, kind.commands.length);
  const on = (field: string) => issuesOn(issues, "kind", "", field);
  const need = (field: string) => issuesOn(needs, "kind", "", field);
  return (
    <section className={cx(panel.body, styles.form)} aria-labelledby="device-title">
      <FormTitle id="device-title" words="Device" />
      <p className={panel.hint}>{`${facts}. Choose one in the outline to edit it, or add a new one.`}</p>
      <ItemIssues messages={[...on(""), ...on("metrics"), ...on("commands")]} />
      <ItemNeeds messages={[...need("metrics"), ...need("commands")]} />
      <Field label="Name" field="kind" issues={on("kind")} needs={need("kind")}>
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
