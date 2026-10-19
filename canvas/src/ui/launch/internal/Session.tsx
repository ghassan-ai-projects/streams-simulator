import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import type { ConnectState, LaunchStore } from "../../../app/launch";
import type { SessionDraft } from "../../../model/launch";
import { firstSessionDraft } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { useStore } from "../../primitives";
import { SessionForm } from "./SessionForm";
import { SessionElsewhere, SessionEnded, SessionFailed, SessionLive } from "./SessionLive";

interface SessionProps {
  readonly launch: LaunchStore;
  readonly plant: Plant;
  readonly onFinished: (run: string) => void;
}

interface OwnProps extends Omit<SessionProps, "plant"> {
  readonly state: Exclude<ConnectState, { phase: "idle" }>;
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
    change({ ...firstSessionDraft, policies: draft.policies, policiesFrom: draft.policiesFrom });
  };
  return { draft, change, forget };
}

export function Session({ launch, plant, onFinished }: SessionProps) {
  const state = useStore(launch.connect);
  const { draft, change, forget } = useDraft(launch, plant.id);
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
    />
  );
  if (mine === null) {
    return state.phase === "live" || state.phase === "starting" ? <SessionElsewhere domain={state.domain} /> : form;
  }
  return mine.phase === "idle" ? (
    form
  ) : (
    <OwnSession state={mine} launch={launch} onFinished={onFinished} form={form} forget={forget} />
  );
}
