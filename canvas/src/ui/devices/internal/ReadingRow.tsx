import {
  chooseInstrument,
  placeKey,
  rangeWords,
  readingFacts,
  startAt,
  takesInstrument,
  twinIssuesOn,
} from "../../../model/devices";
import type { Instrument, Metric, MetricChoice } from "../../../model/devices";
import { panel } from "../../primitives";
import tw from "./Behaviour.module.css";
import { Field, ItemIssues, NumberInput } from "./Field";
import { InstrumentPicker } from "./InstrumentPicker";
import { InterfaceNeeds } from "./InterfaceNeeds";
import { choose, commit } from "./twinProps";
import type { Reveal, TwinProps } from "./twinProps";

interface RowProps extends Pick<TwinProps, "twin" | "issues" | "needs"> {
  readonly reveal: Reveal | null;
  readonly member: string;
  readonly metric: Metric;
  readonly choice: MetricChoice;
  readonly instruments: readonly Instrument[] | null;
}

type NumericFields = Pick<RowProps, "twin" | "member" | "metric" | "choice" | "instruments" | "reveal"> & {
  readonly on: (field: string) => readonly string[];
};

function InstrumentField({ twin, member, metric, choice, instruments, reveal, on }: NumericFields) {
  const place = (field: string) => placeKey({ scope: "metric", kind: member, key: metric.key, field });
  return (
    <Field
      label="Instrument"
      field="instrument"
      place={place("instrument")}
      hint="What measures this reading, with its noise and faults."
      issues={on("instrument")}
    >
      {(aria) => (
        <InstrumentPicker
          metric={metric.name}
          value={choice.instrument}
          instruments={instruments}
          reveal={
            reveal === null
              ? null
              : () => {
                  reveal({ part: "metric", key: metric.key }, "eu_range");
                }
          }
          aria={aria}
          onPick={(ref) => {
            choose(twin, (choices) => chooseInstrument(choices, member, metric.key, ref));
          }}
        />
      )}
    </Field>
  );
}

function InitialField({ twin, member, metric, choice, on }: NumericFields) {
  return (
    <Field
      label="Starts at"
      field="initial"
      place={placeKey({ scope: "metric", kind: member, key: metric.key, field: "initial" })}
      hint="Optional. The true value when the twin starts."
      issues={on("initial")}
    >
      {(aria) => (
        <NumberInput
          aria={aria}
          value={choice.initial}
          onChange={(initial) => {
            twin.edit((choices) => startAt(choices, member, metric.key, initial));
          }}
          onCommit={commit(twin)}
        />
      )}
    </Field>
  );
}

function Numeric({ twin, issues, member, metric, choice, instruments, reveal }: RowProps) {
  const on = (field: string) => twinIssuesOn(issues, { scope: "metric", kind: member, key: metric.key, field });
  const fields = { twin, member, metric, choice, instruments, reveal, on };
  return (
    <>
      <InstrumentField {...fields} />
      <InitialField {...fields} />
    </>
  );
}

export function ReadingRow(props: RowProps) {
  const { issues, member, metric } = props;
  const numeric = takesInstrument(metric);
  const facts = numeric
    ? `${readingFacts(metric)} · ${rangeWords(metric.range.min, metric.range.max)}`
    : readingFacts(metric);
  return (
    <div role="group" aria-label={`Reading ${metric.name}`} className={tw.reading}>
      <h4 className={tw.readingHead}>
        <strong>{metric.name}</strong>
        <span className={panel.hint}>{facts}</span>
      </h4>
      <ItemIssues messages={twinIssuesOn(issues, { scope: "metric", kind: member, key: metric.key, field: "" })} />
      <InterfaceNeeds needs={props.needs} part="metric" itemKey={metric.key} reveal={props.reveal} />
      {numeric ? (
        <Numeric {...props} />
      ) : (
        <p className={panel.hint}>A plain reading: the twin reports it as it is, with no instrument.</p>
      )}
    </div>
  );
}
