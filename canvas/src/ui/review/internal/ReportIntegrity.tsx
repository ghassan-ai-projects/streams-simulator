import type { Report, VerifyState } from "../../../model/review";
import { fileLine, integrityLine, replayLine } from "../../../model/review";
import { grouped } from "../../../model/shared/units";
import { spaced } from "../../../model/shared/words";
import { Rows, Section } from "./ReportParts";
import parts from "./ReportParts.module.css";
import styles from "./ReportPanel.module.css";

function ReplayCheck({ verify }: { readonly verify: VerifyState }) {
  const detail = verify.phase === "ready" ? verify.value.detail : "";
  return (
    <Section title="Replay check">
      <p className={styles.verdict}>{replayLine(verify)}</p>
      <details className={styles.inset}>
        <summary>For developers</summary>
        <p className={parts.muted}>Checked with the streamsim verify command over the run folder.</p>
        {detail !== "" && <pre>{detail}</pre>}
      </details>
    </Section>
  );
}

const count = (report: Report, key: string) => (report.ended ? grouped(report.counts[key] ?? 0) : "n/a");

function Deliveries({ report }: { readonly report: Report }) {
  const dropped = Object.entries(report.dropped).map(([reason, n]) => [spaced(reason), grouped(n)]);
  return (
    <Section title="Deliveries" note={report.ended ? undefined : "run not closed"}>
      <Rows
        rows={[...["emitted", "delivered", "dropped"].map((key) => [key, count(report, key)]), ...dropped]}
        columns={["text", "number"]}
        headers={["Records", "Count"]}
      />
    </Section>
  );
}

function FileCheck({ report }: { readonly report: Report }) {
  return (
    <Section title="File check">
      <p className={parts.muted}>{fileLine(report)}</p>
      <details className={parts.muted}>
        <summary>For developers</summary>
        <p className={parts.muted}>
          The report compares the files in the run folder with the digests the run recorded.
        </p>
        <Rows
          rows={report.checks.map((check) => [{ text: spaced(check.name), title: check.name }, spaced(check.status)])}
          columns={["text", "text"]}
          headers={["File", "Check"]}
        />
      </details>
    </Section>
  );
}

export function Integrity({ report, verify }: { readonly report: Report; readonly verify: VerifyState }) {
  return (
    <details className={styles.integrity}>
      <summary>{integrityLine(report, verify)}</summary>
      <ReplayCheck verify={verify} />
      <FileCheck report={report} />
      <Deliveries report={report} />
    </details>
  );
}
