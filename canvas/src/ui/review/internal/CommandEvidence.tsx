import type { SessionCommand } from "../../../model/review";
import {
  acceptanceDetail,
  acceptanceWords,
  commandSide,
  effectWords,
  reportWords,
  runClock,
} from "../../../model/review";
import type { PolicyActions } from "./PolicyRefusal";
import { PolicyRefusal } from "./PolicyRefusal";
import cells from "./ReportTable.module.css";
import styles from "./SessionEvidence.module.css";

const receiptTones: Readonly<Record<SessionCommand["receipt"], string | undefined>> = {
  pending: styles.warn,
  accepted: styles.chip,
  refused: styles.alarm,
};

interface StageProps {
  readonly label: string;
  readonly tone?: string | undefined;
  readonly words: string;
  readonly detail?: string;
}

function Stage({ label, tone, words, detail = "" }: StageProps) {
  return (
    <li className={styles.stage}>
      <span className={styles.stageName}>{label}</span>
      <span className={tone}>{words}</span>
      {detail !== "" && <span className={styles.detail}>{detail}</span>}
    </li>
  );
}

interface RowProps {
  readonly command: SessionCommand;
  readonly startNs: number;
  readonly policy: PolicyActions;
}

function CommandRow({ command, startNs, policy }: RowProps) {
  return (
    <tr>
      <th scope="row">
        {command.command}
        <p className={styles.detail}>
          <span className={styles.id} title={command.device}>
            {command.device}
          </span>{" "}
          · {commandSide(command)} ·{" "}
          <span className={styles.id} title={command.commandId}>
            {command.commandId}
          </span>
        </p>
      </th>
      <td>
        <ul className={styles.stages} aria-label={`Stages of ${command.commandId}`}>
          <Stage label="Sent at" tone={cells.clock} words={runClock(command.sentAtNs, startNs)} />
          <Stage
            label="Acceptance"
            tone={receiptTones[command.receipt]}
            words={acceptanceWords(command)}
            detail={acceptanceDetail(command)}
          />
          {command.policyDigest !== "" && <PolicyRefusal command={command} {...policy} />}
          <Stage label="Board report" words={reportWords[command.report]} />
          <Stage
            label="Effect evidence"
            tone={command.effect === "observed" ? styles.ok : styles.chip}
            words={effectWords(command)}
            detail={command.effectEvidence}
          />
        </ul>
      </td>
    </tr>
  );
}

interface EvidenceProps {
  readonly commands: readonly SessionCommand[];
  readonly startNs: number;
  readonly policy: PolicyActions;
}

export function CommandEvidence({ commands, startNs, policy }: EvidenceProps) {
  return (
    <div className={cells.scroll} tabIndex={0} role="region" aria-label="Command evidence, scrollable">
      <table className={cells.table}>
        <caption>Command evidence</caption>
        <thead>
          <tr>
            {["Command", "Stages"].map((label) => (
              <th key={label} scope="col">
                {label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {commands.map((command) => (
            <CommandRow key={command.commandId} command={command} startNs={startNs} policy={policy} />
          ))}
        </tbody>
      </table>
    </div>
  );
}
