import type { ReactNode } from "react";
import { useEffect, useMemo, useRef } from "react";
import type { DeviceSession, Draft } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { draftProblems, stemOf } from "../../../model/devices";
import { screen } from "../../primitives";
import { CommandForm } from "./CommandForm";
import { DeviceBar } from "./DeviceBar";
import { DeviceDetails } from "./DeviceDetails";
import { DeviceOutline } from "./DeviceOutline";
import styles from "./Devices.module.css";
import { KindForm } from "./KindForm";
import { MetricForm } from "./MetricForm";
import type { FormProps } from "./formProps";

interface InterfaceProps {
  readonly draft: Draft;
  readonly devices: DeviceSession;
  readonly navigation: NavigationStore;
  readonly tabs: ReactNode;
  readonly focus: string | null;
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

function useFocusField(label: string | null) {
  const host = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (label !== null) host.current?.querySelector<HTMLInputElement>(`input[aria-label="${label}"]`)?.focus();
  }, [label]);
  return host;
}

export function InterfaceStep({ draft, devices, navigation, tabs, focus }: InterfaceProps) {
  const issues = useMemo(() => [...draftProblems(draft.doc), ...draft.issues], [draft.doc, draft.issues]);
  const host = useFocusField(focus);
  return (
    <div className={screen.screen}>
      <div className={screen.side}>
        {tabs}
        <DeviceOutline draft={draft} devices={devices} issues={issues} />
      </div>
      <section className={screen.center} aria-label={`Editing ${stemOf(draft.path)}`}>
        <DeviceBar draft={draft} devices={devices} navigation={navigation} />
        <div ref={host} className={styles.formHost}>
          <ItemForm draft={draft} devices={devices} issues={issues} />
        </div>
      </section>
      <DeviceDetails draft={draft} devices={devices} issues={issues} />
    </div>
  );
}
