import type { ReactNode } from "react";
import { useEffect, useMemo } from "react";
import type { DeviceSession, Draft, Step, TwinDraft } from "../../../app/devices";
import { problemsOf, twinNeedsOf } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { stemOf } from "../../../model/devices";
import { failureLine, screen, useStore } from "../../primitives";
import { CheckPanel } from "./CheckPanel";
import styles from "./Devices.module.css";
import { NeedsFirst } from "./NeedsFirst";
import { TwinBar } from "./TwinBar";
import { TwinDetails } from "./TwinDetails";
import { TwinForm } from "./TwinForm";
import { TwinSide } from "./TwinSide";
import type { Reveal } from "./twinProps";

interface StepProps {
  readonly draft: Draft;
  readonly devices: DeviceSession;
  readonly navigation: NavigationStore;
  readonly tabs: ReactNode;
  readonly step: Exclude<Step, "interface">;
  readonly reveal: Reveal;
}

type ReadyProps = Omit<StepProps, "draft"> & { readonly device: Draft; readonly twin: TwinDraft };

function Ready({ device, devices, navigation, tabs, step, twin, reveal }: ReadyProps) {
  const session = devices.twin;
  const issues = useMemo(() => [...problemsOf(twin), ...twin.issues], [twin]);
  const needs = twinNeedsOf(device);
  const name = stemOf(device.path);
  return (
    <div className={screen.screen}>
      <div className={screen.side}>
        {tabs}
        {step === "behaviour" && <TwinSide twin={session} draft={twin} />}
      </div>
      <section className={screen.center} aria-label={`${step === "check" ? "Check" : "Behaviour"} of ${name}`}>
        <TwinBar draft={twin} device={device} twin={session} navigation={navigation} step={step} />
        <div className={styles.formHost}>
          {step === "check" ? (
            <CheckPanel draft={twin} device={device} />
          ) : (
            <TwinForm twin={session} draft={twin} issues={issues} needs={needs} reveal={reveal} />
          )}
        </div>
      </section>
      <TwinDetails draft={twin} />
    </div>
  );
}

function Waiting({
  tabs,
  failed,
  words,
}: {
  readonly tabs: ReactNode;
  readonly failed: boolean;
  readonly words: string;
}) {
  return (
    <div className={screen.screen}>
      <div className={screen.side}>{tabs}</div>
      <p className={failed ? screen.failure : screen.loading} role={failed ? "alert" : "status"}>
        {words}
      </p>
    </div>
  );
}

function Lacking({
  device,
  tabs,
  reveal,
}: {
  readonly device: Draft;
  readonly tabs: ReactNode;
  readonly reveal: Reveal;
}) {
  return (
    <div className={screen.screen}>
      <div className={screen.side}>{tabs}</div>
      <section className={screen.center} aria-label={`Behaviour of ${stemOf(device.path)}`}>
        <div className={styles.formHost}>
          <NeedsFirst device={device} reveal={reveal} />
        </div>
      </section>
    </div>
  );
}

export function TwinStep({ draft, devices, navigation, tabs, step, reveal }: StepProps) {
  const session = devices.twin;
  const state = useStore(session);
  useEffect(() => {
    void session.open(draft.path);
    return () => {
      void session.persist();
    };
  }, [session, draft.path]);
  if (state.phase === "ready" && state.path === draft.path) {
    const props = { device: draft, devices, navigation, tabs, step, twin: state, reveal };
    return <Ready {...props} />;
  }
  if (state.phase === "failed" && twinNeedsOf(draft).length > 0) {
    return <Lacking device={draft} tabs={tabs} reveal={reveal} />;
  }
  if (state.phase === "failed") {
    const words = `The behaviour could not be opened: ${failureLine(state.failure.diagnostic)}`;
    return <Waiting tabs={tabs} failed words={words} />;
  }
  return <Waiting tabs={tabs} failed={false} words="Opening the behaviour…" />;
}
