import { useState } from "react";
import type { Report } from "../../../model/review";
import { humanise } from "../../../model/shared/words";
import { immediateWords, runClock, storyOf, storyWindow } from "../../../model/review";
import { ActionStory } from "./ActionStory";
import type { Cell } from "./ReportParts";
import { Rows, Section } from "./ReportParts";
import styles from "./ReportParts.module.css";

export interface Picked {
  readonly atNs: number;
  readonly entity: string;
}

export interface EventActions {
  readonly pick: (atNs: number, entity: string) => void;
  readonly seek: (positionNs: number) => void;
}

const named = (id: string): Cell => ({ text: humanise(id), title: id });
const words = (text: string): Cell => ({ text, title: text });

interface EventsProps {
  readonly report: Report;
  readonly picked: Picked | null;
  readonly actions?: EventActions | undefined;
}

interface CallsProps {
  readonly report: Report;
  readonly chosen: Picked | null;
  readonly onPick: ((index: number) => void) | undefined;
}

function Calls({ report, chosen, onPick }: CallsProps) {
  const at = (ns: number) => runClock(ns, report.startNs);
  const current = report.effectorCalls.findIndex((call) => chosen?.atNs === call.atNs && chosen.entity === call.entity);
  return (
    <Section title="Effector calls">
      <Rows
        rows={report.effectorCalls.map((call) => [
          at(call.atNs),
          call.entity,
          named(call.effector),
          words(call.immediate ? `${call.outcome} · ${immediateWords}` : call.outcome),
        ])}
        columns={["clock", "id", "effector", "outcome"]}
        headers={["Time", "Instance", "Effector", "Outcome"]}
        current={current}
        {...(onPick === undefined ? {} : { onSelect: onPick })}
      />
    </Section>
  );
}

function endCell(d: Report["disturbances"][number], at: (ns: number) => string): Cell {
  const endNs = d.untilNs ?? d.clearedNs;
  if (endNs === null) return { text: "–", title: "Did not end before the run ended" };
  const by = d.clearedBy === "" ? "" : ` by ${d.clearedBy}`;
  return { text: at(endNs), title: `Ends at ${at(endNs)}${by}` };
}

function Disturbances({ report, picked, actions }: EventsProps) {
  const at = (ns: number) => runClock(ns, report.startNs);
  return (
    <Section title="Disturbances">
      <Rows
        rows={report.disturbances.map((d) => [
          { text: at(d.onsetNs), title: `set at ${at(d.injectedNs)}` },
          d.target,
          named(d.name),
          endCell(d, at),
        ])}
        columns={["clock", "id", "label", "clock"]}
        headers={["Starts", "Instance", "Disturbance", "Ends"]}
        current={report.disturbances.findIndex((d) => picked?.atNs === d.onsetNs && picked.entity === d.target)}
        {...(actions === undefined
          ? {}
          : {
              onSelect: (index) => {
                pickDisturbance(report, index, actions);
              },
            })}
      />
    </Section>
  );
}

function pickDisturbance(report: Report, index: number, actions: EventActions): void {
  const disturbance = report.disturbances[index];
  if (disturbance !== undefined) actions.pick(disturbance.onsetNs, disturbance.target);
}

function useChosen(picked: Picked | null) {
  const [chosen, setChosen] = useState<Picked | null>(picked);
  const [seen, setSeen] = useState(picked);
  if (picked !== seen) {
    setSeen(picked);
    if (picked !== null) setChosen(picked);
  }
  return [chosen, setChosen] as const;
}

export function Events({ report, picked, actions }: EventsProps) {
  const [chosen, setChosen] = useChosen(picked);
  const story = storyOf(report, chosen);
  if (report.disturbances.length === 0 && report.effectorCalls.length === 0) {
    return (
      <Section title="Disturbances and effector calls">
        <p className={styles.muted}>None in this run.</p>
      </Section>
    );
  }
  const pick = (index: number) => {
    const call = report.effectorCalls[index];
    if (call === undefined) return;
    setChosen({ atNs: call.atNs, entity: call.entity });
    actions?.pick(call.atNs, call.entity);
  };
  return (
    <>
      <Disturbances report={report} picked={picked} actions={actions} />
      <Calls report={report} chosen={chosen} onPick={actions === undefined ? undefined : pick} />
      {story !== null && actions !== undefined && (
        <ActionStory
          story={story}
          actions={{
            showBefore: (each) => {
              actions.seek(storyWindow(report, each.atNs, "before"));
            },
            showAfter: (each) => {
              actions.seek(storyWindow(report, each.atNs, "after"));
            },
            hide: () => {
              setChosen(null);
            },
          }}
        />
      )}
    </>
  );
}
