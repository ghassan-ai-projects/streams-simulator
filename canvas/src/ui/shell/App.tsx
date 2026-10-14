import { useEffect } from "react";
import type { EditorStore } from "../../app/editorStore";
import type { PanelStore } from "../../app/panelStore";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantStore } from "../../app/plantStore";
import type { LaunchStore } from "../../app/launchStore";
import type { ReviewStore } from "../../app/reviewStore";
import type { SessionStore } from "../../app/sessionStore";
import type { ThemeStore } from "../../app/themeStore";
import type { WorkspaceStore } from "../../app/workspaceStore";
import type { Route } from "../../model/route";
import { EditorScreen } from "../editor/EditorScreen";
import { HomeScreen } from "../home/HomeScreen";
import { PairingScreen } from "../pairing/PairingScreen";
import { PlantScreen } from "../plant/PlantScreen";
import { RunScreen } from "../run/RunScreen";
import { KeyHelp } from "./KeyHelp";
import { PanelToggles } from "./PanelToggles";
import { RunIndicator } from "./RunIndicator";
import { Shell } from "./Shell";
import { TopBar } from "./TopBar";
import { Unreachable } from "./Unreachable";
import { useStore } from "../useStore";
import { useTheme } from "../useTheme";

export interface Stores {
  readonly themes: ThemeStore;
  readonly session: SessionStore;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly plants: PlantStore;
  readonly review: ReviewStore;
  readonly launch: LaunchStore;
  readonly editor: EditorStore;
  readonly panels: PanelStore;
}

function useSession({ session, workspace }: Stores) {
  const state = useStore(session);
  const listing = useStore(workspace);
  useEffect(() => {
    if (listing.phase === "failed" && listing.failure.code === "unpaired") void session.check();
  }, [listing, session]);
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
        launch={stores.launch}
        domain={route.domain}
        hall={route.hall}
        startLaunching={route.launch === true}
      />
    );
  }
  if (route.screen === "edit") {
    return <EditorScreen editor={stores.editor} navigation={stores.navigation} composition={route.composition} />;
  }
  if (route.screen === "run") {
    return (
      <RunScreen
        review={stores.review}
        plants={stores.plants}
        workspace={stores.workspace}
        navigation={stores.navigation}
        launch={stores.launch}
        worldId={route.worldId}
      />
    );
  }
  return (
    <HomeScreen
      workspace={stores.workspace}
      navigation={stores.navigation}
      editor={stores.editor}
      badgeFor={(worldId) => stores.review.badgeFor(worldId)}
    />
  );
}

export function App({ stores }: { readonly stores: Stores }) {
  const { preference } = useTheme(stores.themes);
  const state = useSession(stores);
  const route = useStore(stores.navigation);
  const panels = useStore(stores.panels);
  if (state.phase === "unpaired") {
    return <PairingScreen session={stores.session} />;
  }
  const topBar = (
    <TopBar
      status={<RunIndicator launch={stores.launch} navigation={stores.navigation} />}
      toggles={route.screen === "home" ? null : <PanelToggles panels={stores.panels} />}
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
    <Shell topBar={topBar} left={panels.left} right={panels.right}>
      {state.phase === "checking" && <p role="status">Connecting to the bridge…</p>}
      {state.phase === "unreachable" && <Unreachable failure={state.failure} />}
      {state.phase === "paired" && <Screen stores={stores} route={route} />}
      <KeyHelp />
    </Shell>
  );
}
