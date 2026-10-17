import { useEffect, useState } from "react";
import type { EditorSession } from "../../../app/editor";
import type { Scheduler } from "../../../app/ports";
import type { PlantMaker } from "../../../app/plant";
import type { NavigationStore, PanelStore, ThemeStore } from "../../../app/shell";
import { sessionGate, showBuild } from "../../../app/shell";
import type { PlantSession } from "../../../app/plant";
import type { LaunchStore } from "../../../app/launch";
import type { ReviewSession } from "../../../app/review";
import type { Session } from "../../../app/session";
import type { WorkspaceStore } from "../../../app/workspace";
import type { Route } from "../../../model/shell";
import { EditorScreen } from "../../editor";
import { WorkspaceScreen } from "../../workspace";
import { PairingScreen } from "../../session";
import { PlantScreen } from "../../plant";
import { RunScreen } from "../../review";
import { CrumbSlot } from "../../primitives";
import { KeyHelp } from "./KeyHelp";
import { PanelToggles } from "./PanelToggles";
import { rightPanelName } from "./panelName";
import { RunIndicator } from "./RunIndicator";
import { Shell } from "./Shell";
import { TopBar } from "./TopBar";
import { Unreachable } from "./Unreachable";
import { useStore } from "../../primitives";
import { useTheme } from "./useTheme";

export interface Stores {
  readonly themes: ThemeStore;
  readonly session: Session;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly plants: PlantSession;
  readonly review: ReviewSession;
  readonly launch: LaunchStore;
  readonly scheduler: Scheduler;
  readonly editor: EditorSession;
  readonly panels: PanelStore;
  readonly maker: PlantMaker;
  readonly showDetails: () => void;
}

function useSession({ session, workspace }: Stores) {
  const state = useStore(session);
  useEffect(() => sessionGate(session, workspace), [session, workspace]);
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
        onShowDetails={stores.showDetails}
      />
    );
  }
  if (route.screen === "edit") {
    return <EditorScreen editor={stores.editor} navigation={stores.navigation} composition={route.composition} />;
  }
  if (route.screen === "review") {
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
    <WorkspaceScreen
      workspace={stores.workspace}
      navigation={stores.navigation}
      maker={stores.maker}
      badgeFor={(worldId) => stores.review.badgeFor(worldId)}
    />
  );
}

function ShellToggles({ stores, route }: { readonly stores: Stores; readonly route: Route }) {
  const name = rightPanelName[route.screen];
  if (route.screen !== "edit") return <PanelToggles panels={stores.panels} right={name} />;
  return (
    <PanelToggles
      panels={stores.panels}
      right={name}
      onRightPress={() => {
        showBuild(stores.editor, stores.panels);
      }}
    />
  );
}

export function App({ stores }: { readonly stores: Stores }) {
  const { preference } = useTheme(stores.themes);
  const state = useSession(stores);
  const route = useStore(stores.navigation);
  const panels = useStore(stores.panels);
  const [slot, setSlot] = useState<HTMLElement | null>(null);
  if (state.phase === "unpaired") {
    return <PairingScreen session={stores.session} />;
  }
  const topBar = (
    <TopBar
      crumbs={setSlot}
      help={<KeyHelp screen={route.screen} />}
      status={<RunIndicator launch={stores.launch} navigation={stores.navigation} />}
      toggles={route.screen === "workspace" ? null : <ShellToggles stores={stores} route={route} />}
      preference={preference}
      onCycleTheme={() => {
        stores.themes.cycle();
      }}
      onWorkspace={() => {
        stores.navigation.go({ screen: "workspace" });
      }}
    />
  );
  return (
    <CrumbSlot value={slot}>
      <Shell topBar={topBar} left={panels.left} right={panels.right}>
        {state.phase === "checking" && <p role="status">Connecting to the bridge…</p>}
        {state.phase === "unreachable" && <Unreachable failure={state.failure} />}
        {state.phase === "paired" && <Screen stores={stores} route={route} />}
      </Shell>
    </CrumbSlot>
  );
}
