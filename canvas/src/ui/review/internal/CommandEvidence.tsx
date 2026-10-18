import type { SessionCommand } from "../../../model/review";
import { commandSide, effectWords, receiptWords, refusalWords, reportWords, runClock } from "../../../model/review";
import cells from "./ReportTable.module.css";
import styles from "./SessionEvidence.module.css";

const receiptTones: Readonly<Record<SessionCommand["receipt"], string | undefined>> = {
  pending: styles.warn,
  accepted: styles.chip,
  refused: styles.alarm,
};

function Stage({ tone, words }: { readonly tone: string | undefined; readonly words: string }) {
  return <span className={tone}>{words}</span>;
}

function CommandRow({ command, startNs }: { readonly command: SessionCommand; readonly startNs: number }) {
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
      <td className={cells.clock}>{runClock(command.sentAtNs, startNs)}</td>
      <td>
        <Stage tone={receiptTones[command.receipt]} words={receiptWords[command.receipt]} />
        {command.reason !== "" && <p className={styles.detail}>{refusalWords(command.reason)}</p>}
      </td>
      <td>{reportWords[command.report]}</td>
      <td>
        <Stage tone={command.effect === "observed" ? styles.ok : styles.chip} words={effectWords(command)} />
        {command.effectEvidence !== "" && <p className={styles.detail}>{command.effectEvidence}</p>}
      </td>
    </tr>
  );
}

export function CommandEvidence({
  commands,
  startNs,
}: {
  readonly commands: readonly SessionCommand[];
  readonly startNs: number;
}) {
  return (
    <table className={cells.table}>
      <caption>Command evidence</caption>
      <thead>
        <tr>
          {["Command", "Sent at", "Acceptance", "Board report", "Effect evidence"].map((label) => (
            <th key={label} scope="col">
              {label}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {commands.map((command) => (
          <CommandRow key={command.commandId} command={command} startNs={startNs} />
        ))}
      </tbody>
    </table>
  );
}
