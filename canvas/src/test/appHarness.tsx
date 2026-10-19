import { render } from "@testing-library/react";
import { createDeviceDesk, createDeviceSession } from "../app/devices";
import { createEditorSession } from "../app/editor";
import { createLaunchStore } from "../app/launch";
import { createPlantMaker, createPlantSession } from "../app/plant";
import { createReviewSession } from "../app/review";
import { createSession } from "../app/session";
import { createNavigationStore, createPanelStore, createThemeStore } from "../app/shell";
import type { Clipboard } from "../app/ports";
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

function appStores(bridge: FakeBridge, hash = "#/", clipboard?: Clipboard) {
  const location = memoryLocation(hash);
  const timers = fakeClock();
  const shown: string[] = [];
  const stores = {
    themes: themes(),
    session: createSession(bridge.session),
    workspace: createWorkspaceStore(bridge.workspace),
    navigation: createNavigationStore(location),
    plants: createPlantSession(bridge.plant),
    review: createReviewSession(bridge.review, timers, timers, clipboard),
    liveReview: createReviewSession(bridge.review, timers, timers, clipboard),
    launch: createLaunchStore(bridge.launch, timers, timers),
    scheduler: timers,
    editor: createEditorSession(bridge.editor),
    devices: createDeviceSession(bridge.devices, bridge.workspace.watch),
    desk: createDeviceDesk(bridge.devices),
    panels: createPanelStore({ read: () => null, write: () => undefined }),
    maker: createPlantMaker(bridge.plant),
    showDetails: () => {
      shown.push("details");
    },
  };
  return { stores, location, timers, shown };
}

export function mountApp(bridge: FakeBridge, hash = "#/", clipboard?: Clipboard) {
  const mounted = appStores(bridge, hash, clipboard);
  const view = render(<App stores={mounted.stores} />);
  return { ...mounted, view };
}
