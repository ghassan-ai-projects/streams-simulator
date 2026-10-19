import { ClipboardCheck, Hand, OctagonX } from "lucide-react";
import type { ControlSession } from "../../../app/review";
import type { ControlScope, ScopeChoice, ScopeStatus } from "../../../model/review";
import { cx, shared } from "../../primitives";
import styles from "./Interrupts.module.css";

interface ButtonsProps {
  readonly controls: ControlSession;
  readonly scope: ControlScope;
  readonly status: ScopeStatus;
  readonly ready: boolean;
  readonly noteId: string;
}

interface ScopeFieldProps {
  readonly choices: readonly ScopeChoice[];
  readonly picked: string;
  readonly onPick: (key: string) => void;
}

export function ScopeField({ choices, picked, onPick }: ScopeFieldProps) {
  return (
    <label className={styles.field}>
      Scope
      <select
        aria-label="Scope"
        className={styles.select}
        value={picked}
        onChange={(event) => {
          onPick(event.target.value);
        }}
      >
        {choices.map((choice) => (
          <option key={choice.key} value={choice.key}>
            {choice.label}
          </option>
        ))}
      </select>
    </label>
  );
}

type PartProps = Omit<ButtonsProps, "noteId">;

function HoldButton({ controls, scope, status, ready }: PartProps) {
  return (
    <button
      type="button"
      className={cx(shared.secondary, styles.toggle)}
      aria-pressed={status.held}
      disabled={!ready}
      onClick={() => {
        void controls.hold(scope);
      }}
    >
      <Hand size={14} aria-hidden="true" />
      Hold
    </button>
  );
}

function ReleaseButton({ controls, scope, ready }: PartProps) {
  return (
    <button
      type="button"
      className={shared.secondary}
      disabled={!ready}
      onClick={() => {
        void controls.release(scope);
      }}
    >
      Release
    </button>
  );
}

function ApproveButton({ controls, scope, status, ready }: PartProps) {
  const approving = status.approving;
  return (
    <button
      type="button"
      className={cx(shared.secondary, styles.toggle)}
      aria-pressed={approving}
      disabled={!ready}
      onClick={() => {
        void (approving ? controls.stopApproving(scope) : controls.approveEach(scope));
      }}
    >
      <ClipboardCheck size={14} aria-hidden="true" />
      Approve each command
    </button>
  );
}

function StopButton({ controls, scope, ready, noteId }: ButtonsProps) {
  return (
    <button
      type="button"
      className={cx(shared.secondary, styles.estop)}
      disabled={!ready}
      aria-describedby={noteId}
      onClick={() => {
        void controls.emergencyStop(scope);
      }}
    >
      <OctagonX size={14} aria-hidden="true" />
      Emergency stop
    </button>
  );
}

export function InterruptButtons(props: ButtonsProps) {
  return (
    <>
      <div className={styles.group} role="group" aria-label="Hold and approval">
        <HoldButton {...props} />
        <ReleaseButton {...props} />
        <ApproveButton {...props} />
      </div>
      <StopButton {...props} />
    </>
  );
}
