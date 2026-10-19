import type { TwinReport, TwinRequest } from "../../../model/review";
import { judgeReasonWords, ruleWords, verdictWords } from "../../../model/review";
import { useFocusOnMount } from "../../primitives";
import parts from "./ReportParts.module.css";
import styles from "./Twin.module.css";
import { VerdictChip } from "./TwinVerdict";

function Counts({ report }: { readonly report: TwinReport }) {
  const { summary } = report;
  const entries: readonly (readonly [string, number])[] = [
    ["Readings compared", summary.compared],
    ["Within tolerance", summary.within],
    ["Beyond tolerance", summary.diverged],
    ["Not compared", summary.notCompared],
    ["Commands applied", summary.commandsApplied],
    ["Commands not applied", summary.commandsNotApplied],
    ["Metrics that cannot be judged", summary.metricsCannotJudge],
    ["Readings with no named metric", summary.unattributed],
  ];
  return (
    <div role="group" aria-label="Twin comparison counts">
      <dl className={styles.counts}>
        {entries.map(([label, value]) => (
          <div key={label} className={styles.count}>
            <dt>{label}</dt>
            <dd>{String(value)}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}

function typedWords(request: TwinRequest): string {
  return request.tolerance.trim() === "" ? "No tolerance was typed." : `Tolerances as typed: ${request.tolerance}`;
}

export function TwinResult({ report, request }: { readonly report: TwinReport; readonly request: TwinRequest }) {
  const heading = useFocusOnMount<HTMLHeadingElement>();
  return (
    <div className={styles.result} role="group" aria-label="Twin comparison result">
      <div className={styles.verdictRow}>
        <VerdictChip verdict={report.verdict} />
        <h3 ref={heading} tabIndex={-1} className={styles.heading}>
          {verdictWords[report.verdict].line}
        </h3>
      </div>
      {report.reason !== null && <p className={parts.muted}>{judgeReasonWords[report.reason]}</p>}
      <Counts report={report} />
      <p className={styles.meta}>{ruleWords(report)}</p>
      <p className={styles.meta}>{`Twin: domain ${report.twinDomain}. ${typedWords(request)}`}</p>
    </div>
  );
}
