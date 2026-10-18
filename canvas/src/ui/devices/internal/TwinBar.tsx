import { Cpu, Play, RefreshCw, Wrench } from "lucide-react";
import type { Draft, Step, TwinDraft, TwinSession } from "../../../app/devices";
import { builtOf, canCreate, createBlock, twinStatusOf, verdictOf } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { screen, shared } from "../../primitives";
import { ConfirmButton } from "./ConfirmButton";
import { DeviceTitle } from "./DeviceTitle";
import styles from "./Devices.module.css";

interface BarProps {
  readonly draft: TwinDraft;
  readonly device: Draft;
  readonly twin: TwinSession;
  readonly navigation: NavigationStore;
  readonly step: Exclude<Step, "interface">;
}

type PartProps = Pick<BarProps, "draft" | "twin">;

const tones = {
  busy: styles.statusBusy,
  warn: styles.statusWarn,
  ok: styles.statusOk,
  "": styles.statusIdle,
} as const;

function Reload({ draft, twin }: PartProps) {
  if (!(draft.save.kind === "write-failed" && draft.save.failure.kind === "conflict")) return null;
  return (
    <ConfirmButton
      words="Reload from disk"
      icon={<RefreshCw size={14} aria-hidden="true" />}
      title="Read the choices from disk again"
      question="Reload and lose your choices?"
      confirmWords="Reload and lose them"
      needsConfirm={draft.edited}
      onConfirm={() => {
        void twin.reload();
      }}
    />
  );
}

function Replace({ draft, twin }: PartProps) {
  if (draft.build.kind !== "exists") return null;
  return (
    <button
      type="button"
      className={shared.secondary}
      onClick={() => {
        void twin.replace();
      }}
    >
      Replace it
    </button>
  );
}

function OpenPlant({ draft, navigation }: Pick<BarProps, "draft" | "navigation">) {
  const built = builtOf(draft);
  if (built === null || verdictOf(built) === "failed") return null;
  return (
    <button
      type="button"
      className={shared.secondary}
      onClick={() => {
        navigation.go({ screen: "plant", domain: built.path, hall: null });
      }}
    >
      <Play size={14} aria-hidden="true" />
      Open as a plant
    </button>
  );
}

function Create({ draft, device, twin }: Pick<BarProps, "draft" | "device" | "twin">) {
  const allowed = canCreate(draft, device);
  const again = draft.onDisk || builtOf(draft) !== null;
  return (
    <button
      type="button"
      className={shared.primary}
      aria-disabled={!allowed}
      title={
        createBlock(draft, device) ??
        (again ? "Replace the twin with these choices" : "Create the twin from these choices")
      }
      onClick={() => {
        if (allowed) void twin.create();
      }}
    >
      <Wrench size={14} aria-hidden="true" />
      {again ? "Replace twin" : "Create twin"}
    </button>
  );
}

export function TwinBar({ draft, device, twin, navigation, step }: BarProps) {
  const status = twinStatusOf(draft, device);
  return (
    <div className={screen.bar}>
      <Cpu size={14} aria-hidden="true" />
      <DeviceTitle path={device.path} tag={step === "check" ? "Check" : "Behaviour"} navigation={navigation} />
      <span role="status" className={tones[status.tone]} title={status.detail ?? status.text}>
        {status.text}
      </span>
      <Reload draft={draft} twin={twin} />
      <Replace draft={draft} twin={twin} />
      <OpenPlant draft={draft} navigation={navigation} />
      <Create draft={draft} device={device} twin={twin} />
    </div>
  );
}
