import { Cable, CircleCheck, CircleSlash, CircleX } from "lucide-react";
import type { ConnectSession } from "../../../app/launch";
import type { ReviewSession, ReviewState } from "../../../app/review";
import type { SessionDevice, SessionLive } from "../../../model/review";
import { deviceLine, linkWords, recordedPolicyWords, runClock } from "../../../model/review";
import { CommandEvidence } from "./CommandEvidence";
import { Section } from "./ReportParts";
import parts from "./ReportParts.module.css";
import styles from "./SessionEvidence.module.css";

type Link = SessionLive["host"];

const linkTones: Readonly<Record<Link, string | undefined>> = {
  online: styles.ok,
  offline: styles.alarm,
  unknown: styles.chip,
  not_applicable: styles.chip,
};

const linkIcons = { online: CircleCheck, offline: CircleX, unknown: Cable, not_applicable: CircleSlash } as const;

function LinkChip({ name, state }: { readonly name: string; readonly state: Link }) {
  const Icon = linkIcons[state];
  return (
    <span className={linkTones[state]}>
      <Icon size={13} aria-hidden="true" />
      {`${name}: ${linkWords[state]}`}
    </span>
  );
}

function DeviceRow({ device }: { readonly device: SessionDevice }) {
  const name = device.instance || device.device;
  return (
    <li className={styles.device}>
      <span className={styles.id} title={name}>
        {name}
      </span>
      <span className={device.health === "offline" ? styles.alarm : styles.chip}>{deviceLine(device)}</span>
    </li>
  );
}

function ConnectionHealth({ state, startNs }: { readonly state: ReviewState; readonly startNs: number }) {
  const { live } = state;
  if (state.connection === "running" && !state.caughtUp)
    return <p className={parts.muted}>Catching up to live: reading the session so far…</p>;
  if (live === null) return <p className={parts.muted}>No connection health recorded at this time.</p>;
  const note = state.connection === "saved" ? `as of ${runClock(live.atNs, startNs)}` : undefined;
  return (
    <Section title="Connection health" note={note}>
      <div className={styles.links}>
        <LinkChip name="Host" state={live.host} />
        <LinkChip name="Gateway" state={live.gateway} />
      </div>
      {live.devices.length === 0 ? (
        <p className={parts.muted}>No devices have announced themselves yet.</p>
      ) : (
        <ul className={styles.devices} aria-label="Devices">
          {live.devices.map((device) => (
            <DeviceRow key={device.device} device={device} />
          ))}
        </ul>
      )}
    </Section>
  );
}

interface EvidenceProps {
  readonly review: ReviewSession;
  readonly connect: ConnectSession;
  readonly state: ReviewState;
  readonly startNs: number;
}

function Commands({ review, connect, state, startNs }: EvidenceProps) {
  const commands = state.live?.commands ?? [];
  const policy = { review, connect, domain: state.sessionDomain };
  return (
    <Section title="Commands">
      {state.live?.commandsTruncated === true && (
        <p role="note">Showing the latest commands only. Earlier command records remain in the saved journal.</p>
      )}
      <p className={parts.muted}>{recordedPolicyWords(state.sessionPolicies)}</p>
      <p className={parts.muted}>
        Acceptance and the board&apos;s report do not prove an effect. Host observations are separate evidence.
      </p>
      {commands.length === 0 ? (
        <p className={parts.muted}>No commands recorded at this time.</p>
      ) : (
        <CommandEvidence commands={commands} startNs={startNs} policy={policy} />
      )}
    </Section>
  );
}

export function SessionEvidence({ review, connect, state, startNs }: EvidenceProps) {
  return (
    <>
      <ConnectionHealth state={state} startNs={startNs} />
      <Commands review={review} connect={connect} state={state} startNs={startNs} />
      {state.connection === "running" && (
        <p className={parts.muted}>Recording live. File verification and replay are available after Stop and save.</p>
      )}
    </>
  );
}
