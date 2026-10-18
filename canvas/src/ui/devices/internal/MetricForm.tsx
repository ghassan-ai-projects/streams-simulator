import { changeMetric, issuesOn, rangeHint, typeChoices, unitChoices } from "../../../model/devices";
import type { Metric } from "../../../model/devices";
import { cx, panel, useFocusOnMount } from "../../primitives";
import styles from "./Devices.module.css";
import { Field, ItemIssues, ItemNeeds, NumberInput, RemoveButton, Select, TextBox } from "./Field";
import { FormTitle } from "./FormTitle";
import type { FormProps, MetricEdit, MetricFields } from "./formProps";
import { Behaviour } from "./MetricExtras";
import notes from "./Notes.module.css";

function Identity({ metric, on, need, change }: MetricFields) {
  const ref = useFocusOnMount<HTMLInputElement>({ when: metric.name === "" });
  return (
    <>
      <Field
        label="Name"
        field="name"
        hint="Start with a letter or digit, then letters, digits and . _ : - only."
        issues={on("name")}
        needs={need("name")}
      >
        {(aria) => (
          <TextBox
            aria={aria}
            inputRef={ref}
            value={metric.name}
            onChange={(name) => {
              change({ name });
            }}
          />
        )}
      </Field>
      <Field label="Type" field="data_type" issues={on("data_type")} needs={need("data_type")}>
        {(aria) => (
          <Select
            aria={aria}
            value={metric.dataType}
            choices={typeChoices}
            onChange={(dataType) => {
              change({ dataType });
            }}
          />
        )}
      </Field>
    </>
  );
}

function Units({ metric, on, need, change }: MetricFields) {
  return (
    <Field
      label="Unit"
      field="unit"
      hint="A unit code such as Cel, %, Pa or rpm."
      issues={on("unit")}
      needs={need("unit")}
    >
      {(aria) => (
        <>
          <TextBox
            aria={aria}
            list={`${aria.id}-units`}
            value={metric.unit}
            onChange={(unit) => {
              change({ unit });
            }}
          />
          <datalist id={`${aria.id}-units`}>
            {unitChoices.map(([value, label]) => (
              <option key={value} value={value} label={label} />
            ))}
          </datalist>
        </>
      )}
    </Field>
  );
}

function Bounds({ metric, on, need, change }: MetricFields) {
  const { range } = metric;
  return (
    <Field
      label="Range"
      field="eu_range"
      group
      caution={rangeHint(range)}
      issues={on("eu_range")}
      needs={need("eu_range")}
    >
      {(aria) => (
        <span className={notes.bounds}>
          <span className={notes.bound}>
            <label htmlFor={aria.id}>Minimum</label>
            <NumberInput
              value={range.min}
              aria={aria}
              onChange={(min) => {
                change({ range: { ...range, min } });
              }}
            />
          </span>
          <span className={notes.bound}>
            <label htmlFor={`${aria.id}-max`}>Maximum</label>
            <NumberInput
              value={range.max}
              aria={{ ...aria, id: `${aria.id}-max`, "data-field": undefined }}
              onChange={(max) => {
                change({ range: { ...range, max } });
              }}
            />
          </span>
        </span>
      )}
    </Field>
  );
}

function Numbers(fields: MetricFields) {
  const { metric, on, need, change } = fields;
  const numeric = metric.dataType === "number" || metric.dataType === "integer";
  const ranged = metric.range.min !== null || metric.range.max !== null;
  const stated = metric.resolution !== null || need("resolution").length > 0;
  return (
    <>
      {(numeric || ranged) && <Bounds {...fields} />}
      {(numeric || stated) && (
        <Field
          label="Resolution"
          field="resolution"
          hint="The smallest step the reading can show."
          issues={on("resolution")}
          needs={need("resolution")}
        >
          {(aria) => (
            <NumberInput
              value={metric.resolution}
              aria={aria}
              onChange={(resolution) => {
                change({ resolution });
              }}
            />
          )}
        </Field>
      )}
    </>
  );
}

export function MetricForm({ metric, devices, issues, needs }: FormProps & { readonly metric: Metric }) {
  const on = (field: string) => issuesOn(issues, "metric", metric.key, field);
  const need = (field: string) => issuesOn(needs, "metric", metric.key, field);
  const change = (edit: MetricEdit) => {
    devices.edit((kind) => changeMetric(kind, metric.key, edit));
  };
  const fields = { metric, on, need, change };
  return (
    <section className={cx(panel.body, styles.form)} aria-labelledby="reading-title">
      <FormTitle id="reading-title" words="Reading" />
      <ItemIssues messages={on("")} />
      <ItemNeeds messages={need("")} />
      <Identity {...fields} />
      <Units {...fields} />
      <Numbers {...fields} />
      <Behaviour {...fields} />
      <RemoveButton
        words="Remove this reading"
        onRemove={() => {
          devices.remove({ part: "metric", key: metric.key });
        }}
      />
    </section>
  );
}
