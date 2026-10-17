import { Cpu, RefreshCw, RotateCcw, Save, ShieldCheck } from "lucide-react";
import type { DeviceSession, Draft } from "../../../app/devices";
import { canSave, isBusy, isDirty, saveStatusOf } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { screen, shared } from "../../primitives";
import { DeviceTitle } from "./DeviceTitle";
import styles from "./Devices.module.css";
import { useDeviceKeys } from "./useDeviceKeys";

interface BarProps {
  readonly draft: Draft;
  readonly devices: DeviceSession;
  readonly navigation: NavigationStore;
}

const tones = {
  busy: styles.statusBusy,
  warn: styles.statusWarn,
  ok: styles.statusOk,
  "": styles.statusIdle,
} as const;

function Reload({ draft, devices }: Pick<BarProps, "draft" | "devices">) {
  if (!(draft.save.kind === "write-failed" && draft.save.failure.kind === "conflict")) return null;
  return (
    <button
      type="button"
      className={shared.secondary}
      onClick={() => {
        void devices.reload();
      }}
    >
      <RefreshCw size={14} aria-hidden="true" />
      Reload from disk
    </button>
  );
}

function Revert({ draft, devices }: Pick<BarProps, "draft" | "devices">) {
  return (
    <button
      type="button"
      className={shared.secondary}
      aria-disabled={!isDirty(draft) || isBusy(draft)}
      title="Discard the changes since the last save"
      onClick={() => {
        devices.revert();
      }}
    >
      <RotateCcw size={14} aria-hidden="true" />
      Revert
    </button>
  );
}

function Check({ draft, devices }: Pick<BarProps, "draft" | "devices">) {
  return (
    <button
      type="button"
      className={shared.secondary}
      aria-disabled={isBusy(draft)}
      title="Ask the binary to check the saved file again"
      onClick={() => {
        void devices.check();
      }}
    >
      <ShieldCheck size={14} aria-hidden="true" />
      Check again
    </button>
  );
}

function SaveButton({ draft, devices }: Pick<BarProps, "draft" | "devices">) {
  const savable = canSave(draft);
  return (
    <button
      type="button"
      className={shared.primary}
      aria-disabled={!savable}
      title={savable ? "Save and check (Ctrl+S)" : "Nothing to save"}
      onClick={() => {
        if (savable) void devices.save();
      }}
    >
      <Save size={14} aria-hidden="true" />
      Save
    </button>
  );
}

export function DeviceBar({ draft, devices, navigation }: BarProps) {
  useDeviceKeys(devices);
  const status = saveStatusOf(draft);
  return (
    <div className={screen.bar}>
      <Cpu size={14} aria-hidden="true" />
      <DeviceTitle path={draft.path} tag="Editing" navigation={navigation} />
      <Reload draft={draft} devices={devices} />
      <span role="status" className={tones[status.tone]} title={status.detail ?? status.text}>
        {status.text}
      </span>
      <Revert draft={draft} devices={devices} />
      <Check draft={draft} devices={devices} />
      <SaveButton draft={draft} devices={devices} />
    </div>
  );
}
