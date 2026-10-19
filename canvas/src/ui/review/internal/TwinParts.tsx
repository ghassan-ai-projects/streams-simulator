import type { TwinCommand, TwinDivergence, TwinMetric, TwinReport } from "../../../model/review";
import {
  commandStatusWords,
  divergenceLine,
  inForceWords,
  metricLine,
  noDivergenceWords,
  notAppliedWords,
  notComparedWords,
  originWords,
  runClock,
  unitOfMetric,
} from "../../../model/review";
import { cx, shared } from "../../primitives";
import parts from "./ReportParts.module.css";
import { Section } from "./ReportParts";
import styles from "./Twin.module.css";
import { VerdictChip } from "./TwinVerdict";

function statusWords(command: TwinCommand): string {
  const reason = command.reason === null ? "" : `: ${notAppliedWords[command.reason]}`;
  return `${commandStatusWords[command.status]}${reason}`;
}

function CommandItem({ command, startNs }: { readonly command: TwinCommand; readonly startNs: number }) {
  const at = command.atNs === null ? "" : ` · ${runClock(command.atNs, startNs)}`;
  return (
    <li className={styles.item}>
      <span className={styles.name}>{`${command.commandId} · ${command.command} on ${command.device}${at}`}</span>
      <span>{statusWords(command)}</span>
      {command.origin === "node" && <span className={styles.meta}>{originWords.node}</span>}
      {command.detail !== "" && <span className={styles.meta}>{command.detail}</span>}
    </li>
  );
}

export function CommandsSection({ report, startNs }: { readonly report: TwinReport; readonly startNs: number }) {
  return (
    <Section title="Twin commands">
      <p className={parts.muted}>
        A command reaches the twin only when the board accepted it and the node did not hold it back.
      </p>
      {report.commands.length === 0 ? (
        <p className={parts.muted}>The host sent no command to a real board in this run.</p>
      ) : (
        <ul className={styles.list} aria-label="Commands and the twin">
          {report.commands.map((command) => (
            <CommandItem key={command.commandId} command={command} startNs={startNs} />
          ))}
        </ul>
      )}
    </Section>
  );
}

interface DivergenceProps {
  readonly divergence: TwinDivergence;
  readonly unit: string;
  readonly startNs: number;
  readonly onSeek: (ns: number) => void;
}

function DivergenceItem({ divergence, unit, startNs, onSeek }: DivergenceProps) {
  return (
    <li className={styles.item}>
      <button
        type="button"
        className={cx(shared.secondary, styles.jump)}
        aria-label={`Go to the start of ${divergence.name}`}
        onClick={() => {
          onSeek(divergence.firstNs);
        }}
      >
        {divergence.name}
      </button>
      <span>{divergenceLine(divergence, unit, startNs)}</span>
      <span className={styles.meta}>{`Commands in force: ${inForceWords(divergence)}`}</span>
    </li>
  );
}

interface DivergencesProps {
  readonly report: TwinReport;
  readonly startNs: number;
  readonly onSeek: (ns: number) => void;
}

export function DivergencesSection({ report, startNs, onSeek }: DivergencesProps) {
  return (
    <Section title="Where the twin and the board part ways">
      {report.divergences.length === 0 ? (
        <p className={parts.muted}>{noDivergenceWords[report.verdict]}</p>
      ) : (
        <ul className={styles.list} aria-label="Divergences">
          {report.divergences.map((divergence) => (
            <DivergenceItem
              key={divergence.name}
              divergence={divergence}
              unit={unitOfMetric(report, divergence)}
              startNs={startNs}
              onSeek={onSeek}
            />
          ))}
        </ul>
      )}
    </Section>
  );
}

function reasonsLine(metric: TwinMetric): string {
  const each = metric.notCompared.map((entry) => `${String(entry.count)} × ${notComparedWords[entry.reason]}`);
  return `Not compared: ${each.join("; ")}.`;
}

function MetricItem({ metric }: { readonly metric: TwinMetric }) {
  return (
    <li className={styles.item}>
      <span className={styles.verdictRow}>
        <span className={styles.name}>{`${metric.device}.${metric.metric}`}</span>
        <VerdictChip verdict={metric.verdict} />
      </span>
      <span>{metricLine(metric)}</span>
      {metric.notCompared.length > 0 && <span className={styles.meta}>{reasonsLine(metric)}</span>}
    </li>
  );
}

export function MetricsSection({ report }: { readonly report: TwinReport }) {
  return (
    <Section title="Twin metrics and tolerances">
      {report.metrics.length === 0 ? (
        <p className={parts.muted}>The board reported no metric the twin could be compared on.</p>
      ) : (
        <ul className={styles.list} aria-label="Metrics and tolerances">
          {report.metrics.map((metric) => (
            <MetricItem key={`${metric.device}.${metric.metric}`} metric={metric} />
          ))}
        </ul>
      )}
    </Section>
  );
}
