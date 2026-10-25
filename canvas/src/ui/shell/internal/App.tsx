import { useEffect, useState } from "react";
import { resumeRun, sessionGate, showBuild } from "../../../app/shell";
import type { Route } from "../../../model/shell";
import { DeviceScreen } from "../../devices";
import { EditorScreen } from "../../editor";
import { CaptureScreen, WorkspaceScreen } from "../../workspace";
import { PairingScreen } from "../../session";
import { AssumptionsDisclosure } from "../../plant";
import { PlantRoute } from "./PlantRoute";
import { CrumbSlot } from "../../primitives";
import { KeyHelp } from "./KeyHelp";
import { KeepLiveSession, ReviewRoute } from "./LiveRoutes";
import { PanelToggles } from "./PanelToggles";
import { rightPanelName } from "./panelName";
import { RunIndicator } from "./RunIndicator";
import type { Stores } from "./stores";
import { Shell } from "./Shell";
import { TopBar } from "./TopBar";
import { Unreachable } from "./Unreachable";
import { useStore } from "../../primitives";
import { useTheme } from "./useTheme";

function useSession({ session, workspace, launch }: Stores) {
  const state = useStore(session);
  useEffect(() => sessionGate(session, workspace), [session, workspace]);
  useEffect(() => resumeRun(launch, workspace), [launch, workspace]);
  return state;
}

function Authoring({
  stores,
  route,
}: {
  readonly stores: Stores;
  readonly route: Extract<Route, { readonly screen: "edit" | "device" }>;
}) {
  if (route.screen === "edit") {
    return <EditorScreen editor={stores.editor} navigation={stores.navigation} composition={route.composition} />;
  }
  return (
    <DeviceScreen
      devices={stores.devices}
      navigation={stores.navigation}
      path={route.path}
      assumptions={(domain, twin) => (
        <AssumptionsDisclosure session={stores.provenance.assumptions} path={domain} twin={twin} />
      )}
      {...(route.step === undefined ? {} : { step: route.step })}
    />
  );
}

function Home({ stores, route }: { readonly stores: Stores; readonly route: Route }) {
  if (route.screen === "capture") {
    return <CaptureScreen workspace={stores.workspace} desk={stores.desk} navigation={stores.navigation} />;
  }
  return (
    <WorkspaceScreen
      workspace={stores.workspace}
      navigation={stores.navigation}
      maker={stores.maker}
      desk={stores.desk}
      provenance={stores.provenance}
      badgeFor={(worldId) => stores.review.badgeFor(worldId)}
    />
  );
}

function Screen({ stores, route }: { readonly stores: Stores; readonly route: Route }) {
  if (route.screen === "plant") {
    return <PlantRoute stores={stores} route={route} />;
  }
  if (route.screen === "edit" || route.screen === "device") return <Authoring stores={stores} route={route} />;
  if (route.screen === "review")
    return <ReviewRoute stores={stores} worldId={route.worldId} model={route.model === true} />;
  return <Home stores={stores} route={route} />;
}

function withoutPanels(route: Route): boolean {
  return route.screen === "workspace" || route.screen === "capture";
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
      toggles={withoutPanels(route) ? null : <ShellToggles stores={stores} route={route} />}
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
        {state.phase === "paired" && <KeepLiveSession stores={stores} />}
        {state.phase === "paired" && <Screen stores={stores} route={route} />}
      </Shell>
    </CrumbSlot>
  );
}
