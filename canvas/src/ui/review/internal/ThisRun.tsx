import type { HandoverState } from "../../../app/review";
import type { Report, VerifyState } from "../../../model/review";
import type { Connection } from "../../../model/review";
import {
  evidenceWords,
  executionWords,
  openedReplayWords,
  purposeWords,
  replayStatusWords,
  scoringWords,
  visibilityWords,
} from "../../../model/review";
import type { ReactNode } from "react";
import { OpenQuestions } from "./OpenQuestions";
import type { ReplayControl } from "./OpenQuestions";
import { Integrity } from "./ReportIntegrity";
import { Section } from "./ReportParts";
import styles from "./ThisRun.module.css";

export interface ThisRunProps {
  readonly report: Report;
  readonly verify: VerifyState;
  readonly connection: Connection;
  readonly replay: ReplayControl;
  readonly handover: HandoverState;
}

function Fact({ label, children }: { readonly label: string; readonly children: ReactNode }) {
  return (
    <div className={styles.fact} data-part={label.toLowerCase().replace(/ /g, "-")}>
      <dt>{label}</dt>
      <dd>{children}</dd>
    </div>
  );
}

function Scoring({ report }: { readonly report: Report }) {
  return (
    <Fact label="Scoring">
      <p className={styles.line}>{scoringWords(report)}</p>
      {report.scoring === null && (
        <details className={styles.developers}>
          <summary>For developers</summary>
          <p>
            Nothing was scored: no consumer verdict was recorded for this run. A consumer attaches through the operator
            endpoint; the canvas does not attach one.
          </p>
        </details>
      )}
    </Fact>
  );
}

function HandoverNotice({ handover }: { readonly handover: HandoverState }) {
  if (handover.phase === "failed")
    return (
      <p role="alert" className={styles.failure}>
        {handover.message}
      </p>
    );
  if (handover.phase !== "saved") return null;
  return (
    <p role="status" className={styles.line}>
      {`Handover saved as ${handover.name}.`}
    </p>
  );
}

function OpenedReplay({ report, verify }: Pick<ThisRunProps, "report" | "verify">) {
  const words = openedReplayWords(report, verify);
  return words === "" ? null : (
    <p className={styles.line} data-part="opened-replay">
      {words}
    </p>
  );
}

function Facts({ report, verify }: Pick<ThisRunProps, "report" | "verify">) {
  return (
    <dl className={styles.facts}>
      <Fact label="Execution">
        <p className={styles.line}>{executionWords(report)}</p>
      </Fact>
      <Fact label="Recording integrity">
        <Integrity report={report} verify={verify} />
        <p
          className={styles.line}
          data-part="replay-status"
        >{`Replay in this report: ${replayStatusWords(report.replay)}`}</p>
        <OpenedReplay report={report} verify={verify} />
      </Fact>
      <Fact label="Evidence">
        {evidenceWords(report).map((line) => (
          <p key={line} className={styles.line}>
            {line}
          </p>
        ))}
      </Fact>
      <Scoring report={report} />
    </dl>
  );
}

export function ThisRun(props: ThisRunProps) {
  const { report, connection, replay, handover } = props;
  return (
    <Section title="This run">
      {report.name !== "" && <p className={styles.name}>{report.name}</p>}
      <p className={styles.purpose} data-part="purpose">
        {purposeWords(report)}
      </p>
      <Facts report={report} verify={props.verify} />
      <p className={styles.visibility} data-part="visibility">
        {visibilityWords(report, connection)}
      </p>
      <OpenQuestions questions={report.openQuestions} replay={replay} />
      <HandoverNotice handover={handover} />
    </Section>
  );
}
