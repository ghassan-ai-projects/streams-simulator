import { useCallback, useEffect, useMemo, useState } from "react";
import type { DeviceSession, Draft, Step } from "../../../app/devices";
import { hasUnsavedWork, stepStatuses } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { stemOf } from "../../../model/devices";
import type { Selection } from "../../../model/devices";
import { failureLine, screen, useStore } from "../../primitives";
import { InterfaceStep } from "./InterfaceStep";
import { PendingDeviceLeave } from "./LeaveDevice";
import { StepTabs } from "./StepTabs";
import { TwinStep } from "./TwinStep";
import { StepFocusContext } from "./stepFocus";
import type { Pending } from "./stepFocus";

interface ScreenProps {
  readonly devices: DeviceSession;
  readonly navigation: NavigationStore;
  readonly path: string;
}

function useLeaveWarning(devices: DeviceSession) {
  useEffect(() => {
    const warn = (event: BeforeUnloadEvent) => {
      if (hasUnsavedWork(devices.snapshot(), devices.twin.snapshot())) event.preventDefault();
    };
    window.addEventListener("beforeunload", warn);
    return () => {
      window.removeEventListener("beforeunload", warn);
    };
  }, [devices]);
}

function Editing({ draft, devices, navigation }: Omit<ScreenProps, "path"> & { readonly draft: Draft }) {
  useLeaveWarning(devices);
  const [step, setStep] = useState<Step>("interface");
  const [focus, setFocus] = useState<{ readonly field: string; readonly at: number } | null>(null);
  const [pending, setPending] = useState<Pending>("fresh");
  const done = useCallback((next: Pending) => {
    setPending(next);
  }, []);
  const statuses = stepStatuses(draft, useStore(devices.twin));
  const pick = (next: Step) => {
    setFocus(null);
    setPending("fresh");
    setStep(next);
  };
  const reveal = (selection: Selection, field: string) => {
    devices.select(selection);
    setFocus((previous) => ({ field, at: (previous?.at ?? 0) + 1 }));
    setPending("none");
    setStep("interface");
  };
  const tabs = <StepTabs step={step} statuses={statuses} onPick={pick} />;
  const stepFocus = useMemo(() => ({ pending, done }), [pending, done]);
  return (
    <StepFocusContext value={stepFocus}>
      {step === "interface" ? (
        <InterfaceStep draft={draft} devices={devices} navigation={navigation} tabs={tabs} focus={focus} />
      ) : (
        <TwinStep draft={draft} devices={devices} navigation={navigation} tabs={tabs} step={step} reveal={reveal} />
      )}
      <PendingDeviceLeave name={stemOf(draft.path)} devices={devices} navigation={navigation} />
    </StepFocusContext>
  );
}

export function DeviceScreen({ devices, navigation, path }: ScreenProps) {
  const state = useStore(devices);
  useEffect(() => {
    const disposeGuard = navigation.guardUnsaved(() => hasUnsavedWork(devices.snapshot(), devices.twin.snapshot()));
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
