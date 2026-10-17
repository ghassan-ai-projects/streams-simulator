import { render } from "@testing-library/react";
import { createEditorSession } from "../app/editor";
import { createLaunchStore } from "../app/launch";
import { createPlantMaker, createPlantSession } from "../app/plant";
import { createReviewSession } from "../app/review";
import { createSession } from "../app/session";
import { createNavigationStore, createPanelStore, createThemeStore } from "../app/shell";
import { createWorkspaceStore } from "../app/workspace";
import { App } from "../ui/shell";
import type { FakeBridge } from "./fakeBridge";
import { fakeClock } from "./fakeClock";
import { memoryLocation } from "./location";

function themes() {
  let value: string | null = null;
  return createThemeStore(
    { read: () => value, write: (_k, v) => (value = v) },
    { prefersDark: () => false, subscribe: () => () => undefined },
  );
}

function appStores(bridge: FakeBridge, hash = "#/") {
  const location = memoryLocation(hash);
  const timers = fakeClock();
  const shown: string[] = [];
  const stores = {
    themes: themes(),
    session: createSession(bridge.session),
    workspace: createWorkspaceStore(bridge.workspace),
    navigation: createNavigationStore(location),
    plants: createPlantSession(bridge.plant),
    review: createReviewSession(bridge.review, timers, timers),
    launch: createLaunchStore(bridge.launch, timers, timers),
    scheduler: timers,
    editor: createEditorSession(bridge.editor),
    panels: createPanelStore({ read: () => null, write: () => undefined }),
    maker: createPlantMaker(bridge.plant),
    showDetails: () => {
      shown.push("details");
    },
  };
  return { stores, location, timers, shown };
}

export function mountApp(bridge: FakeBridge, hash = "#/") {
  const mounted = appStores(bridge, hash);
  const view = render(<App stores={mounted.stores} />);
  return { ...mounted, view };
}
