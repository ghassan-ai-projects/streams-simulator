import type { ReactNode } from "react";
import { useState } from "react";
import type { DeviceSession } from "../../../app/devices";
import { isBusy, isDirty } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { cx, shared, useFocusOnMount, useStore } from "../../primitives";
import styles from "./Devices.module.css";

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
      <button type="button" className={shared.primary} onClick={onSave}>
        Save and leave
      </button>
    </div>
  );
}

function Frame({
  name,
  onEscape,
  children,
}: {
  readonly name: string;
  readonly onEscape: () => void;
  readonly children: ReactNode;
}) {
  return (
    <div className={shared.scrim}>
      <dialog
        open
        className={cx(shared.dialog, styles.leaveDialog)}
        aria-modal="false"
        aria-labelledby="leave-device-title"
        aria-describedby="leave-device-text"
        onKeyDown={(event) => {
          if (event.key === "Escape") onEscape();
        }}
      >
        <h2 id="leave-device-title" className={styles.leaveTitle}>
          Leave with unsaved changes?
        </h2>
        <p id="leave-device-text" className={styles.leaveText}>
          {`${name} has changes that are not saved. Save them and leave, discard them, or stay and keep editing.`}
        </p>
        {children}
      </dialog>
    </div>
  );
}

function Leaving({ name, devices, navigation }: LeaveProps) {
  const [opener] = useState(() => document.activeElement);
  const busy = isBusy(useStore(devices));
  const stay = () => {
    navigation.stay();
    if (opener instanceof HTMLElement) opener.focus();
  };
  const saveThenLeave = async () => {
    await devices.save();
    if (isDirty(devices.snapshot())) stay();
    else navigation.leave();
  };
  return (
    <Frame name={name} onEscape={stay}>
      <Choices
        busy={busy}
        onStay={stay}
        onDiscard={() => {
          devices.revert();
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
