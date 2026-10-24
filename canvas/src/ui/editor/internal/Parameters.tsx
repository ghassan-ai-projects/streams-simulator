import { useId } from "react";
import type { Draft, EditorSession } from "../../../app/editor";
import { isDirty } from "../../../app/editor";
import { componentOf, equipmentOf, fieldIssues, setParameter, setPreset } from "../../../model/editor";
import type { Equipment } from "../../../model/editor";
import type { ComponentSpec } from "../../../model/plant";
import { panel } from "../../primitives";
import { ParameterField } from "./ParameterField";
import styles from "./Parameters.module.css";

interface ParametersProps {
  readonly draft: Draft;
  readonly editor: EditorSession;
  readonly id: string;
}

function PresetOptions({ spec }: { readonly spec: ComponentSpec }) {
  const presets = spec.presets ?? [];
  const fallback = presets.find((preset) => preset.name === spec.defaultPreset);
  return (
    <>
      <option value="">{fallback === undefined ? "The component's own size" : `Default: ${fallback.title}`}</option>
      {presets.map((preset) => (
        <option key={preset.name} value={preset.name}>
          {preset.title}
        </option>
      ))}
    </>
  );
}

function PresetField({
  spec,
  entry,
  props,
}: {
  readonly spec: ComponentSpec;
  readonly entry: Equipment;
  readonly props: ParametersProps;
}) {
  const { draft, editor, id } = props;
  const fieldId = useId();
  const refusals = isDirty(draft) ? [] : fieldIssues(draft.issues?.listed ?? [], entry.index, "preset");
  return (
    <div className={styles.parameter}>
      <label htmlFor={fieldId} className={styles.name}>
        Size
      </label>
      <select
        id={fieldId}
        value={entry.preset}
        onChange={(event) => {
          editor.edit((doc) => setPreset(doc, id, event.target.value));
        }}
      >
        <PresetOptions spec={spec} />
      </select>
      {refusals.map((message) => (
        <p key={message} className={styles.refusal} role="alert">
          {message}
        </p>
      ))}
    </div>
  );
}

function sharing(entry: Equipment, id: string): string {
  return `Changing these gives ${id} its own entry in the composition; ${String(entry.names.length - 1)} other instances keep theirs.`;
}

function Fields({
  props,
  entry,
  spec,
}: {
  readonly props: ParametersProps;
  readonly entry: Equipment;
  readonly spec: ComponentSpec;
}) {
  const { draft, editor, id } = props;
  const listed = isDirty(draft) ? [] : (draft.issues?.listed ?? []);
  return (
    <>
      {(spec.presets ?? []).length > 0 && <PresetField spec={spec} entry={entry} props={props} />}
      {(spec.parameters ?? []).map((parameter) => (
        <ParameterField
          key={parameter.name}
          spec={parameter}
          value={entry.parameters[parameter.name]}
          refusals={fieldIssues(listed, entry.index, parameter.name)}
          onChange={(value) => {
            editor.edit((doc) => setParameter(doc, id, parameter.name, value));
          }}
        />
      ))}
      {entry.names.length > 1 && <p className={panel.hint}>{sharing(entry, id)}</p>}
    </>
  );
}

function hasFields(spec: ComponentSpec | undefined): spec is ComponentSpec {
  return spec !== undefined && ((spec.parameters ?? []).length > 0 || (spec.presets ?? []).length > 0);
}

export function Parameters(props: ParametersProps) {
  const heading = useId();
  const { draft, id } = props;
  const entry = equipmentOf(draft.doc).find((each) => each.names.includes(id));
  const spec = entry === undefined ? undefined : draft.specs.get(componentOf(entry.use));
  if (entry === undefined || !hasFields(spec)) return null;
  return (
    <section className={panel.group} aria-labelledby={heading}>
      <h3 id={heading}>Size and parameters</h3>
      <Fields props={props} entry={entry} spec={spec} />
    </section>
  );
}
