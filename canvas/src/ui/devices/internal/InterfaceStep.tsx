import type { ReactNode } from "react";
import { useMemo } from "react";
import type { DeviceSession, Draft } from "../../../app/devices";
import { interfaceIssues, twinNeedsOf } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { stemOf } from "../../../model/devices";
import { screen } from "../../primitives";
import { CommandForm } from "./CommandForm";
import { DeviceBar } from "./DeviceBar";
import { DeviceDetails } from "./DeviceDetails";
import { DeviceOutline } from "./DeviceOutline";
import styles from "./Devices.module.css";
import { KindForm } from "./KindForm";
import { MetricForm } from "./MetricForm";
import type { FormProps } from "./formProps";
import { useFieldFocus } from "./stepFocus";

interface InterfaceProps {
  readonly draft: Draft;
  readonly devices: DeviceSession;
  readonly navigation: NavigationStore;
  readonly tabs: ReactNode;
  readonly focus: { readonly field: string; readonly at: number } | null;
}

function ItemForm({ draft, devices, issues, needs }: FormProps & { readonly draft: Draft }) {
  const { doc, selected } = draft;
  const metric = selected.part === "metric" ? doc.metrics.find((entry) => entry.key === selected.key) : undefined;
  const command = selected.part === "command" ? doc.commands.find((entry) => entry.key === selected.key) : undefined;
  const props = { devices, issues, needs };
  if (metric !== undefined) return <MetricForm key={metric.key} metric={metric} {...props} />;
  if (command !== undefined) return <CommandForm key={command.key} command={command} {...props} />;
  return <KindForm kind={doc} {...props} />;
}

export function InterfaceStep({ draft, devices, navigation, tabs, focus }: InterfaceProps) {
  const issues = useMemo(() => interfaceIssues(draft), [draft]);
  const needs = twinNeedsOf(draft);
  const host = useFieldFocus(focus);
  return (
    <div className={screen.screen}>
      <div className={screen.side}>
        {tabs}
        <DeviceOutline draft={draft} devices={devices} issues={issues} needs={needs} />
      </div>
      <section className={screen.center} aria-label={`Editing ${stemOf(draft.path)}`}>
        <DeviceBar draft={draft} devices={devices} navigation={navigation} />
        <div ref={host} className={styles.formHost}>
          <ItemForm draft={draft} devices={devices} issues={issues} needs={needs} />
        </div>
      </section>
      <DeviceDetails draft={draft} devices={devices} issues={issues} needs={needs} />
    </div>
  );
}
