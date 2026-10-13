import { useEffect } from "react";
import type { SessionStore } from "../app/sessionStore";
import type { ThemeStore } from "../app/themeStore";
import type { WorkspaceStore } from "../app/workspaceStore";
import { HomeScreen } from "./home/HomeScreen";
import { PairingScreen } from "./pairing/PairingScreen";
import { Shell } from "./shell/Shell";
import { TopBar } from "./shell/TopBar";
import { Unreachable } from "./shell/Unreachable";
import { useStore } from "./useStore";
import { useTheme } from "./useTheme";

interface AppProps {
  readonly themes: ThemeStore;
  readonly session: SessionStore;
  readonly workspace: WorkspaceStore;
}

function useSession(session: SessionStore, workspace: WorkspaceStore) {
  const state = useStore(session);
  useEffect(() => {
    void session.check();
  }, [session]);
  useEffect(() => {
    if (state.phase !== "paired") {
      return undefined;
    }
    void workspace.open();
    return () => {
      workspace.close();
    };
  }, [state.phase, workspace]);
  return state;
}

export function App({ themes, session, workspace }: AppProps) {
  const { preference } = useTheme(themes);
  const state = useSession(session, workspace);
  const topBar = (
    <TopBar
      preference={preference}
      onCycleTheme={() => {
        themes.cycle();
      }}
    />
  );
  if (state.phase === "unpaired") {
    return <PairingScreen session={session} />;
  }
  return (
    <Shell topBar={topBar}>
      {state.phase === "checking" && <p role="status">Connecting to the bridge…</p>}
      {state.phase === "unreachable" && <Unreachable failure={state.failure} />}
      {state.phase === "paired" && <HomeScreen workspace={workspace} />}
    </Shell>
  );
}
