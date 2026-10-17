import { useState } from "react";
import {
  changeMetric,
  enumOptions,
  idRule,
  issuesOn,
  looksLikeId,
  rangeHint,
  typeChoices,
  unitChoices,
  withEnumOptions,
} from "../../../model/devices";
import type { Metric } from "../../../model/devices";
import { cx, panel, useFocusOnMount } from "../../primitives";
import styles from "./Devices.module.css";
import { Field, ItemIssues, NumberInput, RemoveButton, Select, TextBox } from "./Field";
import type { FormProps, MetricEdit } from "./formProps";

interface MetricFields {
  readonly metric: Metric;
  readonly on: (field: string) => readonly string[];
  readonly change: (edit: MetricEdit) => void;
}

function Identity({ metric, on, change }: MetricFields) {
  const ref = useFocusOnMount<HTMLInputElement>({ when: metric.name === "" });
  const likely = metric.name === "" || looksLikeId(metric.name);
  const caution = likely ? "" : "The binary is likely to refuse this name.";
  return (
    <>
      <Field label="Name" hint={idRule} caution={caution} issues={on("name")}>
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
      <Field label="Type" issues={on("data_type")}>
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

function Units({ metric, on, change }: MetricFields) {
  return (
    <Field label="Unit" hint="A unit code such as Cel, %, Pa or rpm." issues={on("unit")}>
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

function Bounds({ metric, on, change }: MetricFields) {
  const { range } = metric;
  return (
    <Field label="Range" caution={rangeHint(range)} issues={on("eu_range")}>
      {(aria) => (
        <span className={panel.pair}>
          <NumberInput
            value={range.min}
            aria={aria}
            name="Minimum"
            onChange={(min) => {
              change({ range: { ...range, min } });
            }}
          />
          <NumberInput
            value={range.max}
            aria={{ ...aria, id: `${aria.id}-max` }}
            name="Maximum"
            onChange={(max) => {
              change({ range: { ...range, max } });
            }}
          />
        </span>
      )}
    </Field>
  );
}

function Numbers({ metric, on, change }: MetricFields) {
  const numeric = metric.dataType === "number" || metric.dataType === "integer";
  const ranged = metric.range.min !== null || metric.range.max !== null;
  return (
    <>
      {(numeric || ranged) && <Bounds metric={metric} on={on} change={change} />}
      {numeric && (
        <Field label="Resolution" hint="The smallest step the reading can show." issues={on("resolution")}>
          {(aria) => (
            <NumberInput
              value={metric.resolution}
              aria={aria}
              name="Resolution"
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

function Options({ metric, on, change }: MetricFields) {
  const [typed, setTyped] = useState(() => enumOptions(metric));
  return (
    <Field label="Options" hint="The values the reading can have, separated by commas." issues={on("enum_values")}>
      {(aria) => (
        <TextBox
          aria={aria}
          value={typed}
          onChange={(next) => {
            setTyped(next);
            change({ extra: withEnumOptions(metric.extra, next) });
          }}
        />
      )}
    </Field>
  );
}

function Setpoint({ metric, on, change }: MetricFields) {
  return (
    <Field label="Setpoint" issues={on("writable")}>
      {(aria) => (
        <label className={styles.check}>
          <input
            {...aria}
            type="checkbox"
            checked={metric.writable}
            onChange={(event) => {
              change({ writable: event.target.checked });
            }}
          />
          A command can set this reading
        </label>
      )}
    </Field>
  );
}

function Behaviour({ metric, on, change }: MetricFields) {
  return (
    <>
      {metric.dataType === "enum" && <Options metric={metric} on={on} change={change} />}
      <Field label="Scan class" hint="How often it is read, such as fast or slow." issues={on("scan_class")}>
        {(aria) => (
          <TextBox
            aria={aria}
            value={metric.scanClass}
            onChange={(scanClass) => {
              change({ scanClass });
            }}
          />
        )}
      </Field>
      <Setpoint metric={metric} on={on} change={change} />
      <Field label="Description" hint="Up to 200 characters." issues={on("description")}>
        {(aria) => (
          <TextBox
            aria={aria}
            maxLength={200}
            value={metric.description}
            onChange={(description) => {
              change({ description });
            }}
          />
        )}
      </Field>
    </>
  );
}

export function MetricForm({ metric, devices, issues }: FormProps & { readonly metric: Metric }) {
  const on = (field: string) => issuesOn(issues, "metric", metric.key, field);
  const change = (edit: MetricEdit) => {
    devices.edit((kind) => changeMetric(kind, metric.key, edit));
  };
  return (
    <section className={cx(panel.body, styles.form)} aria-labelledby="reading-title">
      <h2 id="reading-title" className={styles.formTitle}>
        Reading
      </h2>
      <ItemIssues messages={on("")} />
      <Identity metric={metric} on={on} change={change} />
      <Units metric={metric} on={on} change={change} />
      <Numbers metric={metric} on={on} change={change} />
      <Behaviour metric={metric} on={on} change={change} />
      <RemoveButton
        words="Remove this reading"
        onRemove={() => {
          devices.remove({ part: "metric", key: metric.key });
        }}
      />
    </section>
  );
}
