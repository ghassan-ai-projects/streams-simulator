import {
  chooseInstrument,
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

interface RowProps extends Pick<TwinProps, "twin" | "issues"> {
  readonly reveal: Reveal | null;
  readonly kind: string;
  readonly metric: Metric;
  readonly choice: MetricChoice;
  readonly instruments: readonly Instrument[] | null;
}

function Numeric({ twin, issues, kind, metric, choice, instruments }: RowProps) {
  const on = (field: string) => twinIssuesOn(issues, { scope: "metric", kind, key: metric.name, field });
  return (
    <>
      <Field label="Instrument" hint="What measures this reading, with its noise and faults." issues={on("instrument")}>
        {(aria) => (
          <InstrumentPicker
            metric={metric.name}
            value={choice.instrument}
            instruments={instruments}
            aria={aria}
            onPick={(ref) => {
              choose(twin, (choices) => chooseInstrument(choices, kind, metric.name, ref));
            }}
          />
        )}
      </Field>
      <Field label="Starts at" hint="Optional. The true value when the twin starts." issues={on("initial")}>
        {(aria) => (
          <NumberInput
            aria={aria}
            name="Starts at"
            value={choice.initial}
            onChange={(initial) => {
              twin.edit((choices) => startAt(choices, kind, metric.name, initial));
            }}
            onCommit={commit(twin)}
          />
        )}
      </Field>
    </>
  );
}

export function ReadingRow(props: RowProps) {
  const { issues, kind, metric } = props;
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
      <ItemIssues messages={twinIssuesOn(issues, { scope: "metric", kind, key: metric.name, field: "" })} />
      <InterfaceNeeds
        issues={issues}
        scope="metric"
        kind={kind}
        name={metric.name}
        selection={{ part: "metric", key: metric.key }}
        reveal={props.reveal}
      />
      {numeric ? (
        <Numeric {...props} />
      ) : (
        <p className={panel.hint}>A plain reading: the twin reports it as it is, with no instrument.</p>
      )}
    </div>
  );
}
