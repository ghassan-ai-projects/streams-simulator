import { useState } from "react";
import { enumOptions, withEnumOptions } from "../../../model/devices";
import { Field, TextBox, Toggle } from "./Field";
import type { MetricFields } from "./formProps";

function Options({ metric, on, need, change }: MetricFields) {
  const [typed, setTyped] = useState(() => enumOptions(metric));
  return (
    <Field
      label="Options"
      field="enum_values"
      hint="The values the reading can have, separated by commas."
      issues={on("enum_values")}
      needs={need("enum_values")}
    >
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

function ScanClass({ metric, on, need, change }: MetricFields) {
  return (
    <Field
      label="Scan class"
      field="scan_class"
      hint="How often it is read, such as fast or slow."
      issues={on("scan_class")}
      needs={need("scan_class")}
    >
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
  );
}

function Setpoint({ metric, on, need, change }: MetricFields) {
  return (
    <Field label="Setpoint" field="writable" issues={on("writable")} needs={need("writable")}>
      {(aria) => (
        <Toggle
          aria={aria}
          checked={metric.writable}
          words="A command can set this reading"
          onChange={(writable) => {
            change({ writable });
          }}
        />
      )}
    </Field>
  );
}

function Description({ metric, on, need, change }: MetricFields) {
  return (
    <Field
      label="Description"
      field="description"
      hint="Up to 200 characters."
      issues={on("description")}
      needs={need("description")}
    >
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
  );
}

export function Behaviour(fields: MetricFields) {
  return (
    <>
      {fields.metric.dataType === "enum" && <Options {...fields} />}
      <ScanClass {...fields} />
      <Setpoint {...fields} />
      <Description {...fields} />
    </>
  );
}
