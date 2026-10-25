import type { ReactNode } from "react";
import { useEffect } from "react";
import type { DeviceSession, Draft } from "../../../app/devices";
import { hasUnsavedWork, stepStatuses } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { stemOf } from "../../../model/devices";
import { failureLine, screen, useStore } from "../../primitives";
import { InterfaceStep } from "./InterfaceStep";
import { PendingDeviceLeave } from "./LeaveDevice";
import { StepTabs } from "./StepTabs";
import { TwinStep } from "./TwinStep";
import { StepFocusContext } from "./stepFocus";
import { useStepNavigation } from "./useStepNavigation";

interface ScreenProps {
  readonly devices: DeviceSession;
  readonly navigation: NavigationStore;
  readonly path: string;
  readonly step?: "behaviour";
  readonly assumptions: (domain: string, twin: string) => ReactNode;
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

function Editing({
  draft,
  devices,
  navigation,
  step: opening,
  assumptions,
}: Omit<ScreenProps, "path"> & { readonly draft: Draft }) {
  useLeaveWarning(devices);
  const { step, focus, stepFocus, pick, reveal } = useStepNavigation(devices, opening);
  const statuses = stepStatuses(draft, useStore(devices.twin));
  const tabs = <StepTabs step={step} statuses={statuses} onPick={pick} />;
  return (
    <StepFocusContext value={stepFocus}>
      {step === "interface" ? (
        <InterfaceStep draft={draft} devices={devices} navigation={navigation} tabs={tabs} focus={focus} />
      ) : (
        <TwinStep
          draft={draft}
          devices={devices}
          navigation={navigation}
          tabs={tabs}
          step={step}
          reveal={reveal}
          assumptions={assumptions}
        />
      )}
      <PendingDeviceLeave name={stemOf(draft.path)} devices={devices} navigation={navigation} />
    </StepFocusContext>
  );
}

export function DeviceScreen({ devices, navigation, path, step, assumptions }: ScreenProps) {
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
    return (
      <Editing
        draft={state}
        devices={devices}
        navigation={navigation}
        assumptions={assumptions}
        {...(step === undefined ? {} : { step })}
      />
    );
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
