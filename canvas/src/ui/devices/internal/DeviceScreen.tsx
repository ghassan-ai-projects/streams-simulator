import { useEffect, useMemo } from "react";
import type { DeviceSession, Draft } from "../../../app/devices";
import { isDirty } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { draftProblems, stemOf } from "../../../model/devices";
import { failureLine, screen, useStore } from "../../primitives";
import { CommandForm } from "./CommandForm";
import { DeviceBar } from "./DeviceBar";
import { DeviceDetails } from "./DeviceDetails";
import { DeviceOutline } from "./DeviceOutline";
import styles from "./Devices.module.css";
import { KindForm } from "./KindForm";
import { PendingDeviceLeave } from "./LeaveDevice";
import { MetricForm } from "./MetricForm";
import type { FormProps } from "./formProps";

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

function ItemForm({ draft, devices, issues }: FormProps & { readonly draft: Draft }) {
  const { doc, selected } = draft;
  const metric = selected.part === "metric" ? doc.metrics.find((entry) => entry.key === selected.key) : undefined;
  const command = selected.part === "command" ? doc.commands.find((entry) => entry.key === selected.key) : undefined;
  if (metric !== undefined) return <MetricForm key={metric.key} metric={metric} devices={devices} issues={issues} />;
  if (command !== undefined)
    return <CommandForm key={command.key} command={command} devices={devices} issues={issues} />;
  return <KindForm kind={doc} devices={devices} issues={issues} />;
}

function Editing({ draft, devices, navigation }: Omit<ScreenProps, "path"> & { readonly draft: Draft }) {
  useLeaveWarning(isDirty(draft));
  const issues = useMemo(() => [...draftProblems(draft.doc), ...draft.issues], [draft.doc, draft.issues]);
  return (
    <>
      <div className={screen.screen}>
        <div className={screen.side}>
          <DeviceOutline draft={draft} devices={devices} issues={issues} />
        </div>
        <section className={screen.center} aria-label={`Editing ${stemOf(draft.path)}`}>
          <DeviceBar draft={draft} devices={devices} navigation={navigation} />
          <div className={styles.formHost}>
            <ItemForm draft={draft} devices={devices} issues={issues} />
          </div>
        </section>
        <DeviceDetails draft={draft} devices={devices} issues={issues} />
      </div>
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
