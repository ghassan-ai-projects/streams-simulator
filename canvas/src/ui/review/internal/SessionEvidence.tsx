import type { ReviewState } from "../../../app/review";
import type { SessionCommand, SessionLive } from "../../../model/review";
import { Section } from "./ReportParts";
import parts from "./ReportParts.module.css";
import cells from "./ReportTable.module.css";

const linkWords = { online: "Online", offline: "Offline", unknown: "Unknown", not_applicable: "No gateway" };
const receiptWords = { pending: "Unconfirmed", accepted: "Accepted", refused: "Refused" };
const reportWords = {
  pending: "No board report",
  succeeded: "Board reports success",
  failed: "Board reports failure",
  unknown: "Unknown",
};

function CommandRows({ commands }: { readonly commands: readonly SessionCommand[] }) {
  return (
    <table className={cells.table}>
      <caption>Command evidence</caption>
      <thead>
        <tr>
          {["Sent", "Acceptance", "Board report", "Effect evidence"].map((label) => (
            <th key={label} scope="col">
              {label}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {commands.map((command) => (
          <tr key={command.commandId}>
            <th scope="row">
              {command.command}
              <br />
              <small>
                {command.device} · {command.commandId}
              </small>
            </th>
            <td>
              {receiptWords[command.receipt]}
              {command.reason !== "" && <p>{command.reason}</p>}
            </td>
            <td>{reportWords[command.report]}</td>
            <td>
              {command.effect === "observed" ? "Host observed" : "Effect unknown"}
              {command.effectEvidence !== "" && <p>{command.effectEvidence}</p>}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function ConnectionHealth({ live }: { readonly live: SessionLive | null }) {
  if (live === null) return <p className={parts.muted}>Waiting for connection health from the session…</p>;
  return (
    <Section title="Connection health">
      <p>
        Host: {linkWords[live.host]} · Gateway: {linkWords[live.gateway]}
      </p>
      {live.devices.length === 0 ? (
        <p className={parts.muted}>No devices have announced themselves yet.</p>
      ) : (
        <ul>
          {live.devices.map((device) => (
            <li key={device.device}>
              {device.instance || device.device} · {linkWords[device.health]} · board {device.board} · boot{" "}
              {device.bootId}
              {device.reboots > 0 && ` · ${String(device.reboots)} reboot(s)`}
            </li>
          ))}
        </ul>
      )}
    </Section>
  );
}

export function SessionEvidence({ state }: { readonly state: ReviewState }) {
  return (
    <>
      <ConnectionHealth live={state.live} />
      <Section title="Commands">
        {state.live?.commandsTruncated === true && (
          <p role="note">Showing the latest 128 commands. Earlier command records remain in the saved journal.</p>
        )}
        <p className={parts.muted}>
          Acceptance and the board&apos;s report do not prove an effect. Host observations are separate evidence.
        </p>
        {state.live === null || state.live.commands.length === 0 ? (
          <p className={parts.muted}>No commands recorded at this time.</p>
        ) : (
          <CommandRows commands={state.live.commands} />
        )}
      </Section>
      {state.connection === "running" && (
        <p className={parts.muted}>Recording live. File verification and replay are available after Stop and save.</p>
      )}
    </>
  );
}
