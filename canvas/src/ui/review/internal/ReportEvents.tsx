import type { Report } from "../../../model/review";
import { humanise } from "../../../model/shared/words";
import { runClock } from "../../../model/review";
import type { Cell } from "./ReportParts";
import { Rows, Section } from "./ReportParts";
import styles from "./RunScreen.module.css";

export interface Picked {
  readonly atNs: number;
  readonly entity: string;
}

const named = (id: string): Cell => ({ text: humanise(id), title: id });
const words = (text: string): Cell => ({ text, title: text });

export function Events({ report, picked }: { readonly report: Report; readonly picked: Picked | null }) {
  const at = (ns: number) => runClock(ns, report.startNs);
  const hit = (ns: number, entity: string) => picked?.atNs === ns && picked.entity === entity;
  if (report.disturbances.length === 0 && report.effectorCalls.length === 0) {
    return (
      <Section title="Disturbances and effector calls">
        <p className={styles.muted}>None in this run.</p>
      </Section>
    );
  }
  return (
    <>
      <Section title="Disturbances">
        <Rows
          rows={report.disturbances.map((d) => [
            { text: at(d.onsetNs), title: `set at ${at(d.injectedNs)}` },
            d.target,
            named(d.name),
          ])}
          columns={["clock", "id", "label"]}
          headers={["Starts", "Instance", "Disturbance"]}
          current={report.disturbances.findIndex((d) => hit(d.onsetNs, d.target))}
        />
      </Section>
      <Section title="Effector calls">
        <Rows
          rows={report.effectorCalls.map((call) => [
            at(call.atNs),
            call.entity,
            named(call.effector),
            words(call.outcome),
          ])}
          columns={["clock", "id", "effector", "outcome"]}
          headers={["Time", "Instance", "Effector", "Outcome"]}
          current={report.effectorCalls.findIndex((call) => hit(call.atNs, call.entity))}
        />
      </Section>
    </>
  );
}
