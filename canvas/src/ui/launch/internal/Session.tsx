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
}

function OwnSession({ state, launch, onFinished, form }: OwnProps) {
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

export function Session({ launch, plant, onFinished }: SessionProps) {
  const state = useStore(launch.connect);
  const [draft, setDraft] = useState<SessionDraft>(firstSessionDraft);
  useEffect(() => {
    void launch.connect.resume(plant.id);
  }, [launch, plant.id]);
  const form = <SessionForm plant={plant} draft={draft} change={setDraft} state={state} connect={launch.connect} />;
  if (state.phase === "idle") return form;
  if (state.domain !== plant.id) {
    return state.phase === "live" || state.phase === "starting" ? <SessionElsewhere domain={state.domain} /> : form;
  }
  return <OwnSession state={state} launch={launch} onFinished={onFinished} form={form} />;
}
