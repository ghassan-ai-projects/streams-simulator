import { useEffect } from "react";
import type { NavigationStore } from "../app/navigationStore";
import type { PlantStore } from "../app/plantStore";
import type { ReviewStore } from "../app/reviewStore";
import type { SessionStore } from "../app/sessionStore";
import type { ThemeStore } from "../app/themeStore";
import type { WorkspaceStore } from "../app/workspaceStore";
import type { Route } from "../model/route";
import { HomeScreen } from "./home/HomeScreen";
import { PairingScreen } from "./pairing/PairingScreen";
import { PlantScreen } from "./plant/PlantScreen";
import { RunScreen } from "./run/RunScreen";
import { Shell } from "./shell/Shell";
import { TopBar } from "./shell/TopBar";
import { Unreachable } from "./shell/Unreachable";
import { useStore } from "./useStore";
import { useTheme } from "./useTheme";

export interface Stores {
  readonly themes: ThemeStore;
  readonly session: SessionStore;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly plants: PlantStore;
  readonly review: ReviewStore;
}

function useSession({ session, workspace }: Stores) {
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

function Screen({ stores, route }: { readonly stores: Stores; readonly route: Route }) {
  if (route.screen === "plant") {
    return (
      <PlantScreen
        plants={stores.plants}
        workspace={stores.workspace}
        navigation={stores.navigation}
        domain={route.domain}
        hall={route.hall}
      />
    );
  }
  if (route.screen === "run") {
    return (
      <RunScreen
        review={stores.review}
        plants={stores.plants}
        workspace={stores.workspace}
        navigation={stores.navigation}
        worldId={route.worldId}
      />
    );
  }
  return <HomeScreen workspace={stores.workspace} navigation={stores.navigation} />;
}

export function App({ stores }: { readonly stores: Stores }) {
  const { preference } = useTheme(stores.themes);
  const state = useSession(stores);
  const route = useStore(stores.navigation);
  if (state.phase === "unpaired") {
    return <PairingScreen session={stores.session} />;
  }
  const topBar = (
    <TopBar
      preference={preference}
      onCycleTheme={() => {
        stores.themes.cycle();
      }}
      onHome={() => {
        stores.navigation.go({ screen: "home" });
      }}
    />
  );
  return (
    <Shell topBar={topBar}>
      {state.phase === "checking" && <p role="status">Connecting to the bridge…</p>}
      {state.phase === "unreachable" && <Unreachable failure={state.failure} />}
      {state.phase === "paired" && <Screen stores={stores} route={route} />}
    </Shell>
  );
}
