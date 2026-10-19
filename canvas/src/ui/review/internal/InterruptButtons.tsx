import { Check, ClipboardCheck, Hand, OctagonX } from "lucide-react";
import type { ControlSession } from "../../../app/review";
import type { ControlScope, ScopeChoice, ScopeStatus } from "../../../model/review";
import { cx, shared } from "../../primitives";
import styles from "./Interrupts.module.css";

export interface Enabled {
  readonly actions: boolean;
  readonly stop: boolean;
}

interface ButtonsProps {
  readonly controls: ControlSession;
  readonly scope: ControlScope;
  readonly status: ScopeStatus;
  readonly enabled: Enabled;
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
      <span className="sr-only">Scope</span>
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

function On({ pressed }: { readonly pressed: boolean }) {
  return pressed ? (
    <span className={styles.mark} aria-hidden="true">
      <Check size={14} />
      On
    </span>
  ) : null;
}

function HoldButton({ controls, scope, status, enabled }: PartProps) {
  return (
    <button
      type="button"
      className={cx(shared.secondary, styles.toggle)}
      aria-pressed={status.held}
      disabled={!enabled.actions}
      onClick={() => {
        void controls.hold(scope);
      }}
    >
      <Hand size={14} aria-hidden="true" />
      Hold
      <On pressed={status.held} />
    </button>
  );
}

function ReleaseButton({ controls, scope, enabled }: PartProps) {
  return (
    <button
      type="button"
      className={shared.secondary}
      disabled={!enabled.actions}
      onClick={() => {
        void controls.release(scope);
      }}
    >
      Release
    </button>
  );
}

function ApproveButton({ controls, scope, status, enabled }: PartProps) {
  const approving = status.approving;
  return (
    <button
      type="button"
      className={cx(shared.secondary, styles.toggle)}
      aria-pressed={approving}
      disabled={!enabled.actions}
      onClick={() => {
        void (approving ? controls.stopApproving(scope) : controls.approveEach(scope));
      }}
    >
      <ClipboardCheck size={14} aria-hidden="true" />
      Approve each command
      <On pressed={approving} />
    </button>
  );
}

function StopButton({ controls, scope, enabled, noteId }: ButtonsProps) {
  return (
    <button
      type="button"
      className={cx(shared.secondary, styles.estop)}
      disabled={!enabled.stop}
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
