import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import type { ConnectState, LaunchStore } from "../../../app/launch";
import type { SessionDraft } from "../../../model/launch";
import { keepForNextRun, unrecoveredWords } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { panel, useStore } from "../../primitives";
import { SessionForm } from "./SessionForm";
import { SessionElsewhere, SessionEnded, SessionFailed, SessionLive } from "./SessionLive";

interface SessionProps {
  readonly launch: LaunchStore;
  readonly plant: Plant;
  readonly onFinished: (run: string) => void;
}

interface OwnProps extends Omit<SessionProps, "plant"> {
  readonly state: ConnectState;
  readonly form: ReactNode;
  readonly forget: () => void;
}

function OwnSession({ state, launch, onFinished, form, forget }: OwnProps) {
  switch (state.phase) {
    case "starting":
    case "live":
      return <SessionLive state={state} connect={launch.connect} />;
    case "ended":
      return (
        <SessionEnded
          run={state.run}
          onOpen={() => {
            launch.connect.reset();
            onFinished(state.run);
          }}
          onAnother={() => {
            launch.connect.reset();
            forget();
          }}
        />
      );
    case "failed":
      return (
        <SessionFailed
          state={state}
          onBack={() => {
            launch.connect.reset();
          }}
        />
      );
    case "idle":
    case "checking":
    case "checked":
    case "refused":
      return form;
  }
}

function useDraft(launch: LaunchStore, plant: string) {
  const [draft, setDraft] = useState<SessionDraft>(() => launch.connect.draftFor(plant));
  useEffect(() => {
    void launch.connect.prefill(plant).then((filled) => {
      if (filled !== null) setDraft(filled);
    });
  }, [launch, plant]);
  const change = (next: SessionDraft) => {
    setDraft(next);
    launch.connect.keepDraft(plant, next);
  };
  const forget = () => {
    change(keepForNextRun(draft));
  };
  return { draft, change, forget };
}

function useUnrecovered(launch: LaunchStore, plant: string): boolean {
  const [unrecovered] = useState(() => launch.connect.unrecovered(plant));
  useEffect(
    () => () => {
      launch.connect.forgetUnrecovered(plant);
    },
    [launch, plant],
  );
  return unrecovered;
}

export function Session({ launch, plant, onFinished }: SessionProps) {
  const state = useStore(launch.connect);
  const { draft, change, forget } = useDraft(launch, plant.id);
  const unrecovered = useUnrecovered(launch, plant.id);
  useEffect(() => {
    void launch.connect.resume();
  }, [launch]);
  const mine = state.phase === "idle" || state.domain === plant.id ? state : null;
  const form = (
    <SessionForm
      plant={plant}
      draft={draft}
      change={change}
      state={mine ?? { phase: "idle" }}
      connect={launch.connect}
      note={
        unrecovered ? (
          <p className={panel.hint} role="note">
            {unrecoveredWords}
          </p>
        ) : null
      }
    />
  );
  if (mine === null && (state.phase === "live" || state.phase === "starting")) {
    return <SessionElsewhere domain={state.domain} />;
  }
  return (
    <OwnSession state={mine ?? { phase: "idle" }} launch={launch} onFinished={onFinished} form={form} forget={forget} />
  );
}
