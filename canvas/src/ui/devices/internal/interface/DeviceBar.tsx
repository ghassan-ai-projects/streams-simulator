import { Cpu, RefreshCw, RotateCcw, Save, ShieldCheck } from "lucide-react";
import { useState } from "react";
import type { DeviceSession, Draft } from "../../../../app/devices";
import {
  canSave,
  isBusy,
  isDirty,
  rangeNeedsOf,
  saveStatusOf,
  twinReadiness,
  writeConflicted,
} from "../../../../app/devices";
import type { NavigationStore } from "../../../../app/shell";
import { screen, shared, useStore } from "../../../primitives";
import { ConfirmButton } from "../common/ConfirmButton";
import { DeviceTitle } from "../common/DeviceTitle";
import { StatusLine } from "../common/StatusLine";
import notes from "../common/Notes.module.css";
import { useDeviceKeys } from "../common/useDeviceKeys";

interface BarProps {
  readonly draft: Draft;
  readonly devices: DeviceSession;
  readonly navigation: NavigationStore;
}

interface AskProps {
  readonly onAsking: (asking: boolean) => void;
}

const readyTones = {
  busy: notes.readyBusy,
  warn: notes.readyWarn,
  ok: notes.readyOk,
  "": notes.readyIdle,
} as const;

function Reload({ draft, devices }: Pick<BarProps, "draft" | "devices">) {
  if (!writeConflicted(draft)) return null;
  return (
    <ConfirmButton
      words="Reload from disk"
      icon={<RefreshCw size={14} aria-hidden="true" />}
      title="Read the file from disk again"
      question="Reload and lose your changes?"
      confirmWords="Reload and lose them"
      needsConfirm={isDirty(draft)}
      onConfirm={() => {
        void devices.reload();
      }}
    />
  );
}

function Revert({ draft, devices, onAsking }: Pick<BarProps, "draft" | "devices"> & AskProps) {
  return (
    <ConfirmButton
      words="Revert"
      icon={<RotateCcw size={14} aria-hidden="true" />}
      title="Discard the changes since the last save"
      question="Discard your changes?"
      confirmWords="Discard them"
      needsConfirm
      disabled={!isDirty(draft) || isBusy(draft)}
      onAsking={onAsking}
      onConfirm={() => {
        devices.revert();
      }}
    />
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
  const [asking, setAsking] = useState(false);
  const status = saveStatusOf(draft);
  const readiness = twinReadiness(draft, rangeNeedsOf(useStore(devices.twin)).length);
  return (
    <div className={screen.bar}>
      <Cpu size={14} aria-hidden="true" />
      <DeviceTitle path={draft.path} tag="Editing" navigation={navigation} />
      <Reload draft={draft} devices={devices} />
      {!asking && <StatusLine status={status} />}
      {!asking && readiness.text !== "" && (
        <span role="status" className={readyTones[readiness.tone]} title={readiness.text}>
          {readiness.text}
        </span>
      )}
      <Revert draft={draft} devices={devices} onAsking={setAsking} />
      {!asking && <Check draft={draft} devices={devices} />}
      {!asking && <SaveButton draft={draft} devices={devices} />}
    </div>
  );
}
