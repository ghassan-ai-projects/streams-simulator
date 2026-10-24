import type { ReactNode } from "react";
import { useEffect, useMemo } from "react";
import type { DeviceSession, Draft, Held, Step, TwinDraft } from "../../../app/devices";
import { interfaceNeedsOf, problemsOf, twinNeedsOf } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { stemOf } from "../../../model/devices";
import { cx, failureLine, panel, screen, useStore } from "../../primitives";
import { CheckPanel } from "./CheckPanel";
import { DeviceTitle } from "./DeviceTitle";
import styles from "./Devices.module.css";
import { FormTitle } from "./FormTitle";
import { HeldNotice, RetryButton } from "./HeldNotice";
import { NeedsFirst } from "./NeedsFirst";
import notes from "./Notes.module.css";
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
  const needs = interfaceNeedsOf(device, twin);
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
            <TwinForm
              twin={session}
              draft={twin}
              issues={issues}
              needs={needs}
              reveal={reveal}
              navigation={navigation}
            />
          )}
        </div>
      </section>
      <TwinDetails draft={twin} />
    </div>
  );
}

function Waiting({ tabs, words }: { readonly tabs: ReactNode; readonly words: string }) {
  return (
    <div className={screen.screen}>
      <div className={screen.side}>{tabs}</div>
      <p className={screen.loading} role="status">
        {words}
      </p>
    </div>
  );
}

interface FrameProps {
  readonly tabs: ReactNode;
  readonly name: string;
  readonly bar: ReactNode;
  readonly children: ReactNode;
}

function Frame({ tabs, name, bar, children }: FrameProps) {
  return (
    <div className={screen.screen}>
      <div className={screen.side}>{tabs}</div>
      <section className={screen.center} aria-label={`Behaviour of ${name}`}>
        {bar}
        <div className={styles.formHost}>{children}</div>
      </section>
    </div>
  );
}

interface UnavailableProps extends Omit<FrameProps, "children"> {
  readonly twin: DeviceSession["twin"];
  readonly navigation: NavigationStore;
  readonly failure: string;
  readonly held: readonly Held[];
  readonly lead: string;
  readonly reveal: Reveal;
}

function Unavailable({ twin, navigation, failure, held, lead, reveal, ...frame }: UnavailableProps) {
  return (
    <Frame {...frame}>
      <div className={cx(panel.body, styles.form)}>
        <FormTitle id="unavailable-title" words="The behaviour cannot be opened yet" interim />
        {held.length > 0 ? (
          <HeldNotice held={held} twin={twin} navigation={navigation} lead={lead} reveal={reveal} />
        ) : (
          <>
            <p className={notes.unavailable} role="alert">
              {failure}
            </p>
            <RetryButton twin={twin} />
          </>
        )}
      </div>
    </Frame>
  );
}

interface LackingProps extends Pick<FrameProps, "tabs" | "bar"> {
  readonly device: Draft;
  readonly reveal: Reveal;
}

function Lacking({ device, tabs, bar, reveal }: LackingProps) {
  return (
    <Frame tabs={tabs} bar={bar} name={stemOf(device.path)}>
      <NeedsFirst device={device} reveal={reveal} />
    </Frame>
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
  if (state.phase === "failed") {
    const tag = step === "check" ? "Check" : "Behaviour";
    const bar = (
      <div className={screen.bar}>
        <DeviceTitle path={draft.path} tag={tag} navigation={navigation} />
      </div>
    );
    if (twinNeedsOf(draft).length > 0) return <Lacking device={draft} tabs={tabs} bar={bar} reveal={reveal} />;
    const failure = `The behaviour could not be opened: ${failureLine(state.failure.diagnostic)}`;
    const held = state.held;
    const props = {
      name: stemOf(draft.path),
      tabs,
      bar,
      twin: session,
      navigation,
      failure,
      held,
      lead: draft.path,
      reveal,
    };
    return <Unavailable {...props} />;
  }
  return <Waiting tabs={tabs} words="Opening the behaviour…" />;
}
