import { useEffect, useState } from "react";
import type { DeviceSession, Draft, Step } from "../../../app/devices";
import { isDirty, stepStatuses } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { interfaceLabel, stemOf } from "../../../model/devices";
import type { Selection } from "../../../model/devices";
import { failureLine, screen, useStore } from "../../primitives";
import { InterfaceStep } from "./InterfaceStep";
import { PendingDeviceLeave } from "./LeaveDevice";
import { StepTabs } from "./StepTabs";
import { TwinStep } from "./TwinStep";

interface ScreenProps {
  readonly devices: DeviceSession;
  readonly navigation: NavigationStore;
  readonly path: string;
}

function useLeaveWarning(dirty: boolean) {
  useEffect(() => {
    const warn = (event: BeforeUnloadEvent) => {
      if (dirty) event.preventDefault();
    };
    window.addEventListener("beforeunload", warn);
    return () => {
      window.removeEventListener("beforeunload", warn);
    };
  }, [dirty]);
}

function Editing({ draft, devices, navigation }: Omit<ScreenProps, "path"> & { readonly draft: Draft }) {
  useLeaveWarning(isDirty(draft));
  const [step, setStep] = useState<Step>("interface");
  const [focus, setFocus] = useState<string | null>(null);
  const statuses = stepStatuses(draft, useStore(devices.twin));
  const pick = (next: Step) => {
    setFocus(null);
    setStep(next);
  };
  const reveal = (selection: Selection, field: string) => {
    devices.select(selection);
    setFocus(interfaceLabel(field));
    setStep("interface");
  };
  const tabs = <StepTabs step={step} statuses={statuses} onPick={pick} />;
  return (
    <>
      {step === "interface" ? (
        <InterfaceStep draft={draft} devices={devices} navigation={navigation} tabs={tabs} focus={focus} />
      ) : (
        <TwinStep draft={draft} devices={devices} navigation={navigation} tabs={tabs} step={step} reveal={reveal} />
      )}
      <PendingDeviceLeave name={stemOf(draft.path)} devices={devices} navigation={navigation} />
    </>
  );
}

export function DeviceScreen({ devices, navigation, path }: ScreenProps) {
  const state = useStore(devices);
  useEffect(() => {
    const disposeGuard = navigation.guardUnsaved(() => isDirty(devices.snapshot()));
    void devices.open(path);
    return () => {
      disposeGuard();
      devices.close();
    };
  }, [devices, navigation, path]);
  if (state.phase === "ready" && state.path === path) {
    return <Editing draft={state} devices={devices} navigation={navigation} />;
  }
  const failed = state.phase === "failed" && state.path === path;
  return (
    <p className={failed ? screen.failure : screen.loading} role={failed ? "alert" : "status"}>
      {state.phase === "failed" && failed
        ? `The device could not be opened: ${failureLine(state.failure.diagnostic)}`
        : "Opening the device…"}
    </p>
  );
}
