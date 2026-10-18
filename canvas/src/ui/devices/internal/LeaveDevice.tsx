import type { KeyboardEvent, ReactNode } from "react";
import { useState } from "react";
import type { DeviceSession } from "../../../app/devices";
import { hasUnsavedWork, isBusy, isDirty, saveStatusOf, twinStatusOf } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { cx, shared, useFocusOnMount, useStore } from "../../primitives";
import styles from "./Devices.module.css";
import notes from "./Notes.module.css";

interface LeaveProps {
  readonly name: string;
  readonly devices: DeviceSession;
  readonly navigation: NavigationStore;
}

interface ChoiceProps {
  readonly busy: boolean;
  readonly onStay: () => void;
  readonly onDiscard: () => void;
  readonly onSave: () => void;
}

const focusable = "button:not([disabled]), input, select, textarea, [href], [tabindex]:not([tabindex='-1'])";

function trapTab(event: KeyboardEvent<HTMLElement>) {
  if (event.key !== "Tab") return;
  const items = [...event.currentTarget.querySelectorAll<HTMLElement>(focusable)];
  const [first] = items;
  const last = items.at(-1);
  if (first === undefined || last === undefined) return;
  const edge = event.shiftKey ? first : last;
  if (document.activeElement !== edge) return;
  event.preventDefault();
  (event.shiftKey ? last : first).focus();
}

function Choices({ busy, onStay, onDiscard, onSave }: ChoiceProps) {
  const stayButton = useFocusOnMount<HTMLButtonElement>();
  return (
    <div className={styles.leaveActions}>
      <button type="button" ref={stayButton} className={shared.secondary} onClick={onStay}>
        Stay
      </button>
      <button type="button" className={cx(shared.secondary, styles.remove)} disabled={busy} onClick={onDiscard}>
        Discard and leave
      </button>
      <button type="button" className={shared.primary} disabled={busy} onClick={onSave}>
        Save and leave
      </button>
    </div>
  );
}

function Frame({
  name,
  reason,
  onEscape,
  children,
}: {
  readonly name: string;
  readonly reason: string;
  readonly onEscape: () => void;
  readonly children: ReactNode;
}) {
  return (
    <div className={shared.scrim}>
      <dialog
        open
        className={cx(shared.dialog, styles.leaveDialog)}
        aria-modal="true"
        aria-labelledby="leave-device-title"
        aria-describedby="leave-device-text"
        onKeyDown={(event) => {
          if (event.key === "Escape") onEscape();
          trapTab(event);
        }}
      >
        <h2 id="leave-device-title" className={styles.leaveTitle}>
          Leave with unsaved changes?
        </h2>
        <p id="leave-device-text" className={styles.leaveText}>
          {`${name} has changes that are not saved. Save them and leave, discard them, or stay and keep editing.`}
        </p>
        {reason !== "" && (
          <p role="alert" className={notes.reasonLine}>
            {reason}
          </p>
        )}
        {children}
      </dialog>
    </div>
  );
}

function refusalOf(devices: DeviceSession): string {
  const device = devices.snapshot();
  if (device.phase === "ready" && isDirty(device)) return saveStatusOf(device).text;
  const twin = devices.twin.snapshot();
  const failed = twin.phase === "ready" && twin.save.kind === "write-failed";
  return twin.phase === "ready" && failed ? twinStatusOf(twin, device).text : "Something is still not saved.";
}

function Leaving({ name, devices, navigation }: LeaveProps) {
  const [opener] = useState(() => document.activeElement);
  const [reason, setReason] = useState("");
  const busy = isBusy(useStore(devices));
  const stay = () => {
    navigation.stay();
    if (opener instanceof HTMLElement) opener.focus();
  };
  const saveThenLeave = async () => {
    await devices.save();
    await devices.twin.persist();
    if (hasUnsavedWork(devices.snapshot(), devices.twin.snapshot())) setReason(refusalOf(devices));
    else navigation.leave();
  };
  return (
    <Frame name={name} reason={reason} onEscape={stay}>
      <Choices
        busy={busy}
        onStay={stay}
        onDiscard={() => {
          devices.revert();
          devices.twin.discard();
          navigation.leave();
        }}
        onSave={() => {
          void saveThenLeave();
        }}
      />
    </Frame>
  );
}

export function PendingDeviceLeave(props: LeaveProps) {
  const leaving = useStore(props.navigation.leaving);
  return leaving === null ? null : <Leaving {...props} />;
}
